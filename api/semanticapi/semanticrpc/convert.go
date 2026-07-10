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

package semanticrpc

import (
	"encoding/json"

	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
)

func PositionToProto(p semanticapi.Position) *Position {
	return &Position{Line: p.Line, Character: p.Character}
}

func PositionFromProto(p *Position) semanticapi.Position {
	if p == nil {
		return semanticapi.Position{}
	}
	return semanticapi.Position{Line: p.Line, Character: p.Character}
}

func RangeToProto(r semanticapi.Range) *Range {
	return &Range{
		Start: PositionToProto(r.Start),
		End:   PositionToProto(r.End),
	}
}

func RangeFromProto(r *Range) semanticapi.Range {
	if r == nil {
		return semanticapi.Range{}
	}
	return semanticapi.Range{
		Start: PositionFromProto(r.Start),
		End:   PositionFromProto(r.End),
	}
}

func RangePtrFromProto(r *Range) *semanticapi.Range {
	if r == nil {
		return nil
	}
	rng := RangeFromProto(r)
	return &rng
}

func LocationToProto(l semanticapi.Location) *Location {
	return &Location{Uri: l.URI, Range: RangeToProto(l.Range)}
}

func LocationFromProto(l *Location) semanticapi.Location {
	if l == nil {
		return semanticapi.Location{}
	}
	return semanticapi.Location{URI: l.Uri, Range: RangeFromProto(l.Range)}
}

func LocationsToProto(locs []semanticapi.Location) []*Location {
	out := make([]*Location, len(locs))
	for i, l := range locs {
		out[i] = LocationToProto(l)
	}
	return out
}

func LocationsFromProto(locs []*Location) []semanticapi.Location {
	out := make([]semanticapi.Location, len(locs))
	for i, l := range locs {
		out[i] = LocationFromProto(l)
	}
	return out
}

func TextEditToProto(e semanticapi.TextEdit) *TextEdit {
	return &TextEdit{Range: RangeToProto(e.Range), NewText: e.NewText}
}

func TextEditFromProto(e *TextEdit) semanticapi.TextEdit {
	if e == nil {
		return semanticapi.TextEdit{}
	}
	return semanticapi.TextEdit{Range: RangeFromProto(e.Range), NewText: e.NewText}
}

func TextEditsToProto(edits []semanticapi.TextEdit) []*TextEdit {
	out := make([]*TextEdit, len(edits))
	for i, e := range edits {
		out[i] = TextEditToProto(e)
	}
	return out
}

func TextEditsFromProto(edits []*TextEdit) []semanticapi.TextEdit {
	out := make([]semanticapi.TextEdit, len(edits))
	for i, e := range edits {
		out[i] = TextEditFromProto(e)
	}
	return out
}

func AnnotatedTextEditToProto(e semanticapi.AnnotatedTextEdit) *AnnotatedTextEdit {
	return &AnnotatedTextEdit{
		Range:        RangeToProto(e.Range),
		NewText:      e.NewText,
		AnnotationId: e.AnnotationID,
	}
}

func AnnotatedTextEditFromProto(e *AnnotatedTextEdit) semanticapi.AnnotatedTextEdit {
	if e == nil {
		return semanticapi.AnnotatedTextEdit{}
	}
	return semanticapi.AnnotatedTextEdit{
		Range:        RangeFromProto(e.Range),
		NewText:      e.NewText,
		AnnotationID: e.AnnotationId,
	}
}

func AnnotatedTextEditsToProto(edits []semanticapi.AnnotatedTextEdit) []*AnnotatedTextEdit {
	out := make([]*AnnotatedTextEdit, len(edits))
	for i, e := range edits {
		out[i] = AnnotatedTextEditToProto(e)
	}
	return out
}

func AnnotatedTextEditsFromProto(edits []*AnnotatedTextEdit) []semanticapi.AnnotatedTextEdit {
	out := make([]semanticapi.AnnotatedTextEdit, len(edits))
	for i, e := range edits {
		out[i] = AnnotatedTextEditFromProto(e)
	}
	return out
}

func InsertReplaceEditToProto(e *semanticapi.InsertReplaceEdit) *InsertReplaceEdit {
	if e == nil {
		return nil
	}
	return &InsertReplaceEdit{
		NewText: e.NewText,
		Insert:  RangeToProto(e.Insert),
		Replace: RangeToProto(e.Replace),
	}
}

func InsertReplaceEditFromProto(e *InsertReplaceEdit) *semanticapi.InsertReplaceEdit {
	if e == nil {
		return nil
	}
	return &semanticapi.InsertReplaceEdit{
		NewText: e.NewText,
		Insert:  RangeFromProto(e.Insert),
		Replace: RangeFromProto(e.Replace),
	}
}

func MarkedStringToProto(m semanticapi.MarkedString) *MarkedString {
	return &MarkedString{
		Language: m.Language,
		Value:    m.Value,
		IsRaw:    m.IsRaw,
	}
}

func MarkedStringFromProto(m *MarkedString) semanticapi.MarkedString {
	if m == nil {
		return semanticapi.MarkedString{}
	}
	return semanticapi.MarkedString{
		Language: m.Language,
		Value:    m.Value,
		IsRaw:    m.IsRaw,
	}
}

func MarkedStringsToProto(ms []semanticapi.MarkedString) []*MarkedString {
	out := make([]*MarkedString, len(ms))
	for i, m := range ms {
		out[i] = MarkedStringToProto(m)
	}
	return out
}

func MarkedStringsFromProto(ms []*MarkedString) []semanticapi.MarkedString {
	out := make([]semanticapi.MarkedString, len(ms))
	for i, m := range ms {
		out[i] = MarkedStringFromProto(m)
	}
	return out
}

func LocationLinkToProto(l semanticapi.LocationLink) *LocationLink {
	ret := &LocationLink{
		TargetUri:            l.TargetURI,
		TargetRange:          RangeToProto(l.TargetRange),
		TargetSelectionRange: RangeToProto(l.TargetSelectionRange),
	}
	if l.OriginSelectionRange != nil {
		ret.OriginSelectionRange = RangeToProto(*l.OriginSelectionRange)
		ret.HasOriginSelectionRange = true
	}
	return ret
}

func LocationLinkFromProto(l *LocationLink) semanticapi.LocationLink {
	if l == nil {
		return semanticapi.LocationLink{}
	}
	ret := semanticapi.LocationLink{
		TargetURI:            l.TargetUri,
		TargetRange:          RangeFromProto(l.TargetRange),
		TargetSelectionRange: RangeFromProto(l.TargetSelectionRange),
	}
	if l.HasOriginSelectionRange {
		ret.OriginSelectionRange = RangePtrFromProto(l.OriginSelectionRange)
	}
	return ret
}

func LocationLinksToProto(ls []semanticapi.LocationLink) []*LocationLink {
	out := make([]*LocationLink, len(ls))
	for i, l := range ls {
		out[i] = LocationLinkToProto(l)
	}
	return out
}

func LocationLinksFromProto(ls []*LocationLink) []semanticapi.LocationLink {
	out := make([]semanticapi.LocationLink, len(ls))
	for i, l := range ls {
		out[i] = LocationLinkFromProto(l)
	}
	return out
}

func TextDocumentIdentifierToProto(t semanticapi.TextDocumentIdentifier) *TextDocumentIdentifier {
	return &TextDocumentIdentifier{Uri: t.URI}
}

func TextDocumentIdentifierFromProto(t *TextDocumentIdentifier) semanticapi.TextDocumentIdentifier {
	if t == nil {
		return semanticapi.TextDocumentIdentifier{}
	}
	return semanticapi.TextDocumentIdentifier{URI: t.Uri}
}

func VersionedTextDocumentIdentifierToProto(t semanticapi.VersionedTextDocumentIdentifier) *VersionedTextDocumentIdentifier {
	return &VersionedTextDocumentIdentifier{Uri: t.URI, Version: t.Version}
}

func VersionedTextDocumentIdentifierFromProto(t *VersionedTextDocumentIdentifier) semanticapi.VersionedTextDocumentIdentifier {
	if t == nil {
		return semanticapi.VersionedTextDocumentIdentifier{}
	}
	return semanticapi.VersionedTextDocumentIdentifier{URI: t.Uri, Version: t.Version}
}

func TextDocumentItemToProto(t semanticapi.TextDocumentItem) *TextDocumentItem {
	return &TextDocumentItem{
		Uri:        t.URI,
		LanguageId: t.LanguageID,
		Version:    t.Version,
		Text:       t.Text,
	}
}

func TextDocumentItemFromProto(t *TextDocumentItem) semanticapi.TextDocumentItem {
	if t == nil {
		return semanticapi.TextDocumentItem{}
	}
	return semanticapi.TextDocumentItem{
		URI:        t.Uri,
		LanguageID: t.LanguageId,
		Version:    t.Version,
		Text:       t.Text,
	}
}

func MarkupKindToProto(k semanticapi.MarkupKind) uint32 {
	switch k {
	case semanticapi.MarkupKindMarkdown:
		return 2
	default:
		return 1
	}
}

func MarkupKindFromProto(k uint32) semanticapi.MarkupKind {
	switch k {
	case 2:
		return semanticapi.MarkupKindMarkdown
	default:
		return semanticapi.MarkupKindPlainText
	}
}

func MarkupContentToProto(m semanticapi.MarkupContent) *MarkupContent {
	return &MarkupContent{Kind: MarkupKindToProto(m.Kind), Value: m.Value}
}

