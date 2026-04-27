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

package debugrpc

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/go-dap"
	"github.com/unstablebuild/rune-go-sdk/api/debugapi"
	"github.com/unstablebuild/rune-go-sdk/joincontext"
	"google.golang.org/grpc"
)

var _ debugapi.Debugger = (*Client)(nil)

// Client implements debugapi.Debugger by making gRPC calls to a DebugService server.
type Client struct {
	cc              grpc.ClientConnInterface
	client          DebuggerClient
	clientCtx       context.Context
	clientCancelCtx func()
}

// NewClient creates a new Client connected to the given gRPC connection.
func NewClient(ctx context.Context, cc grpc.ClientConnInterface) *Client {
	clientCtx, cancel := context.WithCancel(ctx)
	return &Client{
		cc:              cc,
		client:          NewDebuggerClient(cc),
		clientCtx:       clientCtx,
		clientCancelCtx: cancel,
	}
}

// Close closes the client and cancels any ongoing operations.
func (c *Client) Close() error {
	if c.clientCancelCtx != nil {
		c.clientCancelCtx()
	}
	return nil
}

// CreateSession implements debugapi.Debugger. It opens a
// server-streaming RPC, reads the initial SessionOpened message
// (blocking until it arrives), then spawns a goroutine that
// drains the stream and forwards events to subscriber. The
// goroutine exits when the stream is closed by the server or
// when the client context is cancelled.
func (c *Client) CreateSession(
	ctx context.Context, langID string,
	client debugapi.ClientCapabilities,
	subscriber debugapi.EventSubscriber,
) (string, *dap.Capabilities, error) {
	// The stream must outlive ctx (which typically belongs to
	// the caller's single request), so we bind it to
	// c.clientCtx only. The initial handshake respects ctx via
	// a short-lived joined context.
	stream, err := c.client.CreateSession(c.clientCtx, &CreateSessionRequest{
		LangId: langID,
		Client: clientCapabilitiesToProto(client),
	})
	if err != nil {
		return "", nil, fmt.Errorf("create session: %w", err)
	}

	// Wait for the first SessionOpened message, respecting the
	// caller's ctx so slow adapter startup can be aborted.
	opened, err := recvSessionOpened(ctx, stream)
	if err != nil {
		return "", nil, err
	}

	go drainSessionEvents(stream, subscriber)

	return opened.GetSessionId(),
		capabilitiesFromProto(opened.GetCapabilities()), nil
}

// recvSessionOpened receives messages until the first
// SessionOpened arrives, or ctx/stream terminates.
func recvSessionOpened(
	ctx context.Context, stream grpc.ServerStreamingClient[SessionEvent],
) (*SessionOpened, error) {
	type result struct {
		msg *SessionEvent
		err error
	}
	ch := make(chan result, 1)
	go func() {
		msg, err := stream.Recv()
		ch <- result{msg, err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-ch:
		if r.err != nil {
			return nil, fmt.Errorf("recv session opened: %w", r.err)
		}
		opened, ok := r.msg.GetPayload().(*SessionEvent_Opened)
		if !ok || opened == nil {
			return nil, errors.New(
				"first session event must be SessionOpened")
		}
		return opened.Opened, nil
	}
}

// drainSessionEvents reads the remainder of the session stream,
// forwarding DAP events and terminating on SessionClosed or
// stream EOF/error.
func drainSessionEvents(
	stream grpc.ServerStreamingClient[SessionEvent],
	subscriber debugapi.EventSubscriber,
) {
	reason := "canceled"
	defer func() {
		subscriber.OnClose(reason)
	}()
	for {
		msg, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) && reason == "canceled" {
				reason = "terminated"
			}
			return
		}
		switch p := msg.GetPayload().(type) {
		case *SessionEvent_DapEvent:
			if ev := dapEventFromProto(p.DapEvent); ev != nil {
				subscriber.OnEvent(ev)
			}
		case *SessionEvent_Closed:
			if r := p.Closed.GetReason(); r != "" {
				reason = r
			} else {
				reason = "terminated"
			}
			return
		case *SessionEvent_Opened:
			// Protocol error: Opened must be first. Ignore.
		}
	}
}

