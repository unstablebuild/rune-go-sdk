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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

var (
	defaultAttr = term.Attributes{}
)

type compWithAttr struct {
	tui.Component
	attr term.Attributes
}

func (c *compWithAttr) SetAttr(attr term.Attributes) (ret term.Attributes) {
	ret = c.attr
	c.attr = attr
	return
}

func newCompWithAttr(c tui.Component) *compWithAttr {
	return &compWithAttr{Component: c, attr: defaultAttr}
}

type focusListTestList struct {
	*FocusList
}

func (l *focusListTestList) MaxOffset() int {
	return l.list.MaxOffset()
}

func (l *focusListTestList) SeekDown() (ok bool) {
	if l.list.SeekDown() {
		l.focusIdx--
		ok = true
	}
	return
}

func (l *focusListTestList) SeekUp() (ok bool) {
	if l.list.SeekUp() {
		l.focusIdx++
		ok = true
	}
	return
}

func (l *focusListTestList) SeekEnd() (ok bool) {
	for l.SeekDown() {
		ok = true
	}
	return
}

func (l *focusListTestList) SeekStart() (ok bool) {
	for l.SeekUp() {
		ok = true
	}
	return
}

func (l *focusListTestList) CanSeekDown() bool {
	return l.list.CanSeekDown()
}

func (l *focusListTestList) CanSeekUp() bool {
	return l.list.CanSeekUp()
}

func (l *focusListTestList) PushBackList(other testList) {
	l.FocusList.PushBackList(other.(*focusListTestList).FocusList)
}
func (l *focusListTestList) PushFrontList(other testList) {
	l.FocusList.PushFrontList(other.(*focusListTestList).FocusList)
}

func (l *focusListTestList) PushBack(c tui.Component) ListNode {
	return l.FocusList.PushBack(newCompWithAttr(c))
}

func (l *focusListTestList) PushFront(c tui.Component) ListNode {
	return l.FocusList.PushFront(newCompWithAttr(c))
}

func (l *focusListTestList) Remove(e *ListNode) tui.Component {
	return l.FocusList.Remove(e)
}

func (l *focusListTestList) Sort(less func(a, b tui.Component) bool) {
	l.FocusList.Sort(func(a, b tui.Component) bool {
		return less(a.(*compWithAttr).Component, b.(*compWithAttr).Component)
	})
}

func newFocusTestList(i int) testList {
	ret := &focusListTestList{FocusList: NewFocusList()}
	ret.SetElementHeight(i)
	return ret
}

func TestFocusListFocus(t *testing.T) {
	t.Run("Focus returns false if list is empty", func(t *testing.T) {
		l := NewFocusList()
		_, ok := l.Focus()
		assert.False(t, ok)
	})

	t.Run("Focus returns true and the focus of the list", func(t *testing.T) {
		el1 := &compWithAttr{&TestComponent{Ch: 'c'}, defaultAttr}
		el2 := &compWithAttr{&TestComponent{Ch: 'a'}, defaultAttr}
		l := NewFocusList()

		l.PushFront(el1)
		n, ok := l.Focus()
		assert.True(t, ok)
		assert.Equal(t, n.Value(), el1)

		l.PushFront(el2)
		n, ok = l.Focus()
		assert.True(t, ok)
		assert.Equal(t, n.Value(), el1)

		l.FocusUp()
		n, ok = l.Focus()
		assert.True(t, ok)
		assert.Equal(t, n.Value(), el2)
	})

	t.Run("focus sequence", func(t *testing.T) {
		l := NewFocusList()
		require.False(t, l.CanFocusDown())
		require.False(t, l.CanFocusUp())

		l.PushBack(newCompWithAttr(&TestComponent{}))
		require.False(t, l.CanFocusDown())
		require.False(t, l.CanFocusUp())

		l.PushBack(newCompWithAttr(&TestComponent{}))
		require.True(t, l.CanFocusDown())
		require.False(t, l.CanFocusUp())

		l.PushFront(newCompWithAttr(&TestComponent{}))
		require.True(t, l.CanFocusDown())
		require.True(t, l.CanFocusUp())

		l.FocusUp()
		require.True(t, l.CanFocusDown())
		require.False(t, l.CanFocusUp())

		l.FocusEnd()
		require.False(t, l.CanFocusDown())
		require.True(t, l.CanFocusUp())

		l.FocusStart()
		require.True(t, l.CanFocusDown())
		require.False(t, l.CanFocusUp())

		l.FocusDown()
		// resize should not affect
		l.Resize(math.MaxInt32, math.MaxInt32)

		require.True(t, l.CanFocusDown())
		require.True(t, l.CanFocusUp())
	})
}

