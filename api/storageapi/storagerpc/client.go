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

package storagerpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"strings"

	"github.com/unstablebuild/rune-go-sdk/api/storageapi"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/docmarshal"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/storagerpc/docpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	marshaler      docmarshal.Marshaler
	cc             grpc.ClientConnInterface
	pb             docpb.DocumentStoreClient
	ownsConnection bool
	partitions     []string
	closeFn        func() error
}

// NewClient returns a grpc-based client that satisfies Service
// by relaying operations to remote datastore server. See NewServer
// for more details.
func NewClient(
	addr net.Addr, m docmarshal.Marshaler, opts ...grpc.DialOption,
) (storageapi.Service, error) {
	opts = append(opts, grpc.WithContextDialer(
		func(ctx context.Context, _ string) (net.Conn, error) {
			var d net.Dialer
			d.Deadline, _ = ctx.Deadline()
			conn, err := net.Dial(addr.Network(), addr.String())
			if err != nil {
				return nil, err
			}
			if tcpConn, ok := conn.(*net.TCPConn); ok {
				// Make sure to set keep alive so that the connection doesn't die
				err := tcpConn.SetKeepAlive(true)
				if err != nil {
					return nil, fmt.Errorf("tcp conn set keep alive: %w", err)
				}
			}
			return conn, err
		},
	))
	cc, err := grpc.NewClient("passthrough:///", opts...)
	if err != nil {
		return nil, err
	}

	ret := new(Client)
	ret.ownsConnection = true
	ret.Init(cc, m)
	return ret, nil
}

func (c *Client) Init(cc grpc.ClientConnInterface, m docmarshal.Marshaler) {
	c.cc = cc
	c.pb = docpb.NewDocumentStoreClient(cc)
	c.marshaler = m
	c.partitions = nil
	c.closeFn = func() error {
		if closer, ok := c.cc.(io.Closer); ok && c.ownsConnection {
			return closer.Close()
		}
		return nil
	}
}

func (c *Client) partitionContext(ctx context.Context) context.Context {
	return ContextWithPartitions(ctx, c.partitions...)
}

func (c *Client) Create(
	ctx context.Context, ID string, data interface{},
) error {
	bytes, err := c.encodeCreateData(data)
	if err != nil {
		return err
	}

	req := docpb.CreateDocumentRequest{Id: ID, Data: bytes}
	res, err := c.pb.Create(c.partitionContext(ctx), &req)
	if err != nil {
		return convertRpcError(err)
	}
	if res.GetAlreadyExists() {
		return storageapi.ErrAlreadyExists
	}
	return nil
}

func (c *Client) Set(
	ctx context.Context, ID string, data interface{},
) error {
	bytes, err := c.encodeCreateData(data)
	if err != nil {
		return err
	}

	req := docpb.SetDocumentRequest{Id: ID, Data: bytes}
	_, err = c.pb.Set(c.partitionContext(ctx), &req)
	if err != nil {
		return convertRpcError(err)
	}
	return nil
}

func (c *Client) Update(
	ctx context.Context, ID string, updates []storageapi.Update,
	preconds ...storageapi.Precondition,
) error {
	if len(updates) == 0 {
		panic("invalid arguments: empty updates")
	}
	u := makeProtoUpdates(c.marshaler, updates)
	p := makeProtoPreconditions(c.marshaler, preconds...)
	req := docpb.UpdateDocumentRequest{Id: ID, Updates: u, Preconditions: p}
	res, err := c.pb.Update(c.partitionContext(ctx), &req)
	if err != nil {
		return convertRpcError(err)
	}
	if res.GetNotFound() {
		return storageapi.ErrNotFound
	}
	if res.GetPreconditionFailed() {
		return storageapi.ErrPreconditionFailed
	}
	return nil
}

