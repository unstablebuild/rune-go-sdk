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

package syntaxrpc_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi"
	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi/syntaxrpc"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"github.com/unstablebuild/rune-go-sdk/term"
	tcell "github.com/unstablebuild/tcell/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type stubParser struct {
	lastURI     workspaceapi.URI
	lastContent string
	locations   []textapi.Location
}

func (s *stubParser) Search(string, []string) (iterator.Iterator[syntaxapi.Result], error) {
	return iterator.Empty[syntaxapi.Result](), nil
}

func (s *stubParser) SearchNode(syntaxapi.NodeCaptureName) (iterator.Iterator[syntaxapi.Result], error) {
	return iterator.Empty[syntaxapi.Result](), nil
}

func (s *stubParser) Query(workspaceapi.URI, string, []string) (iterator.Iterator[syntaxapi.Result], error) {
	return iterator.Empty[syntaxapi.Result](), nil
}

func (s *stubParser) QueryNode(workspaceapi.URI, syntaxapi.NodeCaptureName) (iterator.Iterator[syntaxapi.Result], error) {
	return iterator.Empty[syntaxapi.Result](), nil
}

func (s *stubParser) Highlight(
	uri workspaceapi.URI, content string,
) (iterator.Iterator[textapi.Location], error) {
	s.lastURI = uri
	s.lastContent = content
	return iterator.FromSlice(s.locations), nil
}

func TestHighlight(t *testing.T) {
	tests := []struct {
		name      string
		uri       string
		content   string
		locations []textapi.Location
	}{
		{
			name:    "single location",
			uri:     "file:///tmp/test.go",
			content: "package main",
			locations: []textapi.Location{
				{
					From:    term.Coordinates{X: 0, Y: 0},
					To:      term.Coordinates{X: 7, Y: 0},
					Attr:    term.Attributes(tcell.Style{Fg: tcell.ColorBlue}),
					Message: "keyword",
				},
			},
		},
		{
			name:    "multiple locations",
			uri:     "file:///workspace/main.rs",
			content: "fn main() {}",
			locations: []textapi.Location{
				{
					From:    term.Coordinates{X: 0, Y: 0},
					To:      term.Coordinates{X: 2, Y: 0},
					Attr:    term.Attributes(tcell.Style{Fg: tcell.ColorRed}),
					Message: "keyword",
				},
				{
					From:    term.Coordinates{X: 3, Y: 0},
					To:      term.Coordinates{X: 7, Y: 0},
					Attr:    term.Attributes(tcell.Style{Fg: tcell.ColorGreen, Attrs: tcell.AttrBold}),
					Message: "function",
				},
			},
		},
		{
			name:      "empty locations",
			uri:       "file:///tmp/empty.txt",
			content:   "",
			locations: []textapi.Location{},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubParser{locations: tt.locations}
			srv := grpc.NewServer()
			syntaxrpc.RegisterSyntaxServer(srv, syntaxrpc.NewServer(stub))

			// Use a short path to avoid unix socket path length limits.
			tmpDir, err := os.MkdirTemp("", "syn")
			require.NoError(t, err)
			t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })
			sockPath := filepath.Join(tmpDir, fmt.Sprintf("%d.sock", i))
			lis, err := net.Listen("unix", sockPath)
			require.NoError(t, err)
			go func() { _ = srv.Serve(lis) }()
			t.Cleanup(srv.Stop)

			conn, err := grpc.NewClient(
				"unix:"+sockPath,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			require.NoError(t, err)
			t.Cleanup(func() { _ = conn.Close() })

			ctx := context.Background()
			client := syntaxrpc.NewClient(ctx, conn)

			uri, err := workspaceapi.ParseURI(tt.uri)
			require.NoError(t, err)

			it, err := client.Highlight(uri, tt.content)
			require.NoError(t, err)

			got, err := iterator.ToSlice(ctx, it)
			require.NoError(t, err)

			require.Equal(t, tt.uri, stub.lastURI.String())
			require.Equal(t, tt.content, stub.lastContent)
			require.Equal(t, tt.locations, got)
		})
	}
}
