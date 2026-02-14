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

// Entry represents a file or directory in the explorer.
// This is the primary data structure for oil.nvim-like editing.
// Entries are identified by their ID (from Unicode Private Use Area),
// which persists across edits allowing move/rename detection.
type Entry struct {
	ID       rune             // Unique identifier (0 for new entries)
	Name     string           // Filename without path
	URI      workspaceapi.URI // Full URI
	IsDir    bool             // True if directory
	Expanded bool             // True if directory is expanded
	Depth    int              // Nesting level (0 = root children)
}

// ParentURI returns the URI of this entry's parent directory.
func (e Entry) ParentURI() workspaceapi.URI {
	return workspaceapi.Dir(e.URI)
}

// EntryMap provides O(1) lookup of entries by ID.
type EntryMap map[rune]*Entry

// Add adds an entry to the map. If ID is 0, it's ignored (new entries).
func (m EntryMap) Add(e *Entry) {
	if e.ID != 0 {
		m[e.ID] = e
	}
}

// Get returns the entry with the given ID, or nil if not found.
func (m EntryMap) Get(id rune) *Entry {
	return m[id]
}

// collectEntries extracts all entries from a node tree into a flat list.
func collectEntries(root *node) []*Entry {
	var entries []*Entry
	var walk func(n *node)
	walk = func(n *node) {
		for _, child := range n.children {
			entries = append(entries, &Entry{
				ID:       child.id,
				Name:     child.name,
				URI:      child.uri,
				IsDir:    child.isDir,
				Expanded: child.expanded,
				Depth:    child.depth,
			})
			if child.isDir && child.expanded {
				walk(child)
			}
		}
	}
	walk(root)
	return entries
}

// buildEntryMap creates a map of ID -> Entry from a list of entries.
func buildEntryMap(entries []*Entry) EntryMap {
	m := make(EntryMap)
	for _, e := range entries {
		m.Add(e)
	}
	return m
}
