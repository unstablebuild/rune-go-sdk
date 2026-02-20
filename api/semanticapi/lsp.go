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

import "context"

// LSP defines the interface for a Language Server Protocol server.
type LSP interface {
	Initialize(ctx context.Context, params InitializeParams) (InitializeResult, error)
	Initialized(ctx context.Context) error
	Shutdown(ctx context.Context) error
	Exit(ctx context.Context) error

	DidOpen(ctx context.Context, params DidOpenTextDocumentParams) error
	DidChange(ctx context.Context, params DidChangeTextDocumentParams) error
	DidClose(ctx context.Context, params DidCloseTextDocumentParams) error
	DidSave(ctx context.Context, params DidSaveTextDocumentParams) error

	Completion(ctx context.Context, params CompletionParams) (CompletionResult, error)
	Hover(ctx context.Context, params HoverParams) (*Hover, error)
	SignatureHelp(ctx context.Context, params SignatureHelpParams) (*SignatureHelp, error)
	Definition(ctx context.Context, params DefinitionParams) (LocationResult, error)
	Declaration(ctx context.Context, params DeclarationParams) (LocationResult, error)
	TypeDefinition(ctx context.Context, params TypeDefinitionParams) (LocationResult, error)
	Implementation(ctx context.Context, params ImplementationParams) (LocationResult, error)
	References(ctx context.Context, params ReferenceParams) ([]Location, error)
	DocumentHighlight(ctx context.Context, params DocumentHighlightParams) ([]DocumentHighlight, error)
	DocumentSymbol(ctx context.Context, params DocumentSymbolParams) (DocumentSymbolResult, error)
	CodeAction(ctx context.Context, params CodeActionParams) ([]CodeActionResult, error)
	CodeLens(ctx context.Context, params CodeLensParams) ([]CodeLens, error)
	Formatting(ctx context.Context, params DocumentFormattingParams) ([]TextEdit, error)
	RangeFormatting(ctx context.Context, params DocumentRangeFormattingParams) ([]TextEdit, error)
	Rename(ctx context.Context, params RenameParams) (*WorkspaceEdit, error)
	PrepareRename(ctx context.Context, params PrepareRenameParams) (*PrepareRenameResult, error)
	FoldingRange(ctx context.Context, params FoldingRangeParams) ([]FoldingRange, error)
	SelectionRange(ctx context.Context, params SelectionRangeParams) ([]SelectionRange, error)

	SemanticTokensFull(ctx context.Context, params SemanticTokensParams) (*SemanticTokens, error)
	SemanticTokensRange(ctx context.Context, params SemanticTokensRangeParams) (*SemanticTokens, error)

	Diagnostic(ctx context.Context, params DocumentDiagnosticParams) (DocumentDiagnosticReport, error)
	WorkspaceDiagnostic(ctx context.Context, params WorkspaceDiagnosticParams) (WorkspaceDiagnosticReport, error)

	WorkspaceSymbol(ctx context.Context, params WorkspaceSymbolParams) ([]SymbolInformation, error)
	ExecuteCommand(ctx context.Context, params ExecuteCommandParams) (string, error)

	PrepareCallHierarchy(ctx context.Context, params CallHierarchyPrepareParams) ([]CallHierarchyItem, error)
	CallHierarchyIncomingCalls(ctx context.Context, params CallHierarchyIncomingCallsParams) ([]CallHierarchyIncomingCall, error)
	CallHierarchyOutgoingCalls(ctx context.Context, params CallHierarchyOutgoingCallsParams) ([]CallHierarchyOutgoingCall, error)

	CompletionResolve(ctx context.Context, item CompletionItem) (CompletionItem, error)
	CodeLensResolve(ctx context.Context, lens CodeLens) (CodeLens, error)
	DocumentColor(ctx context.Context, params DocumentColorParams) ([]ColorInformation, error)
	ColorPresentation(ctx context.Context, params ColorPresentationParams) ([]ColorPresentation, error)
	DocumentLink(ctx context.Context, params DocumentLinkParams) ([]DocumentLink, error)
	DocumentLinkResolve(ctx context.Context, link DocumentLink) (DocumentLink, error)
	OnTypeFormatting(ctx context.Context, params DocumentOnTypeFormattingParams) ([]TextEdit, error)
	LinkedEditingRange(ctx context.Context, params LinkedEditingRangeParams) (*LinkedEditingRanges, error)
	Moniker(ctx context.Context, params MonikerParams) ([]Moniker, error)
	WillSaveWaitUntil(ctx context.Context, params WillSaveTextDocumentParams) ([]TextEdit, error)
	SemanticTokensFullDelta(ctx context.Context, params SemanticTokensDeltaParams) (*SemanticTokensDelta, error)

	PrepareTypeHierarchy(ctx context.Context, params TypeHierarchyPrepareParams) ([]TypeHierarchyItem, error)
	TypeHierarchySupertypes(ctx context.Context, params TypeHierarchySupertypesParams) ([]TypeHierarchyItem, error)
	TypeHierarchySubtypes(ctx context.Context, params TypeHierarchySubtypesParams) ([]TypeHierarchyItem, error)

	InlayHint(ctx context.Context, params InlayHintParams) ([]InlayHint, error)
	InlayHintResolve(ctx context.Context, hint InlayHint) (InlayHint, error)
	InlineValue(ctx context.Context, params InlineValueParams) ([]InlineValue, error)

	WillCreateFiles(ctx context.Context, params CreateFilesParams) (*WorkspaceEdit, error)
	WillRenameFiles(ctx context.Context, params RenameFilesParams) (*WorkspaceEdit, error)
	WillDeleteFiles(ctx context.Context, params DeleteFilesParams) (*WorkspaceEdit, error)

	WillSave(ctx context.Context, params WillSaveTextDocumentParams) error
	DidChangeConfiguration(ctx context.Context, params DidChangeConfigurationParams) error
	DidChangeWatchedFiles(ctx context.Context, params DidChangeWatchedFilesParams) error
	DidChangeWorkspaceFolders(ctx context.Context, params DidChangeWorkspaceFoldersParams) error
	WorkDoneProgressCancel(ctx context.Context, params WorkDoneProgressCancelParams) error
	SetTrace(ctx context.Context, params SetTraceParams) error
	DidCreateFiles(ctx context.Context, params CreateFilesParams) error
	DidRenameFiles(ctx context.Context, params RenameFilesParams) error
	DidDeleteFiles(ctx context.Context, params DeleteFilesParams) error
}
