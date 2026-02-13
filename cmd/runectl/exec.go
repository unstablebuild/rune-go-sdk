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
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/unstablebuild/rune-go-sdk/api/extensionapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"golang.org/x/term"
)

func newExecCmd(a *app) *cobra.Command {
	var (
		format string
		dir    string
	)

	cmd := &cobra.Command{
		Use:   "exec [flags] -- <command> [args...]",
		Short: "Execute a command using the workspace executor",
		Long: `Execute a command using the workspace executor.

Use -- to separate exec flags from the command and its arguments.

Examples:
  runectl exec -- ls -la
  runectl exec --dir /tmp -- cat file.txt
  runectl exec -F json -- echo hello
  runectl exec -- bash`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			w, err := a.getWorkspace()
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			executor := w.Executor(ctx)
			defer func() { _ = executor.Close() }()

			return runInteractive(ctx, w, executor, args, dir)
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)
	cmd.Flags().StringVarP(
		&dir, "dir", "d", "",
		"Working directory for the command",
	)
	return cmd
}

func runInteractive(
	ctx context.Context,
	w *extensionapi.Workspace,
	executor workspaceapi.Executor,
	args []string,
	dir string,
) error {
	terminal := w.Terminal(ctx)
	pty, err := terminal.StartPty()
	if err != nil {
		return fmt.Errorf("start pty: %w", err)
	}
	defer func() { _ = pty.Master.Close() }()
	defer func() { _ = pty.Slave.Close() }()

	resetPtySize := func() {
		width, height, err := term.GetSize(int(os.Stdin.Fd()))
		if err == nil {
			err := terminal.SetPtySize(pty, width, height)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not propagate pty size: %v", err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "warning: could not get pty size: %v", err)
		}
	}

	isTerm := term.IsTerminal(int(os.Stdin.Fd()))
	if isTerm {
		stdin := int(os.Stdin.Fd())
		oldState, err := term.MakeRaw(stdin)
		if err == nil {
			defer func() { _ = term.Restore(stdin, oldState) }()
		}
		resetPtySize()
	}

	execCmd := workspaceapi.Cmd{
		Path:   args[0],
		Args:   args[1:],
		Dir:    dir,
		Stdin:  pty.Slave,
		Stdout: pty.Slave,
		Stderr: pty.Slave,
		SysProcAttr: &syscall.SysProcAttr{
			Setsid:  true,
			Setctty: true,
		},
	}

	ch := make(chan error, 1)
	watcher := workspaceapi.ChanProcessWatcher(ch)
	execCmd.Watcher = watcher

	_, err = executor.Start(ctx, execCmd)
	if err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	var done sync.Mutex
	done.Lock()
	go func() { _, _ = io.Copy(pty.Master, os.Stdin) }()
	go func() {
		defer done.Unlock()
		_, _ = io.Copy(os.Stdout, pty.Master)
	}()

	// monitor local pty size changes so we can propagate to remote pty
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)

	var procErr error
	for {
		select {
		case procErr = <-ch:
		case <-sig:
			fmt.Fprintf(os.Stderr, "received SIGWINCH")
			resetPtySize()
			continue
		}
		break
	}

	// unblock i/o goroutines
	_ = pty.Master.Close()
	done.Lock()
	return procErr
}
