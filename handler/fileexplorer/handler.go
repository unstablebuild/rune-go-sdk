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
	"context"
	"strings"

	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/component/fileexplorer"
	"github.com/unstablebuild/rune-go-sdk/handler"
	"github.com/unstablebuild/rune-go-sdk/handler/inputbox"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/tcell/v3"
)

// Handler implements tui.Handler for file explorer interaction.
// It wraps a fileexplorer.Component and provides oil.nvim-like editing.
type Handler struct {
	comp          *fileexplorer.Component
	mode          Mode
	cursor        int
	voffset       int
	width, height int
	attrs         term.Attributes
	cursorAttrs   term.Attributes

	// Edit mode state
	editBox    *inputbox.Handler
	editRow    int
	editPrefix []term.Cell // Row ID + indent + icon (preserved during edit)
	editIsDir  bool

	// Confirm mode state
	pendingOps []fileexplorer.Operation
	confirmSel int // 0 = Yes, 1 = No

	// Callbacks
	onOpen  func(uri workspaceapi.URI)
	onExit  func()
	onApply func(error)
}

var (
	_ handler.ScrollableFloatingWithAttributes = (*Handler)(nil)
)

// New creates a new Handler wrapping the given fileexplorer.Component.
func New(comp *fileexplorer.Component, opts ...Option) *Handler {
	h := &Handler{
		comp: comp,
		mode: ModeView,
		cursorAttrs: term.Attributes{
			Attrs: tcell.AttrReverse,
		},
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Resize satisfies tui.Component.
func (h *Handler) Resize(width, height int) {
	h.width, h.height = width, height
	h.comp.Resize(width, height)
	if h.editBox != nil {
		h.editBox.Resize(width-h.editPrefixWidth(), 1)
	}
	h.clampCursor()
	h.updateVoffset()
}

// Draw satisfies tui.Component.
func (h *Handler) Draw(w term.Writer) {
	switch h.mode {
	case ModeView:
		h.drawView(w)
	case ModeEdit:
		h.drawEdit(w)
	case ModeConfirm:
		h.drawView(w)
		h.drawConfirmDialog(w)
	}
}

// Handle satisfies tui.Handler.
func (h *Handler) Handle(ev term.Event) (exit, handled bool) {
	if ev.Type != term.EventKey {
		return false, false
	}

	switch h.mode {
	case ModeView:
		return h.handleView(ev)
	case ModeEdit:
		return h.handleEdit(ev)
	case ModeConfirm:
		return h.handleConfirm(ev)
	}
	return false, false
}

// Cursor satisfies tui.Handler.
func (h *Handler) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	switch h.mode {
	case ModeEdit:
		c, style, show := h.editBox.Cursor()
		c.X += h.editPrefixWidth()
		c.Y = h.editRow - h.voffset
		return c, style, show
	case ModeConfirm:
		return term.Coordinates{}, term.CursorStyleDefault, false
	default:
		return term.Coordinates{X: 0, Y: h.cursor - h.voffset},
			term.CursorStyleSteadyBar, false
	}
}

// Selection satisfies tui.Handler.
func (h *Handler) Selection() (string, bool) {
	if h.mode == ModeEdit && h.editBox != nil {
		return h.editBox.Selection()
	}
	return "", false
}

// SetAttr satisfies handler.WithAttributes.
func (h *Handler) SetAttr(attr term.Attributes) term.Attributes {
	prev := h.attrs
	h.attrs = attr
	return prev
}

// Dimensions satisfies component.Floating.
func (h *Handler) Dimensions() (int, int) {
	return h.comp.Dimensions()
}

// SeekUp satisfies component.Scrollable.
func (h *Handler) SeekUp() bool {
	if h.voffset > 0 {
		h.voffset--
		return true
	}
	return false
}

// SeekDown satisfies component.Scrollable.
func (h *Handler) SeekDown() bool {
	max := h.MaxSeekOffset()
	if h.voffset < max {
		h.voffset++
		return true
	}
	return false
}

// SeekOffset satisfies component.Scrollable.
func (h *Handler) SeekOffset() int {
	return h.voffset
}

// MaxSeekOffset satisfies component.Scrollable.
func (h *Handler) MaxSeekOffset() int {
	cells := h.comp.Cells()
	if len(cells) <= h.height {
		return 0
	}
	return len(cells) - h.height
}

// Mode returns the current mode.
func (h *Handler) Mode() Mode {
	return h.mode
}

// Component returns the underlying fileexplorer.Component.
func (h *Handler) Component() *fileexplorer.Component {
	return h.comp
}

func (h *Handler) handleView(ev term.Event) (exit, handled bool) {
	switch {
	case ev.Key == term.KeyEsc, ev.Ch == 'q':
		if h.onExit != nil {
			h.onExit()
		}
		return true, true

	case ev.Key == term.KeyArrowUp, ev.Ch == 'k':
		h.moveCursorUp()
		return false, true

	case ev.Key == term.KeyArrowDown, ev.Ch == 'j':
		h.moveCursorDown()
		return false, true

	case ev.Key == term.KeyEnter, ev.Ch == 'l':
		h.expandOrOpen()
		return false, true

	case ev.Ch == 'h':
		h.collapseOrParent()
		return false, true

	case ev.Ch == 'r', ev.Ch == 'i':
		h.enterEditMode()
		return false, true

	case ev.Ch == 'a':
		h.addFile()
		return false, true

	case ev.Ch == 'A':
		h.addDir()
		return false, true

	case ev.Ch == 'd':
		h.deleteLine()
		return false, true

	case ev.Ch == 's' && ev.Mod == term.ModCtrl:
		h.showConfirmDialog()
		return false, true

	case ev.Key == term.KeyPgup:
		h.pageUp()
		return false, true

	case ev.Key == term.KeyPgdn:
		h.pageDown()
		return false, true

	case ev.Key == term.KeyHome, ev.Ch == 'g' && ev.Mod == 0:
		h.cursor = 0
		h.updateVoffset()
		return false, true

	case ev.Ch == 'G':
		cells := h.comp.Cells()
		if len(cells) > 0 {
			h.cursor = len(cells) - 1
		}
		h.updateVoffset()
		return false, true
	}

	return false, false
}

func (h *Handler) handleEdit(ev term.Event) (exit, handled bool) {
	switch ev.Key {
	case term.KeyEsc:
		h.cancelEdit()
		return false, true
	case term.KeyEnter:
		h.commitEdit()
		return false, true
	}

	// Delegate to inputbox
	_, handled = h.editBox.Handle(ev)
	return false, handled
}

func (h *Handler) handleConfirm(ev term.Event) (exit, handled bool) {
	switch {
	case ev.Key == term.KeyEsc, ev.Ch == 'n', ev.Ch == 'N':
		h.mode = ModeView
		h.pendingOps = nil
		return false, true

	case ev.Key == term.KeyEnter, ev.Ch == 'y', ev.Ch == 'Y':
		if h.confirmSel == 0 {
			h.applyChanges()
		}
		h.mode = ModeView
		h.pendingOps = nil
		return false, true

	case ev.Key == term.KeyArrowLeft, ev.Key == term.KeyArrowRight,
		ev.Key == term.KeyTab, ev.Ch == 'h', ev.Ch == 'l':
		h.confirmSel = 1 - h.confirmSel
		return false, true

	case ev.Ch == 'j', ev.Key == term.KeyArrowDown:
		// Navigate operation list (future enhancement)
		return false, true

	case ev.Ch == 'k', ev.Key == term.KeyArrowUp:
		// Navigate operation list (future enhancement)
		return false, true
	}

	return false, false
}

func (h *Handler) drawView(w term.Writer) {
	cells := h.comp.Cells()

	// Draw background
	for y := range h.height {
		for x := range h.width {
			w.UnionAttributes(term.Coordinates{X: x, Y: y}, h.attrs)
		}
	}

	// Draw visible rows
	for y := range h.height {
		rowIdx := h.voffset + y
		if rowIdx >= len(cells) {
			break
		}
		row := cells[rowIdx]

		// Skip row ID (first cell)
		var x int
		for i, cell := range row {
			if i == 0 {
				continue // Skip row ID
			}
			if x >= h.width {
				break
			}
			w.SetCell(term.Coordinates{X: x, Y: y}, cell)
			x++
			if cell.Width > 1 {
				x += int(cell.Width) - 1
			}
		}

		// Apply cursor highlight
		if rowIdx == h.cursor {
			for x := range h.width {
				w.UnionAttributes(term.Coordinates{X: x, Y: y}, h.cursorAttrs)
			}
		}
	}
}

func (h *Handler) drawEdit(w term.Writer) {
	cells := h.comp.Cells()

	// Draw background
	for y := range h.height {
		for x := range h.width {
			w.UnionAttributes(term.Coordinates{X: x, Y: y}, h.attrs)
		}
	}

	// Draw visible rows
	for y := range h.height {
		rowIdx := h.voffset + y
		if rowIdx >= len(cells) {
			break
		}

		if rowIdx == h.editRow {
			// Draw edit row: prefix + inputbox
			h.drawEditRow(w, y)
		} else {
			row := cells[rowIdx]
			var x int
			for i, cell := range row {
				if i == 0 {
					continue // Skip row ID
				}
				if x >= h.width {
					break
				}
				w.SetCell(term.Coordinates{X: x, Y: y}, cell)
				x++
				if cell.Width > 1 {
					x += int(cell.Width) - 1
				}
			}
		}

		// Apply cursor highlight for edit row
		if rowIdx == h.editRow {
			for x := range h.width {
				w.UnionAttributes(term.Coordinates{X: x, Y: y}, h.cursorAttrs)
			}
		}
	}
}

func (h *Handler) drawEditRow(w term.Writer, y int) {
	// Draw prefix (row ID is skipped, start from index 1)
	var x int
	for i, cell := range h.editPrefix {
		if i == 0 {
			continue // Skip row ID
		}
		if x >= h.width {
			break
		}
		w.SetCell(term.Coordinates{X: x, Y: y}, cell)
		x++
		if cell.Width > 1 {
			x += int(cell.Width) - 1
		}
	}

	// Draw inputbox
	prefixWidth := h.editPrefixWidth()
	editWriter := &offsetWriter{
		w:       w,
		offsetX: prefixWidth,
		offsetY: y,
	}
	h.editBox.Draw(editWriter)
}

func (h *Handler) drawConfirmDialog(w term.Writer) {
	lines := h.buildConfirmLines()

	// Calculate dimensions
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	dialogWidth := maxWidth + 4
	dialogHeight := len(lines) + 4 // +4 for border and buttons

	// Center the dialog
	startX := (h.width - dialogWidth) / 2
	startY := (h.height - dialogHeight) / 2
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	// Draw background
	bgAttrs := term.Attributes{Bg: tcell.ColorDarkBlue}
	for y := startY; y < startY+dialogHeight && y < h.height; y++ {
		for x := startX; x < startX+dialogWidth && x < h.width; x++ {
			w.SetCell(term.Coordinates{X: x, Y: y}, term.Cell{
				Ch:         ' ',
				Width:      1,
				Attributes: bgAttrs,
			})
		}
	}

	// Draw border
	h.drawBox(w, startX, startY, dialogWidth, dialogHeight, bgAttrs)

	// Draw content
	contentAttrs := term.Attributes{Bg: tcell.ColorDarkBlue, Fg: tcell.ColorWhite}
	for i, line := range lines {
		y := startY + 1 + i
		if y >= h.height-1 {
			break
		}
		for x, ch := range line {
			px := startX + 2 + x
			if px >= startX+dialogWidth-1 {
				break
			}
			w.SetCell(term.Coordinates{X: px, Y: y}, term.Cell{
				Ch:         ch,
				Width:      1,
				Attributes: contentAttrs,
			})
		}
	}

	// Draw buttons
	buttonY := startY + dialogHeight - 2
	yesX := startX + dialogWidth/2 - 8
	noX := startX + dialogWidth/2 + 2

	yesAttrs := contentAttrs
	noAttrs := contentAttrs
	if h.confirmSel == 0 {
		yesAttrs.Attrs = tcell.AttrReverse
	} else {
		noAttrs.Attrs = tcell.AttrReverse
	}

	h.drawButton(w, yesX, buttonY, "[Yes]", yesAttrs)
	h.drawButton(w, noX, buttonY, "[No]", noAttrs)
}

func (h *Handler) buildConfirmLines() []string {
	lines := []string{"Apply the following changes?", ""}
	for _, op := range h.pendingOps {
		var line string
		switch op.Type {
		case fileexplorer.OpCreate:
			line = "  + Create: " + op.NewURI.Name()
		case fileexplorer.OpMkdir:
			line = "  + Mkdir:  " + op.NewURI.Name() + "/"
		case fileexplorer.OpDelete:
			line = "  - Delete: " + op.URI.Name()
		case fileexplorer.OpRename:
			line = "  ~ Rename: " + op.URI.Name() + " -> " + op.NewURI.Name()
		}
		lines = append(lines, line)
	}
	return lines
}

func (h *Handler) drawBox(
	w term.Writer, x, y, width, height int, attrs term.Attributes,
) {
	fgAttrs := term.Attributes{Bg: attrs.Bg, Fg: tcell.ColorWhite}

	// Corners
	w.SetCell(term.Coordinates{X: x, Y: y},
		term.Cell{Ch: '┌', Width: 1, Attributes: fgAttrs})
	w.SetCell(term.Coordinates{X: x + width - 1, Y: y},
		term.Cell{Ch: '┐', Width: 1, Attributes: fgAttrs})
	w.SetCell(term.Coordinates{X: x, Y: y + height - 1},
		term.Cell{Ch: '└', Width: 1, Attributes: fgAttrs})
	w.SetCell(term.Coordinates{X: x + width - 1, Y: y + height - 1},
		term.Cell{Ch: '┘', Width: 1, Attributes: fgAttrs})

	// Horizontal lines
	for i := x + 1; i < x+width-1; i++ {
		w.SetCell(term.Coordinates{X: i, Y: y},
			term.Cell{Ch: '─', Width: 1, Attributes: fgAttrs})
		w.SetCell(term.Coordinates{X: i, Y: y + height - 1},
			term.Cell{Ch: '─', Width: 1, Attributes: fgAttrs})
	}

	// Vertical lines
	for i := y + 1; i < y+height-1; i++ {
		w.SetCell(term.Coordinates{X: x, Y: i},
			term.Cell{Ch: '│', Width: 1, Attributes: fgAttrs})
		w.SetCell(term.Coordinates{X: x + width - 1, Y: i},
			term.Cell{Ch: '│', Width: 1, Attributes: fgAttrs})
	}
}

func (h *Handler) drawButton(
	w term.Writer, x, y int, label string, attrs term.Attributes,
) {
	for i, ch := range label {
		w.SetCell(term.Coordinates{X: x + i, Y: y}, term.Cell{
			Ch:         ch,
			Width:      1,
			Attributes: attrs,
		})
	}
}

func (h *Handler) moveCursorUp() {
	if h.cursor > 0 {
		h.cursor--
		h.updateVoffset()
	}
}

func (h *Handler) moveCursorDown() {
	cells := h.comp.Cells()
	if h.cursor < len(cells)-1 {
		h.cursor++
		h.updateVoffset()
	}
}

func (h *Handler) pageUp() {
	h.cursor -= h.height
	if h.cursor < 0 {
		h.cursor = 0
	}
	h.updateVoffset()
}

func (h *Handler) pageDown() {
	cells := h.comp.Cells()
	h.cursor += h.height
	if h.cursor >= len(cells) {
		h.cursor = len(cells) - 1
	}
	if h.cursor < 0 {
		h.cursor = 0
	}
	h.updateVoffset()
}

func (h *Handler) expandOrOpen() {
	uri, isFile := h.comp.ExpandNodeAt(term.Coordinates{Y: h.cursor})
	if isFile && h.onOpen != nil {
		h.onOpen(uri)
	}
	h.clampCursor()
}

func (h *Handler) collapseOrParent() {
	// Try to collapse current node if it's a directory
	h.comp.CollapseLevel(term.Coordinates{Y: h.cursor})
	h.clampCursor()
}

func (h *Handler) enterEditMode() {
	cells := h.comp.Cells()
	if h.cursor >= len(cells) {
		return
	}

	row := cells[h.cursor]
	if len(row) < 2 {
		return
	}

	// Find where the filename starts (after indent + icon + space)
	nameStart := h.findNameStart(row)
	if nameStart >= len(row) {
		return
	}

	// Extract prefix (row ID + indent + icon + space)
	h.editPrefix = make([]term.Cell, nameStart)
	copy(h.editPrefix, row[:nameStart])

	// Extract filename
	var nameBuilder strings.Builder
	for _, cell := range row[nameStart:] {
		nameBuilder.WriteRune(cell.Ch)
	}
	name := nameBuilder.String()
	h.editIsDir = strings.HasSuffix(name, "/")
	name = strings.TrimSuffix(name, "/")

	// Create inputbox with filename
	h.editBox = inputbox.New()
	h.editBox.Resize(h.width-h.editPrefixWidth(), 1)
	for _, ch := range name {
		h.editBox.Handle(term.Event{Type: term.EventKey, Ch: ch})
	}

	h.editRow = h.cursor
	h.mode = ModeEdit
}

func (h *Handler) cancelEdit() {
	h.mode = ModeView
	h.editBox = nil
	h.editPrefix = nil
}

func (h *Handler) commitEdit() {
	if h.editBox == nil {
		h.cancelEdit()
		return
	}

	newName := h.editBox.Text()
	if newName == "" {
		h.cancelEdit()
		return
	}

	// Reconstruct the row
	var newRow []term.Cell
	newRow = append(newRow, h.editPrefix...)

	// Add filename
	for _, ch := range newName {
		newRow = append(newRow, term.Cell{Ch: ch, Width: 1})
	}

	// Add trailing slash if directory
	if h.editIsDir {
		newRow = append(newRow, term.Cell{Ch: '/', Width: 1})
	}

	// Update cells
	cells := h.comp.Cells()
	if h.editRow < len(cells) {
		cells[h.editRow] = newRow
		h.comp.SetCells(cells)
	}

	h.mode = ModeView
	h.editBox = nil
	h.editPrefix = nil
}

func (h *Handler) addFile() {
	h.addEntry(false)
}

func (h *Handler) addDir() {
	h.addEntry(true)
}

func (h *Handler) addEntry(isDir bool) {
	cells := h.comp.Cells()
	if len(cells) == 0 {
		return
	}

	depth := h.depthAtCursor(cells)
	name := "newfile.txt"
	if isDir {
		name = "newdir"
	}

	newLine := h.buildLine(depth, name, isDir)
	pos := min(h.cursor+1, len(cells))

	newCells := make([][]term.Cell, 0, len(cells)+1)
	newCells = append(newCells, cells[:pos]...)
	newCells = append(newCells, newLine)
	newCells = append(newCells, cells[pos:]...)

	h.comp.SetCells(newCells)
	h.cursor = pos
	h.clampCursor()

	// Enter edit mode for the new entry
	h.enterEditMode()
}

func (h *Handler) deleteLine() {
	cells := h.comp.Cells()
	if len(cells) == 0 || h.cursor >= len(cells) {
		return
	}

	newCells := append(cells[:h.cursor], cells[h.cursor+1:]...)
	h.comp.SetCells(newCells)
	h.clampCursor()
}

func (h *Handler) showConfirmDialog() {
	changes := h.comp.Changes()
	if len(changes) == 0 {
		return
	}

	h.pendingOps = OrderOperations(changes)
	h.confirmSel = 0
	h.mode = ModeConfirm
}

func (h *Handler) applyChanges() {
	if len(h.pendingOps) == 0 {
		return
	}

	err := h.comp.ApplyChanges(h.pendingOps)
	if h.onApply != nil {
		h.onApply(err)
	}
}

func (h *Handler) clampCursor() {
	cells := h.comp.Cells()
	if h.cursor >= len(cells) {
		if len(cells) > 0 {
			h.cursor = len(cells) - 1
		} else {
			h.cursor = 0
		}
	}
	if h.cursor < 0 {
		h.cursor = 0
	}
}

func (h *Handler) updateVoffset() {
	if h.height <= 0 {
		return
	}

	if h.cursor < h.voffset {
		h.voffset = h.cursor
	} else if h.cursor >= h.voffset+h.height {
		h.voffset = h.cursor - h.height + 1
	}

	max := h.MaxSeekOffset()
	if h.voffset > max {
		h.voffset = max
	}
	if h.voffset < 0 {
		h.voffset = 0
	}
}

func (h *Handler) findNameStart(row []term.Cell) int {
	// Format: [rowID][indent...][icon][space][name]
	// rowID is at index 0, then pairs of (indent, space), then icon, space
	i := 1 // Skip row ID
	for i+1 < len(row) {
		ch := row[i].Ch
		if ch != '│' && ch != ' ' {
			// Found icon, skip icon + space
			if i+2 <= len(row) {
				return i + 2
			}
			return i
		}
		if ch == '│' {
			i += 2 // Skip indent + space pair
		} else {
			i++
		}
	}
	return i
}

func (h *Handler) editPrefixWidth() int {
	if len(h.editPrefix) <= 1 {
		return 0
	}
	// Count visible width (skip row ID at index 0)
	width := 0
	for i := 1; i < len(h.editPrefix); i++ {
		width++
		if h.editPrefix[i].Width > 1 {
			width += int(h.editPrefix[i].Width) - 1
		}
	}
	return width
}

func (h *Handler) depthAtCursor(cells [][]term.Cell) int {
	if h.cursor >= len(cells) {
		return 0
	}
	row := cells[h.cursor]
	depth := 0
	i := 1 // Skip row ID
	for i+1 < len(row) {
		if row[i].Ch == '│' && row[i+1].Ch == ' ' {
			depth++
			i += 2
		} else {
			break
		}
	}
	return depth
}

func (h *Handler) buildLine(depth int, name string, isDir bool) []term.Cell {
	var b strings.Builder
	for range depth {
		b.WriteString("│ ")
	}
	b.WriteString("  ") // icon placeholder + space
	b.WriteString(name)
	if isDir {
		b.WriteRune('/')
	}

	lineCells := term.StringToCells(b.String())
	if len(lineCells) == 0 {
		return nil
	}

	// Prepend a zero row ID (new node)
	row := make([]term.Cell, 0, len(lineCells[0])+1)
	row = append(row, term.Cell{Ch: 0, Width: 0})
	row = append(row, lineCells[0]...)
	return row
}

// offsetWriter wraps a Writer and applies an offset to all coordinates.
type offsetWriter struct {
	w                term.Writer
	offsetX, offsetY int
}

func (o *offsetWriter) Context() context.Context {
	return o.w.Context()
}

func (o *offsetWriter) SetCell(pos term.Coordinates, cell term.Cell) {
	o.w.SetCell(term.Coordinates{
		X: pos.X + o.offsetX,
		Y: pos.Y + o.offsetY,
	}, cell)
}

func (o *offsetWriter) UnionAttributes(pos term.Coordinates, attr term.Attributes) {
	o.w.UnionAttributes(term.Coordinates{
		X: pos.X + o.offsetX,
		Y: pos.Y + o.offsetY,
	}, attr)
}
