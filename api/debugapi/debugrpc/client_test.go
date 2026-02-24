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

	"github.com/google/go-dap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/debugapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// === Tests (public) ===

func TestClient_Initialize(t *testing.T) {
	tests := []struct {
		name         string
		args         *dap.InitializeRequestArguments
		wantCaps     *dap.Capabilities
		wantCalled   bool
	}{
		{
			name: "basic initialization",
			args: &dap.InitializeRequestArguments{
				ClientID:   "test-client",
				ClientName: "Test Client",
				AdapterID:  "test-adapter",
				Locale:     "en-US",
			},
			wantCaps: &dap.Capabilities{
				SupportsConfigurationDoneRequest: true,
				SupportsFunctionBreakpoints:      true,
			},
			wantCalled: true,
		},
		{
			name: "minimal initialization",
			args: &dap.InitializeRequestArguments{
				AdapterID: "minimal-adapter",
			},
			wantCaps: &dap.Capabilities{
				SupportsConfigurationDoneRequest: true,
			},
			wantCalled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(t)
			env.mock.capabilities = tc.wantCaps
			ctx := context.Background()

			caps, err := env.client.Initialize(ctx, tc.args)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCalled, env.mock.initializeCalled)
			assert.Equal(t, tc.wantCaps.SupportsConfigurationDoneRequest, caps.SupportsConfigurationDoneRequest)
		})
	}
}

func TestClient_Launch(t *testing.T) {
	tests := []struct {
		name string
		args debugapi.LaunchRequestArguments
	}{
		{
			name: "launch with program only",
			args: debugapi.LaunchRequestArguments{
				Program: "/path/to/program",
			},
		},
		{
			name: "launch with all options",
			args: debugapi.LaunchRequestArguments{
				Program:     "/path/to/program",
				Args:        []string{"--flag", "value"},
				Cwd:         "/working/dir",
				Env:         map[string]string{"FOO": "bar"},
				StopOnEntry: true,
				NoDebug:     false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(t)
			ctx := context.Background()

			err := env.client.Launch(ctx, tc.args)
			require.NoError(t, err)
			assert.True(t, env.mock.launchCalled)
			assert.Equal(t, tc.args.Program, env.mock.lastLaunchArgs.Program)
		})
	}
}

func TestClient_Attach(t *testing.T) {
	tests := []struct {
		name string
		args debugapi.AttachRequestArguments
	}{
		{
			name: "attach by PID",
			args: debugapi.AttachRequestArguments{
				PID: 12345,
			},
		},
		{
			name: "attach by program",
			args: debugapi.AttachRequestArguments{
				Program: "/path/to/program",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(t)
			ctx := context.Background()

			err := env.client.Attach(ctx, tc.args)
			require.NoError(t, err)
			assert.True(t, env.mock.attachCalled)
		})
	}
}

func TestClient_SessionManagement(t *testing.T) {
	tests := []struct {
		name   string
		action func(*Client, context.Context) error
		verify func(*mockDebugger) bool
	}{
		{
			name:   "ConfigurationDone",
			action: func(c *Client, ctx context.Context) error { return c.ConfigurationDone(ctx) },
			verify: func(m *mockDebugger) bool { return m.configurationDoneCalled },
		},
		{
			name: "Disconnect",
			action: func(c *Client, ctx context.Context) error {
				return c.Disconnect(ctx, &dap.DisconnectArguments{Restart: true, TerminateDebuggee: true})
			},
			verify: func(m *mockDebugger) bool { return m.disconnectCalled },
		},
		{
			name: "Terminate",
			action: func(c *Client, ctx context.Context) error {
				return c.Terminate(ctx, &dap.TerminateArguments{Restart: false})
			},
			verify: func(m *mockDebugger) bool { return m.terminateCalled },
		},
		{
			name:   "Restart",
			action: func(c *Client, ctx context.Context) error { return c.Restart(ctx) },
			verify: func(m *mockDebugger) bool { return m.restartCalled },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(t)
			ctx := context.Background()

			err := tc.action(env.client, ctx)
			require.NoError(t, err)
			assert.True(t, tc.verify(env.mock))
		})
	}
}

