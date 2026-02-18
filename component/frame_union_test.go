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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/component/comptest"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestDrawFrameUnionNoFrame(t *testing.T) {
	one := &TestComponent{Ch: 'X'}
	two := &TestComponent{Ch: 'B'}
	three := &TestComponent{Ch: 'b'}
	four := &TestComponent{Ch: 'x'}
	five := &TestComponent{Ch: '\''}
	main := &TestComponent{Ch: 'A'}
	f := NewFrameUnion(main)
	f.Frame = false
	f.Resize(20, 16)
	f.UnionTop(one, 1)

	w := term.NewStringWriter(20, 20)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
XXXXXXXXXXXXXXXXXXXX
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
                    
                    
                    
                    `,
		}, {
			Action: func() {
				f.Resize(2, 3)
			}, Expected: `
XX                  
AA                  
AA                  
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    `,
		}, {
			Action: func() {
				f.Resize(3, 3)
			}, Expected: `
XXX                 
AAA                 
AAA                 
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    `,
		}, {
			Action: func() {
				f.Resize(2, 2)
			}, Expected: `
XX                  
AA                  
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    `,
		}, {
			Action: func() {
				f.UnionBottom(two, 1)
				f.Resize(2, 2)
			}, Expected: `
AA                  
AA                  
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    `,
		}, {
			Action: func() {
				f.Resize(20, 16)
			}, Expected: `
XXXXXXXXXXXXXXXXXXXX
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
BBBBBBBBBBBBBBBBBBBB
                    
                    
                    
                    `,
		}, {
			Action: func() {
				f.UnionBottom(three, 1)
				f.UnionTop(four, 1)
				f.UnionBottom(five, 1)
				f.Resize(20, 20)
			}, Expected: `
