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
	"log/slog"

	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
	"github.com/unstablebuild/rune-go-sdk/joincontext"
	"google.golang.org/grpc"
)

var _ semanticapi.LSP = (*Client)(nil)

// Client satisfies semanticapi.LSP by calling a remote LSP server over gRPC.
type Client struct {
	cc              grpc.ClientConnInterface
	lsp             LSPClient
	log             *slog.Logger
	clientCtx       context.Context
	clientCancelCtx func()
}

// NewClient allocates storage for a new Client and initializes it.
func NewClient(ctx context.Context, cc grpc.ClientConnInterface) *Client {
	ret := new(Client)
	ret.Init(ctx, cc)
	return ret
}

// Init initializes this Client with the given client connection and parent context.
func (c *Client) Init(ctx context.Context, cc grpc.ClientConnInterface) {
	c.cc = cc
	c.lsp = NewLSPClient(cc)
	c.clientCtx, c.clientCancelCtx = context.WithCancel(ctx)
	c.log = slog.Default().With("struct", "semanticrpc.Client")
}

func (c *Client) Initialize(ctx context.Context, params semanticapi.InitializeParams) (semanticapi.InitializeResult, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &InitializeRequest{
		RootUri:          params.RootURI,
		Capabilities:     params.Capabilities,
		WorkspaceFolders: WorkspaceFoldersToProto(params.WorkspaceFolders),
		Trace:            string(params.Trace),
	}
	if params.ProcessID != nil {
		req.ProcessId = int32(*params.ProcessID)
		req.HasProcessId = true
	}
	res, err := c.lsp.Initialize(ctx, req)
	if err != nil {
		return semanticapi.InitializeResult{}, err
	}
	return semanticapi.InitializeResult{
		Capabilities: ServerCapabilitiesFromProto(res.GetCapabilities()),
	}, nil
}

func (c *Client) Initialized(ctx context.Context) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	_, err := c.lsp.Initialized(ctx, &InitializedRequest{})
	return err
}

func (c *Client) Shutdown(ctx context.Context) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	_, err := c.lsp.Shutdown(ctx, &ShutdownRequest{})
	return err
}

func (c *Client) Exit(ctx context.Context) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	_, err := c.lsp.Exit(ctx, &ExitRequest{})
	return err
}

func (c *Client) DidOpen(ctx context.Context, params semanticapi.DidOpenTextDocumentParams) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DidOpenRequest{TextDocument: TextDocumentItemToProto(params.TextDocument)}
	_, err := c.lsp.DidOpen(ctx, req)
	return err
}

func (c *Client) DidChange(ctx context.Context, params semanticapi.DidChangeTextDocumentParams) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DidChangeRequest{
		TextDocument:   VersionedTextDocumentIdentifierToProto(params.TextDocument),
		ContentChanges: ContentChangesToProto(params.ContentChanges),
	}
	_, err := c.lsp.DidChange(ctx, req)
	return err
}

func (c *Client) DidClose(ctx context.Context, params semanticapi.DidCloseTextDocumentParams) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DidCloseRequest{TextDocument: TextDocumentIdentifierToProto(params.TextDocument)}
	_, err := c.lsp.DidClose(ctx, req)
	return err
}

func (c *Client) DidSave(ctx context.Context, params semanticapi.DidSaveTextDocumentParams) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DidSaveRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Text:         params.Text,
	}
	_, err := c.lsp.DidSave(ctx, req)
	return err
}

func (c *Client) Completion(ctx context.Context, params semanticapi.CompletionParams) (semanticapi.CompletionResult, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &CompletionRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Position:     PositionToProto(params.Position),
	}
	if params.Context != nil {
		req.ContextTriggerKind = int32(params.Context.TriggerKind)
		req.ContextTriggerCharacter = params.Context.TriggerCharacter
	}
	res, err := c.lsp.Completion(ctx, req)
	if err != nil {
		return semanticapi.CompletionResult{}, err
	}
	return CompletionResultFromProto(res.GetResult()), nil
}