// Launch implements debugapi.Debugger.
func (c *Client) Launch(ctx context.Context, sessionID string, args debugapi.LaunchRequestArguments) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &LaunchRequest{
		SessionId:   sessionID,
		Program:     args.Program,
		Args:        args.Args,
		Cwd:         args.Cwd,
		Env:         args.Env,
		StopOnEntry: args.StopOnEntry,
		NoDebug:     args.NoDebug,
	}

	_, err := c.client.Launch(ctx, req)
	return err
}

// Attach implements debugapi.Debugger.
func (c *Client) Attach(ctx context.Context, sessionID string, args debugapi.AttachRequestArguments) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &AttachRequest{
		SessionId: sessionID,
		Pid:       int32(args.PID),
		Program:   args.Program,
	}

	_, err := c.client.Attach(ctx, req)
	return err
}

// ConfigurationDone implements debugapi.Debugger.
func (c *Client) ConfigurationDone(ctx context.Context, sessionID string) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	_, err := c.client.ConfigurationDone(ctx, &ConfigurationDoneRequest{
		SessionId: sessionID,
	})
	return err
}

// Disconnect implements debugapi.Debugger.
func (c *Client) Disconnect(ctx context.Context, sessionID string, args *dap.DisconnectArguments) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DisconnectRequest{
		SessionId:         sessionID,
		Restart:           args.Restart,
		TerminateDebuggee: args.TerminateDebuggee,
		SuspendDebuggee:   args.SuspendDebuggee,
	}

	_, err := c.client.Disconnect(ctx, req)
	return err
}

// Terminate implements debugapi.Debugger.
func (c *Client) Terminate(ctx context.Context, sessionID string, args *dap.TerminateArguments) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &TerminateRequest{
		SessionId: sessionID,
		Restart:   args.Restart,
	}

	_, err := c.client.Terminate(ctx, req)
	return err
}

// Restart implements debugapi.Debugger.
func (c *Client) Restart(ctx context.Context, sessionID string) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	_, err := c.client.Restart(ctx, &RestartRequest{SessionId: sessionID})
	return err
}

// SetBreakpoints implements debugapi.Debugger.
func (c *Client) SetBreakpoints(ctx context.Context, sessionID string, args *dap.SetBreakpointsArguments) ([]dap.Breakpoint, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &SetBreakpointsRequest{
		SessionId:      sessionID,
		Source:         sourceToProtoFromDap(&args.Source),
		Breakpoints:    sourceBreakpointsToProto(args.Breakpoints),
		SourceModified: args.SourceModified,
	}

	resp, err := c.client.SetBreakpoints(ctx, req)
	if err != nil {
		return nil, err
	}

	return breakpointsFromProto(resp.GetBreakpoints()), nil
}

// SetFunctionBreakpoints implements debugapi.Debugger.
func (c *Client) SetFunctionBreakpoints(ctx context.Context, sessionID string, args *dap.SetFunctionBreakpointsArguments) ([]dap.Breakpoint, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &SetFunctionBreakpointsRequest{
		SessionId:   sessionID,
		Breakpoints: functionBreakpointsToProto(args.Breakpoints),
	}

	resp, err := c.client.SetFunctionBreakpoints(ctx, req)
	if err != nil {
		return nil, err
	}

	return breakpointsFromProto(resp.GetBreakpoints()), nil
}

// SetExceptionBreakpoints implements debugapi.Debugger.
func (c *Client) SetExceptionBreakpoints(ctx context.Context, sessionID string, args *dap.SetExceptionBreakpointsArguments) ([]dap.Breakpoint, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &SetExceptionBreakpointsRequest{
		SessionId: sessionID,
		Filters:   args.Filters,
	}

	resp, err := c.client.SetExceptionBreakpoints(ctx, req)
	if err != nil {
		return nil, err
	}

	return breakpointsFromProto(resp.GetBreakpoints()), nil
}

// Continue implements debugapi.Debugger.
func (c *Client) Continue(ctx context.Context, sessionID string, args *dap.ContinueArguments) (*dap.ContinueResponseBody, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &ContinueRequest{
		SessionId:    sessionID,
		ThreadId:     int32(args.ThreadId),
		SingleThread: args.SingleThread,
	}

	resp, err := c.client.Continue(ctx, req)
	if err != nil {
		return nil, err
	}

	return &dap.ContinueResponseBody{
		AllThreadsContinued: resp.GetAllThreadsContinued(),
	}, nil
}

