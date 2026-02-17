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
// The interface follows the Debug Adapter Protocol (DAP) shape, enabling
// implementations for Go (Delve), Python (debugpy), Rust/C/C++ (LLDB/GDB),
// and other languages.
package debugapi

import (
	"context"
	"errors"
)

// Common errors returned by Debugger implementations.
var (
	// ErrNotSupported is returned when an operation is not supported by the debugger.
	ErrNotSupported = errors.New("operation not supported by this debugger")

	// ErrNotConnected is returned when the debugger is not connected.
	ErrNotConnected = errors.New("debugger not connected")

	// ErrAlreadyRunning is returned when trying to start while already running.
	ErrAlreadyRunning = errors.New("debuggee already running")

	// ErrNotRunning is returned when trying to pause while not running.
	ErrNotRunning = errors.New("debuggee not running")

	// ErrInvalidBreakpoint is returned when a breakpoint operation fails.
	ErrInvalidBreakpoint = errors.New("invalid breakpoint")
)

// Debugger is the core interface for debugging programs.
// It abstracts operations common across debuggers following the DAP shape.
type Debugger interface {
	// === Session Management ===

	// Initialize configures the debugger with client capabilities and returns
	// the debugger's capabilities.
	Initialize(ctx context.Context, args InitializeArguments) (*Capabilities, error)

	// Launch starts debugging a new process.
	Launch(ctx context.Context, args LaunchArguments) error

	// Attach connects to an already-running process.
	Attach(ctx context.Context, args AttachArguments) error

	// ConfigurationDone signals that all configuration is complete.
	ConfigurationDone(ctx context.Context) error

	// Disconnect ends the debug session.
	Disconnect(ctx context.Context, args DisconnectArguments) error

	// === Execution Control ===

	// Continue resumes execution of the given thread (or all threads if threadID is 0).
	Continue(ctx context.Context, threadID int64) error

	// Pause suspends execution of the given thread (or all threads if threadID is 0).
	Pause(ctx context.Context, threadID int64) error

	// Next executes one step, stepping over function calls.
	Next(ctx context.Context, threadID int64) error

	// StepIn steps into function calls.
	StepIn(ctx context.Context, threadID int64) error

	// StepOut continues until the current function returns.
	StepOut(ctx context.Context, threadID int64) error

	// === Breakpoints ===

	// SetBreakpoints sets breakpoints for a source file, replacing all previous
	// breakpoints in that file.
	SetBreakpoints(ctx context.Context, args SetBreakpointsArguments) ([]Breakpoint, error)

	// SetFunctionBreakpoints sets breakpoints on function names.
	SetFunctionBreakpoints(ctx context.Context, args SetFunctionBreakpointsArguments) ([]Breakpoint, error)

	// === State Inspection ===

	// Threads returns all threads in the debuggee.
	Threads(ctx context.Context) ([]Thread, error)

	// StackTrace returns the call stack for a thread.
	StackTrace(ctx context.Context, args StackTraceArguments) ([]StackFrame, int, error)

	// Scopes returns the variable scopes for a stack frame.
	Scopes(ctx context.Context, frameID int64) ([]Scope, error)

	// Variables returns variables within a scope or container.
	Variables(ctx context.Context, args VariablesArguments) ([]Variable, error)

	// Evaluate evaluates an expression in the context of a stack frame.
	Evaluate(ctx context.Context, args EvaluateArguments) (*Variable, error)

	// SetVariable modifies the value of a variable.
	SetVariable(ctx context.Context, args SetVariableArguments) (*Variable, error)

	// === Memory & Disassembly ===

	// ReadMemory reads bytes from the debuggee's memory.
	ReadMemory(ctx context.Context, args ReadMemoryArguments) ([]byte, error)

	// Disassemble returns disassembled instructions.
	Disassemble(ctx context.Context, args DisassembleArguments) ([]DisassembledInstruction, error)

	// === Events ===

	// Events returns a channel that receives debugger events (stopped, exited, etc.).
	Events() <-chan Event
}

// === Argument Types ===

// InitializeArguments contains arguments for Initialize.
type InitializeArguments struct {
	ClientID   string
	ClientName string
	AdapterID  string
	Locale     string
}

// LaunchArguments contains arguments for launching a debug session.
type LaunchArguments struct {
	Program     string            // Path to executable or script
	Args        []string          // Command-line arguments
	Cwd         string            // Working directory
	Env         map[string]string // Environment variables
	StopOnEntry bool              // Stop at program entry point
	NoDebug     bool              // Run without debugging
}

// AttachArguments contains arguments for attaching to a process.
type AttachArguments struct {
	PID     int    // Process ID to attach to
	Program string // Optional: path to executable (for symbols)
}

// DisconnectArguments contains arguments for Disconnect.
type DisconnectArguments struct {
	Restart           bool // Restart after disconnect
	TerminateDebuggee bool // Terminate the debuggee
	SuspendDebuggee   bool // Suspend the debuggee
}

// SetBreakpointsArguments contains arguments for SetBreakpoints.
type SetBreakpointsArguments struct {
	Source      Source             // Source file
	Breakpoints []SourceBreakpoint // Breakpoints to set
}

// SetFunctionBreakpointsArguments contains arguments for SetFunctionBreakpoints.
type SetFunctionBreakpointsArguments struct {
	Breakpoints []FunctionBreakpoint
}

// StackTraceArguments contains arguments for StackTrace.
type StackTraceArguments struct {
	ThreadID   int64
	StartFrame int
	Levels     int
}

