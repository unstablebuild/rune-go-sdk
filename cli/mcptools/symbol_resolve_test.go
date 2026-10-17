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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi"
	"github.com/unstablebuild/rune-go-sdk/term"
)

const (
	// reviewGo references term.Writer, which structuresGo, in a dependency, defines.
	reviewGo     = "file:///work/internal/difftab/review.go"
	structuresGo = "file:///mod/rune-go-sdk/term/structures.go"
)

// match is a symbol index match at line and char of uri.
func match(uri string, line, char int) syntaxapi.Match {
	return syntaxapi.Match{URI: uri, Pos: term.Coordinates{X: char, Y: line}}
}

func TestResolveSymbolIndexFirst(t *testing.T) {
	lsp := newFakeLSP()
	parser := &fakeParser{index: map[string][]syntaxapi.Match{
		"term.Writer": {match(reviewGo, 41, 7), match(reviewGo, 41, 7), match(mainGo, 3, 2)},
	}}
	got, err := ResolveSymbol(t.Context(), lsp, parser, "term.Writer", HintDefault)
	require.NoError(t, err)
	require.Equal(t, Resolution{Sites: []SymbolPosition{
		{URI: reviewGo, Position: pos(41, 7)},
		{URI: mainGo, Position: pos(3, 2)},
	}}, got, "each site once, in the index's order, and no hint")
	require.Empty(t, lsp.methods(), "the language server is not searched")
	require.Equal(t, 1, parser.closed)

	best, err := ResolveSymbolBest(t.Context(), lsp, parser, "term.Writer", HintDefault)
	require.NoError(t, err)
	require.Equal(t, got.Sites[0], best)
}

// lines returns a match at column 0 of each of the first n lines of mainGo.
func lines(n int) []syntaxapi.Match {
	var matches []syntaxapi.Match
	for line := range n {
		matches = append(matches, match(mainGo, line, 0))
	}
	return matches
}

func TestResolveSymbolIndexCapped(t *testing.T) {
	tests := []struct {
		name      string
		matches   []syntaxapi.Match
		wantSites int
		wantReads int
		wantHint  bool
	}{
		{name: "under the cap", matches: lines(MaxSites - 1), wantSites: MaxSites - 1,
			wantReads: MaxSites},
		{name: "at the cap", matches: lines(MaxSites), wantSites: MaxSites,
			wantReads: MaxSites + 1},
		{name: "at the cap with repeats", wantSites: MaxSites, wantReads: 2*MaxSites + 1,
			matches: append(lines(MaxSites), lines(MaxSites)...)},
		{name: "over the cap", matches: lines(MaxSites + 4), wantSites: MaxSites,
			wantReads: MaxSites + 1, wantHint: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &fakeParser{index: map[string][]syntaxapi.Match{"pkg.Many": tt.matches}}
			got, err := ResolveSymbol(t.Context(), newFakeLSP(), parser, "pkg.Many",
				HintDefault)
			require.NoError(t, err)
			require.Len(t, got.Sites, tt.wantSites)
			require.Equal(t, SymbolPosition{URI: mainGo, Position: pos(7, 0)}, got.Sites[7])
			require.Len(t, parser.called("NextMatch"), tt.wantReads,
				"reads one site past the cap at most")
			require.Equal(t, 1, parser.closed)
			if tt.wantHint {
				require.Equal(t, `Only the first 16 places that define or reference `+
					`"pkg.Many" were looked at; there are more, so this answer may be `+
					`incomplete. Qualify the name further, as pkg.Type.Method, to narrow it.`,
					got.Hint)
			} else {
				require.Empty(t, got.Hint)
			}
		})
	}
}

// An index that fails partway is not trusted: the workspace's symbols answer instead.
func TestResolveSymbolIndexReadError(t *testing.T) {
	lsp := newFakeLSP()
	parser := &fakeParser{
		iterErr: errors.New("index closed"),
		index: map[string][]syntaxapi.Match{
			"MyStruct.MyMethod": {match(reviewGo, 41, 7), match(mainGo, 3, 2)},
		},
	}
	got, err := ResolveSymbol(t.Context(), lsp, parser, "MyStruct.MyMethod", HintDefault)
	require.NoError(t, err)
	require.Equal(t, []SymbolPosition{{URI: typesGo, Position: pos(30, 0)}}, got.Sites)
	require.Equal(t, []string{"WorkspaceSymbol"}, lsp.methods())
	require.Equal(t, 1, parser.closed)
}