func (c *Client) Hover(ctx context.Context, params semanticapi.HoverParams) (*semanticapi.Hover, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &HoverRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Position:     PositionToProto(params.Position),
	}
	res, err := c.lsp.Hover(ctx, req)
	if err != nil {
		return nil, err
	}
	return HoverFromProto(res.GetResult(), res.GetHasResult()), nil
}

func (c *Client) SignatureHelp(ctx context.Context, params semanticapi.SignatureHelpParams) (*semanticapi.SignatureHelp, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &SignatureHelpRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Position:     PositionToProto(params.Position),
	}
	res, err := c.lsp.SignatureHelp(ctx, req)
	if err != nil {
		return nil, err
	}
	return SignatureHelpFromProto(res.GetResult(), res.GetHasResult()), nil
}

func (c *Client) Definition(ctx context.Context, params semanticapi.DefinitionParams) (semanticapi.LocationResult, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DefinitionRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Position:     PositionToProto(params.Position),
	}
	res, err := c.lsp.Definition(ctx, req)
	if err != nil {
		return semanticapi.LocationResult{}, err
	}
	return LocationResultFromProto(res), nil
}

func (c *Client) Declaration(ctx context.Context, params semanticapi.DeclarationParams) (semanticapi.LocationResult, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DeclarationRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Position:     PositionToProto(params.Position),
	}
	res, err := c.lsp.Declaration(ctx, req)
	if err != nil {
		return semanticapi.LocationResult{}, err
	}
	return LocationResultFromProtoDecl(res), nil
}

func (c *Client) TypeDefinition(ctx context.Context, params semanticapi.TypeDefinitionParams) (semanticapi.LocationResult, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &TypeDefinitionRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Position:     PositionToProto(params.Position),
	}
	res, err := c.lsp.TypeDefinition(ctx, req)
	if err != nil {
		return semanticapi.LocationResult{}, err
	}
	return LocationResultFromProtoTypeDef(res), nil
}

func (c *Client) Implementation(ctx context.Context, params semanticapi.ImplementationParams) (semanticapi.LocationResult, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &ImplementationRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Position:     PositionToProto(params.Position),
	}
	res, err := c.lsp.Implementation(ctx, req)
	if err != nil {
		return semanticapi.LocationResult{}, err
	}
	return LocationResultFromProtoImpl(res), nil
}

func (c *Client) References(ctx context.Context, params semanticapi.ReferenceParams) ([]semanticapi.Location, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &ReferencesRequest{
		TextDocument:       TextDocumentIdentifierToProto(params.TextDocument),
		Position:           PositionToProto(params.Position),
		IncludeDeclaration: params.Context.IncludeDeclaration,
	}
	res, err := c.lsp.References(ctx, req)
	if err != nil {
		return nil, err
	}
	return LocationsFromProto(res.GetLocations()), nil
}

func (c *Client) DocumentHighlight(ctx context.Context, params semanticapi.DocumentHighlightParams) ([]semanticapi.DocumentHighlight, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DocumentHighlightRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Position:     PositionToProto(params.Position),
	}
	res, err := c.lsp.DocumentHighlight(ctx, req)
	if err != nil {
		return nil, err
	}
	return DocumentHighlightsFromProto(res.GetHighlights()), nil
}

func (c *Client) DocumentSymbol(ctx context.Context, params semanticapi.DocumentSymbolParams) (semanticapi.DocumentSymbolResult, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DocumentSymbolRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
	}
	res, err := c.lsp.DocumentSymbol(ctx, req)
	if err != nil {
		return semanticapi.DocumentSymbolResult{}, err
	}
	return DocumentSymbolResultFromProto(res), nil
}

