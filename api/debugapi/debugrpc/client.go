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
	"io"
	"sync"
	"time"

	"github.com/unstablebuild/rune-go-sdk/api/debugapi"
	"google.golang.org/grpc"
)

const defaultTimeout = 30 * time.Second

var _ debugapi.Debugger = (*Client)(nil)

// Client implements debugapi.Debugger by making gRPC calls to a DebugService server.
type Client struct {
	cc              grpc.ClientConnInterface
	client          DebugServiceClient
	clientCtx       context.Context
	clientCancelCtx func()

	eventsMu sync.Mutex
	events   chan debugapi.Event
}

// NewClient creates a new Client connected to the given gRPC connection.
func NewClient(ctx context.Context, cc grpc.ClientConnInterface) *Client {
	clientCtx, cancel := context.WithCancel(ctx)
	return &Client{
		cc:              cc,
		client:          NewDebugServiceClient(cc),
		clientCtx:       clientCtx,
		clientCancelCtx: cancel,
		events:          make(chan debugapi.Event, 100),
	}
}

// Close closes the client and cancels any ongoing operations.
func (c *Client) Close() error {
	if c.clientCancelCtx != nil {
		c.clientCancelCtx()
	}
	close(c.events)
	return nil
}

func (c *Client) ctxWithTimeout() (context.Context, func()) {
	return context.WithTimeout(c.clientCtx, defaultTimeout)
}

// Initialize implements debugapi.Debugger.
func (c *Client) Initialize(ctx context.Context, args debugapi.InitializeArguments) (*debugapi.Capabilities, error) {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	req := &InitializeRequest{
		ClientId:   args.ClientID,
		ClientName: args.ClientName,
		AdapterId:  args.AdapterID,
		Locale:     args.Locale,
	}

	resp, err := c.client.Initialize(ctx, req)
	if err != nil {
		return nil, err
	}

	// Start event subscription after initialization
	go c.subscribeEvents()

	return capabilitiesFromProto(resp.GetCapabilities()), nil
}

