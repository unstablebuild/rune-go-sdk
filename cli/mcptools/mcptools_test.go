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
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi"
)

func TestCoreTools(t *testing.T) {
	s := newTestServer()
	require.Equal(t, coreTools, toolNames(listTools(t, s.MCPServer)))
}

func TestThemes(t *testing.T) {
	require.Len(t, themeTools, 9, "every Theme constant is listed")
	for theme, tools := range themeTools {
		t.Run(string(theme), func(t *testing.T) {
			s := newTestServer(WithThemes(theme))
			want := slices.Concat(coreTools, tools)
			slices.Sort(want)
			require.Equal(t, want, toolNames(listTools(t, s.MCPServer)))
		})
	}
}

func TestThemesCombine(t *testing.T) {
	s := newTestServer(WithThemes(ThemeRename), WithThemes(ThemeFormatting, ThemeRename))
	want := slices.Concat(coreTools, themeTools[ThemeRename], themeTools[ThemeFormatting])
	slices.Sort(want)
	require.Equal(t, want, toolNames(listTools(t, s.MCPServer)))
}

func TestAllThemes(t *testing.T) {
	s := newTestServer(WithAllThemes())
	names := toolNames(listTools(t, s.MCPServer))
	require.Equal(t, allTools(), names)
	require.Len(t, names, 33)
}

func TestUnknownThemeAddsNothing(t *testing.T) {
	s := newTestServer(WithThemes("no-such-theme"))
	require.Equal(t, coreTools, toolNames(listTools(t, s.MCPServer)))
}

// toolThemes decides registration; a key that names no tool would silently leave a tool
// in the core set.
func TestToolThemesNameRegisteredTools(t *testing.T) {
	var themed []string
	for _, tools := range themeTools {
		themed = append(themed, tools...)
	}
	slices.Sort(themed)
	got := make([]string, 0, len(toolThemes))
	for tool, theme := range toolThemes {
		got = append(got, tool)
		require.Contains(t, themeTools[theme], tool)
	}
	slices.Sort(got)
	require.Equal(t, themed, got)
}

func TestRegisterSeparately(t *testing.T) {
	t.Run("syntax", func(t *testing.T) {
		s := server.NewMCPServer("t", "1")
		RegisterSyntaxTools(context.Background(), s, &fakeParser{}, WithAllThemes())
		require.Equal(t,
			[]string{"syntax_query", "syntax_query_node", "syntax_search",
				"syntax_search_node"},
			toolNames(listTools(t, s)))
	})
	t.Run("lsp", func(t *testing.T) {
		s := server.NewMCPServer("t", "1")
		RegisterLSPTools(context.Background(), s, newFakeLSP())
		require.Equal(t,
			[]string{
				"lsp_definition", "lsp_diagnostics", "lsp_document_symbols",
				"lsp_documentation", "lsp_implementation", "lsp_references",
				"lsp_workspace_symbols",
			},
			toolNames(listTools(t, s)))
	})
	t.Run("options are per call", func(t *testing.T) {
		s := server.NewMCPServer("t", "1")
		RegisterSyntaxTools(context.Background(), s, &fakeParser{},
			WithAlwaysLoad("syntax_search"))
		RegisterLSPTools(context.Background(), s, newFakeLSP(), WithThemes(ThemeRename))
		tools := listTools(t, s)
		want := slices.Concat(coreTools, themeTools[ThemeRename])
		slices.Sort(want)
		require.Equal(t, want, toolNames(tools))
		require.Equal(t, true, tools["syntax_search"].Meta[AlwaysLoadMetaKey])
		// The LSP call kept the default always-load set.
		require.Equal(t, true, tools["lsp_definition"].Meta[AlwaysLoadMetaKey])
		require.Nil(t, tools["syntax_query"].Meta)
	})
}

func TestAlwaysLoad(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
		want []string
	}{
		{
			name: "default",
			opts: []Option{WithAllThemes()},
			want: []string{
				"lsp_definition", "lsp_diagnostics", "lsp_document_symbols",
				"lsp_implementation", "lsp_references", "lsp_workspace_symbols",
			},
		},
		{name: "none", opts: []Option{WithAlwaysLoad()}},
		{
			name: "unregistered tools are ignored",
			opts: []Option{WithAlwaysLoad("syntax_search", "lsp_rename", "no_such_tool")},
			want: []string{"syntax_search"},
		},
		{
			name: "theme tool",
			opts: []Option{WithAlwaysLoad("lsp_rename"), WithThemes(ThemeRename)},
			want: []string{"lsp_rename"},
		},
		{
			name: "last option wins",
			opts: []Option{WithAlwaysLoad("syntax_search"),
				WithAlwaysLoad("syntax_query")},
			want: []string{"syntax_query"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(tt.opts...)
			var got []string
			for name, tl := range listTools(t, s.MCPServer) {
				switch tl.Meta[AlwaysLoadMetaKey] {
				case true:
					got = append(got, name)
				case nil:
					require.Nil(t, tl.Meta, "%s carries other _meta", name)
				default:
					require.Failf(t, "bad meta", "%s: %v", name, tl.Meta)
				}
			}
			slices.Sort(got)
			require.Equal(t, tt.want, got)
		})
	}
}