// VariablesArguments contains arguments for Variables.
type VariablesArguments struct {
	VariablesReference int64
	Filter             string // "indexed", "named", or empty for all
	Start              int    // For indexed variables
	Count              int    // For indexed variables
}

// EvaluateArguments contains arguments for Evaluate.
type EvaluateArguments struct {
	Expression string
	FrameID    int64
	Context    string // "watch", "repl", "hover", "clipboard"
}

// SetVariableArguments contains arguments for SetVariable.
type SetVariableArguments struct {
	VariablesReference int64
	Name               string
	Value              string
}

// ReadMemoryArguments contains arguments for ReadMemory.
type ReadMemoryArguments struct {
	MemoryReference string
	Offset          int64
	Count           int
}

// DisassembleArguments contains arguments for Disassemble.
type DisassembleArguments struct {
	MemoryReference   string
	Offset            int64
	InstructionOffset int
	InstructionCount  int
}

// === Data Types ===

// Capabilities describes debugger capabilities.
type Capabilities struct {
	SupportsConfigurationDoneRequest bool
	SupportsFunctionBreakpoints      bool
	SupportsConditionalBreakpoints   bool
	SupportsHitConditionalBreakpoints bool
	SupportsEvaluateForHovers        bool
	SupportsStepBack                 bool
	SupportsSetVariable              bool
	SupportsRestartFrame             bool
	SupportsGotoTargetsRequest       bool
	SupportsStepInTargetsRequest     bool
	SupportsCompletionsRequest       bool
	SupportsModulesRequest           bool
	SupportsRestartRequest           bool
	SupportsExceptionOptions         bool
	SupportsValueFormattingOptions   bool
	SupportsExceptionInfoRequest     bool
	SupportTerminateDebuggee         bool
	SupportsDelayedStackTraceLoading bool
	SupportsLoadedSourcesRequest     bool
	SupportsLogPoints                bool
	SupportsTerminateThreadsRequest  bool
	SupportsSetExpression            bool
	SupportsTerminateRequest         bool
	SupportsDataBreakpoints          bool
	SupportsReadMemoryRequest        bool
	SupportsWriteMemoryRequest       bool
	SupportsDisassembleRequest       bool
	SupportsCancelRequest            bool
	SupportsBreakpointLocationsRequest bool
	SupportsClipboardContext         bool
	SupportsSingleThreadExecutionRequests bool
}

// Source represents a source file.
type Source struct {
	Name string
	Path string
}

// SourceBreakpoint represents a breakpoint in source code.
type SourceBreakpoint struct {
	Line         int
	Column       int
	Condition    string
	HitCondition string
	LogMessage   string
}

// FunctionBreakpoint represents a breakpoint on a function.
type FunctionBreakpoint struct {
	Name         string
	Condition    string
	HitCondition string
}

// Breakpoint represents an actual breakpoint set in the debuggee.
type Breakpoint struct {
	ID        int
	Verified  bool
	Message   string
	Source    *Source
	Line      int
	Column    int
	EndLine   int
	EndColumn int
}

// Thread represents a thread in the debuggee.
type Thread struct {
	ID   int64
	Name string
}

// StackFrame represents a stack frame.
type StackFrame struct {
	ID                          int64
	Name                        string
	Source                      *Source
	Line                        int
	Column                      int
	EndLine                     int
	EndColumn                   int
	InstructionPointerReference string
}

// Scope represents a variable scope.
type Scope struct {
	Name               string
	PresentationHint   string // "arguments", "locals", "registers"
	VariablesReference int64
	NamedVariables     int
	IndexedVariables   int
	Expensive          bool
}

// Variable represents a variable or expression result.
type Variable struct {
	Name               string
	Value              string
	Type               string
	VariablesReference int64
	NamedVariables     int
	IndexedVariables   int
	MemoryReference    string
}

// DisassembledInstruction represents a single disassembled instruction.
type DisassembledInstruction struct {
	Address          string
	InstructionBytes string
	Instruction      string
	Symbol           string
	Line             int
	Column           int
	EndLine          int
	EndColumn        int
}

// === Events ===

// EventType identifies the type of debugger event.
type EventType string

const (
	EventTypeStopped    EventType = "stopped"
	EventTypeContinued  EventType = "continued"
	EventTypeExited     EventType = "exited"
	EventTypeTerminated EventType = "terminated"
	EventTypeThread     EventType = "thread"
	EventTypeOutput     EventType = "output"
	EventTypeBreakpoint EventType = "breakpoint"
	EventTypeModule     EventType = "module"
)

// StopReason indicates why execution stopped.
type StopReason string

const (
	StopReasonStep              StopReason = "step"
	StopReasonBreakpoint        StopReason = "breakpoint"
	StopReasonException         StopReason = "exception"
	StopReasonPause             StopReason = "pause"
	StopReasonEntry             StopReason = "entry"
	StopReasonGoto              StopReason = "goto"
	StopReasonFunctionBreakpoint StopReason = "function breakpoint"
	StopReasonDataBreakpoint    StopReason = "data breakpoint"
)

// Event represents a debugger event.
type Event struct {
	Type EventType

	// For stopped events
	Reason            StopReason
	ThreadID          int64
	AllThreadsStopped bool
	HitBreakpointIDs  []int

	// For exited events
	ExitCode int

	// For thread events
	ThreadReason string // "started", "exited"

	// For output events
	Category string // "console", "stdout", "stderr", etc.
	Output   string

	// For breakpoint events
	BreakpointReason string // "changed", "new", "removed"
	Breakpoint       *Breakpoint
}
