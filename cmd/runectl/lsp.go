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
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

// parsePosition parses line and col strings as uint32.
func parsePosition(
	lineStr, colStr string,
) (semanticapi.Position, error) {
	line, err := strconv.ParseUint(lineStr, 10, 32)
	if err != nil {
		return semanticapi.Position{}, fmt.Errorf(
			"invalid line %q: %w", lineStr, err,
		)
	}
	col, err := strconv.ParseUint(colStr, 10, 32)
	if err != nil {
		return semanticapi.Position{}, fmt.Errorf(
			"invalid column %q: %w", colStr, err,
		)
	}
	return semanticapi.Position{
		Line:      uint32(line),
		Character: uint32(col),
	}, nil
}

func (a *app) getLSP(
	ctx context.Context,
) (semanticapi.LSP, error) {
	w, err := a.getWorkspace()
	if err != nil {
		return nil, err
	}
	return w.LSP(ctx), nil
}

type flatLocation struct {
	URI       string `json:"uri"`
	StartLine uint32 `json:"start_line"`
	StartChar uint32 `json:"start_char"`
	EndLine   uint32 `json:"end_line"`
	EndChar   uint32 `json:"end_char"`
}

func toFlatLocations(
	locs []semanticapi.Location,
) []flatLocation {
	out := make([]flatLocation, len(locs))
	for i, l := range locs {
		out[i] = flatLocation{
			URI:       l.URI,
			StartLine: l.Range.Start.Line,
			StartChar: l.Range.Start.Character,
			EndLine:   l.Range.End.Line,
			EndChar:   l.Range.End.Character,
		}
	}
	return out
}

func toFlatLocationsFromResult(
	result semanticapi.LocationResult,
) []flatLocation {
	var locs []semanticapi.Location
	if result.Location != nil {
		locs = []semanticapi.Location{*result.Location}
	} else if len(result.Locations) > 0 {
		locs = result.Locations
	} else if len(result.LocationLinks) > 0 {
		out := make([]flatLocation, len(result.LocationLinks))
		for i, ll := range result.LocationLinks {
			out[i] = flatLocation{
				URI:       ll.TargetURI,
				StartLine: ll.TargetRange.Start.Line,
				StartChar: ll.TargetRange.Start.Character,
				EndLine:   ll.TargetRange.End.Line,
				EndChar:   ll.TargetRange.End.Character,
			}
		}
		return out
	}
	return toFlatLocations(locs)
}

func printLocations(
	ctx context.Context,
	format string,
	result semanticapi.LocationResult,
) error {
	flat := toFlatLocationsFromResult(result)
	if format == "" {
		for _, f := range flat {
			fmt.Printf(
				"%s %d:%d-%d:%d\n",
				f.URI,
				f.StartLine, f.StartChar,
				f.EndLine, f.EndChar,
			)
		}
		return nil
	}
	it := iterator.FromSlice(flat)
	return printIterator(ctx, format, it, []string{
		"URI", "StartLine", "StartChar",
		"EndLine", "EndChar",
	})
}

func symbolKindString(k semanticapi.SymbolKind) string {
	switch k {
	case semanticapi.SymbolKindFile:
		return "File"
	case semanticapi.SymbolKindModule:
		return "Module"
	case semanticapi.SymbolKindNamespace:
		return "Namespace"
	case semanticapi.SymbolKindPackage:
		return "Package"
	case semanticapi.SymbolKindClass:
		return "Class"
	case semanticapi.SymbolKindMethod:
		return "Method"
	case semanticapi.SymbolKindProperty:
		return "Property"
	case semanticapi.SymbolKindField:
		return "Field"
	case semanticapi.SymbolKindConstructor:
		return "Constructor"
	case semanticapi.SymbolKindEnum:
		return "Enum"
	case semanticapi.SymbolKindInterface:
		return "Interface"
	case semanticapi.SymbolKindFunction:
		return "Function"
	case semanticapi.SymbolKindVariable:
		return "Variable"
	case semanticapi.SymbolKindConstant:
		return "Constant"
	case semanticapi.SymbolKindString:
		return "String"
	case semanticapi.SymbolKindNumber:
		return "Number"
	case semanticapi.SymbolKindBoolean:
		return "Boolean"
	case semanticapi.SymbolKindArray:
		return "Array"
	case semanticapi.SymbolKindObject:
		return "Object"
	case semanticapi.SymbolKindKey:
		return "Key"
	case semanticapi.SymbolKindNull:
		return "Null"
	case semanticapi.SymbolKindEnumMember:
		return "EnumMember"
	case semanticapi.SymbolKindStruct:
		return "Struct"
	case semanticapi.SymbolKindEvent:
		return "Event"
	case semanticapi.SymbolKindOperator:
		return "Operator"
	case semanticapi.SymbolKindTypeParameter:
		return "TypeParameter"
	default:
		return fmt.Sprintf("Unknown(%d)", k)
	}
}

