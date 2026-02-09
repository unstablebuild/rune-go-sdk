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
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unstablebuild/rune-go-sdk/api/extensionapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"golang.org/x/oauth2"
)

var errNotInRune = errors.New( //nolint:staticcheck
	"This CLI must be invoked from one of" +
		" Rune's plugin commands (! or !!)",
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "runectl: %s\n", err)
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
		ExtensionID: "runectl",
	}
	w, err := extensionapi.NewWorkspace(cfg, meta)
	if err != nil {
		return nil, err
	}
	a.workspace = w
	return w, nil
}

func looksLikeURI(s string) bool {
	return strings.Contains(s, "://")
}

func (a *app) resolveURIArg(
	ctx context.Context, arg string,
) (workspaceapi.URI, error) {
	if looksLikeURI(arg) {
		return workspaceapi.ParseURI(arg)
	}
	absPath, err := filepath.Abs(arg)
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

func newRootCmd() *cobra.Command {
	a := &app{}

	cmd := &cobra.Command{
		Use:   "runectl",
		Short: "CLI for the Rune IDE",
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: false,
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		newDatadirCmd(a),
		newURICmd(a),
		newOpenCmd(a),
		newNotifyCmd(a),
		newWMCmd(a),
		newStorageCmd(a),
		newEditorCmd(a),
		newSyntaxCmd(a),
		newExecCmd(a),
		newSignalCmd(a),
		newLSPCmd(a),
	)

	return cmd
}