// The tables in cases_test.go must cover every tool, or the tests that range over them
// would miss one.
func TestCasesCoverEveryTool(t *testing.T) {
	var names []string
	for _, c := range toolCases() {
		names = append(names, c.tool)
	}
	require.Equal(t, allTools(), names)
	require.Equal(t, allTools(),
		toolNames(listTools(t, newTestServer(WithAllThemes()).MCPServer)))
}

// paramTypes is the JSON schema type of every tool argument.
var paramTypes = map[string]string{
	"symbol": "string", "new_name": "string", "uri": "string",
	"query": "string", "node_types": "string", "ch": "string",
	"name": "string", "kind": "string", "command": "string",
	"line": "number", "character": "number", "tab_size": "number",
	"start_line": "number", "start_character": "number",
	"end_line": "number", "end_character": "number",
	"include_declaration": "boolean", "insert_spaces": "boolean",
	"captures": "array", "languages": "array",
}

func TestToolSchemas(t *testing.T) {
	tools := listTools(t, newTestServer(WithAllThemes()).MCPServer)
	for _, c := range toolCases() {
		t.Run(c.tool, func(t *testing.T) {
			tl := tools[c.tool]
			require.NotEmpty(t, tl.Description)
			require.Equal(t, "object", tl.InputSchema.Type)
			got := slices.Clone(tl.InputSchema.Required)
			slices.Sort(got)
			if len(got) == 0 {
				got = nil
			}
			require.Equal(t, c.required, got)
			for _, name := range c.required {
				require.Contains(t, tl.InputSchema.Properties, name)
			}
			// The valid call uses only arguments the schema lists.
			for name := range c.args {
				require.Contains(t, tl.InputSchema.Properties, name)
			}
			for name, prop := range tl.InputSchema.Properties {
				want, ok := paramTypes[name]
				require.True(t, ok, "unexpected argument %q", name)
				require.Equal(t, want, prop["type"], name)
				if want == "array" {
					// Strict function-calling schemas reject arrays without items.
					require.Equal(t, map[string]any{
						"type": "string"}, prop["items"], name)
				}
			}
		})
	}
}

var toolRef = regexp.MustCompile(`\b(?:lsp|syntax)_[a-z_]+\b`)

// Descriptions steer the model to other tools by name; each name must be one it can call.
func TestDescriptionsReferenceRegisteredTools(t *testing.T) {
	tools := listTools(t, newTestServer(WithAllThemes()).MCPServer)
	for name, tl := range tools {
		for _, ref := range toolRef.FindAllString(tl.Description, -1) {
			require.Contains(t, tools, ref, "%s's description names %s", name, ref)
		}
	}
}

func TestEveryToolCall(t *testing.T) {
	for _, c := range toolCases() {
		t.Run(c.tool, func(t *testing.T) {
			s := newTestServer(WithAllThemes())
			res := callTool(t, s.MCPServer, c.tool, c.args)
			require.False(t, res.IsError, resultText(t, res))
			require.JSONEq(t, c.want, resultText(t, res))
			require.Nil(t, res.StructuredContent)
			backend := &s.lsp.recorder
			if isSyntaxTool(c.tool) {
				backend = &s.parser.recorder
			}
			require.Len(t, backend.called(c.method), 1)
		})
	}
}

func TestMissingRequiredArgument(t *testing.T) {
	for _, c := range toolCases() {
		for _, arg := range c.required {
			t.Run(c.tool+"/"+arg, func(t *testing.T) {
				s := newTestServer(WithAllThemes())
				res := callTool(t, s.MCPServer, c.tool, without(c.args, arg))
				require.True(t, res.IsError)
				require.Contains(t, resultText(t, res), fmt.Sprintf("%q", arg))
				require.Empty(t, s.lsp.methods(), "no backend call")
				require.Empty(t, s.parser.methods(), "no backend call")
			})
		}
	}
}

func TestNoArguments(t *testing.T) {
	s := newTestServer(WithAllThemes())
	for _, c := range toolCases() {
		res := callTool(t, s.MCPServer, c.tool, nil)
		require.Equal(t, len(c.required) > 0, res.IsError, c.tool)
	}
}

