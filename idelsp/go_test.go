// Unstable Build LLC ("COMPANY") CONFIDENTIAL
//
// Unpublished Copyright (c) 2017-2026 Unstable Build, All Rights Reserved.
//
// NOTICE: All information contained herein is, and remains the property of COMPANY.
// The intellectual and technical concepts contained herein are proprietary to
// COMPANY and may be covered by U.S. and Foreign Patents, patents in process,
// and are protected by trade secret or copyright law. Dissemination of this information
// or reproduction of this material is strictly forbidden unless prior written permission
// is obtained from COMPANY. Access to the source code contained herein is hereby
// forbidden to anyone except current COMPANY employees, managers or contractors who
// have executed Confidentiality and Non-disclosure agreements explicitly covering such access.
//
// The copyright notice above does not evidence any actual or intended publication or
// disclosure of this source code, which includes information that is confidential and/or
// proprietary, and is a trade secret, of COMPANY. ANY REPRODUCTION, MODIFICATION,
// DISTRIBUTION, PUBLIC  PERFORMANCE, OR PUBLIC DISPLAY OF OR THROUGH USE OF THIS SOURCE CODE
// WITHOUT  THE EXPRESS WRITTEN CONSENT OF COMPANY IS STRICTLY PROHIBITED, AND IN
// VIOLATION OF APPLICABLE LAWS AND INTERNATIONAL TREATIES. THE RECEIPT OR POSSESSION OF
// THIS SOURCE CODE AND/OR RELATED INFORMATION DOES NOT CONVEY OR IMPLY ANY RIGHTS TO
// REPRODUCE, DISCLOSE OR DISTRIBUTE ITS CONTENTS, OR TO MANUFACTURE, USE, OR SELL
// ANYTHING THAT IT MAY DESCRIBE, IN WHOLE OR IN PART.

package idelsp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"unstable.build/go-tui/text"
	"unstable.build/go-tui/text/vi"
	"unstable.build/go-tui/workspace"

	"github.com/unstablebuild/blue/iterator"
	"github.com/unstablebuild/rune-go-sdk/api/config"
	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
)

