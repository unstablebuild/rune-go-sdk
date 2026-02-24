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

package idedebug

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/go-dap"
	"github.com/unstablebuild/rune-go-sdk/api/debugapi"
)

// Initialize configures the debug adapter with client
// capabilities and retrieves the adapter's capabilities.
func (m *Manager) Initialize(
	ctx context.Context,
	args *dap.InitializeRequestArguments,
) (*dap.Capabilities, error) {
	adapterID := args.AdapterID
	if adapterID == "" {
		return nil, errors.New("adapter ID is required")
	}
	cfg, ok := _debugAdapters[adapterID]
	if !ok {
		return nil, fmt.Errorf("unknown debug adapter: %s", adapterID)
	}

	srv, err := m.getOrCreateServer(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("initialize server: %w", err)
	}
	return srv.caps, nil
}

// Launch starts the debuggee. The request is sent without
// waiting for a response because in DAP the LaunchResponse
// only arrives after ConfigurationDone.
func (m *Manager) Launch(
	_ context.Context,
	args debugapi.LaunchRequestArguments,
) error {
	srv, err := m.serverForLaunch(args.Program)
	if err != nil {
		return err
	}
	launchArgs := map[string]any{
		"mode":        "debug",
		"program":     args.Program,
		"stopOnEntry": args.StopOnEntry,
		"noDebug":     args.NoDebug,
	}
	if len(args.Args) > 0 {
		launchArgs["args"] = args.Args
	}
	if args.Cwd != "" {
		launchArgs["cwd"] = args.Cwd
	}
	if len(args.Env) > 0 {
		launchArgs["env"] = args.Env
	}
	argsJSON, err := json.Marshal(launchArgs)
	if err != nil {
		return fmt.Errorf("marshal launch args: %w", err)
	}
	req := &dap.LaunchRequest{
		Request:   srv.newRequest("launch"),
		Arguments: argsJSON,
	}
	return srv.writeRequest(req)
}

// Attach connects to an already running debuggee.
// Like Launch, the response only arrives after
// ConfigurationDone.
func (m *Manager) Attach(
	_ context.Context,
	args debugapi.AttachRequestArguments,
) error {
	srv, err := m.activeServer()
	if err != nil {
		return err
	}
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("marshal attach args: %w", err)
	}
	req := &dap.AttachRequest{
		Request:   srv.newRequest("attach"),
		Arguments: argsJSON,
	}
	return srv.writeRequest(req)
}

// ConfigurationDone signals that configuration is done.
func (m *Manager) ConfigurationDone(
	ctx context.Context,
) error {
	srv, err := m.activeServer()
	if err != nil {
		return err
	}
	req := &dap.ConfigurationDoneRequest{
		Request: srv.newRequest("configurationDone"),
	}
	_, err = srv.sendRequest(ctx, req)
	return err
}

// Disconnect ends the debug session.
func (m *Manager) Disconnect(
	ctx context.Context,
	args *dap.DisconnectArguments,
) error {
	srv, err := m.activeServer()
	if err != nil {
		return err
	}
	req := &dap.DisconnectRequest{
		Request:   srv.newRequest("disconnect"),
		Arguments: args,
	}
	_, err = srv.sendRequest(ctx, req)
	return err
}

// Terminate requests graceful termination.
func (m *Manager) Terminate(
	ctx context.Context,
	args *dap.TerminateArguments,
) error {
	srv, err := m.activeServer()
	if err != nil {
		return err
	}
	req := &dap.TerminateRequest{
		Request:   srv.newRequest("terminate"),
		Arguments: args,
	}
	_, err = srv.sendRequest(ctx, req)
	return err
}

// Restart restarts the debug session.
func (m *Manager) Restart(ctx context.Context) error {
	srv, err := m.activeServer()
	if err != nil {
		return err
	}
	req := &dap.RestartRequest{
		Request: srv.newRequest("restart"),
	}
	_, err = srv.sendRequest(ctx, req)
	return err
}