// Next implements debugapi.Debugger.
func (c *Client) Next(ctx context.Context, sessionID string, args *dap.NextArguments) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &NextRequest{
		SessionId:    sessionID,
		ThreadId:     int32(args.ThreadId),
		SingleThread: args.SingleThread,
		Granularity:  string(args.Granularity),
	}

	_, err := c.client.Next(ctx, req)
	return err
}

// StepIn implements debugapi.Debugger.
func (c *Client) StepIn(ctx context.Context, sessionID string, args *dap.StepInArguments) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &StepInRequest{
		SessionId:    sessionID,
		ThreadId:     int32(args.ThreadId),
		SingleThread: args.SingleThread,
		TargetId:     int32(args.TargetId),
		Granularity:  string(args.Granularity),
	}

	_, err := c.client.StepIn(ctx, req)
	return err
}

// StepOut implements debugapi.Debugger.
func (c *Client) StepOut(ctx context.Context, sessionID string, args *dap.StepOutArguments) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &StepOutRequest{
		SessionId:    sessionID,
		ThreadId:     int32(args.ThreadId),
		SingleThread: args.SingleThread,
		Granularity:  string(args.Granularity),
	}

	_, err := c.client.StepOut(ctx, req)
	return err
}

// StepBack implements debugapi.Debugger.
func (c *Client) StepBack(ctx context.Context, sessionID string, args *dap.StepBackArguments) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &StepBackRequest{
		SessionId:    sessionID,
		ThreadId:     int32(args.ThreadId),
		SingleThread: args.SingleThread,
		Granularity:  string(args.Granularity),
	}

	_, err := c.client.StepBack(ctx, req)
	return err
}

// ReverseContinue implements debugapi.Debugger.
func (c *Client) ReverseContinue(ctx context.Context, sessionID string, args *dap.ReverseContinueArguments) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &ReverseContinueRequest{
		SessionId:    sessionID,
		ThreadId:     int32(args.ThreadId),
		SingleThread: args.SingleThread,
	}

	_, err := c.client.ReverseContinue(ctx, req)
	return err
}

// Pause implements debugapi.Debugger.
func (c *Client) Pause(ctx context.Context, sessionID string, args *dap.PauseArguments) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	_, err := c.client.Pause(ctx, &PauseRequest{
		SessionId: sessionID,
		ThreadId:  int32(args.ThreadId),
	})
	return err
}

// Threads implements debugapi.Debugger.
func (c *Client) Threads(ctx context.Context, sessionID string) ([]dap.Thread, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	resp, err := c.client.Threads(ctx, &ThreadsRequest{SessionId: sessionID})
	if err != nil {
		return nil, err
	}

	return threadsFromProto(resp.GetThreads()), nil
}

// StackTrace implements debugapi.Debugger.
func (c *Client) StackTrace(ctx context.Context, sessionID string, args *dap.StackTraceArguments) (*dap.StackTraceResponseBody, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &StackTraceRequest{
		SessionId:  sessionID,
		ThreadId:   int32(args.ThreadId),
		StartFrame: int32(args.StartFrame),
		Levels:     int32(args.Levels),
	}

	resp, err := c.client.StackTrace(ctx, req)
	if err != nil {
		return nil, err
	}

	return &dap.StackTraceResponseBody{
		StackFrames: stackFramesFromProto(resp.GetStackFrames()),
		TotalFrames: int(resp.GetTotalFrames()),
	}, nil
}

// Scopes implements debugapi.Debugger.
func (c *Client) Scopes(ctx context.Context, sessionID string, args *dap.ScopesArguments) ([]dap.Scope, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	resp, err := c.client.Scopes(ctx, &ScopesRequest{
		SessionId: sessionID,
		FrameId:   int32(args.FrameId),
	})
	if err != nil {
		return nil, err
	}

	return scopesFromProto(resp.GetScopes()), nil
}

// Variables implements debugapi.Debugger.
func (c *Client) Variables(ctx context.Context, sessionID string, args *dap.VariablesArguments) ([]dap.Variable, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &VariablesRequest{
		SessionId:          sessionID,
		VariablesReference: int32(args.VariablesReference),
		Filter:             string(args.Filter),
		Start:              int32(args.Start),
		Count:              int32(args.Count),
	}

	resp, err := c.client.Variables(ctx, req)
	if err != nil {
		return nil, err
	}

	return variablesFromProto(resp.GetVariables()), nil
}