func MarkupContentFromProto(m *MarkupContent) semanticapi.MarkupContent {
	if m == nil {
		return semanticapi.MarkupContent{}
	}
	return semanticapi.MarkupContent{Kind: MarkupKindFromProto(m.Kind), Value: m.Value}
}

func MarkupContentPtrFromProto(m *MarkupContent) *semanticapi.MarkupContent {
	if m == nil {
		return nil
	}
	mc := MarkupContentFromProto(m)
	return &mc
}

func MarkupContentPtrToProto(m *semanticapi.MarkupContent) *MarkupContent {
	if m == nil {
		return nil
	}
	return MarkupContentToProto(*m)
}

func CompletionItemToProto(c semanticapi.CompletionItem) *CompletionItem {
	ret := &CompletionItem{
		Label:               c.Label,
		Kind:                int32(c.Kind),
		Detail:              c.Detail,
		Documentation:       MarkupContentPtrToProto(c.Documentation),
		InsertText:          c.InsertText,
		InsertTextFormat:    int32(c.InsertTextFormat),
		DocumentationString: c.DocumentationString,
	}
	if c.TextEdit != nil {
		ret.TextEdit = TextEditToProto(*c.TextEdit)
	}
	if len(c.Tags) > 0 {
		tags := make([]int32, len(c.Tags))
		for i, t := range c.Tags {
			tags[i] = int32(t)
		}
		ret.Tags = tags
	}
	return ret
}

func CompletionItemFromProto(c *CompletionItem) semanticapi.CompletionItem {
	if c == nil {
		return semanticapi.CompletionItem{}
	}
	ret := semanticapi.CompletionItem{
		Label:               c.Label,
		Kind:                semanticapi.CompletionItemKind(c.Kind),
		Detail:              c.Detail,
		Documentation:       MarkupContentPtrFromProto(c.Documentation),
		InsertText:          c.InsertText,
		InsertTextFormat:    semanticapi.InsertTextFormat(c.InsertTextFormat),
		DocumentationString: c.DocumentationString,
	}
	if c.TextEdit != nil {
		te := TextEditFromProto(c.TextEdit)
		ret.TextEdit = &te
	}
	if len(c.Tags) > 0 {
		tags := make([]semanticapi.CompletionItemTag, len(c.Tags))
		for i, t := range c.Tags {
			tags[i] = semanticapi.CompletionItemTag(t)
		}
		ret.Tags = tags
	}
	return ret
}

func CompletionResultToProto(r semanticapi.CompletionResult) *CompletionResult {
	items := make([]*CompletionItem, len(r.Items))
	for i, item := range r.Items {
		items[i] = CompletionItemToProto(item)
	}
	return &CompletionResult{IsIncomplete: r.IsIncomplete, Items: items}
}

func CompletionResultFromProto(r *CompletionResult) semanticapi.CompletionResult {
	if r == nil {
		return semanticapi.CompletionResult{}
	}
	items := make([]semanticapi.CompletionItem, len(r.Items))
	for i, item := range r.Items {
		items[i] = CompletionItemFromProto(item)
	}
	return semanticapi.CompletionResult{IsIncomplete: r.IsIncomplete, Items: items}
}

func HoverToProto(h *semanticapi.Hover) (*Hover, bool) {
	if h == nil {
		return nil, false
	}
	ret := &Hover{Contents: MarkupContentToProto(h.Contents)}
	if h.Range != nil {
		ret.Range = RangeToProto(*h.Range)
		ret.HasRange = true
	}
	return ret, true
}

func HoverFromProto(h *Hover, hasResult bool) *semanticapi.Hover {
	if !hasResult || h == nil {
		return nil
	}
	ret := &semanticapi.Hover{Contents: MarkupContentFromProto(h.Contents)}
	if h.HasRange {
		ret.Range = RangePtrFromProto(h.Range)
	}
	return ret
}

func ParameterInformationToProto(p semanticapi.ParameterInformation) *ParameterInformation {
	ret := &ParameterInformation{
		Label:         p.Label,
		Documentation: MarkupContentPtrToProto(p.Documentation),
	}
	if p.LabelOffsets != nil {
		ret.LabelOffsetStart = p.LabelOffsets[0]
		ret.LabelOffsetEnd = p.LabelOffsets[1]
		ret.HasLabelOffsets = true
	}
	return ret
}

func ParameterInformationFromProto(p *ParameterInformation) semanticapi.ParameterInformation {
	if p == nil {
		return semanticapi.ParameterInformation{}
	}
	ret := semanticapi.ParameterInformation{
		Label:         p.Label,
		Documentation: MarkupContentPtrFromProto(p.Documentation),
	}
	if p.HasLabelOffsets {
		ret.LabelOffsets = &[2]uint32{p.LabelOffsetStart, p.LabelOffsetEnd}
	}
	return ret
}

func SignatureInformationToProto(s semanticapi.SignatureInformation) *SignatureInformation {
	params := make([]*ParameterInformation, len(s.Parameters))
	for i, p := range s.Parameters {
		params[i] = ParameterInformationToProto(p)
	}
	return &SignatureInformation{
		Label:         s.Label,
		Documentation: MarkupContentPtrToProto(s.Documentation),
		Parameters:    params,
	}
}

func SignatureInformationFromProto(s *SignatureInformation) semanticapi.SignatureInformation {
	if s == nil {
		return semanticapi.SignatureInformation{}
	}
	params := make([]semanticapi.ParameterInformation, len(s.Parameters))
	for i, p := range s.Parameters {
		params[i] = ParameterInformationFromProto(p)
	}
	return semanticapi.SignatureInformation{
		Label:         s.Label,
		Documentation: MarkupContentPtrFromProto(s.Documentation),
		Parameters:    params,
	}
}

func SignatureHelpToProto(s *semanticapi.SignatureHelp) (*SignatureHelp, bool) {
	if s == nil {
		return nil, false
	}
	sigs := make([]*SignatureInformation, len(s.Signatures))
	for i, sig := range s.Signatures {
		sigs[i] = SignatureInformationToProto(sig)
	}
	return &SignatureHelp{
		Signatures:      sigs,
		ActiveSignature: s.ActiveSignature,
		ActiveParameter: s.ActiveParameter,
	}, true
}

func SignatureHelpFromProto(s *SignatureHelp, hasResult bool) *semanticapi.SignatureHelp {
	if !hasResult || s == nil {
		return nil
	}
	sigs := make([]semanticapi.SignatureInformation, len(s.Signatures))
	for i, sig := range s.Signatures {
		sigs[i] = SignatureInformationFromProto(sig)
	}
	return &semanticapi.SignatureHelp{
		Signatures:      sigs,
		ActiveSignature: s.ActiveSignature,
		ActiveParameter: s.ActiveParameter,
	}
}

func DiagnosticToProto(d semanticapi.Diagnostic) *Diagnostic {
	ret := &Diagnostic{
		Range:    RangeToProto(d.Range),
		Severity: int32(d.Severity),
		Code:     d.Code,
		Source:   d.Source,
		Message:  d.Message,
		Data:     d.Data,
	}
	if d.CodeDescription != nil {
		ret.CodeDescriptionHref = d.CodeDescription.Href
	}
	if len(d.Tags) > 0 {
		tags := make([]int32, len(d.Tags))
		for i, t := range d.Tags {
			tags[i] = int32(t)
		}
		ret.Tags = tags
	}
	if len(d.RelatedInformation) > 0 {
		ri := make([]*DiagnosticRelatedInformation, len(d.RelatedInformation))
		for i, r := range d.RelatedInformation {
			ri[i] = &DiagnosticRelatedInformation{
				Location: LocationToProto(r.Location),
				Message:  r.Message,
			}
		}
		ret.RelatedInformation = ri
	}
	return ret
}

func DiagnosticFromProto(d *Diagnostic) semanticapi.Diagnostic {
	if d == nil {
		return semanticapi.Diagnostic{}
	}
	ret := semanticapi.Diagnostic{
		Range:    RangeFromProto(d.Range),
		Severity: semanticapi.DiagnosticSeverity(d.Severity),
		Code:     d.Code,
		Source:   d.Source,
		Message:  d.Message,
		Data:     d.Data,
	}
	if d.CodeDescriptionHref != "" {
		ret.CodeDescription = &semanticapi.CodeDescription{Href: d.CodeDescriptionHref}
	}
	if len(d.Tags) > 0 {
		tags := make([]semanticapi.DiagnosticTag, len(d.Tags))
		for i, t := range d.Tags {
			tags[i] = semanticapi.DiagnosticTag(t)
		}
		ret.Tags = tags
	}
	if len(d.RelatedInformation) > 0 {
		ri := make([]semanticapi.DiagnosticRelatedInformation, len(d.RelatedInformation))
		for i, r := range d.RelatedInformation {
			ri[i] = semanticapi.DiagnosticRelatedInformation{
				Location: LocationFromProto(r.Location),
				Message:  r.Message,
			}
		}
		ret.RelatedInformation = ri
	}
	return ret
}

func DiagnosticsToProto(diags []semanticapi.Diagnostic) []*Diagnostic {
	out := make([]*Diagnostic, len(diags))
	for i, d := range diags {
		out[i] = DiagnosticToProto(d)
	}
	return out
}

func DiagnosticsFromProto(diags []*Diagnostic) []semanticapi.Diagnostic {
	out := make([]semanticapi.Diagnostic, len(diags))
	for i, d := range diags {
		out[i] = DiagnosticFromProto(d)
	}
	return out
}