func severityString(
	s semanticapi.DiagnosticSeverity,
) string {
	switch s {
	case semanticapi.DiagnosticSeverityError:
		return "Error"
	case semanticapi.DiagnosticSeverityWarning:
		return "Warning"
	case semanticapi.DiagnosticSeverityInformation:
		return "Information"
	case semanticapi.DiagnosticSeverityHint:
		return "Hint"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

func newLSPCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lsp",
		Short: "Language Server Protocol commands",
	}

	cmd.AddCommand(
		newLSPHoverCmd(a),
		newLSPDefinitionCmd(a),
		newLSPReferencesCmd(a),
		newLSPSymbolsCmd(a),
		newLSPWorkspaceSymbolsCmd(a),
		newLSPDiagnosticsCmd(a),
		newLSPRenameCmd(a),
		newLSPCodeActionsCmd(a),
	)

	return cmd
}

func newLSPHoverCmd(a *app) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "hover <file> <line> <col>",
		Short: "Get hover info at a position",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			uri, err := a.resolveURIArg(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			pos, err := parsePosition(args[1], args[2])
			if err != nil {
				return err
			}
			lsp, err := a.getLSP(cmd.Context())
			if err != nil {
				return err
			}
			h, err := lsp.Hover(cmd.Context(), semanticapi.HoverParams{
				TextDocument: semanticapi.TextDocumentIdentifier{
					URI: uri.String(),
				},
				Position: pos,
			})
			if err != nil {
				return err
			}
			if h == nil {
				return printString(format, "", []string{"Contents"})
			}
			type hoverResult struct {
				Contents string `json:"contents"`
				Kind     string `json:"kind"`
			}
			r := hoverResult{
				Contents: h.Contents.Value,
				Kind:     string(h.Contents.Kind),
			}
			return printResult(
				cmd.Context(), format, r,
				func(v hoverResult) {
					fmt.Println(v.Contents)
				},
				[]string{"Contents", "Kind"},
			)
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)

	return cmd
}

func newLSPDefinitionCmd(a *app) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "definition <file> <line> <col>",
		Short: "Go to definition",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			uri, err := a.resolveURIArg(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			pos, err := parsePosition(args[1], args[2])
			if err != nil {
				return err
			}
			lsp, err := a.getLSP(cmd.Context())
			if err != nil {
				return err
			}
			locs, err := lsp.Definition(
				cmd.Context(), semanticapi.DefinitionParams{
					TextDocument: semanticapi.TextDocumentIdentifier{
						URI: uri.String(),
					},
					Position: pos,
				},
			)
			if err != nil {
				return err
			}
			return printLocations(cmd.Context(), format, locs)
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)

	return cmd
}

