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
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkedStringJSON(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expected       MarkedString
		marshalCompare string
	}{
		{
			name:           "plain string",
			json:           `"hello world"`,
			expected:       MarkedString{Value: "hello world", IsRaw: true},
			marshalCompare: `"hello world"`,
		},
		{
			name:           "object with language",
			json:           `{"language":"go","value":"func main()"}`,
			expected:       MarkedString{Language: "go", Value: "func main()"},
			marshalCompare: `{"language":"go","value":"func main()"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual MarkedString
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			b, err := json.Marshal(actual)
			require.NoError(t, err)
			assert.JSONEq(t, tt.marshalCompare, string(b))
		})
	}
}

func TestHoverJSONUnionType(t *testing.T) {
	tests := []struct {
		name             string
		json             string
		expected         Hover
		marshalCompare   string
		skipMarshalCheck bool
	}{
		{
			name: "MarkupContent",
			json: `{"contents":{"kind":"markdown","value":"# Hello"}}`,
			expected: Hover{
				Contents: MarkupContent{Kind: MarkupKindMarkdown, Value: "# Hello"},
			},
			marshalCompare: `{"contents":{"kind":"markdown","value":"# Hello"}}`,
		},
		{
			name: "MarkedString plain string",
			json: `{"contents":"plain text hover"}`,
			expected: Hover{
				ContentsMarked: &MarkedString{Value: "plain text hover", IsRaw: true},
			},
			marshalCompare: `{"contents":"plain text hover"}`,
		},
		{
			name: "MarkedString object",
			json: `{"contents":{"language":"typescript","value":"const x = 1"}}`,
			expected: Hover{
				ContentsMarked: &MarkedString{Language: "typescript", Value: "const x = 1"},
			},
			marshalCompare: `{"contents":{"language":"typescript","value":"const x = 1"}}`,
		},
		{
			name: "MarkedString array",
			json: `{"contents":["first",{"language":"go","value":"package main"}]}`,
			expected: Hover{
				ContentsMarkedA: []MarkedString{
					{Value: "first", IsRaw: true},
					{Language: "go", Value: "package main"},
				},
			},
			marshalCompare: `{"contents":["first",{"language":"go","value":"package main"}]}`,
		},
		{
			name: "with range",
			json: `{"contents":{"kind":"plaintext","value":"doc"},"range":{"start":{"line":1,"character":2},"end":{"line":1,"character":5}}}`,
			expected: Hover{
				Contents: MarkupContent{Kind: MarkupKindPlainText, Value: "doc"},
				Range:    &Range{Start: Position{Line: 1, Character: 2}, End: Position{Line: 1, Character: 5}},
			},
			marshalCompare: `{"contents":{"kind":"plaintext","value":"doc"},"range":{"start":{"line":1,"character":2},"end":{"line":1,"character":5}}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual Hover
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			if !tt.skipMarshalCheck {
				b, err := json.Marshal(actual)
				require.NoError(t, err)
				assert.JSONEq(t, tt.marshalCompare, string(b))
			}
		})
	}
}

