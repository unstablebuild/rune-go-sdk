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
	"os"

	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
)

// Unicode Supplementary Private Use Area-A: U+F0000 to U+FFFFF
// We use these as invisible row IDs that survive string conversion.
const (
	firstRowID rune = 0xF0000
	lastRowID  rune = 0xFFFFF
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

// Component renders a file tree. It implements component.Floating
// and supports expand/collapse, external cell editing with change
// detection, and filesystem operations.
//
// Architecture:
//   - baseTree: represents the last known filesystem state
//   - viewTree: represents what's currently displayed/edited
//   - cells: rendering of viewTree, with invisible row IDs
//
// Row IDs use Unicode Private Use Area (U+F0000-U+FFFFF) as the
// first character of each row. This allows tracking node identity
// across moves/renames even if cells are converted to string and back.
type Component struct {
	fs       workspaceapi.FileSystem
	root     workspaceapi.URI
	cfg      Config
	width    int
	height   int
	cells    [][]term.Cell
	baseTree *node
	viewTree *node
	nextID   rune
}

var _ component.Floating = (*Component)(nil)

// New creates a new Component with the given fs, root URI and config.
func New(
	fs workspaceapi.FileSystem,
	root workspaceapi.URI,
	cfg Config,
) (*Component, error) {
	c := &Component{
		fs:     fs,
		root:   root,
		cfg:    cfg,
		nextID: firstRowID,
	}

	// Build base tree from filesystem
	c.baseTree = &node{
		uri:      root,
		isDir:    true,
		expanded: true,
		depth:    -1,
	}
	if err := readChildren(fs, c.baseTree); err != nil {
		return nil, err
	}
	c.assignIDs(c.baseTree)

	// Copy to view tree
	c.viewTree = c.baseTree.deepCopy()

	// Generate cells from view tree
	c.rebuildCells()

	return c, nil
}

// Resize stores clip dimensions for Draw.
func (c *Component) Resize(width, height int) {
	c.width, c.height = width, height
}

// Draw draws cells clipped to width/height, skipping the first
// cell of each row (the invisible row ID).
func (c *Component) Draw(w term.Writer) {
	for y, row := range c.cells {
		if y >= c.height {
			break
		}
		// Skip first cell (row ID)
		if len(row) < 2 {
			continue
		}
		var offset int
		for x, cell := range row[1:] {
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
// cell buffer (excluding the hidden row ID column).
func (c *Component) Dimensions() (int, int) {
	maxWidth := 0
	for _, row := range c.cells {
		// Subtract 1 for the hidden row ID
		w := len(row) - 1
		if w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth, len(c.cells)
}

// NodeAt returns the URI of the node at the given coordinates.
// Returns false if out of range or if the row doesn't correspond
// to a known node.
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
		// Sync collapsed state to baseTree
		if baseNode := c.findNodeByID(c.baseTree, n.id); baseNode != nil {
			baseNode.expanded = false
		}
	} else {
		if n.children == nil {
			if err := readChildren(c.fs, n); err != nil {
				return workspaceapi.URI{}, false
			}
			c.assignIDs(n)

			// Also add children to baseTree to keep it in sync
			// (expanding reveals existing FS nodes, not new ones)
			if baseNode := c.findNodeByID(c.baseTree, n.id); baseNode != nil {
				if baseNode.children == nil {
					_ = readChildren(c.fs, baseNode)
					c.copyIDsToBase(n, baseNode)
				}
				baseNode.expanded = true
			}
		}
		n.expanded = true
	}
	c.rebuildCells()
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
	c.expandAtDepth(c.viewTree, targetDepth)
	c.syncExpandedToBase()
	c.rebuildCells()
}

// CollapseLevel collapses all directories at the same depth
// as the node at pos.
func (c *Component) CollapseLevel(pos term.Coordinates) {
	n := c.nodeAtRow(pos.Y)
	if n == nil {
		return
	}
	targetDepth := n.depth
	collapseAtDepth(c.viewTree, targetDepth)
	collapseAtDepth(c.baseTree, targetDepth)
	c.rebuildCells()
}

// CollapseAll collapses every directory in the tree.
func (c *Component) CollapseAll() {
	collapseAll(c.viewTree)
	collapseAll(c.baseTree)
	c.rebuildCells()
}

// Cells returns the current cell buffer.
func (c *Component) Cells() [][]term.Cell {
	return c.cells
}

// SetCells replaces the cell buffer and rebuilds the view tree
// from the new cells. Row IDs are preserved where present,
// allowing tracking of moved/renamed nodes.
func (c *Component) SetCells(cells [][]term.Cell) {
	c.cells = cells
	c.viewTree = c.parseTreeFromCells(cells)
}

// Changes compares the view tree against the base tree and
// returns the detected operations. Nodes are matched by their
// row ID, allowing proper rename/move detection.
func (c *Component) Changes() []Operation {
	return computeChanges(c.baseTree, c.viewTree)
}

// ChangeSet compares the view tree against the base tree and
// returns a ChangeSet with operations and any detected conflicts.
func (c *Component) ChangeSet() *ChangeSet {
	return computeChangeSet(c.baseTree, c.viewTree)
}

// ApplyChanges executes the given operations on the filesystem,
// then syncs the base tree with the view tree.
func (c *Component) ApplyChanges(ops []Operation) error {
	for _, op := range ops {
		if err := c.applyOperation(op); err != nil {
			return err
		}
	}

	// Assign IDs to any new nodes (id == 0)
	c.assignIDs(c.viewTree)

	// Copy view tree to base tree
	c.baseTree = c.viewTree.deepCopy()

	// Regenerate cells to keep everything in sync
	c.rebuildCells()

	return nil
}

// Init re-reads the filesystem and rebuilds both trees,
// preserving expanded directories in the view tree.
func (c *Component) Init() error {
	expanded := collectExpandedPaths(c.viewTree)

	c.baseTree = &node{
		uri:      c.root,
		isDir:    true,
		expanded: true,
		depth:    -1,
	}
	if err := readChildren(c.fs, c.baseTree); err != nil {
		return err
	}
	c.assignIDs(c.baseTree)

	if err := restoreExpanded(c.fs, c.baseTree, expanded); err != nil {
		return err
	}

	c.viewTree = c.baseTree.deepCopy()
	c.rebuildCells()

	return nil
}

// rebuildCells generates cells from the view tree.
func (c *Component) rebuildCells() {
	flat := flatten(c.viewTree)
	c.cells = renderCellsWithIDs(flat, c.cfg)
}

// nodeAtRow finds the node corresponding to the given cell row
// by matching the row ID.
func (c *Component) nodeAtRow(y int) *node {
	if y < 0 || y >= len(c.cells) {
		return nil
	}
	row := c.cells[y]
	if len(row) == 0 {
		return nil
	}
	rowID := row[0].Ch
	return c.findNodeByID(c.viewTree, rowID)
}

// findNodeByID searches the tree for a node with the given ID.
func (c *Component) findNodeByID(root *node, id rune) *node {
	if root.id == id {
		return root
	}
	for _, child := range root.children {
		if found := c.findNodeByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

// assignIDs assigns unique IDs to all nodes that don't have one.
func (c *Component) assignIDs(root *node) {
	var walk func(n *node)
	walk = func(n *node) {
		if n.id == 0 && n != root {
			n.id = c.nextID
			c.nextID++
			if c.nextID > lastRowID {
				c.nextID = firstRowID // wrap around
			}
		}
		for _, child := range n.children {
			walk(child)
		}
	}
	walk(root)
}

// copyIDsToBase copies IDs from view node children to base node children
// by matching names.
func (c *Component) copyIDsToBase(viewNode, baseNode *node) {
	for _, vc := range viewNode.children {
		for _, bc := range baseNode.children {
			if vc.name == bc.name && vc.isDir == bc.isDir {
				bc.id = vc.id
				break
			}
		}
	}
}

// syncExpandedToBase walks viewTree and syncs expanded directories
// to baseTree, ensuring both trees have matching children with IDs.
func (c *Component) syncExpandedToBase() {
	var walk func(viewNode *node)
	walk = func(viewNode *node) {
		for _, vc := range viewNode.children {
			if !vc.isDir {
				continue
			}
			baseNode := c.findNodeByID(c.baseTree, vc.id)
			if baseNode == nil {
				continue
			}
			if vc.expanded {
				if baseNode.children == nil && vc.children != nil {
					_ = readChildren(c.fs, baseNode)
					c.copyIDsToBase(vc, baseNode)
				}
				baseNode.expanded = true
			} else {
				baseNode.expanded = false
			}
			if vc.expanded {
				walk(vc)
			}
		}
	}
	walk(c.viewTree)
}

// parseTreeFromCells builds a view tree from the given cells,
// preserving row IDs where present.
func (c *Component) parseTreeFromCells(cells [][]term.Cell) *node {
	root := &node{
		uri:      c.root,
		isDir:    true,
		expanded: true,
		depth:    -1,
	}

	entries := parseCellsWithIDs(cells, c.cfg)
	if len(entries) == 0 {
		return root
	}

	// Build tree from parsed entries
	stack := []*node{root}
	for _, e := range entries {
		// Pop stack until we find the parent
		for len(stack) > 1 && stack[len(stack)-1].depth >= e.depth {
			stack = stack[:len(stack)-1]
		}
		parent := stack[len(stack)-1]

		child := &node{
			id:     e.id,
			name:   e.name,
			uri:    workspaceapi.Join(parent.uri, e.name),
			isDir:  e.isDir,
			depth:  e.depth,
			parent: parent,
		}

		// For directories, check if they were expanded in the old view tree
		if child.isDir {
			if old := c.findNodeByID(c.viewTree, e.id); old != nil {
				child.expanded = old.expanded
				// Copy children if expanded (they'll be in subsequent entries)
			}
		}

		parent.children = append(parent.children, child)
		if e.isDir {
			stack = append(stack, child)
		}
	}

	return root
}

// expandAtDepth expands all directories at the given depth.
func (c *Component) expandAtDepth(n *node, depth int) {
	for _, child := range n.children {
		if child.isDir && child.depth == depth {
			if child.children == nil {
				_ = readChildren(c.fs, child)
				c.assignIDs(child)
			}
			child.expanded = true
		}
		if child.isDir && child.expanded {
			c.expandAtDepth(child, depth)
		}
	}
}

// applyOperation executes a single operation on the filesystem.
func (c *Component) applyOperation(op Operation) error {
	switch op.Type {
	case OpCreate:
		f, err := c.fs.OpenFile(
			op.NewURI.Path(),
			os.O_CREATE|os.O_WRONLY,
			0644,
		)
		if err != nil {
			return err
		}
		return f.Close()

	case OpMkdir:
		return c.fs.MkdirAll(op.NewURI.Path(), 0755)

	case OpDelete:
		return c.fs.Remove(op.URI.Path())

	case OpRename, OpMove:
		// Copy content to new location, delete old
		info, err := c.fs.Stat(op.URI.Path())
		if err != nil {
			return err
		}
		if info.IsDir() {
			// For directories, just create at new location
			// (children will be handled by subsequent operations)
			return c.fs.MkdirAll(op.NewURI.Path(), 0755)
		}
		if err := c.copyFile(op.URI.Path(), op.NewURI.Path()); err != nil {
			return err
		}
		return c.fs.Remove(op.URI.Path())

	case OpCopy:
		info, err := c.fs.Stat(op.URI.Path())
		if err != nil {
			return err
		}
		if info.IsDir() {
			return c.fs.MkdirAll(op.NewURI.Path(), 0755)
		}
		return c.copyFile(op.URI.Path(), op.NewURI.Path())
	}
	return nil
}

func (c *Component) copyFile(srcPath, dstPath string) (err error) {
	src, err := c.fs.OpenFile(srcPath, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := src.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	dst, err := c.fs.OpenFile(
		dstPath,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0644,
	)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := dst.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	buf := make([]byte, 32*1024)
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
		}
		if readErr != nil {
			break
		}
	}
	return nil
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
