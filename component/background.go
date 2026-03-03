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

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Background represents a background Component. It wraps a tui.Component
// to make sure that all cells are reset to cell, before comp is drawn.
//
// If cell.Ch is non-zero, then the Background overrides the underlying
// component's attributes in calls to SetCell such that cell.Bg and
// cell.Fg are always used.
type Background struct {
	root          tui.Component
	width, height int
	cell          term.Cell
	buffer        [][]term.Cell
	zero          []term.Cell
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
	b.resizeCells(width, height)
}

// Draw satisfies tui.Component
func (b *Background) Draw(w term.Writer) {
	if b.cell.Ch == 0 {
		for y := 0; y < b.height; y++ {
			for x := 0; x < b.width; x++ {
				w.SetCell(term.Coordinates{X: x, Y: y}, b.cell)
			}
		}
		b.root.Draw(w)
		return
	}

	writer := bgWriter{Background: b}

	b.zeroCells()
	b.root.Draw(&writer)
	b.draw(w, b.buffer)
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

func (b *Background) zeroCells() {
	for i := 0; i < b.height; i++ {
		copy(b.buffer[i], b.zero)
	}
}

func (b *Background) resizeCells(width, height int) {
	if cap(b.zero) >= width {
		b.zero = b.zero[:width]
		clear(b.zero)
	} else {
		b.zero = make([]term.Cell, width)
	}
	if cap(b.buffer) >= height {
		b.buffer = b.buffer[:height]
	} else {
		b.buffer = make([][]term.Cell, height)
	}
	for i := range height {
		if cap(b.buffer[i]) >= width {
			b.buffer[i] = b.buffer[i][:width]
		} else {
			b.buffer[i] = make([]term.Cell, width)
		}
	}
}

func (b *Background) draw(w term.Writer, cells [][]term.Cell) {
	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			cell := b.cell
			if y < len(cells) && x < len(cells[y]) {
				ocell := cells[y][x]
				if ocell.Ch != 0 {
					cell.Ch = ocell.Ch
					cell.Combining = ocell.Combining
					cell.Width = ocell.Width
				}
				if ocell.Fg != 0 {
					cell.Fg = ocell.Fg
				}
				// but no background
			}
			w.SetCell(term.Coordinates{X: x, Y: y}, cell)
		}
	}
}

// bgWriter satisfies term.Writer with a Buffer.
type bgWriter struct {
	*Background
	ctx context.Context
}

// SetCell satisfies term.Writer
func (w *bgWriter) SetCell(pos term.Coordinates, c term.Cell) {
	if pos.X >= w.width || pos.Y >= w.height || pos.X < 0 || pos.Y < 0 {
		return
	}

	w.buffer[pos.Y][pos.X] = c
}

// UnionAttributes satisfies term.Writer
func (w *bgWriter) UnionAttributes(pos term.Coordinates, attr term.Attributes) {
	if pos.X >= w.width || pos.Y >= w.height || pos.X < 0 || pos.Y < 0 {
		return
	}
	w.buffer[pos.Y][pos.X].Attributes = term.AttributesUnion(
		w.buffer[pos.Y][pos.X].Attributes, attr)
}

func (w *bgWriter) Context() context.Context {
	return w.ctx
}
