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

package textrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/unstablebuild/rune-go-sdk/api/browserapi/browserrpc"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/debug"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// subset of Editor_SubscribeResourceOpenerClient
type resourceOpenerClientStream interface {
	RecvMsg(any) error
	Send(*ClientResourceOpenerMessage) error
	CloseSend() error
}

type resourceOpenerServerStream struct {
	ctx       context.Context
	cancelCtx func()
	stream    resourceOpenerClientStream
	sendChan  chan *ClientResourceOpenerMessage
	h         textapi.ResourceOpenHandler

	// open request id -> cancel func of its handler's context
	opens sync.Map
}

func newResourceOpenerServerStream(
	ctx context.Context,
	stream resourceOpenerClientStream,
	h textapi.ResourceOpenHandler,
) *resourceOpenerServerStream {
	ctx, cancelCtx := context.WithCancel(ctx)
	return &resourceOpenerServerStream{
		ctx:       ctx,
		cancelCtx: cancelCtx,
		stream:    stream,
		sendChan:  make(chan *ClientResourceOpenerMessage),
		h:         h,
	}
}

func (s *resourceOpenerServerStream) sendMessages() {
	defer func() {
		_ = s.stream.CloseSend()
	}()
	for {
		select {
		case <-s.ctx.Done():
			return
		case msg := <-s.sendChan:
			if err := s.stream.Send(msg); err != nil {
				slog.Error(
					"unable to send resource opener message, "+
						"terminating stream",
					"error", err,
				)
				s.cancelCtx()
				return
			}
		}
	}
}

func (s *resourceOpenerServerStream) receiveMessages() {
	go debug.CapturePanicReport(s.sendMessages)
	defer s.cancelCtx()
	for {
		var msg ServerResourceOpenerMessage
		if err := s.stream.RecvMsg(&msg); err != nil {
			if !errors.Is(err, io.EOF) &&
				status.Code(err) != codes.Canceled {
				slog.Error(
					"receive server resource opener message",
					"error", err,
				)
			}
			return
		}
		switch tpe := msg.GetType(); tpe {
		case ServerResourceOpenerMessage_Open:
			s.open(msg.GetOpen())
		case ServerResourceOpenerMessage_OpenCancel:
			cancelRequest(&s.opens, msg.GetOpenCancel().GetId())
		default:
			slog.Error(
				"extraneous server resource opener message",
				"type", tpe,
			)
			return
		}
	}
}

// open runs the handler off the receive loop, which must stay free to take
// the editor's cancel for this very request.
func (s *resourceOpenerServerStream) open(req *OpenResourceRequest) {
	id := req.GetId()
	ctx, cancel := context.WithCancel(s.ctx)
	s.opens.Store(id, context.CancelFunc(cancel))
	go debug.CapturePanicReport(func() {
		defer func() {
			s.opens.Delete(id)
			cancel()
		}()
		s.sendResponse(id, s.runOpen(ctx, req))
	})
}

func (s *resourceOpenerServerStream) runOpen(
	ctx context.Context, req *OpenResourceRequest,
) error {
	uri, err := NewURIFromProto(req.GetUri())
	if err != nil {
		return fmt.Errorf("uri from protobuf: %w", err)
	}
	return s.h.OpenResource(ctx, uri, browserrpc.NewWindow(req.GetWindowId()))
}

func (s *resourceOpenerServerStream) sendResponse(id int64, err error) {
	res := OpenResourceResponse{Id: id}
	if err != nil {
		res.Error = err.Error()
	}
	msg := &ClientResourceOpenerMessage{
		Type: ClientResourceOpenerMessage_Open,
		Open: &res,
	}
	select {
	case s.sendChan <- msg:
	case <-s.ctx.Done():
	}
}
