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
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/component/comptest"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

func TestNewList(t *testing.T) {
	l := NewList(1)

	l.Resize(8, 4)

	if l.ElementHeight() != 1 || l.width != 8 || l.height != 4 {
		t.Errorf("sizes not initialized correctly: %+v", l)
	}

	var l2 List

	l2.Init(1)
	l.Resize(8, 4)

	if l.ElementHeight() != 1 || l.width != 8 || l.height != 4 {
		t.Errorf("sizes not initialized correctly: %+v", l)
	}
}

func testEmptyListDraw(t *testing.T, constructor func(int) testList) {
	l := constructor(1)
	w := term.NewStringWriter(8, 4)

	assert.NotPanics(t, func() {
		l.Draw(w)
		l.Resize(10, 10)
		l.Draw(w)
	})
}

type testList interface {
	Back() (ListNode, bool)
	CanSeekDown() bool
	CanSeekUp() bool
	Draw(w term.Writer)
	ElementAt(pos term.Coordinates) (ListNode, bool)
	ElementHeight() int
	Front() (ListNode, bool)
	Len() int
	PushBackList(other testList)
	PushFrontList(other testList)
	PushBack(c tui.Component) ListNode
	PushFront(c tui.Component) ListNode
	Remove(e *ListNode) tui.Component
	Resize(width, height int)
	SeekDown() bool
	SeekEnd() (ok bool)
	SeekStart() (ok bool)
	SeekUp() bool
	MaxOffset() int
	SetElementHeight(height int)
	Sort(func(a, b tui.Component) bool)
	Reset()
}

type listTestList struct {
	*List
}

func (l *listTestList) PushBackList(other testList) {
	l.List.PushBackList(other.(*listTestList).List)
}
func (l *listTestList) PushFrontList(other testList) {
	l.List.PushFrontList(other.(*listTestList).List)
}
func (l *listTestList) Sort(less func(a, b tui.Component) bool) {
	l.List.Sort(less)
}

func newTestList(i int) testList {
	ret := &listTestList{List: NewList(i)}
	return ret
}

func makeListOfTwo() (*List, []*TestComponent) {
	l := NewList(1)
	el1 := &TestComponent{Ch: 'A'}
	el2 := &TestComponent{Ch: 'b'}

	l.PushBack(el1)
	l.PushBack(el2)
	return l, []*TestComponent{el1, el2}
}

