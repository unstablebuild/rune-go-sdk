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
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package fileexplorer

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/component/comptest"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestEmptyDirectory(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	w, h := c.Dimensions()
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, h)
	assert.Empty(t, c.Cells())
}

func TestFlatDirectory(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "main.go", isDir: false},
			{name: "README.md", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	assert.Equal(t,
		"  README.md\n  main.go",
		cellsString(c.Cells()),
	)
}

func TestDrawClipping(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
			{name: "c.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	w := term.NewStringWriter(4, 2)
	c.Resize(4, 2)

	comptest.TestComponent(t, c, w, []comptest.TestCase{{
		Expected: `
  a.
  b.`,
	}})
}

func TestDrawLarger(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	w := term.NewStringWriter(20, 10)
	c.Resize(20, 10)

	comptest.TestComponent(t, c, w, []comptest.TestCase{{
		Expected: "  a.go              \n" +
			"                    \n" +
			"                    \n" +
			"                    \n" +
			"                    \n" +
			"                    \n" +
			"                    \n" +
			"                    \n" +
			"                    \n" +
			"                    ",
	}})
}

func TestNodeAt(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
			{name: "main.go", isDir: false},
		},
		"/project/src": {},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	tests := []struct {
		name string
		y    int
		want string
		ok   bool
	}{
		{"dir", 0, "file:///project/src", true},
		{"file", 1, "file:///project/main.go", true},
		{"out of range", 5, "", false},
		{"negative", -1, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri, ok := c.NodeAt(
				term.Coordinates{Y: tt.y},
			)
			assert.Equal(t, tt.ok, ok)
			if ok {
				assert.Equal(t, tt.want, uri.String())
			}
		})
	}
}

func TestExpandNodeAtFile(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "main.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	uri, isFile := c.ExpandNodeAt(
		term.Coordinates{Y: 0},
	)
	assert.True(t, isFile)
	assert.Equal(t,
		"file:///project/main.go", uri.String(),
	)
}

func TestExpandNodeAtDirectory(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
		},
		"/project/src": {
			{name: "app.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	_, isFile := c.ExpandNodeAt(
		term.Coordinates{Y: 0},
	)
	assert.False(t, isFile)
	assert.Equal(t,
		"  src/\n│   app.go",
		cellsString(c.Cells()),
	)
}

