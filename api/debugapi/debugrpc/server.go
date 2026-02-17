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

	"github.com/google/go-dap"
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
	args := &dap.InitializeRequestArguments{
		ClientID:                     req.GetClientId(),
		ClientName:                   req.GetClientName(),
		AdapterID:                    req.GetAdapterId(),
		Locale:                       req.GetLocale(),
		LinesStartAt1:                req.GetLinesStartAt_1(),
		ColumnsStartAt1:              req.GetColumnsStartAt_1(),
		PathFormat:                   req.GetPathFormat(),
		SupportsVariableType:         req.GetSupportsVariableType(),
		SupportsVariablePaging:       req.GetSupportsVariablePaging(),
		SupportsRunInTerminalRequest: req.GetSupportsRunInTerminalRequest(),
		SupportsMemoryReferences:     req.GetSupportsMemoryReferences(),
		SupportsProgressReporting:    req.GetSupportsProgressReporting(),
		SupportsInvalidatedEvent:     req.GetSupportsInvalidatedEvent(),
		SupportsMemoryEvent:          req.GetSupportsMemoryEvent(),
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
	args := debugapi.LaunchRequestArguments{
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
	args := debugapi.AttachRequestArguments{
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
	args := &dap.DisconnectArguments{
		Restart:           req.GetRestart(),
		TerminateDebuggee: req.GetTerminateDebuggee(),
		SuspendDebuggee:   req.GetSuspendDebuggee(),
	}

	if err := s.debugger.Disconnect(ctx, args); err != nil {
		return nil, err
	}

	return &DisconnectResponse{}, nil
}

// Terminate implements DebugServiceServer.
func (s *Server) Terminate(ctx context.Context, req *TerminateRequest) (*TerminateResponse, error) {
	args := &dap.TerminateArguments{
		Restart: req.GetRestart(),
	}

	if err := s.debugger.Terminate(ctx, args); err != nil {
		return nil, err
	}

	return &TerminateResponse{}, nil
}

// Restart implements DebugServiceServer.
func (s *Server) Restart(ctx context.Context, req *RestartRequest) (*RestartResponse, error) {
	if err := s.debugger.Restart(ctx); err != nil {
		return nil, err
	}
	return &RestartResponse{}, nil
}

// SetBreakpoints implements DebugServiceServer.
func (s *Server) SetBreakpoints(ctx context.Context, req *SetBreakpointsRequest) (*SetBreakpointsResponse, error) {
	args := &dap.SetBreakpointsArguments{
		Source:         sourceFromProto(req.GetSource()),
		Breakpoints:    sourceBreakpointsFromProto(req.GetBreakpoints()),
		SourceModified: req.GetSourceModified(),
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
	args := &dap.SetFunctionBreakpointsArguments{
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

// SetExceptionBreakpoints implements DebugServiceServer.
func (s *Server) SetExceptionBreakpoints(ctx context.Context, req *SetExceptionBreakpointsRequest) (*SetExceptionBreakpointsResponse, error) {
	args := &dap.SetExceptionBreakpointsArguments{
		Filters: req.GetFilters(),
	}

	bps, err := s.debugger.SetExceptionBreakpoints(ctx, args)
	if err != nil {
		return nil, err
	}

	return &SetExceptionBreakpointsResponse{
		Breakpoints: breakpointsToProto(bps),
	}, nil
}

// Continue implements DebugServiceServer.
func (s *Server) Continue(ctx context.Context, req *ContinueRequest) (*ContinueResponse, error) {
	args := &dap.ContinueArguments{
		ThreadId:     int(req.GetThreadId()),
		SingleThread: req.GetSingleThread(),
	}

	resp, err := s.debugger.Continue(ctx, args)
	if err != nil {
		return nil, err
	}

	return &ContinueResponse{
		AllThreadsContinued: resp.AllThreadsContinued,
	}, nil
}

// Pause implements DebugServiceServer.
func (s *Server) Pause(ctx context.Context, req *PauseRequest) (*PauseResponse, error) {
	args := &dap.PauseArguments{
		ThreadId: int(req.GetThreadId()),
	}

	if err := s.debugger.Pause(ctx, args); err != nil {
		return nil, err
	}
	return &PauseResponse{}, nil
}

// Next implements DebugServiceServer.
func (s *Server) Next(ctx context.Context, req *NextRequest) (*NextResponse, error) {
	args := &dap.NextArguments{
		ThreadId:     int(req.GetThreadId()),
		SingleThread: req.GetSingleThread(),
		Granularity:  dap.SteppingGranularity(req.GetGranularity()),
	}

	if err := s.debugger.Next(ctx, args); err != nil {
		return nil, err
	}
	return &NextResponse{}, nil
}

// StepIn implements DebugServiceServer.
func (s *Server) StepIn(ctx context.Context, req *StepInRequest) (*StepInResponse, error) {
	args := &dap.StepInArguments{
		ThreadId:     int(req.GetThreadId()),
		SingleThread: req.GetSingleThread(),
		TargetId:     int(req.GetTargetId()),
		Granularity:  dap.SteppingGranularity(req.GetGranularity()),
	}

	if err := s.debugger.StepIn(ctx, args); err != nil {
		return nil, err
	}
	return &StepInResponse{}, nil
}

// StepOut implements DebugServiceServer.
func (s *Server) StepOut(ctx context.Context, req *StepOutRequest) (*StepOutResponse, error) {
	args := &dap.StepOutArguments{
		ThreadId:     int(req.GetThreadId()),
		SingleThread: req.GetSingleThread(),
		Granularity:  dap.SteppingGranularity(req.GetGranularity()),
	}

	if err := s.debugger.StepOut(ctx, args); err != nil {
		return nil, err
	}
	return &StepOutResponse{}, nil
}

// StepBack implements DebugServiceServer.
func (s *Server) StepBack(ctx context.Context, req *StepBackRequest) (*StepBackResponse, error) {
	args := &dap.StepBackArguments{
		ThreadId:     int(req.GetThreadId()),
		SingleThread: req.GetSingleThread(),
		Granularity:  dap.SteppingGranularity(req.GetGranularity()),
	}

	if err := s.debugger.StepBack(ctx, args); err != nil {
		return nil, err
	}
	return &StepBackResponse{}, nil
}

// ReverseContinue implements DebugServiceServer.
func (s *Server) ReverseContinue(ctx context.Context, req *ReverseContinueRequest) (*ReverseContinueResponse, error) {
	args := &dap.ReverseContinueArguments{
		ThreadId:     int(req.GetThreadId()),
		SingleThread: req.GetSingleThread(),
	}

	if err := s.debugger.ReverseContinue(ctx, args); err != nil {
		return nil, err
	}
	return &ReverseContinueResponse{}, nil
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
	args := &dap.StackTraceArguments{
		ThreadId:   int(req.GetThreadId()),
		StartFrame: int(req.GetStartFrame()),
		Levels:     int(req.GetLevels()),
	}

	resp, err := s.debugger.StackTrace(ctx, args)
	if err != nil {
		return nil, err
	}

	return &StackTraceResponse{
		StackFrames: stackFramesToProto(resp.StackFrames),
		TotalFrames: int32(resp.TotalFrames),
	}, nil
}

// Scopes implements DebugServiceServer.
func (s *Server) Scopes(ctx context.Context, req *ScopesRequest) (*ScopesResponse, error) {
	args := &dap.ScopesArguments{
		FrameId: int(req.GetFrameId()),
	}

	scopes, err := s.debugger.Scopes(ctx, args)
	if err != nil {
		return nil, err
	}

	return &ScopesResponse{
		Scopes: scopesToProto(scopes),
	}, nil
}

// Variables implements DebugServiceServer.
func (s *Server) Variables(ctx context.Context, req *VariablesRequest) (*VariablesResponse, error) {
	args := &dap.VariablesArguments{
		VariablesReference: int(req.GetVariablesReference()),
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

// SetVariable implements DebugServiceServer.
func (s *Server) SetVariable(ctx context.Context, req *SetVariableRequest) (*SetVariableResponse, error) {
	args := &dap.SetVariableArguments{
		VariablesReference: int(req.GetVariablesReference()),
		Name:               req.GetName(),
		Value:              req.GetValue(),
	}

	resp, err := s.debugger.SetVariable(ctx, args)
	if err != nil {
		return nil, err
	}

	return &SetVariableResponse{
		Value:              resp.Value,
		Type:               resp.Type,
		VariablesReference: int32(resp.VariablesReference),
		NamedVariables:     int32(resp.NamedVariables),
		IndexedVariables:   int32(resp.IndexedVariables),
	}, nil
}

// Source implements DebugServiceServer.
func (s *Server) Source(ctx context.Context, req *SourceRequest) (*SourceResponse, error) {
	args := &dap.SourceArguments{
		SourceReference: int(req.GetSourceReference()),
		Source:          sourceFromProtoPtr(req.GetSource()),
	}

	resp, err := s.debugger.Source(ctx, args)
	if err != nil {
		return nil, err
	}

	return &SourceResponse{
		Content:  resp.Content,
		MimeType: resp.MimeType,
	}, nil
}

// Evaluate implements DebugServiceServer.
func (s *Server) Evaluate(ctx context.Context, req *EvaluateRequest) (*EvaluateResponse, error) {
	args := &dap.EvaluateArguments{
		Expression: req.GetExpression(),
		FrameId:    int(req.GetFrameId()),
		Context:    req.GetContext(),
	}

	resp, err := s.debugger.Evaluate(ctx, args)
	if err != nil {
		return nil, err
	}

	return &EvaluateResponse{
		Result:             resp.Result,
		Type:               resp.Type,
		VariablesReference: int32(resp.VariablesReference),
		NamedVariables:     int32(resp.NamedVariables),
		IndexedVariables:   int32(resp.IndexedVariables),
		MemoryReference:    resp.MemoryReference,
	}, nil
}

// SetExpression implements DebugServiceServer.
func (s *Server) SetExpression(ctx context.Context, req *SetExpressionRequest) (*SetExpressionResponse, error) {
	args := &dap.SetExpressionArguments{
		Expression: req.GetExpression(),
		Value:      req.GetValue(),
		FrameId:    int(req.GetFrameId()),
	}

	resp, err := s.debugger.SetExpression(ctx, args)
	if err != nil {
		return nil, err
	}

	return &SetExpressionResponse{
		Value:              resp.Value,
		Type:               resp.Type,
		VariablesReference: int32(resp.VariablesReference),
		NamedVariables:     int32(resp.NamedVariables),
		IndexedVariables:   int32(resp.IndexedVariables),
	}, nil
}

// Completions implements DebugServiceServer.
func (s *Server) Completions(ctx context.Context, req *CompletionsRequest) (*CompletionsResponse, error) {
	args := &dap.CompletionsArguments{
		FrameId: int(req.GetFrameId()),
		Text:    req.GetText(),
		Column:  int(req.GetColumn()),
		Line:    int(req.GetLine()),
	}

	items, err := s.debugger.Completions(ctx, args)
	if err != nil {
		return nil, err
	}

	return &CompletionsResponse{
		Targets: completionItemsToProto(items),
	}, nil
}

// ExceptionInfo implements DebugServiceServer.
func (s *Server) ExceptionInfo(ctx context.Context, req *ExceptionInfoRequest) (*ExceptionInfoResponse, error) {
	args := &dap.ExceptionInfoArguments{
		ThreadId: int(req.GetThreadId()),
	}

	resp, err := s.debugger.ExceptionInfo(ctx, args)
	if err != nil {
		return nil, err
	}

	return &ExceptionInfoResponse{
		ExceptionId: resp.ExceptionId,
		Description: resp.Description,
		BreakMode:   string(resp.BreakMode),
	}, nil
}

// Modules implements DebugServiceServer.
func (s *Server) Modules(ctx context.Context, req *ModulesRequest) (*ModulesResponse, error) {
	args := &dap.ModulesArguments{
		StartModule: int(req.GetStartModule()),
		ModuleCount: int(req.GetModuleCount()),
	}

	resp, err := s.debugger.Modules(ctx, args)
	if err != nil {
		return nil, err
	}

	return &ModulesResponse{
		Modules:      modulesToProto(resp.Modules),
		TotalModules: int32(resp.TotalModules),
	}, nil
}

// LoadedSources implements DebugServiceServer.
func (s *Server) LoadedSources(ctx context.Context, req *LoadedSourcesRequest) (*LoadedSourcesResponse, error) {
	sources, err := s.debugger.LoadedSources(ctx)
	if err != nil {
		return nil, err
	}

	return &LoadedSourcesResponse{
		Sources: sourcesToProto(sources),
	}, nil
}

// ReadMemory implements DebugServiceServer.
func (s *Server) ReadMemory(ctx context.Context, req *ReadMemoryRequest) (*ReadMemoryResponse, error) {
	args := &dap.ReadMemoryArguments{
		MemoryReference: req.GetMemoryReference(),
		Offset:          int(req.GetOffset()),
		Count:           int(req.GetCount()),
	}

	resp, err := s.debugger.ReadMemory(ctx, args)
	if err != nil {
		return nil, err
	}

	return &ReadMemoryResponse{
		Address:         resp.Address,
		UnreadableBytes: int32(resp.UnreadableBytes),
		Data:            []byte(resp.Data),
	}, nil
}

// WriteMemory implements DebugServiceServer.
func (s *Server) WriteMemory(ctx context.Context, req *WriteMemoryRequest) (*WriteMemoryResponse, error) {
	args := &dap.WriteMemoryArguments{
		MemoryReference: req.GetMemoryReference(),
		Offset:          int(req.GetOffset()),
		Data:            string(req.GetData()),
		AllowPartial:    req.GetAllowPartial(),
	}

	resp, err := s.debugger.WriteMemory(ctx, args)
	if err != nil {
		return nil, err
	}

	return &WriteMemoryResponse{
		Offset:       int64(resp.Offset),
		BytesWritten: int32(resp.BytesWritten),
	}, nil
}

// Disassemble implements DebugServiceServer.
func (s *Server) Disassemble(ctx context.Context, req *DisassembleRequest) (*DisassembleResponse, error) {
	args := &dap.DisassembleArguments{
		MemoryReference:   req.GetMemoryReference(),
		Offset:            int(req.GetOffset()),
		InstructionOffset: int(req.GetInstructionOffset()),
		InstructionCount:  int(req.GetInstructionCount()),
		ResolveSymbols:    req.GetResolveSymbols(),
	}

	instructions, err := s.debugger.Disassemble(ctx, args)
	if err != nil {
		return nil, err
	}

	return &DisassembleResponse{
		Instructions: instructionsToProto(instructions),
	}, nil
}

// GotoTargets implements DebugServiceServer.
func (s *Server) GotoTargets(ctx context.Context, req *GotoTargetsRequest) (*GotoTargetsResponse, error) {
	args := &dap.GotoTargetsArguments{
		Source: sourceFromProto(req.GetSource()),
		Line:   int(req.GetLine()),
		Column: int(req.GetColumn()),
	}

	targets, err := s.debugger.GotoTargets(ctx, args)
	if err != nil {
		return nil, err
	}

	return &GotoTargetsResponse{
		Targets: gotoTargetsToProto(targets),
	}, nil
}

// Goto implements DebugServiceServer.
func (s *Server) Goto(ctx context.Context, req *GotoRequest) (*GotoResponse, error) {
	args := &dap.GotoArguments{
		ThreadId: int(req.GetThreadId()),
		TargetId: int(req.GetTargetId()),
	}

	if err := s.debugger.Goto(ctx, args); err != nil {
		return nil, err
	}

	return &GotoResponse{}, nil
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
			protoEvent := eventToProto(ev)
			if err := stream.Send(protoEvent); err != nil {
				return err
			}
		}
	}
}

// === Conversion Functions ===

func capabilitiesToProto(c *dap.Capabilities) *Capabilities {
	if c == nil {
		return nil
	}
	return &Capabilities{
		SupportsConfigurationDoneRequest:          c.SupportsConfigurationDoneRequest,
		SupportsFunctionBreakpoints:               c.SupportsFunctionBreakpoints,
		SupportsConditionalBreakpoints:            c.SupportsConditionalBreakpoints,
		SupportsHitConditionalBreakpoints:         c.SupportsHitConditionalBreakpoints,
		SupportsEvaluateForHovers:                 c.SupportsEvaluateForHovers,
		SupportsStepBack:                          c.SupportsStepBack,
		SupportsSetVariable:                       c.SupportsSetVariable,
		SupportsRestartFrame:                      c.SupportsRestartFrame,
		SupportsGotoTargetsRequest:                c.SupportsGotoTargetsRequest,
		SupportsStepInTargetsRequest:              c.SupportsStepInTargetsRequest,
		SupportsCompletionsRequest:                c.SupportsCompletionsRequest,
		SupportsModulesRequest:                    c.SupportsModulesRequest,
		SupportsRestartRequest:                    c.SupportsRestartRequest,
		SupportsExceptionOptions:                  c.SupportsExceptionOptions,
		SupportsValueFormattingOptions:            c.SupportsValueFormattingOptions,
		SupportsExceptionInfoRequest:              c.SupportsExceptionInfoRequest,
		SupportTerminateDebuggee:                  c.SupportTerminateDebuggee,
		SupportSuspendDebuggee:                    c.SupportSuspendDebuggee,
		SupportsDelayedStackTraceLoading:          c.SupportsDelayedStackTraceLoading,
		SupportsLoadedSourcesRequest:              c.SupportsLoadedSourcesRequest,
		SupportsLogPoints:                         c.SupportsLogPoints,
		SupportsTerminateThreadsRequest:           c.SupportsTerminateThreadsRequest,
		SupportsSetExpression:                     c.SupportsSetExpression,
		SupportsTerminateRequest:                  c.SupportsTerminateRequest,
		SupportsDataBreakpoints:                   c.SupportsDataBreakpoints,
		SupportsReadMemoryRequest:                 c.SupportsReadMemoryRequest,
		SupportsWriteMemoryRequest:                c.SupportsWriteMemoryRequest,
		SupportsDisassembleRequest:                c.SupportsDisassembleRequest,
		SupportsCancelRequest:                     c.SupportsCancelRequest,
		SupportsBreakpointLocationsRequest:        c.SupportsBreakpointLocationsRequest,
		SupportsClipboardContext:                  c.SupportsClipboardContext,
		SupportsSteppingGranularity:               c.SupportsSteppingGranularity,
		SupportsInstructionBreakpoints:            c.SupportsInstructionBreakpoints,
		SupportsExceptionFilterOptions:            c.SupportsExceptionFilterOptions,
		SupportsSingleThreadExecutionRequests:     c.SupportsSingleThreadExecutionRequests,
	}
}

func sourceFromProto(s *Source) dap.Source {
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

func sourceFromProtoPtr(s *Source) *dap.Source {
	if s == nil {
		return nil
	}
	src := sourceFromProto(s)
	return &src
}

func sourceToProto(s *dap.Source) *Source {
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

func sourcesToProto(sources []dap.Source) []*Source {
	result := make([]*Source, len(sources))
	for i := range sources {
		result[i] = sourceToProto(&sources[i])
	}
	return result
}

func sourceBreakpointsFromProto(bps []*SourceBreakpoint) []dap.SourceBreakpoint {
	result := make([]dap.SourceBreakpoint, len(bps))
	for i, bp := range bps {
		result[i] = dap.SourceBreakpoint{
			Line:         int(bp.GetLine()),
			Column:       int(bp.GetColumn()),
			Condition:    bp.GetCondition(),
			HitCondition: bp.GetHitCondition(),
			LogMessage:   bp.GetLogMessage(),
		}
	}
	return result
}

func functionBreakpointsFromProto(bps []*FunctionBreakpoint) []dap.FunctionBreakpoint {
	result := make([]dap.FunctionBreakpoint, len(bps))
	for i, bp := range bps {
		result[i] = dap.FunctionBreakpoint{
			Name:         bp.GetName(),
			Condition:    bp.GetCondition(),
			HitCondition: bp.GetHitCondition(),
		}
	}
	return result
}

func breakpointsToProto(bps []dap.Breakpoint) []*Breakpoint {
	result := make([]*Breakpoint, len(bps))
	for i, bp := range bps {
		result[i] = &Breakpoint{
			Id:                   int32(bp.Id),
			Verified:             bp.Verified,
			Message:              bp.Message,
			Source:               sourceToProto(bp.Source),
			Line:                 int32(bp.Line),
			Column:               int32(bp.Column),
			EndLine:              int32(bp.EndLine),
			EndColumn:            int32(bp.EndColumn),
			InstructionReference: bp.InstructionReference,
			Offset:               int32(bp.Offset),
		}
	}
	return result
}

func threadsToProto(threads []dap.Thread) []*Thread {
	result := make([]*Thread, len(threads))
	for i, t := range threads {
		result[i] = &Thread{
			Id:   int32(t.Id),
			Name: t.Name,
		}
	}
	return result
}

func stackFramesToProto(frames []dap.StackFrame) []*StackFrame {
	result := make([]*StackFrame, len(frames))
	for i, f := range frames {
		var moduleId int32
		if mid, ok := f.ModuleId.(int); ok {
			moduleId = int32(mid)
		}
		result[i] = &StackFrame{
			Id:                          int32(f.Id),
			Name:                        f.Name,
			Source:                      sourceToProto(f.Source),
			Line:                        int32(f.Line),
			Column:                      int32(f.Column),
			EndLine:                     int32(f.EndLine),
			EndColumn:                   int32(f.EndColumn),
			CanRestart:                  f.CanRestart,
			InstructionPointerReference: f.InstructionPointerReference,
			ModuleId:                    moduleId,
			PresentationHint:            f.PresentationHint,
		}
	}
	return result
}

func scopesToProto(scopes []dap.Scope) []*Scope {
	result := make([]*Scope, len(scopes))
	for i, s := range scopes {
		result[i] = &Scope{
			Name:               s.Name,
			PresentationHint:   s.PresentationHint,
			VariablesReference: int32(s.VariablesReference),
			NamedVariables:     int32(s.NamedVariables),
			IndexedVariables:   int32(s.IndexedVariables),
			Expensive:          s.Expensive,
			Source:             sourceToProto(s.Source),
			Line:               int32(s.Line),
			Column:             int32(s.Column),
			EndLine:            int32(s.EndLine),
			EndColumn:          int32(s.EndColumn),
		}
	}
	return result
}

func variablesToProto(vars []dap.Variable) []*Variable {
	result := make([]*Variable, len(vars))
	for i, v := range vars {
		var hint string
		if v.PresentationHint != nil {
			hint = v.PresentationHint.Kind
		}
		result[i] = &Variable{
			Name:               v.Name,
			Value:              v.Value,
			Type:               v.Type,
			PresentationHint:   hint,
			EvaluateName:       v.EvaluateName,
			VariablesReference: int32(v.VariablesReference),
			NamedVariables:     int32(v.NamedVariables),
			IndexedVariables:   int32(v.IndexedVariables),
			MemoryReference:    v.MemoryReference,
		}
	}
	return result
}

func instructionsToProto(instructions []dap.DisassembledInstruction) []*DisassembledInstruction {
	result := make([]*DisassembledInstruction, len(instructions))
	for i, inst := range instructions {
		result[i] = &DisassembledInstruction{
			Address:          inst.Address,
			InstructionBytes: inst.InstructionBytes,
			Instruction:      inst.Instruction,
			Symbol:           inst.Symbol,
			Location:         sourceToProto(inst.Location),
			Line:             int32(inst.Line),
			Column:           int32(inst.Column),
			EndLine:          int32(inst.EndLine),
			EndColumn:        int32(inst.EndColumn),
		}
	}
	return result
}

func completionItemsToProto(items []dap.CompletionItem) []*CompletionItem {
	result := make([]*CompletionItem, len(items))
	for i, item := range items {
		result[i] = &CompletionItem{
			Label:           item.Label,
			Text:            item.Text,
			SortText:        item.SortText,
			Detail:          item.Detail,
			Type:            string(item.Type),
			Start:           int32(item.Start),
			Length:          int32(item.Length),
			SelectionStart:  int32(item.SelectionStart),
			SelectionLength: int32(item.SelectionLength),
		}
	}
	return result
}

func modulesToProto(modules []dap.Module) []*Module {
	result := make([]*Module, len(modules))
	for i, m := range modules {
		var id string
		switch v := m.Id.(type) {
		case int:
			id = string(rune(v))
		case string:
			id = v
		}
		result[i] = &Module{
			Id:             id,
			Name:           m.Name,
			Path:           m.Path,
			IsOptimized:    m.IsOptimized,
			IsUserCode:     m.IsUserCode,
			Version:        m.Version,
			SymbolStatus:   m.SymbolStatus,
			SymbolFilePath: m.SymbolFilePath,
			DateTimeStamp:  m.DateTimeStamp,
			AddressRange:   m.AddressRange,
		}
	}
	return result
}

func gotoTargetsToProto(targets []dap.GotoTarget) []*GotoTarget {
	result := make([]*GotoTarget, len(targets))
	for i, t := range targets {
		result[i] = &GotoTarget{
			Id:                          int32(t.Id),
			Label:                        t.Label,
			Line:                         int32(t.Line),
			Column:                       int32(t.Column),
			EndLine:                      int32(t.EndLine),
			EndColumn:                    int32(t.EndColumn),
			InstructionPointerReference: t.InstructionPointerReference,
		}
	}
	return result
}

func eventToProto(ev dap.EventMessage) *Event {
	protoEvent := &Event{
		Event: ev.GetEvent().Event,
	}
	return protoEvent
}