func TestClient_ExecutionControl(t *testing.T) {
	tests := []struct {
		name     string
		action   func(*Client, context.Context) error
		verify   func(*mockDebugger) bool
	}{
		{
			name: "Continue",
			action: func(c *Client, ctx context.Context) error {
				_, err := c.Continue(ctx, &dap.ContinueArguments{ThreadId: 1})
				return err
			},
			verify: func(m *mockDebugger) bool { return m.continueCalled },
		},
		{
			name: "Pause",
			action: func(c *Client, ctx context.Context) error {
				return c.Pause(ctx, &dap.PauseArguments{ThreadId: 1})
			},
			verify: func(m *mockDebugger) bool { return m.pauseCalled },
		},
		{
			name: "Next",
			action: func(c *Client, ctx context.Context) error {
				return c.Next(ctx, &dap.NextArguments{ThreadId: 1})
			},
			verify: func(m *mockDebugger) bool { return m.nextCalled },
		},
		{
			name: "StepIn",
			action: func(c *Client, ctx context.Context) error {
				return c.StepIn(ctx, &dap.StepInArguments{ThreadId: 1})
			},
			verify: func(m *mockDebugger) bool { return m.stepInCalled },
		},
		{
			name: "StepOut",
			action: func(c *Client, ctx context.Context) error {
				return c.StepOut(ctx, &dap.StepOutArguments{ThreadId: 1})
			},
			verify: func(m *mockDebugger) bool { return m.stepOutCalled },
		},
		{
			name: "StepBack",
			action: func(c *Client, ctx context.Context) error {
				return c.StepBack(ctx, &dap.StepBackArguments{ThreadId: 1})
			},
			verify: func(m *mockDebugger) bool { return m.stepBackCalled },
		},
		{
			name: "ReverseContinue",
			action: func(c *Client, ctx context.Context) error {
				return c.ReverseContinue(ctx, &dap.ReverseContinueArguments{ThreadId: 1})
			},
			verify: func(m *mockDebugger) bool { return m.reverseContinueCalled },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(t)
			ctx := context.Background()

			err := tc.action(env.client, ctx)
			require.NoError(t, err)
			assert.True(t, tc.verify(env.mock))
		})
	}
}

func TestClient_Breakpoints(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(*mockDebugger)
		action     func(*Client, context.Context) (interface{}, error)
		wantLen    int
	}{
		{
			name: "SetBreakpoints",
			setupMock: func(m *mockDebugger) {
				m.breakpoints = []dap.Breakpoint{
					{Id: 1, Verified: true, Line: 10},
					{Id: 2, Verified: true, Line: 20},
				}
			},
			action: func(c *Client, ctx context.Context) (interface{}, error) {
				return c.SetBreakpoints(ctx, &dap.SetBreakpointsArguments{
					Source: dap.Source{Path: "/path/to/main.go"},
					Breakpoints: []dap.SourceBreakpoint{
						{Line: 10},
						{Line: 20},
					},
				})
			},
			wantLen: 2,
		},
		{
			name: "SetFunctionBreakpoints",
			setupMock: func(m *mockDebugger) {
				m.breakpoints = []dap.Breakpoint{
					{Id: 1, Verified: true},
				}
			},
			action: func(c *Client, ctx context.Context) (interface{}, error) {
				return c.SetFunctionBreakpoints(ctx, &dap.SetFunctionBreakpointsArguments{
					Breakpoints: []dap.FunctionBreakpoint{
						{Name: "main.foo"},
					},
				})
			},
			wantLen: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(t)
			tc.setupMock(env.mock)
			ctx := context.Background()

			result, err := tc.action(env.client, ctx)
			require.NoError(t, err)

			bps := result.([]dap.Breakpoint)
			assert.Len(t, bps, tc.wantLen)
		})
	}
}

