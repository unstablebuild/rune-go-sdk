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

package searchrpc

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/unstablebuild/rune-go-sdk/api/llmapi"
	"github.com/unstablebuild/rune-go-sdk/api/searchapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ searchapi.Searcher = (*Client)(nil)

// Client satisfies searchapi.Searcher via gRPC.
type Client struct {
	pb SearcherClient
}

// NewClient returns a Client that calls the Searcher service on cc.
func NewClient(cc grpc.ClientConnInterface) *Client {
	return &Client{pb: NewSearcherClient(cc)}
}

// RegexpSearch satisfies searchapi.Searcher.
func (c *Client) RegexpSearch(
	ctx context.Context, req searchapi.RegexpRequest,
) (iterator.Iterator[searchapi.RegexpResult], error) {
	callCtx, cancel := context.WithCancel(ctx)
	stream, err := c.pb.RegexpSearch(callCtx, ToProtoRegexpRequest(req))
	if err != nil {
		cancel()
		return nil, fromStatus(err)
	}
	return &streamIterator[RegexpResponse, searchapi.RegexpResult]{
		stream: stream,
		cancel: cancel,
		unpack: func(resp *RegexpResponse) []searchapi.RegexpResult {
			out := make([]searchapi.RegexpResult, len(resp.GetResults()))
			for i, r := range resp.GetResults() {
				out[i] = FromProtoRegexpResult(r)
			}
			return out
		},
	}, nil
}

// SemanticSearch satisfies searchapi.Searcher.
func (c *Client) SemanticSearch(
	ctx context.Context, req searchapi.SemanticRequest,
) (iterator.Iterator[searchapi.SemanticResult], error) {
	callCtx, cancel := context.WithCancel(ctx)
	stream, err := c.pb.SemanticSearch(callCtx, ToProtoSemanticRequest(req))
	if err != nil {
		cancel()
		return nil, fromStatus(err)
	}
	return &streamIterator[SemanticResponse, searchapi.SemanticResult]{
		stream: stream,
		cancel: cancel,
		unpack: func(resp *SemanticResponse) []searchapi.SemanticResult {
			return []searchapi.SemanticResult{FromProtoSemanticResult(resp.GetResult())}
		},
	}, nil
}

type streamIterator[R any, T any] struct {
	stream grpc.ServerStreamingClient[R]
	cancel context.CancelFunc
	unpack func(*R) []T
	batch  []T
	err    error
	done   bool
}

func (it *streamIterator[R, T]) Next(context.Context) (T, bool) {
	var zero T
	for len(it.batch) == 0 {
		if it.done {
			return zero, false
		}
		resp, err := it.stream.Recv()
		if err != nil {
			it.done = true
			if !errors.Is(err, io.EOF) {
				it.err = fromStatus(err)
			}
			return zero, false
		}
		it.batch = it.unpack(resp)
	}
	ret := it.batch[0]
	it.batch = it.batch[1:]
	return ret, true
}

func (it *streamIterator[R, T]) Err() error { return it.err }

// Close cancels the stream, which stops the search on the server.
func (it *streamIterator[R, T]) Close() error {
	it.cancel()
	return nil
}

func fromStatus(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch st.Code() {
	case codes.Unimplemented:
		return fmt.Errorf("%s: %w", st.Message(), searchapi.ErrUnsupported)
	case codes.NotFound:
		return fmt.Errorf("%s: %w", st.Message(), llmapi.ErrModelNotFound)
	case codes.Canceled:
		return fmt.Errorf("%s: %w", st.Message(), context.Canceled)
	case codes.DeadlineExceeded:
		return fmt.Errorf("%s: %w", st.Message(), context.DeadlineExceeded)
	default:
		return err
	}
}
