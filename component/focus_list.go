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
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
	"github.com/unstablebuild/tcell/v3"
)

const defaultElementHeight = 1

var (
	defaultTextAttr   = term.Attributes{}
	highlightTextAttr = term.Attributes{
		Attrs: tcell.AttrBold,
	}
)

// FocusList wraps a List to provide an element Focus. It takes tui.Component
// components.
type FocusList struct {
	Inverted      bool
	list          *List
	focus         ListNode
	focusIdx      int
	height, width int
	textAttr      term.Attributes
	focusAttr     term.Attributes
}

// NewFocusList allocates storage for a new FocusList and initializes it.
func NewFocusList() *FocusList {
	ret := new(FocusList)
	ret.Init()
	return ret
}

// Init initializes this FocusList with the default element height of 1.
func (l *FocusList) Init() {
	l.InitWithAttr(defaultTextAttr, highlightTextAttr)
}

// InitWithAttr initializes this FocusList with the default element height of 1,
// and text as the text attributes and focus as the focus attributes.
func (l *FocusList) InitWithAttr(text, focus term.Attributes) {
	l.list = NewList(defaultElementHeight)
	l.focus = ListNode{}
	l.focusIdx = 0
	l.textAttr = text
	l.focusAttr = focus
}

func (l *FocusList) switchFocus(newFocus ListNode) {
	l.focus = newFocus
}

func (l *FocusList) trySetFirstFocus(node ListNode) bool {
	if l.focus.Value() == nil {
		l.switchFocus(node)
		return true
	}
	return false
}

// SetFocusAttr sets the attributes of the focus nodes.
// The current focus node is changed and any future focused
// nodes will inherit the given attr.
func (l *FocusList) SetFocusAttr(attr term.Attributes) {
	l.focusAttr = attr
	l.switchFocus(l.focus)
}

// SetFocus sets the focus of this FocusList to node.
func (l *FocusList) SetFocus(node ListNode) {
	if node.l != l.list {
		panic("ListNode does not belong to FocusList")
	}

	defer l.switchFocus(node)

	// In order to find the new focusIdx, first attempt to scan
	// the current view to optimize SetFocus for calls with a node currently
	// rendered on screen.
	idx := l.focusIdx
	for n, ok := l.focus.Next(); ok; n, ok = n.Next() {
		idx++
		if n == node {
			l.focusIdx = idx
			return
		}
		// Ignore element height and use height as a vague representation
		// of the current view over this FocusList.
		if idx > l.focusIdx+l.height {
			break
		}
	}

	idx = l.focusIdx
	for n, ok := l.focus.Prev(); ok; n, ok = n.Prev() {
		idx--
		if n == node {
			l.focusIdx = idx
			return
		}
		if idx < l.focusIdx-l.height {
			break
		}
	}

	// fallback to iterating entire list
	l.focusIdx = 0
	for n, ok := l.list.Front(); ok; n, ok = n.Next() {
		if n == node {
			return
		}
		l.focusIdx++
	}

	panic("could not finde ListNode in FocusList")
}

// Reset resets the contents of this FocusList.
func (l *FocusList) Reset() {
	l.switchFocus(ListNode{})
	l.focusIdx = 0
	l.list.Reset()
}

// Back returns the last node of list l or nil if the list is empty.
func (l *FocusList) Back() (ListNode, bool) {
	n, ok := l.list.Back()
	return n, ok
}

// Front returns the first node of list l or nil if the list is empty.
func (l *FocusList) Front() (ListNode, bool) {
	n, ok := l.list.Front()
	return n, ok
}

// Draw satisfies tui.Component
func (l *FocusList) Draw(w term.Writer) {
	l.list.Draw(w)
	if l.focus.el != nil {
		focus := l.focus.el.Value.(*Virtual[tui.Component])
		at := focus.Position()
		for x := range focus.Width() {
			at.X = x
			w.UnionAttributes(at, l.focusAttr)
		}
	}
}

// ElementAt returns the element at pos Coordinates or panics if
// coordinates are out of the bounds of this List.
func (l *FocusList) ElementAt(pos term.Coordinates) (ListNode, bool) {
	node, ok := l.list.ElementAt(pos)
	return node, ok
}

// ElementHeight returns the height for each element of this list.
func (l *FocusList) ElementHeight() int {
	return l.list.ElementHeight()
}

// Len returns the number of nodes of list l in O(1).
func (l *FocusList) Len() int {
	return l.list.Len()
}

// PushBack inserts a new element c at the back of list l and
// returns the linked node.
func (l *FocusList) PushBack(c tui.Component) ListNode {
	n := l.list.PushBack(c)
	l.trySetFirstFocus(n)
	return n
}

