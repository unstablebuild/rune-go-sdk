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

package handler

import (
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Grid is the handler counterpart to component.Grid: it lays handlers
// out in an equally sized grid, each []tui.Handler of its matrix a row,
// and routes events to them as Group does. Only the handlers that fit
// in it, as laid out by Resize, can take the focus.
type Grid struct {
	routed
	matrix        [][]tui.Handler
	width, height int
	cells         []Child
}

// NewGrid returns a Grid of matrix, whose nil handlers leave their cell
// empty.
func NewGrid(matrix [][]tui.Handler) *Grid {
	g := &Grid{matrix: matrix}
	g.layout = func() Layout {
		return Layout{Width: g.width, Height: g.height, Children: g.cells}
	}
	return g
}

// Resize satisfies tui.Component.
func (g *Grid) Resize(width, height int) {
	g.width, g.height = width, height
	g.cells = g.cells[:0]
	if len(g.matrix) == 0 {
		return
	}
	cellHeight := height / len(g.matrix)
	for y, row := range g.matrix {
		top := y * cellHeight
		if len(row) == 0 {
			continue
		}
		cellWidth := width / len(row)
		for x, h := range row {
			left := x * cellWidth
			if h == nil {
				continue
			}
			h.Resize(cellWidth, cellHeight)
			g.cells = append(g.cells, Child{
				Handler: h, Position: term.Coordinates{X: left, Y: top}, Width: cellWidth, Height: cellHeight,
				CaptureDrags: true,
			})
		}
	}
}

// Draw satisfies tui.Component.
func (g *Grid) Draw(w term.Writer) {
	for _, c := range g.cells {
		c.Draw(&component.VirtualWriter{Writer: w, Offset: c.Position, Height: c.Height, Width: c.Width})
	}
}
