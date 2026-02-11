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
	// OpRename indicates a file or directory should be renamed.
	OpRename
)

// Operation represents a single filesystem change detected by
// diffing the edited cell buffer against the internal tree.
type Operation struct {
	Type   OperationType
	URI    workspaceapi.URI
	NewURI workspaceapi.URI
}

func computeChanges(
	oldRoot *node, entries []parsedEntry, cfg Config,
) []Operation {
	newTree := buildVirtualTree(entries)
	var ops []Operation
	diffChildren(oldRoot, newTree, cfg, &ops)
	return ops
}

func buildVirtualTree(entries []parsedEntry) *node {
	root := &node{depth: 0}
	stack := []*node{root}
	for _, e := range entries {
		for len(stack) > 1 && stack[len(stack)-1].depth >= e.depth {
			stack = stack[:len(stack)-1]
		}
		parent := stack[len(stack)-1]
		child := &node{
			name:  e.name,
			isDir: e.isDir,
			depth: e.depth,
		}
		parent.children = append(parent.children, child)
		if e.isDir {
			stack = append(stack, child)
		}
	}
	return root
}

func diffChildren(
	old, new *node, cfg Config, ops *[]Operation,
) {
	oldByName := make(map[string]*node, len(old.children))
	for _, c := range old.children {
		oldByName[c.name] = c
	}
	newByName := make(map[string]*node, len(new.children))
	for _, c := range new.children {
		newByName[c.name] = c
	}

	var removed []*node
	var added []*node

	for _, c := range old.children {
		if _, ok := newByName[c.name]; !ok {
			removed = append(removed, c)
		}
	}
	for _, c := range new.children {
		if _, ok := oldByName[c.name]; !ok {
			added = append(added, c)
		}
	}

	if len(removed) == 1 && len(added) == 1 {
		*ops = append(*ops, Operation{
			Type:   OpRename,
			URI:    removed[0].uri,
			NewURI: workspaceapi.Join(
				old.uri, added[0].name,
			),
		})
	} else {
		for _, c := range removed {
			*ops = append(*ops, Operation{
				Type: OpDelete,
				URI:  c.uri,
			})
		}
		for _, c := range added {
			if c.isDir {
				*ops = append(*ops, Operation{
					Type: OpMkdir,
					NewURI: workspaceapi.Join(
						old.uri, c.name,
					),
				})
			} else {
				*ops = append(*ops, Operation{
					Type: OpCreate,
					NewURI: workspaceapi.Join(
						old.uri, c.name,
					),
				})
			}
		}
	}

	for _, oc := range old.children {
		if !oc.expanded {
			continue
		}
		nc, ok := newByName[oc.name]
		if !ok {
			continue
		}
		diffChildren(oc, nc, cfg, ops)
	}
}