func (c *Client) Get(
	ctx context.Context, ID string, doc interface{},
) error {
	req := docpb.GetDocumentRequest{Id: ID}
	res, err := c.pb.Get(c.partitionContext(ctx), &req)
	if err != nil {
		return convertRpcError(err)
	}
	if res.GetNotFound() {
		return storageapi.ErrNotFound
	}

	data := res.GetData()
	err = storageapi.SafeDecode(c.marshaler, doc, data)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Delete(
	ctx context.Context, ID string,
) error {
	req := docpb.DeleteDocumentRequest{Id: ID}
	_, err := c.pb.Delete(c.partitionContext(ctx), &req)
	if err != nil {
		return convertRpcError(err)
	}
	return nil
}

type rpcIterator struct {
	marshaler docmarshal.Marshaler
	cc        docpb.DocumentStore_ListClient
	next      *docpb.ListDocumentResponse
	nextErr   error
}

func (l *rpcIterator) HasNext() bool {
	if l.next != nil {
		return true
	}

	m := new(docpb.ListDocumentResponse)
	l.nextErr = l.cc.RecvMsg(m)
	if l.nextErr == io.EOF {
		return false
	}
	l.next = m
	return true
}

func (l *rpcIterator) NextTo(doc interface{}) error {
	if l.next == nil && l.nextErr == nil {
		if !l.HasNext() {
			return io.EOF
		}
	}
	next := l.next
	err := l.nextErr
	l.nextErr = nil
	l.next = nil
	if err != nil {
		return convertRpcError(err)
	}

	if errStr := next.GetError(); errStr != "" {
		switch {
		case strings.Contains(errStr, storageapi.ErrNotFound.Error()):
			return storageapi.ErrNotFound
		case strings.Contains(errStr, storageapi.ErrAlreadyExists.Error()):
			return storageapi.ErrAlreadyExists
		case strings.Contains(errStr, storageapi.ErrPreconditionFailed.Error()):
			return storageapi.ErrPreconditionFailed
		case strings.Contains(errStr, storageapi.ErrPermissionDenied.Error()):
			return storageapi.ErrPermissionDenied
		default:
			return errors.New(errStr)
		}
	}
	return storageapi.SafeDecode(l.marshaler, doc, next.GetData())
}

func (l *rpcIterator) Close() error {
	return l.cc.CloseSend()
}

func (c *Client) List(
	ctx context.Context, filters []storageapi.Filter,
) (storageapi.Iterator, error) {
	f, err := makeProtoFilters(c.marshaler, filters)
	if err != nil {
		return nil, err
	}
	req := docpb.ListDocumentRequest{Filters: f}
	req.Fields = fieldsFromContext(ctx)
	res, err := c.pb.List(c.partitionContext(ctx), &req)
	if err != nil {
		return nil, convertRpcError(err)
	}
	return &rpcIterator{marshaler: c.marshaler, cc: res}, nil
}

func (c *Client) Close() error {
	if c.closeFn != nil {
		return c.closeFn()
	}
	return nil
}

func (c *Client) Partition(name string) (storageapi.Service, error) {
	ret := *c
	ret.partitions = append(append([]string(nil), c.partitions...), name)
	return &ret, nil
}

func (c *Client) encodeCreateData(data interface{}) ([]byte, error) {
	if data == nil {
		panic("invalid nil data argument to Create/Set")
	}
	data, err := storageapi.DerefCreateValue(reflect.ValueOf(data))
	if err != nil {
		return nil, err
	}

	return storageapi.Encode(c.marshaler, data, true), nil
}

func convertRpcError(err error) error {
	status, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch status.Code() {
	case codes.FailedPrecondition:
		return storageapi.ErrPreconditionFailed
	case codes.NotFound:
		return storageapi.ErrNotFound
	case codes.AlreadyExists:
		return storageapi.ErrAlreadyExists
	case codes.Unauthenticated, codes.PermissionDenied:
		return storageapi.ErrPermissionDenied
	default:
		return err
	}
}
