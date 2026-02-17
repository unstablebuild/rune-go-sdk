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
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/debugapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// mockDebugger is a mock implementation of debugapi.Debugger for testing.
type mockDebugger struct {
	// Tracks method calls for verification
	initializeCalled          bool
	launchCalled              bool
	attachCalled              bool
	configurationDoneCalled   bool
	disconnectCalled          bool
	continueCalled            bool
	pauseCalled               bool
	nextCalled                bool
	stepInCalled              bool
	stepOutCalled             bool
	setBreakpointsCalled      bool
	setFunctionBreakpointsCalled bool
	threadsCalled             bool
	stackTraceCalled          bool
	scopesCalled              bool
	variablesCalled           bool
	evaluateCalled            bool
	setVariableCalled         bool
	readMemoryCalled          bool
	disassembleCalled         bool

	// Store arguments for verification
	lastInitializeArgs          debugapi.InitializeArguments
	lastLaunchArgs              debugapi.LaunchArguments
	lastAttachArgs              debugapi.AttachArguments
	lastDisconnectArgs          debugapi.DisconnectArguments
	lastContinueThreadID        int64
	lastPauseThreadID           int64
	lastNextThreadID            int64
	lastStepInThreadID          int64
	lastStepOutThreadID         int64
	lastSetBreakpointsArgs      debugapi.SetBreakpointsArguments
	lastSetFunctionBreakpointsArgs debugapi.SetFunctionBreakpointsArguments
	lastStackTraceArgs          debugapi.StackTraceArguments
	lastScopesFrameID           int64
	lastVariablesArgs           debugapi.VariablesArguments
	lastEvaluateArgs            debugapi.EvaluateArguments
	lastSetVariableArgs         debugapi.SetVariableArguments
	lastReadMemoryArgs          debugapi.ReadMemoryArguments
	lastDisassembleArgs         debugapi.DisassembleArguments

	// Return values
	capabilities        *debugapi.Capabilities
	threads             []debugapi.Thread
	stackFrames         []debugapi.StackFrame
	totalFrames         int
	scopes              []debugapi.Scope
	variables           []debugapi.Variable
	evaluateResult      *debugapi.Variable
	setVariableResult   *debugapi.Variable
	breakpoints         []debugapi.Breakpoint
	functionBreakpoints []debugapi.Breakpoint
	memoryData          []byte
	instructions        []debugapi.DisassembledInstruction

	events chan debugapi.Event
}

func newMockDebugger() *mockDebugger {
	return &mockDebugger{
		capabilities: &debugapi.Capabilities{
			SupportsConfigurationDoneRequest: true,
			SupportsFunctionBreakpoints:      true,
			SupportsConditionalBreakpoints:   true,
			SupportsSetVariable:              true,
			SupportsReadMemoryRequest:        true,
			SupportsDisassembleRequest:       true,
		},
		threads: []debugapi.Thread{
			{ID: 1, Name: "main"},
			{ID: 2, Name: "worker"},
		},
		stackFrames: []debugapi.StackFrame{
			{ID: 1, Name: "main.main", Line: 10},
			{ID: 2, Name: "main.foo", Line: 20},
		},
		totalFrames: 2,
		scopes: []debugapi.Scope{
			{Name: "Locals", VariablesReference: 1000},
			{Name: "Arguments", VariablesReference: 1001},
		},
		variables: []debugapi.Variable{
			{Name: "x", Value: "42", Type: "int"},
			{Name: "y", Value: "hello", Type: "string"},
		},
		evaluateResult:    &debugapi.Variable{Name: "result", Value: "100", Type: "int"},
		setVariableResult: &debugapi.Variable{Name: "x", Value: "99", Type: "int"},
		breakpoints: []debugapi.Breakpoint{
			{ID: 1, Verified: true, Line: 10},
		},
		functionBreakpoints: []debugapi.Breakpoint{
			{ID: 2, Verified: true},
		},
		memoryData:   []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f},
		instructions: []debugapi.DisassembledInstruction{
			{Address: "0x1000", Instruction: "mov eax, 1"},
			{Address: "0x1004", Instruction: "ret"},
		},
		events: make(chan debugapi.Event, 10),
	}
}

