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
	"fmt"
	"strings"

	"github.com/unstablebuild/rune-go-sdk/api/storageapi"
	"github.com/unstablebuild/rune-go-sdk/cli"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

type storageCLI struct {
	app  *app
	fs   *cli.FlagSet
	cmds map[string]cli.CLI
}

func newStorageCLI(a *app) *storageCLI {
	c := &storageCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl storage")
	c.cmds = map[string]cli.CLI{
		"create": newStorageCreateCLI(a),
		"set":    newStorageSetCLI(a),
		"get":    newStorageGetCLI(a),
		"update": newStorageUpdateCLI(a),
		"delete": newStorageDeleteCLI(a),
		"list":   newStorageListCLI(a),
	}
	return c
}

func (c *storageCLI) Run(
	ctx context.Context, args []string,
) error {
	return cli.ParseAndRunCommand(
		ctx, c, c.fs, c.cmds, args,
	)
}

func (c *storageCLI) Man() cli.Manual {
	var subMan []cli.Manual
	for _, name := range []string{
		"create", "set", "get",
		"update", "delete", "list",
	} {
		subMan = append(subMan, c.cmds[name].Man())
	}
	return cli.Manual{
		Name:     "storage",
		Summary:  "Document storage commands",
		Synopsis: "<command>",
		Commands: subMan,
		Options:  *c.fs,
	}
}

type storageCreateCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newStorageCreateCLI(a *app) *storageCreateCLI {
	c := &storageCreateCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl storage create")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *storageCreateCLI) Run(
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
	var doc map[string]any
	if err := json.Unmarshal(
		[]byte(rargs[1]), &doc,
	); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	err = w.Storage(ctx).Create(ctx, rargs[0], doc)
	if err != nil {
		return err
	}
	return printOK(c.format)
}

func (c *storageCreateCLI) Man() cli.Manual {
	return cli.Manual{
		Name: "create",
		Summary: "Create a document",
		Synopsis: "[options] <id> <json-doc>",
		Options: *c.fs,
	}
}

type storageSetCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newStorageSetCLI(a *app) *storageSetCLI {
	c := &storageSetCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl storage set")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *storageSetCLI) Run(
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
	var doc map[string]any
	if err := json.Unmarshal(
		[]byte(rargs[1]), &doc,
	); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	err = w.Storage(ctx).Set(ctx, rargs[0], doc)
	if err != nil {
		return err
	}
	return printOK(c.format)
}

func (c *storageSetCLI) Man() cli.Manual {
	return cli.Manual{
		Name:     "set",
		Summary:  "Create or update a document",
		Synopsis: "[options] <id> <json-doc>",
		Options:  *c.fs,
	}
}

type storageGetCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newStorageGetCLI(a *app) *storageGetCLI {
	c := &storageGetCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl storage get")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *storageGetCLI) Run(
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
	if c.format == "table" {
		return fmt.Errorf(
			"table format not supported" +
				" for storage get",
		)
	}
	var doc map[string]any
	err = w.Storage(ctx).Get(ctx, rargs[0], &doc)
	if err != nil {
		return err
	}
	format := c.format
	if format == "" {
		format = "json"
	}
	it := iterator.FromSlice([]map[string]any{doc})
	return printIterator(ctx, format, it, nil)
}

func (c *storageGetCLI) Man() cli.Manual {
	return cli.Manual{
		Name:     "get",
		Summary:  "Get a document",
		Synopsis: "[options] <id>",
		Options:  *c.fs,
	}
}

type storageUpdateCLI struct {
	app     *app
	fs      *cli.FlagSet
	format  string
	precond string
}

func newStorageUpdateCLI(a *app) *storageUpdateCLI {
	c := &storageUpdateCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl storage update")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	c.fs.StringVar(
		&c.precond, "p", "",
		"JSON preconditions",
	)
	return c
}

