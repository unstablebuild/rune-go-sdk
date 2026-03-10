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

package textrpc

import (
	"context"
	"fmt"
	"time"

	"github.com/unstablebuild/rune-go-sdk/api/browserapi/browserrpc"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/debug"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/term/termrpc"
	"google.golang.org/grpc"
)

const (
	defaultTimeout = 4 * time.Second
)

var _ textapi.Handler = Token{}

// Token wraps a browser.Token to satisfy editor.Handler.
type Token struct {
	browserrpc.Token
	workspaceapi.URI
}

// Resource satisfies text.Handler
func (t Token) Resource() workspaceapi.URI {
	return t.URI
}

// SetWrap satisfies text.Handler
func (t Token) SetWrap(wrap bool) {
}

// SetCursorAtScroll satisfies text.Handler
func (t Token) SetCursorAtScroll(term.Coordinates) bool {
	return false
}

// ShowCommandBar satisfies text.Handler.
func (t Token) ShowCommandBar(show bool) {
}

// SeekUp satisfies text.Handler.
func (t Token) SeekUp() bool {
	return false
}

// SeekDown satisfies text.Handler.
func (t Token) SeekDown() bool {
	return false
}

// SeekOffset satisfies text.Handler.
func (t Token) SeekOffset() int {
	return 0
}

// MaxSeekOffset satisfies text.Handler.
func (t Token) MaxSeekOffset() int {
	return 0
}

var _ textapi.Editor = (*Client)(nil)

// Client satisfies textapi.Editor by calling a remote editor over grpc.
type Client struct {
	browser         *browserrpc.Client
	cc              grpc.ClientConnInterface
	ed              EditorClient
	clientCtx       context.Context
	clientCancelCtx func()
}

// NewClient allocates storage for a new Client and initializes it.
func NewClient(ctx context.Context, cc grpc.ClientConnInterface) *Client {
	ret := new(Client)
	ret.Init(ctx, cc)
	return ret
}

// Init initializes this Client with broker and client.
func (c *Client) Init(ctx context.Context, cc grpc.ClientConnInterface) {
	c.ed = NewEditorClient(cc)
	c.cc = cc
	c.browser = browserrpc.NewClient(ctx, cc)
	c.clientCtx, c.clientCancelCtx = context.WithCancel(ctx)
}

// Client returns the underlying grpc EditorClient.
func (c *Client) Client() EditorClient {
	return c.ed
}

// Editor satisfies textapi.Editor
func (c *Client) Editor(file workspaceapi.URI) (textapi.Handler, error) {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	req := EditorRequest{ResourceName: NewURI(file)}

	_, err := c.ed.Editor(ctx, &req)
	if err != nil {
		return nil, err
	}

	return Token{URI: file}, nil
}

// SubscribeEvents requests the editor server to subscribe sub to ev.
func (c *Client) SubscribeEvents(
	evs []textapi.EventType, h textapi.EventHandler,
) error {
	stream, err := c.ed.SubscribeEvent(c.clientCtx)
	if err != nil {
		return err
	}

	var req SubscribeEventRequest
	for _, ev := range evs {
		req.Type = append(req.Type, protoType(textapi.Event{Type: ev}))
	}

	err = stream.Send(&req)
	if err != nil {
		return fmt.Errorf("stream send request: %v", err)
	}

	handler := newEventStreamServer(c.clientCtx, stream, h)
	go debug.CapturePanicReport(handler.receiveEvents)

	return nil
}

// SetLocationList requests the editor server to set l as the new location list for h.
// Note that h is expected to be the return valu of Edit or a dispatched event, delivered
// via an EventHandler.
func (c *Client) SetLocationList(
	h textapi.Handler, pri textapi.LocationPriority, ID string, l textapi.LocationList,
) error {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	token := h.(Token)
	req := makeLocationListRequest(token.URI, pri, ID, l)
	_, err := c.ed.SetLocationList(ctx, &req)
	return err
}

