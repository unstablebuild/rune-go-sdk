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
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi"
)

// symbolArgDescription describes the symbol argument of the tools that take a symbol name.
const symbolArgDescription = `Symbol to look up. Qualify the name with its container ` +
	`path, dot-separated. Any trailing part of the path works:
  go:      <package>.<Symbol> (e.g. "term.Writer")
  python:  <module>.<symbol> or <package>.<module>.<symbol> (e.g. "_impl.Widget")
  rust:    <module>.<Symbol> (e.g. "geometry.Shape")
  methods: <Type>.<method> or <container-path>.<Type>.<method> (e.g. "orders.Order.total")
A qualified name also finds symbols of dependencies. A bare name only finds symbols ` +
	`defined in the workspace, by exact name.`

// symToolOpts returns a single required symbol string parameter for symbol-based tools.
func symToolOpts(desc string) []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithDescription(desc),
		mcp.WithString("symbol", mcp.Required(), mcp.Description(symbolArgDescription)),
	}
}

// dedupFlatLocations removes duplicate FlatLocation entries by (URI, StartLine,
// StartChar).
func dedupFlatLocations(locs []FlatLocation) []FlatLocation {
	type key struct {
		URI       string
		StartLine uint32
		StartChar uint32
	}
	return dedup(locs, func(l FlatLocation) key {
		return key{
			URI:       l.URI,
			StartLine: l.StartLine,
			StartChar: l.StartChar,
		}
	})
}

// RegisterLSPTools registers the LSP tools backed by lsp on s. Calls to lsp and parser are
// made with ctx. Only the core tools are registered unless opts enable more themes.
//
// The tools that take a symbol name resolve it through parser's symbol index first, then
// through lsp's workspace symbols; see [ResolveSymbol]. parser may be nil.
func RegisterLSPTools( //nolint:funlen
	ctx context.Context,
	s *server.MCPServer,
	lsp semanticapi.LSP,
	parser syntaxapi.Parser,
	opts ...Option,
) {
	r := newRegistrar(s, opts)
	bgCtx := ctx
	registerLSPDocumentation(r, lsp, parser, bgCtx)
	registerLSPDefinition(r, lsp, parser, bgCtx)
	registerLSPDeclaration(r, lsp, parser, bgCtx)
	registerLSPTypeDefinition(r, lsp, parser, bgCtx)
	registerLSPImplementation(r, lsp, parser, bgCtx)
	registerLSPReferences(r, lsp, parser, bgCtx)
	registerLSPCompletion(r, lsp, bgCtx)
	registerLSPSignatureHelp(r, lsp, bgCtx)
	registerLSPRename(r, lsp, parser, bgCtx)
	registerLSPPrepareRename(r, lsp, parser, bgCtx)
	registerLSPCodeActions(r, lsp, bgCtx)
	registerLSPFormatting(r, lsp, bgCtx)
	registerLSPRangeFormatting(r, lsp, bgCtx)
	registerLSPOnTypeFormatting(r, lsp, bgCtx)
	registerLSPSymbols(r, lsp, bgCtx)
	registerLSPWorkspaceSymbols(r, lsp, bgCtx)
	registerLSPDiagnostics(r, lsp, bgCtx)
	registerLSPWorkspaceDiagnostics(r, lsp, bgCtx)
	registerLSPDocHighlight(r, lsp, bgCtx)
	registerLSPCodeLens(r, lsp, bgCtx)
	registerLSPFoldingRange(r, lsp, bgCtx)
	registerLSPSelectionRange(r, lsp, bgCtx)
	registerLSPCallHierarchyPrepare(r, lsp, parser, bgCtx)
	registerLSPCallHierarchyIncoming(r, lsp, bgCtx)
	registerLSPCallHierarchyOutgoing(r, lsp, bgCtx)
	registerLSPTypeHierarchyPrepare(r, lsp, parser, bgCtx)
	registerLSPTypeHierarchySupertypes(r, lsp, bgCtx)
	registerLSPTypeHierarchySubtypes(r, lsp, bgCtx)
	registerLSPExecuteCommand(r, lsp, bgCtx)
}

// posToolOpts returns common uri + line + character tool options.
func posToolOpts(desc string) []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithDescription(desc),
		mcp.WithString("uri", mcp.Required(),
			mcp.Description("Document URI. Example: file:///path/to/file.go")),
		mcp.WithNumber("line", mcp.Required(),
			mcp.Description("0-based line number in the document")),
		mcp.WithNumber("character", mcp.Required(),
			mcp.Description("0-based character offset within the line")),
	}
}

// rangeToolOpts returns common uri + start/end range tool options.
func rangeToolOpts(desc string) []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithDescription(desc),
		mcp.WithString("uri", mcp.Required(),
			mcp.Description("Document URI. Example: file:///path/to/file.go")),
		mcp.WithNumber("start_line", mcp.Required(),
			mcp.Description("0-based start line")),
		mcp.WithNumber("start_character", mcp.Required(),
			mcp.Description("0-based start character")),
		mcp.WithNumber("end_line", mcp.Required(), mcp.Description("0-based end line")),
		mcp.WithNumber("end_character", mcp.Required(),
			mcp.Description("0-based end character")),
	}
}

// docToolOpts returns a single-file uri tool option.
func docToolOpts(desc string) []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithDescription(desc),
		mcp.WithString("uri", mcp.Required(),
			mcp.Description("Document URI. Example: file:///path/to/file.go")),
	}
}

func makeTextDocPos(
	uri string, line, char uint32,
) (semanticapi.TextDocumentIdentifier, semanticapi.Position) {
	return semanticapi.TextDocumentIdentifier{URI: uri},
		semanticapi.Position{Line: line, Character: char}
}

