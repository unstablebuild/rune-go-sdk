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

package handlerrpc

import (
	"context"

	"github.com/unstablebuild/rune-go-sdk/term"
	termrpc "github.com/unstablebuild/rune-go-sdk/term/termrpc"
	"github.com/unstablebuild/rune-go-sdk/tui"
	"github.com/unstablebuild/tcell/v3"
)

var zeroCell = termrpc.Cell{}

var _ term.Writer = drawResponseWriter{}

// NewDrawResponse converts a tui.Component into a DrawResponse.
func NewDrawResponse(
	ctx context.Context, comp tui.Component, width, height int,
) *DrawStreamResponse {
	resp := &DrawStreamResponse{
		Cursor: &DrawStreamResponse_Cursor{
			Position: &termrpc.Coordinates{},
		},
	}
	w := newDrawResponseWriter(ctx, width, height)
	comp.Draw(w)
	resp.Rows = w.rows
	return resp
}

type drawResponseWriter struct {
	width, height int
	rows          []*termrpc.CellRow
	ctx           context.Context
}

// SetCell satisfies term.Writer
func (r drawResponseWriter) SetCell(pos term.Coordinates, c term.Cell) {
	if pos.Y >= r.height || pos.X >= r.width || pos.X < 0 || pos.Y < 0 {
		return
	}

	var cell *termrpc.Cell
	if r.rows[pos.Y].Cells[pos.X] == &zeroCell {
		cell = new(termrpc.Cell)
	} else {
		cell = r.rows[pos.Y].Cells[pos.X]
	}
	cell.Character = uint32(c.Ch)
	cell.Foreground = uint64(c.Fg)
	cell.Background = uint64(c.Bg)
	cell.Attrs = int64(c.Attrs)
	cell.Width = uint32(c.Width)
	cell.Bytes = uint32(c.Bytes)
	for _, c := range c.Combining {
		cell.Combining = append(cell.Combining, uint32(c))
	}

	r.rows[pos.Y].Cells[pos.X] = cell
}

func (r drawResponseWriter) UnionAttributes(pos term.Coordinates, attr term.Attributes) {
	if pos.Y >= r.height || pos.X >= r.width || pos.X < 0 || pos.Y < 0 {
		return
	}

	var cell *termrpc.Cell
	if r.rows[pos.Y].Cells[pos.X] == &zeroCell {
		cell = new(termrpc.Cell)
	} else {
		cell = r.rows[pos.Y].Cells[pos.X]
	}

	uattr := term.AttributesUnion(term.Attributes{
		Fg:    tcell.Color(cell.Foreground),
		Bg:    tcell.Color(cell.Background),
		Attrs: tcell.AttrMask(cell.Attrs),
	}, attr)

	cell.Foreground = uint64(uattr.Fg)
	cell.Background = uint64(uattr.Bg)
	cell.Attrs = int64(uattr.Attrs)

	r.rows[pos.Y].Cells[pos.X] = cell
}

// Flush satisfies term.Writer
func (r drawResponseWriter) Flush() error {
	return nil
}

// Clear satisfies term.Writer
func (r drawResponseWriter) Clear(term.Attributes) error {
	return nil
}

// SetCursor satisfies term.Writer
func (r drawResponseWriter) SetCursor(term.Coordinates) {
}

func (r drawResponseWriter) Context() context.Context {
	return r.ctx
}

func newDrawResponseWriter(ctx context.Context, width, height int) drawResponseWriter {
	cellRowSlab := make([]termrpc.CellRow, height)
	cellRowWidthSlab := make([]*termrpc.Cell, height*width)
	rows := make([]*termrpc.CellRow, height)
	for i := 0; i < height; i++ {
		rows[i] = &cellRowSlab[i]
		rows[i].Cells = cellRowWidthSlab[i*width : (i+1)*width]
		for j := 0; j < width; j++ {
			// SetCell substitutes zeroCell for a newly allocated cell;
			// this allows us to speed up client/server communication
			rows[i].Cells[j] = &zeroCell
		}
	}
	return drawResponseWriter{
		ctx:    ctx,
		width:  width,
		height: height,
		rows:   rows,
	}
}
