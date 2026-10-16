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

package mcptools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

// RegisterSyntaxTools registers the syntax tree tools backed by p on s. Calls to p are
// made with ctx.
func RegisterSyntaxTools(
	ctx context.Context,
	s *server.MCPServer,
	p syntaxapi.Parser,
	opts ...Option,
) {
	r := newRegistrar(s, opts)
	bgCtx := ctx
	registerSyntaxSearch(r, p, bgCtx)
	registerSyntaxSearchNode(r, p, bgCtx)
	registerSyntaxQuery(r, p, bgCtx)
	registerSyntaxQueryNode(r, p, bgCtx)
}

func registerSyntaxSearch(r *registrar, p syntaxapi.Parser, bgCtx context.Context) {
	r.add(
		mcp.NewTool("syntax_search",
			mcp.WithDescription(
				"Searches all workspace files using a tree-sitter query.\n\n"+
					"Use this tool when:\n"+
					"- Finding code patterns across the workspace\n"+
					"- Structural search beyond text matching\n\n"+
					"Do NOT use this tool when:\n"+
					"- You know the file — use syntax_query\n"+
					"- Searching common tree nodes declarations/definitions (functions,"+
					" methods, scopes, variables, types and namespaces) "+
					"— use syntax_search_node\n"+
					"- Need semantics like how's this node being used — use lsp_* tools\n\n"+
					"Returns [{file, text, position, capture_name}].",
			),
			mcpQuery,
			mcp.WithArray("captures",
				mcp.WithStringItems(),
				mcp.Description(
					"Capture name filters. Only matches for listed capture names are "+
						"returned. Example: [\"definition.function\"]",
				),
			),
			mcp.WithArray("languages",
				mcp.WithStringItems(),
				mcp.Description(
					"Language filters. Only files of the listed languages are "+
						"searched. Example: [\"go\", \"python\"]",
				),
			),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query, err := req.RequireString("query")
			if err != nil {
				return mcpErr(err), nil
			}
			captures := req.GetStringSlice("captures", nil)
			languages := req.GetStringSlice("languages", nil)
			sit, err := p.Search(query, captures, languages...)
			if err != nil {
				return mcpErr(err), nil
			}
			return collectSyntaxResults(bgCtx, sit)
		},
	)
}

func registerSyntaxSearchNode(r *registrar, p syntaxapi.Parser, bgCtx context.Context) {
	r.add(
		mcp.NewTool("syntax_search_node",
			mcp.WithDescription(
				"Searches all workspace files for predefined node categories.\n\n"+
					"Use this tool when:\n"+
					"- Finding all functions, variables, types workspace-wide\n"+
					"- You know the category but not the tree-sitter query\n\n"+
					"Do NOT use this tool when:\n"+
					"- You have a tree-sitter query — use syntax_search\n"+
					"- Only need one file — use syntax_query_node\n"+
					"- Need semantics like how's this node being used — use lsp_* tools\n\n"+
					"Returns [{file, text, position, capture_name}].",
			),
			mcp.WithString("node_types",
				mcp.Required(),
				mcp.Description(
					"Pipe-separated node types. Valid: scope, namespace, reference, "+
						"func, var, method, type. Example: \"func|method\"",
				),
			),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			raw, err := req.RequireString("node_types")
			if err != nil {
				return mcpErr(err), nil
			}
			nt, err := ParseNodeType(raw)
			if err != nil {
				return mcpErr(err), nil
			}
			sit, err := p.SearchNode(nt)
			if err != nil {
				return mcpErr(err), nil
			}
			return collectSyntaxResults(bgCtx, sit)
		},
	)
}

var mcpQuery = mcp.WithString("query", mcp.Required(),
	mcp.Description(
		"Tree-sitter query pattern. A query consists of one or more patterns, where "+
			"each pattern is an S-expression that matches a certain set of nodes in a "+
			"syntax tree. The expression to match a given node consists of a pair of "+
			"parentheses containing two things: the node's type, and optionally, a "+
			"series of other S-expressions that match the node's children. For example,"+
			" this pattern would match any binary_expression node whose children are "+
			"both number_literal nodes: (binary_expression (number_literal) "+
			"(number_literal))",
	),
)

