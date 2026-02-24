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
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

func newMCPCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server over stdio",
		Args:  cobra.NoArgs,
		RunE: func(
			cmd *cobra.Command, _ []string,
		) error {
			w, err := a.getWorkspace()
			if err != nil {
				return err
			}
			s := server.NewMCPServer(
				"runectl", "0.1.0",
				server.WithRecovery(),
			)
			ctx := cmd.Context()
			registerSyntaxTools(s, w, ctx)
			registerLSPTools(s, w, ctx)
			registerDebugTools(s, w, ctx)
			return server.ServeStdio(s)
		},
	}
	return cmd
}

// mcpJSON marshals v as JSON text and returns it as
// an MCP tool result.
func mcpJSON(v any) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(data)), nil
}

// mcpErr returns an MCP error result from an error.
func mcpErr(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(err.Error())
}

// mcpPos extracts uri, line, and character from an
// MCP request.
func mcpPos(
	req mcp.CallToolRequest,
) (string, uint32, uint32, error) {
	uri, err := req.RequireString("uri")
	if err != nil {
		return "", 0, 0, err
	}
	line, err := req.RequireFloat("line")
	if err != nil {
		return "", 0, 0, err
	}
	char, err := req.RequireFloat("character")
	if err != nil {
		return "", 0, 0, err
	}
	return uri, uint32(line), uint32(char), nil
}

// mcpRange extracts range parameters from an MCP
// request.
func mcpRange(
	req mcp.CallToolRequest,
) (string, [4]uint32, error) {
	uri, err := req.RequireString("uri")
	if err != nil {
		return "", [4]uint32{}, err
	}
	sl, err := req.RequireFloat("start_line")
	if err != nil {
		return "", [4]uint32{}, err
	}
	sc, err := req.RequireFloat("start_character")
	if err != nil {
		return "", [4]uint32{}, err
	}
	el, err := req.RequireFloat("end_line")
	if err != nil {
		return "", [4]uint32{}, err
	}
	ec, err := req.RequireFloat("end_character")
	if err != nil {
		return "", [4]uint32{}, err
	}
	return uri, [4]uint32{
		uint32(sl), uint32(sc),
		uint32(el), uint32(ec),
	}, nil
}