// PushBackList inserts a copy of an other list at the back of list l. The
// lists l and other must NOT be the same or nil.
func (l *FocusList) PushBackList(other *FocusList) {
	other.Iterate(func(c tui.Component) {
		l.PushBack(c)
	})
}

// PushFront inserts a new element c with value v at the front of list l and
// returns e.
func (l *FocusList) PushFront(c tui.Component) ListNode {
	n := l.list.PushFront(c)
	if !l.trySetFirstFocus(n) {
		l.focusIdx++
	}
	return n
}

// PushFrontList inserts a copy of an other list at the front of list l. The
// lists l and other must NOT be the same or nil.
func (l *FocusList) PushFrontList(other *FocusList) {
	for node, ok := other.Back(); ok; node, ok = node.Prev() {
		l.PushFront(node.Value())
	}
}

// Remove removes a node.
func (l *FocusList) Remove(n *ListNode) tui.Component {
	// if deleting what's under the cursor try shifting it.
	if l.focus == *n {
		if ok := l.FocusDown(); !ok {
			if ok := l.FocusUp(); !ok {
				// if we reached this point most likely we are removing the only element
				// in the list or something strange, so let's play safe and put the focus
				// away
				l.focus = ListNode{}
				l.focusIdx = 0
			}
		}
	}
	ret := l.list.Remove(n)
	return ret
}

// Iterate iterates over all elements in l.
func (l *FocusList) Iterate(fn func(tui.Component)) {
	for node, ok := l.list.Front(); ok; node, ok = node.Next() {
		fn(node.Value())
	}
}

// IterateVisible iterates only the visible elements in l.
func (l *FocusList) IterateVisible(fn func(tui.Component)) {
	offset := l.list.Offset()
	lastVisible := l.height/l.list.ElementHeight() + offset
	node, ok := l.list.Head()
	for i := offset; ok && i < lastVisible; i++ {
		fn(node.Value())
		node, ok = node.Next()
	}
}

// Resize satisfies tui.Compontent
func (l *FocusList) Resize(width, height int) {
	l.height, l.width = height, width
	l.list.Resize(width, height)
}

// SetElementHeight sets the height for each element of this list.
func (l *FocusList) SetElementHeight(height int) {
	l.list.SetElementHeight(height)
}

// FocusDown sets the focus to the node after the current focus
// and returns true, or if the focus is already the last node,
// it does nothing and returns false.
func (l *FocusList) FocusDown() bool {
	next, ok := l.focus.Next()
	if ok {
		l.focusIdx++
		l.switchFocus(next)
		if l.focusIdx-l.list.Offset() >= l.height-1 {
			l.list.SeekDown()
		}
	}
	return ok
}

// CanFocusDown returns true if FocusDown would return true.
func (l *FocusList) CanFocusDown() bool {
	_, ok := l.focus.Next()
	return ok
}

// CanFocusUp returns true if FocusUp would return true.
func (l *FocusList) CanFocusUp() bool {
	_, ok := l.focus.Prev()
	return ok
}

// FocusUp sets the focus to the node before the current focus
// and returns true, or if the focus is already the first node,
// it does nothing and returns false.
func (l *FocusList) FocusUp() bool {
	prev, ok := l.focus.Prev()
	if ok {
		l.focusIdx--
		l.switchFocus(prev)
		if l.focusIdx-l.list.Offset() <= 0 {
			l.list.SeekUp()
		}
	}
	return ok
}

// FocusStart sets the focus to the first node of l.
func (l *FocusList) FocusStart() (ok bool) {
	front, ok := l.Front()
	if !ok {
		return ok
	}
	l.list.SeekStart()
	l.focusIdx = 0
	l.switchFocus(front)
	return ok
}

// FocusEnd sets the focus to the last node of l.
func (l *FocusList) FocusEnd() (ok bool) {
	back, ok := l.Back()
	if !ok {
		return ok
	}
	l.list.SeekEnd()
	l.focusIdx = l.list.Len() - 1
	l.switchFocus(back)
	return ok
}

// Focus returns the current node in focus.
func (l *FocusList) Focus() (ListNode, bool) {
	if l.focus.Value() == nil {
		return ListNode{}, false
	}
	return l.focus, true
}

// Offset returns this list's current seek offset.
func (l *FocusList) Offset() int {
	return l.list.Offset()
}

// FocusOffset returns this list's focus index in the underlying list.
func (l *FocusList) FocusOffset() int {
	return l.focusIdx
}

// Sort sorts the elements of this list with the provided less function.
// It also resets the current focus node, according to the Inverted
// property in FocusList.
func (l *FocusList) Sort(less func(a, b tui.Component) bool) {
	l.list.Sort(func(a, b tui.Component) bool {
		return less(a, b)
	})

	if l.Inverted {
		l.FocusEnd()
	} else {
		l.FocusStart()
	}
}