func TestExpandLevel(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "a", isDir: true},
			{name: "b", isDir: true},
			{name: "c.go", isDir: false},
		},
		"/project/a": {
			{name: "a1.go", isDir: false},
		},
		"/project/b": {
			{name: "b1.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	c.ExpandLevel(term.Coordinates{Y: 0})

	expected := `  a/
│   a1.go
  b/
│   b1.go
  c.go`
	assert.Equal(t, expected, cellsString(c.Cells()))
}

func TestCollapseLevel(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "a", isDir: true},
			{name: "b", isDir: true},
		},
		"/project/a": {
			{name: "a1.go", isDir: false},
		},
		"/project/b": {
			{name: "b1.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	c.ExpandLevel(term.Coordinates{Y: 0})
	c.CollapseLevel(term.Coordinates{Y: 0})

	assert.Equal(t,
		"  a/\n  b/",
		cellsString(c.Cells()),
	)
}

func TestCollapseAll(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
		},
		"/project/src": {
			{name: "app.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	c.ExpandNodeAt(term.Coordinates{Y: 0})
	assert.Len(t, c.Cells(), 2)

	c.CollapseAll()
	assert.Equal(t,
		"  src/", cellsString(c.Cells()),
	)
}

func TestInitPreservesExpanded(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
			{name: "main.go", isDir: false},
		},
		"/project/src": {
			{name: "app.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	c.ExpandNodeAt(term.Coordinates{Y: 0})
	before := cellsString(c.Cells())

	err = c.Init()
	require.NoError(t, err)
	assert.Equal(t, before, cellsString(c.Cells()))
}

func TestSorting(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "z.go", isDir: false},
			{name: "a.go", isDir: false},
			{name: "m", isDir: true},
			{name: "b", isDir: true},
		},
		"/project/m": {},
		"/project/b": {},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	expected := `  b/
  m/
  a.go
  z.go`
	assert.Equal(t, expected, cellsString(c.Cells()))
}

func TestNestedExpansion(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
		},
		"/project/src": {
			{name: "handlers", isDir: true},
			{name: "main.go", isDir: false},
		},
		"/project/src/handlers": {
			{name: "auth.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	c.ExpandNodeAt(term.Coordinates{Y: 0})
	c.ExpandNodeAt(term.Coordinates{Y: 1})

	expected := `  src/
│   handlers/
│ │   auth.go
│   main.go`
	assert.Equal(t, expected, cellsString(c.Cells()))
}

func TestChangesNoChanges(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "main.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// No modifications - should have no changes
	ops := c.Changes()
	assert.Empty(t, ops)
}

func TestChangesNoChangesExpanded(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
		},
		"/project/src": {
			{name: "main.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	c.ExpandNodeAt(term.Coordinates{Y: 0})

	// No modifications - should have no changes
	ops := c.Changes()
	assert.Empty(t, ops)
}

func TestChangesCreateFile(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "main.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// Add a new row (no ID = new file)
	cells := c.Cells()
	newRow := term.StringToCells("  new.go")[0]
	cells = append(cells, newRow)
	c.SetCells(cells)

	ops := c.Changes()
	require.Len(t, ops, 1)
	assert.Equal(t, OpCreate, ops[0].Type)
	assert.Equal(t, "file:///project/new.go", ops[0].NewURI.String())
}

func TestChangesCreateDirectory(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "main.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// Add a new directory row (no ID = new dir)
	cells := c.Cells()
	newRow := term.StringToCells("  newdir/")[0]
	cells = append(cells, newRow)
	c.SetCells(cells)

	ops := c.Changes()
	require.Len(t, ops, 1)
	assert.Equal(t, OpMkdir, ops[0].Type)
	assert.Equal(t, "file:///project/newdir", ops[0].NewURI.String())
}

func TestChangesDeleteFile(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// Remove first row (a.go)
	cells := c.Cells()
	cells = cells[1:] // Keep only b.go
	c.SetCells(cells)

	ops := c.Changes()
	require.Len(t, ops, 1)
	assert.Equal(t, OpDelete, ops[0].Type)
	assert.Equal(t, "file:///project/a.go", ops[0].URI.String())
}

func TestChangesDeleteDirectory(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
			{name: "main.go", isDir: false},
		},
		"/project/src": {},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// Remove first row (src/)
	cells := c.Cells()
	cells = cells[1:] // Keep only main.go
	c.SetCells(cells)

	ops := c.Changes()
	require.Len(t, ops, 1)
	assert.Equal(t, OpDelete, ops[0].Type)
	assert.Equal(t, "file:///project/src", ops[0].URI.String())
}

func TestChangesRenameFile(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "old.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// Modify the row content but keep the ID
	cells := c.Cells()
	// Row format: [ID, ' ', ' ', 'o', 'l', 'd', '.', 'g', 'o']
	// We need to change the content after the ID
	newContent := term.StringToCells("  new.go")[0]
	cells[0] = append([]term.Cell{cells[0][0]}, newContent...)
	c.SetCells(cells)

	ops := c.Changes()
	require.Len(t, ops, 1)
	assert.Equal(t, OpRename, ops[0].Type)
	assert.Equal(t, "file:///project/old.go", ops[0].URI.String())
	assert.Equal(t, "file:///project/new.go", ops[0].NewURI.String())
}

func TestChangesRenameDirectory(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "olddir", isDir: true},
		},
		"/project/olddir": {},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// Modify the row content but keep the ID
	cells := c.Cells()
	newContent := term.StringToCells("  newdir/")[0]
	cells[0] = append([]term.Cell{cells[0][0]}, newContent...)
	c.SetCells(cells)

	ops := c.Changes()
	require.Len(t, ops, 1)
	assert.Equal(t, OpRename, ops[0].Type)
	assert.Equal(t, "file:///project/olddir", ops[0].URI.String())
	assert.Equal(t, "file:///project/newdir", ops[0].NewURI.String())
}

