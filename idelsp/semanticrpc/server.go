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

package semanticrpc

import (
	"context"
	"encoding/json"

	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
	"github.com/unstablebuild/rune-go-sdk/api/semanticapi/semanticrpc"
)

type Server struct {
	semanticrpc.UnimplementedLSPServer
	impl      semanticapi.LSP
	ctx       context.Context
	cancelCtx func()
}

// NewServer returns an LSPServer that delegates to impl.
func NewServer(impl semanticapi.LSP) *Server {
	ctx, cancelCtx := context.WithCancel(context.Background())
	return &Server{impl: impl, ctx: ctx, cancelCtx: cancelCtx}
}

func blueCtxFirst(parent context.Context, other context.Context) (context.Context, func()) {
	deadline, hasDeadline := parent.Deadline()
	if d, ok := other.Deadline(); ok {
		if !hasDeadline || d.Before(deadline) {
			deadline, hasDeadline = d, true
		}
	}

	var ctx context.Context
	var cancel func()
	if hasDeadline {
		ctx, cancel = context.WithDeadline(parent, deadline)
	} else {
		ctx, cancel = context.WithCancel(parent)
	}
	return ctx, cancel
}

func (s *Server) Initialize(ctx context.Context, req *semanticrpc.InitializeRequest) (*semanticrpc.InitializeResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.InitializeParams{
		RootURI:          req.GetRootUri(),
		Capabilities:     req.GetCapabilities(),
		WorkspaceFolders: semanticrpc.WorkspaceFoldersFromProto(req.GetWorkspaceFolders()),
		Trace:            semanticapi.TraceValue(req.GetTrace()),
	}
	if req.GetHasProcessId() {
		pid := int(req.GetProcessId())
		params.ProcessID = &pid
	}
	result, err := s.impl.Initialize(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.InitializeResponse{
		Capabilities: semanticrpc.ServerCapabilitiesToProto(result.Capabilities),
	}, nil
}

func (s *Server) Initialized(ctx context.Context, req *semanticrpc.InitializedRequest) (*semanticrpc.InitializedResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	err := s.impl.Initialized(ctx)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.InitializedResponse{}, nil
}

func (s *Server) Shutdown(ctx context.Context, req *semanticrpc.ShutdownRequest) (*semanticrpc.ShutdownResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	err := s.impl.Shutdown(ctx)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.ShutdownResponse{}, nil
}

func (s *Server) Exit(ctx context.Context, req *semanticrpc.ExitRequest) (*semanticrpc.ExitResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	err := s.impl.Exit(ctx)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.ExitResponse{}, nil
}

func (s *Server) DidOpen(ctx context.Context, req *semanticrpc.DidOpenRequest) (*semanticrpc.DidOpenResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.DidOpenTextDocumentParams{
		TextDocument: semanticrpc.TextDocumentItemFromProto(req.GetTextDocument()),
	}
	err := s.impl.DidOpen(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.DidOpenResponse{}, nil
}

func (s *Server) DidChange(ctx context.Context, req *semanticrpc.DidChangeRequest) (*semanticrpc.DidChangeResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.DidChangeTextDocumentParams{
		TextDocument:   semanticrpc.VersionedTextDocumentIdentifierFromProto(req.GetTextDocument()),
		ContentChanges: semanticrpc.ContentChangesFromProto(req.GetContentChanges()),
	}
	err := s.impl.DidChange(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.DidChangeResponse{}, nil
}

func (s *Server) DidClose(ctx context.Context, req *semanticrpc.DidCloseRequest) (*semanticrpc.DidCloseResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.DidCloseTextDocumentParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
	}
	err := s.impl.DidClose(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.DidCloseResponse{}, nil
}

func (s *Server) DidSave(ctx context.Context, req *semanticrpc.DidSaveRequest) (*semanticrpc.DidSaveResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.DidSaveTextDocumentParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Text:         req.GetText(),
	}
	err := s.impl.DidSave(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.DidSaveResponse{}, nil
}

func (s *Server) Completion(ctx context.Context, req *semanticrpc.CompletionRequest) (*semanticrpc.CompletionResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.CompletionParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Position:     semanticrpc.PositionFromProto(req.GetPosition()),
	}
	if req.GetContextTriggerKind() != 0 {
		params.Context = &semanticapi.CompletionContext{
			TriggerKind:      semanticapi.CompletionTriggerKind(req.GetContextTriggerKind()),
			TriggerCharacter: req.GetContextTriggerCharacter(),
		}
	}
	result, err := s.impl.Completion(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.CompletionResponse{Result: semanticrpc.CompletionResultToProto(result)}, nil
}

func (s *Server) Hover(ctx context.Context, req *semanticrpc.HoverRequest) (*semanticrpc.HoverResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.HoverParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Position:     semanticrpc.PositionFromProto(req.GetPosition()),
	}
	result, err := s.impl.Hover(ctx, params)
	if err != nil {
		return nil, err
	}
	h, hasResult := semanticrpc.HoverToProto(result)
	return &semanticrpc.HoverResponse{Result: h, HasResult: hasResult}, nil
}

func (s *Server) SignatureHelp(ctx context.Context, req *semanticrpc.SignatureHelpRequest) (*semanticrpc.SignatureHelpResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.SignatureHelpParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Position:     semanticrpc.PositionFromProto(req.GetPosition()),
	}
	result, err := s.impl.SignatureHelp(ctx, params)
	if err != nil {
		return nil, err
	}
	sh, hasResult := semanticrpc.SignatureHelpToProto(result)
	return &semanticrpc.SignatureHelpResponse{Result: sh, HasResult: hasResult}, nil
}

func (s *Server) Definition(ctx context.Context, req *semanticrpc.DefinitionRequest) (*semanticrpc.DefinitionResponse, error) {
	params := semanticapi.DefinitionParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Position:     semanticrpc.PositionFromProto(req.GetPosition()),
	}
	result, err := s.impl.Definition(ctx, params)
	if err != nil {
		return nil, err
	}
	loc, locs, links := semanticrpc.LocationResultToProto(result)
	return &semanticrpc.DefinitionResponse{
		Location:      loc,
		Locations:     locs,
		LocationLinks: links,
	}, nil
}

