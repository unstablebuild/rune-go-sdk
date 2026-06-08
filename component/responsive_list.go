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

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// ResponsiveList differs from List in that it enables clients to
// set the min width and height of the children components by using a push
// API with Responsive components, rather than tui.Components.
//
// It also seeks in rows rather than elements, so long elements
// greater than height can be seeked top to bottom, one row at a time.
//
// Alignment responds to AlignmentTop or AlignmentBottom for
// content alignment.
type ResponsiveList struct {
	Alignment

	list List

	// in number of rows
	totalHeight int
	offset      int

	// cannot be used directly here otherwise we would need to somehow
	// override ListNode to also set true this ResponsiveList's
	// returned ListNode, to set dirty on SetValue
	// dirty bool
}

// NewResponsiveList allocates storage for a new ResponsiveList and initializes it.
func NewResponsiveList() (l *ResponsiveList) {
	l = new(ResponsiveList)
	l.Init()
	return
}

// Init initializes this list. It can be used to reset its internal state.
func (l *ResponsiveList) Init() {
	l.Reset()
	l.list.Init(1)
}

// Reset resets the contents of this List.
func (l *ResponsiveList) Reset() {
	l.list.Reset()
	l.offset = 0
	l.totalHeight = 0
	l.list.dirty = true
}

// PushBackList inserts a copy of an other list at the back of list l. The
// lists l and other must NOT be the same or nil.
func (l *ResponsiveList) PushBackList(other *ResponsiveList) {
	if l == other {
		panic("other list cannot be self: components can't be deep cloned")
	}
	other.Iterate(func(c Responsive) {
		l.PushBack(c)
	})
}

// PushFrontList inserts a copy of an other list at the front of list l. The
// lists l and other must NOT be the same or nil.
func (l *ResponsiveList) PushFrontList(other *ResponsiveList) {
	if l == other {
		panic("other list cannot be self: components can't be deep cloned")
	}
	for node, ok := other.list.Back(); ok; node, ok = node.Prev() {
		l.PushFront(node.Value().(Responsive))
	}
}

// InsertAfter inserts a new element c immediately after mark and
// returns the linked node. If mark is not an element of l, the list
// is not modified.
func (l *ResponsiveList) InsertAfter(c Responsive, mark ListNode) ListNode {
	defer l.calculateTotalHeight()
	l.list.dirty = true
	return l.list.InsertAfter(c, mark)
}

// InsertBefore inserts a new element c immediately before mark
// and returns the linked node. If mark is not an element of l,
// the list is not modified.
func (l *ResponsiveList) InsertBefore(c Responsive, mark ListNode) ListNode {
	defer l.calculateTotalHeight()
	l.list.dirty = true
	return l.list.InsertBefore(c, mark)
}

// PushBack inserts a new element c at the back of list l and
// returns the linked node.
func (l *ResponsiveList) PushBack(c Responsive) ListNode {
	defer l.calculateTotalHeight()
	l.list.dirty = true
	return l.list.PushBack(c)
}

// PushFront inserts a new element c at the front of list l and
// returns the linked node.
func (l *ResponsiveList) PushFront(c Responsive) ListNode {
	defer l.calculateTotalHeight()
	l.list.dirty = true
	return l.list.PushFront(c)
}

// Remove removes e from l if e is a node of list l. It returns the element
// value e.Value.
func (l *ResponsiveList) Remove(e *ListNode) Responsive {
	defer l.calculateTotalHeight()
	l.list.dirty = true
	return l.list.Remove(e).(Responsive)
}

// CanSeekDown returns whether SeekDown would seek one row down.
func (l *ResponsiveList) CanSeekDown() bool {
	return l.offset < l.MaxOffset()
}

// MaxOffset returns the max seek offset of this list,
// given the current dimensions.
func (l *ResponsiveList) MaxOffset() int {
	return int(math.Max(0, float64(l.totalHeight-l.list.height)))
}

// Offset returns the current seek offset of this list, measured in rows
// from the start of the seekable range (0 means fully sought toward the
// start; MaxOffset means fully sought toward the end). The value is
// independent of Alignment.
func (l *ResponsiveList) Offset() int {
	return l.offset
}

// CanSeekUp returns whether SeekUp would seek one row up.
func (l *ResponsiveList) CanSeekUp() bool {
	return l.offset > 0
}

// SeekDown shifts the contents of this list one row down.
func (l *ResponsiveList) SeekDown() bool {
	if l.Alignment == AlignmentBottom {
		return l.seekUp()
	}
	return l.seekDown()
}

// SeekUp shifts the contents of this list one row up.
func (l *ResponsiveList) SeekUp() bool {
	if l.Alignment == AlignmentBottom {
		return l.seekDown()
	}
	return l.seekUp()
}

// SeekEnd shifts the contents of this list such that the end of the last element
// is drawn at the end of the draw area.
func (l *ResponsiveList) SeekEnd() (ok bool) {
	if l.Alignment == AlignmentBottom {
		return l.seekStart()
	}
	return l.seekEnd()
}

// SeekStart shifts the contents of this list to the start of the list.
func (l *ResponsiveList) SeekStart() (ok bool) {
	if l.Alignment == AlignmentBottom {
		return l.seekEnd()
	}
	return l.seekStart()
}

