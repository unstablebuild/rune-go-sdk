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

package mcptools

// The tests in this file serve the tools from a real MCP server over its stdio and
// streamable HTTP transports and drive them with a real MCP client, so they cover what a
// client sees on the wire: the handshake, tools/list with schemas and _meta, and every
// tool's result.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/require"
)

// connect serves s over a transport and returns an initialized client of it.
type connect func(t *testing.T, s *server.MCPServer) *client.Client

var transports = map[string]connect{"stdio": connectStdio, "http": connectHTTP}

// connectStdio serves s with the stdio transport over a pair of pipes, as a process
// serving on its stdin and stdout would.
func connectStdio(t *testing.T, s *server.MCPServer) *client.Client {
	t.Helper()
	serverIn, clientOut := io.Pipe()
	clientIn, serverOut := io.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	stdio := server.NewStdioServer(s)
	stdio.SetErrorLogger(log.New(io.Discard, "", 0))
	done := make(chan error, 1)
	go func() {
		done <- stdio.Listen(ctx, serverIn, serverOut)
		_ = serverOut.Close()
	}()

	c := client.NewClient(transport.NewIO(clientIn, clientOut, nil))
	require.NoError(t, c.Start(ctx))
	t.Cleanup(func() {
		// Closing the client closes the server's stdin.
		require.NoError(t, c.Close())
		select {
		case err := <-done:
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Errorf("stdio server: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("stdio server did not stop when its input closed")
		}
		cancel()
	})
	initialize(t, c)
	return c
}

// connectHTTP serves s with the streamable HTTP transport on a loopback port.
func connectHTTP(t *testing.T, s *server.MCPServer) *client.Client {
	t.Helper()
	ts := server.NewTestStreamableHTTPServer(s)
	t.Cleanup(ts.Close)
	c, err := client.NewStreamableHttpClient(ts.URL)
	require.NoError(t, err)
	require.NoError(t, c.Start(context.Background()))
	t.Cleanup(func() { _ = c.Close() })
	initialize(t, c)
	return c
}

func initialize(t *testing.T, c *client.Client) {
	t.Helper()
	var req mcp.InitializeRequest
	req.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	req.Params.ClientInfo = mcp.Implementation{Name: "mcptools-test-client",
		Version: "0.0.1"}
	res, err := c.Initialize(ctxTimeout(t), req)
	require.NoError(t, err)
	require.Equal(t, "mcptools-test", res.ServerInfo.Name)
	require.Equal(t, "1.2.3", res.ServerInfo.Version)
	require.NotNil(t, res.Capabilities.Tools, "tools capability advertised")
}

func ctxTimeout(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func clientCall(
	t *testing.T,
	c *client.Client,
	name string,
	args map[string]any,
) (*mcp.CallToolResult, error) {
	t.Helper()
	var req mcp.CallToolRequest
	req.Params.Name = name
	req.Params.Arguments = args
	return c.CallTool(ctxTimeout(t), req)
}

func listClientTools(t *testing.T, c *client.Client) map[string]mcp.Tool {
	t.Helper()
	res, err := c.ListTools(ctxTimeout(t), mcp.ListToolsRequest{})
	require.NoError(t, err)
	require.Empty(t, res.NextCursor, "every tool on one page")
	out := make(map[string]mcp.Tool, len(res.Tools))
	for _, tl := range res.Tools {
		require.NotContains(t, out, tl.Name, "tool listed twice")
		out[tl.Name] = tl
	}
	return out
}

func TestIntegrationListTools(t *testing.T) {
	for name, connect := range transports {
		t.Run(name, func(t *testing.T) {
			t.Run("core", func(t *testing.T) {
				c := connect(t, newTestServer().MCPServer)
				names := slices.Sorted(maps.Keys(listClientTools(t, c)))
				require.Equal(t, coreTools, names)
			})
			t.Run("all", func(t *testing.T) {
				c := connect(t, newTestServer(WithAllThemes()).MCPServer)
				tools := listClientTools(t, c)
				require.Equal(t, allTools(), slices.Sorted(maps.Keys(tools)))
				for _, tc := range toolCases() {
					tl := tools[tc.tool]
					require.NotEmpty(t, tl.Description, tc.tool)
					require.Equal(t, "object", tl.InputSchema.Type, tc.tool)
					required := slices.Clone(tl.InputSchema.Required)
					slices.Sort(required)
					if len(required) == 0 {
						required = nil
					}
					require.Equal(t, tc.required, required, tc.tool)
				}
			})
			t.Run("always load", func(t *testing.T) {
				c := connect(t, newTestServer(WithAllThemes()).MCPServer)
				var marked []string
				for name, tl := range listClientTools(t, c) {
					if tl.Meta == nil {
						continue
					}
					require.Equal(t, map[string]any{AlwaysLoadMetaKey: true},
						tl.Meta.AdditionalFields, name)
					marked = append(marked, name)
				}
				slices.Sort(marked)
				require.Equal(t, []string{
					"lsp_definition", "lsp_diagnostics", "lsp_document_symbols",
					"lsp_implementation", "lsp_references", "lsp_workspace_symbols",
				}, marked)
			})
		})
	}
}

func TestIntegrationCallEveryTool(t *testing.T) {
	for name, connect := range transports {
		t.Run(name, func(t *testing.T) {
			s := newTestServer(WithAllThemes())
			c := connect(t, s.MCPServer)
			for _, tc := range toolCases() {
				t.Run(tc.tool, func(t *testing.T) {
					res, err := clientCall(t, c, tc.tool, tc.args)
					require.NoError(t, err)
					require.False(t, res.IsError, resultText(t, *res))
					require.JSONEq(t, tc.want, resultText(t, *res))
					require.Nil(t, res.StructuredContent)
				})
			}
			require.Len(t, s.lsp.called("Hover"), 1)
			require.Len(t, s.parser.called("Search"), 1)
		})
	}
}

func TestIntegrationErrors(t *testing.T) {
	for name, connect := range transports {
		t.Run(name, func(t *testing.T) {
			s := newTestServer()
			s.lsp.failOn("Diagnostic", errors.New("gopls exited"))
			c := connect(t, s.MCPServer)

			t.Run("missing argument is a tool error", func(t *testing.T) {
				res, err := clientCall(t, c, "lsp_definition", map[string]any{})
				require.NoError(t, err)
				require.True(t, res.IsError)
				require.Contains(t, resultText(t, *res), `"symbol"`)
			})
			t.Run("invalid argument is a tool error", func(t *testing.T) {
				res, err := clientCall(t, c, "syntax_query_node",
					map[string]any{"uri": mainGo, "node_types": "nope"})
				require.NoError(t, err)
				require.True(t, res.IsError)
				require.Contains(t, resultText(t, *res), `invalid node type "nope"`)
			})
			t.Run("backend failure is a tool error", func(t *testing.T) {
				res, err := clientCall(t, c, "lsp_diagnostics",
					map[string]any{"uri": mainGo})
				require.NoError(t, err)
				require.True(t, res.IsError)
				require.Equal(t, "gopls exited", resultText(t, *res))
			})
			t.Run("unknown tool is a protocol error", func(t *testing.T) {
				_, err := clientCall(t, c, "no_such_tool", nil)
				require.ErrorContains(t, err, "no_such_tool")
			})
			t.Run("disabled theme is a protocol error", func(t *testing.T) {
				_, err := clientCall(t, c, "lsp_rename",
					map[string]any{"symbol": "MyFunc", "new_name": "X"})
				require.ErrorContains(t, err, "lsp_rename")
				require.Empty(t, s.lsp.called("Rename"))
			})
			t.Run("the connection survives errors", func(t *testing.T) {
				res, err := clientCall(t, c, "lsp_definition",
					map[string]any{"symbol": "MyFunc"})
				require.NoError(t, err)
				require.False(t, res.IsError, resultText(t, *res))
			})
		})
	}
}

func TestIntegrationConcurrentCalls(t *testing.T) {
	for name, connect := range transports {
		t.Run(name, func(t *testing.T) {
			s := newTestServer(WithAllThemes())
			c := connect(t, s.MCPServer)
			cases := toolCases()
			const workers = 8
			var wg sync.WaitGroup
			errs := make(chan error, workers*len(cases))
			for w := range workers {
				wg.Go(func() {
					for i := range cases {
						tc := cases[(i+w)%len(cases)]
						var req mcp.CallToolRequest
						req.Params.Name = tc.tool
						req.Params.Arguments = tc.args
						ctx, cancel := context.WithTimeout(context.Background(),
							10*time.Second)
						res, err := c.CallTool(ctx, req)
						cancel()
						if err != nil {
							errs <- fmt.Errorf("%s: %w", tc.tool, err)
							continue
						}
						if got := firstText(res); res.IsError || !jsonEqual(tc.want, got) {
							errs <- fmt.Errorf("%s: got %s", tc.tool, got)
						}
					}
				})
			}
			wg.Wait()
			close(errs)
			for err := range errs {
				t.Error(err)
			}
			require.Len(t, s.lsp.called("Hover"), workers)
		})
	}
}

func firstText(res *mcp.CallToolResult) string {
	if len(res.Content) != 1 {
		return fmt.Sprintf("%d content blocks", len(res.Content))
	}
	if tc, ok := res.Content[0].(mcp.TextContent); ok {
		return tc.Text
	}
	return fmt.Sprintf("%T", res.Content[0])
}

// jsonEqual reports whether a and b are equal JSON documents, as require.JSONEq would,
// for use off the test goroutine.
func jsonEqual(a, b string) bool {
	var av, bv any
	if json.Unmarshal([]byte(a), &av) != nil || json.Unmarshal([]byte(b), &bv) != nil {
		return false
	}
	return reflect.DeepEqual(av, bv)
}