func (s *Server) Declaration(ctx context.Context, req *semanticrpc.DeclarationRequest) (*semanticrpc.DeclarationResponse, error) {
	params := semanticapi.DeclarationParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Position:     semanticrpc.PositionFromProto(req.GetPosition()),
	}
	result, err := s.impl.Declaration(ctx, params)
	if err != nil {
		return nil, err
	}
	loc, locs, links := semanticrpc.LocationResultToProto(result)
	return &semanticrpc.DeclarationResponse{
		Location:      loc,
		Locations:     locs,
		LocationLinks: links,
	}, nil
}

func (s *Server) TypeDefinition(ctx context.Context, req *semanticrpc.TypeDefinitionRequest) (*semanticrpc.TypeDefinitionResponse, error) {
	params := semanticapi.TypeDefinitionParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Position:     semanticrpc.PositionFromProto(req.GetPosition()),
	}
	result, err := s.impl.TypeDefinition(ctx, params)
	if err != nil {
		return nil, err
	}
	loc, locs, links := semanticrpc.LocationResultToProto(result)
	return &semanticrpc.TypeDefinitionResponse{
		Location:      loc,
		Locations:     locs,
		LocationLinks: links,
	}, nil
}

func (s *Server) Implementation(ctx context.Context, req *semanticrpc.ImplementationRequest) (*semanticrpc.ImplementationResponse, error) {
	params := semanticapi.ImplementationParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Position:     semanticrpc.PositionFromProto(req.GetPosition()),
	}
	result, err := s.impl.Implementation(ctx, params)
	if err != nil {
		return nil, err
	}
	loc, locs, links := semanticrpc.LocationResultToProto(result)
	return &semanticrpc.ImplementationResponse{
		Location:      loc,
		Locations:     locs,
		LocationLinks: links,
	}, nil
}

