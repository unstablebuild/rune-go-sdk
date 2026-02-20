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

package semanticapi

import (
	"encoding/json"

	"github.com/google/uuid"
)

// Position represents a position in a text document (0-based line and character).
type Position struct {
	Line      uint32 `json:"line"`
	Character uint32 `json:"character"`
}

// Range represents a range in a text document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location inside a resource.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// TextEdit represents a textual edit applicable to a text document.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// AnnotatedTextEdit is a TextEdit with an annotation identifier.
type AnnotatedTextEdit struct {
	Range        Range  `json:"range"`
	NewText      string `json:"newText"`
	AnnotationID string `json:"annotationId"`
}

// InsertReplaceEdit represents an insert/replace edit with separate ranges.
type InsertReplaceEdit struct {
	NewText string `json:"newText"`
	Insert  Range  `json:"insert"`
	Replace Range  `json:"replace"`
}

// TextDocumentIdentifier identifies a text document using a URI.
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// VersionedTextDocumentIdentifier identifies a specific version of a text document.
type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int32  `json:"version"`
}

// TextDocumentItem represents an item to transfer a text document from the
// client to the server.
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int32  `json:"version"`
	Text       string `json:"text"`
}

// TextDocumentPositionParams is a parameter literal used in requests to pass
// a text document and a position inside that document.
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// MarkupKind describes the content type of a Hover or CompletionItem.
type MarkupKind string

const (
	MarkupKindPlainText MarkupKind = "plaintext"
	MarkupKindMarkdown  MarkupKind = "markdown"
)

// MarkupContent represents human-readable text with a rendering format.
type MarkupContent struct {
	Kind  MarkupKind `json:"kind"`
	Value string     `json:"value"`
}

// MarkedString represents a code block with language. Deprecated in favor of MarkupContent.
// Per LSP spec: MarkedString = string | { language: string; value: string }
type MarkedString struct {
	Language string `json:"language,omitempty"`
	Value    string `json:"value"`
	IsRaw    bool   `json:"-"` // true if it was just a plain string
}

// MarshalJSON implements json.Marshaler for MarkedString.
func (m MarkedString) MarshalJSON() ([]byte, error) {
	if m.IsRaw {
		return json.Marshal(m.Value)
	}
	type obj struct {
		Language string `json:"language"`
		Value    string `json:"value"`
	}
	return json.Marshal(obj{Language: m.Language, Value: m.Value})
}

// UnmarshalJSON implements json.Unmarshaler for MarkedString.
func (m *MarkedString) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		m.IsRaw = true
		return json.Unmarshal(data, &m.Value)
	}
	type obj struct {
		Language string `json:"language"`
		Value    string `json:"value"`
	}
	var o obj
	if err := json.Unmarshal(data, &o); err != nil {
		return err
	}
	m.Language = o.Language
	m.Value = o.Value
	m.IsRaw = false
	return nil
}

// CompletionItemKind enumerates the kind of a completion entry.
type CompletionItemKind int

const (
	CompletionItemKindText          CompletionItemKind = 1
	CompletionItemKindMethod        CompletionItemKind = 2
	CompletionItemKindFunction      CompletionItemKind = 3
	CompletionItemKindConstructor   CompletionItemKind = 4
	CompletionItemKindField         CompletionItemKind = 5
	CompletionItemKindVariable      CompletionItemKind = 6
	CompletionItemKindClass         CompletionItemKind = 7
	CompletionItemKindInterface     CompletionItemKind = 8
	CompletionItemKindModule        CompletionItemKind = 9
	CompletionItemKindProperty      CompletionItemKind = 10
	CompletionItemKindUnit          CompletionItemKind = 11
	CompletionItemKindValue         CompletionItemKind = 12
	CompletionItemKindEnum          CompletionItemKind = 13
	CompletionItemKindKeyword       CompletionItemKind = 14
	CompletionItemKindSnippet       CompletionItemKind = 15
	CompletionItemKindColor         CompletionItemKind = 16
	CompletionItemKindFile          CompletionItemKind = 17
	CompletionItemKindReference     CompletionItemKind = 18
	CompletionItemKindFolder        CompletionItemKind = 19
	CompletionItemKindEnumMember    CompletionItemKind = 20
	CompletionItemKindConstant      CompletionItemKind = 21
	CompletionItemKindStruct        CompletionItemKind = 22
	CompletionItemKindEvent         CompletionItemKind = 23
	CompletionItemKindOperator      CompletionItemKind = 24
	CompletionItemKindTypeParameter CompletionItemKind = 25
)

// SymbolKind enumerates the kind of a symbol.
type SymbolKind int

const (
	SymbolKindFile          SymbolKind = 1
	SymbolKindModule        SymbolKind = 2
	SymbolKindNamespace     SymbolKind = 3
	SymbolKindPackage       SymbolKind = 4
	SymbolKindClass         SymbolKind = 5
	SymbolKindMethod        SymbolKind = 6
	SymbolKindProperty      SymbolKind = 7
	SymbolKindField         SymbolKind = 8
	SymbolKindConstructor   SymbolKind = 9
	SymbolKindEnum          SymbolKind = 10
	SymbolKindInterface     SymbolKind = 11
	SymbolKindFunction      SymbolKind = 12
	SymbolKindVariable      SymbolKind = 13
	SymbolKindConstant      SymbolKind = 14
	SymbolKindString        SymbolKind = 15
	SymbolKindNumber        SymbolKind = 16
	SymbolKindBoolean       SymbolKind = 17
	SymbolKindArray         SymbolKind = 18
	SymbolKindObject        SymbolKind = 19
	SymbolKindKey           SymbolKind = 20
	SymbolKindNull          SymbolKind = 21
	SymbolKindEnumMember    SymbolKind = 22
	SymbolKindStruct        SymbolKind = 23
	SymbolKindEvent         SymbolKind = 24
	SymbolKindOperator      SymbolKind = 25
	SymbolKindTypeParameter SymbolKind = 26
)

// DiagnosticSeverity enumerates the severity of a diagnostic.
type DiagnosticSeverity int

const (
	DiagnosticSeverityError       DiagnosticSeverity = 1
	DiagnosticSeverityWarning     DiagnosticSeverity = 2
	DiagnosticSeverityInformation DiagnosticSeverity = 3
	DiagnosticSeverityHint        DiagnosticSeverity = 4
)

// DocumentHighlightKind enumerates the kind of a document highlight.
type DocumentHighlightKind int

const (
	DocumentHighlightKindText  DocumentHighlightKind = 1
	DocumentHighlightKindRead  DocumentHighlightKind = 2
	DocumentHighlightKindWrite DocumentHighlightKind = 3
)

// CodeActionKind enumerates code action kinds.
type CodeActionKind string

const (
	CodeActionKindQuickFix              CodeActionKind = "quickfix"
	CodeActionKindRefactor              CodeActionKind = "refactor"
	CodeActionKindRefactorExtract       CodeActionKind = "refactor.extract"
	CodeActionKindRefactorInline        CodeActionKind = "refactor.inline"
	CodeActionKindRefactorRewrite       CodeActionKind = "refactor.rewrite"
	CodeActionKindSource                CodeActionKind = "source"
	CodeActionKindSourceOrganizeImports CodeActionKind = "source.organizeImports"
)

// InsertTextFormat defines how an insert text is interpreted.
type InsertTextFormat int

const (
	InsertTextFormatPlainText InsertTextFormat = 1
	InsertTextFormatSnippet   InsertTextFormat = 2
)

// FoldingRangeKind enumerates the kind of a folding range.
type FoldingRangeKind string

const (
	FoldingRangeKindComment FoldingRangeKind = "comment"
	FoldingRangeKindImports FoldingRangeKind = "imports"
	FoldingRangeKindRegion  FoldingRangeKind = "region"
)

// CodeActionTriggerKind enumerates the kind of a code action trigger.
type CodeActionTriggerKind int

const (
	CodeActionTriggerKindInvoked   CodeActionTriggerKind = 1
	CodeActionTriggerKindAutomatic CodeActionTriggerKind = 2
)

// CompletionTriggerKind enumerates how a completion was triggered.
type CompletionTriggerKind int

const (
	CompletionTriggerKindInvoked                         CompletionTriggerKind = 1
	CompletionTriggerKindTriggerCharacter                CompletionTriggerKind = 2
	CompletionTriggerKindTriggerForIncompleteCompletions CompletionTriggerKind = 3
)

// CompletionItemTag enumerates tags for completion items.
type CompletionItemTag int

const (
	CompletionItemTagDeprecated CompletionItemTag = 1
)

// DiagnosticTag enumerates tags for diagnostics.
type DiagnosticTag int

const (
	DiagnosticTagUnnecessary DiagnosticTag = 1
	DiagnosticTagDeprecated  DiagnosticTag = 2
)

// CodeDescription represents a URI to open with more information about a diagnostic error.
type CodeDescription struct {
	Href string `json:"href"`
}

// DiagnosticRelatedInformation represents a related message and source code location for a diagnostic.
type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

// LocationLink represents a link between a source and a target location.
type LocationLink struct {
	OriginSelectionRange *Range `json:"originSelectionRange,omitempty"`
	TargetURI            string `json:"targetUri"`
	TargetRange          Range  `json:"targetRange"`
	TargetSelectionRange Range  `json:"targetSelectionRange"`
}

// CompletionContext contains additional information about the context in which
// a completion request is triggered.
type CompletionContext struct {
	TriggerKind      CompletionTriggerKind `json:"triggerKind"`
	TriggerCharacter string                `json:"triggerCharacter,omitempty"`
}

// TextDocumentEdit represents an edit to a versioned text document.
// Per LSP spec, edits can be (TextEdit | AnnotatedTextEdit)[].
type TextDocumentEdit struct {
	TextDocument   VersionedTextDocumentIdentifier `json:"textDocument"`
	Edits          []TextEdit                      `json:"-"`
	AnnotatedEdits []AnnotatedTextEdit             `json:"-"`
	UseAnnotated   bool                            `json:"-"`
}

