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

package commandrpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/unstablebuild/rune-go-sdk/api/commandapi"
	"github.com/unstablebuild/rune-go-sdk/debug"
	"google.golang.org/grpc"
)

var _ commandapi.CommandRegistry = (*Client)(nil)

// Client satisfies commandapi.CommandRegister by calling a
// remote command service over grpc.
type Client struct {
	cmd             CommandClient
	clientCtx       context.Context
	clientCancelCtx func()
}

// NewClient allocates storage for a new Client and
// initializes it.
func NewClient(
	ctx context.Context, cc grpc.ClientConnInterface,
) *Client {
	ret := new(Client)
	ret.Init(ctx, cc)
	return ret
}

// Init initializes this Client.
func (c *Client) Init(
	ctx context.Context, cc grpc.ClientConnInterface,
) {
	c.cmd = NewCommandClient(cc)
	c.clientCtx, c.clientCancelCtx = context.WithCancel(ctx)
}

// RegisterCommand satisfies commandapi.CommandRegister.
func (c *Client) RegisterCommand(
	man commandapi.CommandManual, h commandapi.CommandHandler,
) error {
	stream, err := c.cmd.SubscribeCommand(c.clientCtx)
	if err != nil {
		return err
	}
	rpcMan := makeProtoManual(man)
	req := SubscribeCommandRequest{Command: &rpcMan}
	sendMsg := ClientCommandMessage{
		Request: &req,
		Type:    ClientCommandMessage_Request,
	}

	if err := stream.Send(&sendMsg); err != nil {
		return fmt.Errorf(
			"send subscribe command request: %w", err,
		)
	}

	var recvMsg ServerCommandMessage
	err = stream.RecvMsg(&recvMsg)
	if err != nil {
		return fmt.Errorf(
			"send subscribe command request: %w", err,
		)
	}

	if recvMsg.GetType() != ServerCommandMessage_Response ||
		recvMsg.GetResponse() == nil {
		return errors.New(
			"recv subscribe command response: nil response",
		)
	}

	srvStream := newCommandServerStream(
		c.clientCtx, stream, h,
	)
	go debug.CapturePanicReport(srvStream.receiveMessages)

	return nil
}

// RegisterREPLCommand satisfies commandapi.CommandRegister.
func (c *Client) RegisterREPLCommand(
	man commandapi.CommandManual, h commandapi.REPLHandler,
) error {
	stream, err := c.cmd.SubscribeREPLCommand(c.clientCtx)
	if err != nil {
		return err
	}
	rpcMan := makeProtoManual(man)
	req := SubscribeREPLCommandRequest{Command: &rpcMan}
	sendMsg := ClientREPLCommandMessage{
		Request: &req,
		Type:    ClientREPLCommandMessage_Request,
	}

	if err := stream.Send(&sendMsg); err != nil {
		return fmt.Errorf(
			"send subscribe repl command request: %w", err,
		)
	}

	var recvMsg ServerREPLCommandMessage
	err = stream.RecvMsg(&recvMsg)
	if err != nil {
		return fmt.Errorf(
			"recv subscribe repl command response: %w", err,
		)
	}

	if recvMsg.GetType() != ServerREPLCommandMessage_Response ||
		recvMsg.GetResponse() == nil {
		return errors.New(
			"recv subscribe repl command response: " +
				"nil response",
		)
	}

	srvStream := newREPLCommandServerStream(
		c.clientCtx, stream, h,
	)
	go debug.CapturePanicReport(srvStream.receiveMessages)

	return nil
}

// Close closes all resources associated with this client.
func (c *Client) Close() {
	if c.clientCancelCtx != nil {
		c.clientCancelCtx()
		c.clientCancelCtx = nil
	}
}

func makeProtoManual(
	man commandapi.CommandManual,
) CommandManual {
	var cmds []*CommandManual
	for _, cmd := range man.Commands {
		childManual := new(CommandManual)
		*childManual = makeProtoManual(cmd)
		cmds = append(cmds, childManual)
	}
	ret := CommandManual{
		Name:     man.Name,
		Summary:  man.Summary,
		Synopsis: man.Synopsis,
		Commands: cmds,
	}
	return ret // nolint:govet
}
