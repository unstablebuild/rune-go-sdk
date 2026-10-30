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
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/llmapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
)

// stubLLMServer implements the GetModel RPC for testing the client.
type stubLLMServer struct {
	UnimplementedLLMServer
	resp         *GetModelResponse
	err          error
	availability func(context.Context, *CanGenerateImagesRequest) (*CanGenerateImagesResponse, error)
	generate     func(context.Context, *GenerateImageRequest) (*GenerateImageResponse, error)
}

func (s *stubLLMServer) CanGenerateImages(ctx context.Context, req *CanGenerateImagesRequest) (*CanGenerateImagesResponse, error) {
	if s.availability == nil {
		return s.UnimplementedLLMServer.CanGenerateImages(ctx, req)
	}
	return s.availability(ctx, req)
}

func (s *stubLLMServer) GenerateImage(ctx context.Context, req *GenerateImageRequest) (*GenerateImageResponse, error) {
	if s.generate == nil {
		return s.UnimplementedLLMServer.GenerateImage(ctx, req)
	}
	return s.generate(ctx, req)
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
	gs := grpc.NewServer(grpc.MaxRecvMsgSize(imageRPCMessageLimit), grpc.MaxSendMsgSize(imageRPCMessageLimit))
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

func TestClientCanGenerateImages(t *testing.T) {
	model := llmapi.ModelEntry{Name: "image-model", Provider: "provider", BaseURL: "https://example.test"}
	for _, tc := range []struct {
		name      string
		available bool
		code      codes.Code
	}{
		{name: "available", available: true},
		{name: "unavailable"},
		{name: "older host", code: codes.Unimplemented},
		{name: "permission denied", code: codes.PermissionDenied},
		{name: "canceled", code: codes.Canceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := &stubLLMServer{}
			if tc.code != codes.Unimplemented {
				srv.availability = func(_ context.Context, req *CanGenerateImagesRequest) (*CanGenerateImagesResponse, error) {
					assert.Equal(t, model, FromProtoModelEntry(req.GetModel()))
					return &CanGenerateImagesResponse{Available: tc.available}, status.Error(tc.code, "test error")
				}
			}
			var service llmapi.Service = dialStub(t, srv)
			got, err := service.CanGenerateImages(context.Background(), model)
			wantCode := tc.code
			if wantCode == codes.Unimplemented {
				wantCode = codes.OK
			}
			assert.Equal(t, wantCode, status.Code(err))
			assert.Equal(t, tc.available, got)
		})
	}
}

func TestClientGenerateImage(t *testing.T) {
	model := llmapi.ModelEntry{Name: "image-model", Provider: "provider"}
	request := llmapi.ImageRequest{Prompt: "draw a tree", TransparentBackground: true, Images: [][]byte{{1, 2}, {3, 4}}}
	for _, tc := range []struct {
		name     string
		size     int
		code     codes.Code
		wantCode codes.Code
	}{
		{name: "round trip", size: 16},
		{name: "above default receive limit", size: 5 << 20},
		{name: "above response budget", size: MaxImageMessageSize, wantCode: codes.ResourceExhausted},
		{name: "above transport limit", size: imageRPCMessageLimit, wantCode: codes.ResourceExhausted},
		{name: "unsupported", code: codes.Unimplemented, wantCode: codes.Unimplemented},
		{name: "provider error", code: codes.Internal, wantCode: codes.Internal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			image := llmapi.Image{PNG: make([]byte, tc.size), GenerationID: "generation", RequestID: "request"}
			srv := &stubLLMServer{}
			if tc.code != codes.Unimplemented {
				srv.generate = func(_ context.Context, req *GenerateImageRequest) (*GenerateImageResponse, error) {
					assert.Equal(t, model, FromProtoModelEntry(req.GetModel()))
					assert.Equal(t, request, FromProtoImageRequest(req.GetRequest()))
					return &GenerateImageResponse{Image: ToProtoImage(image)}, status.Error(tc.code, "test error")
				}
			}
			var service llmapi.Service = dialStub(t, srv)
			got, err := service.GenerateImage(context.Background(), model, request)
			require.Equal(t, tc.wantCode, status.Code(err), "%v", err)
			if err == nil {
				assert.Equal(t, image, got)
			} else {
				assert.Equal(t, llmapi.Image{}, got)
			}
		})
	}
}

func TestClientImageRequestSize(t *testing.T) {
	for _, tc := range []struct {
		name    string
		model   llmapi.ModelEntry
		request llmapi.ImageRequest
	}{
		{name: "prompt", request: llmapi.ImageRequest{Prompt: strings.Repeat("x", MaxImageMessageSize)}},
		{name: "combined references", request: llmapi.ImageRequest{Images: [][]byte{make([]byte, 8<<20), make([]byte, 8<<20)}}},
		{name: "model metadata", model: llmapi.ModelEntry{Name: strings.Repeat("x", MaxImageMessageSize)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := new(Client).GenerateImage(context.Background(), tc.model, tc.request)
			require.Equal(t, codes.ResourceExhausted, status.Code(err))
			assert.Contains(t, err.Error(), "16 MiB RPC limit")
		})
	}
	for _, delta := range []int{0, 1} {
		t.Run(fmt.Sprintf("budget plus %d", delta), func(t *testing.T) {
			req := llmapi.ImageRequest{Images: [][]byte{make([]byte, MaxImageMessageSize)}}
			overhead := proto.Size(&GenerateImageRequest{Model: ToProtoModelEntry(llmapi.ModelEntry{}), Request: ToProtoImageRequest(req)}) - MaxImageMessageSize
			req.Images[0] = req.Images[0][:MaxImageMessageSize-overhead+delta]
			c := dialStub(t, &stubLLMServer{})
			_, err := c.GenerateImage(context.Background(), llmapi.ModelEntry{}, req)
			want := codes.Unimplemented
			if delta > 0 {
				want = codes.ResourceExhausted
			}
			assert.Equal(t, want, status.Code(err))
		})
	}
	_, err := new(Client).CanGenerateImages(context.Background(), llmapi.ModelEntry{Name: strings.Repeat("x", MaxImageMessageSize)})
	assert.Equal(t, codes.ResourceExhausted, status.Code(err))
}