// MarshalJSON implements custom JSON marshaling for TextDocumentEdit.
func (t TextDocumentEdit) MarshalJSON() ([]byte, error) {
	type withEdits struct {
		TextDocument VersionedTextDocumentIdentifier `json:"textDocument"`
		Edits        json.RawMessage                 `json:"edits"`
	}
	w := withEdits{TextDocument: t.TextDocument}
	var b []byte
	var err error
	if t.UseAnnotated && len(t.AnnotatedEdits) > 0 {
		b, err = json.Marshal(t.AnnotatedEdits)
	} else {
		b, err = json.Marshal(t.Edits)
	}
	if err != nil {
		return nil, err
	}
	w.Edits = b
	return json.Marshal(w)
}

// UnmarshalJSON implements custom JSON unmarshaling for TextDocumentEdit.
func (t *TextDocumentEdit) UnmarshalJSON(data []byte) error {
	type withEdits struct {
		TextDocument VersionedTextDocumentIdentifier `json:"textDocument"`
		Edits        []json.RawMessage               `json:"edits"`
	}
	var w withEdits
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	t.TextDocument = w.TextDocument
	for _, raw := range w.Edits {
		// Check if it has annotationId field (AnnotatedTextEdit)
		var ate AnnotatedTextEdit
		if err := json.Unmarshal(raw, &ate); err == nil && ate.AnnotationID != "" {
			t.AnnotatedEdits = append(t.AnnotatedEdits, ate)
			t.UseAnnotated = true
		} else {
			var te TextEdit
			if err := json.Unmarshal(raw, &te); err == nil {
				t.Edits = append(t.Edits, te)
			}
		}
	}
	return nil
}

// ReferenceContext contains additional information for a references request.
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// CodeActionContext contains additional diagnostic information about the context
// in which a code action is run.
type CodeActionContext struct {
	Diagnostics []Diagnostic          `json:"diagnostics"`
	Only        []CodeActionKind      `json:"only,omitempty"`
	TriggerKind CodeActionTriggerKind `json:"triggerKind,omitempty"`
}

// FormattingOptions describes options for formatting.
type FormattingOptions struct {
	TabSize                uint32 `json:"tabSize"`
	InsertSpaces           bool   `json:"insertSpaces"`
	TrimTrailingWhitespace bool   `json:"trimTrailingWhitespace,omitempty"`
	InsertFinalNewline     bool   `json:"insertFinalNewline,omitempty"`
	TrimFinalNewlines      bool   `json:"trimFinalNewlines,omitempty"`
}

// WorkspaceFoldersChangeEvent represents a workspace folders change event.
type WorkspaceFoldersChangeEvent struct {
	Added   []WorkspaceFolder `json:"added"`
	Removed []WorkspaceFolder `json:"removed"`
}

// CompletionOptions describes the options for the completion provider.
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
	ResolveProvider   bool     `json:"resolveProvider,omitempty"`
}

// SignatureHelpOptions describes the options for the signature help provider.
type SignatureHelpOptions struct {
	TriggerCharacters   []string `json:"triggerCharacters,omitempty"`
	RetriggerCharacters []string `json:"retriggerCharacters,omitempty"`
}

// CodeActionOptions describes the options for the code action provider.
type CodeActionOptions struct {
	CodeActionKinds []CodeActionKind `json:"codeActionKinds,omitempty"`
	ResolveProvider bool             `json:"resolveProvider,omitempty"`
}

// CodeLensOptions describes the options for the code lens provider.
type CodeLensOptions struct {
	ResolveProvider bool `json:"resolveProvider,omitempty"`
}

// RenameOptions describes the options for the rename provider.
type RenameOptions struct {
	PrepareProvider bool `json:"prepareProvider,omitempty"`
}

// ExecuteCommandOptions describes the options for the execute command provider.
type ExecuteCommandOptions struct {
	Commands []string `json:"commands"`
}

// SemanticTokensLegend describes the legend used by the semantic tokens provider.
type SemanticTokensLegend struct {
	TokenTypes     []string `json:"tokenTypes"`
	TokenModifiers []string `json:"tokenModifiers"`
}

// SemanticTokensOptions describes the options for the semantic tokens provider.
type SemanticTokensOptions struct {
	Legend SemanticTokensLegend `json:"legend"`
	Full   bool                 `json:"full,omitempty"`
	Range  bool                 `json:"range,omitempty"`
}

// DiagnosticOptions describes the options for the diagnostic provider.
type DiagnosticOptions struct {
	Identifier            string `json:"identifier,omitempty"`
	InterFileDependencies bool   `json:"interFileDependencies"`
	WorkspaceDiagnostics  bool   `json:"workspaceDiagnostics"`
}

// DocumentLinkOptions describes the options for the document link provider.
type DocumentLinkOptions struct {
	ResolveProvider bool `json:"resolveProvider,omitempty"`
}

// DocumentOnTypeFormattingOptions describes the options for the on type formatting provider.
type DocumentOnTypeFormattingOptions struct {
	FirstTriggerCharacter string   `json:"firstTriggerCharacter"`
	MoreTriggerCharacter  []string `json:"moreTriggerCharacter,omitempty"`
}

// InlayHintOptions describes the options for the inlay hint provider.
type InlayHintOptions struct {
	ResolveProvider bool `json:"resolveProvider,omitempty"`
}

// CompletionItem represents a completion entry.
type CompletionItem struct {
	Label               string              `json:"label"`
	Kind                CompletionItemKind  `json:"kind,omitempty"`
	Tags                []CompletionItemTag `json:"tags,omitempty"`
	Detail              string              `json:"detail,omitempty"`
	Documentation       *MarkupContent      `json:"-"`
	DocumentationString string              `json:"-"`
	InsertText          string              `json:"insertText,omitempty"`
	InsertTextFormat    InsertTextFormat    `json:"insertTextFormat,omitempty"`
	TextEdit            *TextEdit           `json:"-"`
	TextEditIR          *InsertReplaceEdit  `json:"-"` // InsertReplaceEdit variant
}

// MarshalJSON implements custom JSON marshaling for CompletionItem.
// The "documentation" field is either a string or a MarkupContent object per LSP spec.
// The "textEdit" field is either a TextEdit or an InsertReplaceEdit per LSP spec.
func (c CompletionItem) MarshalJSON() ([]byte, error) {
	type withFields struct {
		Label            string              `json:"label"`
		Kind             CompletionItemKind  `json:"kind,omitempty"`
		Tags             []CompletionItemTag `json:"tags,omitempty"`
		Detail           string              `json:"detail,omitempty"`
		Documentation    json.RawMessage     `json:"documentation,omitempty"`
		InsertText       string              `json:"insertText,omitempty"`
		InsertTextFormat InsertTextFormat    `json:"insertTextFormat,omitempty"`
		TextEdit         json.RawMessage     `json:"textEdit,omitempty"`
	}
	w := withFields{
		Label:            c.Label,
		Kind:             c.Kind,
		Tags:             c.Tags,
		Detail:           c.Detail,
		InsertText:       c.InsertText,
		InsertTextFormat: c.InsertTextFormat,
	}
	if c.Documentation != nil {
		b, err := json.Marshal(c.Documentation)
		if err != nil {
			return nil, err
		}
		w.Documentation = b
	} else if c.DocumentationString != "" {
		b, err := json.Marshal(c.DocumentationString)
		if err != nil {
			return nil, err
		}
		w.Documentation = b
	}
	if c.TextEditIR != nil {
		b, err := json.Marshal(c.TextEditIR)
		if err != nil {
			return nil, err
		}
		w.TextEdit = b
	} else if c.TextEdit != nil {
		b, err := json.Marshal(c.TextEdit)
		if err != nil {
			return nil, err
		}
		w.TextEdit = b
	}
	return json.Marshal(w)
}

// UnmarshalJSON implements custom JSON unmarshaling for CompletionItem.
func (c *CompletionItem) UnmarshalJSON(data []byte) error {
	type withFields struct {
		Label            string              `json:"label"`
		Kind             CompletionItemKind  `json:"kind,omitempty"`
		Tags             []CompletionItemTag `json:"tags,omitempty"`
		Detail           string              `json:"detail,omitempty"`
		Documentation    json.RawMessage     `json:"documentation,omitempty"`
		InsertText       string              `json:"insertText,omitempty"`
		InsertTextFormat InsertTextFormat    `json:"insertTextFormat,omitempty"`
		TextEdit         json.RawMessage     `json:"textEdit,omitempty"`
	}
	var w withFields
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	c.Label = w.Label
	c.Kind = w.Kind
	c.Tags = w.Tags
	c.Detail = w.Detail
	c.InsertText = w.InsertText
	c.InsertTextFormat = w.InsertTextFormat

	if len(w.Documentation) > 0 {
		if w.Documentation[0] == '{' {
			var mc MarkupContent
			if err := json.Unmarshal(w.Documentation, &mc); err == nil {
				c.Documentation = &mc
			}
		} else {
			var s string
			if err := json.Unmarshal(w.Documentation, &s); err == nil {
				c.DocumentationString = s
			}
		}
	}
	if len(w.TextEdit) > 0 {
		// InsertReplaceEdit has "insert" and "replace" fields, TextEdit has "range"
		var ir InsertReplaceEdit
		if err := json.Unmarshal(w.TextEdit, &ir); err == nil && ir.Insert.Start != ir.Insert.End {
			c.TextEditIR = &ir
		} else {
			var te TextEdit
			if err := json.Unmarshal(w.TextEdit, &te); err == nil {
				c.TextEdit = &te
			}
		}
	}
	return nil
}

// CompletionResult represents the result of a completion request.
// Per LSP spec: CompletionItem[] | CompletionList | null
// When unmarshaling from CompletionItem[], IsIncomplete is set to false.
type CompletionResult struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
	IsItemsOnly  bool             `json:"-"` // true if response was just CompletionItem[]
}

// MarshalJSON implements json.Marshaler for CompletionResult.
func (r CompletionResult) MarshalJSON() ([]byte, error) {
	if r.IsItemsOnly {
		return json.Marshal(r.Items)
	}
	type completionList struct {
		IsIncomplete bool             `json:"isIncomplete"`
		Items        []CompletionItem `json:"items"`
	}
	return json.Marshal(completionList{IsIncomplete: r.IsIncomplete, Items: r.Items})
}

