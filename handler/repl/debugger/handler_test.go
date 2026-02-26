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

package debugger

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-dap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/debugapi"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/handler/repl"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

var _ repl.CommandHandler = (*Handler)(nil)

func collectStrings(
	t *testing.T, h *Handler, cmd repl.Command,
) ([]string, error) {
	t.Helper()
	ctx := context.Background()
	iter, err := h.HandleCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if iter == nil {
		return nil, nil
	}
	defer func() { _ = iter.Close() }()
	var lines []string
	for {
		item, ok := iter.Next(ctx)
		if !ok {
			break
		}
		lines = append(lines, renderResponsive(t, item))
	}
	return lines, iter.Err()
}

func renderResponsive(
	t *testing.T, r component.Responsive,
) string {
	t.Helper()
	const width = 120
	h := r.Height(width)
	if h < 1 {
		h = 1
	}
	r.Resize(width, h)
	s, ok := r.(fmt.Stringer)
	if !ok {
		t.Fatalf("component %T does not implement fmt.Stringer", r)
	}
	return s.String()
}

func TestNew(t *testing.T) {
	m := newMockDebugger()
	caps := &dap.Capabilities{
		SupportsConfigurationDoneRequest: true,
	}
	h := New(m, caps)
	assert.NotNil(t, h)
	assert.Equal(t, caps, h.capabilities)
	assert.NotEmpty(t, h.commands)
}

func TestSelectFirstThread(t *testing.T) {
	m := newMockDebugger()
	m.threadsResp = []dap.Thread{
		{Id: 7, Name: "main"},
		{Id: 8, Name: "worker"},
	}
	h := New(m, &dap.Capabilities{})
	assert.Equal(t, 0, h.threadID)

	err := h.SelectFirstThread(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 7, h.threadID)
}

func TestSelectFirstThreadEmpty(t *testing.T) {
	m := newMockDebugger()
	m.threadsResp = nil
	h := New(m, &dap.Capabilities{})

	err := h.SelectFirstThread(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, h.threadID)
}

