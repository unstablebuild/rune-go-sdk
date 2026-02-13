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