// SetVariable implements debugapi.Debugger.
func (c *Client) SetVariable(ctx context.Context, sessionID string, args *dap.SetVariableArguments) (*dap.SetVariableResponseBody, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &SetVariableRequest{
		SessionId:          sessionID,
		VariablesReference: int32(args.VariablesReference),
		Name:               args.Name,
		Value:              args.Value,
	}

	resp, err := c.client.SetVariable(ctx, req)
	if err != nil {
		return nil, err
	}

	return &dap.SetVariableResponseBody{
		Value:              resp.GetValue(),
		Type:               resp.GetType(),
		VariablesReference: int(resp.GetVariablesReference()),
		NamedVariables:     int(resp.GetNamedVariables()),
		IndexedVariables:   int(resp.GetIndexedVariables()),
	}, nil
}

// Source implements debugapi.Debugger.
func (c *Client) Source(ctx context.Context, sessionID string, args *dap.SourceArguments) (*dap.SourceResponseBody, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &SourceRequest{
		SessionId:       sessionID,
		SourceReference: int32(args.SourceReference),
		Source:          sourceToProtoFromDapPtr(args.Source),
	}

	resp, err := c.client.Source(ctx, req)
	if err != nil {
		return nil, err
	}

	return &dap.SourceResponseBody{
		Content:  resp.GetContent(),
		MimeType: resp.GetMimeType(),
	}, nil
}

// Evaluate implements debugapi.Debugger.
func (c *Client) Evaluate(ctx context.Context, sessionID string, args *dap.EvaluateArguments) (*dap.EvaluateResponseBody, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &EvaluateRequest{
		SessionId:  sessionID,
		Expression: args.Expression,
		FrameId:    int32(args.FrameId),
		Context:    string(args.Context),
	}

	resp, err := c.client.Evaluate(ctx, req)
	if err != nil {
		return nil, err
	}

	return &dap.EvaluateResponseBody{
		Result:             resp.GetResult(),
		Type:               resp.GetType(),
		VariablesReference: int(resp.GetVariablesReference()),
		NamedVariables:     int(resp.GetNamedVariables()),
		IndexedVariables:   int(resp.GetIndexedVariables()),
		MemoryReference:    resp.GetMemoryReference(),
	}, nil
}

// SetExpression implements debugapi.Debugger.
func (c *Client) SetExpression(ctx context.Context, sessionID string, args *dap.SetExpressionArguments) (*dap.SetExpressionResponseBody, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &SetExpressionRequest{
		SessionId:  sessionID,
		Expression: args.Expression,
		Value:      args.Value,
		FrameId:    int32(args.FrameId),
	}

	resp, err := c.client.SetExpression(ctx, req)
	if err != nil {
		return nil, err
	}

	return &dap.SetExpressionResponseBody{
		Value:              resp.GetValue(),
		Type:               resp.GetType(),
		VariablesReference: int(resp.GetVariablesReference()),
		NamedVariables:     int(resp.GetNamedVariables()),
		IndexedVariables:   int(resp.GetIndexedVariables()),
	}, nil
}

// Completions implements debugapi.Debugger.
func (c *Client) Completions(ctx context.Context, sessionID string, args *dap.CompletionsArguments) ([]dap.CompletionItem, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &CompletionsRequest{
		SessionId: sessionID,
		FrameId:   int32(args.FrameId),
		Text:      args.Text,
		Column:    int32(args.Column),
		Line:      int32(args.Line),
	}

	resp, err := c.client.Completions(ctx, req)
	if err != nil {
		return nil, err
	}

	return completionItemsFromProto(resp.GetTargets()), nil
}

// ExceptionInfo implements debugapi.Debugger.
func (c *Client) ExceptionInfo(ctx context.Context, sessionID string, args *dap.ExceptionInfoArguments) (*dap.ExceptionInfoResponseBody, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &ExceptionInfoRequest{
		SessionId: sessionID,
		ThreadId:  int32(args.ThreadId),
	}

	resp, err := c.client.ExceptionInfo(ctx, req)
	if err != nil {
		return nil, err
	}

	return &dap.ExceptionInfoResponseBody{
		ExceptionId: resp.GetExceptionId(),
		Description: resp.GetDescription(),
		BreakMode:   dap.ExceptionBreakMode(resp.GetBreakMode()),
	}, nil
}