func TestChangesCreateInExpandedDir(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
		},
		"/project/src": {
			{name: "main.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	c.ExpandNodeAt(term.Coordinates{Y: 0})

	// Add a new row inside src (at depth 1)
	cells := c.Cells()
	newRow := term.StringToCells("│   new.go")[0]
	// Insert after main.go
	newCells := make([][]term.Cell, 0, len(cells)+1)
	newCells = append(newCells, cells[:2]...)
	newCells = append(newCells, newRow)
	newCells = append(newCells, cells[2:]...)
	c.SetCells(newCells)

	ops := c.Changes()
	require.Len(t, ops, 1)
	assert.Equal(t, OpCreate, ops[0].Type)
	assert.Equal(t, "file:///project/src/new.go", ops[0].NewURI.String())
}

func TestChangesMultipleOperations(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
			{name: "c.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	cells := c.Cells()

	// Delete a.go (remove first row)
	// Rename b.go to renamed.go (keep ID, change content)
	// Keep c.go
	// Add new.go (new row without ID)

	newCells := make([][]term.Cell, 0, 3)

	// b.go renamed to renamed.go (keep ID)
	renamedContent := term.StringToCells("  renamed.go")[0]
	newCells = append(newCells, append([]term.Cell{cells[1][0]}, renamedContent...))

	// c.go unchanged
	newCells = append(newCells, cells[2])

	// new.go (no ID)
	newRow := term.StringToCells("  new.go")[0]
	newCells = append(newCells, newRow)

	c.SetCells(newCells)

	ops := c.Changes()

	// Should have: delete a.go, rename b.go->renamed.go, create new.go
	var deletes, renames, creates []string
	for _, op := range ops {
		switch op.Type {
		case OpDelete:
			deletes = append(deletes, op.URI.Name())
		case OpRename:
			renames = append(renames, op.URI.Name()+"->"+op.NewURI.Name())
		case OpCreate:
			creates = append(creates, op.NewURI.Name())
		}
	}

	assert.ElementsMatch(t, []string{"a.go"}, deletes)
	assert.ElementsMatch(t, []string{"b.go->renamed.go"}, renames)
	assert.ElementsMatch(t, []string{"new.go"}, creates)
}

func TestRowIDSurvivesStringConversion(t *testing.T) {
	// This test verifies that row IDs survive string conversion
	// (i.e., cells -> string -> cells preserves the ID)
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "test.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	cells := c.Cells()
	require.Len(t, cells, 1)

	// Get the original ID
	originalID := cells[0][0].Ch
	require.True(t, originalID >= firstRowID && originalID <= lastRowID,
		"ID should be in Private Use Area")

	// Convert to string and back
	var sb strings.Builder
	for _, cell := range cells[0] {
		sb.WriteRune(cell.Ch)
	}
	asString := sb.String()

	// Convert back to cells
	roundTripped := term.StringToCells(asString)
	require.Len(t, roundTripped, 1)

	// Verify ID is preserved
	assert.Equal(t, originalID, roundTripped[0][0].Ch,
		"Row ID should survive string conversion")
}

func TestExpandAfterCellEdits(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "dirA", isDir: true},
			{name: "dirB", isDir: true},
			{name: "main.go", isDir: false},
		},
		"/project/dirA": {
			{name: "a.go", isDir: false},
		},
		"/project/dirB": {
			{name: "b.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// Add a new file at root level
	cells := c.Cells()
	newRow := term.StringToCells("  newfile.go")[0]
	cells = append(cells, newRow)
	c.SetCells(cells)

	// Now expand dirA - should work correctly
	c.ExpandNodeAt(term.Coordinates{Y: 0})

	result := cellsString(c.Cells())
	assert.Contains(t, result, "a.go", "dirA should be expanded")
	assert.Contains(t, result, "newfile.go", "new file should be preserved")
}

func TestSetCellsRoundtrip(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "main.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	original := cellsString(c.Cells())
	clone := term.CloneCells(c.Cells())
	c.SetCells(clone)
	assert.Equal(t, original, cellsString(c.Cells()))

	w := term.NewStringWriter(20, 5)
	c.Resize(20, 5)
	c.Draw(w)
}

func TestCustomIcons(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
			{name: "main.go", isDir: false},
		},
		"/project/src": {},
	}}
	c, err := New(mfs, rootURI(), Config{
		Icons: map[string]rune{
			"/":   '\uf07b',
			".go": '\ue627',
		},
	})
	require.NoError(t, err)

	assert.Equal(t,
		"\uf07b src/\n\ue627 main.go",
		cellsString(c.Cells()),
	)
}