func (m *mockDebugger) Initialize(ctx context.Context, args debugapi.InitializeArguments) (*debugapi.Capabilities, error) {
	m.initializeCalled = true
	m.lastInitializeArgs = args
	return m.capabilities, nil
}

func (m *mockDebugger) Launch(ctx context.Context, args debugapi.LaunchArguments) error {
	m.launchCalled = true
	m.lastLaunchArgs = args
	return nil
}

func (m *mockDebugger) Attach(ctx context.Context, args debugapi.AttachArguments) error {
	m.attachCalled = true
	m.lastAttachArgs = args
	return nil
}

func (m *mockDebugger) ConfigurationDone(ctx context.Context) error {
	m.configurationDoneCalled = true
	return nil
}

func (m *mockDebugger) Disconnect(ctx context.Context, args debugapi.DisconnectArguments) error {
	m.disconnectCalled = true
	m.lastDisconnectArgs = args
	return nil
}

func (m *mockDebugger) Continue(ctx context.Context, threadID int64) error {
	m.continueCalled = true
	m.lastContinueThreadID = threadID
	return nil
}

func (m *mockDebugger) Pause(ctx context.Context, threadID int64) error {
	m.pauseCalled = true
	m.lastPauseThreadID = threadID
	return nil
}

func (m *mockDebugger) Next(ctx context.Context, threadID int64) error {
	m.nextCalled = true
	m.lastNextThreadID = threadID
	return nil
}

func (m *mockDebugger) StepIn(ctx context.Context, threadID int64) error {
	m.stepInCalled = true
	m.lastStepInThreadID = threadID
	return nil
}

func (m *mockDebugger) StepOut(ctx context.Context, threadID int64) error {
	m.stepOutCalled = true
	m.lastStepOutThreadID = threadID
	return nil
}

func (m *mockDebugger) SetBreakpoints(ctx context.Context, args debugapi.SetBreakpointsArguments) ([]debugapi.Breakpoint, error) {
	m.setBreakpointsCalled = true
	m.lastSetBreakpointsArgs = args
	return m.breakpoints, nil
}

func (m *mockDebugger) SetFunctionBreakpoints(ctx context.Context, args debugapi.SetFunctionBreakpointsArguments) ([]debugapi.Breakpoint, error) {
	m.setFunctionBreakpointsCalled = true
	m.lastSetFunctionBreakpointsArgs = args
	return m.functionBreakpoints, nil
}

func (m *mockDebugger) Threads(ctx context.Context) ([]debugapi.Thread, error) {
	m.threadsCalled = true
	return m.threads, nil
}

func (m *mockDebugger) StackTrace(ctx context.Context, args debugapi.StackTraceArguments) ([]debugapi.StackFrame, int, error) {
	m.stackTraceCalled = true
	m.lastStackTraceArgs = args
	return m.stackFrames, m.totalFrames, nil
}

func (m *mockDebugger) Scopes(ctx context.Context, frameID int64) ([]debugapi.Scope, error) {
	m.scopesCalled = true
	m.lastScopesFrameID = frameID
	return m.scopes, nil
}

func (m *mockDebugger) Variables(ctx context.Context, args debugapi.VariablesArguments) ([]debugapi.Variable, error) {
	m.variablesCalled = true
	m.lastVariablesArgs = args
	return m.variables, nil
}

func (m *mockDebugger) Evaluate(ctx context.Context, args debugapi.EvaluateArguments) (*debugapi.Variable, error) {
	m.evaluateCalled = true
	m.lastEvaluateArgs = args
	return m.evaluateResult, nil
}

func (m *mockDebugger) SetVariable(ctx context.Context, args debugapi.SetVariableArguments) (*debugapi.Variable, error) {
	m.setVariableCalled = true
	m.lastSetVariableArgs = args
	return m.setVariableResult, nil
}

func (m *mockDebugger) ReadMemory(ctx context.Context, args debugapi.ReadMemoryArguments) ([]byte, error) {
	m.readMemoryCalled = true
	m.lastReadMemoryArgs = args
	return m.memoryData, nil
}