func TestDiagnosticJSONUnionType(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expected       Diagnostic
		marshalCompare string
	}{
		{
			name: "string code",
			json: `{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":5}},"message":"error","code":"E001"}`,
			expected: Diagnostic{
				Range:   Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 5}},
				Message: "error",
				Code:    "E001",
			},
			marshalCompare: `{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":5}},"message":"error","code":"E001"}`,
		},
		{
			name: "integer code",
			json: `{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":5}},"message":"error","code":123}`,
			expected: Diagnostic{
				Range:     Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 5}},
				Message:   "error",
				CodeInt:   123,
				CodeIsInt: true,
			},
			marshalCompare: `{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":5}},"message":"error","code":123}`,
		},
		{
			name: "no code",
			json: `{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":5}},"message":"error"}`,
			expected: Diagnostic{
				Range:   Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 5}},
				Message: "error",
			},
			marshalCompare: `{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":5}},"message":"error"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual Diagnostic
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			b, err := json.Marshal(actual)
			require.NoError(t, err)
			assert.JSONEq(t, tt.marshalCompare, string(b))
		})
	}
}

func TestSignatureInformationJSONUnionType(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expected       SignatureInformation
		marshalCompare string
	}{
		{
			name: "MarkupContent documentation",
			json: `{"label":"func()","documentation":{"kind":"markdown","value":"# Doc"}}`,
			expected: SignatureInformation{
				Label:         "func()",
				Documentation: &MarkupContent{Kind: MarkupKindMarkdown, Value: "# Doc"},
			},
			marshalCompare: `{"label":"func()","documentation":{"kind":"markdown","value":"# Doc"}}`,
		},
		{
			name: "string documentation",
			json: `{"label":"func()","documentation":"plain doc"}`,
			expected: SignatureInformation{
				Label:               "func()",
				DocumentationString: "plain doc",
			},
			marshalCompare: `{"label":"func()","documentation":"plain doc"}`,
		},
		{
			name:           "no documentation",
			json:           `{"label":"func()"}`,
			expected:       SignatureInformation{Label: "func()"},
			marshalCompare: `{"label":"func()"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual SignatureInformation
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			b, err := json.Marshal(actual)
			require.NoError(t, err)
			assert.JSONEq(t, tt.marshalCompare, string(b))
		})
	}
}

func TestParameterInformationJSONUnionType(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expected       ParameterInformation
		marshalCompare string
	}{
		{
			name: "string label and MarkupContent doc",
			json: `{"label":"param","documentation":{"kind":"plaintext","value":"doc"}}`,
			expected: ParameterInformation{
				Label:         "param",
				Documentation: &MarkupContent{Kind: MarkupKindPlainText, Value: "doc"},
			},
			marshalCompare: `{"label":"param","documentation":{"kind":"plaintext","value":"doc"}}`,
		},
		{
			name: "offset label and string doc",
			json: `{"label":[5,10],"documentation":"string doc"}`,
			expected: ParameterInformation{
				LabelOffsets:        &[2]uint32{5, 10},
				DocumentationString: "string doc",
			},
			marshalCompare: `{"label":[5,10],"documentation":"string doc"}`,
		},
		{
			name:           "string label only",
			json:           `{"label":"x"}`,
			expected:       ParameterInformation{Label: "x"},
			marshalCompare: `{"label":"x"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual ParameterInformation
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			b, err := json.Marshal(actual)
			require.NoError(t, err)
			assert.JSONEq(t, tt.marshalCompare, string(b))
		})
	}
}

func TestInlayHintJSONUnionType(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expected       InlayHint
		marshalCompare string
	}{
		{
			name: "string label and MarkupContent tooltip",
			json: `{"position":{"line":1,"character":2},"label":"hint","tooltip":{"kind":"markdown","value":"# tip"}}`,
			expected: InlayHint{
				Position: Position{Line: 1, Character: 2},
				Label:    "hint",
				Tooltip:  &MarkupContent{Kind: MarkupKindMarkdown, Value: "# tip"},
			},
			marshalCompare: `{"position":{"line":1,"character":2},"label":"hint","tooltip":{"kind":"markdown","value":"# tip"}}`,
		},
		{
			name: "string label and string tooltip",
			json: `{"position":{"line":0,"character":0},"label":"x","tooltip":"simple tip"}`,
			expected: InlayHint{
				Position:      Position{Line: 0, Character: 0},
				Label:         "x",
				TooltipString: "simple tip",
			},
			marshalCompare: `{"position":{"line":0,"character":0},"label":"x","tooltip":"simple tip"}`,
		},
		{
			name: "label parts",
			json: `{"position":{"line":0,"character":0},"label":[{"value":"part1"},{"value":"part2"}]}`,
			expected: InlayHint{
				Position:   Position{Line: 0, Character: 0},
				LabelParts: []InlayHintLabelPart{{Value: "part1"}, {Value: "part2"}},
			},
			marshalCompare: `{"position":{"line":0,"character":0},"label":[{"value":"part1"},{"value":"part2"}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual InlayHint
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			b, err := json.Marshal(actual)
			require.NoError(t, err)
			assert.JSONEq(t, tt.marshalCompare, string(b))
		})
	}
}

