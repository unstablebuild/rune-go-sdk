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

package llmrpc

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/llmapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// stubLLMServer implements the GetModel RPC for testing the client.
type stubLLMServer struct {
	UnimplementedLLMServer
	resp *GetModelResponse
	err  error
}

func (s *stubLLMServer) GetModel(
	_ context.Context, _ *GetModelRequest,
) (*GetModelResponse, error) {
	return s.resp, s.err
}

// dialStub starts an in-process gRPC server backed by srv and returns a
// Client wired to it.
func dialStub(t *testing.T, srv *stubLLMServer) *Client {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	gs := grpc.NewServer()
	RegisterLLMServer(gs, srv)
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

func TestClientGetModel(t *testing.T) {
	tests := []struct {
		name      string
		resp      *GetModelResponse
		serverErr error
		wantEntry llmapi.ModelEntry
		wantErrIs error
		wantErr   bool
	}{
		{
			name: "found",
			resp: &GetModelResponse{
				Found: true,
				Model: ToProtoModelEntry(llmapi.ModelEntry{
					Name:          "gpt-5",
					Provider:      "openai",
					ContextWindow: 128000,
				}),
			},
			wantEntry: llmapi.ModelEntry{
				Name:          "gpt-5",
				Provider:      "openai",
				ContextWindow: 128000,
			},
		},
		{
			name:      "not found",
			resp:      &GetModelResponse{Found: false},
			wantErrIs: llmapi.ErrModelNotFound,
			wantErr:   true,
		},
		{
			name:      "transport error",
			serverErr: status.Error(codes.Internal, "boom"),
			wantErr:   true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := dialStub(t, &stubLLMServer{resp: tc.resp, err: tc.serverErr})
			entry, err := c.GetModel(context.Background(),
				llmapi.ModelEntry{Name: "gpt-5", Provider: "openai"})
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrIs != nil {
					assert.ErrorIs(t, err, tc.wantErrIs)
				} else {
					assert.False(t, errors.Is(err, llmapi.ErrModelNotFound))
				}
				assert.Equal(t, llmapi.ModelEntry{}, entry)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantEntry, entry)
		})
	}
}
