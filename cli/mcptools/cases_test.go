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

import (
	"context"
	"encoding/json"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/require"
)

// coreTools are registered without options.
var coreTools = []string{
	"lsp_definition",
	"lsp_diagnostics",
	"lsp_document_symbols",
	"lsp_documentation",
	"lsp_implementation",
	"lsp_references",
	"lsp_workspace_symbols",
	"syntax_query",
	"syntax_query_node",
	"syntax_search",
	"syntax_search_node",
}

// themeTools lists each theme's tools. It is written out rather than derived from
// toolThemes so that a tool moved to the wrong theme fails the tests.
var themeTools = map[Theme][]string{
	ThemeRename:      {"lsp_prepare_rename", "lsp_rename"},
	ThemeCodeActions: {"lsp_code_actions", "lsp_execute_command"},
	ThemeFormatting: {"lsp_formatting", "lsp_on_type_formatting",
		"lsp_range_formatting"},
	ThemeCompletion: {"lsp_completion", "lsp_signature_help"},
	ThemeCallHierarchy: {"lsp_incoming_calls", "lsp_outgoing_calls",
		"lsp_prepare_call_hierarchy"},
	ThemeTypeHierarchy: {"lsp_prepare_type_hierarchy", "lsp_type_subtypes",
		"lsp_type_supertypes"},
	ThemeTypeInfo:             {"lsp_declaration", "lsp_type_definition"},
	ThemeWorkspaceDiagnostics: {"lsp_workspace_diagnostics"},
	ThemeDocumentInfo: {"lsp_code_lens", "lsp_document_highlight",
		"lsp_folding_range", "lsp_selection_range"},
}

// allTools are registered with WithAllThemes, sorted.
func allTools() []string {
	all := slices.Clone(coreTools)
	for _, tools := range themeTools {
		all = append(all, tools...)
	}
	slices.Sort(all)
	return all
}

// toolCase is a valid call of one tool and its exact result against fakeLSP and
// fakeParser.
type toolCase struct {
	tool string
	args map[string]any
	// required are the arguments the schema requires.
	required []string
	// method is the backend method the tool answers from.
	method string
	want   string
}

const (
	mainGo  = "file:///src/main.go"
	typesGo = "file:///src/types.go"
)

var (
	symbolArg = []string{"symbol"}
	posArgs   = []string{"character", "line", "uri"}
	docArgs   = []string{"uri"}
	itemArgs  = []string{
		"end_character", "end_line", "kind", "name",
		"start_character", "start_line", "uri",
	}
)

func at(uri string, line, char int) map[string]any {
	return map[string]any{"uri": uri, "line": line, "character": char}
}

func item(name, kind, uri string, sl, sc, el, ec int) map[string]any {
	return map[string]any{
		"name": name, "kind": kind, "uri": uri,
		"start_line": sl, "start_character": sc,
		"end_line": el, "end_character": ec,
	}
}