func (c *Client) CodeAction(ctx context.Context, params semanticapi.CodeActionParams) ([]semanticapi.CodeActionResult, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &CodeActionRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Range:        RangeToProto(params.Range),
		Diagnostics:  DiagnosticsToProto(params.Context.Diagnostics),
	}
	res, err := c.lsp.CodeAction(ctx, req)
	if err != nil {
		return nil, err
	}
	return CodeActionResultsFromProto(res), nil
}

func (c *Client) CodeLens(ctx context.Context, params semanticapi.CodeLensParams) ([]semanticapi.CodeLens, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &CodeLensRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
	}
	res, err := c.lsp.CodeLens(ctx, req)
	if err != nil {
		return nil, err
	}
	return CodeLensesFromProto(res.GetLenses()), nil
}

func (c *Client) Formatting(ctx context.Context, params semanticapi.DocumentFormattingParams) ([]semanticapi.TextEdit, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &FormattingRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		TabSize:      params.Options.TabSize,
		InsertSpaces: params.Options.InsertSpaces,
	}
	res, err := c.lsp.Formatting(ctx, req)
	if err != nil {
		return nil, err
	}
	return TextEditsFromProto(res.GetEdits()), nil
}

func (c *Client) RangeFormatting(ctx context.Context, params semanticapi.DocumentRangeFormattingParams) ([]semanticapi.TextEdit, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &RangeFormattingRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Range:        RangeToProto(params.Range),
		TabSize:      params.Options.TabSize,
		InsertSpaces: params.Options.InsertSpaces,
	}
	res, err := c.lsp.RangeFormatting(ctx, req)
	if err != nil {
		return nil, err
	}
	return TextEditsFromProto(res.GetEdits()), nil
}

func (c *Client) Rename(ctx context.Context, params semanticapi.RenameParams) (*semanticapi.WorkspaceEdit, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &RenameRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Position:     PositionToProto(params.Position),
		NewName:      params.NewName,
	}
	res, err := c.lsp.Rename(ctx, req)
	if err != nil {
		return nil, err
	}
	if !res.GetHasResult() {
		return nil, nil
	}
	return WorkspaceEditFromProto(res.GetResult()), nil
}

func (c *Client) PrepareRename(ctx context.Context, params semanticapi.PrepareRenameParams) (*semanticapi.PrepareRenameResult, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &PrepareRenameRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Position:     PositionToProto(params.Position),
	}
	res, err := c.lsp.PrepareRename(ctx, req)
	if err != nil {
		return nil, err
	}
	return PrepareRenameResultFromProto(res.GetResult(), res.GetHasResult()), nil
}

func (c *Client) FoldingRange(ctx context.Context, params semanticapi.FoldingRangeParams) ([]semanticapi.FoldingRange, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &FoldingRangeRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
	}
	res, err := c.lsp.FoldingRange(ctx, req)
	if err != nil {
		return nil, err
	}
	return FoldingRangesFromProto(res.GetRanges()), nil
}

func (c *Client) SelectionRange(ctx context.Context, params semanticapi.SelectionRangeParams) ([]semanticapi.SelectionRange, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	positions := make([]*Position, len(params.Positions))
	for i, p := range params.Positions {
		positions[i] = PositionToProto(p)
	}
	req := &SelectionRangeRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Positions:    positions,
	}
	res, err := c.lsp.SelectionRange(ctx, req)
	if err != nil {
		return nil, err
	}
	return SelectionRangesFromProto(res.GetRanges()), nil
}

func (c *Client) SemanticTokensFull(ctx context.Context, params semanticapi.SemanticTokensParams) (*semanticapi.SemanticTokens, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &SemanticTokensFullRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
	}
	res, err := c.lsp.SemanticTokensFull(ctx, req)
	if err != nil {
		return nil, err
	}
	return SemanticTokensFromProto(res.GetResult(), res.GetHasResult()), nil
}

