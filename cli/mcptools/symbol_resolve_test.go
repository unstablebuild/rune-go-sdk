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
)

func TestResolveSymbol(t *testing.T) {
	tests := []struct {
		name    string
		symbol  string
		want    []SymbolPosition
		wantErr string
	}{
		{name: "function", symbol: "MyFunc",
			want: []SymbolPosition{{URI: mainGo, Position: pos(5, 0)}}},
		{name: "interface", symbol: "MyInterface",
			want: []SymbolPosition{{URI: typesGo, Position: pos(15, 5)}}},
		{name: "variable", symbol: "myVar",
			want: []SymbolPosition{{URI: mainGo, Position: pos(3, 4)}}},
		{name: "method by its own name", symbol: "MyMethod", want: []SymbolPosition{
			{URI: typesGo, Position: pos(30, 0)}}},
		{name: "method qualified", symbol: "MyStruct.MyMethod", want: []SymbolPosition{
			{URI: typesGo, Position: pos(30, 0)}}},
		{name: "prefix", symbol: "My", wantErr: `symbol "My" not found`},
		{name: "case differs", symbol: "myfunc", wantErr: `symbol "myfunc" not found`},
		{name: "suffix without a dot", symbol: "Func",
			wantErr: `symbol "Func" not found`},
		{name: "missing", symbol: "Nope", wantErr: `symbol "Nope" not found`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lsp := newFakeLSP()
			got, err := ResolveSymbol(t.Context(), lsp, tt.symbol, HintDefault)
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
		got, err := ResolveSymbol(t.Context(), lsp, "Reader", hint)
		require.NoError(t, err)
		require.Equal(t, want, got, "hint %d", hint)

		best, err := ResolveSymbolBest(t.Context(), lsp, "Reader", hint)
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
	got, err := ResolveSymbol(t.Context(), lsp, "Dup", HintDefault)
	require.NoError(t, err)
	require.Equal(t, []SymbolPosition{
		{URI: "file:///dup.go", Position: pos(4, 2)},
		{URI: "file:///dup.go", Position: pos(4, 3)},
	}, got)
}

func TestResolveSymbolLookupError(t *testing.T) {
	lsp := newFakeLSP()
	boom := errors.New("server not ready")
	lsp.failOn("WorkspaceSymbol", boom)
	_, err := ResolveSymbol(t.Context(), lsp, "MyFunc", HintDefault)
	require.ErrorIs(t, err, boom)
	_, err = ResolveSymbolBest(t.Context(), lsp, "MyFunc", HintDefault)
	require.ErrorIs(t, err, boom)
}

func TestResolveSymbolBestNotFound(t *testing.T) {
	_, err := ResolveSymbolBest(t.Context(), newFakeLSP(), "Nope", HintDefault)
	require.EqualError(t, err, `symbol "Nope" not found`)
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
