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

	"github.com/unstablebuild/rune-go-sdk/cli"
)

type uriCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newURICLI(a *app) *uriCLI {
	c := &uriCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl uri")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *uriCLI) Run(
	ctx context.Context, args []string,
) (retErr error) {
	defer func() {
		retErr = formatError(c.format, retErr)
	}()
	rargs, _, ok, err := cli.ParseUsage(c, c.fs, 1, args)
	if !ok || err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	uri, err := w.FileSystem(ctx).URI(rargs[0])
	if err != nil {
		return err
	}
	return printString(
		c.format, uri.String(), []string{"URI"},
	)
}

func (c *uriCLI) Man() cli.Manual {
	return cli.Manual{
		Name:     "uri",
		Summary:  "Resolve a path to a workspace URI",
		Synopsis: "[options] <path>",
		Options:  *c.fs,
	}
}