func TestE2E(t *testing.T) {
	t.Parallel()
	goplsBin := findGopls(t)
	tmpDir := setupTestWorkspace(t, "go")

	uri := makeURI(t, "file://"+tmpDir)
	scheme, err := workspace.NewFileScheme(context.Background(), config.NopConfig(), uri)
	require.NoError(t, err)
	w := workspace.NewSchemeWorkspace(uri, scheme)
	textcomp, err := text.NewComponent(vi.Editor(), w, text.DefaultConfig())
	require.NoError(t, err)

	mainPath := filepath.Join(tmpDir, "main.go")
	mainContent, err := os.ReadFile(mainPath)
	require.NoError(t, err)

	mainTestPath := filepath.Join(tmpDir, "main_test.go")
	mainTestContent, err := os.ReadFile(mainTestPath)
	require.NoError(t, err)

	utilPath := filepath.Join(tmpDir, "util.go")
	utilContent, err := os.ReadFile(utilPath)
	require.NoError(t, err)

	mainURI := "file://" + mainPath
	utilURI := "file://" + utilPath
	testURI := "file://" + mainTestPath

	mainWSURI, err := workspaceapi.ParseURI(mainURI)
	require.NoError(t, err)
	utilWSURI, err := workspaceapi.ParseURI(utilURI)
	require.NoError(t, err)
	testWSURI, err := workspaceapi.ParseURI(testURI)
	require.NoError(t, err)

	tests := []struct {
		name string
		fn   func(t *testing.T, mgr *Manager)
	}{
		{
			name: "Hover",
			fn: func(t *testing.T, mgr *Manager) {
				result, err := mgr.Hover(t.Context(),
					semanticapi.HoverParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
						Position: semanticapi.Position{
							Line: 33, Character: 20,
						},
					},
				)
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, semanticapi.MarkupKindMarkdown, result.Contents.Kind)
				assert.Contains(t, result.Contents.Value,
					"func (g *Greeter) Greet() string")
				assert.Contains(t, result.Contents.Value,
					"Greet returns a greeting message.")
				require.NotNil(t, result.Range)
				assert.Equal(t, semanticapi.Range{
					Start: semanticapi.Position{Line: 33, Character: 18},
					End:   semanticapi.Position{Line: 33, Character: 23},
				}, *result.Range)
			},
		},
		{
			name: "Definition",
			fn: func(t *testing.T, mgr *Manager) {
				result, err := mgr.Definition(t.Context(),
					semanticapi.DefinitionParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
						Position: semanticapi.Position{
							Line: 45, Character: 13,
						},
					},
				)
				locs := result.Locations
				require.NoError(t, err)
				require.Len(t, locs, 1)
				assert.Equal(t, mainURI, locs[0].URI)
				assert.Equal(t, semanticapi.Range{
					Start: semanticapi.Position{Line: 38, Character: 5},
					End:   semanticapi.Position{Line: 38, Character: 8},
				}, locs[0].Range)
			},
		},
		{
			name: "References",
			fn: func(t *testing.T, mgr *Manager) {
				// Add references
				locs, err := mgr.References(t.Context(),
					semanticapi.ReferenceParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
						Position: semanticapi.Position{
							Line: 38, Character: 6,
						},
						Context: semanticapi.ReferenceContext{
							IncludeDeclaration: true,
						},
					},
				)
				require.NoError(t, err)
				require.Len(t, locs, 4)
				assert.ElementsMatch(t, []semanticapi.Location{
					{
						URI: mainURI,
						Range: semanticapi.Range{
							Start: semanticapi.Position{Line: 38, Character: 5},
							End:   semanticapi.Position{Line: 38, Character: 8},
						},
					},
					{
						URI: mainURI,
						Range: semanticapi.Range{
							Start: semanticapi.Position{Line: 45, Character: 13},
							End:   semanticapi.Position{Line: 45, Character: 16},
						},
					},
					{
						URI: testURI,
						Range: semanticapi.Range{
							Start: semanticapi.Position{Line: 35, Character: 2},
							End:   semanticapi.Position{Line: 35, Character: 5},
						},
					},
					{
						URI: testURI,
						Range: semanticapi.Range{
							Start: semanticapi.Position{Line: 28, Character: 4},
							End:   semanticapi.Position{Line: 28, Character: 7},
						},
					},
				}, locs)
			},
		},
		{
			name: "DocumentSymbol",
			fn: func(t *testing.T, mgr *Manager) {
				result, err := mgr.DocumentSymbol(t.Context(),
					semanticapi.DocumentSymbolParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
					},
				)
				syms := result.SymbolInformation
				require.NoError(t, err)
				require.Len(t, syms, 7)
				names := make([]string, len(syms))
				for i, s := range syms {
					names[i] = s.Name
				}
				assert.Contains(t, names, "Greeter")
				assert.Contains(t, names, "Add")
				assert.Contains(t, names, "main")
				assert.Equal(t, semanticapi.SymbolInformation{
					Name: "Add",
					Kind: semanticapi.SymbolKindFunction,
					Location: semanticapi.Location{
						URI: mainURI,
						Range: semanticapi.Range{
							Start: semanticapi.Position{
								Character: 0,
								Line:      38,
							},
							End: semanticapi.Position{
								Character: 1,
								Line:      40,
							},
						},
					},
				}, syms[2])
			},
		},
		{
			name: "WorkspaceSymbol",
			fn: func(t *testing.T, mgr *Manager) {
				syms, err := mgr.WorkspaceSymbol(t.Context(),
					semanticapi.WorkspaceSymbolParams{
						Query: "main.Add",
					},
				)
				require.NoError(t, err)
				require.NotEmpty(t, syms)
				assert.Equal(t, semanticapi.SymbolInformation{
					Name: "main.Add",
					Kind: semanticapi.SymbolKindFunction,
					Location: semanticapi.Location{
						URI: mainURI,
						Range: semanticapi.Range{
							Start: semanticapi.Position{
								Character: 5,
								Line:      38,
							},
							End: semanticapi.Position{
								Character: 8,
								Line:      38,
							},
						},
					},
				}, syms[0])
			},
		},
		{
			name: "Completion",
			fn: func(t *testing.T, mgr *Manager) {
				result, err := mgr.Completion(t.Context(),
					semanticapi.CompletionParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
						Position: semanticapi.Position{
							Line: 34, Character: 12,
						},
					},
				)
				require.NoError(t, err)
				require.NotEmpty(t, result.Items)
				var sprintf *semanticapi.CompletionItem
				for i := range result.Items {
					if result.Items[i].Label == "Sprintf" {
						sprintf = &result.Items[i]
						break
					}
				}
				require.NotNil(t, sprintf, "expected Sprintf in completion list")
				assert.Equal(t, semanticapi.CompletionItemKindFunction, sprintf.Kind)
			},
		},
		{
			name: "Formatting",
			fn: func(t *testing.T, mgr *Manager) {
				edits, err := mgr.Formatting(t.Context(),
					semanticapi.DocumentFormattingParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
						Options: semanticapi.FormattingOptions{
							TabSize:      4,
							InsertSpaces: false,
						},
					},
				)
				require.NoError(t, err)
				assert.Empty(t, edits)
			},
		},
		{
			name: "FoldingRange",
			fn: func(t *testing.T, mgr *Manager) {
				ranges, err := mgr.FoldingRange(t.Context(),
					semanticapi.FoldingRangeParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
					},
				)
				require.NoError(t, err)
				require.NotEmpty(t, ranges)
				var addFold *semanticapi.FoldingRange
				for i := range ranges {
					if ranges[i].StartLine == 38 {
						addFold = &ranges[i]
						break
					}
				}
				require.NotNil(t, addFold, "expected folding range for Add")
				assert.Equal(t, uint32(38), addFold.StartLine)
			},
		},
		{
			// references to the symbol scoped to this file, like references
			// but only for the same file.
			name: "DocumentHighlight",
			fn: func(t *testing.T, mgr *Manager) {
				highlights, err := mgr.DocumentHighlight(t.Context(),
					semanticapi.DocumentHighlightParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
						Position: semanticapi.Position{
							Line: 38, Character: 5,
						},
					},
				)
				require.NoError(t, err)
				require.NotEmpty(t, highlights)
				assert.GreaterOrEqual(t, len(highlights), 2)
			},
		},
		{
			// test that we can rename at position before calling Rename
			name: "PrepareRename",
			fn: func(t *testing.T, mgr *Manager) {
				result, err := mgr.PrepareRename(t.Context(),
					semanticapi.PrepareRenameParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
						Position: semanticapi.Position{
							Line: 38, Character: 5,
						},
					},
				)
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, "Add", result.Placeholder)
				assert.Equal(t, semanticapi.Range{
					Start: semanticapi.Position{Line: 38, Character: 5},
					End:   semanticapi.Position{Line: 38, Character: 8},
				}, result.Range)
			},
		},
		{
			name: "Rename",
			fn: func(t *testing.T, mgr *Manager) {
				edit, err := mgr.Rename(t.Context(),
					semanticapi.RenameParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: utilURI,
						},
						Position: semanticapi.Position{
							Line: 26, Character: 5,
						},
						NewName: "Mul",
					},
				)
				require.NoError(t, err)
				require.NotNil(t, edit)
				require.NotEmpty(t, edit.Changes)
				edits, ok := edit.Changes[utilURI]
				require.True(t, ok, "expected edits in util.go")
				require.NotEmpty(t, edits)
				for _, e := range edits {
					assert.Equal(t, "Mul", e.NewText)
				}
			},
		},
		{
			// given a set of positions, selection ranges that user
			// might be interested in selecting
			name: "SelectionRange",
			fn: func(t *testing.T, mgr *Manager) {
				ranges, err := mgr.SelectionRange(t.Context(),
					semanticapi.SelectionRangeParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
						Positions: []semanticapi.Position{
							{Line: 39, Character: 1},
						},
					},
				)
				require.NoError(t, err)
				require.Len(t, ranges, 1)
			},
		},
		{
			name: "SemanticTokensRange",
			fn: func(t *testing.T, mgr *Manager) {
				tokens, err := mgr.SemanticTokensRange(t.Context(),
					semanticapi.SemanticTokensRangeParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
						Range: semanticapi.Range{
							Start: semanticapi.Position{Line: 9, Character: 0},
							End:   semanticapi.Position{Line: 19, Character: 0},
						},
					},
				)
				require.NoError(t, err)
				require.NotNil(t, tokens)
			},
		},
		{
			// required to call CallHierarchyIncoming/Outgoing calls
			name: "PrepareCallHierarchy",
			fn: func(t *testing.T, mgr *Manager) {
				items, err := mgr.PrepareCallHierarchy(t.Context(),
					semanticapi.CallHierarchyPrepareParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
						Position: semanticapi.Position{
							Line: 38, Character: 5,
						},
					},
				)
				require.NoError(t, err)
				require.Len(t, items, 1)
				assert.Equal(t, "Add", items[0].Name)
				assert.Equal(t, semanticapi.SymbolKindFunction, items[0].Kind)
			},
		},
		{
			// what functions are alling this
			name: "CallHierarchyIncomingCalls",
			fn: func(t *testing.T, mgr *Manager) {
				items, err := mgr.PrepareCallHierarchy(t.Context(),
					semanticapi.CallHierarchyPrepareParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
						Position: semanticapi.Position{
							Line: 38, Character: 5,
						},
					},
				)
				require.NoError(t, err)
				require.NotEmpty(t, items)

				calls, err := mgr.CallHierarchyIncomingCalls(t.Context(),
					semanticapi.CallHierarchyIncomingCallsParams{
						Item: items[0],
					},
				)
				require.NoError(t, err)
				require.Len(t, calls, 3)
				assert.Equal(t, "main", calls[0].From.Name)
			},
		},
		{
			// what functions is this calling
			name: "CallHierarchyOutgoingCalls",
			fn: func(t *testing.T, mgr *Manager) {
				items, err := mgr.PrepareCallHierarchy(t.Context(),
					semanticapi.CallHierarchyPrepareParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
						Position: semanticapi.Position{
							Line: 42, Character: 5,
						},
					},
				)
				require.NoError(t, err)
				require.NotEmpty(t, items)

				calls, err := mgr.CallHierarchyOutgoingCalls(t.Context(),
					semanticapi.CallHierarchyOutgoingCallsParams{
						Item: items[0],
					},
				)
				require.NoError(t, err)
				require.Len(t, calls, 3)
				var found bool
				for _, c := range calls {
					if c.To.Name == "Add" {
						found = true
						break
					}
				}
				assert.True(t, found, "expected outgoing call to Add")
			},
		},
		{
			name: "CodeAction",
			fn: func(t *testing.T, mgr *Manager) {
				actions, err := mgr.CodeAction(t.Context(),
					semanticapi.CodeActionParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
						Range: semanticapi.Range{
							Start: semanticapi.Position{Line: 9, Character: 0},
							End:   semanticapi.Position{Line: 27, Character: 0},
						},
					},
				)
				require.NoError(t, err)
				require.NotEmpty(t, actions)
				assert.Len(t, actions, 5)
			},
		},
		{
			name: "TypeDefinition",
			fn: func(t *testing.T, mgr *Manager) {
				// TypeDefinition of variable "g" at line 43
				result, err := mgr.TypeDefinition(t.Context(),
					semanticapi.TypeDefinitionParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
						Position: semanticapi.Position{
							Line: 43, Character: 1,
						},
					},
				)
				require.NoError(t, err)
				locs := result.Locations
				require.Len(t, locs, 1)
				assert.Equal(t, mainURI, locs[0].URI)
				assert.Equal(t, semanticapi.Range{
					Start: semanticapi.Position{Line: 28, Character: 5},
					End:   semanticapi.Position{Line: 28, Character: 12},
				}, locs[0].Range)
			},
		},
		{
			name: "SignatureHelp",
			fn: func(t *testing.T, mgr *Manager) {
				result, err := mgr.SignatureHelp(t.Context(),
					semanticapi.SignatureHelpParams{
						TextDocument: semanticapi.TextDocumentIdentifier{
							URI: mainURI,
						},
						Position: semanticapi.Position{
							Line: 45, Character: 17,
						},
					},
				)
				require.NoError(t, err)
				assert.ElementsMatch(t, []semanticapi.SignatureInformation{
					{
						Label: "Add(a int, b int) int",
						Documentation: &semanticapi.MarkupContent{
							Kind:  "markdown",
							Value: "Add adds two integers.",
						},
						Parameters: []semanticapi.ParameterInformation{
							{
								Label:         "a int",
								LabelOffsets:  nil,
								Documentation: nil,
							},
							{
								Label:         "b int",
								LabelOffsets:  nil,
								Documentation: nil,
							},
						},
					},
				}, result.Signatures)
			},
		},
	}

	t.Run("lazyly initialized via Handle EventTypeOpen file", func(t *testing.T) {
		t.Parallel()
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var wg sync.WaitGroup
				cfg := Config{
					MaxRetries: 1,
					Callback: &testCallback{
						onShowMessage: func(params semanticapi.ShowMessageParams) {
							if strings.Contains(params.Message, "Finished loading packages") {
								wg.Done()
							}
						},
					},
				}
				mgr := New(
					uri,
					scheme,
					scheme,
					textcomp,
					&stubPkgManager{bin: goplsBin},
					nil, // notifications
					textcomp,
					cfg,
				)

				ctx := context.Background()

				wg.Add(1)
				mgr.Handle(ctx, textapi.Event{
					Type:    textapi.EventTypeOpen,
					URI:     mainWSURI,
					Content: string(mainContent),
				})
				mgr.Handle(ctx, textapi.Event{
					Type:    textapi.EventTypeOpen,
					URI:     utilWSURI,
					Content: string(utilContent),
				})
				mgr.Handle(ctx, textapi.Event{
					Type:    textapi.EventTypeOpen,
					URI:     testWSURI,
					Content: string(mainTestContent),
				})

				wg.Wait()
				tt.fn(t, mgr)
				require.NoError(t, mgr.Close())
			})
		}
	})

	t.Run("initialized via lsp.Initialize", func(t *testing.T) {
		t.Parallel()
		// the following set of the API is only supported via
		// Initialize, w hich allows configuring gopls with custom configuration.
		initializedTests := make([]struct {
			name string
			fn   func(t *testing.T, mgr *Manager)
		}, len(tests))
		copy(initializedTests, tests)
		initializedTests = append(initializedTests, []struct {
			name string
			fn   func(t *testing.T, mgr *Manager)
		}{
			{
				name: "Declaration",
				fn: func(t *testing.T, mgr *Manager) {
					// Declaration of "Add" at its call site (line 36, char 13).
					// gopls doesn't support textDocument/declaration for Go.
					_, err := mgr.Declaration(t.Context(),
						semanticapi.DeclarationParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: mainURI,
							},
							Position: semanticapi.Position{
								Line: 36, Character: 13,
							},
						},
					)
					// gopls returns "method not found" for declaration.
					require.Error(t, err)
					assert.Contains(t, err.Error(), "Declaration")
				},
			},
			{
				name: "Implementation",
				fn: func(t *testing.T, mgr *Manager) {
					// Implementation on Speaker.Speak interface method (line 42, char 1).
					// Should return the Robot.Speak implementation.
					result, err := mgr.Implementation(t.Context(),
						semanticapi.ImplementationParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: mainURI,
							},
							Position: semanticapi.Position{
								Line: 51, Character: 1,
							},
						},
					)
					require.NoError(t, err)
					locs := result.Locations
					require.Len(t, locs, 1)
					assert.Equal(t, mainURI, locs[0].URI)
					// Robot.Speak is at line 60, chars 16-21
					assert.Equal(t, semanticapi.Range{
						Start: semanticapi.Position{Line: 60, Character: 16},
						End:   semanticapi.Position{Line: 60, Character: 21},
					}, locs[0].Range)
				},
			},
			{
				name: "CodeLens",
				fn: func(t *testing.T, mgr *Manager) {
					// Request code lenses for the test file.
					lenses, err := mgr.CodeLens(t.Context(),
						semanticapi.CodeLensParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: testURI,
							},
						},
					)
					require.NoError(t, err)
					expected := []semanticapi.CodeLens{
						{
							Range: semanticapi.Range{
								Start: semanticapi.Position{Line: 23, Character: 0},
								End:   semanticapi.Position{Line: 23, Character: 0},
							},
							Command: &semanticapi.Command{
								Title:   "run file benchmarks",
								Command: "gopls.run_tests",
								Arguments: []json.RawMessage{
									json.RawMessage(
										fmt.Sprintf(`{"URI":"%s","Tests":null,"Benchmarks":["BenchmarkAdd"]}`, testURI)),
								},
							},
						},
						{
							Range: semanticapi.Range{
								Start: semanticapi.Position{Line: 27, Character: 0},
								End:   semanticapi.Position{Line: 27, Character: 0},
							},
							Command: &semanticapi.Command{
								Title:   "run test",
								Command: "gopls.run_tests",
								Arguments: []json.RawMessage{
									json.RawMessage(
										fmt.Sprintf(`{"URI":"%s","Tests":["TestAdd"],"Benchmarks":null}`, testURI)),
								},
							},
						},
						{
							Range: semanticapi.Range{
								Start: semanticapi.Position{Line: 33, Character: 0},
								End:   semanticapi.Position{Line: 33, Character: 0},
							},
							Command: &semanticapi.Command{
								Title:   "run benchmark",
								Command: "gopls.run_tests",
								Arguments: []json.RawMessage{
									json.RawMessage(
										fmt.Sprintf(`{"URI":"%s","Tests":null,"Benchmarks":["BenchmarkAdd"]}`, testURI)),
								},
							},
						},
					}
					require.Equal(t, expected, lenses)

				},
			},
			{
				name: "RangeFormatting",
				fn: func(t *testing.T, mgr *Manager) {
					// RangeFormatting on the Add function (lines 29-31).
					_, err := mgr.RangeFormatting(t.Context(),
						semanticapi.DocumentRangeFormattingParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: mainURI,
							},
							Range: semanticapi.Range{
								Start: semanticapi.Position{Line: 29, Character: 0},
								End:   semanticapi.Position{Line: 32, Character: 0},
							},
							Options: semanticapi.FormattingOptions{
								TabSize:      4,
								InsertSpaces: false,
							},
						},
					)
					// gopls doesn't support rangeFormatting.
					require.Error(t, err)
					assert.Contains(t, err.Error(), "RangeFormatting")
				},
			},
			{
				name: "Diagnostic",
				fn: func(t *testing.T, mgr *Manager) {
					// Introduce an error by adding an unused variable.
					// Send a didChange to add "unused := 42" at line 35.
					err := mgr.DidChange(t.Context(),
						semanticapi.DidChangeTextDocumentParams{
							TextDocument: semanticapi.VersionedTextDocumentIdentifier{
								URI:     utilURI,
								Version: 2,
							},
							ContentChanges: []semanticapi.TextDocumentContentChangeEvent{
								{
									// Replace entire content with error code.
									Text: `package main

// Broken function with unused variable.
func Broken() {
	unused := 42
}
`,
								},
							},
						},
					)
					require.NoError(t, err)

					// Wait for gopls to process the change.
					time.Sleep(500 * time.Millisecond)

					// Pull diagnostics for the file with error.
					report, err := mgr.Diagnostic(t.Context(),
						semanticapi.DocumentDiagnosticParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: utilURI,
							},
						},
					)
					require.NoError(t, err)
					// Should have at least one diagnostic for unused variable.
					require.NotEmpty(t, report.Items)
					// Verify the diagnostic for unused variable.
					diag := report.Items[0]
					assert.Equal(t, semanticapi.DiagnosticSeverityError, diag.Severity)
					assert.Contains(t, diag.Message, "declared and not used")
					assert.Equal(t, uint32(4), diag.Range.Start.Line)
				},
			},
			{
				name: "ExecuteCommand",
				fn: func(t *testing.T, mgr *Manager) {
					arg := map[string]string{"URI": mainURI}
					argBytes, _ := json.Marshal(arg)
					result, err := mgr.ExecuteCommand(t.Context(),
						semanticapi.ExecuteCommandParams{
							Command:   "gopls.list_known_packages",
							Arguments: []json.RawMessage{argBytes},
						},
					)
					require.NoError(t, err)
					assert.Contains(t, result, "Packages")
					assert.Contains(t, result, "fmt")
				},
			},
			{
				name: "CompletionResolve",
				fn: func(t *testing.T, mgr *Manager) {
					// CompletionResolve returns item unchanged in this impl.
					item := semanticapi.CompletionItem{
						Label:  "testFunction",
						Kind:   semanticapi.CompletionItemKindFunction,
						Detail: "func testFunction()",
					}
					resolved, err := mgr.CompletionResolve(t.Context(), item)
					require.NoError(t, err)
					assert.Equal(t, item, resolved)
				},
			},
			{
				name: "CodeLensResolve",
				fn: func(t *testing.T, mgr *Manager) {
					// CodeLensResolve returns lens unchanged in this impl.
					lens := semanticapi.CodeLens{
						Range: semanticapi.Range{
							Start: semanticapi.Position{Line: 33, Character: 0},
							End:   semanticapi.Position{Line: 33, Character: 4},
						},
						Command: &semanticapi.Command{
							Title:   "run",
							Command: "gopls.run",
						},
					}
					resolved, err := mgr.CodeLensResolve(t.Context(), lens)
					require.NoError(t, err)
					assert.Equal(t, lens, resolved)
				},
			},
			{
				name: "DocumentColor",
				fn: func(t *testing.T, mgr *Manager) {
					colors, err := mgr.DocumentColor(t.Context(),
						semanticapi.DocumentColorParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: mainURI,
							},
						},
					)
					require.NoError(t, err)
					// Go files have no color literals; expect empty.
					assert.Empty(t, colors)
				},
			},
			{
				name: "ColorPresentation",
				fn: func(t *testing.T, mgr *Manager) {
					presentations, err := mgr.ColorPresentation(t.Context(),
						semanticapi.ColorPresentationParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: mainURI,
							},
							Color: semanticapi.Color{
								Red: 1.0, Green: 0.0, Blue: 0.0, Alpha: 1.0,
							},
							Range: semanticapi.Range{
								Start: semanticapi.Position{Line: 16, Character: 7},
								End:   semanticapi.Position{Line: 16, Character: 12},
							},
						},
					)
					require.NoError(t, err)
					// Go doesn't have color literals; expect empty.
					assert.Empty(t, presentations)
				},
			},
			{
				name: "DocumentLink",
				fn: func(t *testing.T, mgr *Manager) {
					links, err := mgr.DocumentLink(t.Context(),
						semanticapi.DocumentLinkParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: mainURI,
							},
						},
					)
					require.NoError(t, err)
					// gopls returns links for import paths and URLs in comments
					require.Len(t, links, 1)
					assert.Equal(t, semanticapi.DocumentLink{
						Range: semanticapi.Range{
							Start: semanticapi.Position{Line: 25, Character: 8},
							End:   semanticapi.Position{Line: 25, Character: 11},
						},
						Target: "https://pkg.go.dev/fmt",
					}, links[0])
				},
			},
			{
				name: "DocumentLinkResolve",
				fn: func(t *testing.T, mgr *Manager) {
					link := semanticapi.DocumentLink{
						Range: semanticapi.Range{
							Start: semanticapi.Position{Line: 25, Character: 7},
							End:   semanticapi.Position{Line: 25, Character: 12},
						},
						Target: "fmt",
					}
					resolved, err := mgr.DocumentLinkResolve(t.Context(), link)
					require.NoError(t, err)
					assert.Equal(t, link, resolved)
				},
			},
			{
				name: "OnTypeFormatting",
				fn: func(t *testing.T, mgr *Manager) {
					// gopls doesn't support onTypeFormatting.
					edits, err := mgr.OnTypeFormatting(t.Context(),
						semanticapi.DocumentOnTypeFormattingParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: mainURI,
							},
							Position: semanticapi.Position{
								Line: 30, Character: 0,
							},
							Character: "\n",
							Options: semanticapi.FormattingOptions{
								TabSize:      4,
								InsertSpaces: false,
							},
						},
					)
					require.NoError(t, err)
					// gopls returns nil for unsupported features.
					assert.Nil(t, edits)
				},
			},
			{
				name: "LinkedEditingRange",
				fn: func(t *testing.T, mgr *Manager) {
					// Go doesn't have linked editing ranges (like HTML tags).
					ranges, err := mgr.LinkedEditingRange(t.Context(),
						semanticapi.LinkedEditingRangeParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: mainURI,
							},
							Position: semanticapi.Position{
								Line: 29, Character: 5,
							},
						},
					)
					require.NoError(t, err)
					// Go files have no linked editing ranges.
					assert.Nil(t, ranges)
				},
			},
			{
				name: "Moniker",
				fn: func(t *testing.T, mgr *Manager) {
					// gopls doesn't fully support monikers.
					monikers, err := mgr.Moniker(t.Context(),
						semanticapi.MonikerParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: mainURI,
							},
							Position: semanticapi.Position{
								Line: 29, Character: 5,
							},
						},
					)
					require.NoError(t, err)
					// gopls returns nil for unsupported features.
					assert.Nil(t, monikers)
				},
			},
			{
				name: "WillSaveWaitUntil",
				fn: func(t *testing.T, mgr *Manager) {
					// gopls doesn't support willSaveWaitUntil.
					edits, err := mgr.WillSaveWaitUntil(t.Context(),
						semanticapi.WillSaveTextDocumentParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: mainURI,
							},
							Reason: semanticapi.TextDocumentSaveReasonManual,
						},
					)
					require.NoError(t, err)
					// gopls returns nil for unsupported features.
					assert.Nil(t, edits)
				},
			},
			{
				name: "SemanticTokensFullDelta",
				fn: func(t *testing.T, mgr *Manager) {
					// gopls doesn't support semanticTokens/full/delta.
					_, err := mgr.SemanticTokensFullDelta(t.Context(),
						semanticapi.SemanticTokensDeltaParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: mainURI,
							},
							PreviousResultID: "",
						},
					)
					require.Error(t, err)
					assert.Contains(t, err.Error(), "SemanticTokensFullDelta")
				},
			},
			{
				name: "PrepareTypeHierarchy",
				fn: func(t *testing.T, mgr *Manager) {
					// Prepare type hierarchy on Greeter struct (line 28, char 6).
					items, err := mgr.PrepareTypeHierarchy(t.Context(),
						semanticapi.TypeHierarchyPrepareParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: mainURI,
							},
							Position: semanticapi.Position{
								Line: 28, Character: 6,
							},
						},
					)
					require.NoError(t, err)
					require.Len(t, items, 1)
					assert.Equal(t, "Greeter", items[0].Name)
					// gopls returns SymbolKindClass for Go structs.
					assert.Equal(t, semanticapi.SymbolKindClass, items[0].Kind)
					assert.Equal(t, mainURI, items[0].URI)
					assert.Equal(t, semanticapi.Range{
						Start: semanticapi.Position{Line: 28, Character: 5},
						End:   semanticapi.Position{Line: 28, Character: 12},
					}, items[0].SelectionRange)
				},
			},
			{
				name: "TypeHierarchySupertypes",
				fn: func(t *testing.T, mgr *Manager) {
					// Create a synthetic item since PrepareTypeHierarchy returns nil.
					// This tests that the call works even with no real data.
					item := semanticapi.TypeHierarchyItem{
						Name: "Greeter",
						Kind: semanticapi.SymbolKindStruct,
						URI:  mainURI,
						Range: semanticapi.Range{
							Start: semanticapi.Position{Line: 28, Character: 0},
							End:   semanticapi.Position{Line: 31, Character: 1},
						},
						SelectionRange: semanticapi.Range{
							Start: semanticapi.Position{Line: 28, Character: 5},
							End:   semanticapi.Position{Line: 28, Character: 12},
						},
					}
					supertypes, err := mgr.TypeHierarchySupertypes(t.Context(),
						semanticapi.TypeHierarchySupertypesParams{
							Item: item,
						},
					)
					require.NoError(t, err)
					// Greeter has no supertypes (no embedded types).
					assert.Nil(t, supertypes)
				},
			},
			{
				name: "TypeHierarchySubtypes",
				fn: func(t *testing.T, mgr *Manager) {
					// Create a synthetic item since PrepareTypeHierarchy returns nil.
					item := semanticapi.TypeHierarchyItem{
						Name: "Greeter",
						Kind: semanticapi.SymbolKindStruct,
						URI:  mainURI,
						Range: semanticapi.Range{
							Start: semanticapi.Position{Line: 19, Character: 0},
							End:   semanticapi.Position{Line: 22, Character: 1},
						},
						SelectionRange: semanticapi.Range{
							Start: semanticapi.Position{Line: 19, Character: 5},
							End:   semanticapi.Position{Line: 19, Character: 12},
						},
					}
					subtypes, err := mgr.TypeHierarchySubtypes(t.Context(),
						semanticapi.TypeHierarchySubtypesParams{
							Item: item,
						},
					)
					require.NoError(t, err)
					// Greeter has no subtypes.
					assert.Nil(t, subtypes)
				},
			},
			{
				name: "InlayHint",
				fn: func(t *testing.T, mgr *Manager) {
					hints, err := mgr.InlayHint(t.Context(),
						semanticapi.InlayHintParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: mainURI,
							},
							Range: semanticapi.Range{
								Start: semanticapi.Position{Line: 0, Character: 0},
								End:   semanticapi.Position{Line: 60, Character: 0},
							},
						},
					)
					require.NoError(t, err)
					expected := []semanticapi.InlayHint{
						{
							Position:     semanticapi.Position{Line: 34, Character: 20},
							LabelParts:   []semanticapi.InlayHintLabelPart{{Value: "format:"}},
							Kind:         semanticapi.InlayHintKindParameter,
							PaddingRight: true,
						},
						{
							Position:     semanticapi.Position{Line: 34, Character: 34},
							LabelParts:   []semanticapi.InlayHintLabelPart{{Value: "a...:"}},
							Kind:         semanticapi.InlayHintKindParameter,
							PaddingRight: true,
						},
						{
							Position:     semanticapi.Position{Line: 44, Character: 13},
							LabelParts:   []semanticapi.InlayHintLabelPart{{Value: "a...:"}},
							Kind:         semanticapi.InlayHintKindParameter,
							PaddingRight: true,
						},
						{
							Position:     semanticapi.Position{Line: 45, Character: 13},
							LabelParts:   []semanticapi.InlayHintLabelPart{{Value: "a...:"}},
							Kind:         semanticapi.InlayHintKindParameter,
							PaddingRight: true,
						},
						{
							Position:     semanticapi.Position{Line: 45, Character: 17},
							LabelParts:   []semanticapi.InlayHintLabelPart{{Value: "a:"}},
							Kind:         semanticapi.InlayHintKindParameter,
							PaddingRight: true,
						},
						{
							Position:     semanticapi.Position{Line: 45, Character: 20},
							LabelParts:   []semanticapi.InlayHintLabelPart{{Value: "b:"}},
							Kind:         semanticapi.InlayHintKindParameter,
							PaddingRight: true,
						},
						{
							Position:     semanticapi.Position{Line: 61, Character: 20},
							LabelParts:   []semanticapi.InlayHintLabelPart{{Value: "format:"}},
							Kind:         semanticapi.InlayHintKindParameter,
							PaddingRight: true,
						},
						{
							Position:     semanticapi.Position{Line: 61, Character: 48},
							LabelParts:   []semanticapi.InlayHintLabelPart{{Value: "a...:"}},
							Kind:         semanticapi.InlayHintKindParameter,
							PaddingRight: true,
						},
						{
							Position:    semanticapi.Position{Line: 43, Character: 2},
							LabelParts:  []semanticapi.InlayHintLabelPart{{Value: "*Greeter"}},
							Kind:        semanticapi.InlayHintKindType,
							PaddingLeft: true,
						},
					}
					assert.ElementsMatch(t, expected, hints)
				},
			},
			{
				name: "InlayHintResolve",
				fn: func(t *testing.T, mgr *Manager) {
					hint := semanticapi.InlayHint{
						Position: semanticapi.Position{Line: 36, Character: 17},
						Label:    "a:",
						Kind:     semanticapi.InlayHintKindParameter,
					}
					resolved, err := mgr.InlayHintResolve(t.Context(), hint)
					require.NoError(t, err)
					assert.Equal(t, hint, resolved)
				},
			},
			{
				name: "InlineValue",
				fn: func(t *testing.T, mgr *Manager) {
					// gopls doesn't support inlineValue (debugger feature).
					values, err := mgr.InlineValue(t.Context(),
						semanticapi.InlineValueParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: mainURI,
							},
							Range: semanticapi.Range{
								Start: semanticapi.Position{Line: 29, Character: 0},
								End:   semanticapi.Position{Line: 32, Character: 0},
							},
						},
					)
					require.NoError(t, err)
					// gopls returns nil for unsupported features.
					assert.Nil(t, values)
				},
			},
			{
				name: "WillCreateFiles",
				fn: func(t *testing.T, mgr *Manager) {
					// gopls doesn't support willCreateFiles.
					_, err := mgr.WillCreateFiles(t.Context(),
						semanticapi.CreateFilesParams{
							Files: []semanticapi.FileCreate{
								{URI: "file:///tmp/newfile.go"},
							},
						},
					)
					require.Error(t, err)
					assert.Contains(t, err.Error(), "WillCreateFiles")
				},
			},
			{
				name: "WillRenameFiles",
				fn: func(t *testing.T, mgr *Manager) {
					// gopls doesn't support willRenameFiles.
					_, err := mgr.WillRenameFiles(t.Context(),
						semanticapi.RenameFilesParams{
							Files: []semanticapi.FileRename{
								{
									OldURI: utilURI,
									NewURI: "file:///tmp/renamed.go",
								},
							},
						},
					)
					require.Error(t, err)
					assert.Contains(t, err.Error(), "WillRenameFiles")
				},
			},
			{
				name: "WillDeleteFiles",
				fn: func(t *testing.T, mgr *Manager) {
					// gopls doesn't support willDeleteFiles.
					_, err := mgr.WillDeleteFiles(t.Context(),
						semanticapi.DeleteFilesParams{
							Files: []semanticapi.FileDelete{
								{URI: "file:///tmp/todelete.go"},
							},
						},
					)
					require.Error(t, err)
					assert.Contains(t, err.Error(), "WillDeleteFiles")
				},
			},
			{
				name: "SemanticTokensFull",
				fn: func(t *testing.T, mgr *Manager) {
					tokens, err := mgr.SemanticTokensFull(t.Context(),
						semanticapi.SemanticTokensParams{
							TextDocument: semanticapi.TextDocumentIdentifier{
								URI: mainURI,
							},
						},
					)
					require.NoError(t, err)
					require.NotNil(t, tokens)
					assert.Equal(t, 525, len(tokens.Data))
				},
			},
		}...)
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var wg sync.WaitGroup
				cfg := Config{
					MaxRetries: 1,
					Callback: &testCallback{
						onShowMessage: func(params semanticapi.ShowMessageParams) {
							if strings.Contains(params.Message, "Finished loading packages") ||
								strings.Contains(params.Message, "background refresh finished") {
								wg.Done()
							}
						},
					},
				}
				mgr := New(
					uri,
					scheme,
					scheme,
					textcomp,
					&stubPkgManager{bin: goplsBin},
					nil, // notifications
					textcomp,
					cfg,
				)

				ctx := context.Background()

				params := autoInitParams(uri.String())
				initializeOpts, err := json.Marshal(map[string]any{
					"semanticTokens": true,
					"langID":         "go",
					"command":        "gopls serve",
					"codelenses": map[string]bool{
						"gc_details":         true,
						"generate":           true,
						"regenerate_cgo":     true,
						"run_govulncheck":    true,
						"test":               true,
						"tidy":               true,
						"upgrade_dependency": true,
						"vendor":             true,
					},
					"hints": map[string]bool{
						"assignVariableTypes":    true,
						"compositeLiteralFields": true,
						"compositeLiteralTypes":  true,
						"constantValues":         true,
						"functionTypeParameters": true,
						"parameterNames":         true,
						"rangeVariableTypes":     true,
					},
				})
				require.NoError(t, err)
				params.InitializeOptions = initializeOpts

				wg.Add(1)
				_, err = mgr.Initialize(ctx, params)
				require.NoError(t, err)
				require.NoError(t, mgr.DidOpen(ctx, semanticapi.DidOpenTextDocumentParams{
					TextDocument: semanticapi.TextDocumentItem{
						URI:        mainURI,
						LanguageID: "go",
						Version:    0,
						Text:       string(mainContent),
					},
				}))
				require.NoError(t, mgr.DidOpen(ctx, semanticapi.DidOpenTextDocumentParams{
					TextDocument: semanticapi.TextDocumentItem{
						URI:        utilURI,
						LanguageID: "go",
						Version:    0,
						Text:       string(utilContent),
					},
				}))
				require.NoError(t, mgr.DidOpen(ctx, semanticapi.DidOpenTextDocumentParams{
					TextDocument: semanticapi.TextDocumentItem{
						URI:        testURI,
						LanguageID: "go",
						Version:    0,
						Text:       string(mainTestContent),
					},
				}))

				wg.Wait()
				tt.fn(t, mgr)
				require.NoError(t, mgr.Close())
			})
		}
	})
}

