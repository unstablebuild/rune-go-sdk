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
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/unstablebuild/rune-go-sdk/api/extensionapi"
	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
)

// symToolOpts returns a single required symbol string
// parameter for symbol-based tools.
func symToolOpts(desc string) []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithDescription(desc),
		mcp.WithString("symbol", mcp.Required(),
			mcp.Description(
				"Symbol name to look up. "+
					"Must be an exact match.",
			),
		),
	}
}

// dedupFlatLocations removes duplicate flatLocation
// entries by (URI, StartLine, StartChar).
func dedupFlatLocations(
	locs []flatLocation,
) []flatLocation {
	type key struct {
		URI       string
		StartLine uint32
		StartChar uint32
	}
	return dedup(locs, func(l flatLocation) key {
		return key{
			URI:       l.URI,
			StartLine: l.StartLine,
			StartChar: l.StartChar,
		}
	})
}

func registerLSPTools( //nolint:funlen
	s *server.MCPServer,
	w *extensionapi.Workspace,
	bgCtx context.Context,
) {
	lsp := w.LSP(bgCtx)
	registerLSPDocumentation(s, lsp, bgCtx)
	registerLSPDefinition(s, lsp, bgCtx)
	registerLSPDeclaration(s, lsp, bgCtx)
	registerLSPTypeDefinition(s, lsp, bgCtx)
	registerLSPImplementation(s, lsp, bgCtx)
	registerLSPReferences(s, lsp, bgCtx)
	registerLSPCompletion(s, lsp, bgCtx)
	registerLSPSignatureHelp(s, lsp, bgCtx)
	registerLSPRename(s, lsp, bgCtx)
	registerLSPPrepareRename(s, lsp, bgCtx)
	registerLSPCodeActions(s, lsp, bgCtx)
	registerLSPFormatting(s, lsp, bgCtx)
	registerLSPRangeFormatting(s, lsp, bgCtx)
	registerLSPOnTypeFormatting(s, lsp, bgCtx)
	registerLSPSymbols(s, lsp, bgCtx)
	registerLSPWorkspaceSymbols(s, lsp, bgCtx)
	registerLSPDiagnostics(s, lsp, bgCtx)
	registerLSPWorkspaceDiagnostics(s, lsp, bgCtx)
	registerLSPDocHighlight(s, lsp, bgCtx)
	registerLSPCodeLens(s, lsp, bgCtx)
	registerLSPFoldingRange(s, lsp, bgCtx)
	registerLSPSelectionRange(s, lsp, bgCtx)
	registerLSPCallHierarchyPrepare(s, lsp, bgCtx)
	registerLSPCallHierarchyIncoming(s, lsp, bgCtx)
	registerLSPCallHierarchyOutgoing(s, lsp, bgCtx)
	registerLSPTypeHierarchyPrepare(s, lsp, bgCtx)
	registerLSPTypeHierarchySupertypes(s, lsp, bgCtx)
	registerLSPTypeHierarchySubtypes(s, lsp, bgCtx)
	registerLSPExecuteCommand(s, lsp, bgCtx)
}

// posToolOpts returns common uri + line + character tool
// options.
func posToolOpts(desc string) []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithDescription(desc),
		mcp.WithString("uri", mcp.Required(),
			mcp.Description(
				"Document URI. "+
					"Example: "+
					"file:///path/to/file.go",
			),
		),
		mcp.WithNumber("line", mcp.Required(),
			mcp.Description(
				"0-based line number "+
					"in the document",
			),
		),
		mcp.WithNumber("character", mcp.Required(),
			mcp.Description(
				"0-based character offset "+
					"within the line",
			),
		),
	}
}

// rangeToolOpts returns common uri + start/end range tool
// options.
func rangeToolOpts(desc string) []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithDescription(desc),
		mcp.WithString("uri", mcp.Required(),
			mcp.Description(
				"Document URI. "+
					"Example: "+
					"file:///path/to/file.go",
			),
		),
		mcp.WithNumber("start_line",
			mcp.Required(),
			mcp.Description("0-based start line"),
		),
		mcp.WithNumber("start_character",
			mcp.Required(),
			mcp.Description(
				"0-based start character",
			),
		),
		mcp.WithNumber("end_line",
			mcp.Required(),
			mcp.Description("0-based end line"),
		),
		mcp.WithNumber("end_character",
			mcp.Required(),
			mcp.Description(
				"0-based end character",
			),
		),
	}
}

// docToolOpts returns a single-file uri tool option.
func docToolOpts(desc string) []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithDescription(desc),
		mcp.WithString("uri", mcp.Required(),
			mcp.Description(
				"Document URI. "+
					"Example: "+
					"file:///path/to/file.go",
			),
		),
	}
}

func makeTextDocPos(
	uri string, line, char uint32,
) (semanticapi.TextDocumentIdentifier, semanticapi.Position) {
	return semanticapi.TextDocumentIdentifier{URI: uri},
		semanticapi.Position{
			Line: line, Character: char,
		}
}

func makeRange(r [4]uint32) semanticapi.Range {
	return semanticapi.Range{
		Start: semanticapi.Position{
			Line: r[0], Character: r[1],
		},
		End: semanticapi.Position{
			Line: r[2], Character: r[3],
		},
	}
}

func registerLSPDocumentation(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_documentation",
			symToolOpts(
				"Returns type info and docs "+
					"for a symbol by name.\n\n"+
					"Use this tool when:\n"+
					"- You need type or docs "+
					"for a symbol\n"+
					"- Inspecting what a symbol "+
					"refers to\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Need the source "+
					"— use lsp_definition\n"+
					"- Need all usages "+
					"— use lsp_references\n\n"+
					"Returns {contents, kind} "+
					"or null.",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			sp, err := resolveSymbolBest(
				bgCtx, lsp, sym, hintDefault,
			)
			if err != nil {
				return mcpErr(err), nil
			}
			td, pos := makeTextDocPos(
				sp.URI, sp.Position.Line,
				sp.Position.Character,
			)
			h, err := lsp.Hover(bgCtx,
				semanticapi.HoverParams{
					TextDocument: td,
					Position:     pos,
				})
			if err != nil {
				return mcpErr(err), nil
			}
			if h == nil {
				return mcpJSON(nil)
			}
			return mcpJSON(map[string]any{
				"contents": h.Contents.Value,
				"kind":     string(h.Contents.Kind),
			})
		},
	)
}

