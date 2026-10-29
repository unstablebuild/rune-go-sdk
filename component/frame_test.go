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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unstablebuild/rune-go-sdk/component/comptest"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestDrawFrame(t *testing.T) {
	u := &TestComponent{Ch: 'X'}
	f := NewFrame(u)

	f.Resize(8, 4)

	w := term.NewStringWriter(9, 5)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
┌──────┐ 
│XXXXXX│ 
│XXXXXX│ 
└──────┘ 
         `,
		}, {
			Action: func() { f.SetContent(&TestComponent{Ch: '*'}) }, Expected: `
┌──────┐ 
│******│ 
│******│ 
└──────┘ 
         `,
		}, {
			Action: func() { f.Resize(4, 4) }, Expected: `
┌──┐     
│**│     
│**│     
└──┘     
         `,
		}, {
			Action: func() { f.SetContent(&TestComponent{Ch: 'T'}) }, Expected: `
┌──┐     
│TT│     
│TT│     
└──┘     
         `,
		}, {
			Action: func() { f.Resize(2, 2) }, Expected: `
TT       
TT       
         
         
         `,
		}, {
			Action: func() { f.Resize(8, 4) }, Expected: `
┌──────┐ 
│TTTTTT│ 
│TTTTTT│ 
└──────┘ 
         `,
		}, {
			Action: func() {
				fb := FrameCharSetDefault()
				fb.HorizontalTop = '┄'
				fb.VerticalRight = '┊'
				f.FrameCharSet = fb
			}, Expected: `
┌┄┄┄┄┄┄┐ 
│TTTTTT┊ 
│TTTTTT┊ 
└──────┘ 
         `,
		}, {
			Action: func() {
				f.FrameCharSet = FrameCharSetDefault()
				f.SetContent(NewString("123"))
				f.Resize(8, 1)
			}, Expected: `
123      
         
         
         
         `,
		},
	}

	comptest.TestComponent(t, f, w, tests)
}

func TestComponentDimensions(t *testing.T) {
	f := NewFrame(StaticFloating(&TestComponent{}, 2, 2))
	actualWidth, actualHeight := f.Dimensions()
	assert.Equal(t, 4, actualWidth)
	assert.Equal(t, 4, actualHeight)
}

func TestFrameCharSetGraphicsEmpty(t *testing.T) {
	tests := []struct {
		name string
		g    FrameCharSetGraphics
		want bool
	}{
		{"zero", FrameCharSetGraphics{}, true},
		{"overlay alone", FrameCharSetGraphics{Overlay: true}, true},
		{"one image", FrameCharSetGraphics{VerticalRight: "file:///r.png"}, false},
		{"overlay image", FrameCharSetGraphics{TopLeft: "file:///tl.png", Overlay: true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.g.Empty())
		})
	}
}

func TestScrollBarCharSetAt(t *testing.T) {
	full := ScrollBarCharSet{
		Top: 't', Middle: 'm', Bottom: 'b',
		Graphics: ScrollBarCharSetGraphics{Top: "T", Middle: "M", Bottom: "B"},
	}
	middle := ScrollBarCharSet{Middle: 'm', Graphics: ScrollBarCharSetGraphics{Middle: "M"}}
	tests := []struct {
		name      string
		cs        ScrollBarCharSet
		height    int
		want      string
		wantImage string
	}{
		{"every slot", full, 4, "tmmb", "TMMB"},
		{"two cells are both ends", full, 2, "tb", "TB"},
		{"one cell is the middle", full, 1, "m", "M"},
		{"ends default to the middle", middle, 3, "mmm", "MMM"},
		{
			"runes and images default apart",
			ScrollBarCharSet{
				Top: 't', Middle: 'm',
				Graphics: ScrollBarCharSetGraphics{Middle: "M", Bottom: "B"},
			},
			3, "tmm", "MMB",
		},
		{"nothing set", ScrollBarCharSet{}, 2, "\x00\x00", "__"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var runes, images []rune
			for i := range tt.height {
				ch, uri := tt.cs.At(i, tt.height)
				runes = append(runes, ch)
				if uri == "" {
					uri = "_"
				}
				images = append(images, []rune(uri)...)
			}
			assert.Equal(t, tt.want, string(runes))
			assert.Equal(t, tt.wantImage, string(images))
		})
	}
}

func TestFrameDrawScrollBarCharSet(t *testing.T) {
	tests := []struct {
		name     string
		cs       ScrollBarCharSet
		expected string
	}{
		{
			name: "no runes draw the right edge",
			expected: `
┌──┐
│XX│
│XX│
│XX│
│XX│
└──┘`,
		},
		{
			name: "every slot",
			cs:   ScrollBarCharSet{Top: '▗', Middle: '▐', Bottom: '▝'},
			expected: `
