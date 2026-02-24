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

package extensionapi

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"

	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
	"github.com/unstablebuild/rune-go-sdk/api/browserapi/browserrpc"
	"github.com/unstablebuild/rune-go-sdk/api/config"
	"github.com/unstablebuild/rune-go-sdk/api/config/configrpc"
	"github.com/unstablebuild/rune-go-sdk/api/debugapi"
	"github.com/unstablebuild/rune-go-sdk/api/debugapi/debugrpc"
	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
	"github.com/unstablebuild/rune-go-sdk/api/semanticapi/semanticrpc"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/docmarshal/doctoml"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/storagerpc"
	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi"
	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi/syntaxrpc"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/api/textapi/textrpc"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi/workspacerpc"
	"github.com/unstablebuild/rune-go-sdk/term"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/credentials/oauth"
)

// Workspace abstracts resources associated with a Workspace.
type Workspace struct {
	dataDir string
	conn    grpc.ClientConnInterface
	editor  *textrpc.Client
	config  config.Config
	meta    Metadata
}

// DataDir returns the data directory to store data to re-use across sessions.
func (w *Workspace) DataDir(ctx context.Context) string {
	return w.dataDir
}

// WindowManager returns the workspace's window manager, which can be used
// to create and manipulate the workspace's browser tabs, windows and bars.
func (w *Workspace) WindowManager(ctx context.Context) browserapi.WindowManager {
	return browserrpc.NewClient(ctx, w.conn)
}

// ResourceOpener returns the workspace's resource opener which allows
// clients to access resources in the workspace. Extensions that need more control
// over the underlying files can use FileSystem instead.
func (w *Workspace) ResourceOpener(ctx context.Context) browserapi.ResourceOpener {
	return browserrpc.NewClient(ctx, w.conn)
}

// Notifications returns the workspace's notifications.
func (w *Workspace) Notifications(ctx context.Context) browserapi.Notifications {
	return browserrpc.NewClient(ctx, w.conn)
}

// FileSystem returns the workspace's file system, which can be used to
// directly manipulate raw files in the workspace. Extensions that
// want to simply open resources as tabs in a workspace should use ResourceOpener.
func (w *Workspace) FileSystem(ctx context.Context) workspaceapi.FileSystem {
	return workspacerpc.NewClient(ctx, w.conn)
}

// Executor returns the workspace's executor, which can be used
// to start and stop processes.
func (w *Workspace) Executor(ctx context.Context) workspaceapi.Executor {
	return workspacerpc.NewClient(ctx, w.conn)
}

// Terminal returns the workspace's pty capability which can be used to
// implement a terminal.
func (w *Workspace) Terminal(ctx context.Context) workspaceapi.Terminal {
	return workspacerpc.NewClient(ctx, w.conn)
}

// Editor returns the workspace's editor which can be used to monitor
// and edit files, move the cursor, add location lists with or without
// color attributes, and even set the default background and foreground
// of a resource.
func (w *Workspace) Editor(ctx context.Context) textapi.Editor {
	return textrpc.NewClient(ctx, w.conn)
}

// LSP returns the workspace's LSP client, which can be used to interact
// with language servers.
func (w *Workspace) LSP(ctx context.Context) semanticapi.LSP {
	return semanticrpc.NewClient(ctx, w.conn)
}

// Debugger returns the workspace's Debugger client, which can be used to interact
// with DAP-compatible debuggers.
func (w *Workspace) Debugger(ctx context.Context) debugapi.Debugger {
	return debugrpc.NewClient(ctx, w.conn)
}

// RegisterCommand registers a new command which is to be dispatched to the
// given CommandHandler. This is the primary mechanism extensions should use
// to run new functionality.
func (w *Workspace) RegisterCommand(
	cmd textapi.CommandManual, h textapi.CommandHandler,
) error {
	return w.editorClient(context.Background()).SubscribeCommand(cmd, h)
}

// Parser returns the workspace's syntax parser, which can be used
// to perform AST-level searches and parsing across workspace files.
func (w *Workspace) Parser(ctx context.Context) syntaxapi.Parser {
	return syntaxrpc.NewClient(ctx, w.conn)
}

// Storage returns the workspace's storage facility, which can be used to
// persist extension data across sessions.
func (w *Workspace) Storage(ctx context.Context) storageapi.Service {
	c := new(storagerpc.Client)
	c.Init(w.conn, doctoml.Marshaler())
	return storageapi.WithPartition(c, w.meta.ExtensionID)
}

// Interrupter returns the workspace's event loop interrupter, which
// can be used to request new draws when state changes asynchronously.
func (w *Workspace) Interrupter(ctx context.Context) term.Interrupter {
	return browserrpc.NewClient(ctx, w.conn)
}

// Config returns a copy of the workspace's global configuration.
func (w *Workspace) Config(ctx context.Context) config.Config {
	if w.config == nil {
		c, err := configrpc.FetchConfig(w.conn)
		if err != nil {
			return config.ErrConfig(fmt.Errorf("fetch config: %w", err))
		}
		w.config = c
	}
	return w.config
}

// RawConn returns the underlying raw connection to the extension host.
// This method should generally not be used by extensions.
func (w *Workspace) RawConn() grpc.ClientConnInterface {
	return w.conn
}

// NewWorkspace returns a new Workspace. Extensions should use
// ServeWorkspaceExtension, which performs the stdin/stdout exchange
// necessary to receive a valid Config.
func NewWorkspace(req Config, meta Metadata) (*Workspace, error) {
	ret := new(Workspace)
	ret.dataDir = req.DataDir
	opts := []grpc.DialOption{
		grpc.WithContextDialer(
			func(ctx context.Context, _ string) (net.Conn, error) {
				addr, err := net.ResolveUnixAddr("unix", req.Socket)
				if err != nil {
					return nil, err
				}
				var d net.Dialer
				return d.DialContext(ctx, addr.Network(), addr.String())
			},
		)}

	if len(req.Certificate) != 0 {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(req.Certificate) {
			return nil, fmt.Errorf("credentials: failed to append certificates")
		}
		tlsCfg := tls.Config{InsecureSkipVerify: true, RootCAs: certPool}
		transportCreds := credentials.NewTLS(&tlsCfg)
		opts = append(opts, grpc.WithTransportCredentials(transportCreds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if req.Token != nil {
		tokenSource := oauth2.StaticTokenSource(req.Token)
		rpcCreds := oauth.TokenSource{TokenSource: tokenSource}
		opts = append(opts, grpc.WithPerRPCCredentials(rpcCreds))
	}

	// passthrough is used to indicate that dns
	// resolution SHOULD NOT be performed on target
	conn, err := grpc.NewClient("passthrough:", opts...)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}
	ret.conn = conn
	ret.meta = meta
	return ret, nil
}

func (w *Workspace) editorClient(ctx context.Context) *textrpc.Client {
	if w.editor == nil {
		w.editor = w.Editor(ctx).(*textrpc.Client)
	}
	return w.editor
}
