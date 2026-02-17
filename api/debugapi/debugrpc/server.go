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

	"github.com/unstablebuild/rune-go-sdk/api/debugapi"
	"google.golang.org/grpc"
)

var _ DebugServiceServer = (*Server)(nil)

// Server implements DebugServiceServer by wrapping a debugapi.Debugger.
type Server struct {
	UnimplementedDebugServiceServer
	debugger debugapi.Debugger
}

// NewServer creates a new Server wrapping the given Debugger.
func NewServer(d debugapi.Debugger) *Server {
	return &Server{debugger: d}
}

// Register registers this server with the given gRPC server.
func (s *Server) Register(srv *grpc.Server) {
	RegisterDebugServiceServer(srv, s)
}

// Initialize implements DebugServiceServer.
func (s *Server) Initialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error) {
	args := debugapi.InitializeArguments{
		ClientID:   req.GetClientId(),
		ClientName: req.GetClientName(),
		AdapterID:  req.GetAdapterId(),
		Locale:     req.GetLocale(),
	}

	caps, err := s.debugger.Initialize(ctx, args)
	if err != nil {
		return nil, err
	}

	return &InitializeResponse{
		Capabilities: capabilitiesToProto(caps),
	}, nil
}

// Launch implements DebugServiceServer.
func (s *Server) Launch(ctx context.Context, req *LaunchRequest) (*LaunchResponse, error) {
	args := debugapi.LaunchArguments{
		Program:     req.GetProgram(),
		Args:        req.GetArgs(),
		Cwd:         req.GetCwd(),
		Env:         req.GetEnv(),
		StopOnEntry: req.GetStopOnEntry(),
		NoDebug:     req.GetNoDebug(),
	}

	if err := s.debugger.Launch(ctx, args); err != nil {
		return nil, err
	}

	return &LaunchResponse{}, nil
}

// Attach implements DebugServiceServer.
func (s *Server) Attach(ctx context.Context, req *AttachRequest) (*AttachResponse, error) {
	args := debugapi.AttachArguments{
		PID:     int(req.GetPid()),
		Program: req.GetProgram(),
	}

	if err := s.debugger.Attach(ctx, args); err != nil {
		return nil, err
	}

	return &AttachResponse{}, nil
}

// ConfigurationDone implements DebugServiceServer.
func (s *Server) ConfigurationDone(ctx context.Context, req *ConfigurationDoneRequest) (*ConfigurationDoneResponse, error) {
	if err := s.debugger.ConfigurationDone(ctx); err != nil {
		return nil, err
	}
	return &ConfigurationDoneResponse{}, nil
}

// Disconnect implements DebugServiceServer.
func (s *Server) Disconnect(ctx context.Context, req *DisconnectRequest) (*DisconnectResponse, error) {
	args := debugapi.DisconnectArguments{
		Restart:           req.GetRestart(),
		TerminateDebuggee: req.GetTerminateDebuggee(),
		SuspendDebuggee:   req.GetSuspendDebuggee(),
	}

	if err := s.debugger.Disconnect(ctx, args); err != nil {
		return nil, err
	}

	return &DisconnectResponse{}, nil
}

// Continue implements DebugServiceServer.
func (s *Server) Continue(ctx context.Context, req *ContinueRequest) (*ContinueResponse, error) {
	if err := s.debugger.Continue(ctx, req.GetThreadId()); err != nil {
		return nil, err
	}
	return &ContinueResponse{}, nil
}

// Pause implements DebugServiceServer.
func (s *Server) Pause(ctx context.Context, req *PauseRequest) (*PauseResponse, error) {
	if err := s.debugger.Pause(ctx, req.GetThreadId()); err != nil {
		return nil, err
	}
	return &PauseResponse{}, nil
}

// Next implements DebugServiceServer.
func (s *Server) Next(ctx context.Context, req *NextRequest) (*NextResponse, error) {
	if err := s.debugger.Next(ctx, req.GetThreadId()); err != nil {
		return nil, err
	}
	return &NextResponse{}, nil
}