func (c *Client) SemanticTokensRange(ctx context.Context, params semanticapi.SemanticTokensRangeParams) (*semanticapi.SemanticTokens, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &SemanticTokensRangeRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Range:        RangeToProto(params.Range),
	}
	res, err := c.lsp.SemanticTokensRange(ctx, req)
	if err != nil {
		return nil, err
	}
	return SemanticTokensFromProto(res.GetResult(), res.GetHasResult()), nil
}

func (c *Client) Diagnostic(ctx context.Context, params semanticapi.DocumentDiagnosticParams) (semanticapi.DocumentDiagnosticReport, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DiagnosticRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
	}
	res, err := c.lsp.Diagnostic(ctx, req)
	if err != nil {
		return semanticapi.DocumentDiagnosticReport{}, err
	}
	return DocumentDiagnosticReportFromProto(res.GetReport()), nil
}

func (c *Client) WorkspaceDiagnostic(
	ctx context.Context,
	params semanticapi.WorkspaceDiagnosticParams,
) (semanticapi.WorkspaceDiagnosticReport, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &WorkspaceDiagnosticRequest{
		Identifier:        params.Identifier,
		PreviousResultIds: PreviousResultIDsToProto(params.PreviousResultIDs),
	}
	res, err := c.lsp.WorkspaceDiagnostic(ctx, req)
	if err != nil {
		return semanticapi.WorkspaceDiagnosticReport{}, err
	}
	return semanticapi.WorkspaceDiagnosticReport{
		Items: WorkspaceDocumentDiagnosticReportsFromProto(res.GetItems()),
	}, nil
}

func (c *Client) WorkspaceSymbol(ctx context.Context, params semanticapi.WorkspaceSymbolParams) ([]semanticapi.SymbolInformation, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &WorkspaceSymbolRequest{Query: params.Query}
	res, err := c.lsp.WorkspaceSymbol(ctx, req)
	if err != nil {
		return nil, err
	}
	return SymbolInformationsFromProto(res.GetSymbols()), nil
}

func (c *Client) ExecuteCommand(ctx context.Context, params semanticapi.ExecuteCommandParams) (string, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	args := make([]string, len(params.Arguments))
	for i, a := range params.Arguments {
		args[i] = string(a)
	}
	req := &ExecuteCommandRequest{
		Command:   params.Command,
		Arguments: args,
	}
	res, err := c.lsp.ExecuteCommand(ctx, req)
	if err != nil {
		return "", err
	}
	return res.GetResult(), nil
}

func (c *Client) PrepareCallHierarchy(ctx context.Context, params semanticapi.CallHierarchyPrepareParams) ([]semanticapi.CallHierarchyItem, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &PrepareCallHierarchyRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Position:     PositionToProto(params.Position),
	}
	res, err := c.lsp.PrepareCallHierarchy(ctx, req)
	if err != nil {
		return nil, err
	}
	return CallHierarchyItemsFromProto(res.GetItems()), nil
}

func (c *Client) CallHierarchyIncomingCalls(ctx context.Context, params semanticapi.CallHierarchyIncomingCallsParams) ([]semanticapi.CallHierarchyIncomingCall, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &CallHierarchyIncomingCallsRequest{
		Item: CallHierarchyItemToProto(params.Item),
	}
	res, err := c.lsp.CallHierarchyIncomingCalls(ctx, req)
	if err != nil {
		return nil, err
	}
	return CallHierarchyIncomingCallsFromProto(res.GetCalls()), nil
}

func (c *Client) CallHierarchyOutgoingCalls(ctx context.Context, params semanticapi.CallHierarchyOutgoingCallsParams) ([]semanticapi.CallHierarchyOutgoingCall, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &CallHierarchyOutgoingCallsRequest{
		Item: CallHierarchyItemToProto(params.Item),
	}
	res, err := c.lsp.CallHierarchyOutgoingCalls(ctx, req)
	if err != nil {
		return nil, err
	}
	return CallHierarchyOutgoingCallsFromProto(res.GetCalls()), nil
}