func TestE2ECallback(t *testing.T) {
	t.Parallel()
	goplsBin := findGopls(t)
	tmpDir := setupTestWorkspace(t, "")

	mainPath := filepath.Join(tmpDir, "main.go")
	testURI := "file://" + mainPath

	errorContent := `package main

func broken() {
	unused := 42
}
`

	ctx, cancel := context.WithTimeout(
		context.Background(), 10*time.Second,
	)
	defer cancel()

	var wg sync.WaitGroup
	var receivedDiagnostics []semanticapi.PublishDiagnosticsParams
	callback := &testCallback{
		onDiagnostics: func(p semanticapi.PublishDiagnosticsParams) {
			if len(p.Diagnostics) != 0 {
				defer wg.Done()
				receivedDiagnostics = append(receivedDiagnostics, p)
			}
		},
	}

	uri := makeURI(t, "file://"+tmpDir)
	scheme, err := workspace.NewFileScheme(context.Background(), config.NopConfig(), uri)
	require.NoError(t, err)
	w := workspace.NewSchemeWorkspace(uri, scheme)
	ed := vi.Editor()
	opener, err := text.NewComponent(ed, w, text.DefaultConfig())
	require.NoError(t, err)
	err = os.WriteFile(mainPath, []byte(errorContent), 0777)
	require.NoError(t, err)

	mgr := New(
		uri,
		scheme,
		scheme,
		ed,
		&stubPkgManager{bin: goplsBin},
		nil, // notifications
		opener,
		Config{Callback: callback, MaxRetries: 1},
	)

	params := autoInitParams(uri.String())
	initOptions := []byte(`{"langID": "go", "command": "gopls serve"}`)
	params.InitializeOptions = initOptions
	wg.Add(1)
	_, err = mgr.Initialize(ctx, params)
	require.NoError(t, err)
	require.NoError(t, mgr.DidOpen(ctx, semanticapi.DidOpenTextDocumentParams{
		TextDocument: semanticapi.TextDocumentItem{
			URI:        testURI,
			LanguageID: "go",
			Version:    0,
			Text:       errorContent,
		},
	}))
	wg.Wait()

	require.NotNil(t, receivedDiagnostics, "expected to receive diagnostics for main.go")
	require.Len(t, receivedDiagnostics, 1)
	expected := []semanticapi.PublishDiagnosticsParams{
		{
			URI:     testURI,
			Version: 0,
			Diagnostics: []semanticapi.Diagnostic{
				{
					Range: semanticapi.Range{
						Start: semanticapi.Position{
							Line:      3,
							Character: 1,
						},
						End: semanticapi.Position{
							Line:      3,
							Character: 7,
						},
					},
					Severity: 1,
					Code:     "UnusedVar",
					CodeDescription: &semanticapi.CodeDescription{
						Href: "https://pkg.go.dev/golang.org/x" +
							"/tools/internal/typesinternal#UnusedVar",
					},
					Source:  "compiler",
					Message: "declared and not used: unused",
					Tags: []semanticapi.DiagnosticTag{
						semanticapi.DiagnosticTagUnnecessary,
					},
					RelatedInformation: nil,
					Data:               nil,
				},
			},
		},
	}
	assert.ElementsMatch(t, expected, receivedDiagnostics)
}

