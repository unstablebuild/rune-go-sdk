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

func TestContainerDimensions(t *testing.T) {
	l := NewContainer()
	l.Resize(4, 4)

	row1 := l.AddRow()
	row1.AddComponent(testResponsive('a', 2), MaxCols)

	row2 := l.AddRow()
	row2.AddComponent(testResponsive('a', 3), MaxCols/2)
	row2.AddComponent(testResponsive('a', 3), MaxCols/2)

	width, height := l.Dimensions()
	assert.Equal(t, 6, width)
	assert.Equal(t, 5, height)
}

func TestContainerDraw(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		l := NewContainer()
		l.Resize(4, 4)
		w := term.NewStringWriter(20, 9)
		tests := []comptest.TestCase{
			{
				Action: nil, Expected: `
                    
                    
                    
                    
                    
                    
                    
                    
                    `,
			},
		}
		comptest.TestComponent(t, l, w, tests)
	})

	t.Run("happy path", func(t *testing.T) {
		l := NewContainer()
		r1 := l.AddRow()
		a := testResponsive('a', 1)
		r1.AddComponent(a, 6)
		r1.AddComponent(testResponsive('b', 1), 6)
		r2 := l.AddRow()
		r2.AddComponent(testResponsive('c', 1), 6)
		r2.AddComponent(testResponsive('d', 1), 6)
		r3 := l.AddRow()
		r3.AddComponent(testResponsive('e', 1), 6)
		r3.AddComponent(testResponsive('f', 1), 6)
		l.Resize(20, 4)

		w := term.NewStringWriter(20, 9)

		tests := []comptest.TestCase{
			{
				Action: nil, Expected: `
aaaaaaaaaabbbbbbbbbb
ccccccccccdddddddddd
eeeeeeeeeeffffffffff
                    
                    
                    
                    
                    
                    `,
			}, {
				Action: func() { l.Resize(20, 9) }, Expected: `
aaaaaaaaaabbbbbbbbbb
ccccccccccdddddddddd
eeeeeeeeeeffffffffff
                    
                    
                    
                    
                    
                    `,
			}, {
				Action: func() { l.Resize(2, 2) }, Expected: `
ab                  
cd                  
                    
                    
                    
                    
                    
                    
                    `,
			}, {
				Action: func() { require.False(t, l.ScrollUp()) }, Expected: `
ab                  
cd                  
                    
                    
                    
                    
                    
                    
                    `,
			}, {
				Action: func() { require.True(t, l.ScrollDown()) }, Expected: `
cd                  
ef                  
                    
                    
                    
                    
                    
                    
                    `,
			}, {
				Action: func() { require.True(t, l.ScrollDown()) }, Expected: `
ef                  
                    
                    
                    
                    
                    
                    
                    
                    `,
			}, {
				Action: func() { require.False(t, l.ScrollDown()) }, Expected: `
ef                  
                    
                    
                    
                    
                    
                    
                    
                    `,
			}, {
				Action: func() {
					for l.ScrollUp() {
					}
					l.Resize(20, 9)
					row := l.AddRow()
					row.AddComponent(testResponsive('Z', 1), 3)
					row.AddComponent(testResponsive('Y', 1), 6)
					row.AddComponent(testResponsive('X', 1), 3)
				}, Expected: `
aaaaaaaaaabbbbbbbbbb
ccccccccccdddddddddd
eeeeeeeeeeffffffffff
ZZZZZYYYYYYYYYYXXXXX
                    
                    
                    
                    
                    `,
			}, {
				Action: func() {
					row := l.AddRow()
					row.AddComponent(testResponsive('3', 3), 3)
					row.AddComponent(testResponsive('2', 2), 3)
					row.AddComponent(testResponsive('1', 1), 3)
				}, Expected: `
aaaaaaaaaabbbbbbbbbb
ccccccccccdddddddddd
eeeeeeeeeeffffffffff
ZZZZZYYYYYYYYYYXXXXX
333332222211111     
333332222211111     
333332222211111     
                    
                    `,
			}, {
				Action: func() {
					// changing height requirements without container
					// or row "knowing" about it
					a.WantHeight = 2
				}, Expected: `
aaaaaaaaaabbbbbbbbbb
aaaaaaaaaabbbbbbbbbb
ccccccccccdddddddddd
eeeeeeeeeeffffffffff
ZZZZZYYYYYYYYYYXXXXX
333332222211111     
333332222211111     
333332222211111     
                    `,
			}, {
				Action: func() {
					a.WantHeight = 4
				}, Expected: `
aaaaaaaaaabbbbbbbbbb
aaaaaaaaaabbbbbbbbbb
aaaaaaaaaabbbbbbbbbb
aaaaaaaaaabbbbbbbbbb
ccccccccccdddddddddd
eeeeeeeeeeffffffffff
ZZZZZYYYYYYYYYYXXXXX
333332222211111     
333332222211111     `,
			}, {
				Action: func() {
					require.True(t, l.ScrollDown())
				}, Expected: `
aaaaaaaaaabbbbbbbbbb
aaaaaaaaaabbbbbbbbbb
aaaaaaaaaabbbbbbbbbb
ccccccccccdddddddddd
eeeeeeeeeeffffffffff
ZZZZZYYYYYYYYYYXXXXX
333332222211111     
333332222211111     
333332222211111     `,
			}, {
				Action: func() {
					for l.ScrollDown() {
					}
					row := l.AddRow()
					row.AddComponent(testResponsive('@', 3), 6)
					row.AddComponent(testResponsive('#', 2), 5)
					row.AddComponent(testResponsive('$', 1), 5)
				}, Expected: `
333332222211111     
@@@@@@@@@@########$$
@@@@@@@@@@########$$
@@@@@@@@@@########$$
                    
                    
                    
                    
                    `,
			},
		}
		comptest.TestComponent(t, l, w, tests)
	})

	t.Run("zero row", func(t *testing.T) {
		l := NewContainer()
		r1 := l.AddRow()
		r1.AddComponent(testResponsive('a', 1), 12)
		_ = l.AddRow()
		r3 := l.AddRow()
		r3.AddComponent(testResponsive('b', 1), 12)
		l.Resize(4, 4)
		w := term.NewStringWriter(20, 9)
		tests := []comptest.TestCase{
			{
				Action: nil, Expected: `
aaaa                
bbbb                
                    
                    
                    
                    
                    
                    
                    `,
			},
		}
		comptest.TestComponent(t, l, w, tests)
	})

}

func testResponsive(ch rune, wantHeight int) *TestResponsive {
	return &TestResponsive{
		WantWidth:  wantHeight,
		WantHeight: wantHeight,
		TestComponent: TestComponent{
			Ch: ch,
		},
	}
}
