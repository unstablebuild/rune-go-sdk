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
	"reflect"
	"testing"

	"github.com/unstablebuild/rune-go-sdk/term"
	"google.golang.org/protobuf/proto"
)

// recordWriter is a term.Writer that keeps the last cell written to every
// position, so a decoded frame can be compared cell by cell.
type recordWriter struct {
	width, height int
	cells         []term.Cell
}

func newRecordWriter(width, height int) *recordWriter {
	return &recordWriter{
		width:  width,
		height: height,
		cells:  make([]term.Cell, width*height),
	}
}

func (r *recordWriter) SetCell(pos term.Coordinates, c term.Cell) {
	if pos.X < 0 || pos.Y < 0 || pos.X >= r.width || pos.Y >= r.height {
		return
	}
	r.cells[pos.Y*r.width+pos.X] = c
}

func (r *recordWriter) UnionAttributes(pos term.Coordinates, attr term.Attributes) {
	if pos.X < 0 || pos.Y < 0 || pos.X >= r.width || pos.Y >= r.height {
		return
	}
	cell := &r.cells[pos.Y*r.width+pos.X]
	cell.SetAttributes(term.AttributesUnion(cell.Attributes(), attr))
}

func (r *recordWriter) Context() context.Context { return context.Background() }

func (r *recordWriter) DrawImage(term.Image) bool { return false }

// drawFunc adapts a draw callback into a tui.Component.
type drawFunc func(term.Writer)

func (d drawFunc) Resize(int, int)    {}
func (d drawFunc) Draw(w term.Writer) { d(w) }

// drawFrames renders draw into both wire formats and returns the decoded
// grids, after a marshal/unmarshal round trip through the wire.
func drawFrames(
	t *testing.T, width, height int, draw func(term.Writer),
) (rows, packed *recordWriter) {
	t.Helper()
	ctx := context.Background()

	decode := func(usePacked bool) *recordWriter {
		w := newDrawResponseWriter(ctx, width, height, usePacked)
		draw(w)
		var resp DrawStreamResponse
		w.fill(&resp)

		wire, err := proto.Marshal(&resp)
		if err != nil {
			t.Fatalf("marshal draw response (packed=%v): %v", usePacked, err)
		}
		var decoded DrawStreamResponse
		if err := proto.Unmarshal(wire, &decoded); err != nil {
			t.Fatalf("unmarshal draw response (packed=%v): %v", usePacked, err)
		}

		out := newRecordWriter(width, height)
		if usePacked {
			if decoded.GetPacked() == nil {
				t.Fatal("packed writer did not emit a packed frame")
			}
			if decoded.GetRows() != nil {
				t.Fatal("packed writer also emitted legacy rows")
			}
			decoded.GetPacked().WriteTo(out)
			return out
		}
		if decoded.GetPacked() != nil {
			t.Fatal("rows writer emitted a packed frame")
		}
		for y, row := range decoded.GetRows() {
			for x, c := range row.GetCells() {
				out.SetCell(term.Coordinates{X: x, Y: y}, c.ToModel())
			}
		}
		return out
	}

	return decode(false), decode(true)
}

