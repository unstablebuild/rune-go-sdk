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

// Package debugapi defines a common interface for debuggers across languages.
// The interface follows the Debug Adapter Protocol (DAP) specification,
// using types from github.com/google/go-dap.
//
// See: https://microsoft.github.io/debug-adapter-protocol/specification
package debugapi

import (
	"context"
	"errors"

	"github.com/google/go-dap"
)

// ClientCapabilities describes the DAP client's capabilities that
// are forwarded into the adapter's Initialize request when a
// session is created. All fields are optional.
type ClientCapabilities struct {
	ClientID                     string
	ClientName                   string
	Locale                       string
	LinesStartAt1                bool
	ColumnsStartAt1              bool
	PathFormat                   string
	SupportsVariableType         bool
	SupportsVariablePaging       bool
	SupportsRunInTerminalRequest bool
	SupportsMemoryReferences     bool
	SupportsProgressReporting    bool
	SupportsInvalidatedEvent     bool
	SupportsMemoryEvent          bool
}

// Common errors returned by Debugger implementations.
var (
	// ErrNotSupported is returned when an operation is not supported.
	ErrNotSupported = errors.New("operation not supported by this debugger")

	// ErrNotConnected is returned when the debugger is not connected.
	ErrNotConnected = errors.New("debugger not connected")

	// ErrAlreadyRunning is returned when trying to start while running.
	ErrAlreadyRunning = errors.New("debuggee already running")

	// ErrNotRunning is returned when trying to pause while not running.
	ErrNotRunning = errors.New("debuggee not running")

	// ErrSessionNotFound is returned when a session with the given
	// ID does not exist.
	ErrSessionNotFound = errors.New("debug session not found")

	// ErrNoAdapterConfigured is returned from CreateSession when no
	// debug adapter is configured for the requested language.
	ErrNoAdapterConfigured = errors.New(
		"no debug adapter configured for language")
)

// EventSubscriber receives DAP events for a debug session. It is
// supplied to CreateSession and remains bound to the session's
// lifetime.
//
// OnEvent is called for every DAP event the adapter emits.
// OnClose is called exactly once when the session ends. Reason
// is one of "terminated", "disconnected", "adapter_failed",
// "canceled".
//
// Implementations must be safe to call from multiple goroutines
// and must not block for long — the transport serialises all
// events for the session on a single stream.
type EventSubscriber interface {
	OnEvent(ev dap.EventMessage)
	OnClose(reason string)
}