// StepIn implements DebugServiceServer.
func (s *Server) StepIn(ctx context.Context, req *StepInRequest) (*StepInResponse, error) {
	if err := s.debugger.StepIn(ctx, req.GetThreadId()); err != nil {
		return nil, err
	}
	return &StepInResponse{}, nil
}

// StepOut implements DebugServiceServer.
func (s *Server) StepOut(ctx context.Context, req *StepOutRequest) (*StepOutResponse, error) {
	if err := s.debugger.StepOut(ctx, req.GetThreadId()); err != nil {
		return nil, err
	}
	return &StepOutResponse{}, nil
}

// SetBreakpoints implements DebugServiceServer.
func (s *Server) SetBreakpoints(ctx context.Context, req *SetBreakpointsRequest) (*SetBreakpointsResponse, error) {
	args := debugapi.SetBreakpointsArguments{
		Source:      sourceFromProto(req.GetSource()),
		Breakpoints: sourceBreakpointsFromProto(req.GetBreakpoints()),
	}

	bps, err := s.debugger.SetBreakpoints(ctx, args)
	if err != nil {
		return nil, err
	}

	return &SetBreakpointsResponse{
		Breakpoints: breakpointsToProto(bps),
	}, nil
}

// SetFunctionBreakpoints implements DebugServiceServer.
func (s *Server) SetFunctionBreakpoints(ctx context.Context, req *SetFunctionBreakpointsRequest) (*SetFunctionBreakpointsResponse, error) {
	args := debugapi.SetFunctionBreakpointsArguments{
		Breakpoints: functionBreakpointsFromProto(req.GetBreakpoints()),
	}

	bps, err := s.debugger.SetFunctionBreakpoints(ctx, args)
	if err != nil {
		return nil, err
	}

	return &SetFunctionBreakpointsResponse{
		Breakpoints: breakpointsToProto(bps),
	}, nil
}

// Threads implements DebugServiceServer.
func (s *Server) Threads(ctx context.Context, req *ThreadsRequest) (*ThreadsResponse, error) {
	threads, err := s.debugger.Threads(ctx)
	if err != nil {
		return nil, err
	}

	return &ThreadsResponse{
		Threads: threadsToProto(threads),
	}, nil
}

// StackTrace implements DebugServiceServer.
func (s *Server) StackTrace(ctx context.Context, req *StackTraceRequest) (*StackTraceResponse, error) {
	args := debugapi.StackTraceArguments{
		ThreadID:   req.GetThreadId(),
		StartFrame: int(req.GetStartFrame()),
		Levels:     int(req.GetLevels()),
	}

	frames, total, err := s.debugger.StackTrace(ctx, args)
	if err != nil {
		return nil, err
	}

	return &StackTraceResponse{
		StackFrames: stackFramesToProto(frames),
		TotalFrames: int32(total),
	}, nil
}

// Scopes implements DebugServiceServer.
func (s *Server) Scopes(ctx context.Context, req *ScopesRequest) (*ScopesResponse, error) {
	scopes, err := s.debugger.Scopes(ctx, req.GetFrameId())
	if err != nil {
		return nil, err
	}

	return &ScopesResponse{
		Scopes: scopesToProto(scopes),
	}, nil
}

// Variables implements DebugServiceServer.
func (s *Server) Variables(ctx context.Context, req *VariablesRequest) (*VariablesResponse, error) {
	args := debugapi.VariablesArguments{
		VariablesReference: req.GetVariablesReference(),
		Filter:             req.GetFilter(),
		Start:              int(req.GetStart()),
		Count:              int(req.GetCount()),
	}

	vars, err := s.debugger.Variables(ctx, args)
	if err != nil {
		return nil, err
	}

	return &VariablesResponse{
		Variables: variablesToProto(vars),
	}, nil
}

// Evaluate implements DebugServiceServer.
func (s *Server) Evaluate(ctx context.Context, req *EvaluateRequest) (*EvaluateResponse, error) {
	args := debugapi.EvaluateArguments{
		Expression: req.GetExpression(),
		FrameID:    req.GetFrameId(),
		Context:    req.GetContext(),
	}

	v, err := s.debugger.Evaluate(ctx, args)
	if err != nil {
		return nil, err
	}

	return &EvaluateResponse{
		Result: variableToProto(v),
	}, nil
}

