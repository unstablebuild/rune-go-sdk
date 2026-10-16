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
	"fmt"
	"strconv"
	"strings"

	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi"
)

// FlatLocation is the JSON shape of a location in tool output.
type FlatLocation struct {
	URI       string `json:"uri"`
	StartLine uint32 `json:"start_line"`
	StartChar uint32 `json:"start_char"`
	EndLine   uint32 `json:"end_line"`
	EndChar   uint32 `json:"end_char"`
}

// ToFlatLocations converts locs to their tool output shape.
func ToFlatLocations(locs []semanticapi.Location) []FlatLocation {
	out := make([]FlatLocation, len(locs))
	for i, l := range locs {
		out[i] = FlatLocation{
			URI:       l.URI,
			StartLine: l.Range.Start.Line,
			StartChar: l.Range.Start.Character,
			EndLine:   l.Range.End.Line,
			EndChar:   l.Range.End.Character,
		}
	}
	return out
}

// ToFlatLocationsFromResult converts any form of a location result to its tool output
// shape.
func ToFlatLocationsFromResult(result semanticapi.LocationResult) []FlatLocation {
	var locs []semanticapi.Location
	if result.Location != nil {
		locs = []semanticapi.Location{*result.Location}
	} else if len(result.Locations) > 0 {
		locs = result.Locations
	} else if len(result.LocationLinks) > 0 {
		out := make([]FlatLocation, len(result.LocationLinks))
		for i, ll := range result.LocationLinks {
			out[i] = FlatLocation{
				URI:       ll.TargetURI,
				StartLine: ll.TargetRange.Start.Line,
				StartChar: ll.TargetRange.Start.Character,
				EndLine:   ll.TargetRange.End.Line,
				EndChar:   ll.TargetRange.End.Character,
			}
		}
		return out
	}
	return ToFlatLocations(locs)
}

