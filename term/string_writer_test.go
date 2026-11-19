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
	"testing"

	"github.com/stretchr/testify/assert"
)

// this test is not very thorough because StringWriter is used by
// many tui.Component tests and so it's already indirectly tested.
func TestStringWriter(t *testing.T) {
	width, height := 5, 6
	writer := NewStringWriter(width, height)

	c := 'A'
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			if i > j-1 {
				writer.SetCell(Coordinates{X: j, Y: i}, Cell{Ch: c})
			}
		}
		c++
	}

	// we use StringWriter mostly for tests so it's important that it panics on oob
	assert.Panics(t, func() {
		writer.SetCell(Coordinates{X: width + 1, Y: height + 1}, Cell{Ch: '='})
	})

	if err := writer.Flush(); err != nil {
		t.Fatal(err)
	}

	expected := "A    \nBB   \nCCC  \nDDDD \nEEEEE\n     "
	assert.Equal(t, expected, writer.String())
}

func TestStringWriterSetCellKeepsBackground(t *testing.T) {
	teal := NewRGBColor(0, 128, 128)
	plum := NewRGBColor(128, 0, 128)
	white := NewRGBColor(255, 255, 255)
	pos := Coordinates{X: 1, Y: 1}
	keep := NewCell('x', 1, Attributes{Fg: white, Bg: plum, Attrs: AttrBold | AttrKeepBackground})

	for _, tc := range []struct {
		name     string
		existing Cell
		expected Cell
	}{
		{
			name:     "coloured background underneath",
			existing: NewCell('C', 1, Attributes{Fg: plum, Bg: teal, Attrs: AttrItalic}),
			expected: NewCell('x', 1, Attributes{Fg: white, Bg: teal, Attrs: AttrBold}),
		},
		{
			name:     "default background underneath",
			existing: NewCell('C', 1, Attributes{Fg: plum}),
			expected: NewCell('x', 1, Attributes{Fg: white, Attrs: AttrBold}),
		},
		{
			name:     "blank cell underneath",
			expected: NewCell('x', 1, Attributes{Fg: white, Attrs: AttrBold}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writer := NewStringWriter(3, 3)
			writer.SetCell(pos, tc.existing)
			writer.SetCell(pos, keep)
			assert.Equal(t, tc.expected, writer.Cells()[pos.Y*3+pos.X])
		})
	}

	t.Run("union keeps its own background semantics", func(t *testing.T) {
		writer := NewStringWriter(3, 3)
		writer.SetCell(pos, NewCell('C', 1, Attributes{Bg: teal}))
		writer.UnionAttributes(pos, Attributes{Fg: white, Attrs: AttrKeepBackground})
		assert.Equal(t, NewCell('C', 1, Attributes{Fg: white, Bg: teal}), writer.Cells()[pos.Y*3+pos.X])
	})
}