// MoveToPrevLocation requests the editor server to move cursor to the previous location
// in location list identified by ID.
func (c *Client) MoveToPrevLocation(h textapi.Handler, ID string) error {
	err := c.moveToLocation(h, ID, false)
	return err
}

// MoveToNextLocation requests the editor server to move cursor to the next location
// in location list identified by ID.
func (c *Client) MoveToNextLocation(h textapi.Handler, ID string) error {
	err := c.moveToLocation(h, ID, true)
	return err
}

// SetCursor requests the editor server to move cursor to pos
func (c *Client) SetCursor(h textapi.Handler, pos term.Coordinates) error {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	token := h.(Token)
	var protoPos termrpc.Coordinates
	protoPos.FromModel(pos)
	req := SetCursorRequest{Pos: &protoPos, ResourceName: NewURI(token.URI)}
	_, err := c.ed.SetCursor(ctx, &req)
	return err
}

// Cursor requests the editor server to move cursor to pos
func (c *Client) Cursor(h textapi.Handler) (term.Coordinates, error) {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	token := h.(Token)
	req := CursorRequest{ResourceName: NewURI(token.URI)}
	res, err := c.ed.Cursor(ctx, &req)
	if err != nil {
		return term.Coordinates{}, err
	}
	return res.GetPos().ToModel(), nil
}

// CellEditor satisfies textapi.Editor.
func (c *Client) CellEditor(h textapi.Handler) textapi.CellEditor {
	token := h.(Token)
	return clientWriter{client: c, uri: token.URI}
}

// CellView satisfies textapi.Editor.
func (c *Client) CellView(h textapi.Handler) textapi.CellView {
	token := h.(Token)
	return clientView{client: c, uri: token.URI}
}

// SetDefaultAttributes satisfies textapi.Editor.
func (c *Client) SetDefaultAttributes(h textapi.Handler, attrs term.Attributes) error {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	token := h.(Token)
	var rpcAttrs termrpc.Attributes
	rpcAttrs.FromModel(attrs)
	req := SetDefaultAttributesRequest{
		ResourceName: NewURI(token.URI),
		Attributes:   &rpcAttrs,
	}
	_, err := c.ed.SetDefaultAttributes(ctx, &req)
	return err
}

// Close closes all resources associated with this client.
func (c *Client) Close() (ret error) {
	if c.clientCancelCtx != nil {
		c.clientCancelCtx()
		c.clientCancelCtx = nil
		c.cc = nil
	}
	return ret
}

func (c *Client) ctxWithTimeout() (context.Context, func()) {
	ctx, cancel := context.WithTimeout(c.clientCtx, defaultTimeout)
	return ctx, cancel
}

func makeLocationListRequest(
	uri workspaceapi.URI, priority textapi.LocationPriority,
	listID string, l textapi.LocationList,
) SetLocationListRequest {
	req := SetLocationListRequest{
		ResourceName: NewURI(uri),
		ListId:       listID,
		Priority:     uint32(priority),
	}

	for loc, ok := l.Current(); ok; loc, ok = l.Next() {
		var from, to termrpc.Coordinates
		var attr termrpc.Attributes
		from.FromModel(loc.From)
		to.FromModel(loc.To)
		attr.FromModel(loc.Attr)
		req.Locations = append(req.Locations, &SetLocationListRequest_Location{
			From: &from,
			To:   &to,
			Attr: &attr,
			Msg:  loc.Message,
			Icon: loc.Icon,
		})
	}
	return req // nolint:govet
}

func (c *Client) moveToLocation(h textapi.Handler, ID string, next bool) (err error) {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	token := h.(Token)
	req := MoveToLocationRequest{ResourceName: NewURI(token.URI), ListId: ID}
	if next {
		_, err = c.ed.MoveToNextLocation(ctx, &req)
	} else {
		_, err = c.ed.MoveToPrevLocation(ctx, &req)
	}
	return err
}