// Modules implements debugapi.Debugger.
func (c *Client) Modules(ctx context.Context, sessionID string, args *dap.ModulesArguments) (*dap.ModulesResponseBody, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &ModulesRequest{
		SessionId:   sessionID,
		StartModule: int32(args.StartModule),
		ModuleCount: int32(args.ModuleCount),
	}

	resp, err := c.client.Modules(ctx, req)
	if err != nil {
		return nil, err
	}

	return &dap.ModulesResponseBody{
		Modules:      modulesFromProto(resp.GetModules()),
		TotalModules: int(resp.GetTotalModules()),
	}, nil
}

// LoadedSources implements debugapi.Debugger.
func (c *Client) LoadedSources(ctx context.Context, sessionID string) ([]dap.Source, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	resp, err := c.client.LoadedSources(ctx, &LoadedSourcesRequest{
		SessionId: sessionID,
	})
	if err != nil {
		return nil, err
	}

	return sourcesFromProto(resp.GetSources()), nil
}

// ReadMemory implements debugapi.Debugger.
func (c *Client) ReadMemory(ctx context.Context, sessionID string, args *dap.ReadMemoryArguments) (*dap.ReadMemoryResponseBody, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &ReadMemoryRequest{
		SessionId:       sessionID,
		MemoryReference: args.MemoryReference,
		Offset:          int64(args.Offset),
		Count:           int32(args.Count),
	}

	resp, err := c.client.ReadMemory(ctx, req)
	if err != nil {
		return nil, err
	}

	return &dap.ReadMemoryResponseBody{
		Address:         resp.GetAddress(),
		UnreadableBytes: int(resp.GetUnreadableBytes()),
		Data:            string(resp.GetData()),
	}, nil
}

// WriteMemory implements debugapi.Debugger.
func (c *Client) WriteMemory(ctx context.Context, sessionID string, args *dap.WriteMemoryArguments) (*dap.WriteMemoryResponseBody, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &WriteMemoryRequest{
		SessionId:       sessionID,
		MemoryReference: args.MemoryReference,
		Offset:          int64(args.Offset),
		Data:            []byte(args.Data),
		AllowPartial:    args.AllowPartial,
	}

	resp, err := c.client.WriteMemory(ctx, req)
	if err != nil {
		return nil, err
	}

	return &dap.WriteMemoryResponseBody{
		Offset:       int(resp.GetOffset()),
		BytesWritten: int(resp.GetBytesWritten()),
	}, nil
}

// Disassemble implements debugapi.Debugger.
func (c *Client) Disassemble(ctx context.Context, sessionID string, args *dap.DisassembleArguments) ([]dap.DisassembledInstruction, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &DisassembleRequest{
		SessionId:         sessionID,
		MemoryReference:   args.MemoryReference,
		Offset:            int64(args.Offset),
		InstructionOffset: int32(args.InstructionOffset),
		InstructionCount:  int32(args.InstructionCount),
		ResolveSymbols:    args.ResolveSymbols,
	}

	resp, err := c.client.Disassemble(ctx, req)
	if err != nil {
		return nil, err
	}

	return instructionsFromProto(resp.GetInstructions()), nil
}

// GotoTargets implements debugapi.Debugger.
func (c *Client) GotoTargets(ctx context.Context, sessionID string, args *dap.GotoTargetsArguments) ([]dap.GotoTarget, error) {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &GotoTargetsRequest{
		SessionId: sessionID,
		Source:    sourceToProtoFromDap(&args.Source),
		Line:      int32(args.Line),
		Column:    int32(args.Column),
	}

	resp, err := c.client.GotoTargets(ctx, req)
	if err != nil {
		return nil, err
	}

	return gotoTargetsFromProto(resp.GetTargets()), nil
}

// Goto implements debugapi.Debugger.
func (c *Client) Goto(ctx context.Context, sessionID string, args *dap.GotoArguments) error {
	ctx, cancel := joincontext.New(ctx, c.clientCtx)
	defer cancel()

	req := &GotoRequest{
		SessionId: sessionID,
		ThreadId:  int32(args.ThreadId),
		TargetId:  int32(args.TargetId),
	}

	_, err := c.client.Goto(ctx, req)
	return err
}

