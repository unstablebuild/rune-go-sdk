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

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"github.com/unstablebuild/rune-go-sdk/api/extensionapi"
	"github.com/unstablebuild/rune-go-sdk/cli/mcptools"
)

func newMCPCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server over stdio",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			w, err := a.getWorkspace()
			if err != nil {
				return err
			}
			s := server.NewMCPServer("runectl", "0.1.0", server.WithRecovery())
			registerMCPTools(cmd.Context(), s, w)
			return server.ServeStdio(s)
		},
	}
	return cmd
}

// registerMCPTools registers every syntax and LSP tool.
func registerMCPTools(
	ctx context.Context,
	s *server.MCPServer,
	w *extensionapi.Workspace,
) {
	parser := w.Parser(ctx)
	mcptools.RegisterSyntaxTools(ctx, s, parser, mcptools.WithAllThemes())
	mcptools.RegisterLSPTools(ctx, s, w.LSP(ctx), parser, mcptools.WithAllThemes())
}