func (s *Server) References(ctx context.Context, req *semanticrpc.ReferencesRequest) (*semanticrpc.ReferencesResponse, error) {
	params := semanticapi.ReferenceParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Position:     semanticrpc.PositionFromProto(req.GetPosition()),
		Context: semanticapi.ReferenceContext{
			IncludeDeclaration: req.GetIncludeDeclaration(),
		},
	}
	result, err := s.impl.References(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.ReferencesResponse{Locations: semanticrpc.LocationsToProto(result)}, nil
}

func (s *Server) DocumentHighlight(ctx context.Context, req *semanticrpc.DocumentHighlightRequest) (*semanticrpc.DocumentHighlightResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.DocumentHighlightParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Position:     semanticrpc.PositionFromProto(req.GetPosition()),
	}
	result, err := s.impl.DocumentHighlight(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.DocumentHighlightResponse{Highlights: semanticrpc.DocumentHighlightsToProto(result)}, nil
}

func (s *Server) DocumentSymbol(ctx context.Context, req *semanticrpc.DocumentSymbolRequest) (*semanticrpc.DocumentSymbolResponse, error) {
	params := semanticapi.DocumentSymbolParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
	}
	result, err := s.impl.DocumentSymbol(ctx, params)
	if err != nil {
		return nil, err
	}
	syms, symInfo := semanticrpc.DocumentSymbolResultToProto(result)
	return &semanticrpc.DocumentSymbolResponse{
		Symbols:           syms,
		SymbolInformation: symInfo,
	}, nil
}

func (s *Server) CodeAction(ctx context.Context, req *semanticrpc.CodeActionRequest) (*semanticrpc.CodeActionResponse, error) {
	params := semanticapi.CodeActionParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Range:        semanticrpc.RangeFromProto(req.GetRange()),
		Context: semanticapi.CodeActionContext{
			Diagnostics: semanticrpc.DiagnosticsFromProto(req.GetDiagnostics()),
		},
	}
	result, err := s.impl.CodeAction(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.CodeActionResponse{Items: semanticrpc.CodeActionResultsToProto(result)}, nil
}

func (s *Server) CodeLens(ctx context.Context, req *semanticrpc.CodeLensRequest) (*semanticrpc.CodeLensResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.CodeLensParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
	}
	result, err := s.impl.CodeLens(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.CodeLensResponse{Lenses: semanticrpc.CodeLensesToProto(result)}, nil
}

func (s *Server) Formatting(ctx context.Context, req *semanticrpc.FormattingRequest) (*semanticrpc.FormattingResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.DocumentFormattingParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Options: semanticapi.FormattingOptions{
			TabSize:      req.GetTabSize(),
			InsertSpaces: req.GetInsertSpaces(),
		},
	}
	result, err := s.impl.Formatting(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.FormattingResponse{Edits: semanticrpc.TextEditsToProto(result)}, nil
}

func (s *Server) RangeFormatting(ctx context.Context, req *semanticrpc.RangeFormattingRequest) (*semanticrpc.RangeFormattingResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.DocumentRangeFormattingParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Range:        semanticrpc.RangeFromProto(req.GetRange()),
		Options: semanticapi.FormattingOptions{
			TabSize:      req.GetTabSize(),
			InsertSpaces: req.GetInsertSpaces(),
		},
	}
	result, err := s.impl.RangeFormatting(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.RangeFormattingResponse{Edits: semanticrpc.TextEditsToProto(result)}, nil
}

