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
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"github.com/unstablebuild/rune-go-sdk/api/extensionapi"
	"github.com/unstablebuild/rune-go-sdk/cli"
	"golang.org/x/oauth2"
)

var errNotInRune = errors.New( //nolint:staticcheck
	"This CLI must be invoked from one of" +
		" Rune's plugin commands (! or !!)",
)

func main() {
	root := newRootCLI()
	if err := cli.Run(context.Background(), root); err != nil {
		fmt.Fprintf(os.Stderr, "runectrl: %s\n", err)
		os.Exit(1)
	}
}

type app struct {
	workspace *extensionapi.Workspace
}

func (a *app) getWorkspace() (
	*extensionapi.Workspace, error,
) {
	if a.workspace != nil {
		return a.workspace, nil
	}
	socket := os.Getenv("RUNE_SOCKET")
	if socket == "" {
		return nil, errNotInRune
	}
	datadir := os.Getenv("RUNE_DATADIR")
	if datadir == "" {
		return nil, errNotInRune
	}
	cfg := extensionapi.Config{
		Socket:  socket,
		DataDir: datadir,
	}
	if tok := os.Getenv("RUNE_TOKEN"); tok != "" {
		cfg.Token = &oauth2.Token{AccessToken: tok}
	}
	if cert := os.Getenv("RUNE_CERT"); cert != "" {
		data, err := base64.StdEncoding.DecodeString(cert)
		if err != nil {
			return nil, fmt.Errorf("decode cert: %w", err)
		}
		cfg.Certificate = data
	}
	meta := extensionapi.Metadata{
		ExtensionID: "runectrl",
	}
	w, err := extensionapi.NewWorkspace(cfg, meta)
	if err != nil {
		return nil, err
	}
	a.workspace = w
	return w, nil
}

type rootCLI struct {
	app  *app
	fs   *cli.FlagSet
	cmds map[string]cli.CLI
}

func newRootCLI() *rootCLI {
	a := &app{}
	fs := cli.NewFlagSet("runectrl")
	cmds := map[string]cli.CLI{
		"datadir": newDatadirCLI(a),
		"uri":     newURICLI(a),
		"open":    newOpenCLI(a),
		"notify":  newNotifyCLI(a),
		"wm":      newWMCLI(a),
		"storage": newStorageCLI(a),
		"editor":  newEditorCLI(a),
		"syntax":  newSyntaxCLI(a),
	}
	return &rootCLI{app: a, fs: fs, cmds: cmds}
}

func (r *rootCLI) Run(
	ctx context.Context, args []string,
) error {
	return cli.ParseAndRunCommand(
		ctx, r, r.fs, r.cmds, args,
	)
}

func (r *rootCLI) Man() cli.Manual {
	var subMan []cli.Manual
	for _, cmd := range []string{
		"datadir", "uri", "open", "notify",
		"wm", "storage", "editor", "syntax",
	} {
		subMan = append(subMan, r.cmds[cmd].Man())
	}
	return cli.Manual{
		Name:     "runectrl",
		Summary:  "CLI for Rune workspace APIs",
		Synopsis: "<command>",
		Commands: subMan,
		Options:  *r.fs,
	}
}