func TestClient_StateInspection(t *testing.T) {
	t.Run("Threads", func(t *testing.T) {
		env := newTestEnv(t)
		env.mock.threads = []dap.Thread{
			{Id: 1, Name: "main"},
			{Id: 2, Name: "worker"},
		}
		ctx := context.Background()

		threads, err := env.client.Threads(ctx)
		require.NoError(t, err)
		assert.Len(t, threads, 2)
		assert.Equal(t, 1, threads[0].Id)
		assert.Equal(t, "main", threads[0].Name)
	})

	t.Run("StackTrace", func(t *testing.T) {
		env := newTestEnv(t)
		env.mock.stackTraceResponse = &dap.StackTraceResponseBody{
			StackFrames: []dap.StackFrame{
				{Id: 1, Name: "main.main", Line: 10},
				{Id: 2, Name: "main.foo", Line: 20},
			},
			TotalFrames: 2,
		}
		ctx := context.Background()

		resp, err := env.client.StackTrace(ctx, &dap.StackTraceArguments{ThreadId: 1})
		require.NoError(t, err)
		assert.Len(t, resp.StackFrames, 2)
		assert.Equal(t, 2, resp.TotalFrames)
	})

	t.Run("Scopes", func(t *testing.T) {
		env := newTestEnv(t)
		env.mock.scopes = []dap.Scope{
			{Name: "Locals", VariablesReference: 1000},
			{Name: "Arguments", VariablesReference: 1001},
		}
		ctx := context.Background()

		scopes, err := env.client.Scopes(ctx, &dap.ScopesArguments{FrameId: 1})
		require.NoError(t, err)
		assert.Len(t, scopes, 2)
		assert.Equal(t, "Locals", scopes[0].Name)
	})

	t.Run("Variables", func(t *testing.T) {
		env := newTestEnv(t)
		env.mock.variables = []dap.Variable{
			{Name: "x", Value: "42", Type: "int"},
			{Name: "y", Value: "hello", Type: "string"},
		}
		ctx := context.Background()

		vars, err := env.client.Variables(ctx, &dap.VariablesArguments{VariablesReference: 1000})
		require.NoError(t, err)
		assert.Len(t, vars, 2)
		assert.Equal(t, "x", vars[0].Name)
	})
}

func TestClient_Evaluate(t *testing.T) {
	tests := []struct {
		name       string
		args       *dap.EvaluateArguments
		wantResult string
	}{
		{
			name:       "evaluate expression",
			args:       &dap.EvaluateArguments{Expression: "x + y", FrameId: 1, Context: "repl"},
			wantResult: "100",
		},
		{
			name:       "evaluate watch",
			args:       &dap.EvaluateArguments{Expression: "myVar", FrameId: 1, Context: "watch"},
			wantResult: "100",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(t)
			env.mock.evaluateResponse = &dap.EvaluateResponseBody{
				Result: tc.wantResult,
				Type:   "int",
			}
			ctx := context.Background()

			result, err := env.client.Evaluate(ctx, tc.args)
			require.NoError(t, err)
			assert.Equal(t, tc.wantResult, result.Result)
		})
	}
}

func TestClient_Memory(t *testing.T) {
	t.Run("ReadMemory", func(t *testing.T) {
		env := newTestEnv(t)
		env.mock.readMemoryResponse = &dap.ReadMemoryResponseBody{
			Address: "0x1000",
			Data:    "SGVsbG8=", // "Hello" in base64
		}
		ctx := context.Background()

		resp, err := env.client.ReadMemory(ctx, &dap.ReadMemoryArguments{
			MemoryReference: "0x1000",
			Count:           5,
		})
		require.NoError(t, err)
		assert.Equal(t, "0x1000", resp.Address)
	})

	t.Run("WriteMemory", func(t *testing.T) {
		env := newTestEnv(t)
		env.mock.writeMemoryResponse = &dap.WriteMemoryResponseBody{
			BytesWritten: 5,
		}
		ctx := context.Background()

		resp, err := env.client.WriteMemory(ctx, &dap.WriteMemoryArguments{
			MemoryReference: "0x1000",
			Data:            "SGVsbG8=",
		})
		require.NoError(t, err)
		assert.Equal(t, 5, resp.BytesWritten)
	})
}

func TestClient_Disassemble(t *testing.T) {
	env := newTestEnv(t)
	env.mock.disassembleResponse = []dap.DisassembledInstruction{
		{Address: "0x1000", Instruction: "mov eax, 1"},
		{Address: "0x1004", Instruction: "ret"},
	}
	ctx := context.Background()

	instructions, err := env.client.Disassemble(ctx, &dap.DisassembleArguments{
		MemoryReference:  "0x1000",
		InstructionCount: 10,
	})
	require.NoError(t, err)
	assert.Len(t, instructions, 2)
	assert.Equal(t, "0x1000", instructions[0].Address)
}

// === Private test helpers ===

type testEnv struct {
	mock     *mockDebugger
	server   *grpc.Server
	client   *Client
	listener net.Listener
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	socketPath := filepath.Join(os.TempDir(), "debugapi_test.sock")
	_ = os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)

	mock := newMockDebugger()
	srv := grpc.NewServer()
	debugServer := NewServer(mock)
	debugServer.Register(srv)

	go func() {
		_ = srv.Serve(listener)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//nolint:staticcheck // grpc.DialContext is deprecated but still works for 1.x
	conn, err := grpc.DialContext(ctx, "unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), //nolint:staticcheck // deprecated but works for 1.x
	)
	require.NoError(t, err)

	client := NewClient(context.Background(), conn)

	t.Cleanup(func() {
		_ = client.Close()
		srv.GracefulStop()
		_ = listener.Close()
		_ = os.Remove(socketPath)
	})

	return &testEnv{
		mock:     mock,
		server:   srv,
		client:   client,
		listener: listener,
	}
}

