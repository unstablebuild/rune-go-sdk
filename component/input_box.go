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

package component

import (
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
	"github.com/unstablebuild/tcell/v3"
)

// InputBox implements a multi-line text input handler with wrapping.
type InputBox struct {
	text     []rune
	cursor   int
	voffset  int
	width    int
	height   int
	attrs    term.Attributes
	selected bool
}

var _ tui.Handler = (*InputBox)(nil)
var _ WithAttributes = (*InputBox)(nil)
var _ Responsive = (*InputBox)(nil)

// NewInputBox creates and initializes a new InputBox with default attributes.
func NewInputBox() *InputBox {
	return &InputBox{
		text:   make([]rune, 0, 64),
		cursor: 0,
		attrs:  term.Attributes{},
	}
}

// SetAttr satisfies WithAttributes.
func (ib *InputBox) SetAttr(attr term.Attributes) (ret term.Attributes) {
	ret = ib.attrs
	ib.attrs = attr
	return
}

// Text returns the current text content of the input box.
func (ib *InputBox) Text() string {
	return string(ib.text)
}

// Clear resets the input box to empty state.
func (ib *InputBox) Clear() {
	ib.text = ib.text[:0]
	ib.cursor = 0
	ib.selected = false
	ib.updateVoffset()
}

// Resize satisfies tui.Component
func (ib *InputBox) Resize(width, height int) {
	ib.width = width
	ib.height = height
	ib.updateVoffset()
}

// Height satisfies Responsive. Returns the number of lines needed to
// display all text content and cursor given the specified width.
func (ib *InputBox) Height(width int) int {
	if width <= 0 {
		return 0
	}

	if len(ib.text) == 0 {
		return 1 // at least 1 line for empty input/cursor
	}

	// Calculate lines based on text length
	textLines := (len(ib.text) + width - 1) / width

	// When text fills lines exactly and cursor is at end, it wraps to next line
	if len(ib.text)%width == 0 && ib.cursor == len(ib.text) {
		return textLines + 1
	}

	return textLines
}

// updateVoffset adjusts the vertical offset to keep the cursor visible
// and show as much text as possible.
func (ib *InputBox) updateVoffset() {
	if ib.width <= 0 || ib.height <= 0 {
		return
	}

	// Calculate which line the cursor is on
	cursorLine := ib.cursor / ib.width

	// Calculate total lines needed (text + cursor position)
	totalLines := cursorLine + 1
	if len(ib.text) > 0 {
		textLines := (len(ib.text) + ib.width - 1) / ib.width
		if textLines > totalLines {
			totalLines = textLines
		}
	}

	// If all content fits, show from top
	if totalLines <= ib.height {
		ib.voffset = 0
		return
	}

	// Adjust voffset to keep cursor visible
	if cursorLine < ib.voffset {
		ib.voffset = cursorLine
	} else if cursorLine >= ib.voffset+ib.height {
		ib.voffset = cursorLine - ib.height + 1
	}

	// Additionally, don't scroll past the text content
	// If the first visible line has no text, scroll up to show text
	if ib.voffset > 0 && ib.voffset*ib.width >= len(ib.text) {
		// First visible line is beyond text, scroll up to last text line
		if len(ib.text) > 0 {
			ib.voffset = (len(ib.text) - 1) / ib.width
		} else {
			ib.voffset = 0
		}
		// Ensure cursor is still visible
		if cursorLine >= ib.voffset+ib.height {
			ib.voffset = cursorLine - ib.height + 1
		}
	}
}

// Draw satisfies tui.Component
func (ib *InputBox) Draw(w term.Writer) {
	if ib.height == 0 || ib.width == 0 {
		return
	}

	// Calculate starting index in text for visible area
	startIdx := ib.voffset * ib.width
	endIdx := startIdx + ib.height*ib.width
	if endIdx > len(ib.text) {
		endIdx = len(ib.text)
	}

	// Draw visible text
	y := 0
	x := 0
	for i := startIdx; i < endIdx; i++ {
		attrs := ib.attrs
		// Highlight selected text
		if ib.selected {
			attrs.Attrs |= tcell.AttrReverse
		}
		w.SetCell(term.Coordinates{X: x, Y: y}, term.Cell{
			Ch:         ib.text[i],
			Attributes: attrs,
			Width:      1,
		})
		x++
		if x >= ib.width {
			x = 0
			y++
			if y >= ib.height {
				break
			}
		}
	}
}

