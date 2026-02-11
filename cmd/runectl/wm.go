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
	"strconv"

	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/cli"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
)

// windowID implements browserapi.Window.
type windowID uint64

func (w windowID) WindowID() uint64 { return uint64(w) }

func resolveHandler(
	ctx context.Context,
	a *app,
	uri string,
) (browserapi.Handler, error) {
	w, err := a.getWorkspace()
	if err != nil {
		return nil, err
	}
	parsed, err := workspaceapi.ParseURI(uri)
	if err != nil {
		return nil, err
	}
	return w.ResourceOpener(ctx).Open(parsed)
}

func parseWindowID(s string) (windowID, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid window ID %q: %w", s, err,
		)
	}
	return windowID(v), nil
}

func parseOrientation(
	s string,
) (browserapi.Orientation, error) {
	switch s {
	case "default":
		return browserapi.OrientationDefault, nil
	case "top":
		return browserapi.OrientationTop, nil
	case "bottom":
		return browserapi.OrientationBottom, nil
	case "left":
		return browserapi.OrientationLeft, nil
	case "right":
		return browserapi.OrientationRight, nil
	default:
		return 0, fmt.Errorf(
			"invalid orientation %q:"+
				" use default|top|bottom|left|right", s,
		)
	}
}

type wmCLI struct {
	app  *app
	fs   *cli.FlagSet
	cmds map[string]cli.CLI
}

func newWMCLI(a *app) *wmCLI {
	c := &wmCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl wm")
	c.cmds = map[string]cli.CLI{
		"focus":       newWMFocusCLI(a),
		"split":       newWMSplitCLI(a),
		"floating":    newWMFloatingCLI(a),
		"set-content": newWMSetContentCLI(a),
		"close":       newWMCloseCLI(a),
	}
	return c
}

func (c *wmCLI) Run(
	ctx context.Context, args []string,
) error {
	return cli.ParseAndRunCommand(
		ctx, c, c.fs, c.cmds, args,
	)
}

func (c *wmCLI) Man() cli.Manual {
	var subMan []cli.Manual
	for _, name := range []string{
		"focus", "split", "floating",
		"set-content", "close",
	} {
		subMan = append(subMan, c.cmds[name].Man())
	}
	return cli.Manual{
		Name:     "wm",
		Summary:  "Window manager commands",
		Synopsis: "<command>",
		Commands: subMan,
		Options:  *c.fs,
	}
}

type wmFocusCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newWMFocusCLI(a *app) *wmFocusCLI {
	c := &wmFocusCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl wm focus")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *wmFocusCLI) Run(
	ctx context.Context, args []string,
) (retErr error) {
	defer func() {
		retErr = formatError(c.format, retErr)
	}()
	_, _, ok, err := cli.ParseUsage(c, c.fs, 0, args)
	if !ok || err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	win, err := w.WindowManager(ctx).Focus()
	if err != nil {
		return err
	}
	return printResult(
		ctx, c.format, win.WindowID(),
		func(v uint64) { fmt.Println(v) },
		[]string{"WindowID"},
	)
}

func (c *wmFocusCLI) Man() cli.Manual {
	return cli.Manual{
		Name:     "focus",
		Summary:  "Get the focused window ID",
		Synopsis: "[options]",
		Options:  *c.fs,
	}
}

type wmSplitCLI struct {
	app         *app
	fs          *cli.FlagSet
	format      string
	orientation string
}

func newWMSplitCLI(a *app) *wmSplitCLI {
	c := &wmSplitCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl wm split")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	c.fs.StringVar(
		&c.orientation, "o", "default",
		"Orientation: default|top|bottom|left|right",
	)
	return c
}

func (c *wmSplitCLI) Run(
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
	orient, err := parseOrientation(c.orientation)
	if err != nil {
		return err
	}
	wid, err := parseWindowID(rargs[0])
	if err != nil {
		return err
	}
	h, err := resolveHandler(ctx, c.app, rargs[1])
	if err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	win, err := w.WindowManager(ctx).Split(
		orient, wid, h,
	)
	if err != nil {
		return err
	}
	return printResult(
		ctx, c.format, win.WindowID(),
		func(v uint64) { fmt.Println(v) },
		[]string{"WindowID"},
	)
}