// SymbolKindString returns the name of k.
func SymbolKindString(k semanticapi.SymbolKind) string {
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

// SeverityString returns the name of s.
func SeverityString(s semanticapi.DiagnosticSeverity) string {
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

// CompletionItemKindString returns the name of k.
func CompletionItemKindString(k semanticapi.CompletionItemKind) string {
	switch k {
	case semanticapi.CompletionItemKindText:
		return "Text"
	case semanticapi.CompletionItemKindMethod:
		return "Method"
	case semanticapi.CompletionItemKindFunction:
		return "Function"
	case semanticapi.CompletionItemKindConstructor:
		return "Constructor"
	case semanticapi.CompletionItemKindField:
		return "Field"
	case semanticapi.CompletionItemKindVariable:
		return "Variable"
	case semanticapi.CompletionItemKindClass:
		return "Class"
	case semanticapi.CompletionItemKindInterface:
		return "Interface"
	case semanticapi.CompletionItemKindModule:
		return "Module"
	case semanticapi.CompletionItemKindProperty:
		return "Property"
	case semanticapi.CompletionItemKindUnit:
		return "Unit"
	case semanticapi.CompletionItemKindValue:
		return "Value"
	case semanticapi.CompletionItemKindEnum:
		return "Enum"
	case semanticapi.CompletionItemKindKeyword:
		return "Keyword"
	case semanticapi.CompletionItemKindSnippet:
		return "Snippet"
	case semanticapi.CompletionItemKindColor:
		return "Color"
	case semanticapi.CompletionItemKindFile:
		return "File"
	case semanticapi.CompletionItemKindReference:
		return "Reference"
	case semanticapi.CompletionItemKindFolder:
		return "Folder"
	case semanticapi.CompletionItemKindEnumMember:
		return "EnumMember"
	case semanticapi.CompletionItemKindConstant:
		return "Constant"
	case semanticapi.CompletionItemKindStruct:
		return "Struct"
	case semanticapi.CompletionItemKindEvent:
		return "Event"
	case semanticapi.CompletionItemKindOperator:
		return "Operator"
	case semanticapi.CompletionItemKindTypeParameter:
		return "TypeParameter"
	default:
		return fmt.Sprintf("Unknown(%d)", k)
	}
}

// FlatSymbol is the JSON shape of a document symbol in tool output.
type FlatSymbol struct {
	Name      string `json:"name"`
	Detail    string `json:"detail"`
	Kind      string `json:"kind"`
	Depth     int    `json:"depth"`
	StartLine uint32 `json:"start_line"`
	StartChar uint32 `json:"start_char"`
	EndLine   uint32 `json:"end_line"`
	EndChar   uint32 `json:"end_char"`
}

// FlattenSymbols flattens a document symbol tree, recording each symbol's nesting depth
// starting at depth.
func FlattenSymbols(syms []semanticapi.DocumentSymbol, depth int) []FlatSymbol {
	out := make([]FlatSymbol, 0, len(syms))
	for _, s := range syms {
		out = append(out, FlatSymbol{
			Name:      s.Name,
			Detail:    s.Detail,
			Kind:      SymbolKindString(s.Kind),
			Depth:     depth,
			StartLine: s.Range.Start.Line,
			StartChar: s.Range.Start.Character,
			EndLine:   s.Range.End.Line,
			EndChar:   s.Range.End.Character,
		})
		out = append(out, FlattenSymbols(s.Children, depth+1)...)
	}
	return out
}

// FlattenSymbolResult flattens either form of a document symbol result.
func FlattenSymbolResult(result semanticapi.DocumentSymbolResult) []FlatSymbol {
	if len(result.DocumentSymbols) > 0 {
		return FlattenSymbols(result.DocumentSymbols, 0)
	}
	// Convert SymbolInformation to flat format
	out := make([]FlatSymbol, 0, len(result.SymbolInformation))
	for _, si := range result.SymbolInformation {
		out = append(out, FlatSymbol{
			Name:      si.Name,
			Kind:      SymbolKindString(si.Kind),
			Depth:     0,
			StartLine: si.Location.Range.Start.Line,
			StartChar: si.Location.Range.Start.Character,
			EndLine:   si.Location.Range.End.Line,
			EndChar:   si.Location.Range.End.Character,
		})
	}
	return out
}

// FlatWorkspaceSymbol is the JSON shape of a workspace symbol in tool output.
type FlatWorkspaceSymbol struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	URI       string `json:"uri"`
	StartLine uint32 `json:"start_line"`
	StartChar uint32 `json:"start_char"`
}

// FlatDiagnostic is the JSON shape of a diagnostic in tool output.
type FlatDiagnostic struct {
	Severity  string `json:"severity"`
	StartLine uint32 `json:"start_line"`
	StartChar uint32 `json:"start_char"`
	EndLine   uint32 `json:"end_line"`
	EndChar   uint32 `json:"end_char"`
	Message   string `json:"message"`
	Source    string `json:"source"`
	Code      string `json:"code"`
}

// FlatWorkspaceDiagnostic is the JSON shape of a workspace diagnostic in tool output.
type FlatWorkspaceDiagnostic struct {
	URI       string `json:"uri"`
	Severity  string `json:"severity"`
	StartLine uint32 `json:"start_line"`
	StartChar uint32 `json:"start_char"`
	EndLine   uint32 `json:"end_line"`
	EndChar   uint32 `json:"end_char"`
	Message   string `json:"message"`
	Source    string `json:"source"`
	Code      string `json:"code"`
}

// diagnosticCode returns d's code, which language servers send as either a string or an
// integer.
func diagnosticCode(d semanticapi.Diagnostic) string {
	if d.CodeIsInt {
		return strconv.Itoa(d.CodeInt)
	}
	return d.Code
}

// FlatCodeAction is the JSON shape of a code action in tool output.
type FlatCodeAction struct {
	Title string `json:"title"`
	Kind  string `json:"kind"`
}

// FlatCompletion is the JSON shape of a completion item in tool output.
type FlatCompletion struct {
	Label  string `json:"label"`
	Kind   string `json:"kind"`
	Detail string `json:"detail"`
}

// FlatSignatureHelp is the JSON shape of signature help in tool output.
type FlatSignatureHelp struct {
	Label           string `json:"label"`
	ActiveSignature uint32 `json:"active_signature"`
	ActiveParameter uint32 `json:"active_parameter"`
	Parameters      string `json:"parameters"`
}

// FlatTextEdit is the JSON shape of a text edit in tool output.
type FlatTextEdit struct {
	StartLine uint32 `json:"start_line"`
	StartChar uint32 `json:"start_char"`
	EndLine   uint32 `json:"end_line"`
	EndChar   uint32 `json:"end_char"`
	NewText   string `json:"new_text"`
}

// FlatPrepareRename is the JSON shape of a prepare-rename result in tool output.
type FlatPrepareRename struct {
	StartLine   uint32 `json:"start_line"`
	StartChar   uint32 `json:"start_char"`
	EndLine     uint32 `json:"end_line"`
	EndChar     uint32 `json:"end_char"`
	Placeholder string `json:"placeholder"`
}

// DocumentHighlightKindString returns the name of k.
func DocumentHighlightKindString(k semanticapi.DocumentHighlightKind) string {
	switch k {
	case semanticapi.DocumentHighlightKindText:
		return "Text"
	case semanticapi.DocumentHighlightKindRead:
		return "Read"
	case semanticapi.DocumentHighlightKindWrite:
		return "Write"
	default:
		return fmt.Sprintf("Unknown(%d)", k)
	}
}

// FlatDocumentHighlight is the JSON shape of a document highlight in tool output.
type FlatDocumentHighlight struct {
	Kind      string `json:"kind"`
	StartLine uint32 `json:"start_line"`
	StartChar uint32 `json:"start_char"`
	EndLine   uint32 `json:"end_line"`
	EndChar   uint32 `json:"end_char"`
}

// FlatCodeLens is the JSON shape of a code lens in tool output.
type FlatCodeLens struct {
	StartLine uint32 `json:"start_line"`
	StartChar uint32 `json:"start_char"`
	EndLine   uint32 `json:"end_line"`
	EndChar   uint32 `json:"end_char"`
	Title     string `json:"title"`
	Command   string `json:"command"`
}

// FoldingRangeKindString returns the name of k.
func FoldingRangeKindString(k semanticapi.FoldingRangeKind) string {
	switch k {
	case semanticapi.FoldingRangeKindComment:
		return "comment"
	case semanticapi.FoldingRangeKindImports:
		return "imports"
	case semanticapi.FoldingRangeKindRegion:
		return "region"
	default:
		return string(k)
	}
}

// FlatFoldingRange is the JSON shape of a folding range in tool output.
type FlatFoldingRange struct {
	StartLine uint32 `json:"start_line"`
	StartChar uint32 `json:"start_char"`
	EndLine   uint32 `json:"end_line"`
	EndChar   uint32 `json:"end_char"`
	Kind      string `json:"kind"`
}

// FlatSelectionRange is the JSON shape of a selection range in tool output.
type FlatSelectionRange struct {
	StartLine uint32 `json:"start_line"`
	StartChar uint32 `json:"start_char"`
	EndLine   uint32 `json:"end_line"`
	EndChar   uint32 `json:"end_char"`
	Depth     int    `json:"depth"`
}

// FlattenSelectionRanges flattens each selection range and its parents.
func FlattenSelectionRanges(ranges []semanticapi.SelectionRange) []FlatSelectionRange {
	flat := make([]FlatSelectionRange, 0, len(ranges))
	for _, r := range ranges {
		depth := 0
		for current := &r; current != nil; current = current.Parent {
			flat = append(flat, FlatSelectionRange{
				StartLine: current.Range.Start.Line,
				StartChar: current.Range.Start.Character,
				EndLine:   current.Range.End.Line,
				EndChar:   current.Range.End.Character,
				Depth:     depth,
			})
			depth++
		}
	}
	return flat
}

// FlatCallHierarchyItem is the JSON shape of a call hierarchy item in tool output.
type FlatCallHierarchyItem struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	URI       string `json:"uri"`
	StartLine uint32 `json:"start_line"`
	StartChar uint32 `json:"start_char"`
	EndLine   uint32 `json:"end_line"`
	EndChar   uint32 `json:"end_char"`
}

