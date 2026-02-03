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
	"github.com/unstablebuild/tcell/v3"
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
		f.SetAttr(term.Attributes{Fg: tcell.ColorRed, Bg: tcell.ColorGreen})

		assert.Equal(t, tcell.ColorGreen, f.Bg)
		assert.Equal(t, tcell.ColorRed, f.Fg)

		prev := f.Content().(WithAttributes).SetAttr(term.Attributes{})
		assert.Equal(t, tcell.ColorGreen, prev.Bg)
		assert.Equal(t, tcell.ColorRed, prev.Fg)
	})

	t.Run("component does not WithAttributes", func(t *testing.T) {
		f := NewFrame(&TestComponent{Ch: 'a'})
		f.SetAttr(term.Attributes{Fg: tcell.ColorRed, Bg: tcell.ColorGreen})

		assert.Equal(t, tcell.ColorGreen, f.Bg)
		assert.Equal(t, tcell.ColorRed, f.Fg)
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
