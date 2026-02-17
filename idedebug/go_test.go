package idedebug

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/google/go-dap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unstablebuild/rune-go-sdk/api/debugapi"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

func TestE2E(t *testing.T) {
	t.Parallel()
	dlvBin := findDlv(t)
	tmpDir := setupTestWorkspace(t, "go")

	uri := makeURI(t, "file://"+tmpDir)
	mainPath := filepath.Join(tmpDir, "main.go")

	// Line number for "sum := Add(x, y)" in go/main.go
	const breakpointLine = 13

	tests := []struct {
		name string
		fn   func(t *testing.T, mgr *Manager)
	}{
		{
			name: "Threads",
			fn: func(t *testing.T, mgr *Manager) {
				threads, err := mgr.Threads(
					t.Context(),
				)
				require.NoError(t, err)
				require.NotEmpty(t, threads)
			},
		},
		{
			name: "StackTrace",
			fn: func(t *testing.T, mgr *Manager) {
				threads, err := mgr.Threads(
					t.Context(),
				)
				require.NoError(t, err)
				require.NotEmpty(t, threads)

				st, err := mgr.StackTrace(
					t.Context(),
					&dap.StackTraceArguments{
						ThreadId: threads[0].Id,
					},
				)
				require.NoError(t, err)
				require.NotEmpty(t, st.StackFrames)
				assert.Contains(t,
					st.StackFrames[0].Source.Path,
					"main.go",
				)
				assert.Equal(t,
					breakpointLine,
					st.StackFrames[0].Line,
				)
			},
		},
		{
			name: "Scopes",
			fn: func(t *testing.T, mgr *Manager) {
				threads, err := mgr.Threads(
					t.Context(),
				)
				require.NoError(t, err)
				require.NotEmpty(t, threads)

				st, err := mgr.StackTrace(
					t.Context(),
					&dap.StackTraceArguments{
						ThreadId: threads[0].Id,
					},
				)
				require.NoError(t, err)
				require.NotEmpty(t, st.StackFrames)

				scopes, err := mgr.Scopes(
					t.Context(),
					&dap.ScopesArguments{
						FrameId: st.StackFrames[0].Id,
					},
				)
				require.NoError(t, err)
				require.NotEmpty(t, scopes)
				assert.Equal(t,
					"Locals", scopes[0].Name,
				)
			},
		},
		{
			name: "Variables",
			fn: func(t *testing.T, mgr *Manager) {
				threads, err := mgr.Threads(
					t.Context(),
				)
				require.NoError(t, err)
				require.NotEmpty(t, threads)

				st, err := mgr.StackTrace(
					t.Context(),
					&dap.StackTraceArguments{
						ThreadId: threads[0].Id,
					},
				)
				require.NoError(t, err)
				require.NotEmpty(t, st.StackFrames)

				scopes, err := mgr.Scopes(
					t.Context(),
					&dap.ScopesArguments{
						FrameId: st.StackFrames[0].Id,
					},
				)
				require.NoError(t, err)
				require.NotEmpty(t, scopes)

				vars, err := mgr.Variables(
					t.Context(),
					&dap.VariablesArguments{
						VariablesReference: scopes[0].
							VariablesReference,
					},
				)
				require.NoError(t, err)
				require.NotEmpty(t, vars)

				varMap := make(map[string]string)
				for _, v := range vars {
					varMap[v.Name] = v.Value
				}
				assert.Equal(t, "10", varMap["x"])
				assert.Equal(t, "20", varMap["y"])
			},
		},
		{
			name: "Evaluate",
			fn: func(t *testing.T, mgr *Manager) {
				threads, err := mgr.Threads(
					t.Context(),
				)
				require.NoError(t, err)
				require.NotEmpty(t, threads)

				st, err := mgr.StackTrace(
					t.Context(),
					&dap.StackTraceArguments{
						ThreadId: threads[0].Id,
					},
				)
				require.NoError(t, err)
				require.NotEmpty(t, st.StackFrames)

				result, err := mgr.Evaluate(
					t.Context(),
					&dap.EvaluateArguments{
						Expression: "x + y",
						FrameId:    st.StackFrames[0].Id,
					},
				)
				require.NoError(t, err)
				assert.Equal(t, "30", result.Result)
			},
		},
		{
			name: "ContinueAndTerminate",
			fn: func(t *testing.T, mgr *Manager) {
				_, err := mgr.Continue(
					t.Context(),
					&dap.ContinueArguments{
						ThreadId: 1,
					},
				)
				require.NoError(t, err)
				waitForEvent(
					t, mgr.Events(), "terminated",
				)
			},
		},
	}

	t.Run("lazy init via Handle EventTypeOpen",
		func(t *testing.T) {
			t.Parallel()
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					executor := newTestExecutor(t)
					mgr := New(
						uri,
						executor,
						&stubPkgManager{bin: dlvBin},
						Config{MaxRetries: 1},
					)

					mainWSURI, err :=
						workspaceapi.ParseURI(
							"file://" + mainPath,
						)
					require.NoError(t, err)

					mgr.Handle(
						context.Background(),
						textapi.Event{
							Type: textapi.EventTypeOpen,
							URI:  mainWSURI,
						},
					)

					waitForEvent(
						t, mgr.Events(), "initialized",
					)

					setupDebugSession(
						t, mgr, tmpDir, mainPath,
						breakpointLine,
					)
					tt.fn(t, mgr)
					require.NoError(t, mgr.Close())
				})
			}
		})

	t.Run("explicit init via Initialize",
		func(t *testing.T) {
			t.Parallel()
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					executor := newTestExecutor(t)
					mgr := New(
						uri,
						executor,
						&stubPkgManager{bin: dlvBin},
						Config{MaxRetries: 1},
					)

					caps, err := mgr.Initialize(
						context.Background(),
						&dap.InitializeRequestArguments{
							AdapterID: "go",
						},
					)
					require.NoError(t, err)
					require.NotNil(t, caps)

					waitForEvent(
						t, mgr.Events(), "initialized",
					)

					setupDebugSession(
						t, mgr, tmpDir, mainPath,
						breakpointLine,
					)
					tt.fn(t, mgr)
					require.NoError(t, mgr.Close())
				})
			}
		})
}