func PreviousResultIDsToProto(ids []semanticapi.PreviousResultID) []*PreviousResultId {
	out := make([]*PreviousResultId, len(ids))
	for i, id := range ids {
		out[i] = &PreviousResultId{
			Uri:   id.URI,
			Value: id.Value,
		}
	}
	return out
}

func WorkspaceDocumentDiagnosticReportFromProto(
	r *WorkspaceDocumentDiagnosticReport,
) semanticapi.WorkspaceDocumentDiagnosticReport {
	if r == nil {
		return semanticapi.WorkspaceDocumentDiagnosticReport{}
	}
	return semanticapi.WorkspaceDocumentDiagnosticReport{
		Kind:     r.Kind,
		ResultID: r.ResultId,
		URI:      r.Uri,
		Version:  r.Version,
		Items:    DiagnosticsFromProto(r.Items),
	}
}

func WorkspaceDocumentDiagnosticReportsFromProto(
	reports []*WorkspaceDocumentDiagnosticReport,
) []semanticapi.WorkspaceDocumentDiagnosticReport {
	out := make([]semanticapi.WorkspaceDocumentDiagnosticReport, len(reports))
	for i, r := range reports {
		out[i] = WorkspaceDocumentDiagnosticReportFromProto(r)
	}
	return out
}

func DocumentHighlightToProto(h semanticapi.DocumentHighlight) *DocumentHighlight {
	return &DocumentHighlight{
		Range: RangeToProto(h.Range),
		Kind:  int32(h.Kind),
	}
}

func DocumentHighlightFromProto(h *DocumentHighlight) semanticapi.DocumentHighlight {
	if h == nil {
		return semanticapi.DocumentHighlight{}
	}
	return semanticapi.DocumentHighlight{
		Range: RangeFromProto(h.Range),
		Kind:  semanticapi.DocumentHighlightKind(h.Kind),
	}
}

func DocumentHighlightsToProto(hs []semanticapi.DocumentHighlight) []*DocumentHighlight {
	out := make([]*DocumentHighlight, len(hs))
	for i, h := range hs {
		out[i] = DocumentHighlightToProto(h)
	}
	return out
}

func DocumentHighlightsFromProto(hs []*DocumentHighlight) []semanticapi.DocumentHighlight {
	out := make([]semanticapi.DocumentHighlight, len(hs))
	for i, h := range hs {
		out[i] = DocumentHighlightFromProto(h)
	}
	return out
}

func DocumentSymbolToProto(s semanticapi.DocumentSymbol) *DocumentSymbol {
	children := make([]*DocumentSymbol, len(s.Children))
	for i, c := range s.Children {
		children[i] = DocumentSymbolToProto(c)
	}
	return &DocumentSymbol{
		Name:           s.Name,
		Detail:         s.Detail,
		Kind:           int32(s.Kind),
		Range:          RangeToProto(s.Range),
		SelectionRange: RangeToProto(s.SelectionRange),
		Children:       children,
	}
}

func DocumentSymbolFromProto(s *DocumentSymbol) semanticapi.DocumentSymbol {
	if s == nil {
		return semanticapi.DocumentSymbol{}
	}
	children := make([]semanticapi.DocumentSymbol, len(s.Children))
	for i, c := range s.Children {
		children[i] = DocumentSymbolFromProto(c)
	}
	return semanticapi.DocumentSymbol{
		Name:           s.Name,
		Detail:         s.Detail,
		Kind:           semanticapi.SymbolKind(s.Kind),
		Range:          RangeFromProto(s.Range),
		SelectionRange: RangeFromProto(s.SelectionRange),
		Children:       children,
	}
}

func DocumentSymbolsToProto(syms []semanticapi.DocumentSymbol) []*DocumentSymbol {
	out := make([]*DocumentSymbol, len(syms))
	for i, s := range syms {
		out[i] = DocumentSymbolToProto(s)
	}
	return out
}

func DocumentSymbolsFromProto(syms []*DocumentSymbol) []semanticapi.DocumentSymbol {
	out := make([]semanticapi.DocumentSymbol, len(syms))
	for i, s := range syms {
		out[i] = DocumentSymbolFromProto(s)
	}
	return out
}

func SymbolInformationToProto(s semanticapi.SymbolInformation) *SymbolInformation {
	return &SymbolInformation{
		Name:     s.Name,
		Kind:     int32(s.Kind),
		Location: LocationToProto(s.Location),
	}
}

func SymbolInformationFromProto(s *SymbolInformation) semanticapi.SymbolInformation {
	if s == nil {
		return semanticapi.SymbolInformation{}
	}
	return semanticapi.SymbolInformation{
		Name:     s.Name,
		Kind:     semanticapi.SymbolKind(s.Kind),
		Location: LocationFromProto(s.Location),
	}
}

func SymbolInformationsToProto(syms []semanticapi.SymbolInformation) []*SymbolInformation {
	out := make([]*SymbolInformation, len(syms))
	for i, s := range syms {
		out[i] = SymbolInformationToProto(s)
	}
	return out
}

func SymbolInformationsFromProto(syms []*SymbolInformation) []semanticapi.SymbolInformation {
	out := make([]semanticapi.SymbolInformation, len(syms))
	for i, s := range syms {
		out[i] = SymbolInformationFromProto(s)
	}
	return out
}

func LocationResultToProto(r semanticapi.LocationResult) (*Location, []*Location, []*LocationLink) {
	if r.Location != nil {
		return LocationToProto(*r.Location), nil, nil
	}
	if len(r.LocationLinks) > 0 {
		return nil, nil, LocationLinksToProto(r.LocationLinks)
	}
	if len(r.Locations) > 0 {
		return nil, LocationsToProto(r.Locations), nil
	}
	return nil, nil, nil
}

func LocationResultFromProto(res *DefinitionResponse) semanticapi.LocationResult {
	if res == nil {
		return semanticapi.LocationResult{}
	}
	if len(res.GetLocationLinks()) > 0 {
		return semanticapi.LocationResult{LocationLinks: LocationLinksFromProto(res.GetLocationLinks())}
	}
	if res.GetLocation() != nil {
		loc := LocationFromProto(res.GetLocation())
		return semanticapi.LocationResult{Location: &loc}
	}
	if len(res.GetLocations()) > 0 {
		return semanticapi.LocationResult{Locations: LocationsFromProto(res.GetLocations())}
	}
	return semanticapi.LocationResult{}
}

func LocationResultFromProtoDecl(res *DeclarationResponse) semanticapi.LocationResult {
	if res == nil {
		return semanticapi.LocationResult{}
	}
	if len(res.GetLocationLinks()) > 0 {
		return semanticapi.LocationResult{LocationLinks: LocationLinksFromProto(res.GetLocationLinks())}
	}
	if res.GetLocation() != nil {
		loc := LocationFromProto(res.GetLocation())
		return semanticapi.LocationResult{Location: &loc}
	}
	if len(res.GetLocations()) > 0 {
		return semanticapi.LocationResult{Locations: LocationsFromProto(res.GetLocations())}
	}
	return semanticapi.LocationResult{}
}

func LocationResultFromProtoTypeDef(res *TypeDefinitionResponse) semanticapi.LocationResult {
	if res == nil {
		return semanticapi.LocationResult{}
	}
	if len(res.GetLocationLinks()) > 0 {
		return semanticapi.LocationResult{LocationLinks: LocationLinksFromProto(res.GetLocationLinks())}
	}
	if res.GetLocation() != nil {
		loc := LocationFromProto(res.GetLocation())
		return semanticapi.LocationResult{Location: &loc}
	}
	if len(res.GetLocations()) > 0 {
		return semanticapi.LocationResult{Locations: LocationsFromProto(res.GetLocations())}
	}
	return semanticapi.LocationResult{}
}

func LocationResultFromProtoImpl(res *ImplementationResponse) semanticapi.LocationResult {
	if res == nil {
		return semanticapi.LocationResult{}
	}
	if len(res.GetLocationLinks()) > 0 {
		return semanticapi.LocationResult{LocationLinks: LocationLinksFromProto(res.GetLocationLinks())}
	}
	if res.GetLocation() != nil {
		loc := LocationFromProto(res.GetLocation())
		return semanticapi.LocationResult{Location: &loc}
	}
	if len(res.GetLocations()) > 0 {
		return semanticapi.LocationResult{Locations: LocationsFromProto(res.GetLocations())}
	}
	return semanticapi.LocationResult{}
}

func DocumentSymbolResultToProto(r semanticapi.DocumentSymbolResult) ([]*DocumentSymbol, []*SymbolInformation) {
	if len(r.DocumentSymbols) > 0 {
		return DocumentSymbolsToProto(r.DocumentSymbols), nil
	}
	if len(r.SymbolInformation) > 0 {
		return nil, SymbolInformationsToProto(r.SymbolInformation)
	}
	return nil, nil
}

func DocumentSymbolResultFromProto(res *DocumentSymbolResponse) semanticapi.DocumentSymbolResult {
	if res == nil {
		return semanticapi.DocumentSymbolResult{}
	}
	if len(res.GetSymbolInformation()) > 0 {
		return semanticapi.DocumentSymbolResult{
			SymbolInformation: SymbolInformationsFromProto(res.GetSymbolInformation()),
		}
	}
	return semanticapi.DocumentSymbolResult{
		DocumentSymbols: DocumentSymbolsFromProto(res.GetSymbols()),
	}
}

