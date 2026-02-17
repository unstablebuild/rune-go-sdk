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
	"log/slog"

	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
	"unstable.build/go-tui/ide/idelsp/jsonrpc2"
)

// callbackAdapter adapts semanticapi.LSPCallback to jsonrpc2.Handler.
type callbackAdapter struct {
	cb semanticapi.LSPCallback
}

// Compile-time assertion that callbackAdapter implements jsonrpc2.Handler.
var _ jsonrpc2.Handler = (*callbackAdapter)(nil)

func newCallbackAdapter(cb semanticapi.LSPCallback) jsonrpc2.Handler {
	if cb == nil {
		return jsonrpc2.HandlerFunc(func(
			ctx context.Context, req *jsonrpc2.Request,
		) (any, error) {
			return nil, jsonrpc2.ErrNotHandled
		})
	}
	return &callbackAdapter{cb: cb}
}

// Handle implements jsonrpc2.Handler, processing server-initiated messages.
func (a *callbackAdapter) Handle(
	ctx context.Context, req *jsonrpc2.Request,
) (any, error) {
	// Notifications have no ID; handle and return nil.
	if !req.IsCall() {
		a.handleNotification(ctx, req.Method, req.Params)
		return nil, nil
	}
	// Requests have an ID; handle and return result.
	return a.handleRequest(ctx, req.Method, req.Params)
}

// handleNotification handles server-initiated notifications.
func (a *callbackAdapter) handleNotification(
	ctx context.Context, method string, params json.RawMessage,
) {
	var err error
	switch method {
	case "window/showMessage":
		var p semanticapi.ShowMessageParams
		if json.Unmarshal(params, &p) == nil {
			err = a.cb.ShowMessage(ctx, p)
		}
	case "window/logMessage":
		var p semanticapi.LogMessageParams
		if json.Unmarshal(params, &p) == nil {
			err = a.cb.LogMessage(ctx, p)
		}
	case "textDocument/publishDiagnostics":
		var p semanticapi.PublishDiagnosticsParams
		if json.Unmarshal(params, &p) == nil {
			err = a.cb.PublishDiagnostics(ctx, p)
		}
	case "$/progress":
		var p semanticapi.ProgressParams
		if json.Unmarshal(params, &p) == nil {
			err = a.cb.Progress(ctx, p)
		}
	case "$/logTrace":
		var p semanticapi.LogTraceParams
		if json.Unmarshal(params, &p) == nil {
			err = a.cb.LogTrace(ctx, p)
		}
	default:
		slog.Debug("idelsp: unknown notification", "method", method)
	}
	if err != nil {
		slog.Warn("idelsp: callback error",
			"method", method, "error", err)
	}
}

// handleRequest handles server-initiated requests.
func (a *callbackAdapter) handleRequest(
	ctx context.Context, method string, params json.RawMessage,
) (any, error) {
	switch method {
	case "window/showDocument":
		var p semanticapi.ShowDocumentParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		return a.cb.ShowDocument(ctx, p)

	case "window/showMessageRequest":
		var p semanticapi.ShowMessageRequestParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		return a.cb.ShowMessageRequest(ctx, p)

	case "window/workDoneProgress/create":
		var p semanticapi.WorkDoneProgressCreateParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		return nil, a.cb.WorkDoneProgressCreate(ctx, p)

	case "workspace/applyEdit":
		var p semanticapi.ApplyWorkspaceEditParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		return a.cb.ApplyEdit(ctx, p)

	case "workspace/workspaceFolders":
		return a.cb.WorkspaceFolders(ctx)

	case "workspace/configuration":
		var p semanticapi.ConfigurationParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		return a.cb.Configuration(ctx, p)

	case "client/registerCapability":
		var p semanticapi.RegistrationParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		return nil, a.cb.RegisterCapability(ctx, p)

	case "client/unregisterCapability":
		var p semanticapi.UnregistrationParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		return nil, a.cb.UnregisterCapability(ctx, p)

	case "workspace/codeLens/refresh":
		return nil, a.cb.CodeLensRefresh(ctx)

	case "workspace/semanticTokens/refresh":
		return nil, a.cb.SemanticTokensRefresh(ctx)

	case "workspace/inlayHint/refresh":
		return nil, a.cb.InlayHintRefresh(ctx)

	case "workspace/diagnostic/refresh":
		return nil, a.cb.DiagnosticRefresh(ctx)

	default:
		return nil, fmt.Errorf("unknown request method: %s", method)
	}
}
