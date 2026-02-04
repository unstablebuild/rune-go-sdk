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
	"container/list"
	"math"
	"sort"

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// ListNode is an element of a component List.
type ListNode struct {
	l  *List
	el *list.Element
}

// List satisfies tui.Component by drawing a set of
// children components with a fixed height and a width
// that's equal to the maximum width available for this component.
//
// The height of the components can be set via the constructors and
// it can be modified later with SetElementHeight.
type List struct {
	elementHeight int
	// API indirectly exposes internal list so we need
	// to make sure that the elements are properly position and sized
	// before drawing
	dirty         bool
	width, height int
	offset        struct {
		value int
		head  ListNode
	}
	list list.List // list of Virtual
}

// NewList allocates storage for a new List and initializes it.
func NewList(elementHeight int) (l *List) {
	l = new(List)
	l.Init(elementHeight)
	return l
}

func (l *List) listNode(el *list.Element) ListNode {
	return ListNode{
		l:  l,
		el: el,
	}
}

func (l *List) setMaxOffset() {
	l.offset.value = l.MaxOffset()
	l.offset.head, _ = l.Back()
	for i := 1; i < l.height/l.elementHeight; i++ {
		l.offset.head, _ = l.offset.head.Prev()
	}
}

/* offset.value and offset.head are out of sync
   because an element has been inserted or removed before offset.head */

func (l *List) findNodeBeforeOffset(node ListNode) bool {
	for n, ok := l.offset.head, true; ok; n, ok = n.Prev() {
		if n == node {
			return true
		}
	}
	return false
}

func (l *List) fixOffsetRemove(removed, nextAfterOffsetHead ListNode) {
	if l.offset.value > l.MaxOffset() {
		l.setMaxOffset()
		return
	}
	if removed == l.offset.head {
		if nextAfterOffsetHead.el == nil {
			l.setMaxOffset()
			return
		}
		l.offset.head = nextAfterOffsetHead
		return
	}
	found := l.findNodeBeforeOffset(removed)
	if found {
		l.offset.head, _ = l.offset.head.Next()
	}
}

func (l *List) fixOffsetAdd(node ListNode) {
	found := l.findNodeBeforeOffset(node)
	if found {
		l.offset.head, _ = l.offset.head.Prev()
	}
}

// Reset resets the contents of this List.
func (l *List) Reset() {
	l.list.Init()
	l.offset.value = 0
	l.offset.head, _ = l.Front()
}

// Init initializes this List.
func (l *List) Init(elementHeight int) {
	if elementHeight <= 0 {
		panic("element height cannot be smaller than or equal to 0")
	}
	l.elementHeight = elementHeight
	l.list.Init()
}

// SetElementHeight sets the height for each element of this list.
func (l *List) SetElementHeight(height int) {
	l.elementHeight = height
	l.dirty = true
}

// ElementHeight returns the height for each element of this list.
func (l *List) ElementHeight() int {
	return l.elementHeight
}

// CanSeekUp returns whether SeekUp would seek one row up.
func (l *List) CanSeekUp() bool {
	return l.offset.value > 0
}

// Offset returns this list's current seek offset.
func (l *List) Offset() int {
	return l.offset.value
}

// Head returns the first node of list l, starting at the current seek offset
// or false if the list is empty.
func (l *List) Head() (ListNode, bool) {
	return l.offset.head, l.Len() > 0
}

// MaxOffset returns this list's max seek offset.
func (l *List) MaxOffset() int {
	if l.height == 0 {
		return 0
	}
	denom := l.elementHeight
	if denom > l.height {
		denom = l.height
	}
	return int(math.Max(0, float64(l.list.Len()-l.height/denom)))
}

// CanSeekDown returns whether SeekUp would seek one row down.
func (l *List) CanSeekDown() bool {
	return l.offset.value < l.MaxOffset()
}

// SeekUp shifts the contents of this list one row up.
func (l *List) SeekUp() bool {
	ok := l.CanSeekUp()
	if ok {
		l.offset.value--
		l.offset.head, _ = l.offset.head.Prev()
		l.dirty = true
	}
	return ok
}

