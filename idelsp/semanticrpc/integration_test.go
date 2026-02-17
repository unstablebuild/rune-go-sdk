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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
	"github.com/unstablebuild/rune-go-sdk/api/semanticapi/semanticrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var sockCounter atomic.Uint64

// stubLSP is a configurable test double for semanticapi.LSP.
// Each field is a function that, when set, overrides the default behaviour.
// Default behaviour returns zero values (happy path) so tests only need to
// configure the methods they care about.
type stubLSP struct {
	onInitialize                 func(context.Context, semanticapi.InitializeParams) (semanticapi.InitializeResult, error)
	onInitialized                func(context.Context) error
	onShutdown                   func(context.Context) error
	onExit                       func(context.Context) error
	onDidOpen                    func(context.Context, semanticapi.DidOpenTextDocumentParams) error
	onDidChange                  func(context.Context, semanticapi.DidChangeTextDocumentParams) error
	onDidClose                   func(context.Context, semanticapi.DidCloseTextDocumentParams) error
	onDidSave                    func(context.Context, semanticapi.DidSaveTextDocumentParams) error
	onCompletion                 func(context.Context, semanticapi.CompletionParams) (semanticapi.CompletionResult, error)
	onHover                      func(context.Context, semanticapi.HoverParams) (*semanticapi.Hover, error)
	onSignatureHelp              func(context.Context, semanticapi.SignatureHelpParams) (*semanticapi.SignatureHelp, error)
	onDefinition                 func(context.Context, semanticapi.DefinitionParams) (semanticapi.LocationResult, error)
	onDeclaration                func(context.Context, semanticapi.DeclarationParams) (semanticapi.LocationResult, error)
	onTypeDefinition             func(context.Context, semanticapi.TypeDefinitionParams) (semanticapi.LocationResult, error)
	onImplementation             func(context.Context, semanticapi.ImplementationParams) (semanticapi.LocationResult, error)
	onReferences                 func(context.Context, semanticapi.ReferenceParams) ([]semanticapi.Location, error)
	onDocumentHighlight          func(context.Context, semanticapi.DocumentHighlightParams) ([]semanticapi.DocumentHighlight, error)
	onDocumentSymbol             func(context.Context, semanticapi.DocumentSymbolParams) (semanticapi.DocumentSymbolResult, error)
	onCodeAction                 func(context.Context, semanticapi.CodeActionParams) ([]semanticapi.CodeActionResult, error)
	onCodeLens                   func(context.Context, semanticapi.CodeLensParams) ([]semanticapi.CodeLens, error)
	onFormatting                 func(context.Context, semanticapi.DocumentFormattingParams) ([]semanticapi.TextEdit, error)
	onRangeFormatting            func(context.Context, semanticapi.DocumentRangeFormattingParams) ([]semanticapi.TextEdit, error)
	onRename                     func(context.Context, semanticapi.RenameParams) (*semanticapi.WorkspaceEdit, error)
	onPrepareRename              func(context.Context, semanticapi.PrepareRenameParams) (*semanticapi.PrepareRenameResult, error)
	onFoldingRange               func(context.Context, semanticapi.FoldingRangeParams) ([]semanticapi.FoldingRange, error)
	onSelectionRange             func(context.Context, semanticapi.SelectionRangeParams) ([]semanticapi.SelectionRange, error)
	onSemanticTokensFull         func(context.Context, semanticapi.SemanticTokensParams) (*semanticapi.SemanticTokens, error)
	onSemanticTokensRange        func(context.Context, semanticapi.SemanticTokensRangeParams) (*semanticapi.SemanticTokens, error)
	onDiagnostic                 func(context.Context, semanticapi.DocumentDiagnosticParams) (semanticapi.DocumentDiagnosticReport, error)
	onWorkspaceSymbol            func(context.Context, semanticapi.WorkspaceSymbolParams) ([]semanticapi.SymbolInformation, error)
	onExecuteCommand             func(context.Context, semanticapi.ExecuteCommandParams) (string, error)
	onPrepareCallHierarchy       func(context.Context, semanticapi.CallHierarchyPrepareParams) ([]semanticapi.CallHierarchyItem, error)
	onCallHierarchyIncomingCalls func(context.Context, semanticapi.CallHierarchyIncomingCallsParams) ([]semanticapi.CallHierarchyIncomingCall, error)
	onCallHierarchyOutgoingCalls func(context.Context, semanticapi.CallHierarchyOutgoingCallsParams) ([]semanticapi.CallHierarchyOutgoingCall, error)
	onCompletionResolve          func(context.Context, semanticapi.CompletionItem) (semanticapi.CompletionItem, error)
	onCodeLensResolve            func(context.Context, semanticapi.CodeLens) (semanticapi.CodeLens, error)
	onDocumentColor              func(context.Context, semanticapi.DocumentColorParams) ([]semanticapi.ColorInformation, error)
	onColorPresentation          func(context.Context, semanticapi.ColorPresentationParams) ([]semanticapi.ColorPresentation, error)
	onDocumentLink               func(context.Context, semanticapi.DocumentLinkParams) ([]semanticapi.DocumentLink, error)
	onDocumentLinkResolve        func(context.Context, semanticapi.DocumentLink) (semanticapi.DocumentLink, error)
	onOnTypeFormatting           func(context.Context, semanticapi.DocumentOnTypeFormattingParams) ([]semanticapi.TextEdit, error)
	onLinkedEditingRange         func(context.Context, semanticapi.LinkedEditingRangeParams) (*semanticapi.LinkedEditingRanges, error)
	onMoniker                    func(context.Context, semanticapi.MonikerParams) ([]semanticapi.Moniker, error)
	onWillSaveWaitUntil          func(context.Context, semanticapi.WillSaveTextDocumentParams) ([]semanticapi.TextEdit, error)
	onSemanticTokensFullDelta    func(context.Context, semanticapi.SemanticTokensDeltaParams) (*semanticapi.SemanticTokensDelta, error)
	onPrepareTypeHierarchy       func(context.Context, semanticapi.TypeHierarchyPrepareParams) ([]semanticapi.TypeHierarchyItem, error)
	onTypeHierarchySupertypes    func(context.Context, semanticapi.TypeHierarchySupertypesParams) ([]semanticapi.TypeHierarchyItem, error)
	onTypeHierarchySubtypes      func(context.Context, semanticapi.TypeHierarchySubtypesParams) ([]semanticapi.TypeHierarchyItem, error)
	onInlayHint                  func(context.Context, semanticapi.InlayHintParams) ([]semanticapi.InlayHint, error)
	onInlayHintResolve           func(context.Context, semanticapi.InlayHint) (semanticapi.InlayHint, error)
	onInlineValue                func(context.Context, semanticapi.InlineValueParams) ([]semanticapi.InlineValue, error)
	onWillCreateFiles            func(context.Context, semanticapi.CreateFilesParams) (*semanticapi.WorkspaceEdit, error)
	onWillRenameFiles            func(context.Context, semanticapi.RenameFilesParams) (*semanticapi.WorkspaceEdit, error)
	onWillDeleteFiles            func(context.Context, semanticapi.DeleteFilesParams) (*semanticapi.WorkspaceEdit, error)
	onWillSave                   func(context.Context, semanticapi.WillSaveTextDocumentParams) error
	onDidChangeConfiguration     func(context.Context, semanticapi.DidChangeConfigurationParams) error
	onDidChangeWatchedFiles      func(context.Context, semanticapi.DidChangeWatchedFilesParams) error
	onDidChangeWorkspaceFolders  func(context.Context, semanticapi.DidChangeWorkspaceFoldersParams) error
	onWorkDoneProgressCancel     func(context.Context, semanticapi.WorkDoneProgressCancelParams) error
	onSetTrace                   func(context.Context, semanticapi.SetTraceParams) error
	onDidCreateFiles             func(context.Context, semanticapi.CreateFilesParams) error
	onDidRenameFiles             func(context.Context, semanticapi.RenameFilesParams) error
	onDidDeleteFiles             func(context.Context, semanticapi.DeleteFilesParams) error
}

