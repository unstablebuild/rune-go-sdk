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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi"
)

func TestSymbolKindString(t *testing.T) {
	// In LSP order, from 1.
	names := []string{
		"File", "Module", "Namespace", "Package", "Class", "Method",
		"Property", "Field", "Constructor", "Enum", "Interface",
		"Function", "Variable", "Constant", "String", "Number",
		"Boolean", "Array", "Object", "Key", "Null", "EnumMember",
		"Struct", "Event", "Operator", "TypeParameter",
	}
	for i, name := range names {
		k := semanticapi.SymbolKind(i + 1)
		require.Equal(t, name, SymbolKindString(k))
		require.Equal(t, k, ParseSymbolKind(name), name)
		require.Equal(t, k, ParseSymbolKind(strings.ToLower(name)), name)
		require.Equal(t, k, ParseSymbolKind(strings.ToUpper(name)), name)
	}
	require.Equal(t, "Unknown(0)", SymbolKindString(0))
	require.Equal(t, "Unknown(27)", SymbolKindString(27))
}

func TestParseSymbolKindDefaultsToFunction(t *testing.T) {
	for _, s := range []string{"", "func", "Unknown(27)", "type parameter"} {
		require.Equal(t, semanticapi.SymbolKindFunction, ParseSymbolKind(s), s)
	}
}

func TestSeverityString(t *testing.T) {
	tests := map[semanticapi.DiagnosticSeverity]string{
		0: "Unknown(0)", 1: "Error", 2: "Warning", 3: "Information", 4: "Hint",
		5: "Unknown(5)",
	}
	for s, want := range tests {
		require.Equal(t, want, SeverityString(s))
	}
}

func TestCompletionItemKindString(t *testing.T) {
	names := []string{
		"Text", "Method", "Function", "Constructor", "Field", "Variable",
		"Class", "Interface", "Module", "Property", "Unit", "Value",
		"Enum", "Keyword", "Snippet", "Color", "File", "Reference",
		"Folder", "EnumMember", "Constant", "Struct", "Event",
		"Operator", "TypeParameter",
	}
	for i, name := range names {
		require.Equal(t, name,
			CompletionItemKindString(semanticapi.CompletionItemKind(i+1)))
	}
	require.Equal(t, "Unknown(0)", CompletionItemKindString(0))
	require.Equal(t, "Unknown(26)", CompletionItemKindString(26))
}

func TestDocumentHighlightKindString(t *testing.T) {
	tests := map[semanticapi.DocumentHighlightKind]string{
		0: "Unknown(0)", 1: "Text", 2: "Read", 3: "Write", 4: "Unknown(4)",
	}
	for k, want := range tests {
		require.Equal(t, want, DocumentHighlightKindString(k))
	}
}

func TestFoldingRangeKindString(t *testing.T) {
	tests := map[semanticapi.FoldingRangeKind]string{
		semanticapi.FoldingRangeKindComment: "comment",
		semanticapi.FoldingRangeKindImports: "imports",
		semanticapi.FoldingRangeKindRegion:  "region",
		"custom":                            "custom",
		"":                                  "",
	}
	for k, want := range tests {
		require.Equal(t, want, FoldingRangeKindString(k))
	}
}

func TestToFlatLocations(t *testing.T) {
	require.Equal(t, []FlatLocation{}, ToFlatLocations(nil))
	require.Equal(t,
		[]FlatLocation{
			{URI: "file:///a", StartLine: 1, StartChar: 2, EndLine: 3, EndChar: 4}},
		ToFlatLocations([]semanticapi.Location{loc("file:///a", 1, 2, 3, 4)}))
}