func CodeActionResultToProto(r semanticapi.CodeActionResult) *CodeActionResultItem {
	if r.CodeAction != nil {
		return &CodeActionResultItem{CodeAction: CodeActionToProto(*r.CodeAction)}
	}
	if r.Command != nil {
		return &CodeActionResultItem{Command: CommandToProto(r.Command)}
	}
	return nil
}

func CodeActionResultFromProto(r *CodeActionResultItem) semanticapi.CodeActionResult {
	if r == nil {
		return semanticapi.CodeActionResult{}
	}
	if r.CodeAction != nil {
		ca := CodeActionFromProto(r.CodeAction)
		return semanticapi.CodeActionResult{CodeAction: &ca}
	}
	if r.Command != nil {
		return semanticapi.CodeActionResult{Command: CommandFromProto(r.Command)}
	}
	return semanticapi.CodeActionResult{}
}

func CodeActionResultsToProto(rs []semanticapi.CodeActionResult) []*CodeActionResultItem {
	out := make([]*CodeActionResultItem, len(rs))
	for i, r := range rs {
		out[i] = CodeActionResultToProto(r)
	}
	return out
}

func CodeActionResultsFromProto(res *CodeActionResponse) []semanticapi.CodeActionResult {
	if res == nil {
		return nil
	}
	// Try new items field first
	if len(res.GetItems()) > 0 {
		out := make([]semanticapi.CodeActionResult, len(res.GetItems()))
		for i, item := range res.GetItems() {
			out[i] = CodeActionResultFromProto(item)
		}
		return out
	}
	// Fall back to legacy actions field
	if len(res.GetActions()) > 0 {
		out := make([]semanticapi.CodeActionResult, len(res.GetActions()))
		for i, a := range res.GetActions() {
			ca := CodeActionFromProto(a)
			out[i] = semanticapi.CodeActionResult{CodeAction: &ca}
		}
		return out
	}
	return nil
}

func CommandToProto(c *semanticapi.Command) *Command {
	if c == nil {
		return nil
	}
	args := make([]string, len(c.Arguments))
	for i, a := range c.Arguments {
		args[i] = string(a)
	}
	return &Command{Title: c.Title, Command: c.Command, Arguments: args}
}

func CommandFromProto(c *Command) *semanticapi.Command {
	if c == nil {
		return nil
	}
	args := make([]json.RawMessage, len(c.Arguments))
	for i, a := range c.Arguments {
		args[i] = json.RawMessage(a)
	}
	return &semanticapi.Command{Title: c.Title, Command: c.Command, Arguments: args}
}

func CodeActionToProto(a semanticapi.CodeAction) *CodeAction {
	ret := &CodeAction{
		Title:       a.Title,
		Kind:        string(a.Kind),
		Diagnostics: DiagnosticsToProto(a.Diagnostics),
		Command:     CommandToProto(a.Command),
		Group:       a.Group,
	}
	if a.Edit != nil {
		ret.Edit = WorkspaceEditToProto(a.Edit)
	}
	return ret
}

func CodeActionFromProto(a *CodeAction) semanticapi.CodeAction {
	if a == nil {
		return semanticapi.CodeAction{}
	}
	ret := semanticapi.CodeAction{
		Title:       a.Title,
		Kind:        semanticapi.CodeActionKind(a.Kind),
		Diagnostics: DiagnosticsFromProto(a.Diagnostics),
		Command:     CommandFromProto(a.Command),
		Group:       a.Group,
	}
	if a.Edit != nil {
		ret.Edit = WorkspaceEditFromProto(a.Edit)
	}
	return ret
}

func CodeActionsToProto(actions []semanticapi.CodeAction) []*CodeAction {
	out := make([]*CodeAction, len(actions))
	for i, a := range actions {
		out[i] = CodeActionToProto(a)
	}
	return out
}

func CodeActionsFromProto(actions []*CodeAction) []semanticapi.CodeAction {
	out := make([]semanticapi.CodeAction, len(actions))
	for i, a := range actions {
		out[i] = CodeActionFromProto(a)
	}
	return out
}

func CodeLensToProto(l semanticapi.CodeLens) *CodeLens {
	return &CodeLens{
		Range:   RangeToProto(l.Range),
		Command: CommandToProto(l.Command),
	}
}

func CodeLensFromProto(l *CodeLens) semanticapi.CodeLens {
	if l == nil {
		return semanticapi.CodeLens{}
	}
	return semanticapi.CodeLens{
		Range:   RangeFromProto(l.Range),
		Command: CommandFromProto(l.Command),
	}
}

func CodeLensesToProto(lenses []semanticapi.CodeLens) []*CodeLens {
	out := make([]*CodeLens, len(lenses))
	for i, l := range lenses {
		out[i] = CodeLensToProto(l)
	}
	return out
}

func CodeLensesFromProto(lenses []*CodeLens) []semanticapi.CodeLens {
	out := make([]semanticapi.CodeLens, len(lenses))
	for i, l := range lenses {
		out[i] = CodeLensFromProto(l)
	}
	return out
}

func TextDocumentEditToProto(e *semanticapi.TextDocumentEdit) *TextDocumentEdit {
	if e == nil {
		return nil
	}
	return &TextDocumentEdit{
		TextDocument: VersionedTextDocumentIdentifierToProto(e.TextDocument),
		Edits:        TextEditsToProto(e.Edits),
	}
}

func TextDocumentEditFromProto(e *TextDocumentEdit) *semanticapi.TextDocumentEdit {
	if e == nil {
		return nil
	}
	return &semanticapi.TextDocumentEdit{
		TextDocument: VersionedTextDocumentIdentifierFromProto(e.TextDocument),
		Edits:        TextEditsFromProto(e.Edits),
	}
}

func DocumentChangeToProto(d semanticapi.DocumentChange) *DocumentChange {
	ret := &DocumentChange{}
	if d.TextDocumentEdit != nil {
		ret.TextDocumentEdit = TextDocumentEditToProto(d.TextDocumentEdit)
	}
	if d.CreateFile != nil {
		ret.CreateFile = &CreateFile{
			Kind:         d.CreateFile.Kind,
			Uri:          d.CreateFile.URI,
			AnnotationId: d.CreateFile.AnnotationID,
		}
		if d.CreateFile.Options != nil {
			ret.CreateFile.Options = &CreateFileOptions{
				Overwrite:      d.CreateFile.Options.Overwrite,
				IgnoreIfExists: d.CreateFile.Options.IgnoreIfExists,
			}
		}
	}
	if d.RenameFile != nil {
		ret.RenameFile = &RenameFileOp{
			Kind:         d.RenameFile.Kind,
			OldUri:       d.RenameFile.OldURI,
			NewUri:       d.RenameFile.NewURI,
			AnnotationId: d.RenameFile.AnnotationID,
		}
		if d.RenameFile.Options != nil {
			ret.RenameFile.Options = &RenameFileOptions{
				Overwrite:      d.RenameFile.Options.Overwrite,
				IgnoreIfExists: d.RenameFile.Options.IgnoreIfExists,
			}
		}
	}
	if d.DeleteFile != nil {
		ret.DeleteFile = &DeleteFile{
			Kind:         d.DeleteFile.Kind,
			Uri:          d.DeleteFile.URI,
			AnnotationId: d.DeleteFile.AnnotationID,
		}
		if d.DeleteFile.Options != nil {
			ret.DeleteFile.Options = &DeleteFileOptions{
				Recursive:         d.DeleteFile.Options.Recursive,
				IgnoreIfNotExists: d.DeleteFile.Options.IgnoreIfNotExists,
			}
		}
	}
	return ret
}

func DocumentChangeFromProto(d *DocumentChange) semanticapi.DocumentChange {
	if d == nil {
		return semanticapi.DocumentChange{}
	}
	ret := semanticapi.DocumentChange{}
	if d.TextDocumentEdit != nil {
		ret.TextDocumentEdit = TextDocumentEditFromProto(d.TextDocumentEdit)
	}
	if d.CreateFile != nil {
		ret.CreateFile = &semanticapi.CreateFile{
			Kind:         d.CreateFile.Kind,
			URI:          d.CreateFile.Uri,
			AnnotationID: d.CreateFile.AnnotationId,
		}
		if d.CreateFile.Options != nil {
			ret.CreateFile.Options = &semanticapi.CreateFileOptions{
				Overwrite:      d.CreateFile.Options.Overwrite,
				IgnoreIfExists: d.CreateFile.Options.IgnoreIfExists,
			}
		}
	}
	if d.RenameFile != nil {
		ret.RenameFile = &semanticapi.RenameFile{
			Kind:         d.RenameFile.Kind,
			OldURI:       d.RenameFile.OldUri,
			NewURI:       d.RenameFile.NewUri,
			AnnotationID: d.RenameFile.AnnotationId,
		}
		if d.RenameFile.Options != nil {
			ret.RenameFile.Options = &semanticapi.RenameFileOptions{
				Overwrite:      d.RenameFile.Options.Overwrite,
				IgnoreIfExists: d.RenameFile.Options.IgnoreIfExists,
			}
		}
	}
	if d.DeleteFile != nil {
		ret.DeleteFile = &semanticapi.DeleteFile{
			Kind:         d.DeleteFile.Kind,
			URI:          d.DeleteFile.Uri,
			AnnotationID: d.DeleteFile.AnnotationId,
		}
		if d.DeleteFile.Options != nil {
			ret.DeleteFile.Options = &semanticapi.DeleteFileOptions{
				Recursive:         d.DeleteFile.Options.Recursive,
				IgnoreIfNotExists: d.DeleteFile.Options.IgnoreIfNotExists,
			}
		}
	}
	return ret
}