func (s *stubLSP) Initialize(ctx context.Context, p semanticapi.InitializeParams) (semanticapi.InitializeResult, error) {
	if s.onInitialize != nil {
		return s.onInitialize(ctx, p)
	}
	return semanticapi.InitializeResult{}, nil
}
func (s *stubLSP) Initialized(ctx context.Context) error {
	if s.onInitialized != nil {
		return s.onInitialized(ctx)
	}
	return nil
}
func (s *stubLSP) Shutdown(ctx context.Context) error {
	if s.onShutdown != nil {
		return s.onShutdown(ctx)
	}
	return nil
}
func (s *stubLSP) Exit(ctx context.Context) error {
	if s.onExit != nil {
		return s.onExit(ctx)
	}
	return nil
}
func (s *stubLSP) DidOpen(ctx context.Context, p semanticapi.DidOpenTextDocumentParams) error {
	if s.onDidOpen != nil {
		return s.onDidOpen(ctx, p)
	}
	return nil
}
func (s *stubLSP) DidChange(ctx context.Context, p semanticapi.DidChangeTextDocumentParams) error {
	if s.onDidChange != nil {
		return s.onDidChange(ctx, p)
	}
	return nil
}
func (s *stubLSP) DidClose(ctx context.Context, p semanticapi.DidCloseTextDocumentParams) error {
	if s.onDidClose != nil {
		return s.onDidClose(ctx, p)
	}
	return nil
}
func (s *stubLSP) DidSave(ctx context.Context, p semanticapi.DidSaveTextDocumentParams) error {
	if s.onDidSave != nil {
		return s.onDidSave(ctx, p)
	}
	return nil
}
func (s *stubLSP) Completion(ctx context.Context, p semanticapi.CompletionParams) (semanticapi.CompletionResult, error) {
	if s.onCompletion != nil {
		return s.onCompletion(ctx, p)
	}
	return semanticapi.CompletionResult{}, nil
}
func (s *stubLSP) Hover(ctx context.Context, p semanticapi.HoverParams) (*semanticapi.Hover, error) {
	if s.onHover != nil {
		return s.onHover(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) SignatureHelp(ctx context.Context, p semanticapi.SignatureHelpParams) (*semanticapi.SignatureHelp, error) {
	if s.onSignatureHelp != nil {
		return s.onSignatureHelp(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) Definition(ctx context.Context, p semanticapi.DefinitionParams) (semanticapi.LocationResult, error) {
	if s.onDefinition != nil {
		return s.onDefinition(ctx, p)
	}
	return semanticapi.LocationResult{}, nil
}
func (s *stubLSP) Declaration(ctx context.Context, p semanticapi.DeclarationParams) (semanticapi.LocationResult, error) {
	if s.onDeclaration != nil {
		return s.onDeclaration(ctx, p)
	}
	return semanticapi.LocationResult{}, nil
}
func (s *stubLSP) TypeDefinition(ctx context.Context, p semanticapi.TypeDefinitionParams) (semanticapi.LocationResult, error) {
	if s.onTypeDefinition != nil {
		return s.onTypeDefinition(ctx, p)
	}
	return semanticapi.LocationResult{}, nil
}
func (s *stubLSP) Implementation(ctx context.Context, p semanticapi.ImplementationParams) (semanticapi.LocationResult, error) {
	if s.onImplementation != nil {
		return s.onImplementation(ctx, p)
	}
	return semanticapi.LocationResult{}, nil
}
func (s *stubLSP) References(ctx context.Context, p semanticapi.ReferenceParams) ([]semanticapi.Location, error) {
	if s.onReferences != nil {
		return s.onReferences(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) DocumentHighlight(ctx context.Context, p semanticapi.DocumentHighlightParams) ([]semanticapi.DocumentHighlight, error) {
	if s.onDocumentHighlight != nil {
		return s.onDocumentHighlight(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) DocumentSymbol(ctx context.Context, p semanticapi.DocumentSymbolParams) (semanticapi.DocumentSymbolResult, error) {
	if s.onDocumentSymbol != nil {
		return s.onDocumentSymbol(ctx, p)
	}
	return semanticapi.DocumentSymbolResult{}, nil
}
func (s *stubLSP) CodeAction(ctx context.Context, p semanticapi.CodeActionParams) ([]semanticapi.CodeActionResult, error) {
	if s.onCodeAction != nil {
		return s.onCodeAction(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) CodeLens(ctx context.Context, p semanticapi.CodeLensParams) ([]semanticapi.CodeLens, error) {
	if s.onCodeLens != nil {
		return s.onCodeLens(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) Formatting(ctx context.Context, p semanticapi.DocumentFormattingParams) ([]semanticapi.TextEdit, error) {
	if s.onFormatting != nil {
		return s.onFormatting(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) RangeFormatting(ctx context.Context, p semanticapi.DocumentRangeFormattingParams) ([]semanticapi.TextEdit, error) {
	if s.onRangeFormatting != nil {
		return s.onRangeFormatting(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) Rename(ctx context.Context, p semanticapi.RenameParams) (*semanticapi.WorkspaceEdit, error) {
	if s.onRename != nil {
		return s.onRename(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) PrepareRename(ctx context.Context, p semanticapi.PrepareRenameParams) (*semanticapi.PrepareRenameResult, error) {
	if s.onPrepareRename != nil {
		return s.onPrepareRename(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) FoldingRange(ctx context.Context, p semanticapi.FoldingRangeParams) ([]semanticapi.FoldingRange, error) {
	if s.onFoldingRange != nil {
		return s.onFoldingRange(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) SelectionRange(ctx context.Context, p semanticapi.SelectionRangeParams) ([]semanticapi.SelectionRange, error) {
	if s.onSelectionRange != nil {
		return s.onSelectionRange(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) SemanticTokensFull(ctx context.Context, p semanticapi.SemanticTokensParams) (*semanticapi.SemanticTokens, error) {
	if s.onSemanticTokensFull != nil {
		return s.onSemanticTokensFull(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) SemanticTokensRange(ctx context.Context, p semanticapi.SemanticTokensRangeParams) (*semanticapi.SemanticTokens, error) {
	if s.onSemanticTokensRange != nil {
		return s.onSemanticTokensRange(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) Diagnostic(ctx context.Context, p semanticapi.DocumentDiagnosticParams) (semanticapi.DocumentDiagnosticReport, error) {
	if s.onDiagnostic != nil {
		return s.onDiagnostic(ctx, p)
	}
	return semanticapi.DocumentDiagnosticReport{}, nil
}
func (s *stubLSP) WorkspaceSymbol(ctx context.Context, p semanticapi.WorkspaceSymbolParams) ([]semanticapi.SymbolInformation, error) {
	if s.onWorkspaceSymbol != nil {
		return s.onWorkspaceSymbol(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) ExecuteCommand(ctx context.Context, p semanticapi.ExecuteCommandParams) (string, error) {
	if s.onExecuteCommand != nil {
		return s.onExecuteCommand(ctx, p)
	}
	return "", nil
}
func (s *stubLSP) PrepareCallHierarchy(ctx context.Context, p semanticapi.CallHierarchyPrepareParams) ([]semanticapi.CallHierarchyItem, error) {
	if s.onPrepareCallHierarchy != nil {
		return s.onPrepareCallHierarchy(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) CallHierarchyIncomingCalls(ctx context.Context, p semanticapi.CallHierarchyIncomingCallsParams) ([]semanticapi.CallHierarchyIncomingCall, error) {
	if s.onCallHierarchyIncomingCalls != nil {
		return s.onCallHierarchyIncomingCalls(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) CallHierarchyOutgoingCalls(ctx context.Context, p semanticapi.CallHierarchyOutgoingCallsParams) ([]semanticapi.CallHierarchyOutgoingCall, error) {
	if s.onCallHierarchyOutgoingCalls != nil {
		return s.onCallHierarchyOutgoingCalls(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) CompletionResolve(ctx context.Context, item semanticapi.CompletionItem) (semanticapi.CompletionItem, error) {
	if s.onCompletionResolve != nil {
		return s.onCompletionResolve(ctx, item)
	}
	return semanticapi.CompletionItem{}, nil
}
func (s *stubLSP) CodeLensResolve(ctx context.Context, lens semanticapi.CodeLens) (semanticapi.CodeLens, error) {
	if s.onCodeLensResolve != nil {
		return s.onCodeLensResolve(ctx, lens)
	}
	return semanticapi.CodeLens{}, nil
}
func (s *stubLSP) DocumentColor(ctx context.Context, p semanticapi.DocumentColorParams) ([]semanticapi.ColorInformation, error) {
	if s.onDocumentColor != nil {
		return s.onDocumentColor(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) ColorPresentation(ctx context.Context, p semanticapi.ColorPresentationParams) ([]semanticapi.ColorPresentation, error) {
	if s.onColorPresentation != nil {
		return s.onColorPresentation(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) DocumentLink(ctx context.Context, p semanticapi.DocumentLinkParams) ([]semanticapi.DocumentLink, error) {
	if s.onDocumentLink != nil {
		return s.onDocumentLink(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) DocumentLinkResolve(ctx context.Context, link semanticapi.DocumentLink) (semanticapi.DocumentLink, error) {
	if s.onDocumentLinkResolve != nil {
		return s.onDocumentLinkResolve(ctx, link)
	}
	return semanticapi.DocumentLink{}, nil
}
func (s *stubLSP) OnTypeFormatting(ctx context.Context, p semanticapi.DocumentOnTypeFormattingParams) ([]semanticapi.TextEdit, error) {
	if s.onOnTypeFormatting != nil {
		return s.onOnTypeFormatting(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) LinkedEditingRange(ctx context.Context, p semanticapi.LinkedEditingRangeParams) (*semanticapi.LinkedEditingRanges, error) {
	if s.onLinkedEditingRange != nil {
		return s.onLinkedEditingRange(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) Moniker(ctx context.Context, p semanticapi.MonikerParams) ([]semanticapi.Moniker, error) {
	if s.onMoniker != nil {
		return s.onMoniker(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) WillSaveWaitUntil(ctx context.Context, p semanticapi.WillSaveTextDocumentParams) ([]semanticapi.TextEdit, error) {
	if s.onWillSaveWaitUntil != nil {
		return s.onWillSaveWaitUntil(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) SemanticTokensFullDelta(ctx context.Context, p semanticapi.SemanticTokensDeltaParams) (*semanticapi.SemanticTokensDelta, error) {
	if s.onSemanticTokensFullDelta != nil {
		return s.onSemanticTokensFullDelta(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) PrepareTypeHierarchy(ctx context.Context, p semanticapi.TypeHierarchyPrepareParams) ([]semanticapi.TypeHierarchyItem, error) {
	if s.onPrepareTypeHierarchy != nil {
		return s.onPrepareTypeHierarchy(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) TypeHierarchySupertypes(ctx context.Context, p semanticapi.TypeHierarchySupertypesParams) ([]semanticapi.TypeHierarchyItem, error) {
	if s.onTypeHierarchySupertypes != nil {
		return s.onTypeHierarchySupertypes(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) TypeHierarchySubtypes(ctx context.Context, p semanticapi.TypeHierarchySubtypesParams) ([]semanticapi.TypeHierarchyItem, error) {
	if s.onTypeHierarchySubtypes != nil {
		return s.onTypeHierarchySubtypes(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) InlayHint(ctx context.Context, p semanticapi.InlayHintParams) ([]semanticapi.InlayHint, error) {
	if s.onInlayHint != nil {
		return s.onInlayHint(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) InlayHintResolve(ctx context.Context, hint semanticapi.InlayHint) (semanticapi.InlayHint, error) {
	if s.onInlayHintResolve != nil {
		return s.onInlayHintResolve(ctx, hint)
	}
	return semanticapi.InlayHint{}, nil
}
func (s *stubLSP) InlineValue(ctx context.Context, p semanticapi.InlineValueParams) ([]semanticapi.InlineValue, error) {
	if s.onInlineValue != nil {
		return s.onInlineValue(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) WillCreateFiles(ctx context.Context, p semanticapi.CreateFilesParams) (*semanticapi.WorkspaceEdit, error) {
	if s.onWillCreateFiles != nil {
		return s.onWillCreateFiles(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) WillRenameFiles(ctx context.Context, p semanticapi.RenameFilesParams) (*semanticapi.WorkspaceEdit, error) {
	if s.onWillRenameFiles != nil {
		return s.onWillRenameFiles(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) WillDeleteFiles(ctx context.Context, p semanticapi.DeleteFilesParams) (*semanticapi.WorkspaceEdit, error) {
	if s.onWillDeleteFiles != nil {
		return s.onWillDeleteFiles(ctx, p)
	}
	return nil, nil
}
func (s *stubLSP) WillSave(ctx context.Context, p semanticapi.WillSaveTextDocumentParams) error {
	if s.onWillSave != nil {
		return s.onWillSave(ctx, p)
	}
	return nil
}
func (s *stubLSP) DidChangeConfiguration(ctx context.Context, p semanticapi.DidChangeConfigurationParams) error {
	if s.onDidChangeConfiguration != nil {
		return s.onDidChangeConfiguration(ctx, p)
	}
	return nil
}
func (s *stubLSP) DidChangeWatchedFiles(ctx context.Context, p semanticapi.DidChangeWatchedFilesParams) error {
	if s.onDidChangeWatchedFiles != nil {
		return s.onDidChangeWatchedFiles(ctx, p)
	}
	return nil
}
func (s *stubLSP) DidChangeWorkspaceFolders(ctx context.Context, p semanticapi.DidChangeWorkspaceFoldersParams) error {
	if s.onDidChangeWorkspaceFolders != nil {
		return s.onDidChangeWorkspaceFolders(ctx, p)
	}
	return nil
}
func (s *stubLSP) WorkDoneProgressCancel(ctx context.Context, p semanticapi.WorkDoneProgressCancelParams) error {
	if s.onWorkDoneProgressCancel != nil {
		return s.onWorkDoneProgressCancel(ctx, p)
	}
	return nil
}
func (s *stubLSP) SetTrace(ctx context.Context, p semanticapi.SetTraceParams) error {
	if s.onSetTrace != nil {
		return s.onSetTrace(ctx, p)
	}
	return nil
}
func (s *stubLSP) DidCreateFiles(ctx context.Context, p semanticapi.CreateFilesParams) error {
	if s.onDidCreateFiles != nil {
		return s.onDidCreateFiles(ctx, p)
	}
	return nil
}
func (s *stubLSP) DidRenameFiles(ctx context.Context, p semanticapi.RenameFilesParams) error {
	if s.onDidRenameFiles != nil {
		return s.onDidRenameFiles(ctx, p)
	}
	return nil
}
func (s *stubLSP) DidDeleteFiles(ctx context.Context, p semanticapi.DeleteFilesParams) error {
	if s.onDidDeleteFiles != nil {
		return s.onDidDeleteFiles(ctx, p)
	}
	return nil
}

// testEnv holds a running gRPC server and a connected client for integration tests.
type testEnv struct {
	stub   *stubLSP
	client semanticapi.LSP
	stop   func()
}

// newTestEnv starts a gRPC server on a unix socket with the given stub and
// returns a connected client. Cleanup is automatic via t.Cleanup.
func newTestEnv(t *testing.T, stub *stubLSP) testEnv {
	t.Helper()

	// Use a short path to avoid exceeding macOS's 104-char unix socket limit.
	// t.TempDir() embeds the full test/subtest name which can be very long.
	n := sockCounter.Add(1)
	sock := filepath.Join(os.TempDir(), fmt.Sprintf("lsp_test_%d_%d.sock", os.Getpid(), n))
	t.Cleanup(func() { _ = os.Remove(sock) })
	lis, err := net.Listen("unix", sock)
	require.NoError(t, err)

	srv := grpc.NewServer()
	semanticrpc.RegisterLSPServer(srv, NewServer(stub))
	go func() { _ = srv.Serve(lis) }()

	cc, err := grpc.NewClient(
		"passthrough:///unix://"+sock,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", sock)
		}),
	)
	require.NoError(t, err)

	client := semanticrpc.NewClient(context.Background(), cc)

	t.Cleanup(func() {
		_ = cc.Close()
		srv.GracefulStop()
	})

	return testEnv{stub: stub, client: client, stop: srv.GracefulStop}
}

func TestInitialize(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		stub := &stubLSP{
			onInitialize: func(_ context.Context, p semanticapi.InitializeParams) (semanticapi.InitializeResult, error) {
				assert.Equal(t, "file:///workspace", p.RootURI)
				return semanticapi.InitializeResult{
					Capabilities: semanticapi.ServerCapabilities{
						CompletionProvider:    &semanticapi.CompletionOptions{},
						HoverProvider:         true,
						DefinitionProvider:    true,
						ReferencesProvider:    true,
						CallHierarchyProvider: true,
					},
				}, nil
			},
		}
		env := newTestEnv(t, stub)
		result, err := env.client.Initialize(context.Background(), semanticapi.InitializeParams{
			RootURI: "file:///workspace",
		})
		require.NoError(t, err)
		assert.NotNil(t, result.Capabilities.CompletionProvider)
		assert.True(t, result.Capabilities.HoverProvider)
		assert.True(t, result.Capabilities.DefinitionProvider)
		assert.True(t, result.Capabilities.ReferencesProvider)
		assert.True(t, result.Capabilities.CallHierarchyProvider)
		assert.Nil(t, result.Capabilities.CodeActionProvider)
	})

	t.Run("error path", func(t *testing.T) {
		stub := &stubLSP{
			onInitialize: func(context.Context, semanticapi.InitializeParams) (semanticapi.InitializeResult, error) {
				return semanticapi.InitializeResult{}, errors.New("init failed")
			},
		}
		env := newTestEnv(t, stub)
		_, err := env.client.Initialize(context.Background(), semanticapi.InitializeParams{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "init failed")
	})
}

func TestInitialized(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		called := false
		stub := &stubLSP{
			onInitialized: func(context.Context) error {
				called = true
				return nil
			},
		}
		env := newTestEnv(t, stub)
		err := env.client.Initialized(context.Background())
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("error path", func(t *testing.T) {
		stub := &stubLSP{
			onInitialized: func(context.Context) error {
				return errors.New("not ready")
			},
		}
		env := newTestEnv(t, stub)
		err := env.client.Initialized(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not ready")
	})
}

func TestShutdown(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		env := newTestEnv(t, &stubLSP{})
		require.NoError(t, env.client.Shutdown(context.Background()))
	})

	t.Run("error path", func(t *testing.T) {
		stub := &stubLSP{
			onShutdown: func(context.Context) error { return errors.New("shutdown err") },
		}
		env := newTestEnv(t, stub)
		err := env.client.Shutdown(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "shutdown err")
	})
}

func TestExit(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		env := newTestEnv(t, &stubLSP{})
		require.NoError(t, env.client.Exit(context.Background()))
	})

	t.Run("error path", func(t *testing.T) {
		stub := &stubLSP{
			onExit: func(context.Context) error { return errors.New("exit err") },
		}
		env := newTestEnv(t, stub)
		err := env.client.Exit(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exit err")
	})
}

func TestDidOpen(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var got semanticapi.DidOpenTextDocumentParams
		stub := &stubLSP{
			onDidOpen: func(_ context.Context, p semanticapi.DidOpenTextDocumentParams) error {
				got = p
				return nil
			},
		}
		env := newTestEnv(t, stub)
		params := semanticapi.DidOpenTextDocumentParams{
			TextDocument: semanticapi.TextDocumentItem{
				URI:        "file:///main.go",
				LanguageID: "go",
				Version:    1,
				Text:       "package main",
			},
		}
		require.NoError(t, env.client.DidOpen(context.Background(), params))
		assert.Equal(t, "file:///main.go", got.TextDocument.URI)
		assert.Equal(t, "go", got.TextDocument.LanguageID)
		assert.Equal(t, int32(1), got.TextDocument.Version)
		assert.Equal(t, "package main", got.TextDocument.Text)
	})

	t.Run("error path", func(t *testing.T) {
		stub := &stubLSP{
			onDidOpen: func(context.Context, semanticapi.DidOpenTextDocumentParams) error {
				return errors.New("open failed")
			},
		}
		env := newTestEnv(t, stub)
		err := env.client.DidOpen(context.Background(), semanticapi.DidOpenTextDocumentParams{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "open failed")
	})
}

func TestDidChange(t *testing.T) {
	t.Run("happy path with range change", func(t *testing.T) {
		var got semanticapi.DidChangeTextDocumentParams
		stub := &stubLSP{
			onDidChange: func(_ context.Context, p semanticapi.DidChangeTextDocumentParams) error {
				got = p
				return nil
			},
		}
		env := newTestEnv(t, stub)
		rng := semanticapi.Range{
			Start: semanticapi.Position{Line: 5, Character: 0},
			End:   semanticapi.Position{Line: 5, Character: 10},
		}
		params := semanticapi.DidChangeTextDocumentParams{
			TextDocument: semanticapi.VersionedTextDocumentIdentifier{
				URI: "file:///main.go", Version: 2,
			},
			ContentChanges: []semanticapi.TextDocumentContentChangeEvent{
				{Range: &rng, Text: "new text"},
			},
		}
		require.NoError(t, env.client.DidChange(context.Background(), params))
		assert.Equal(t, "file:///main.go", got.TextDocument.URI)
		assert.Equal(t, int32(2), got.TextDocument.Version)
		require.Len(t, got.ContentChanges, 1)
		require.NotNil(t, got.ContentChanges[0].Range)
		assert.Equal(t, uint32(5), got.ContentChanges[0].Range.Start.Line)
		assert.Equal(t, "new text", got.ContentChanges[0].Text)
	})

	t.Run("full document change (nil range)", func(t *testing.T) {
		var got semanticapi.DidChangeTextDocumentParams
		stub := &stubLSP{
			onDidChange: func(_ context.Context, p semanticapi.DidChangeTextDocumentParams) error {
				got = p
				return nil
			},
		}
		env := newTestEnv(t, stub)
		params := semanticapi.DidChangeTextDocumentParams{
			TextDocument: semanticapi.VersionedTextDocumentIdentifier{
				URI: "file:///main.go", Version: 3,
			},
			ContentChanges: []semanticapi.TextDocumentContentChangeEvent{
				{Text: "entire new content"},
			},
		}
		require.NoError(t, env.client.DidChange(context.Background(), params))
		require.Len(t, got.ContentChanges, 1)
		assert.Nil(t, got.ContentChanges[0].Range)
		assert.Equal(t, "entire new content", got.ContentChanges[0].Text)
	})
}

func TestDidClose(t *testing.T) {
	var got semanticapi.DidCloseTextDocumentParams
	stub := &stubLSP{
		onDidClose: func(_ context.Context, p semanticapi.DidCloseTextDocumentParams) error {
			got = p
			return nil
		},
	}
	env := newTestEnv(t, stub)
	require.NoError(t, env.client.DidClose(context.Background(), semanticapi.DidCloseTextDocumentParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
	}))
	assert.Equal(t, "file:///main.go", got.TextDocument.URI)
}

func TestDidSave(t *testing.T) {
	var got semanticapi.DidSaveTextDocumentParams
	stub := &stubLSP{
		onDidSave: func(_ context.Context, p semanticapi.DidSaveTextDocumentParams) error {
			got = p
			return nil
		},
	}
	env := newTestEnv(t, stub)
	require.NoError(t, env.client.DidSave(context.Background(), semanticapi.DidSaveTextDocumentParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
		Text:         "saved content",
	}))
	assert.Equal(t, "file:///main.go", got.TextDocument.URI)
	assert.Equal(t, "saved content", got.Text)
}

func TestCompletion(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		doc := &semanticapi.MarkupContent{Kind: semanticapi.MarkupKindMarkdown, Value: "doc"}
		te := semanticapi.TextEdit{
			Range:   semanticapi.Range{Start: semanticapi.Position{Line: 1, Character: 0}, End: semanticapi.Position{Line: 1, Character: 3}},
			NewText: "fmt",
		}
		stub := &stubLSP{
			onCompletion: func(_ context.Context, p semanticapi.CompletionParams) (semanticapi.CompletionResult, error) {
				assert.Equal(t, "file:///main.go", p.TextDocument.URI)
				assert.Equal(t, uint32(10), p.Position.Line)
				return semanticapi.CompletionResult{
					IsIncomplete: true,
					Items: []semanticapi.CompletionItem{
						{
							Label:            "fmt",
							Kind:             semanticapi.CompletionItemKindModule,
							Detail:           "package fmt",
							Documentation:    doc,
							InsertText:       "fmt",
							InsertTextFormat: semanticapi.InsertTextFormatPlainText,
							TextEdit:         &te,
						},
					},
				}, nil
			},
		}
		env := newTestEnv(t, stub)
		result, err := env.client.Completion(context.Background(), semanticapi.CompletionParams{
			TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
			Position:     semanticapi.Position{Line: 10, Character: 5},
		})
		require.NoError(t, err)
		assert.True(t, result.IsIncomplete)
		require.Len(t, result.Items, 1)
		item := result.Items[0]
		assert.Equal(t, "fmt", item.Label)
		assert.Equal(t, semanticapi.CompletionItemKindModule, item.Kind)
		assert.Equal(t, "package fmt", item.Detail)
		require.NotNil(t, item.Documentation)
		assert.Equal(t, semanticapi.MarkupKindMarkdown, item.Documentation.Kind)
		assert.Equal(t, "doc", item.Documentation.Value)
		assert.Equal(t, "fmt", item.InsertText)
		assert.Equal(t, semanticapi.InsertTextFormatPlainText, item.InsertTextFormat)
		require.NotNil(t, item.TextEdit)
		assert.Equal(t, "fmt", item.TextEdit.NewText)
	})

	t.Run("error path", func(t *testing.T) {
		stub := &stubLSP{
			onCompletion: func(context.Context, semanticapi.CompletionParams) (semanticapi.CompletionResult, error) {
				return semanticapi.CompletionResult{}, errors.New("completion err")
			},
		}
		env := newTestEnv(t, stub)
		_, err := env.client.Completion(context.Background(), semanticapi.CompletionParams{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "completion err")
	})
}

func TestHover(t *testing.T) {
	t.Run("happy path with result", func(t *testing.T) {
		rng := semanticapi.Range{
			Start: semanticapi.Position{Line: 1, Character: 0},
			End:   semanticapi.Position{Line: 1, Character: 5},
		}
		stub := &stubLSP{
			onHover: func(context.Context, semanticapi.HoverParams) (*semanticapi.Hover, error) {
				return &semanticapi.Hover{
					Contents: semanticapi.MarkupContent{Kind: semanticapi.MarkupKindMarkdown, Value: "func Foo()"},
					Range:    &rng,
				}, nil
			},
		}
		env := newTestEnv(t, stub)
		result, err := env.client.Hover(context.Background(), semanticapi.HoverParams{
			TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
			Position:     semanticapi.Position{Line: 1, Character: 2},
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "func Foo()", result.Contents.Value)
		assert.Equal(t, semanticapi.MarkupKindMarkdown, result.Contents.Kind)
		require.NotNil(t, result.Range)
		assert.Equal(t, uint32(1), result.Range.Start.Line)
	})

	t.Run("nil result", func(t *testing.T) {
		env := newTestEnv(t, &stubLSP{})
		result, err := env.client.Hover(context.Background(), semanticapi.HoverParams{})
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("error path", func(t *testing.T) {
		stub := &stubLSP{
			onHover: func(context.Context, semanticapi.HoverParams) (*semanticapi.Hover, error) {
				return nil, errors.New("hover err")
			},
		}
		env := newTestEnv(t, stub)
		_, err := env.client.Hover(context.Background(), semanticapi.HoverParams{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "hover err")
	})
}

func TestSignatureHelp(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		doc := &semanticapi.MarkupContent{Kind: semanticapi.MarkupKindPlainText, Value: "param doc"}
		stub := &stubLSP{
			onSignatureHelp: func(context.Context, semanticapi.SignatureHelpParams) (*semanticapi.SignatureHelp, error) {
				return &semanticapi.SignatureHelp{
					Signatures: []semanticapi.SignatureInformation{
						{
							Label:         "func Foo(x int, y string)",
							Documentation: &semanticapi.MarkupContent{Kind: semanticapi.MarkupKindMarkdown, Value: "Foo does things"},
							Parameters: []semanticapi.ParameterInformation{
								{Label: "x int", Documentation: doc},
								{Label: "y string"},
							},
						},
					},
					ActiveSignature: 0,
					ActiveParameter: 1,
				}, nil
			},
		}
		env := newTestEnv(t, stub)
		result, err := env.client.SignatureHelp(context.Background(), semanticapi.SignatureHelpParams{
			TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
			Position:     semanticapi.Position{Line: 10, Character: 20},
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Signatures, 1)
		assert.Equal(t, "func Foo(x int, y string)", result.Signatures[0].Label)
		require.Len(t, result.Signatures[0].Parameters, 2)
		assert.Equal(t, "x int", result.Signatures[0].Parameters[0].Label)
		require.NotNil(t, result.Signatures[0].Parameters[0].Documentation)
		assert.Equal(t, "param doc", result.Signatures[0].Parameters[0].Documentation.Value)
		assert.Nil(t, result.Signatures[0].Parameters[1].Documentation)
		assert.Equal(t, uint32(1), result.ActiveParameter)
	})

	t.Run("nil result", func(t *testing.T) {
		env := newTestEnv(t, &stubLSP{})
		result, err := env.client.SignatureHelp(context.Background(), semanticapi.SignatureHelpParams{})
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestDefinition(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		stub := &stubLSP{
			onDefinition: func(_ context.Context, p semanticapi.DefinitionParams) (semanticapi.LocationResult, error) {
				assert.Equal(t, uint32(5), p.Position.Line)
				return semanticapi.LocationResult{
					Locations: []semanticapi.Location{
						{URI: "file:///other.go", Range: semanticapi.Range{
							Start: semanticapi.Position{Line: 10, Character: 0},
							End:   semanticapi.Position{Line: 10, Character: 5},
						}},
					},
				}, nil
			},
		}
		env := newTestEnv(t, stub)
		result, err := env.client.Definition(context.Background(), semanticapi.DefinitionParams{
			TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
			Position:     semanticapi.Position{Line: 5, Character: 3},
		})
		require.NoError(t, err)
		require.Len(t, result.Locations, 1)
		assert.Equal(t, "file:///other.go", result.Locations[0].URI)
		assert.Equal(t, uint32(10), result.Locations[0].Range.Start.Line)
	})

	t.Run("error path", func(t *testing.T) {
		stub := &stubLSP{
			onDefinition: func(context.Context, semanticapi.DefinitionParams) (semanticapi.LocationResult, error) {
				return semanticapi.LocationResult{}, errors.New("def err")
			},
		}
		env := newTestEnv(t, stub)
		_, err := env.client.Definition(context.Background(), semanticapi.DefinitionParams{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "def err")
	})
}

func TestDeclaration(t *testing.T) {
	stub := &stubLSP{
		onDeclaration: func(context.Context, semanticapi.DeclarationParams) (semanticapi.LocationResult, error) {
			return semanticapi.LocationResult{
				Locations: []semanticapi.Location{{URI: "file:///decl.go"}},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	result, err := env.client.Declaration(context.Background(), semanticapi.DeclarationParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
	})
	require.NoError(t, err)
	require.Len(t, result.Locations, 1)
	assert.Equal(t, "file:///decl.go", result.Locations[0].URI)
}

func TestTypeDefinition(t *testing.T) {
	stub := &stubLSP{
		onTypeDefinition: func(context.Context, semanticapi.TypeDefinitionParams) (semanticapi.LocationResult, error) {
			return semanticapi.LocationResult{
				Locations: []semanticapi.Location{{URI: "file:///types.go"}},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	result, err := env.client.TypeDefinition(context.Background(), semanticapi.TypeDefinitionParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
	})
	require.NoError(t, err)
	require.Len(t, result.Locations, 1)
	assert.Equal(t, "file:///types.go", result.Locations[0].URI)
}

func TestImplementation(t *testing.T) {
	stub := &stubLSP{
		onImplementation: func(context.Context, semanticapi.ImplementationParams) (semanticapi.LocationResult, error) {
			return semanticapi.LocationResult{
				Locations: []semanticapi.Location{
					{URI: "file:///impl1.go"},
					{URI: "file:///impl2.go"},
				},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	result, err := env.client.Implementation(context.Background(), semanticapi.ImplementationParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
	})
	require.NoError(t, err)
	require.Len(t, result.Locations, 2)
}

func TestReferences(t *testing.T) {
	t.Run("happy path with include declaration", func(t *testing.T) {
		var got semanticapi.ReferenceParams
		stub := &stubLSP{
			onReferences: func(_ context.Context, p semanticapi.ReferenceParams) ([]semanticapi.Location, error) {
				got = p
				return []semanticapi.Location{{URI: "file:///ref.go"}}, nil
			},
		}
		env := newTestEnv(t, stub)
		locs, err := env.client.References(context.Background(), semanticapi.ReferenceParams{
			TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
			Position:     semanticapi.Position{Line: 3, Character: 7},
			Context:      semanticapi.ReferenceContext{IncludeDeclaration: true},
		})
		require.NoError(t, err)
		require.Len(t, locs, 1)
		assert.True(t, got.Context.IncludeDeclaration)
	})
}

func TestDocumentHighlight(t *testing.T) {
	stub := &stubLSP{
		onDocumentHighlight: func(context.Context, semanticapi.DocumentHighlightParams) ([]semanticapi.DocumentHighlight, error) {
			return []semanticapi.DocumentHighlight{
				{
					Range: semanticapi.Range{Start: semanticapi.Position{Line: 1}, End: semanticapi.Position{Line: 1, Character: 5}},
					Kind:  semanticapi.DocumentHighlightKindWrite,
				},
				{
					Range: semanticapi.Range{Start: semanticapi.Position{Line: 5}, End: semanticapi.Position{Line: 5, Character: 5}},
					Kind:  semanticapi.DocumentHighlightKindRead,
				},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	highlights, err := env.client.DocumentHighlight(context.Background(), semanticapi.DocumentHighlightParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
	})
	require.NoError(t, err)
	require.Len(t, highlights, 2)
	assert.Equal(t, semanticapi.DocumentHighlightKindWrite, highlights[0].Kind)
	assert.Equal(t, semanticapi.DocumentHighlightKindRead, highlights[1].Kind)
}

func TestDocumentSymbol(t *testing.T) {
	stub := &stubLSP{
		onDocumentSymbol: func(context.Context, semanticapi.DocumentSymbolParams) (semanticapi.DocumentSymbolResult, error) {
			return semanticapi.DocumentSymbolResult{
				DocumentSymbols: []semanticapi.DocumentSymbol{
					{
						Name:   "main",
						Detail: "func",
						Kind:   semanticapi.SymbolKindFunction,
						Range:  semanticapi.Range{Start: semanticapi.Position{Line: 3}, End: semanticapi.Position{Line: 10}},
						Children: []semanticapi.DocumentSymbol{
							{Name: "x", Kind: semanticapi.SymbolKindVariable},
						},
					},
				},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	result, err := env.client.DocumentSymbol(context.Background(), semanticapi.DocumentSymbolParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
	})
	require.NoError(t, err)
	syms := result.DocumentSymbols
	require.Len(t, syms, 1)
	assert.Equal(t, "main", syms[0].Name)
	assert.Equal(t, semanticapi.SymbolKindFunction, syms[0].Kind)
	require.Len(t, syms[0].Children, 1)
	assert.Equal(t, "x", syms[0].Children[0].Name)
}

func TestCodeAction(t *testing.T) {
	stub := &stubLSP{
		onCodeAction: func(_ context.Context, p semanticapi.CodeActionParams) ([]semanticapi.CodeActionResult, error) {
			assert.Len(t, p.Context.Diagnostics, 1)
			return []semanticapi.CodeActionResult{
				{
					CodeAction: &semanticapi.CodeAction{
						Title: "Organize Imports",
						Kind:  semanticapi.CodeActionKindSourceOrganizeImports,
						Edit: &semanticapi.WorkspaceEdit{
							Changes: map[string][]semanticapi.TextEdit{
								"file:///main.go": {{
									Range:   semanticapi.Range{Start: semanticapi.Position{Line: 2}, End: semanticapi.Position{Line: 4}},
									NewText: "import \"fmt\"\n",
								}},
							},
						},
					},
				},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	actions, err := env.client.CodeAction(context.Background(), semanticapi.CodeActionParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
		Range:        semanticapi.Range{Start: semanticapi.Position{Line: 2}, End: semanticapi.Position{Line: 4}},
		Context: semanticapi.CodeActionContext{
			Diagnostics: []semanticapi.Diagnostic{
				{Message: "unused import", Severity: semanticapi.DiagnosticSeverityWarning},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, actions, 1)
	require.NotNil(t, actions[0].CodeAction)
	assert.Equal(t, "Organize Imports", actions[0].CodeAction.Title)
	assert.Equal(t, semanticapi.CodeActionKindSourceOrganizeImports, actions[0].CodeAction.Kind)
	require.NotNil(t, actions[0].CodeAction.Edit)
	edits, ok := actions[0].CodeAction.Edit.Changes["file:///main.go"]
	require.True(t, ok)
	require.Len(t, edits, 1)
	assert.Equal(t, "import \"fmt\"\n", edits[0].NewText)
}

func TestCodeLens(t *testing.T) {
	stub := &stubLSP{
		onCodeLens: func(context.Context, semanticapi.CodeLensParams) ([]semanticapi.CodeLens, error) {
			return []semanticapi.CodeLens{
				{
					Range:   semanticapi.Range{Start: semanticapi.Position{Line: 5}},
					Command: &semanticapi.Command{Title: "Run Test", Command: "test.run", Arguments: []json.RawMessage{json.RawMessage(`"TestFoo"`)}},
				},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	lenses, err := env.client.CodeLens(context.Background(), semanticapi.CodeLensParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main_test.go"},
	})
	require.NoError(t, err)
	require.Len(t, lenses, 1)
	require.NotNil(t, lenses[0].Command)
	assert.Equal(t, "Run Test", lenses[0].Command.Title)
	assert.Equal(t, []json.RawMessage{json.RawMessage(`"TestFoo"`)}, lenses[0].Command.Arguments)
}

func TestFormatting(t *testing.T) {
	stub := &stubLSP{
		onFormatting: func(_ context.Context, p semanticapi.DocumentFormattingParams) ([]semanticapi.TextEdit, error) {
			assert.Equal(t, uint32(4), p.Options.TabSize)
			assert.True(t, p.Options.InsertSpaces)
			return []semanticapi.TextEdit{
				{Range: semanticapi.Range{Start: semanticapi.Position{Line: 0}, End: semanticapi.Position{Line: 0, Character: 3}}, NewText: "\t"},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	edits, err := env.client.Formatting(context.Background(), semanticapi.DocumentFormattingParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
		Options:      semanticapi.FormattingOptions{TabSize: 4, InsertSpaces: true},
	})
	require.NoError(t, err)
	require.Len(t, edits, 1)
	assert.Equal(t, "\t", edits[0].NewText)
}

func TestRangeFormatting(t *testing.T) {
	stub := &stubLSP{
		onRangeFormatting: func(_ context.Context, p semanticapi.DocumentRangeFormattingParams) ([]semanticapi.TextEdit, error) {
			assert.Equal(t, uint32(5), p.Range.Start.Line)
			return []semanticapi.TextEdit{{NewText: "formatted"}}, nil
		},
	}
	env := newTestEnv(t, stub)
	edits, err := env.client.RangeFormatting(context.Background(), semanticapi.DocumentRangeFormattingParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
		Range:        semanticapi.Range{Start: semanticapi.Position{Line: 5}, End: semanticapi.Position{Line: 10}},
	})
	require.NoError(t, err)
	require.Len(t, edits, 1)
}

func TestRename(t *testing.T) {
	t.Run("happy path with result", func(t *testing.T) {
		stub := &stubLSP{
			onRename: func(_ context.Context, p semanticapi.RenameParams) (*semanticapi.WorkspaceEdit, error) {
				assert.Equal(t, "newName", p.NewName)
				return &semanticapi.WorkspaceEdit{
					Changes: map[string][]semanticapi.TextEdit{
						"file:///main.go":  {{NewText: "newName"}},
						"file:///other.go": {{NewText: "newName"}, {NewText: "newName"}},
					},
				}, nil
			},
		}
		env := newTestEnv(t, stub)
		result, err := env.client.Rename(context.Background(), semanticapi.RenameParams{
			TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
			Position:     semanticapi.Position{Line: 3, Character: 5},
			NewName:      "newName",
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Changes["file:///main.go"], 1)
		assert.Len(t, result.Changes["file:///other.go"], 2)
	})

	t.Run("nil result", func(t *testing.T) {
		env := newTestEnv(t, &stubLSP{})
		result, err := env.client.Rename(context.Background(), semanticapi.RenameParams{NewName: "x"})
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("error path", func(t *testing.T) {
		stub := &stubLSP{
			onRename: func(context.Context, semanticapi.RenameParams) (*semanticapi.WorkspaceEdit, error) {
				return nil, errors.New("rename err")
			},
		}
		env := newTestEnv(t, stub)
		_, err := env.client.Rename(context.Background(), semanticapi.RenameParams{NewName: "x"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rename err")
	})
}

func TestPrepareRename(t *testing.T) {
	t.Run("happy path with result", func(t *testing.T) {
		stub := &stubLSP{
			onPrepareRename: func(context.Context, semanticapi.PrepareRenameParams) (*semanticapi.PrepareRenameResult, error) {
				return &semanticapi.PrepareRenameResult{
					Range:       semanticapi.Range{Start: semanticapi.Position{Line: 3, Character: 5}, End: semanticapi.Position{Line: 3, Character: 8}},
					Placeholder: "foo",
				}, nil
			},
		}
		env := newTestEnv(t, stub)
		result, err := env.client.PrepareRename(context.Background(), semanticapi.PrepareRenameParams{
			TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
			Position:     semanticapi.Position{Line: 3, Character: 5},
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "foo", result.Placeholder)
		assert.Equal(t, uint32(5), result.Range.Start.Character)
	})

	t.Run("nil result", func(t *testing.T) {
		env := newTestEnv(t, &stubLSP{})
		result, err := env.client.PrepareRename(context.Background(), semanticapi.PrepareRenameParams{})
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestFoldingRange(t *testing.T) {
	stub := &stubLSP{
		onFoldingRange: func(context.Context, semanticapi.FoldingRangeParams) ([]semanticapi.FoldingRange, error) {
			return []semanticapi.FoldingRange{
				{StartLine: 1, EndLine: 5, Kind: semanticapi.FoldingRangeKindImports},
				{StartLine: 7, EndLine: 20, Kind: semanticapi.FoldingRangeKindRegion},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	ranges, err := env.client.FoldingRange(context.Background(), semanticapi.FoldingRangeParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
	})
	require.NoError(t, err)
	require.Len(t, ranges, 2)
	assert.Equal(t, semanticapi.FoldingRangeKindImports, ranges[0].Kind)
	assert.Equal(t, uint32(7), ranges[1].StartLine)
}

func TestSelectionRange(t *testing.T) {
	stub := &stubLSP{
		onSelectionRange: func(_ context.Context, p semanticapi.SelectionRangeParams) ([]semanticapi.SelectionRange, error) {
			assert.Len(t, p.Positions, 2)
			return []semanticapi.SelectionRange{
				{
					Range: semanticapi.Range{Start: semanticapi.Position{Line: 1}, End: semanticapi.Position{Line: 1, Character: 10}},
					Parent: &semanticapi.SelectionRange{
						Range: semanticapi.Range{Start: semanticapi.Position{Line: 0}, End: semanticapi.Position{Line: 5}},
					},
				},
				{
					Range: semanticapi.Range{Start: semanticapi.Position{Line: 3}, End: semanticapi.Position{Line: 3, Character: 5}},
				},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	ranges, err := env.client.SelectionRange(context.Background(), semanticapi.SelectionRangeParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
		Positions:    []semanticapi.Position{{Line: 1, Character: 5}, {Line: 3, Character: 2}},
	})
	require.NoError(t, err)
	require.Len(t, ranges, 2)
	require.NotNil(t, ranges[0].Parent)
	assert.Equal(t, uint32(0), ranges[0].Parent.Range.Start.Line)
	assert.Nil(t, ranges[1].Parent)
}

func TestSemanticTokensFull(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		stub := &stubLSP{
			onSemanticTokensFull: func(context.Context, semanticapi.SemanticTokensParams) (*semanticapi.SemanticTokens, error) {
				return &semanticapi.SemanticTokens{
					ResultID: "abc123",
					Data:     []uint32{0, 0, 7, 1, 0, 1, 0, 4, 2, 0},
				}, nil
			},
		}
		env := newTestEnv(t, stub)
		result, err := env.client.SemanticTokensFull(context.Background(), semanticapi.SemanticTokensParams{
			TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "abc123", result.ResultID)
		assert.Equal(t, []uint32{0, 0, 7, 1, 0, 1, 0, 4, 2, 0}, result.Data)
	})

	t.Run("nil result", func(t *testing.T) {
		env := newTestEnv(t, &stubLSP{})
		result, err := env.client.SemanticTokensFull(context.Background(), semanticapi.SemanticTokensParams{})
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestSemanticTokensRange(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		stub := &stubLSP{
			onSemanticTokensRange: func(_ context.Context, p semanticapi.SemanticTokensRangeParams) (*semanticapi.SemanticTokens, error) {
				assert.Equal(t, uint32(5), p.Range.Start.Line)
				return &semanticapi.SemanticTokens{Data: []uint32{1, 2, 3}}, nil
			},
		}
		env := newTestEnv(t, stub)
		result, err := env.client.SemanticTokensRange(context.Background(), semanticapi.SemanticTokensRangeParams{
			TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
			Range:        semanticapi.Range{Start: semanticapi.Position{Line: 5}, End: semanticapi.Position{Line: 10}},
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, []uint32{1, 2, 3}, result.Data)
	})

	t.Run("nil result", func(t *testing.T) {
		env := newTestEnv(t, &stubLSP{})
		result, err := env.client.SemanticTokensRange(context.Background(), semanticapi.SemanticTokensRangeParams{})
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestDiagnostic(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		stub := &stubLSP{
			onDiagnostic: func(context.Context, semanticapi.DocumentDiagnosticParams) (semanticapi.DocumentDiagnosticReport, error) {
				return semanticapi.DocumentDiagnosticReport{
					Kind:     "full",
					ResultID: "diag-1",
					Items: []semanticapi.Diagnostic{
						{
							Range:    semanticapi.Range{Start: semanticapi.Position{Line: 10}, End: semanticapi.Position{Line: 10, Character: 20}},
							Severity: semanticapi.DiagnosticSeverityError,
							Code:     "E001",
							Source:   "golint",
							Message:  "undefined: foo",
						},
					},
					RelatedDocuments: map[string]semanticapi.DocumentDiagnosticReport{
						"file:///other.go": {
							Items: []semanticapi.Diagnostic{
								{Message: "related", Severity: semanticapi.DiagnosticSeverityHint},
							},
						},
					},
				}, nil
			},
		}
		env := newTestEnv(t, stub)
		report, err := env.client.Diagnostic(context.Background(), semanticapi.DocumentDiagnosticParams{
			TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
		})
		require.NoError(t, err)
		assert.Equal(t, "full", report.Kind)
		assert.Equal(t, "diag-1", report.ResultID)
		require.Len(t, report.Items, 1)
		assert.Equal(t, "undefined: foo", report.Items[0].Message)
		assert.Equal(t, semanticapi.DiagnosticSeverityError, report.Items[0].Severity)
		assert.Equal(t, "E001", report.Items[0].Code)
		assert.Equal(t, "golint", report.Items[0].Source)
		require.Contains(t, report.RelatedDocuments, "file:///other.go")
		assert.Len(t, report.RelatedDocuments["file:///other.go"].Items, 1)
	})

	t.Run("error path", func(t *testing.T) {
		stub := &stubLSP{
			onDiagnostic: func(context.Context, semanticapi.DocumentDiagnosticParams) (semanticapi.DocumentDiagnosticReport, error) {
				return semanticapi.DocumentDiagnosticReport{}, errors.New("diag err")
			},
		}
		env := newTestEnv(t, stub)
		_, err := env.client.Diagnostic(context.Background(), semanticapi.DocumentDiagnosticParams{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "diag err")
	})
}

func TestWorkspaceSymbol(t *testing.T) {
	stub := &stubLSP{
		onWorkspaceSymbol: func(_ context.Context, p semanticapi.WorkspaceSymbolParams) ([]semanticapi.SymbolInformation, error) {
			assert.Equal(t, "Foo", p.Query)
			return []semanticapi.SymbolInformation{
				{
					Name:     "FooBar",
					Kind:     semanticapi.SymbolKindFunction,
					Location: semanticapi.Location{URI: "file:///pkg.go", Range: semanticapi.Range{Start: semanticapi.Position{Line: 42}}},
				},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	syms, err := env.client.WorkspaceSymbol(context.Background(), semanticapi.WorkspaceSymbolParams{Query: "Foo"})
	require.NoError(t, err)
	require.Len(t, syms, 1)
	assert.Equal(t, "FooBar", syms[0].Name)
	assert.Equal(t, semanticapi.SymbolKindFunction, syms[0].Kind)
	assert.Equal(t, "file:///pkg.go", syms[0].Location.URI)
}

func TestExecuteCommand(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		stub := &stubLSP{
			onExecuteCommand: func(_ context.Context, p semanticapi.ExecuteCommandParams) (string, error) {
				assert.Equal(t, "editor.action.organizeImports", p.Command)
				assert.Equal(t, []json.RawMessage{json.RawMessage(`"arg1"`), json.RawMessage(`"arg2"`)}, p.Arguments)
				return `{"ok":true}`, nil
			},
		}
		env := newTestEnv(t, stub)
		result, err := env.client.ExecuteCommand(context.Background(), semanticapi.ExecuteCommandParams{
			Command:   "editor.action.organizeImports",
			Arguments: []json.RawMessage{json.RawMessage(`"arg1"`), json.RawMessage(`"arg2"`)},
		})
		require.NoError(t, err)
		assert.Equal(t, `{"ok":true}`, result)
	})

	t.Run("error path", func(t *testing.T) {
		stub := &stubLSP{
			onExecuteCommand: func(context.Context, semanticapi.ExecuteCommandParams) (string, error) {
				return "", errors.New("cmd err")
			},
		}
		env := newTestEnv(t, stub)
		_, err := env.client.ExecuteCommand(context.Background(), semanticapi.ExecuteCommandParams{Command: "x"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cmd err")
	})
}

func TestPrepareCallHierarchy(t *testing.T) {
	stub := &stubLSP{
		onPrepareCallHierarchy: func(context.Context, semanticapi.CallHierarchyPrepareParams) ([]semanticapi.CallHierarchyItem, error) {
			return []semanticapi.CallHierarchyItem{
				{
					Name:           "Foo",
					Kind:           semanticapi.SymbolKindFunction,
					URI:            "file:///main.go",
					Range:          semanticapi.Range{Start: semanticapi.Position{Line: 10}, End: semanticapi.Position{Line: 20}},
					SelectionRange: semanticapi.Range{Start: semanticapi.Position{Line: 10, Character: 5}, End: semanticapi.Position{Line: 10, Character: 8}},
				},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	items, err := env.client.PrepareCallHierarchy(context.Background(), semanticapi.CallHierarchyPrepareParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
		Position:     semanticapi.Position{Line: 10, Character: 6},
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Foo", items[0].Name)
	assert.Equal(t, semanticapi.SymbolKindFunction, items[0].Kind)
	assert.Equal(t, "file:///main.go", items[0].URI)
}

func TestCallHierarchyIncomingCalls(t *testing.T) {
	stub := &stubLSP{
		onCallHierarchyIncomingCalls: func(_ context.Context, p semanticapi.CallHierarchyIncomingCallsParams) ([]semanticapi.CallHierarchyIncomingCall, error) {
			assert.Equal(t, "Foo", p.Item.Name)
			return []semanticapi.CallHierarchyIncomingCall{
				{
					From: semanticapi.CallHierarchyItem{Name: "Bar", Kind: semanticapi.SymbolKindFunction, URI: "file:///bar.go"},
					FromRanges: []semanticapi.Range{
						{Start: semanticapi.Position{Line: 15, Character: 3}, End: semanticapi.Position{Line: 15, Character: 6}},
					},
				},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	calls, err := env.client.CallHierarchyIncomingCalls(context.Background(), semanticapi.CallHierarchyIncomingCallsParams{
		Item: semanticapi.CallHierarchyItem{Name: "Foo", Kind: semanticapi.SymbolKindFunction, URI: "file:///main.go"},
	})
	require.NoError(t, err)
	require.Len(t, calls, 1)
	assert.Equal(t, "Bar", calls[0].From.Name)
	require.Len(t, calls[0].FromRanges, 1)
	assert.Equal(t, uint32(15), calls[0].FromRanges[0].Start.Line)
}

func TestCallHierarchyOutgoingCalls(t *testing.T) {
	stub := &stubLSP{
		onCallHierarchyOutgoingCalls: func(_ context.Context, p semanticapi.CallHierarchyOutgoingCallsParams) ([]semanticapi.CallHierarchyOutgoingCall, error) {
			assert.Equal(t, "Foo", p.Item.Name)
			return []semanticapi.CallHierarchyOutgoingCall{
				{
					To:         semanticapi.CallHierarchyItem{Name: "Baz", Kind: semanticapi.SymbolKindMethod, URI: "file:///baz.go"},
					FromRanges: []semanticapi.Range{{Start: semanticapi.Position{Line: 12}}},
				},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	calls, err := env.client.CallHierarchyOutgoingCalls(context.Background(), semanticapi.CallHierarchyOutgoingCallsParams{
		Item: semanticapi.CallHierarchyItem{Name: "Foo"},
	})
	require.NoError(t, err)
	require.Len(t, calls, 1)
	assert.Equal(t, "Baz", calls[0].To.Name)
	assert.Equal(t, semanticapi.SymbolKindMethod, calls[0].To.Kind)
}

func TestCodeActionWithCommand(t *testing.T) {
	stub := &stubLSP{
		onCodeAction: func(context.Context, semanticapi.CodeActionParams) ([]semanticapi.CodeActionResult, error) {
			return []semanticapi.CodeActionResult{
				{
					CodeAction: &semanticapi.CodeAction{
						Title:   "Run fix",
						Kind:    semanticapi.CodeActionKindQuickFix,
						Command: &semanticapi.Command{Title: "Fix it", Command: "fix.it", Arguments: []json.RawMessage{json.RawMessage(`"a"`)}},
					},
				},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	actions, err := env.client.CodeAction(context.Background(), semanticapi.CodeActionParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
	})
	require.NoError(t, err)
	require.Len(t, actions, 1)
	require.NotNil(t, actions[0].CodeAction)
	assert.Nil(t, actions[0].CodeAction.Edit)
	require.NotNil(t, actions[0].CodeAction.Command)
	assert.Equal(t, "Fix it", actions[0].CodeAction.Command.Title)
	assert.Equal(t, "fix.it", actions[0].CodeAction.Command.Command)
	assert.Equal(t, []json.RawMessage{json.RawMessage(`"a"`)}, actions[0].CodeAction.Command.Arguments)
}

func TestCompletionResolve(t *testing.T) {
	stub := &stubLSP{
		onCompletionResolve: func(_ context.Context, item semanticapi.CompletionItem) (semanticapi.CompletionItem, error) {
			item.Detail = "resolved detail"
			return item, nil
		},
	}
	env := newTestEnv(t, stub)
	result, err := env.client.CompletionResolve(context.Background(), semanticapi.CompletionItem{
		Label: "fmt",
		Kind:  semanticapi.CompletionItemKindModule,
	})
	require.NoError(t, err)
	assert.Equal(t, "fmt", result.Label)
	assert.Equal(t, "resolved detail", result.Detail)
}

func TestCodeLensResolve(t *testing.T) {
	stub := &stubLSP{
		onCodeLensResolve: func(_ context.Context, lens semanticapi.CodeLens) (semanticapi.CodeLens, error) {
			lens.Command = &semanticapi.Command{Title: "Resolved", Command: "resolved.cmd"}
			return lens, nil
		},
	}
	env := newTestEnv(t, stub)
	result, err := env.client.CodeLensResolve(context.Background(), semanticapi.CodeLens{
		Range: semanticapi.Range{Start: semanticapi.Position{Line: 5}},
	})
	require.NoError(t, err)
	require.NotNil(t, result.Command)
	assert.Equal(t, "Resolved", result.Command.Title)
}

func TestDocumentColor(t *testing.T) {
	stub := &stubLSP{
		onDocumentColor: func(_ context.Context, p semanticapi.DocumentColorParams) ([]semanticapi.ColorInformation, error) {
			return []semanticapi.ColorInformation{
				{
					Range: semanticapi.Range{Start: semanticapi.Position{Line: 1}, End: semanticapi.Position{Line: 1, Character: 10}},
					Color: semanticapi.Color{Red: 1.0, Green: 0.0, Blue: 0.0, Alpha: 1.0},
				},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	result, err := env.client.DocumentColor(context.Background(), semanticapi.DocumentColorParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
	})
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, 1.0, result[0].Color.Red)
	assert.Equal(t, 0.0, result[0].Color.Blue)
}

func TestDocumentLink(t *testing.T) {
	stub := &stubLSP{
		onDocumentLink: func(_ context.Context, p semanticapi.DocumentLinkParams) ([]semanticapi.DocumentLink, error) {
			return []semanticapi.DocumentLink{
				{
					Range:   semanticapi.Range{Start: semanticapi.Position{Line: 3}, End: semanticapi.Position{Line: 3, Character: 20}},
					Target:  "https://example.com",
					Tooltip: "Go to example",
				},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	result, err := env.client.DocumentLink(context.Background(), semanticapi.DocumentLinkParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
	})
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "https://example.com", result[0].Target)
	assert.Equal(t, "Go to example", result[0].Tooltip)
}

func TestPrepareTypeHierarchy(t *testing.T) {
	stub := &stubLSP{
		onPrepareTypeHierarchy: func(_ context.Context, p semanticapi.TypeHierarchyPrepareParams) ([]semanticapi.TypeHierarchyItem, error) {
			return []semanticapi.TypeHierarchyItem{
				{
					Name:           "MyInterface",
					Kind:           semanticapi.SymbolKindInterface,
					URI:            "file:///types.go",
					Range:          semanticapi.Range{Start: semanticapi.Position{Line: 5}, End: semanticapi.Position{Line: 10}},
					SelectionRange: semanticapi.Range{Start: semanticapi.Position{Line: 5, Character: 5}, End: semanticapi.Position{Line: 5, Character: 16}},
				},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	items, err := env.client.PrepareTypeHierarchy(context.Background(), semanticapi.TypeHierarchyPrepareParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///types.go"},
		Position:     semanticapi.Position{Line: 5, Character: 8},
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "MyInterface", items[0].Name)
	assert.Equal(t, semanticapi.SymbolKindInterface, items[0].Kind)
}

func TestInlayHint(t *testing.T) {
	stub := &stubLSP{
		onInlayHint: func(_ context.Context, p semanticapi.InlayHintParams) ([]semanticapi.InlayHint, error) {
			return []semanticapi.InlayHint{
				{
					Position:     semanticapi.Position{Line: 10, Character: 15},
					Label:        "int",
					Kind:         semanticapi.InlayHintKindType,
					PaddingLeft:  true,
					PaddingRight: false,
				},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	result, err := env.client.InlayHint(context.Background(), semanticapi.InlayHintParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
		Range:        semanticapi.Range{Start: semanticapi.Position{Line: 0}, End: semanticapi.Position{Line: 20}},
	})
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "int", result[0].Label)
	assert.Equal(t, semanticapi.InlayHintKindType, result[0].Kind)
	assert.True(t, result[0].PaddingLeft)
}

func TestWillSave(t *testing.T) {
	called := false
	stub := &stubLSP{
		onWillSave: func(_ context.Context, p semanticapi.WillSaveTextDocumentParams) error {
			assert.Equal(t, "file:///main.go", p.TextDocument.URI)
			assert.Equal(t, semanticapi.TextDocumentSaveReasonManual, p.Reason)
			called = true
			return nil
		},
	}
	env := newTestEnv(t, stub)
	err := env.client.WillSave(context.Background(), semanticapi.WillSaveTextDocumentParams{
		TextDocument: semanticapi.TextDocumentIdentifier{URI: "file:///main.go"},
		Reason:       semanticapi.TextDocumentSaveReasonManual,
	})
	require.NoError(t, err)
	assert.True(t, called)
}

func TestDidChangeConfiguration(t *testing.T) {
	called := false
	stub := &stubLSP{
		onDidChangeConfiguration: func(_ context.Context, p semanticapi.DidChangeConfigurationParams) error {
			assert.Equal(t, json.RawMessage(`{"setting":"value"}`), p.Settings)
			called = true
			return nil
		},
	}
	env := newTestEnv(t, stub)
	err := env.client.DidChangeConfiguration(context.Background(), semanticapi.DidChangeConfigurationParams{
		Settings: json.RawMessage(`{"setting":"value"}`),
	})
	require.NoError(t, err)
	assert.True(t, called)
}

func TestSetTrace(t *testing.T) {
	called := false
	stub := &stubLSP{
		onSetTrace: func(_ context.Context, p semanticapi.SetTraceParams) error {
			assert.Equal(t, semanticapi.TraceValueVerbose, p.Value)
			called = true
			return nil
		},
	}
	env := newTestEnv(t, stub)
	err := env.client.SetTrace(context.Background(), semanticapi.SetTraceParams{
		Value: semanticapi.TraceValueVerbose,
	})
	require.NoError(t, err)
	assert.True(t, called)
}

func TestWillCreateFiles(t *testing.T) {
	stub := &stubLSP{
		onWillCreateFiles: func(_ context.Context, p semanticapi.CreateFilesParams) (*semanticapi.WorkspaceEdit, error) {
			require.Len(t, p.Files, 1)
			assert.Equal(t, "file:///new.go", p.Files[0].URI)
			return &semanticapi.WorkspaceEdit{
				Changes: map[string][]semanticapi.TextEdit{
					"file:///go.mod": {{NewText: "// updated"}},
				},
			}, nil
		},
	}
	env := newTestEnv(t, stub)
	result, err := env.client.WillCreateFiles(context.Background(), semanticapi.CreateFilesParams{
		Files: []semanticapi.FileCreate{{URI: "file:///new.go"}},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Changes["file:///go.mod"], 1)
}

func TestDidChangeWorkspaceFolders(t *testing.T) {
	called := false
	stub := &stubLSP{
		onDidChangeWorkspaceFolders: func(_ context.Context, p semanticapi.DidChangeWorkspaceFoldersParams) error {
			require.Len(t, p.Event.Added, 1)
			require.Len(t, p.Event.Removed, 1)
			assert.Equal(t, "file:///new", p.Event.Added[0].URI)
			assert.Equal(t, "file:///old", p.Event.Removed[0].URI)
			called = true
			return nil
		},
	}
	env := newTestEnv(t, stub)
	err := env.client.DidChangeWorkspaceFolders(context.Background(), semanticapi.DidChangeWorkspaceFoldersParams{
		Event: semanticapi.WorkspaceFoldersChangeEvent{
			Added:   []semanticapi.WorkspaceFolder{{URI: "file:///new", Name: "new"}},
			Removed: []semanticapi.WorkspaceFolder{{URI: "file:///old", Name: "old"}},
		},
	})
	require.NoError(t, err)
	assert.True(t, called)
}
