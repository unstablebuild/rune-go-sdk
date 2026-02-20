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

package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/require"
)

// mcpTestEnv wraps testEnv with an MCP server for
// testing tool calls.
type mcpTestEnv struct {
	*testEnv
	srv *server.MCPServer
}

func newMCPTestEnv(t *testing.T) *mcpTestEnv {
	t.Helper()
	env := newTestEnv(t)

	a := &app{}
	w, err := a.getWorkspace()
	require.NoError(t, err)

	s := server.NewMCPServer(
		"runectl-test", "0.0.1",
		server.WithRecovery(),
	)
	ctx := context.Background()
	registerSyntaxTools(s, w, ctx)
	registerLSPTools(s, w, ctx)

	return &mcpTestEnv{testEnv: env, srv: s}
}

// callTool sends a tools/call JSON-RPC message and
// returns the result text content.
func (e *mcpTestEnv) callTool(
	t *testing.T,
	name string,
	args map[string]any,
) string {
	t.Helper()
	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      name,
			"arguments": args,
		},
	}
	data, err := json.Marshal(msg)
	require.NoError(t, err)

	resp := e.srv.HandleMessage(
		context.Background(), data,
	)
	jsonResp, ok := resp.(mcp.JSONRPCResponse)
	require.True(t, ok,
		"expected JSONRPCResponse, got %T: %+v",
		resp, resp,
	)

	resultJSON, err := json.Marshal(jsonResp.Result)
	require.NoError(t, err)

	var toolResult mcp.CallToolResult
	err = json.Unmarshal(resultJSON, &toolResult)
	require.NoError(t, err)
	require.False(t, toolResult.IsError,
		"tool returned error: %+v",
		toolResult.Content,
	)
	require.NotEmpty(t, toolResult.Content)

	text, ok := toolResult.Content[0].(mcp.TextContent)
	require.True(t, ok,
		"expected TextContent, got %T",
		toolResult.Content[0],
	)
	return text.Text
}

func TestMCPSyntax(t *testing.T) {
	tests := []struct {
		name   string
		tool   string
		args   map[string]any
		check  func(*testing.T, string)
	}{
		{
			name: "search",
			tool: "syntax_search",
			args: map[string]any{
				"query": "(function_declaration) @fn",
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out,
					"file:///src/main.go",
				)
				require.Contains(t, out,
					"func main()",
				)
			},
		},
		{
			name: "search_node",
			tool: "syntax_search_node",
			args: map[string]any{
				"node_types": "func",
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "MyFunc")
			},
		},
		{
			name: "query",
			tool: "syntax_query",
			args: map[string]any{
				"uri":   "file:///test.go",
				"query": "(package_clause) @pkg",
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out,
					"package main",
				)
			},
		},
		{
			name: "query_node",
			tool: "syntax_query_node",
			args: map[string]any{
				"uri":        "file:///test.go",
				"node_types": "namespace",
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "main")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newMCPTestEnv(t)
			defer env.cleanup()

			out := env.callTool(
				t, tt.tool, tt.args,
			)
			tt.check(t, out)
		})
	}
}