func capabilitiesFromProto(c *Capabilities) *dap.Capabilities {
	if c == nil {
		return nil
	}
	return &dap.Capabilities{
		SupportsConfigurationDoneRequest:      c.GetSupportsConfigurationDoneRequest(),
		SupportsFunctionBreakpoints:           c.GetSupportsFunctionBreakpoints(),
		SupportsConditionalBreakpoints:        c.GetSupportsConditionalBreakpoints(),
		SupportsHitConditionalBreakpoints:     c.GetSupportsHitConditionalBreakpoints(),
		SupportsEvaluateForHovers:             c.GetSupportsEvaluateForHovers(),
		SupportsStepBack:                      c.GetSupportsStepBack(),
		SupportsSetVariable:                   c.GetSupportsSetVariable(),
		SupportsRestartFrame:                  c.GetSupportsRestartFrame(),
		SupportsGotoTargetsRequest:            c.GetSupportsGotoTargetsRequest(),
		SupportsStepInTargetsRequest:          c.GetSupportsStepInTargetsRequest(),
		SupportsCompletionsRequest:            c.GetSupportsCompletionsRequest(),
		SupportsModulesRequest:                c.GetSupportsModulesRequest(),
		SupportsRestartRequest:                c.GetSupportsRestartRequest(),
		SupportsExceptionOptions:              c.GetSupportsExceptionOptions(),
		SupportsValueFormattingOptions:        c.GetSupportsValueFormattingOptions(),
		SupportsExceptionInfoRequest:          c.GetSupportsExceptionInfoRequest(),
		SupportTerminateDebuggee:              c.GetSupportTerminateDebuggee(),
		SupportSuspendDebuggee:                c.GetSupportSuspendDebuggee(),
		SupportsDelayedStackTraceLoading:      c.GetSupportsDelayedStackTraceLoading(),
		SupportsLoadedSourcesRequest:          c.GetSupportsLoadedSourcesRequest(),
		SupportsLogPoints:                     c.GetSupportsLogPoints(),
		SupportsTerminateThreadsRequest:       c.GetSupportsTerminateThreadsRequest(),
		SupportsSetExpression:                 c.GetSupportsSetExpression(),
		SupportsTerminateRequest:              c.GetSupportsTerminateRequest(),
		SupportsDataBreakpoints:               c.GetSupportsDataBreakpoints(),
		SupportsReadMemoryRequest:             c.GetSupportsReadMemoryRequest(),
		SupportsWriteMemoryRequest:            c.GetSupportsWriteMemoryRequest(),
		SupportsDisassembleRequest:            c.GetSupportsDisassembleRequest(),
		SupportsCancelRequest:                 c.GetSupportsCancelRequest(),
		SupportsBreakpointLocationsRequest:    c.GetSupportsBreakpointLocationsRequest(),
		SupportsClipboardContext:              c.GetSupportsClipboardContext(),
		SupportsSteppingGranularity:           c.GetSupportsSteppingGranularity(),
		SupportsInstructionBreakpoints:        c.GetSupportsInstructionBreakpoints(),
		SupportsExceptionFilterOptions:        c.GetSupportsExceptionFilterOptions(),
		SupportsSingleThreadExecutionRequests: c.GetSupportsSingleThreadExecutionRequests(),
	}
}

func sourceToProtoFromDap(s *dap.Source) *Source {
	if s == nil {
		return nil
	}
	return &Source{
		Name:             s.Name,
		Path:             s.Path,
		SourceReference:  int32(s.SourceReference),
		PresentationHint: string(s.PresentationHint),
		Origin:           s.Origin,
	}
}

func sourceToProtoFromDapPtr(s *dap.Source) *Source {
	if s == nil {
		return nil
	}
	return sourceToProtoFromDap(s)
}

func sourceBreakpointsToProto(bps []dap.SourceBreakpoint) []*SourceBreakpoint {
	result := make([]*SourceBreakpoint, len(bps))
	for i, bp := range bps {
		result[i] = &SourceBreakpoint{
			Line:         int32(bp.Line),
			Column:       int32(bp.Column),
			Condition:    bp.Condition,
			HitCondition: bp.HitCondition,
			LogMessage:   bp.LogMessage,
		}
	}
	return result
}