type mockDebugger struct {
	initializeCalled        bool
	launchCalled            bool
	attachCalled            bool
	configurationDoneCalled bool
	disconnectCalled        bool
	terminateCalled         bool
	restartCalled           bool
	continueCalled          bool
	pauseCalled             bool
	nextCalled              bool
	stepInCalled            bool
	stepOutCalled           bool
	stepBackCalled          bool
	reverseContinueCalled   bool
	setBreakpointsCalled    bool
	threadsCalled           bool
	stackTraceCalled        bool
	scopesCalled            bool
	variablesCalled         bool
	evaluateCalled          bool
	readMemoryCalled        bool
	writeMemoryCalled       bool
	disassembleCalled       bool

	lastLaunchArgs debugapi.LaunchRequestArguments

	capabilities          *dap.Capabilities
	threads               []dap.Thread
	stackTraceResponse    *dap.StackTraceResponseBody
	scopes                []dap.Scope
	variables             []dap.Variable
	evaluateResponse      *dap.EvaluateResponseBody
	breakpoints           []dap.Breakpoint
	readMemoryResponse    *dap.ReadMemoryResponseBody
	writeMemoryResponse   *dap.WriteMemoryResponseBody
	disassembleResponse   []dap.DisassembledInstruction

	events chan dap.EventMessage
}

func newMockDebugger() *mockDebugger {
	return &mockDebugger{
		capabilities: &dap.Capabilities{
			SupportsConfigurationDoneRequest: true,
			SupportsFunctionBreakpoints:      true,
		},
		threads: []dap.Thread{
			{Id: 1, Name: "main"},
		},
		stackTraceResponse: &dap.StackTraceResponseBody{
			StackFrames: []dap.StackFrame{
				{Id: 1, Name: "main.main", Line: 10},
			},
			TotalFrames: 1,
		},
		scopes: []dap.Scope{
			{Name: "Locals", VariablesReference: 1000},
		},
		variables: []dap.Variable{
			{Name: "x", Value: "42", Type: "int"},
		},
		evaluateResponse: &dap.EvaluateResponseBody{
			Result: "100",
			Type:   "int",
		},
		breakpoints: []dap.Breakpoint{
			{Id: 1, Verified: true, Line: 10},
		},
		events: make(chan dap.EventMessage, 10),
	}
}

func (m *mockDebugger) Initialize(ctx context.Context, args *dap.InitializeRequestArguments) (*dap.Capabilities, error) {
	m.initializeCalled = true
	return m.capabilities, nil
}

func (m *mockDebugger) Launch(ctx context.Context, args debugapi.LaunchRequestArguments) error {
	m.launchCalled = true
	m.lastLaunchArgs = args
	return nil
}

func (m *mockDebugger) Attach(ctx context.Context, args debugapi.AttachRequestArguments) error {
	m.attachCalled = true
	return nil
}

func (m *mockDebugger) ConfigurationDone(ctx context.Context) error {
	m.configurationDoneCalled = true
	return nil
}

func (m *mockDebugger) Disconnect(ctx context.Context, args *dap.DisconnectArguments) error {
	m.disconnectCalled = true
	return nil
}

func (m *mockDebugger) Terminate(ctx context.Context, args *dap.TerminateArguments) error {
	m.terminateCalled = true
	return nil
}

func (m *mockDebugger) Restart(ctx context.Context) error {
	m.restartCalled = true
	return nil
}

func (m *mockDebugger) SetBreakpoints(ctx context.Context, args *dap.SetBreakpointsArguments) ([]dap.Breakpoint, error) {
	m.setBreakpointsCalled = true
	return m.breakpoints, nil
}

func (m *mockDebugger) SetFunctionBreakpoints(ctx context.Context, args *dap.SetFunctionBreakpointsArguments) ([]dap.Breakpoint, error) {
	return m.breakpoints, nil
}

func (m *mockDebugger) SetExceptionBreakpoints(ctx context.Context, args *dap.SetExceptionBreakpointsArguments) ([]dap.Breakpoint, error) {
	return m.breakpoints, nil
}

func (m *mockDebugger) Continue(ctx context.Context, args *dap.ContinueArguments) (*dap.ContinueResponseBody, error) {
	m.continueCalled = true
	return &dap.ContinueResponseBody{AllThreadsContinued: true}, nil
}