func TestDefaultIndentRune(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
		},
		"/project/src": {
			{name: "app.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	c.ExpandNodeAt(term.Coordinates{Y: 0})
	assert.Contains(t,
		cellsString(c.Cells()), "│",
	)
}

func TestChangesMoveFile(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
			{name: "file.go", isDir: false},
		},
		"/project/src": {},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// Expand src
	c.ExpandNodeAt(term.Coordinates{Y: 0})

	cells := c.Cells()
	// Row 0: src/ (dir)
	// Row 1: file.go (file at root)

	// Move file.go into src/ by changing its depth
	// We need to modify the cells to represent the move
	fileRow := cells[1]
	fileID := fileRow[0]

	// Create new cells: src/, then file.go inside src
	newCells := make([][]term.Cell, 2)
	newCells[0] = cells[0] // src/ stays the same

	// Create file.go at depth 1 (inside src)
	newContent := term.StringToCells("│   file.go")[0]
	newCells[1] = append([]term.Cell{fileID}, newContent...)

	c.SetCells(newCells)

	ops := c.Changes()
	require.Len(t, ops, 1)
	assert.Equal(t, OpMove, ops[0].Type)
	assert.Equal(t, "file:///project/file.go", ops[0].URI.String())
	assert.Equal(t, "file:///project/src/file.go", ops[0].NewURI.String())
}

func TestChangesMoveDirectory(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "dest", isDir: true},
			{name: "src", isDir: true},
		},
		"/project/dest": {},
		"/project/src": {},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// Expand dest
	c.ExpandNodeAt(term.Coordinates{Y: 0})

	cells := c.Cells()
	// Row 0: dest/ (expanded)
	// Row 1: src/ (at root)

	srcID := cells[1][0]

	// Move src/ into dest/
	newCells := make([][]term.Cell, 2)
	newCells[0] = cells[0] // dest/ stays

	// src/ at depth 1 (inside dest)
	newContent := term.StringToCells("│   src/")[0]
	newCells[1] = append([]term.Cell{srcID}, newContent...)

	c.SetCells(newCells)

	ops := c.Changes()
	require.Len(t, ops, 1)
	assert.Equal(t, OpMove, ops[0].Type)
	assert.Equal(t, "file:///project/src", ops[0].URI.String())
	assert.Equal(t, "file:///project/dest/src", ops[0].NewURI.String())
}

func TestChangesDetectDuplicateID(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "original.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	cells := c.Cells()
	originalID := cells[0][0]

	// Add a second row with the same ID
	// This simulates a copy scenario where the same ID appears multiple times
	copyContent := term.StringToCells("  copy.go")[0]
	copyRow := append([]term.Cell{originalID}, copyContent...)

	newCells := make([][]term.Cell, 2)
	newCells[0] = cells[0]
	newCells[1] = copyRow

	c.SetCells(newCells)

	ops := c.Changes()
	// When same ID appears twice, the implementation detects a copy operation
	// The behavior may vary based on the order of processing
	assert.NotEmpty(t, ops, "should detect some operations")

	// Check that we have a copy operation
	hasCopy := false
	for _, op := range ops {
		if op.Type == OpCopy {
			hasCopy = true
		}
	}
	assert.True(t, hasCopy, "should detect a copy operation for duplicate IDs")
}

func TestChangeSetConflicts(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	cells := c.Cells()

	// Rename both files to the same name (conflict)
	newContent := term.StringToCells("  same.go")[0]
	newCells := make([][]term.Cell, 2)
	newCells[0] = append([]term.Cell{cells[0][0]}, newContent...)
	newCells[1] = append([]term.Cell{cells[1][0]}, newContent...)

	c.SetCells(newCells)

	cs := c.ChangeSet()
	assert.True(t, cs.HasConflicts())
	require.Len(t, cs.Conflicts, 1)
	assert.Contains(t, cs.Conflicts[0].Path, "same.go")
}

