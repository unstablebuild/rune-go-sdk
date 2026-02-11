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

	"github.com/unstablebuild/rune-go-sdk/api/browserapi/browserrpc"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/cli"
)

type openCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newOpenCLI(a *app) *openCLI {
	c := &openCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl open")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *openCLI) Run(
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
	parsed, err := workspaceapi.ParseURI(rargs[0])
	if err != nil {
		return err
	}
	h, err := w.ResourceOpener(ctx).Open(parsed)
	if err != nil {
		return err
	}
	uri := rargs[0]
	if tok, ok := h.(browserrpc.Token); ok {
		uri = tok.URI
	}
	return printString(
		c.format, uri, []string{"Resource"},
	)
}

func (c *openCLI) Man() cli.Manual {
	return cli.Manual{
		Name:     "open",
		Summary:  "Open a resource",
		Synopsis: "[options] <uri>",
		Options:  *c.fs,
	}
}