func (c *wmSplitCLI) Man() cli.Manual {
	return cli.Manual{
		Name: "split",
		Summary: "Split a window",
		Synopsis: "[options] <window-id>" +
			" <handler-uri>",
		Options: *c.fs,
	}
}

type wmFloatingCLI struct {
	app     *app
	fs      *cli.FlagSet
	format  string
	align   string
	offsetX int
	offsetY int
}

func newWMFloatingCLI(a *app) *wmFloatingCLI {
	c := &wmFloatingCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl wm floating")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	c.fs.StringVar(
		&c.align, "a", "",
		"Alignment: default|left|right|"+
			"top|bottom|centered",
	)
	c.fs.IntVar(&c.offsetX, "x", 0, "X offset")
	c.fs.IntVar(&c.offsetY, "y", 0, "Y offset")
	return c
}

func (c *wmFloatingCLI) Run(
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
	h, err := resolveHandler(ctx, c.app, rargs[0])
	if err != nil {
		return err
	}
	fh, ok := h.(browserapi.Floating)
	if !ok {
		return fmt.Errorf(
			"handler does not implement Floating",
		)
	}
	alignment, err := parseAlignment(c.align)
	if err != nil {
		return err
	}
	cfg := browserapi.FloatingConfig{
		Alignment: alignment,
		Offset: term.Coordinates{
			X: c.offsetX, Y: c.offsetY,
		},
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	win, err := w.WindowManager(ctx).Floating(fh, cfg)
	if err != nil {
		return err
	}
	return printResult(
		ctx, c.format, win.WindowID(),
		func(v uint64) { fmt.Println(v) },
		[]string{"WindowID"},
	)
}

func (c *wmFloatingCLI) Man() cli.Manual {
	return cli.Manual{
		Name:     "floating",
		Summary:  "Create a floating window",
		Synopsis: "[options] <handler-uri>",
		Options:  *c.fs,
	}
}

func parseAlignment(
	s string,
) (component.Alignment, error) {
	switch s {
	case "", "default":
		return 0, nil
	case "left":
		return component.AlignmentLeft, nil
	case "right":
		return component.AlignmentRight, nil
	case "top":
		return component.AlignmentTop, nil
	case "bottom":
		return component.AlignmentBottom, nil
	case "centered":
		return component.AlignmentCentered, nil
	default:
		return 0, fmt.Errorf(
			"invalid alignment %q", s,
		)
	}
}

type wmSetContentCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newWMSetContentCLI(a *app) *wmSetContentCLI {
	c := &wmSetContentCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl wm set-content")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *wmSetContentCLI) Run(
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
	wid, err := parseWindowID(rargs[0])
	if err != nil {
		return err
	}
	h, err := resolveHandler(ctx, c.app, rargs[1])
	if err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	err = w.WindowManager(ctx).SetWindowContent(wid, h)
	if err != nil {
		return err
	}
	return printOK(c.format)
}

func (c *wmSetContentCLI) Man() cli.Manual {
	return cli.Manual{
		Name: "set-content",
		Summary: "Set window content",
		Synopsis: "[options] <window-id>" +
			" <handler-uri>",
		Options: *c.fs,
	}
}

type wmCloseCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newWMCloseCLI(a *app) *wmCloseCLI {
	c := &wmCloseCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl wm close")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *wmCloseCLI) Run(
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
	wid, err := parseWindowID(rargs[0])
	if err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	err = w.WindowManager(ctx).CloseWindow(wid)
	if err != nil {
		return err
	}
	return printOK(c.format)
}

func (c *wmCloseCLI) Man() cli.Manual {
	return cli.Manual{
		Name:     "close",
		Summary:  "Close a window",
		Synopsis: "[options] <window-id>",
		Options:  *c.fs,
	}
}