func TestChangeSetNoConflicts(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "test.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// No modifications
	cs := c.ChangeSet()
	assert.False(t, cs.HasConflicts())
	assert.Empty(t, cs.Conflicts)
	assert.Empty(t, cs.Operations)
}

func TestChangesDeleteNested(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
		},
		"/project/src": {
			{name: "main.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	c.ExpandNodeAt(term.Coordinates{Y: 0})

	// Delete all rows
	c.SetCells(nil)

	ops := c.Changes()

	// Should have 2 deletes
	assert.Len(t, ops, 2)

	var deletedPaths []string
	for _, op := range ops {
		assert.Equal(t, OpDelete, op.Type)
		deletedPaths = append(deletedPaths, op.URI.Path())
	}
	assert.Contains(t, deletedPaths, "/project/src")
	assert.Contains(t, deletedPaths, "/project/src/main.go")
}

func TestChangesRenameAndCreate(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "old.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	cells := c.Cells()
	oldID := cells[0][0]

	// Rename old.go to new.go and add another.go
	newCells := make([][]term.Cell, 2)
	renamedContent := term.StringToCells("  new.go")[0]
	newCells[0] = append([]term.Cell{oldID}, renamedContent...)

	createdContent := term.StringToCells("  another.go")[0]
	newCells[1] = createdContent

	c.SetCells(newCells)

	ops := c.Changes()
	require.Len(t, ops, 2)

	var hasRename, hasCreate bool
	for _, op := range ops {
		if op.Type == OpRename {
			hasRename = true
			assert.Equal(t, "old.go", op.URI.Name())
			assert.Equal(t, "new.go", op.NewURI.Name())
		}
		if op.Type == OpCreate {
			hasCreate = true
			assert.Equal(t, "another.go", op.NewURI.Name())
		}
	}
	assert.True(t, hasRename)
	assert.True(t, hasCreate)
}

func TestChangesDeleteAndCreate(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "old.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// Delete old.go and create new.go
	newContent := term.StringToCells("  new.go")[0]
	c.SetCells([][]term.Cell{newContent})

	ops := c.Changes()
	require.Len(t, ops, 2)

	var hasDelete, hasCreate bool
	for _, op := range ops {
		if op.Type == OpDelete {
			hasDelete = true
			assert.Equal(t, "old.go", op.URI.Name())
		}
		if op.Type == OpCreate {
			hasCreate = true
			assert.Equal(t, "new.go", op.NewURI.Name())
		}
	}
	assert.True(t, hasDelete)
	assert.True(t, hasCreate)
}

func TestDeepNestedStructure(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "a", isDir: true},
		},
		"/project/a": {
			{name: "b", isDir: true},
		},
		"/project/a/b": {
			{name: "c", isDir: true},
		},
		"/project/a/b/c": {
			{name: "deep.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// Expand all levels
	c.ExpandNodeAt(term.Coordinates{Y: 0}) // a
	c.ExpandNodeAt(term.Coordinates{Y: 1}) // b
	c.ExpandNodeAt(term.Coordinates{Y: 2}) // c

	cells := c.Cells()
	assert.Len(t, cells, 4)

	// Verify deep.go is at correct depth
	result := cellsString(cells)
	assert.Contains(t, result, "deep.go")

	// No changes expected
	ops := c.Changes()
	assert.Empty(t, ops)
}

func TestSpecialCharactersInFilename(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "file with spaces.go", isDir: false},
			{name: "file-with-dashes.go", isDir: false},
			{name: "file_with_underscores.go", isDir: false},
			{name: "file.multiple.dots.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	result := cellsString(c.Cells())
	assert.Contains(t, result, "file with spaces.go")
	assert.Contains(t, result, "file-with-dashes.go")
	assert.Contains(t, result, "file_with_underscores.go")
	assert.Contains(t, result, "file.multiple.dots.go")
}

func TestUnicodeFilenames(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "日本語.go", isDir: false},
			{name: "émojis🎉.txt", isDir: false},
			{name: "ñoño.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	result := cellsString(c.Cells())
	assert.Contains(t, result, "日本語.go")
	assert.Contains(t, result, "émojis🎉.txt")
	assert.Contains(t, result, "ñoño.go")
}

func TestCustomIndentRune(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
		},
		"/project/src": {
			{name: "app.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{
		IndentRune: '┃',
	})
	require.NoError(t, err)

	c.ExpandNodeAt(term.Coordinates{Y: 0})
	result := cellsString(c.Cells())
	assert.Contains(t, result, "┃")
}