┌──┐
│XX▗
│XX▐
│XX▝
│XX│
└──┘`,
		},
		{
			name: "the ends default to the middle",
			cs:   ScrollBarCharSet{Middle: '▐'},
			expected: `
┌──┐
│XX▐
│XX▐
│XX▐
│XX│
└──┘`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFrame(scrollingComponent{TestComponent: &TestComponent{Ch: 'X'}})
			f.ScrollBarCharSet = tt.cs
			comptest.TestComponent(t, f, term.NewStringWriter(4, 6), []comptest.TestCase{
				{Action: func() { f.Resize(4, 6) }, Expected: tt.expected},
			})
		})
	}
}

// scrollingComponent has half as many rows past its bottom as a frame
// six rows high shows, so that the frame draws a scroll bar over the
// top three quarters of its right edge.
type scrollingComponent struct {
	*TestComponent
}

func (scrollingComponent) SeekUp() bool       { return false }
func (scrollingComponent) SeekDown() bool     { return false }
func (scrollingComponent) SeekOffset() int    { return 0 }
func (scrollingComponent) MaxSeekOffset() int { return 2 }

func TestFrameResponsive(t *testing.T) {
	t.Run("adds frame height to content's Height", func(t *testing.T) {
		f := NewFrame(testResponsive('a', 10))
		assert.Equal(t, 12, f.Height(10))
	})
	t.Run("it's conistent with Resize with width < 3 behaviour", func(t *testing.T) {
		f := NewFrame(testResponsive('a', 10))
		assert.Equal(t, 10, f.Height(2))
	})
	t.Run("it's conistent with Resize with height < 3 behaviour", func(t *testing.T) {
		f := NewFrame(testResponsive('a', 0))
		assert.Equal(t, 3, f.Height(10))
	})
}

func TestFrameWithAttributes(t *testing.T) {
	t.Run("component satisfies WithAttributes", func(t *testing.T) {
		f := NewFrame(&TestComponent{Ch: 'a'})
		f.SetAttr(term.Attributes{Fg: term.ColorRed, Bg: term.ColorGreen})

		assert.Equal(t, term.ColorGreen, f.Bg)
		assert.Equal(t, term.ColorRed, f.Fg)

		prev := f.Content().(WithAttributes).SetAttr(term.Attributes{})
		assert.Equal(t, term.ColorGreen, prev.Bg)
		assert.Equal(t, term.ColorRed, prev.Fg)
	})

	t.Run("component does not WithAttributes", func(t *testing.T) {
		f := NewFrame(&TestComponent{Ch: 'a'})
		f.SetAttr(term.Attributes{Fg: term.ColorRed, Bg: term.ColorGreen})

		assert.Equal(t, term.ColorGreen, f.Bg)
		assert.Equal(t, term.ColorRed, f.Fg)
	})
}

func TestFrameResizeNegative(t *testing.T) {
	t.Run("negative height does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			b := NewBackground(&TestComponent{}, term.Cell{})
			f := NewFrame(b)
			f.Resize(10, -1)
		})
	})
	t.Run("negative width does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			b := NewBackground(&TestComponent{}, term.Cell{})
			f := NewFrame(b)
			f.Resize(-1, 10)
		})
	})
}

func TestFrameScrollbarCalculate(t *testing.T) {
	suite := []struct {
		offset, maxOffset, height      int
		expectedOffset, expectedHeight int
	}{
		{
			offset:         0,
			maxOffset:      0,
			height:         10,
			expectedOffset: 1,
			expectedHeight: 8,
		},
		{
			offset:         10,
			maxOffset:      10,
			height:         12,
			expectedOffset: 6,
			expectedHeight: 5,
		},
		{
			offset:         10,
			maxOffset:      20,
			height:         12,
			expectedOffset: 4,
			expectedHeight: 4,
		},
		{
			offset:         500,
			maxOffset:      1000,
			height:         12,
			expectedOffset: 5,
			expectedHeight: 1,
		},
	}

	for i, test := range suite {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			actualOffset, actualHeight := calculateScrollBar(
				test.offset, test.maxOffset, test.height)
			assert.Equal(t, test.expectedOffset, actualOffset, "offset")
			assert.Equal(t, test.expectedHeight, actualHeight, "height")
		})
	}

	t.Run("doesn't panic if offset > maxOffset", func(t *testing.T) {
		assert.NotPanics(t, func() {
			calculateScrollBar(999, 1, 10)
		})
	})

	t.Run("doesn't panic if everything is 0", func(t *testing.T) {
		assert.NotPanics(t, func() {
			calculateScrollBar(0, 0, 0)
		})
	})
}