func newLSPReferencesCmd(a *app) *cobra.Command {
	var format string
	var includeDecl bool

	cmd := &cobra.Command{
		Use:   "references <file> <line> <col>",
		Short: "Find all references",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			uri, err := a.resolveURIArg(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			pos, err := parsePosition(args[1], args[2])
			if err != nil {
				return err
			}
			lsp, err := a.getLSP(cmd.Context())
			if err != nil {
				return err
			}
			locs, err := lsp.References(
				cmd.Context(), semanticapi.ReferenceParams{
					TextDocument: semanticapi.TextDocumentIdentifier{
						URI: uri.String(),
					},
					Position: pos,
					Context: semanticapi.ReferenceContext{
						IncludeDeclaration: includeDecl,
					},
				},
			)
			if err != nil {
				return err
			}
			return printLocations(cmd.Context(), format,
				semanticapi.LocationResult{Locations: locs})
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)
	cmd.Flags().BoolVarP(
		&includeDecl, "declaration", "d", false,
		"Include declaration",
	)

	return cmd
}

type flatSymbol struct {
	Name      string `json:"name"`
	Detail    string `json:"detail"`
	Kind      string `json:"kind"`
	Depth     int    `json:"depth"`
	StartLine uint32 `json:"start_line"`
	StartChar uint32 `json:"start_char"`
	EndLine   uint32 `json:"end_line"`
	EndChar   uint32 `json:"end_char"`
}

func flattenSymbols(
	syms []semanticapi.DocumentSymbol, depth int,
) []flatSymbol {
	var out []flatSymbol
	for _, s := range syms {
		out = append(out, flatSymbol{
			Name:      s.Name,
			Detail:    s.Detail,
			Kind:      symbolKindString(s.Kind),
			Depth:     depth,
			StartLine: s.Range.Start.Line,
			StartChar: s.Range.Start.Character,
			EndLine:   s.Range.End.Line,
			EndChar:   s.Range.End.Character,
		})
		out = append(
			out,
			flattenSymbols(s.Children, depth+1)...,
		)
	}
	return out
}

func flattenSymbolResult(
	result semanticapi.DocumentSymbolResult,
) []flatSymbol {
	if len(result.DocumentSymbols) > 0 {
		return flattenSymbols(result.DocumentSymbols, 0)
	}
	// Convert SymbolInformation to flat format
	var out []flatSymbol
	for _, si := range result.SymbolInformation {
		out = append(out, flatSymbol{
			Name:      si.Name,
			Kind:      symbolKindString(si.Kind),
			Depth:     0,
			StartLine: si.Location.Range.Start.Line,
			StartChar: si.Location.Range.Start.Character,
			EndLine:   si.Location.Range.End.Line,
			EndChar:   si.Location.Range.End.Character,
		})
	}
	return out
}

func newLSPSymbolsCmd(a *app) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "symbols <file>",
		Short: "List document symbols",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			uri, err := a.resolveURIArg(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			lsp, err := a.getLSP(cmd.Context())
			if err != nil {
				return err
			}
			syms, err := lsp.DocumentSymbol(
				cmd.Context(), semanticapi.DocumentSymbolParams{
					TextDocument: semanticapi.TextDocumentIdentifier{
						URI: uri.String(),
					},
				},
			)
			if err != nil {
				return err
			}
			flat := flattenSymbolResult(syms)
			if format == "" {
				for _, s := range flat {
					indent := strings.Repeat("  ", s.Depth)
					fmt.Printf(
						"%s%s [%s] %d:%d-%d:%d\n",
						indent, s.Name, s.Kind,
						s.StartLine, s.StartChar,
						s.EndLine, s.EndChar,
					)
				}
				return nil
			}
			it := iterator.FromSlice(flat)
			return printIterator(cmd.Context(), format, it, []string{
				"Name", "Detail", "Kind", "Depth",
				"StartLine", "StartChar",
				"EndLine", "EndChar",
			})
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)

	return cmd
}

type flatWorkspaceSymbol struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	URI       string `json:"uri"`
	StartLine uint32 `json:"start_line"`
	StartChar uint32 `json:"start_char"`
}

func newLSPWorkspaceSymbolsCmd(a *app) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "workspace-symbols <query>",
		Short: "Search workspace symbols",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			lsp, err := a.getLSP(cmd.Context())
			if err != nil {
				return err
			}
			syms, err := lsp.WorkspaceSymbol(
				cmd.Context(), semanticapi.WorkspaceSymbolParams{
					Query: args[0],
				},
			)
			if err != nil {
				return err
			}
			flat := make([]flatWorkspaceSymbol, len(syms))
			for i, s := range syms {
				flat[i] = flatWorkspaceSymbol{
					Name:      s.Name,
					Kind:      symbolKindString(s.Kind),
					URI:       s.Location.URI,
					StartLine: s.Location.Range.Start.Line,
					StartChar: s.Location.Range.Start.Character,
				}
			}
			if format == "" {
				for _, s := range flat {
					fmt.Printf(
						"%s %s %s:%d:%d\n",
						s.Name, s.Kind,
						s.URI,
						s.StartLine, s.StartChar,
					)
				}
				return nil
			}
			it := iterator.FromSlice(flat)
			return printIterator(cmd.Context(), format, it, []string{
				"Name", "Kind", "URI",
				"StartLine", "StartChar",
			})
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)

	return cmd
}

type flatDiagnostic struct {
	Severity  string `json:"severity"`
	StartLine uint32 `json:"start_line"`
	StartChar uint32 `json:"start_char"`
	EndLine   uint32 `json:"end_line"`
	EndChar   uint32 `json:"end_char"`
	Message   string `json:"message"`
	Source    string `json:"source"`
	Code      string `json:"code"`
}