func WorkspaceEditToProto(e *semanticapi.WorkspaceEdit) *WorkspaceEdit {
	if e == nil {
		return nil
	}
	changes := make(map[string]*TextEditList, len(e.Changes))
	for uri, edits := range e.Changes {
		changes[uri] = &TextEditList{Edits: TextEditsToProto(edits)}
	}
	ret := &WorkspaceEdit{Changes: changes}
	if len(e.DocumentChanges) > 0 {
		dc := make([]*DocumentChange, len(e.DocumentChanges))
		for i, d := range e.DocumentChanges {
			dc[i] = DocumentChangeToProto(d)
		}
		ret.DocumentChanges = dc
	}
	return ret
}

func WorkspaceEditFromProto(e *WorkspaceEdit) *semanticapi.WorkspaceEdit {
	if e == nil {
		return nil
	}
	changes := make(map[string][]semanticapi.TextEdit, len(e.Changes))
	for uri, list := range e.Changes {
		changes[uri] = TextEditsFromProto(list.GetEdits())
	}
	ret := &semanticapi.WorkspaceEdit{Changes: changes}
	if len(e.DocumentChanges) > 0 {
		dc := make([]semanticapi.DocumentChange, len(e.DocumentChanges))
		for i, d := range e.DocumentChanges {
			dc[i] = DocumentChangeFromProto(d)
		}
		ret.DocumentChanges = dc
	}
	return ret
}

func FoldingRangeToProto(f semanticapi.FoldingRange) *FoldingRange {
	return &FoldingRange{
		StartLine:      f.StartLine,
		StartCharacter: f.StartCharacter,
		EndLine:        f.EndLine,
		EndCharacter:   f.EndCharacter,
		Kind:           string(f.Kind),
	}
}

func FoldingRangeFromProto(f *FoldingRange) semanticapi.FoldingRange {
	if f == nil {
		return semanticapi.FoldingRange{}
	}
	return semanticapi.FoldingRange{
		StartLine:      f.StartLine,
		StartCharacter: f.StartCharacter,
		EndLine:        f.EndLine,
		EndCharacter:   f.EndCharacter,
		Kind:           semanticapi.FoldingRangeKind(f.Kind),
	}
}

func FoldingRangesToProto(ranges []semanticapi.FoldingRange) []*FoldingRange {
	out := make([]*FoldingRange, len(ranges))
	for i, f := range ranges {
		out[i] = FoldingRangeToProto(f)
	}
	return out
}

func FoldingRangesFromProto(ranges []*FoldingRange) []semanticapi.FoldingRange {
	out := make([]semanticapi.FoldingRange, len(ranges))
	for i, f := range ranges {
		out[i] = FoldingRangeFromProto(f)
	}
	return out
}

func SelectionRangeToProto(s semanticapi.SelectionRange) *SelectionRange {
	ret := &SelectionRange{Range: RangeToProto(s.Range)}
	if s.Parent != nil {
		ret.Parent = SelectionRangeToProto(*s.Parent)
	}
	return ret
}

func SelectionRangeFromProto(s *SelectionRange) semanticapi.SelectionRange {
	if s == nil {
		return semanticapi.SelectionRange{}
	}
	ret := semanticapi.SelectionRange{Range: RangeFromProto(s.Range)}
	if s.Parent != nil {
		parent := SelectionRangeFromProto(s.Parent)
		ret.Parent = &parent
	}
	return ret
}

func SelectionRangesToProto(ranges []semanticapi.SelectionRange) []*SelectionRange {
	out := make([]*SelectionRange, len(ranges))
	for i, s := range ranges {
		out[i] = SelectionRangeToProto(s)
	}
	return out
}

func SelectionRangesFromProto(ranges []*SelectionRange) []semanticapi.SelectionRange {
	out := make([]semanticapi.SelectionRange, len(ranges))
	for i, s := range ranges {
		out[i] = SelectionRangeFromProto(s)
	}
	return out
}

func SemanticTokensToProto(t *semanticapi.SemanticTokens) (*SemanticTokens, bool) {
	if t == nil {
		return nil, false
	}
	return &SemanticTokens{ResultId: t.ResultID, Data: t.Data}, true
}

func SemanticTokensFromProto(t *SemanticTokens, hasResult bool) *semanticapi.SemanticTokens {
	if !hasResult || t == nil {
		return nil
	}
	return &semanticapi.SemanticTokens{ResultID: t.ResultId, Data: t.Data}
}

func CallHierarchyItemToProto(c semanticapi.CallHierarchyItem) *CallHierarchyItem {
	return &CallHierarchyItem{
		Name:           c.Name,
		Kind:           int32(c.Kind),
		Uri:            c.URI,
		Range:          RangeToProto(c.Range),
		SelectionRange: RangeToProto(c.SelectionRange),
	}
}

func CallHierarchyItemFromProto(c *CallHierarchyItem) semanticapi.CallHierarchyItem {
	if c == nil {
		return semanticapi.CallHierarchyItem{}
	}
	return semanticapi.CallHierarchyItem{
		Name:           c.Name,
		Kind:           semanticapi.SymbolKind(c.Kind),
		URI:            c.Uri,
		Range:          RangeFromProto(c.Range),
		SelectionRange: RangeFromProto(c.SelectionRange),
	}
}

func CallHierarchyItemsToProto(items []semanticapi.CallHierarchyItem) []*CallHierarchyItem {
	out := make([]*CallHierarchyItem, len(items))
	for i, item := range items {
		out[i] = CallHierarchyItemToProto(item)
	}
	return out
}

func CallHierarchyItemsFromProto(items []*CallHierarchyItem) []semanticapi.CallHierarchyItem {
	out := make([]semanticapi.CallHierarchyItem, len(items))
	for i, item := range items {
		out[i] = CallHierarchyItemFromProto(item)
	}
	return out
}

func RangesToProto(ranges []semanticapi.Range) []*Range {
	out := make([]*Range, len(ranges))
	for i, r := range ranges {
		out[i] = RangeToProto(r)
	}
	return out
}

func RangesFromProto(ranges []*Range) []semanticapi.Range {
	out := make([]semanticapi.Range, len(ranges))
	for i, r := range ranges {
		out[i] = RangeFromProto(r)
	}
	return out
}

func CallHierarchyIncomingCallToProto(c semanticapi.CallHierarchyIncomingCall) *CallHierarchyIncomingCall {
	return &CallHierarchyIncomingCall{
		From:       CallHierarchyItemToProto(c.From),
		FromRanges: RangesToProto(c.FromRanges),
	}
}

func CallHierarchyIncomingCallFromProto(c *CallHierarchyIncomingCall) semanticapi.CallHierarchyIncomingCall {
	if c == nil {
		return semanticapi.CallHierarchyIncomingCall{}
	}
	return semanticapi.CallHierarchyIncomingCall{
		From:       CallHierarchyItemFromProto(c.From),
		FromRanges: RangesFromProto(c.FromRanges),
	}
}

func CallHierarchyIncomingCallsToProto(calls []semanticapi.CallHierarchyIncomingCall) []*CallHierarchyIncomingCall {
	out := make([]*CallHierarchyIncomingCall, len(calls))
	for i, c := range calls {
		out[i] = CallHierarchyIncomingCallToProto(c)
	}
	return out
}

func CallHierarchyIncomingCallsFromProto(calls []*CallHierarchyIncomingCall) []semanticapi.CallHierarchyIncomingCall {
	out := make([]semanticapi.CallHierarchyIncomingCall, len(calls))
	for i, c := range calls {
		out[i] = CallHierarchyIncomingCallFromProto(c)
	}
	return out
}

func CallHierarchyOutgoingCallToProto(c semanticapi.CallHierarchyOutgoingCall) *CallHierarchyOutgoingCall {
	return &CallHierarchyOutgoingCall{
		To:         CallHierarchyItemToProto(c.To),
		FromRanges: RangesToProto(c.FromRanges),
	}
}

func CallHierarchyOutgoingCallFromProto(c *CallHierarchyOutgoingCall) semanticapi.CallHierarchyOutgoingCall {
	if c == nil {
		return semanticapi.CallHierarchyOutgoingCall{}
	}
	return semanticapi.CallHierarchyOutgoingCall{
		To:         CallHierarchyItemFromProto(c.To),
		FromRanges: RangesFromProto(c.FromRanges),
	}
}

func CallHierarchyOutgoingCallsToProto(calls []semanticapi.CallHierarchyOutgoingCall) []*CallHierarchyOutgoingCall {
	out := make([]*CallHierarchyOutgoingCall, len(calls))
	for i, c := range calls {
		out[i] = CallHierarchyOutgoingCallToProto(c)
	}
	return out
}

func CallHierarchyOutgoingCallsFromProto(calls []*CallHierarchyOutgoingCall) []semanticapi.CallHierarchyOutgoingCall {
	out := make([]semanticapi.CallHierarchyOutgoingCall, len(calls))
	for i, c := range calls {
		out[i] = CallHierarchyOutgoingCallFromProto(c)
	}
	return out
}

