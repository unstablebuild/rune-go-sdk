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

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-dap"
	"github.com/spf13/cobra"
	"github.com/unstablebuild/rune-go-sdk/api/debugapi"
	"github.com/unstablebuild/rune-go-sdk/handler/repl"
	"github.com/unstablebuild/rune-go-sdk/handler/repl/debugger"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// debugAdapter describes how to launch a DAP server for a
// given language.
type debugAdapter struct {
	LangID  string `json:"langID"`
	Command string `json:"command"`
}

// debugAdapters maps language names to their DAP adapter
// configuration. The {addr} placeholder in Command is
// replaced by the IDE with the actual listen address.
var debugAdapters = map[string]debugAdapter{
	"go": {LangID: "go", Command: "dlv dap --listen={addr}"},
}

func supportedLanguages() string {
	langs := make([]string, 0, len(debugAdapters))
	for k := range debugAdapters {
		langs = append(langs, k)
	}
	return strings.Join(langs, ", ")
}

func newDebuggerCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debugger",
		Short: "Interactive debugger REPL",
	}
	cmd.AddCommand(
		newDebuggerLaunchCmd(a),
		newDebuggerAttachCmd(a),
	)
	return cmd
}

func newDebuggerLaunchCmd(a *app) *cobra.Command {
	var (
		stopOnEntry bool
		cwd         string
		envVars     []string
	)

	cmd := &cobra.Command{
		Use:   "launch [flags] <language> <program> [args...]",
		Short: "Launch a program under the debugger",
		Long: `Launch a program under the debugger and start an interactive REPL.

Examples:
  runectl debugger launch go ./myapp
  runectl debugger launch go --stop-on-entry ./myapp arg1 arg2
  runectl debugger launch go --cwd /tmp --env FOO=bar ./myapp`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			env := parseEnvVars(envVars)
			launchArgs := debugapi.LaunchRequestArguments{
				Program:     args[1],
				Args:        args[2:],
				Cwd:         cwd,
				Env:         env,
				StopOnEntry: stopOnEntry,
			}
			return runDebugger(cmd.Context(), a, args[0], func(
				ctx context.Context, dbg debugapi.Debugger,
			) error {
				return dbg.Launch(ctx, launchArgs)
			})
		},
	}

	cmd.Flags().BoolVar(
		&stopOnEntry, "stop-on-entry", false,
		"Stop immediately after launch",
	)
	cmd.Flags().StringVar(
		&cwd, "cwd", "",
		"Working directory for the debuggee",
	)
	cmd.Flags().StringArrayVar(
		&envVars, "env", nil,
		"Environment variables (KEY=VALUE)",
	)
	return cmd
}

func newDebuggerAttachCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach <language> <pid>",
		Short: "Attach the debugger to a running process",
		Long: `Attach the debugger to a running process and start an interactive REPL.

Examples:
  runectl debugger attach go 12345`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pid, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid pid: %w", err)
			}
			attachArgs := debugapi.AttachRequestArguments{
				PID: pid,
			}
			return runDebugger(cmd.Context(), a, args[0], func(
				ctx context.Context, dbg debugapi.Debugger,
			) error {
				return dbg.Attach(ctx, attachArgs)
			})
		},
	}
	return cmd
}

func runDebugger(
	ctx context.Context,
	a *app,
	language string,
	start func(context.Context, debugapi.Debugger) error,
) error {
	adapter, ok := debugAdapters[language]
	if !ok {
		return fmt.Errorf(
			"language %q unsupported (supported: %s)",
			language, supportedLanguages(),
		)
	}

	initOpts, err := json.Marshal(adapter)
	if err != nil {
		return fmt.Errorf("marshal init options: %w", err)
	}

	w, err := a.getWorkspace()
	if err != nil {
		return err
	}

	dbg := w.Debugger(ctx)

	caps, err := dbg.Initialize(ctx, &debugapi.InitializeRequestArguments{
		InitializeRequestArguments: dap.InitializeRequestArguments{
			ClientID:        "runectl",
			ClientName:      "runectl debugger",
			AdapterID:       language,
			LinesStartAt1:   true,
			ColumnsStartAt1: true,
			PathFormat:      "path",
		},
		InitializeOptions: initOpts,
	})
	if err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	if err := start(ctx, dbg); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	if err := dbg.ConfigurationDone(ctx); err != nil {
		return fmt.Errorf("configuration done: %w", err)
	}

	handler := debugger.New(dbg, caps)
	if err := handler.SelectFirstThread(ctx); err != nil {
		slog.Debug("select initial thread", "err", err)
	}
	r := repl.New(
		handler,
		term.ScheduleNextTick,
		term.FuncInterrupter(func(context.Context) error {
			if !term.PublishEvent(term.Event{Type: term.EventInterrupt}) {
				return errors.New("could not publish interrupt")
			}
			return nil
		}),
		repl.WithPrompt("(debug) "),
	)

	tuiErr := tui.Run(r)

	r.Wait()

	// Use a fresh context with timeout for cleanup since the parent
	// context may have been cancelled when the TUI exited.
	disconnectCtx, cancel := context.WithTimeout(
		context.Background(), 5*time.Second,
	)
	defer cancel()
	if err := dbg.Disconnect(disconnectCtx, &dap.DisconnectArguments{
		TerminateDebuggee: true,
	}); err != nil {
		slog.Debug("disconnect", "err", err)
	}

	return tuiErr
}

func parseEnvVars(envVars []string) map[string]string {
	if len(envVars) == 0 {
		return nil
	}
	env := make(map[string]string, len(envVars))
	for _, e := range envVars {
		k, v, _ := strings.Cut(e, "=")
		env[k] = v
	}
	return env
}