func setupDebugSession(
	t *testing.T,
	mgr *Manager,
	program string,
	mainPath string,
	breakpointLine int,
) {
	t.Helper()
	ctx := t.Context()

	err := mgr.Launch(ctx, debugapi.LaunchRequestArguments{
		Program: program,
	})
	require.NoError(t, err)

	bps, err := mgr.SetBreakpoints(
		ctx,
		&dap.SetBreakpointsArguments{
			Source: dap.Source{
				Path: mainPath,
			},
			Breakpoints: []dap.SourceBreakpoint{
				{Line: breakpointLine},
			},
		},
	)
	require.NoError(t, err)
	require.Len(t, bps, 1)
	assert.True(t, bps[0].Verified)

	err = mgr.ConfigurationDone(ctx)
	require.NoError(t, err)

	waitForEvent(t, mgr.Events(), "stopped")
}

func findDlv(t *testing.T) string {
	t.Helper()
	dlvBin, err := exec.LookPath("dlv")
	if err != nil {
		for _, p := range []string{
			filepath.Join(
				os.Getenv("HOME"),
				".rune", "bin", "dlv",
			),
			filepath.Join(
				os.Getenv("HOME"),
				"go", "bin", "dlv",
			),
		} {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		t.Skip(
			"dlv not found, skipping e2e test",
		)
	}
	return dlvBin
}

func setupTestWorkspace(
	t *testing.T, testdataDir string,
) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	if testdataDir == "" {
		return tmpDir
	}

	entries, err := os.ReadDir(testdataDir)
	require.NoError(t, err)
	for _, e := range entries {
		src := filepath.Join(testdataDir, e.Name())
		dst := filepath.Join(tmpDir, e.Name())
		data, err := os.ReadFile(src)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(
			dst, data, 0644,
		))
	}
	return tmpDir
}

func waitForEvent(
	t *testing.T,
	ch <-chan dap.EventMessage,
	eventType string,
) {
	t.Helper()
	timeout := time.After(30 * time.Second)
	for {
		select {
		case ev := <-ch:
			if ev.GetEvent().Event == eventType {
				return
			}
		case <-timeout:
			t.Fatalf(
				"timeout waiting for %s event",
				eventType,
			)
		}
	}
}

func makeURI(
	t *testing.T, uri string,
) workspaceapi.URI {
	ret, err := workspaceapi.ParseURI(uri)
	require.NoError(t, err)
	return ret
}

type stubPkgManager struct {
	bin string
}

func (p *stubPkgManager) LibDir(
	_ context.Context, _ string,
) (iterator.Iterator[string], error) {
	return iterator.FromSlice(
		[]string{p.bin},
	), nil
}

type testExecutor struct {
	mu        sync.Mutex
	processes map[int]*os.Process
	nextPid   int
}

func newTestExecutor(t *testing.T) *testExecutor {
	e := &testExecutor{
		processes: make(map[int]*os.Process),
	}
	t.Cleanup(func() {
		e.mu.Lock()
		defer e.mu.Unlock()
		for _, p := range e.processes {
			_ = p.Kill()
		}
	})
	return e
}

func (e *testExecutor) StartCommand(
	_ context.Context,
	cmd workspaceapi.Cmd,
) (workspaceapi.Pid, error) {
	c := exec.Command(cmd.Path, cmd.Args...)
	if cmd.Dir != "" {
		c.Dir = cmd.Dir
	}
	c.Env = cmd.Env
	c.Stdin = cmd.Stdin
	c.Stdout = cmd.Stdout
	if cmd.Stderr != nil {
		c.Stderr = cmd.Stderr
	}

	if err := c.Start(); err != nil {
		return 0, fmt.Errorf(
			"start command: %w", err,
		)
	}

	e.mu.Lock()
	e.nextPid++
	pid := workspaceapi.Pid(e.nextPid)
	e.processes[int(pid)] = c.Process
	e.mu.Unlock()

	if cmd.Watcher != nil {
		ch := cmd.Watcher.WatchProcess()
		go func() {
			err := c.Wait()
			if ch != nil {
				ch <- err
			}
		}()
	}

	return pid, nil
}

func (e *testExecutor) Signal(
	pid workspaceapi.Pid, sig syscall.Signal,
) error {
	e.mu.Lock()
	p, ok := e.processes[int(pid)]
	e.mu.Unlock()
	if !ok {
		return fmt.Errorf(
			"process not found: %d", pid,
		)
	}
	return p.Signal(sig)
}

func (e *testExecutor) Close() error {
	return nil
}