func ServerCapabilitiesToProto(c semanticapi.ServerCapabilities) *ServerCapabilities {
	ret := &ServerCapabilities{
		HoverProvider:                   c.HoverProvider,
		DeclarationProvider:             c.DeclarationProvider,
		DefinitionProvider:              c.DefinitionProvider,
		TypeDefinitionProvider:          c.TypeDefinitionProvider,
		ImplementationProvider:          c.ImplementationProvider,
		ReferencesProvider:              c.ReferencesProvider,
		DocumentHighlightProvider:       c.DocumentHighlightProvider,
		DocumentSymbolProvider:          c.DocumentSymbolProvider,
		DocumentFormattingProvider:      c.DocumentFormattingProvider,
		DocumentRangeFormattingProvider: c.DocumentRangeFormattingProvider,
		FoldingRangeProvider:            c.FoldingRangeProvider,
		SelectionRangeProvider:          c.SelectionRangeProvider,
		WorkspaceSymbolProvider:         c.WorkspaceSymbolProvider,
		CallHierarchyProvider:           c.CallHierarchyProvider,
		DocumentColorProvider:           c.DocumentColorProvider,
		LinkedEditingRangeProvider:      c.LinkedEditingRangeProvider,
		MonikerProvider:                 c.MonikerProvider,
		TypeHierarchyProvider:           c.TypeHierarchyProvider,
		InlineValueProvider:             c.InlineValueProvider,
	}
	if c.CompletionProvider != nil {
		ret.CompletionProvider = true
		ret.CompletionTriggerCharacters = c.CompletionProvider.TriggerCharacters
		ret.CompletionResolveProvider = c.CompletionProvider.ResolveProvider
	}
	if c.SignatureHelpProvider != nil {
		ret.SignatureHelpProvider = true
		ret.SignatureHelpTriggerCharacters = c.SignatureHelpProvider.TriggerCharacters
		ret.SignatureHelpRetriggerCharacters = c.SignatureHelpProvider.RetriggerCharacters
	}
	if c.CodeActionProvider != nil {
		ret.CodeActionProvider = true
		kinds := make([]string, len(c.CodeActionProvider.CodeActionKinds))
		for i, k := range c.CodeActionProvider.CodeActionKinds {
			kinds[i] = string(k)
		}
		ret.CodeActionKinds = kinds
		ret.CodeActionResolveProvider = c.CodeActionProvider.ResolveProvider
	}
	if c.CodeLensProvider != nil {
		ret.CodeLensProvider = true
		ret.CodeLensResolveProvider = c.CodeLensProvider.ResolveProvider
	}
	if c.RenameProvider != nil {
		ret.RenameProvider = true
		ret.RenamePrepareProvider = c.RenameProvider.PrepareProvider
	}
	if c.ExecuteCommandProvider != nil {
		ret.ExecuteCommandProvider = true
		ret.ExecuteCommandCommands = c.ExecuteCommandProvider.Commands
	}
	if c.SemanticTokensProvider != nil {
		ret.SemanticTokensProvider = true
		ret.SemanticTokensTokenTypes = c.SemanticTokensProvider.Legend.TokenTypes
		ret.SemanticTokensTokenModifiers = c.SemanticTokensProvider.Legend.TokenModifiers
		ret.SemanticTokensFull = c.SemanticTokensProvider.Full
		ret.SemanticTokensRange = c.SemanticTokensProvider.Range
	}
	if c.DiagnosticProvider != nil {
		ret.DiagnosticProvider = true
		ret.DiagnosticIdentifier = c.DiagnosticProvider.Identifier
		ret.DiagnosticInterFileDependencies = c.DiagnosticProvider.InterFileDependencies
		ret.DiagnosticWorkspaceDiagnostics = c.DiagnosticProvider.WorkspaceDiagnostics
	}
	if c.DocumentLinkProvider != nil {
		ret.DocumentLinkProvider = true
		ret.DocumentLinkResolveProvider = c.DocumentLinkProvider.ResolveProvider
	}
	if c.DocumentOnTypeFormattingProvider != nil {
		ret.DocumentOnTypeFormattingProvider = true
		ret.OnTypeFormattingFirstTriggerCharacter = c.DocumentOnTypeFormattingProvider.FirstTriggerCharacter
		ret.OnTypeFormattingMoreTriggerCharacters = c.DocumentOnTypeFormattingProvider.MoreTriggerCharacter
	}
	if c.InlayHintProvider != nil {
		ret.InlayHintProvider = true
		ret.InlayHintResolveProvider = c.InlayHintProvider.ResolveProvider
	}
	ret.Experimental = c.Experimental
	return ret
}

func ServerCapabilitiesFromProto(c *ServerCapabilities) semanticapi.ServerCapabilities {
	if c == nil {
		return semanticapi.ServerCapabilities{}
	}
	ret := semanticapi.ServerCapabilities{
		HoverProvider:                   c.HoverProvider,
		DeclarationProvider:             c.DeclarationProvider,
		DefinitionProvider:              c.DefinitionProvider,
		TypeDefinitionProvider:          c.TypeDefinitionProvider,
		ImplementationProvider:          c.ImplementationProvider,
		ReferencesProvider:              c.ReferencesProvider,
		DocumentHighlightProvider:       c.DocumentHighlightProvider,
		DocumentSymbolProvider:          c.DocumentSymbolProvider,
		DocumentFormattingProvider:      c.DocumentFormattingProvider,
		DocumentRangeFormattingProvider: c.DocumentRangeFormattingProvider,
		FoldingRangeProvider:            c.FoldingRangeProvider,
		SelectionRangeProvider:          c.SelectionRangeProvider,
		WorkspaceSymbolProvider:         c.WorkspaceSymbolProvider,
		CallHierarchyProvider:           c.CallHierarchyProvider,
		DocumentColorProvider:           c.DocumentColorProvider,
		LinkedEditingRangeProvider:      c.LinkedEditingRangeProvider,
		MonikerProvider:                 c.MonikerProvider,
		TypeHierarchyProvider:           c.TypeHierarchyProvider,
		InlineValueProvider:             c.InlineValueProvider,
	}
	if c.CompletionProvider {
		ret.CompletionProvider = &semanticapi.CompletionOptions{
			TriggerCharacters: c.CompletionTriggerCharacters,
			ResolveProvider:   c.CompletionResolveProvider,
		}
	}
	if c.SignatureHelpProvider {
		ret.SignatureHelpProvider = &semanticapi.SignatureHelpOptions{
			TriggerCharacters:   c.SignatureHelpTriggerCharacters,
			RetriggerCharacters: c.SignatureHelpRetriggerCharacters,
		}
	}
	if c.CodeActionProvider {
		kinds := make([]semanticapi.CodeActionKind, len(c.CodeActionKinds))
		for i, k := range c.CodeActionKinds {
			kinds[i] = semanticapi.CodeActionKind(k)
		}
		ret.CodeActionProvider = &semanticapi.CodeActionOptions{
			CodeActionKinds: kinds,
			ResolveProvider: c.CodeActionResolveProvider,
		}
	}
	if c.CodeLensProvider {
		ret.CodeLensProvider = &semanticapi.CodeLensOptions{
			ResolveProvider: c.CodeLensResolveProvider,
		}
	}
	if c.RenameProvider {
		ret.RenameProvider = &semanticapi.RenameOptions{
			PrepareProvider: c.RenamePrepareProvider,
		}
	}
	if c.ExecuteCommandProvider {
		ret.ExecuteCommandProvider = &semanticapi.ExecuteCommandOptions{
			Commands: c.ExecuteCommandCommands,
		}
	}
	if c.SemanticTokensProvider {
		ret.SemanticTokensProvider = &semanticapi.SemanticTokensOptions{
			Legend: semanticapi.SemanticTokensLegend{
				TokenTypes:     c.SemanticTokensTokenTypes,
				TokenModifiers: c.SemanticTokensTokenModifiers,
			},
			Full:  c.SemanticTokensFull,
			Range: c.SemanticTokensRange,
		}
	}
	if c.DiagnosticProvider {
		ret.DiagnosticProvider = &semanticapi.DiagnosticOptions{
			Identifier:            c.DiagnosticIdentifier,
			InterFileDependencies: c.DiagnosticInterFileDependencies,
			WorkspaceDiagnostics:  c.DiagnosticWorkspaceDiagnostics,
		}
	}
	if c.DocumentLinkProvider {
		ret.DocumentLinkProvider = &semanticapi.DocumentLinkOptions{
			ResolveProvider: c.DocumentLinkResolveProvider,
		}
	}
	if c.DocumentOnTypeFormattingProvider {
		ret.DocumentOnTypeFormattingProvider = &semanticapi.DocumentOnTypeFormattingOptions{
			FirstTriggerCharacter: c.OnTypeFormattingFirstTriggerCharacter,
			MoreTriggerCharacter:  c.OnTypeFormattingMoreTriggerCharacters,
		}
	}
	if c.InlayHintProvider {
		ret.InlayHintProvider = &semanticapi.InlayHintOptions{
			ResolveProvider: c.InlayHintResolveProvider,
		}
	}
	if len(c.Experimental) > 0 {
		ret.Experimental = json.RawMessage(c.Experimental)
	}
	return ret
}

func DocumentDiagnosticReportToProto(r semanticapi.DocumentDiagnosticReport) *DocumentDiagnosticReport {
	ret := &DocumentDiagnosticReport{
		Kind:     r.Kind,
		ResultId: r.ResultID,
		Items:    DiagnosticsToProto(r.Items),
	}
	if len(r.RelatedDocuments) > 0 {
		ret.RelatedDocuments = make(map[string]*DiagnosticList, len(r.RelatedDocuments))
		for uri, report := range r.RelatedDocuments {
			ret.RelatedDocuments[uri] = &DiagnosticList{Diagnostics: DiagnosticsToProto(report.Items)}
		}
	}
	return ret
}

