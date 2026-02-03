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
	"github.com/unstablebuild/rune-go-sdk/tui"
)

type responsiveTestList struct {
	elementHeight int
	*ResponsiveList
}

type testListResponsive struct {
	tui.Component
	height *int
}

func (t *testListResponsive) Height(width int) int {
	return *t.height
}

func (l *responsiveTestList) newTestResponsive(c tui.Component) Responsive {
	// emulate List behaviour, and test max width default
	v := &testListResponsive{Component: c, height: &l.elementHeight}
	return v
}

func (l *responsiveTestList) PushBackList(other testList) {
	l.ResponsiveList.PushBackList(other.(*responsiveTestList).ResponsiveList)
}
func (l *responsiveTestList) PushFrontList(other testList) {
	l.ResponsiveList.PushFrontList(other.(*responsiveTestList).ResponsiveList)
}

func (l *responsiveTestList) PushBack(c tui.Component) ListNode {
	return l.ResponsiveList.PushBack(l.newTestResponsive(c))
}

func (l *responsiveTestList) PushFront(c tui.Component) ListNode {
	return l.ResponsiveList.PushFront(l.newTestResponsive(c))
}

func (l *responsiveTestList) Remove(e *ListNode) tui.Component {
	return l.ResponsiveList.Remove(e).(*testListResponsive).Component
}

func (l *responsiveTestList) Sort(less func(a, b tui.Component) bool) {
	l.ResponsiveList.Sort(func(a, b Responsive) bool {
		return less(a.(*testListResponsive).Component, b.(*testListResponsive).Component)
	})
}

func (l *responsiveTestList) SetElementHeight(i int) {
	// setting element height from here bypassing ResponsiveList
	// allows for all test Responsive components to be resized
	// and emulate SetElementHeight behaviour.
	l.elementHeight = i
	l.Resize(l.list.width, l.list.height)
}

func (l *responsiveTestList) ElementHeight() int {
	return l.elementHeight
}

func newResponsiveTestList(i int) testList {
	ret := &responsiveTestList{ResponsiveList: NewResponsiveList()}
	ret.elementHeight = i
	return ret
}

func TestResponsiveListSort(t *testing.T) {
	testListSort(t, newResponsiveTestList)
}

func TestResponsiveListFrontBack(t *testing.T) {
	testFrontBack(t, newResponsiveTestList)
}

func TestResponsiveListListRemove(t *testing.T) {
	testListRemove(t, newResponsiveTestList)
}

func TestResponsiveListEmptyDraw(t *testing.T) {
	testEmptyListDraw(t, newResponsiveTestList)
}

func TestResponsiveListResponsiveness(t *testing.T) {
	l := NewResponsiveList()
	l.Resize(8, 4)

	w := term.NewStringWriter(8, 4)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
        
        
        
        `,
		}, {
			Action: func() {
				l.PushBack(&TestResponsive{
					WantHeight: 2,
					TestComponent: TestComponent{
						Ch: 'X',
					},
				},
				)
			}, Expected: `
XXXXXXXX
XXXXXXXX
        
        `,
		}, {
			Action: func() {
				l.PushBack(&TestResponsive{
					WantHeight: 2,
					TestComponent: TestComponent{
						Ch: 'Y',
					},
				},
				)
			}, Expected: `
XXXXXXXX
XXXXXXXX
YYYYYYYY
YYYYYYYY`,
		}, {
			Action: func() {
				l.PushBack(&TestResponsive{
					WantHeight: 3,
					TestComponent: TestComponent{
						Ch: 'Z',
					},
				},
				)
				l.SeekEnd()
			}, Expected: `
YYYYYYYY
ZZZZZZZZ
ZZZZZZZZ
ZZZZZZZZ`,
		}, {
			Action: func() {
				el, ok := l.ElementAt(term.Coordinates{})
				require.True(t, ok)
				v := el.Value().(*TestResponsive)
				assert.Equal(t, 'Y', v.Ch)

				el, ok = l.ElementAt(term.Coordinates{Y: 1})
				require.True(t, ok)
				v = el.Value().(*TestResponsive)
				assert.Equal(t, 'Z', v.Ch)

				el, ok = l.ElementAt(term.Coordinates{Y: 3})
				require.True(t, ok)
				v = el.Value().(*TestResponsive)
				assert.Equal(t, 'Z', v.Ch)
			}, Expected: `
YYYYYYYY
ZZZZZZZZ
ZZZZZZZZ
ZZZZZZZZ`,
		},
	}

	comptest.TestComponent(t, l, w, tests)
}

func TestResponsiveListResizeLarger(t *testing.T) {
	l := NewResponsiveList()

	w := term.NewStringWriter(8, 4)

	tests := []comptest.TestCase{
		{
			Action: func() {
				l.PushBack(&TestResponsive{
					WantHeight: 2,
					TestComponent: TestComponent{
						Ch: 'X',
					},
				},
				)
				l.PushBack(&TestResponsive{
					WantHeight: 2,
					TestComponent: TestComponent{
						Ch: 'Y',
					},
				},
				)
				l.PushBack(&TestResponsive{
					WantHeight: 3,
					TestComponent: TestComponent{
						Ch: 'Z',
					},
				},
				)
				assert.True(t, l.SeekEnd())
				l.Resize(8, 4)
				assert.True(t, l.SeekEnd())
			}, Expected: `
YYYYYYYY
ZZZZZZZZ
ZZZZZZZZ
ZZZZZZZZ`,
		}, {
			Action: func() {
				el, ok := l.ElementAt(term.Coordinates{})
				require.True(t, ok)
				v := el.Value().(*TestResponsive)
				assert.Equal(t, 'Y', v.Ch)

				el, ok = l.ElementAt(term.Coordinates{Y: 1})
				require.True(t, ok)
				v = el.Value().(*TestResponsive)
				assert.Equal(t, 'Z', v.Ch)

				el, ok = l.ElementAt(term.Coordinates{Y: 2})
				require.True(t, ok)
				v = el.Value().(*TestResponsive)
				assert.Equal(t, 'Z', v.Ch)

				el, ok = l.ElementAt(term.Coordinates{Y: 3})
				require.True(t, ok)
				v = el.Value().(*TestResponsive)
				assert.Equal(t, 'Z', v.Ch)
			}, Expected: `
YYYYYYYY
ZZZZZZZZ
ZZZZZZZZ
ZZZZZZZZ`,
		},
	}

	comptest.TestComponent(t, l, w, tests)
}

func TestResponsiveListScroll(t *testing.T) {
	l := NewResponsiveList()

	w := term.NewStringWriter(8, 4)
	l.PushBack(testResponsive('a', 2))
	l.PushBack(testResponsive('b', 1))
	l.PushBack(testResponsive('c', 3))
	l.Resize(7, 3)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
aaaaaaa 
aaaaaaa 
bbbbbbb 
        `,
		}, {
			Action: func() {
				require.True(t, l.SeekDown())
			}, Expected: `
aaaaaaa 
bbbbbbb 
ccccccc 
        `,
		}, {
			Action: func() {
				require.True(t, l.SeekDown())
			}, Expected: `
bbbbbbb 
ccccccc 
ccccccc 
        `,
		}, {
			Action: func() {
				require.True(t, l.SeekDown())
			}, Expected: `
ccccccc 
ccccccc 
ccccccc 
        `,
		}, {
			Action: func() {
				assert.False(t, l.SeekDown())
				require.True(t, l.SeekUp())
			}, Expected: `
bbbbbbb 
ccccccc 
ccccccc 
        `,
		}, {
			Action: func() {
				require.True(t, l.SeekUp())
			}, Expected: `
aaaaaaa 
bbbbbbb 
ccccccc 
        `,
		}, {
			Action: func() {
				require.True(t, l.SeekUp())
				assert.False(t, l.SeekUp())
			}, Expected: `
aaaaaaa 
aaaaaaa 
bbbbbbb 
        `,
		},
	}

	comptest.TestComponent(t, l, w, tests)
}