func TestUnknownCommand(t *testing.T) {
	m := newMockDebugger()
	h := New(m, &dap.Capabilities{})
	_, err := h.HandleCommand(
		context.Background(),
		repl.Command{Name: "nonexistent"},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

func TestEmptyCommand(t *testing.T) {
	m := newMockDebugger()
	h := New(m, &dap.Capabilities{})
	iter, err := h.HandleCommand(
		context.Background(),
		repl.Command{},
	)
	require.NoError(t, err)
	defer func() { _ = iter.Close() }()
	_, ok := iter.Next(context.Background())
	assert.False(t, ok)
}

func TestContinue(t *testing.T) {
	m := newMockDebugger()
	m.continueResp = &dap.ContinueResponseBody{
		AllThreadsContinued: true,
	}
	h := New(m, &dap.Capabilities{})
	h.threadID = 1

	items, err := collectStrings(t, h, repl.Command{
		Name: "continue",
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "continuing", items[0])
	assert.True(t, m.continueCalled)
	assert.Equal(t, 1, m.lastContinueArgs.ThreadId)
}

func TestContinueAlias(t *testing.T) {
	m := newMockDebugger()
	m.continueResp = &dap.ContinueResponseBody{}
	h := New(m, &dap.Capabilities{})

	_, err := collectStrings(t, h, repl.Command{Name: "c"})
	require.NoError(t, err)
	assert.True(t, m.continueCalled)
}

func TestNext(t *testing.T) {
	m := newMockDebugger()
	m.stackTraceResp = &dap.StackTraceResponseBody{
		StackFrames: []dap.StackFrame{
			{Id: 1, Name: "main.foo", Line: 10,
				Source: &dap.Source{Path: "/a/main.go"}},
		},
	}
	h := New(m, &dap.Capabilities{})
	h.threadID = 2

	_, err := collectStrings(t, h, repl.Command{Name: "next"})
	require.NoError(t, err)
	assert.True(t, m.nextCalled)
	assert.Equal(t, 2, m.lastNextArgs.ThreadId)
}

func TestStep(t *testing.T) {
	m := newMockDebugger()
	m.stackTraceResp = &dap.StackTraceResponseBody{
		StackFrames: []dap.StackFrame{
			{Id: 1, Name: "main.bar", Line: 20,
				Source: &dap.Source{Path: "/a/main.go"}},
		},
	}
	h := New(m, &dap.Capabilities{})

	_, err := collectStrings(t, h, repl.Command{Name: "step"})
	require.NoError(t, err)
	assert.True(t, m.stepInCalled)
}

func TestStepOut(t *testing.T) {
	m := newMockDebugger()
	m.stackTraceResp = &dap.StackTraceResponseBody{
		StackFrames: []dap.StackFrame{
			{Id: 1, Name: "main.baz", Line: 30,
				Source: &dap.Source{Path: "/a/main.go"}},
		},
	}
	h := New(m, &dap.Capabilities{})

	_, err := collectStrings(t, h, repl.Command{Name: "stepout"})
	require.NoError(t, err)
	assert.True(t, m.stepOutCalled)
}

func TestPause(t *testing.T) {
	m := newMockDebugger()
	h := New(m, &dap.Capabilities{})
	h.threadID = 3

	_, err := collectStrings(t, h, repl.Command{Name: "pause"})
	require.NoError(t, err)
	assert.True(t, m.pauseCalled)
	assert.Equal(t, 3, m.lastPauseArgs.ThreadId)
}

func TestRestart(t *testing.T) {
	m := newMockDebugger()
	h := New(m, &dap.Capabilities{})
	h.frameIndex = 5
	h.frameID = 10

	_, err := collectStrings(t, h, repl.Command{Name: "restart"})
	require.NoError(t, err)
	assert.True(t, m.restartCalled)
	assert.Equal(t, 0, h.frameIndex)
	assert.Equal(t, 0, h.frameID)
}

func TestBreak(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
		wantBP  bool
	}{
		{
			name:    "no args",
			args:    nil,
			wantErr: "usage:",
		},
		{
			name:    "invalid location",
			args:    []string{"nocolon"},
			wantErr: "invalid location",
		},
		{
			name:    "invalid line",
			args:    []string{"main.go:abc"},
			wantErr: "invalid line",
		},
		{
			name:   "valid breakpoint",
			args:   []string{"main.go:42"},
			wantBP: true,
		},
		{
			name:   "breakpoint with condition",
			args:   []string{"main.go:10", "x", ">", "5"},
			wantBP: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := newMockDebugger()
			m.setBreakpointsResp = []dap.Breakpoint{
				{Id: 1, Verified: true, Line: 42,
					Source: &dap.Source{Path: "main.go"}},
			}
			h := New(m, &dap.Capabilities{})

			_, err := collectStrings(t, h, repl.Command{
				Name: "break", Args: tc.args,
			})
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			if tc.wantBP {
				assert.True(t, m.setBreakpointsCalled)
				assert.NotEmpty(t, h.breakpoints)
			}
		})
	}
}

func TestBreakpoints(t *testing.T) {
	m := newMockDebugger()
	h := New(m, &dap.Capabilities{})

	// Empty
	items, err := collectStrings(t, h, repl.Command{Name: "breakpoints"})
	require.NoError(t, err)
	assert.Len(t, items, 1)

	// With breakpoints
	h.breakpoints["main.go"] = []trackedBreakpoint{
		{Breakpoint: dap.Breakpoint{Id: 1, Verified: true, Line: 10,
			Source: &dap.Source{Path: "main.go"}}},
		{Breakpoint: dap.Breakpoint{Id: 2, Verified: false, Line: 20,
			Source: &dap.Source{Path: "main.go"}}},
	}
	items, err = collectStrings(t, h, repl.Command{Name: "bp"})
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestClear(t *testing.T) {
	m := newMockDebugger()
	m.setBreakpointsResp = []dap.Breakpoint{
		{Id: 2, Verified: true, Line: 20,
			Source: &dap.Source{Path: "main.go"}},
	}
	h := New(m, &dap.Capabilities{})
	h.breakpoints["main.go"] = []trackedBreakpoint{
		{Breakpoint: dap.Breakpoint{Id: 1, Verified: true, Line: 10}},
		{Breakpoint: dap.Breakpoint{Id: 2, Verified: true, Line: 20}},
	}

	_, err := collectStrings(t, h, repl.Command{
		Name: "clear", Args: []string{"1"},
	})
	require.NoError(t, err)
	assert.True(t, m.setBreakpointsCalled)
}

func TestClearNotFound(t *testing.T) {
	m := newMockDebugger()
	h := New(m, &dap.Capabilities{})

	_, err := collectStrings(t, h, repl.Command{
		Name: "clear", Args: []string{"99"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestClearAll(t *testing.T) {
	m := newMockDebugger()
	h := New(m, &dap.Capabilities{})
	h.breakpoints["a.go"] = []trackedBreakpoint{
		{Breakpoint: dap.Breakpoint{Id: 1}},
	}
	h.breakpoints["b.go"] = []trackedBreakpoint{
		{Breakpoint: dap.Breakpoint{Id: 2}},
	}

	_, err := collectStrings(t, h, repl.Command{Name: "clearall"})
	require.NoError(t, err)
	assert.Empty(t, h.breakpoints)
}

func TestCondition(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing args",
			args:    nil,
			wantErr: "usage:",
		},
		{
			name:    "missing expression",
			args:    []string{"1"},
			wantErr: "usage:",
		},
		{
			name:    "invalid id",
			args:    []string{"abc", "x > 5"},
			wantErr: "invalid breakpoint id",
		},
		{
			name:    "breakpoint not found",
			args:    []string{"99", "x > 5"},
			wantErr: "not found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := newMockDebugger()
			h := New(m, &dap.Capabilities{})

			_, err := collectStrings(t, h, repl.Command{
				Name: "condition", Args: tc.args,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}

	t.Run("applies condition", func(t *testing.T) {
		m := newMockDebugger()
		m.setBreakpointsResp = []dap.Breakpoint{
			{Id: 1, Verified: true, Line: 10,
				Source: &dap.Source{Path: "main.go"}},
		}
		h := New(m, &dap.Capabilities{})
		h.breakpoints["main.go"] = []trackedBreakpoint{
			{Breakpoint: dap.Breakpoint{Id: 1, Verified: true, Line: 10,
				Source: &dap.Source{Path: "main.go"}}},
		}

		_, err := collectStrings(t, h, repl.Command{
			Name: "condition", Args: []string{"1", "x", ">", "5"},
		})
		require.NoError(t, err)
		assert.True(t, m.setBreakpointsCalled)

		// Verify the adapter received the condition.
		require.Len(t, m.lastSetBreakpointsArgs.Breakpoints, 1)
		assert.Equal(t, "x > 5",
			m.lastSetBreakpointsArgs.Breakpoints[0].Condition)

		// Verify tracked breakpoint stores the condition.
		require.Len(t, h.breakpoints["main.go"], 1)
		assert.Equal(t, "x > 5", h.breakpoints["main.go"][0].condition)
	})
}

func TestPrint(t *testing.T) {
	m := newMockDebugger()
	m.evaluateResp = &dap.EvaluateResponseBody{
		Result: "42",
		Type:   "int",
	}
	h := New(m, &dap.Capabilities{})
	h.frameID = 5

	items, err := collectStrings(t, h, repl.Command{
		Name: "print", Args: []string{"x", "+", "1"},
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "42 (int)", items[0])
	assert.Equal(t, "x + 1", m.lastEvaluateArgs.Expression)
	assert.Equal(t, 5, m.lastEvaluateArgs.FrameId)
}

func TestLocals(t *testing.T) {
	m := newMockDebugger()
	m.scopesResp = []dap.Scope{
		{Name: "Locals", VariablesReference: 10},
		{Name: "Arguments", VariablesReference: 20},
	}
	m.variablesResp = []dap.Variable{
		{Name: "x", Value: "1", Type: "int"},
		{Name: "y", Value: "hello", Type: "string"},
	}
	h := New(m, &dap.Capabilities{})

	items, err := collectStrings(t, h, repl.Command{Name: "locals"})
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "x = 1 (int)", items[0])
	assert.Equal(t, "y = hello (string)", items[1])
}

func TestArgs(t *testing.T) {
	m := newMockDebugger()
	m.scopesResp = []dap.Scope{
		{Name: "Arguments", VariablesReference: 20},
	}
	m.variablesResp = []dap.Variable{
		{Name: "n", Value: "5", Type: "int"},
	}
	h := New(m, &dap.Capabilities{})

	items, err := collectStrings(t, h, repl.Command{Name: "args"})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "n = 5 (int)", items[0])
}

func TestThreads(t *testing.T) {
	m := newMockDebugger()
	m.threadsResp = []dap.Thread{
		{Id: 1, Name: "main"},
		{Id: 2, Name: "worker"},
	}
	h := New(m, &dap.Capabilities{})
	h.threadID = 1

	items, err := collectStrings(t, h, repl.Command{Name: "threads"})
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "* Thread 1: main", items[0])
	assert.Equal(t, "  Thread 2: worker", items[1])
}

func TestThread(t *testing.T) {
	m := newMockDebugger()
	m.stackTraceResp = &dap.StackTraceResponseBody{
		StackFrames: []dap.StackFrame{
			{Id: 10, Name: "worker.run", Line: 5,
				Source: &dap.Source{Path: "/a/worker.go"}},
		},
	}
	h := New(m, &dap.Capabilities{})

	_, err := collectStrings(t, h, repl.Command{
		Name: "thread", Args: []string{"2"},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, h.threadID)
	assert.Equal(t, 0, h.frameIndex)
}

func TestStack(t *testing.T) {
	m := newMockDebugger()
	m.stackTraceResp = &dap.StackTraceResponseBody{
		StackFrames: []dap.StackFrame{
			{Id: 1, Name: "main.foo", Line: 10},
			{Id: 2, Name: "main.bar", Line: 20},
			{Id: 3, Name: "runtime.main", Line: 100},
		},
	}
	h := New(m, &dap.Capabilities{})

	items, err := collectStrings(t, h, repl.Command{Name: "stack"})
	require.NoError(t, err)
	require.Len(t, items, 3)
	assert.Contains(t, items[0], "main.foo")
	assert.Contains(t, items[1], "main.bar")
	assert.Contains(t, items[2], "runtime.main")
}

func TestStackWithDepth(t *testing.T) {
	m := newMockDebugger()
	m.stackTraceResp = &dap.StackTraceResponseBody{
		StackFrames: []dap.StackFrame{
			{Id: 1, Name: "main.foo", Line: 10},
		},
	}
	h := New(m, &dap.Capabilities{})

	_, err := collectStrings(t, h, repl.Command{
		Name: "stack", Args: []string{"1"},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, m.lastStackTraceArgs.Levels)
}

func TestFrame(t *testing.T) {
	m := newMockDebugger()
	m.stackTraceResp = &dap.StackTraceResponseBody{
		StackFrames: []dap.StackFrame{
			{Id: 1, Name: "main.foo", Line: 10,
				Source: &dap.Source{Path: "/a/main.go"}},
			{Id: 2, Name: "main.bar", Line: 20,
				Source: &dap.Source{Path: "/a/main.go"}},
		},
	}
	h := New(m, &dap.Capabilities{})

	_, err := collectStrings(t, h, repl.Command{
		Name: "frame", Args: []string{"1"},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, h.frameIndex)
	assert.Equal(t, 2, h.frameID)
}

func TestFrameOutOfRange(t *testing.T) {
	m := newMockDebugger()
	m.stackTraceResp = &dap.StackTraceResponseBody{
		StackFrames: []dap.StackFrame{
			{Id: 1, Name: "main.foo", Line: 10},
		},
	}
	h := New(m, &dap.Capabilities{})

	_, err := collectStrings(t, h, repl.Command{
		Name: "frame", Args: []string{"5"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")
}

func TestUpDown(t *testing.T) {
	m := newMockDebugger()
	m.stackTraceResp = &dap.StackTraceResponseBody{
		StackFrames: []dap.StackFrame{
			{Id: 1, Name: "a", Line: 1,
				Source: &dap.Source{Path: "/a.go"}},
			{Id: 2, Name: "b", Line: 2,
				Source: &dap.Source{Path: "/a.go"}},
			{Id: 3, Name: "c", Line: 3,
				Source: &dap.Source{Path: "/a.go"}},
		},
	}
	h := New(m, &dap.Capabilities{})

	// up 2
	_, err := collectStrings(t, h, repl.Command{
		Name: "up", Args: []string{"2"},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, h.frameIndex)

	// down 1
	_, err = collectStrings(t, h, repl.Command{
		Name: "down",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, h.frameIndex)
}

func TestDownClamps(t *testing.T) {
	m := newMockDebugger()
	m.stackTraceResp = &dap.StackTraceResponseBody{
		StackFrames: []dap.StackFrame{
			{Id: 1, Name: "a", Line: 1,
				Source: &dap.Source{Path: "/a.go"}},
		},
	}
	h := New(m, &dap.Capabilities{})
	h.frameIndex = 0

	_, err := collectStrings(t, h, repl.Command{
		Name: "down", Args: []string{"10"},
	})
	require.NoError(t, err)
	assert.Equal(t, 0, h.frameIndex)
}

func TestUpNonPositive(t *testing.T) {
	m := newMockDebugger()
	h := New(m, &dap.Capabilities{})

	_, err := collectStrings(t, h, repl.Command{
		Name: "up", Args: []string{"0"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "positive")

	_, err = collectStrings(t, h, repl.Command{
		Name: "up", Args: []string{"-1"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "positive")
}

func TestDownNonPositive(t *testing.T) {
	m := newMockDebugger()
	h := New(m, &dap.Capabilities{})

	_, err := collectStrings(t, h, repl.Command{
		Name: "down", Args: []string{"0"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "positive")
}

func TestHelp(t *testing.T) {
	m := newMockDebugger()
	h := New(m, &dap.Capabilities{})

	items, err := collectStrings(t, h, repl.Command{Name: "help"})
	require.NoError(t, err)
	assert.True(t, len(items) > 10, "expected many help lines")
}

func TestHelpSpecific(t *testing.T) {
	m := newMockDebugger()
	h := New(m, &dap.Capabilities{})

	items, err := collectStrings(t, h, repl.Command{
		Name: "help", Args: []string{"continue"},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, items)
}

func TestHelpUnknown(t *testing.T) {
	m := newMockDebugger()
	h := New(m, &dap.Capabilities{})

	_, err := collectStrings(t, h, repl.Command{
		Name: "help", Args: []string{"bogus"},
	})
	assert.Error(t, err)
}

func TestComplete(t *testing.T) {
	m := newMockDebugger()
	h := New(m, &dap.Capabilities{})

	iter, err := h.Complete(context.Background(), "con", nil)
	require.NoError(t, err)
	defer func() { _ = iter.Close() }()
	results, err := iterator.ToSlice(context.Background(), iter)
	require.NoError(t, err)
	assert.Contains(t, results, "continue")
	assert.Contains(t, results, "condition")
}

func TestCompleteEmpty(t *testing.T) {
	m := newMockDebugger()
	h := New(m, &dap.Capabilities{})

	iter, err := h.Complete(context.Background(), "", nil)
	require.NoError(t, err)
	defer func() { _ = iter.Close() }()
	results, err := iterator.ToSlice(context.Background(), iter)
	require.NoError(t, err)
	assert.True(t, len(results) > 10)
}

func TestCompleteWithArgs(t *testing.T) {
	m := newMockDebugger()
	h := New(m, &dap.Capabilities{})

	iter, err := h.Complete(
		context.Background(), "break", []string{"main.go:10"},
	)
	require.NoError(t, err)
	defer func() { _ = iter.Close() }()
	results, err := iterator.ToSlice(context.Background(), iter)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestParseFileLocation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantFile string
		wantLine int
		wantErr  bool
	}{
		{
			name:     "valid",
			input:    "main.go:42",
			wantFile: "main.go",
			wantLine: 42,
		},
		{
			name:     "with path",
			input:    "/foo/bar/baz.go:100",
			wantFile: "/foo/bar/baz.go",
			wantLine: 100,
		},
		{
			name:    "no colon",
			input:   "main.go",
			wantErr: true,
		},
		{
			name:    "bad line",
			input:   "main.go:abc",
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			file, line, err := parseFileLocation(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantFile, file)
			assert.Equal(t, tc.wantLine, line)
		})
	}
}

func TestWhatis(t *testing.T) {
	m := newMockDebugger()
	m.evaluateResp = &dap.EvaluateResponseBody{
		Result: "42",
		Type:   "int",
	}
	h := New(m, &dap.Capabilities{})

	items, err := collectStrings(t, h, repl.Command{
		Name: "whatis", Args: []string{"x"},
	})
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "hover", string(m.lastEvaluateArgs.Context))
}

func TestSet(t *testing.T) {
	m := newMockDebugger()
	m.scopesResp = []dap.Scope{
		{Name: "Locals", VariablesReference: 10},
	}
	m.setVariableResp = &dap.SetVariableResponseBody{
		Value: "99",
	}
	h := New(m, &dap.Capabilities{})

	items, err := collectStrings(t, h, repl.Command{
		Name: "set", Args: []string{"x", "=", "99"},
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "x = 99", items[0])
	assert.True(t, m.setVariableCalled)
}

// === Mock Debugger ===

type mockDebugger struct {
	debugapi.Debugger

	continueCalled    bool
	lastContinueArgs  dap.ContinueArguments
	continueResp      *dap.ContinueResponseBody
	nextCalled        bool
	lastNextArgs      dap.NextArguments
	stepInCalled      bool
	stepOutCalled     bool
	pauseCalled       bool
	lastPauseArgs     dap.PauseArguments
	restartCalled     bool
	disconnectCalled  bool
	setBreakpointsCalled    bool
	lastSetBreakpointsArgs dap.SetBreakpointsArguments
	evaluateCalled         bool
	lastEvaluateArgs  dap.EvaluateArguments
	evaluateResp      *dap.EvaluateResponseBody
	setVariableCalled bool
	setVariableResp   *dap.SetVariableResponseBody

	scopesResp        []dap.Scope
	variablesResp     []dap.Variable
	threadsResp       []dap.Thread
	stackTraceResp    *dap.StackTraceResponseBody
	lastStackTraceArgs dap.StackTraceArguments
	setBreakpointsResp []dap.Breakpoint
	sourceResp        *dap.SourceResponseBody
	loadedSourcesResp []dap.Source
	modulesResp       *dap.ModulesResponseBody
}

func newMockDebugger() *mockDebugger {
	return &mockDebugger{
		continueResp: &dap.ContinueResponseBody{},
		stackTraceResp: &dap.StackTraceResponseBody{
			StackFrames: []dap.StackFrame{
				{Id: 1, Name: "main.main", Line: 1,
					Source: &dap.Source{Path: "main.go"}},
			},
		},
		evaluateResp: &dap.EvaluateResponseBody{Result: "nil"},
		modulesResp:  &dap.ModulesResponseBody{},
	}
}

func (m *mockDebugger) Initialize(
	_ context.Context, _ *debugapi.InitializeRequestArguments,
) (*dap.Capabilities, error) {
	return &dap.Capabilities{}, nil
}

func (m *mockDebugger) Launch(
	_ context.Context, _ debugapi.LaunchRequestArguments,
) error {
	return nil
}

func (m *mockDebugger) Attach(
	_ context.Context, _ debugapi.AttachRequestArguments,
) error {
	return nil
}

func (m *mockDebugger) ConfigurationDone(_ context.Context) error {
	return nil
}

func (m *mockDebugger) Disconnect(
	_ context.Context, _ *dap.DisconnectArguments,
) error {
	m.disconnectCalled = true
	return nil
}

func (m *mockDebugger) Terminate(
	_ context.Context, _ *dap.TerminateArguments,
) error {
	return nil
}

func (m *mockDebugger) Restart(_ context.Context) error {
	m.restartCalled = true
	return nil
}

func (m *mockDebugger) SetBreakpoints(
	_ context.Context, args *dap.SetBreakpointsArguments,
) ([]dap.Breakpoint, error) {
	m.setBreakpointsCalled = true
	m.lastSetBreakpointsArgs = *args
	return m.setBreakpointsResp, nil
}

func (m *mockDebugger) SetFunctionBreakpoints(
	_ context.Context, _ *dap.SetFunctionBreakpointsArguments,
) ([]dap.Breakpoint, error) {
	return nil, nil
}

func (m *mockDebugger) SetExceptionBreakpoints(
	_ context.Context, _ *dap.SetExceptionBreakpointsArguments,
) ([]dap.Breakpoint, error) {
	return nil, nil
}

func (m *mockDebugger) Continue(
	_ context.Context, args *dap.ContinueArguments,
) (*dap.ContinueResponseBody, error) {
	m.continueCalled = true
	m.lastContinueArgs = *args
	return m.continueResp, nil
}

func (m *mockDebugger) Next(
	_ context.Context, args *dap.NextArguments,
) error {
	m.nextCalled = true
	m.lastNextArgs = *args
	return nil
}

func (m *mockDebugger) StepIn(
	_ context.Context, _ *dap.StepInArguments,
) error {
	m.stepInCalled = true
	return nil
}

func (m *mockDebugger) StepOut(
	_ context.Context, _ *dap.StepOutArguments,
) error {
	m.stepOutCalled = true
	return nil
}

func (m *mockDebugger) StepBack(
	_ context.Context, _ *dap.StepBackArguments,
) error {
	return nil
}

func (m *mockDebugger) ReverseContinue(
	_ context.Context, _ *dap.ReverseContinueArguments,
) error {
	return nil
}

func (m *mockDebugger) Pause(
	_ context.Context, args *dap.PauseArguments,
) error {
	m.pauseCalled = true
	m.lastPauseArgs = *args
	return nil
}

func (m *mockDebugger) Threads(
	_ context.Context,
) ([]dap.Thread, error) {
	return m.threadsResp, nil
}

func (m *mockDebugger) StackTrace(
	_ context.Context, args *dap.StackTraceArguments,
) (*dap.StackTraceResponseBody, error) {
	m.lastStackTraceArgs = *args
	return m.stackTraceResp, nil
}

func (m *mockDebugger) Scopes(
	_ context.Context, _ *dap.ScopesArguments,
) ([]dap.Scope, error) {
	return m.scopesResp, nil
}

func (m *mockDebugger) Variables(
	_ context.Context, _ *dap.VariablesArguments,
) ([]dap.Variable, error) {
	return m.variablesResp, nil
}

func (m *mockDebugger) SetVariable(
	_ context.Context, _ *dap.SetVariableArguments,
) (*dap.SetVariableResponseBody, error) {
	m.setVariableCalled = true
	return m.setVariableResp, nil
}

func (m *mockDebugger) Source(
	_ context.Context, _ *dap.SourceArguments,
) (*dap.SourceResponseBody, error) {
	if m.sourceResp != nil {
		return m.sourceResp, nil
	}
	return &dap.SourceResponseBody{Content: "line1\nline2\n"}, nil
}

func (m *mockDebugger) Evaluate(
	_ context.Context, args *dap.EvaluateArguments,
) (*dap.EvaluateResponseBody, error) {
	m.evaluateCalled = true
	m.lastEvaluateArgs = *args
	return m.evaluateResp, nil
}

func (m *mockDebugger) SetExpression(
	_ context.Context, _ *dap.SetExpressionArguments,
) (*dap.SetExpressionResponseBody, error) {
	return nil, nil
}

func (m *mockDebugger) Completions(
	_ context.Context, _ *dap.CompletionsArguments,
) ([]dap.CompletionItem, error) {
	return nil, nil
}

func (m *mockDebugger) ExceptionInfo(
	_ context.Context, _ *dap.ExceptionInfoArguments,
) (*dap.ExceptionInfoResponseBody, error) {
	return nil, nil
}

func (m *mockDebugger) Modules(
	_ context.Context, _ *dap.ModulesArguments,
) (*dap.ModulesResponseBody, error) {
	return m.modulesResp, nil
}

func (m *mockDebugger) LoadedSources(
	_ context.Context,
) ([]dap.Source, error) {
	return m.loadedSourcesResp, nil
}

func (m *mockDebugger) ReadMemory(
	_ context.Context, _ *dap.ReadMemoryArguments,
) (*dap.ReadMemoryResponseBody, error) {
	return nil, nil
}

func (m *mockDebugger) WriteMemory(
	_ context.Context, _ *dap.WriteMemoryArguments,
) (*dap.WriteMemoryResponseBody, error) {
	return nil, nil
}

func (m *mockDebugger) Disassemble(
	_ context.Context, _ *dap.DisassembleArguments,
) ([]dap.DisassembledInstruction, error) {
	return nil, nil
}

func (m *mockDebugger) GotoTargets(
	_ context.Context, _ *dap.GotoTargetsArguments,
) ([]dap.GotoTarget, error) {
	return nil, nil
}

func (m *mockDebugger) Goto(
	_ context.Context, _ *dap.GotoArguments,
) error {
	return nil
}
