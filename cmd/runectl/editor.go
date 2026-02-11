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
	"strconv"

	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/cli"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/tcell/v3"
)

type editorCLI struct {
	app  *app
	fs   *cli.FlagSet
	cmds map[string]cli.CLI
}

func newEditorCLI(a *app) *editorCLI {
	c := &editorCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl editor")
	c.cmds = map[string]cli.CLI{
		"print":     newEditorPrintCLI(a),
		"color":     newEditorColorCLI(a),
		"edit":      newEditorEditCLI(a),
		"locations": newEditorLocationsCLI(a),
		"cursor":    newEditorCursorCLI(a),
	}
	return c
}

func (c *editorCLI) Run(
	ctx context.Context, args []string,
) error {
	return cli.ParseAndRunCommand(
		ctx, c, c.fs, c.cmds, args,
	)
}

func (c *editorCLI) Man() cli.Manual {
	var subMan []cli.Manual
	for _, name := range []string{
		"print", "color",
		"edit", "locations", "cursor",
	} {
		subMan = append(subMan, c.cmds[name].Man())
	}
	return cli.Manual{
		Name:     "editor",
		Summary:  "Editor commands",
		Synopsis: "<command>",
		Commands: subMan,
		Options:  *c.fs,
	}
}

func getEditorHandler(
	ctx context.Context, a *app, uri string,
) (textapi.Handler, error) {
	w, err := a.getWorkspace()
	if err != nil {
		return nil, err
	}
	parsed, err := workspaceapi.ParseURI(uri)
	if err != nil {
		return nil, err
	}
	return w.Editor(ctx).Editor(parsed)
}

type editorPrintCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newEditorPrintCLI(a *app) *editorPrintCLI {
	c := &editorPrintCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl editor print")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *editorPrintCLI) Run(
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
	h, err := getEditorHandler(ctx, c.app, rargs[0])
	if err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	cells, err := w.Editor(ctx).CellView(h).RawCells()
	if err != nil {
		return err
	}
	text := term.CellsToString(cells)
	return printString(
		c.format, text, []string{"Content"},
	)
}

func (c *editorPrintCLI) Man() cli.Manual {
	return cli.Manual{
		Name:     "print",
		Summary:  "Print editor text content",
		Synopsis: "[options] <uri>",
		Options:  *c.fs,
	}
}

type editorColorCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newEditorColorCLI(a *app) *editorColorCLI {
	c := &editorColorCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl editor color")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *editorColorCLI) Run(
	ctx context.Context, args []string,
) (retErr error) {
	defer func() {
		retErr = formatError(c.format, retErr)
	}()
	rargs, rest, ok, err := cli.ParseUsage(
		c, c.fs, 2, args,
	)
	if !ok || err != nil {
		return err
	}
	h, err := getEditorHandler(ctx, c.app, rargs[0])
	if err != nil {
		return err
	}
	var attr term.Attributes
	attr.Bg = tcell.GetColor(rargs[1])
	if len(rest) > 0 {
		attr.Fg = tcell.GetColor(rest[0])
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	err = w.Editor(ctx).SetDefaultAttributes(h, attr)
	if err != nil {
		return err
	}
	return printOK(c.format)
}

func (c *editorColorCLI) Man() cli.Manual {
	return cli.Manual{
		Name:     "color",
		Summary:  "Set editor default colors",
		Synopsis: "[options] <uri> <bg> [<fg>]",
		Options:  *c.fs,
	}
}

type editorEditCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newEditorEditCLI(a *app) *editorEditCLI {
	c := &editorEditCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl editor edit")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

type editResult struct {
	FromX int    `json:"from_x"`
	FromY int    `json:"from_y"`
	ToX   int    `json:"to_x"`
	ToY   int    `json:"to_y"`
	Old   string `json:"old"`
}

func (c *editorEditCLI) Run(
	ctx context.Context, args []string,
) (retErr error) {
	defer func() {
		retErr = formatError(c.format, retErr)
	}()
	rargs, _, ok, err := cli.ParseUsage(
		c, c.fs, 6, args,
	)
	if !ok || err != nil {
		return err
	}
	h, err := getEditorHandler(ctx, c.app, rargs[0])
	if err != nil {
		return err
	}
	startX, err := strconv.Atoi(rargs[1])
	if err != nil {
		return fmt.Errorf("invalid start-x: %w", err)
	}
	startY, err := strconv.Atoi(rargs[2])
	if err != nil {
		return fmt.Errorf("invalid start-y: %w", err)
	}
	endX, err := strconv.Atoi(rargs[3])
	if err != nil {
		return fmt.Errorf("invalid end-x: %w", err)
	}
	endY, err := strconv.Atoi(rargs[4])
	if err != nil {
		return fmt.Errorf("invalid end-y: %w", err)
	}
	text := rargs[5]
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	start := term.Coordinates{X: startX, Y: startY}
	end := term.Coordinates{X: endX, Y: endY}
	from, to, old, err := w.Editor(ctx).CellEditor(
		h,
	).Edit(ctx, start, end, text)
	if err != nil {
		return err
	}
	res := editResult{
		FromX: from.X, FromY: from.Y,
		ToX: to.X, ToY: to.Y,
		Old: old,
	}
	return printResult(
		ctx, c.format, res,
		func(v editResult) {
			fmt.Printf(
				"%d %d %d %d %s\n",
				v.FromX, v.FromY,
				v.ToX, v.ToY, v.Old,
			)
		},
		[]string{
			"FromX", "FromY", "ToX", "ToY", "Old",
		},
	)
}

func (c *editorEditCLI) Man() cli.Manual {
	return cli.Manual{
		Name: "edit",
		Summary: "Edit text in a buffer",
		Synopsis: "[options] <uri> <start-x> <start-y>" +
			" <end-x> <end-y> <text>",
		Options: *c.fs,
	}
}

type editorLocationsCLI struct {
	app  *app
	fs   *cli.FlagSet
	cmds map[string]cli.CLI
}

func newEditorLocationsCLI(a *app) *editorLocationsCLI {
	c := &editorLocationsCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl editor locations")
	c.cmds = map[string]cli.CLI{
		"set":  newLocationsSetCLI(a),
		"next": newLocationsNextCLI(a),
		"prev": newLocationsPrevCLI(a),
	}
	return c
}

func (c *editorLocationsCLI) Run(
	ctx context.Context, args []string,
) error {
	return cli.ParseAndRunCommand(
		ctx, c, c.fs, c.cmds, args,
	)
}

func (c *editorLocationsCLI) Man() cli.Manual {
	var subMan []cli.Manual
	for _, name := range []string{
		"set", "next", "prev",
	} {
		subMan = append(subMan, c.cmds[name].Man())
	}
	return cli.Manual{
		Name:     "locations",
		Summary:  "Location list commands",
		Synopsis: "<command>",
		Commands: subMan,
		Options:  *c.fs,
	}
}

type locationsSetCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newLocationsSetCLI(a *app) *locationsSetCLI {
	c := &locationsSetCLI{app: a}
	c.fs = cli.NewFlagSet(
		"runectrl editor locations set",
	)
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

type jsonLocation struct {
	From    jsonCoord `json:"from"`
	To      jsonCoord `json:"to"`
	Message string    `json:"message"`
}

type jsonCoord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (c *locationsSetCLI) Run(
	ctx context.Context, args []string,
) (retErr error) {
	defer func() {
		retErr = formatError(c.format, retErr)
	}()
	rargs, _, ok, err := cli.ParseUsage(
		c, c.fs, 4, args,
	)
	if !ok || err != nil {
		return err
	}
	priority, err := parseLocationPriority(rargs[1])
	if err != nil {
		return err
	}
	id := rargs[2]
	var locs []jsonLocation
	if err := json.Unmarshal(
		[]byte(rargs[3]), &locs,
	); err != nil {
		return fmt.Errorf(
			"invalid locations JSON: %w", err,
		)
	}
	locations := make([]textapi.Location, len(locs))
	for i, l := range locs {
		locations[i] = textapi.Location{
			From: term.Coordinates{
				X: l.From.X, Y: l.From.Y,
			},
			To: term.Coordinates{
				X: l.To.X, Y: l.To.Y,
			},
			Message: l.Message,
		}
	}
	h, err := getEditorHandler(ctx, c.app, rargs[0])
	if err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	err = w.Editor(ctx).SetLocationList(
		h, priority, id,
		textapi.LocationSlice(locations),
	)
	if err != nil {
		return err
	}
	return printOK(c.format)
}

func (c *locationsSetCLI) Man() cli.Manual {
	return cli.Manual{
		Name: "set",
		Summary: "Set a location list",
		Synopsis: "[options] <uri> <priority> <id>" +
			" <json-locations>\n" +
			"  priority: info|warning|error|critical\n" +
			"  json-locations: " +
			`[{"from":{"x":N,"y":N},` +
			`"to":{"x":N,"y":N},` +
			`"message":"..."},...]`,
		Options: *c.fs,
	}
}

type locationsNextCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newLocationsNextCLI(a *app) *locationsNextCLI {
	c := &locationsNextCLI{app: a}
	c.fs = cli.NewFlagSet(
		"runectrl editor locations next",
	)
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *locationsNextCLI) Run(
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
	h, err := getEditorHandler(ctx, c.app, rargs[0])
	if err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	err = w.Editor(ctx).MoveToNextLocation(h, rargs[1])
	if err != nil {
		return err
	}
	return printOK(c.format)
}

func (c *locationsNextCLI) Man() cli.Manual {
	return cli.Manual{
		Name:     "next",
		Summary:  "Move to next location",
		Synopsis: "[options] <uri> <id>",
		Options:  *c.fs,
	}
}

type locationsPrevCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newLocationsPrevCLI(a *app) *locationsPrevCLI {
	c := &locationsPrevCLI{app: a}
	c.fs = cli.NewFlagSet(
		"runectrl editor locations prev",
	)
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *locationsPrevCLI) Run(
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
	h, err := getEditorHandler(ctx, c.app, rargs[0])
	if err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	err = w.Editor(ctx).MoveToPrevLocation(h, rargs[1])
	if err != nil {
		return err
	}
	return printOK(c.format)
}

func (c *locationsPrevCLI) Man() cli.Manual {
	return cli.Manual{
		Name:     "prev",
		Summary:  "Move to previous location",
		Synopsis: "[options] <uri> <id>",
		Options:  *c.fs,
	}
}

type editorCursorCLI struct {
	app  *app
	fs   *cli.FlagSet
	cmds map[string]cli.CLI
}

func newEditorCursorCLI(a *app) *editorCursorCLI {
	c := &editorCursorCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl editor cursor")
	c.cmds = map[string]cli.CLI{
		"get": newCursorGetCLI(a),
		"set": newCursorSetCLI(a),
	}
	return c
}

func (c *editorCursorCLI) Run(
	ctx context.Context, args []string,
) error {
	return cli.ParseAndRunCommand(
		ctx, c, c.fs, c.cmds, args,
	)
}

func (c *editorCursorCLI) Man() cli.Manual {
	var subMan []cli.Manual
	for _, name := range []string{"get", "set"} {
		subMan = append(subMan, c.cmds[name].Man())
	}
	return cli.Manual{
		Name:     "cursor",
		Summary:  "Cursor commands",
		Synopsis: "<command>",
		Commands: subMan,
		Options:  *c.fs,
	}
}

type cursorGetCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newCursorGetCLI(a *app) *cursorGetCLI {
	c := &cursorGetCLI{app: a}
	c.fs = cli.NewFlagSet(
		"runectrl editor cursor get",
	)
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *cursorGetCLI) Run(
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
	h, err := getEditorHandler(ctx, c.app, rargs[0])
	if err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	coords, err := w.Editor(ctx).Cursor(h)
	if err != nil {
		return err
	}
	return printResult(
		ctx, c.format, coords,
		func(v term.Coordinates) {
			fmt.Printf("%d %d\n", v.X, v.Y)
		},
		[]string{"X", "Y"},
	)
}

func (c *cursorGetCLI) Man() cli.Manual {
	return cli.Manual{
		Name:     "get",
		Summary:  "Get cursor position",
		Synopsis: "[options] <uri>",
		Options:  *c.fs,
	}
}

type cursorSetCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newCursorSetCLI(a *app) *cursorSetCLI {
	c := &cursorSetCLI{app: a}
	c.fs = cli.NewFlagSet(
		"runectrl editor cursor set",
	)
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *cursorSetCLI) Run(
	ctx context.Context, args []string,
) (retErr error) {
	defer func() {
		retErr = formatError(c.format, retErr)
	}()
	rargs, _, ok, err := cli.ParseUsage(
		c, c.fs, 3, args,
	)
	if !ok || err != nil {
		return err
	}
	x, err := strconv.Atoi(rargs[1])
	if err != nil {
		return fmt.Errorf("invalid x: %w", err)
	}
	y, err := strconv.Atoi(rargs[2])
	if err != nil {
		return fmt.Errorf("invalid y: %w", err)
	}
	h, err := getEditorHandler(ctx, c.app, rargs[0])
	if err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	err = w.Editor(ctx).SetCursor(
		h, term.Coordinates{X: x, Y: y},
	)
	if err != nil {
		return err
	}
	return printOK(c.format)
}

func (c *cursorSetCLI) Man() cli.Manual {
	return cli.Manual{
		Name:     "set",
		Summary:  "Set cursor position",
		Synopsis: "[options] <uri> <x> <y>",
		Options:  *c.fs,
	}
}

func parseLocationPriority(
	s string,
) (textapi.LocationPriority, error) {
	switch s {
	case "info":
		return textapi.LocationPriorityInfo, nil
	case "warning":
		return textapi.LocationPriorityWarning, nil
	case "error":
		return textapi.LocationPriorityError, nil
	case "critical":
		return textapi.LocationPriorityCritical, nil
	default:
		return 0, fmt.Errorf(
			"invalid priority %q:"+
				" use info|warning|error|critical", s,
		)
	}
}