// Launch implements debugapi.Debugger.
func (c *Client) Launch(ctx context.Context, args debugapi.LaunchArguments) error {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	req := &LaunchRequest{
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
func (c *Client) Attach(ctx context.Context, args debugapi.AttachArguments) error {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	req := &AttachRequest{
		Pid:     int32(args.PID),
		Program: args.Program,
	}

	_, err := c.client.Attach(ctx, req)
	return err
}

// ConfigurationDone implements debugapi.Debugger.
func (c *Client) ConfigurationDone(ctx context.Context) error {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	_, err := c.client.ConfigurationDone(ctx, &ConfigurationDoneRequest{})
	return err
}

// Disconnect implements debugapi.Debugger.
func (c *Client) Disconnect(ctx context.Context, args debugapi.DisconnectArguments) error {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	req := &DisconnectRequest{
		Restart:           args.Restart,
		TerminateDebuggee: args.TerminateDebuggee,
		SuspendDebuggee:   args.SuspendDebuggee,
	}

	_, err := c.client.Disconnect(ctx, req)
	return err
}

// Continue implements debugapi.Debugger.
func (c *Client) Continue(ctx context.Context, threadID int64) error {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	_, err := c.client.Continue(ctx, &ContinueRequest{ThreadId: threadID})
	return err
}

// Pause implements debugapi.Debugger.
func (c *Client) Pause(ctx context.Context, threadID int64) error {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	_, err := c.client.Pause(ctx, &PauseRequest{ThreadId: threadID})
	return err
}

// Next implements debugapi.Debugger.
func (c *Client) Next(ctx context.Context, threadID int64) error {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	_, err := c.client.Next(ctx, &NextRequest{ThreadId: threadID})
	return err
}

// StepIn implements debugapi.Debugger.
func (c *Client) StepIn(ctx context.Context, threadID int64) error {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	_, err := c.client.StepIn(ctx, &StepInRequest{ThreadId: threadID})
	return err
}

// StepOut implements debugapi.Debugger.
func (c *Client) StepOut(ctx context.Context, threadID int64) error {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	_, err := c.client.StepOut(ctx, &StepOutRequest{ThreadId: threadID})
	return err
}

// SetBreakpoints implements debugapi.Debugger.
func (c *Client) SetBreakpoints(ctx context.Context, args debugapi.SetBreakpointsArguments) ([]debugapi.Breakpoint, error) {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	req := &SetBreakpointsRequest{
		Source:      sourceToProto(&args.Source),
		Breakpoints: sourceBreakpointsToProto(args.Breakpoints),
	}

	resp, err := c.client.SetBreakpoints(ctx, req)
	if err != nil {
		return nil, err
	}

	return breakpointsFromProto(resp.GetBreakpoints()), nil
}

// SetFunctionBreakpoints implements debugapi.Debugger.
func (c *Client) SetFunctionBreakpoints(ctx context.Context, args debugapi.SetFunctionBreakpointsArguments) ([]debugapi.Breakpoint, error) {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	req := &SetFunctionBreakpointsRequest{
		Breakpoints: functionBreakpointsToProto(args.Breakpoints),
	}

	resp, err := c.client.SetFunctionBreakpoints(ctx, req)
	if err != nil {
		return nil, err
	}

	return breakpointsFromProto(resp.GetBreakpoints()), nil
}

// Threads implements debugapi.Debugger.
func (c *Client) Threads(ctx context.Context) ([]debugapi.Thread, error) {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	resp, err := c.client.Threads(ctx, &ThreadsRequest{})
	if err != nil {
		return nil, err
	}

	return threadsFromProto(resp.GetThreads()), nil
}

// StackTrace implements debugapi.Debugger.
func (c *Client) StackTrace(ctx context.Context, args debugapi.StackTraceArguments) ([]debugapi.StackFrame, int, error) {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	req := &StackTraceRequest{
		ThreadId:   args.ThreadID,
		StartFrame: int32(args.StartFrame),
		Levels:     int32(args.Levels),
	}

	resp, err := c.client.StackTrace(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	return stackFramesFromProto(resp.GetStackFrames()), int(resp.GetTotalFrames()), nil
}

// Scopes implements debugapi.Debugger.
func (c *Client) Scopes(ctx context.Context, frameID int64) ([]debugapi.Scope, error) {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	resp, err := c.client.Scopes(ctx, &ScopesRequest{FrameId: frameID})
	if err != nil {
		return nil, err
	}

	return scopesFromProto(resp.GetScopes()), nil
}

// Variables implements debugapi.Debugger.
func (c *Client) Variables(ctx context.Context, args debugapi.VariablesArguments) ([]debugapi.Variable, error) {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	req := &VariablesRequest{
		VariablesReference: args.VariablesReference,
		Filter:             args.Filter,
		Start:              int32(args.Start),
		Count:              int32(args.Count),
	}

	resp, err := c.client.Variables(ctx, req)
	if err != nil {
		return nil, err
	}

	return variablesFromProto(resp.GetVariables()), nil
}

// Evaluate implements debugapi.Debugger.
func (c *Client) Evaluate(ctx context.Context, args debugapi.EvaluateArguments) (*debugapi.Variable, error) {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	req := &EvaluateRequest{
		Expression: args.Expression,
		FrameId:    args.FrameID,
		Context:    args.Context,
	}

	resp, err := c.client.Evaluate(ctx, req)
	if err != nil {
		return nil, err
	}

	return variableFromProto(resp.GetResult()), nil
}

// SetVariable implements debugapi.Debugger.
func (c *Client) SetVariable(ctx context.Context, args debugapi.SetVariableArguments) (*debugapi.Variable, error) {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	req := &SetVariableRequest{
		VariablesReference: args.VariablesReference,
		Name:               args.Name,
		Value:              args.Value,
	}

	resp, err := c.client.SetVariable(ctx, req)
	if err != nil {
		return nil, err
	}

	return variableFromProto(resp.GetResult()), nil
}

// ReadMemory implements debugapi.Debugger.
func (c *Client) ReadMemory(ctx context.Context, args debugapi.ReadMemoryArguments) ([]byte, error) {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	req := &ReadMemoryRequest{
		MemoryReference: args.MemoryReference,
		Offset:          args.Offset,
		Count:           int32(args.Count),
	}

	resp, err := c.client.ReadMemory(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetData(), nil
}

// Disassemble implements debugapi.Debugger.
func (c *Client) Disassemble(ctx context.Context, args debugapi.DisassembleArguments) ([]debugapi.DisassembledInstruction, error) {
	ctx, cancel := c.ctxWithTimeout()
	defer cancel()

	req := &DisassembleRequest{
		MemoryReference:   args.MemoryReference,
		Offset:            args.Offset,
		InstructionOffset: int32(args.InstructionOffset),
		InstructionCount:  int32(args.InstructionCount),
	}

	resp, err := c.client.Disassemble(ctx, req)
	if err != nil {
		return nil, err
	}

	return instructionsFromProto(resp.GetInstructions()), nil
}

// Events implements debugapi.Debugger.
func (c *Client) Events() <-chan debugapi.Event {
	return c.events
}

func (c *Client) subscribeEvents() {
	stream, err := c.client.SubscribeEvents(c.clientCtx, &SubscribeEventsRequest{})
	if err != nil {
		return
	}

	for {
		ev, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			return
		}

		select {
		case c.events <- *eventFromProto(ev):
		case <-c.clientCtx.Done():
			return
		}
	}
}

// === Conversion Functions ===

func capabilitiesFromProto(c *Capabilities) *debugapi.Capabilities {
	if c == nil {
		return nil
	}
	return &debugapi.Capabilities{
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
		SupportsSingleThreadExecutionRequests: c.GetSupportsSingleThreadExecutionRequests(),
	}
}

func sourceBreakpointsToProto(bps []debugapi.SourceBreakpoint) []*SourceBreakpoint {
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

func functionBreakpointsToProto(bps []debugapi.FunctionBreakpoint) []*FunctionBreakpoint {
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

func breakpointsFromProto(bps []*Breakpoint) []debugapi.Breakpoint {
	result := make([]debugapi.Breakpoint, len(bps))
	for i, bp := range bps {
		result[i] = debugapi.Breakpoint{
			ID:        int(bp.GetId()),
			Verified:  bp.GetVerified(),
			Message:   bp.GetMessage(),
			Source:    sourceFromProtoPtr(bp.GetSource()),
			Line:      int(bp.GetLine()),
			Column:    int(bp.GetColumn()),
			EndLine:   int(bp.GetEndLine()),
			EndColumn: int(bp.GetEndColumn()),
		}
	}
	return result
}

func sourceFromProtoPtr(s *Source) *debugapi.Source {
	if s == nil {
		return nil
	}
	return &debugapi.Source{
		Name: s.GetName(),
		Path: s.GetPath(),
	}
}

func threadsFromProto(threads []*Thread) []debugapi.Thread {
	result := make([]debugapi.Thread, len(threads))
	for i, t := range threads {
		result[i] = debugapi.Thread{
			ID:   t.GetId(),
			Name: t.GetName(),
		}
	}
	return result
}

func stackFramesFromProto(frames []*StackFrame) []debugapi.StackFrame {
	result := make([]debugapi.StackFrame, len(frames))
	for i, f := range frames {
		result[i] = debugapi.StackFrame{
			ID:                          f.GetId(),
			Name:                        f.GetName(),
			Source:                      sourceFromProtoPtr(f.GetSource()),
			Line:                        int(f.GetLine()),
			Column:                      int(f.GetColumn()),
			EndLine:                     int(f.GetEndLine()),
			EndColumn:                   int(f.GetEndColumn()),
			InstructionPointerReference: f.GetInstructionPointerReference(),
		}
	}
	return result
}

func scopesFromProto(scopes []*Scope) []debugapi.Scope {
	result := make([]debugapi.Scope, len(scopes))
	for i, s := range scopes {
		result[i] = debugapi.Scope{
			Name:               s.GetName(),
			PresentationHint:   s.GetPresentationHint(),
			VariablesReference: s.GetVariablesReference(),
			NamedVariables:     int(s.GetNamedVariables()),
			IndexedVariables:   int(s.GetIndexedVariables()),
			Expensive:          s.GetExpensive(),
		}
	}
	return result
}

func variablesFromProto(vars []*Variable) []debugapi.Variable {
	result := make([]debugapi.Variable, len(vars))
	for i, v := range vars {
		result[i] = debugapi.Variable{
			Name:               v.GetName(),
			Value:              v.GetValue(),
			Type:               v.GetType(),
			VariablesReference: v.GetVariablesReference(),
			NamedVariables:     int(v.GetNamedVariables()),
			IndexedVariables:   int(v.GetIndexedVariables()),
			MemoryReference:    v.GetMemoryReference(),
		}
	}
	return result
}

func variableFromProto(v *Variable) *debugapi.Variable {
	if v == nil {
		return nil
	}
	return &debugapi.Variable{
		Name:               v.GetName(),
		Value:              v.GetValue(),
		Type:               v.GetType(),
		VariablesReference: v.GetVariablesReference(),
		NamedVariables:     int(v.GetNamedVariables()),
		IndexedVariables:   int(v.GetIndexedVariables()),
		MemoryReference:    v.GetMemoryReference(),
	}
}

func instructionsFromProto(instructions []*DisassembledInstruction) []debugapi.DisassembledInstruction {
	result := make([]debugapi.DisassembledInstruction, len(instructions))
	for i, inst := range instructions {
		result[i] = debugapi.DisassembledInstruction{
			Address:          inst.GetAddress(),
			InstructionBytes: inst.GetInstructionBytes(),
			Instruction:      inst.GetInstruction(),
			Symbol:           inst.GetSymbol(),
			Line:             int(inst.GetLine()),
			Column:           int(inst.GetColumn()),
			EndLine:          int(inst.GetEndLine()),
			EndColumn:        int(inst.GetEndColumn()),
		}
	}
	return result
}

func eventFromProto(e *Event) *debugapi.Event {
	if e == nil {
		return nil
	}

	ev := &debugapi.Event{
		Type:              eventTypeFromProto(e.GetType()),
		Reason:            stopReasonFromProto(e.GetReason()),
		ThreadID:          e.GetThreadId(),
		AllThreadsStopped: e.GetAllThreadsStopped(),
		ExitCode:          int(e.GetExitCode()),
		ThreadReason:      e.GetThreadReason(),
		Category:          e.GetCategory(),
		Output:            e.GetOutput(),
		BreakpointReason:  e.GetBreakpointReason(),
	}

	for _, id := range e.GetHitBreakpointIds() {
		ev.HitBreakpointIDs = append(ev.HitBreakpointIDs, int(id))
	}

	if bp := e.GetBreakpoint(); bp != nil {
		ev.Breakpoint = &debugapi.Breakpoint{
			ID:        int(bp.GetId()),
			Verified:  bp.GetVerified(),
			Message:   bp.GetMessage(),
			Source:    sourceFromProtoPtr(bp.GetSource()),
			Line:      int(bp.GetLine()),
			Column:    int(bp.GetColumn()),
			EndLine:   int(bp.GetEndLine()),
			EndColumn: int(bp.GetEndColumn()),
		}
	}

	return ev
}

func eventTypeFromProto(t Event_Type) debugapi.EventType {
	switch t {
	case Event_STOPPED:
		return debugapi.EventTypeStopped
	case Event_CONTINUED:
		return debugapi.EventTypeContinued
	case Event_EXITED:
		return debugapi.EventTypeExited
	case Event_TERMINATED:
		return debugapi.EventTypeTerminated
	case Event_THREAD:
		return debugapi.EventTypeThread
	case Event_OUTPUT:
		return debugapi.EventTypeOutput
	case Event_BREAKPOINT:
		return debugapi.EventTypeBreakpoint
	case Event_MODULE:
		return debugapi.EventTypeModule
	default:
		return debugapi.EventTypeStopped
	}
}

func stopReasonFromProto(r Event_StopReason) debugapi.StopReason {
	switch r {
	case Event_STEP:
		return debugapi.StopReasonStep
	case Event_BREAKPOINT_HIT:
		return debugapi.StopReasonBreakpoint
	case Event_EXCEPTION:
		return debugapi.StopReasonException
	case Event_PAUSE:
		return debugapi.StopReasonPause
	case Event_ENTRY:
		return debugapi.StopReasonEntry
	case Event_GOTO:
		return debugapi.StopReasonGoto
	case Event_FUNCTION_BREAKPOINT:
		return debugapi.StopReasonFunctionBreakpoint
	case Event_DATA_BREAKPOINT:
		return debugapi.StopReasonDataBreakpoint
	default:
		return debugapi.StopReasonStep
	}
}
