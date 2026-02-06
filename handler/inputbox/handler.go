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

package inputbox

import (
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/handler"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/tcell/v3"
)

// Handler implements a multi-line text input handler with wrapping,
// text selection and, word-wise cursor movement. Note that text selection is
// limited by your terminal's ability to handle shift+arrow keys.
type Handler struct {
	text        []rune
	cursor      int
	voffset     int
	width       int
	height      int
	attrs       term.Attributes
	selAnchor   int // -1 means no selection
	placeholder *component.ResponsiveString
}

var _ handler.WithAttributesResponsive = (*Handler)(nil)

// New creates and initializes a new input box with default attributes.
func New(opts ...Option) *Handler {
	ret := &Handler{
		text:      make([]rune, 0, 64),
		selAnchor: -1,
	}
	for _, o := range opts {
		o(ret)
	}
	return ret
}

// SetAttr satisfies WithAttributes.
func (ib *Handler) SetAttr(attr term.Attributes) (ret term.Attributes) {
	ret = ib.attrs
	ib.attrs = attr
	ib.updatePlaceholderBackground()
	return
}

// Text returns the current text content of the input box.
func (ib *Handler) Text() string {
	return string(ib.text)
}

// Clear resets the input box to empty state.
func (ib *Handler) Clear() {
	ib.text = ib.text[:0]
	ib.cursor = 0
	ib.selAnchor = -1
	ib.updateVoffset()
}

// Resize satisfies tui.Handler
func (ib *Handler) Resize(width, height int) {
	ib.width = width
	ib.height = height
	if ib.placeholder != nil {
		ib.placeholder.Resize(width, height)
		ib.updatePlaceholderBackground()
	}
	ib.updateVoffset()
}

// Height satisfies Responsive. Returns the number of lines needed to
// display all text content and cursor given the specified width.
func (ib *Handler) Height(width int) int {
	if width <= 0 {
		return 0
	}

	if len(ib.text) == 0 {
		return 1
	}

	textLines := (len(ib.text) + width - 1) / width

	// cursor is at end; wraps to next line
	if len(ib.text)%width == 0 && ib.cursor == len(ib.text) {
		return textLines + 1
	}

	return textLines
}

// Draw satisfies tui.Component.
func (ib *Handler) Draw(w term.Writer) {
	if ib.height == 0 || ib.width == 0 {
		return
	}

	for y := range ib.height {
		for x := range ib.width {
			w.UnionAttributes(term.Coordinates{X: x, Y: y}, ib.attrs)
		}
	}

	if len(ib.text) == 0 && ib.placeholder != nil {
		ib.placeholder.Draw(w)
		return
	}

	startIdx := ib.voffset * ib.width
	endIdx := min(startIdx+ib.height*ib.width, len(ib.text))

	selStart, selEnd := ib.selectionRange()

	var x, y int
	for i := startIdx; i < endIdx; i++ {
		attrs := ib.attrs
		if i >= selStart && i < selEnd {
			attrs.Attrs |= tcell.AttrReverse
		}
		w.SetCell(term.Coordinates{X: x, Y: y}, term.Cell{
			Ch:         ib.text[i],
			Attributes: attrs,
			Width:      1,
		})
		x++
		if x < ib.width {
			continue
		}
		x = 0
		y++
		if y >= ib.height {
			break
		}
	}
}

