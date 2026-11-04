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

	acp "github.com/coder/acp-go-sdk"
	"github.com/unstablebuild/rune-go-sdk/api/acpapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	UnimplementedACPServer
	svc acpapi.Service
}

// NewServer returns an ACPServer that serves svc. acpapi's sentinel errors
// reach clients as the same errors.
func NewServer(svc acpapi.Service) ACPServer {
	return &server{svc: svc}
}

func (s *server) Agents(_ *AgentsRequest, stream grpc.ServerStreamingServer[AgentInfo]) error {
	ctx := stream.Context()
	it := s.svc.Agents(ctx)
	defer func() { _ = it.Close() }()
	for {
		a, ok := it.Next(ctx)
		if !ok {
			return toStatus(it.Err())
		}
		if err := stream.Send(&AgentInfo{Name: a.Name, Running: a.Running}); err != nil {
			return err
		}
	}
}

func (s *server) Sessions(_ *SessionsRequest, stream grpc.ServerStreamingServer[SessionInfo]) error {
	ctx := stream.Context()
	it := s.svc.Sessions(ctx)
	defer func() { _ = it.Close() }()
	for {
		info, ok := it.Next(ctx)
		if !ok {
			return toStatus(it.Err())
		}
		if err := stream.Send(toProtoInfo(info)); err != nil {
			return err
		}
	}
}

func (s *server) NewSession(ctx context.Context, req *NewSessionRequest) (*SessionInfo, error) {
	var sreq acp.NewSessionRequest
	if len(req.GetSession()) > 0 {
		if err := json.Unmarshal(req.GetSession(), &sreq); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "session: %v", err)
		}
	}
	sess, err := s.svc.NewSession(ctx, acpapi.NewSessionRequest{
		Agent:   req.GetAgent(),
		Name:    req.GetName(),
		Session: sreq,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return toProtoInfo(sess.Info()), nil
}

func (s *server) OpenSession(ctx context.Context, req *OpenSessionRequest) (*SessionInfo, error) {
	sess, err := s.svc.OpenSession(ctx, req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}
	return toProtoInfo(sess.Info()), nil
}

func (s *server) Events(req *EventsRequest, stream grpc.ServerStreamingServer[Event]) error {
	ctx := stream.Context()
	sess, err := s.svc.OpenSession(ctx, req.GetSession())
	if err != nil {
		return toStatus(err)
	}
	it := sess.Events()
	defer func() { _ = it.Close() }()
	for {
		ev, ok := it.Next(ctx)
		if !ok {
			break
		}
		b, err := json.Marshal(ev)
		if err != nil {
			return status.Errorf(codes.Internal, "event %d: %v", ev.Seq, err)
		}
		if err := stream.Send(&Event{Event: b}); err != nil {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return toStatus(err)
	}
	if err := it.Err(); err != nil {
		return status.Error(codes.Unavailable, err.Error())
	}
	return nil
}

func (s *server) Prompt(ctx context.Context, req *PromptRequest) (*PromptResponse, error) {
	var content []acp.ContentBlock
	if err := json.Unmarshal(req.GetContent(), &content); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "content: %v", err)
	}
	sess, err := s.svc.OpenSession(ctx, req.GetSession())
	if err != nil {
		return nil, toStatus(err)
	}
	if err := sess.Prompt(ctx, content); err != nil {
		return nil, toStatus(err)
	}
	return &PromptResponse{}, nil
}

func (s *server) Cancel(ctx context.Context, req *CancelRequest) (*CancelResponse, error) {
	sess, err := s.svc.OpenSession(ctx, req.GetSession())
	if err != nil {
		return nil, toStatus(err)
	}
	if err := sess.Cancel(ctx); err != nil {
		return nil, toStatus(err)
	}
	return &CancelResponse{}, nil
}

func (s *server) Respond(ctx context.Context, req *RespondRequest) (*RespondResponse, error) {
	var o acp.RequestPermissionOutcome
	if err := json.Unmarshal(req.GetOutcome(), &o); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "outcome: %v", err)
	}
	sess, err := s.svc.OpenSession(ctx, req.GetSession())
	if err != nil {
		return nil, toStatus(err)
	}
	if err := sess.Respond(ctx, acpapi.PermissionID(req.GetPermissionId()), o); err != nil {
		return nil, toStatus(err)
	}
	return &RespondResponse{}, nil
}

func (s *server) SetConfigOption(ctx context.Context, req *SetConfigOptionRequest) (*SetConfigOptionResponse, error) {
	var r acp.SetSessionConfigOptionRequest
	if err := json.Unmarshal(req.GetRequest(), &r); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "request: %v", err)
	}
	sess, err := s.svc.OpenSession(ctx, req.GetSession())
	if err != nil {
		return nil, toStatus(err)
	}
	if err := sess.SetConfigOption(ctx, r); err != nil {
		return nil, toStatus(err)
	}
	return &SetConfigOptionResponse{}, nil
}

func (s *server) CloseSession(ctx context.Context, req *CloseSessionRequest) (*CloseSessionResponse, error) {
	sess, err := s.svc.OpenSession(ctx, req.GetSession())
	if err != nil {
		return nil, toStatus(err)
	}
	if err := sess.Close(); err != nil {
		return nil, toStatus(err)
	}
	return &CloseSessionResponse{}, nil
}

func toProtoInfo(info acpapi.SessionInfo) *SessionInfo {
	return &SessionInfo{Name: info.Name, Agent: info.Agent, Id: string(info.ID)}
}

// sentinels pairs each acpapi error with the code it travels as. Errors that
// share a code are told apart by their message.
var sentinels = []struct {
	err  error
	code codes.Code
}{
	{acpapi.ErrNotFound, codes.NotFound},
	{acpapi.ErrExists, codes.AlreadyExists},
	{acpapi.ErrBusy, codes.FailedPrecondition},
	{acpapi.ErrResolved, codes.FailedPrecondition},
	{acpapi.ErrClosed, codes.FailedPrecondition},
}

func toStatus(err error) error {
	if err == nil {
		return nil
	}
	for _, s := range sentinels {
		if errors.Is(err, s.err) {
			return status.Error(s.code, err.Error())
		}
	}
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	return status.Error(codes.Unknown, err.Error())
}