func TestSetFocus(t *testing.T) {
	tsuite := []struct {
		description string
		sut         func(*testing.T, *FocusList)
	}{
		{
			"extraneous ListNode should panic",
			func(t *testing.T, l *FocusList) {
				assert.Panics(t, func() {
					l.SetFocus(ListNode{})
				})
			},
		},
		{
			"setting current focus is no-op",
			func(t *testing.T, l *FocusList) {
				node := l.PushBack(newCompWithAttr(&TestComponent{}))
				actual, ok := l.Focus()
				require.True(t, ok)
				assert.Equal(t, node, actual)
				assert.Equal(t, 0, l.FocusOffset())

				l.SetFocus(node)
				actual, ok = l.Focus()
				require.True(t, ok)
				assert.Equal(t, node, actual)

				assert.Equal(t, 0, l.FocusOffset())
			},
		},
		{
			"set next",
			func(t *testing.T, l *FocusList) {
				node1 := l.PushBack(newCompWithAttr(&TestComponent{}))
				node2 := l.PushBack(newCompWithAttr(&TestComponent{}))

				actual, ok := l.Focus()
				require.True(t, ok)
				assert.Equal(t, node1, actual)

				l.SetFocus(node2)
				actual, ok = l.Focus()
				require.True(t, ok)
				assert.Equal(t, node2, actual)

				assert.Equal(t, 1, l.FocusOffset())
			},
		},
		{
			"set prev",
			func(t *testing.T, l *FocusList) {
				node1 := l.PushBack(newCompWithAttr(&TestComponent{}))
				node2 := l.PushBack(newCompWithAttr(&TestComponent{}))
				require.True(t, l.FocusDown())
				assert.Equal(t, 1, l.FocusOffset())

				actual, ok := l.Focus()
				require.True(t, ok)
				assert.Equal(t, node2, actual)

				l.SetFocus(node1)
				actual, ok = l.Focus()
				require.True(t, ok)
				assert.Equal(t, node1, actual)

				assert.Equal(t, 0, l.FocusOffset())
			},
		},
	}

	for _, tcase := range tsuite {
		t.Run(tcase.description, func(t *testing.T) {
			l := NewFocusList()
			tcase.sut(t, l)
		})
	}
}

func TestFocusListDraw(t *testing.T) {
	testListDraw(t, nil, newFocusTestList)
}

func TestFocusListSort(t *testing.T) {
	testListSort(t, newFocusTestList)

	t.Run("focus and seeks to start of list upon call Sort", func(t *testing.T) {
		l := NewFocusList()
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'z'}))
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'x'}))
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'a'}))
		l.Resize(1, 1)
		require.True(t, l.FocusEnd())
		require.True(t, l.CanFocusUp())
		require.False(t, l.CanFocusDown())

		l.Sort(func(a, b tui.Component) bool {
			return a.(*compWithAttr).Component.(*TestComponent).Ch <
				b.(*compWithAttr).Component.(*TestComponent).Ch
		})

		assert.False(t, l.CanFocusUp())
		assert.True(t, l.CanFocusDown())
	})
}

