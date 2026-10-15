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
	"context"
	"image"

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Background represents a background Component. It wraps a tui.Component
// to make sure that all cells are reset to cell, before comp is drawn.
//
// If cell.Ch is non-zero, the cells comp draws keep only their character
// and, when set, their foreground color, and take the rest from cell, so
// comp is always drawn on cell's background.
type Background struct {
	root          tui.Component
	width, height int
	cell          term.Cell
}

var _ Floating = (*Background)(nil)
var _ Responsive = (*Background)(nil)
var _ WithAttributes = (*Background)(nil)

// WithBackground is deprecated. Use NewBackground instead.
func WithBackground(comp tui.Component, cell term.Cell) *Background {
	return NewBackground(comp, cell)
}

// NewBackground allocates storage for a new Background and initializes it.
func NewBackground(comp tui.Component, cell term.Cell) *Background {
	ret := new(Background)
	ret.Init(comp, cell)
	return ret
}

// Init initializes this Background with comp and cell.
func (b *Background) Init(comp tui.Component, c term.Cell) {
	b.root = comp
	b.cell = c
}

// Resize satisfies tui.Component
func (b *Background) Resize(width, height int) {
	width = max(0, width)
	height = max(0, height)
	b.width = width
	b.height = height
	b.root.Resize(width, height)
}

// Draw satisfies tui.Component
func (b *Background) Draw(w term.Writer) {
	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			w.SetCell(term.Coordinates{X: x, Y: y}, b.cell)
		}
	}
	if b.cell.Ch == 0 {
		b.root.Draw(w)
		return
	}
	b.root.Draw(&bgWriter{Background: b, out: w})
}

// Content returns the inner component.
func (b *Background) Content() tui.Component {
	return b.root
}

// Dimensions satisfies Floating if underlying tui.Component
// satisfies Floating, or panics if it doesn't.
func (s *Background) Dimensions() (width, height int) {
	return s.root.(Floating).Dimensions()
}

// SetAttr satisfies WithAttributes if underlying tui.Component
// satisfies WithAttributes, or panics if it doesn't.
func (s *Background) SetAttr(attr term.Attributes) term.Attributes {
	s.cell.Bg = attr.Bg
	s.cell.Fg = attr.Fg
	s.cell.Attrs = attr.Attrs
	return s.root.(WithAttributes).SetAttr(attr)
}

// Height satisfies Responsive if underlying tui.Component
// satisfies Responsive, or panics if it doesn't.
func (s *Background) Height(width int) int {
	return s.root.(Responsive).Height(width)
}

// bgWriter draws a Background's component on out, the writer the
// Background draws on, merging its cells with the Background's.
type bgWriter struct {
	*Background
	out term.Writer
}

// SetCell satisfies term.Writer
func (w *bgWriter) SetCell(pos term.Coordinates, c term.Cell) {
	if w.outOfBounds(pos) {
		return
	}
	cell := w.cell
	if c.Ch != 0 {
		cell.Ch = c.Ch
		cell.SetCombining(c.CombiningRunes())
		cell.Width = c.Width
	}
	if c.Fg != 0 {
		cell.Fg = c.Fg
	}
	w.out.SetCell(pos, cell)
}

// UnionAttributes satisfies term.Writer. Only the foreground color is
// united, since the other attributes of the component's cells are the
// Background's.
func (w *bgWriter) UnionAttributes(pos term.Coordinates, attr term.Attributes) {
	if w.outOfBounds(pos) {
		return
	}
	w.out.UnionAttributes(pos, term.Attributes{Fg: attr.Fg})
}

func (w *bgWriter) Context() context.Context {
	return w.out.Context()
}

// DrawImage satisfies term.Writer
func (w *bgWriter) DrawImage(img term.Image) bool {
	if !img.Overflow {
		var ok bool
		if img, ok = img.Clipped(image.Rect(0, 0, w.width, w.height)); !ok {
			return true
		}
	}
	return w.out.DrawImage(img)
}

func (w *bgWriter) outOfBounds(pos term.Coordinates) bool {
	return pos.X >= w.width || pos.Y >= w.height || pos.X < 0 || pos.Y < 0
}