// toolCases returns a case for every tool, sorted by tool.
func toolCases() []toolCase {
	cases := []toolCase{
		{
			tool: "lsp_documentation", args: map[string]any{"symbol": "MyFunc"},
			required: symbolArg, method: "Hover",
			want: `{"contents":"func MyFunc()","kind":"markdown"}`,
		},
		{
			tool: "lsp_definition", args: map[string]any{"symbol": "MyFunc"},
			required: symbolArg, method: "Definition",
			want: `[{"uri":"file:///src/main.go","start_line":5,"start_char":0,` +
				`"end_line":5,"end_char":6}]`,
		},
		{
			tool: "lsp_declaration", args: map[string]any{"symbol": "MyFunc"},
			required: symbolArg, method: "Declaration",
			want: `[{"uri":"file:///src/decl.go","start_line":7,"start_char":1,` +
				`"end_line":7,"end_char":9}]`,
		},
		{
			tool: "lsp_type_definition", args: map[string]any{"symbol": "myVar"},
			required: symbolArg, method: "TypeDefinition",
			want: `[{"uri":"file:///src/types.go","start_line":20,"start_char":5,` +
				`"end_line":25,"end_char":1}]`,
		},
		{
			tool: "lsp_implementation", args: map[string]any{"symbol": "MyInterface"},
			required: symbolArg, method: "Implementation",
			want: `[{"uri":"file:///src/impl1.go","start_line":10,"start_char":0,` +
				`"end_line":10,"end_char":10},` +
				`{"uri":"file:///src/impl2.go","start_line":25,"start_char":0,` +
				`"end_line":25,"end_char":10}]`,
		},
		{
			tool: "lsp_references", args: map[string]any{"symbol": "MyFunc",
				"include_declaration": true},
			required: symbolArg, method: "References",
			want: `[{"uri":"file:///src/main.go","start_line":5,"start_char":0,` +
				`"end_line":5,"end_char":6},` +
				`{"uri":"file:///src/util.go","start_line":20,"start_char":3,` +
				`"end_line":20,"end_char":13}]`,
		},
		{
			tool: "lsp_completion", args: at(mainGo, 3, 7),
			required: posArgs, method: "Completion",
			want: `[{"label":"fmt","kind":"Module","detail":"package fmt"},` +
				`{"label":"Println","kind":"Function","detail":"func Println(a ...any)"}]`,
		},
		{
			tool: "lsp_signature_help", args: at(mainGo, 3, 7),
			required: posArgs, method: "SignatureHelp",
			want: `{"label":"func Printf(format string, a ...any)","active_signature":1,` +
				`"active_parameter":1,"parameters":"format string, a ...any"}`,
		},
		{
			tool: "lsp_rename", args: map[string]any{"symbol": "MyFunc",
				"new_name": "Renamed"},
			required: []string{"new_name", "symbol"}, method: "Rename",
			want: `{"file:///src/main.go":2,"file:///src/util.go":1}`,
		},
		{
			tool: "lsp_prepare_rename", args: map[string]any{"symbol": "MyFunc"},
			required: symbolArg, method: "PrepareRename",
			want: `{"start_line":5,"start_char":0,"end_line":5,"end_char":6,` +
				`"placeholder":"MyFunc"}`,
		},
		{
			tool: "lsp_code_actions", args: at(mainGo, 3, 7),
			required: posArgs, method: "CodeAction",
			want: `[{"title":"Extract variable","kind":"refactor.extract"},` +
				`{"title":"Run test","kind":"command"}]`,
		},
		{
			tool: "lsp_formatting", args: map[string]any{"uri": mainGo},
			required: docArgs, method: "Formatting",
			want: `[{"start_line":0,"start_char":0,"end_line":0,"end_char":4,` +
				`"new_text":"\t"}]`,
		},
		{
			tool: "lsp_range_formatting",
			args: map[string]any{
				"uri": mainGo, "start_line": 2, "start_character": 1,
				"end_line": 4, "end_character": 0,
			},
			required: []string{"end_character", "end_line", "start_character",
				"start_line", "uri"},
			method: "RangeFormatting",
			want: `[{"start_line":2,"start_char":1,"end_line":4,"end_char":0,` +
				`"new_text":"fixed"}]`,
		},
		{
			tool:     "lsp_on_type_formatting",
			args:     map[string]any{"uri": mainGo, "line": 3, "character": 7, "ch": "}"},
			required: []string{"ch", "character", "line", "uri"}, method: "OnTypeFormatting",
			want: `[{"start_line":3,"start_char":7,"end_line":3,"end_char":7,` +
				`"new_text":"}\n"}]`,
		},
		{
			tool: "lsp_document_symbols", args: map[string]any{"uri": mainGo},
			required: docArgs, method: "DocumentSymbol",
			want: `[{"name":"main","detail":"func()","kind":"Function","depth":0,` +
				`"start_line":5,"start_char":0,"end_line":10,"end_char":1},` +
				`{"name":"x","detail":"","kind":"Variable","depth":1,` +
				`"start_line":6,"start_char":1,"end_line":6,"end_char":10}]`,
		},
		{
			tool: "lsp_workspace_symbols", args: map[string]any{"query": "MyStruct"},
			required: []string{"query"}, method: "WorkspaceSymbol",
			want: `[{"name":"MyStruct","kind":"Struct","uri":"file:///src/types.go",` +
				`"start_line":20,"start_char":5},` +
				`{"name":"MyStruct.MyMethod","kind":"Method",` +
				`"uri":"file:///src/types.go","start_line":30,"start_char":0}]`,
		},
		{
			tool: "lsp_diagnostics", args: map[string]any{"uri": mainGo},
			required: docArgs, method: "Diagnostic",
			want: `[{"severity":"Error","start_line":3,"start_char":10,"end_line":3,` +
				`"end_char":15,` +
				`"message":"undefined: x","source":"compiler","code":"E001"},` +
				`{"severity":"Warning","start_line":4,"start_char":0,"end_line":4,` +
				`"end_char":2,` +
				`"message":"Cannot find name 'y'.","source":"ts","code":"2304"}]`,
		},
		{
			tool: "lsp_workspace_diagnostics", args: map[string]any{},
			method: "WorkspaceDiagnostic",
			want: `[{"uri":"file:///src/main.go","severity":"Error","start_line":3,` +
				`"start_char":10,` +
				`"end_line":3,"end_char":15,"message":"undefined: x",` +
				`"source":"compiler","code":"E001"},` +
				`{"uri":"file:///src/util.go","severity":"Hint","start_line":7,` +
				`"start_char":0,` +
				`"end_line":7,"end_char":5,"message":"simplify","source":"","code":""},` +
				`{"uri":"file:///src/util.go","severity":"Information","start_line":9,` +
				`"start_char":0,` +
				`"end_line":9,"end_char":1,"message":"note","source":"","code":"7"}]`,
		},
		{
			tool: "lsp_document_highlight", args: at(mainGo, 3, 7),
			required: posArgs, method: "DocumentHighlight",
			want: `[{"kind":"Read","start_line":5,"start_char":5,"end_line":5,` +
				`"end_char":9},` +
				`{"kind":"Write","start_line":10,"start_char":3,"end_line":10,` +
				`"end_char":7},` +
				`{"kind":"Text","start_line":12,"start_char":0,"end_line":12,"end_char":4}]`,
		},
		{
			tool: "lsp_code_lens", args: map[string]any{"uri": mainGo},
			required: docArgs, method: "CodeLens",
			want: `[{"start_line":5,"start_char":0,"end_line":5,"end_char":10,` +
				`"title":"Run test","command":"test.run"},` +
				`{"start_line":7,"start_char":0,"end_line":7,"end_char":3,"title":"",` +
				`"command":""}]`,
		},
		{
			tool: "lsp_folding_range", args: map[string]any{"uri": mainGo},
			required: docArgs, method: "FoldingRange",
			want: `[{"start_line":5,"start_char":0,"end_line":20,"end_char":1,` +
				`"kind":"region"},` +
				`{"start_line":1,"start_char":0,"end_line":3,"end_char":0,` +
				`"kind":"imports"},` +
				`{"start_line":22,"start_char":2,"end_line":24,"end_char":0,` +
				`"kind":"comment"},` +
				`{"start_line":30,"start_char":0,"end_line":31,"end_char":0,"kind":""}]`,
		},
		{
			tool: "lsp_selection_range", args: at(mainGo, 3, 7),
			required: posArgs, method: "SelectionRange",
			want: `[{"start_line":3,"start_char":7,"end_line":3,"end_char":11,"depth":0},` +
				`{"start_line":3,"start_char":0,"end_line":8,"end_char":1,"depth":1},` +
				`{"start_line":0,"start_char":0,"end_line":100,"end_char":0,"depth":2}]`,
		},
		{
			tool: "lsp_prepare_call_hierarchy", args: map[string]any{"symbol": "MyFunc"},
			required: symbolArg, method: "PrepareCallHierarchy",
			want: `[{"name":"MyFunc","kind":"Function","uri":"file:///src/main.go",` +
				`"start_line":5,"start_char":0,"end_line":10,"end_char":1}]`,
		},
		{
			tool:     "lsp_incoming_calls",
			args:     item("MyFunc", "Function", mainGo, 5, 0, 10, 1),
			required: itemArgs, method: "CallHierarchyIncomingCalls",
			want: `[{"from_name":"runApp","from_kind":"Function",` +
				`"from_uri":"file:///src/app.go",` +
				`"from_start_line":20,"from_start_char":0}]`,
		},
		{
			tool:     "lsp_outgoing_calls",
			args:     item("MyFunc", "Function", mainGo, 5, 0, 10, 1),
			required: itemArgs, method: "CallHierarchyOutgoingCalls",
			want: `[{"to_name":"Println","to_kind":"Function",` +
				`"to_uri":"file:///go/src/fmt/print.go",` +
				`"to_start_line":100,"to_start_char":0}]`,
		},
		{
			tool: "lsp_prepare_type_hierarchy", args: map[string]any{
				"symbol": "MyInterface"},
			required: symbolArg, method: "PrepareTypeHierarchy",
			want: `[{"name":"MyInterface","kind":"Interface",` +
				`"uri":"file:///src/types.go","detail":"interface",` +
				`"start_line":15,"start_char":0,"end_line":18,"end_char":1}]`,
		},
		{
			tool:     "lsp_type_supertypes",
			args:     item("MyStruct", "Struct", typesGo, 20, 5, 25, 1),
			required: itemArgs, method: "TypeHierarchySupertypes",
			want: `[{"name":"io.Reader","kind":"Interface",` +
				`"uri":"file:///go/src/io/io.go","detail":"interface",` +
				`"start_line":50,"start_char":0,"end_line":53,"end_char":1}]`,
		},
		{
			tool:     "lsp_type_subtypes",
			args:     item("MyInterface", "Interface", typesGo, 15, 5, 18, 1),
			required: itemArgs, method: "TypeHierarchySubtypes",
			want: `[{"name":"MyReader","kind":"Struct","uri":"file:///src/reader.go",` +
				`"detail":"struct",` +
				`"start_line":10,"start_char":0,"end_line":15,"end_char":1},` +
				`{"name":"BufferedReader","kind":"Struct",` +
				`"uri":"file:///src/buffered.go","detail":"",` +
				`"start_line":8,"start_char":0,"end_line":12,"end_char":1}]`,
		},
		{
			tool: "lsp_execute_command", args: map[string]any{"command": "test.run"},
			required: []string{"command"}, method: "ExecuteCommand",
			want: `{"result":"ran test.run"}`,
		},
		{
			tool: "syntax_search",
			args: map[string]any{
				"query":     "(function_declaration) @fn",
				"captures":  []any{"fn"},
				"languages": []any{"go"},
			},
			required: []string{"query"}, method: "Search",
			want: `[{"file":"file:///src/main.go","text":"func main()","from_x":0,` +
				`"from_y":5,"to_x":11,"to_y":5,"capture_name":"fn"},` +
				`{"file":"file:///src/main.go","text":"func main()","from_x":2,` +
				`"from_y":9,"to_x":13,"to_y":9,"capture_name":"fn"}]`,
		},
		{
			tool: "syntax_search_node", args: map[string]any{"node_types": "func|method"},
			required: []string{"node_types"}, method: "SearchNode",
			want: `[{"file":"file:///src/main.go","text":"MyFunc","from_x":0,` +
				`"from_y":5,"to_x":6,"to_y":5,"capture_name":"definition.func"},` +
				`{"file":"file:///src/main.go","text":"MyFunc","from_x":2,"from_y":9,` +
				`"to_x":8,"to_y":9,"capture_name":"definition.func"}]`,
		},
		{
			tool: "syntax_query", args: map[string]any{"uri": mainGo,
				"query": "(package_clause) @pkg"},
			required: []string{"query", "uri"}, method: "Query",
			want: `[{"file":"file:///src/main.go","text":"package main","from_x":0,` +
				`"from_y":5,"to_x":12,"to_y":5,"capture_name":"pkg"},` +
				`{"file":"file:///src/main.go","text":"package main","from_x":2,` +
				`"from_y":9,"to_x":14,"to_y":9,"capture_name":"pkg"}]`,
		},
		{
			tool: "syntax_query_node", args: map[string]any{"uri": mainGo,
				"node_types": "namespace"},
			required: []string{"node_types", "uri"}, method: "QueryNode",
			want: `[{"file":"file:///src/main.go","text":"main","from_x":0,"from_y":5,` +
				`"to_x":4,"to_y":5,"capture_name":"definition.namespace"},` +
				`{"file":"file:///src/main.go","text":"main","from_x":2,"from_y":9,` +
				`"to_x":6,"to_y":9,"capture_name":"definition.namespace"}]`,
		},
	}
	slices.SortFunc(cases, func(a, b toolCase) int {
		return strings.Compare(a.tool, b.tool)
	})
	return cases
}