func registerLSPDefinition(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_definition",
			symToolOpts(
				"Finds the definition of a "+
					"symbol by name.\n\n"+
					"Use this tool when:\n"+
					"- Finding where a function, "+
					"variable, or type is "+
					"defined\n"+
					"- Navigating from usage to "+
					"implementation\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Need the declaration "+
					"— use lsp_declaration\n"+
					"- Need the type def "+
					"— use lsp_type_definition\n"+
					"- Need all usages "+
					"— use lsp_references\n\n"+
					"Returns [{uri, start_line, "+
					"start_char, end_line, "+
					"end_char}].",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			positions, err := resolveSymbol(
				bgCtx, lsp, sym, hintDefault,
			)
			if err != nil {
				return mcpErr(err), nil
			}
			var all []flatLocation
			for _, sp := range positions {
				td, pos := makeTextDocPos(
					sp.URI, sp.Position.Line,
					sp.Position.Character,
				)
				locs, err := lsp.Definition(bgCtx,
					semanticapi.DefinitionParams{
						TextDocument: td,
						Position:     pos,
					})
				if err != nil {
					return mcpErr(err), nil
				}
				all = append(all,
					toFlatLocationsFromResult(locs)...,
				)
			}
			return mcpJSON(dedupFlatLocations(all))
		},
	)
}

func registerLSPDeclaration(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_declaration",
			symToolOpts(
				"Finds the declaration of a "+
					"symbol by name.\n\n"+
					"Use this tool when:\n"+
					"- Finding the interface or "+
					"forward declaration\n"+
					"- Need declaration rather "+
					"than definition\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Need the implementation "+
					"— use lsp_definition\n"+
					"- Need all implementations "+
					"— use lsp_implementation\n\n"+
					"Returns [{uri, start_line, "+
					"start_char, end_line, "+
					"end_char}].",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			positions, err := resolveSymbol(
				bgCtx, lsp, sym, hintDefault,
			)
			if err != nil {
				return mcpErr(err), nil
			}
			var all []flatLocation
			for _, sp := range positions {
				td, pos := makeTextDocPos(
					sp.URI, sp.Position.Line,
					sp.Position.Character,
				)
				locs, err := lsp.Declaration(bgCtx,
					semanticapi.DeclarationParams{
						TextDocument: td,
						Position:     pos,
					})
				if err != nil {
					return mcpErr(err), nil
				}
				all = append(all,
					toFlatLocationsFromResult(locs)...,
				)
			}
			return mcpJSON(dedupFlatLocations(all))
		},
	)
}

func registerLSPTypeDefinition(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_type_definition",
			symToolOpts(
				"Finds the type definition of "+
					"a symbol by name.\n\n"+
					"Use this tool when:\n"+
					"- Finding the struct or "+
					"interface behind a value\n"+
					"- Need the type def rather "+
					"than the value def\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Need the symbol def "+
					"— use lsp_definition\n"+
					"- Need type hierarchy — use "+
					"lsp_prepare_type_hierarchy"+
					"\n\n"+
					"Returns [{uri, start_line, "+
					"start_char, end_line, "+
					"end_char}].",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			positions, err := resolveSymbol(
				bgCtx, lsp, sym, hintTypeDefinition,
			)
			if err != nil {
				return mcpErr(err), nil
			}
			var all []flatLocation
			for _, sp := range positions {
				td, pos := makeTextDocPos(
					sp.URI, sp.Position.Line,
					sp.Position.Character,
				)
				locs, err := lsp.TypeDefinition(bgCtx,
					semanticapi.TypeDefinitionParams{
						TextDocument: td,
						Position:     pos,
					})
				if err != nil {
					return mcpErr(err), nil
				}
				all = append(all,
					toFlatLocationsFromResult(locs)...,
				)
			}
			return mcpJSON(dedupFlatLocations(all))
		},
	)
}

func registerLSPImplementation(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_implementation",
			symToolOpts(
				"Finds all implementations of "+
					"a symbol by name.\n\n"+
					"Use this tool when:\n"+
					"- Finding concrete types "+
					"implementing an interface\n"+
					"- Listing all "+
					"implementations\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Need the interface "+
					"— use lsp_declaration\n"+
					"- Need all refs "+
					"— use lsp_references\n\n"+
					"Returns [{uri, start_line, "+
					"start_char, end_line, "+
					"end_char}].",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			positions, err := resolveSymbol(
				bgCtx, lsp, sym, hintImplementation,
			)
			if err != nil {
				return mcpErr(err), nil
			}
			var all []flatLocation
			for _, sp := range positions {
				td, pos := makeTextDocPos(
					sp.URI, sp.Position.Line,
					sp.Position.Character,
				)
				locs, err := lsp.Implementation(bgCtx,
					semanticapi.ImplementationParams{
						TextDocument: td,
						Position:     pos,
					})
				if err != nil {
					return mcpErr(err), nil
				}
				all = append(all,
					toFlatLocationsFromResult(locs)...,
				)
			}
			return mcpJSON(dedupFlatLocations(all))
		},
	)
}