func TestMCPLSP(t *testing.T) {
	tests := []struct {
		name   string
		tool   string
		args   map[string]any
		check  func(*testing.T, string)
	}{
		{
			name: "hover",
			tool: "lsp_hover",
			args: map[string]any{
				"uri":       "file:///src/main.go",
				"line":      0,
				"character": 5,
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out,
					"func main()",
				)
			},
		},
		{
			name: "definition",
			tool: "lsp_definition",
			args: map[string]any{
				"uri":       "file:///src/main.go",
				"line":      0,
				"character": 5,
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out,
					"file:///src/main.go",
				)
				require.Contains(t, out, `"start_line":10`)
			},
		},
		{
			name: "references",
			tool: "lsp_references",
			args: map[string]any{
				"uri":       "file:///src/main.go",
				"line":      0,
				"character": 5,
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out,
					"file:///src/main.go",
				)
				require.Contains(t, out,
					"file:///src/util.go",
				)
			},
		},
		{
			name: "completion",
			tool: "lsp_completion",
			args: map[string]any{
				"uri":       "file:///src/main.go",
				"line":      0,
				"character": 5,
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "fmt")
				require.Contains(t, out, "Println")
			},
		},
		{
			name: "doc_symbols",
			tool: "lsp_document_symbols",
			args: map[string]any{
				"uri": "file:///src/main.go",
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "main")
				require.Contains(t, out, "Function")
			},
		},
		{
			name: "ws_symbols",
			tool: "lsp_workspace_symbols",
			args: map[string]any{
				"query": "My",
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "MyFunc")
			},
		},
		{
			name: "diagnostics",
			tool: "lsp_diagnostics",
			args: map[string]any{
				"uri": "file:///src/main.go",
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out,
					"undefined variable",
				)
				require.Contains(t, out, "Error")
			},
		},
		{
			name: "ws_diags",
			tool: "lsp_workspace_diagnostics",
			args: map[string]any{},
			check: func(t *testing.T, out string) {
				require.Contains(t, out,
					"undefined variable",
				)
				require.Contains(t, out,
					"unused import",
				)
				require.Contains(t, out, "Error")
				require.Contains(t, out, "Warning")
				require.Contains(t, out,
					"file:///src/main.go",
				)
				require.Contains(t, out,
					"file:///src/util.go",
				)
			},
		},
		{
			name: "rename",
			tool: "lsp_rename",
			args: map[string]any{
				"uri":       "file:///src/main.go",
				"line":      5,
				"character": 5,
				"new_name":  "newMain",
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out,
					"file:///src/main.go",
				)
			},
		},
		{
			name: "code_actions",
			tool: "lsp_code_actions",
			args: map[string]any{
				"uri":       "file:///src/main.go",
				"line":      0,
				"character": 0,
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out,
					"Extract variable",
				)
				require.Contains(t, out,
					"Organize imports",
				)
			},
		},
		{
			name: "formatting",
			tool: "lsp_formatting",
			args: map[string]any{
				"uri": "file:///src/main.go",
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "package")
			},
		},
		{
			name: "declaration",
			tool: "lsp_declaration",
			args: map[string]any{
				"uri":       "file:///src/main.go",
				"line":      0,
				"character": 5,
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out,
					"file:///src/types.go",
				)
			},
		},
		{
			name: "type_definition",
			tool: "lsp_type_definition",
			args: map[string]any{
				"uri":       "file:///src/main.go",
				"line":      0,
				"character": 5,
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out,
					"file:///src/types.go",
				)
			},
		},
		{
			name: "implementation",
			tool: "lsp_implementation",
			args: map[string]any{
				"uri":       "file:///src/main.go",
				"line":      0,
				"character": 5,
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out,
					"file:///src/impl1.go",
				)
				require.Contains(t, out,
					"file:///src/impl2.go",
				)
			},
		},
		{
			name: "sig_help",
			tool: "lsp_signature_help",
			args: map[string]any{
				"uri":       "file:///src/main.go",
				"line":      0,
				"character": 5,
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "Println")
			},
		},
		{
			name: "doc_highlight",
			tool: "lsp_document_highlight",
			args: map[string]any{
				"uri":       "file:///src/main.go",
				"line":      5,
				"character": 5,
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "Read")
				require.Contains(t, out, "Write")
			},
		},
		{
			name: "code_lens",
			tool: "lsp_code_lens",
			args: map[string]any{
				"uri": "file:///src/main.go",
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "Run test")
				require.Contains(t, out, "test.run")
			},
		},
		{
			name: "folding_range",
			tool: "lsp_folding_range",
			args: map[string]any{
				"uri": "file:///src/main.go",
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "region")
				require.Contains(t, out, "imports")
			},
		},
		{
			name: "prep_call_hier",
			tool: "lsp_prepare_call_hierarchy",
			args: map[string]any{
				"uri":       "file:///src/main.go",
				"line":      5,
				"character": 5,
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "main")
				require.Contains(t, out, "Function")
			},
		},
		{
			name: "prep_type_hier",
			tool: "lsp_prepare_type_hierarchy",
			args: map[string]any{
				"uri":       "file:///src/main.go",
				"line":      15,
				"character": 5,
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "Reader")
				require.Contains(t, out, "Interface")
			},
		},
		{
			name: "execute_command",
			tool: "lsp_execute_command",
			args: map[string]any{
				"command": "test.run",
			},
			check: func(t *testing.T, out string) {
				require.Contains(t, out,
					"command executed: test.run",
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newMCPTestEnv(t)
			defer env.cleanup()

			out := env.callTool(
				t, tt.tool, tt.args,
			)
			if tt.check != nil {
				tt.check(t, out)
			}
		})
	}
}

func TestMCPToolError(t *testing.T) {
	env := newMCPTestEnv(t)
	defer env.cleanup()

	// Call a tool with missing required parameter.
	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "lsp_hover",
			"arguments": map[string]any{},
		},
	}
	data, err := json.Marshal(msg)
	require.NoError(t, err)

	resp := env.srv.HandleMessage(
		context.Background(), data,
	)
	jsonResp, ok := resp.(mcp.JSONRPCResponse)
	require.True(t, ok,
		"expected JSONRPCResponse, got %T",
		resp,
	)

	resultJSON, err := json.Marshal(jsonResp.Result)
	require.NoError(t, err)

	var toolResult mcp.CallToolResult
	err = json.Unmarshal(resultJSON, &toolResult)
	require.NoError(t, err)
	require.True(t, toolResult.IsError,
		"expected tool error for missing params",
	)
}

func TestMCPUnknownTool(t *testing.T) {
	env := newMCPTestEnv(t)
	defer env.cleanup()

	msg := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "nonexistent_tool",
			"arguments": {}
		}
	}`

	resp := env.srv.HandleMessage(
		context.Background(), []byte(msg),
	)
	// Should return an error response.
	_, isErr := resp.(mcp.JSONRPCError)
	_, isResp := resp.(mcp.JSONRPCResponse)
	require.True(t, isErr || isResp,
		"expected error or response, got %T",
		resp,
	)
}