// isSyntaxTool reports whether tool is answered by the parser rather than the language
// server.
func isSyntaxTool(tool string) bool {
	return strings.HasPrefix(tool, "syntax_")
}

// without returns a copy of args without key.
func without(args map[string]any, key string) map[string]any {
	out := maps.Clone(args)
	delete(out, key)
	return out
}

// testServer is an MCP server with every tool registered over the fakes it holds.
type testServer struct {
	*server.MCPServer
	lsp    *fakeLSP
	parser *fakeParser
}

func newTestServer(opts ...Option) *testServer {
	return newTestServerContext(context.Background(), opts...)
}

func newTestServerContext(ctx context.Context, opts ...Option) *testServer {
	s := &testServer{
		MCPServer: server.NewMCPServer("mcptools-test", "1.2.3"),
		lsp:       newFakeLSP(),
		parser:    &fakeParser{},
	}
	RegisterSyntaxTools(ctx, s.MCPServer, s.parser, opts...)
	RegisterLSPTools(ctx, s.MCPServer, s.lsp, opts...)
	return s
}

// request sends one JSON-RPC request to s and returns its response message.
func request(
	t *testing.T,
	s *server.MCPServer,
	method string,
	params any,
) mcp.JSONRPCMessage {
	t.Helper()
	data, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": method, "params": params,
	})
	require.NoError(t, err)
	return s.HandleMessage(context.Background(), data)
}