func testListDraw(t *testing.T, skipCases []int, constructor func(int) testList) {
	l := constructor(1)
	l2 := constructor(1)

	l.Resize(8, 4)

	l2.Resize(8, 4)

	w := term.NewStringWriter(8, 4)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
        
        
        
        `,
		}, {
			Action: func() { l.PushBack(&TestComponent{Ch: 'X'}) }, Expected: `
XXXXXXXX
        
        
        `,
		}, {
			Action: func() { l.PushBack(&TestComponent{Ch: 'Y'}) }, Expected: `
XXXXXXXX
YYYYYYYY
        
        `,
		}, {
			Action: func() { l.SeekDown() }, Expected: `
XXXXXXXX
YYYYYYYY
        
        `,
		}, {
			Action: func() { l.SeekUp() }, Expected: `
XXXXXXXX
YYYYYYYY
        
        `,
		}, {
			Action: func() { l.PushFront(&TestComponent{Ch: 'Z'}) }, Expected: `
ZZZZZZZZ
XXXXXXXX
YYYYYYYY
        `,
		}, {
			Action: func() {
				l2.PushFront(&TestComponent{Ch: '$'})
				l2.PushFront(&TestComponent{Ch: '#'})
				l.PushBackList(l2)
			}, Expected: `
ZZZZZZZZ
XXXXXXXX
YYYYYYYY
########`,
		}, {
			Action: func() { l.SeekUp() }, Expected: `
ZZZZZZZZ
XXXXXXXX
YYYYYYYY
########`,
		}, {
			Action: func() { l.SeekDown() }, Expected: `
XXXXXXXX
YYYYYYYY
########
$$$$$$$$`,
		}, {
			Action: func() { l.SeekEnd() }, Expected: `
XXXXXXXX
YYYYYYYY
########
$$$$$$$$`,
		}, {
			Action: func() {
				l.Resize(4, 4)
				el, ok := l.ElementAt(term.Coordinates{Y: 0})
				require.True(t, ok)
				assert.Equal(t, 'X', getTestComponent(el.Value()).Ch)
				el, ok = l.ElementAt(term.Coordinates{Y: 1})
				require.True(t, ok)
				assert.Equal(t, 'Y', getTestComponent(el.Value()).Ch)
				el, ok = l.ElementAt(term.Coordinates{Y: 2})
				require.True(t, ok)
				assert.Equal(t, '#', getTestComponent(el.Value()).Ch)
				el, ok = l.ElementAt(term.Coordinates{Y: 3})
				require.True(t, ok)
				assert.Equal(t, '$', getTestComponent(el.Value()).Ch)
				_, ok = l.ElementAt(term.Coordinates{Y: 4})
				require.False(t, ok)
				assert.Panics(t, func() {
					l.ElementAt(term.Coordinates{Y: -1})
				})
			}, Expected: `
XXXX    
YYYY    
####    
$$$$    `,
		}, {
			Action: func() { l.SetElementHeight(2) }, Expected: `
XXXX    
XXXX    
YYYY    
YYYY    `,
		}, {
			Action: func() { l.Resize(8, 4) }, Expected: `
XXXXXXXX
XXXXXXXX
YYYYYYYY
YYYYYYYY`,
		}, {
			Action: func() { l.SetElementHeight(3) }, Expected: `
XXXXXXXX
XXXXXXXX
XXXXXXXX
        `,
		}, {
			Action: func() { l.SeekDown() }, Expected: `
YYYYYYYY
YYYYYYYY
YYYYYYYY
        `,
		}, {
			Action: func() { l.SeekEnd() }, Expected: `
$$$$$$$$
$$$$$$$$
$$$$$$$$
        `,
		}, {
			Action: func() { l.SeekStart() }, Expected: `
ZZZZZZZZ
ZZZZZZZZ
ZZZZZZZZ
        `,
		}, {
			Action: func() {
				l.SetElementHeight(1)
				l.Sort(func(a, b tui.Component) bool {
					return a.(*TestComponent).Ch < b.(*TestComponent).Ch
				})
			}, Expected: `
########
$$$$$$$$
XXXXXXXX
YYYYYYYY`,
		}, {
			Action: func() {
				front, ok := l.Front()
				require.True(t, ok)
				l.Remove(&front)
			}, Expected: `
$$$$$$$$
XXXXXXXX
YYYYYYYY
ZZZZZZZZ`,
		},
		{
			Action: func() {
				require.False(t, l.SeekEnd())
				require.Equal(t, 4, l.Len())
				for i := 'a'; i < 'a'+16; i++ {
					l.PushFront(&TestComponent{Ch: rune(i)})
				}
				require.Equal(t, 20, l.Len())
				require.True(t, l.SeekEnd())
			}, Expected: `
$$$$$$$$
XXXXXXXX
YYYYYYYY
ZZZZZZZZ`,
		}, {
			Action: func() {
				z, ok := l.Back()
				require.True(t, ok)
				l.Remove(&z)
			}, Expected: `
aaaaaaaa
$$$$$$$$
XXXXXXXX
YYYYYYYY`,
		}, {
			Action: func() {
				y, ok := l.Back()
				require.True(t, ok)
				x, ok := y.Prev()
				require.True(t, ok)
				dollas, ok := x.Prev()
				require.True(t, ok)
				head, ok := dollas.Prev()
				require.True(t, ok)
				l.Remove(&head)
			}, Expected: `
bbbbbbbb
$$$$$$$$
XXXXXXXX
YYYYYYYY`,
		}, {
			Action: func() {
				y, ok := l.Back()
				require.True(t, ok)
				x, ok := y.Prev()
				require.True(t, ok)
				dollas, ok := x.Prev()
				require.True(t, ok)
				b, ok := dollas.Prev()
				require.True(t, ok)
				c, ok := b.Prev()
				require.True(t, ok)
				l.Remove(&c)
			}, Expected: `
bbbbbbbb
$$$$$$$$
XXXXXXXX
YYYYYYYY`,
		}, {
			Action: func() {
				y, ok := l.Back()
				require.True(t, ok)
				x, ok := y.Prev()
				require.True(t, ok)
				dollas, ok := x.Prev()
				require.True(t, ok)
				b, ok := dollas.Prev()
				require.True(t, ok)
				for i := 0; i < 4; i++ {
					require.True(t, l.SeekUp())
				}
				l.Remove(&b)
			}, Expected: `
gggggggg
ffffffff
eeeeeeee
dddddddd`,
		}, {
			Action: func() {
				for node, ok := l.Back(); ok; node, ok = l.Front() {
					l.Remove(&node)
				}
				require.Equal(t, 0, l.Len())
				l.PushBack(&TestComponent{Ch: '%'})
			}, Expected: `
%%%%%%%%
        
        
        `,
		},
	}

	var filteredTests []comptest.TestCase
	for i, test := range tests {
		include := true
		for _, skipCase := range skipCases {
			if i == skipCase {
				include = false
				break
			}
		}
		if include {
			filteredTests = append(filteredTests, test)
		}
	}

	comptest.TestComponent(t, l, w, filteredTests)
}

