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
	"path/filepath"
	"slices"
	"strings"

	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/term"
)

type node struct {
	id       rune // Unique identifier from Private Use Area
	name     string
	uri      workspaceapi.URI
	isDir    bool
	expanded bool
	depth    int
	parent   *node
	children []*node
}

// deepCopy creates a deep copy of the node and all its children.
func (n *node) deepCopy() *node {
	if n == nil {
		return nil
	}
	copy := &node{
		id:       n.id,
		name:     n.name,
		uri:      n.uri,
		isDir:    n.isDir,
		expanded: n.expanded,
		depth:    n.depth,
		// parent is set below when copying children
	}
	if len(n.children) > 0 {
		copy.children = make([]*node, len(n.children))
		for i, child := range n.children {
			copy.children[i] = child.deepCopy()
			copy.children[i].parent = copy
		}
	}
	return copy
}

type parsedEntry struct {
	id    rune // Row ID (0 if not present)
	name  string
	isDir bool
	depth int
}

func readChildren(
	fs workspaceapi.FileSystem, n *node,
) error {
	entries, err := fs.ReadDir(n.uri.Path())
	if err != nil {
		return err
	}
	children := make([]*node, 0, len(entries))
	for _, e := range entries {
		child := &node{
			name:   e.Name(),
			uri:    workspaceapi.Join(n.uri, e.Name()),
			isDir:  e.IsDir(),
			depth:  n.depth + 1,
			parent: n,
		}
		children = append(children, child)
	}
	sortChildren(children)
	n.children = children
	return nil
}

func sortChildren(children []*node) {
	slices.SortFunc(children, func(a, b *node) int {
		if a.isDir != b.isDir {
			if a.isDir {
				return -1
			}
			return 1
		}
		return strings.Compare(a.name, b.name)
	})
}

func flatten(root *node) []*node {
	var flat []*node
	var walk func(n *node)
	walk = func(n *node) {
		for _, child := range n.children {
			flat = append(flat, child)
			if child.isDir && child.expanded {
				walk(child)
			}
		}
	}
	walk(root)
	return flat
}

// renderCellsWithIDs renders nodes to cells, prepending each row
// with the node's ID as the first cell.
func renderCellsWithIDs(
	flat []*node, cfg Config,
) [][]term.Cell {
	rows := make([][]term.Cell, 0, len(flat))
	for _, n := range flat {
		line := renderLine(n, cfg)
		lineCells := term.StringToCells(line)
		if len(lineCells) == 0 {
			continue
		}
		// Prepend the row ID as first cell
		row := make([]term.Cell, 0, len(lineCells[0])+1)
		row = append(row, term.Cell{Ch: n.id, Width: 0})
		row = append(row, lineCells[0]...)
		rows = append(rows, row)
	}
	return rows
}

func renderLine(n *node, cfg Config) string {
	var b strings.Builder
	indent := indentRune(cfg)
	for range n.depth {
		b.WriteRune(indent)
		b.WriteRune(' ')
	}
	b.WriteRune(iconFor(n, cfg))
	b.WriteRune(' ')
	b.WriteString(n.name)
	if n.isDir {
		b.WriteRune('/')
	}
	return b.String()
}

func indentRune(cfg Config) rune {
	if cfg.IndentRune != 0 {
		return cfg.IndentRune
	}
	return '│'
}

func iconFor(n *node, cfg Config) rune {
	if cfg.Icons == nil {
		return ' '
	}
	if n.isDir {
		if icon, ok := cfg.Icons["/"]; ok {
			return icon
		}
	} else {
		ext := filepath.Ext(n.name)
		if icon, ok := cfg.Icons[ext]; ok {
			return icon
		}
	}
	if icon, ok := cfg.Icons[""]; ok {
		return icon
	}
	return ' '
}

// parseCellsWithIDs parses cells back into entries, extracting
// the row ID from the first cell of each row.
func parseCellsWithIDs(
	cells [][]term.Cell, cfg Config,
) []parsedEntry {
	entries := make([]parsedEntry, 0, len(cells))
	indent := indentRune(cfg)
	for _, row := range cells {
		if len(row) == 0 {
			continue
		}

		// First cell is the row ID (if in Private Use Area)
		var id rune
		startIdx := 0
		if row[0].Ch >= firstRowID && row[0].Ch <= lastRowID {
			id = row[0].Ch
			startIdx = 1
		}

		// Parse the rest of the row
		line := cellsToRowString(row[startIdx:])
		entry := parseLine(line, indent)
		entry.id = id
		entries = append(entries, entry)
	}
	return entries
}

func cellsToRowString(row []term.Cell) string {
	var b strings.Builder
	for _, c := range row {
		b.WriteRune(c.Ch)
		for _, comb := range c.Combining {
			b.WriteRune(comb)
		}
	}
	return b.String()
}

func parseLine(line string, indent rune) parsedEntry {
	var depth int
	runes := []rune(line)
	i := 0
	for i+1 < len(runes) &&
		runes[i] == indent && runes[i+1] == ' ' {
		depth++
		i += 2
	}
	// skip icon + space
	if i+1 < len(runes) {
		i += 2
	}
	name := string(runes[i:])
	isDir := strings.HasSuffix(name, "/")
	if isDir {
		name = strings.TrimSuffix(name, "/")
	}
	return parsedEntry{
		name:  name,
		isDir: isDir,
		depth: depth,
	}
}

func collectExpandedPaths(root *node) map[string]bool {
	expanded := make(map[string]bool)
	var walk func(n *node)
	walk = func(n *node) {
		for _, child := range n.children {
			if child.isDir && child.expanded {
				expanded[child.uri.String()] = true
				walk(child)
			}
		}
	}
	walk(root)
	return expanded
}

func restoreExpanded(
	fs workspaceapi.FileSystem,
	root *node,
	expanded map[string]bool,
) error {
	for _, child := range root.children {
		if !child.isDir {
			continue
		}
		if !expanded[child.uri.String()] {
			continue
		}
		if err := readChildren(fs, child); err != nil {
			return err
		}
		child.expanded = true
		if err := restoreExpanded(fs, child, expanded); err != nil {
			return err
		}
	}
	return nil
}
