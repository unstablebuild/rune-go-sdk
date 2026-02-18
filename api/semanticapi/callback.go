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
	"context"
	"encoding/json"
)

// LSPCallback defines the interface for server→client LSP callbacks.
// This includes both notifications (no return value expected) and
// requests (return value expected from client).
type LSPCallback interface {

	// ShowMessage displays a message in the UI (window/showMessage).
	ShowMessage(ctx context.Context, params ShowMessageParams) error
	// LogMessage logs a message (window/logMessage).
	LogMessage(ctx context.Context, params LogMessageParams) error
	// PublishDiagnostics publishes diagnostics for a document
	// (textDocument/publishDiagnostics).
	PublishDiagnostics(ctx context.Context, params PublishDiagnosticsParams) error
	// Progress reports progress ($/progress).
	Progress(ctx context.Context, params ProgressParams) error
	// LogTrace logs a trace message ($/logTrace).
	LogTrace(ctx context.Context, params LogTraceParams) error

	// ShowDocument requests the client to display a document (window/showDocument).
	ShowDocument(ctx context.Context, params ShowDocumentParams) (ShowDocumentResult, error)
	// ShowMessageRequest shows a message with action items
	// (window/showMessageRequest).
	ShowMessageRequest(
		ctx context.Context, params ShowMessageRequestParams,
	) (*MessageActionItem, error)
	// WorkDoneProgressCreate requests creation of a progress token
	// (window/workDoneProgress/create).
	WorkDoneProgressCreate(ctx context.Context, params WorkDoneProgressCreateParams) error
	// ApplyEdit applies a workspace edit (workspace/applyEdit).
	ApplyEdit(
		ctx context.Context, params ApplyWorkspaceEditParams,
	) (ApplyWorkspaceEditResult, error)
	// WorkspaceFolders returns the workspace folders (workspace/workspaceFolders).
	WorkspaceFolders(ctx context.Context) ([]WorkspaceFolder, error)
	// Configuration fetches configuration from the client (workspace/configuration).
	Configuration(ctx context.Context, params ConfigurationParams) ([]json.RawMessage, error)
	// RegisterCapability dynamically registers capabilities
	// (client/registerCapability).
	RegisterCapability(ctx context.Context, params RegistrationParams) error
	// UnregisterCapability dynamically unregisters capabilities
	// (client/unregisterCapability).
	UnregisterCapability(ctx context.Context, params UnregistrationParams) error

	// CodeLensRefresh requests the client to refresh code lenses
	// (workspace/codeLens/refresh).
	CodeLensRefresh(ctx context.Context) error
	// SemanticTokensRefresh requests the client to refresh semantic tokens
	// (workspace/semanticTokens/refresh).
	SemanticTokensRefresh(ctx context.Context) error
	// InlayHintRefresh requests the client to refresh inlay hints
	// (workspace/inlayHint/refresh).
	InlayHintRefresh(ctx context.Context) error
	// DiagnosticRefresh requests the client to refresh diagnostics
	// (workspace/diagnostic/refresh).
	DiagnosticRefresh(ctx context.Context) error
}
