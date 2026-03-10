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

	"github.com/unstablebuild/rune-go-sdk/api/browserapi/browserrpc"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/debug"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// subset of Editor_SubscribeCommandClient
type commandClientStream interface {
	RecvMsg(any) error
	Send(*ClientCommandMessage) error
	CloseSend() error
}

type commandServerStream struct {
	ctx       context.Context
	cancelCtx func()
	stream    commandClientStream
	sendChan  chan *ClientCommandMessage
	h         textapi.CommandHandler
}

func newCommandServerStream(
	ctx context.Context,
	stream commandClientStream,
	h textapi.CommandHandler,
) *commandServerStream {
	ctx, cancelCtx := context.WithCancel(ctx)
	return &commandServerStream{
		stream:    stream,
		h:         h,
		cancelCtx: cancelCtx,
		ctx:       ctx,
		sendChan:  make(chan *ClientCommandMessage),
	}
}

func (s *commandServerStream) sendMessages() {
	defer func() {
		_ = s.stream.CloseSend()
	}()
	for {
		select {
		case <-s.ctx.Done():
			return
		case msg := <-s.sendChan:
			err := s.stream.Send(msg)
			if err != nil {
				slog.Error(
					"unable to send stream message back, "+
						"terminating stream",
					"error", err,
				)
				s.cancelCtx()
				return
			}
		}
	}
}

func (s *commandServerStream) receiveMessages() {
	go debug.CapturePanicReport(s.sendMessages)
	defer s.cancelCtx()
	for {
		var reqMsg ServerCommandMessage
		err := s.stream.RecvMsg(&reqMsg)
		if err != nil {
			if !errors.Is(err, io.EOF) &&
				status.Code(err) != codes.Canceled {
				slog.Error(
					"receive server command message",
					"error", err,
				)
			}
			return
		}
		tpe := reqMsg.GetType()
		switch tpe {
		case ServerCommandMessage_Handle:
			s.handleCommand(reqMsg.GetHandle())
		case ServerCommandMessage_Complete:
			err = s.handleComplete(reqMsg.GetComplete())
			if err != nil {
				var resp CompleteCommandDone
				resp.Error = err.Error()
				resp.Id = reqMsg.GetComplete().GetId()
				msg := &ClientCommandMessage{
					Type:         ClientCommandMessage_CompleteDone,
					CompleteDone: &resp,
				}
				select {
				case <-s.ctx.Done():
					return
				case s.sendChan <- msg:
				}
			}
		default:
			slog.Error(
				"extraneous server command message",
				"type", tpe,
			)
			return
		}
	}
}

func (s *commandServerStream) handleCommand(
	req *HandleCommandRequest,
) {
	err := s.doHandleCommand(req)
	var res HandleCommandResponse
	if err != nil {
		res.Error = err.Error()
	}

	var respMsg ClientCommandMessage
	respMsg.Type = ClientCommandMessage_Handle
	respMsg.Handle = &res
	select {
	case s.sendChan <- &respMsg:
	case <-s.ctx.Done():
	}
}

func (s *commandServerStream) doHandleCommand(
	req *HandleCommandRequest,
) error {
	var cmd textapi.Command
	err := s.commandFromProto(&cmd, req)
	if err != nil {
		return fmt.Errorf("command from protobuf: %w", err)
	}

	// NOTE: commands cancels are not currently being
	// propagated from client to server. We must add an extra
	// message that we handle here to do so.
	return s.h.HandleCommand(s.ctx, cmd)
}

func (s *commandServerStream) handleComplete(
	req *CompleteCommandRequest,
) error {
	if req.Name == "" {
		return errors.New(
			"invalid complete request: missing command name",
		)
	}
	if req.Id == 0 {
		return errors.New(
			"invalid complete request: missing request id",
		)
	}

	slog.Debug(
		"streaming completer's iterator",
		"name", req.Name, "args", req.Args,
	)
	defer slog.Debug(
		"done streaming completer's iterator",
		"name", req.Name, "args", req.Args,
	)

	ctx := s.ctx
	completer, err := s.h.Complete(ctx, req.Name, req.Args)
	if err != nil {
		return err
	}

	id := req.Id
	go debug.CapturePanicReport(func() {
		var resp CompleteCommandDone
		err := s.streamValues(id, completer)
		if err != nil {
			resp.Error = err.Error()
		}
		resp.Id = id
		msg := &ClientCommandMessage{
			Type:         ClientCommandMessage_CompleteDone,
			CompleteDone: &resp,
		}
		select {
		case <-s.ctx.Done():
		case s.sendChan <- msg:
		}
	})

	return nil
}

func (s *commandServerStream) streamValues(
	id int64,
	completer iterator.Iterator[string],
) error {
	for {
		next, ok := completer.Next(s.ctx)
		if !ok {
			return completer.Err()
		}
		resp := CompleteCommandValue{Id: id, Value: next}
		respMsg := &ClientCommandMessage{
			CompleteValue: &resp,
			Type:          ClientCommandMessage_CompleteValue,
		}
		select {
		case s.sendChan <- respMsg:
		case <-s.ctx.Done():
			return s.ctx.Err()
		}
	}
}

func (s *commandServerStream) commandFromProto(
	e *textapi.Command, pe *HandleCommandRequest,
) (err error) {
	if pe.ResourceName != nil {
		e.URI, err = NewURIFromProto(pe.GetResourceName())
		if err != nil {
			return
		}
		e.Resource = Token{
			URI: e.URI,
		}
	}
	e.Cursor.Window = pe.GetCursorWindow().ToModel()
	e.Cursor.Content = pe.GetCursorContent().ToModel()
	e.Args = pe.GetArgs()
	e.Name = pe.GetName()
	e.Window = browserrpc.NewWindow(
		uint64(pe.GetWindowId()),
	)
	return err
}