func TestPackedFrameMatchesRows(t *testing.T) {
	const width, height = 7, 4

	tests := []struct {
		name string
		draw func(term.Writer)
	}{{
		name: "empty frame",
		draw: func(term.Writer) {},
	}, {
		name: "sparse cells leave the rest zeroed",
		draw: func(w term.Writer) {
			w.SetCell(term.Coordinates{X: 0, Y: 0}, term.Cell{Ch: 'a', Width: 1, Bytes: 1})
			w.SetCell(term.Coordinates{X: 6, Y: 3}, term.Cell{Ch: 'z', Width: 1, Bytes: 1})
		},
	}, {
		name: "colors and attributes",
		draw: func(w term.Writer) {
			w.SetCell(term.Coordinates{X: 2, Y: 1}, term.Cell{
				Ch:    'A',
				Fg:    term.Color(0x00ff00),
				Bg:    term.Color(0x0000ff),
				Attrs: term.AttrBold | term.AttrUnderline,
				Width: 1,
				Bytes: 1,
			})
		},
	}, {
		name: "wide glyph",
		draw: func(w term.Writer) {
			w.SetCell(term.Coordinates{X: 1, Y: 1}, term.Cell{Ch: '世', Width: 2, Bytes: 3})
			w.SetCell(term.Coordinates{X: 2, Y: 1}, term.Cell{Ch: 0, Width: 0, Bytes: 0})
		},
	}, {
		name: "combining runes",
		draw: func(w term.Writer) {
			cell := term.Cell{Ch: 'e', Width: 1, Bytes: 3}
			cell.SetCombining([]rune{0x0301})
			w.SetCell(term.Coordinates{X: 3, Y: 2}, cell)

			emoji := term.Cell{Ch: 0x1f469, Width: 2, Bytes: 11}
			emoji.SetCombining([]rune{0x200d, 0x1f4bb})
			w.SetCell(term.Coordinates{X: 4, Y: 2}, emoji)
		},
	}, {
		name: "combining runes overwritten by a plain cell",
		draw: func(w term.Writer) {
			cell := term.Cell{Ch: 'e', Width: 1, Bytes: 3}
			cell.SetCombining([]rune{0x0301})
			w.SetCell(term.Coordinates{X: 3, Y: 2}, cell)
			w.SetCell(term.Coordinates{X: 3, Y: 2}, term.Cell{Ch: 'x', Width: 1, Bytes: 1})
		},
	}, {
		name: "combining runes replaced",
		draw: func(w term.Writer) {
			cell := term.Cell{Ch: 'e', Width: 1, Bytes: 3}
			cell.SetCombining([]rune{0x0301})
			w.SetCell(term.Coordinates{X: 3, Y: 2}, cell)

			other := term.Cell{Ch: 'a', Width: 1, Bytes: 4}
			other.SetCombining([]rune{0x0308, 0x0301})
			w.SetCell(term.Coordinates{X: 3, Y: 2}, other)
		},
	}, {
		name: "union attributes over an existing cell",
		draw: func(w term.Writer) {
			w.SetCell(term.Coordinates{X: 2, Y: 2}, term.Cell{
				Ch:    'u',
				Fg:    term.Color(0x00ff00),
				Attrs: term.AttrBold,
				Width: 1,
				Bytes: 1,
			})
			w.UnionAttributes(term.Coordinates{X: 2, Y: 2}, term.Attributes{
				Bg:    term.Color(0xff0000),
				Attrs: term.AttrUnderline,
			})
		},
	}, {
		name: "union attributes over an untouched cell",
		draw: func(w term.Writer) {
			w.UnionAttributes(term.Coordinates{X: 5, Y: 0}, term.Attributes{
				Fg:    term.Color(0x123456),
				Attrs: term.AttrReverse,
			})
		},
	}, {
		name: "out of bounds writes are ignored",
		draw: func(w term.Writer) {
			w.SetCell(term.Coordinates{X: -1, Y: 0}, term.Cell{Ch: 'a'})
			w.SetCell(term.Coordinates{X: 0, Y: -1}, term.Cell{Ch: 'b'})
			w.SetCell(term.Coordinates{X: width, Y: 0}, term.Cell{Ch: 'c'})
			w.SetCell(term.Coordinates{X: 0, Y: height}, term.Cell{Ch: 'd'})
			w.UnionAttributes(term.Coordinates{X: width, Y: height}, term.Attributes{
				Attrs: term.AttrBold,
			})
			w.SetCell(term.Coordinates{X: 1, Y: 1}, term.Cell{Ch: 'e', Width: 1, Bytes: 1})
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, packed := drawFrames(t, width, height, tt.draw)
			for i := range rows.cells {
				if !reflect.DeepEqual(rows.cells[i], packed.cells[i]) {
					t.Errorf("cell (%d,%d): rows=%+v packed=%+v",
						i%width, i/width, rows.cells[i], packed.cells[i])
				}
			}
		})
	}
}

func TestNewDrawResponseFormat(t *testing.T) {
	comp := drawFunc(func(w term.Writer) {
		w.SetCell(term.Coordinates{X: 0, Y: 0}, term.Cell{Ch: 'a', Width: 1, Bytes: 1})
	})

	resp := NewDrawResponse(context.Background(), comp, 4, 2, false)
	if resp.GetPacked() != nil || len(resp.GetRows()) != 2 {
		t.Errorf("packed=false: got packed=%v rows=%d",
			resp.GetPacked() != nil, len(resp.GetRows()))
	}

	resp = NewDrawResponse(context.Background(), comp, 4, 2, true)
	if resp.GetPacked() == nil || len(resp.GetRows()) != 0 {
		t.Errorf("packed=true: got packed=%v rows=%d",
			resp.GetPacked() != nil, len(resp.GetRows()))
	}
	if got := resp.GetPacked().GetChars()[0]; got != uint32('a') {
		t.Errorf("packed chars[0] = %d, want %d", got, 'a')
	}
}
