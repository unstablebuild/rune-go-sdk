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
	"slices"
	"strings"

	"github.com/unstablebuild/rune-go-sdk/component/fileexplorer"
)

// OrderOperations sorts operations to ensure correct execution order:
// 1. Deletes - deepest paths first (children before parents)
// 2. Renames/Moves - ordered to avoid conflicts
// 3. Creates/Mkdirs - shallowest paths first (parents before children)
// 4. Copies - after creates (source must exist)
func OrderOperations(ops []fileexplorer.Operation) []fileexplorer.Operation {
	if len(ops) == 0 {
		return ops
	}

	var deletes, renames, creates, copies []fileexplorer.Operation
	for _, op := range ops {
		switch op.Type {
		case fileexplorer.OpDelete:
			deletes = append(deletes, op)
		case fileexplorer.OpRename, fileexplorer.OpMove:
			renames = append(renames, op)
		case fileexplorer.OpCreate, fileexplorer.OpMkdir:
			creates = append(creates, op)
		case fileexplorer.OpCopy:
			copies = append(copies, op)
		}
	}

	deletes = orderDeletes(deletes)
	renames = orderRenames(renames)
	creates = orderCreates(creates)
	copies = orderCreates(copies) // Copies also need parent dirs first

	result := make([]fileexplorer.Operation, 0, len(ops))
	result = append(result, deletes...)
	result = append(result, renames...)
	result = append(result, creates...)
	result = append(result, copies...)
	return result
}

// orderDeletes sorts delete operations so that deeper paths come first.
// This ensures children are deleted before their parents.
func orderDeletes(ops []fileexplorer.Operation) []fileexplorer.Operation {
	slices.SortFunc(ops, func(a, b fileexplorer.Operation) int {
		depthA := pathDepth(a.URI.Path())
		depthB := pathDepth(b.URI.Path())
		// Deeper paths first (descending order)
		return depthB - depthA
	})
	return ops
}

// orderRenames sorts rename operations to avoid conflicts.
// If A->B and B->C exist, B->C must happen first.
func orderRenames(ops []fileexplorer.Operation) []fileexplorer.Operation {
	if len(ops) <= 1 {
		return ops
	}

	// Build a graph of dependencies
	// An operation depends on another if its source path would be
	// overwritten by the other operation's destination
	graph := make(map[int][]int) // op index -> indices it depends on
	for i, op := range ops {
		for j, other := range ops {
			if i == j {
				continue
			}
			// If op's source is other's destination, op must come after other
			if op.URI.Path() == other.NewURI.Path() {
				graph[i] = append(graph[i], j)
			}
		}
	}

	// Topological sort
	result := make([]fileexplorer.Operation, 0, len(ops))
	visited := make(map[int]bool)
	inStack := make(map[int]bool)

	var visit func(int) bool
	visit = func(i int) bool {
		if inStack[i] {
			// Cycle detected, skip
			return false
		}
		if visited[i] {
			return true
		}
		inStack[i] = true
		for _, dep := range graph[i] {
			if !visit(dep) {
				return false
			}
		}
		inStack[i] = false
		visited[i] = true
		result = append(result, ops[i])
		return true
	}

	for i := range ops {
		visit(i)
	}

	return result
}

// orderCreates sorts create operations so that shallower paths come first.
// This ensures parent directories are created before their children.
func orderCreates(ops []fileexplorer.Operation) []fileexplorer.Operation {
	slices.SortFunc(ops, func(a, b fileexplorer.Operation) int {
		depthA := pathDepth(a.NewURI.Path())
		depthB := pathDepth(b.NewURI.Path())
		// Shallower paths first (ascending order)
		return depthA - depthB
	})
	return ops
}

// pathDepth returns the number of path components.
func pathDepth(path string) int {
	path = strings.Trim(path, "/")
	if path == "" {
		return 0
	}
	return strings.Count(path, "/") + 1
}
