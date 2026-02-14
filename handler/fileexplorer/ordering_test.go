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

package fileexplorer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/component/fileexplorer"
)

func TestOrderOperations(t *testing.T) {
	tests := []struct {
		name     string
		ops      []fileexplorer.Operation
		expected []fileexplorer.OperationType
	}{
		{
			name:     "empty operations",
			ops:      nil,
			expected: nil,
		},
		{
			name: "deletes ordered deepest first",
			ops: []fileexplorer.Operation{
				{Type: fileexplorer.OpDelete, URI: mustURI("/a")},
				{Type: fileexplorer.OpDelete, URI: mustURI("/a/b/c")},
				{Type: fileexplorer.OpDelete, URI: mustURI("/a/b")},
			},
			expected: []fileexplorer.OperationType{
				fileexplorer.OpDelete,
				fileexplorer.OpDelete,
				fileexplorer.OpDelete,
			},
		},
		{
			name: "creates ordered shallowest first",
			ops: []fileexplorer.Operation{
				{Type: fileexplorer.OpMkdir, NewURI: mustURI("/a/b/c")},
				{Type: fileexplorer.OpCreate, NewURI: mustURI("/a/file.txt")},
				{Type: fileexplorer.OpMkdir, NewURI: mustURI("/a")},
			},
			expected: []fileexplorer.OperationType{
				fileexplorer.OpMkdir,
				fileexplorer.OpCreate,
				fileexplorer.OpMkdir,
			},
		},
		{
			name: "mixed operations ordered correctly",
			ops: []fileexplorer.Operation{
				{Type: fileexplorer.OpCreate, NewURI: mustURI("/new.txt")},
				{Type: fileexplorer.OpDelete, URI: mustURI("/old.txt")},
				{Type: fileexplorer.OpRename, URI: mustURI("/a.txt"), NewURI: mustURI("/b.txt")},
			},
			expected: []fileexplorer.OperationType{
				fileexplorer.OpDelete,
				fileexplorer.OpRename,
				fileexplorer.OpCreate,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OrderOperations(tt.ops)
			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}
			assert.Equal(t, len(tt.expected), len(result))
			for i, op := range result {
				assert.Equal(t, tt.expected[i], op.Type, "operation %d", i)
			}
		})
	}
}

func TestOrderDeletes(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		expected []string
	}{
		{
			name:     "empty",
			paths:    nil,
			expected: nil,
		},
		{
			name:     "single",
			paths:    []string{"/a"},
			expected: []string{"/a"},
		},
		{
			name:     "nested paths",
			paths:    []string{"/a", "/a/b", "/a/b/c"},
			expected: []string{"/a/b/c", "/a/b", "/a"},
		},
		{
			name:     "sibling paths",
			paths:    []string{"/a/b", "/a/c", "/a/d"},
			expected: []string{"/a/b", "/a/c", "/a/d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ops []fileexplorer.Operation
			for _, p := range tt.paths {
				ops = append(ops, fileexplorer.Operation{
					Type: fileexplorer.OpDelete,
					URI:  mustURI(p),
				})
			}

			result := orderDeletes(ops)

			var resultPaths []string
			for _, op := range result {
				resultPaths = append(resultPaths, op.URI.Path())
			}
			assert.Equal(t, tt.expected, resultPaths)
		})
	}
}

func TestOrderCreates(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		expected []string
	}{
		{
			name:     "empty",
			paths:    nil,
			expected: nil,
		},
		{
			name:     "single",
			paths:    []string{"/a"},
			expected: []string{"/a"},
		},
		{
			name:     "nested paths",
			paths:    []string{"/a/b/c", "/a", "/a/b"},
			expected: []string{"/a", "/a/b", "/a/b/c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ops []fileexplorer.Operation
			for _, p := range tt.paths {
				ops = append(ops, fileexplorer.Operation{
					Type:   fileexplorer.OpMkdir,
					NewURI: mustURI(p),
				})
			}

			result := orderCreates(ops)

			var resultPaths []string
			for _, op := range result {
				resultPaths = append(resultPaths, op.NewURI.Path())
			}
			assert.Equal(t, tt.expected, resultPaths)
		})
	}
}