func (s *Server) Rename(ctx context.Context, req *semanticrpc.RenameRequest) (*semanticrpc.RenameResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.RenameParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Position:     semanticrpc.PositionFromProto(req.GetPosition()),
		NewName:      req.GetNewName(),
	}
	result, err := s.impl.Rename(ctx, params)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return &semanticrpc.RenameResponse{HasResult: false}, nil
	}
	return &semanticrpc.RenameResponse{Result: semanticrpc.WorkspaceEditToProto(result), HasResult: true}, nil
}

func (s *Server) PrepareRename(ctx context.Context, req *semanticrpc.PrepareRenameRequest) (*semanticrpc.PrepareRenameResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.PrepareRenameParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Position:     semanticrpc.PositionFromProto(req.GetPosition()),
	}
	result, err := s.impl.PrepareRename(ctx, params)
	if err != nil {
		return nil, err
	}
	r, hasResult := semanticrpc.PrepareRenameResultToProto(result)
	return &semanticrpc.PrepareRenameResponse{Result: r, HasResult: hasResult}, nil
}

func (s *Server) FoldingRange(ctx context.Context, req *semanticrpc.FoldingRangeRequest) (*semanticrpc.FoldingRangeResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.FoldingRangeParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
	}
	result, err := s.impl.FoldingRange(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.FoldingRangeResponse{Ranges: semanticrpc.FoldingRangesToProto(result)}, nil
}

func (s *Server) SelectionRange(ctx context.Context, req *semanticrpc.SelectionRangeRequest) (*semanticrpc.SelectionRangeResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	positions := make([]semanticapi.Position, len(req.GetPositions()))
	for i, p := range req.GetPositions() {
		positions[i] = semanticrpc.PositionFromProto(p)
	}
	params := semanticapi.SelectionRangeParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Positions:    positions,
	}
	result, err := s.impl.SelectionRange(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.SelectionRangeResponse{Ranges: semanticrpc.SelectionRangesToProto(result)}, nil
}

func (s *Server) SemanticTokensFull(ctx context.Context, req *semanticrpc.SemanticTokensFullRequest) (*semanticrpc.SemanticTokensFullResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.SemanticTokensParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
	}
	result, err := s.impl.SemanticTokensFull(ctx, params)
	if err != nil {
		return nil, err
	}
	t, hasResult := semanticrpc.SemanticTokensToProto(result)
	return &semanticrpc.SemanticTokensFullResponse{Result: t, HasResult: hasResult}, nil
}

func (s *Server) SemanticTokensRange(ctx context.Context, req *semanticrpc.SemanticTokensRangeRequest) (*semanticrpc.SemanticTokensRangeResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.SemanticTokensRangeParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Range:        semanticrpc.RangeFromProto(req.GetRange()),
	}
	result, err := s.impl.SemanticTokensRange(ctx, params)
	if err != nil {
		return nil, err
	}
	t, hasResult := semanticrpc.SemanticTokensToProto(result)
	return &semanticrpc.SemanticTokensRangeResponse{Result: t, HasResult: hasResult}, nil
}

func (s *Server) Diagnostic(ctx context.Context, req *semanticrpc.DiagnosticRequest) (*semanticrpc.DiagnosticResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.DocumentDiagnosticParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
	}
	result, err := s.impl.Diagnostic(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.DiagnosticResponse{Report: semanticrpc.DocumentDiagnosticReportToProto(result)}, nil
}