func (m *mockDebugger) Disassemble(ctx context.Context, args debugapi.DisassembleArguments) ([]debugapi.DisassembledInstruction, error) {
	m.disassembleCalled = true
	m.lastDisassembleArgs = args
	return m.instructions, nil
}

func (m *mockDebugger) Events() <-chan debugapi.Event {
	return m.events
}

// testEnv sets up a test environment with server and client over Unix socket.
type testEnv struct {
	mock     *mockDebugger
	server   *grpc.Server
	client   *Client
	listener net.Listener
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	// Create Unix socket
	socketPath := filepath.Join(os.TempDir(), "debugapi_test.sock")
	os.Remove(socketPath) // Clean up any previous socket

	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)

	// Create mock and server
	mock := newMockDebugger()
	srv := grpc.NewServer()
	debugServer := NewServer(mock)
	debugServer.Register(srv)

	// Start server in background
	go func() {
		_ = srv.Serve(listener)
	}()

	// Create client connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	require.NoError(t, err)

	client := NewClient(context.Background(), conn)

	t.Cleanup(func() {
		client.Close()
		srv.GracefulStop()
		listener.Close()
		os.Remove(socketPath)
	})

	return &testEnv{
		mock:     mock,
		server:   srv,
		client:   client,
		listener: listener,
	}
}

func TestClientServer_Initialize(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	args := debugapi.InitializeArguments{
		ClientID:   "test-client",
		ClientName: "Test Client",
		AdapterID:  "test-adapter",
		Locale:     "en-US",
	}

	caps, err := env.client.Initialize(ctx, args)
	require.NoError(t, err)

	assert.True(t, env.mock.initializeCalled)
	assert.Equal(t, args.ClientID, env.mock.lastInitializeArgs.ClientID)
	assert.Equal(t, args.ClientName, env.mock.lastInitializeArgs.ClientName)
	assert.Equal(t, args.AdapterID, env.mock.lastInitializeArgs.AdapterID)
	assert.Equal(t, args.Locale, env.mock.lastInitializeArgs.Locale)

	assert.True(t, caps.SupportsConfigurationDoneRequest)
	assert.True(t, caps.SupportsFunctionBreakpoints)
}

func TestClientServer_Launch(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	args := debugapi.LaunchArguments{
		Program:     "/path/to/program",
		Args:        []string{"--flag", "value"},
		Cwd:         "/working/dir",
		Env:         map[string]string{"FOO": "bar"},
		StopOnEntry: true,
		NoDebug:     false,
	}

	err := env.client.Launch(ctx, args)
	require.NoError(t, err)

	assert.True(t, env.mock.launchCalled)
	assert.Equal(t, args.Program, env.mock.lastLaunchArgs.Program)
	assert.Equal(t, args.Args, env.mock.lastLaunchArgs.Args)
	assert.Equal(t, args.Cwd, env.mock.lastLaunchArgs.Cwd)
	assert.Equal(t, args.Env, env.mock.lastLaunchArgs.Env)
	assert.Equal(t, args.StopOnEntry, env.mock.lastLaunchArgs.StopOnEntry)
}

func TestClientServer_Attach(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	args := debugapi.AttachArguments{
		PID:     12345,
		Program: "/path/to/program",
	}

	err := env.client.Attach(ctx, args)
	require.NoError(t, err)

	assert.True(t, env.mock.attachCalled)
	assert.Equal(t, args.PID, env.mock.lastAttachArgs.PID)
	assert.Equal(t, args.Program, env.mock.lastAttachArgs.Program)
}

func TestClientServer_ConfigurationDone(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	err := env.client.ConfigurationDone(ctx)
	require.NoError(t, err)

	assert.True(t, env.mock.configurationDoneCalled)
}

func TestClientServer_Disconnect(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	args := debugapi.DisconnectArguments{
		Restart:           true,
		TerminateDebuggee: true,
		SuspendDebuggee:   false,
	}

	err := env.client.Disconnect(ctx, args)
	require.NoError(t, err)

	assert.True(t, env.mock.disconnectCalled)
	assert.Equal(t, args.Restart, env.mock.lastDisconnectArgs.Restart)
	assert.Equal(t, args.TerminateDebuggee, env.mock.lastDisconnectArgs.TerminateDebuggee)
}