func (c *storageUpdateCLI) Run(
	ctx context.Context, args []string,
) (retErr error) {
	defer func() {
		retErr = formatError(c.format, retErr)
	}()
	_, rest, ok, err := cli.ParseUsage(
		c, c.fs, 0, args,
	)
	if !ok || err != nil {
		return err
	}
	if len(rest) < 3 || len(rest)%2 == 0 {
		return fmt.Errorf(
			"usage: storage update [options] <id>" +
				" <field.path> <value>" +
				" [<field.path> <value> ...]",
		)
	}
	id := rest[0]
	updates, err := parseUpdateArgs(rest[1:])
	if err != nil {
		return err
	}
	var preconds []storageapi.Precondition
	if c.precond != "" {
		preconds, err = parsePreconditions(c.precond)
		if err != nil {
			return err
		}
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	err = w.Storage(ctx).Update(
		ctx, id, updates, preconds...,
	)
	if err != nil {
		return err
	}
	return printOK(c.format)
}

func (c *storageUpdateCLI) Man() cli.Manual {
	return cli.Manual{
		Name: "update",
		Summary: "Update document fields",
		Synopsis: "[options] <id> <field.path>" +
			" <value> [<field.path> <value> ...]",
		Options: *c.fs,
	}
}

type storageDeleteCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newStorageDeleteCLI(a *app) *storageDeleteCLI {
	c := &storageDeleteCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl storage delete")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *storageDeleteCLI) Run(
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
	err = w.Storage(ctx).Delete(ctx, rargs[0])
	if err != nil {
		return err
	}
	return printOK(c.format)
}

func (c *storageDeleteCLI) Man() cli.Manual {
	return cli.Manual{
		Name:     "delete",
		Summary:  "Delete a document",
		Synopsis: "[options] <id>",
		Options:  *c.fs,
	}
}

type storageListCLI struct {
	app     *app
	fs      *cli.FlagSet
	format  string
	filters multiFlag
}

func newStorageListCLI(a *app) *storageListCLI {
	c := &storageListCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl storage list")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	c.fs.Var(
		&c.filters, "f",
		"JSON filter (repeatable)",
	)
	return c
}

func (c *storageListCLI) Run(
	ctx context.Context, args []string,
) (retErr error) {
	defer func() {
		retErr = formatError(c.format, retErr)
	}()
	_, _, ok, err := cli.ParseUsage(
		c, c.fs, 0, args,
	)
	if !ok || err != nil {
		return err
	}
	if c.format == "table" {
		return fmt.Errorf(
			"table format not supported" +
				" for storage list",
		)
	}
	filters, err := parseFilters(c.filters)
	if err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	sit, err := w.Storage(ctx).List(ctx, filters)
	if err != nil {
		return err
	}
	defer func() { _ = sit.Close() }()

	it := storageIterToGeneric(ctx, sit)
	defer func() { _ = it.Close() }()

	format := c.format
	if format == "" {
		format = "json"
	}
	return printIterator(ctx, format, it, nil)
}

func (c *storageListCLI) Man() cli.Manual {
	return cli.Manual{
		Name:     "list",
		Summary:  "List documents",
		Synopsis: "[options]",
		Options:  *c.fs,
	}
}

type multiFlag []string

func (m *multiFlag) String() string { return "" }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func parseUpdateArgs(
	args []string,
) ([]storageapi.Update, error) {
	if len(args)%2 != 0 {
		return nil, fmt.Errorf(
			"update requires pairs of" +
				" <field.path> <value>",
		)
	}
	updates := make(
		[]storageapi.Update, 0, len(args)/2,
	)
	for i := 0; i < len(args); i += 2 {
		path := strings.Split(args[i], ".")
		for _, p := range path {
			if p == "" {
				return nil, fmt.Errorf(
					"empty segment in field path %q",
					args[i],
				)
			}
		}
		var value any
		if err := json.Unmarshal(
			[]byte(args[i+1]), &value,
		); err != nil {
			value = args[i+1]
		}
		updates = append(updates, storageapi.Update{
			FieldPath: path, Value: value,
		})
	}
	return updates, nil
}

type jsonPrecondition struct {
	Path  []string `json:"path"`
	Value any      `json:"value"`
}

func parsePreconditions(
	s string,
) ([]storageapi.Precondition, error) {
	var raw []jsonPrecondition
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		return nil, fmt.Errorf(
			"invalid preconditions JSON: %w", err,
		)
	}
	ret := make([]storageapi.Precondition, len(raw))
	for i, p := range raw {
		ret[i] = storageapi.Precondition{
			FieldPath: p.Path, Value: p.Value,
		}
	}
	return ret, nil
}

type jsonFilter struct {
	Path  []string `json:"path"`
	Op    string   `json:"op"`
	Value any      `json:"value"`
}

func parseFilters(
	raw []string,
) ([]storageapi.Filter, error) {
	var ret []storageapi.Filter
	for _, s := range raw {
		var f jsonFilter
		if err := json.Unmarshal(
			[]byte(s), &f,
		); err != nil {
			return nil, fmt.Errorf(
				"invalid filter JSON %q: %w", s, err,
			)
		}
		ret = append(ret, storageapi.Filter{
			Field: storageapi.Field{
				FieldPath: f.Path,
				Value:     f.Value,
			},
			Op: storageapi.Op(f.Op),
		})
	}
	return ret, nil
}

func storageIterToGeneric(
	_ context.Context,
	sit storageapi.Iterator,
) iterator.Iterator[map[string]any] {
	return iterator.FromFunc(
		func(ctx context.Context) (
			map[string]any, bool, error,
		) {
			if !sit.HasNext() {
				return nil, false, nil
			}
			var doc map[string]any
			if err := sit.NextTo(&doc); err != nil {
				return nil, false, err
			}
			return doc, true, nil
		},
		sit.Close,
	)
}