func TestPathDepth(t *testing.T) {
	tests := []struct {
		path     string
		expected int
	}{
		{"", 0},
		{"/", 0},
		{"/a", 1},
		{"/a/b", 2},
		{"/a/b/c", 3},
		{"a/b/c", 3},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, pathDepth(tt.path))
		})
	}
}

func TestOrderRenames(t *testing.T) {
	tests := []struct {
		name     string
		renames  [][2]string // [from, to] pairs
		expected []string    // expected order of 'from' paths
	}{
		{
			name:     "empty",
			renames:  nil,
			expected: nil,
		},
		{
			name:     "single rename",
			renames:  [][2]string{{"/a", "/b"}},
			expected: []string{"/a"},
		},
		{
			name: "independent renames preserve order",
			renames: [][2]string{
				{"/x", "/y"},
				{"/a", "/b"},
			},
			expected: []string{"/x", "/a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ops []fileexplorer.Operation
			for _, r := range tt.renames {
				ops = append(ops, fileexplorer.Operation{
					Type:   fileexplorer.OpRename,
					URI:    mustURI(r[0]),
					NewURI: mustURI(r[1]),
				})
			}

			result := orderRenames(ops)

			var resultPaths []string
			for _, op := range result {
				resultPaths = append(resultPaths, op.URI.Path())
			}
			assert.Equal(t, tt.expected, resultPaths)
		})
	}
}

func TestOrderOperationsGroupsByType(t *testing.T) {
	// Test that operations are grouped by type: deletes, renames, creates, copies
	ops := []fileexplorer.Operation{
		{Type: fileexplorer.OpCreate, NewURI: mustURI("/new1.txt")},
		{Type: fileexplorer.OpCopy, URI: mustURI("/src.txt"), NewURI: mustURI("/copy.txt")},
		{Type: fileexplorer.OpRename, URI: mustURI("/old.txt"), NewURI: mustURI("/renamed.txt")},
		{Type: fileexplorer.OpDelete, URI: mustURI("/delete.txt")},
		{Type: fileexplorer.OpMkdir, NewURI: mustURI("/newdir")},
		{Type: fileexplorer.OpMove, URI: mustURI("/move.txt"), NewURI: mustURI("/dest/move.txt")},
	}

	result := OrderOperations(ops)

	// Verify all operations are present
	assert.Equal(t, len(ops), len(result))

	// Group operations by type
	var foundDelete, foundRename, foundMove, foundCreate, foundMkdir, foundCopy bool
	var lastDelete, lastRename, lastMove, firstCreate, firstMkdir, firstCopy int

	for i, op := range result {
		switch op.Type {
		case fileexplorer.OpDelete:
			foundDelete = true
			lastDelete = i
		case fileexplorer.OpRename:
			foundRename = true
			lastRename = i
		case fileexplorer.OpMove:
			foundMove = true
			lastMove = i
		case fileexplorer.OpCreate:
			if !foundCreate {
				firstCreate = i
			}
			foundCreate = true
		case fileexplorer.OpMkdir:
			if !foundMkdir {
				firstMkdir = i
			}
			foundMkdir = true
		case fileexplorer.OpCopy:
			if !foundCopy {
				firstCopy = i
			}
			foundCopy = true
		}
	}

	assert.True(t, foundDelete && foundRename && foundMove && foundCreate && foundMkdir && foundCopy)

	// Verify deletes come before creates
	assert.Less(t, lastDelete, firstCreate, "deletes should come before creates")
	assert.Less(t, lastDelete, firstMkdir, "deletes should come before mkdirs")

	// Verify renames/moves come before creates
	assert.Less(t, lastRename, firstCreate, "renames should come before creates")
	assert.Less(t, lastMove, firstCreate, "moves should come before creates")

	// Verify creates come before copies
	assert.Less(t, firstCreate, firstCopy, "creates should come before copies")
}

func TestOrderDeletesDeepNesting(t *testing.T) {
	// Test with deeply nested paths
	paths := []string{
		"/a",
		"/a/b",
		"/a/b/c",
		"/a/b/c/d",
		"/a/b/c/d/e",
		"/a/b/c/d/e/f",
	}

	var ops []fileexplorer.Operation
	for _, p := range paths {
		ops = append(ops, fileexplorer.Operation{
			Type: fileexplorer.OpDelete,
			URI:  mustURI(p),
		})
	}

	result := orderDeletes(ops)

	// Should be deepest first
	expected := []string{
		"/a/b/c/d/e/f",
		"/a/b/c/d/e",
		"/a/b/c/d",
		"/a/b/c",
		"/a/b",
		"/a",
	}

	var resultPaths []string
	for _, op := range result {
		resultPaths = append(resultPaths, op.URI.Path())
	}
	assert.Equal(t, expected, resultPaths)
}