XXXXXXXXXXXXXXXXXXXX
xxxxxxxxxxxxxxxxxxxx
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAA
''''''''''''''''''''
bbbbbbbbbbbbbbbbbbbb
BBBBBBBBBBBBBBBBBBBB`,
		},
	}

	comptest.TestComponent(t, f, w, tests)
}

func TestDrawFrameUnionWithFrame(t *testing.T) {
	one := NewFrame(&TestComponent{Ch: 'X'})
	two := NewFrame(&TestComponent{Ch: 'B'})
	three := NewFrame(&TestComponent{Ch: 'b'})
	four := NewFrame(&TestComponent{Ch: 'x'})
	five := NewFrame(&TestComponent{Ch: '\''})
	six := NewFrame(&TestComponent{Ch: '6'})
	seven := NewFrame(&TestComponent{Ch: '7'})
	eight := NewFrame(&TestComponent{Ch: '8'})
	nine := NewFrame(&TestComponent{Ch: '9'})
	main := NewFrame(&TestComponent{Ch: 'A'})
	f := NewFrameUnion(main)
	f.Resize(20, 16)
	f.UnionTop(one, 3)

	w := term.NewStringWriter(20, 20)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
┌──────────────────┐
│XXXXXXXXXXXXXXXXXX│
├──────────────────┤
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
└──────────────────┘
                    
                    
                    
                    `,
		}, {
			Action: func() {
				f.Left = '┊'
				f.Right = '┊'
			}, Expected: `
┌──────────────────┐
│XXXXXXXXXXXXXXXXXX│
┊──────────────────┊
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
└──────────────────┘
                    
                    
                    
                    `,
		}, {
			Action: func() {
				f.Resize(2, 3)
			}, Expected: `
AA                  
AA                  
AA                  
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    `,
		}, {
			Action: func() {
				f.Resize(3, 3)
			}, Expected: `
┌─┐                 
│A│                 
└─┘                 
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    `,
		}, {
			Action: func() {
				f.Resize(2, 2)
			}, Expected: `
AA                  
AA                  
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    `,
		}, {
			Action: func() {
				f.UnionBottom(two, 3)
				f.Resize(2, 2)
			}, Expected: `
AA                  
AA                  
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    
                    `,
		}, {
			Action: func() {
				f.Resize(20, 16)
			}, Expected: `
┌──────────────────┐
│XXXXXXXXXXXXXXXXXX│
┊──────────────────┊
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
┊──────────────────┊
│BBBBBBBBBBBBBBBBBB│
└──────────────────┘
                    
                    
                    
                    `,
		}, {
			Action: func() {
				f.UnionBottom(three, 3)
				f.UnionTop(four, 3)
				f.UnionBottom(five, 3)
				f.UnionLeft(six, 3)
				f.UnionLeft(seven, 3)
				f.UnionRight(eight, 3)
				f.UnionRight(nine, 3)
				f.Resize(20, 20)
			}, Expected: `
┌──────────────────┐
│XXXXXXXXXXXXXXXXXX│
┊──────────────────┊
│xxxxxxxxxxxxxxxxxx│
┊─┬─┬──────────┬─┬─┊
│6│7│AAAAAAAAAA│9│8│
│6│7│AAAAAAAAAA│9│8│
│6│7│AAAAAAAAAA│9│8│
│6│7│AAAAAAAAAA│9│8│
│6│7│AAAAAAAAAA│9│8│
│6│7│AAAAAAAAAA│9│8│
│6│7│AAAAAAAAAA│9│8│
│6│7│AAAAAAAAAA│9│8│
┊─┴─┴──────────┴─┴─┊
│''''''''''''''''''│
┊──────────────────┊
│bbbbbbbbbbbbbbbbbb│
┊──────────────────┊
│BBBBBBBBBBBBBBBBBB│
└──────────────────┘`,
		},
	}

	comptest.TestComponent(t, f, w, tests)

	t.Run("ComponentAt returns the component at position offset", func(t *testing.T) {
		c, ok := f.ComponentAt(term.Coordinates{})
		require.True(t, ok)
		assert.Equal(t, one, c.C)
		assert.Equal(t, term.Coordinates{}, c.Position())

		c, ok = f.ComponentAt(term.Coordinates{X: 19})
		require.True(t, ok)
		assert.Equal(t, one, c.C)
		assert.Equal(t, term.Coordinates{}, c.Position())

		c, ok = f.ComponentAt(term.Coordinates{Y: 2, X: 19})
		require.True(t, ok)
		assert.Equal(t, one, c.C)
		assert.Equal(t, term.Coordinates{}, c.Position())

		c, ok = f.ComponentAt(term.Coordinates{Y: 3, X: 19})
		require.True(t, ok)
		assert.Equal(t, four, c.C)
		assert.Equal(t, term.Coordinates{Y: 2}, c.Position())

		c, ok = f.ComponentAt(term.Coordinates{Y: 5, X: 19})
		require.True(t, ok)
		assert.Equal(t, eight, c.C)
		assert.Equal(t, term.Coordinates{Y: 4, X: 17}, c.Position())

		c, ok = f.ComponentAt(term.Coordinates{Y: 5, X: 17})
		require.True(t, ok)
		assert.Equal(t, nine, c.C)

		c, ok = f.ComponentAt(term.Coordinates{Y: 5, X: 13})
		require.True(t, ok)
		assert.Equal(t, main, c.C)

		c, ok = f.ComponentAt(term.Coordinates{Y: 12, X: 3})
		require.True(t, ok)
		assert.Equal(t, seven, c.C)

		c, ok = f.ComponentAt(term.Coordinates{Y: 15, X: 3})
		require.True(t, ok)
		assert.Equal(t, five, c.C)

		c, ok = f.ComponentAt(term.Coordinates{Y: 19, X: 3})
		require.True(t, ok)
		assert.Equal(t, two, c.C)
		assert.Equal(t, term.Coordinates{Y: 17}, c.Position())

		c, ok = f.ComponentAt(term.Coordinates{Y: 16, X: 3})
		require.True(t, ok)
		assert.Equal(t, three, c.C)
		assert.Equal(t, term.Coordinates{Y: 15}, c.Position())

		pos := f.MainPosition()
		assert.Equal(t, term.Coordinates{X: 4, Y: 4}, pos)
	})
}