// SetVariable implements DebugServiceServer.
func (s *Server) SetVariable(ctx context.Context, req *SetVariableRequest) (*SetVariableResponse, error) {
	args := debugapi.SetVariableArguments{
		VariablesReference: req.GetVariablesReference(),
		Name:               req.GetName(),
		Value:              req.GetValue(),
	}

	v, err := s.debugger.SetVariable(ctx, args)
	if err != nil {
		return nil, err
	}

	return &SetVariableResponse{
		Result: variableToProto(v),
	}, nil
}

// ReadMemory implements DebugServiceServer.
func (s *Server) ReadMemory(ctx context.Context, req *ReadMemoryRequest) (*ReadMemoryResponse, error) {
	args := debugapi.ReadMemoryArguments{
		MemoryReference: req.GetMemoryReference(),
		Offset:          req.GetOffset(),
		Count:           int(req.GetCount()),
	}

	data, err := s.debugger.ReadMemory(ctx, args)
	if err != nil {
		return nil, err
	}

	return &ReadMemoryResponse{
		Data: data,
	}, nil
}

// Disassemble implements DebugServiceServer.
func (s *Server) Disassemble(ctx context.Context, req *DisassembleRequest) (*DisassembleResponse, error) {
	args := debugapi.DisassembleArguments{
		MemoryReference:   req.GetMemoryReference(),
		Offset:            req.GetOffset(),
		InstructionOffset: int(req.GetInstructionOffset()),
		InstructionCount:  int(req.GetInstructionCount()),
	}

	instructions, err := s.debugger.Disassemble(ctx, args)
	if err != nil {
		return nil, err
	}

	return &DisassembleResponse{
		Instructions: instructionsToProto(instructions),
	}, nil
}

// SubscribeEvents implements DebugServiceServer.
func (s *Server) SubscribeEvents(req *SubscribeEventsRequest, stream DebugService_SubscribeEventsServer) error {
	events := s.debugger.Events()
	if events == nil {
		return nil
	}

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case ev, ok := <-events:
			if !ok {
				return nil
			}
			if err := stream.Send(eventToProto(&ev)); err != nil {
				return err
			}
		}
	}
}

// === Conversion Functions ===

func capabilitiesToProto(c *debugapi.Capabilities) *Capabilities {
	if c == nil {
		return nil
	}
	return &Capabilities{
		SupportsConfigurationDoneRequest:      c.SupportsConfigurationDoneRequest,
		SupportsFunctionBreakpoints:           c.SupportsFunctionBreakpoints,
		SupportsConditionalBreakpoints:        c.SupportsConditionalBreakpoints,
		SupportsHitConditionalBreakpoints:     c.SupportsHitConditionalBreakpoints,
		SupportsEvaluateForHovers:             c.SupportsEvaluateForHovers,
		SupportsStepBack:                      c.SupportsStepBack,
		SupportsSetVariable:                   c.SupportsSetVariable,
		SupportsRestartFrame:                  c.SupportsRestartFrame,
		SupportsGotoTargetsRequest:            c.SupportsGotoTargetsRequest,
		SupportsStepInTargetsRequest:          c.SupportsStepInTargetsRequest,
		SupportsCompletionsRequest:            c.SupportsCompletionsRequest,
		SupportsModulesRequest:                c.SupportsModulesRequest,
		SupportsRestartRequest:                c.SupportsRestartRequest,
		SupportsExceptionOptions:              c.SupportsExceptionOptions,
		SupportsValueFormattingOptions:        c.SupportsValueFormattingOptions,
		SupportsExceptionInfoRequest:          c.SupportsExceptionInfoRequest,
		SupportTerminateDebuggee:              c.SupportTerminateDebuggee,
		SupportsDelayedStackTraceLoading:      c.SupportsDelayedStackTraceLoading,
		SupportsLoadedSourcesRequest:          c.SupportsLoadedSourcesRequest,
		SupportsLogPoints:                     c.SupportsLogPoints,
		SupportsTerminateThreadsRequest:       c.SupportsTerminateThreadsRequest,
		SupportsSetExpression:                 c.SupportsSetExpression,
		SupportsTerminateRequest:              c.SupportsTerminateRequest,
		SupportsDataBreakpoints:               c.SupportsDataBreakpoints,
		SupportsReadMemoryRequest:             c.SupportsReadMemoryRequest,
		SupportsWriteMemoryRequest:            c.SupportsWriteMemoryRequest,
		SupportsDisassembleRequest:            c.SupportsDisassembleRequest,
		SupportsCancelRequest:                 c.SupportsCancelRequest,
		SupportsBreakpointLocationsRequest:    c.SupportsBreakpointLocationsRequest,
		SupportsClipboardContext:              c.SupportsClipboardContext,
		SupportsSingleThreadExecutionRequests: c.SupportsSingleThreadExecutionRequests,
	}
}

