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

// Handler implements a multi-line text input handler with wrapping.
type Handler struct {
	text        []rune
	cursor      int
	voffset     int
	width       int
	height      int
	attrs       term.Attributes
	selected    bool
	placeholder *component.ResponsiveString
}

var _ handler.WithAttributesResponsive = (*Handler)(nil)

// New creates and initializes a new input box with default attributes.
func New(opts ...Option) *Handler {
	ret := &Handler{
		text:   make([]rune, 0, 64),
		cursor: 0,
		attrs:  term.Attributes{},
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
	ib.selected = false
	ib.updateVoffset()
}

// Resize satisfies tui.Handler
func (ib *Handler) Resize(width, height int) {
	ib.width = width
	ib.height = height
	if ib.placeholder != nil {
		ib.placeholder.Resize(width, height)
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

	if len(ib.text) == 0 && ib.placeholder != nil {
		ib.placeholder.Draw(w)
		return
	}

	startIdx := ib.voffset * ib.width
	endIdx := min(startIdx+ib.height*ib.width, len(ib.text))

	var x, y int
	for i := startIdx; i < endIdx; i++ {
		attrs := ib.attrs
		if ib.selected {
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

	// emacs-style cursor movement
	case ev.Ch == 'a' && ev.Mod == term.ModCtrl:
		// <ctrl-a> beginning of line
		ib.cursor = 0
		ib.selected = false
		return false, true

	case ev.Ch == 'e' && ev.Mod == term.ModCtrl:
		// <ctrl-e> end of line
		ib.cursor = len(ib.text)
		ib.selected = false
		return false, true

	case ev.Ch == 'b' && ev.Mod == term.ModCtrl:
		// <ctrl+b> backward char
		ib.selected = false
		if ib.cursor > 0 {
			ib.cursor--
		}
		return false, true

	case ev.Ch == 'f' && ev.Mod == term.ModCtrl:
		// <ctrl-f> forward char
		ib.selected = false
		if ib.cursor < len(ib.text) {
			ib.cursor++
		}
		return false, true

	case ev.Ch == 'd' && ev.Mod == term.ModCtrl:
		// <ctrl-d> delete char at cursor
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
		// <ctrl-h> backspace
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
		// <ctrl-k> kill to end of line
		if ib.cursor < len(ib.text) {
			ib.text = ib.text[:ib.cursor]
		}
		return false, true

	case ev.Ch == 'u' && ev.Mod == term.ModCtrl:
		// <ctrl+u> kill entire line
		ib.Clear()
		return false, true

	case ev.Key == term.KeySpace:
		ev.Ch = ' '
		fallthrough

	case ev.Ch != 0:
		ib.selected = false
		ib.text = append(ib.text[:ib.cursor], append([]rune{ev.Ch}, ib.text[ib.cursor:]...)...)
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
	if ib.selected && len(ib.text) > 0 {
		return string(ib.text), true
	}
	return "", false
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