// SetBreakpoints sets breakpoints for a source file.
func (m *Manager) SetBreakpoints(
	ctx context.Context,
	args *dap.SetBreakpointsArguments,
) ([]dap.Breakpoint, error) {
	var srv *debugServer
	var err error
	if args.Source.Path != "" {
		srv, err = m.serverForFile(args.Source.Path)
	} else {
		srv, err = m.activeServer()
	}
	if err != nil {
		return nil, err
	}
	req := &dap.SetBreakpointsRequest{
		Request:   srv.newRequest("setBreakpoints"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	bpResp, ok := resp.(*dap.SetBreakpointsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return bpResp.Body.Breakpoints, nil
}

// SetFunctionBreakpoints sets breakpoints on functions.
func (m *Manager) SetFunctionBreakpoints(
	ctx context.Context,
	args *dap.SetFunctionBreakpointsArguments,
) ([]dap.Breakpoint, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.SetFunctionBreakpointsRequest{
		Request:   srv.newRequest("setFunctionBreakpoints"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.SetFunctionBreakpointsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return r.Body.Breakpoints, nil
}

// SetExceptionBreakpoints configures exception bps.
func (m *Manager) SetExceptionBreakpoints(
	ctx context.Context,
	args *dap.SetExceptionBreakpointsArguments,
) ([]dap.Breakpoint, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.SetExceptionBreakpointsRequest{
		Request:   srv.newRequest("setExceptionBreakpoints"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.SetExceptionBreakpointsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return r.Body.Breakpoints, nil
}

// Continue resumes execution of all threads.
func (m *Manager) Continue(
	ctx context.Context,
	args *dap.ContinueArguments,
) (*dap.ContinueResponseBody, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.ContinueRequest{
		Request:   srv.newRequest("continue"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.ContinueResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return &r.Body, nil
}

// Next executes one step over.
func (m *Manager) Next(
	ctx context.Context,
	args *dap.NextArguments,
) error {
	srv, err := m.activeServer()
	if err != nil {
		return err
	}
	req := &dap.NextRequest{
		Request:   srv.newRequest("next"),
		Arguments: *args,
	}
	_, err = srv.sendRequest(ctx, req)
	return err
}

// StepIn steps into a function call.
func (m *Manager) StepIn(
	ctx context.Context,
	args *dap.StepInArguments,
) error {
	srv, err := m.activeServer()
	if err != nil {
		return err
	}
	req := &dap.StepInRequest{
		Request:   srv.newRequest("stepIn"),
		Arguments: *args,
	}
	_, err = srv.sendRequest(ctx, req)
	return err
}

// StepOut steps out of the current function.
func (m *Manager) StepOut(
	ctx context.Context,
	args *dap.StepOutArguments,
) error {
	srv, err := m.activeServer()
	if err != nil {
		return err
	}
	req := &dap.StepOutRequest{
		Request:   srv.newRequest("stepOut"),
		Arguments: *args,
	}
	_, err = srv.sendRequest(ctx, req)
	return err
}

// StepBack executes one backward step.
func (m *Manager) StepBack(
	ctx context.Context,
	args *dap.StepBackArguments,
) error {
	srv, err := m.activeServer()
	if err != nil {
		return err
	}
	req := &dap.StepBackRequest{
		Request:   srv.newRequest("stepBack"),
		Arguments: *args,
	}
	_, err = srv.sendRequest(ctx, req)
	return err
}

// ReverseContinue resumes backward execution.
func (m *Manager) ReverseContinue(
	ctx context.Context,
	args *dap.ReverseContinueArguments,
) error {
	srv, err := m.activeServer()
	if err != nil {
		return err
	}
	req := &dap.ReverseContinueRequest{
		Request: srv.newRequest("reverseContinue"),
		Arguments: *args,
	}
	_, err = srv.sendRequest(ctx, req)
	return err
}

// Pause suspends execution.
func (m *Manager) Pause(
	ctx context.Context,
	args *dap.PauseArguments,
) error {
	srv, err := m.activeServer()
	if err != nil {
		return err
	}
	req := &dap.PauseRequest{
		Request:   srv.newRequest("pause"),
		Arguments: *args,
	}
	_, err = srv.sendRequest(ctx, req)
	return err
}

// Threads retrieves all threads.
func (m *Manager) Threads(
	ctx context.Context,
) ([]dap.Thread, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.ThreadsRequest{
		Request: srv.newRequest("threads"),
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.ThreadsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return r.Body.Threads, nil
}

// StackTrace returns the call stack for a thread.
func (m *Manager) StackTrace(
	ctx context.Context,
	args *dap.StackTraceArguments,
) (*dap.StackTraceResponseBody, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.StackTraceRequest{
		Request:   srv.newRequest("stackTrace"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.StackTraceResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return &r.Body, nil
}

// Scopes returns variable scopes for a stack frame.
func (m *Manager) Scopes(
	ctx context.Context,
	args *dap.ScopesArguments,
) ([]dap.Scope, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.ScopesRequest{
		Request:   srv.newRequest("scopes"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.ScopesResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return r.Body.Scopes, nil
}

// Variables retrieves child variables.
func (m *Manager) Variables(
	ctx context.Context,
	args *dap.VariablesArguments,
) ([]dap.Variable, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.VariablesRequest{
		Request:   srv.newRequest("variables"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.VariablesResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return r.Body.Variables, nil
}

// SetVariable modifies a variable's value.
func (m *Manager) SetVariable(
	ctx context.Context,
	args *dap.SetVariableArguments,
) (*dap.SetVariableResponseBody, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.SetVariableRequest{
		Request:   srv.newRequest("setVariable"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.SetVariableResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return &r.Body, nil
}

// Source retrieves source code.
func (m *Manager) Source(
	ctx context.Context,
	args *dap.SourceArguments,
) (*dap.SourceResponseBody, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.SourceRequest{
		Request:   srv.newRequest("source"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.SourceResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return &r.Body, nil
}

// Evaluate evaluates an expression.
func (m *Manager) Evaluate(
	ctx context.Context,
	args *dap.EvaluateArguments,
) (*dap.EvaluateResponseBody, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.EvaluateRequest{
		Request:   srv.newRequest("evaluate"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.EvaluateResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return &r.Body, nil
}

// SetExpression assigns a value to an expression.
func (m *Manager) SetExpression(
	ctx context.Context,
	args *dap.SetExpressionArguments,
) (*dap.SetExpressionResponseBody, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.SetExpressionRequest{
		Request:   srv.newRequest("setExpression"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.SetExpressionResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return &r.Body, nil
}

// Completions provides completion suggestions.
func (m *Manager) Completions(
	ctx context.Context,
	args *dap.CompletionsArguments,
) ([]dap.CompletionItem, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.CompletionsRequest{
		Request:   srv.newRequest("completions"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.CompletionsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return r.Body.Targets, nil
}

// ExceptionInfo retrieves exception details.
func (m *Manager) ExceptionInfo(
	ctx context.Context,
	args *dap.ExceptionInfoArguments,
) (*dap.ExceptionInfoResponseBody, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.ExceptionInfoRequest{
		Request:   srv.newRequest("exceptionInfo"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.ExceptionInfoResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return &r.Body, nil
}

// Modules retrieves loaded modules.
func (m *Manager) Modules(
	ctx context.Context,
	args *dap.ModulesArguments,
) (*dap.ModulesResponseBody, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.ModulesRequest{
		Request:   srv.newRequest("modules"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.ModulesResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return &r.Body, nil
}

// LoadedSources retrieves all loaded sources.
func (m *Manager) LoadedSources(
	ctx context.Context,
) ([]dap.Source, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.LoadedSourcesRequest{
		Request: srv.newRequest("loadedSources"),
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.LoadedSourcesResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return r.Body.Sources, nil
}

// ReadMemory reads bytes from memory.
func (m *Manager) ReadMemory(
	ctx context.Context,
	args *dap.ReadMemoryArguments,
) (*dap.ReadMemoryResponseBody, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.ReadMemoryRequest{
		Request:   srv.newRequest("readMemory"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.ReadMemoryResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return &r.Body, nil
}

// WriteMemory writes bytes to memory.
func (m *Manager) WriteMemory(
	ctx context.Context,
	args *dap.WriteMemoryArguments,
) (*dap.WriteMemoryResponseBody, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.WriteMemoryRequest{
		Request:   srv.newRequest("writeMemory"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.WriteMemoryResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return &r.Body, nil
}

// Disassemble returns disassembled instructions.
func (m *Manager) Disassemble(
	ctx context.Context,
	args *dap.DisassembleArguments,
) ([]dap.DisassembledInstruction, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.DisassembleRequest{
		Request:   srv.newRequest("disassemble"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.DisassembleResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return r.Body.Instructions, nil
}

// GotoTargets returns possible goto targets.
func (m *Manager) GotoTargets(
	ctx context.Context,
	args *dap.GotoTargetsArguments,
) ([]dap.GotoTarget, error) {
	srv, err := m.activeServer()
	if err != nil {
		return nil, err
	}
	req := &dap.GotoTargetsRequest{
		Request:   srv.newRequest("gotoTargets"),
		Arguments: *args,
	}
	resp, err := srv.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*dap.GotoTargetsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	return r.Body.Targets, nil
}

// Goto sets execution to continue from a target.
func (m *Manager) Goto(
	ctx context.Context,
	args *dap.GotoArguments,
) error {
	srv, err := m.activeServer()
	if err != nil {
		return err
	}
	req := &dap.GotoRequest{
		Request:   srv.newRequest("goto"),
		Arguments: *args,
	}
	_, err = srv.sendRequest(ctx, req)
	return err
}

func (m *Manager) serverForLaunch(
	program string,
) (*debugServer, error) {
	if program != "" {
		srv, err := m.serverForFile(program)
		if err == nil {
			return srv, nil
		}
	}
	return m.activeServer()
}
