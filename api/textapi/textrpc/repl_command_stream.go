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
	"io"
	"log/slog"
	"sync"

	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/debug"
	"github.com/unstablebuild/rune-go-sdk/handler/repl"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/term/termrpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// subset of Editor_SubscribeREPLCommandClient
type replClientStream interface {
	RecvMsg(any) error
	Send(*ClientREPLCommandMessage) error
	CloseSend() error
}

type replCommandServerStream struct {
	ctx       context.Context
	cancelCtx func()
	stream    replClientStream
	sendChan  chan *ClientREPLCommandMessage
	h         textapi.REPLHandler

	// completion id -> cancel func of its streaming context
	completions sync.Map
	// handle request id -> cancel func of its command's context
	handles sync.Map
	// help request id -> cancel func of its help's context
	helps sync.Map
}

func newREPLCommandServerStream(
	ctx context.Context,
	stream replClientStream,
	h textapi.REPLHandler,
) *replCommandServerStream {
	ctx, cancelCtx := context.WithCancel(ctx)
	return &replCommandServerStream{
		stream:    stream,
		h:         h,
		cancelCtx: cancelCtx,
		ctx:       ctx,
		sendChan:  make(chan *ClientREPLCommandMessage),
	}
}

func (s *replCommandServerStream) sendMessages() {
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
					"unable to send repl stream message, "+
						"terminating stream",
					"error", err,
				)
				s.cancelCtx()
				return
			}
		}
	}
}

func (s *replCommandServerStream) receiveMessages() {
	go debug.CapturePanicReport(s.sendMessages)
	defer s.cancelCtx()
	for {
		var reqMsg ServerREPLCommandMessage
		err := s.stream.RecvMsg(&reqMsg)
		if err != nil {
			if !errors.Is(err, io.EOF) &&
				status.Code(err) != codes.Canceled {
				slog.Error(
					"receive server repl command message",
					"error", err,
				)
			}
			return
		}
		tpe := reqMsg.GetType()
		switch tpe {
		case ServerREPLCommandMessage_Handle:
			s.handleCommand(reqMsg.GetHandle())
		case ServerREPLCommandMessage_Complete:
			err = s.handleComplete(reqMsg.GetComplete())
			if err != nil {
				var resp CompleteCommandDone
				resp.Error = err.Error()
				resp.Id = reqMsg.GetComplete().GetId()
				msg := &ClientREPLCommandMessage{
					Type:         ClientREPLCommandMessage_CompleteDone,
					CompleteDone: &resp,
				}
				select {
				case <-s.ctx.Done():
					return
				case s.sendChan <- msg:
				}
			}
		case ServerREPLCommandMessage_Help:
			s.handleHelp(reqMsg.GetHelp())
		case ServerREPLCommandMessage_CompleteCancel:
			s.cancelComplete(reqMsg.GetCompleteCancel().GetId())
		case ServerREPLCommandMessage_HandleCancel:
			cancelRequest(&s.handles, reqMsg.GetHandleCancel().GetId())
		case ServerREPLCommandMessage_HelpCancel:
			cancelRequest(&s.helps, reqMsg.GetHelpCancel().GetId())
		default:
			slog.Error(
				"extraneous server repl command message",
				"type", tpe,
			)
			return
		}
	}
}

// handleCommand dispatches off the receive loop, which must stay free to
// take the editor's cancel for this very command. The handler therefore
// only ever overlaps with handlers that have already been cancelled.
func (s *replCommandServerStream) handleCommand(
	req *HandleREPLCommandRequest,
) {
	cmd := repl.Command{
		Name: req.GetName(),
		Args: req.GetArgs(),
	}
	width := int(req.GetWidth())
	id := req.GetId()

	ctx, cancel := context.WithCancel(s.ctx)
	s.handles.Store(id, context.CancelFunc(cancel))
	go debug.CapturePanicReport(func() {
		defer func() {
			s.handles.Delete(id)
			cancel()
		}()

		pw := &grpcProgressWriter{stream: s, ctx: ctx, id: id}
		iter, err := s.h.HandleCommand(ctx, cmd, pw)
		if err != nil {
			s.sendHandleDone(id, err)
			return
		}
		defer func() { _ = iter.Close() }()

		err = s.streamResponsive(
			ctx, iter, width,
			func(rows []*termrpc.CellRow) *ClientREPLCommandMessage {
				return &ClientREPLCommandMessage{
					Type: ClientREPLCommandMessage_HandleValue,
					HandleValue: &HandleREPLCommandValue{
						Rows: rows,
						Id:   id,
					},
				}
			},
		)
		s.sendHandleDone(id, err)
	})
}

func cancelRequest(m *sync.Map, id int64) {
	if cancel, ok := m.LoadAndDelete(id); ok {
		cancel.(context.CancelFunc)()
	}
}

// grpcProgressWriter is a repl.ProgressWriter that forwards Progress
// calls to the client over the REPL gRPC stream, allowing the editor's
// real progress bar to render progress reported by an out-of-process
// extension's command handler.
type grpcProgressWriter struct {
	stream *replCommandServerStream
	ctx    context.Context
	id     int64
}

