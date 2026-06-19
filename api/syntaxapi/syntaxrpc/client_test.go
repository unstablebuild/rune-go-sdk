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
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type stubSyntaxServer struct {
	UnimplementedSyntaxServer
	matches   []*ResolveSymbolMatch
	progress  []*ResolveSymbolProgress
	noDot     bool
	serverErr error
	gotName   string
}

func (s *stubSyntaxServer) ResolveSymbol(
	req *ResolveSymbolRequest, stream grpc.ServerStreamingServer[ResolveSymbolResponse],
) error {
	s.gotName = req.GetName()
	if s.noDot {
		return status.Error(codes.InvalidArgument, errNoDotDetail)
	}
	if s.serverErr != nil {
		return s.serverErr
	}
	for _, p := range s.progress {
		if err := stream.Send(&ResolveSymbolResponse{
			Payload: &ResolveSymbolResponse_Progress{Progress: p},
		}); err != nil {
			return err
		}
	}
	for _, m := range s.matches {
		if err := stream.Send(&ResolveSymbolResponse{
			Payload: &ResolveSymbolResponse_Match{Match: m},
		}); err != nil {
			return err
		}
	}
	return nil
}

func dialSyntaxStub(t *testing.T, srv *stubSyntaxServer) *Client {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	gs := grpc.NewServer()
	RegisterSyntaxServer(gs, srv)
	go func() { _ = gs.Serve(lis) }()
	t.Cleanup(gs.Stop)

	cc, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cc.Close() })
	return NewClient(context.Background(), cc)
}

func TestClientResolveSymbol(t *testing.T) {
	srv := &stubSyntaxServer{
		progress: []*ResolveSymbolProgress{
			{Message: "Searching references…", Found: 0, Step: 0, Total: 4},
			{Message: "Searching definitions…", Found: 1, Step: 2, Total: 4},
		},
		matches: []*ResolveSymbolMatch{
			{Uri: "file:///a.go", Line: 10, Character: 5, Display: "pkg.Sym", ImportPath: "example/pkg"},
			{Uri: "file:///b.go", Line: 20, Character: 1, Display: "pkg.Sym"},
		},
	}
	c := dialSyntaxStub(t, srv)

	var reports []string
	prog := syntaxapi.ProgressFunc(func(msg string, _ int, _, _ int64) {
		reports = append(reports, msg)
	})

	it, err := c.ResolveSymbol(context.Background(), "pkg.Sym", prog)
	require.NoError(t, err)
	matches, err := iterator.ToSlice(context.Background(), it)
	require.NoError(t, err)
	require.NoError(t, it.Err())

	assert.Equal(t, "pkg.Sym", srv.gotName)
	assert.Equal(t, []string{"Searching references…", "Searching definitions…"}, reports)
	require.Len(t, matches, 2)
	assert.Equal(t, "file:///a.go", matches[0].URI)
	assert.Equal(t, 10, matches[0].Pos.Y)
	assert.Equal(t, 5, matches[0].Pos.X)
	assert.Equal(t, "pkg.Sym", matches[0].Display)
	assert.Equal(t, "example/pkg", matches[0].ImportPath)
	assert.Equal(t, "file:///b.go", matches[1].URI)
}

func TestClientResolveSymbolNilProgress(t *testing.T) {
	srv := &stubSyntaxServer{
		progress: []*ResolveSymbolProgress{{Message: "step", Total: 1}},
		matches:  []*ResolveSymbolMatch{{Uri: "file:///a.go", Line: 1}},
	}
	c := dialSyntaxStub(t, srv)

	it, err := c.ResolveSymbol(context.Background(), "pkg.Sym", nil)
	require.NoError(t, err)
	matches, err := iterator.ToSlice(context.Background(), it)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	assert.Equal(t, "file:///a.go", matches[0].URI)
}

func TestClientResolveSymbolErrNoDot(t *testing.T) {
	c := dialSyntaxStub(t, &stubSyntaxServer{noDot: true})

	it, err := c.ResolveSymbol(context.Background(), "Sym", nil)
	require.NoError(t, err)
	matches, err := iterator.ToSlice(context.Background(), it)
	assert.ErrorIs(t, err, syntaxapi.ErrNoDot)
	assert.ErrorIs(t, it.Err(), syntaxapi.ErrNoDot)
	assert.Empty(t, matches)
}

func TestClientResolveSymbolTransportError(t *testing.T) {
	c := dialSyntaxStub(t, &stubSyntaxServer{serverErr: status.Error(codes.Internal, "boom")})

	it, err := c.ResolveSymbol(context.Background(), "pkg.Sym", nil)
	require.NoError(t, err)
	_, err = iterator.ToSlice(context.Background(), it)
	require.Error(t, err)
	assert.NotErrorIs(t, err, syntaxapi.ErrNoDot)
}