func sourceFromProto(s *Source) debugapi.Source {
	if s == nil {
		return debugapi.Source{}
	}
	return debugapi.Source{
		Name: s.GetName(),
		Path: s.GetPath(),
	}
}

func sourceToProto(s *debugapi.Source) *Source {
	if s == nil {
		return nil
	}
	return &Source{
		Name: s.Name,
		Path: s.Path,
	}
}

func sourceBreakpointsFromProto(bps []*SourceBreakpoint) []debugapi.SourceBreakpoint {
	result := make([]debugapi.SourceBreakpoint, len(bps))
	for i, bp := range bps {
		result[i] = debugapi.SourceBreakpoint{
			Line:         int(bp.GetLine()),
			Column:       int(bp.GetColumn()),
			Condition:    bp.GetCondition(),
			HitCondition: bp.GetHitCondition(),
			LogMessage:   bp.GetLogMessage(),
		}
	}
	return result
}

func functionBreakpointsFromProto(bps []*FunctionBreakpoint) []debugapi.FunctionBreakpoint {
	result := make([]debugapi.FunctionBreakpoint, len(bps))
	for i, bp := range bps {
		result[i] = debugapi.FunctionBreakpoint{
			Name:         bp.GetName(),
			Condition:    bp.GetCondition(),
			HitCondition: bp.GetHitCondition(),
		}
	}
	return result
}

func breakpointsToProto(bps []debugapi.Breakpoint) []*Breakpoint {
	result := make([]*Breakpoint, len(bps))
	for i, bp := range bps {
		result[i] = &Breakpoint{
			Id:        int32(bp.ID),
			Verified:  bp.Verified,
			Message:   bp.Message,
			Source:    sourceToProto(bp.Source),
			Line:      int32(bp.Line),
			Column:    int32(bp.Column),
			EndLine:   int32(bp.EndLine),
			EndColumn: int32(bp.EndColumn),
		}
	}
	return result
}

func threadsToProto(threads []debugapi.Thread) []*Thread {
	result := make([]*Thread, len(threads))
	for i, t := range threads {
		result[i] = &Thread{
			Id:   t.ID,
			Name: t.Name,
		}
	}
	return result
}

func stackFramesToProto(frames []debugapi.StackFrame) []*StackFrame {
	result := make([]*StackFrame, len(frames))
	for i, f := range frames {
		result[i] = &StackFrame{
			Id:                          f.ID,
			Name:                        f.Name,
			Source:                      sourceToProto(f.Source),
			Line:                        int32(f.Line),
			Column:                      int32(f.Column),
			EndLine:                     int32(f.EndLine),
			EndColumn:                   int32(f.EndColumn),
			InstructionPointerReference: f.InstructionPointerReference,
		}
	}
	return result
}

func scopesToProto(scopes []debugapi.Scope) []*Scope {
	result := make([]*Scope, len(scopes))
	for i, s := range scopes {
		result[i] = &Scope{
			Name:               s.Name,
			PresentationHint:   s.PresentationHint,
			VariablesReference: s.VariablesReference,
			NamedVariables:     int32(s.NamedVariables),
			IndexedVariables:   int32(s.IndexedVariables),
			Expensive:          s.Expensive,
		}
	}
	return result
}