func (c *Client) CompletionResolve(ctx context.Context, item semanticapi.CompletionItem) (semanticapi.CompletionItem, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &CompletionResolveRequest{Item: CompletionItemToProto(item)}
	res, err := c.lsp.CompletionResolve(ctx, req)
	if err != nil {
		return semanticapi.CompletionItem{}, err
	}
	return CompletionItemFromProto(res.GetItem()), nil
}

func (c *Client) CodeLensResolve(ctx context.Context, lens semanticapi.CodeLens) (semanticapi.CodeLens, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &CodeLensResolveRequest{Lens: CodeLensToProto(lens)}
	res, err := c.lsp.CodeLensResolve(ctx, req)
	if err != nil {
		return semanticapi.CodeLens{}, err
	}
	return CodeLensFromProto(res.GetLens()), nil
}

func (c *Client) DocumentColor(ctx context.Context, params semanticapi.DocumentColorParams) ([]semanticapi.ColorInformation, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DocumentColorRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
	}
	res, err := c.lsp.DocumentColor(ctx, req)
	if err != nil {
		return nil, err
	}
	return ColorInformationsFromProto(res.GetColors()), nil
}

func (c *Client) ColorPresentation(ctx context.Context, params semanticapi.ColorPresentationParams) ([]semanticapi.ColorPresentation, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &ColorPresentationRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Color:        ColorToProto(params.Color),
		Range:        RangeToProto(params.Range),
	}
	res, err := c.lsp.ColorPresentation(ctx, req)
	if err != nil {
		return nil, err
	}
	return ColorPresentationsFromProto(res.GetPresentations()), nil
}

func (c *Client) DocumentLink(ctx context.Context, params semanticapi.DocumentLinkParams) ([]semanticapi.DocumentLink, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DocumentLinkRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
	}
	res, err := c.lsp.DocumentLink(ctx, req)
	if err != nil {
		return nil, err
	}
	return DocumentLinksFromProto(res.GetLinks()), nil
}

func (c *Client) DocumentLinkResolve(ctx context.Context, link semanticapi.DocumentLink) (semanticapi.DocumentLink, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DocumentLinkResolveRequest{Link: DocumentLinkToProto(link)}
	res, err := c.lsp.DocumentLinkResolve(ctx, req)
	if err != nil {
		return semanticapi.DocumentLink{}, err
	}
	return DocumentLinkFromProto(res.GetLink()), nil
}

func (c *Client) OnTypeFormatting(ctx context.Context, params semanticapi.DocumentOnTypeFormattingParams) ([]semanticapi.TextEdit, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &OnTypeFormattingRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Position:     PositionToProto(params.Position),
		Character:    params.Character,
		TabSize:      params.Options.TabSize,
		InsertSpaces: params.Options.InsertSpaces,
	}
	res, err := c.lsp.OnTypeFormatting(ctx, req)
	if err != nil {
		return nil, err
	}
	return TextEditsFromProto(res.GetEdits()), nil
}

func (c *Client) LinkedEditingRange(ctx context.Context, params semanticapi.LinkedEditingRangeParams) (*semanticapi.LinkedEditingRanges, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &LinkedEditingRangeRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Position:     PositionToProto(params.Position),
	}
	res, err := c.lsp.LinkedEditingRange(ctx, req)
	if err != nil {
		return nil, err
	}
	return LinkedEditingRangesFromProto(res.GetResult(), res.GetHasResult()), nil
}

func (c *Client) Moniker(ctx context.Context, params semanticapi.MonikerParams) ([]semanticapi.Moniker, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &MonikerRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Position:     PositionToProto(params.Position),
	}
	res, err := c.lsp.Moniker(ctx, req)
	if err != nil {
		return nil, err
	}
	return MonikersFromProto(res.GetMonikers()), nil
}