func TestInlayHintLabelPartJSONUnionType(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expected       InlayHintLabelPart
		marshalCompare string
	}{
		{
			name: "MarkupContent tooltip",
			json: `{"value":"part","tooltip":{"kind":"markdown","value":"# tip"}}`,
			expected: InlayHintLabelPart{
				Value:   "part",
				Tooltip: &MarkupContent{Kind: MarkupKindMarkdown, Value: "# tip"},
			},
			marshalCompare: `{"value":"part","tooltip":{"kind":"markdown","value":"# tip"}}`,
		},
		{
			name: "string tooltip",
			json: `{"value":"part","tooltip":"simple"}`,
			expected: InlayHintLabelPart{
				Value:         "part",
				TooltipString: "simple",
			},
			marshalCompare: `{"value":"part","tooltip":"simple"}`,
		},
		{
			name:           "no tooltip",
			json:           `{"value":"part"}`,
			expected:       InlayHintLabelPart{Value: "part"},
			marshalCompare: `{"value":"part"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual InlayHintLabelPart
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			b, err := json.Marshal(actual)
			require.NoError(t, err)
			assert.JSONEq(t, tt.marshalCompare, string(b))
		})
	}
}

func TestCompletionItemTextEditUnionType(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expected       CompletionItem
		marshalCompare string
	}{
		{
			name: "TextEdit",
			json: `{"label":"foo","textEdit":{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":3}},"newText":"foobar"}}`,
			expected: CompletionItem{
				Label: "foo",
				TextEdit: &TextEdit{
					Range:   Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 3}},
					NewText: "foobar",
				},
			},
			marshalCompare: `{"label":"foo","textEdit":{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":3}},"newText":"foobar"}}`,
		},
		{
			name: "InsertReplaceEdit",
			json: `{"label":"bar","textEdit":{"newText":"barbaz","insert":{"start":{"line":1,"character":0},"end":{"line":1,"character":3}},"replace":{"start":{"line":1,"character":0},"end":{"line":1,"character":6}}}}`,
			expected: CompletionItem{
				Label: "bar",
				TextEditIR: &InsertReplaceEdit{
					NewText: "barbaz",
					Insert:  Range{Start: Position{Line: 1, Character: 0}, End: Position{Line: 1, Character: 3}},
					Replace: Range{Start: Position{Line: 1, Character: 0}, End: Position{Line: 1, Character: 6}},
				},
			},
			marshalCompare: `{"label":"bar","textEdit":{"newText":"barbaz","insert":{"start":{"line":1,"character":0},"end":{"line":1,"character":3}},"replace":{"start":{"line":1,"character":0},"end":{"line":1,"character":6}}}}`,
		},
		{
			name:           "no textEdit",
			json:           `{"label":"baz"}`,
			expected:       CompletionItem{Label: "baz"},
			marshalCompare: `{"label":"baz"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual CompletionItem
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			b, err := json.Marshal(actual)
			require.NoError(t, err)
			assert.JSONEq(t, tt.marshalCompare, string(b))
		})
	}
}

func TestTextDocumentEditUnionType(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expected       TextDocumentEdit
		marshalCompare string
	}{
		{
			name: "TextEdit array",
			json: `{"textDocument":{"uri":"file:///a.go","version":1},"edits":[{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":3}},"newText":"foo"}]}`,
			expected: TextDocumentEdit{
				TextDocument: VersionedTextDocumentIdentifier{URI: "file:///a.go", Version: 1},
				Edits: []TextEdit{
					{Range: Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 3}}, NewText: "foo"},
				},
			},
			marshalCompare: `{"textDocument":{"uri":"file:///a.go","version":1},"edits":[{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":3}},"newText":"foo"}]}`,
		},
		{
			name: "AnnotatedTextEdit array",
			json: `{"textDocument":{"uri":"file:///b.go","version":2},"edits":[{"range":{"start":{"line":1,"character":0},"end":{"line":1,"character":5}},"newText":"bar","annotationId":"ann1"}]}`,
			expected: TextDocumentEdit{
				TextDocument: VersionedTextDocumentIdentifier{URI: "file:///b.go", Version: 2},
				AnnotatedEdits: []AnnotatedTextEdit{
					{Range: Range{Start: Position{Line: 1, Character: 0}, End: Position{Line: 1, Character: 5}}, NewText: "bar", AnnotationID: "ann1"},
				},
				UseAnnotated: true,
			},
			marshalCompare: `{"textDocument":{"uri":"file:///b.go","version":2},"edits":[{"range":{"start":{"line":1,"character":0},"end":{"line":1,"character":5}},"newText":"bar","annotationId":"ann1"}]}`,
		},
		{
			name: "mixed edits",
			json: `{"textDocument":{"uri":"file:///c.go","version":3},"edits":[{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":1}},"newText":"a"},{"range":{"start":{"line":1,"character":0},"end":{"line":1,"character":1}},"newText":"b","annotationId":"id2"}]}`,
			expected: TextDocumentEdit{
				TextDocument: VersionedTextDocumentIdentifier{URI: "file:///c.go", Version: 3},
				Edits: []TextEdit{
					{Range: Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 1}}, NewText: "a"},
				},
				AnnotatedEdits: []AnnotatedTextEdit{
					{Range: Range{Start: Position{Line: 1, Character: 0}, End: Position{Line: 1, Character: 1}}, NewText: "b", AnnotationID: "id2"},
				},
				UseAnnotated: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual TextDocumentEdit
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			if tt.marshalCompare != "" {
				b, err := json.Marshal(actual)
				require.NoError(t, err)
				assert.JSONEq(t, tt.marshalCompare, string(b))
			}
		})
	}
}