func TestPositionArguments(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    uint32
		wantErr string
	}{
		{name: "zero", value: 0, want: 0},
		{name: "whole float", value: 7.0, want: 7},
		{name: "numeric string", value: "7", want: 7},
		{name: "max", value: 4294967295, want: 4294967295},
		{name: "negative", value: -1, wantErr: `"line" must be a whole number`},
		{name: "fraction", value: 1.5, wantErr: `"line" must be a whole number`},
		{name: "too large", value: 4294967296, wantErr: `"line" must be a whole number`},
		{name: "text", value: "seven", wantErr: `"line" cannot be converted`},
		{name: "bool", value: true, wantErr: `"line" is not a float64`},
		{name: "null", value: nil, wantErr: `"line" is not a float64`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(WithAllThemes())
			args := at(mainGo, 0, 2)
			args["line"] = tt.value
			res := callTool(t, s.MCPServer, "lsp_completion", args)
			if tt.wantErr != "" {
				require.True(t, res.IsError)
				require.Contains(t, resultText(t, res), tt.wantErr)
				require.Empty(t, s.lsp.methods())
				return
			}
			require.False(t, res.IsError, resultText(t, res))
			p := s.lsp.called("Completion")[0].params.(semanticapi.CompletionParams)
			require.Equal(t, pos(tt.want, 2), p.Position)
		})
	}
}

func TestRangeArguments(t *testing.T) {
	for _, arg := range []string{"start_line", "start_character", "end_line",
		"end_character"} {
		t.Run(arg, func(t *testing.T) {
			s := newTestServer(WithAllThemes())
			args := item("f", "Function", mainGo, 1, 2, 3, 4)
			args[arg] = -3
			res := callTool(t, s.MCPServer, "lsp_incoming_calls", args)
			require.True(t, res.IsError)
			require.Contains(t, resultText(t, res),
				fmt.Sprintf("%q must be a whole number", arg))
		})
	}
}

func TestFormattingOptions(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		want    semanticapi.FormattingOptions
		wantErr string
	}{
		{name: "defaults", args: map[string]any{},
			want: semanticapi.FormattingOptions{TabSize: 4, InsertSpaces: true}},
		{name: "null tab size", args: map[string]any{
			"tab_size": nil}, want: semanticapi.FormattingOptions{TabSize: 4,
			InsertSpaces: true}},
		{
			name: "explicit",
			args: map[string]any{"tab_size": 2, "insert_spaces": false},
			want: semanticapi.FormattingOptions{TabSize: 2, InsertSpaces: false},
		},
		{name: "negative", args: map[string]any{"tab_size": -2},
			wantErr: `"tab_size" must be a whole number`},
		{name: "fraction", args: map[string]any{"tab_size": 2.5},
			wantErr: `"tab_size" must be a whole number`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(WithAllThemes())
			doc := map[string]any{"uri": mainGo}
			ranged := map[string]any{
				"uri": mainGo, "start_line": 0, "start_character": 0,
				"end_line": 1, "end_character": 0,
			}
			for k, v := range tt.args {
				doc[k], ranged[k] = v, v
			}
			whole := callTool(t, s.MCPServer, "lsp_formatting", doc)
			part := callTool(t, s.MCPServer, "lsp_range_formatting", ranged)
			if tt.wantErr != "" {
				require.Contains(t, resultText(t, whole), tt.wantErr)
				require.Contains(t, resultText(t, part), tt.wantErr)
				require.Empty(t, s.lsp.methods())
				return
			}
			require.False(t, whole.IsError, resultText(t, whole))
			require.False(t, part.IsError, resultText(t, part))
			wp := s.lsp.called(
				"Formatting")[0].params.(semanticapi.DocumentFormattingParams)
			pp := s.lsp.called(
				"RangeFormatting")[0].params.(semanticapi.DocumentRangeFormattingParams)
			require.Equal(t, tt.want, wp.Options)
			require.Equal(t, tt.want, pp.Options)
		})
	}
}

