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
	"fmt"

	"github.com/spf13/cobra"
)

func newResolveCmd(a *app) *cobra.Command {
	var (
		format string
		dir    string
	)

	cmd := &cobra.Command{
		Use:   "resolve [flags] <command>",
		Short: "Print the file the workspace executor runs for a command",
		Long: `Print the file the workspace executor runs for a command.

The command is resolved against the workspace executor's PATH, the way
exec would run it.

Examples:
  runectl resolve go
  runectl resolve --dir /tmp ./script.sh
  runectl resolve -F json gopls`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()

			w, err := a.getWorkspace()
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			executor := w.Executor(ctx)
			defer func() { _ = executor.Close() }()

			path, err := executor.LookPath(ctx, args[0], dir)
			if err != nil {
				return fmt.Errorf("resolve: %w", err)
			}
			return printString(format, path, []string{"Path"})
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)
	cmd.Flags().StringVarP(
		&dir, "dir", "d", "",
		"Working directory to resolve a relative command against",
	)
	return cmd
}
