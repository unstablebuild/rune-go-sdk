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

package syntaxrpc

import (
	"context"

	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/term/termrpc"
	"google.golang.org/grpc"
)

// Server implements SyntaxServer by delegating to a syntaxapi.Parser.
type Server struct {
	UnimplementedSyntaxServer
	parser syntaxapi.Parser
}

// NewServer returns a Server that delegates to parser.
func NewServer(parser syntaxapi.Parser) *Server {
	return &Server{parser: parser}
}

// Highlight implements SyntaxServer.
func (s *Server) Highlight(
	req *HighlightRequest,
	stream grpc.ServerStreamingServer[HighlightResponse],
) error {
	uri, err := workspaceapi.ParseURI(req.GetUri())
	if err != nil {
		return err
	}
	it, err := s.parser.Highlight(uri, req.GetContent())
	if err != nil {
		return err
	}
	defer func() { _ = it.Close() }()
	for {
		loc, ok := it.Next(context.Background())
		if !ok {
			return it.Err()
		}
		var from, to termrpc.Coordinates
		from.FromModel(loc.From)
		to.FromModel(loc.To)
		var attr termrpc.Attributes
		attr.FromModel(loc.Attr)
		resp := HighlightResponse{
			From:    &from,
			To:      &to,
			Attr:    &attr,
			Message: loc.Message,
		}
		if err := stream.Send(&resp); err != nil {
			return err
		}
	}
}