func variablesToProto(vars []debugapi.Variable) []*Variable {
	result := make([]*Variable, len(vars))
	for i, v := range vars {
		result[i] = &Variable{
			Name:               v.Name,
			Value:              v.Value,
			Type:               v.Type,
			VariablesReference: v.VariablesReference,
			NamedVariables:     int32(v.NamedVariables),
			IndexedVariables:   int32(v.IndexedVariables),
			MemoryReference:    v.MemoryReference,
		}
	}
	return result
}

func variableToProto(v *debugapi.Variable) *Variable {
	if v == nil {
		return nil
	}
	return &Variable{
		Name:               v.Name,
		Value:              v.Value,
		Type:               v.Type,
		VariablesReference: v.VariablesReference,
		NamedVariables:     int32(v.NamedVariables),
		IndexedVariables:   int32(v.IndexedVariables),
		MemoryReference:    v.MemoryReference,
	}
}

func instructionsToProto(instructions []debugapi.DisassembledInstruction) []*DisassembledInstruction {
	result := make([]*DisassembledInstruction, len(instructions))
	for i, inst := range instructions {
		result[i] = &DisassembledInstruction{
			Address:          inst.Address,
			InstructionBytes: inst.InstructionBytes,
			Instruction:      inst.Instruction,
			Symbol:           inst.Symbol,
			Line:             int32(inst.Line),
			Column:           int32(inst.Column),
			EndLine:          int32(inst.EndLine),
			EndColumn:        int32(inst.EndColumn),
		}
	}
	return result
}

func eventToProto(e *debugapi.Event) *Event {
	if e == nil {
		return nil
	}

	ev := &Event{
		Type:              eventTypeToProto(e.Type),
		Reason:            stopReasonToProto(e.Reason),
		ThreadId:          e.ThreadID,
		AllThreadsStopped: e.AllThreadsStopped,
		ExitCode:          int32(e.ExitCode),
		ThreadReason:      e.ThreadReason,
		Category:          e.Category,
		Output:            e.Output,
		BreakpointReason:  e.BreakpointReason,
	}

	for _, id := range e.HitBreakpointIDs {
		ev.HitBreakpointIds = append(ev.HitBreakpointIds, int32(id))
	}

	if e.Breakpoint != nil {
		ev.Breakpoint = &Breakpoint{
			Id:        int32(e.Breakpoint.ID),
			Verified:  e.Breakpoint.Verified,
			Message:   e.Breakpoint.Message,
			Source:    sourceToProto(e.Breakpoint.Source),
			Line:      int32(e.Breakpoint.Line),
			Column:    int32(e.Breakpoint.Column),
			EndLine:   int32(e.Breakpoint.EndLine),
			EndColumn: int32(e.Breakpoint.EndColumn),
		}
	}

	return ev
}

func eventTypeToProto(t debugapi.EventType) Event_Type {
	switch t {
	case debugapi.EventTypeStopped:
		return Event_STOPPED
	case debugapi.EventTypeContinued:
		return Event_CONTINUED
	case debugapi.EventTypeExited:
		return Event_EXITED
	case debugapi.EventTypeTerminated:
		return Event_TERMINATED
	case debugapi.EventTypeThread:
		return Event_THREAD
	case debugapi.EventTypeOutput:
		return Event_OUTPUT
	case debugapi.EventTypeBreakpoint:
		return Event_BREAKPOINT
	case debugapi.EventTypeModule:
		return Event_MODULE
	default:
		return Event_STOPPED
	}
}

func stopReasonToProto(r debugapi.StopReason) Event_StopReason {
	switch r {
	case debugapi.StopReasonStep:
		return Event_STEP
	case debugapi.StopReasonBreakpoint:
		return Event_BREAKPOINT_HIT
	case debugapi.StopReasonException:
		return Event_EXCEPTION
	case debugapi.StopReasonPause:
		return Event_PAUSE
	case debugapi.StopReasonEntry:
		return Event_ENTRY
	case debugapi.StopReasonGoto:
		return Event_GOTO
	case debugapi.StopReasonFunctionBreakpoint:
		return Event_FUNCTION_BREAKPOINT
	case debugapi.StopReasonDataBreakpoint:
		return Event_DATA_BREAKPOINT
	default:
		return Event_STEP
	}
}