func (c *Client) WillSaveWaitUntil(ctx context.Context, params semanticapi.WillSaveTextDocumentParams) ([]semanticapi.TextEdit, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &WillSaveWaitUntilRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Reason:       int32(params.Reason),
	}
	res, err := c.lsp.WillSaveWaitUntil(ctx, req)
	if err != nil {
		return nil, err
	}
	return TextEditsFromProto(res.GetEdits()), nil
}

func (c *Client) SemanticTokensFullDelta(ctx context.Context, params semanticapi.SemanticTokensDeltaParams) (*semanticapi.SemanticTokensDelta, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &SemanticTokensFullDeltaRequest{
		TextDocument:     TextDocumentIdentifierToProto(params.TextDocument),
		PreviousResultId: params.PreviousResultID,
	}
	res, err := c.lsp.SemanticTokensFullDelta(ctx, req)
	if err != nil {
		return nil, err
	}
	return SemanticTokensDeltaFromProto(res.GetResult(), res.GetHasResult()), nil
}

func (c *Client) PrepareTypeHierarchy(ctx context.Context, params semanticapi.TypeHierarchyPrepareParams) ([]semanticapi.TypeHierarchyItem, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &PrepareTypeHierarchyRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Position:     PositionToProto(params.Position),
	}
	res, err := c.lsp.PrepareTypeHierarchy(ctx, req)
	if err != nil {
		return nil, err
	}
	return TypeHierarchyItemsFromProto(res.GetItems()), nil
}

func (c *Client) TypeHierarchySupertypes(ctx context.Context, params semanticapi.TypeHierarchySupertypesParams) ([]semanticapi.TypeHierarchyItem, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &TypeHierarchySupertypesRequest{
		Item: TypeHierarchyItemToProto(params.Item),
	}
	res, err := c.lsp.TypeHierarchySupertypes(ctx, req)
	if err != nil {
		return nil, err
	}
	return TypeHierarchyItemsFromProto(res.GetItems()), nil
}

func (c *Client) TypeHierarchySubtypes(ctx context.Context, params semanticapi.TypeHierarchySubtypesParams) ([]semanticapi.TypeHierarchyItem, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &TypeHierarchySubtypesRequest{
		Item: TypeHierarchyItemToProto(params.Item),
	}
	res, err := c.lsp.TypeHierarchySubtypes(ctx, req)
	if err != nil {
		return nil, err
	}
	return TypeHierarchyItemsFromProto(res.GetItems()), nil
}

func (c *Client) InlayHint(ctx context.Context, params semanticapi.InlayHintParams) ([]semanticapi.InlayHint, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &InlayHintRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Range:        RangeToProto(params.Range),
	}
	res, err := c.lsp.InlayHint(ctx, req)
	if err != nil {
		return nil, err
	}
	return InlayHintsFromProto(res.GetHints()), nil
}

func (c *Client) InlayHintResolve(ctx context.Context, hint semanticapi.InlayHint) (semanticapi.InlayHint, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &InlayHintResolveRequest{Hint: InlayHintToProto(hint)}
	res, err := c.lsp.InlayHintResolve(ctx, req)
	if err != nil {
		return semanticapi.InlayHint{}, err
	}
	return InlayHintFromProto(res.GetHint()), nil
}

func (c *Client) InlineValue(ctx context.Context, params semanticapi.InlineValueParams) ([]semanticapi.InlineValue, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &InlineValueRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Range:        RangeToProto(params.Range),
	}
	res, err := c.lsp.InlineValue(ctx, req)
	if err != nil {
		return nil, err
	}
	return InlineValuesFromProto(res.GetValues()), nil
}

func (c *Client) WillCreateFiles(ctx context.Context, params semanticapi.CreateFilesParams) (*semanticapi.WorkspaceEdit, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &WillCreateFilesRequest{Files: FileCreatesToProto(params.Files)}
	res, err := c.lsp.WillCreateFiles(ctx, req)
	if err != nil {
		return nil, err
	}
	if !res.GetHasResult() {
		return nil, nil
	}
	return WorkspaceEditFromProto(res.GetResult()), nil
}