// TestArgumentsReachBackend checks what each tool asks of the language server or parser.
func TestArgumentsReachBackend(t *testing.T) {
	doc := semanticapi.TextDocumentIdentifier{URI: mainGo}
	types := semanticapi.TextDocumentIdentifier{URI: typesGo}
	fn := pos(5, 0)
	iface := pos(15, 5)
	only := func(t *testing.T, r *recorder, method string) any {
		t.Helper()
		calls := r.called(method)
		require.Len(t, calls, 1)
		return calls[0].params
	}
	tests := []struct {
		tool  string
		args  map[string]any
		check func(t *testing.T, s *testServer)
	}{
		{"lsp_documentation", nil, func(t *testing.T, s *testServer) {
			require.Equal(t, semanticapi.WorkspaceSymbolParams{Query: "MyFunc"},
				only(t, &s.lsp.recorder, "WorkspaceSymbol"))
			require.Equal(t, semanticapi.HoverParams{TextDocument: doc, Position: fn},
				only(t, &s.lsp.recorder, "Hover"))
		}},
		{"lsp_definition", nil, func(t *testing.T, s *testServer) {
			require.Equal(t,
				semanticapi.DefinitionParams{TextDocument: doc, Position: fn},
				only(t, &s.lsp.recorder, "Definition"))
		}},
		{"lsp_declaration", nil, func(t *testing.T, s *testServer) {
			require.Equal(t,
				semanticapi.DeclarationParams{TextDocument: doc, Position: fn},
				only(t, &s.lsp.recorder, "Declaration"))
		}},
		{"lsp_type_definition", nil, func(t *testing.T, s *testServer) {
			require.Equal(t,
				semanticapi.TypeDefinitionParams{TextDocument: doc, Position: pos(3, 4)},
				only(t, &s.lsp.recorder, "TypeDefinition"))
		}},
		{"lsp_implementation", nil, func(t *testing.T, s *testServer) {
			require.Equal(t,
				semanticapi.ImplementationParams{TextDocument: types, Position: iface},
				only(t, &s.lsp.recorder, "Implementation"))
		}},
		{"lsp_references", nil, func(t *testing.T, s *testServer) {
			p := only(t, &s.lsp.recorder, "References").(semanticapi.ReferenceParams)
			require.Equal(t, doc, p.TextDocument)
			require.Equal(t, fn, p.Position)
			require.True(t, p.Context.IncludeDeclaration)
		}},
		{"lsp_references", map[string]any{
			"symbol": "MyFunc"}, func(t *testing.T, s *testServer) {
			p := only(t, &s.lsp.recorder, "References").(semanticapi.ReferenceParams)
			require.False(t, p.Context.IncludeDeclaration, "defaults to false")
		}},
		{"lsp_completion", nil, func(t *testing.T, s *testServer) {
			require.Equal(t,
				semanticapi.CompletionParams{TextDocument: doc, Position: pos(3, 7)},
				only(t, &s.lsp.recorder, "Completion"))
		}},
		{"lsp_signature_help", nil, func(t *testing.T, s *testServer) {
			require.Equal(t,
				semanticapi.SignatureHelpParams{TextDocument: doc, Position: pos(3, 7)},
				only(t, &s.lsp.recorder, "SignatureHelp"))
		}},
		{"lsp_rename", nil, func(t *testing.T, s *testServer) {
			require.Equal(t,
				semanticapi.RenameParams{TextDocument: doc, Position: fn,
					NewName: "Renamed"},
				only(t, &s.lsp.recorder, "Rename"))
		}},
		{"lsp_prepare_rename", nil, func(t *testing.T, s *testServer) {
			require.Equal(t,
				semanticapi.PrepareRenameParams{TextDocument: doc, Position: fn},
				only(t, &s.lsp.recorder, "PrepareRename"))
		}},
		{"lsp_code_actions", nil, func(t *testing.T, s *testServer) {
			p := only(t, &s.lsp.recorder, "CodeAction").(semanticapi.CodeActionParams)
			require.Equal(t, doc, p.TextDocument)
			require.Equal(t, rng(3, 7, 3, 7), p.Range)
			require.NotNil(t, p.Context.Diagnostics, "sent as [] rather than null")
			require.Empty(t, p.Context.Diagnostics)
		}},
		{"lsp_range_formatting", nil, func(t *testing.T, s *testServer) {
			p := only(t, &s.lsp.recorder,
				"RangeFormatting").(semanticapi.DocumentRangeFormattingParams)
			require.Equal(t, doc, p.TextDocument)
			require.Equal(t, rng(2, 1, 4, 0), p.Range)
		}},
		{"lsp_on_type_formatting", nil, func(t *testing.T, s *testServer) {
			require.Equal(t, semanticapi.DocumentOnTypeFormattingParams{
				TextDocument: doc, Position: pos(3, 7), Character: "}",
				Options: semanticapi.FormattingOptions{TabSize: 4, InsertSpaces: true},
			}, only(t, &s.lsp.recorder, "OnTypeFormatting"))
		}},
		{"lsp_document_symbols", nil, func(t *testing.T, s *testServer) {
			require.Equal(t, semanticapi.DocumentSymbolParams{TextDocument: doc},
				only(t, &s.lsp.recorder, "DocumentSymbol"))
		}},
		{"lsp_workspace_symbols", nil, func(t *testing.T, s *testServer) {
			require.Equal(t, semanticapi.WorkspaceSymbolParams{Query: "MyStruct"},
				only(t, &s.lsp.recorder, "WorkspaceSymbol"))
		}},
		{"lsp_diagnostics", nil, func(t *testing.T, s *testServer) {
			require.Equal(t, semanticapi.DocumentDiagnosticParams{TextDocument: doc},
				only(t, &s.lsp.recorder, "Diagnostic"))
		}},
		{"lsp_workspace_diagnostics", nil, func(t *testing.T, s *testServer) {
			require.Equal(t, semanticapi.WorkspaceDiagnosticParams{},
				only(t, &s.lsp.recorder, "WorkspaceDiagnostic"))
		}},
		{"lsp_document_highlight", nil, func(t *testing.T, s *testServer) {
			require.Equal(t,
				semanticapi.DocumentHighlightParams{TextDocument: doc,
					Position: pos(3, 7)},
				only(t, &s.lsp.recorder, "DocumentHighlight"))
		}},
		{"lsp_code_lens", nil, func(t *testing.T, s *testServer) {
			require.Equal(t, semanticapi.CodeLensParams{TextDocument: doc},
				only(t, &s.lsp.recorder, "CodeLens"))
		}},
		{"lsp_folding_range", nil, func(t *testing.T, s *testServer) {
			require.Equal(t, semanticapi.FoldingRangeParams{TextDocument: doc},
				only(t, &s.lsp.recorder, "FoldingRange"))
		}},
		{"lsp_selection_range", nil, func(t *testing.T, s *testServer) {
			require.Equal(t, semanticapi.SelectionRangeParams{
				TextDocument: doc, Positions: []semanticapi.Position{pos(3, 7)},
			}, only(t, &s.lsp.recorder, "SelectionRange"))
		}},
		{"lsp_prepare_call_hierarchy", nil, func(t *testing.T, s *testServer) {
			require.Equal(t,
				semanticapi.CallHierarchyPrepareParams{TextDocument: doc, Position: fn},
				only(t, &s.lsp.recorder, "PrepareCallHierarchy"))
		}},
		{"lsp_incoming_calls", nil, func(t *testing.T, s *testServer) {
			item := semanticapi.CallHierarchyItem{
				Name: "MyFunc", Kind: semanticapi.SymbolKindFunction, URI: mainGo,
				Range: rng(5, 0, 10, 1), SelectionRange: rng(5, 0, 10, 1),
			}
			require.Equal(t, semanticapi.CallHierarchyIncomingCallsParams{Item: item},
				only(t, &s.lsp.recorder, "CallHierarchyIncomingCalls"))
		}},
		{"lsp_outgoing_calls", item("run", "method", mainGo, 1, 2, 3, 4), func(
			t *testing.T, s *testServer,
		) {
			item := semanticapi.CallHierarchyItem{
				Name: "run", Kind: semanticapi.SymbolKindMethod, URI: mainGo,
				Range: rng(1, 2, 3, 4), SelectionRange: rng(1, 2, 3, 4),
			}
			require.Equal(t, semanticapi.CallHierarchyOutgoingCallsParams{Item: item},
				only(t, &s.lsp.recorder, "CallHierarchyOutgoingCalls"))
		}},
		{"lsp_prepare_type_hierarchy", nil, func(t *testing.T, s *testServer) {
			require.Equal(t,
				semanticapi.TypeHierarchyPrepareParams{TextDocument: types, Position: iface},
				only(t, &s.lsp.recorder, "PrepareTypeHierarchy"))
		}},
		{"lsp_type_supertypes", nil, func(t *testing.T, s *testServer) {
			item := semanticapi.TypeHierarchyItem{
				Name: "MyStruct", Kind: semanticapi.SymbolKindStruct, URI: typesGo,
				Range: rng(20, 5, 25, 1), SelectionRange: rng(20, 5, 25, 1),
			}
			require.Equal(t, semanticapi.TypeHierarchySupertypesParams{Item: item},
				only(t, &s.lsp.recorder, "TypeHierarchySupertypes"))
		}},
		{"lsp_type_subtypes", nil, func(t *testing.T, s *testServer) {
			item := semanticapi.TypeHierarchyItem{
				Name: "MyInterface", Kind: semanticapi.SymbolKindInterface, URI: typesGo,
				Range: rng(15, 5, 18, 1), SelectionRange: rng(15, 5, 18, 1),
			}
			require.Equal(t, semanticapi.TypeHierarchySubtypesParams{Item: item},
				only(t, &s.lsp.recorder, "TypeHierarchySubtypes"))
		}},
		{"lsp_execute_command", nil, func(t *testing.T, s *testServer) {
			require.Equal(t, semanticapi.ExecuteCommandParams{Command: "test.run"},
				only(t, &s.lsp.recorder, "ExecuteCommand"))
		}},
		{"syntax_search", nil, func(t *testing.T, s *testServer) {
			require.Equal(t, searchParams{
				query: "(function_declaration) " +
					"@fn", captures: []string{"fn"}, languages: []string{"go"},
			}, only(t, &s.parser.recorder, "Search"))
		}},
		{"syntax_search", map[string]any{
			"query": "(x) @x"}, func(t *testing.T, s *testServer) {
			require.Equal(t, searchParams{query: "(x) @x"},
				only(t, &s.parser.recorder, "Search"))
		}},
		{"syntax_search_node", nil, func(t *testing.T, s *testServer) {
			nodeTypes := syntaxapi.NodeCaptureDefinitionFunc |
				syntaxapi.NodeCaptureDefinitionMethod
			require.Equal(t, searchParams{nodeTypes: nodeTypes},
				only(t, &s.parser.recorder, "SearchNode"))
		}},
		{"syntax_query", nil, func(t *testing.T, s *testServer) {
			require.Equal(t, searchParams{query: "(package_clause) @pkg", uri: mainGo},
				only(t, &s.parser.recorder, "Query"))
		}},
		{"syntax_query", map[string]any{"uri": mainGo, "query": "(x) @x",
			"captures": []any{"x"}}, func(t *testing.T, s *testServer) {
			require.Equal(t,
				searchParams{query: "(x) @x", uri: mainGo, captures: []string{"x"}},
				only(t, &s.parser.recorder, "Query"))
		}},
		{"syntax_query_node", nil, func(t *testing.T, s *testServer) {
			nodeTypes := syntaxapi.NodeCaptureDefinitionNamespace
			require.Equal(t, searchParams{nodeTypes: nodeTypes, uri: mainGo},
				only(t, &s.parser.recorder, "QueryNode"))
		}},
	}
	cases := make(map[string]toolCase)
	for _, c := range toolCases() {
		cases[c.tool] = c
	}
	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			args := tt.args
			if args == nil {
				args = cases[tt.tool].args
			}
			s := newTestServer(WithAllThemes())
			res := callTool(t, s.MCPServer, tt.tool, args)
			require.False(t, res.IsError, resultText(t, res))
			tt.check(t, s)
		})
	}
}

