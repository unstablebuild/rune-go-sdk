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
	"fmt"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/component/comptest"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestDrawBackgroundNoZero(t *testing.T) {
	u := &TestComponent{Ch: 'X'}
	s := NewSpan(u, SpanConfig{
		PadHorizontalPerc: 0.2,
		PadVerticalPerc:   0.2,
	})
	b := WithBackground(s, term.Cell{Ch: 'O'})

	b.Resize(9, 5)

	w := term.NewStringWriter(9, 5)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
XXXXXXXXO
XXXXXXXXO
XXXXXXXXO
XXXXXXXXO
OOOOOOOOO`,
		},
	}

	comptest.TestComponent(t, b, w, tests)
}

// imageComponent places img and records the context it is drawn with.
type imageComponent struct {
	img term.Image
	ok  bool
	ctx context.Context
}

func (c *imageComponent) Resize(int, int) {}

func (c *imageComponent) Draw(w term.Writer) {
	c.ctx = w.Context()
	c.ok = w.DrawImage(c.img)
}

// drawRecorder records what is drawn on it, in order.
type drawRecorder struct {
	term.NoopWriter
	drawn []string
}

func (w *drawRecorder) SetCell(pos term.Coordinates, c term.Cell) {
	w.drawn = append(w.drawn, fmt.Sprintf("%d,%d %q", pos.X, pos.Y, c.Ch))
}

func (w *drawRecorder) DrawImage(img term.Image) bool {
	w.drawn = append(w.drawn, fmt.Sprintf("image %d,%d", img.Pos.X, img.Pos.Y))
	return true
}

func (w *drawRecorder) Context() context.Context { return context.Background() }

// coveredImage draws a cell, then an image, then a cell over the image.
type coveredImage struct{}

func (coveredImage) Resize(int, int) {}

func (coveredImage) Draw(w term.Writer) {
	w.SetCell(term.Coordinates{}, term.Cell{Ch: 'A'})
	w.DrawImage(term.Image{Pos: term.Coordinates{X: 1}, Width: 1, Height: 1})
	w.SetCell(term.Coordinates{X: 1}, term.Cell{Ch: 'B'})
}

func TestDrawBackgroundKeepsTheDrawOrder(t *testing.T) {
	tests := []struct {
		name     string
		cell     term.Cell
		expected []string
	}{
		{
			name: "with a background character",
			cell: term.Cell{Ch: 'O'},
			expected: []string{
				"0,0 'O'", "1,0 'O'", "0,0 'A'", "image 1,0", "1,0 'B'",
			},
		},
		{
			name: "without a background character",
			expected: []string{
				`0,0 '\x00'`, `1,0 '\x00'`, "0,0 'A'", "image 1,0", "1,0 'B'",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := NewBackground(coveredImage{}, test.cell)
			b.Resize(2, 1)
			rec := &drawRecorder{}

			b.Draw(rec)

			assert.Equal(t, test.expected, rec.drawn,
				"the background is drawn first, then the component in its order")
		})
	}
}

func TestDrawBackgroundImages(t *testing.T) {
	const width, height = 4, 3
	tests := []struct {
		name        string
		cell        term.Cell
		img         term.Image
		graphic     bool
		expectedOK  bool
		expectedVis []image.Rectangle
	}{
		{
			name:        "places the image",
			cell:        term.Cell{Ch: 'O'},
			img:         term.Image{Pos: term.Coordinates{X: 1, Y: 1}, Width: 2, Height: 2},
			graphic:     true,
			expectedOK:  true,
			expectedVis: []image.Rectangle{image.Rect(1, 1, 3, 3)},
		},
		{
			name:        "clips the image to the background",
			cell:        term.Cell{Ch: 'O'},
			img:         term.Image{Pos: term.Coordinates{X: 2, Y: 1}, Width: 4, Height: 4},
			graphic:     true,
			expectedOK:  true,
			expectedVis: []image.Rectangle{image.Rect(2, 1, 4, 3)},
		},
		{
			name:       "drops an image outside the background",
			cell:       term.Cell{Ch: 'O'},
			img:        term.Image{Pos: term.Coordinates{X: 5, Y: 5}, Width: 2, Height: 2},
			graphic:    true,
			expectedOK: true,
		},
		{
			name:        "keeps an overflowing image whole",
			cell:        term.Cell{Ch: 'O'},
			img:         term.Image{Pos: term.Coordinates{X: 2, Y: 1}, Width: 4, Height: 4, Overflow: true},
			graphic:     true,
			expectedOK:  true,
			expectedVis: []image.Rectangle{image.Rect(2, 1, 6, 5)},
		},
		{
			name:        "reports a writer that draws no images",
			cell:        term.Cell{Ch: 'O'},
			img:         term.Image{Pos: term.Coordinates{X: 1, Y: 1}, Width: 2, Height: 2},
			expectedVis: []image.Rectangle{image.Rect(1, 1, 3, 3)},
		},
		{
			name:        "places the image without a background character",
			img:         term.Image{Pos: term.Coordinates{X: 1, Y: 1}, Width: 2, Height: 2},
			graphic:     true,
			expectedOK:  true,
			expectedVis: []image.Rectangle{image.Rect(1, 1, 3, 3)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content := &imageComponent{img: test.img}
			b := NewBackground(content, test.cell)
			b.Resize(width, height)
			rec := &imageRecorder{graphic: test.graphic}

			b.Draw(rec)

			require.NotNil(t, content.ctx, "the content sees the writer's context")
			assert.Equal(t, rec.Context(), content.ctx)
			assert.Equal(t, test.expectedOK, content.ok)
			var visible []image.Rectangle
			for _, img := range rec.images {
				visible = append(visible, img.Visible())
			}
			assert.Equal(t, test.expectedVis, visible)
		})
	}
}

func TestResizeBackground(t *testing.T) {
	t.Run("zero width does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			b := WithBackground(&TestComponent{}, term.Cell{})
			b.Resize(0, 1)
		})
	})
	t.Run("zero height does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			b := WithBackground(&TestComponent{}, term.Cell{})
			b.Resize(1, 0)
		})
	})
	t.Run("negative width does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			b := WithBackground(&TestComponent{}, term.Cell{})
			b.Resize(-1, 5)
		})
	})
	t.Run("negative height does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			b := WithBackground(&TestComponent{}, term.Cell{})
			b.Resize(5, -1)
		})
	})
	t.Run("both negative does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			b := WithBackground(&TestComponent{}, term.Cell{})
			b.Resize(-1, -1)
		})
	})
}

func TestResizeBackgroundBufferReuse(t *testing.T) {
	u := &TestComponent{Ch: 'X'}
	b := WithBackground(u, term.Cell{Ch: 'O'})

	tsuite := []struct {
		name             string
		width, height    int
		expectW, expectH int
	}{
		{"initial", 10, 5, 10, 5},
		{"shrink both", 5, 3, 5, 3},
		{"grow back", 10, 5, 10, 5},
		{"grow width only", 15, 5, 15, 5},
		{"grow height only", 15, 8, 15, 8},
		{"shrink to 1x1", 1, 1, 1, 1},
		{"back to normal", 10, 5, 10, 5},
	}

	for _, tcase := range tsuite {
		t.Run(tcase.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				b.Resize(tcase.width, tcase.height)
				w := term.NewStringWriter(
					tcase.expectW, tcase.expectH,
				)
				b.Draw(w)
			})
		})
	}
}

func BenchmarkBackgroundResize(b *testing.B) {
	bsuite := []struct {
		name          string
		width, height int
	}{
		{"80x24", 80, 24},
		{"200x50", 200, 50},
		{"400x100", 400, 100},
	}

	for _, bc := range bsuite {
		b.Run(bc.name+"/same_size", func(b *testing.B) {
			bg := WithBackground(
				&TestComponent{}, term.Cell{Ch: 'O'},
			)
			bg.Resize(bc.width, bc.height)
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				bg.Resize(bc.width, bc.height)
			}
		})
		b.Run(bc.name+"/shrink_grow", func(b *testing.B) {
			bg := WithBackground(
				&TestComponent{}, term.Cell{Ch: 'O'},
			)
			bg.Resize(bc.width, bc.height)
			half := bc.width / 2
			halfH := bc.height / 2
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				bg.Resize(half, halfH)
				bg.Resize(bc.width, bc.height)
			}
		})
	}
}

func TestDrawBackground(t *testing.T) {
	u := &TestComponent{Ch: 'X'}
	s := NewSpan(u, SpanConfig{
		PadHorizontalPerc: 0.2,
		PadVerticalPerc:   0.2,
	})
	b := WithBackground(s, term.Cell{Ch: 0})

	b.Resize(9, 5)

	w := term.NewStringWriter(9, 5)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
XXXXXXXX 
XXXXXXXX 
XXXXXXXX 
XXXXXXXX 
         `,
		},
	}

	comptest.TestComponent(t, b, w, tests)
}