// Handle satisfies tui.Handler
func (ib *Handler) Handle(ev term.Event) (exit, handled bool) {
	if ev.Type != term.EventKey {
		return false, false
	}

	if ib.hasSelection() && ev.Ch != 0 &&
		ev.Mod != term.ModCtrl && ev.Mod != term.ModAlt {
		ib.deleteSelection()
	}

	switch ev.Key {
	case term.KeyEnter:
		return true, true
	case term.KeyEsc:
		return true, true
	}

	defer ib.updateVoffset()

	switch {
	// shift+ctrl/alt word selection
	case ev.Key == term.KeyArrowLeft &&
		ev.Mod == term.ModCtrlShift,
		ev.Key == term.KeyArrowLeft &&
			ev.Mod == term.ModAltShift:
		ib.startSelection()
		ib.moveWordLeft()
		return false, true

	case ev.Key == term.KeyArrowRight &&
		ev.Mod == term.ModCtrlShift,
		ev.Key == term.KeyArrowRight &&
			ev.Mod == term.ModAltShift:
		ib.startSelection()
		ib.moveWordRight()
		return false, true

	// ctrl/alt word movement
	case ev.Key == term.KeyArrowLeft &&
		ev.Mod == term.ModCtrl,
		ev.Key == term.KeyArrowLeft &&
			ev.Mod == term.ModAlt:
		ib.clearSelection()
		ib.moveWordLeft()
		return false, true

	case ev.Key == term.KeyArrowRight &&
		ev.Mod == term.ModCtrl,
		ev.Key == term.KeyArrowRight &&
			ev.Mod == term.ModAlt:
		ib.clearSelection()
		ib.moveWordRight()
		return false, true

	// word deletion
	case ev.Key == term.KeyBackspace &&
		ev.Mod == term.ModCtrl,
		ev.Key == term.KeyBackspace &&
			ev.Mod == term.ModAlt:
		if ib.hasSelection() {
			ib.deleteSelection()
		} else {
			ib.deleteWordBackward()
		}
		return false, true

	case ev.Key == term.KeyDelete &&
		ev.Mod == term.ModCtrl,
		ev.Key == term.KeyDelete &&
			ev.Mod == term.ModAlt:
		if ib.hasSelection() {
			ib.deleteSelection()
		} else {
			ib.deleteWordForward()
		}
		return false, true

	// shift+arrow char selection
	case ev.Key == term.KeyArrowLeft &&
		ev.Mod == term.ModShift:
		ib.startSelection()
		if ib.cursor > 0 {
			ib.cursor--
		}
		return false, true

	case ev.Key == term.KeyArrowRight &&
		ev.Mod == term.ModShift:
		ib.startSelection()
		if ib.cursor < len(ib.text) {
			ib.cursor++
		}
		return false, true

	// shift+home/end selection
	case ev.Key == term.KeyHome &&
		ev.Mod == term.ModShift:
		ib.startSelection()
		ib.cursor = 0
		return false, true

	case ev.Key == term.KeyEnd &&
		ev.Mod == term.ModShift:
		ib.startSelection()
		ib.cursor = len(ib.text)
		return false, true

	case ev.Key == term.KeyBackspace:
		if ib.hasSelection() {
			ib.deleteSelection()
		} else if ib.cursor > 0 {
			ib.text = append(
				ib.text[:ib.cursor-1],
				ib.text[ib.cursor:]...,
			)
			ib.cursor--
		}
		return false, true

	case ev.Key == term.KeyDelete:
		if ib.hasSelection() {
			ib.deleteSelection()
		} else if ib.cursor < len(ib.text) {
			ib.text = append(
				ib.text[:ib.cursor],
				ib.text[ib.cursor+1:]...,
			)
		}
		return false, true

	case ev.Key == term.KeyArrowLeft:
		ib.clearSelection()
		if ib.cursor > 0 {
			ib.cursor--
		}
		return false, true

	case ev.Key == term.KeyArrowRight:
		ib.clearSelection()
		if ib.cursor < len(ib.text) {
			ib.cursor++
		}
		return false, true

	case ev.Key == term.KeyHome:
		ib.clearSelection()
		ib.cursor = 0
		return false, true

	case ev.Key == term.KeyEnd:
		ib.clearSelection()
		ib.cursor = len(ib.text)
		return false, true

	// emacs-style cursor movement
	case ev.Ch == 'a' && ev.Mod == term.ModCtrl:
		// <ctrl-a> beginning of line
		ib.clearSelection()
		ib.cursor = 0
		return false, true

	case ev.Ch == 'e' && ev.Mod == term.ModCtrl:
		// <ctrl-e> end of line
		ib.clearSelection()
		ib.cursor = len(ib.text)
		return false, true

	case ev.Ch == 'b' && ev.Mod == term.ModCtrl:
		// <ctrl+b> backward char
		ib.clearSelection()
		if ib.cursor > 0 {
			ib.cursor--
		}
		return false, true

	case ev.Ch == 'f' && ev.Mod == term.ModCtrl:
		// <ctrl-f> forward char
		ib.clearSelection()
		if ib.cursor < len(ib.text) {
			ib.cursor++
		}
		return false, true

	case ev.Ch == 'd' && ev.Mod == term.ModCtrl:
		// <ctrl-d> delete char at cursor
		if ib.hasSelection() {
			ib.deleteSelection()
		} else if ib.cursor < len(ib.text) {
			ib.text = append(
				ib.text[:ib.cursor],
				ib.text[ib.cursor+1:]...,
			)
		}
		return false, true

	case ev.Ch == 'h' && ev.Mod == term.ModCtrl:
		// <ctrl-h> backspace
		if ib.hasSelection() {
			ib.deleteSelection()
		} else if ib.cursor > 0 {
			ib.text = append(
				ib.text[:ib.cursor-1],
				ib.text[ib.cursor:]...,
			)
			ib.cursor--
		}
		return false, true

	case ev.Ch == 'k' && ev.Mod == term.ModCtrl:
		// <ctrl-k> kill to end of line
		ib.clearSelection()
		if ib.cursor < len(ib.text) {
			ib.text = ib.text[:ib.cursor]
		}
		return false, true

	case ev.Ch == 'u' && ev.Mod == term.ModCtrl:
		// <ctrl+u> kill entire line
		ib.Clear()
		return false, true

	case ev.Ch == 'b' && ev.Mod == term.ModAlt:
		// <alt-b> backward word
		ib.clearSelection()
		ib.moveWordLeft()
		return false, true

	case ev.Ch == 'f' && ev.Mod == term.ModAlt:
		// <alt-f> forward word
		ib.clearSelection()
		ib.moveWordRight()
		return false, true

	case ev.Ch == 'd' && ev.Mod == term.ModAlt:
		// <alt-d> delete word forward
		if ib.hasSelection() {
			ib.deleteSelection()
		} else {
			ib.deleteWordForward()
		}
		return false, true

	case ev.Ch == 't' && ev.Mod == term.ModCtrl:
		// <ctrl-t> transpose characters
		ib.clearSelection()
		ib.transpose()
		return false, true

	case ev.Key == term.KeySpace:
		ev.Ch = ' '
		fallthrough

	case ev.Ch != 0:
		ib.clearSelection()
		ib.text = append(
			ib.text[:ib.cursor],
			append([]rune{ev.Ch}, ib.text[ib.cursor:]...)...,
		)
		ib.cursor++
		return false, true
	}

	return false, false
}