func TestDrawFrameUnionUnionNoFrame(t *testing.T) {
	main := &TestComponent{Ch: 'A'}
	one := NewFrame(&TestComponent{Ch: '1'})
	two := NewFrame(&TestComponent{Ch: '2'})
	three := NewFrame(&TestComponent{Ch: '3'})
	four := NewFrame(&TestComponent{Ch: '4'})
	five := NewFrame(&TestComponent{Ch: '5'})
	six := NewFrame(&TestComponent{Ch: '6'})
	seven := NewFrame(&TestComponent{Ch: '7'})
	eight := NewFrame(&TestComponent{Ch: '8'})

	suite := []struct {
		description string
		setup       func(f *FrameUnion)
		expected    string
	}{

		{
			description: "only no frame unions",
			setup: func(f *FrameUnion) {
				f.UnionTopFrame(four, 1, false)
				f.UnionBottomFrame(two, 1, false)
				f.UnionLeftFrame(six, 1, false)
				f.UnionRightFrame(eight, 1, false)
			},
			expected: `
44444444444444444444
6┌────────────────┐8
6│AAAAAAAAAAAAAAAA│8
6│AAAAAAAAAAAAAAAA│8
6│AAAAAAAAAAAAAAAA│8
6│AAAAAAAAAAAAAAAA│8
6│AAAAAAAAAAAAAAAA│8
6│AAAAAAAAAAAAAAAA│8
6│AAAAAAAAAAAAAAAA│8
6│AAAAAAAAAAAAAAAA│8
6│AAAAAAAAAAAAAAAA│8
6│AAAAAAAAAAAAAAAA│8
6│AAAAAAAAAAAAAAAA│8
6│AAAAAAAAAAAAAAAA│8
6└────────────────┘8
22222222222222222222
                    
                    
                    
                    `,
		},
		{
			description: "mixed top bottom unions, start with no frame, no left, or right",
			setup: func(f *FrameUnion) {
				f.UnionTopFrame(four, 1, false)
				f.UnionTop(three, 3)
				f.UnionBottomFrame(two, 1, false)
				f.UnionBottom(one, 3)
				f.UnionLeft(five, 3)
				f.UnionRight(seven, 3)
			},
			expected: `
44444444444444444444
┌──────────────────┐
│333333333333333333│
├─┬──────────────┬─┤
│5│AAAAAAAAAAAAAA│7│
│5│AAAAAAAAAAAAAA│7│
│5│AAAAAAAAAAAAAA│7│
│5│AAAAAAAAAAAAAA│7│
│5│AAAAAAAAAAAAAA│7│
│5│AAAAAAAAAAAAAA│7│
│5│AAAAAAAAAAAAAA│7│
│5│AAAAAAAAAAAAAA│7│
├─┴──────────────┴─┤
│111111111111111111│
└──────────────────┘
22222222222222222222
                    
                    
                    
                    `,
		},
		{
			description: "mixed top bottom unions, start with frame, no left, or right",
			setup: func(f *FrameUnion) {
				f.UnionTop(three, 3)
				f.UnionTopFrame(four, 1, false)
				f.UnionBottom(one, 3)
				f.UnionBottomFrame(two, 1, false)
				f.UnionLeft(five, 3)
				f.UnionRight(seven, 3)
			},
			expected: `
┌──────────────────┐
│333333333333333333│
└──────────────────┘
44444444444444444444
┌─┬──────────────┬─┐
│5│AAAAAAAAAAAAAA│7│
│5│AAAAAAAAAAAAAA│7│
│5│AAAAAAAAAAAAAA│7│
│5│AAAAAAAAAAAAAA│7│
│5│AAAAAAAAAAAAAA│7│
│5│AAAAAAAAAAAAAA│7│
└─┴──────────────┴─┘
22222222222222222222
┌──────────────────┐
│111111111111111111│
└──────────────────┘
                    
                    
                    
                    `,
		},
		{
			description: "mixed left right unions, start with no frame, no top , or bottom",
			setup: func(f *FrameUnion) {
				f.UnionTop(three, 3)
				f.UnionBottom(one, 3)
				f.UnionLeftFrame(four, 1, false)
				f.UnionLeft(five, 3)
				f.UnionRightFrame(two, 1, false)
				f.UnionRight(seven, 3)
			},
			expected: `
┌──────────────────┐
│333333333333333333│
└┬─┬────────────┬─┬┘
4│5│AAAAAAAAAAAA│7│2
4│5│AAAAAAAAAAAA│7│2
4│5│AAAAAAAAAAAA│7│2
4│5│AAAAAAAAAAAA│7│2
4│5│AAAAAAAAAAAA│7│2
4│5│AAAAAAAAAAAA│7│2
4│5│AAAAAAAAAAAA│7│2
4│5│AAAAAAAAAAAA│7│2
4│5│AAAAAAAAAAAA│7│2
4│5│AAAAAAAAAAAA│7│2
┌┴─┴────────────┴─┴┐
│111111111111111111│
└──────────────────┘
                    
                    
                    
                    `,
		},
		{
			description: "mixed left right unions, start with frame, top and bottom frame",
			setup: func(f *FrameUnion) {
				f.UnionTop(three, 3)
				f.UnionBottom(one, 3)
				f.UnionLeft(five, 3)
				f.UnionLeftFrame(four, 1, false)
				f.UnionRight(seven, 3)
				f.UnionRightFrame(two, 1, false)
			},
			expected: `
┌──────────────────┐
│333333333333333333│
├─┐4┌──────────┐2┌─┤
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
├─┘4└──────────┘2└─┤
│111111111111111111│
└──────────────────┘
                    
                    
                    
                    `,
		},
		{
			description: "mixed left right unions, start with frame, last top first bottom no frame",
			setup: func(f *FrameUnion) {
				f.UnionTopFrame(three, 1, false)
				f.UnionBottomFrame(one, 1, false)
				f.UnionLeft(five, 3)
				f.UnionLeftFrame(four, 1, false)
				f.UnionRight(seven, 3)
				f.UnionRightFrame(two, 1, false)
			},
			expected: `
33333333333333333333
┌─┐4┌──────────┐2┌─┐
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
│5│4│AAAAAAAAAA│2│7│
└─┘4└──────────┘2└─┘
11111111111111111111
                    
                    
                    
                    `,
		},
	}

	for _, test := range suite {
		t.Run(test.description, func(t *testing.T) {
			main := NewFrame(main)
			f := NewFrameUnion(main)
			test.setup(f)
			f.Resize(20, 16)

			w := term.NewStringWriter(20, 20)
			tests := []comptest.TestCase{
				{Action: nil, Expected: test.expected},
			}

			comptest.TestComponent(t, f, w, tests)

		})
	}

}