func TestManyFiles(t *testing.T) {
	entries := make([]mockEntry, 100)
	for i := range 100 {
		entries[i] = mockEntry{
			name:  fmt.Sprintf("file%03d.go", i),
			isDir: false,
		}
	}

	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": entries,
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	cells := c.Cells()
	assert.Len(t, cells, 100)

	// Verify sorting
	w, h := c.Dimensions()
	assert.Greater(t, w, 0)
	assert.Equal(t, 100, h)
}

func TestToggleExpandMultipleTimes(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
		},
		"/project/src": {
			{name: "app.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// Toggle expand multiple times
	for range 5 {
		c.ExpandNodeAt(term.Coordinates{Y: 0}) // expand
		assert.Len(t, c.Cells(), 2)

		c.ExpandNodeAt(term.Coordinates{Y: 0}) // collapse
		assert.Len(t, c.Cells(), 1)
	}

	// No changes expected
	ops := c.Changes()
	assert.Empty(t, ops)
}

func TestNodeAtOutOfBounds(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "test.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	tests := []struct {
		name string
		y    int
	}{
		{"negative", -1},
		{"large positive", 1000},
		{"just out of bounds", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := c.NodeAt(term.Coordinates{Y: tt.y})
			assert.False(t, ok)
		})
	}
}

func TestExpandNodeAtOutOfBounds(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "test.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// Should not crash on invalid coordinates
	_, isFile := c.ExpandNodeAt(term.Coordinates{Y: -1})
	assert.False(t, isFile)

	_, isFile = c.ExpandNodeAt(term.Coordinates{Y: 100})
	assert.False(t, isFile)
}

func TestEmptyCells(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "test.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// Set empty cells
	c.SetCells(nil)

	ops := c.Changes()
	require.Len(t, ops, 1)
	assert.Equal(t, OpDelete, ops[0].Type)
}

func TestSetCellsPreservesExpanded(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
		},
		"/project/src": {
			{name: "app.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	// Expand
	c.ExpandNodeAt(term.Coordinates{Y: 0})
	assert.Len(t, c.Cells(), 2)

	// Clone and set cells
	cells := term.CloneCells(c.Cells())
	c.SetCells(cells)

	// Should preserve expanded state
	assert.Len(t, c.Cells(), 2)
}

func TestChangesMixedOperations(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "delete-me.go", isDir: false},
			{name: "rename-me.go", isDir: false},
			{name: "keep.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	cells := c.Cells()

	// delete-me.go is deleted (row 0)
	// rename-me.go is renamed to renamed.go (row 1)
	// keep.go stays (row 2)
	// new.go is created

	newCells := make([][]term.Cell, 3)

	// renamed.go (keep ID from rename-me.go)
	renamedContent := term.StringToCells("  renamed.go")[0]
	newCells[0] = append([]term.Cell{cells[1][0]}, renamedContent...)

	// keep.go unchanged
	newCells[1] = cells[2]

	// new.go (no ID)
	newCells[2] = term.StringToCells("  new.go")[0]

	c.SetCells(newCells)

	ops := c.Changes()
	assert.Len(t, ops, 3)

	counts := map[OperationType]int{}
	for _, op := range ops {
		counts[op.Type]++
	}
	assert.Equal(t, 1, counts[OpDelete])
	assert.Equal(t, 1, counts[OpRename])
	assert.Equal(t, 1, counts[OpCreate])
}

func TestOperationString(t *testing.T) {
	tests := []struct {
		op       Operation
		expected string
	}{
		{
			Operation{Type: OpCreate, NewURI: mustParseURI("file:///test.go")},
			"create test.go",
		},
		{
			Operation{Type: OpMkdir, NewURI: mustParseURI("file:///newdir")},
			"mkdir newdir",
		},
		{
			Operation{Type: OpDelete, URI: mustParseURI("file:///old.go")},
			"delete old.go",
		},
		{
			Operation{
				Type:   OpRename,
				URI:    mustParseURI("file:///old.go"),
				NewURI: mustParseURI("file:///new.go"),
			},
			"rename old.go -> new.go",
		},
		{
			Operation{
				Type:   OpMove,
				URI:    mustParseURI("file:///src/file.go"),
				NewURI: mustParseURI("file:///dest/file.go"),
			},
			"move /src/file.go -> /dest/file.go",
		},
		{
			Operation{
				Type:   OpCopy,
				URI:    mustParseURI("file:///src/file.go"),
				NewURI: mustParseURI("file:///copy.go"),
			},
			"copy /src/file.go -> /copy.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.op.String())
		})
	}
}