func TestIterateVisible(t *testing.T) {
	tsuite := []struct {
		desc     string
		op       func(*testing.T, *FocusList)
		expected string
	}{
		{"at start", func(*testing.T, *FocusList) {}, "ab"},
		{"middle", func(t *testing.T, l *FocusList) {
			assert.True(t, l.FocusDown())
		}, "bc"},
		{"end ", func(t *testing.T, l *FocusList) {
			assert.True(t, l.FocusEnd())
		}, "cd"},
	}
	for _, tcase := range tsuite {
		t.Run(tcase.desc, func(t *testing.T) {
			l := NewFocusList()
			l.PushBack(newCompWithAttr(&TestComponent{Ch: 'a'}))
			l.PushBack(newCompWithAttr(&TestComponent{Ch: 'b'}))
			l.PushBack(newCompWithAttr(&TestComponent{Ch: 'c'}))
			l.PushBack(newCompWithAttr(&TestComponent{Ch: 'd'}))
			l.Resize(2, 2)

			tcase.op(t, l)

			var actualRunes []rune
			l.IterateVisible(func(c tui.Component) {
				actualRunes = append(actualRunes, c.(*compWithAttr).Component.(*TestComponent).Ch)
			})
			assert.Equal(t, tcase.expected, string(actualRunes))
		})
	}
}

func TestFocusListFrontBack(t *testing.T) {
	testFrontBack(t, newFocusTestList)
}

func TestFocusListListRemove(t *testing.T) {
	testListRemove(t, newFocusTestList)
}

func TestFocusListMaxOffset(t *testing.T) {
	testListMaxOffset(t, newFocusTestList)
}

func TestFocusListEmptyDraw(t *testing.T) {
	testEmptyListDraw(t, newFocusTestList)
}

func makeFocusList(n int) *FocusList {
	ret := NewFocusList()
	for i := 0; i < n; i++ {
		ret.PushBack(newCompWithAttr(&TestComponent{Ch: rune(i)}))
	}
	// set to a sensible size
	ret.Resize(100, 50)
	return ret
}

func TestFocusListRemove(t *testing.T) {
	t.Run("remove twice the same node", func(t *testing.T) {
		list := makeFocusList(3)
		ok := list.FocusStart()
		require.True(t, ok)

		node, ok := list.Focus()
		require.True(t, ok)

		list.Remove(&node)

		assert.Panics(t, func() {
			list.Remove(&node)
		})
	})

	t.Run("remove focus node that is the first node of the list", func(t *testing.T) {
		l := NewFocusList()
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'a'}))
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'b'}))
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'c'}))

		ok := l.FocusStart()
		require.True(t, ok)

		node, ok := l.Focus()
		require.True(t, ok)
		require.Equal(t, 'a', node.Value().(*compWithAttr).Component.(*TestComponent).Ch)

		l.Remove(&node)
		require.Equal(t, 2, l.Len())

		node, ok = l.Focus()
		require.True(t, ok)

		assert.Equal(t, 'b', node.Value().(*compWithAttr).Component.(*TestComponent).Ch)
	})

	t.Run("remove focus node that is the last node of the list", func(t *testing.T) {
		l := NewFocusList()
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'a'}))
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'b'}))
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'c'}))

		ok := l.FocusEnd()
		require.True(t, ok)

		node, ok := l.Focus()
		require.True(t, ok)
		require.Equal(t, 'c', node.Value().(*compWithAttr).Component.(*TestComponent).Ch)

		l.Remove(&node)
		require.Equal(t, 2, l.Len())

		node, ok = l.Focus()
		require.True(t, ok)

		require.Equal(t, 'b', node.Value().(*compWithAttr).Component.(*TestComponent).Ch)
	})

	t.Run("remove the only node in the list and call focus", func(t *testing.T) {
		l := NewFocusList()
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'a'}))

		ok := l.FocusStart()
		require.True(t, ok)

		node, ok := l.Focus()
		require.True(t, ok)
		require.Equal(t, 'a', node.Value().(*compWithAttr).Component.(*TestComponent).Ch)

		l.Remove(&node)

		_, ok = l.Focus()
		assert.False(t, ok)
	})

	t.Run("setting focus on a removed node panics", func(t *testing.T) {
		l := NewFocusList()
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'a'}))

		ok := l.FocusStart()
		require.True(t, ok)

		node, ok := l.Focus()
		require.True(t, ok)
		require.Equal(t, 'a', node.Value().(*compWithAttr).Component.(*TestComponent).Ch)

		l.Remove(&node)

		assert.Panics(t, func() {
			l.SetFocus(node)
		})
	})

	t.Run("remove on a node in the middle of the list then set focus on a node before the node that was removed", func(t *testing.T) {
		l := NewFocusList()
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'a'}))
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'b'}))
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'c'}))
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'd'}))
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'e'}))

		ok := l.FocusStart() // 'a'
		require.True(t, ok)

		ok = l.FocusDown() // 'b'
		require.True(t, ok)
		nodeB, ok := l.Focus()
		require.True(t, ok)

		ok = l.FocusDown() // 'c'
		require.True(t, ok)

		nodeC, ok := l.Focus()
		require.True(t, ok)
		require.Equal(t, 'c', nodeC.Value().(*compWithAttr).Component.(*TestComponent).Ch)

		l.Remove(&nodeC)
		require.Equal(t, 4, l.Len())

		l.SetFocus(nodeB)
		nodeB, ok = l.Focus()
		require.True(t, ok)
		assert.Equal(t, 'b', nodeB.Value().(*compWithAttr).Component.(*TestComponent).Ch)
	})

	t.Run("remove on a node in the middle of the list then set focus on a node after the node that was removed", func(t *testing.T) {
		l := NewFocusList()
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'a'}))
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'b'}))
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'c'}))
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'd'}))
		l.PushBack(newCompWithAttr(&TestComponent{Ch: 'e'}))

		ok := l.FocusEnd() // 'e'
		require.True(t, ok)
		nodeD, ok := l.Focus()
		require.True(t, ok)

		ok = l.FocusStart() // 'a'
		require.True(t, ok)

		ok = l.FocusDown() // 'b'
		require.True(t, ok)

		ok = l.FocusDown() // 'c'
		require.True(t, ok)
		nodeC, ok := l.Focus()
		require.True(t, ok)
		require.Equal(t, 'c', nodeC.Value().(*compWithAttr).Component.(*TestComponent).Ch)

		l.Remove(&nodeC)
		require.Equal(t, 4, l.Len())

		l.SetFocus(nodeD)
		nodeD, ok = l.Focus()
		require.True(t, ok)
		assert.Equal(t, 'e', nodeD.Value().(*compWithAttr).Component.(*TestComponent).Ch)
	})

}

