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
)

var zeroCell = termrpc.Cell{}

var _ term.Writer = drawResponseWriter{}

// NewDrawResponse converts a tui.Component into a DrawResponse. When packed
// is set, the frame is encoded columnar in DrawStreamResponse.Packed instead
// of the legacy per-cell DrawStreamResponse.Rows.
func NewDrawResponse(
	ctx context.Context, comp tui.Component, width, height int, packed bool,
) *DrawStreamResponse {
	resp := &DrawStreamResponse{
		Cursor: &DrawStreamResponse_Cursor{
			Position: &termrpc.Coordinates{},
		},
	}
	w := newDrawResponseWriter(ctx, width, height, packed)
	comp.Draw(w)
	w.fill(resp)
	return resp
}

type drawResponseWriter struct {
	width, height int
	rows          []*termrpc.CellRow
	ctx           context.Context
	// cellSlab is a pre-allocated contiguous block of width*height cells.
	// SetCell and UnionAttributes allocate from the slab via nextCell
	// instead of calling new(termrpc.Cell) per cell, reducing heap
	// allocations from O(cells written) to 1. The counter is a pointer
	// so that value-receiver methods (required by term.Writer) can
	// advance it.
	cellSlab []termrpc.Cell
	nextCell *int
	// packed is non-nil when the peer negotiated the columnar wire
	// format. It supersedes rows and cellSlab: cells are written into
	// flat planes, so a frame costs a fixed number of allocations on
	// both the encode and the decode side.
	packed *packedWriter
}

// packedWriter holds the mutable state of a columnar frame. It is
// referenced by pointer so that the value-receiver term.Writer methods
// can mutate it.
type packedWriter struct {
	cells *termrpc.PackedCells
	// combiningAt maps a cell index to its entry in cells.Combining. It
	// is allocated on the first cell carrying combining marks, which is
	// rare, so the common frame pays nothing for it.
	combiningAt map[uint32]int
}

// set records the combining marks of the cell at index i, overwriting any
// previously recorded marks for that cell.
func (p *packedWriter) setCombining(i uint32, runes []rune) {
	idx, ok := p.combiningAt[i]
	if !ok {
		if len(runes) == 0 {
			return
		}
		if p.combiningAt == nil {
			p.combiningAt = make(map[uint32]int)
		}
		idx = len(p.cells.Combining)
		p.combiningAt[i] = idx
		p.cells.Combining = append(p.cells.Combining,
			&termrpc.PackedCells_Combining{Index: i})
	}
	entry := p.cells.Combining[idx]
	entry.Runes = entry.Runes[:0]
	for _, r := range runes {
		entry.Runes = append(entry.Runes, uint32(r))
	}
}

// SetCell satisfies term.Writer
func (r drawResponseWriter) SetCell(pos term.Coordinates, c term.Cell) {
	if pos.Y >= r.height || pos.X >= r.width || pos.X < 0 || pos.Y < 0 {
		return
	}

	if p := r.packed; p != nil {
		i := pos.Y*r.width + pos.X
		p.cells.Chars[i] = uint32(c.Ch)
		p.cells.Fg[i] = uint32(c.Fg)
		p.cells.Bg[i] = uint32(c.Bg)
		p.cells.Attrs[i] = uint32(c.Attrs)
		p.cells.Widths[i] = uint32(c.Width)
		p.cells.Bytes[i] = uint32(c.Bytes)
		p.setCombining(uint32(i), c.CombiningRunes())
		return
	}

	var cell *termrpc.Cell
	if r.rows[pos.Y].Cells[pos.X] == &zeroCell {
		cell = &r.cellSlab[*r.nextCell]
		*r.nextCell++
	} else {
		cell = r.rows[pos.Y].Cells[pos.X]
	}
	cell.Character = uint32(c.Ch)
	cell.Foreground = uint32(c.Fg)
	cell.Background = uint32(c.Bg)
	cell.Attrs = uint32(c.Attrs)
	cell.Width = uint32(c.Width)
	cell.Bytes = uint32(c.Bytes)
	runes := c.CombiningRunes()
	combining := make([]uint32, 0, len(runes))
	for _, r := range runes {
		combining = append(combining, uint32(r))
	}
	cell.Combining = combining

	r.rows[pos.Y].Cells[pos.X] = cell
}

func (r drawResponseWriter) UnionAttributes(pos term.Coordinates, attr term.Attributes) {
	if pos.Y >= r.height || pos.X >= r.width || pos.X < 0 || pos.Y < 0 {
		return
	}

	if p := r.packed; p != nil {
		i := pos.Y*r.width + pos.X
		uattr := term.AttributesUnion(term.Attributes{
			Fg:    term.Color(p.cells.Fg[i]),
			Bg:    term.Color(p.cells.Bg[i]),
			Attrs: term.AttrMask(p.cells.Attrs[i]),
		}, attr)
		p.cells.Fg[i] = uint32(uattr.Fg)
		p.cells.Bg[i] = uint32(uattr.Bg)
		p.cells.Attrs[i] = uint32(uattr.Attrs)
		return
	}

	var cell *termrpc.Cell
	if r.rows[pos.Y].Cells[pos.X] == &zeroCell {
		cell = &r.cellSlab[*r.nextCell]
		*r.nextCell++
	} else {
		cell = r.rows[pos.Y].Cells[pos.X]
	}

	uattr := term.AttributesUnion(term.Attributes{
		Fg:    term.Color(cell.Foreground),
		Bg:    term.Color(cell.Background),
		Attrs: term.AttrMask(cell.Attrs),
	}, attr)

	cell.Foreground = uint32(uattr.Fg)
	cell.Background = uint32(uattr.Bg)
	cell.Attrs = uint32(uattr.Attrs)

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

// fill writes the drawn frame into resp, in the format this writer was
// created for.
func (r drawResponseWriter) fill(resp *DrawStreamResponse) {
	if r.packed != nil {
		resp.Packed = r.packed.cells
		return
	}
	resp.Rows = r.rows
}

func newDrawResponseWriter(
	ctx context.Context, width, height int, packed bool,
) drawResponseWriter {
	total := width * height
	if packed {
		// One slab backs every plane: the planes are only read by the
		// proto marshaller, so they can share storage.
		slab := make([]uint32, 6*total)
		return drawResponseWriter{
			ctx:    ctx,
			width:  width,
			height: height,
			packed: &packedWriter{
				cells: &termrpc.PackedCells{
					Width:  uint32(width),
					Height: uint32(height),
					Chars:  slab[0*total : 1*total],
					Fg:     slab[1*total : 2*total],
					Bg:     slab[2*total : 3*total],
					Attrs:  slab[3*total : 4*total],
					Widths: slab[4*total : 5*total],
					Bytes:  slab[5*total : 6*total],
				},
			},
		}
	}
	cellRowSlab := make([]termrpc.CellRow, height)
	cellRowWidthSlab := make([]*termrpc.Cell, total)
	cellSlab := make([]termrpc.Cell, total)
	rows := make([]*termrpc.CellRow, height)
	for i := 0; i < height; i++ {
		rows[i] = &cellRowSlab[i]
		rows[i].Cells = cellRowWidthSlab[i*width : (i+1)*width]
		for j := 0; j < width; j++ {
			// SetCell substitutes zeroCell for a newly allocated cell
			// from the pre-allocated slab; this avoids per-cell heap
			// allocations during Draw.
			rows[i].Cells[j] = &zeroCell
		}
	}
	var n int
	return drawResponseWriter{
		ctx:      ctx,
		width:    width,
		height:   height,
		rows:     rows,
		cellSlab: cellSlab,
		nextCell: &n,
	}
}