func TestBackendErrors(t *testing.T) {
	for _, c := range toolCases() {
		t.Run(c.tool, func(t *testing.T) {
			s := newTestServer(WithAllThemes())
			err := fmt.Errorf("%s: server crashed", c.method)
			s.lsp.failOn(c.method, err)
			s.parser.failOn(c.method, err)
			res := callTool(t, s.MCPServer, c.tool, c.args)
			require.True(t, res.IsError)
			require.Equal(t, err.Error(), resultText(t, res))
		})
	}
}

// symbolTools take a symbol name and resolve it through the workspace's symbols first.
var symbolTools = []string{
	"lsp_declaration", "lsp_definition", "lsp_documentation",
	"lsp_implementation", "lsp_prepare_call_hierarchy",
	"lsp_prepare_rename", "lsp_prepare_type_hierarchy",
	"lsp_references", "lsp_rename", "lsp_type_definition",
}

func TestSymbolToolsResolveFirst(t *testing.T) {
	cases := make(map[string]toolCase)
	for _, c := range toolCases() {
		cases[c.tool] = c
	}
	for _, tool := range symbolTools {
		t.Run(tool, func(t *testing.T) {
			c := cases[tool]
			require.Equal(t, symbolArg, c.required[len(c.required)-1:])

			s := newTestServer(WithAllThemes())
			callTool(t, s.MCPServer, tool, c.args)
			require.Equal(t, []string{"WorkspaceSymbol", c.method}, s.lsp.methods())

			t.Run("not found", func(t *testing.T) {
				s := newTestServer(WithAllThemes())
				res := callTool(t, s.MCPServer, tool,
					map[string]any{"symbol": "Nope", "new_name": "x"})
				require.True(t, res.IsError)
				require.Equal(t, `symbol "Nope" not found`, resultText(t, res))
				require.Equal(t, []string{"WorkspaceSymbol"}, s.lsp.methods())
			})
			t.Run("prefix is not a match", func(t *testing.T) {
				s := newTestServer(WithAllThemes())
				res := callTool(t, s.MCPServer, tool,
					map[string]any{"symbol": "My", "new_name": "x"})
				require.True(t, res.IsError)
				require.Equal(t, `symbol "My" not found`, resultText(t, res))
			})
			t.Run("lookup fails", func(t *testing.T) {
				s := newTestServer(WithAllThemes())
				s.lsp.failOn("WorkspaceSymbol", errors.New("index not ready"))
				res := callTool(t, s.MCPServer, tool, c.args)
				require.True(t, res.IsError)
				require.Equal(t, "index not ready", resultText(t, res))
				require.Empty(t, s.lsp.called(c.method))
			})
		})
	}
}