func TestListNode(t *testing.T) {
	l := NewList(1)
	c1, c2 := NewString(""), NewString("")
	c3, c4 := NewString(""), NewString("")
	el1 := l.PushFront(c1)
	el2 := l.PushBack(c2)
	el4 := l.InsertAfter(c4, el2)
	el3 := l.InsertBefore(c3, el4)

	t.Run("Value returns the component instance passed in PushFront", func(t *testing.T) {
		assert.Equal(t, c1, el1.Value())
	})

	t.Run("Value returns the component instance passed in PushBack", func(t *testing.T) {
		assert.Equal(t, c2, el2.Value())
	})

	t.Run("Value returns the component instance passed in PushFront", func(t *testing.T) {
		assert.Equal(t, c3, el3.Value())
	})

	t.Run("Value returns the component instance passed in PushBack", func(t *testing.T) {
		assert.Equal(t, c4, el4.Value())
	})

	t.Run("Value on non-linked element returns false", func(t *testing.T) {
		assert.Nil(t, ListNode{}.Value())
	})

	t.Run("Prev returns the previous node", func(t *testing.T) {
		prev, ok := el2.Prev()
		require.True(t, ok)
		assert.Equal(t, el1, prev)
	})

	t.Run("Prev returns false if first node", func(t *testing.T) {
		_, ok := el1.Prev()
		require.False(t, ok)
	})

	t.Run("Next returns the next node", func(t *testing.T) {
		next, ok := el3.Next()
		require.True(t, ok)
		assert.Equal(t, el4, next)
	})

	t.Run("Next returns false if last node", func(t *testing.T) {
		_, ok := el4.Next()
		require.False(t, ok)
	})
}

func testListRemove(t *testing.T, constructor func(int) testList) {
	t.Run("Remove returns element in list", func(t *testing.T) {
		el := &TestComponent{}
		l := constructor(1)
		n := l.PushBack(el)

		ret := l.Remove(&n)
		assert.Equal(t, el, getTestComponent(ret))
	})

	t.Run("Remove called twice on same node panics", func(t *testing.T) {
		el := &TestComponent{}
		l := constructor(1)
		n := l.PushBack(el)
		assert.Equal(t, 1, l.Len())

		ret := l.Remove(&n)
		assert.Equal(t, 0, l.Len())
		assert.NotNil(t, ret)

		assert.Panics(t, func() {
			l.Remove(&n)
		})
	})
}