func TestToFlatLocationsFromResult(t *testing.T) {
	single := loc("file:///single", 1, 0, 1, 5)
	many := []semanticapi.Location{loc("file:///a", 2, 0, 2, 3),
		loc("file:///b", 3, 1, 3, 4)}
	links := []semanticapi.LocationLink{{
		TargetURI:            "file:///link",
		TargetRange:          rng(4, 0, 9, 1),
		TargetSelectionRange: rng(4, 5, 4, 8),
	}}
	tests := []struct {
		name   string
		result semanticapi.LocationResult
		want   []FlatLocation
	}{
		{name: "empty", want: []FlatLocation{}},
		{
			name:   "location",
			result: semanticapi.LocationResult{Location: &single},
			want: []FlatLocation{
				{URI: "file:///single", StartLine: 1, EndLine: 1, EndChar: 5}},
		},
		{
			name:   "locations",
			result: semanticapi.LocationResult{Locations: many},
			want: []FlatLocation{
				{URI: "file:///a", StartLine: 2, EndLine: 2, EndChar: 3},
				{URI: "file:///b", StartLine: 3, StartChar: 1, EndLine: 3, EndChar: 4},
			},
		},
		{
			name:   "links use the target range",
			result: semanticapi.LocationResult{LocationLinks: links},
			want: []FlatLocation{
				{URI: "file:///link", StartLine: 4, EndLine: 9, EndChar: 1}},
		},
		{
			name: "location wins",
			result: semanticapi.LocationResult{Location: &single, Locations: many,
				LocationLinks: links},
			want: []FlatLocation{
				{URI: "file:///single", StartLine: 1, EndLine: 1, EndChar: 5}},
		},
		{
			name:   "locations win over links",
			result: semanticapi.LocationResult{Locations: many[:1], LocationLinks: links},
			want: []FlatLocation{
				{URI: "file:///a", StartLine: 2, EndLine: 2, EndChar: 3}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, ToFlatLocationsFromResult(tt.result))
		})
	}
}

func TestFlattenSymbols(t *testing.T) {
	type symbol = semanticapi.DocumentSymbol
	sym := func(name string, children ...symbol) symbol {
		return semanticapi.DocumentSymbol{
			Name: name, Detail: name + "!", Kind: semanticapi.SymbolKindField,
			Range: rng(1, 2, 3, 4), SelectionRange: rng(9, 9, 9, 9), Children: children,
		}
	}
	tree := []semanticapi.DocumentSymbol{sym("a", sym("b", sym("c")), sym("d")), sym("e")}
	flat := func(name string, depth int) FlatSymbol {
		return FlatSymbol{
			Name: name, Detail: name + "!", Kind: "Field", Depth: depth,
			StartLine: 1, StartChar: 2, EndLine: 3, EndChar: 4,
		}
	}
	require.Equal(t,
		[]FlatSymbol{flat("a", 0), flat("b", 1), flat("c", 2), flat("d", 1),
			flat("e", 0)},
		FlattenSymbols(tree, 0))
	require.Equal(t, []FlatSymbol{flat("e", 3)}, FlattenSymbols(tree[1:], 3))
	require.Equal(t, []FlatSymbol{}, FlattenSymbols(nil, 0))
}

func TestFlattenSymbolResult(t *testing.T) {
	doc := []semanticapi.DocumentSymbol{
		{Name: "doc", Kind: semanticapi.SymbolKindClass, Range: rng(0, 0, 9, 0)}}
	info := []semanticapi.SymbolInformation{
		{Name: "info", Kind: semanticapi.SymbolKindConstant,
			Location: loc("file:///x", 5, 1, 5, 7)},
	}
	require.Equal(t, []FlatSymbol{{Name: "doc", Kind: "Class", EndLine: 9}},
		FlattenSymbolResult(
			semanticapi.DocumentSymbolResult{DocumentSymbols: doc,
				SymbolInformation: info}),
		"document symbols win")
	require.Equal(t,
		[]FlatSymbol{{
			Name: "info", Kind: "Constant",
			StartLine: 5, StartChar: 1, EndLine: 5, EndChar: 7,
		}},
		FlattenSymbolResult(semanticapi.DocumentSymbolResult{SymbolInformation: info}))
	require.Equal(t, []FlatSymbol{},
		FlattenSymbolResult(semanticapi.DocumentSymbolResult{}))
}

func TestFlattenSelectionRanges(t *testing.T) {
	ranges := []semanticapi.SelectionRange{
		{
			Range: rng(1, 1, 1, 2),
			Parent: &semanticapi.SelectionRange{
				Range:  rng(1, 0, 2, 0),
				Parent: &semanticapi.SelectionRange{Range: rng(0, 0, 9, 0)},
			},
		},
		{Range: rng(5, 5, 5, 6)},
	}
	require.Equal(t, []FlatSelectionRange{
		{StartLine: 1, StartChar: 1, EndLine: 1, EndChar: 2, Depth: 0},
		{StartLine: 1, StartChar: 0, EndLine: 2, EndChar: 0, Depth: 1},
		{StartLine: 0, StartChar: 0, EndLine: 9, EndChar: 0, Depth: 2},
		{StartLine: 5, StartChar: 5, EndLine: 5, EndChar: 6, Depth: 0},
	}, FlattenSelectionRanges(ranges))
	require.Equal(t, []FlatSelectionRange{}, FlattenSelectionRanges(nil))
}

