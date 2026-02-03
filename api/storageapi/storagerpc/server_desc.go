// Unstable Build LLC ("COMPANY") CONFIDENTIAL
//
// Unpublished Copyright (c) 2018-2024 Unstable Build, All Rights Reserved.
//
// NOTICE: All information contained herein is, and remains the property of COMPANY.
// The intellectual and technical concepts contained herein are proprietary to
// COMPANY and may be covered by U.S. and Foreign Patents, patents in process,
// and are protected by trade secret or copyright law. Dissemination of this information
// or reproduction of this material is strictly forbidden unless prior written permission
// is obtained from COMPANY. Access to the source code contained herein is hereby
// forbidden to anyone except current COMPANY employees, managers or contractors who
// have executed Confidentiality and Non-disclosure agreements explicitly covering such access.
//
// The copyright notice above does not evidence any actual or intended publication or
// disclosure of this source code, which includes information that is confidential and/or
// proprietary, and is a trade secret, of COMPANY. ANY REPRODUCTION, MODIFICATION,
// DISTRIBUTION, PUBLIC  PERFORMANCE, OR PUBLIC DISPLAY OF OR THROUGH USE OF THIS SOURCE CODE
// WITHOUT  THE EXPRESS WRITTEN CONSENT OF COMPANY IS STRICTLY PROHIBITED, AND IN
// VIOLATION OF APPLICABLE LAWS AND INTERNATIONAL TREATIES. THE RECEIPT OR POSSESSION OF
// THIS SOURCE CODE AND/OR RELATED INFORMATION DOES NOT CONVEY OR IMPLY ANY RIGHTS TO
// REPRODUCE, DISCLOSE OR DISTRIBUTE ITS CONTENTS, OR TO MANUFACTURE, USE, OR SELL
// ANYTHING THAT IT MAY DESCRIBE, IN WHOLE OR IN PART.

package storagerpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/unstablebuild/rune-go-sdk/api/storageapi/storagerpc/docmarshal"
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