// Resize resizes this list to fit within width and height.
func (l *ResponsiveList) Resize(width, height int) {
	defer l.calculateTotalHeight()

	l.list.width, l.list.height = width, height
	for el, ok := l.list.Front(); ok; el, ok = el.Next() {
		comp := el.el.Value.(*Virtual[tui.Component])
		height := comp.C.(Responsive).Height(l.list.width)
		comp.Resize(l.list.width, height)
	}
	l.list.dirty = false
}

// Draw draws this list's elements with the current seek offset.
func (l *ResponsiveList) Draw(w term.Writer) {
	if l.list.dirty {
		l.Resize(l.list.width, l.list.height)
	}
	if l.Alignment == AlignmentBottom {
		l.drawBottom(w)
	} else {
		l.drawTop(w)
	}
}

// ElementAt returns the element at pos Coordinates or panics if
// coordinates are out of the bounds of this ResponsiveList.
func (l *ResponsiveList) ElementAt(pos term.Coordinates) (ListNode, bool) {
	if pos.X < 0 || pos.Y < 0 {
		panic("negative coordinates")
	}
	if l.Alignment == AlignmentBottom {
		return l.elementAtBottom(pos)
	}
	return l.elementAtTop(pos)
}

// Sort sorts the elements of this list with the provided less function.
func (l *ResponsiveList) Sort(less func(a, b Responsive) bool) {
	l.list.Sort(func(a, b tui.Component) bool {
		return less(a.(Responsive), b.(Responsive))
	})
}

// Iterate iterates over all elements in l.
func (l *ResponsiveList) Iterate(fn func(Responsive)) {
	for node, ok := l.list.Front(); ok; node, ok = node.Next() {
		fn(node.Value().(Responsive))
	}
}

// Back returns the last node of list l or false if the list is empty.
func (l *ResponsiveList) Back() (ListNode, bool) {
	return l.list.Back()
}

// Front returns the first node of list l or false if the list is empty.
func (l *ResponsiveList) Front() (ListNode, bool) {
	return l.list.Front()
}

// Len returns the number of nodes of list l in O(1).
func (l *ResponsiveList) Len() int {
	return l.list.Len()
}

// SizeHeight returns the height of this List, in the last call to Resize.
func (l *ResponsiveList) SizeHeight() int {
	return l.list.Height()
}

// SizeWidth returns the width of this List, in the last call to Resize.
func (l *ResponsiveList) SizeWidth() int {
	return l.list.Width()
}

// Height satisfies Responsive.
func (l *ResponsiveList) Height(width int) (ret int) {
	l.Iterate(func(r Responsive) {
		ret += r.Height(width)
	})
	return
}

func (l *ResponsiveList) seekDown() bool {
	ok := l.CanSeekDown()
	if ok {
		l.offset++
	}
	return ok
}
func (l *ResponsiveList) seekUp() bool {
	ok := l.CanSeekUp()
	if ok {
		l.offset--
	}
	return ok
}
func (l *ResponsiveList) seekEnd() (ok bool) {
	prev := l.offset
	l.offset = l.MaxOffset()
	return prev != l.offset
}

func (l *ResponsiveList) seekStart() (ok bool) {
	for l.SeekUp() {
		ok = true
	}
	return
}

func (l *ResponsiveList) calculateTotalHeight() {
	l.totalHeight = 0

	for el, ok := l.list.Front(); ok; el, ok = el.Next() {
		comp := el.el.Value.(*Virtual[tui.Component])
		l.totalHeight += comp.C.(Responsive).Height(l.list.width)
	}
}

func (l *ResponsiveList) drawTop(w term.Writer) {
	pos := term.Coordinates{Y: -l.offset}
	l.doDraw(w, pos)
}

func (l *ResponsiveList) drawBottom(w term.Writer) {
	pos := term.Coordinates{Y: -l.MaxOffset() + l.offset}
	l.doDraw(w, pos)
}

func (l *ResponsiveList) doDraw(w term.Writer, pos term.Coordinates) {
	// elements could be partially drawn outside bounds
	w = term.BoundsCheckWriter(l.list.width, l.list.height, w)
	for el, ok := l.list.Front(); ok; el, ok = el.Next() {
		comp := el.el.Value.(*Virtual[tui.Component])
		height := comp.C.(Responsive).Height(l.list.width)
		nextY := pos.Y + height
		if nextY < 0 {
			pos.Y = nextY
			continue
		}
		comp.Move(pos)
		comp.Draw(w)
		if nextY >= l.list.height {
			break
		}
		pos.Y = nextY
	}
}

func (l *ResponsiveList) elementAtTop(target term.Coordinates) (ListNode, bool) {
	start := term.Coordinates{Y: -l.offset}
	return l.doElementAt(start, target)
}

func (l *ResponsiveList) elementAtBottom(target term.Coordinates) (ListNode, bool) {
	start := term.Coordinates{Y: -l.MaxOffset() + l.offset}
	return l.doElementAt(start, target)
}

func (l *ResponsiveList) doElementAt(pos, target term.Coordinates) (ListNode, bool) {
	for el, ok := l.list.Front(); ok; el, ok = el.Next() {
		comp := el.el.Value.(*Virtual[tui.Component])
		height := comp.C.(Responsive).Height(l.list.width)
		pos.Y += height
		if pos.Y > target.Y {
			return el, true
		}
	}
	return ListNode{}, false
}
