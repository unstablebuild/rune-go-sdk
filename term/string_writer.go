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
	"bytes"
	"context"
	"fmt"
)

var _ Writer = (*StringWriter)(nil)

// StringWriter satisfies Writer by rendering the cells into a plain string.
type StringWriter struct {
	cellbuf       []Cell
	buffer        bytes.Buffer
	width, height int
	CursorCh      rune
	SetContext    context.Context
	BackgroundCh  rune
	ForegroundCh  rune
}

// NewStringWriter allocates storage for a new StringWriter and
// initializes it.
func NewStringWriter(width, height int) (t *StringWriter) {
	t = new(StringWriter)
	t.Init(width, height)
	return
}

// Init initializes this StringWriter with the given height and width.
func (w *StringWriter) Init(width, height int) {
	w.Resize(width, height)
	w.CursorCh = '▐'
	w.SetContext = context.Background()
}

// Context returns context.Background
func (w *StringWriter) Context() context.Context {
	return w.SetContext
}

// Resize satisfies Writer.
func (w *StringWriter) Resize(width, height int) {
	w.width, w.height = width, height
	w.cellbuf = make([]Cell, width*height)
}

// Reset resets this writer.
func (w *StringWriter) Reset() {
	w.Init(w.width, w.height)
	w.buffer.Reset()
}

func outOfBounds(height, width int, pos Coordinates) bool {
	return pos.X >= width || pos.Y >= height || pos.X < 0 || pos.Y < 0
}

// SetCell satisfies Writer.
func (w *StringWriter) SetCell(pos Coordinates, cell Cell) {
	if outOfBounds(w.height, w.width, pos) {
		panic(fmt.Sprintf("SetCell(x=%d;y=%d): out of bounds: width=%d;height=%d",
			pos.X, pos.Y, w.width, w.height))
	}
	idx := pos.Y*w.width + pos.X
	w.cellbuf[idx] = cell
}

// UnionAttributes satisfies Writer.
func (w *StringWriter) UnionAttributes(pos Coordinates, attr Attributes) {
	if outOfBounds(w.height, w.width, pos) {
		panic(fmt.Sprintf("SetCell(x=%d;y=%d): out of bounds: width=%d;height=%d",
			pos.X, pos.Y, w.width, w.height))
	}
	idx := pos.Y*w.width + pos.X
	w.cellbuf[idx].Attributes = AttributesUnion(w.cellbuf[idx].Attributes, attr)
}

// Flush flushes the contents of this writer into the underlying cell buffer.
func (w *StringWriter) Flush() (err error) {
	for i, c := range w.cellbuf {
		if i != 0 && i%w.width == 0 {
			w.buffer.WriteRune('\n')
		}
		ch := c.Ch
		switch ch {
		case '\t', '\n', 0:
			ch = ' '
		}
		if w.BackgroundCh != 0 && c.Bg != 0 {
			ch = w.BackgroundCh
		}
		if w.ForegroundCh != 0 && c.Fg != 0 {
			ch = w.ForegroundCh
		}
		w.buffer.WriteRune(ch)
	}
	return
}

// Cells returns the internal cell slice.
func (w *StringWriter) Cells() []Cell {
	return w.cellbuf
}

// Clear satisfies Writer. Note that attr are ignored as they
// can't be represented in a string.
func (w *StringWriter) Clear(attr Attributes) (err error) {
	w.cellbuf = make([]Cell, w.width*w.height)
	w.buffer.Reset()
	return
}

// String returns the string representation of the contents of this Writer.
func (w *StringWriter) String() string {
	return w.buffer.String()
}

// SetCursor satisfies Writer by substituting the rune
// at pos for a pre-defined cursor-like rune.
func (w *StringWriter) SetCursor(pos Coordinates) {
	i := pos.X + pos.Y*w.width
	if i < len(w.cellbuf) {
		w.cellbuf[i].Ch = w.CursorCh
	}
}