// Handle satisfies tui.Handler
func (ib *InputBox) Handle(ev term.Event) (exit, handled bool) {
	if ev.Type != term.EventKey {
		return false, false
	}

	// Handle selection deletion on next input
	if ib.selected && ev.Ch != 0 {
		ib.Clear()
	}

	switch ev.Key {
	case term.KeyEnter:
		return true, true
	case term.KeyEsc:
		return true, true
	}

	defer ib.updateVoffset()

	switch {
	case ev.Key == term.KeyBackspace:
		if ib.selected {
			ib.Clear()
		} else if ib.cursor > 0 {
			ib.text = append(ib.text[:ib.cursor-1], ib.text[ib.cursor:]...)
			ib.cursor--
		}
		return false, true

	case ev.Key == term.KeyDelete:
		if ib.selected {
			ib.Clear()
		} else if ib.cursor < len(ib.text) {
			ib.text = append(ib.text[:ib.cursor], ib.text[ib.cursor+1:]...)
		}
		return false, true

	case ev.Key == term.KeyArrowLeft:
		ib.selected = false
		if ib.cursor > 0 {
			ib.cursor--
		}
		return false, true

	case ev.Key == term.KeyArrowRight:
		ib.selected = false
		if ib.cursor < len(ib.text) {
			ib.cursor++
		}
		return false, true

	case ev.Key == term.KeyHome:
		ib.cursor = 0
		ib.selected = false
		return false, true

	case ev.Key == term.KeyEnd:
		ib.cursor = len(ib.text)
		ib.selected = false
		return false, true

	// Emacs-style cursor movement
	case ev.Ch == 'a' && ev.Mod == term.ModCtrl:
		// Ctrl+A: beginning of line
		ib.cursor = 0
		ib.selected = false
		return false, true

	case ev.Ch == 'e' && ev.Mod == term.ModCtrl:
		// Ctrl+E: end of line
		ib.cursor = len(ib.text)
		ib.selected = false
		return false, true

	case ev.Ch == 'b' && ev.Mod == term.ModCtrl:
		// Ctrl+B: backward char
		ib.selected = false
		if ib.cursor > 0 {
			ib.cursor--
		}
		return false, true

	case ev.Ch == 'f' && ev.Mod == term.ModCtrl:
		// Ctrl+F: forward char
		ib.selected = false
		if ib.cursor < len(ib.text) {
			ib.cursor++
		}
		return false, true

	case ev.Ch == 'd' && ev.Mod == term.ModCtrl:
		// Ctrl+D: delete char at cursor
		if ib.selected {
			ib.Clear()
		} else if ib.cursor < len(ib.text) {
			ib.text = append(
				ib.text[:ib.cursor],
				ib.text[ib.cursor+1:]...,
			)
		}
		return false, true

	case ev.Ch == 'h' && ev.Mod == term.ModCtrl:
		// Ctrl+H: backspace
		if ib.selected {
			ib.Clear()
		} else if ib.cursor > 0 {
			ib.text = append(
				ib.text[:ib.cursor-1],
				ib.text[ib.cursor:]...,
			)
			ib.cursor--
		}
		return false, true

	case ev.Ch == 'k' && ev.Mod == term.ModCtrl:
		// Ctrl+K: kill to end of line
		if ib.cursor < len(ib.text) {
			ib.text = ib.text[:ib.cursor]
		}
		return false, true

	case ev.Ch == 'u' && ev.Mod == term.ModCtrl:
		// Ctrl+U: kill entire line
		ib.Clear()
		return false, true

	case ev.Key == term.KeySpace:
		ev.Ch = ' '
		fallthrough

	case ev.Ch != 0 && ev.Ch >= 32: // printable character
		ib.selected = false
		// Insert character at cursor
		ib.text = append(ib.text[:ib.cursor], append([]rune{ev.Ch}, ib.text[ib.cursor:]...)...)
		ib.cursor++
		return false, true
	}

	return false, false
}

// Cursor satisfies tui.Handler
func (ib *InputBox) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	cursorLine := ib.cursor / ib.width
	cursorCol := ib.cursor % ib.width
	cursorY := cursorLine - ib.voffset
	return term.Coordinates{X: cursorCol, Y: cursorY},
		term.CursorStyleSteadyBlock, true
}

// Selection satisfies tui.Handler
func (ib *InputBox) Selection() (string, bool) {
	if ib.selected && len(ib.text) > 0 {
		return string(ib.text), true
	}
	return "", false
}