func TestDocumentChangeUnionType(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected DocumentChange
	}{
		{
			name: "TextDocumentEdit",
			json: `{"textDocument":{"uri":"file:///a.go","version":1},"edits":[]}`,
			expected: DocumentChange{
				TextDocumentEdit: &TextDocumentEdit{
					TextDocument: VersionedTextDocumentIdentifier{URI: "file:///a.go", Version: 1},
				},
			},
		},
		{
			name: "CreateFile",
			json: `{"kind":"create","uri":"file:///new.go"}`,
			expected: DocumentChange{
				CreateFile: &CreateFile{Kind: "create", URI: "file:///new.go"},
			},
		},
		{
			name: "RenameFile",
			json: `{"kind":"rename","oldUri":"file:///old.go","newUri":"file:///new.go"}`,
			expected: DocumentChange{
				RenameFile: &RenameFile{Kind: "rename", OldURI: "file:///old.go", NewURI: "file:///new.go"},
			},
		},
		{
			name: "DeleteFile",
			json: `{"kind":"delete","uri":"file:///del.go"}`,
			expected: DocumentChange{
				DeleteFile: &DeleteFile{Kind: "delete", URI: "file:///del.go"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual DocumentChange
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			// Test round-trip
			b, err := json.Marshal(actual)
			require.NoError(t, err)
			var roundTrip DocumentChange
			err = json.Unmarshal(b, &roundTrip)
			require.NoError(t, err)
			assert.Equal(t, actual, roundTrip)
		})
	}
}