func TestHierarchyItems(t *testing.T) {
	call := semanticapi.CallHierarchyItem{
		Name: "f", Kind: semanticapi.SymbolKindMethod, URI: "file:///f",
		Range: rng(1, 2, 3, 4), SelectionRange: rng(9, 9, 9, 9),
	}
	require.Equal(t, FlatCallHierarchyItem{
		Name: "f", Kind: "Method", URI: "file:///f",
		StartLine: 1, StartChar: 2, EndLine: 3, EndChar: 4,
	}, ToFlatCallHierarchyItem(call))

	typ := semanticapi.TypeHierarchyItem{
		Name: "T", Kind: semanticapi.SymbolKindStruct, Detail: "struct", URI: "file:///t",
		Range: rng(5, 6, 7, 8), SelectionRange: rng(9, 9, 9, 9),
	}
	require.Equal(t, FlatTypeHierarchyItem{
		Name: "T", Kind: "Struct", URI: "file:///t", Detail: "struct",
		StartLine: 5, StartChar: 6, EndLine: 7, EndChar: 8,
	}, ToFlatTypeHierarchyItem(typ))
}

func TestParseNodeType(t *testing.T) {
	singles := map[string]syntaxapi.NodeCaptureName{
		"scope":     syntaxapi.NodeCaptureScope,
		"namespace": syntaxapi.NodeCaptureDefinitionNamespace,
		"reference": syntaxapi.NodeCaptureReference,
		"func":      syntaxapi.NodeCaptureDefinitionFunc,
		"var":       syntaxapi.NodeCaptureDefinitionVar,
		"method":    syntaxapi.NodeCaptureDefinitionMethod,
		"type":      syntaxapi.NodeCaptureDefinitionType,
	}
	var all syntaxapi.NodeCaptureName
	var names []string
	for name, want := range singles {
		got, err := ParseNodeType(name)
		require.NoError(t, err)
		require.Equal(t, want, got, name)
		require.Zero(t, all&want, "%s shares a bit", name)
		all |= want
		names = append(names, name)
	}
	got, err := ParseNodeType(strings.Join(names, "|"))
	require.NoError(t, err)
	require.Equal(t, all, got)

	got, err = ParseNodeType("func|func")
	require.NoError(t, err)
	require.Equal(t, syntaxapi.NodeCaptureDefinitionFunc, got)

	const valid = "scope|namespace|reference|func|var|method|type"
	for _, bad := range []string{
		"", "|", "func|", "|func", "Func", " func", "func |type", "funcs",
	} {
		_, err := ParseNodeType(bad)
		require.ErrorContains(t, err, "invalid node type", "%q", bad)
		require.ErrorContains(t, err, valid, "%q", bad)
	}
}

func TestDiagnosticCode(t *testing.T) {
	require.Equal(t, "", diagnosticCode(semanticapi.Diagnostic{}))
	require.Equal(t, "E001", diagnosticCode(semanticapi.Diagnostic{Code: "E001"}))
	require.Equal(t, "2304",
		diagnosticCode(semanticapi.Diagnostic{CodeInt: 2304, CodeIsInt: true}))
	require.Equal(t, "0", diagnosticCode(semanticapi.Diagnostic{CodeIsInt: true}))
	require.Equal(t, "-1",
		diagnosticCode(semanticapi.Diagnostic{CodeInt: -1, CodeIsInt: true}))
}

func TestToUint(t *testing.T) {
	for _, f := range []float64{0, 1, 42, 4294967295} {
		got, err := toUint("n", f)
		require.NoError(t, err)
		require.Equal(t, uint32(f), got)
	}
	for _, f := range []float64{-1, -0.5, 0.5, 1e-9, 4294967296, 1e20} {
		_, err := toUint("n", f)
		require.ErrorContains(t, err,
			`argument "n" must be a whole number from 0 to 4294967295`, "%v", f)
	}
}
