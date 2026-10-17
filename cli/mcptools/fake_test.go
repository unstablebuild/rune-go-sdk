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
	"sync"

	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"github.com/unstablebuild/rune-go-sdk/term"
)

// call is one call a fake received.
type call struct {
	method string
	params any
	ctx    context.Context
}

// recorder records the calls a fake receives and answers each with the error configured
// for its method.
type recorder struct {
	mu    sync.Mutex
	calls []call
	errs  map[string]error
}

func (r *recorder) record(ctx context.Context, method string, params any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, call{method: method, params: params, ctx: ctx})
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.errs[method]
}

// recordNoContext records a call of a method that takes no context; its call has a nil
// ctx.
func (r *recorder) recordNoContext(method string, params any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, call{method: method, params: params})
	return r.errs[method]
}

// failOn makes method answer err.
func (r *recorder) failOn(method string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.errs == nil {
		r.errs = make(map[string]error)
	}
	r.errs[method] = err
}

// called returns the calls made to method, in order.
func (r *recorder) called(method string) []call {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []call
	for _, c := range r.calls {
		if c.method == method {
			out = append(out, c)
		}
	}
	return out
}

// methods returns the names of the methods called, in order.
func (r *recorder) methods() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.calls))
	for i, c := range r.calls {
		out[i] = c.method
	}
	return out
}

func pos(line, char uint32) semanticapi.Position {
	return semanticapi.Position{Line: line, Character: char}
}

func rng(sl, sc, el, ec uint32) semanticapi.Range {
	return semanticapi.Range{Start: pos(sl, sc), End: pos(el, ec)}
}

func loc(uri string, sl, sc, el, ec uint32) semanticapi.Location {
	return semanticapi.Location{URI: uri, Range: rng(sl, sc, el, ec)}
}

// widen returns the range of n characters from p.
func widen(p semanticapi.Position, n uint32) semanticapi.Range {
	return semanticapi.Range{Start: p, End: pos(p.Line, p.Character+n)}
}

// defaultSymbols are the workspace symbols of fakeLSP.
func defaultSymbols() []semanticapi.SymbolInformation {
	return []semanticapi.SymbolInformation{
		{Name: "MyFunc", Kind: semanticapi.SymbolKindFunction,
			Location: loc("file:///src/main.go", 5, 0, 5, 10)},
		{Name: "MyInterface", Kind: semanticapi.SymbolKindInterface,
			Location: loc("file:///src/types.go", 15, 5, 18, 1)},
		{Name: "MyStruct", Kind: semanticapi.SymbolKindStruct,
			Location: loc("file:///src/types.go", 20, 5, 25, 1)},
		{Name: "myVar", Kind: semanticapi.SymbolKindVariable,
			Location: loc("file:///src/main.go", 3, 4, 3, 9)},
		{Name: "MyStruct.MyMethod", Kind: semanticapi.SymbolKindMethod,
			Location: loc("file:///src/types.go", 30, 0, 35, 1)},
	}
}

// fakeLSP implements every semanticapi.LSP method a tool calls; any other method panics
// through the nil embedded interface. Results that depend on the request echo it, so
// tests can see what the tools passed.
type fakeLSP struct {
	semanticapi.LSP
	recorder

	symbols []semanticapi.SymbolInformation
	// defs answers Definition in a document, by URI, as a language server does at a
	// reference. Definition elsewhere echoes the requested position.
	defs map[string]semanticapi.Location
	// hovers answers Hover in a document, by URI. Hover elsewhere answers MyFunc's.
	hovers map[string]string
	// empty makes every method but WorkspaceSymbol answer its zero value, so symbol tools
	// still resolve.
	empty bool
}

func newFakeLSP() *fakeLSP {
	return &fakeLSP{symbols: defaultSymbols()}
}

func (f *fakeLSP) Hover(
	ctx context.Context,
	p semanticapi.HoverParams,
) (*semanticapi.Hover, error) {
	if err := f.record(ctx, "Hover", p); err != nil || f.empty {
		return nil, err
	}
	if h, ok := f.hovers[p.TextDocument.URI]; ok {
		return &semanticapi.Hover{Contents: semanticapi.MarkupContent{
			Kind: "markdown", Value: h,
		}}, nil
	}
	return &semanticapi.Hover{Contents: semanticapi.MarkupContent{
		Kind: "markdown", Value: "func MyFunc()",
	}}, nil
}