func (c *Client) WillRenameFiles(ctx context.Context, params semanticapi.RenameFilesParams) (*semanticapi.WorkspaceEdit, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &WillRenameFilesRequest{Files: FileRenamesToProto(params.Files)}
	res, err := c.lsp.WillRenameFiles(ctx, req)
	if err != nil {
		return nil, err
	}
	if !res.GetHasResult() {
		return nil, nil
	}
	return WorkspaceEditFromProto(res.GetResult()), nil
}

func (c *Client) WillDeleteFiles(ctx context.Context, params semanticapi.DeleteFilesParams) (*semanticapi.WorkspaceEdit, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &WillDeleteFilesRequest{Files: FileDeletesToProto(params.Files)}
	res, err := c.lsp.WillDeleteFiles(ctx, req)
	if err != nil {
		return nil, err
	}
	if !res.GetHasResult() {
		return nil, nil
	}
	return WorkspaceEditFromProto(res.GetResult()), nil
}

func (c *Client) WillSave(ctx context.Context, params semanticapi.WillSaveTextDocumentParams) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &WillSaveRequest{
		TextDocument: TextDocumentIdentifierToProto(params.TextDocument),
		Reason:       int32(params.Reason),
	}
	_, err := c.lsp.WillSave(ctx, req)
	return err
}

func (c *Client) DidChangeConfiguration(ctx context.Context, params semanticapi.DidChangeConfigurationParams) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DidChangeConfigurationRequest{Settings: []byte(params.Settings)}
	_, err := c.lsp.DidChangeConfiguration(ctx, req)
	return err
}

func (c *Client) DidChangeWatchedFiles(ctx context.Context, params semanticapi.DidChangeWatchedFilesParams) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DidChangeWatchedFilesRequest{Changes: FileEventsToProto(params.Changes)}
	_, err := c.lsp.DidChangeWatchedFiles(ctx, req)
	return err
}

func (c *Client) DidChangeWorkspaceFolders(ctx context.Context, params semanticapi.DidChangeWorkspaceFoldersParams) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DidChangeWorkspaceFoldersRequest{
		Added:   WorkspaceFoldersToProto(params.Event.Added),
		Removed: WorkspaceFoldersToProto(params.Event.Removed),
	}
	_, err := c.lsp.DidChangeWorkspaceFolders(ctx, req)
	return err
}

func (c *Client) WorkDoneProgressCancel(ctx context.Context, params semanticapi.WorkDoneProgressCancelParams) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &WorkDoneProgressCancelRequest{Token: params.Token}
	_, err := c.lsp.WorkDoneProgressCancel(ctx, req)
	return err
}

func (c *Client) SetTrace(ctx context.Context, params semanticapi.SetTraceParams) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &SetTraceRequest{Value: string(params.Value)}
	_, err := c.lsp.SetTrace(ctx, req)
	return err
}

func (c *Client) DidCreateFiles(ctx context.Context, params semanticapi.CreateFilesParams) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DidCreateFilesRequest{Files: FileCreatesToProto(params.Files)}
	_, err := c.lsp.DidCreateFiles(ctx, req)
	return err
}

func (c *Client) DidRenameFiles(ctx context.Context, params semanticapi.RenameFilesParams) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DidRenameFilesRequest{Files: FileRenamesToProto(params.Files)}
	_, err := c.lsp.DidRenameFiles(ctx, req)
	return err
}

func (c *Client) DidDeleteFiles(ctx context.Context, params semanticapi.DeleteFilesParams) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DidDeleteFilesRequest{Files: FileDeletesToProto(params.Files)}
	_, err := c.lsp.DidDeleteFiles(ctx, req)
	return err
}
