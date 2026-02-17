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

package idedebug

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

// PkgManager abstracts the ability to resolve package
// directories for a given package identifier.
type PkgManager interface {
	// LibDir returns an iterator of directory paths where
	// the package's binaries may be found.
	LibDir(ctx context.Context, pkgID string) (
		iterator.Iterator[string], error,
	)
}

type debugConfig struct {
	id      string
	command string
	args    []string
}

var debugAdapters = map[string]debugConfig{
	"go": {
		id:      "go",
		command: "dlv",
		args:    []string{"dap"},
	},
}

func debugAdapterForFile(
	filename workspaceapi.URI,
) (debugConfig, error) {
	return debugAdapterForFilename(filename.Path())
}

func doDebugAdapterForFile(
	filename string,
) (string, error) {
	ext := filepath.Ext(filename)
	switch ext {
	case ".go":
		return "go", nil
	default:
		return "", errors.New("unsupported language")
	}
}

func debugAdapterForFilename(
	filename string,
) (debugConfig, error) {
	id, err := doDebugAdapterForFile(
		filepath.Base(filename),
	)
	if err != nil {
		return debugConfig{}, err
	}
	cfg, ok := debugAdapters[id]
	if !ok {
		return debugConfig{}, fmt.Errorf(
			"%s language debug adapter is not supported yet",
			id,
		)
	}
	return cfg, nil
}