func DocumentDiagnosticReportFromProto(r *DocumentDiagnosticReport) semanticapi.DocumentDiagnosticReport {
	if r == nil {
		return semanticapi.DocumentDiagnosticReport{}
	}
	ret := semanticapi.DocumentDiagnosticReport{
		Kind:     r.Kind,
		ResultID: r.ResultId,
		Items:    DiagnosticsFromProto(r.Items),
	}
	if len(r.RelatedDocuments) > 0 {
		ret.RelatedDocuments = make(map[string]semanticapi.DocumentDiagnosticReport, len(r.RelatedDocuments))
		for uri, list := range r.RelatedDocuments {
			ret.RelatedDocuments[uri] = semanticapi.DocumentDiagnosticReport{
				Items: DiagnosticsFromProto(list.GetDiagnostics()),
			}
		}
	}
	return ret
}

func ContentChangeToProto(c semanticapi.TextDocumentContentChangeEvent) *TextDocumentContentChangeEvent {
	ret := &TextDocumentContentChangeEvent{Text: c.Text}
	if c.Range != nil {
		ret.Range = RangeToProto(*c.Range)
		ret.HasRange = true
	}
	return ret
}

func ContentChangeFromProto(c *TextDocumentContentChangeEvent) semanticapi.TextDocumentContentChangeEvent {
	if c == nil {
		return semanticapi.TextDocumentContentChangeEvent{}
	}
	ret := semanticapi.TextDocumentContentChangeEvent{Text: c.Text}
	if c.HasRange {
		ret.Range = RangePtrFromProto(c.Range)
	}
	return ret
}

func ContentChangesToProto(changes []semanticapi.TextDocumentContentChangeEvent) []*TextDocumentContentChangeEvent {
	out := make([]*TextDocumentContentChangeEvent, len(changes))
	for i, c := range changes {
		out[i] = ContentChangeToProto(c)
	}
	return out
}

func ContentChangesFromProto(changes []*TextDocumentContentChangeEvent) []semanticapi.TextDocumentContentChangeEvent {
	out := make([]semanticapi.TextDocumentContentChangeEvent, len(changes))
	for i, c := range changes {
		out[i] = ContentChangeFromProto(c)
	}
	return out
}

func PrepareRenameResultToProto(r *semanticapi.PrepareRenameResult) (*PrepareRenameResult, bool) {
	if r == nil {
		return nil, false
	}
	return &PrepareRenameResult{
		Range:       RangeToProto(r.Range),
		Placeholder: r.Placeholder,
	}, true
}

func PrepareRenameResultFromProto(r *PrepareRenameResult, hasResult bool) *semanticapi.PrepareRenameResult {
	if !hasResult || r == nil {
		return nil
	}
	return &semanticapi.PrepareRenameResult{
		Range:       RangeFromProto(r.Range),
		Placeholder: r.Placeholder,
	}
}

func ColorToProto(c semanticapi.Color) *Color {
	return &Color{Red: c.Red, Green: c.Green, Blue: c.Blue, Alpha: c.Alpha}
}

func ColorFromProto(c *Color) semanticapi.Color {
	if c == nil {
		return semanticapi.Color{}
	}
	return semanticapi.Color{Red: c.Red, Green: c.Green, Blue: c.Blue, Alpha: c.Alpha}
}

func ColorInformationToProto(c semanticapi.ColorInformation) *ColorInformation {
	return &ColorInformation{Range: RangeToProto(c.Range), Color: ColorToProto(c.Color)}
}

func ColorInformationFromProto(c *ColorInformation) semanticapi.ColorInformation {
	if c == nil {
		return semanticapi.ColorInformation{}
	}
	return semanticapi.ColorInformation{Range: RangeFromProto(c.Range), Color: ColorFromProto(c.Color)}
}

func ColorInformationsToProto(cs []semanticapi.ColorInformation) []*ColorInformation {
	out := make([]*ColorInformation, len(cs))
	for i, c := range cs {
		out[i] = ColorInformationToProto(c)
	}
	return out
}

func ColorInformationsFromProto(cs []*ColorInformation) []semanticapi.ColorInformation {
	out := make([]semanticapi.ColorInformation, len(cs))
	for i, c := range cs {
		out[i] = ColorInformationFromProto(c)
	}
	return out
}

func ColorPresentationToProto(c semanticapi.ColorPresentation) *ColorPresentation {
	ret := &ColorPresentation{Label: c.Label, AdditionalTextEdits: TextEditsToProto(c.AdditionalTextEdits)}
	if c.TextEdit != nil {
		ret.TextEdit = TextEditToProto(*c.TextEdit)
	}
	return ret
}

func ColorPresentationFromProto(c *ColorPresentation) semanticapi.ColorPresentation {
	if c == nil {
		return semanticapi.ColorPresentation{}
	}
	ret := semanticapi.ColorPresentation{Label: c.Label, AdditionalTextEdits: TextEditsFromProto(c.AdditionalTextEdits)}
	if c.TextEdit != nil {
		te := TextEditFromProto(c.TextEdit)
		ret.TextEdit = &te
	}
	return ret
}

func ColorPresentationsToProto(cs []semanticapi.ColorPresentation) []*ColorPresentation {
	out := make([]*ColorPresentation, len(cs))
	for i, c := range cs {
		out[i] = ColorPresentationToProto(c)
	}
	return out
}

func ColorPresentationsFromProto(cs []*ColorPresentation) []semanticapi.ColorPresentation {
	out := make([]semanticapi.ColorPresentation, len(cs))
	for i, c := range cs {
		out[i] = ColorPresentationFromProto(c)
	}
	return out
}

func DocumentLinkToProto(l semanticapi.DocumentLink) *DocumentLink {
	return &DocumentLink{Range: RangeToProto(l.Range), Target: l.Target, Tooltip: l.Tooltip}
}

func DocumentLinkFromProto(l *DocumentLink) semanticapi.DocumentLink {
	if l == nil {
		return semanticapi.DocumentLink{}
	}
	return semanticapi.DocumentLink{Range: RangeFromProto(l.Range), Target: l.Target, Tooltip: l.Tooltip}
}

func DocumentLinksToProto(ls []semanticapi.DocumentLink) []*DocumentLink {
	out := make([]*DocumentLink, len(ls))
	for i, l := range ls {
		out[i] = DocumentLinkToProto(l)
	}
	return out
}

func DocumentLinksFromProto(ls []*DocumentLink) []semanticapi.DocumentLink {
	out := make([]semanticapi.DocumentLink, len(ls))
	for i, l := range ls {
		out[i] = DocumentLinkFromProto(l)
	}
	return out
}

func LinkedEditingRangesToProto(r *semanticapi.LinkedEditingRanges) (*LinkedEditingRanges, bool) {
	if r == nil {
		return nil, false
	}
	return &LinkedEditingRanges{Ranges: RangesToProto(r.Ranges), WordPattern: r.WordPattern}, true
}

func LinkedEditingRangesFromProto(r *LinkedEditingRanges, hasResult bool) *semanticapi.LinkedEditingRanges {
	if !hasResult || r == nil {
		return nil
	}
	return &semanticapi.LinkedEditingRanges{Ranges: RangesFromProto(r.Ranges), WordPattern: r.WordPattern}
}

func MonikerToProto(m semanticapi.Moniker) *Moniker {
	return &Moniker{
		Scheme:     m.Scheme,
		Identifier: m.Identifier,
		Unique:     string(m.Unique),
		Kind:       string(m.Kind),
	}
}

func MonikerFromProto(m *Moniker) semanticapi.Moniker {
	if m == nil {
		return semanticapi.Moniker{}
	}
	return semanticapi.Moniker{
		Scheme:     m.Scheme,
		Identifier: m.Identifier,
		Unique:     semanticapi.MonikerUniquenessLevel(m.Unique),
		Kind:       semanticapi.MonikerKind(m.Kind),
	}
}

func MonikersToProto(ms []semanticapi.Moniker) []*Moniker {
	out := make([]*Moniker, len(ms))
	for i, m := range ms {
		out[i] = MonikerToProto(m)
	}
	return out
}

func MonikersFromProto(ms []*Moniker) []semanticapi.Moniker {
	out := make([]semanticapi.Moniker, len(ms))
	for i, m := range ms {
		out[i] = MonikerFromProto(m)
	}
	return out
}

func FileEventToProto(e semanticapi.FileEvent) *FileEvent {
	return &FileEvent{Uri: e.URI, Type: int32(e.Type)}
}

func FileEventFromProto(e *FileEvent) semanticapi.FileEvent {
	if e == nil {
		return semanticapi.FileEvent{}
	}
	return semanticapi.FileEvent{URI: e.Uri, Type: semanticapi.FileChangeType(e.Type)}
}

func FileEventsToProto(es []semanticapi.FileEvent) []*FileEvent {
	out := make([]*FileEvent, len(es))
	for i, e := range es {
		out[i] = FileEventToProto(e)
	}
	return out
}

func FileEventsFromProto(es []*FileEvent) []semanticapi.FileEvent {
	out := make([]semanticapi.FileEvent, len(es))
	for i, e := range es {
		out[i] = FileEventFromProto(e)
	}
	return out
}

func FileCreatesToProto(fs []semanticapi.FileCreate) []*FileCreate {
	out := make([]*FileCreate, len(fs))
	for i, f := range fs {
		out[i] = &FileCreate{Uri: f.URI}
	}
	return out
}

