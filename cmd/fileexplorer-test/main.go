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

// Command fileexplorer-test provides a CLI for testing the fileexplorer
// handler with scripted key sequences.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/component/fileexplorer"
	fileexplorerhandler "github.com/unstablebuild/rune-go-sdk/handler/fileexplorer"
	"github.com/unstablebuild/rune-go-sdk/term/termtest"
)

func main() {
	root := flag.String("root", ".", "root directory to explore")
	width := flag.Int("width", 40, "terminal width")
	height := flag.Int("height", 10, "terminal height")
	flag.Parse()

	absRoot, err := workspaceapi.CurrentUserHostURI(*root)
	if err != nil {
		log.Fatalf("invalid root path: %v", err)
	}

	fs := &localFS{}
	comp, err := fileexplorer.New(fs, absRoot, fileexplorer.Config{
		Icons: map[string]rune{
			"/":   '\uf07b',
			".go": '\ue627',
			".md": '\ue73e',
			"":    '\uf15b',
		},
	})
	if err != nil {
		log.Fatalf("failed to create explorer: %v", err)
	}

	h := fileexplorerhandler.New(comp)
	harness := termtest.NewHarness(h, *width, *height)

	// Print initial state
	fmt.Print(harness.Render())
	fmt.Println()
	fmt.Println("---")

	// Read from stdin
	if err := harness.Run(os.Stdin, os.Stdout); err != nil {
		log.Fatalf("run error: %v", err)
	}
}

// localFS implements workspaceapi.FileSystem using the local filesystem.
type localFS struct{}

func (l *localFS) URI(path string) (workspaceapi.URI, error) {
	return workspaceapi.CurrentUserHostURI(path)
}

func (l *localFS) OpenFile(
	path string, flag int, mode os.FileMode,
) (workspaceapi.File, error) {
	return os.OpenFile(path, flag, mode)
}

func (l *localFS) Remove(path string) error {
	return os.Remove(path)
}

func (l *localFS) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (l *localFS) ReadDir(name string) ([]os.DirEntry, error) {
	return os.ReadDir(name)
}

func (l *localFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
