// Copyright 2026 Unstable Build, LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package acprpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	acp "github.com/coder/acp-go-sdk"
	"github.com/unstablebuild/rune-go-sdk/api/acpapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ acpapi.Service = (*Client)(nil)

// Client satisfies acpapi.Service via gRPC.
type Client struct {
	pb  ACPClient
	ctx context.Context
}

// NewClient returns a Client that calls the ACP service on cc. ctx bounds
// the event streams of the sessions it returns.
func NewClient(ctx context.Context, cc grpc.ClientConnInterface) *Client {
	return &Client{pb: NewACPClient(cc), ctx: ctx}
}

// Agents satisfies acpapi.Service.
func (c *Client) Agents(ctx context.Context) iterator.Iterator[acpapi.AgentInfo] {
	callCtx, cancel := context.WithCancel(ctx)
	stream, err := c.pb.Agents(callCtx, &AgentsRequest{})
	if err != nil {
		cancel()
		return iterator.Error[acpapi.AgentInfo](fromStatus(err))
	}
	return &streamIterator[AgentInfo, acpapi.AgentInfo]{
		stream: stream,
		cancel: cancel,
		conv: func(a *AgentInfo) (acpapi.AgentInfo, error) {
			return acpapi.AgentInfo{Name: a.GetName(), Running: a.GetRunning()}, nil
		},
	}
}

// Sessions satisfies acpapi.Service.
func (c *Client) Sessions(ctx context.Context) iterator.Iterator[acpapi.SessionInfo] {
	callCtx, cancel := context.WithCancel(ctx)
	stream, err := c.pb.Sessions(callCtx, &SessionsRequest{})
	if err != nil {
		cancel()
		return iterator.Error[acpapi.SessionInfo](fromStatus(err))
	}
	return &streamIterator[SessionInfo, acpapi.SessionInfo]{
		stream: stream,
		cancel: cancel,
		conv:   func(s *SessionInfo) (acpapi.SessionInfo, error) { return fromProtoInfo(s), nil },
	}
}

// NewSession satisfies acpapi.Service.
func (c *Client) NewSession(ctx context.Context, req acpapi.NewSessionRequest) (acpapi.Session, error) {
	b, err := json.Marshal(req.Session)
	if err != nil {
		return nil, fmt.Errorf("acp: encode session/new: %w", err)
	}
	info, err := c.pb.NewSession(ctx, &NewSessionRequest{Agent: req.Agent, Name: req.Name, Session: b})
	if err != nil {
		return nil, fromStatus(err)
	}
	return &session{c: c, info: fromProtoInfo(info)}, nil
}

// OpenSession satisfies acpapi.Service.
func (c *Client) OpenSession(ctx context.Context, name string) (acpapi.Session, error) {
	info, err := c.pb.OpenSession(ctx, &OpenSessionRequest{Name: name})
	if err != nil {
		return nil, fromStatus(err)
	}
	return &session{c: c, info: fromProtoInfo(info)}, nil
}

type session struct {
	c    *Client
	info acpapi.SessionInfo
}

func (s *session) Info() acpapi.SessionInfo { return s.info }

func (s *session) Events() iterator.Iterator[acpapi.Event] {
	ctx, cancel := context.WithCancel(s.c.ctx)
	stream, err := s.c.pb.Events(ctx, &EventsRequest{Session: s.info.Name})
	if err != nil {
		cancel()
		return iterator.Error[acpapi.Event](fromStatus(err))
	}
	it := &eventIterator{cancel: cancel, ch: make(chan eventOrErr, 64), done: make(chan struct{})}
	go it.pump(stream)
	return it
}

func (s *session) Prompt(ctx context.Context, content []acp.ContentBlock) error {
	b, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("acp: encode prompt: %w", err)
	}
	_, err = s.c.pb.Prompt(ctx, &PromptRequest{Session: s.info.Name, Content: b})
	return fromStatus(err)
}

func (s *session) Cancel(ctx context.Context) error {
	_, err := s.c.pb.Cancel(ctx, &CancelRequest{Session: s.info.Name})
	return fromStatus(err)
}

func (s *session) Respond(ctx context.Context, id acpapi.PermissionID, o acp.RequestPermissionOutcome) error {
	b, err := json.Marshal(o)
	if err != nil {
		return fmt.Errorf("acp: encode outcome: %w", err)
	}
	_, err = s.c.pb.Respond(ctx, &RespondRequest{Session: s.info.Name, PermissionId: int64(id), Outcome: b})
	return fromStatus(err)
}

