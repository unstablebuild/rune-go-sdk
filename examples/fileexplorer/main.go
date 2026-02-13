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
	"strings"

	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/component/fileexplorer"
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

// ExplorerDemo implements tui.Handler for the file explorer.
type ExplorerDemo struct {
	explorer      *fileexplorer.Component
	width, height int
	cursor        int
	status        string
	showPrompt    bool
	promptYes     bool
	changes       []fileexplorer.Operation
}

// NewExplorerDemo creates a new file explorer demo.
func NewExplorerDemo(root workspaceapi.URI) (*ExplorerDemo, error) {
	fs := &localFS{}
	exp, err := fileexplorer.New(fs, root, fileexplorer.Config{
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

	return &ExplorerDemo{
		explorer: exp,
		status:   "Enter:expand/collapse, a/A:add file/dir, d:delete, Tab/Shift+Tab:indent, Ctrl+S:save",
	}, nil
}

// Resize satisfies tui.Component.
func (d *ExplorerDemo) Resize(width, height int) {
	d.width, d.height = width, height
	d.explorer.Resize(width, height-1) // reserve 1 line for status
}

// Draw satisfies tui.Component.
func (d *ExplorerDemo) Draw(w term.Writer) {
	d.explorer.Draw(w)
	d.drawCursor(w)
	d.drawStatusLine(w)

	if d.showPrompt {
		d.drawPrompt(w)
	}
}

// Handle satisfies tui.Handler.
func (d *ExplorerDemo) Handle(ev term.Event) (exit, handled bool) {
	if ev.Type != term.EventKey {
		return false, false
	}

	if d.showPrompt {
		return d.handlePrompt(ev)
	}

	switch {
	case ev.Key == term.KeyEsc:
		return true, true

	case ev.Key == term.KeyArrowUp:
		if d.cursor > 0 {
			d.cursor--
		}
		return false, true

	case ev.Key == term.KeyArrowDown:
		cells := d.explorer.Cells()
		if d.cursor < len(cells)-1 {
			d.cursor++
		}
		return false, true

	case ev.Key == term.KeyEnter:
		d.handleEnter()
		return false, true

	case ev.Ch == 's' && ev.Mod == term.ModCtrl:
		d.showChangesPrompt()
		return false, true

	case ev.Ch == 'r' && ev.Mod == term.ModCtrl:
		if err := d.explorer.Init(); err != nil {
			d.status = fmt.Sprintf("Reload error: %v", err)
		} else {
			d.status = "Reloaded from filesystem"
			d.cursor = 0
		}
		return false, true

	case ev.Ch == 'd':
		d.deleteLine()
		return false, true

	case ev.Ch == 'a':
		d.addFile()
		return false, true

	case ev.Ch == 'A':
		d.addDir()
		return false, true

	case ev.Key == term.KeyTab && ev.Mod == 0:
		d.indentLine()
		return false, true

	case ev.Key == term.KeyTab && ev.Mod == term.ModShift:
		d.dedentLine()
		return false, true
	}

	return false, false
}

// Cursor satisfies tui.Handler.
func (d *ExplorerDemo) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	return term.Coordinates{X: 0, Y: d.cursor}, term.CursorStyleSteadyBar, false
}

// Selection satisfies tui.Handler.
func (d *ExplorerDemo) Selection() (string, bool) {
	return "", false
}

func (d *ExplorerDemo) handleEnter() {
	uri, isFile := d.explorer.ExpandNodeAt(
		term.Coordinates{Y: d.cursor},
	)
	if isFile {
		d.status = fmt.Sprintf("Selected: %s", uri.String())
	} else {
		cells := d.explorer.Cells()
		if d.cursor >= len(cells) && len(cells) > 0 {
			d.cursor = len(cells) - 1
		}
		d.status = "Press Enter to expand/collapse, Ctrl+S to show changes"
	}
}

func (d *ExplorerDemo) showChangesPrompt() {
	d.changes = d.explorer.Changes()
	if len(d.changes) == 0 {
		d.status = "No changes detected"
		return
	}
	d.showPrompt = true
	d.promptYes = true
}

func (d *ExplorerDemo) handlePrompt(
	ev term.Event,
) (exit, handled bool) {
	switch {
	case ev.Key == term.KeyEsc:
		d.showPrompt = false
		d.status = "Cancelled"
		return false, true

	case ev.Key == term.KeyArrowLeft,
		ev.Key == term.KeyArrowRight,
		ev.Key == term.KeyTab:
		d.promptYes = !d.promptYes
		return false, true

	case ev.Key == term.KeyEnter:
		d.showPrompt = false
		if d.promptYes {
			d.executeChanges()
		} else {
			d.status = "Cancelled"
		}
		return false, true

	case ev.Ch == 'y' || ev.Ch == 'Y':
		d.showPrompt = false
		d.executeChanges()
		return false, true

	case ev.Ch == 'n' || ev.Ch == 'N':
		d.showPrompt = false
		d.status = "Cancelled"
		return false, true
	}

	return false, false
}

func (d *ExplorerDemo) executeChanges() {
	var msgs []string
	for _, op := range d.changes {
		var msg string
		switch op.Type {
		case fileexplorer.OpCreate:
			msg = fmt.Sprintf("Create: %s", op.NewURI.Path())
		case fileexplorer.OpMkdir:
			msg = fmt.Sprintf("Mkdir: %s", op.NewURI.Path())
		case fileexplorer.OpDelete:
			msg = fmt.Sprintf("Delete: %s", op.URI.Path())
		case fileexplorer.OpRename:
			msg = fmt.Sprintf("Rename: %s -> %s",
				op.URI.Name(), op.NewURI.Name())
		}
		msgs = append(msgs, msg)
	}
	d.status = fmt.Sprintf("Would execute: %s", strings.Join(msgs, "; "))
}

func (d *ExplorerDemo) deleteLine() {
	cells := d.explorer.Cells()
	if len(cells) == 0 || d.cursor >= len(cells) {
		return
	}
	newCells := append(cells[:d.cursor], cells[d.cursor+1:]...)
	d.explorer.SetCells(newCells)
	if d.cursor >= len(newCells) && len(newCells) > 0 {
		d.cursor = len(newCells) - 1
	}
	d.status = "Deleted line (Ctrl+S to review changes)"
}

func (d *ExplorerDemo) addFile() {
	cells := d.explorer.Cells()
	if len(cells) == 0 {
		return
	}
	depth := d.depthAtCursor(cells)
	newLine := d.buildLine(depth, "newfile.txt", false)
	pos := min(d.cursor+1, len(cells))
	newCells := make([][]term.Cell, 0, len(cells)+1)
	newCells = append(newCells, cells[:pos]...)
	newCells = append(newCells, newLine...)
	newCells = append(newCells, cells[pos:]...)
	d.explorer.SetCells(newCells)
	d.cursor = pos
	d.status = "Added newfile.txt (Ctrl+S to review changes)"
}

func (d *ExplorerDemo) addDir() {
	cells := d.explorer.Cells()
	if len(cells) == 0 {
		return
	}
	depth := d.depthAtCursor(cells)
	newLine := d.buildLine(depth, "newdir", true)
	pos := min(d.cursor+1, len(cells))
	newCells := make([][]term.Cell, 0, len(cells)+1)
	newCells = append(newCells, cells[:pos]...)
	newCells = append(newCells, newLine...)
	newCells = append(newCells, cells[pos:]...)
	d.explorer.SetCells(newCells)
	d.cursor = pos
	d.status = "Added newdir/ (Ctrl+S to review changes)"
}

func (d *ExplorerDemo) depthAtCursor(cells [][]term.Cell) int {
	if d.cursor >= len(cells) {
		return 0
	}
	row := cells[d.cursor]
	depth := 0
	i := 0
	for i+1 < len(row) {
		if row[i].Ch == '│' && row[i+1].Ch == ' ' {
			depth++
			i += 2
		} else {
			break
		}
	}
	return depth
}

func (d *ExplorerDemo) buildLine(depth int, name string, isDir bool) [][]term.Cell {
	var b strings.Builder
	for range depth {
		b.WriteString("│ ")
	}
	b.WriteString("  ") // icon placeholder + space
	b.WriteString(name)
	if isDir {
		b.WriteRune('/')
	}
	return term.StringToCells(b.String())
}

func (d *ExplorerDemo) indentLine() {
	cells := d.explorer.Cells()
	if d.cursor >= len(cells) {
		return
	}

	// Insert "│ " at the beginning of the line
	indent := term.StringToCells("│ ")[0]
	row := cells[d.cursor]
	newRow := make([]term.Cell, 0, len(row)+2)
	newRow = append(newRow, indent...)
	newRow = append(newRow, row...)

	cells[d.cursor] = newRow
	d.explorer.SetCells(cells)
	d.status = "Indented line (Tab to indent more, Shift+Tab to dedent)"
}

func (d *ExplorerDemo) dedentLine() {
	cells := d.explorer.Cells()
	if d.cursor >= len(cells) {
		return
	}

	row := cells[d.cursor]
	// Check if line starts with "│ " and remove it
	if len(row) >= 2 && row[0].Ch == '│' && row[1].Ch == ' ' {
		cells[d.cursor] = row[2:]
		d.explorer.SetCells(cells)
		d.status = "Dedented line (Tab to indent, Shift+Tab to dedent more)"
	} else {
		d.status = "Cannot dedent further (already at root level)"
	}
}

func (d *ExplorerDemo) drawCursor(w term.Writer) {
	cells := d.explorer.Cells()
	if d.cursor >= len(cells) {
		return
	}
	row := cells[d.cursor]
	for x := range row {
		if x >= d.width {
			break
		}
		w.UnionAttributes(
			term.Coordinates{X: x, Y: d.cursor},
			term.Attributes{Attrs: tcell.AttrReverse},
		)
	}
	// fill rest of line with reverse
	for x := len(row); x < d.width; x++ {
		w.UnionAttributes(
			term.Coordinates{X: x, Y: d.cursor},
			term.Attributes{Attrs: tcell.AttrReverse},
		)
	}
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
	for x, c := range statusCells[0] {
		if x >= d.width {
			break
		}
		c.Bg = tcell.ColorGray
		c.Fg = tcell.ColorWhite
		w.SetCell(term.Coordinates{X: x, Y: y}, c)
	}
}

func (d *ExplorerDemo) drawPrompt(w term.Writer) {
	// Build prompt content
	var lines []string
	lines = append(lines, "Apply the following changes?")
	lines = append(lines, "")
	for _, op := range d.changes {
		var line string
		switch op.Type {
		case fileexplorer.OpCreate:
			line = fmt.Sprintf("  + Create: %s", op.NewURI.Name())
		case fileexplorer.OpMkdir:
			line = fmt.Sprintf("  + Mkdir:  %s/", op.NewURI.Name())
		case fileexplorer.OpDelete:
			line = fmt.Sprintf("  - Delete: %s", op.URI.Name())
		case fileexplorer.OpRename:
			line = fmt.Sprintf("  ~ Rename: %s -> %s",
				op.URI.Name(), op.NewURI.Name())
		}
		lines = append(lines, line)
	}
	lines = append(lines, "")

	// Calculate dimensions
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	promptWidth := maxWidth + 4
	promptHeight := len(lines) + 3 // +3 for padding and buttons

	// Center the prompt
	startX := (d.width - promptWidth) / 2
	startY := (d.height - promptHeight) / 2
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	// Draw background
	for y := startY; y < startY+promptHeight && y < d.height; y++ {
		for x := startX; x < startX+promptWidth && x < d.width; x++ {
			w.SetCell(term.Coordinates{X: x, Y: y}, term.Cell{
				Ch:         ' ',
				Width:      1,
				Attributes: term.Attributes{Bg: tcell.ColorDarkBlue},
			})
		}
	}

	// Draw border
	d.drawBox(w, startX, startY, promptWidth, promptHeight)

	// Draw content
	for i, line := range lines {
		y := startY + 1 + i
		if y >= d.height-1 {
			break
		}
		cells := term.StringToCells(line)
		for x, c := range cells[0] {
			px := startX + 2 + x
			if px >= startX+promptWidth-1 {
				break
			}
			c.Bg = tcell.ColorDarkBlue
			c.Fg = tcell.ColorWhite
			w.SetCell(term.Coordinates{X: px, Y: y}, c)
		}
	}

	// Draw buttons
	buttonY := startY + promptHeight - 2
	yesX := startX + promptWidth/2 - 8
	noX := startX + promptWidth/2 + 2

	d.drawButton(w, yesX, buttonY, "[Yes]", d.promptYes)
	d.drawButton(w, noX, buttonY, "[No]", !d.promptYes)
}

func (d *ExplorerDemo) drawBox(
	w term.Writer, x, y, width, height int,
) {
	attrs := term.Attributes{
		Bg: tcell.ColorDarkBlue,
		Fg: tcell.ColorWhite,
	}

	// corners
	w.SetCell(term.Coordinates{X: x, Y: y},
		term.Cell{Ch: '┌', Width: 1, Attributes: attrs})
	w.SetCell(term.Coordinates{X: x + width - 1, Y: y},
		term.Cell{Ch: '┐', Width: 1, Attributes: attrs})
	w.SetCell(term.Coordinates{X: x, Y: y + height - 1},
		term.Cell{Ch: '└', Width: 1, Attributes: attrs})
	w.SetCell(term.Coordinates{X: x + width - 1, Y: y + height - 1},
		term.Cell{Ch: '┘', Width: 1, Attributes: attrs})

	// horizontal lines
	for i := x + 1; i < x+width-1; i++ {
		w.SetCell(term.Coordinates{X: i, Y: y},
			term.Cell{Ch: '─', Width: 1, Attributes: attrs})
		w.SetCell(term.Coordinates{X: i, Y: y + height - 1},
			term.Cell{Ch: '─', Width: 1, Attributes: attrs})
	}

	// vertical lines
	for i := y + 1; i < y+height-1; i++ {
		w.SetCell(term.Coordinates{X: x, Y: i},
			term.Cell{Ch: '│', Width: 1, Attributes: attrs})
		w.SetCell(term.Coordinates{X: x + width - 1, Y: i},
			term.Cell{Ch: '│', Width: 1, Attributes: attrs})
	}
}

func (d *ExplorerDemo) drawButton(
	w term.Writer, x, y int, label string, selected bool,
) {
	attrs := term.Attributes{
		Bg: tcell.ColorDarkBlue,
		Fg: tcell.ColorWhite,
	}
	if selected {
		attrs.Attrs = tcell.AttrReverse
	}
	cells := term.StringToCells(label)
	for i, c := range cells[0] {
		c.Attributes = attrs
		w.SetCell(term.Coordinates{X: x + i, Y: y}, c)
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