// A name several symbols share: tools that aggregate ask about each and merge the
// answers; tools that act on one symbol take the best match only.
func TestSymbolSharedByMany(t *testing.T) {
	shared := func() *testServer {
		s := newTestServer(WithAllThemes())
		s.lsp.symbols = []semanticapi.SymbolInformation{
			{Name: "Read", Kind: semanticapi.SymbolKindVariable,
				Location: loc("file:///a.go", 1, 0, 1, 4)},
			{Name: "File.Read", Kind: semanticapi.SymbolKindMethod,
				Location: loc("file:///b.go", 2, 5, 2, 9)},
			{Name: "Reader", Kind: semanticapi.SymbolKindInterface,
				Location: loc("file:///c.go", 3, 0, 3, 6)},
			// The same symbol twice: asked about once.
			{Name: "Read", Kind: semanticapi.SymbolKindFunction,
				Location: loc("file:///d.go", 4, 0, 4, 4)},
			{Name: "Read", Kind: semanticapi.SymbolKindFunction,
				Location: loc("file:///d.go", 4, 0, 4, 4)},
		}
		return s
	}
	positions := func(calls []call) []string {
		var out []string
		for _, c := range calls {
			var td semanticapi.TextDocumentIdentifier
			var p semanticapi.Position
			switch v := c.params.(type) {
			case semanticapi.DefinitionParams:
				td, p = v.TextDocument, v.Position
			case semanticapi.RenameParams:
				td, p = v.TextDocument, v.Position
			}
			out = append(out, fmt.Sprintf("%s:%d:%d", td.URI, p.Line, p.Character))
		}
		return out
	}

	t.Run("definition merges", func(t *testing.T) {
		s := shared()
		res := callTool(t, s.MCPServer, "lsp_definition",
			map[string]any{"symbol": "Read"})
		require.False(t, res.IsError, resultText(t, res))
		// Preferred kinds (method, function) first; the variable last.
		require.Equal(t,
			[]string{"file:///b.go:2:5", "file:///d.go:4:0", "file:///a.go:1:0"},
			positions(s.lsp.called("Definition")))
		require.JSONEq(t, `[`+
			`{"uri":"file:///b.go","start_line":2,"start_char":5,"end_line":2,`+
			`"end_char":11},`+
			`{"uri":"file:///d.go","start_line":4,"start_char":0,"end_line":4,`+
			`"end_char":6},`+
			`{"uri":"file:///a.go","start_line":1,"start_char":0,"end_line":1,"end_char":6}]`,
			resultText(t, res))
	})
	t.Run("references merge without duplicates", func(t *testing.T) {
		s := shared()
		res := callTool(t, s.MCPServer, "lsp_references",
			map[string]any{"symbol": "Read"})
		require.False(t, res.IsError, resultText(t, res))
		require.Len(t, s.lsp.called("References"), 3)
		// Each answer also holds util.go:20:3, kept once.
		require.JSONEq(t, `[`+
			`{"uri":"file:///b.go","start_line":2,"start_char":5,"end_line":2,`+
			`"end_char":11},`+
			`{"uri":"file:///src/util.go","start_line":20,"start_char":3,"end_line":20,`+
			`"end_char":13},`+
			`{"uri":"file:///d.go","start_line":4,"start_char":0,"end_line":4,`+
			`"end_char":6},`+
			`{"uri":"file:///a.go","start_line":1,"start_char":0,"end_line":1,"end_char":6}]`,
			resultText(t, res))
	})
	t.Run("rename takes the best match", func(t *testing.T) {
		s := shared()
		res := callTool(t, s.MCPServer, "lsp_rename",
			map[string]any{"symbol": "Read", "new_name": "Load"})
		require.False(t, res.IsError, resultText(t, res))
		require.Equal(t, []string{"file:///b.go:2:5"}, positions(s.lsp.called("Rename")))
	})
	t.Run("a failure on any match fails the call", func(t *testing.T) {
		s := shared()
		s.lsp.failOn("Definition", errors.New("timeout"))
		res := callTool(t, s.MCPServer, "lsp_definition",
			map[string]any{"symbol": "Read"})
		require.True(t, res.IsError)
		require.Equal(t, "timeout", resultText(t, res))
	})
}