// SeekDown shifts the contents of this list one row down.
func (l *List) SeekDown() bool {
	ok := l.CanSeekDown()
	if ok {
		l.offset.value++
		l.offset.head, _ = l.offset.head.Next()
		l.dirty = true
	}
	return ok
}

// SeekEnd shifts the contents of this list such that the last element
// is drawn at the top of the list.
func (l *List) SeekEnd() (ok bool) {
	if l.CanSeekDown() {
		l.setMaxOffset()
		ok = true
		l.dirty = true
	}
	return
}

// SeekStart shifts the contents of this list such that the first element
// is drawn at the top of the list.
func (l *List) SeekStart() (ok bool) {
	if l.offset.value != 0 {
		ok = true
		l.offset.value = 0
		l.offset.head, _ = l.Front()
		l.dirty = true
	}
	return
}

// Resize resizes this list to fit within width and height.
func (l *List) Resize(width, height int) {
	l.width, l.height = width, height
	lastVisible := l.height/l.elementHeight + l.offset.value
	for i, el := l.offset.value, l.offset.head.el; i < lastVisible && el != nil; el, i = el.Next(), i+1 {
		comp := el.Value.(*Virtual[tui.Component])
		comp.Resize(l.width, l.elementHeight)
		ypos := (i - l.offset.value) * l.elementHeight
		comp.Move(term.Coordinates{X: 0, Y: ypos})
	}
	l.dirty = false
}

// Draw draws this list's elements with the current seek offset.
func (l *List) Draw(w term.Writer) {
	if l.dirty {
		l.Resize(l.width, l.height)
	}

	lastVisible := l.height/l.elementHeight + l.offset.value
	for i, el := l.offset.value, l.offset.head.el; i < lastVisible && el != nil; i, el = i+1, el.Next() {
		el.Value.(*Virtual[tui.Component]).Draw(w)
	}
}

// Back returns the last node of list l or false if the list is empty.
func (l *List) Back() (ListNode, bool) {
	b := l.list.Back()
	if b == nil {
		return ListNode{}, false
	}
	return l.listNode(b), true
}

// Front returns the first node of list l or false if the list is empty.
func (l *List) Front() (ListNode, bool) {
	f := l.list.Front()
	if f == nil {
		return ListNode{}, false
	}
	return l.listNode(f), true
}

// Len returns the number of nodes of list l in O(1).
func (l *List) Len() int { return l.list.Len() }

// PushBackList inserts a copy of an other list at the back of list l. The
// lists l and other must NOT be the same or nil.
func (l *List) PushBackList(other *List) {
	if l == other {
		panic("other list cannot be self: components can't be deep cloned")
	}
	l.list.PushBackList(&other.list)
	l.dirty = true
	if l.Len() == other.Len() {
		l.offset.value = 0
		l.offset.head, _ = l.Front()
	}
}

// PushFrontList inserts a copy of an other list at the front of list l. The
// lists l and other must NOT be the same or nil.
func (l *List) PushFrontList(other *List) {
	if l == other {
		panic("other list cannot be self: components can't be deep cloned")
	}
	l.list.PushFrontList(&other.list)
	l.dirty = true
	if l.Len() == other.Len() {
		l.offset.value = 0
		l.offset.head, _ = l.Front()
	} else {
		l.offset.value += other.list.Len()
	}
}

// InsertAfter inserts a new element c immediately after mark and
// returns the linked node. If mark is not an element of l, the list
// is not modified.
func (l *List) InsertAfter(c tui.Component, mark ListNode) ListNode {
	if mark.l != l {
		panic("node to insert not belonging to this list")
	}
	v := &Virtual[tui.Component]{C: c}
	l.dirty = true
	l.fixOffsetAdd(mark)
	return l.listNode(l.list.InsertAfter(v, mark.el))
}

// InsertBefore inserts a new element c immediately before mark
// and returns the linked node. If mark is not an element of l,
// the list is not modified.
func (l *List) InsertBefore(c tui.Component, mark ListNode) ListNode {
	if mark.l != l {
		panic("node to insert not belonging to this list")
	}
	v := &Virtual[tui.Component]{C: c}
	l.dirty = true
	l.fixOffsetAdd(mark)
	return l.listNode(l.list.InsertBefore(v, mark.el))
}