func registerLSPReferences(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	opts := symToolOpts(
		"Finds all references to a " +
			"symbol by name.\n\n" +
			"Use this tool when:\n" +
			"- Finding all usages of a " +
			"symbol across the workspace\n" +
			"- Understanding the impact of " +
			"modifying a symbol\n\n" +
			"Do NOT use this tool when:\n" +
			"- Only need the definition " +
			"— use lsp_definition\n" +
			"- Want to rename " +
			"— use lsp_rename\n\n" +
			"Returns [{uri, start_line, " +
			"start_char, end_line, " +
			"end_char}].",
	)
	opts = append(opts,
		mcp.WithBoolean("include_declaration",
			mcp.Description("Include declaration"),
		),
	)
	s.AddTool(
		mcp.NewTool("lsp_references", opts...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			positions, err := resolveSymbol(
				bgCtx, lsp, sym, hintDefault,
			)
			if err != nil {
				return mcpErr(err), nil
			}
			inclDecl := req.GetBool(
				"include_declaration", false,
			)
			var all []flatLocation
			for _, sp := range positions {
				td, pos := makeTextDocPos(
					sp.URI, sp.Position.Line,
					sp.Position.Character,
				)
				locs, err := lsp.References(bgCtx,
					semanticapi.ReferenceParams{
						TextDocument: td,
						Position:     pos,
						Context: semanticapi.ReferenceContext{
							IncludeDeclaration: inclDecl,
						},
					})
				if err != nil {
					return mcpErr(err), nil
				}
				all = append(all,
					toFlatLocations(locs)...,
				)
			}
			return mcpJSON(dedupFlatLocations(all))
		},
	)
}