func FileCreatesFromProto(fs []*FileCreate) []semanticapi.FileCreate {
	out := make([]semanticapi.FileCreate, len(fs))
	for i, f := range fs {
		if f != nil {
			out[i] = semanticapi.FileCreate{URI: f.Uri}
		}
	}
	return out
}

func FileRenamesToProto(fs []semanticapi.FileRename) []*FileRename {
	out := make([]*FileRename, len(fs))
	for i, f := range fs {
		out[i] = &FileRename{OldUri: f.OldURI, NewUri: f.NewURI}
	}
	return out
}

func FileRenamesFromProto(fs []*FileRename) []semanticapi.FileRename {
	out := make([]semanticapi.FileRename, len(fs))
	for i, f := range fs {
		if f != nil {
			out[i] = semanticapi.FileRename{OldURI: f.OldUri, NewURI: f.NewUri}
		}
	}
	return out
}

func FileDeletesToProto(fs []semanticapi.FileDelete) []*FileDelete {
	out := make([]*FileDelete, len(fs))
	for i, f := range fs {
		out[i] = &FileDelete{Uri: f.URI}
	}
	return out
}

func FileDeletesFromProto(fs []*FileDelete) []semanticapi.FileDelete {
	out := make([]semanticapi.FileDelete, len(fs))
	for i, f := range fs {
		if f != nil {
			out[i] = semanticapi.FileDelete{URI: f.Uri}
		}
	}
	return out
}

func WorkspaceFoldersToProto(fs []semanticapi.WorkspaceFolder) []*WorkspaceFolder {
	out := make([]*WorkspaceFolder, len(fs))
	for i, f := range fs {
		out[i] = &WorkspaceFolder{Uri: f.URI, Name: f.Name}
	}
	return out
}

func WorkspaceFoldersFromProto(fs []*WorkspaceFolder) []semanticapi.WorkspaceFolder {
	out := make([]semanticapi.WorkspaceFolder, len(fs))
	for i, f := range fs {
		if f != nil {
			out[i] = semanticapi.WorkspaceFolder{URI: f.Uri, Name: f.Name}
		}
	}
	return out
}

func SemanticTokensDeltaToProto(d *semanticapi.SemanticTokensDelta) (*SemanticTokensDelta, bool) {
	if d == nil {
		return nil, false
	}
	edits := make([]*SemanticTokensEdit, len(d.Edits))
	for i, e := range d.Edits {
		edits[i] = &SemanticTokensEdit{Start: e.Start, DeleteCount: e.DeleteCount, Data: e.Data}
	}
	return &SemanticTokensDelta{ResultId: d.ResultID, Edits: edits}, true
}

func SemanticTokensDeltaFromProto(d *SemanticTokensDelta, hasResult bool) *semanticapi.SemanticTokensDelta {
	if !hasResult || d == nil {
		return nil
	}
	edits := make([]semanticapi.SemanticTokensEdit, len(d.Edits))
	for i, e := range d.Edits {
		if e != nil {
			edits[i] = semanticapi.SemanticTokensEdit{Start: e.Start, DeleteCount: e.DeleteCount, Data: e.Data}
		}
	}
	return &semanticapi.SemanticTokensDelta{ResultID: d.ResultId, Edits: edits}
}

func TypeHierarchyItemToProto(t semanticapi.TypeHierarchyItem) *TypeHierarchyItem {
	tags := make([]int32, len(t.Tags))
	for i, tag := range t.Tags {
		tags[i] = int32(tag)
	}
	return &TypeHierarchyItem{
		Name:           t.Name,
		Kind:           int32(t.Kind),
		Tags:           tags,
		Detail:         t.Detail,
		Uri:            t.URI,
		Range:          RangeToProto(t.Range),
		SelectionRange: RangeToProto(t.SelectionRange),
	}
}

func TypeHierarchyItemFromProto(t *TypeHierarchyItem) semanticapi.TypeHierarchyItem {
	if t == nil {
		return semanticapi.TypeHierarchyItem{}
	}
	tags := make([]semanticapi.SymbolTag, len(t.Tags))
	for i, tag := range t.Tags {
		tags[i] = semanticapi.SymbolTag(tag)
	}
	return semanticapi.TypeHierarchyItem{
		Name:           t.Name,
		Kind:           semanticapi.SymbolKind(t.Kind),
		Tags:           tags,
		Detail:         t.Detail,
		URI:            t.Uri,
		Range:          RangeFromProto(t.Range),
		SelectionRange: RangeFromProto(t.SelectionRange),
	}
}

func TypeHierarchyItemsToProto(items []semanticapi.TypeHierarchyItem) []*TypeHierarchyItem {
	out := make([]*TypeHierarchyItem, len(items))
	for i, item := range items {
		out[i] = TypeHierarchyItemToProto(item)
	}
	return out
}

func TypeHierarchyItemsFromProto(items []*TypeHierarchyItem) []semanticapi.TypeHierarchyItem {
	out := make([]semanticapi.TypeHierarchyItem, len(items))
	for i, item := range items {
		out[i] = TypeHierarchyItemFromProto(item)
	}
	return out
}

func InlayHintLabelPartToProto(p semanticapi.InlayHintLabelPart) *InlayHintLabelPart {
	ret := &InlayHintLabelPart{
		Value:   p.Value,
		Command: CommandToProto(p.Command),
	}
	if p.Location != nil {
		ret.Location = LocationToProto(*p.Location)
	}
	if p.Tooltip != nil {
		ret.TooltipMarkup = MarkupContentToProto(*p.Tooltip)
	}
	return ret
}

func InlayHintLabelPartFromProto(p *InlayHintLabelPart) semanticapi.InlayHintLabelPart {
	if p == nil {
		return semanticapi.InlayHintLabelPart{}
	}
	ret := semanticapi.InlayHintLabelPart{
		Value:   p.Value,
		Command: CommandFromProto(p.Command),
	}
	if p.Location != nil {
		loc := LocationFromProto(p.Location)
		ret.Location = &loc
	}
	if p.TooltipMarkup != nil {
		mc := MarkupContentFromProto(p.TooltipMarkup)
		ret.Tooltip = &mc
	}
	return ret
}

func InlayHintToProto(h semanticapi.InlayHint) *InlayHint {
	ret := &InlayHint{
		Position:     PositionToProto(h.Position),
		Label:        h.Label,
		Kind:         int32(h.Kind),
		TextEdits:    TextEditsToProto(h.TextEdits),
		PaddingLeft:  h.PaddingLeft,
		PaddingRight: h.PaddingRight,
	}
	if len(h.LabelParts) > 0 {
		parts := make([]*InlayHintLabelPart, len(h.LabelParts))
		for i, p := range h.LabelParts {
			parts[i] = InlayHintLabelPartToProto(p)
		}
		ret.LabelParts = parts
	}
	if h.Tooltip != nil {
		ret.TooltipMarkup = MarkupContentToProto(*h.Tooltip)
	}
	return ret
}

func InlayHintFromProto(h *InlayHint) semanticapi.InlayHint {
	if h == nil {
		return semanticapi.InlayHint{}
	}
	ret := semanticapi.InlayHint{
		Position:     PositionFromProto(h.Position),
		Label:        h.Label,
		Kind:         semanticapi.InlayHintKind(h.Kind),
		TextEdits:    TextEditsFromProto(h.TextEdits),
		PaddingLeft:  h.PaddingLeft,
		PaddingRight: h.PaddingRight,
	}
	if len(h.LabelParts) > 0 {
		parts := make([]semanticapi.InlayHintLabelPart, len(h.LabelParts))
		for i, p := range h.LabelParts {
			parts[i] = InlayHintLabelPartFromProto(p)
		}
		ret.LabelParts = parts
	}
	if h.TooltipMarkup != nil {
		ret.Tooltip = MarkupContentPtrFromProto(h.TooltipMarkup)
	}
	return ret
}

func InlayHintsToProto(hs []semanticapi.InlayHint) []*InlayHint {
	out := make([]*InlayHint, len(hs))
	for i, h := range hs {
		out[i] = InlayHintToProto(h)
	}
	return out
}

func InlayHintsFromProto(hs []*InlayHint) []semanticapi.InlayHint {
	out := make([]semanticapi.InlayHint, len(hs))
	for i, h := range hs {
		out[i] = InlayHintFromProto(h)
	}
	return out
}

func InlineValueToProto(v semanticapi.InlineValue) *InlineValue {
	return &InlineValue{
		Range:        RangeToProto(v.Range),
		Text:         v.Text,
		VariableName: v.VariableName,
		Expression:   v.Expression,
	}
}

func InlineValueFromProto(v *InlineValue) semanticapi.InlineValue {
	if v == nil {
		return semanticapi.InlineValue{}
	}
	return semanticapi.InlineValue{
		Range:        RangeFromProto(v.Range),
		Text:         v.Text,
		VariableName: v.VariableName,
		Expression:   v.Expression,
	}
}

func InlineValuesToProto(vs []semanticapi.InlineValue) []*InlineValue {
	out := make([]*InlineValue, len(vs))
	for i, v := range vs {
		out[i] = InlineValueToProto(v)
	}
	return out
}

func InlineValuesFromProto(vs []*InlineValue) []semanticapi.InlineValue {
	out := make([]semanticapi.InlineValue, len(vs))
	for i, v := range vs {
		out[i] = InlineValueFromProto(v)
	}
	return out
}