func TestWorkspaceEditDocumentChanges(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected WorkspaceEdit
	}{
		{
			name: "TextDocumentEdit only",
			json: `{"documentChanges":[{"textDocument":{"uri":"file:///a.go","version":1},"edits":[]}]}`,
			expected: WorkspaceEdit{
				DocumentChanges: []DocumentChange{
					{
						TextDocumentEdit: &TextDocumentEdit{
							TextDocument: VersionedTextDocumentIdentifier{URI: "file:///a.go", Version: 1},
						},
					},
				},
			},
		},
		{
			name: "mixed operations",
			json: `{"documentChanges":[{"textDocument":{"uri":"file:///a.go","version":1},"edits":[]},{"kind":"create","uri":"file:///b.go"},{"kind":"delete","uri":"file:///c.go"}]}`,
			expected: WorkspaceEdit{
				DocumentChanges: []DocumentChange{
					{
						TextDocumentEdit: &TextDocumentEdit{
							TextDocument: VersionedTextDocumentIdentifier{URI: "file:///a.go", Version: 1},
						},
					},
					{CreateFile: &CreateFile{Kind: "create", URI: "file:///b.go"}},
					{DeleteFile: &DeleteFile{Kind: "delete", URI: "file:///c.go"}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual WorkspaceEdit
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestPrepareRenameResultUnionType(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expected       PrepareRenameResult
		marshalCompare string
	}{
		{
			name: "Range with placeholder",
			json: `{"range":{"start":{"line":1,"character":5},"end":{"line":1,"character":10}},"placeholder":"oldName"}`,
			expected: PrepareRenameResult{
				Range:       Range{Start: Position{Line: 1, Character: 5}, End: Position{Line: 1, Character: 10}},
				Placeholder: "oldName",
			},
			marshalCompare: `{"range":{"start":{"line":1,"character":5},"end":{"line":1,"character":10}},"placeholder":"oldName"}`,
		},
		{
			name: "Range only",
			json: `{"start":{"line":2,"character":0},"end":{"line":2,"character":8}}`,
			expected: PrepareRenameResult{
				Range:       Range{Start: Position{Line: 2, Character: 0}, End: Position{Line: 2, Character: 8}},
				IsRangeOnly: true,
			},
			marshalCompare: `{"start":{"line":2,"character":0},"end":{"line":2,"character":8}}`,
		},
		{
			name: "defaultBehavior true",
			json: `{"defaultBehavior":true}`,
			expected: PrepareRenameResult{
				DefaultBehavior: true,
				IsDefault:       true,
			},
			marshalCompare: `{"defaultBehavior":true}`,
		},
		{
			name: "defaultBehavior false",
			json: `{"defaultBehavior":false}`,
			expected: PrepareRenameResult{
				DefaultBehavior: false,
				IsDefault:       true,
			},
			marshalCompare: `{"defaultBehavior":false}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual PrepareRenameResult
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			b, err := json.Marshal(actual)
			require.NoError(t, err)
			assert.JSONEq(t, tt.marshalCompare, string(b))
		})
	}
}

func TestProgressTokenUnionType(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expected       ProgressToken
		marshalCompare string
	}{
		{
			name:           "string token",
			json:           `"progress-1"`,
			expected:       ProgressToken{StringValue: "progress-1"},
			marshalCompare: `"progress-1"`,
		},
		{
			name:           "integer token",
			json:           `42`,
			expected:       ProgressToken{IntegerValue: 42, IsInteger: true},
			marshalCompare: `42`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual ProgressToken
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			b, err := json.Marshal(actual)
			require.NoError(t, err)
			assert.JSONEq(t, tt.marshalCompare, string(b))
		})
	}
}

func TestNewWorkDoneTokenReturnsValidUUID(t *testing.T) {
	token := NewWorkDoneToken()
	require.NotNil(t, token)
	assert.False(t, token.IsInteger)
	assert.NotEmpty(t, token.StringValue)

	_, err := uuid.Parse(token.StringValue)
	require.NoError(t, err)
}

func TestNewWorkDoneTokenSuccessiveCallsAreDistinct(t *testing.T) {
	a := NewWorkDoneToken()
	b := NewWorkDoneToken()
	assert.NotEqual(t, a.StringValue, b.StringValue)
}

func TestCreateFileJSON(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expected       CreateFile
		marshalCompare string
	}{
		{
			name: "with options",
			json: `{"kind":"create","uri":"file:///new.go","options":{"overwrite":true}}`,
			expected: CreateFile{
				Kind:    "create",
				URI:     "file:///new.go",
				Options: &CreateFileOptions{Overwrite: true},
			},
			marshalCompare: `{"kind":"create","uri":"file:///new.go","options":{"overwrite":true}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual CreateFile
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			b, err := json.Marshal(actual)
			require.NoError(t, err)
			assert.JSONEq(t, tt.marshalCompare, string(b))
		})
	}
}

func TestRenameFileJSON(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expected       RenameFile
		marshalCompare string
	}{
		{
			name: "basic",
			json: `{"kind":"rename","oldUri":"file:///old.go","newUri":"file:///new.go"}`,
			expected: RenameFile{
				Kind:   "rename",
				OldURI: "file:///old.go",
				NewURI: "file:///new.go",
			},
			marshalCompare: `{"kind":"rename","oldUri":"file:///old.go","newUri":"file:///new.go"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual RenameFile
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			b, err := json.Marshal(actual)
			require.NoError(t, err)
			assert.JSONEq(t, tt.marshalCompare, string(b))
		})
	}
}

func TestDeleteFileJSON(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expected       DeleteFile
		marshalCompare string
	}{
		{
			name: "with options",
			json: `{"kind":"delete","uri":"file:///del.go","options":{"recursive":true}}`,
			expected: DeleteFile{
				Kind:    "delete",
				URI:     "file:///del.go",
				Options: &DeleteFileOptions{Recursive: true},
			},
			marshalCompare: `{"kind":"delete","uri":"file:///del.go","options":{"recursive":true}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual DeleteFile
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			b, err := json.Marshal(actual)
			require.NoError(t, err)
			assert.JSONEq(t, tt.marshalCompare, string(b))
		})
	}
}

func TestInsertReplaceEditJSON(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expected       InsertReplaceEdit
		marshalCompare string
	}{
		{
			name: "basic",
			json: `{"newText":"hello","insert":{"start":{"line":0,"character":0},"end":{"line":0,"character":2}},"replace":{"start":{"line":0,"character":0},"end":{"line":0,"character":5}}}`,
			expected: InsertReplaceEdit{
				NewText: "hello",
				Insert:  Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 2}},
				Replace: Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 5}},
			},
			marshalCompare: `{"newText":"hello","insert":{"start":{"line":0,"character":0},"end":{"line":0,"character":2}},"replace":{"start":{"line":0,"character":0},"end":{"line":0,"character":5}}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual InsertReplaceEdit
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			b, err := json.Marshal(actual)
			require.NoError(t, err)
			assert.JSONEq(t, tt.marshalCompare, string(b))
		})
	}
}

func TestAnnotatedTextEditJSON(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expected       AnnotatedTextEdit
		marshalCompare string
	}{
		{
			name: "basic",
			json: `{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":3}},"newText":"foo","annotationId":"change-1"}`,
			expected: AnnotatedTextEdit{
				Range:        Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 3}},
				NewText:      "foo",
				AnnotationID: "change-1",
			},
			marshalCompare: `{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":3}},"newText":"foo","annotationId":"change-1"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual AnnotatedTextEdit
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			b, err := json.Marshal(actual)
			require.NoError(t, err)
			assert.JSONEq(t, tt.marshalCompare, string(b))
		})
	}
}

func TestLocationResultUnionType(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected LocationResult
	}{
		{
			name: "single Location",
			json: `{"uri":"file:///a.go","range":{"start":{"line":1,"character":0},"end":{"line":1,"character":5}}}`,
			expected: LocationResult{
				Location: &Location{
					URI:   "file:///a.go",
					Range: Range{Start: Position{Line: 1, Character: 0}, End: Position{Line: 1, Character: 5}},
				},
			},
		},
		{
			name: "Location array",
			json: `[{"uri":"file:///a.go","range":{"start":{"line":1,"character":0},"end":{"line":1,"character":5}}},{"uri":"file:///b.go","range":{"start":{"line":2,"character":0},"end":{"line":2,"character":3}}}]`,
			expected: LocationResult{
				Locations: []Location{
					{URI: "file:///a.go", Range: Range{Start: Position{Line: 1, Character: 0}, End: Position{Line: 1, Character: 5}}},
					{URI: "file:///b.go", Range: Range{Start: Position{Line: 2, Character: 0}, End: Position{Line: 2, Character: 3}}},
				},
			},
		},
		{
			name: "LocationLink array",
			json: `[{"targetUri":"file:///a.go","targetRange":{"start":{"line":1,"character":0},"end":{"line":1,"character":5}},"targetSelectionRange":{"start":{"line":1,"character":0},"end":{"line":1,"character":3}}}]`,
			expected: LocationResult{
				LocationLinks: []LocationLink{
					{
						TargetURI:            "file:///a.go",
						TargetRange:          Range{Start: Position{Line: 1, Character: 0}, End: Position{Line: 1, Character: 5}},
						TargetSelectionRange: Range{Start: Position{Line: 1, Character: 0}, End: Position{Line: 1, Character: 3}},
					},
				},
			},
		},
		{
			name:     "empty array",
			json:     `[]`,
			expected: LocationResult{Locations: []Location{}},
		},
		{
			name:     "null",
			json:     `null`,
			expected: LocationResult{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual LocationResult
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			// Test round-trip (skip null)
			if tt.json != "null" {
				b, err := json.Marshal(actual)
				require.NoError(t, err)
				var roundTrip LocationResult
				err = json.Unmarshal(b, &roundTrip)
				require.NoError(t, err)
				assert.Equal(t, actual, roundTrip)
			}
		})
	}
}

func TestDocumentSymbolResultUnionType(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected DocumentSymbolResult
	}{
		{
			name: "DocumentSymbol array",
			json: `[{"name":"main","kind":12,"range":{"start":{"line":0,"character":0},"end":{"line":10,"character":0}},"selectionRange":{"start":{"line":0,"character":5},"end":{"line":0,"character":9}}}]`,
			expected: DocumentSymbolResult{
				DocumentSymbols: []DocumentSymbol{
					{
						Name:           "main",
						Kind:           SymbolKindFunction,
						Range:          Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 10, Character: 0}},
						SelectionRange: Range{Start: Position{Line: 0, Character: 5}, End: Position{Line: 0, Character: 9}},
					},
				},
			},
		},
		{
			name: "SymbolInformation array",
			json: `[{"name":"main","kind":12,"location":{"uri":"file:///a.go","range":{"start":{"line":0,"character":0},"end":{"line":10,"character":0}}}}]`,
			expected: DocumentSymbolResult{
				SymbolInformation: []SymbolInformation{
					{
						Name: "main",
						Kind: SymbolKindFunction,
						Location: Location{
							URI:   "file:///a.go",
							Range: Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 10, Character: 0}},
						},
					},
				},
			},
		},
		{
			name:     "empty array",
			json:     `[]`,
			expected: DocumentSymbolResult{DocumentSymbols: []DocumentSymbol{}},
		},
		{
			name:     "null",
			json:     `null`,
			expected: DocumentSymbolResult{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual DocumentSymbolResult
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			// Test round-trip (skip null and empty)
			if tt.json != "null" && (len(tt.expected.DocumentSymbols) > 0 ||
				len(tt.expected.SymbolInformation) > 0) {
				b, err := json.Marshal(actual)
				require.NoError(t, err)
				assert.JSONEq(t, tt.json, string(b))
			}
		})
	}
}

func TestCompletionResultUnionType(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected CompletionResult
	}{
		{
			name: "CompletionList",
			json: `{"isIncomplete":true,"items":[{"label":"foo"},{"label":"bar"}]}`,
			expected: CompletionResult{
				IsIncomplete: true,
				Items:        []CompletionItem{{Label: "foo"}, {Label: "bar"}},
			},
		},
		{
			name: "CompletionItem array",
			json: `[{"label":"foo"},{"label":"bar"},{"label":"baz"}]`,
			expected: CompletionResult{
				Items:       []CompletionItem{{Label: "foo"}, {Label: "bar"}, {Label: "baz"}},
				IsItemsOnly: true,
			},
		},
		{
			name: "empty CompletionList",
			json: `{"isIncomplete":false,"items":[]}`,
			expected: CompletionResult{
				IsIncomplete: false,
				Items:        []CompletionItem{},
			},
		},
		{
			name:     "null",
			json:     `null`,
			expected: CompletionResult{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual CompletionResult
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			// Test round-trip (skip null)
			if tt.json != "null" {
				b, err := json.Marshal(actual)
				require.NoError(t, err)
				var roundTrip CompletionResult
				err = json.Unmarshal(b, &roundTrip)
				require.NoError(t, err)
				assert.Equal(t, actual.IsIncomplete, roundTrip.IsIncomplete)
				assert.Equal(t, len(actual.Items), len(roundTrip.Items))
			}
		})
	}
}

func TestCodeActionResultUnionType(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected CodeActionResult
	}{
		{
			name: "CodeAction with kind",
			json: `{"title":"Organize Imports","kind":"source.organizeImports"}`,
			expected: CodeActionResult{
				CodeAction: &CodeAction{Title: "Organize Imports", Kind: "source.organizeImports"},
			},
		},
		{
			name: "CodeAction with edit",
			json: `{"title":"Fix typo","edit":{"changes":{}}}`,
			expected: CodeActionResult{
				CodeAction: &CodeAction{Title: "Fix typo", Edit: &WorkspaceEdit{Changes: map[string][]TextEdit{}}},
			},
		},
		{
			name: "Command",
			json: `{"title":"Run Test","command":"test.run","arguments":[]}`,
			expected: CodeActionResult{
				Command: &Command{Title: "Run Test", Command: "test.run", Arguments: []json.RawMessage{}},
			},
		},
		{
			name:     "null",
			json:     `null`,
			expected: CodeActionResult{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual CodeActionResult
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			// Test round-trip (skip null)
			if tt.json != "null" {
				b, err := json.Marshal(actual)
				require.NoError(t, err)
				var roundTrip CodeActionResult
				err = json.Unmarshal(b, &roundTrip)
				require.NoError(t, err)
				assert.Equal(t, actual.Command != nil, roundTrip.Command != nil)
				assert.Equal(t, actual.CodeAction != nil, roundTrip.CodeAction != nil)
			}
		})
	}
}

func TestCodeActionResultArray(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected []CodeActionResult
	}{
		{
			name: "mixed Command and CodeAction",
			json: `[{"title":"Run","command":"test.run"},{"title":"Fix","kind":"quickfix"}]`,
			expected: []CodeActionResult{
				{Command: &Command{Title: "Run", Command: "test.run"}},
				{CodeAction: &CodeAction{Title: "Fix", Kind: "quickfix"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual []CodeActionResult
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestSemanticTokensResultUnionType(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected SemanticTokensResult
	}{
		{
			name: "SemanticTokens",
			json: `{"data":[1,2,3,4,5]}`,
			expected: SemanticTokensResult{
				SemanticTokens: &SemanticTokens{Data: []uint32{1, 2, 3, 4, 5}},
			},
		},
		{
			name: "SemanticTokens with resultId",
			json: `{"resultId":"abc123","data":[1,2,3]}`,
			expected: SemanticTokensResult{
				SemanticTokens: &SemanticTokens{ResultID: "abc123", Data: []uint32{1, 2, 3}},
			},
		},
		{
			name: "SemanticTokensDelta",
			json: `{"edits":[{"start":0,"deleteCount":5,"data":[1,2,3]}]}`,
			expected: SemanticTokensResult{
				SemanticTokensDelta: &SemanticTokensDelta{
					Edits: []SemanticTokensEdit{{Start: 0, DeleteCount: 5, Data: []uint32{1, 2, 3}}},
				},
			},
		},
		{
			name: "SemanticTokensDelta with resultId",
			json: `{"resultId":"def456","edits":[]}`,
			expected: SemanticTokensResult{
				SemanticTokensDelta: &SemanticTokensDelta{ResultID: "def456", Edits: []SemanticTokensEdit{}},
			},
		},
		{
			name:     "null",
			json:     `null`,
			expected: SemanticTokensResult{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual SemanticTokensResult
			err := json.Unmarshal([]byte(tt.json), &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

			// Test round-trip (skip null)
			if tt.json != "null" {
				b, err := json.Marshal(actual)
				require.NoError(t, err)
				var roundTrip SemanticTokensResult
				err = json.Unmarshal(b, &roundTrip)
				require.NoError(t, err)
				assert.Equal(t, actual.SemanticTokens != nil, roundTrip.SemanticTokens != nil)
				assert.Equal(t, actual.SemanticTokensDelta != nil, roundTrip.SemanticTokensDelta != nil)
			}
		})
	}
}