func registerSyntaxQuery(r *registrar, p syntaxapi.Parser, bgCtx context.Context) {
	r.add(
		mcp.NewTool("syntax_query",
			mcp.WithDescription(
				"Queries a single file using a tree-sitter query.\n\n"+
					"Use this tool when:\n"+
					"- Structural search within a specific known file\n"+
					"- You have a tree-sitter query and the file URI\n\n"+
					"Do NOT use this tool when:\n"+
					"- Want to search all files — use syntax_search\n"+
					"- Searching categories — use syntax_query_node\n\n"+
					"Returns [{file, text, position, capture_name}].",
			),
			mcp.WithString("uri", mcp.Required(),
				mcp.Description("Document URI. Example: file:///path/to/file.go")),
			mcpQuery,
			mcp.WithArray("captures",
				mcp.WithStringItems(),
				mcp.Description(
					"Capture name filters. Only matches for listed names are returned."),
			),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uriStr, err := req.RequireString("uri")
			if err != nil {
				return mcpErr(err), nil
			}
			uri, err := workspaceapi.ParseURI(uriStr)
			if err != nil {
				return mcpErr(err), nil
			}
			query, err := req.RequireString("query")
			if err != nil {
				return mcpErr(err), nil
			}
			captures := req.GetStringSlice("captures", nil)
			sit, err := p.Query(uri, query, captures)
			if err != nil {
				return mcpErr(err), nil
			}
			return collectSyntaxResults(bgCtx, sit)
		},
	)
}

func registerSyntaxQueryNode(r *registrar, p syntaxapi.Parser, bgCtx context.Context) {
	r.add(
		mcp.NewTool("syntax_query_node",
			mcp.WithDescription(
				"Queries a single file for predefined node categories.\n\n"+
					"Use this tool when:\n"+
					"- Finding functions, variables, types in a specific file\n"+
					"- You know the file URI and node category\n\n"+
					"Do NOT use this tool when:\n"+
					"- Want all files — use syntax_search_node\n"+
					"- You have a tree-sitter query — use syntax_query\n\n"+
					"Returns [{file, text, position, capture_name}].",
			),
			mcp.WithString("uri", mcp.Required(),
				mcp.Description("Document URI. Example: file:///path/to/file.go")),
			mcp.WithString("node_types",
				mcp.Required(),
				mcp.Description(
					"Pipe-separated node types. Valid: scope, namespace, reference, "+
						"func, var, method, type. Example: \"func|method\"",
				),
			),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uriStr, err := req.RequireString("uri")
			if err != nil {
				return mcpErr(err), nil
			}
			uri, err := workspaceapi.ParseURI(uriStr)
			if err != nil {
				return mcpErr(err), nil
			}
			raw, err := req.RequireString("node_types")
			if err != nil {
				return mcpErr(err), nil
			}
			nt, err := ParseNodeType(raw)
			if err != nil {
				return mcpErr(err), nil
			}
			sit, err := p.QueryNode(uri, nt)
			if err != nil {
				return mcpErr(err), nil
			}
			return collectSyntaxResults(bgCtx, sit)
		},
	)
}

// collectSyntaxResults drains a syntax result iterator and returns the results as JSON.
func collectSyntaxResults(
	ctx context.Context,
	sit iterator.Iterator[syntaxapi.Result],
) (*mcp.CallToolResult, error) {
	defer func() { _ = sit.Close() }()
	results := []SearchResult{}
	for {
		r, ok := sit.Next(ctx)
		if !ok {
			break
		}
		results = append(results, SearchResult{
			File:        fmt.Sprint(r.File),
			Text:        r.Text,
			FromX:       r.From.X,
			FromY:       r.From.Y,
			ToX:         r.To.X,
			ToY:         r.To.Y,
			CaptureName: r.CaptureName,
		})
	}
	if err := sit.Err(); err != nil {
		return mcpErr(err), nil
	}
	return mcpJSON(results)
}
