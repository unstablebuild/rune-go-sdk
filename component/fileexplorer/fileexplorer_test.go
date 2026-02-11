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

func TestChanges(t *testing.T) {
	tests := []struct {
		name   string
		dirs   map[string][]mockEntry
		expand []int
		cells  string
		ops    []expectedOp
	}{
		{
			name: "no changes",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "main.go", isDir: false},
				},
			},
			cells: `  main.go`,
		},
		{
			name: "no changes in expanded tree",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "src", isDir: true},
				},
				"/project/src": {
					{name: "main.go", isDir: false},
				},
			},
			expand: []int{0},
			cells: `  src/
│   main.go`,
		},
		{
			name: "create file",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "main.go", isDir: false},
				},
			},
			cells: `  main.go
  new.go`,
			ops: []expectedOp{
				{OpCreate, "", "file:///project/new.go"},
			},
		},
		{
			name: "create directory",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "main.go", isDir: false},
				},
			},
			cells: `  main.go
  newdir/`,
			ops: []expectedOp{
				{OpMkdir, "", "file:///project/newdir"},
			},
		},
		{
			name: "delete file",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "a.go", isDir: false},
					{name: "b.go", isDir: false},
				},
			},
			cells: `  b.go`,
			ops: []expectedOp{
				{OpDelete, "file:///project/a.go", ""},
			},
		},
		{
			name: "delete directory",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "src", isDir: true},
					{name: "main.go", isDir: false},
				},
				"/project/src": {},
			},
			cells: `  main.go`,
			ops: []expectedOp{
				{OpDelete, "file:///project/src", ""},
			},
		},
		{
			name: "rename file",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "old.go", isDir: false},
				},
			},
			cells: `  new.go`,
			ops: []expectedOp{
				{
					OpRename,
					"file:///project/old.go",
					"file:///project/new.go",
				},
			},
		},
		{
			name: "rename directory",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "olddir", isDir: true},
				},
				"/project/olddir": {},
			},
			cells: `  newdir/`,
			ops: []expectedOp{
				{
					OpRename,
					"file:///project/olddir",
					"file:///project/newdir",
				},
			},
		},
		{
			name: "create file in expanded subdirectory",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "src", isDir: true},
				},
				"/project/src": {
					{name: "main.go", isDir: false},
				},
			},
			expand: []int{0},
			cells: `  src/
│   main.go
│   new.go`,
			ops: []expectedOp{
				{
					OpCreate, "",
					"file:///project/src/new.go",
				},
			},
		},
		{
			name: "delete file from expanded subdirectory",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "src", isDir: true},
				},
				"/project/src": {
					{name: "a.go", isDir: false},
					{name: "b.go", isDir: false},
				},
			},
			expand: []int{0},
			cells: `  src/
│   b.go`,
			ops: []expectedOp{
				{
					OpDelete,
					"file:///project/src/a.go", "",
				},
			},
		},
		{
			name: "rename file in expanded subdirectory",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "src", isDir: true},
				},
				"/project/src": {
					{name: "old.go", isDir: false},
				},
			},
			expand: []int{0},
			cells: `  src/
│   new.go`,
			ops: []expectedOp{
				{
					OpRename,
					"file:///project/src/old.go",
					"file:///project/src/new.go",
				},
			},
		},
		{
			name: "create directory in expanded subdirectory",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "src", isDir: true},
				},
				"/project/src": {
					{name: "main.go", isDir: false},
				},
			},
			expand: []int{0},
			cells: `  src/
│   sub/
│   main.go`,
			ops: []expectedOp{
				{
					OpMkdir, "",
					"file:///project/src/sub",
				},
			},
		},
		{
			name: "delete expanded directory with contents",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "src", isDir: true},
					{name: "readme.txt", isDir: false},
				},
				"/project/src": {
					{name: "main.go", isDir: false},
				},
			},
			expand: []int{0},
			cells:  `  readme.txt`,
			ops: []expectedOp{
				{OpDelete, "file:///project/src", ""},
			},
		},
		{
			name: "rename expanded directory",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "src", isDir: true},
				},
				"/project/src": {
					{name: "main.go", isDir: false},
				},
			},
			expand: []int{0},
			cells: `  lib/
│   main.go`,
			ops: []expectedOp{
				{
					OpRename,
					"file:///project/src",
					"file:///project/lib",
				},
			},
		},
		{
			name: "multiple deletes",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "a.go", isDir: false},
					{name: "b.go", isDir: false},
					{name: "c.go", isDir: false},
					{name: "d.go", isDir: false},
				},
			},
			cells: `  a.go
  d.go`,
			ops: []expectedOp{
				{OpDelete, "file:///project/b.go", ""},
				{OpDelete, "file:///project/c.go", ""},
			},
		},
		{
			name: "multiple creates",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "main.go", isDir: false},
				},
			},
			cells: `  main.go
  new.txt
  sub/`,
			ops: []expectedOp{
				{OpCreate, "", "file:///project/new.txt"},
				{OpMkdir, "", "file:///project/sub"},
			},
		},
		{
			name: "mixed operations at root",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "a.go", isDir: false},
					{name: "b.go", isDir: false},
					{name: "c.go", isDir: false},
				},
			},
			cells: `  b.go
  d.go
  sub/`,
			ops: []expectedOp{
				{OpDelete, "file:///project/a.go", ""},
				{OpDelete, "file:///project/c.go", ""},
				{OpCreate, "", "file:///project/d.go"},
				{OpMkdir, "", "file:///project/sub"},
			},
		},
		{
			name: "nested operations across expanded dirs",
			dirs: map[string][]mockEntry{
				"/project": {
					{name: "lib", isDir: true},
					{name: "src", isDir: true},
				},
				"/project/lib": {
					{name: "x.go", isDir: false},
				},
				"/project/src": {
					{name: "a.go", isDir: false},
					{name: "b.go", isDir: false},
				},
			},
			expand: []int{0, 2},
			cells: `  lib/
│   x.go
│   y.go
  src/
│   a.go
│   new.go`,
			ops: []expectedOp{
				{
					OpCreate, "",
					"file:///project/lib/y.go",
				},
				{
					OpRename,
					"file:///project/src/b.go",
					"file:///project/src/new.go",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New(
				&mockFS{dirs: tt.dirs},
				rootURI(), Config{},
			)
			require.NoError(t, err)
			for _, row := range tt.expand {
				c.ExpandNodeAt(
					term.Coordinates{Y: row},
				)
			}
			c.SetCells(term.StringToCells(tt.cells))
			ops := c.Changes()
			require.Len(t, ops, len(tt.ops),
				"unexpected number of operations")
			for i, exp := range tt.ops {
				assert.Equal(t, exp.opType,
					ops[i].Type, "op[%d].Type", i)
				assert.Equal(t, exp.uri,
					ops[i].URI.String(),
					"op[%d].URI", i)
				assert.Equal(t, exp.newURI,
					ops[i].NewURI.String(),
					"op[%d].NewURI", i)
			}
		})
	}
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

type expectedOp struct {
	opType OperationType
	uri    string
	newURI string
}

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
	return term.CellsToString(cells)
}
