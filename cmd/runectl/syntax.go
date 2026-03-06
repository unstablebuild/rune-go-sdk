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
	"strings"

	"github.com/spf13/cobra"
	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

type searchResult struct {
	File        string `json:"file"`
	Text        string `json:"text"`
	FromX       int    `json:"from_x"`
	FromY       int    `json:"from_y"`
	ToX         int    `json:"to_x"`
	ToY         int    `json:"to_y"`
	CaptureName string `json:"capture_name"`
}

func syntaxIterToGeneric(
	sit iterator.Iterator[syntaxapi.Result],
) iterator.Iterator[searchResult] {
	return iterator.FromFunc(
		func(ctx context.Context) (
			searchResult, bool, error,
		) {
			r, ok := sit.Next(ctx)
			if !ok {
				return searchResult{}, false,
					sit.Err()
			}
			return searchResult{
				File:        fmt.Sprint(r.File),
				Text:        r.Text,
				FromX:       r.From.X,
				FromY:       r.From.Y,
				ToX:         r.To.X,
				ToY:         r.To.Y,
				CaptureName: r.CaptureName,
			}, true, nil
		},
		sit.Close,
	)
}

func printSyntaxResults(
	ctx context.Context,
	format string,
	sit iterator.Iterator[syntaxapi.Result],
) error {
	defer func() { _ = sit.Close() }()
	it := syntaxIterToGeneric(sit)
	defer func() { _ = it.Close() }()
	if format == "" {
		format = "json"
	}
	return printIterator(ctx, format, it, []string{
		"File", "Text",
		"FromX", "FromY", "ToX", "ToY",
		"CaptureName",
	})
}

func parseSingleNodeType(s string) (syntaxapi.NodeCaptureName, error) {
	switch s {
	case "scope":
		return syntaxapi.NodeCaptureScope, nil
	case "namespace":
		return syntaxapi.NodeCaptureDefinitionNamespace, nil
	case "reference":
		return syntaxapi.NodeCaptureReference, nil
	case "func":
		return syntaxapi.NodeCaptureDefinitionFunc, nil
	case "var":
		return syntaxapi.NodeCaptureDefinitionVar, nil
	case "method":
		return syntaxapi.NodeCaptureDefinitionMethod, nil
	case "type":
		return syntaxapi.NodeCaptureDefinitionType, nil
	default:
		return 0, fmt.Errorf(
			"invalid node type %q: use "+
				"scope|namespace|reference|func|var|method|type",
			s,
		)
	}
}

func parseNodeType(s string) (syntaxapi.NodeCaptureName, error) {
	var result syntaxapi.NodeCaptureName
	for _, part := range strings.Split(s, "|") {
		nodeType, err := parseSingleNodeType(part)
		if err != nil {
			return 0, err
		}
		result |= nodeType
	}
	return result, nil
}

func newSyntaxCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "syntax",
		Short: "AST-level search commands",
	}

	cmd.AddCommand(
		newSyntaxSearchCmd(a),
		newSyntaxSearchNodeCmd(a),
		newSyntaxQueryCmd(a),
		newSyntaxQueryNodeCmd(a),
	)

	return cmd
}

func newSyntaxSearchCmd(a *app) *cobra.Command {
	var format string
	var captures []string
	var languages []string

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search workspace using a tree-sitter query",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			w, err := a.getWorkspace()
			if err != nil {
				return err
			}
			sit, err := w.Parser(cmd.Context()).Search(
				args[0], captures, languages...,
			)
			if err != nil {
				return err
			}
			return printSyntaxResults(cmd.Context(), format, sit)
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)
	cmd.Flags().StringArrayVarP(
		&captures, "capture", "c", nil,
		"Capture name filter (repeatable)",
	)
	cmd.Flags().StringArrayVarP(
		&languages, "lang", "L", nil,
		"Language filter (repeatable, e.g. go, python)",
	)

	return cmd
}

func newSyntaxSearchNodeCmd(a *app) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "searchnode <node-type>",
		Short: "Search workspace for known node types",
		Long: `Search workspace for known node types.

Node types: scope|namespace|reference|func|var|method|type`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			nodeType, err := parseNodeType(args[0])
			if err != nil {
				return err
			}
			w, err := a.getWorkspace()
			if err != nil {
				return err
			}
			sit, err := w.Parser(cmd.Context()).SearchNode(nodeType)
			if err != nil {
				return err
			}
			return printSyntaxResults(cmd.Context(), format, sit)
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)

	return cmd
}

func newSyntaxQueryCmd(a *app) *cobra.Command {
	var format string
	var captures []string

	cmd := &cobra.Command{
		Use:   "query <file> <query>",
		Short: "Query a file using a tree-sitter query",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			uri, err := a.resolveURIArg(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			w, err := a.getWorkspace()
			if err != nil {
				return err
			}
			sit, err := w.Parser(cmd.Context()).Query(
				uri, args[1], captures,
			)
			if err != nil {
				return err
			}
			return printSyntaxResults(cmd.Context(), format, sit)
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)
	cmd.Flags().StringArrayVarP(
		&captures, "capture", "c", nil,
		"Capture name filter (repeatable)",
	)

	return cmd
}

func newSyntaxQueryNodeCmd(a *app) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "querynode <file> <node-type>",
		Short: "Query a file for known node types",
		Long: `Query a file for known node types.

Node types: scope|namespace|reference|func|var|method|type`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			uri, err := a.resolveURIArg(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			nodeType, err := parseNodeType(args[1])
			if err != nil {
				return err
			}
			w, err := a.getWorkspace()
			if err != nil {
				return err
			}
			sit, err := w.Parser(cmd.Context()).QueryNode(uri, nodeType)
			if err != nil {
				return err
			}
			return printSyntaxResults(cmd.Context(), format, sit)
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)

	return cmd
}