func (f *fakeLSP) Definition(
	ctx context.Context,
	p semanticapi.DefinitionParams,
) (semanticapi.LocationResult, error) {
	if err := f.record(ctx, "Definition", p); err != nil || f.empty {
		return semanticapi.LocationResult{}, err
	}
	if l, ok := f.defs[p.TextDocument.URI]; ok {
		return semanticapi.LocationResult{Locations: []semanticapi.Location{l}}, nil
	}
	return semanticapi.LocationResult{Locations: []semanticapi.Location{{
		URI: p.TextDocument.URI, Range: widen(p.Position, 6),
	}}}, nil
}

func (f *fakeLSP) Declaration(
	ctx context.Context,
	p semanticapi.DeclarationParams,
) (semanticapi.LocationResult, error) {
	if err := f.record(ctx, "Declaration", p); err != nil || f.empty {
		return semanticapi.LocationResult{}, err
	}
	l := loc("file:///src/decl.go", 7, 1, 7, 9)
	return semanticapi.LocationResult{Location: &l}, nil
}

func (f *fakeLSP) TypeDefinition(
	ctx context.Context,
	p semanticapi.TypeDefinitionParams,
) (semanticapi.LocationResult, error) {
	if err := f.record(ctx, "TypeDefinition", p); err != nil || f.empty {
		return semanticapi.LocationResult{}, err
	}
	return semanticapi.LocationResult{LocationLinks: []semanticapi.LocationLink{{
		TargetURI:            "file:///src/types.go",
		TargetRange:          rng(20, 5, 25, 1),
		TargetSelectionRange: rng(20, 5, 20, 13),
	}}}, nil
}

func (f *fakeLSP) Implementation(
	ctx context.Context,
	p semanticapi.ImplementationParams,
) (semanticapi.LocationResult, error) {
	if err := f.record(ctx, "Implementation", p); err != nil || f.empty {
		return semanticapi.LocationResult{}, err
	}
	return semanticapi.LocationResult{Locations: []semanticapi.Location{
		loc("file:///src/impl1.go", 10, 0, 10, 10),
		loc("file:///src/impl2.go", 25, 0, 25, 10),
		// Same start as the first: deduplicated away.
		loc("file:///src/impl1.go", 10, 0, 10, 12),
	}}, nil
}

func (f *fakeLSP) References(
	ctx context.Context,
	p semanticapi.ReferenceParams,
) ([]semanticapi.Location, error) {
	if err := f.record(ctx, "References", p); err != nil || f.empty {
		return nil, err
	}
	return []semanticapi.Location{
		{URI: p.TextDocument.URI, Range: widen(p.Position, 6)},
		loc("file:///src/util.go", 20, 3, 20, 13),
	}, nil
}

func (f *fakeLSP) Completion(
	ctx context.Context,
	p semanticapi.CompletionParams,
) (semanticapi.CompletionResult, error) {
	if err := f.record(ctx, "Completion", p); err != nil || f.empty {
		return semanticapi.CompletionResult{}, err
	}
	return semanticapi.CompletionResult{Items: []semanticapi.CompletionItem{
		{Label: "fmt", Kind: semanticapi.CompletionItemKindModule, Detail: "package fmt"},
		{Label: "Println", Kind: semanticapi.CompletionItemKindFunction,
			Detail: "func Println(a ...any)"},
	}}, nil
}

func (f *fakeLSP) SignatureHelp(
	ctx context.Context,
	p semanticapi.SignatureHelpParams,
) (*semanticapi.SignatureHelp, error) {
	if err := f.record(ctx, "SignatureHelp", p); err != nil || f.empty {
		return nil, err
	}
	return &semanticapi.SignatureHelp{
		Signatures: []semanticapi.SignatureInformation{
			{Label: "func Print(a ...any)", Parameters: []semanticapi.ParameterInformation{
				{Label: "a ...any"}}},
			{Label: "func Printf(format string, a " +
				"...any)", Parameters: []semanticapi.ParameterInformation{
				{Label: "format string"}, {Label: "a ...any"},
			}},
		},
		ActiveSignature: 1,
		ActiveParameter: 1,
	}, nil
}

func (f *fakeLSP) Rename(
	ctx context.Context,
	p semanticapi.RenameParams,
) (*semanticapi.WorkspaceEdit, error) {
	if err := f.record(ctx, "Rename", p); err != nil || f.empty {
		return nil, err
	}
	edit := semanticapi.TextEdit{Range: widen(p.Position, 6), NewText: p.NewName}
	return &semanticapi.WorkspaceEdit{Changes: map[string][]semanticapi.TextEdit{
		p.TextDocument.URI:    {edit, edit},
		"file:///src/util.go": {edit},
	}}, nil
}