func newLSPDiagnosticsCmd(a *app) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "diagnostics <file>",
		Short: "Get diagnostics for a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			uri, err := a.resolveURIArg(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			lsp, err := a.getLSP(cmd.Context())
			if err != nil {
				return err
			}
			report, err := lsp.Diagnostic(
				cmd.Context(), semanticapi.DocumentDiagnosticParams{
					TextDocument: semanticapi.TextDocumentIdentifier{
						URI: uri.String(),
					},
				},
			)
			if err != nil {
				return err
			}
			flat := make([]flatDiagnostic, len(report.Items))
			for i, d := range report.Items {
				flat[i] = flatDiagnostic{
					Severity:  severityString(d.Severity),
					StartLine: d.Range.Start.Line,
					StartChar: d.Range.Start.Character,
					EndLine:   d.Range.End.Line,
					EndChar:   d.Range.End.Character,
					Message:   d.Message,
					Source:    d.Source,
					Code:      d.Code,
				}
			}
			if format == "" {
				for _, d := range flat {
					fmt.Printf(
						"[%s] %d:%d: %s (%s, %s)\n",
						d.Severity,
						d.StartLine, d.StartChar,
						d.Message, d.Source, d.Code,
					)
				}
				return nil
			}
			it := iterator.FromSlice(flat)
			return printIterator(cmd.Context(), format, it, []string{
				"Severity", "StartLine", "StartChar",
				"EndLine", "EndChar",
				"Message", "Source", "Code",
			})
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)

	return cmd
}

type flatRenameEdit struct {
	URI   string `json:"uri"`
	Edits int    `json:"edits"`
}

func newLSPRenameCmd(a *app) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "rename <file> <line> <col> <new-name>",
		Short: "Rename a symbol",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			uri, err := a.resolveURIArg(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			pos, err := parsePosition(args[1], args[2])
			if err != nil {
				return err
			}
			newName := args[3]
			lsp, err := a.getLSP(cmd.Context())
			if err != nil {
				return err
			}
			edit, err := lsp.Rename(
				cmd.Context(), semanticapi.RenameParams{
					TextDocument: semanticapi.TextDocumentIdentifier{
						URI: uri.String(),
					},
					Position: pos,
					NewName:  newName,
				},
			)
			if err != nil {
				return err
			}
			if edit == nil {
				return printString(
					format, "no edits",
					[]string{"Result"},
				)
			}
			var flat []flatRenameEdit
			for u, edits := range edit.Changes {
				flat = append(flat, flatRenameEdit{
					URI:   u,
					Edits: len(edits),
				})
			}
			if format == "" {
				for _, f := range flat {
					fmt.Printf("%s: %d edits\n", f.URI, f.Edits)
				}
				return nil
			}
			it := iterator.FromSlice(flat)
			return printIterator(cmd.Context(), format, it, []string{
				"URI", "Edits",
			})
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)

	return cmd
}

type flatCodeAction struct {
	Title string `json:"title"`
	Kind  string `json:"kind"`
}

func newLSPCodeActionsCmd(a *app) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "code-actions <file> <line> <col>",
		Short: "List code actions at a position",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			uri, err := a.resolveURIArg(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			pos, err := parsePosition(args[1], args[2])
			if err != nil {
				return err
			}
			lsp, err := a.getLSP(cmd.Context())
			if err != nil {
				return err
			}
			r := semanticapi.Range{Start: pos, End: pos}
			actions, err := lsp.CodeAction(
				cmd.Context(), semanticapi.CodeActionParams{
					TextDocument: semanticapi.TextDocumentIdentifier{
						URI: uri.String(),
					},
					Range: r,
					Context: semanticapi.CodeActionContext{
						Diagnostics: []semanticapi.Diagnostic{},
					},
				},
			)
			if err != nil {
				return err
			}
			flat := make([]flatCodeAction, 0, len(actions))
			for _, a := range actions {
				if a.CodeAction != nil {
					flat = append(flat, flatCodeAction{
						Title: a.CodeAction.Title,
						Kind:  string(a.CodeAction.Kind),
					})
				} else if a.Command != nil {
					flat = append(flat, flatCodeAction{
						Title: a.Command.Title,
						Kind:  "command",
					})
				}
			}
			if format == "" {
				for _, a := range flat {
					fmt.Printf("[%s] %q\n", a.Kind, a.Title)
				}
				return nil
			}
			it := iterator.FromSlice(flat)
			return printIterator(cmd.Context(), format, it, []string{
				"Title", "Kind",
			})
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)

	return cmd
}
