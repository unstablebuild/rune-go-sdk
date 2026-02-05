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
	"fmt"
	"strings"

	"github.com/unstablebuild/rune-go-sdk/api/storageapi/docmarshal"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/storagerpc/docpb"
	"google.golang.org/grpc"
)

// RegisterCollectionDocumentService registers the given srv with the given registrar
// with a custom service description name, that enables registering one service per
// collection. Client.InitWithCollection should be used on the client side to talk
// to a service registered via this function.
func RegisterCollectionDocumentService(
	registrar grpc.ServiceRegistrar, srv docpb.DocumentStoreServer, collection string,
) {
	desc := docpb.DocumentStore_ServiceDesc
	desc.ServiceName = fmt.Sprintf("proto.DocumentStore.%s", collection)
	var newMethods []grpc.MethodDesc
	for _, method := range desc.Methods {
		method.Handler = updateMethodInfoUnaryHandler(desc.ServiceName, method.Handler)
		newMethods = append(newMethods, method)
	}
	desc.Methods = newMethods
	// stream method name is not mangling because it operates at a lower level
	// and the full method is defined when the path is matched, which is how
	// unary should work.
	registrar.RegisterService(&desc, srv)
}

func updateMethodInfoUnaryHandler(
	newServiceName string,
	// for some reason unary handlers are a private type!
	prev func(
		srv interface{}, ctx context.Context,
		dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor,
	) (interface{}, error),
) func(
	srv interface{}, ctx context.Context, dec func(interface{}) error,
	interceptor grpc.UnaryServerInterceptor) (
	interface{}, error,
) {
	return func(
		srv interface{}, ctx context.Context,
		dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor,
	) (interface{}, error) {
		if interceptor != nil {
			interceptor = updateMethodInfoUnaryInterceptor(newServiceName, interceptor)
		}
		return prev(srv, ctx, dec, interceptor)
	}
}

func updateMethodInfoUnaryInterceptor(
	newServiceName string, prev grpc.UnaryServerInterceptor,
) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (interface{}, error) {
		info.FullMethod = strings.ReplaceAll(info.FullMethod,
			"proto.DocumentStore", newServiceName)
		return prev(ctx, req, info, handler)
	}
}

// InitWithCollection initializes this Client with the given grpc connection,
// encoding marhshaler, and uses collection to suffix the service descriptor to
// enable multiple collection services registered in the same server. Server
// should be registered with grpc via RegisterCollectionDocumentService.
func (c *Client) InitWithCollection(
	cc grpc.ClientConnInterface, m docmarshal.Marshaler, collection string,
) {
	c.cc = cc
	c.pb = newDocumentStoreClient(cc, collection)
	c.marshaler = m
}

type documentStoreClient struct {
	cc         grpc.ClientConnInterface
	collection string
}

func newDocumentStoreClient(
	cc grpc.ClientConnInterface, collection string,
) docpb.DocumentStoreClient {
	return &documentStoreClient{cc, collection}
}

func (c *documentStoreClient) Create(
	ctx context.Context, in *docpb.CreateDocumentRequest, opts ...grpc.CallOption,
) (*docpb.CreateDocumentResponse, error) {
	out := new(docpb.CreateDocumentResponse)
	err := c.cc.Invoke(ctx,
		fmt.Sprintf("/proto.DocumentStore.%s/Create", c.collection), in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *documentStoreClient) Set(
	ctx context.Context, in *docpb.SetDocumentRequest, opts ...grpc.CallOption,
) (*docpb.DocumentResponse, error) {
	out := new(docpb.DocumentResponse)
	err := c.cc.Invoke(ctx,
		fmt.Sprintf("/proto.DocumentStore.%s/Set", c.collection), in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *documentStoreClient) Update(
	ctx context.Context, in *docpb.UpdateDocumentRequest, opts ...grpc.CallOption,
) (*docpb.UpdateDocumentResponse, error) {
	out := new(docpb.UpdateDocumentResponse)
	err := c.cc.Invoke(ctx,
		fmt.Sprintf("/proto.DocumentStore.%s/Update", c.collection), in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *documentStoreClient) Get(
	ctx context.Context, in *docpb.GetDocumentRequest, opts ...grpc.CallOption,
) (*docpb.GetDocumentResponse, error) {
	out := new(docpb.GetDocumentResponse)
	err := c.cc.Invoke(ctx,
		fmt.Sprintf("/proto.DocumentStore.%s/Get", c.collection), in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *documentStoreClient) Delete(
	ctx context.Context, in *docpb.DeleteDocumentRequest, opts ...grpc.CallOption,
) (*docpb.DocumentResponse, error) {
	out := new(docpb.DocumentResponse)
	err := c.cc.Invoke(ctx,
		fmt.Sprintf("/proto.DocumentStore.%s/Delete", c.collection), in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *documentStoreClient) List(
	ctx context.Context, in *docpb.ListDocumentRequest, opts ...grpc.CallOption,
) (docpb.DocumentStore_ListClient, error) {
	stream, err := c.cc.NewStream(ctx, &docpb.DocumentStore_ServiceDesc.Streams[0],
		fmt.Sprintf("/proto.DocumentStore.%s/List", c.collection), opts...)
	if err != nil {
		return nil, err
	}
	x := &documentStoreListClient{stream}
	if err := x.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type documentStoreListClient struct {
	grpc.ClientStream
}

func (x *documentStoreListClient) Recv() (*docpb.ListDocumentResponse, error) {
	m := new(docpb.ListDocumentResponse)
	if err := x.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}
