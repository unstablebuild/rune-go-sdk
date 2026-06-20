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
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"github.com/unstablebuild/rune-go-sdk/term"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ syntaxapi.Parser = (*Client)(nil)

// Client satisfies syntaxapi.Parser via gRPC.
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

// Search satisfies syntaxapi.Parser.
func (c *Client) Search(
	query string,
	captureNames []string,
	languages ...string,
) (iterator.Iterator[syntaxapi.Result], error) {
	req := SearchRequest{Query: query, CaptureNames: captureNames, Languages: languages}
	stream, err := c.pb.Search(c.ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("syntax search: %w", err)
	}
	return &rpcIterator{stream: stream}, nil
}

// SearchNode satisfies syntaxapi.Parser.
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

// Query satisfies syntaxapi.Parser.
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

// QueryNode satisfies syntaxapi.Parser.
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

// Highlight satisfies syntaxapi.Parser.
func (c *Client) Highlight(
	uri workspaceapi.URI,
	content string,
) (iterator.Iterator[textapi.Location], error) {
	req := HighlightRequest{Uri: uri.String(), Content: content}
	stream, err := c.pb.Highlight(c.ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("syntax highlight: %w", err)
	}
	return &highlightIterator{stream: stream}, nil
}

// errNoDotDetail is the gRPC status message used to carry syntaxapi.ErrNoDot
// across the process boundary so the client can reconstruct it.
const errNoDotDetail = "syntaxapi.ErrNoDot"

// ResolveSymbol satisfies syntaxapi.Parser.
func (c *Client) ResolveSymbol(
	ctx context.Context,
	name string,
	progress syntaxapi.Progress,
) (iterator.Iterator[syntaxapi.Match], error) {
	stream, err := c.pb.ResolveSymbol(ctx, &ResolveSymbolRequest{Name: name})
	if err != nil {
		return nil, fmt.Errorf("syntax resolve symbol: %w", err)
	}
	return &resolveIterator{stream: stream, progress: progress}, nil
}

type resolveIterator struct {
	stream   grpc.ServerStreamingClient[ResolveSymbolResponse]
	progress syntaxapi.Progress
	err      error
}

func (it *resolveIterator) Next(_ context.Context) (syntaxapi.Match, bool) {
	for {
		resp, err := it.stream.Recv()
		if err == io.EOF {
			return syntaxapi.Match{}, false
		}
		if err != nil {
			if st, ok := status.FromError(err); ok &&
				st.Code() == codes.InvalidArgument && st.Message() == errNoDotDetail {
				it.err = syntaxapi.ErrNoDot
			} else {
				it.err = err
			}
			return syntaxapi.Match{}, false
		}
		if p := resp.GetProgress(); p != nil {
			if it.progress != nil {
				it.progress.Report(p.GetMessage(), int(p.GetFound()), p.GetStep(), p.GetTotal())
			}
			continue
		}
		m := resp.GetMatch()
		if m == nil {
			continue
		}
		return syntaxapi.Match{
			URI: m.GetUri(),
			Pos: term.Coordinates{
				X: int(m.GetCharacter()),
				Y: int(m.GetLine()),
			},
			Display:    m.GetDisplay(),
			ImportPath: m.GetImportPath(),
		}, true
	}
}

func (it *resolveIterator) Err() error {
	return it.err
}

func (it *resolveIterator) Close() error {
	return it.stream.CloseSend()
}

// ListReferencedSymbols satisfies syntaxapi.Parser.
func (c *Client) ListReferencedSymbols(
	ctx context.Context,
) (iterator.Iterator[string], error) {
	stream, err := c.pb.ListReferencedSymbols(ctx, &ListReferencedSymbolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("syntax list referenced symbols: %w", err)
	}
	return &listReferencedIterator{stream: stream}, nil
}

type listReferencedIterator struct {
	stream grpc.ServerStreamingClient[ListReferencedSymbolsResponse]
	err    error
}

func (it *listReferencedIterator) Next(_ context.Context) (string, bool) {
	resp, err := it.stream.Recv()
	if err == io.EOF {
		return "", false
	}
	if err != nil {
		it.err = err
		return "", false
	}
	return resp.GetName(), true
}

func (it *listReferencedIterator) Err() error {
	return it.err
}

func (it *listReferencedIterator) Close() error {
	return it.stream.CloseSend()
}

type highlightIterator struct {
	stream grpc.ServerStreamingClient[HighlightResponse]
	err    error
}

func (it *highlightIterator) Next(_ context.Context) (textapi.Location, bool) {
	resp, err := it.stream.Recv()
	if err == io.EOF {
		return textapi.Location{}, false
	}
	if err != nil {
		it.err = err
		return textapi.Location{}, false
	}
	loc := textapi.Location{
		From:    resp.GetFrom().ToModel(),
		To:      resp.GetTo().ToModel(),
		Attr:    resp.GetAttr().ToModel(),
		Message: resp.GetMessage(),
	}
	return loc, true
}

func (it *highlightIterator) Err() error {
	return it.err
}

func (it *highlightIterator) Close() error {
	return it.stream.CloseSend()
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