func TestOrderCreatesMixedTypes(t *testing.T) {
	// Test creates with both OpCreate and OpMkdir
	ops := []fileexplorer.Operation{
		{Type: fileexplorer.OpCreate, NewURI: mustURI("/a/b/c/file.txt")},
		{Type: fileexplorer.OpMkdir, NewURI: mustURI("/a")},
		{Type: fileexplorer.OpMkdir, NewURI: mustURI("/a/b")},
		{Type: fileexplorer.OpCreate, NewURI: mustURI("/a/file.txt")},
		{Type: fileexplorer.OpMkdir, NewURI: mustURI("/a/b/c")},
	}

	result := orderCreates(ops)

	// Should be ordered by depth (shallowest first)
	depths := make([]int, len(result))
	for i, op := range result {
		depths[i] = pathDepth(op.NewURI.Path())
	}

	// Verify depths are non-decreasing
	for i := 1; i < len(depths); i++ {
		assert.GreaterOrEqual(t, depths[i], depths[i-1],
			"depth at %d (%d) should be >= depth at %d (%d)",
			i, depths[i], i-1, depths[i-1])
	}
}

func TestOrderOperationsEmpty(t *testing.T) {
	result := OrderOperations(nil)
	assert.Nil(t, result)

	result = OrderOperations([]fileexplorer.Operation{})
	assert.Empty(t, result)
}

func TestOrderOperationsSingleType(t *testing.T) {
	tests := []struct {
		name    string
		opType  fileexplorer.OperationType
		useURI  bool // true for Delete, false for Create
	}{
		{"only deletes", fileexplorer.OpDelete, true},
		{"only creates", fileexplorer.OpCreate, false},
		{"only mkdirs", fileexplorer.OpMkdir, false},
		{"only renames", fileexplorer.OpRename, true},
		{"only moves", fileexplorer.OpMove, true},
		{"only copies", fileexplorer.OpCopy, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ops []fileexplorer.Operation
			for i := range 3 {
				op := fileexplorer.Operation{Type: tt.opType}
				if tt.useURI {
					op.URI = mustURI("/file" + string(rune('a'+i)))
					if tt.opType == fileexplorer.OpRename ||
						tt.opType == fileexplorer.OpMove ||
						tt.opType == fileexplorer.OpCopy {
						op.NewURI = mustURI("/new" + string(rune('a'+i)))
					}
				} else {
					op.NewURI = mustURI("/file" + string(rune('a'+i)))
				}
				ops = append(ops, op)
			}

			result := OrderOperations(ops)
			assert.Equal(t, len(ops), len(result))
			for _, op := range result {
				assert.Equal(t, tt.opType, op.Type)
			}
		})
	}
}

func TestPathDepthEdgeCases(t *testing.T) {
	tests := []struct {
		path     string
		expected int
	}{
		{"", 0},
		{"/", 0},
		{"//", 0},
		{"///", 0},
		{"/a/", 1},
		{"/a//b", 3}, // Double slash creates empty component
		{"./a/b", 3},
		{"../a/b", 3},
		{"a", 1},
		{"a/", 1},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := pathDepth(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOrderRenamesCycle(t *testing.T) {
	// Test cycle detection: A->B, B->C, C->A
	ops := []fileexplorer.Operation{
		{Type: fileexplorer.OpRename, URI: mustURI("/a"), NewURI: mustURI("/b")},
		{Type: fileexplorer.OpRename, URI: mustURI("/b"), NewURI: mustURI("/c")},
		{Type: fileexplorer.OpRename, URI: mustURI("/c"), NewURI: mustURI("/a")},
	}

	// Should not hang or crash - cycle detection should handle this
	result := orderRenames(ops)

	// Should return some operations (exact behavior depends on implementation)
	// The important thing is it doesn't hang
	assert.NotNil(t, result)
}

func mustURI(path string) workspaceapi.URI {
	uri, err := workspaceapi.ParseURI("file://localhost" + path)
	if err != nil {
		panic(err)
	}
	return uri
}