func TestClientServer_ExecutionControl(t *testing.T) {
	tests := []struct {
		name     string
		action   func(*Client, context.Context, int64) error
		verify   func(*mockDebugger) bool
		threadID int64
	}{
		{
			name:     "Continue",
			action:   func(c *Client, ctx context.Context, tid int64) error { return c.Continue(ctx, tid) },
			verify:   func(m *mockDebugger) bool { return m.continueCalled && m.lastContinueThreadID == 1 },
			threadID: 1,
		},
		{
			name:     "Pause",
			action:   func(c *Client, ctx context.Context, tid int64) error { return c.Pause(ctx, tid) },
			verify:   func(m *mockDebugger) bool { return m.pauseCalled && m.lastPauseThreadID == 2 },
			threadID: 2,
		},
		{
			name:     "Next",
			action:   func(c *Client, ctx context.Context, tid int64) error { return c.Next(ctx, tid) },
			verify:   func(m *mockDebugger) bool { return m.nextCalled && m.lastNextThreadID == 3 },
			threadID: 3,
		},
		{
			name:     "StepIn",
			action:   func(c *Client, ctx context.Context, tid int64) error { return c.StepIn(ctx, tid) },
			verify:   func(m *mockDebugger) bool { return m.stepInCalled && m.lastStepInThreadID == 4 },
			threadID: 4,
		},
		{
			name:     "StepOut",
			action:   func(c *Client, ctx context.Context, tid int64) error { return c.StepOut(ctx, tid) },
			verify:   func(m *mockDebugger) bool { return m.stepOutCalled && m.lastStepOutThreadID == 5 },
			threadID: 5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := setupTestEnv(t)
			ctx := context.Background()

			err := tc.action(env.client, ctx, tc.threadID)
			require.NoError(t, err)
			assert.True(t, tc.verify(env.mock))
		})
	}
}

func TestClientServer_SetBreakpoints(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	args := debugapi.SetBreakpointsArguments{
		Source: debugapi.Source{
			Name: "main.go",
			Path: "/path/to/main.go",
		},
		Breakpoints: []debugapi.SourceBreakpoint{
			{Line: 10, Condition: "x > 5"},
			{Line: 20, LogMessage: "value is {x}"},
		},
	}

	bps, err := env.client.SetBreakpoints(ctx, args)
	require.NoError(t, err)

	assert.True(t, env.mock.setBreakpointsCalled)
	assert.Equal(t, args.Source.Path, env.mock.lastSetBreakpointsArgs.Source.Path)
	assert.Len(t, env.mock.lastSetBreakpointsArgs.Breakpoints, 2)
	assert.Len(t, bps, 1)
	assert.Equal(t, 1, bps[0].ID)
	assert.True(t, bps[0].Verified)
}

func TestClientServer_SetFunctionBreakpoints(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	args := debugapi.SetFunctionBreakpointsArguments{
		Breakpoints: []debugapi.FunctionBreakpoint{
			{Name: "main.foo", Condition: "i == 0"},
			{Name: "main.bar"},
		},
	}

	bps, err := env.client.SetFunctionBreakpoints(ctx, args)
	require.NoError(t, err)

	assert.True(t, env.mock.setFunctionBreakpointsCalled)
	assert.Len(t, env.mock.lastSetFunctionBreakpointsArgs.Breakpoints, 2)
	assert.Len(t, bps, 1)
}

func TestClientServer_Threads(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	threads, err := env.client.Threads(ctx)
	require.NoError(t, err)

	assert.True(t, env.mock.threadsCalled)
	assert.Len(t, threads, 2)
	assert.Equal(t, int64(1), threads[0].ID)
	assert.Equal(t, "main", threads[0].Name)
	assert.Equal(t, int64(2), threads[1].ID)
	assert.Equal(t, "worker", threads[1].Name)
}

