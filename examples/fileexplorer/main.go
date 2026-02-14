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
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/component/fileexplorer"
	fileexplorerhandler "github.com/unstablebuild/rune-go-sdk/handler/fileexplorer"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
	"github.com/unstablebuild/tcell/v3"
)

func main() {
	root := flag.String("root", ".", "root directory to explore")
	flag.Parse()

	absRoot, err := workspaceapi.CurrentUserHostURI(*root)
	if err != nil {
		log.Fatalf("invalid root path: %v", err)
	}

	demo, err := NewExplorerDemo(absRoot)
	if err != nil {
		log.Fatalf("failed to create explorer: %v", err)
	}

	if err := tui.Run(demo); err != nil {
		log.Fatal(err)
	}
}

// ExplorerDemo wraps the file explorer handler and adds a status line.
type ExplorerDemo struct {
	handler       *fileexplorerhandler.Handler
	width, height int
	status        string
}

// NewExplorerDemo creates a new file explorer demo.
func NewExplorerDemo(root workspaceapi.URI) (*ExplorerDemo, error) {
	fs := &localFS{}
	comp, err := fileexplorer.New(fs, root, fileexplorer.Config{
		Icons: map[string]rune{
			"/":   '\uf07b', // folder icon
			".go": '\ue627', // go icon
			".md": '\ue73e', // markdown icon
			"":    '\uf15b', // default file icon
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create explorer: %w", err)
	}

	demo := &ExplorerDemo{
		status: "j/k:navigate, Enter/l:expand, h:collapse, r:rename, a/A:add, d:delete, Ctrl+S:save, q:quit",
	}

	demo.handler = fileexplorerhandler.New(comp,
		fileexplorerhandler.WithOnOpen(func(uri workspaceapi.URI) {
			demo.status = fmt.Sprintf("Selected: %s", uri.String())
		}),
		fileexplorerhandler.WithOnApply(func(err error) {
			if err != nil {
				demo.status = fmt.Sprintf("Error: %v", err)
			} else {
				demo.status = "Changes applied successfully"
			}
		}),
	)

	return demo, nil
}

// Resize satisfies tui.Component.
func (d *ExplorerDemo) Resize(width, height int) {
	d.width, d.height = width, height
	d.handler.Resize(width, height-1) // reserve 1 line for status
}

// Draw satisfies tui.Component.
func (d *ExplorerDemo) Draw(w term.Writer) {
	d.handler.Draw(w)
	d.drawStatusLine(w)
}

// Handle satisfies tui.Handler.
func (d *ExplorerDemo) Handle(ev term.Event) (exit, handled bool) {
	// Update status based on mode
	switch d.handler.Mode() {
	case fileexplorerhandler.ModeEdit:
		d.status = "Edit mode: type name, Enter:commit, Esc:cancel"
	case fileexplorerhandler.ModeConfirm:
		d.status = "Confirm: y/Enter:apply, n/Esc:cancel, arrows:select"
	default:
		if d.status == "Edit mode: type name, Enter:commit, Esc:cancel" ||
			d.status == "Confirm: y/Enter:apply, n/Esc:cancel, arrows:select" {
			d.status = "j/k:navigate, Enter/l:expand, h:collapse, r:rename, a/A:add, d:delete, Ctrl+S:save, q:quit"
		}
	}

	return d.handler.Handle(ev)
}

// Cursor satisfies tui.Handler.
func (d *ExplorerDemo) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	return d.handler.Cursor()
}

// Selection satisfies tui.Handler.
func (d *ExplorerDemo) Selection() (string, bool) {
	return d.handler.Selection()
}

func (d *ExplorerDemo) drawStatusLine(w term.Writer) {
	y := d.height - 1
	for x := range d.width {
		w.SetCell(term.Coordinates{X: x, Y: y}, term.Cell{
			Ch:         ' ',
			Width:      1,
			Attributes: term.Attributes{Bg: tcell.ColorGray},
		})
	}
	statusCells := term.StringToCells(d.status)
	if len(statusCells) == 0 {
		return
	}
	for x, c := range statusCells[0] {
		if x >= d.width {
			break
		}
		c.Bg = tcell.ColorGray
		c.Fg = tcell.ColorWhite
		w.SetCell(term.Coordinates{X: x, Y: y}, c)
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