// A name the index cannot resolve falls back to the language server's workspace symbols.
func TestResolveSymbolIndexFallback(t *testing.T) {
	failing := &fakeParser{}
	failing.failOn("ResolveSymbol", errors.New("index not ready"))
	method := []SymbolPosition{{URI: typesGo, Position: pos(30, 0)}}
	tests := []struct {
		name     string
		parser   syntaxapi.Parser
		symbol   string
		want     []SymbolPosition
		wantHint string
	}{
		{name: "no dot", parser: &fakeParser{}, symbol: "MyFunc",
			want:     []SymbolPosition{{URI: mainGo, Position: pos(5, 0)}},
			wantHint: bareHint("MyFunc")},
		{name: "no match", parser: &fakeParser{}, symbol: "MyStruct.MyMethod", want: method},
		{name: "index fails", parser: failing, symbol: "MyStruct.MyMethod", want: method},
		{name: "no parser", symbol: "MyStruct.MyMethod", want: method},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lsp := newFakeLSP()
			got, err := ResolveSymbol(t.Context(), lsp, tt.parser, tt.symbol, HintDefault)
			require.NoError(t, err)
			require.Equal(t, Resolution{Sites: tt.want, Hint: tt.wantHint}, got)
			require.Equal(t, []string{"WorkspaceSymbol"}, lsp.methods())
		})
	}
}

func TestResolveSymbolWorkspaceCapped(t *testing.T) {
	lsp := newFakeLSP()
	lsp.symbols = nil
	for line := range MaxSites + 3 {
		lsp.symbols = append(lsp.symbols, semanticapi.SymbolInformation{
			Name: "Read", Kind: semanticapi.SymbolKindFunction,
			Location: loc(mainGo, uint32(line), 0, uint32(line), 4),
		})
	}
	got, err := ResolveSymbol(t.Context(), lsp, nil, "Read", HintDefault)
	require.NoError(t, err)
	require.Len(t, got.Sites, MaxSites)
	require.Contains(t, got.Hint, "Only the first 16 places")
	require.Contains(t, got.Hint, bareHint("Read"), "both hints")
}