func (s *Server) WorkspaceSymbol(ctx context.Context, req *semanticrpc.WorkspaceSymbolRequest) (*semanticrpc.WorkspaceSymbolResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.WorkspaceSymbolParams{Query: req.GetQuery()}
	result, err := s.impl.WorkspaceSymbol(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.WorkspaceSymbolResponse{Symbols: semanticrpc.SymbolInformationsToProto(result)}, nil
}

func (s *Server) ExecuteCommand(ctx context.Context, req *semanticrpc.ExecuteCommandRequest) (*semanticrpc.ExecuteCommandResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	args := make([]json.RawMessage, len(req.GetArguments()))
	for i, a := range req.GetArguments() {
		args[i] = json.RawMessage(a)
	}
	params := semanticapi.ExecuteCommandParams{
		Command:   req.GetCommand(),
		Arguments: args,
	}
	result, err := s.impl.ExecuteCommand(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.ExecuteCommandResponse{Result: result}, nil
}

func (s *Server) PrepareCallHierarchy(ctx context.Context, req *semanticrpc.PrepareCallHierarchyRequest) (*semanticrpc.PrepareCallHierarchyResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.CallHierarchyPrepareParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Position:     semanticrpc.PositionFromProto(req.GetPosition()),
	}
	result, err := s.impl.PrepareCallHierarchy(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.PrepareCallHierarchyResponse{Items: semanticrpc.CallHierarchyItemsToProto(result)}, nil
}

func (s *Server) CallHierarchyIncomingCalls(ctx context.Context, req *semanticrpc.CallHierarchyIncomingCallsRequest) (*semanticrpc.CallHierarchyIncomingCallsResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.CallHierarchyIncomingCallsParams{
		Item: semanticrpc.CallHierarchyItemFromProto(req.GetItem()),
	}
	result, err := s.impl.CallHierarchyIncomingCalls(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.CallHierarchyIncomingCallsResponse{Calls: semanticrpc.CallHierarchyIncomingCallsToProto(result)}, nil
}

func (s *Server) CallHierarchyOutgoingCalls(ctx context.Context, req *semanticrpc.CallHierarchyOutgoingCallsRequest) (*semanticrpc.CallHierarchyOutgoingCallsResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.CallHierarchyOutgoingCallsParams{
		Item: semanticrpc.CallHierarchyItemFromProto(req.GetItem()),
	}
	result, err := s.impl.CallHierarchyOutgoingCalls(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.CallHierarchyOutgoingCallsResponse{Calls: semanticrpc.CallHierarchyOutgoingCallsToProto(result)}, nil
}

func (s *Server) CompletionResolve(ctx context.Context, req *semanticrpc.CompletionResolveRequest) (*semanticrpc.CompletionResolveResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	item := semanticrpc.CompletionItemFromProto(req.GetItem())
	result, err := s.impl.CompletionResolve(ctx, item)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.CompletionResolveResponse{Item: semanticrpc.CompletionItemToProto(result)}, nil
}

func (s *Server) CodeLensResolve(ctx context.Context, req *semanticrpc.CodeLensResolveRequest) (*semanticrpc.CodeLensResolveResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	lens := semanticrpc.CodeLensFromProto(req.GetLens())
	result, err := s.impl.CodeLensResolve(ctx, lens)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.CodeLensResolveResponse{Lens: semanticrpc.CodeLensToProto(result)}, nil
}

func (s *Server) DocumentColor(ctx context.Context, req *semanticrpc.DocumentColorRequest) (*semanticrpc.DocumentColorResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.DocumentColorParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
	}
	result, err := s.impl.DocumentColor(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.DocumentColorResponse{Colors: semanticrpc.ColorInformationsToProto(result)}, nil
}

func (s *Server) ColorPresentation(ctx context.Context, req *semanticrpc.ColorPresentationRequest) (*semanticrpc.ColorPresentationResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.ColorPresentationParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Color:        semanticrpc.ColorFromProto(req.GetColor()),
		Range:        semanticrpc.RangeFromProto(req.GetRange()),
	}
	result, err := s.impl.ColorPresentation(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.ColorPresentationResponse{Presentations: semanticrpc.ColorPresentationsToProto(result)}, nil
}

func (s *Server) DocumentLink(ctx context.Context, req *semanticrpc.DocumentLinkRequest) (*semanticrpc.DocumentLinkResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.DocumentLinkParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
	}
	result, err := s.impl.DocumentLink(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.DocumentLinkResponse{Links: semanticrpc.DocumentLinksToProto(result)}, nil
}

func (s *Server) DocumentLinkResolve(ctx context.Context, req *semanticrpc.DocumentLinkResolveRequest) (*semanticrpc.DocumentLinkResolveResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	link := semanticrpc.DocumentLinkFromProto(req.GetLink())
	result, err := s.impl.DocumentLinkResolve(ctx, link)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.DocumentLinkResolveResponse{Link: semanticrpc.DocumentLinkToProto(result)}, nil
}

func (s *Server) OnTypeFormatting(ctx context.Context, req *semanticrpc.OnTypeFormattingRequest) (*semanticrpc.OnTypeFormattingResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.DocumentOnTypeFormattingParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Position:     semanticrpc.PositionFromProto(req.GetPosition()),
		Character:    req.GetCharacter(),
		Options: semanticapi.FormattingOptions{
			TabSize:      req.GetTabSize(),
			InsertSpaces: req.GetInsertSpaces(),
		},
	}
	result, err := s.impl.OnTypeFormatting(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.OnTypeFormattingResponse{Edits: semanticrpc.TextEditsToProto(result)}, nil
}

func (s *Server) LinkedEditingRange(ctx context.Context, req *semanticrpc.LinkedEditingRangeRequest) (*semanticrpc.LinkedEditingRangeResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.LinkedEditingRangeParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Position:     semanticrpc.PositionFromProto(req.GetPosition()),
	}
	result, err := s.impl.LinkedEditingRange(ctx, params)
	if err != nil {
		return nil, err
	}
	r, hasResult := semanticrpc.LinkedEditingRangesToProto(result)
	return &semanticrpc.LinkedEditingRangeResponse{Result: r, HasResult: hasResult}, nil
}

func (s *Server) Moniker(ctx context.Context, req *semanticrpc.MonikerRequest) (*semanticrpc.MonikerResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.MonikerParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Position:     semanticrpc.PositionFromProto(req.GetPosition()),
	}
	result, err := s.impl.Moniker(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.MonikerResponse{Monikers: semanticrpc.MonikersToProto(result)}, nil
}

func (s *Server) WillSaveWaitUntil(ctx context.Context, req *semanticrpc.WillSaveWaitUntilRequest) (*semanticrpc.WillSaveWaitUntilResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.WillSaveTextDocumentParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Reason:       semanticapi.TextDocumentSaveReason(req.GetReason()),
	}
	result, err := s.impl.WillSaveWaitUntil(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.WillSaveWaitUntilResponse{Edits: semanticrpc.TextEditsToProto(result)}, nil
}

func (s *Server) SemanticTokensFullDelta(ctx context.Context, req *semanticrpc.SemanticTokensFullDeltaRequest) (*semanticrpc.SemanticTokensFullDeltaResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.SemanticTokensDeltaParams{
		TextDocument:     semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		PreviousResultID: req.GetPreviousResultId(),
	}
	result, err := s.impl.SemanticTokensFullDelta(ctx, params)
	if err != nil {
		return nil, err
	}
	r, hasResult := semanticrpc.SemanticTokensDeltaToProto(result)
	return &semanticrpc.SemanticTokensFullDeltaResponse{Result: r, HasResult: hasResult}, nil
}

func (s *Server) PrepareTypeHierarchy(ctx context.Context, req *semanticrpc.PrepareTypeHierarchyRequest) (*semanticrpc.PrepareTypeHierarchyResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.TypeHierarchyPrepareParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Position:     semanticrpc.PositionFromProto(req.GetPosition()),
	}
	result, err := s.impl.PrepareTypeHierarchy(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.PrepareTypeHierarchyResponse{Items: semanticrpc.TypeHierarchyItemsToProto(result)}, nil
}

func (s *Server) TypeHierarchySupertypes(ctx context.Context, req *semanticrpc.TypeHierarchySupertypesRequest) (*semanticrpc.TypeHierarchySupertypesResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.TypeHierarchySupertypesParams{
		Item: semanticrpc.TypeHierarchyItemFromProto(req.GetItem()),
	}
	result, err := s.impl.TypeHierarchySupertypes(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.TypeHierarchySupertypesResponse{Items: semanticrpc.TypeHierarchyItemsToProto(result)}, nil
}

func (s *Server) TypeHierarchySubtypes(ctx context.Context, req *semanticrpc.TypeHierarchySubtypesRequest) (*semanticrpc.TypeHierarchySubtypesResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.TypeHierarchySubtypesParams{
		Item: semanticrpc.TypeHierarchyItemFromProto(req.GetItem()),
	}
	result, err := s.impl.TypeHierarchySubtypes(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.TypeHierarchySubtypesResponse{Items: semanticrpc.TypeHierarchyItemsToProto(result)}, nil
}

func (s *Server) InlayHint(ctx context.Context, req *semanticrpc.InlayHintRequest) (*semanticrpc.InlayHintResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.InlayHintParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Range:        semanticrpc.RangeFromProto(req.GetRange()),
	}
	result, err := s.impl.InlayHint(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.InlayHintResponse{Hints: semanticrpc.InlayHintsToProto(result)}, nil
}

func (s *Server) InlayHintResolve(ctx context.Context, req *semanticrpc.InlayHintResolveRequest) (*semanticrpc.InlayHintResolveResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	hint := semanticrpc.InlayHintFromProto(req.GetHint())
	result, err := s.impl.InlayHintResolve(ctx, hint)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.InlayHintResolveResponse{Hint: semanticrpc.InlayHintToProto(result)}, nil
}

func (s *Server) InlineValue(ctx context.Context, req *semanticrpc.InlineValueRequest) (*semanticrpc.InlineValueResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.InlineValueParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Range:        semanticrpc.RangeFromProto(req.GetRange()),
	}
	result, err := s.impl.InlineValue(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.InlineValueResponse{Values: semanticrpc.InlineValuesToProto(result)}, nil
}

func (s *Server) WillCreateFiles(ctx context.Context, req *semanticrpc.WillCreateFilesRequest) (*semanticrpc.WillCreateFilesResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.CreateFilesParams{
		Files: semanticrpc.FileCreatesFromProto(req.GetFiles()),
	}
	result, err := s.impl.WillCreateFiles(ctx, params)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return &semanticrpc.WillCreateFilesResponse{HasResult: false}, nil
	}
	return &semanticrpc.WillCreateFilesResponse{Result: semanticrpc.WorkspaceEditToProto(result), HasResult: true}, nil
}

func (s *Server) WillRenameFiles(ctx context.Context, req *semanticrpc.WillRenameFilesRequest) (*semanticrpc.WillRenameFilesResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.RenameFilesParams{
		Files: semanticrpc.FileRenamesFromProto(req.GetFiles()),
	}
	result, err := s.impl.WillRenameFiles(ctx, params)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return &semanticrpc.WillRenameFilesResponse{HasResult: false}, nil
	}
	return &semanticrpc.WillRenameFilesResponse{Result: semanticrpc.WorkspaceEditToProto(result), HasResult: true}, nil
}

func (s *Server) WillDeleteFiles(ctx context.Context, req *semanticrpc.WillDeleteFilesRequest) (*semanticrpc.WillDeleteFilesResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.DeleteFilesParams{
		Files: semanticrpc.FileDeletesFromProto(req.GetFiles()),
	}
	result, err := s.impl.WillDeleteFiles(ctx, params)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return &semanticrpc.WillDeleteFilesResponse{HasResult: false}, nil
	}
	return &semanticrpc.WillDeleteFilesResponse{Result: semanticrpc.WorkspaceEditToProto(result), HasResult: true}, nil
}

func (s *Server) WillSave(ctx context.Context, req *semanticrpc.WillSaveRequest) (*semanticrpc.WillSaveResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.WillSaveTextDocumentParams{
		TextDocument: semanticrpc.TextDocumentIdentifierFromProto(req.GetTextDocument()),
		Reason:       semanticapi.TextDocumentSaveReason(req.GetReason()),
	}
	err := s.impl.WillSave(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.WillSaveResponse{}, nil
}

func (s *Server) DidChangeConfiguration(ctx context.Context, req *semanticrpc.DidChangeConfigurationRequest) (*semanticrpc.DidChangeConfigurationResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.DidChangeConfigurationParams{
		Settings: json.RawMessage(req.GetSettings()),
	}
	err := s.impl.DidChangeConfiguration(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.DidChangeConfigurationResponse{}, nil
}

func (s *Server) DidChangeWatchedFiles(ctx context.Context, req *semanticrpc.DidChangeWatchedFilesRequest) (*semanticrpc.DidChangeWatchedFilesResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.DidChangeWatchedFilesParams{
		Changes: semanticrpc.FileEventsFromProto(req.GetChanges()),
	}
	err := s.impl.DidChangeWatchedFiles(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.DidChangeWatchedFilesResponse{}, nil
}

func (s *Server) DidChangeWorkspaceFolders(ctx context.Context, req *semanticrpc.DidChangeWorkspaceFoldersRequest) (*semanticrpc.DidChangeWorkspaceFoldersResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.DidChangeWorkspaceFoldersParams{
		Event: semanticapi.WorkspaceFoldersChangeEvent{
			Added:   semanticrpc.WorkspaceFoldersFromProto(req.GetAdded()),
			Removed: semanticrpc.WorkspaceFoldersFromProto(req.GetRemoved()),
		},
	}
	err := s.impl.DidChangeWorkspaceFolders(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.DidChangeWorkspaceFoldersResponse{}, nil
}

func (s *Server) WorkDoneProgressCancel(ctx context.Context, req *semanticrpc.WorkDoneProgressCancelRequest) (*semanticrpc.WorkDoneProgressCancelResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.WorkDoneProgressCancelParams{
		Token: req.GetToken(),
	}
	err := s.impl.WorkDoneProgressCancel(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.WorkDoneProgressCancelResponse{}, nil
}

func (s *Server) SetTrace(ctx context.Context, req *semanticrpc.SetTraceRequest) (*semanticrpc.SetTraceResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.SetTraceParams{
		Value: semanticapi.TraceValue(req.GetValue()),
	}
	err := s.impl.SetTrace(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.SetTraceResponse{}, nil
}

func (s *Server) DidCreateFiles(ctx context.Context, req *semanticrpc.DidCreateFilesRequest) (*semanticrpc.DidCreateFilesResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.CreateFilesParams{
		Files: semanticrpc.FileCreatesFromProto(req.GetFiles()),
	}
	err := s.impl.DidCreateFiles(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.DidCreateFilesResponse{}, nil
}

func (s *Server) DidRenameFiles(ctx context.Context, req *semanticrpc.DidRenameFilesRequest) (*semanticrpc.DidRenameFilesResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.RenameFilesParams{
		Files: semanticrpc.FileRenamesFromProto(req.GetFiles()),
	}
	err := s.impl.DidRenameFiles(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.DidRenameFilesResponse{}, nil
}

func (s *Server) DidDeleteFiles(ctx context.Context, req *semanticrpc.DidDeleteFilesRequest) (*semanticrpc.DidDeleteFilesResponse, error) {
	ctx, cancel := blueCtxFirst(ctx, s.ctx)
	defer cancel()
	params := semanticapi.DeleteFilesParams{
		Files: semanticrpc.FileDeletesFromProto(req.GetFiles()),
	}
	err := s.impl.DidDeleteFiles(ctx, params)
	if err != nil {
		return nil, err
	}
	return &semanticrpc.DidDeleteFilesResponse{}, nil
}

func (s *Server) Close() error {
	s.cancelCtx()
	return nil
}
