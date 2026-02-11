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
	"fmt"
	"io"

	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"google.golang.org/grpc"
)

var _ syntaxapi.Searcher = (*Client)(nil)

// Client satisfies syntaxapi.Searcher via gRPC.
type Client struct {
	cc  grpc.ClientConnInterface
	pb  SyntaxClient
	ctx context.Context
}

// NewClient allocates storage for a new Client and
// initializes it with the given connection.
func NewClient(ctx context.Context, cc grpc.ClientConnInterface) *Client {
	ret := new(Client)
	ret.Init(ctx, cc)
	return ret
}

// Init initializes this client with cc.
func (c *Client) Init(ctx context.Context, cc grpc.ClientConnInterface) {
	c.cc = cc
	c.pb = NewSyntaxClient(cc)
	c.ctx = ctx
}

// Search satisfies syntaxapi.Searcher.
func (c *Client) Search(
	query string,
	captureNames []string,
) (iterator.Iterator[syntaxapi.Result], error) {
	req := SearchRequest{Query: query, CaptureNames: captureNames}
	stream, err := c.pb.Search(c.ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("syntax search: %w", err)
	}
	return &rpcIterator{stream: stream}, nil
}

// SearchNode satisfies syntaxapi.Searcher.
func (c *Client) SearchNode(
	nodeTypes syntaxapi.NodeCaptureName,
) (iterator.Iterator[syntaxapi.Result], error) {
	req := SearchNodeRequest{NodeTypes: uint32(nodeTypes)}
	stream, err := c.pb.SearchNode(c.ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("syntax search node: %w", err)
	}
	return &rpcIterator{stream: stream}, nil
}

// Query satisfies syntaxapi.Searcher.
func (c *Client) Query(
	file workspaceapi.URI,
	query string,
	captureNames []string,
) (iterator.Iterator[syntaxapi.Result], error) {
	req := QueryRequest{
		Uri:          file.String(),
		Query:        query,
		CaptureNames: captureNames,
	}
	stream, err := c.pb.Query(c.ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("syntax query: %w", err)
	}
	return &rpcIterator{stream: stream}, nil
}

// QueryNode satisfies syntaxapi.Searcher.
func (c *Client) QueryNode(
	file workspaceapi.URI,
	nodeTypes syntaxapi.NodeCaptureName,
) (iterator.Iterator[syntaxapi.Result], error) {
	req := QueryNodeRequest{Uri: file.String(), NodeTypes: uint32(nodeTypes)}
	stream, err := c.pb.QueryNode(c.ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("syntax query node: %w", err)
	}
	return &rpcIterator{stream: stream}, nil
}

type rpcIterator struct {
	stream grpc.ServerStreamingClient[SearchResponse]
	err    error
}

func (it *rpcIterator) Next(_ context.Context) (syntaxapi.Result, bool) {
	resp, err := it.stream.Recv()
	if err == io.EOF {
		return syntaxapi.Result{}, false
	}
	if err != nil {
		it.err = err
		return syntaxapi.Result{}, false
	}

	uri, err := workspaceapi.ParseURI(resp.GetUri())
	if err != nil {
		it.err = fmt.Errorf("parse uri: %w", err)
		return syntaxapi.Result{}, false
	}

	result := syntaxapi.Result{
		File:        uri,
		Text:        resp.GetText(),
		From:        resp.GetFrom().ToModel(),
		To:          resp.GetTo().ToModel(),
		CaptureName: resp.GetCaptureName(),
	}
	return result, true
}

func (it *rpcIterator) Err() error {
	return it.err
}

func (it *rpcIterator) Close() error {
	return it.stream.CloseSend()
}