func functionBreakpointsToProto(bps []dap.FunctionBreakpoint) []*FunctionBreakpoint {
	result := make([]*FunctionBreakpoint, len(bps))
	for i, bp := range bps {
		result[i] = &FunctionBreakpoint{
			Name:         bp.Name,
			Condition:    bp.Condition,
			HitCondition: bp.HitCondition,
		}
	}
	return result
}

func breakpointsFromProto(bps []*Breakpoint) []dap.Breakpoint {
	result := make([]dap.Breakpoint, len(bps))
	for i, bp := range bps {
		result[i] = dap.Breakpoint{
			Id:                   int(bp.GetId()),
			Verified:             bp.GetVerified(),
			Message:              bp.GetMessage(),
			Source:               sourceFromProtoDapPtr(bp.GetSource()),
			Line:                 int(bp.GetLine()),
			Column:               int(bp.GetColumn()),
			EndLine:              int(bp.GetEndLine()),
			EndColumn:            int(bp.GetEndColumn()),
			InstructionReference: bp.GetInstructionReference(),
			Offset:               int(bp.GetOffset()),
		}
	}
	return result
}

func sourceFromProtoDap(s *Source) dap.Source {
	if s == nil {
		return dap.Source{}
	}
	return dap.Source{
		Name:             s.GetName(),
		Path:             s.GetPath(),
		SourceReference:  int(s.GetSourceReference()),
		PresentationHint: s.GetPresentationHint(),
		Origin:           s.GetOrigin(),
	}
}

func sourceFromProtoDapPtr(s *Source) *dap.Source {
	if s == nil {
		return nil
	}
	src := sourceFromProtoDap(s)
	return &src
}

func threadsFromProto(threads []*Thread) []dap.Thread {
	result := make([]dap.Thread, len(threads))
	for i, t := range threads {
		result[i] = dap.Thread{
			Id:   int(t.GetId()),
			Name: t.GetName(),
		}
	}
	return result
}

func stackFramesFromProto(frames []*StackFrame) []dap.StackFrame {
	result := make([]dap.StackFrame, len(frames))
	for i, f := range frames {
		result[i] = dap.StackFrame{
			Id:                          int(f.GetId()),
			Name:                        f.GetName(),
			Source:                      sourceFromProtoDapPtr(f.GetSource()),
			Line:                        int(f.GetLine()),
			Column:                      int(f.GetColumn()),
			EndLine:                     int(f.GetEndLine()),
			EndColumn:                   int(f.GetEndColumn()),
			CanRestart:                  f.GetCanRestart(),
			InstructionPointerReference: f.GetInstructionPointerReference(),
			ModuleId:                    int(f.GetModuleId()),
			PresentationHint:            f.GetPresentationHint(),
		}
	}
	return result
}

func scopesFromProto(scopes []*Scope) []dap.Scope {
	result := make([]dap.Scope, len(scopes))
	for i, s := range scopes {
		result[i] = dap.Scope{
			Name:               s.GetName(),
			PresentationHint:   s.GetPresentationHint(),
			VariablesReference: int(s.GetVariablesReference()),
			NamedVariables:     int(s.GetNamedVariables()),
			IndexedVariables:   int(s.GetIndexedVariables()),
			Expensive:          s.GetExpensive(),
			Source:             sourceFromProtoDapPtr(s.GetSource()),
			Line:               int(s.GetLine()),
			Column:             int(s.GetColumn()),
			EndLine:            int(s.GetEndLine()),
			EndColumn:          int(s.GetEndColumn()),
		}
	}
	return result
}

func variablesFromProto(vars []*Variable) []dap.Variable {
	result := make([]dap.Variable, len(vars))
	for i, v := range vars {
		var hint *dap.VariablePresentationHint
		if h := v.GetPresentationHint(); h != "" {
			hint = &dap.VariablePresentationHint{Kind: h}
		}
		result[i] = dap.Variable{
			Name:               v.GetName(),
			Value:              v.GetValue(),
			Type:               v.GetType(),
			PresentationHint:   hint,
			EvaluateName:       v.GetEvaluateName(),
			VariablesReference: int(v.GetVariablesReference()),
			NamedVariables:     int(v.GetNamedVariables()),
			IndexedVariables:   int(v.GetIndexedVariables()),
			MemoryReference:    v.GetMemoryReference(),
		}
	}
	return result
}