func makeRange(r [4]uint32) semanticapi.Range {
	return semanticapi.Range{
		Start: semanticapi.Position{Line: r[0], Character: r[1]},
		End:   semanticapi.Position{Line: r[2], Character: r[3]},
	}
}

func registerLSPDocumentation(
	r *registrar,
	lsp semanticapi.LSP,
	parser syntaxapi.Parser,
	bgCtx context.Context,
) {
	r.add(
		mcp.NewTool("lsp_documentation",
			symToolOpts(
				"Returns type info and docs for a symbol by name.\n\n"+
					"Use this tool when:\n"+
					"- You need type or docs for a symbol\n"+
					"- Inspecting what a symbol refers to\n\n"+
					"Do NOT use this tool when:\n"+
					"- Need the source — use lsp_definition\n"+
					"- Need all usages — use lsp_references\n\n"+
					"Returns [{uri, start_line, start_char, contents, kind}], one per "+
					"distinct symbol the name refers to, each at a place that defines or "+
					"references it.",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			res, err := ResolveSymbol(bgCtx, lsp, parser, sym, HintDefault)
			if err != nil {
				return mcpErr(err), nil
			}
			docs := []FlatDocumentation{}
			for _, sp := range res.Sites {
				td, pos := makeTextDocPos(sp.URI, sp.Position.Line, sp.Position.Character)
				h, err := lsp.Hover(bgCtx,
					semanticapi.HoverParams{TextDocument: td, Position: pos})
				if err != nil {
					return mcpErr(err), nil
				}
				if h == nil || h.Contents.Value == "" {
					continue
				}
				docs = append(docs, FlatDocumentation{
					URI:       sp.URI,
					StartLine: sp.Position.Line,
					StartChar: sp.Position.Character,
					Contents:  h.Contents.Value,
					Kind:      string(h.Contents.Kind),
				})
			}
			// References to one symbol hover the same; keep the first of each.
			return mcpJSONHint(dedup(docs, func(d FlatDocumentation) string {
				return d.Contents
			}), res.Hint)
		},
	)
}

func registerLSPDefinition(
	r *registrar,
	lsp semanticapi.LSP,
	parser syntaxapi.Parser,
	bgCtx context.Context,
) {
	r.add(
		mcp.NewTool("lsp_definition",
			symToolOpts(
				"Finds the definition of a symbol by name.\n\n"+
					"Use this tool when:\n"+
					"- Finding where a function, variable, or type is defined\n"+
					"- Navigating from usage to implementation\n\n"+
					"Do NOT use this tool when:\n"+
					"- Need the declaration — use lsp_declaration\n"+
					"- Need the type def — use lsp_type_definition\n"+
					"- Need all usages — use lsp_references\n\n"+
					"Returns [{uri, start_line, start_char, end_line, end_char}].",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			res, err := ResolveSymbol(bgCtx, lsp, parser, sym, HintDefault)
			if err != nil {
				return mcpErr(err), nil
			}
			var all []FlatLocation
			for _, sp := range res.Sites {
				td, pos := makeTextDocPos(sp.URI, sp.Position.Line, sp.Position.Character)
				locs, err := lsp.Definition(bgCtx,
					semanticapi.DefinitionParams{TextDocument: td, Position: pos})
				if err != nil {
					return mcpErr(err), nil
				}
				all = append(all, ToFlatLocationsFromResult(locs)...)
			}
			return mcpJSONHint(dedupFlatLocations(all), res.Hint)
		},
	)
}

func registerLSPDeclaration(
	r *registrar,
	lsp semanticapi.LSP,
	parser syntaxapi.Parser,
	bgCtx context.Context,
) {
	r.add(
		mcp.NewTool("lsp_declaration",
			symToolOpts(
				"Finds the declaration of a symbol by name.\n\n"+
					"Use this tool when:\n"+
					"- Finding the interface or forward declaration\n"+
					"- Need declaration rather than definition\n\n"+
					"Do NOT use this tool when:\n"+
					"- Need the implementation — use lsp_definition\n"+
					"- Need all implementations — use lsp_implementation\n\n"+
					"Returns [{uri, start_line, start_char, end_line, end_char}].",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			res, err := ResolveSymbol(bgCtx, lsp, parser, sym, HintDefault)
			if err != nil {
				return mcpErr(err), nil
			}
			var all []FlatLocation
			for _, sp := range res.Sites {
				td, pos := makeTextDocPos(sp.URI, sp.Position.Line, sp.Position.Character)
				locs, err := lsp.Declaration(bgCtx,
					semanticapi.DeclarationParams{TextDocument: td, Position: pos})
				if err != nil {
					return mcpErr(err), nil
				}
				all = append(all, ToFlatLocationsFromResult(locs)...)
			}
			return mcpJSONHint(dedupFlatLocations(all), res.Hint)
		},
	)
}

func registerLSPTypeDefinition(
	r *registrar,
	lsp semanticapi.LSP,
	parser syntaxapi.Parser,
	bgCtx context.Context,
) {
	r.add(
		mcp.NewTool("lsp_type_definition",
			symToolOpts(
				"Finds the type definition of a symbol by name.\n\n"+
					"Use this tool when:\n"+
					"- Finding the struct or interface behind a value\n"+
					"- Need the type def rather than the value def\n\n"+
					"Do NOT use this tool when:\n"+
					"- Need the symbol def — use lsp_definition\n"+
					"- Need type hierarchy — use lsp_prepare_type_hierarchy\n\n"+
					"Returns [{uri, start_line, start_char, end_line, end_char}].",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			res, err := ResolveSymbol(bgCtx, lsp, parser, sym, HintTypeDefinition)
			if err != nil {
				return mcpErr(err), nil
			}
			var all []FlatLocation
			for _, sp := range res.Sites {
				td, pos := makeTextDocPos(sp.URI, sp.Position.Line, sp.Position.Character)
				locs, err := lsp.TypeDefinition(bgCtx,
					semanticapi.TypeDefinitionParams{TextDocument: td, Position: pos})
				if err != nil {
					return mcpErr(err), nil
				}
				all = append(all, ToFlatLocationsFromResult(locs)...)
			}
			return mcpJSONHint(dedupFlatLocations(all), res.Hint)
		},
	)
}

func registerLSPImplementation(
	r *registrar,
	lsp semanticapi.LSP,
	parser syntaxapi.Parser,
	bgCtx context.Context,
) {
	r.add(
		mcp.NewTool("lsp_implementation",
			symToolOpts(
				"Finds all implementations of a symbol by name.\n\n"+
					"Use this tool when:\n"+
					"- Finding concrete types implementing an interface\n"+
					"- Listing all implementations\n\n"+
					"Do NOT use this tool when:\n"+
					"- Need the interface — use lsp_declaration\n"+
					"- Need all refs — use lsp_references\n\n"+
					"Returns [{uri, start_line, start_char, end_line, end_char}].",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			res, err := ResolveSymbol(bgCtx, lsp, parser, sym, HintImplementation)
			if err != nil {
				return mcpErr(err), nil
			}
			var all []FlatLocation
			for _, sp := range res.Sites {
				td, pos := makeTextDocPos(sp.URI, sp.Position.Line, sp.Position.Character)
				locs, err := lsp.Implementation(bgCtx,
					semanticapi.ImplementationParams{TextDocument: td, Position: pos})
				if err != nil {
					return mcpErr(err), nil
				}
				all = append(all, ToFlatLocationsFromResult(locs)...)
			}
			return mcpJSONHint(dedupFlatLocations(all), res.Hint)
		},
	)
}

func registerLSPReferences(
	r *registrar,
	lsp semanticapi.LSP,
	parser syntaxapi.Parser,
	bgCtx context.Context,
) {
	opts := symToolOpts(
		"Finds all references to a symbol by name.\n\n" +
			"Use this tool when:\n" +
			"- Finding all usages of a symbol across the workspace\n" +
			"- Understanding the impact of modifying a symbol\n\n" +
			"Do NOT use this tool when:\n" +
			"- Only need the definition — use lsp_definition\n" +
			"- Want to rename — use lsp_rename\n\n" +
			"Returns [{uri, start_line, start_char, end_line, end_char}].",
	)
	opts = append(opts,
		mcp.WithBoolean("include_declaration", mcp.Description("Include declaration")))
	r.add(
		mcp.NewTool("lsp_references", opts...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			res, err := ResolveSymbol(bgCtx, lsp, parser, sym, HintDefault)
			if err != nil {
				return mcpErr(err), nil
			}
			inclDecl := req.GetBool("include_declaration", false)
			var all []FlatLocation
			for _, sp := range res.Sites {
				td, pos := makeTextDocPos(sp.URI, sp.Position.Line, sp.Position.Character)
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
				all = append(all, ToFlatLocations(locs)...)
			}
			return mcpJSONHint(dedupFlatLocations(all), res.Hint)
		},
	)
}

func registerLSPCompletion(r *registrar, lsp semanticapi.LSP, bgCtx context.Context) {
	r.add(
		mcp.NewTool("lsp_completion",
			posToolOpts(
				"Returns code completions at a position in a file.\n\n"+
					"Use this tool when:\n"+
					"- You need completions at a cursor position\n"+
					"- Discovering available methods or symbols\n\n"+
					"Do NOT use this tool when:\n"+
					"- Need signature details — use lsp_signature_help\n"+
					"- Need hover docs — use lsp_documentation\n\n"+
					"Returns [{label, kind, detail}].",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uri, line, char, err := mcpPos(req)
			if err != nil {
				return mcpErr(err), nil
			}
			td, pos := makeTextDocPos(uri, line, char)
			result, err := lsp.Completion(bgCtx,
				semanticapi.CompletionParams{TextDocument: td, Position: pos})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make([]FlatCompletion, len(result.Items))
			for i, item := range result.Items {
				flat[i] = FlatCompletion{
					Label:  item.Label,
					Kind:   CompletionItemKindString(item.Kind),
					Detail: item.Detail,
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPSignatureHelp(r *registrar, lsp semanticapi.LSP, bgCtx context.Context) {
	r.add(
		mcp.NewTool("lsp_signature_help",
			posToolOpts(
				"Returns function signature help at a position.\n\n"+
					"Use this tool when:\n"+
					"- You need parameter names and types for a call\n"+
					"- Identifying which param is active at cursor\n\n"+
					"Do NOT use this tool when:\n"+
					"- Need completions — use lsp_completion\n"+
					"- Need hover docs — use lsp_documentation\n\n"+
					"Returns {label, active_signature, active_parameter, parameters} "+
					"of the active signature, or null.",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uri, line, char, err := mcpPos(req)
			if err != nil {
				return mcpErr(err), nil
			}
			td, pos := makeTextDocPos(uri, line, char)
			result, err := lsp.SignatureHelp(bgCtx,
				semanticapi.SignatureHelpParams{TextDocument: td, Position: pos})
			if err != nil {
				return mcpErr(err), nil
			}
			if result == nil || len(result.Signatures) == 0 {
				return mcpJSON(nil)
			}
			// A server may report an active signature past the end; LSP says to treat
			// that as the first.
			active := result.ActiveSignature
			if int(active) >= len(result.Signatures) {
				active = 0
			}
			sig := result.Signatures[active]
			params := make([]string, len(sig.Parameters))
			for i, p := range sig.Parameters {
				params[i] = p.Label
			}
			return mcpJSON(FlatSignatureHelp{
				Label:           sig.Label,
				ActiveSignature: active,
				ActiveParameter: result.ActiveParameter,
				Parameters:      strings.Join(params, ", "),
			})
		},
	)
}

func registerLSPRename(
	r *registrar,
	lsp semanticapi.LSP,
	parser syntaxapi.Parser,
	bgCtx context.Context,
) {
	opts := symToolOpts(
		"Dry-run rename: shows which files and how many edits a rename would produce, " +
			"but does NOT apply any changes.\n\n" +
			"Use this tool when:\n" +
			"- Previewing the scope of a rename before editing files yourself\n" +
			"- Checking which files would be affected by a rename\n\n" +
			"Before calling:\n" +
			"- Use lsp_prepare_rename to verify the rename is valid\n\n" +
			"Do NOT use this tool when:\n" +
			"- Just checking if rename is possible — use lsp_prepare_rename\n" +
			"- Finding refs without modifying — use lsp_references\n\n" +
			"Returns {uri: edit_count} map or null. Edits are NOT applied.",
	)
	opts = append(opts,
		mcp.WithString("new_name", mcp.Required(),
			mcp.Description("New name for the symbol")),
	)
	r.add(
		mcp.NewTool("lsp_rename", opts...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			newName, err := req.RequireString("new_name")
			if err != nil {
				return mcpErr(err), nil
			}
			sp, err := ResolveSymbolBest(bgCtx, lsp, parser, sym, HintDefault)
			if err != nil {
				return mcpErr(err), nil
			}
			td, pos := makeTextDocPos(sp.URI, sp.Position.Line, sp.Position.Character)
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
				return mcpJSONHint(nil, renameHint(sym))
			}
			result := make(map[string]int)
			for u, edits := range edit.Changes {
				result[u] = len(edits)
			}
			return mcpJSONHint(result, renameHint(sym))
		},
	)
}

// renameHint is the hint of a rename tool, which acts on the first site of sym only: a
// bare name may have picked one of several symbols.
func renameHint(sym string) string {
	if strings.Contains(sym, ".") {
		return ""
	}
	return bareHint(sym)
}

func registerLSPPrepareRename(
	r *registrar,
	lsp semanticapi.LSP,
	parser syntaxapi.Parser,
	bgCtx context.Context,
) {
	r.add(
		mcp.NewTool("lsp_prepare_rename",
			symToolOpts(
				"Checks if rename is valid for a symbol by name.\n\n"+
					"Use this tool when:\n"+
					"- Verifying a symbol can be renamed before lsp_rename\n"+
					"- Getting the current name and range of the symbol\n\n"+
					"Do NOT use this tool when:\n"+
					"- Want to perform the rename — use lsp_rename\n\n"+
					"Returns {range, placeholder} or null.",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			sp, err := ResolveSymbolBest(bgCtx, lsp, parser, sym, HintDefault)
			if err != nil {
				return mcpErr(err), nil
			}
			td, pos := makeTextDocPos(sp.URI, sp.Position.Line, sp.Position.Character)
			result, err := lsp.PrepareRename(bgCtx,
				semanticapi.PrepareRenameParams{TextDocument: td, Position: pos})
			if err != nil {
				return mcpErr(err), nil
			}
			if result == nil {
				return mcpJSONHint(nil, renameHint(sym))
			}
			return mcpJSONHint(FlatPrepareRename{
				StartLine:   result.Range.Start.Line,
				StartChar:   result.Range.Start.Character,
				EndLine:     result.Range.End.Line,
				EndChar:     result.Range.End.Character,
				Placeholder: result.Placeholder,
			}, renameHint(sym))
		},
	)
}

func registerLSPCodeActions(r *registrar, lsp semanticapi.LSP, bgCtx context.Context) {
	r.add(
		mcp.NewTool("lsp_code_actions",
			posToolOpts(
				"Lists available code actions (refactors, fixes) at a position.\n\n"+
					"Use this tool when:\n"+
					"- Discovering available refactorings or quick fixes\n"+
					"- Looking for Extract variable, Organize imports\n\n"+
					"Do NOT use this tool when:\n"+
					"- Need to format — use lsp_formatting\n"+
					"- Need diagnostics — use lsp_diagnostics\n\n"+
					"Returns [{title, kind}].",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uri, line, char, err := mcpPos(req)
			if err != nil {
				return mcpErr(err), nil
			}
			td, pos := makeTextDocPos(uri, line, char)
			r := semanticapi.Range{Start: pos, End: pos}
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
			flat := make([]FlatCodeAction, 0, len(actions))
			for _, a := range actions {
				if a.CodeAction != nil {
					flat = append(flat, FlatCodeAction{
						Title: a.CodeAction.Title,
						Kind:  string(a.CodeAction.Kind),
					})
				} else if a.Command != nil {
					flat = append(flat, FlatCodeAction{
						Title: a.Command.Title,
						Kind:  "command",
					})
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPFormatting(r *registrar, lsp semanticapi.LSP, bgCtx context.Context) {
	opts := docToolOpts(
		"Formats an entire document using the language server.\n\n" +
			"Use this tool when:\n" +
			"- You need to auto-format a file\n" +
			"- Applying consistent code style to a document\n\n" +
			"Do NOT use this tool when:\n" +
			"- Only need a range — use lsp_range_formatting\n" +
			"- Triggered by typing — use lsp_on_type_formatting\n\n" +
			"Returns [{start_line, start_char, end_line, end_char, new_text}].",
	)
	opts = append(opts,
		mcp.WithNumber("tab_size", mcp.Description("Tab size (default 4)")),
		mcp.WithBoolean("insert_spaces", mcp.Description("Use spaces (default true)")))
	r.add(
		mcp.NewTool("lsp_formatting", opts...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uri, err := req.RequireString("uri")
			if err != nil {
				return mcpErr(err), nil
			}
			tabSize, err := optionalUint(req, "tab_size", 4)
			if err != nil {
				return mcpErr(err), nil
			}
			insertSpaces := req.GetBool("insert_spaces", true)
			td := semanticapi.TextDocumentIdentifier{URI: uri}
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
	r *registrar,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	opts := rangeToolOpts(
		"Formats a specific range within a document.\n\n" +
			"Use this tool when:\n" +
			"- You need to format only a portion of a file\n" +
			"- Applying formatting to a selected code region\n\n" +
			"Do NOT use this tool when:\n" +
			"- Want the entire file — use lsp_formatting\n" +
			"- Triggered by typing — use lsp_on_type_formatting\n\n" +
			"Returns [{start_line, start_char, end_line, end_char, new_text}].",
	)
	opts = append(opts,
		mcp.WithNumber("tab_size", mcp.Description("Tab size (default 4)")),
		mcp.WithBoolean("insert_spaces", mcp.Description("Use spaces (default true)")))
	r.add(
		mcp.NewTool("lsp_range_formatting", opts...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uri, rng, err := mcpRange(req)
			if err != nil {
				return mcpErr(err), nil
			}
			tabSize, err := optionalUint(req, "tab_size", 4)
			if err != nil {
				return mcpErr(err), nil
			}
			insertSpaces := req.GetBool("insert_spaces", true)
			td := semanticapi.TextDocumentIdentifier{URI: uri}
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
	r *registrar,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	r.add(
		mcp.NewTool("lsp_on_type_formatting",
			mcp.WithDescription(
				"Returns formatting edits triggered by typing a character.\n\n"+
					"Use this tool when:\n"+
					"- Getting format adjustments after typing }, ;, or newline\n"+
					"- Simulating auto-formatting\n\n"+
					"Do NOT use this tool when:\n"+
					"- Want to format the whole document — use lsp_formatting\n"+
					"- Want to format a range — use lsp_range_formatting\n\n"+
					"Returns [{start_line, start_char, end_line, end_char, new_text}].",
			),
			mcp.WithString("uri", mcp.Required(), mcp.Description("File URI")),
			mcp.WithNumber("line", mcp.Required()),
			mcp.WithNumber("character", mcp.Required()),
			mcp.WithString("ch", mcp.Required(), mcp.Description("The typed character")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func registerLSPSymbols(r *registrar, lsp semanticapi.LSP, bgCtx context.Context) {
	r.add(
		mcp.NewTool("lsp_document_symbols",
			docToolOpts(
				"Lists all symbols in a document (functions, classes, variables).\n\n"+
					"Use this tool when:\n"+
					"- You need an overview of a file's structure\n"+
					"- Enumerating all definitions in a file\n\n"+
					"Do NOT use this tool when:\n"+
					"- Searching across workspace — use lsp_workspace_symbols\n"+
					"- Need a specific symbol — use lsp_definition\n\n"+
					"Returns nested [{name, kind, range, children}].",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uri, err := req.RequireString("uri")
			if err != nil {
				return mcpErr(err), nil
			}
			td := semanticapi.TextDocumentIdentifier{URI: uri}
			syms, err := lsp.DocumentSymbol(bgCtx,
				semanticapi.DocumentSymbolParams{TextDocument: td})
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(FlattenSymbolResult(syms))
		},
	)
}

func registerLSPWorkspaceSymbols(
	r *registrar,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	r.add(
		mcp.NewTool("lsp_workspace_symbols",
			mcp.WithDescription(
				"Searches for symbols by name across all workspace files.\n\n"+
					"Use this tool when:\n"+
					"- Finding a symbol by name across the project\n"+
					"- Locating where a function or type is defined\n\n"+
					"Do NOT use this tool when:\n"+
					"- Want symbols from one file — use lsp_document_symbols\n"+
					"- Need structural patterns — use syntax_search\n\n"+
					"Returns [{name, kind, uri, start_line, start_char}].",
			),
			mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query, err := req.RequireString("query")
			if err != nil {
				return mcpErr(err), nil
			}
			syms, err := lsp.WorkspaceSymbol(bgCtx,
				semanticapi.WorkspaceSymbolParams{Query: query})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make([]FlatWorkspaceSymbol, len(syms))
			for i, s := range syms {
				flat[i] = FlatWorkspaceSymbol{
					Name:      s.Name,
					Kind:      SymbolKindString(s.Kind),
					URI:       s.Location.URI,
					StartLine: s.Location.Range.Start.Line,
					StartChar: s.Location.Range.Start.Character,
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPDiagnostics(r *registrar, lsp semanticapi.LSP, bgCtx context.Context) {
	r.add(
		mcp.NewTool("lsp_diagnostics",
			docToolOpts(
				"Returns diagnostics (errors, warnings) for a file.\n\n"+
					"Use this tool when:\n"+
					"- Checking a file for compilation errors\n"+
					"- Getting linter or language server diagnostics\n\n"+
					"Do NOT use this tool when:\n"+
					"- Need fix suggestions — use lsp_code_actions\n"+
					"- Need to format — use lsp_formatting\n\n"+
					"Returns [{severity, range, message, source, code}].",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uri, err := req.RequireString("uri")
			if err != nil {
				return mcpErr(err), nil
			}
			td := semanticapi.TextDocumentIdentifier{URI: uri}
			report, err := lsp.Diagnostic(bgCtx,
				semanticapi.DocumentDiagnosticParams{TextDocument: td})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make([]FlatDiagnostic, len(report.Items))
			for i, d := range report.Items {
				flat[i] = FlatDiagnostic{
					Severity:  SeverityString(d.Severity),
					StartLine: d.Range.Start.Line,
					StartChar: d.Range.Start.Character,
					EndLine:   d.Range.End.Line,
					EndChar:   d.Range.End.Character,
					Message:   d.Message,
					Source:    d.Source,
					Code:      diagnosticCode(d),
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPWorkspaceDiagnostics(
	r *registrar,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	r.add(
		mcp.NewTool("lsp_workspace_diagnostics",
			mcp.WithDescription(
				"Returns diagnostics (errors, warnings) for all files in the workspace.\n\n"+
					"Use this tool when:\n"+
					"- Checking for errors or warnings across the entire workspace\n"+
					"- Need a project-wide error overview\n\n"+
					"Do NOT use this tool when:\n"+
					"- Only need diagnostics for a single known file "+
					"— use lsp_diagnostics\n\n"+
					"Returns [{severity, range, message, source, code, uri}].",
			),
		),
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			report, err := lsp.WorkspaceDiagnostic(bgCtx,
				semanticapi.WorkspaceDiagnosticParams{})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make([]FlatWorkspaceDiagnostic, 0, len(report.Items))
			for _, doc := range report.Items {
				for _, d := range doc.Items {
					flat = append(flat,
						FlatWorkspaceDiagnostic{
							URI:       doc.URI,
							Severity:  SeverityString(d.Severity),
							StartLine: d.Range.Start.Line,
							StartChar: d.Range.Start.Character,
							EndLine:   d.Range.End.Line,
							EndChar:   d.Range.End.Character,
							Message:   d.Message,
							Source:    d.Source,
							Code:      diagnosticCode(d),
						},
					)
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPDocHighlight(r *registrar, lsp semanticapi.LSP, bgCtx context.Context) {
	r.add(
		mcp.NewTool("lsp_document_highlight",
			posToolOpts(
				"Highlights occurrences of the symbol at a position in the file.\n\n"+
					"Use this tool when:\n"+
					"- Seeing all read/write occurrences of a variable\n"+
					"- Understanding symbol usage within one document\n\n"+
					"Do NOT use this tool when:\n"+
					"- Need cross-file refs — use lsp_references\n"+
					"- Need the definition — use lsp_definition\n\n"+
					"Returns [{kind, range}].",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uri, line, char, err := mcpPos(req)
			if err != nil {
				return mcpErr(err), nil
			}
			td, pos := makeTextDocPos(uri, line, char)
			hl, err := lsp.DocumentHighlight(bgCtx,
				semanticapi.DocumentHighlightParams{TextDocument: td, Position: pos})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make([]FlatDocumentHighlight, len(hl))
			for i, h := range hl {
				flat[i] = FlatDocumentHighlight{
					Kind:      DocumentHighlightKindString(h.Kind),
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

func registerLSPCodeLens(r *registrar, lsp semanticapi.LSP, bgCtx context.Context) {
	r.add(
		mcp.NewTool("lsp_code_lens",
			docToolOpts(
				"Returns code lenses (inline actionable annotations) for a file.\n\n"+
					"Use this tool when:\n"+
					"- Discovering inline actions like Run test\n"+
					"- Finding commands associated with code ranges\n\n"+
					"Do NOT use this tool when:\n"+
					"- Want to execute a command — use lsp_execute_command\n"+
					"- Need code actions — use lsp_code_actions\n\n"+
					"Returns [{range, title, command}].",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uri, err := req.RequireString("uri")
			if err != nil {
				return mcpErr(err), nil
			}
			td := semanticapi.TextDocumentIdentifier{URI: uri}
			lenses, err := lsp.CodeLens(bgCtx,
				semanticapi.CodeLensParams{TextDocument: td})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make([]FlatCodeLens, len(lenses))
			for i, l := range lenses {
				var title, command string
				if l.Command != nil {
					title = l.Command.Title
					command = l.Command.Command
				}
				flat[i] = FlatCodeLens{
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

func registerLSPFoldingRange(r *registrar, lsp semanticapi.LSP, bgCtx context.Context) {
	r.add(
		mcp.NewTool("lsp_folding_range",
			docToolOpts(
				"Returns foldable ranges (functions, blocks, imports) in a file.\n\n"+
					"Use this tool when:\n"+
					"- Identifying logical sections for collapsing\n"+
					"- Understanding the block structure of a file\n\n"+
					"Do NOT use this tool when:\n"+
					"- Need symbol definitions — use lsp_document_symbols\n"+
					"- Need selection expansion — use lsp_selection_range\n\n"+
					"Returns [{start_line, start_char, end_line, end_char, kind}].",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uri, err := req.RequireString("uri")
			if err != nil {
				return mcpErr(err), nil
			}
			td := semanticapi.TextDocumentIdentifier{URI: uri}
			ranges, err := lsp.FoldingRange(bgCtx,
				semanticapi.FoldingRangeParams{TextDocument: td})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make([]FlatFoldingRange, len(ranges))
			for i, r := range ranges {
				flat[i] = FlatFoldingRange{
					StartLine: r.StartLine,
					StartChar: r.StartCharacter,
					EndLine:   r.EndLine,
					EndChar:   r.EndCharacter,
					Kind:      FoldingRangeKindString(r.Kind),
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPSelectionRange(r *registrar, lsp semanticapi.LSP, bgCtx context.Context) {
	r.add(
		mcp.NewTool("lsp_selection_range",
			posToolOpts(
				"Returns expanding selection ranges at a position.\n\n"+
					"Use this tool when:\n"+
					"- You need smart selection expansion at a cursor\n"+
					"- Understanding syntactic nesting at a point\n\n"+
					"Do NOT use this tool when:\n"+
					"- Need foldable regions — use lsp_folding_range\n"+
					"- Need symbol structure — use lsp_document_symbols\n\n"+
					"Returns nested ranges, each with a parent range.",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uri, line, char, err := mcpPos(req)
			if err != nil {
				return mcpErr(err), nil
			}
			td := semanticapi.TextDocumentIdentifier{URI: uri}
			pos := semanticapi.Position{Line: line, Character: char}
			ranges, err := lsp.SelectionRange(bgCtx,
				semanticapi.SelectionRangeParams{
					TextDocument: td,
					Positions:    []semanticapi.Position{pos},
				})
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(FlattenSelectionRanges(ranges))
		},
	)
}

func registerLSPCallHierarchyPrepare(
	r *registrar,
	lsp semanticapi.LSP,
	parser syntaxapi.Parser,
	bgCtx context.Context,
) {
	r.add(
		mcp.NewTool("lsp_prepare_call_hierarchy",
			symToolOpts(
				"Prepares call hierarchy info for a symbol by name.\n\n"+
					"Use this tool when:\n"+
					"- You want to explore the call graph of a function\n"+
					"- First step before lsp_incoming_calls or lsp_outgoing_calls\n\n"+
					"Do NOT use this tool when:\n"+
					"- Already have an item — use lsp_incoming_calls or "+
					"lsp_outgoing_calls\n"+
					"- Need all references — use lsp_references\n\n"+
					"Returns [{name, kind, uri, range}].",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			res, err := ResolveSymbol(bgCtx, lsp, parser, sym, HintCallHierarchy)
			if err != nil {
				return mcpErr(err), nil
			}
			flat := []FlatCallHierarchyItem{}
			for _, sp := range res.Sites {
				td, pos := makeTextDocPos(sp.URI, sp.Position.Line, sp.Position.Character)
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
					flat = append(flat, ToFlatCallHierarchyItem(item))
				}
			}
			// Sites of one symbol prepare the same item.
			return mcpJSONHint(dedup(flat, func(i FlatCallHierarchyItem) FlatCallHierarchyItem {
				return i
			}), res.Hint)
		},
	)
}

func registerLSPCallHierarchyIncoming(
	r *registrar,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	r.add(
		mcp.NewTool("lsp_incoming_calls",
			mcp.WithDescription(
				"Returns functions that call the given hierarchy item.\n\n"+
					"Use this tool when:\n"+
					"- Finding all callers of a function\n"+
					"- Tracing the call graph backwards\n\n"+
					"Before calling:\n"+
					"- Use lsp_prepare_call_hierarchy to get item params\n\n"+
					"Do NOT use this tool when:\n"+
					"- Need what it calls — use lsp_outgoing_calls\n"+
					"- Need all refs — use lsp_references\n\n"+
					"Returns [{from_name, from_kind, from_uri, from_start_line}].",
			),
			mcp.WithString("name", mcp.Required()),
			mcp.WithString("kind", mcp.Required(),
				mcp.Description("Symbol kind (e.g. function)")),
			mcp.WithString("uri", mcp.Required()),
			mcp.WithNumber("start_line", mcp.Required()),
			mcp.WithNumber("start_character", mcp.Required()),
			mcp.WithNumber("end_line", mcp.Required()),
			mcp.WithNumber("end_character", mcp.Required()),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			item, err := mcpCallHierarchyItem(req)
			if err != nil {
				return mcpErr(err), nil
			}
			calls, err := lsp.CallHierarchyIncomingCalls(bgCtx,
				semanticapi.CallHierarchyIncomingCallsParams{Item: item})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make([]FlatIncomingCall, len(calls))
			for i, c := range calls {
				flat[i] = FlatIncomingCall{
					FromName:      c.From.Name,
					FromKind:      SymbolKindString(c.From.Kind),
					FromURI:       c.From.URI,
					FromStartLine: c.From.Range.Start.Line,
					FromStartChar: c.From.Range.Start.Character,
				}
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPCallHierarchyOutgoing(
	r *registrar,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	r.add(
		mcp.NewTool("lsp_outgoing_calls",
			mcp.WithDescription(
				"Returns functions called by the given hierarchy item.\n\n"+
					"Use this tool when:\n"+
					"- Finding all functions called by a function\n"+
					"- Tracing the call graph forward\n\n"+
					"Before calling:\n"+
					"- Use lsp_prepare_call_hierarchy to get item params\n\n"+
					"Do NOT use this tool when:\n"+
					"- Need callers — use lsp_incoming_calls\n"+
					"- Need all refs — use lsp_references\n\n"+
					"Returns [{to_name, to_kind, to_uri, to_start_line}].",
			),
			mcp.WithString("name", mcp.Required()),
			mcp.WithString("kind", mcp.Required(),
				mcp.Description("Symbol kind (e.g. function)")),
			mcp.WithString("uri", mcp.Required()),
			mcp.WithNumber("start_line", mcp.Required()),
			mcp.WithNumber("start_character", mcp.Required()),
			mcp.WithNumber("end_line", mcp.Required()),
			mcp.WithNumber("end_character", mcp.Required()),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			item, err := mcpCallHierarchyItem(req)
			if err != nil {
				return mcpErr(err), nil
			}
			calls, err := lsp.CallHierarchyOutgoingCalls(bgCtx,
				semanticapi.CallHierarchyOutgoingCallsParams{Item: item})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make([]FlatOutgoingCall, len(calls))
			for i, c := range calls {
				flat[i] = FlatOutgoingCall{
					ToName:      c.To.Name,
					ToKind:      SymbolKindString(c.To.Kind),
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
	r *registrar,
	lsp semanticapi.LSP,
	parser syntaxapi.Parser,
	bgCtx context.Context,
) {
	r.add(
		mcp.NewTool("lsp_prepare_type_hierarchy",
			symToolOpts(
				"Returns type hierarchy items for a symbol by name.\n\n"+
					"Use this tool when:\n"+
					"- Exploring inheritance chains for a type\n"+
					"- Need to understand class/interface hierarchy\n\n"+
					"Do NOT use this tool when:\n"+
					"- Want interface implementations — use lsp_implementation\n"+
					"- Need definition location — use lsp_type_definition\n\n"+
					"Returns [{name, kind, uri, range}]. Pass items to "+
					"lsp_type_supertypes or lsp_type_subtypes.",
			)...),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sym, err := req.RequireString("symbol")
			if err != nil {
				return mcpErr(err), nil
			}
			res, err := ResolveSymbol(bgCtx, lsp, parser, sym, HintTypeHierarchy)
			if err != nil {
				return mcpErr(err), nil
			}
			flat := []FlatTypeHierarchyItem{}
			for _, sp := range res.Sites {
				td, pos := makeTextDocPos(sp.URI, sp.Position.Line, sp.Position.Character)
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
					flat = append(flat, ToFlatTypeHierarchyItem(item))
				}
			}
			// Sites of one symbol prepare the same item.
			return mcpJSONHint(dedup(flat, func(i FlatTypeHierarchyItem) FlatTypeHierarchyItem {
				return i
			}), res.Hint)
		},
	)
}

func registerLSPTypeHierarchySupertypes(
	r *registrar,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	r.add(
		mcp.NewTool("lsp_type_supertypes",
			mcp.WithDescription(
				"Returns parent types (supertypes) of a type hierarchy item.\n\n"+
					"Use this tool when:\n"+
					"- Walking up an inheritance chain\n"+
					"- Finding base classes or implemented interfaces\n\n"+
					"Do NOT use this tool when:\n"+
					"- Need children — use lsp_type_subtypes\n"+
					"- Haven't prepared yet — call lsp_prepare_type_hierarchy first\n\n"+
					"Returns [{name, kind, uri, range}].",
			),
			mcp.WithString("name", mcp.Required()),
			mcp.WithString("kind", mcp.Required(),
				mcp.Description("Symbol kind (e.g. interface)")),
			mcp.WithString("uri", mcp.Required()),
			mcp.WithNumber("start_line", mcp.Required()),
			mcp.WithNumber("start_character", mcp.Required()),
			mcp.WithNumber("end_line", mcp.Required()),
			mcp.WithNumber("end_character", mcp.Required()),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			item, err := mcpTypeHierarchyItem(req)
			if err != nil {
				return mcpErr(err), nil
			}
			items, err := lsp.TypeHierarchySupertypes(bgCtx,
				semanticapi.TypeHierarchySupertypesParams{Item: item})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make([]FlatTypeHierarchyItem, len(items))
			for i, item := range items {
				flat[i] = ToFlatTypeHierarchyItem(item)
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPTypeHierarchySubtypes(
	r *registrar,
	lsp semanticapi.LSP,
	bgCtx context.Context,
) {
	r.add(
		mcp.NewTool("lsp_type_subtypes",
			mcp.WithDescription(
				"Returns child types (subtypes) of a type hierarchy item.\n\n"+
					"Use this tool when:\n"+
					"- Walking down an inheritance chain\n"+
					"- Finding all classes that extend a base type\n\n"+
					"Do NOT use this tool when:\n"+
					"- Need parents — use lsp_type_supertypes\n"+
					"- Haven't prepared yet — call lsp_prepare_type_hierarchy first\n\n"+
					"Returns [{name, kind, uri, range}].",
			),
			mcp.WithString("name", mcp.Required()),
			mcp.WithString("kind", mcp.Required(),
				mcp.Description("Symbol kind (e.g. interface)")),
			mcp.WithString("uri", mcp.Required()),
			mcp.WithNumber("start_line", mcp.Required()),
			mcp.WithNumber("start_character", mcp.Required()),
			mcp.WithNumber("end_line", mcp.Required()),
			mcp.WithNumber("end_character", mcp.Required()),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			item, err := mcpTypeHierarchyItem(req)
			if err != nil {
				return mcpErr(err), nil
			}
			items, err := lsp.TypeHierarchySubtypes(bgCtx,
				semanticapi.TypeHierarchySubtypesParams{Item: item})
			if err != nil {
				return mcpErr(err), nil
			}
			flat := make([]FlatTypeHierarchyItem, len(items))
			for i, item := range items {
				flat[i] = ToFlatTypeHierarchyItem(item)
			}
			return mcpJSON(flat)
		},
	)
}

func registerLSPExecuteCommand(r *registrar, lsp semanticapi.LSP, bgCtx context.Context) {
	r.add(
		mcp.NewTool("lsp_execute_command",
			mcp.WithDescription(
				"Executes a server-side workspace command by name.\n\n"+
					"Use this tool when:\n"+
					"- Triggering a command from lsp_code_lens or lsp_code_actions\n"+
					"- Running refactoring or build commands\n\n"+
					"Do NOT use this tool when:\n"+
					"- Want code edits — use lsp_rename or lsp_formatting\n"+
					"- Need diagnostics — use lsp_diagnostics\n\n"+
					"Returns command execution result as JSON.",
			),
			mcp.WithString("command", mcp.Required(),
				mcp.Description("Command identifier. Example: \"test.run\"")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			command, err := req.RequireString("command")
			if err != nil {
				return mcpErr(err), nil
			}
			result, err := lsp.ExecuteCommand(bgCtx,
				semanticapi.ExecuteCommandParams{Command: command})
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(map[string]string{"result": result})
		},
	)
}
func flattenTextEdits(edits []semanticapi.TextEdit) []FlatTextEdit {
	flat := make([]FlatTextEdit, len(edits))
	for i, e := range edits {
		flat[i] = FlatTextEdit{
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
		Kind:           ParseSymbolKind(kindStr),
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
		Kind:           ParseSymbolKind(kindStr),
		URI:            uri,
		Range:          r,
		SelectionRange: r,
	}, nil
}