func (f *fakeLSP) PrepareRename(
	ctx context.Context,
	p semanticapi.PrepareRenameParams,
) (*semanticapi.PrepareRenameResult, error) {
	if err := f.record(ctx, "PrepareRename", p); err != nil || f.empty {
		return nil, err
	}
	return &semanticapi.PrepareRenameResult{Range: widen(p.Position, 6),
		Placeholder: "MyFunc"}, nil
}

func (f *fakeLSP) CodeAction(
	ctx context.Context,
	p semanticapi.CodeActionParams,
) ([]semanticapi.CodeActionResult, error) {
	if err := f.record(ctx, "CodeAction", p); err != nil || f.empty {
		return nil, err
	}
	return []semanticapi.CodeActionResult{
		{CodeAction: &semanticapi.CodeAction{Title: "Extract variable",
			Kind: "refactor.extract"}},
		{Command: &semanticapi.Command{Title: "Run test", Command: "test.run"}},
		// Neither form: skipped.
		{},
	}, nil
}

func (f *fakeLSP) Formatting(
	ctx context.Context,
	p semanticapi.DocumentFormattingParams,
) ([]semanticapi.TextEdit, error) {
	if err := f.record(ctx, "Formatting", p); err != nil || f.empty {
		return nil, err
	}
	return []semanticapi.TextEdit{{Range: rng(0, 0, 0, 4), NewText: "\t"}}, nil
}

func (f *fakeLSP) RangeFormatting(
	ctx context.Context,
	p semanticapi.DocumentRangeFormattingParams,
) ([]semanticapi.TextEdit, error) {
	if err := f.record(ctx, "RangeFormatting", p); err != nil || f.empty {
		return nil, err
	}
	return []semanticapi.TextEdit{{Range: p.Range, NewText: "fixed"}}, nil
}

func (f *fakeLSP) OnTypeFormatting(
	ctx context.Context,
	p semanticapi.DocumentOnTypeFormattingParams,
) ([]semanticapi.TextEdit, error) {
	if err := f.record(ctx, "OnTypeFormatting", p); err != nil || f.empty {
		return nil, err
	}
	return []semanticapi.TextEdit{
		{Range: widen(p.Position, 0), NewText: p.Character + "\n"}}, nil
}

func (f *fakeLSP) DocumentSymbol(
	ctx context.Context,
	p semanticapi.DocumentSymbolParams,
) (semanticapi.DocumentSymbolResult, error) {
	if err := f.record(ctx, "DocumentSymbol", p); err != nil || f.empty {
		return semanticapi.DocumentSymbolResult{}, err
	}
	return semanticapi.DocumentSymbolResult{DocumentSymbols: []semanticapi.DocumentSymbol{{
		Name: "main", Detail: "func()", Kind: semanticapi.SymbolKindFunction,
		Range: rng(5, 0, 10, 1),
		Children: []semanticapi.DocumentSymbol{
			{Name: "x", Kind: semanticapi.SymbolKindVariable, Range: rng(6, 1, 6, 10)},
		},
	}}}, nil
}