// UnmarshalJSON implements json.Unmarshaler for CompletionResult.
func (r *CompletionResult) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	// If it's an array, it's CompletionItem[]
	if data[0] == '[' {
		r.IsItemsOnly = true
		r.IsIncomplete = false
		return json.Unmarshal(data, &r.Items)
	}
	// Otherwise it's CompletionList
	type completionList struct {
		IsIncomplete bool             `json:"isIncomplete"`
		Items        []CompletionItem `json:"items"`
	}
	var cl completionList
	if err := json.Unmarshal(data, &cl); err != nil {
		return err
	}
	r.IsIncomplete = cl.IsIncomplete
	r.Items = cl.Items
	return nil
}

// Hover represents the result of a hover request.
// Per LSP spec, contents can be: MarkedString | MarkedString[] | MarkupContent
type Hover struct {
	Contents        MarkupContent  `json:"-"`
	ContentsMarked  *MarkedString  `json:"-"`
	ContentsMarkedA []MarkedString `json:"-"`
	Range           *Range         `json:"range,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for Hover.
// The "contents" field is either a MarkedString, MarkedString[], or MarkupContent.
func (h Hover) MarshalJSON() ([]byte, error) {
	type withContents struct {
		Contents json.RawMessage `json:"contents"`
		Range    *Range          `json:"range,omitempty"`
	}
	w := withContents{Range: h.Range}
	var b []byte
	var err error
	switch {
	case len(h.ContentsMarkedA) > 0:
		b, err = json.Marshal(h.ContentsMarkedA)
	case h.ContentsMarked != nil:
		b, err = json.Marshal(h.ContentsMarked)
	default:
		b, err = json.Marshal(h.Contents)
	}
	if err != nil {
		return nil, err
	}
	w.Contents = b
	return json.Marshal(w)
}

// UnmarshalJSON implements custom JSON unmarshaling for Hover.
func (h *Hover) UnmarshalJSON(data []byte) error {
	type withContents struct {
		Contents json.RawMessage `json:"contents"`
		Range    *Range          `json:"range,omitempty"`
	}
	var w withContents
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	h.Range = w.Range
	if len(w.Contents) == 0 {
		return nil
	}
	// Try MarkupContent first (has "kind" field)
	if w.Contents[0] == '{' {
		var mc MarkupContent
		if err := json.Unmarshal(w.Contents, &mc); err == nil && mc.Kind != "" {
			h.Contents = mc
			return nil
		}
		// Try MarkedString object {language, value}
		var ms MarkedString
		if err := json.Unmarshal(w.Contents, &ms); err == nil {
			h.ContentsMarked = &ms
			return nil
		}
	}
	// Try MarkedString array
	if w.Contents[0] == '[' {
		var arr []MarkedString
		if err := json.Unmarshal(w.Contents, &arr); err == nil {
			h.ContentsMarkedA = arr
			return nil
		}
	}
	// Try plain string (MarkedString raw)
	var s string
	if err := json.Unmarshal(w.Contents, &s); err == nil {
		h.ContentsMarked = &MarkedString{Value: s, IsRaw: true}
	}
	return nil
}

// ParameterInformation represents a parameter of a callable-signature.
type ParameterInformation struct {
	Label               string         `json:"-"`
	LabelOffsets        *[2]uint32     `json:"-"`
	Documentation       *MarkupContent `json:"-"`
	DocumentationString string         `json:"-"`
}

// MarshalJSON implements custom JSON marshaling for ParameterInformation.
// The "label" field is either a string or a [start, end] tuple per LSP spec.
// The "documentation" field is either a string or a MarkupContent object per LSP spec.
func (p ParameterInformation) MarshalJSON() ([]byte, error) {
	type withFields struct {
		Label         json.RawMessage `json:"label"`
		Documentation json.RawMessage `json:"documentation,omitempty"`
	}
	var w withFields
	if p.LabelOffsets != nil {
		b, err := json.Marshal(p.LabelOffsets)
		if err != nil {
			return nil, err
		}
		w.Label = b
	} else {
		b, err := json.Marshal(p.Label)
		if err != nil {
			return nil, err
		}
		w.Label = b
	}
	if p.Documentation != nil {
		b, err := json.Marshal(p.Documentation)
		if err != nil {
			return nil, err
		}
		w.Documentation = b
	} else if p.DocumentationString != "" {
		b, err := json.Marshal(p.DocumentationString)
		if err != nil {
			return nil, err
		}
		w.Documentation = b
	}
	return json.Marshal(w)
}

// UnmarshalJSON implements custom JSON unmarshaling for ParameterInformation.
func (p *ParameterInformation) UnmarshalJSON(data []byte) error {
	type withFields struct {
		Label         json.RawMessage `json:"label"`
		Documentation json.RawMessage `json:"documentation,omitempty"`
	}
	var w withFields
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	if len(w.Label) > 0 && w.Label[0] == '[' {
		var offsets [2]uint32
		if err := json.Unmarshal(w.Label, &offsets); err == nil {
			p.LabelOffsets = &offsets
		}
	} else if len(w.Label) > 0 {
		_ = json.Unmarshal(w.Label, &p.Label)
	}
	if len(w.Documentation) > 0 {
		if w.Documentation[0] == '{' {
			var mc MarkupContent
			if err := json.Unmarshal(w.Documentation, &mc); err == nil {
				p.Documentation = &mc
				return nil
			}
		}
		var s string
		if err := json.Unmarshal(w.Documentation, &s); err == nil {
			p.DocumentationString = s
		}
	}
	return nil
}

// SignatureInformation represents the signature of something callable.
type SignatureInformation struct {
	Label               string                 `json:"label"`
	Documentation       *MarkupContent         `json:"-"`
	DocumentationString string                 `json:"-"`
	Parameters          []ParameterInformation `json:"parameters,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for SignatureInformation.
// The "documentation" field is either a string or a MarkupContent object per LSP spec.
func (s SignatureInformation) MarshalJSON() ([]byte, error) {
	type withDoc struct {
		Label         string                 `json:"label"`
		Documentation json.RawMessage        `json:"documentation,omitempty"`
		Parameters    []ParameterInformation `json:"parameters,omitempty"`
	}
	w := withDoc{Label: s.Label, Parameters: s.Parameters}
	if s.Documentation != nil {
		b, err := json.Marshal(s.Documentation)
		if err != nil {
			return nil, err
		}
		w.Documentation = b
	} else if s.DocumentationString != "" {
		b, err := json.Marshal(s.DocumentationString)
		if err != nil {
			return nil, err
		}
		w.Documentation = b
	}
	return json.Marshal(w)
}

// UnmarshalJSON implements custom JSON unmarshaling for SignatureInformation.
func (s *SignatureInformation) UnmarshalJSON(data []byte) error {
	type withDoc struct {
		Label         string                 `json:"label"`
		Documentation json.RawMessage        `json:"documentation,omitempty"`
		Parameters    []ParameterInformation `json:"parameters,omitempty"`
	}
	var w withDoc
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	s.Label = w.Label
	s.Parameters = w.Parameters
	if len(w.Documentation) > 0 {
		if w.Documentation[0] == '{' {
			var mc MarkupContent
			if err := json.Unmarshal(w.Documentation, &mc); err == nil {
				s.Documentation = &mc
				return nil
			}
		}
		var str string
		if err := json.Unmarshal(w.Documentation, &str); err == nil {
			s.DocumentationString = str
		}
	}
	return nil
}

// SignatureHelp represents the signature of something callable.
type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature uint32                 `json:"activeSignature"`
	ActiveParameter uint32                 `json:"activeParameter"`
}

