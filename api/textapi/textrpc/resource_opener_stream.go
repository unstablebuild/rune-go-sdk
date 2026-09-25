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

	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/debug"
	handlerrpc "github.com/unstablebuild/rune-go-sdk/handler/handlerrpc"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// subset of Editor_SubscribeResourceOpenerClient
type resourceOpenerClientStream interface {
	RecvMsg(any) error
	Send(*ClientResourceOpenerMessage) error
	CloseSend() error
}

// handlerStreamOpener opens the stream a resource's handler is served on.
type handlerStreamOpener interface {
	OpenResource(ctx context.Context, opts ...grpc.CallOption) (
		Editor_OpenResourceClient, error,
	)
}

type resourceOpenerServerStream struct {
	ctx       context.Context
	cancelCtx func()
	stream    resourceOpenerClientStream
	sendChan  chan *ClientResourceOpenerMessage
	h         textapi.ResourceOpenHandler
	ed        handlerStreamOpener

	// open request id -> cancel func of its handler's context
	opens sync.Map
}

func newResourceOpenerServerStream(
	ctx context.Context,
	stream resourceOpenerClientStream,
	ed handlerStreamOpener,
	h textapi.ResourceOpenHandler,
) *resourceOpenerServerStream {
	ctx, cancelCtx := context.WithCancel(ctx)
	return &resourceOpenerServerStream{
		ctx:       ctx,
		cancelCtx: cancelCtx,
		stream:    stream,
		sendChan:  make(chan *ClientResourceOpenerMessage),
		h:         h,
		ed:        ed,
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
// the editor's cancel for this very request. A failure is reported on the
// opener stream; a handler is served on its own stream, which is what
// reports success.
func (s *resourceOpenerServerStream) open(req *OpenResourceRequest) {
	id := req.GetId()
	ctx, cancel := context.WithCancel(s.ctx)
	s.opens.Store(id, context.CancelFunc(cancel))
	go debug.CapturePanicReport(func() {
		defer func() {
			s.opens.Delete(id)
			cancel()
		}()
		h, err := s.runOpen(ctx, req)
		if err == nil {
			err = s.serve(id, h)
			if err != nil {
				_ = h.Close()
			}
		}
		if err != nil {
			s.sendResponse(id, err)
		}
	})
}

func (s *resourceOpenerServerStream) runOpen(
	ctx context.Context, req *OpenResourceRequest,
) (browserapi.Handler, error) {
	uri, err := NewURIFromProto(req.GetUri())
	if err != nil {
		return nil, fmt.Errorf("uri from protobuf: %w", err)
	}
	h, err := s.h.OpenResource(ctx, uri)
	if err != nil {
		return nil, err
	}
	if h == nil {
		return nil, errors.New("resource opener returned no handler")
	}
	return h, nil
}

// serve installs h as the answer to the open request id, on a handler
// stream that outlives the request, the same way a tab's handler is served.
func (s *resourceOpenerServerStream) serve(id int64, h browserapi.Handler) error {
	stream, err := s.ed.OpenResource(s.ctx)
	if err != nil {
		return fmt.Errorf("open resource stream: %w", err)
	}
	sendMsg := OpenResourceMessage{
		Type:    handlerrpc.MessageType_Request,
		Request: &OpenResourceStreamRequest{Id: id},
	}
	if err := stream.SendMsg(&sendMsg); err != nil {
		return fmt.Errorf("send open resource request: %w", err)
	}
	var recvMsg handlerrpc.ServerMessage
	if err := stream.RecvMsg(&recvMsg); err != nil {
		return fmt.Errorf("recv open resource response: %w", err)
	}
	if recvMsg.GetResponse() == nil {
		return fmt.Errorf("recv nil open resource response: %v", &recvMsg)
	}
	server := handlerrpc.NewServerStream(stream, h, func() *OpenResourceMessage {
		return new(OpenResourceMessage)
	})
	go debug.CapturePanicReport(server.ReceiveMessages)
	return nil
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