func (m *mockDebugger) Next(ctx context.Context, args *dap.NextArguments) error {
	m.nextCalled = true
	return nil
}

func (m *mockDebugger) StepIn(ctx context.Context, args *dap.StepInArguments) error {
	m.stepInCalled = true
	return nil
}

func (m *mockDebugger) StepOut(ctx context.Context, args *dap.StepOutArguments) error {
	m.stepOutCalled = true
	return nil
}

func (m *mockDebugger) StepBack(ctx context.Context, args *dap.StepBackArguments) error {
	m.stepBackCalled = true
	return nil
}

func (m *mockDebugger) ReverseContinue(ctx context.Context, args *dap.ReverseContinueArguments) error {
	m.reverseContinueCalled = true
	return nil
}

func (m *mockDebugger) Pause(ctx context.Context, args *dap.PauseArguments) error {
	m.pauseCalled = true
	return nil
}

func (m *mockDebugger) Threads(ctx context.Context) ([]dap.Thread, error) {
	m.threadsCalled = true
	return m.threads, nil
}

func (m *mockDebugger) StackTrace(ctx context.Context, args *dap.StackTraceArguments) (*dap.StackTraceResponseBody, error) {
	m.stackTraceCalled = true
	return m.stackTraceResponse, nil
}

func (m *mockDebugger) Scopes(ctx context.Context, args *dap.ScopesArguments) ([]dap.Scope, error) {
	m.scopesCalled = true
	return m.scopes, nil
}

func (m *mockDebugger) Variables(ctx context.Context, args *dap.VariablesArguments) ([]dap.Variable, error) {
	m.variablesCalled = true
	return m.variables, nil
}

func (m *mockDebugger) SetVariable(ctx context.Context, args *dap.SetVariableArguments) (*dap.SetVariableResponseBody, error) {
	return &dap.SetVariableResponseBody{Value: "99"}, nil
}

func (m *mockDebugger) Source(ctx context.Context, args *dap.SourceArguments) (*dap.SourceResponseBody, error) {
	return &dap.SourceResponseBody{Content: "source code"}, nil
}

func (m *mockDebugger) Evaluate(ctx context.Context, args *dap.EvaluateArguments) (*dap.EvaluateResponseBody, error) {
	m.evaluateCalled = true
	return m.evaluateResponse, nil
}

func (m *mockDebugger) SetExpression(ctx context.Context, args *dap.SetExpressionArguments) (*dap.SetExpressionResponseBody, error) {
	return &dap.SetExpressionResponseBody{Value: "new_value"}, nil
}

func (m *mockDebugger) Completions(ctx context.Context, args *dap.CompletionsArguments) ([]dap.CompletionItem, error) {
	return []dap.CompletionItem{{Label: "item1"}}, nil
}

func (m *mockDebugger) ExceptionInfo(ctx context.Context, args *dap.ExceptionInfoArguments) (*dap.ExceptionInfoResponseBody, error) {
	return &dap.ExceptionInfoResponseBody{ExceptionId: "err1"}, nil
}

func (m *mockDebugger) Modules(ctx context.Context, args *dap.ModulesArguments) (*dap.ModulesResponseBody, error) {
	return &dap.ModulesResponseBody{}, nil
}

func (m *mockDebugger) LoadedSources(ctx context.Context) ([]dap.Source, error) {
	return []dap.Source{}, nil
}

func (m *mockDebugger) ReadMemory(ctx context.Context, args *dap.ReadMemoryArguments) (*dap.ReadMemoryResponseBody, error) {
	m.readMemoryCalled = true
	return m.readMemoryResponse, nil
}

func (m *mockDebugger) WriteMemory(ctx context.Context, args *dap.WriteMemoryArguments) (*dap.WriteMemoryResponseBody, error) {
	m.writeMemoryCalled = true
	return m.writeMemoryResponse, nil
}

func (m *mockDebugger) Disassemble(ctx context.Context, args *dap.DisassembleArguments) ([]dap.DisassembledInstruction, error) {
	m.disassembleCalled = true
	return m.disassembleResponse, nil
}

func (m *mockDebugger) GotoTargets(ctx context.Context, args *dap.GotoTargetsArguments) ([]dap.GotoTarget, error) {
	return []dap.GotoTarget{}, nil
}

func (m *mockDebugger) Goto(ctx context.Context, args *dap.GotoArguments) error {
	return nil
}

func (m *mockDebugger) Events() <-chan dap.EventMessage {
	return m.events
}