func instructionsFromProto(instructions []*DisassembledInstruction) []dap.DisassembledInstruction {
	result := make([]dap.DisassembledInstruction, len(instructions))
	for i, inst := range instructions {
		result[i] = dap.DisassembledInstruction{
			Address:          inst.GetAddress(),
			InstructionBytes: inst.GetInstructionBytes(),
			Instruction:      inst.GetInstruction(),
			Symbol:           inst.GetSymbol(),
			Location:         sourceFromProtoDapPtr(inst.GetLocation()),
			Line:             int(inst.GetLine()),
			Column:           int(inst.GetColumn()),
			EndLine:          int(inst.GetEndLine()),
			EndColumn:        int(inst.GetEndColumn()),
		}
	}
	return result
}

func completionItemsFromProto(items []*CompletionItem) []dap.CompletionItem {
	result := make([]dap.CompletionItem, len(items))
	for i, item := range items {
		result[i] = dap.CompletionItem{
			Label:           item.GetLabel(),
			Text:            item.GetText(),
			SortText:        item.GetSortText(),
			Detail:          item.GetDetail(),
			Type:            dap.CompletionItemType(item.GetType()),
			Start:           int(item.GetStart()),
			Length:          int(item.GetLength()),
			SelectionStart:  int(item.GetSelectionStart()),
			SelectionLength: int(item.GetSelectionLength()),
		}
	}
	return result
}

func modulesFromProto(modules []*Module) []dap.Module {
	result := make([]dap.Module, len(modules))
	for i, m := range modules {
		result[i] = dap.Module{
			Id:             m.GetId(),
			Name:           m.GetName(),
			Path:           m.GetPath(),
			IsOptimized:    m.GetIsOptimized(),
			IsUserCode:     m.GetIsUserCode(),
			Version:        m.GetVersion(),
			SymbolStatus:   m.GetSymbolStatus(),
			SymbolFilePath: m.GetSymbolFilePath(),
			DateTimeStamp:  m.GetDateTimeStamp(),
			AddressRange:   m.GetAddressRange(),
		}
	}
	return result
}

func sourcesFromProto(sources []*Source) []dap.Source {
	result := make([]dap.Source, len(sources))
	for i, s := range sources {
		result[i] = sourceFromProtoDap(s)
	}
	return result
}

func gotoTargetsFromProto(targets []*GotoTarget) []dap.GotoTarget {
	result := make([]dap.GotoTarget, len(targets))
	for i, t := range targets {
		result[i] = dap.GotoTarget{
			Id:                          int(t.GetId()),
			Label:                       t.GetLabel(),
			Line:                        int(t.GetLine()),
			Column:                      int(t.GetColumn()),
			EndLine:                     int(t.GetEndLine()),
			EndColumn:                   int(t.GetEndColumn()),
			InstructionPointerReference: t.GetInstructionPointerReference(),
		}
	}
	return result
}

// clientCapabilitiesToProto converts a debugapi.ClientCapabilities
// into its wire form.
func clientCapabilitiesToProto(c debugapi.ClientCapabilities) *ClientCapabilities {
	return &ClientCapabilities{
		ClientId:                     c.ClientID,
		ClientName:                   c.ClientName,
		Locale:                       c.Locale,
		LinesStartAt_1:               c.LinesStartAt1,
		ColumnsStartAt_1:             c.ColumnsStartAt1,
		PathFormat:                   c.PathFormat,
		SupportsVariableType:         c.SupportsVariableType,
		SupportsVariablePaging:       c.SupportsVariablePaging,
		SupportsRunInTerminalRequest: c.SupportsRunInTerminalRequest,
		SupportsMemoryReferences:     c.SupportsMemoryReferences,
		SupportsProgressReporting:    c.SupportsProgressReporting,
		SupportsInvalidatedEvent:     c.SupportsInvalidatedEvent,
		SupportsMemoryEvent:          c.SupportsMemoryEvent,
	}
}

// dapEventFromProto reconstructs a concrete dap.EventMessage from
// its wire form. Returns nil if the message cannot be decoded as
// a DAP event.
func dapEventFromProto(e *Event) dap.EventMessage {
	if e == nil || len(e.GetBody()) == 0 {
		return nil
	}
	msg, err := dap.DecodeProtocolMessage(e.GetBody())
	if err != nil {
		return nil
	}
	ev, _ := msg.(dap.EventMessage)
	return ev
}
