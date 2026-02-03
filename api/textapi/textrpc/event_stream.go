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
	"time"

	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const (
	handleReceiveMessageTimeout = 3 * time.Second
)

type eventStreamServer struct {
	log       *slog.Logger
	stream    Editor_SubscribeEventClient
	parentCtx context.Context
	handler   textapi.EventHandler
}

func newEventStreamServer(
	parentCtx context.Context, stream Editor_SubscribeEventClient,
	handler textapi.EventHandler,
) eventStreamServer {
	return eventStreamServer{
		log:       slog.Default().With("struct", "textrpc.eventStreamServer"),
		stream:    stream,
		parentCtx: parentCtx,
		handler:   handler,
	}
}

func (s eventStreamServer) receiveEvents() {
	defer s.log.Debug("done receiving events")
	for {
		protoEv, err := s.stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) && status.Code(err) != codes.Canceled {
				s.log.Error("stream recv error", "error", err)
			}
			break
		}

		var ev textapi.Event
		if err := fromProto(&ev, protoEv); err != nil {
			s.log.Error("decode proto event", "error", err)
			continue
		}

		// set a timeout to prevent a deadlock via client stream buffer exhaustion
		ctx, cancel := context.WithTimeout(s.parentCtx, handleReceiveMessageTimeout)
		s.log.Debug("handle event", "type", ev.Type)
		exit := s.handler.Handle(ctx, ev)
		if err := ctx.Err(); err != nil {
			s.log.Error("could not dispatch event in time", "type", ev.Type, "error", err)
			select {
			case <-s.parentCtx.Done():
				cancel()
				return
			default:
			}
		}
		cancel()
		if exit {
			break
		}
	}

	// do not attempt to send unsubscribe if client is closing
	select {
	case <-s.parentCtx.Done():
		return
	default:
	}

	req := SubscribeEventRequest{Unsubscribe: true}
	s.log.Debug("send unsubscribe")
	err := s.stream.Send(&req)
	if err != nil {
		s.log.Error("send unsubscribe", "error", err)
	}
	if err := s.stream.CloseSend(); err != nil {
		s.log.Error("stream close send", "error", err)
	}
}
