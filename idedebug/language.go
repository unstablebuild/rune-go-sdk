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
