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
	"path/filepath"
	"strings"

	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/cli"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

type syntaxCLI struct {
	app  *app
	fs   *cli.FlagSet
	cmds map[string]cli.CLI
}

func newSyntaxCLI(a *app) *syntaxCLI {
	c := &syntaxCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl syntax")
	c.cmds = map[string]cli.CLI{
		"search":     newSyntaxSearchCLI(a),
		"searchnode": newSyntaxSearchNodeCLI(a),
		"query":      newSyntaxQueryCLI(a),
		"querynode":  newSyntaxQueryNodeCLI(a),
	}
	return c
}

func (c *syntaxCLI) Run(
	ctx context.Context, args []string,
) error {
	return cli.ParseAndRunCommand(
		ctx, c, c.fs, c.cmds, args,
	)
}

func (c *syntaxCLI) Man() cli.Manual {
	var subMan []cli.Manual
	for _, name := range []string{
		"search", "searchnode", "query", "querynode",
	} {
		subMan = append(subMan, c.cmds[name].Man())
	}
	return cli.Manual{
		Name:     "syntax",
		Summary:  "AST-level search commands",
		Synopsis: "<command>",
		Commands: subMan,
		Options:  *c.fs,
	}
}

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

func (a *app) pathToURI(
	ctx context.Context, path string,
) (workspaceapi.URI, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return workspaceapi.URI{}, fmt.Errorf(
			"resolve path: %w", err,
		)
	}
	w, err := a.getWorkspace()
	if err != nil {
		return workspaceapi.URI{}, err
	}
	return w.FileSystem(ctx).URI(absPath)
}

type syntaxSearchCLI struct {
	app      *app
	fs       *cli.FlagSet
	format   string
	captures multiFlag
}

func newSyntaxSearchCLI(a *app) *syntaxSearchCLI {
	c := &syntaxSearchCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl syntax search")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	c.fs.Var(
		&c.captures, "c",
		"Capture name filter (repeatable)",
	)
	return c
}

func (c *syntaxSearchCLI) Run(
	ctx context.Context, args []string,
) (retErr error) {
	defer func() {
		retErr = formatError(c.format, retErr)
	}()
	rargs, _, ok, err := cli.ParseUsage(
		c, c.fs, 1, args,
	)
	if !ok || err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	sit, err := w.Searcher(ctx).Search(
		rargs[0], []string(c.captures),
	)
	if err != nil {
		return err
	}
	return printSyntaxResults(ctx, c.format, sit)
}

func (c *syntaxSearchCLI) Man() cli.Manual {
	return cli.Manual{
		Name: "search",
		Summary: "Search workspace using a tree-sitter query",
		Synopsis: "[options] <query>",
		Options: *c.fs,
	}
}

type syntaxSearchNodeCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newSyntaxSearchNodeCLI(a *app) *syntaxSearchNodeCLI {
	c := &syntaxSearchNodeCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl syntax searchnode")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *syntaxSearchNodeCLI) Run(
	ctx context.Context, args []string,
) (retErr error) {
	defer func() {
		retErr = formatError(c.format, retErr)
	}()
	rargs, _, ok, err := cli.ParseUsage(
		c, c.fs, 1, args,
	)
	if !ok || err != nil {
		return err
	}
	nodeType, err := parseNodeType(rargs[0])
	if err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	sit, err := w.Searcher(ctx).SearchNode(nodeType)
	if err != nil {
		return err
	}
	return printSyntaxResults(ctx, c.format, sit)
}

func (c *syntaxSearchNodeCLI) Man() cli.Manual {
	return cli.Manual{
		Name: "searchnode",
		Summary: "Search workspace for known node types",
		Synopsis: "[options] <node-type>\n" +
			"  node-type: scope|namespace|reference|" +
			"func|var|method|type",
		Options: *c.fs,
	}
}

type syntaxQueryCLI struct {
	app      *app
	fs       *cli.FlagSet
	format   string
	captures multiFlag
}

func newSyntaxQueryCLI(a *app) *syntaxQueryCLI {
	c := &syntaxQueryCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl syntax query")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	c.fs.Var(
		&c.captures, "c",
		"Capture name filter (repeatable)",
	)
	return c
}

func (c *syntaxQueryCLI) Run(
	ctx context.Context, args []string,
) (retErr error) {
	defer func() {
		retErr = formatError(c.format, retErr)
	}()
	rargs, _, ok, err := cli.ParseUsage(
		c, c.fs, 2, args,
	)
	if !ok || err != nil {
		return err
	}
	uri, err := c.app.pathToURI(ctx, rargs[0])
	if err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	sit, err := w.Searcher(ctx).Query(
		uri, rargs[1], []string(c.captures),
	)
	if err != nil {
		return err
	}
	return printSyntaxResults(ctx, c.format, sit)
}

func (c *syntaxQueryCLI) Man() cli.Manual {
	return cli.Manual{
		Name: "query",
		Summary: "Query a file using a tree-sitter query",
		Synopsis: "[options] <file> <query>",
		Options: *c.fs,
	}
}

type syntaxQueryNodeCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newSyntaxQueryNodeCLI(a *app) *syntaxQueryNodeCLI {
	c := &syntaxQueryNodeCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl syntax querynode")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *syntaxQueryNodeCLI) Run(
	ctx context.Context, args []string,
) (retErr error) {
	defer func() {
		retErr = formatError(c.format, retErr)
	}()
	rargs, _, ok, err := cli.ParseUsage(
		c, c.fs, 2, args,
	)
	if !ok || err != nil {
		return err
	}
	uri, err := c.app.pathToURI(ctx, rargs[0])
	if err != nil {
		return err
	}
	nodeType, err := parseNodeType(rargs[1])
	if err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	sit, err := w.Searcher(ctx).QueryNode(uri, nodeType)
	if err != nil {
		return err
	}
	return printSyntaxResults(ctx, c.format, sit)
}

func (c *syntaxQueryNodeCLI) Man() cli.Manual {
	return cli.Manual{
		Name: "querynode",
		Summary: "Query a file for known node types",
		Synopsis: "[options] <file> <node-type>\n" +
			"  node-type: scope|namespace|reference|" +
			"func|var|method|type",
		Options: *c.fs,
	}
}