func (f *fakeLSP) WorkspaceSymbol(
	ctx context.Context,
	p semanticapi.WorkspaceSymbolParams,
) ([]semanticapi.SymbolInformation, error) {
	if err := f.record(ctx, "WorkspaceSymbol", p); err != nil {
		return nil, err
	}
	var out []semanticapi.SymbolInformation
	for _, s := range f.symbols {
		if strings.Contains(s.Name, p.Query) {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f *fakeLSP) Diagnostic(
	ctx context.Context,
	p semanticapi.DocumentDiagnosticParams,
) (semanticapi.DocumentDiagnosticReport, error) {
	if err := f.record(ctx, "Diagnostic", p); err != nil || f.empty {
		return semanticapi.DocumentDiagnosticReport{}, err
	}
	return semanticapi.DocumentDiagnosticReport{Kind: "full", Items: []semanticapi.Diagnostic{
		{
			Range: rng(3, 10, 3, 15), Severity: semanticapi.DiagnosticSeverityError,
			Code: "E001", Source: "compiler", Message: "undefined: x",
		},
		{
			Range: rng(4, 0, 4, 2), Severity: semanticapi.DiagnosticSeverityWarning,
			CodeInt: 2304, CodeIsInt: true, Source: "ts", Message: "Cannot find name 'y'.",
		},
	}}, nil
}

func (f *fakeLSP) WorkspaceDiagnostic(
	ctx context.Context, p semanticapi.WorkspaceDiagnosticParams,
) (semanticapi.WorkspaceDiagnosticReport, error) {
	if err := f.record(ctx, "WorkspaceDiagnostic", p); err != nil || f.empty {
		return semanticapi.WorkspaceDiagnosticReport{}, err
	}
	items := []semanticapi.WorkspaceDocumentDiagnosticReport{
		{URI: "file:///src/main.go", Items: []semanticapi.Diagnostic{{
			Range: rng(3, 10, 3, 15), Severity: semanticapi.DiagnosticSeverityError,
			Code: "E001", Source: "compiler", Message: "undefined: x",
		}}},
		{URI: "file:///src/util.go", Items: []semanticapi.Diagnostic{
			{Range: rng(7, 0, 7, 5), Severity: semanticapi.DiagnosticSeverityHint,
				Message: "simplify"},
			{Range: rng(9, 0, 9, 1), Severity: semanticapi.DiagnosticSeverityInformation,
				CodeInt: 7, CodeIsInt: true, Message: "note"},
		}},
		// A file without diagnostics adds nothing.
		{URI: "file:///src/clean.go"},
	}
	return semanticapi.WorkspaceDiagnosticReport{Items: items}, nil
}

func (f *fakeLSP) DocumentHighlight(
	ctx context.Context,
	p semanticapi.DocumentHighlightParams,
) ([]semanticapi.DocumentHighlight, error) {
	if err := f.record(ctx, "DocumentHighlight", p); err != nil || f.empty {
		return nil, err
	}
	return []semanticapi.DocumentHighlight{
		{Range: rng(5, 5, 5, 9), Kind: semanticapi.DocumentHighlightKindRead},
		{Range: rng(10, 3, 10, 7), Kind: semanticapi.DocumentHighlightKindWrite},
		{Range: rng(12, 0, 12, 4), Kind: semanticapi.DocumentHighlightKindText},
	}, nil
}

func (f *fakeLSP) CodeLens(
	ctx context.Context,
	p semanticapi.CodeLensParams,
) ([]semanticapi.CodeLens, error) {
	if err := f.record(ctx, "CodeLens", p); err != nil || f.empty {
		return nil, err
	}
	return []semanticapi.CodeLens{
		{Range: rng(5, 0, 5, 10),
			Command: &semanticapi.Command{Title: "Run test", Command: "test.run"}},
		// An unresolved lens has no command.
		{Range: rng(7, 0, 7, 3)},
	}, nil
}

func (f *fakeLSP) FoldingRange(
	ctx context.Context,
	p semanticapi.FoldingRangeParams,
) ([]semanticapi.FoldingRange, error) {
	if err := f.record(ctx, "FoldingRange", p); err != nil || f.empty {
		return nil, err
	}
	return []semanticapi.FoldingRange{
		{StartLine: 5, EndLine: 20, EndCharacter: 1,
			Kind: semanticapi.FoldingRangeKindRegion},
		{StartLine: 1, EndLine: 3, Kind: semanticapi.FoldingRangeKindImports},
		{StartLine: 22, StartCharacter: 2, EndLine: 24,
			Kind: semanticapi.FoldingRangeKindComment},
		{StartLine: 30, EndLine: 31},
	}, nil
}

func (f *fakeLSP) SelectionRange(
	ctx context.Context,
	p semanticapi.SelectionRangeParams,
) ([]semanticapi.SelectionRange, error) {
	if err := f.record(ctx, "SelectionRange", p); err != nil || f.empty {
		return nil, err
	}
	out := make([]semanticapi.SelectionRange, len(p.Positions))
	for i, at := range p.Positions {
		out[i] = semanticapi.SelectionRange{
			Range: widen(at, 4),
			Parent: &semanticapi.SelectionRange{
				Range:  rng(at.Line, 0, at.Line+5, 1),
				Parent: &semanticapi.SelectionRange{Range: rng(0, 0, 100, 0)},
			},
		}
	}
	return out, nil
}

func (f *fakeLSP) PrepareCallHierarchy(
	ctx context.Context, p semanticapi.CallHierarchyPrepareParams,
) ([]semanticapi.CallHierarchyItem, error) {
	if err := f.record(ctx, "PrepareCallHierarchy", p); err != nil || f.empty {
		return nil, err
	}
	return []semanticapi.CallHierarchyItem{{
		Name: "MyFunc", Kind: semanticapi.SymbolKindFunction, URI: p.TextDocument.URI,
		Range: rng(p.Position.Line, 0, p.Position.Line+5, 1), SelectionRange: widen(
			p.Position, 6),
	}}, nil
}

func (f *fakeLSP) CallHierarchyIncomingCalls(
	ctx context.Context, p semanticapi.CallHierarchyIncomingCallsParams,
) ([]semanticapi.CallHierarchyIncomingCall, error) {
	if err := f.record(ctx, "CallHierarchyIncomingCalls", p); err != nil || f.empty {
		return nil, err
	}
	return []semanticapi.CallHierarchyIncomingCall{{
		From: semanticapi.CallHierarchyItem{
			Name: "runApp", Kind: semanticapi.SymbolKindFunction, URI: "file:///src/app.go",
			Range: rng(20, 0, 30, 1), SelectionRange: rng(20, 5, 20, 11),
		},
		FromRanges: []semanticapi.Range{rng(25, 2, 25, 8)},
	}}, nil
}

func (f *fakeLSP) CallHierarchyOutgoingCalls(
	ctx context.Context, p semanticapi.CallHierarchyOutgoingCallsParams,
) ([]semanticapi.CallHierarchyOutgoingCall, error) {
	if err := f.record(ctx, "CallHierarchyOutgoingCalls", p); err != nil || f.empty {
		return nil, err
	}
	return []semanticapi.CallHierarchyOutgoingCall{{
		To: semanticapi.CallHierarchyItem{
			Name: "Println", Kind: semanticapi.SymbolKindFunction,
			URI:   "file:///go/src/fmt/print.go",
			Range: rng(100, 0, 110, 1), SelectionRange: rng(100, 5, 100, 12),
		},
		FromRanges: []semanticapi.Range{rng(7, 2, 7, 13)},
	}}, nil
}

func (f *fakeLSP) PrepareTypeHierarchy(
	ctx context.Context, p semanticapi.TypeHierarchyPrepareParams,
) ([]semanticapi.TypeHierarchyItem, error) {
	if err := f.record(ctx, "PrepareTypeHierarchy", p); err != nil || f.empty {
		return nil, err
	}
	return []semanticapi.TypeHierarchyItem{{
		Name: "MyInterface", Kind: semanticapi.SymbolKindInterface, Detail: "interface",
		URI: p.TextDocument.URI, Range: rng(p.Position.Line, 0, p.Position.Line+3, 1),
		SelectionRange: widen(p.Position, 11),
	}}, nil
}

func (f *fakeLSP) TypeHierarchySupertypes(
	ctx context.Context, p semanticapi.TypeHierarchySupertypesParams,
) ([]semanticapi.TypeHierarchyItem, error) {
	if err := f.record(ctx, "TypeHierarchySupertypes", p); err != nil || f.empty {
		return nil, err
	}
	return []semanticapi.TypeHierarchyItem{{
		Name: "io.Reader", Kind: semanticapi.SymbolKindInterface, Detail: "interface",
		URI: "file:///go/src/io/io.go", Range: rng(50, 0, 53, 1),
	}}, nil
}

func (f *fakeLSP) TypeHierarchySubtypes(
	ctx context.Context, p semanticapi.TypeHierarchySubtypesParams,
) ([]semanticapi.TypeHierarchyItem, error) {
	if err := f.record(ctx, "TypeHierarchySubtypes", p); err != nil || f.empty {
		return nil, err
	}
	return []semanticapi.TypeHierarchyItem{
		{Name: "MyReader", Kind: semanticapi.SymbolKindStruct, Detail: "struct",
			URI: "file:///src/reader.go", Range: rng(10, 0, 15, 1)},
		{Name: "BufferedReader", Kind: semanticapi.SymbolKindStruct,
			URI: "file:///src/buffered.go", Range: rng(8, 0, 12, 1)},
	}, nil
}

func (f *fakeLSP) ExecuteCommand(
	ctx context.Context,
	p semanticapi.ExecuteCommandParams,
) (string, error) {
	if err := f.record(ctx, "ExecuteCommand", p); err != nil || f.empty {
		return "", err
	}
	return "ran " + p.Command, nil
}

// fakeParser implements the syntaxapi.Parser methods the syntax tools call. Its iterators
// record the contexts they are read with and whether they were closed.
type fakeParser struct {
	syntaxapi.Parser
	recorder

	// index holds the matches of ResolveSymbol, by name.
	index map[string][]syntaxapi.Match
	// iterErr makes each iterator fail after its first result.
	iterErr error
	empty   bool
	closed  int
}

// searchParams are the arguments of Search and SearchNode.
type searchParams struct {
	query     string
	nodeTypes syntaxapi.NodeCaptureName
	uri       string
	captures  []string
	languages []string
}

func (f *fakeParser) Search(
	query string, captures []string, languages ...string,
) (iterator.Iterator[syntaxapi.Result], error) {
	p := searchParams{query: query, captures: captures, languages: languages}
	if err := f.recordNoContext("Search", p); err != nil {
		return nil, err
	}
	return f.results("file:///src/main.go", "func main()", "fn"), nil
}

func (f *fakeParser) SearchNode(
	nodeTypes syntaxapi.NodeCaptureName, languages ...string,
) (iterator.Iterator[syntaxapi.Result], error) {
	p := searchParams{nodeTypes: nodeTypes, languages: languages}
	if err := f.recordNoContext("SearchNode", p); err != nil {
		return nil, err
	}
	return f.results("file:///src/main.go", "MyFunc", "definition.func"), nil
}

func (f *fakeParser) Query(
	uri workspaceapi.URI, query string, captures []string,
) (iterator.Iterator[syntaxapi.Result], error) {
	p := searchParams{query: query, uri: uri.String(), captures: captures}
	if err := f.recordNoContext("Query", p); err != nil {
		return nil, err
	}
	return f.results(uri.String(), "package main", "pkg"), nil
}

func (f *fakeParser) QueryNode(
	uri workspaceapi.URI, nodeTypes syntaxapi.NodeCaptureName,
) (iterator.Iterator[syntaxapi.Result], error) {
	p := searchParams{nodeTypes: nodeTypes, uri: uri.String()}
	if err := f.recordNoContext("QueryNode", p); err != nil {
		return nil, err
	}
	return f.results(uri.String(), "main", "definition.namespace"), nil
}

// ResolveSymbol answers from index. Like Rune's symbol index, it refuses a name without a
// dot. Each read of its iterator is recorded as NextMatch.
func (f *fakeParser) ResolveSymbol(
	ctx context.Context, name string, _ syntaxapi.Progress,
) (iterator.Iterator[syntaxapi.Match], error) {
	if err := f.record(ctx, "ResolveSymbol", name); err != nil {
		return nil, err
	}
	if !strings.Contains(name, ".") {
		return nil, syntaxapi.ErrNoDot
	}
	matches := f.index[name]
	i := 0
	return iterator.FromFunc(
		func(ctx context.Context) (syntaxapi.Match, bool, error) {
			_ = f.record(ctx, "NextMatch", nil)
			if f.iterErr != nil && i == 1 {
				return syntaxapi.Match{}, false, f.iterErr
			}
			if i >= len(matches) {
				return syntaxapi.Match{}, false, nil
			}
			i++
			return matches[i-1], true, nil
		},
		f.close,
	), nil
}

func (f *fakeParser) close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed++
	return nil
}

// results returns an iterator over two results in file: text on line 5, then a second
// match on line 9.
func (f *fakeParser) results(
	file, text, capture string,
) iterator.Iterator[syntaxapi.Result] {
	uri, err := workspaceapi.ParseURI(file)
	if err != nil {
		panic(err)
	}
	results := []syntaxapi.Result{
		{File: uri, Text: text, From: term.Coordinates{X: 0, Y: 5},
			To: term.Coordinates{X: len(text), Y: 5}, CaptureName: capture},
		{File: uri, Text: text, From: term.Coordinates{X: 2, Y: 9},
			To: term.Coordinates{X: 2 + len(text), Y: 9}, CaptureName: capture},
	}
	if f.empty {
		results = nil
	}
	i := 0
	return iterator.FromFunc(
		func(ctx context.Context) (syntaxapi.Result, bool, error) {
			_ = f.record(ctx, "Next", nil)
			if f.iterErr != nil && i == 1 {
				return syntaxapi.Result{}, false, f.iterErr
			}
			if i >= len(results) {
				return syntaxapi.Result{}, false, nil
			}
			i++
			return results[i-1], true, nil
		},
		f.close,
	)
}