func registerLSPCompletion(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_completion",
			posToolOpts(
				"Returns code completions at "+
					"a position in a file.\n\n"+
					"Use this tool when:\n"+
					"- You need completions at "+
					"a cursor position\n"+
					"- Discovering available "+
					"methods or symbols\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Need signature details "+
					"— use lsp_signature_help\n"+
					"- Need hover docs "+
					"— use lsp_hover\n\n"+
					"Returns [{label, kind, "+
					"detail}].",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			uri, line, char, err := mcpPos(req)
			if err != nil {
				return mcpErr(err), nil
			}
			td, pos := makeTextDocPos(uri, line, char)
			result, err := lsp.Completion(bgCtx,
				semanticapi.CompletionParams{
					TextDocument: td,
					Position:     pos,
				})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make(
				[]flatCompletion, len(result.Items),
			)
			for i, item := range result.Items {
				flat[i] = flatCompletion{
					Label:  item.Label,
					Kind:   completionItemKindString(item.Kind),
					Detail: item.Detail,
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPSignatureHelp(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_signature_help",
			posToolOpts(
				"Returns function signature "+
					"help at a position.\n\n"+
					"Use this tool when:\n"+
					"- You need parameter names "+
					"and types for a call\n"+
					"- Identifying which param "+
					"is active at cursor\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Need completions "+
					"— use lsp_completion\n"+
					"- Need hover docs "+
					"— use lsp_hover\n\n"+
					"Returns {label, "+
					"active_signature, "+
					"active_parameter} or null.",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			uri, line, char, err := mcpPos(req)
			if err != nil {
				return mcpErr(err), nil
			}
			td, pos := makeTextDocPos(uri, line, char)
			result, err := lsp.SignatureHelp(bgCtx,
				semanticapi.SignatureHelpParams{
					TextDocument: td,
					Position:     pos,
				})
			if err != nil {
				return mcpErr(err), nil
			}
			if result == nil {
				return mcpJSON(nil)
			}
			var params []string
			for _, sig := range result.Signatures {
				for _, p := range sig.Parameters {
					params = append(params, p.Label)
				}
			}
			label := ""
			if len(result.Signatures) > 0 {
				label = result.Signatures[0].Label
			}
			return mcpJSON(flatSignatureHelp{
				Label:           label,
				ActiveSignature: result.ActiveSignature,
				ActiveParameter: result.ActiveParameter,
				Parameters:      strings.Join(params, ", "),
			})
		},
	)
}

func registerLSPRename(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	opts := symToolOpts(
		"Renames a symbol across the " +
			"workspace.\n\n" +
			"Use this tool when:\n" +
			"- Renaming a function, variable, " +
			"type, or other symbol\n" +
			"- You need a safe rename that " +
			"updates all references\n\n" +
			"Before calling:\n" +
			"- Use lsp_prepare_rename to " +
			"verify the rename is valid\n\n" +
			"Do NOT use this tool when:\n" +
			"- Just checking if rename is " +
			"possible — use " +
			"lsp_prepare_rename\n" +
			"- Finding refs without " +
			"modifying — use " +
			"lsp_references\n\n" +
			"Returns {uri: edit_count} map " +
			"or null.",
	)
	opts = append(opts,
		mcp.WithString("new_name", mcp.Required(),
			mcp.Description("New name for the symbol"),
		),
	)
	s.AddTool(
		mcp.NewTool("lsp_rename", opts...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			newName, err := req.RequireString(
				"new_name",
			)
			if err != nil {
				return mcpErr(err), nil
			}
			sp, err := resolveSymbolBest(
				bgCtx, lsp, sym, hintDefault,
			)
			if err != nil {
				return mcpErr(err), nil
			}
			td, pos := makeTextDocPos(
				sp.URI, sp.Position.Line,
				sp.Position.Character,
			)
			edit, err := lsp.Rename(bgCtx,
				semanticapi.RenameParams{
					TextDocument: td,
					Position:     pos,
					NewName:      newName,
				})
			if err != nil {
				return mcpErr(err), nil
			}
			if edit == nil {
				return mcpJSON(nil)
			}
			result := make(map[string]int)
			for u, edits := range edit.Changes {
				result[u] = len(edits)
			}
			return mcpJSON(result)
		},
	)
}

func registerLSPPrepareRename(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_prepare_rename",
			symToolOpts(
				"Checks if rename is valid "+
					"for a symbol by name.\n\n"+
					"Use this tool when:\n"+
					"- Verifying a symbol can be "+
					"renamed before lsp_rename\n"+
					"- Getting the current name "+
					"and range of the symbol\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Want to perform the "+
					"rename — use lsp_rename\n\n"+
					"Returns {range, "+
					"placeholder} or null.",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			sp, err := resolveSymbolBest(
				bgCtx, lsp, sym, hintDefault,
			)
			if err != nil {
				return mcpErr(err), nil
			}
			td, pos := makeTextDocPos(
				sp.URI, sp.Position.Line,
				sp.Position.Character,
			)
			result, err := lsp.PrepareRename(bgCtx,
				semanticapi.PrepareRenameParams{
					TextDocument: td,
					Position:     pos,
				})
			if err != nil {
				return mcpErr(err), nil
			}
			if result == nil {
				return mcpJSON(nil)
			}
			return mcpJSON(flatPrepareRename{
				StartLine:   result.Range.Start.Line,
				StartChar:   result.Range.Start.Character,
				EndLine:     result.Range.End.Line,
				EndChar:     result.Range.End.Character,
				Placeholder: result.Placeholder,
			})
		},
	)
}

func registerLSPCodeActions(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_code_actions",
			posToolOpts(
				"Lists available code actions "+
					"(refactors, fixes) at a "+
					"position.\n\n"+
					"Use this tool when:\n"+
					"- Discovering available "+
					"refactorings or quick fixes\n"+
					"- Looking for Extract "+
					"variable, Organize imports"+
					"\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Need to format "+
					"— use lsp_formatting\n"+
					"- Need diagnostics "+
					"— use lsp_diagnostics\n\n"+
					"Returns [{title, kind}].",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			uri, line, char, err := mcpPos(req)
			if err != nil {
				return mcpErr(err), nil
			}
			td, pos := makeTextDocPos(uri, line, char)
			r := semanticapi.Range{
				Start: pos, End: pos,
			}
			actions, err := lsp.CodeAction(bgCtx,
				semanticapi.CodeActionParams{
					TextDocument: td,
					Range:        r,
					Context: semanticapi.CodeActionContext{
						Diagnostics: []semanticapi.Diagnostic{},
					},
				})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make(
				[]flatCodeAction, 0, len(actions),
			)
			for _, a := range actions {
				if a.CodeAction != nil {
					flat = append(flat, flatCodeAction{
						Title: a.CodeAction.Title,
						Kind: string(
							a.CodeAction.Kind,
						),
					})
				} else if a.Command != nil {
					flat = append(flat, flatCodeAction{
						Title: a.Command.Title,
						Kind:  "command",
					})
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPFormatting(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	opts := docToolOpts(
		"Formats an entire document " +
			"using the language server.\n\n" +
			"Use this tool when:\n" +
			"- You need to auto-format " +
			"a file\n" +
			"- Applying consistent code " +
			"style to a document\n\n" +
			"Do NOT use this tool when:\n" +
			"- Only need a range " +
			"— use lsp_range_formatting\n" +
			"- Triggered by typing " +
			"— use lsp_on_type_formatting" +
			"\n\n" +
			"Returns [{start_line, " +
			"start_char, end_line, " +
			"end_char, new_text}].",
	)
	opts = append(opts,
		mcp.WithNumber("tab_size",
			mcp.Description("Tab size (default 4)"),
		),
		mcp.WithBoolean("insert_spaces",
			mcp.Description(
				"Use spaces (default true)",
			),
		),
	)
	s.AddTool(
		mcp.NewTool("lsp_formatting", opts...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			uri, err := req.RequireString("uri")
			if err != nil {
				return mcpErr(err), nil
			}
			tabSize := uint32(
				req.GetFloat("tab_size", 4),
			)
			insertSpaces := req.GetBool(
				"insert_spaces", true,
			)
			td := semanticapi.TextDocumentIdentifier{
				URI: uri,
			}
			edits, err := lsp.Formatting(bgCtx,
				semanticapi.DocumentFormattingParams{
					TextDocument: td,
					Options: semanticapi.FormattingOptions{
						TabSize:      tabSize,
						InsertSpaces: insertSpaces,
					},
				})
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(flattenTextEdits(edits))
		},
	)
}

func registerLSPRangeFormatting(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	opts := rangeToolOpts(
		"Formats a specific range " +
			"within a document.\n\n" +
			"Use this tool when:\n" +
			"- You need to format only " +
			"a portion of a file\n" +
			"- Applying formatting to a " +
			"selected code region\n\n" +
			"Do NOT use this tool when:\n" +
			"- Want the entire file " +
			"— use lsp_formatting\n" +
			"- Triggered by typing " +
			"— use lsp_on_type_formatting" +
			"\n\n" +
			"Returns [{start_line, " +
			"start_char, end_line, " +
			"end_char, new_text}].",
	)
	opts = append(opts,
		mcp.WithNumber("tab_size",
			mcp.Description("Tab size (default 4)"),
		),
		mcp.WithBoolean("insert_spaces",
			mcp.Description(
				"Use spaces (default true)",
			),
		),
	)
	s.AddTool(
		mcp.NewTool("lsp_range_formatting", opts...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			uri, rng, err := mcpRange(req)
			if err != nil {
				return mcpErr(err), nil
			}
			tabSize := uint32(
				req.GetFloat("tab_size", 4),
			)
			insertSpaces := req.GetBool(
				"insert_spaces", true,
			)
			td := semanticapi.TextDocumentIdentifier{
				URI: uri,
			}
			edits, err := lsp.RangeFormatting(bgCtx,
				semanticapi.DocumentRangeFormattingParams{
					TextDocument: td,
					Range:        makeRange(rng),
					Options: semanticapi.FormattingOptions{
						TabSize:      tabSize,
						InsertSpaces: insertSpaces,
					},
				})
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(flattenTextEdits(edits))
		},
	)
}

func registerLSPOnTypeFormatting(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_on_type_formatting",
			mcp.WithDescription(
				"Returns formatting edits "+
					"triggered by typing a "+
					"character.\n\n"+
					"Use this tool when:\n"+
					"- Getting format adjustments "+
					"after typing }, ;, or "+
					"newline\n"+
					"- Simulating auto-formatting"+
					"\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Want to format the whole "+
					"document — use "+
					"lsp_formatting\n"+
					"- Want to format a range "+
					"— use lsp_range_formatting"+
					"\n\n"+
					"Returns [{start_line, "+
					"start_char, end_line, "+
					"end_char, new_text}].",
			),
			mcp.WithString("uri", mcp.Required(),
				mcp.Description("File URI"),
			),
			mcp.WithNumber("line", mcp.Required()),
			mcp.WithNumber("character", mcp.Required()),
			mcp.WithString("ch", mcp.Required(),
				mcp.Description("The typed character"),
			),
		),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			uri, line, char, err := mcpPos(req)
			if err != nil {
				return mcpErr(err), nil
			}
			ch, err := req.RequireString("ch")
			if err != nil {
				return mcpErr(err), nil
			}
			td, pos := makeTextDocPos(uri, line, char)
			edits, err := lsp.OnTypeFormatting(bgCtx,
				semanticapi.DocumentOnTypeFormattingParams{
					TextDocument: td,
					Position:     pos,
					Character:    ch,
					Options: semanticapi.FormattingOptions{
						TabSize:      4,
						InsertSpaces: true,
					},
				})
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(flattenTextEdits(edits))
		},
	)
}

func registerLSPSymbols(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_document_symbols",
			docToolOpts(
				"Lists all symbols in a "+
					"document (functions, "+
					"classes, variables).\n\n"+
					"Use this tool when:\n"+
					"- You need an overview of "+
					"a file's structure\n"+
					"- Enumerating all "+
					"definitions in a file\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Searching across "+
					"workspace — use "+
					"lsp_workspace_symbols\n"+
					"- Need a specific symbol "+
					"— use lsp_definition\n\n"+
					"Returns nested [{name, "+
					"kind, range, children}].",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			uri, err := req.RequireString("uri")
			if err != nil {
				return mcpErr(err), nil
			}
			td := semanticapi.TextDocumentIdentifier{
				URI: uri,
			}
			syms, err := lsp.DocumentSymbol(bgCtx,
				semanticapi.DocumentSymbolParams{
					TextDocument: td,
				})
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(flattenSymbolResult(syms))
		},
	)
}

func registerLSPWorkspaceSymbols(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_workspace_symbols",
			mcp.WithDescription(
				"Searches for symbols by name "+
					"across all workspace "+
					"files.\n\n"+
					"Use this tool when:\n"+
					"- Finding a symbol by name "+
					"across the project\n"+
					"- Locating where a function "+
					"or type is defined\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Want symbols from one "+
					"file — use "+
					"lsp_document_symbols\n"+
					"- Need structural patterns "+
					"— use syntax_search\n\n"+
					"Returns [{name, kind, uri, "+
					"start_line, start_char}].",
			),
			mcp.WithString("query", mcp.Required(),
				mcp.Description("Search query"),
			),
		),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			query, err := req.RequireString("query")
			if err != nil {
				return mcpErr(err), nil
			}
			syms, err := lsp.WorkspaceSymbol(bgCtx,
				semanticapi.WorkspaceSymbolParams{
					Query: query,
				})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make(
				[]flatWorkspaceSymbol, len(syms),
			)
			for i, s := range syms {
				flat[i] = flatWorkspaceSymbol{
					Name: s.Name,
					Kind: symbolKindString(s.Kind),
					URI:  s.Location.URI,
					StartLine: s.Location.Range.Start.Line,
					StartChar: s.Location.Range.Start.Character,
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPDiagnostics(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_diagnostics",
			docToolOpts(
				"Returns diagnostics (errors, "+
					"warnings) for a file.\n\n"+
					"Use this tool when:\n"+
					"- Checking a file for "+
					"compilation errors\n"+
					"- Getting linter or "+
					"language server "+
					"diagnostics\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Need fix suggestions "+
					"— use lsp_code_actions\n"+
					"- Need to format "+
					"— use lsp_formatting\n\n"+
					"Returns [{severity, range, "+
					"message, source, code}].",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			uri, err := req.RequireString("uri")
			if err != nil {
				return mcpErr(err), nil
			}
			td := semanticapi.TextDocumentIdentifier{
				URI: uri,
			}
			report, err := lsp.Diagnostic(bgCtx,
				semanticapi.DocumentDiagnosticParams{
					TextDocument: td,
				})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make(
				[]flatDiagnostic, len(report.Items),
			)
			for i, d := range report.Items {
				flat[i] = flatDiagnostic{
					Severity: severityString(
						d.Severity,
					),
					StartLine: d.Range.Start.Line,
					StartChar: d.Range.Start.Character,
					EndLine:   d.Range.End.Line,
					EndChar:   d.Range.End.Character,
					Message:   d.Message,
					Source:    d.Source,
					Code:      d.Code,
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPWorkspaceDiagnostics(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_workspace_diagnostics",
			mcp.WithDescription(
				"Returns diagnostics (errors, "+
					"warnings) for all files "+
					"in the workspace.\n\n"+
					"Use this tool when:\n"+
					"- Checking for errors or "+
					"warnings across the "+
					"entire workspace\n"+
					"- Need a project-wide "+
					"error overview\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Only need diagnostics "+
					"for a single known file "+
					"— use lsp_diagnostics\n\n"+
					"Returns [{severity, "+
					"range, message, source, "+
					"code, uri}].",
			),
		),
		func(
			_ context.Context,
			_ mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			report, err := lsp.WorkspaceDiagnostic(
				bgCtx,
				semanticapi.WorkspaceDiagnosticParams{},
			)
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make(
				[]flatWorkspaceDiagnostic,
				0, len(report.Items),
			)
			for _, doc := range report.Items {
				for _, d := range doc.Items {
					flat = append(flat,
						flatWorkspaceDiagnostic{
							URI: doc.URI,
							Severity: severityString(
								d.Severity,
							),
							StartLine: d.Range.Start.Line,
							StartChar: d.Range.Start.Character,
							EndLine:   d.Range.End.Line,
							EndChar:   d.Range.End.Character,
							Message:   d.Message,
							Source:    d.Source,
							Code:      d.Code,
						},
					)
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPDocHighlight(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_document_highlight",
			posToolOpts(
				"Highlights occurrences of "+
					"the symbol at a position "+
					"in the file.\n\n"+
					"Use this tool when:\n"+
					"- Seeing all read/write "+
					"occurrences of a variable"+
					"\n"+
					"- Understanding symbol "+
					"usage within one document"+
					"\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Need cross-file refs "+
					"— use lsp_references\n"+
					"- Need the definition "+
					"— use lsp_definition\n\n"+
					"Returns [{kind, range}].",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			uri, line, char, err := mcpPos(req)
			if err != nil {
				return mcpErr(err), nil
			}
			td, pos := makeTextDocPos(uri, line, char)
			hl, err := lsp.DocumentHighlight(bgCtx,
				semanticapi.DocumentHighlightParams{
					TextDocument: td,
					Position:     pos,
				})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make(
				[]flatDocumentHighlight, len(hl),
			)
			for i, h := range hl {
				flat[i] = flatDocumentHighlight{
					Kind: documentHighlightKindString(
						h.Kind,
					),
					StartLine: h.Range.Start.Line,
					StartChar: h.Range.Start.Character,
					EndLine:   h.Range.End.Line,
					EndChar:   h.Range.End.Character,
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPCodeLens(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_code_lens",
			docToolOpts(
				"Returns code lenses (inline "+
					"actionable annotations) "+
					"for a file.\n\n"+
					"Use this tool when:\n"+
					"- Discovering inline "+
					"actions like Run test\n"+
					"- Finding commands "+
					"associated with code "+
					"ranges\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Want to execute a "+
					"command — use "+
					"lsp_execute_command\n"+
					"- Need code actions "+
					"— use lsp_code_actions\n\n"+
					"Returns [{range, title, "+
					"command}].",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			uri, err := req.RequireString("uri")
			if err != nil {
				return mcpErr(err), nil
			}
			td := semanticapi.TextDocumentIdentifier{
				URI: uri,
			}
			lenses, err := lsp.CodeLens(bgCtx,
				semanticapi.CodeLensParams{
					TextDocument: td,
				})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make([]flatCodeLens, len(lenses))
			for i, l := range lenses {
				var title, command string
				if l.Command != nil {
					title = l.Command.Title
					command = l.Command.Command
				}
				flat[i] = flatCodeLens{
					StartLine: l.Range.Start.Line,
					StartChar: l.Range.Start.Character,
					EndLine:   l.Range.End.Line,
					EndChar:   l.Range.End.Character,
					Title:     title,
					Command:   command,
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPFoldingRange(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_folding_range",
			docToolOpts(
				"Returns foldable ranges "+
					"(functions, blocks, "+
					"imports) in a file.\n\n"+
					"Use this tool when:\n"+
					"- Identifying logical "+
					"sections for collapsing\n"+
					"- Understanding the block "+
					"structure of a file\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Need symbol definitions "+
					"— use lsp_document_symbols"+
					"\n"+
					"- Need selection expansion "+
					"— use "+
					"lsp_selection_range\n\n"+
					"Returns [{start_line, "+
					"start_char, end_line, "+
					"end_char, kind}].",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			uri, err := req.RequireString("uri")
			if err != nil {
				return mcpErr(err), nil
			}
			td := semanticapi.TextDocumentIdentifier{
				URI: uri,
			}
			ranges, err := lsp.FoldingRange(bgCtx,
				semanticapi.FoldingRangeParams{
					TextDocument: td,
				})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make(
				[]flatFoldingRange, len(ranges),
			)
			for i, r := range ranges {
				flat[i] = flatFoldingRange{
					StartLine: r.StartLine,
					StartChar: r.StartCharacter,
					EndLine:   r.EndLine,
					EndChar:   r.EndCharacter,
					Kind: foldingRangeKindString(
						r.Kind,
					),
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPSelectionRange(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_selection_range",
			posToolOpts(
				"Returns expanding selection "+
					"ranges at a position.\n\n"+
					"Use this tool when:\n"+
					"- You need smart selection "+
					"expansion at a cursor\n"+
					"- Understanding syntactic "+
					"nesting at a point\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Need foldable regions "+
					"— use lsp_folding_range\n"+
					"- Need symbol structure "+
					"— use "+
					"lsp_document_symbols\n\n"+
					"Returns nested ranges, "+
					"each with a parent range.",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			uri, line, char, err := mcpPos(req)
			if err != nil {
				return mcpErr(err), nil
			}
			td := semanticapi.TextDocumentIdentifier{
				URI: uri,
			}
			pos := semanticapi.Position{
				Line: line, Character: char,
			}
			ranges, err := lsp.SelectionRange(bgCtx,
				semanticapi.SelectionRangeParams{
					TextDocument: td,
					Positions: []semanticapi.Position{
						pos,
					},
				})
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(
				flattenSelectionRanges(ranges),
			)
		},
	)
}

func registerLSPCallHierarchyPrepare(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_prepare_call_hierarchy",
			symToolOpts(
				"Prepares call hierarchy "+
					"info for a symbol by "+
					"name.\n\n"+
					"Use this tool when:\n"+
					"- You want to explore the "+
					"call graph of a function\n"+
					"- First step before "+
					"lsp_incoming_calls or "+
					"lsp_outgoing_calls\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Already have an item "+
					"— use lsp_incoming_calls "+
					"or lsp_outgoing_calls\n"+
					"- Need all references "+
					"— use lsp_references\n\n"+
					"Returns [{name, kind, "+
					"uri, range}].",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			positions, err := resolveSymbol(
				bgCtx, lsp, sym, hintCallHierarchy,
			)
			if err != nil {
				return mcpErr(err), nil
			}
			var flat []flatCallHierarchyItem
			for _, sp := range positions {
				td, pos := makeTextDocPos(
					sp.URI, sp.Position.Line,
					sp.Position.Character,
				)
				items, err := lsp.PrepareCallHierarchy(
					bgCtx,
					semanticapi.CallHierarchyPrepareParams{
						TextDocument: td,
						Position:     pos,
					})
				if err != nil {
					return mcpErr(err), nil
				}
				for _, item := range items {
					flat = append(flat,
						toFlatCallHierarchyItem(item),
					)
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPCallHierarchyIncoming(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_incoming_calls",
			mcp.WithDescription(
				"Returns functions that call "+
					"the given hierarchy "+
					"item.\n\n"+
					"Use this tool when:\n"+
					"- Finding all callers of "+
					"a function\n"+
					"- Tracing the call graph "+
					"backwards\n\n"+
					"Before calling:\n"+
					"- Use "+
					"lsp_prepare_call_hierarchy"+
					" to get item params\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Need what it calls "+
					"— use lsp_outgoing_calls\n"+
					"- Need all refs "+
					"— use lsp_references\n\n"+
					"Returns [{from_name, "+
					"from_kind, from_uri, "+
					"from_start_line}].",
			),
			mcp.WithString("name", mcp.Required()),
			mcp.WithString("kind", mcp.Required(),
				mcp.Description(
					"Symbol kind (e.g. function)",
				),
			),
			mcp.WithString("uri", mcp.Required()),
			mcp.WithNumber("start_line",
				mcp.Required()),
			mcp.WithNumber("start_character",
				mcp.Required()),
			mcp.WithNumber("end_line",
				mcp.Required()),
			mcp.WithNumber("end_character",
				mcp.Required()),
		),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			item, err := mcpCallHierarchyItem(req)
			if err != nil {
				return mcpErr(err), nil
			}
			calls, err := lsp.CallHierarchyIncomingCalls(
				bgCtx,
				semanticapi.CallHierarchyIncomingCallsParams{
					Item: item,
				})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make(
				[]flatIncomingCall, len(calls),
			)
			for i, c := range calls {
				flat[i] = flatIncomingCall{
					FromName: c.From.Name,
					FromKind: symbolKindString(
						c.From.Kind,
					),
					FromURI: c.From.URI,
					FromStartLine: c.From.Range.Start.Line,
					FromStartChar: c.From.Range.Start.Character,
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPCallHierarchyOutgoing(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_outgoing_calls",
			mcp.WithDescription(
				"Returns functions called by "+
					"the given hierarchy "+
					"item.\n\n"+
					"Use this tool when:\n"+
					"- Finding all functions "+
					"called by a function\n"+
					"- Tracing the call graph "+
					"forward\n\n"+
					"Before calling:\n"+
					"- Use "+
					"lsp_prepare_call_hierarchy"+
					" to get item params\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Need callers "+
					"— use lsp_incoming_calls\n"+
					"- Need all refs "+
					"— use lsp_references\n\n"+
					"Returns [{to_name, "+
					"to_kind, to_uri, "+
					"to_start_line}].",
			),
			mcp.WithString("name", mcp.Required()),
			mcp.WithString("kind", mcp.Required(),
				mcp.Description(
					"Symbol kind (e.g. function)",
				),
			),
			mcp.WithString("uri", mcp.Required()),
			mcp.WithNumber("start_line",
				mcp.Required()),
			mcp.WithNumber("start_character",
				mcp.Required()),
			mcp.WithNumber("end_line",
				mcp.Required()),
			mcp.WithNumber("end_character",
				mcp.Required()),
		),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			item, err := mcpCallHierarchyItem(req)
			if err != nil {
				return mcpErr(err), nil
			}
			calls, err := lsp.CallHierarchyOutgoingCalls(
				bgCtx,
				semanticapi.CallHierarchyOutgoingCallsParams{
					Item: item,
				})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make(
				[]flatOutgoingCall, len(calls),
			)
			for i, c := range calls {
				flat[i] = flatOutgoingCall{
					ToName: c.To.Name,
					ToKind: symbolKindString(
						c.To.Kind,
					),
					ToURI:       c.To.URI,
					ToStartLine: c.To.Range.Start.Line,
					ToStartChar: c.To.Range.Start.Character,
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPTypeHierarchyPrepare(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_prepare_type_hierarchy",
			symToolOpts(
				"Returns type hierarchy items "+
					"for a symbol by name.\n\n"+
					"Use this tool when:\n"+
					"- Exploring inheritance "+
					"chains for a type\n"+
					"- Need to understand "+
					"class/interface hierarchy"+
					"\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Want interface "+
					"implementations "+
					"— use lsp_implementation\n"+
					"- Need definition location"+
					" — use "+
					"lsp_type_definition\n\n"+
					"Returns [{name, kind, "+
					"uri, range}]. Pass items "+
					"to lsp_type_supertypes "+
					"or lsp_type_subtypes.",
			)...),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			positions, err := resolveSymbol(
				bgCtx, lsp, sym, hintTypeHierarchy,
			)
			if err != nil {
				return mcpErr(err), nil
			}
			var flat []flatTypeHierarchyItem
			for _, sp := range positions {
				td, pos := makeTextDocPos(
					sp.URI, sp.Position.Line,
					sp.Position.Character,
				)
				items, err := lsp.PrepareTypeHierarchy(
					bgCtx,
					semanticapi.TypeHierarchyPrepareParams{
						TextDocument: td,
						Position:     pos,
					})
				if err != nil {
					return mcpErr(err), nil
				}
				for _, item := range items {
					flat = append(flat,
						toFlatTypeHierarchyItem(item),
					)
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPTypeHierarchySupertypes(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_type_supertypes",
			mcp.WithDescription(
				"Returns parent types "+
					"(supertypes) of a type "+
					"hierarchy item.\n\n"+
					"Use this tool when:\n"+
					"- Walking up an "+
					"inheritance chain\n"+
					"- Finding base classes or "+
					"implemented interfaces"+
					"\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Need children "+
					"— use lsp_type_subtypes\n"+
					"- Haven't prepared yet "+
					"— call "+
					"lsp_prepare_type_hierarchy"+
					" first\n\n"+
					"Returns [{name, kind, "+
					"uri, range}].",
			),
			mcp.WithString("name", mcp.Required()),
			mcp.WithString("kind", mcp.Required(),
				mcp.Description(
					"Symbol kind (e.g. interface)",
				),
			),
			mcp.WithString("uri", mcp.Required()),
			mcp.WithNumber("start_line",
				mcp.Required()),
			mcp.WithNumber("start_character",
				mcp.Required()),
			mcp.WithNumber("end_line",
				mcp.Required()),
			mcp.WithNumber("end_character",
				mcp.Required()),
		),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			item, err := mcpTypeHierarchyItem(req)
			if err != nil {
				return mcpErr(err), nil
			}
			items, err := lsp.TypeHierarchySupertypes(
				bgCtx,
				semanticapi.TypeHierarchySupertypesParams{
					Item: item,
				})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make(
				[]flatTypeHierarchyItem, len(items),
			)
			for i, item := range items {
				flat[i] = toFlatTypeHierarchyItem(item)
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPTypeHierarchySubtypes(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_type_subtypes",
			mcp.WithDescription(
				"Returns child types "+
					"(subtypes) of a type "+
					"hierarchy item.\n\n"+
					"Use this tool when:\n"+
					"- Walking down an "+
					"inheritance chain\n"+
					"- Finding all classes "+
					"that extend a base type"+
					"\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Need parents — use "+
					"lsp_type_supertypes\n"+
					"- Haven't prepared yet "+
					"— call "+
					"lsp_prepare_type_hierarchy"+
					" first\n\n"+
					"Returns [{name, kind, "+
					"uri, range}].",
			),
			mcp.WithString("name", mcp.Required()),
			mcp.WithString("kind", mcp.Required(),
				mcp.Description(
					"Symbol kind (e.g. interface)",
				),
			),
			mcp.WithString("uri", mcp.Required()),
			mcp.WithNumber("start_line",
				mcp.Required()),
			mcp.WithNumber("start_character",
				mcp.Required()),
			mcp.WithNumber("end_line",
				mcp.Required()),
			mcp.WithNumber("end_character",
				mcp.Required()),
		),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			item, err := mcpTypeHierarchyItem(req)
			if err != nil {
				return mcpErr(err), nil
			}
			items, err := lsp.TypeHierarchySubtypes(
				bgCtx,
				semanticapi.TypeHierarchySubtypesParams{
					Item: item,
				})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make(
				[]flatTypeHierarchyItem, len(items),
			)
			for i, item := range items {
				flat[i] = toFlatTypeHierarchyItem(item)
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPExecuteCommand(
	s *server.MCPServer,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	s.AddTool(
		mcp.NewTool("lsp_execute_command",
			mcp.WithDescription(
				"Executes a server-side "+
					"workspace command by "+
					"name.\n\n"+
					"Use this tool when:\n"+
					"- Triggering a command "+
					"from lsp_code_lens or "+
					"lsp_code_actions\n"+
					"- Running refactoring or "+
					"build commands\n\n"+
					"Do NOT use this tool "+
					"when:\n"+
					"- Want code edits "+
					"— use lsp_rename or "+
					"lsp_formatting\n"+
					"- Need diagnostics "+
					"— use lsp_diagnostics\n\n"+
					"Returns command execution "+
					"result as JSON.",
			),
			mcp.WithString("command", mcp.Required(),
				mcp.Description(
					"Command identifier. "+
						"Example: \"test.run\"",
				),
			),
		),
		func(
			_ context.Context,
			req mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			command, err := req.RequireString(
				"command",
			)
			if err != nil {
				return mcpErr(err), nil
			}
			result, err := lsp.ExecuteCommand(bgCtx,
				semanticapi.ExecuteCommandParams{
					Command: command,
				})
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(map[string]string{
				"result": result,
			})
		},
	)
}
func flattenTextEdits(
	edits []semanticapi.TextEdit,
) []flatTextEdit {
	flat := make([]flatTextEdit, len(edits))
	for i, e := range edits {
		flat[i] = flatTextEdit{
			StartLine: e.Range.Start.Line,
			StartChar: e.Range.Start.Character,
			EndLine:   e.Range.End.Line,
			EndChar:   e.Range.End.Character,
			NewText:   e.NewText,
		}
	}
	return flat
}

func mcpCallHierarchyItem(
	req mcp.CallToolRequest,
) (semanticapi.CallHierarchyItem, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return semanticapi.CallHierarchyItem{}, err
	}
	kindStr, err := req.RequireString("kind")
	if err != nil {
		return semanticapi.CallHierarchyItem{}, err
	}
	uri, rng, err := mcpRange(req)
	if err != nil {
		return semanticapi.CallHierarchyItem{}, err
	}
	r := makeRange(rng)
	return semanticapi.CallHierarchyItem{
		Name:           name,
		Kind:           parseSymbolKind(kindStr),
		URI:            uri,
		Range:          r,
		SelectionRange: r,
	}, nil
}

func mcpTypeHierarchyItem(
	req mcp.CallToolRequest,
) (semanticapi.TypeHierarchyItem, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return semanticapi.TypeHierarchyItem{}, err
	}
	kindStr, err := req.RequireString("kind")
	if err != nil {
		return semanticapi.TypeHierarchyItem{}, err
	}
	uri, rng, err := mcpRange(req)
	if err != nil {
		return semanticapi.TypeHierarchyItem{}, err
	}
	r := makeRange(rng)
	return semanticapi.TypeHierarchyItem{
		Name:           name,
		Kind:           parseSymbolKind(kindStr),
		URI:            uri,
		Range:          r,
		SelectionRange: r,
	}, nil
}