// handle sends one JSON-RPC request to s and returns its result, failing on a JSON-RPC
// error.
func handle(
	t *testing.T,
	s *server.MCPServer,
	method string,
	params any,
) json.RawMessage {
	t.Helper()
	resp := request(t, s, method, params)
	jr, ok := resp.(mcp.JSONRPCResponse)
	require.True(t, ok, "expected JSONRPCResponse, got %T: %+v", resp, resp)
	out, err := json.Marshal(jr.Result)
	require.NoError(t, err)
	return out
}

// listedTool is a tool as tools/list returns it.
type listedTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Meta        map[string]any `json:"_meta"`
	InputSchema struct {
		Type       string                    `json:"type"`
		Properties map[string]map[string]any `json:"properties"`
		Required   []string                  `json:"required"`
	} `json:"inputSchema"`
}

func listTools(t *testing.T, s *server.MCPServer) map[string]listedTool {
	t.Helper()
	var res struct {
		Tools []listedTool `json:"tools"`
	}
	require.NoError(t, json.Unmarshal(handle(t, s, "tools/list", map[string]any{}), &res))
	out := make(map[string]listedTool, len(res.Tools))
	for _, tl := range res.Tools {
		require.NotContains(t, out, tl.Name, "tool listed twice")
		out[tl.Name] = tl
	}
	return out
}

func toolNames(tools map[string]listedTool) []string {
	return slices.Sorted(maps.Keys(tools))
}

func callTool(
	t *testing.T,
	s *server.MCPServer,
	name string,
	args map[string]any,
) mcp.CallToolResult {
	t.Helper()
	var res mcp.CallToolResult
	require.NoError(t, json.Unmarshal(handle(t, s, "tools/call", map[string]any{
		"name": name, "arguments": args,
	}), &res))
	return res
}

// resultText returns the single text block of res.
func resultText(t *testing.T, res mcp.CallToolResult) string {
	t.Helper()
	require.Len(t, res.Content, 1, "%+v", res.Content)
	switch c := res.Content[0].(type) {
	case mcp.TextContent:
		return c.Text
	case *mcp.TextContent:
		return c.Text
	}
	require.Failf(t, "not text", "got %T", res.Content[0])
	return ""
}