// PushBack inserts a new element c at the back of list l and
// returns the linked node.
func (l *List) PushBack(c tui.Component) ListNode {
	v := &Virtual[tui.Component]{C: c}
	l.dirty = true
	ret := l.listNode(l.list.PushBack(v))
	if l.Len() == 1 {
		l.offset.value = 0
		l.offset.head = ret
	}
	return ret
}

// PushFront inserts a new element c at the front of list l and
// returns the linked node.
func (l *List) PushFront(c tui.Component) ListNode {
	v := &Virtual[tui.Component]{C: c}
	l.dirty = true
	ret := l.listNode(l.list.PushFront(v))
	if l.Len() == 1 {
		l.offset.value = 0
		l.offset.head = ret
	} else {
		l.fixOffsetAdd(ret)
	}
	return ret
}

// Remove removes e from l if e is a node of list l.
// It returns the element value e.Value.
func (l *List) Remove(e *ListNode) tui.Component {
	if e.l != l {
		panic("node to remove not belonging to this list")
	}
	l.dirty = true
	nextAfterOffsetHead, _ := l.offset.head.Next()
	ret := l.list.Remove(e.el).(*Virtual[tui.Component]).C
	if l.Len() == 0 {
		l.offset.value = 0
		l.offset.head = ListNode{}
	} else {
		l.fixOffsetRemove(*e, nextAfterOffsetHead)
	}
	e.l = nil
	return ret
}

// SetValue sets the tui.Component value in ListNode.
func (e ListNode) SetValue(c tui.Component) {
	if e.el == nil || e.el.Value == nil {
		panic("trying to SetValue on an un-linked ListNode")
	}
	v := e.el.Value.(*Virtual[tui.Component])
	v.C = c
	// C needs resize
	e.l.dirty = true
}

// Value gets the tui.Component value in ListNode.
func (e ListNode) Value() tui.Component {
	if e.el == nil || e.el.Value == nil {
		return nil
	}
	return e.el.Value.(*Virtual[tui.Component]).C
}

// Prev gets the previous linked node in the list before e.
func (e ListNode) Prev() (ListNode, bool) {
	if e.el == nil || e.el.Prev() == nil {
		return ListNode{}, false
	}
	return e.l.listNode(e.el.Prev()), true
}

// Next gets the next linked node in the list after e.
func (e ListNode) Next() (ListNode, bool) {
	if e.el == nil || e.el.Next() == nil {
		return ListNode{}, false
	}
	return e.l.listNode(e.el.Next()), true
}

// ElementAt returns the element at pos Coordinates or panics if
// coordinates are out of the bounds of this List.
func (l *List) ElementAt(pos term.Coordinates) (ListNode, bool) {
	if pos.X < 0 || pos.Y < 0 {
		panic("negative coordinates")
	}

	el, ok := l.offset.head, l.offset.head != (ListNode{})
	if !ok {
		return ListNode{}, ok
	}

	i := 0
	y := pos.Y
	ok = true
	for ok {
		if i*l.elementHeight+l.elementHeight > y {
			return el, true
		}
		el, ok = el.Next()
		i++
	}

	return ListNode{}, false
}

// Sort in-place sorts this List with the provided less function.
// It resets the current seek offset, if any.
func (l *List) Sort(less func(a, b tui.Component) bool) {
	els := make([]ListNode, l.Len())
	i := 0
	for node, ok := l.Front(); ok; node, ok = node.Next() {
		els[i] = node
		i++
	}

	sort.Slice(els, func(i, j int) bool {
		a := els[i].Value()
		b := els[j].Value()
		return less(a, b)
	})

	for i := l.Len() - 1; i >= 0; i-- {
		l.list.MoveToFront(els[i].el)
	}

	l.offset.value = 0
	l.offset.head, _ = l.Front()
	l.dirty = true
}

// Height returns the height of this List, in the last call to Resize.
func (l *List) Height() int {
	return l.height
}

// Width returns the width of this List, in the last call to Resize.
func (l *List) Width() int {
	return l.width
}