func testFrontBack(t *testing.T, constructor func(int) testList) {
	t.Run("Front/Back returns ok=false if empty list", func(t *testing.T) {
		l := constructor(1)
		_, ok := l.Back()
		assert.False(t, ok)

		_, ok = l.Front()
		assert.False(t, ok)
	})

	t.Run("Front/Back returns same element if list is of size 1", func(t *testing.T) {
		el := &TestComponent{Ch: 'C'}
		l := constructor(1)
		l.PushBack(el)

		front, ok := l.Front()
		assert.True(t, ok)
		back, ok := l.Back()
		assert.True(t, ok)

		assert.Equal(t, front, back)
	})

	t.Run("Back returns last element in list", func(t *testing.T) {
		l, els := makeListOfTwo()

		ret, ok := l.Back()
		assert.True(t, ok)
		assert.Equal(t, els[1], ret.Value())
	})

	t.Run("Front returns last element in list", func(t *testing.T) {
		l, els := makeListOfTwo()

		ret, ok := l.Front()
		assert.True(t, ok)
		assert.Equal(t, els[0], ret.Value())
	})
}

// bridge between testList, responsiveTestList and focusTestList
func getTestComponent(v tui.Component) *TestComponent {
	if a, ok := v.(*compWithAttr); ok {
		return a.Component.(*TestComponent)
	}
	if a, ok := v.(*testListResponsive); ok {
		return a.Component.(*TestComponent)
	}
	return v.(*TestComponent)
}

func sortTestComponent(a, b tui.Component) bool {
	return getTestComponent(a).Ch < getTestComponent(b).Ch
}

func assertEqualNodes(t *testing.T, expected int, l testList) {
	var i int
	for node, ok := l.Front(); ok; node, ok = node.Next() {
		i++
	}
	assert.Equal(t, expected, i)
}

func testListSort(t *testing.T, constructor func(int) testList) {
	t.Run("returns a copy of the list with same element height", func(t *testing.T) {
		l := constructor(1)

		// make sure returned list is correct
		for i := 0; i < 3; i++ {
			l.Sort(func(a, b tui.Component) bool { return false })
			assert.Equal(t, 1, l.ElementHeight())
		}
	})

	t.Run("returns a copy of the list sorted", func(t *testing.T) {
		l := constructor(1)
		l.PushBack(&TestComponent{Ch: 'b'})
		l.PushBack(&TestComponent{Ch: 'a'})
		l.PushBack(&TestComponent{Ch: 'c'})
		// make sure returned list is correct
		for range 3 {
			l.Sort(func(a, b tui.Component) bool {
				return getTestComponent(a).Ch < getTestComponent(b).Ch
			})
			node, ok := l.Front()
			require.True(t, ok)
			assert.Equal(t, 'a', getTestComponent(node.Value()).Ch)

			node, ok = node.Next()
			require.True(t, ok)
			assert.Equal(t, 'b', getTestComponent(node.Value()).Ch)

			node, ok = node.Next()
			require.True(t, ok)
			assert.Equal(t, 'c', getTestComponent(node.Value()).Ch)

			_, ok = node.Next()
			require.False(t, ok)
		}
	})

	t.Run("incremental sorting of pushed elements", func(t *testing.T) {
		l := constructor(1)

		node := l.PushBack(&TestComponent{Ch: 'z'})
		_, ok := node.Next()
		assert.False(t, ok)

		l.Sort(sortTestComponent)

		node, ok = l.Front()
		require.True(t, ok)
		assert.Equal(t, 'z', getTestComponent(node.Value()).Ch)
		assertEqualNodes(t, 1, l)

		node = l.PushBack(&TestComponent{Ch: 'y'})
		_, ok = node.Next()
		assert.False(t, ok)

		l.Sort(sortTestComponent)

		node, ok = l.Front()
		require.True(t, ok)
		assert.Equal(t, 'y', getTestComponent(node.Value()).Ch)

		node, ok = node.Next()
		require.True(t, ok)
		assert.Equal(t, 'z', getTestComponent(node.Value()).Ch)

		assertEqualNodes(t, 2, l)
	})

	t.Run("does not invalidate preaviously leaked nodes", func(t *testing.T) {
		l := constructor(1)

		z := l.PushBack(&TestComponent{Ch: 'z'})
		_, ok := z.Next()
		assert.False(t, ok)

		y := l.PushBack(&TestComponent{Ch: 'y'})
		_, ok = y.Next()
		assert.False(t, ok)

		_, ok = z.Next()
		assert.True(t, ok)

		l.Sort(sortTestComponent)

		_, ok = y.Next()
		assert.True(t, ok)

		_, ok = z.Next()
		assert.False(t, ok)
	})
}