func TestDrawFrameUnionWithFrameBottomOnly(t *testing.T) {
	one := NewFrame(&TestComponent{Ch: 'X'})
	main := NewFrame(&TestComponent{Ch: 'A'})
	f := NewFrameUnion(main)
	f.UnionBottom(one, 3)
	f.Resize(20, 16)

	w := term.NewStringWriter(20, 20)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
┌──────────────────┐
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
│AAAAAAAAAAAAAAAAAA│
├──────────────────┤
│XXXXXXXXXXXXXXXXXX│
└──────────────────┘
                    
                    
                    
                    `,
		},
	}

	comptest.TestComponent(t, f, w, tests)
}

func TestFrameUnionResizeNegative(t *testing.T) {
	suite := []struct {
		description   string
		width, height int
		setup         func(f *FrameUnion)
	}{
		{
			description: "negative height with top union",
			width:       10,
			height:      -1,
			setup: func(f *FrameUnion) {
				f.UnionTop(&TestComponent{Ch: 'T'}, 3)
			},
		},
		{
			description: "negative width with left union",
			width:       -1,
			height:      10,
			setup: func(f *FrameUnion) {
				f.UnionLeft(&TestComponent{Ch: 'L'}, 3)
			},
		},
		{
			description: "negative height with bottom union",
			width:       10,
			height:      -1,
			setup: func(f *FrameUnion) {
				f.UnionBottom(&TestComponent{Ch: 'B'}, 3)
			},
		},
		{
			description: "negative width with right union",
			width:       -1,
			height:      10,
			setup: func(f *FrameUnion) {
				f.UnionRight(&TestComponent{Ch: 'R'}, 3)
			},
		},
	}

	for _, test := range suite {
		t.Run(test.description, func(t *testing.T) {
			assert.NotPanics(t, func() {
				main := &TestComponent{Ch: 'A'}
				f := NewFrameUnion(main)
				test.setup(f)
				f.Resize(test.width, test.height)
			})
		})
	}
}

func TestComponentAtOutOfBounds(t *testing.T) {
	f := NewFrameUnion(&TestComponent{})
	f.Resize(10, 10)

	_, ok := f.ComponentAt(term.Coordinates{X: 10})
	require.False(t, ok)
	_, ok = f.ComponentAt(term.Coordinates{Y: 10})
	require.False(t, ok)
}
