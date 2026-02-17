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
)

// Debugger defines the Debug Adapter Protocol operations.
// All methods correspond to DAP requests as defined in the specification.
//
// See: https://microsoft.github.io/debug-adapter-protocol/specification
type Debugger interface {
	// Initialize configures the debug adapter with client capabilities
	// and retrieves the adapter's capabilities.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Initialize
	Initialize(ctx context.Context, args *dap.InitializeRequestArguments) (*dap.Capabilities, error)

	// Launch starts the debuggee with or without debugging.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Launch
	Launch(ctx context.Context, args LaunchRequestArguments) error

	// Attach connects to an already running debuggee.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Attach
	Attach(ctx context.Context, args AttachRequestArguments) error

	// ConfigurationDone indicates that configuration is complete.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_ConfigurationDone
	ConfigurationDone(ctx context.Context) error

	// Disconnect ends the debug session.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Disconnect
	Disconnect(ctx context.Context, args *dap.DisconnectArguments) error

	// Terminate requests graceful termination of the debuggee.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Terminate
	Terminate(ctx context.Context, args *dap.TerminateArguments) error

	// Restart restarts the debug session.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Restart
	Restart(ctx context.Context) error

	// SetBreakpoints sets breakpoints for a source file.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_SetBreakpoints
	SetBreakpoints(ctx context.Context, args *dap.SetBreakpointsArguments) ([]dap.Breakpoint, error)

	// SetFunctionBreakpoints sets breakpoints on function names.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_SetFunctionBreakpoints
	SetFunctionBreakpoints(ctx context.Context, args *dap.SetFunctionBreakpointsArguments) ([]dap.Breakpoint, error)

	// SetExceptionBreakpoints configures exception breakpoints.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_SetExceptionBreakpoints
	SetExceptionBreakpoints(ctx context.Context, args *dap.SetExceptionBreakpointsArguments) ([]dap.Breakpoint, error)

	// Continue resumes execution of all threads.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Continue
	Continue(ctx context.Context, args *dap.ContinueArguments) (*dap.ContinueResponseBody, error)

	// Next executes one step (stepping over function calls).
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Next
	Next(ctx context.Context, args *dap.NextArguments) error

	// StepIn steps into a function call.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_StepIn
	StepIn(ctx context.Context, args *dap.StepInArguments) error

	// StepOut steps out of the current function.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_StepOut
	StepOut(ctx context.Context, args *dap.StepOutArguments) error

	// StepBack executes one backward step.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_StepBack
	StepBack(ctx context.Context, args *dap.StepBackArguments) error

	// ReverseContinue resumes backward execution.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_ReverseContinue
	ReverseContinue(ctx context.Context, args *dap.ReverseContinueArguments) error

	// Pause suspends execution.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Pause
	Pause(ctx context.Context, args *dap.PauseArguments) error

	// Threads retrieves all threads.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Threads
	Threads(ctx context.Context) ([]dap.Thread, error)

	// StackTrace returns the call stack for a thread.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_StackTrace
	StackTrace(ctx context.Context, args *dap.StackTraceArguments) (*dap.StackTraceResponseBody, error)

	// Scopes returns variable scopes for a stack frame.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Scopes
	Scopes(ctx context.Context, args *dap.ScopesArguments) ([]dap.Scope, error)

	// Variables retrieves child variables.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Variables
	Variables(ctx context.Context, args *dap.VariablesArguments) ([]dap.Variable, error)

	// SetVariable modifies a variable's value.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_SetVariable
	SetVariable(ctx context.Context, args *dap.SetVariableArguments) (*dap.SetVariableResponseBody, error)

	// Source retrieves source code.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Source
	Source(ctx context.Context, args *dap.SourceArguments) (*dap.SourceResponseBody, error)

	// Evaluate evaluates an expression.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Evaluate
	Evaluate(ctx context.Context, args *dap.EvaluateArguments) (*dap.EvaluateResponseBody, error)

	// SetExpression assigns a value to an expression.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_SetExpression
	SetExpression(ctx context.Context, args *dap.SetExpressionArguments) (*dap.SetExpressionResponseBody, error)

	// Completions provides completion suggestions.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Completions
	Completions(ctx context.Context, args *dap.CompletionsArguments) ([]dap.CompletionItem, error)

	// ExceptionInfo retrieves exception details.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_ExceptionInfo
	ExceptionInfo(ctx context.Context, args *dap.ExceptionInfoArguments) (*dap.ExceptionInfoResponseBody, error)

	// Modules retrieves loaded modules.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Modules
	Modules(ctx context.Context, args *dap.ModulesArguments) (*dap.ModulesResponseBody, error)

	// LoadedSources retrieves all loaded sources.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_LoadedSources
	LoadedSources(ctx context.Context) ([]dap.Source, error)

	// ReadMemory reads bytes from memory.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_ReadMemory
	ReadMemory(ctx context.Context, args *dap.ReadMemoryArguments) (*dap.ReadMemoryResponseBody, error)

	// WriteMemory writes bytes to memory.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_WriteMemory
	WriteMemory(ctx context.Context, args *dap.WriteMemoryArguments) (*dap.WriteMemoryResponseBody, error)

	// Disassemble returns disassembled instructions.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Disassemble
	Disassemble(ctx context.Context, args *dap.DisassembleArguments) ([]dap.DisassembledInstruction, error)

	// GotoTargets returns possible goto targets.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_GotoTargets
	GotoTargets(ctx context.Context, args *dap.GotoTargetsArguments) ([]dap.GotoTarget, error)

	// Goto sets execution to continue from a target.
	// DAP: https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Goto
	Goto(ctx context.Context, args *dap.GotoArguments) error

	// Events returns a channel for receiving debugger events.
	Events() <-chan dap.EventMessage
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