func TestListMaxOffset(t *testing.T) {
	testListMaxOffset(t, newTestList)
}

func testListMaxOffset(t *testing.T, constructor func(int) testList) {
	tsuite := []struct {
		description   string
		height        int
		elementHeight int
		elements      int
		want          int
	}{
		{"enough height, zero elements with element height of 1", 10, 1, 0, 0},
		{"enough height, zero elements with element height of 2", 10, 2, 0, 0},
		{"enough height, one element with element height of 1", 10, 1, 1, 0},
		{"enough height, one element with element height of 2", 10, 2, 1, 0},
		{"limited height, one element with element height of 1", 1, 1, 1, 0},
		{"limited height < element height, one element with element height of 2", 1, 2, 1, 0},
		{"limited height < element height, two elements with element height of 1", 1, 2, 2, 1},
		{"limited height < element height, two elements with element height of 2", 1, 2, 2, 1},
		{"limited height, three elements with element height of 1", 1, 1, 3, 2},
		{"limited height < element height, three elements with element height of 2", 1, 2, 3, 2},
		{"limited height > element height, many elements with element height of 1", 10, 1, 30, 20},
		{"limited height > element height, many (even) elements with element height of 2", 10, 2, 30, 25},
		{"limited height > element height, many (odd) elements with element height of 2", 10, 2, 31, 26},
	}

	for _, tcase := range tsuite {
		t.Run(tcase.description, func(t *testing.T) {
			l := constructor(tcase.elementHeight)
			l.Resize(10, tcase.height)

			for i := 0; i < tcase.elements; i++ {
				l.PushBack(&TestComponent{})
			}

			require.Equal(t, tcase.elements, l.Len())
			assert.Equal(t, tcase.want, l.MaxOffset())
		})
	}
}

func TestListDraw(t *testing.T) {
	testListDraw(t, nil, newTestList)
}

func TestListSort(t *testing.T) {
	testListSort(t, newTestList)
}

func TestFrontBack(t *testing.T) {
	testFrontBack(t, newTestList)
}

func TestListRemove(t *testing.T) {
	testListRemove(t, newTestList)
}

func TestEmptyListDraw(t *testing.T) {
	testEmptyListDraw(t, newTestList)
}

func benchmarkListDraw(b *testing.B, n int) {
	l := NewList(1)
	l.Resize(1000, 1000)
	for i := 0; i < n; i++ {
		l.PushBack(NewString(strconv.Itoa(i)))
	}

	var w term.NoopWriter
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Draw(w)
	}
}

func benchmarkListSeekEndDraw(b *testing.B, n int) {
	l := NewList(1)
	l.Resize(1000, 1000)
	for i := 0; i < n; i++ {
		l.PushBack(NewString(strconv.Itoa(i)))
	}

	l.SeekEnd()
	var w term.NoopWriter
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Draw(w)
	}
}

func BenchmarkListDraw1(b *testing.B) {
	benchmarkListDraw(b, 1)
}

func BenchmarkListDraw10(b *testing.B) {
	benchmarkListDraw(b, 10)
}

func BenchmarkListDraw100(b *testing.B) {
	benchmarkListDraw(b, 100)
}

func BenchmarkListDraw1000(b *testing.B) {
	benchmarkListDraw(b, 1000)
}

func BenchmarkListDraw100000(b *testing.B) {
	benchmarkListDraw(b, 100000)
}

func BenchmarkListSeekEndDraw1(b *testing.B) {
	benchmarkListSeekEndDraw(b, 1)
}

func BenchmarkListSeekEndDraw10(b *testing.B) {
	benchmarkListSeekEndDraw(b, 10)
}

func BenchmarkListSeekEndDraw100(b *testing.B) {
	benchmarkListSeekEndDraw(b, 100)
}

func BenchmarkListSeekEndDraw1000(b *testing.B) {
	benchmarkListSeekEndDraw(b, 1000)
}

func BenchmarkListSeekEndDraw100000(b *testing.B) {
	benchmarkListSeekEndDraw(b, 100000)
}
