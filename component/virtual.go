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

// Virtual wraps a component to
// provide virtual coordinates and write bound checking.
// It exposes Move which can be used to move the inner component
// in the virtual coordinate space.
type Virtual[T tui.Component] struct {
	C  T
	vw VirtualWriter
	w  term.Writer // prebox, save allocs
}

// Resize resizes the underlying component and stores size
// to perform bound checking on Draw.
func (c *Virtual[T]) Resize(width, height int) {
	c.C.Resize(width, height)
	c.vw.Height = height
	c.vw.Width = width
}

// Draw uses a virtual writer to perform bound checking and
// if successful draw the inner component in the virtual coordinate space.
func (c *Virtual[T]) Draw(writer term.Writer) {
	if c.w == nil {
		c.w = &c.vw
	}
	c.vw.Writer = writer
	c.C.Draw(c.w)
}

// Move changes the position of this virtual component
// in the virtual coordinate space.
func (c *Virtual[T]) Move(pos term.Coordinates) {
	c.vw.Offset = pos
}

// Width returns the width set in last Resize.
func (c *Virtual[T]) Width() int {
	return c.vw.Width
}

// Height returns the height set in last Resize.
func (c *Virtual[T]) Height() int {
	return c.vw.Height
}

// Position returns this virtual component's position
// in the virtual coordinate space.
func (c *Virtual[T]) Position() term.Coordinates {
	return c.vw.Offset
}

var _ term.Writer = (*VirtualWriter)(nil)

// VirtualWriter wraps the given w with a writer that applies an
// offset and SetCell clipping according to offset, height and width.
type VirtualWriter struct {
	Writer        term.Writer
	Offset        term.Coordinates
	Height, Width int
}

// SetCell satisfies term.Writer.
func (w *VirtualWriter) SetCell(pos term.Coordinates, c term.Cell) {
	if pos.X >= w.Width || pos.Y >= w.Height || pos.Y < 0 || pos.X < 0 {
		return
	}
	pos = term.Coordinates{X: w.Offset.X + pos.X, Y: w.Offset.Y + pos.Y}
	w.Writer.SetCell(pos, c)
}

// UnionAttributes satisfies term.Writer.
func (w *VirtualWriter) UnionAttributes(pos term.Coordinates, attr term.Attributes) {
	if pos.X >= w.Width || pos.Y >= w.Height || pos.Y < 0 || pos.X < 0 {
		return
	}
	pos = term.Coordinates{X: w.Offset.X + pos.X, Y: w.Offset.Y + pos.Y}
	w.Writer.UnionAttributes(pos, attr)
}

// Context satisfies term.Writer.
func (w *VirtualWriter) Context() context.Context {
	return w.Writer.Context()
}