// Cursor satisfies tui.Handler
func (ib *Handler) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	cursorLine := ib.cursor / ib.width
	cursorCol := ib.cursor % ib.width
	cursorY := cursorLine - ib.voffset
	return term.Coordinates{X: cursorCol, Y: cursorY},
		term.CursorStyleSteadyBar, true
}

// Selection satisfies tui.Handler
func (ib *Handler) Selection() (string, bool) {
	if !ib.hasSelection() {
		return "", false
	}
	start, end := ib.selectionRange()
	return string(ib.text[start:end]), true
}

func (ib *Handler) hasSelection() bool {
	return ib.selAnchor >= 0 && ib.selAnchor != ib.cursor
}

func (ib *Handler) selectionRange() (int, int) {
	if ib.selAnchor < 0 {
		return 0, 0
	}
	if ib.selAnchor < ib.cursor {
		return ib.selAnchor, ib.cursor
	}
	return ib.cursor, ib.selAnchor
}

func (ib *Handler) clearSelection() {
	ib.selAnchor = -1
}

// startSelection sets the anchor to the current cursor position
// if no selection is active yet.
func (ib *Handler) startSelection() {
	if ib.selAnchor < 0 {
		ib.selAnchor = ib.cursor
	}
}

func (ib *Handler) deleteSelection() {
	if !ib.hasSelection() {
		return
	}
	start, end := ib.selectionRange()
	ib.text = append(ib.text[:start], ib.text[end:]...)
	ib.cursor = start
	ib.selAnchor = -1
}

func isWordChar(r rune) bool {
	return r == '_' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9')
}

func (ib *Handler) moveWordLeft() {
	for ib.cursor > 0 && !isWordChar(ib.text[ib.cursor-1]) {
		ib.cursor--
	}
	for ib.cursor > 0 && isWordChar(ib.text[ib.cursor-1]) {
		ib.cursor--
	}
}

func (ib *Handler) moveWordRight() {
	n := len(ib.text)
	for ib.cursor < n && !isWordChar(ib.text[ib.cursor]) {
		ib.cursor++
	}
	for ib.cursor < n && isWordChar(ib.text[ib.cursor]) {
		ib.cursor++
	}
}

func (ib *Handler) deleteWordBackward() {
	start := ib.cursor
	for start > 0 && !isWordChar(ib.text[start-1]) {
		start--
	}
	for start > 0 && isWordChar(ib.text[start-1]) {
		start--
	}
	ib.text = append(ib.text[:start], ib.text[ib.cursor:]...)
	ib.cursor = start
}

func (ib *Handler) deleteWordForward() {
	end := ib.cursor
	n := len(ib.text)
	for end < n && !isWordChar(ib.text[end]) {
		end++
	}
	for end < n && isWordChar(ib.text[end]) {
		end++
	}
	ib.text = append(ib.text[:ib.cursor], ib.text[end:]...)
}

func (ib *Handler) transpose() {
	if ib.cursor < 2 {
		return
	}
	ib.text[ib.cursor-2], ib.text[ib.cursor-1] =
		ib.text[ib.cursor-1], ib.text[ib.cursor-2]
}

// updateVoffset adjusts the vertical offset to keep the cursor visible
// and show as much text as possible.
func (ib *Handler) updateVoffset() {
	if ib.width <= 0 || ib.height <= 0 {
		return
	}

	cursorLine := ib.cursor / ib.width
	totalLines := cursorLine + 1
	if len(ib.text) > 0 {
		textLines := (len(ib.text) + ib.width - 1) / ib.width
		if textLines > totalLines {
			totalLines = textLines
		}
	}

	if totalLines <= ib.height {
		ib.voffset = 0
		return
	}

	if cursorLine < ib.voffset {
		ib.voffset = cursorLine
	} else if cursorLine >= ib.voffset+ib.height {
		ib.voffset = cursorLine - ib.height + 1
	}

	if ib.voffset > 0 && ib.voffset*ib.width >= len(ib.text) {
		if len(ib.text) > 0 {
			ib.voffset = (len(ib.text) - 1) / ib.width
		} else {
			ib.voffset = 0
		}
		if cursorLine >= ib.voffset+ib.height {
			ib.voffset = cursorLine - ib.height + 1
		}
	}
}

func (ib *Handler) updatePlaceholderBackground() {
	if ib.placeholder != nil {
		_ = ib.placeholder.SetAttr(term.Attributes{Bg: ib.attrs.Bg})
	}
}