func findGopls(t *testing.T) string {
	t.Helper()
	goplsBin, err := exec.LookPath("gopls")
	if err != nil {
		for _, p := range []string{
			filepath.Join(
				os.Getenv("HOME"),
				".rune", "bin", "gopls",
			),
			filepath.Join(
				os.Getenv("HOME"),
				"go", "bin", "gopls",
			),
		} {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		t.Skip("gopls not found, skipping e2e test")
	}
	return goplsBin
}

func setupTestWorkspace(t *testing.T, testdataDir string) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	if testdataDir == "" {
		return tmpDir
	}

	entries, err := os.ReadDir(testdataDir)
	require.NoError(t, err)
	for _, e := range entries {
		src := filepath.Join(testdataDir, e.Name())
		dst := filepath.Join(tmpDir, e.Name())
		data, err := os.ReadFile(src)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(
			dst, data, 0644,
		))
	}
	return tmpDir
}

type stubPkgManager struct {
	bin string
}

func (p *stubPkgManager) LibDir(
	_ context.Context, _ string,
) (iterator.Iterator[string], error) {
	return iterator.FromSlice(
		[]string{p.bin},
	), nil
}

type testCallback struct {
	mu            sync.Mutex
	onShowMessage func(params semanticapi.ShowMessageParams)
	diagnostics   []semanticapi.PublishDiagnosticsParams
	messages      []semanticapi.ShowMessageParams
	logMessages   []semanticapi.LogMessageParams
	progress      []semanticapi.ProgressParams
	onDiagnostics func(semanticapi.PublishDiagnosticsParams)
}