// Debugger defines the Debug Adapter Protocol operations. All
// methods (other than CreateSession) dispatch to the session
// identified by sessionID; CreateSession must be called first to
// obtain a sessionID.
//
// See: https://microsoft.github.io/debug-adapter-protocol/specification
type Debugger interface {
	// CreateSession starts a new debug session for the given
	// language, spawning the configured adapter and performing
	// the DAP Initialize handshake. The returned sessionID must
	// be passed to all subsequent methods. subscriber receives
	// DAP events for the session's lifetime; it is closed via
	// subscriber.OnClose when the session ends.
	CreateSession(ctx context.Context, langID string,
		client ClientCapabilities, subscriber EventSubscriber,
	) (sessionID string, caps *dap.Capabilities, err error)

	// Launch starts the debuggee with or without debugging.
	Launch(ctx context.Context, sessionID string, args LaunchRequestArguments) error

	// Attach connects to an already running debuggee.
	Attach(ctx context.Context, sessionID string, args AttachRequestArguments) error

	// ConfigurationDone indicates that configuration is complete.
	ConfigurationDone(ctx context.Context, sessionID string) error

	// Disconnect ends the debug session.
	Disconnect(ctx context.Context, sessionID string, args *dap.DisconnectArguments) error

	// Terminate requests graceful termination of the debuggee.
	Terminate(ctx context.Context, sessionID string, args *dap.TerminateArguments) error

	// Restart restarts the debug session.
	Restart(ctx context.Context, sessionID string) error

	// SetBreakpoints sets breakpoints for a source file.
	SetBreakpoints(ctx context.Context, sessionID string, args *dap.SetBreakpointsArguments) ([]dap.Breakpoint, error)

	// SetFunctionBreakpoints sets breakpoints on function names.
	SetFunctionBreakpoints(ctx context.Context, sessionID string, args *dap.SetFunctionBreakpointsArguments) ([]dap.Breakpoint, error)

	// SetExceptionBreakpoints configures exception breakpoints.
	SetExceptionBreakpoints(ctx context.Context, sessionID string, args *dap.SetExceptionBreakpointsArguments) ([]dap.Breakpoint, error)

	// Continue resumes execution of all threads.
	Continue(ctx context.Context, sessionID string, args *dap.ContinueArguments) (*dap.ContinueResponseBody, error)

	// Next executes one step (stepping over function calls).
	Next(ctx context.Context, sessionID string, args *dap.NextArguments) error

	// StepIn steps into a function call.
	StepIn(ctx context.Context, sessionID string, args *dap.StepInArguments) error

	// StepOut steps out of the current function.
	StepOut(ctx context.Context, sessionID string, args *dap.StepOutArguments) error

	// StepBack executes one backward step.
	StepBack(ctx context.Context, sessionID string, args *dap.StepBackArguments) error

	// ReverseContinue resumes backward execution.
	ReverseContinue(ctx context.Context, sessionID string, args *dap.ReverseContinueArguments) error

	// Pause suspends execution.
	Pause(ctx context.Context, sessionID string, args *dap.PauseArguments) error

	// Threads retrieves all threads.
	Threads(ctx context.Context, sessionID string) ([]dap.Thread, error)

	// StackTrace returns the call stack for a thread.
	StackTrace(ctx context.Context, sessionID string, args *dap.StackTraceArguments) (*dap.StackTraceResponseBody, error)

	// Scopes returns variable scopes for a stack frame.
	Scopes(ctx context.Context, sessionID string, args *dap.ScopesArguments) ([]dap.Scope, error)

	// Variables retrieves child variables.
	Variables(ctx context.Context, sessionID string, args *dap.VariablesArguments) ([]dap.Variable, error)

	// SetVariable modifies a variable's value.
	SetVariable(ctx context.Context, sessionID string, args *dap.SetVariableArguments) (*dap.SetVariableResponseBody, error)

	// Source retrieves source code.
	Source(ctx context.Context, sessionID string, args *dap.SourceArguments) (*dap.SourceResponseBody, error)

	// Evaluate evaluates an expression.
	Evaluate(ctx context.Context, sessionID string, args *dap.EvaluateArguments) (*dap.EvaluateResponseBody, error)

	// SetExpression assigns a value to an expression.
	SetExpression(ctx context.Context, sessionID string, args *dap.SetExpressionArguments) (*dap.SetExpressionResponseBody, error)

	// Completions provides completion suggestions.
	Completions(ctx context.Context, sessionID string, args *dap.CompletionsArguments) ([]dap.CompletionItem, error)

	// ExceptionInfo retrieves exception details.
	ExceptionInfo(ctx context.Context, sessionID string, args *dap.ExceptionInfoArguments) (*dap.ExceptionInfoResponseBody, error)

	// Modules retrieves loaded modules.
	Modules(ctx context.Context, sessionID string, args *dap.ModulesArguments) (*dap.ModulesResponseBody, error)

	// LoadedSources retrieves all loaded sources.
	LoadedSources(ctx context.Context, sessionID string) ([]dap.Source, error)

	// ReadMemory reads bytes from memory.
	ReadMemory(ctx context.Context, sessionID string, args *dap.ReadMemoryArguments) (*dap.ReadMemoryResponseBody, error)

	// WriteMemory writes bytes to memory.
	WriteMemory(ctx context.Context, sessionID string, args *dap.WriteMemoryArguments) (*dap.WriteMemoryResponseBody, error)

	// Disassemble returns disassembled instructions.
	Disassemble(ctx context.Context, sessionID string, args *dap.DisassembleArguments) ([]dap.DisassembledInstruction, error)

	// GotoTargets returns possible goto targets.
	GotoTargets(ctx context.Context, sessionID string, args *dap.GotoTargetsArguments) ([]dap.GotoTarget, error)

	// Goto sets execution to continue from a target.
	Goto(ctx context.Context, sessionID string, args *dap.GotoArguments) error
}

// LaunchRequestArguments contains arguments for Launch.
// These are implementation-specific and not defined by DAP.
type LaunchRequestArguments struct {
	Program     string            `json:"program"`
	Args        []string          `json:"args,omitempty"`
	Cwd         string            `json:"cwd,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	StopOnEntry bool              `json:"stopOnEntry,omitempty"`
	NoDebug     bool              `json:"noDebug,omitempty"`
}

// AttachRequestArguments contains arguments for Attach.
// These are implementation-specific and not defined by DAP.
type AttachRequestArguments struct {
	PID     int    `json:"pid,omitempty"`
	Program string `json:"program,omitempty"`
}