func (w *grpcProgressWriter) Progress(
	progress, total int64, units string,
) {
	msg := &ClientREPLCommandMessage{
		Type: ClientREPLCommandMessage_HandleProgress,
		HandleProgress: &HandleREPLCommandProgress{
			Progress: progress,
			Total:    total,
			Units:    units,
			Id:       w.id,
		},
	}
	// Non-blocking send: progress is informational. If the client falls
	// behind or the stream is shutting down, drop the sample rather than
	// stall the handler goroutine.
	select {
	case w.stream.sendChan <- msg:
	case <-w.ctx.Done():
	}
}

func (s *replCommandServerStream) sendHandleDone(id int64, err error) {
	resp := HandleREPLCommandDone{Id: id}
	if err != nil {
		resp.Error = err.Error()
	}
	msg := &ClientREPLCommandMessage{
		Type:       ClientREPLCommandMessage_HandleDone,
		HandleDone: &resp,
	}
	select {
	case s.sendChan <- msg:
	case <-s.ctx.Done():
	}
}

func (s *replCommandServerStream) handleHelp(
	req *HelpCommandRequest,
) {
	width := int(req.GetWidth())
	id := req.GetId()

	ctx, cancel := context.WithCancel(s.ctx)
	s.helps.Store(id, context.CancelFunc(cancel))
	go debug.CapturePanicReport(func() {
		defer func() {
			s.helps.Delete(id)
			cancel()
		}()

		iter, err := s.h.Help(ctx, req.GetArgs())
		if err != nil {
			s.sendHelpDone(id, err)
			return
		}
		defer func() { _ = iter.Close() }()

		err = s.streamResponsive(
			ctx, iter, width,
			func(rows []*termrpc.CellRow) *ClientREPLCommandMessage {
				return &ClientREPLCommandMessage{
					Type: ClientREPLCommandMessage_HelpValue,
					HelpValue: &HelpCommandValue{
						Rows: rows,
						Id:   id,
					},
				}
			},
		)
		s.sendHelpDone(id, err)
	})
}

func (s *replCommandServerStream) sendHelpDone(id int64, err error) {
	resp := HelpCommandDone{Id: id}
	if err != nil {
		resp.Error = err.Error()
	}
	msg := &ClientREPLCommandMessage{
		Type:     ClientREPLCommandMessage_HelpDone,
		HelpDone: &resp,
	}
	select {
	case s.sendChan <- msg:
	case <-s.ctx.Done():
	}
}

func (s *replCommandServerStream) cancelComplete(id int64) {
	if cancel, ok := s.completions.LoadAndDelete(id); ok {
		cancel.(context.CancelFunc)()
	}
}

func (s *replCommandServerStream) handleComplete(
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

	ctx, cancelCtx := context.WithCancel(s.ctx)
	completer, err := s.h.Complete(ctx, req.Name, req.Args)
	if err != nil {
		cancelCtx()
		return err
	}

	id := req.Id
	s.completions.Store(id, context.CancelFunc(cancelCtx))
	go debug.CapturePanicReport(func() {
		defer func() {
			s.completions.Delete(id)
			cancelCtx()
			_ = completer.Close()
		}()
		var resp CompleteCommandDone
		err := s.streamCompleteValues(ctx, id, completer)
		if err != nil && ctx.Err() == nil {
			resp.Error = err.Error()
		}
		resp.Id = id
		msg := &ClientREPLCommandMessage{
			Type:         ClientREPLCommandMessage_CompleteDone,
			CompleteDone: &resp,
		}
		select {
		case <-s.ctx.Done():
		case s.sendChan <- msg:
		}
	})

	return nil
}

func (s *replCommandServerStream) streamCompleteValues(
	ctx context.Context,
	id int64,
	completer iterator.Iterator[string],
) error {
	for {
		next, ok := completer.Next(ctx)
		if !ok {
			return completer.Err()
		}
		resp := CompleteCommandValue{Id: id, Value: next}
		respMsg := &ClientREPLCommandMessage{
			CompleteValue: &resp,
			Type:          ClientREPLCommandMessage_CompleteValue,
		}
		select {
		case s.sendChan <- respMsg:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *replCommandServerStream) streamResponsive(
	ctx context.Context,
	iter iterator.Iterator[component.Responsive],
	width int,
	wrap func([]*termrpc.CellRow) *ClientREPLCommandMessage,
) error {
	for {
		item, ok := iter.Next(ctx)
		if !ok {
			return iter.Err()
		}
		rows := renderResponsive(item, width)
		msg := wrap(rows)
		select {
		case s.sendChan <- msg:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// renderResponsive renders a responsive component at the
// given width and returns the resulting proto cell rows.
func renderResponsive(
	c component.Responsive, width int,
) []*termrpc.CellRow {
	if width <= 0 {
		return nil
	}
	height := c.Height(width)
	if height <= 0 {
		return nil
	}
	w := term.NewStringWriter(width, height)
	c.Resize(width, height)
	c.Draw(w)
	cells := w.Cells()
	rows := make([]*termrpc.CellRow, height)
	for y := range height {
		row := &termrpc.CellRow{
			Cells: make([]*termrpc.Cell, width),
		}
		for x := range width {
			var pc termrpc.Cell
			pc.FromModel(cells[y*width+x])
			row.Cells[x] = &pc
		}
		rows[y] = row
	}
	return rows
}