func (c *testCallback) ShowMessage(
	_ context.Context, params semanticapi.ShowMessageParams,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = append(c.messages, params)
	if c.onShowMessage != nil {
		c.onShowMessage(params)
	}
	return nil
}

func (c *testCallback) LogMessage(
	_ context.Context, params semanticapi.LogMessageParams,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logMessages = append(c.logMessages, params)
	return nil
}

func (c *testCallback) PublishDiagnostics(
	_ context.Context, params semanticapi.PublishDiagnosticsParams,
) error {
	c.mu.Lock()
	c.diagnostics = append(c.diagnostics, params)
	cb := c.onDiagnostics
	c.mu.Unlock()
	if cb != nil {
		cb(params)
	}
	return nil
}

func (c *testCallback) Progress(
	_ context.Context, params semanticapi.ProgressParams,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.progress = append(c.progress, params)
	return nil
}

func (c *testCallback) LogTrace(
	_ context.Context, _ semanticapi.LogTraceParams,
) error {
	return nil
}

func (c *testCallback) ShowDocument(
	_ context.Context, _ semanticapi.ShowDocumentParams,
) (semanticapi.ShowDocumentResult, error) {
	return semanticapi.ShowDocumentResult{Success: true}, nil
}

func (c *testCallback) ShowMessageRequest(
	_ context.Context, _ semanticapi.ShowMessageRequestParams,
) (*semanticapi.MessageActionItem, error) {
	return nil, nil
}