func benchmarkSetFocus(b *testing.B, n int) {
	b.Run("toggle focus between first and middle", func(b *testing.B) {
		l := makeFocusList(n)
		i := 0
		front, _ := l.Front()
		var middle ListNode
		for node, ok := l.Front(); ok; node, ok = node.Next() {
			if n/2 == i {
				middle = node
				break
			}
			i++
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.SetFocus(middle)
			l.SetFocus(front)
		}
	})
	b.Run("toggle focus between first and second", func(b *testing.B) {
		l := makeFocusList(n)
		front, _ := l.Front()
		focus, _ := front.Next()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.SetFocus(focus)
			l.SetFocus(front)
		}
	})
	b.Run("toggle focus between first and last", func(b *testing.B) {
		l := makeFocusList(n)
		front, _ := l.Front()
		focus, _ := l.Back()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.SetFocus(focus)
			l.SetFocus(front)
		}
	})
	b.Run("toggle focus between last and second to last", func(b *testing.B) {
		l := makeFocusList(n)
		last, _ := l.Back()
		secondLast, _ := last.Prev()
		l.SetFocus(secondLast)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.SetFocus(last)
			l.SetFocus(secondLast)
		}
	})
	b.Run("toggle focus between middle and last", func(b *testing.B) {
		l := makeFocusList(n)
		last, _ := l.Back()
		var middle ListNode
		i := 0
		for node, ok := l.Front(); ok; node, ok = node.Next() {
			if n/2 == i {
				middle = node
				break
			}
			i++
		}
		l.SetFocus(middle)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.SetFocus(last)
			l.SetFocus(middle)
		}
	})
}

func BenchmarkSetFocus10(b *testing.B) {
	benchmarkSetFocus(b, 10)
}
func BenchmarkSetFocus100(b *testing.B) {
	benchmarkSetFocus(b, 100)
}
func BenchmarkSetFocus1000(b *testing.B) {
	benchmarkSetFocus(b, 1000)
}
func BenchmarkSetFocus1000000(b *testing.B) {
	benchmarkSetFocus(b, 1000000)
}
