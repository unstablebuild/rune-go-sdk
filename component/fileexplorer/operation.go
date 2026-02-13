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
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
)

// OperationType describes the kind of filesystem operation.
type OperationType int

const (
	// OpCreate indicates a new file should be created.
	OpCreate OperationType = iota
	// OpMkdir indicates a new directory should be created.
	OpMkdir
	// OpDelete indicates a file or directory should be deleted.
	OpDelete
	// OpRename indicates a file or directory should be renamed/moved.
	OpRename
)

// Operation represents a single filesystem change detected by
// comparing the view tree against the base tree.
type Operation struct {
	Type   OperationType
	URI    workspaceapi.URI // Source URI (for delete, rename)
	NewURI workspaceapi.URI // Destination URI (for create, mkdir, rename)
}

// computeChanges compares the view tree against the base tree
// and returns the operations needed to transform base into view.
// Nodes are matched by their row ID, allowing proper rename detection.
func computeChanges(baseTree, viewTree *node) []Operation {
	var ops []Operation

	// Collect all nodes from both trees
	baseNodes := collectNodesByID(baseTree)
	viewNodes := collectNodesByID(viewTree)
	newNodes := collectNewNodes(viewTree)

	// Find deleted nodes: in base but not in view
	for id, baseNode := range baseNodes {
		if _, exists := viewNodes[id]; !exists {
			ops = append(ops, Operation{
				Type: OpDelete,
				URI:  baseNode.uri,
			})
		}
	}

	// Add operations for new nodes (id == 0)
	for _, viewNode := range newNodes {
		if viewNode.isDir {
			ops = append(ops, Operation{
				Type:   OpMkdir,
				NewURI: viewNode.uri,
			})
		} else {
			ops = append(ops, Operation{
				Type:   OpCreate,
				NewURI: viewNode.uri,
			})
		}
	}

	// Find renamed/moved nodes: same ID but different URI
	for id, viewNode := range viewNodes {
		baseNode, exists := baseNodes[id]
		if !exists {
			// Node with ID not in base - treat as new
			if viewNode.isDir {
				ops = append(ops, Operation{
					Type:   OpMkdir,
					NewURI: viewNode.uri,
				})
			} else {
				ops = append(ops, Operation{
					Type:   OpCreate,
					NewURI: viewNode.uri,
				})
			}
			continue
		}

		// Check if renamed or moved
		if baseNode.uri.String() != viewNode.uri.String() {
			ops = append(ops, Operation{
				Type:   OpRename,
				URI:    baseNode.uri,
				NewURI: viewNode.uri,
			})
		}
	}

	return ops
}

// collectNodesByID walks the tree and returns a map of ID to node.
// Only includes nodes with non-zero IDs.
func collectNodesByID(root *node) map[rune]*node {
	nodes := make(map[rune]*node)
	var walk func(n *node)
	walk = func(n *node) {
		if n.id != 0 {
			nodes[n.id] = n
		}
		for _, child := range n.children {
			walk(child)
		}
	}
	walk(root)
	return nodes
}

// collectNewNodes walks the tree and returns all nodes with id == 0.
// These are nodes created by the user that don't exist in the base tree.
func collectNewNodes(root *node) []*node {
	var nodes []*node
	var walk func(n *node)
	walk = func(n *node) {
		if n.id == 0 && n != root {
			nodes = append(nodes, n)
		}
		for _, child := range n.children {
			walk(child)
		}
	}
	walk(root)
	return nodes
}