// Empty answers are [] for list results, so a client never has to tell null from an empty
// list, and null where the tool documents an absent answer.
func TestEmptyResults(t *testing.T) {
	absent := map[string]string{
		"lsp_documentation":   `null`,
		"lsp_signature_help":  `null`,
		"lsp_rename":          `null`,
		"lsp_prepare_rename":  `null`,
		"lsp_execute_command": `{"result":""}`,
	}
	for _, c := range toolCases() {
		t.Run(c.tool, func(t *testing.T) {
			s := newTestServer(WithAllThemes())
			s.lsp.empty, s.parser.empty = true, true
			if c.tool == "lsp_workspace_symbols" {
				s.lsp.symbols = nil
			}
			res := callTool(t, s.MCPServer, c.tool, c.args)
			require.False(t, res.IsError, resultText(t, res))
			want, ok := absent[c.tool]
			if !ok {
				want = `[]`
			}
			require.Equal(t, want, resultText(t, res))
		})
	}
}

func TestSignatureHelpActiveSignature(t *testing.T) {
	tests := []struct {
		name   string
		active uint32
		want   string
	}{
		{"first", 0, `{"label":"a(x)","active_signature":0,"active_parameter":0,` +
			`"parameters":"x"}`},
		{"second", 1, `{"label":"b(y, z)","active_signature":1,"active_parameter":0,` +
			`"parameters":"y, z"}`},
		{"out of range is the first", 5, `{"label":"a(x)","active_signature":0,` +
			`"active_parameter":0,"parameters":"x"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lsp := &signatureLSP{help: &semanticapi.SignatureHelp{
				Signatures: []semanticapi.SignatureInformation{
					{Label: "a(x)",
						Parameters: []semanticapi.ParameterInformation{{Label: "x"}}},
					{Label: "b(y, " +
						"z)", Parameters: []semanticapi.ParameterInformation{{Label: "y"},
						{Label: "z"}}},
				},
				ActiveSignature: tt.active,
			}}
			s := server.NewMCPServer("t", "1")
			RegisterLSPTools(context.Background(), s, lsp, WithAllThemes())
			res := callTool(t, s, "lsp_signature_help", at(mainGo, 0, 0))
			require.JSONEq(t, tt.want, resultText(t, res))
		})
	}
	t.Run("no signatures", func(t *testing.T) {
		s := server.NewMCPServer("t", "1")
		RegisterLSPTools(context.Background(), s,
			&signatureLSP{help: &semanticapi.SignatureHelp{}}, WithAllThemes())
		require.Equal(t, `null`,
			resultText(t, callTool(t, s, "lsp_signature_help", at(mainGo, 0, 0))))
	})
}

type signatureLSP struct {
	semanticapi.LSP
	help *semanticapi.SignatureHelp
}

func (l *signatureLSP) SignatureHelp(
	context.Context,
	semanticapi.SignatureHelpParams,
) (*semanticapi.SignatureHelp, error) {
	return l.help, nil
}

func TestSyntaxArguments(t *testing.T) {
	tests := []struct {
		tool    string
		args    map[string]any
		wantErr string
	}{
		{"syntax_search_node", map[string]any{"node_types": "funcs"},
			`invalid node type "funcs"`},
		{"syntax_search_node", map[string]any{"node_types": "func|"},
			`invalid node type ""`},
		{"syntax_search_node", map[string]any{"node_types": ""}, `invalid node type ""`},
		{"syntax_query_node", map[string]any{"uri": mainGo, "node_types": "Func"},
			`invalid node type "Func"`},
		{"syntax_query", map[string]any{"uri": "src/main.go", "query": "(x) @x"},
			"invalid empty scheme"},
		{"syntax_query_node", map[string]any{"uri": "src/main.go", "node_types": "func"},
			"invalid empty scheme"},
	}
	for _, tt := range tests {
		t.Run(tt.tool+"/"+tt.wantErr, func(t *testing.T) {
			s := newTestServer()
			res := callTool(t, s.MCPServer, tt.tool, tt.args)
			require.True(t, res.IsError)
			require.Contains(t, resultText(t, res), tt.wantErr)
			require.Empty(t, s.parser.methods())
		})
	}
}

func TestSyntaxIterators(t *testing.T) {
	for _, c := range toolCases() {
		if !isSyntaxTool(c.tool) {
			continue
		}
		t.Run(c.tool, func(t *testing.T) {
			t.Run("closed after reading", func(t *testing.T) {
				s := newTestServer()
				callTool(t, s.MCPServer, c.tool, c.args)
				require.Equal(t, 1, s.parser.closed)
			})
			t.Run("error while reading", func(t *testing.T) {
				s := newTestServer()
				s.parser.iterErr = errors.New("tree-sitter: parse failed")
				res := callTool(t, s.MCPServer, c.tool, c.args)
				require.True(t, res.IsError)
				// The fake's iterator wraps the error.
				require.Contains(t, resultText(t, res), "tree-sitter: parse failed")
				require.Equal(t, 1, s.parser.closed)
			})
		})
	}
}

type ctxKey struct{}

// Tools call the backends with the context they were registered with, not the request's,
// as RegisterLSPTools and RegisterSyntaxTools document.
func TestRegistrationContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxKey{}, "registered")
	s := newTestServerContext(ctx, WithAllThemes())
	for _, c := range toolCases() {
		res := callTool(t, s.MCPServer, c.tool, c.args)
		require.False(t, res.IsError, "%s: %s", c.tool, resultText(t, res))
	}
	var seen int
	for _, r := range []*recorder{&s.lsp.recorder, &s.parser.recorder} {
		for _, c := range r.calls {
			if c.ctx == nil {
				// Parser methods take no context; their iterators are read with it
				// instead.
				continue
			}
			seen++
			require.Equal(t, "registered", c.ctx.Value(ctxKey{}), c.method)
		}
	}
	require.Positive(t, seen)
	require.NotEmpty(t, s.parser.called("Next"))
}

func TestCanceledRegistrationContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := newTestServerContext(ctx, WithAllThemes())
	cancel()
	res := callTool(t, s.MCPServer, "lsp_diagnostics", map[string]any{"uri": mainGo})
	require.True(t, res.IsError)
	require.Equal(t, context.Canceled.Error(), resultText(t, res))
}

func TestUnknownAndDisabledTools(t *testing.T) {
	s := newTestServer()
	for _, name := range []string{"no_such_tool", "lsp_rename",
		"lsp_workspace_diagnostics"} {
		resp := request(t, s.MCPServer, "tools/call",
			map[string]any{"name": name, "arguments": map[string]any{}})
		e, ok := resp.(mcp.JSONRPCError)
		require.True(t, ok, "%s: got %T", name, resp)
		require.Contains(t, e.Error.Message, name)
	}
}

func TestConcurrentCalls(t *testing.T) {
	s := newTestServer(WithAllThemes())
	cases := toolCases()
	const workers = 8
	var wg sync.WaitGroup
	errs := make(chan error, workers*len(cases))
	for w := range workers {
		wg.Go(func() {
			for i := range cases {
				c := cases[(i+w)%len(cases)]
				res := callTool(t, s.MCPServer, c.tool, c.args)
				if res.IsError {
					errs <- fmt.Errorf("%s: error result", c.tool)
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
}