func (c *testCallback) WorkDoneProgressCreate(
	_ context.Context, _ semanticapi.WorkDoneProgressCreateParams,
) error {
	return nil
}

func (c *testCallback) ApplyEdit(
	_ context.Context, _ semanticapi.ApplyWorkspaceEditParams,
) (semanticapi.ApplyWorkspaceEditResult, error) {
	return semanticapi.ApplyWorkspaceEditResult{Applied: true}, nil
}

func (c *testCallback) WorkspaceFolders(
	_ context.Context,
) ([]semanticapi.WorkspaceFolder, error) {
	return nil, nil
}

func (c *testCallback) Configuration(
	_ context.Context, _ semanticapi.ConfigurationParams,
) ([]json.RawMessage, error) {
	return nil, nil
}

func (c *testCallback) RegisterCapability(
	_ context.Context, _ semanticapi.RegistrationParams,
) error {
	return nil
}

func (c *testCallback) UnregisterCapability(
	_ context.Context, _ semanticapi.UnregistrationParams,
) error {
	return nil
}

func (c *testCallback) CodeLensRefresh(_ context.Context) error {
	return nil
}

func (c *testCallback) SemanticTokensRefresh(_ context.Context) error {
	return nil
}

func (c *testCallback) InlayHintRefresh(_ context.Context) error {
	return nil
}

func (c *testCallback) DiagnosticRefresh(_ context.Context) error {
	return nil
}

func makeURI(t *testing.T, uri string) workspaceapi.URI {
	ret, err := workspaceapi.ParseURI(uri)
	require.NoError(t, err)
	return ret
}
