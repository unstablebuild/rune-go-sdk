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

// InputBox implements a simple single-line text input handler suitable
// for chat-like TUI interfaces. It supports basic text editing operations
// including character input, backspace, delete, cursor movement, and text selection.
type InputBox struct {
	text     []rune // the input text
	cursor   int    // cursor position (index in text)
	width    int    // available width for rendering
	height   int    // available height for rendering
	attrs    term.Attributes
	selected bool // whether all text is selected (Ctrl+A)
}

// Compile-time assertion that InputBox implements tui.Handler
var _ tui.Handler = (*InputBox)(nil)

// NewInputBox creates and initializes a new InputBox with default attributes.
func NewInputBox() *InputBox {
	return &InputBox{
		text:   make([]rune, 0, 64),
		cursor: 0,
		attrs:  term.Attributes{},
	}
}

// NewInputBoxWithAttrs creates a new InputBox with custom attributes.
func NewInputBoxWithAttrs(attrs term.Attributes) *InputBox {
	return &InputBox{
		text:   make([]rune, 0, 64),
		cursor: 0,
		attrs:  attrs,
	}
}

// Text returns the current text content of the input box.
func (ib *InputBox) Text() string {
	return string(ib.text)
}

// SetText sets the text content and moves cursor to the end.
func (ib *InputBox) SetText(s string) {
	ib.text = []rune(s)
	ib.cursor = len(ib.text)
	ib.selected = false
}

// Clear resets the input box to empty state.
func (ib *InputBox) Clear() {
	ib.text = ib.text[:0]
	ib.cursor = 0
	ib.selected = false
}

// Resize satisfies tui.Component
func (ib *InputBox) Resize(width, height int) {
	ib.width = width
	ib.height = height
}

// Draw satisfies tui.Component
func (ib *InputBox) Draw(w term.Writer) {
	if ib.height == 0 {
		return
	}

	// Draw the text
	for i, ch := range ib.text {
		if i >= ib.width {
			break
		}
		attrs := ib.attrs
		// Highlight selected text
		if ib.selected {
			attrs.Attrs |= tcell.AttrReverse
		}
		w.SetCell(term.Coordinates{X: i, Y: 0}, term.Cell{
			Ch:         ch,
			Attributes: attrs,
			Width:      1,
		})
	}

	// Fill remaining space with blanks
	for i := len(ib.text); i < ib.width; i++ {
		w.SetCell(term.Coordinates{X: i, Y: 0}, term.Cell{
			Ch:         ' ',
			Attributes: ib.attrs,
			Width:      1,
		})
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

	switch {
	case ev.Key == term.KeyEnter:
		return false, true

	case ev.Key == term.KeyEsc:
		return true, true

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

	case ev.Ch == 'a' && ev.Mod == term.ModCtrl:
		// Ctrl+A: select all
		ib.selected = true
		return false, true

	case ev.Ch == 'u' && ev.Mod == term.ModCtrl:
		// Ctrl+U: clear line (common in terminals)
		ib.Clear()
		return false, true

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
	// Show cursor at current position
	x := ib.cursor
	if x >= ib.width {
		x = ib.width - 1
	}
	return term.Coordinates{X: x, Y: 0}, term.CursorStyleSteadyBlock, true
}

// Selection satisfies tui.Handler
func (ib *InputBox) Selection() (string, bool) {
	if ib.selected && len(ib.text) > 0 {
		return string(ib.text), true
	}
	return "", false
}