func TestResolveSymbol(t *testing.T) {
	notFound := func(symbol string) string {
		return `symbol "` + symbol + `" not found. ` + bareHint(symbol)
	}
	tests := []struct {
		name    string
		symbol  string
		want    Resolution
		wantErr string
	}{
		{name: "function", symbol: "MyFunc", want: Resolution{
			Sites: []SymbolPosition{{URI: mainGo, Position: pos(5, 0)}},
			Hint:  bareHint("MyFunc")}},
		{name: "interface", symbol: "MyInterface", want: Resolution{
			Sites: []SymbolPosition{{URI: typesGo, Position: pos(15, 5)}},
			Hint:  bareHint("MyInterface")}},
		{name: "variable", symbol: "myVar", want: Resolution{
			Sites: []SymbolPosition{{URI: mainGo, Position: pos(3, 4)}},
			Hint:  bareHint("myVar")}},
		{name: "method by its own name", symbol: "MyMethod", want: Resolution{
			Sites: []SymbolPosition{{URI: typesGo, Position: pos(30, 0)}},
			Hint:  bareHint("MyMethod")}},
		{name: "method qualified", symbol: "MyStruct.MyMethod", want: Resolution{
			Sites: []SymbolPosition{{URI: typesGo, Position: pos(30, 0)}}}},
		{name: "prefix", symbol: "My", wantErr: notFound("My")},
		{name: "case differs", symbol: "myfunc", wantErr: notFound("myfunc")},
		{name: "suffix without a dot", symbol: "Func", wantErr: notFound("Func")},
		{name: "missing", symbol: "Nope", wantErr: notFound("Nope")},
		{name: "missing qualified", symbol: "pkg.Nope", wantErr: `symbol "pkg.Nope" not found`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lsp := newFakeLSP()
			got, err := ResolveSymbol(t.Context(), lsp, nil, tt.symbol, HintDefault)
			require.Equal(t, semanticapi.WorkspaceSymbolParams{Query: tt.symbol},
				lsp.called("WorkspaceSymbol")[0].params)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestResolveSymbolHints(t *testing.T) {
	at := func(uri string) SymbolPosition {
		return SymbolPosition{URI: "file:///" + uri}
	}
	type symbol = semanticapi.SymbolInformation
	sym := func(name string, kind semanticapi.SymbolKind, uri string) symbol {
		return symbol{Name: name, Kind: kind, Location: loc("file:///"+uri, 0, 0, 0, 1)}
	}
	lsp := newFakeLSP()
	lsp.symbols = []semanticapi.SymbolInformation{
		sym("Reader", semanticapi.SymbolKindVariable, "var"),
		sym("Reader", semanticapi.SymbolKindFunction, "func"),
		sym("Reader", semanticapi.SymbolKindInterface, "iface"),
		sym("Reader", semanticapi.SymbolKindStruct, "struct"),
		sym("Reader", semanticapi.SymbolKindClass, "class"),
		sym("Reader", semanticapi.SymbolKindField, "field"),
		sym("T.Reader", semanticapi.SymbolKindMethod, "method"),
		sym("Reader", semanticapi.SymbolKindConstant, "const"),
		sym("Reader", semanticapi.SymbolKindProperty, "prop"),
		// Not exact matches.
		sym("MyReader", semanticapi.SymbolKindStruct, "prefixed"),
		sym("Reader.Close", semanticapi.SymbolKindMethod, "member"),
	}
	// Preferred kinds come first, each group in the order the language server listed it.
	tests := map[SymbolResolveHint][]string{
		HintDefault: {"func", "iface", "struct", "class", "method", "var",
			"field", "const", "prop"},
		HintImplementation: {"iface", "class", "var", "func", "struct", "field",
			"method", "const", "prop"},
		HintTypeDefinition: {"var", "field", "const", "prop", "func", "iface",
			"struct", "class", "method"},
		HintCallHierarchy: {"func", "method", "var", "iface", "struct", "class",
			"field", "const", "prop"},
		HintTypeHierarchy: {"iface", "struct", "class", "var", "func", "field",
			"method", "const", "prop"},
		SymbolResolveHint(999): {"func", "iface", "struct", "class", "method", "var",
			"field", "const", "prop"},
	}
	for hint, order := range tests {
		want := make([]SymbolPosition, len(order))
		for i, uri := range order {
			want[i] = at(uri)
		}
		got, err := ResolveSymbol(t.Context(), lsp, nil, "Reader", hint)
		require.NoError(t, err)
		require.Equal(t, want, got.Sites, "hint %d", hint)

		best, err := ResolveSymbolBest(t.Context(), lsp, nil, "Reader", hint)
		require.NoError(t, err)
		require.Equal(t, want[0], best, "hint %d", hint)
	}
}

func TestResolveSymbolDeduplicates(t *testing.T) {
	lsp := newFakeLSP()
	same := loc("file:///dup.go", 4, 2, 4, 9)
	lsp.symbols = []semanticapi.SymbolInformation{
		{Name: "Dup", Kind: semanticapi.SymbolKindVariable, Location: same},
		{Name: "Dup", Kind: semanticapi.SymbolKindFunction, Location: same},
		// Same start, different end: still the same position.
		{Name: "Dup", Kind: semanticapi.SymbolKindFunction,
			Location: loc("file:///dup.go", 4, 2, 5, 0)},
		{Name: "Dup", Kind: semanticapi.SymbolKindFunction,
			Location: loc("file:///dup.go", 4, 3, 4, 9)},
	}
	got, err := ResolveSymbol(t.Context(), lsp, nil, "Dup", HintDefault)
	require.NoError(t, err)
	require.Equal(t, []SymbolPosition{
		{URI: "file:///dup.go", Position: pos(4, 2)},
		{URI: "file:///dup.go", Position: pos(4, 3)},
	}, got.Sites)
}

func TestResolveSymbolLookupError(t *testing.T) {
	lsp := newFakeLSP()
	boom := errors.New("server not ready")
	lsp.failOn("WorkspaceSymbol", boom)
	_, err := ResolveSymbol(t.Context(), lsp, nil, "MyFunc", HintDefault)
	require.ErrorIs(t, err, boom)
	_, err = ResolveSymbolBest(t.Context(), lsp, nil, "MyFunc", HintDefault)
	require.ErrorIs(t, err, boom)
}

func TestResolveSymbolBestNotFound(t *testing.T) {
	_, err := ResolveSymbolBest(t.Context(), newFakeLSP(), nil, "Nope", HintDefault)
	require.EqualError(t, err, `symbol "Nope" not found. `+bareHint("Nope"))
}

func TestDedup(t *testing.T) {
	type key struct {
		URI  string
		Line uint32
	}
	byKey := func(p SymbolPosition) key { return key{URI: p.URI, Line: p.Position.Line} }
	items := []SymbolPosition{
		{URI: "a", Position: pos(1, 0)},
		{URI: "a", Position: pos(1, 7)},
		{URI: "b", Position: pos(2, 0)},
		{URI: "a", Position: pos(1, 0)},
	}
	require.Equal(t,
		[]SymbolPosition{{URI: "a", Position: pos(1, 0)},
			{URI: "b", Position: pos(2, 0)}},
		dedup(items, byKey), "first occurrence kept, order preserved")
	require.Equal(t, []SymbolPosition{}, dedup(nil, byKey))
}