// ToFlatCallHierarchyItem converts item to its tool output shape.
func ToFlatCallHierarchyItem(item semanticapi.CallHierarchyItem) FlatCallHierarchyItem {
	return FlatCallHierarchyItem{
		Name:      item.Name,
		Kind:      SymbolKindString(item.Kind),
		URI:       item.URI,
		StartLine: item.Range.Start.Line,
		StartChar: item.Range.Start.Character,
		EndLine:   item.Range.End.Line,
		EndChar:   item.Range.End.Character,
	}
}

// FlatIncomingCall is the JSON shape of an incoming call in tool output.
type FlatIncomingCall struct {
	FromName      string `json:"from_name"`
	FromKind      string `json:"from_kind"`
	FromURI       string `json:"from_uri"`
	FromStartLine uint32 `json:"from_start_line"`
	FromStartChar uint32 `json:"from_start_char"`
}

// FlatOutgoingCall is the JSON shape of an outgoing call in tool output.
type FlatOutgoingCall struct {
	ToName      string `json:"to_name"`
	ToKind      string `json:"to_kind"`
	ToURI       string `json:"to_uri"`
	ToStartLine uint32 `json:"to_start_line"`
	ToStartChar uint32 `json:"to_start_char"`
}

// ParseSymbolKind parses a name returned by SymbolKindString.
func ParseSymbolKind(s string) semanticapi.SymbolKind {
	switch strings.ToLower(s) {
	case "file":
		return semanticapi.SymbolKindFile
	case "module":
		return semanticapi.SymbolKindModule
	case "namespace":
		return semanticapi.SymbolKindNamespace
	case "package":
		return semanticapi.SymbolKindPackage
	case "class":
		return semanticapi.SymbolKindClass
	case "method":
		return semanticapi.SymbolKindMethod
	case "property":
		return semanticapi.SymbolKindProperty
	case "field":
		return semanticapi.SymbolKindField
	case "constructor":
		return semanticapi.SymbolKindConstructor
	case "enum":
		return semanticapi.SymbolKindEnum
	case "interface":
		return semanticapi.SymbolKindInterface
	case "function":
		return semanticapi.SymbolKindFunction
	case "variable":
		return semanticapi.SymbolKindVariable
	case "constant":
		return semanticapi.SymbolKindConstant
	case "string":
		return semanticapi.SymbolKindString
	case "number":
		return semanticapi.SymbolKindNumber
	case "boolean":
		return semanticapi.SymbolKindBoolean
	case "array":
		return semanticapi.SymbolKindArray
	case "object":
		return semanticapi.SymbolKindObject
	case "key":
		return semanticapi.SymbolKindKey
	case "null":
		return semanticapi.SymbolKindNull
	case "enummember":
		return semanticapi.SymbolKindEnumMember
	case "struct":
		return semanticapi.SymbolKindStruct
	case "event":
		return semanticapi.SymbolKindEvent
	case "operator":
		return semanticapi.SymbolKindOperator
	case "typeparameter":
		return semanticapi.SymbolKindTypeParameter
	default:
		return semanticapi.SymbolKindFunction
	}
}