func (s *session) SetConfigOption(ctx context.Context, req acp.SetSessionConfigOptionRequest) error {
	b, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("acp: encode config option: %w", err)
	}
	_, err = s.c.pb.SetConfigOption(ctx, &SetConfigOptionRequest{Session: s.info.Name, Request: b})
	return fromStatus(err)
}

func (s *session) Close() error {
	_, err := s.c.pb.CloseSession(s.c.ctx, &CloseSessionRequest{Session: s.info.Name})
	return fromStatus(err)
}

type eventOrErr struct {
	ev  acpapi.Event
	err error
}

// eventIterator receives on its own goroutine so Next's ctx can bound the
// wait without ending the stream.
type eventIterator struct {
	cancel context.CancelFunc
	ch     chan eventOrErr
	done   chan struct{}
	once   sync.Once
	ended  bool
	err    error
}

func (it *eventIterator) pump(stream grpc.ServerStreamingClient[Event]) {
	defer close(it.ch)
	for {
		msg, err := stream.Recv()
		var item eventOrErr
		switch {
		case errors.Is(err, io.EOF):
			return
		case err != nil:
			item.err = fromEventsStatus(err)
		default:
			if jerr := json.Unmarshal(msg.GetEvent(), &item.ev); jerr != nil {
				item.err = fmt.Errorf("acp: decode event: %w", jerr)
			}
		}
		select {
		case it.ch <- item:
		case <-it.done:
			return
		}
		if item.err != nil {
			return
		}
	}
}

func (it *eventIterator) Next(ctx context.Context) (acpapi.Event, bool) {
	if it.ended {
		return acpapi.Event{}, false
	}
	var item eventOrErr
	var ok bool
	select {
	case item, ok = <-it.ch:
	default:
		select {
		case item, ok = <-it.ch:
		case <-it.done:
			it.ended = true
			it.err = nil
			return acpapi.Event{}, false
		case <-ctx.Done():
			it.err = ctx.Err()
			return acpapi.Event{}, false
		}
	}
	if !ok || it.closed() {
		it.ended = true
		it.err = nil
		return acpapi.Event{}, false
	}
	if item.err != nil {
		it.ended = true
		it.err = item.err
		return acpapi.Event{}, false
	}
	it.err = nil
	return item.ev, true
}

func (it *eventIterator) Err() error { return it.err }

func (it *eventIterator) Close() error {
	it.once.Do(func() {
		close(it.done)
		it.cancel()
	})
	return nil
}

func (it *eventIterator) closed() bool {
	select {
	case <-it.done:
		return true
	default:
		return false
	}
}

type streamIterator[P any, T any] struct {
	stream grpc.ServerStreamingClient[P]
	cancel context.CancelFunc
	conv   func(*P) (T, error)
	err    error
	done   bool
}

func (it *streamIterator[P, T]) Next(context.Context) (T, bool) {
	var zero T
	if it.done {
		return zero, false
	}
	msg, err := it.stream.Recv()
	if err != nil {
		it.done = true
		it.cancel()
		if !errors.Is(err, io.EOF) {
			it.err = fromStatus(err)
		}
		return zero, false
	}
	v, err := it.conv(msg)
	if err != nil {
		it.done = true
		it.cancel()
		it.err = err
		return zero, false
	}
	return v, true
}

func (it *streamIterator[P, T]) Err() error { return it.err }

func (it *streamIterator[P, T]) Close() error {
	it.done = true
	it.cancel()
	return nil
}

func fromProtoInfo(s *SessionInfo) acpapi.SessionInfo {
	return acpapi.SessionInfo{Name: s.GetName(), Agent: s.GetAgent(), ID: acp.SessionId(s.GetId())}
}

// remoteError keeps the server's message while matching the sentinel it was
// sent for.
type remoteError struct {
	msg string
	is  error
}

func (e *remoteError) Error() string { return e.msg }
func (e *remoteError) Unwrap() error { return e.is }

func fromStatus(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	for _, s := range sentinels {
		if st.Code() == s.code && strings.Contains(st.Message(), s.err.Error()) {
			return &remoteError{msg: st.Message(), is: s.err}
		}
	}
	switch st.Code() {
	case codes.Canceled:
		return &remoteError{msg: st.Message(), is: context.Canceled}
	case codes.DeadlineExceeded:
		return &remoteError{msg: st.Message(), is: context.DeadlineExceeded}
	}
	return err
}

// fromEventsStatus maps the status an event stream ends with; UNAVAILABLE
// carries the cause the session ended with.
func fromEventsStatus(err error) error {
	if st, ok := status.FromError(err); ok && st.Code() == codes.Unavailable {
		return errors.New(st.Message())
	}
	return fromStatus(err)
}
