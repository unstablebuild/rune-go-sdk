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

package term

import (
	"context"
)

type (
	// InputMode is the keyboard input mode.
	InputMode int
	// EventType is the type of event being dispatched.
	EventType uint8
	// Modifier is a key modifier like <alt> or <ctrl>.
	Modifier uint8
	// Key is a keyboard key.
	Key uint16
)

// Attributes represents a cell background, foreground and attribute
// bitmask. It is layout-compatible with term.Style; the duplicated
// type exists as the style currency exchanged with Cell.
type Attributes struct {
	Fg    Color
	Bg    Color
	Attrs AttrMask
	// Underline colours an AttrUnderline stroke; when not valid the
	// stroke takes Fg.
	Underline Color
}

// Style returns attr as a term.Style. The values are layout-identical;
// the helper exists for callers that prefer the Style nominal type at
// the renderer boundary.
func (attr Attributes) Style() Style {
	return Style(attr)
}

// CellExtra holds the parts of a Cell that are rarely present. A cell
// keeps them behind one pointer so the common cell stays 24 bytes.
// Cells are copied by value and share their CellExtra, so it must be
// replaced rather than mutated; the Cell setters do that.
type CellExtra struct {
	// Combining holds the remaining grapheme-cluster codepoints that
	// did not fit in Ch.
	Combining []rune
	// Underline colours the cell's underline; when not valid the
	// underline takes Fg.
	Underline Color
}

// Cell represents a location with content on a terminal screen.
// 'Ch' is a unicode character, 'Fg' and 'Bg' are foreground and
// background attributes respectively. Unicode grapheme clusters whose
// codepoints do not fit in a single rune are stored across Ch and
// Extra.Combining.
//
// The field order is chosen so the struct packs into exactly 24 bytes
// with no padding: the pointer leads (8-aligned), the two-byte and
// one-byte fields trail. Style fields are inlined rather than embedding
// Attributes because the embedded struct's internal padding would grow
// Cell to 32 bytes.
type Cell struct {
	// Extra holds the combining marks and the underline colour. nil
	// when the cell has neither (the common case).
	Extra *CellExtra
	// Fg is the foreground color.
	Fg Color
	// Bg is the background color.
	Bg Color
	// Ch is the main character held by this cell.
	// If character cannot fit in the storage provided by the
	// builtin 'rune', then Width() returns > 1 and CombiningRunes
	// returns the rest of data.
	Ch rune
	// Attrs is the text-rendering attribute bitmask.
	Attrs AttrMask
	// Width returns the monospace width of this Cell.
	Width uint8
	// Bytes is the number of bytes consumed by this Cell.
	Bytes uint8
}

// Attributes returns this cell's style as an Attributes value.
func (c Cell) Attributes() Attributes {
	return Attributes{Fg: c.Fg, Bg: c.Bg, Attrs: c.Attrs, Underline: c.UnderlineColor()}
}

// NewCell returns a Cell holding ch with the given monospace width and
// style attr.
func NewCell(ch rune, width uint8, attr Attributes) Cell {
	cell := Cell{
		Ch:    ch,
		Width: width,
		Fg:    attr.Fg,
		Bg:    attr.Bg,
		Attrs: attr.Attrs,
	}
	if attr.Underline.Valid() {
		cell.Extra = &CellExtra{Underline: attr.Underline}
	}
	return cell
}

// SetAttributes replaces this cell's style fields with attr.
func (c *Cell) SetAttributes(attr Attributes) {
	c.Fg = attr.Fg
	c.Bg = attr.Bg
	c.Attrs = attr.Attrs
	c.SetUnderlineColor(attr.Underline)
}

// Style returns this cell's style as a term.Style.
func (c Cell) Style() Style {
	return Style{Fg: c.Fg, Bg: c.Bg, Attrs: c.Attrs, Underline: c.UnderlineColor()}
}

// CombiningRunes returns the combining-mark slice, or nil when the
// cell has no combining marks. Callers must not mutate the returned
// slice in place; allocate a new slice and assign via SetCombining.
func (c Cell) CombiningRunes() []rune {
	if c.Extra == nil {
		return nil
	}
	return c.Extra.Combining
}

// SetCombining replaces this cell's combining-mark slice. Pass nil to
// drop combining marks.
func (c *Cell) SetCombining(runes []rune) {
	c.setExtra(runes, c.UnderlineColor())
}

// UnderlineColor returns the cell's underline colour, ColorDefault
// when the underline takes Fg.
func (c Cell) UnderlineColor() Color {
	if c.Extra == nil {
		return ColorDefault
	}
	return c.Extra.Underline
}

// SetUnderlineColor replaces this cell's underline colour. Pass
// ColorDefault to have the underline take Fg.
func (c *Cell) SetUnderlineColor(color Color) {
	c.setExtra(c.CombiningRunes(), color)
}

func (c *Cell) setExtra(runes []rune, underline Color) {
	if runes == nil && !underline.Valid() {
		c.Extra = nil
		return
	}
	if c.Extra != nil && c.Extra.Underline == underline &&
		len(c.Extra.Combining) == len(runes) && (len(runes) == 0 || &c.Extra.Combining[0] == &runes[0]) {
		return
	}
	c.Extra = &CellExtra{Combining: runes, Underline: underline}
}

// Event represents a terminal event. The 'Mod', 'Key' and 'Ch' fields are
// valid if 'Type' is EventKey. The 'Width' and 'Height' fields are valid if
// 'Type' is EventResize. The 'Err' field is valid if 'Type' is EventError.
type Event struct {
	Type     EventType // one of Event* constants
	Mod      Modifier  // one of Mod* constants or 0
	Key      Key       // one of Key* constants, invalid if 'Ch' is not 0
	Ch       rune      // a unicode character
	Width    int       // width of the screen
	Height   int       // height of the screen
	Err      error     // error in case if input failed
	MouseX   int       // x coord of mouse
	MouseY   int       // y coord of mouse
	Raw      []byte
	UserFunc func()
	Context  context.Context
}

// KeyComb returns this event as a KeyComb, or panics
// if this event is not of type EventKey.
func (e Event) KeyComb() KeyComb {
	if e.Type != EventKey {
		panic("KeyComb called on non-key event")
	}
	return KeyComb{Ch: e.Ch, Mod: e.Mod, Key: e.Key}
}

// KeyComb represents is a key combination. See event for more details.
type KeyComb struct {
	Mod Modifier
	Key Key
	Ch  rune
}

// Writer abstracts termbox write functionality to decouple components from
// termbox, so they're easier to test.
type Writer interface {
	// Context returns the current context of the Writer.
	// This context can be used by tui.Components in combination
	// with term.Interrupter.Interrupt(context.Context) to disambiguate
	// regular calls to Draw from interrupt-driven calls to Draw.
	Context() context.Context
	// SetCell sets the contents of the given cell location.  If
	// the coordinates are out of range, then the operation is ignored.
	SetCell(Coordinates, Cell)
	// UnionAttributes computes the set union between a and b,
	// that is overrides a set of attributes at the given coordinates
	// that contain all the bit flags set in a, b or both, and uses the color
	// defined in b or if not set, uses the color in a.
	UnionAttributes(Coordinates, Attributes)
	// DrawImage places img over its cell rectangle, composited with
	// the cells written there as img.Layer selects, and reports whether
	// this writer renders graphics at all; on false the caller should
	// draw a cell-based fallback.
	// Placements are discarded on the writer's next Clear, so an image
	// is re-placed on every Draw like any other content. Fully clipped
	// placements are dropped and still report true.
	DrawImage(Image) bool
}