// FlatTypeHierarchyItem is the JSON shape of a type hierarchy item in tool output.
type FlatTypeHierarchyItem struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	URI       string `json:"uri"`
	Detail    string `json:"detail"`
	StartLine uint32 `json:"start_line"`
	StartChar uint32 `json:"start_char"`
	EndLine   uint32 `json:"end_line"`
	EndChar   uint32 `json:"end_char"`
}

// ToFlatTypeHierarchyItem converts item to its tool output shape.
func ToFlatTypeHierarchyItem(item semanticapi.TypeHierarchyItem) FlatTypeHierarchyItem {
	return FlatTypeHierarchyItem{
		Name:      item.Name,
		Kind:      SymbolKindString(item.Kind),
		URI:       item.URI,
		Detail:    item.Detail,
		StartLine: item.Range.Start.Line,
		StartChar: item.Range.Start.Character,
		EndLine:   item.Range.End.Line,
		EndChar:   item.Range.End.Character,
	}
}

// SearchResult is the JSON shape of a syntax match in tool output.
type SearchResult struct {
	File        string `json:"file"`
	Text        string `json:"text"`
	FromX       int    `json:"from_x"`
	FromY       int    `json:"from_y"`
	ToX         int    `json:"to_x"`
	ToY         int    `json:"to_y"`
	CaptureName string `json:"capture_name"`
}

func parseSingleNodeType(s string) (syntaxapi.NodeCaptureName, error) {
	switch s {
	case "scope":
		return syntaxapi.NodeCaptureScope, nil
	case "namespace":
		return syntaxapi.NodeCaptureDefinitionNamespace, nil
	case "reference":
		return syntaxapi.NodeCaptureReference, nil
	case "func":
		return syntaxapi.NodeCaptureDefinitionFunc, nil
	case "var":
		return syntaxapi.NodeCaptureDefinitionVar, nil
	case "method":
		return syntaxapi.NodeCaptureDefinitionMethod, nil
	case "type":
		return syntaxapi.NodeCaptureDefinitionType, nil
	default:
		return 0, fmt.Errorf(
			"invalid node type %q: use scope|namespace|reference|func|var|method|type", s)
	}
}

// ParseNodeType parses a pipe-separated list of node types such as "func|type".
func ParseNodeType(s string) (syntaxapi.NodeCaptureName, error) {
	var result syntaxapi.NodeCaptureName
	for _, part := range strings.Split(s, "|") {
		nodeType, err := parseSingleNodeType(part)
		if err != nil {
			return 0, err
		}
		result |= nodeType
	}
	return result, nil
}