// Diagnostic represents a diagnostic, such as a compiler error or warning.
type Diagnostic struct {
	Range              Range                          `json:"range"`
	Severity           DiagnosticSeverity             `json:"severity,omitempty"`
	Code               string                         `json:"-"`
	CodeInt            int                            `json:"-"`
	CodeIsInt          bool                           `json:"-"`
	CodeDescription    *CodeDescription               `json:"codeDescription,omitempty"`
	Source             string                         `json:"source,omitempty"`
	Message            string                         `json:"message"`
	Tags               []DiagnosticTag                `json:"tags,omitempty"`
	RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
	Data               json.RawMessage                `json:"data,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for Diagnostic.
// The "code" field is either an integer or a string per LSP spec.
func (d Diagnostic) MarshalJSON() ([]byte, error) {
	type withCode struct {
		Range              Range                          `json:"range"`
		Severity           DiagnosticSeverity             `json:"severity,omitempty"`
		Code               json.RawMessage                `json:"code,omitempty"`
		CodeDescription    *CodeDescription               `json:"codeDescription,omitempty"`
		Source             string                         `json:"source,omitempty"`
		Message            string                         `json:"message"`
		Tags               []DiagnosticTag                `json:"tags,omitempty"`
		RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
		Data               json.RawMessage                `json:"data,omitempty"`
	}
	w := withCode{
		Range:              d.Range,
		Severity:           d.Severity,
		CodeDescription:    d.CodeDescription,
		Source:             d.Source,
		Message:            d.Message,
		Tags:               d.Tags,
		RelatedInformation: d.RelatedInformation,
		Data:               d.Data,
	}
	if d.CodeIsInt {
		b, err := json.Marshal(d.CodeInt)
		if err != nil {
			return nil, err
		}
		w.Code = b
	} else if d.Code != "" {
		b, err := json.Marshal(d.Code)
		if err != nil {
			return nil, err
		}
		w.Code = b
	}
	return json.Marshal(w)
}

// UnmarshalJSON implements custom JSON unmarshaling for Diagnostic.
func (d *Diagnostic) UnmarshalJSON(data []byte) error {
	type withCode struct {
		Range              Range                          `json:"range"`
		Severity           DiagnosticSeverity             `json:"severity,omitempty"`
		Code               json.RawMessage                `json:"code,omitempty"`
		CodeDescription    *CodeDescription               `json:"codeDescription,omitempty"`
		Source             string                         `json:"source,omitempty"`
		Message            string                         `json:"message"`
		Tags               []DiagnosticTag                `json:"tags,omitempty"`
		RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
		Data               json.RawMessage                `json:"data,omitempty"`
	}
	var w withCode
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	d.Range = w.Range
	d.Severity = w.Severity
	d.CodeDescription = w.CodeDescription
	d.Source = w.Source
	d.Message = w.Message
	d.Tags = w.Tags
	d.RelatedInformation = w.RelatedInformation
	d.Data = w.Data
	if len(w.Code) > 0 {
		if w.Code[0] == '"' {
			_ = json.Unmarshal(w.Code, &d.Code)
		} else {
			d.CodeIsInt = true
			_ = json.Unmarshal(w.Code, &d.CodeInt)
		}
	}
	return nil
}

// DocumentHighlight represents a document highlight, such as all usages of a symbol.
type DocumentHighlight struct {
	Range Range                 `json:"range"`
	Kind  DocumentHighlightKind `json:"kind,omitempty"`
}

// DocumentSymbol represents information about programming constructs like
// variables, classes, interfaces etc.
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           SymbolKind       `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// SymbolInformation represents information about programming constructs like
// variables, classes, interfaces etc. (flat version for workspace symbols).
type SymbolInformation struct {
	Name     string     `json:"name"`
	Kind     SymbolKind `json:"kind"`
	Location Location   `json:"location"`
}

// Command represents a reference to a command.
type Command struct {
	Title     string            `json:"title"`
	Command   string            `json:"command"`
	Arguments []json.RawMessage `json:"arguments,omitempty"`
}

// CodeAction represents a change that can be performed in code.
type CodeAction struct {
	Title       string         `json:"title"`
	Kind        CodeActionKind `json:"kind,omitempty"`
	Diagnostics []Diagnostic   `json:"diagnostics,omitempty"`
	Edit        *WorkspaceEdit `json:"edit,omitempty"`
	Command     *Command       `json:"command,omitempty"`
}

// CodeLens represents a command that should be shown along with source text.
type CodeLens struct {
	Range   Range    `json:"range"`
	Command *Command `json:"command,omitempty"`
}

// CreateFileOptions contains options for creating a file.
type CreateFileOptions struct {
	Overwrite      bool `json:"overwrite,omitempty"`
	IgnoreIfExists bool `json:"ignoreIfExists,omitempty"`
}

// CreateFile represents a create file operation.
type CreateFile struct {
	Kind         string             `json:"kind"` // always "create"
	URI          string             `json:"uri"`
	Options      *CreateFileOptions `json:"options,omitempty"`
	AnnotationID string             `json:"annotationId,omitempty"`
}

// RenameFileOptions contains options for renaming a file.
type RenameFileOptions struct {
	Overwrite      bool `json:"overwrite,omitempty"`
	IgnoreIfExists bool `json:"ignoreIfExists,omitempty"`
}

// RenameFile represents a rename file operation.
type RenameFile struct {
	Kind         string             `json:"kind"` // always "rename"
	OldURI       string             `json:"oldUri"`
	NewURI       string             `json:"newUri"`
	Options      *RenameFileOptions `json:"options,omitempty"`
	AnnotationID string             `json:"annotationId,omitempty"`
}

// DeleteFileOptions contains options for deleting a file.
type DeleteFileOptions struct {
	Recursive         bool `json:"recursive,omitempty"`
	IgnoreIfNotExists bool `json:"ignoreIfNotExists,omitempty"`
}

// DeleteFile represents a delete file operation.
type DeleteFile struct {
	Kind         string             `json:"kind"` // always "delete"
	URI          string             `json:"uri"`
	Options      *DeleteFileOptions `json:"options,omitempty"`
	AnnotationID string             `json:"annotationId,omitempty"`
}

// DocumentChange represents a single change in documentChanges array.
// It can be a TextDocumentEdit, CreateFile, RenameFile, or DeleteFile.
type DocumentChange struct {
	TextDocumentEdit *TextDocumentEdit `json:"-"`
	CreateFile       *CreateFile       `json:"-"`
	RenameFile       *RenameFile       `json:"-"`
	DeleteFile       *DeleteFile       `json:"-"`
}

// MarshalJSON implements json.Marshaler for DocumentChange.
func (d DocumentChange) MarshalJSON() ([]byte, error) {
	switch {
	case d.CreateFile != nil:
		return json.Marshal(d.CreateFile)
	case d.RenameFile != nil:
		return json.Marshal(d.RenameFile)
	case d.DeleteFile != nil:
		return json.Marshal(d.DeleteFile)
	case d.TextDocumentEdit != nil:
		return json.Marshal(d.TextDocumentEdit)
	}
	return []byte("null"), nil
}

// UnmarshalJSON implements json.Unmarshaler for DocumentChange.
func (d *DocumentChange) UnmarshalJSON(data []byte) error {
	// Check for "kind" field to identify file operations
	var probe struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(data, &probe); err == nil && probe.Kind != "" {
		switch probe.Kind {
		case "create":
			d.CreateFile = &CreateFile{}
			return json.Unmarshal(data, d.CreateFile)
		case "rename":
			d.RenameFile = &RenameFile{}
			return json.Unmarshal(data, d.RenameFile)
		case "delete":
			d.DeleteFile = &DeleteFile{}
			return json.Unmarshal(data, d.DeleteFile)
		}
	}
	// Otherwise it's a TextDocumentEdit
	d.TextDocumentEdit = &TextDocumentEdit{}
	return json.Unmarshal(data, d.TextDocumentEdit)
}

// WorkspaceEdit represents changes to many resources managed in the workspace.
// Per LSP spec, documentChanges can be TextDocumentEdit[] or
// (TextDocumentEdit | CreateFile | RenameFile | DeleteFile)[].
type WorkspaceEdit struct {
	Changes         map[string][]TextEdit `json:"changes,omitempty"`
	DocumentChanges []DocumentChange      `json:"documentChanges,omitempty"`
}

// FoldingRange represents a folding range.
type FoldingRange struct {
	StartLine      uint32           `json:"startLine"`
	StartCharacter uint32           `json:"startCharacter,omitempty"`
	EndLine        uint32           `json:"endLine"`
	EndCharacter   uint32           `json:"endCharacter,omitempty"`
	Kind           FoldingRangeKind `json:"kind,omitempty"`
}

// SelectionRange represents a selection range.
type SelectionRange struct {
	Range  Range           `json:"range"`
	Parent *SelectionRange `json:"parent,omitempty"`
}

// SemanticTokens represents semantic tokens for a document.
type SemanticTokens struct {
	ResultID string   `json:"resultId,omitempty"`
	Data     []uint32 `json:"data"`
}

// CallHierarchyItem represents a call hierarchy item.
type CallHierarchyItem struct {
	Name           string     `json:"name"`
	Kind           SymbolKind `json:"kind"`
	URI            string     `json:"uri"`
	Range          Range      `json:"range"`
	SelectionRange Range      `json:"selectionRange"`
}

// CallHierarchyIncomingCall represents an incoming call.
type CallHierarchyIncomingCall struct {
	From       CallHierarchyItem `json:"from"`
	FromRanges []Range           `json:"fromRanges"`
}

// CallHierarchyOutgoingCall represents an outgoing call.
type CallHierarchyOutgoingCall struct {
	To         CallHierarchyItem `json:"to"`
	FromRanges []Range           `json:"fromRanges"`
}

// DocumentDiagnosticReport represents a diagnostic report for a document.
type DocumentDiagnosticReport struct {
	Kind             string                              `json:"kind"`
	ResultID         string                              `json:"resultId,omitempty"`
	Items            []Diagnostic                        `json:"items,omitempty"`
	RelatedDocuments map[string]DocumentDiagnosticReport `json:"relatedDocuments,omitempty"`
}

// TextDocumentSaveReason enumerates the reasons why a text document is saved.
type TextDocumentSaveReason int

const (
	TextDocumentSaveReasonManual     TextDocumentSaveReason = 1
	TextDocumentSaveReasonAfterDelay TextDocumentSaveReason = 2
	TextDocumentSaveReasonFocusOut   TextDocumentSaveReason = 3
)

// FileChangeType enumerates the type of file events.
type FileChangeType int

const (
	FileChangeTypeCreated FileChangeType = 1
	FileChangeTypeChanged FileChangeType = 2
	FileChangeTypeDeleted FileChangeType = 3
)

// TraceValue represents a trace value.
type TraceValue string

const (
	TraceValueOff      TraceValue = "off"
	TraceValueMessages TraceValue = "messages"
	TraceValueVerbose  TraceValue = "verbose"
)

// MonikerUniquenessLevel enumerates the uniqueness level of a moniker.
type MonikerUniquenessLevel string

const (
	MonikerUniquenessLevelDocument MonikerUniquenessLevel = "document"
	MonikerUniquenessLevelProject  MonikerUniquenessLevel = "project"
	MonikerUniquenessLevelGroup    MonikerUniquenessLevel = "group"
	MonikerUniquenessLevelScheme   MonikerUniquenessLevel = "scheme"
	MonikerUniquenessLevelGlobal   MonikerUniquenessLevel = "global"
)

// MonikerKind enumerates the kind of a moniker.
type MonikerKind string

const (
	MonikerKindImport MonikerKind = "import"
	MonikerKindExport MonikerKind = "export"
	MonikerKindLocal  MonikerKind = "local"
)

// InlayHintKind enumerates the kind of an inlay hint.
type InlayHintKind int

const (
	InlayHintKindType      InlayHintKind = 1
	InlayHintKindParameter InlayHintKind = 2
)

// SymbolTag enumerates tags for symbols.
type SymbolTag int

const (
	SymbolTagDeprecated SymbolTag = 1
)

// Color represents a color in RGBA space.
type Color struct {
	Red   float64 `json:"red"`
	Green float64 `json:"green"`
	Blue  float64 `json:"blue"`
	Alpha float64 `json:"alpha"`
}

// ColorInformation contains a range and a color found at that range.
type ColorInformation struct {
	Range Range `json:"range"`
	Color Color `json:"color"`
}

// ColorPresentation describes how a color is presented.
type ColorPresentation struct {
	Label               string     `json:"label"`
	TextEdit            *TextEdit  `json:"textEdit,omitempty"`
	AdditionalTextEdits []TextEdit `json:"additionalTextEdits,omitempty"`
}

// DocumentLink represents a link in a document.
type DocumentLink struct {
	Range   Range  `json:"range"`
	Target  string `json:"target,omitempty"`
	Tooltip string `json:"tooltip,omitempty"`
}

// LinkedEditingRanges represents linked editing ranges.
type LinkedEditingRanges struct {
	Ranges      []Range `json:"ranges"`
	WordPattern string  `json:"wordPattern,omitempty"`
}

// Moniker represents a moniker attached to a symbol.
type Moniker struct {
	Scheme     string                 `json:"scheme"`
	Identifier string                 `json:"identifier"`
	Unique     MonikerUniquenessLevel `json:"unique"`
	Kind       MonikerKind            `json:"kind,omitempty"`
}

// WorkspaceFolder represents a workspace folder.
type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// FileEvent represents a file event.
type FileEvent struct {
	URI  string         `json:"uri"`
	Type FileChangeType `json:"type"`
}

// FileCreate represents a file creation event.
type FileCreate struct {
	URI string `json:"uri"`
}

// FileRename represents a file rename event.
type FileRename struct {
	OldURI string `json:"oldUri"`
	NewURI string `json:"newUri"`
}

// FileDelete represents a file deletion event.
type FileDelete struct {
	URI string `json:"uri"`
}

// SemanticTokensDelta represents a delta of semantic tokens.
type SemanticTokensDelta struct {
	ResultID string               `json:"resultId,omitempty"`
	Edits    []SemanticTokensEdit `json:"edits"`
}

// SemanticTokensEdit represents an edit to semantic tokens.
type SemanticTokensEdit struct {
	Start       uint32   `json:"start"`
	DeleteCount uint32   `json:"deleteCount"`
	Data        []uint32 `json:"data,omitempty"`
}

// TypeHierarchyItem represents an item in a type hierarchy.
type TypeHierarchyItem struct {
	Name           string      `json:"name"`
	Kind           SymbolKind  `json:"kind"`
	Tags           []SymbolTag `json:"tags,omitempty"`
	Detail         string      `json:"detail,omitempty"`
	URI            string      `json:"uri"`
	Range          Range       `json:"range"`
	SelectionRange Range       `json:"selectionRange"`
}

// InlayHint represents an inlay hint.
type InlayHint struct {
	Position      Position             `json:"position"`
	Label         string               `json:"-"`
	LabelParts    []InlayHintLabelPart `json:"-"`
	Kind          InlayHintKind        `json:"kind,omitempty"`
	TextEdits     []TextEdit           `json:"textEdits,omitempty"`
	Tooltip       *MarkupContent       `json:"-"`
	TooltipString string               `json:"-"`
	PaddingLeft   bool                 `json:"paddingLeft,omitempty"`
	PaddingRight  bool                 `json:"paddingRight,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for InlayHint.
// The "label" field is either a string or an array of InlayHintLabelPart per LSP spec.
// The "tooltip" field is either a string or a MarkupContent object per LSP spec.
func (h InlayHint) MarshalJSON() ([]byte, error) {
	type withFields struct {
		Position     Position        `json:"position"`
		Label        json.RawMessage `json:"label"`
		Kind         InlayHintKind   `json:"kind,omitempty"`
		TextEdits    []TextEdit      `json:"textEdits,omitempty"`
		Tooltip      json.RawMessage `json:"tooltip,omitempty"`
		PaddingLeft  bool            `json:"paddingLeft,omitempty"`
		PaddingRight bool            `json:"paddingRight,omitempty"`
	}
	w := withFields{
		Position:     h.Position,
		Kind:         h.Kind,
		TextEdits:    h.TextEdits,
		PaddingLeft:  h.PaddingLeft,
		PaddingRight: h.PaddingRight,
	}
	if len(h.LabelParts) > 0 {
		b, err := json.Marshal(h.LabelParts)
		if err != nil {
			return nil, err
		}
		w.Label = b
	} else {
		b, err := json.Marshal(h.Label)
		if err != nil {
			return nil, err
		}
		w.Label = b
	}
	if h.Tooltip != nil {
		b, err := json.Marshal(h.Tooltip)
		if err != nil {
			return nil, err
		}
		w.Tooltip = b
	} else if h.TooltipString != "" {
		b, err := json.Marshal(h.TooltipString)
		if err != nil {
			return nil, err
		}
		w.Tooltip = b
	}
	return json.Marshal(w)
}

// UnmarshalJSON implements custom JSON unmarshaling for InlayHint.
func (h *InlayHint) UnmarshalJSON(data []byte) error {
	type withFields struct {
		Position     Position        `json:"position"`
		Label        json.RawMessage `json:"label"`
		Kind         InlayHintKind   `json:"kind,omitempty"`
		TextEdits    []TextEdit      `json:"textEdits,omitempty"`
		Tooltip      json.RawMessage `json:"tooltip,omitempty"`
		PaddingLeft  bool            `json:"paddingLeft,omitempty"`
		PaddingRight bool            `json:"paddingRight,omitempty"`
	}
	var w withFields
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	h.Position = w.Position
	h.Kind = w.Kind
	h.TextEdits = w.TextEdits
	h.PaddingLeft = w.PaddingLeft
	h.PaddingRight = w.PaddingRight
	if len(w.Label) > 0 && w.Label[0] == '[' {
		_ = json.Unmarshal(w.Label, &h.LabelParts)
	} else if len(w.Label) > 0 {
		_ = json.Unmarshal(w.Label, &h.Label)
	}
	if len(w.Tooltip) > 0 {
		if w.Tooltip[0] == '{' {
			var mc MarkupContent
			if err := json.Unmarshal(w.Tooltip, &mc); err == nil {
				h.Tooltip = &mc
				return nil
			}
		}
		var s string
		if err := json.Unmarshal(w.Tooltip, &s); err == nil {
			h.TooltipString = s
		}
	}
	return nil
}

// InlayHintLabelPart represents a part of an inlay hint label.
type InlayHintLabelPart struct {
	Value         string         `json:"value"`
	Tooltip       *MarkupContent `json:"-"`
	TooltipString string         `json:"-"`
	Location      *Location      `json:"location,omitempty"`
	Command       *Command       `json:"command,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for InlayHintLabelPart.
// The "tooltip" field is either a string or a MarkupContent object per LSP spec.
func (p InlayHintLabelPart) MarshalJSON() ([]byte, error) {
	type withTooltip struct {
		Value    string          `json:"value"`
		Tooltip  json.RawMessage `json:"tooltip,omitempty"`
		Location *Location       `json:"location,omitempty"`
		Command  *Command        `json:"command,omitempty"`
	}
	w := withTooltip{
		Value:    p.Value,
		Location: p.Location,
		Command:  p.Command,
	}
	if p.Tooltip != nil {
		b, err := json.Marshal(p.Tooltip)
		if err != nil {
			return nil, err
		}
		w.Tooltip = b
	} else if p.TooltipString != "" {
		b, err := json.Marshal(p.TooltipString)
		if err != nil {
			return nil, err
		}
		w.Tooltip = b
	}
	return json.Marshal(w)
}

// UnmarshalJSON implements custom JSON unmarshaling for InlayHintLabelPart.
func (p *InlayHintLabelPart) UnmarshalJSON(data []byte) error {
	type withTooltip struct {
		Value    string          `json:"value"`
		Tooltip  json.RawMessage `json:"tooltip,omitempty"`
		Location *Location       `json:"location,omitempty"`
		Command  *Command        `json:"command,omitempty"`
	}
	var w withTooltip
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	p.Value = w.Value
	p.Location = w.Location
	p.Command = w.Command
	if len(w.Tooltip) > 0 {
		if w.Tooltip[0] == '{' {
			var mc MarkupContent
			if err := json.Unmarshal(w.Tooltip, &mc); err == nil {
				p.Tooltip = &mc
				return nil
			}
		}
		var s string
		if err := json.Unmarshal(w.Tooltip, &s); err == nil {
			p.TooltipString = s
		}
	}
	return nil
}

// InlineValue represents an inline value.
type InlineValue struct {
	Range        Range  `json:"range"`
	Text         string `json:"text,omitempty"`
	VariableName string `json:"variableName,omitempty"`
	Expression   string `json:"expression,omitempty"`
}

// ShowDocumentResult is the result of a showDocument request.
type ShowDocumentResult struct {
	Success bool `json:"success"`
}

// ServerCapabilities describes the capabilities of a language server.
type ServerCapabilities struct {
	CompletionProvider               *CompletionOptions               `json:"completionProvider,omitempty"`
	HoverProvider                    bool                             `json:"hoverProvider,omitempty"`
	SignatureHelpProvider            *SignatureHelpOptions            `json:"signatureHelpProvider,omitempty"`
	DeclarationProvider              bool                             `json:"declarationProvider,omitempty"`
	DefinitionProvider               bool                             `json:"definitionProvider,omitempty"`
	TypeDefinitionProvider           bool                             `json:"typeDefinitionProvider,omitempty"`
	ImplementationProvider           bool                             `json:"implementationProvider,omitempty"`
	ReferencesProvider               bool                             `json:"referencesProvider,omitempty"`
	DocumentHighlightProvider        bool                             `json:"documentHighlightProvider,omitempty"`
	DocumentSymbolProvider           bool                             `json:"documentSymbolProvider,omitempty"`
	CodeActionProvider               *CodeActionOptions               `json:"codeActionProvider,omitempty"`
	CodeLensProvider                 *CodeLensOptions                 `json:"codeLensProvider,omitempty"`
	DocumentFormattingProvider       bool                             `json:"documentFormattingProvider,omitempty"`
	DocumentRangeFormattingProvider  bool                             `json:"documentRangeFormattingProvider,omitempty"`
	RenameProvider                   *RenameOptions                   `json:"renameProvider,omitempty"`
	FoldingRangeProvider             bool                             `json:"foldingRangeProvider,omitempty"`
	SelectionRangeProvider           bool                             `json:"selectionRangeProvider,omitempty"`
	SemanticTokensProvider           *SemanticTokensOptions           `json:"semanticTokensProvider,omitempty"`
	DiagnosticProvider               *DiagnosticOptions               `json:"diagnosticProvider,omitempty"`
	WorkspaceSymbolProvider          bool                             `json:"workspaceSymbolProvider,omitempty"`
	ExecuteCommandProvider           *ExecuteCommandOptions           `json:"executeCommandProvider,omitempty"`
	CallHierarchyProvider            bool                             `json:"callHierarchyProvider,omitempty"`
	DocumentColorProvider            bool                             `json:"documentColorProvider,omitempty"`
	DocumentLinkProvider             *DocumentLinkOptions             `json:"documentLinkProvider,omitempty"`
	DocumentOnTypeFormattingProvider *DocumentOnTypeFormattingOptions `json:"documentOnTypeFormattingProvider,omitempty"`
	LinkedEditingRangeProvider       bool                             `json:"linkedEditingRangeProvider,omitempty"`
	MonikerProvider                  bool                             `json:"monikerProvider,omitempty"`
	TypeHierarchyProvider            bool                             `json:"typeHierarchyProvider,omitempty"`
	InlayHintProvider                *InlayHintOptions                `json:"inlayHintProvider,omitempty"`
	InlineValueProvider              bool                             `json:"inlineValueProvider,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling for ServerCapabilities.
// LSP allows some provider fields to be either bool or options object.
func (c *ServerCapabilities) UnmarshalJSON(data []byte) error {
	// Use a raw map to handle bool | object union types.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	unmarshalBool := func(key string) bool {
		v, ok := raw[key]
		if !ok {
			return false
		}
		var b bool
		if json.Unmarshal(v, &b) == nil {
			return b
		}
		// If it's an object, it's truthy.
		return v[0] == '{'
	}

	// unmarshalBoolOrOptions: if true → &T{}, if object → unmarshal, if false/absent → nil.
	unmarshalPtr := func(key string, target any) {
		v, ok := raw[key]
		if !ok || string(v) == "false" || string(v) == "null" {
			return
		}
		if string(v) == "true" {
			// target is already zero-valued pointer; caller handles.
			return
		}
		_ = json.Unmarshal(v, target)
	}

	unmarshalPtrOrBool := func(key string, alloc func() any) any {
		v, ok := raw[key]
		if !ok || string(v) == "false" || string(v) == "null" {
			return nil
		}
		ptr := alloc()
		if string(v) == "true" {
			return ptr
		}
		_ = json.Unmarshal(v, ptr)
		return ptr
	}

	c.HoverProvider = unmarshalBool("hoverProvider")
	c.DeclarationProvider = unmarshalBool("declarationProvider")
	c.DefinitionProvider = unmarshalBool("definitionProvider")
	c.TypeDefinitionProvider = unmarshalBool("typeDefinitionProvider")
	c.ImplementationProvider = unmarshalBool("implementationProvider")
	c.ReferencesProvider = unmarshalBool("referencesProvider")
	c.DocumentHighlightProvider = unmarshalBool("documentHighlightProvider")
	c.DocumentSymbolProvider = unmarshalBool("documentSymbolProvider")
	c.DocumentFormattingProvider = unmarshalBool("documentFormattingProvider")
	c.DocumentRangeFormattingProvider = unmarshalBool("documentRangeFormattingProvider")
	c.FoldingRangeProvider = unmarshalBool("foldingRangeProvider")
	c.SelectionRangeProvider = unmarshalBool("selectionRangeProvider")
	c.WorkspaceSymbolProvider = unmarshalBool("workspaceSymbolProvider")
	c.CallHierarchyProvider = unmarshalBool("callHierarchyProvider")
	c.DocumentColorProvider = unmarshalBool("documentColorProvider")
	c.LinkedEditingRangeProvider = unmarshalBool("linkedEditingRangeProvider")
	c.MonikerProvider = unmarshalBool("monikerProvider")
	c.TypeHierarchyProvider = unmarshalBool("typeHierarchyProvider")
	c.InlineValueProvider = unmarshalBool("inlineValueProvider")

	if p := unmarshalPtrOrBool("completionProvider", func() any { return &CompletionOptions{} }); p != nil {
		c.CompletionProvider = p.(*CompletionOptions)
	}
	if p := unmarshalPtrOrBool("signatureHelpProvider", func() any { return &SignatureHelpOptions{} }); p != nil {
		c.SignatureHelpProvider = p.(*SignatureHelpOptions)
	}
	if p := unmarshalPtrOrBool("codeActionProvider", func() any { return &CodeActionOptions{} }); p != nil {
		c.CodeActionProvider = p.(*CodeActionOptions)
	}
	if p := unmarshalPtrOrBool("codeLensProvider", func() any { return &CodeLensOptions{} }); p != nil {
		c.CodeLensProvider = p.(*CodeLensOptions)
	}
	if p := unmarshalPtrOrBool("renameProvider", func() any { return &RenameOptions{} }); p != nil {
		c.RenameProvider = p.(*RenameOptions)
	}
	if p := unmarshalPtrOrBool("semanticTokensProvider", func() any { return &SemanticTokensOptions{} }); p != nil {
		c.SemanticTokensProvider = p.(*SemanticTokensOptions)
	}
	if p := unmarshalPtrOrBool("inlayHintProvider", func() any { return &InlayHintOptions{} }); p != nil {
		c.InlayHintProvider = p.(*InlayHintOptions)
	}

	// These are always objects, never bools.
	c.ExecuteCommandProvider = nil
	unmarshalPtr("executeCommandProvider", &c.ExecuteCommandProvider)
	c.DiagnosticProvider = nil
	unmarshalPtr("diagnosticProvider", &c.DiagnosticProvider)
	c.DocumentLinkProvider = nil
	if p := unmarshalPtrOrBool("documentLinkProvider", func() any { return &DocumentLinkOptions{} }); p != nil {
		c.DocumentLinkProvider = p.(*DocumentLinkOptions)
	}
	c.DocumentOnTypeFormattingProvider = nil
	unmarshalPtr("documentOnTypeFormattingProvider", &c.DocumentOnTypeFormattingProvider)

	return nil
}

// These types handle LSP methods that can return multiple different types.

// LocationResult represents the result of definition/declaration/typeDefinition/implementation.
// Per LSP spec: Location | Location[] | LocationLink[] | null
type LocationResult struct {
	Location      *Location      `json:"-"`
	Locations     []Location     `json:"-"`
	LocationLinks []LocationLink `json:"-"`
}

// MarshalJSON implements json.Marshaler for LocationResult.
func (r LocationResult) MarshalJSON() ([]byte, error) {
	switch {
	case len(r.LocationLinks) > 0:
		return json.Marshal(r.LocationLinks)
	case r.Locations != nil:
		return json.Marshal(r.Locations)
	case r.Location != nil:
		return json.Marshal(r.Location)
	}
	return []byte("null"), nil
}

// UnmarshalJSON implements json.Unmarshaler for LocationResult.
func (r *LocationResult) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	// Try single Location first
	if data[0] == '{' {
		var loc Location
		if err := json.Unmarshal(data, &loc); err == nil {
			r.Location = &loc
			return nil
		}
	}
	// Try array
	if data[0] == '[' {
		// Check first element to determine type
		var raw []json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		if len(raw) == 0 {
			r.Locations = []Location{}
			return nil
		}
		// Check if first element has "targetUri" (LocationLink) or "uri" (Location)
		var probe struct {
			TargetURI string `json:"targetUri"`
		}
		if json.Unmarshal(raw[0], &probe) == nil && probe.TargetURI != "" {
			return json.Unmarshal(data, &r.LocationLinks)
		}
		return json.Unmarshal(data, &r.Locations)
	}
	return nil
}

// DocumentSymbolResult represents the result of documentSymbol request.
// Per LSP spec: DocumentSymbol[] | SymbolInformation[] | null
type DocumentSymbolResult struct {
	DocumentSymbols   []DocumentSymbol    `json:"-"`
	SymbolInformation []SymbolInformation `json:"-"`
}

// MarshalJSON implements json.Marshaler for DocumentSymbolResult.
func (r DocumentSymbolResult) MarshalJSON() ([]byte, error) {
	if len(r.DocumentSymbols) > 0 {
		return json.Marshal(r.DocumentSymbols)
	}
	if len(r.SymbolInformation) > 0 {
		return json.Marshal(r.SymbolInformation)
	}
	return []byte("null"), nil
}

// UnmarshalJSON implements json.Unmarshaler for DocumentSymbolResult.
func (r *DocumentSymbolResult) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	if data[0] != '[' {
		return nil
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) == 0 {
		r.DocumentSymbols = []DocumentSymbol{}
		return nil
	}
	// Check if first element has "selectionRange" (DocumentSymbol) or "location" (SymbolInformation)
	var probe struct {
		SelectionRange *Range    `json:"selectionRange"`
		Location       *Location `json:"location"`
	}
	if err := json.Unmarshal(raw[0], &probe); err != nil {
		return err
	}
	if probe.SelectionRange != nil {
		return json.Unmarshal(data, &r.DocumentSymbols)
	}
	return json.Unmarshal(data, &r.SymbolInformation)
}

// CodeActionResult represents an item in the codeAction response array.
// Per LSP spec: (Command | CodeAction)[]
type CodeActionResult struct {
	Command    *Command    `json:"-"`
	CodeAction *CodeAction `json:"-"`
}

// MarshalJSON implements json.Marshaler for CodeActionResult.
func (r CodeActionResult) MarshalJSON() ([]byte, error) {
	if r.CodeAction != nil {
		return json.Marshal(r.CodeAction)
	}
	if r.Command != nil {
		return json.Marshal(r.Command)
	}
	return []byte("null"), nil
}

// UnmarshalJSON implements json.Unmarshaler for CodeActionResult.
func (r *CodeActionResult) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	// Check if it has "title" and "command" fields (could be either)
	// CodeAction has "title" and optionally "kind", "edit", "command" (nested)
	// Command has "title", "command" (string), "arguments"
	var probe struct {
		Title   string `json:"title"`
		Command any    `json:"command"` // string for Command, *Command for CodeAction
		Kind    string `json:"kind"`
		Edit    any    `json:"edit"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	// If "command" is a string, it's a Command; if object or absent with kind/edit, it's CodeAction
	if probe.Kind != "" || probe.Edit != nil {
		r.CodeAction = &CodeAction{}
		return json.Unmarshal(data, r.CodeAction)
	}
	// Check if command is a string
	var cmdProbe struct {
		Command string `json:"command"`
	}
	if json.Unmarshal(data, &cmdProbe) == nil && cmdProbe.Command != "" {
		// Could be either - if there's no "kind" or "edit", treat as Command
		r.Command = &Command{}
		return json.Unmarshal(data, r.Command)
	}
	// Default to CodeAction
	r.CodeAction = &CodeAction{}
	return json.Unmarshal(data, r.CodeAction)
}

// SemanticTokensResult represents the result of semanticTokens/full/delta request.
// Per LSP spec: SemanticTokens | SemanticTokensDelta | null
type SemanticTokensResult struct {
	SemanticTokens      *SemanticTokens      `json:"-"`
	SemanticTokensDelta *SemanticTokensDelta `json:"-"`
}

// MarshalJSON implements json.Marshaler for SemanticTokensResult.
func (r SemanticTokensResult) MarshalJSON() ([]byte, error) {
	if r.SemanticTokensDelta != nil {
		return json.Marshal(r.SemanticTokensDelta)
	}
	if r.SemanticTokens != nil {
		return json.Marshal(r.SemanticTokens)
	}
	return []byte("null"), nil
}

// UnmarshalJSON implements json.Unmarshaler for SemanticTokensResult.
func (r *SemanticTokensResult) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	// SemanticTokensDelta has "edits" field, SemanticTokens has "data" field
	var probe struct {
		Edits []any `json:"edits"`
		Data  []any `json:"data"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	if probe.Edits != nil {
		r.SemanticTokensDelta = &SemanticTokensDelta{}
		return json.Unmarshal(data, r.SemanticTokensDelta)
	}
	r.SemanticTokens = &SemanticTokens{}
	return json.Unmarshal(data, r.SemanticTokens)
}

// InitializeParams contains the parameters for the Initialize request.
type InitializeParams struct {
	ProcessID         *int              `json:"processId"`
	RootURI           string            `json:"rootUri"`
	Capabilities      json.RawMessage   `json:"capabilities,omitempty"`
	WorkspaceFolders  []WorkspaceFolder `json:"workspaceFolders,omitempty"`
	Trace             TraceValue        `json:"trace,omitempty"`
	InitializeOptions json.RawMessage   `json:"initializationOptions,omitempty"`
	WorkDoneToken     *ProgressToken    `json:"workDoneToken,omitempty"`
}

// InitializeResult contains the result of the Initialize request.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

// DidOpenTextDocumentParams contains the parameters for the DidOpen notification.
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// TextDocumentContentChangeEvent describes a change to a text document.
type TextDocumentContentChangeEvent struct {
	Range *Range `json:"range,omitempty"`
	Text  string `json:"text"`
}

// DidChangeTextDocumentParams contains the parameters for the DidChange notification.
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// DidCloseTextDocumentParams contains the parameters for the DidClose notification.
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DidSaveTextDocumentParams contains the parameters for the DidSave notification.
type DidSaveTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Text         string                 `json:"text,omitempty"`
}

// CompletionParams contains the parameters for a Completion request.
type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      *CompletionContext     `json:"context,omitempty"`
}

// HoverParams contains the parameters for a Hover request.
type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// SignatureHelpParams contains the parameters for a SignatureHelp request.
type SignatureHelpParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// DefinitionParams contains the parameters for a Definition request.
type DefinitionParams struct {
	TextDocument  TextDocumentIdentifier `json:"textDocument"`
	Position      Position               `json:"position"`
	WorkDoneToken *ProgressToken         `json:"workDoneToken,omitempty"`
}

// DeclarationParams contains the parameters for a Declaration request.
type DeclarationParams struct {
	TextDocument  TextDocumentIdentifier `json:"textDocument"`
	Position      Position               `json:"position"`
	WorkDoneToken *ProgressToken         `json:"workDoneToken,omitempty"`
}

// TypeDefinitionParams contains the parameters for a TypeDefinition request.
type TypeDefinitionParams struct {
	TextDocument  TextDocumentIdentifier `json:"textDocument"`
	Position      Position               `json:"position"`
	WorkDoneToken *ProgressToken         `json:"workDoneToken,omitempty"`
}

// ImplementationParams contains the parameters for an Implementation request.
type ImplementationParams struct {
	TextDocument  TextDocumentIdentifier `json:"textDocument"`
	Position      Position               `json:"position"`
	WorkDoneToken *ProgressToken         `json:"workDoneToken,omitempty"`
}

// ReferenceParams contains the parameters for a References request.
type ReferenceParams struct {
	TextDocument  TextDocumentIdentifier `json:"textDocument"`
	Position      Position               `json:"position"`
	Context       ReferenceContext       `json:"context"`
	WorkDoneToken *ProgressToken         `json:"workDoneToken,omitempty"`
}

// DocumentHighlightParams contains the parameters for a DocumentHighlight request.
type DocumentHighlightParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// DocumentSymbolParams contains the parameters for a DocumentSymbol request.
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// CodeActionParams contains the parameters for a CodeAction request.
type CodeActionParams struct {
	TextDocument  TextDocumentIdentifier `json:"textDocument"`
	Range         Range                  `json:"range"`
	Context       CodeActionContext      `json:"context"`
	WorkDoneToken *ProgressToken         `json:"workDoneToken,omitempty"`
}

// CodeLensParams contains the parameters for a CodeLens request.
type CodeLensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DocumentFormattingParams contains the parameters for a Formatting request.
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

// DocumentRangeFormattingParams contains the parameters for a RangeFormatting request.
type DocumentRangeFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Options      FormattingOptions      `json:"options"`
}

// RenameParams contains the parameters for a Rename request.
type RenameParams struct {
	TextDocument  TextDocumentIdentifier `json:"textDocument"`
	Position      Position               `json:"position"`
	NewName       string                 `json:"newName"`
	WorkDoneToken *ProgressToken         `json:"workDoneToken,omitempty"`
}

// PrepareRenameParams contains the parameters for a PrepareRename request.
type PrepareRenameParams struct {
	TextDocument  TextDocumentIdentifier `json:"textDocument"`
	Position      Position               `json:"position"`
	WorkDoneToken *ProgressToken         `json:"workDoneToken,omitempty"`
}

// PrepareRenameResult contains the result of a PrepareRename request.
// Per LSP spec, it can be: Range | { range, placeholder } | { defaultBehavior }.
type PrepareRenameResult struct {
	Range           Range  `json:"-"`
	Placeholder     string `json:"-"`
	DefaultBehavior bool   `json:"-"`
	IsRangeOnly     bool   `json:"-"` // true if only Range is set (not placeholder)
	IsDefault       bool   `json:"-"` // true if defaultBehavior variant
}

// MarshalJSON implements json.Marshaler for PrepareRenameResult.
func (p PrepareRenameResult) MarshalJSON() ([]byte, error) {
	if p.IsDefault {
		return json.Marshal(struct {
			DefaultBehavior bool `json:"defaultBehavior"`
		}{DefaultBehavior: p.DefaultBehavior})
	}
	if p.IsRangeOnly {
		return json.Marshal(p.Range)
	}
	return json.Marshal(struct {
		Range       Range  `json:"range"`
		Placeholder string `json:"placeholder"`
	}{Range: p.Range, Placeholder: p.Placeholder})
}

// UnmarshalJSON implements json.Unmarshaler for PrepareRenameResult.
func (p *PrepareRenameResult) UnmarshalJSON(data []byte) error {
	// Try { defaultBehavior } first
	var defBehavior struct {
		DefaultBehavior bool `json:"defaultBehavior"`
	}
	if err := json.Unmarshal(data, &defBehavior); err == nil {
		// Check if it actually has the field (not just zero value)
		var raw map[string]json.RawMessage
		if json.Unmarshal(data, &raw) == nil {
			if _, ok := raw["defaultBehavior"]; ok {
				p.DefaultBehavior = defBehavior.DefaultBehavior
				p.IsDefault = true
				return nil
			}
		}
	}
	// Try { range, placeholder }
	var withPlaceholder struct {
		Range       Range  `json:"range"`
		Placeholder string `json:"placeholder"`
	}
	if err := json.Unmarshal(data, &withPlaceholder); err == nil && withPlaceholder.Placeholder != "" {
		p.Range = withPlaceholder.Range
		p.Placeholder = withPlaceholder.Placeholder
		return nil
	}
	// Try Range only
	var r Range
	if err := json.Unmarshal(data, &r); err == nil {
		p.Range = r
		p.IsRangeOnly = true
		return nil
	}
	return nil
}

// FoldingRangeParams contains the parameters for a FoldingRange request.
type FoldingRangeParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// SelectionRangeParams contains the parameters for a SelectionRange request.
type SelectionRangeParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Positions    []Position             `json:"positions"`
}

// SemanticTokensParams contains the parameters for a SemanticTokensFull request.
type SemanticTokensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// SemanticTokensRangeParams contains the parameters for a SemanticTokensRange request.
type SemanticTokensRangeParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
}

// DocumentDiagnosticParams contains the parameters for a Diagnostic request.
type DocumentDiagnosticParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// PreviousResultID pairs a document URI with a previous diagnostic result ID.
type PreviousResultID struct {
	URI   string `json:"uri"`
	Value string `json:"value"`
}

// WorkspaceDiagnosticParams contains the parameters for a WorkspaceDiagnostic request.
type WorkspaceDiagnosticParams struct {
	Identifier        string             `json:"identifier,omitempty"`
	PreviousResultIDs []PreviousResultID `json:"previousResultIds,omitempty"`
	WorkDoneToken     string             `json:"workDoneToken,omitempty"`
}

// WorkspaceDocumentDiagnosticReport is a per-document diagnostic report
// within a workspace diagnostic response.
type WorkspaceDocumentDiagnosticReport struct {
	Kind     string       `json:"kind"`
	ResultID string       `json:"resultId,omitempty"`
	URI      string       `json:"uri"`
	Version  int32        `json:"version,omitempty"`
	Items    []Diagnostic `json:"items,omitempty"`
}

// WorkspaceDiagnosticReport contains the response from a WorkspaceDiagnostic request.
type WorkspaceDiagnosticReport struct {
	Items []WorkspaceDocumentDiagnosticReport `json:"items"`
}

// WorkspaceSymbolParams contains the parameters for a WorkspaceSymbol request.
type WorkspaceSymbolParams struct {
	Query         string         `json:"query"`
	WorkDoneToken *ProgressToken `json:"workDoneToken,omitempty"`
}

// ExecuteCommandParams contains the parameters for an ExecuteCommand request.
type ExecuteCommandParams struct {
	Command       string            `json:"command"`
	Arguments     []json.RawMessage `json:"arguments,omitempty"`
	WorkDoneToken *ProgressToken    `json:"workDoneToken,omitempty"`
}

// CallHierarchyPrepareParams contains the parameters for a PrepareCallHierarchy request.
type CallHierarchyPrepareParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// CallHierarchyIncomingCallsParams contains the parameters for an IncomingCalls request.
type CallHierarchyIncomingCallsParams struct {
	Item CallHierarchyItem `json:"item"`
}

// CallHierarchyOutgoingCallsParams contains the parameters for an OutgoingCalls request.
type CallHierarchyOutgoingCallsParams struct {
	Item CallHierarchyItem `json:"item"`
}

// DocumentColorParams contains the parameters for a DocumentColor request.
type DocumentColorParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// ColorPresentationParams contains the parameters for a ColorPresentation request.
type ColorPresentationParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Color        Color                  `json:"color"`
	Range        Range                  `json:"range"`
}

// DocumentLinkParams contains the parameters for a DocumentLink request.
type DocumentLinkParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DocumentOnTypeFormattingParams contains the parameters for an OnTypeFormatting request.
type DocumentOnTypeFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Character    string                 `json:"ch"`
	Options      FormattingOptions      `json:"options"`
}

// LinkedEditingRangeParams contains the parameters for a LinkedEditingRange request.
type LinkedEditingRangeParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// MonikerParams contains the parameters for a Moniker request.
type MonikerParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// WillSaveTextDocumentParams contains the parameters for a WillSave notification/request.
type WillSaveTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Reason       TextDocumentSaveReason `json:"reason"`
}

// SemanticTokensDeltaParams contains the parameters for a SemanticTokensFullDelta request.
type SemanticTokensDeltaParams struct {
	TextDocument     TextDocumentIdentifier `json:"textDocument"`
	PreviousResultID string                 `json:"previousResultId"`
}

// TypeHierarchyPrepareParams contains the parameters for a PrepareTypeHierarchy request.
type TypeHierarchyPrepareParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// TypeHierarchySupertypesParams contains the parameters for a TypeHierarchySupertypes request.
type TypeHierarchySupertypesParams struct {
	Item TypeHierarchyItem `json:"item"`
}

// TypeHierarchySubtypesParams contains the parameters for a TypeHierarchySubtypes request.
type TypeHierarchySubtypesParams struct {
	Item TypeHierarchyItem `json:"item"`
}

// InlayHintParams contains the parameters for an InlayHint request.
type InlayHintParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
}

// InlineValueParams contains the parameters for an InlineValue request.
type InlineValueParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
}

// DidChangeConfigurationParams contains the parameters for a DidChangeConfiguration notification.
type DidChangeConfigurationParams struct {
	Settings json.RawMessage `json:"settings,omitempty"`
}

// DidChangeWatchedFilesParams contains the parameters for a DidChangeWatchedFiles notification.
type DidChangeWatchedFilesParams struct {
	Changes []FileEvent `json:"changes"`
}

// DidChangeWorkspaceFoldersParams contains the parameters for a DidChangeWorkspaceFolders notification.
type DidChangeWorkspaceFoldersParams struct {
	Event WorkspaceFoldersChangeEvent `json:"event"`
}

// WorkDoneProgressCancelParams contains the parameters for a WorkDoneProgressCancel notification.
type WorkDoneProgressCancelParams struct {
	Token string `json:"token"`
}

// SetTraceParams contains the parameters for a SetTrace notification.
type SetTraceParams struct {
	Value TraceValue `json:"value"`
}

// CreateFilesParams contains the parameters for file creation events.
type CreateFilesParams struct {
	Files []FileCreate `json:"files"`
}

// RenameFilesParams contains the parameters for file rename events.
type RenameFilesParams struct {
	Files []FileRename `json:"files"`
}

// DeleteFilesParams contains the parameters for file deletion events.
type DeleteFilesParams struct {
	Files []FileDelete `json:"files"`
}

// ShowDocumentParams contains the parameters for a ShowDocument request.
type ShowDocumentParams struct {
	URI       string `json:"uri"`
	External  bool   `json:"external,omitempty"`
	TakeFocus bool   `json:"takeFocus,omitempty"`
	Selection *Range `json:"selection,omitempty"`
}

// LogTraceParams contains the parameters for a LogTrace notification.
type LogTraceParams struct {
	Message string `json:"message"`
	Verbose string `json:"verbose,omitempty"`
}

// MessageType enumerates the type of a message (for ShowMessage/LogMessage).
type MessageType int

const (
	MessageTypeError   MessageType = 1
	MessageTypeWarning MessageType = 2
	MessageTypeInfo    MessageType = 3
	MessageTypeLog     MessageType = 4
	MessageTypeDebug   MessageType = 5
)

// ShowMessageParams contains the parameters for a window/showMessage notification.
type ShowMessageParams struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}

// LogMessageParams contains the parameters for a window/logMessage notification.
type LogMessageParams struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}

// PublishDiagnosticsParams contains the parameters for a
// textDocument/publishDiagnostics notification.
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Version     int32        `json:"version,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// ProgressToken is either a string or an integer.
type ProgressToken struct {
	StringValue  string
	IntegerValue int
	IsInteger    bool
}

// MarshalJSON implements json.Marshaler.
func (p ProgressToken) MarshalJSON() ([]byte, error) {
	if p.IsInteger {
		return json.Marshal(p.IntegerValue)
	}
	return json.Marshal(p.StringValue)
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ProgressToken) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		p.IsInteger = false
		return json.Unmarshal(data, &p.StringValue)
	}
	p.IsInteger = true
	return json.Unmarshal(data, &p.IntegerValue)
}

// NewWorkDoneToken returns a new ProgressToken with a UUID string
// value, suitable for use as a workDoneToken in LSP requests.
func NewWorkDoneToken() *ProgressToken {
	return &ProgressToken{StringValue: uuid.NewString()}
}

// ProgressParams contains the parameters for a $/progress notification.
type ProgressParams struct {
	Token ProgressToken   `json:"token"`
	Value json.RawMessage `json:"value"`
}

// WorkDoneProgressBegin is sent when work starts.
type WorkDoneProgressBegin struct {
	Kind        string `json:"kind"` // always "begin"
	Title       string `json:"title"`
	Cancellable bool   `json:"cancellable,omitempty"`
	Message     string `json:"message,omitempty"`
	Percentage  uint32 `json:"percentage,omitempty"`
}

// WorkDoneProgressReport is sent to report progress.
type WorkDoneProgressReport struct {
	Kind        string `json:"kind"` // always "report"
	Cancellable bool   `json:"cancellable,omitempty"`
	Message     string `json:"message,omitempty"`
	Percentage  uint32 `json:"percentage,omitempty"`
}

// WorkDoneProgressEnd is sent when work ends.
type WorkDoneProgressEnd struct {
	Kind    string `json:"kind"` // always "end"
	Message string `json:"message,omitempty"`
}

// WorkDoneProgressCreateParams contains the parameters for
// window/workDoneProgress/create request.
type WorkDoneProgressCreateParams struct {
	Token ProgressToken `json:"token"`
}

// MessageActionItem represents an action item in a ShowMessageRequest.
type MessageActionItem struct {
	Title string `json:"title"`
}

// ShowMessageRequestParams contains the parameters for a
// window/showMessageRequest request.
type ShowMessageRequestParams struct {
	Type    MessageType         `json:"type"`
	Message string              `json:"message"`
	Actions []MessageActionItem `json:"actions,omitempty"`
}

// ApplyWorkspaceEditParams contains the parameters for a
// workspace/applyEdit request.
type ApplyWorkspaceEditParams struct {
	Label string        `json:"label,omitempty"`
	Edit  WorkspaceEdit `json:"edit"`
}

// ApplyWorkspaceEditResult contains the result of a workspace/applyEdit request.
type ApplyWorkspaceEditResult struct {
	Applied       bool   `json:"applied"`
	FailureReason string `json:"failureReason,omitempty"`
}

// ConfigurationItem represents a configuration section to fetch.
type ConfigurationItem struct {
	ScopeURI string `json:"scopeUri,omitempty"`
	Section  string `json:"section,omitempty"`
}

// ConfigurationParams contains the parameters for a workspace/configuration request.
type ConfigurationParams struct {
	Items []ConfigurationItem `json:"items"`
}

// Registration represents a capability registration.
type Registration struct {
	ID              string          `json:"id"`
	Method          string          `json:"method"`
	RegisterOptions json.RawMessage `json:"registerOptions,omitempty"`
}

// RegistrationParams contains the parameters for a client/registerCapability request.
type RegistrationParams struct {
	Registrations []Registration `json:"registrations"`
}

// Unregistration represents a capability unregistration.
type Unregistration struct {
	ID     string `json:"id"`
	Method string `json:"method"`
}

// UnregistrationParams contains the parameters for a
// client/unregisterCapability request.
type UnregistrationParams struct {
	Unregistrations []Unregistration `json:"unregisterations"`
}
