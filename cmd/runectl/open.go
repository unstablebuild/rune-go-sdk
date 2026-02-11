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
	"github.com/spf13/cobra"
	"github.com/unstablebuild/rune-go-sdk/api/browserapi/browserrpc"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
)

func newOpenCmd(a *app) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "open <uri>",
		Short: "Open a resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			w, err := a.getWorkspace()
			if err != nil {
				return err
			}
			parsed, err := workspaceapi.ParseURI(args[0])
			if err != nil {
				return err
			}
			h, err := w.ResourceOpener(cmd.Context()).Open(parsed)
			if err != nil {
				return err
			}
			uri := args[0]
			if tok, ok := h.(browserrpc.Token); ok {
				uri = tok.URI
			}
			return printString(
				format, uri, []string{"Resource"},
			)
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)

	return cmd
}
