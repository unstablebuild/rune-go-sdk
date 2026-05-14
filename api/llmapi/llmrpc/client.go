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
	"fmt"
	"io"
	"net"

	"github.com/unstablebuild/rune-go-sdk/api/llmapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"google.golang.org/grpc"
)

var _ llmapi.Service = (*Client)(nil)

// Client satisfies llmapi.Service via gRPC.
type Client struct {
	cc  grpc.ClientConnInterface
	pb  LLMClient
	ctx context.Context
}

// NewClient allocates storage for a new Client and initializes it with cc.
func NewClient(ctx context.Context, cc grpc.ClientConnInterface) *Client {
	ret := new(Client)
	ret.Init(ctx, cc)
	return ret
}

// Init initializes this client with cc.
func (c *Client) Init(ctx context.Context, cc grpc.ClientConnInterface) {
	c.cc = cc
	c.pb = NewLLMClient(cc)
	c.ctx = ctx
}

// NewClientFromAddr returns a gRPC-backed llmapi.Service that dials addr.
// Matches the shape of storagerpc.NewClient.
func NewClientFromAddr(addr net.Addr, opts ...grpc.DialOption) (llmapi.Service, error) {
	opts = append(opts, grpc.WithContextDialer(
		func(ctx context.Context, _ string) (net.Conn, error) {
			var d net.Dialer
			d.Deadline, _ = ctx.Deadline()
			return d.DialContext(ctx, addr.Network(), addr.String())
		},
	))
	cc, err := grpc.NewClient("passthrough:///", opts...)
	if err != nil {
		return nil, err
	}
	return NewClient(context.Background(), cc), nil
}

// CreateCompletion satisfies llmapi.Service.
func (c *Client) CreateCompletion(
	ctx context.Context, model llmapi.ModelEntry, req llmapi.Request,
) (iterator.Iterator[llmapi.Event], error) {
	preq, err := ToProtoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("llm: encode request: %w", err)
	}
	callCtx, cancel := context.WithCancel(ctx)
	stream, err := c.pb.CreateCompletion(callCtx, &CreateCompletionRequest{
		Model:   ToProtoModelEntry(model),
		Request: preq,
	})
	if err != nil {
		cancel()
		if ctxErr, ok := ContextWindowExceededFromStatus(err); ok {
			return nil, ctxErr
		}
		return nil, fmt.Errorf("llm: create completion: %w", err)
	}
	return &eventIterator{stream: stream, cancel: cancel}, nil
}

// CountTokens satisfies llmapi.Service.
func (c *Client) CountTokens(
	model llmapi.ModelEntry, msgs []llmapi.Message,
) (int, error) {
	req := &CountTokensRequest{Model: ToProtoModelEntry(model)}
	if len(msgs) > 0 {
		req.Messages = make([]*Message, len(msgs))
		for i, m := range msgs {
			req.Messages[i] = ToProtoMessage(m)
		}
	}
	resp, err := c.pb.CountTokens(c.ctx, req)
	if err != nil {
		return 0, fmt.Errorf("llm: count tokens: %w", err)
	}
	return int(resp.GetCount()), nil
}

// Models satisfies llmapi.Service.
func (c *Client) Models() iterator.Iterator[llmapi.ModelEntry] {
	callCtx, cancel := context.WithCancel(c.ctx)
	stream, err := c.pb.Models(callCtx, &ModelsRequest{})
	if err != nil {
		cancel()
		return &modelsIterator{err: fmt.Errorf("llm: models: %w", err)}
	}
	return &modelsIterator{stream: stream, cancel: cancel}
}

// GetModel satisfies llmapi.Service.
func (c *Client) GetModel(
	ctx context.Context, model llmapi.ModelEntry,
) (llmapi.ModelEntry, bool) {
	resp, err := c.pb.GetModel(ctx, &GetModelRequest{Model: ToProtoModelEntry(model)})
	if err != nil {
		return llmapi.ModelEntry{}, false
	}
	if !resp.GetFound() {
		return llmapi.ModelEntry{}, false
	}
	return FromProtoModelEntry(resp.GetModel()), true
}

type modelsIterator struct {
	stream grpc.ServerStreamingClient[ModelsResponse]
	cancel context.CancelFunc
	err    error
}

func (it *modelsIterator) Next(_ context.Context) (llmapi.ModelEntry, bool) {
	if it.err != nil || it.stream == nil {
		return llmapi.ModelEntry{}, false
	}
	resp, err := it.stream.Recv()
	if err == io.EOF {
		return llmapi.ModelEntry{}, false
	}
	if err != nil {
		it.err = err
		return llmapi.ModelEntry{}, false
	}
	return FromProtoModelEntry(resp.GetModel()), true
}

func (it *modelsIterator) Err() error { return it.err }

func (it *modelsIterator) Close() error {
	if it.cancel != nil {
		it.cancel()
	}
	return nil
}

type eventIterator struct {
	stream grpc.ServerStreamingClient[CreateCompletionResponse]
	cancel context.CancelFunc
	err    error
}

func (it *eventIterator) Next(_ context.Context) (llmapi.Event, bool) {
	resp, err := it.stream.Recv()
	if err == io.EOF {
		return llmapi.Event{}, false
	}
	if err != nil {
		if ctxErr, ok := ContextWindowExceededFromStatus(err); ok {
			it.err = ctxErr
		} else {
			it.err = err
		}
		return llmapi.Event{}, false
	}
	return FromProtoEvent(resp.GetEvent()), true
}

func (it *eventIterator) Err() error { return it.err }

func (it *eventIterator) Close() error {
	if it.cancel != nil {
		it.cancel()
	}
	return nil
}