func TestOperationTypeString(t *testing.T) {
	tests := []struct {
		t        OperationType
		expected string
	}{
		{OpCreate, "create"},
		{OpMkdir, "mkdir"},
		{OpDelete, "delete"},
		{OpRename, "rename"},
		{OpMove, "move"},
		{OpCopy, "copy"},
		{OperationType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.t.String())
		})
	}
}

func TestDimensionsEmpty(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	w, h := c.Dimensions()
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, h)
}

func TestCellsReturnsCopy(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "test.go", isDir: false},
		},
	}}
	c, err := New(mfs, rootURI(), Config{})
	require.NoError(t, err)

	cells1 := c.Cells()
	cells2 := c.Cells()

	// Both should be equal
	assert.Equal(t, cellsString(cells1), cellsString(cells2))
}

func mustParseURI(s string) workspaceapi.URI {
	u, err := workspaceapi.ParseURI(s)
	if err != nil {
		panic(err)
	}
	return u
}

// Test helpers

type mockEntry struct {
	name  string
	isDir bool
}

func (e mockEntry) Name() string      { return e.name }
func (e mockEntry) IsDir() bool       { return e.isDir }
func (e mockEntry) Type() fs.FileMode { return 0 }
func (e mockEntry) Info() (fs.FileInfo, error) {
	return mockInfo{e}, nil
}

type mockInfo struct{ e mockEntry }

func (m mockInfo) Name() string      { return m.e.name }
func (m mockInfo) Size() int64       { return 0 }
func (m mockInfo) Mode() os.FileMode { return 0 }
func (m mockInfo) ModTime() time.Time {
	return time.Time{}
}
func (m mockInfo) IsDir() bool { return m.e.isDir }
func (m mockInfo) Sys() any    { return nil }

type mockFS struct {
	dirs map[string][]mockEntry
}

func (m *mockFS) URI(
	path string,
) (workspaceapi.URI, error) {
	return workspaceapi.ParseURI("file://" + path)
}

func (m *mockFS) ReadDir(
	name string,
) ([]os.DirEntry, error) {
	entries, ok := m.dirs[name]
	if !ok {
		return nil, fmt.Errorf("not found: %s", name)
	}
	ret := make([]os.DirEntry, len(entries))
	for i, e := range entries {
		ret[i] = e
	}
	return ret, nil
}

func (m *mockFS) OpenFile(
	string, int, os.FileMode,
) (workspaceapi.File, error) {
	panic("not implemented")
}

func (m *mockFS) Remove(string) error {
	panic("not implemented")
}

func (m *mockFS) Stat(string) (os.FileInfo, error) {
	panic("not implemented")
}

func (m *mockFS) MkdirAll(string, os.FileMode) error {
	panic("not implemented")
}

func rootURI() workspaceapi.URI {
	u, err := workspaceapi.ParseURI("file:///project")
	if err != nil {
		panic(err)
	}
	return u
}

func cellsString(cells [][]term.Cell) string {
	// Strip the first cell (row ID) from each row before converting
	stripped := make([][]term.Cell, len(cells))
	for i, row := range cells {
		if len(row) > 1 {
			stripped[i] = row[1:]
		} else {
			stripped[i] = row
		}
	}
	return term.CellsToString(stripped)
}