func TestResponsiveListScrollAlignmentBottom(t *testing.T) {
	l := NewResponsiveList()
	l.Alignment = AlignmentBottom

	w := term.NewStringWriter(8, 4)
	l.PushBack(testResponsive('a', 2))
	l.PushBack(testResponsive('b', 1))
	l.PushBack(testResponsive('c', 3))
	l.Resize(7, 3)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
ccccccc 
ccccccc 
ccccccc 
        `,
		}, {
			Action: func() {
				for y := 0; y < 3; y++ {
					el, ok := l.ElementAt(term.Coordinates{Y: y})
					require.True(t, ok, y)
					v := el.Value().(*TestResponsive)
					assert.Equal(t, 'c', v.Ch)
				}
				require.True(t, l.SeekUp())
			}, Expected: `
bbbbbbb 
ccccccc 
ccccccc 
        `,
		}, {
			Action: func() {
				el, ok := l.ElementAt(term.Coordinates{Y: 0})
				require.True(t, ok)
				v := el.Value().(*TestResponsive)
				assert.Equal(t, 'b', v.Ch)
				el, ok = l.ElementAt(term.Coordinates{Y: 1})
				require.True(t, ok)
				v = el.Value().(*TestResponsive)
				assert.Equal(t, 'c', v.Ch)
				require.True(t, l.SeekUp())
			}, Expected: `
aaaaaaa 
bbbbbbb 
ccccccc 
        `,
		}, {
			Action: func() {
				el, ok := l.ElementAt(term.Coordinates{Y: 0})
				require.True(t, ok)
				v := el.Value().(*TestResponsive)
				assert.Equal(t, 'a', v.Ch)
				require.True(t, l.SeekUp())
			}, Expected: `
aaaaaaa 
aaaaaaa 
bbbbbbb 
        `,
		}, {
			Action: func() {
				assert.False(t, l.SeekUp())
				require.True(t, l.SeekDown())
				el, ok := l.ElementAt(term.Coordinates{Y: 0})
				require.True(t, ok)
				v := el.Value().(*TestResponsive)
				assert.Equal(t, 'a', v.Ch)
				el, ok = l.ElementAt(term.Coordinates{Y: 2})
				require.True(t, ok)
				v = el.Value().(*TestResponsive)
				assert.Equal(t, 'c', v.Ch)
			}, Expected: `
aaaaaaa 
bbbbbbb 
ccccccc 
        `,
		}, {
			Action: func() {
				require.True(t, l.SeekDown())
			}, Expected: `
bbbbbbb 
ccccccc 
ccccccc 
        `,
		}, {
			Action: func() {
				require.True(t, l.SeekDown())
				assert.False(t, l.SeekDown())
			}, Expected: `
ccccccc 
ccccccc 
ccccccc 
        `,
		},
	}

	comptest.TestComponent(t, l, w, tests)
}