func TestClientServer_StackTrace(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	args := debugapi.StackTraceArguments{
		ThreadID:   1,
		StartFrame: 0,
		Levels:     20,
	}

	frames, total, err := env.client.StackTrace(ctx, args)
	require.NoError(t, err)

	assert.True(t, env.mock.stackTraceCalled)
	assert.Equal(t, args.ThreadID, env.mock.lastStackTraceArgs.ThreadID)
	assert.Equal(t, args.StartFrame, env.mock.lastStackTraceArgs.StartFrame)
	assert.Equal(t, args.Levels, env.mock.lastStackTraceArgs.Levels)
	assert.Len(t, frames, 2)
	assert.Equal(t, 2, total)
	assert.Equal(t, "main.main", frames[0].Name)
}

func TestClientServer_Scopes(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	scopes, err := env.client.Scopes(ctx, 1)
	require.NoError(t, err)

	assert.True(t, env.mock.scopesCalled)
	assert.Equal(t, int64(1), env.mock.lastScopesFrameID)
	assert.Len(t, scopes, 2)
	assert.Equal(t, "Locals", scopes[0].Name)
	assert.Equal(t, int64(1000), scopes[0].VariablesReference)
}

func TestClientServer_Variables(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	args := debugapi.VariablesArguments{
		VariablesReference: 1000,
		Filter:             "named",
		Start:              0,
		Count:              10,
	}

	vars, err := env.client.Variables(ctx, args)
	require.NoError(t, err)

	assert.True(t, env.mock.variablesCalled)
	assert.Equal(t, args.VariablesReference, env.mock.lastVariablesArgs.VariablesReference)
	assert.Len(t, vars, 2)
	assert.Equal(t, "x", vars[0].Name)
	assert.Equal(t, "42", vars[0].Value)
}

func TestClientServer_Evaluate(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	args := debugapi.EvaluateArguments{
		Expression: "x + y",
		FrameID:    1,
		Context:    "repl",
	}

	result, err := env.client.Evaluate(ctx, args)
	require.NoError(t, err)

	assert.True(t, env.mock.evaluateCalled)
	assert.Equal(t, args.Expression, env.mock.lastEvaluateArgs.Expression)
	assert.Equal(t, args.FrameID, env.mock.lastEvaluateArgs.FrameID)
	assert.Equal(t, args.Context, env.mock.lastEvaluateArgs.Context)
	assert.Equal(t, "result", result.Name)
	assert.Equal(t, "100", result.Value)
}

func TestClientServer_SetVariable(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	args := debugapi.SetVariableArguments{
		VariablesReference: 1000,
		Name:               "x",
		Value:              "99",
	}

	result, err := env.client.SetVariable(ctx, args)
	require.NoError(t, err)

	assert.True(t, env.mock.setVariableCalled)
	assert.Equal(t, args.VariablesReference, env.mock.lastSetVariableArgs.VariablesReference)
	assert.Equal(t, args.Name, env.mock.lastSetVariableArgs.Name)
	assert.Equal(t, args.Value, env.mock.lastSetVariableArgs.Value)
	assert.Equal(t, "99", result.Value)
}

func TestClientServer_ReadMemory(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	args := debugapi.ReadMemoryArguments{
		MemoryReference: "0x1000",
		Offset:          0,
		Count:           5,
	}

	data, err := env.client.ReadMemory(ctx, args)
	require.NoError(t, err)

	assert.True(t, env.mock.readMemoryCalled)
	assert.Equal(t, args.MemoryReference, env.mock.lastReadMemoryArgs.MemoryReference)
	assert.Equal(t, []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}, data)
}

func TestClientServer_Disassemble(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	args := debugapi.DisassembleArguments{
		MemoryReference:  "0x1000",
		InstructionCount: 10,
	}

	instructions, err := env.client.Disassemble(ctx, args)
	require.NoError(t, err)

	assert.True(t, env.mock.disassembleCalled)
	assert.Equal(t, args.MemoryReference, env.mock.lastDisassembleArgs.MemoryReference)
	assert.Len(t, instructions, 2)
	assert.Equal(t, "0x1000", instructions[0].Address)
	assert.Equal(t, "mov eax, 1", instructions[0].Instruction)
}
