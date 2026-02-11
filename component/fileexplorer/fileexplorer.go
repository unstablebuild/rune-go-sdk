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
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
)

// Config defines options for the file explorer component.
type Config struct {
	// Icons maps file extensions to icon runes.
	// "/" key is used for directories, "" for the default
	// fallback icon.
	Icons map[string]rune
	// IndentRune is drawn for each depth level.
	// Defaults to '│' (U+2502) when zero.
	IndentRune rune
}

// Component renders a file tree. It implements
// component.Floating and supports expand/collapse,
// node lookup, external cell editing with change detection,
// and filesystem reload.
type Component struct {
	fs     workspaceapi.FileSystem
	root   workspaceapi.URI
	cfg    Config
	width  int
	height int
	cells  [][]term.Cell
	tree   *node
	flat   []*node
}

var _ component.Floating = (*Component)(nil)

// New creates a new Component with the given fs, root URI and config.
func New(
	fs workspaceapi.FileSystem,
	root workspaceapi.URI,
	cfg Config,
) (*Component, error) {
	c := &Component{
		fs:   fs,
		root: root,
		cfg:  cfg,
		tree: &node{
			uri:      root,
			isDir:    true,
			expanded: true,
			depth:    -1,
		},
	}
	if err := readChildren(fs, c.tree); err != nil {
		return nil, err
	}
	c.rebuild()
	return c, nil
}

// Resize stores clip dimensions for Draw.
func (c *Component) Resize(width, height int) {
	c.width, c.height = width, height
}

// Draw draws cells clipped to width/height.
func (c *Component) Draw(w term.Writer) {
	for y, row := range c.cells {
		if y >= c.height {
			break
		}
		var offset int
		for x, cell := range row {
			xi := x + offset
			if cell.Width > 1 {
				offset += int(cell.Width) - 1
			}
			if xi >= c.width {
				break
			}
			w.SetCell(
				term.Coordinates{X: xi, Y: y},
				cell,
			)
		}
	}
}

// Dimensions returns the optimal width and height of the
// cell buffer.
func (c *Component) Dimensions() (int, int) {
	return term.CalculateOptimalWidth(c.cells),
		len(c.cells)
}

// NodeAt returns the URI of the node at the given
// coordinates. The y coordinate maps to the flat node
// index. Returns false if out of range.
func (c *Component) NodeAt(
	pos term.Coordinates,
) (workspaceapi.URI, bool) {
	n := c.nodeAtRow(pos.Y)
	if n == nil {
		return workspaceapi.URI{}, false
	}
	return n.uri, true
}

// ExpandNodeAt expands or collapses the node at pos.
// For files, returns (uri, true). For directories,
// toggles expand state and returns (URI{}, false).
func (c *Component) ExpandNodeAt(
	pos term.Coordinates,
) (workspaceapi.URI, bool) {
	n := c.nodeAtRow(pos.Y)
	if n == nil {
		return workspaceapi.URI{}, false
	}
	if !n.isDir {
		return n.uri, true
	}
	if n.expanded {
		n.expanded = false
	} else {
		if n.children == nil {
			if err := readChildren(c.fs, n); err != nil {
				return workspaceapi.URI{}, false
			}
		}
		n.expanded = true
	}
	c.rebuild()
	return workspaceapi.URI{}, false
}

// ExpandLevel expands all directories at the same depth
// as the node at pos.
func (c *Component) ExpandLevel(pos term.Coordinates) {
	n := c.nodeAtRow(pos.Y)
	if n == nil {
		return
	}
	targetDepth := n.depth
	c.expandAtDepth(c.tree, targetDepth)
	c.rebuild()
}

// CollapseLevel collapses all directories at the same depth
// as the node at pos.
func (c *Component) CollapseLevel(pos term.Coordinates) {
	n := c.nodeAtRow(pos.Y)
	if n == nil {
		return
	}
	targetDepth := n.depth
	collapseAtDepth(c.tree, targetDepth)
	c.rebuild()
}

// CollapseAll collapses every directory in the tree.
func (c *Component) CollapseAll() {
	collapseAll(c.tree)
	c.rebuild()
}

// Cells returns the current cell buffer.
func (c *Component) Cells() [][]term.Cell {
	return c.cells
}

// SetCells replaces the cell buffer with the given cells.
func (c *Component) SetCells(cells [][]term.Cell) {
	c.cells = cells
}

// Changes diffs the current cell buffer against the
// internal tree and returns the detected operations.
func (c *Component) Changes() []Operation {
	entries := parseCells(c.cells, c.cfg)
	return computeChanges(c.tree, entries, c.cfg)
}

// Init re-reads the filesystem and rebuilds the tree,
// preserving expanded directories.
func (c *Component) Init() error {
	expanded := collectExpandedPaths(c.tree)
	c.tree = &node{
		uri:      c.root,
		isDir:    true,
		expanded: true,
		depth:    -1,
	}
	if err := readChildren(c.fs, c.tree); err != nil {
		return err
	}
	if err := restoreExpanded(
		c.fs, c.tree, expanded,
	); err != nil {
		return err
	}
	c.rebuild()
	return nil
}

func (c *Component) rebuild() {
	c.flat = flatten(c.tree)
	c.cells = renderCells(c.flat, c.cfg)
}

func (c *Component) nodeAtRow(y int) *node {
	if y < 0 || y >= len(c.flat) {
		return nil
	}
	return c.flat[y]
}

func (c *Component) expandAtDepth(
	n *node, depth int,
) {
	for _, child := range n.children {
		if child.isDir && child.depth == depth {
			if child.children == nil {
				_ = readChildren(c.fs, child)
			}
			child.expanded = true
		}
		if child.isDir && child.expanded {
			c.expandAtDepth(child, depth)
		}
	}
}

func collapseAtDepth(n *node, depth int) {
	for _, child := range n.children {
		if child.isDir && child.depth == depth {
			child.expanded = false
		}
		if child.isDir && child.expanded {
			collapseAtDepth(child, depth)
		}
	}
}

func collapseAll(n *node) {
	for _, child := range n.children {
		if child.isDir {
			child.expanded = false
			collapseAll(child)
		}
	}
}
