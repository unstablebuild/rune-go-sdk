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

package handler

import (
	"slices"

	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// FrameUnion is the handler counterpart to component.FrameUnion: it
// routes events to its main handler and the handlers stacked around it
// as Group does, keys to the main handler unless SetFocus moves them.
// Where members overlap, as framed ones do, the pointer is over the one
// ComponentAt returns.
//
// Use NewFrameUnion, or Init, to initialize a FrameUnion.
type FrameUnion struct {
	component.FrameUnion
	routed
	main          tui.Handler
	top, bottom   []tui.Handler
	left, right   []tui.Handler
	capturing     []tui.Handler
	width, height int
}

// NewFrameUnion allocates storage for a new FrameUnion and initializes it.
func NewFrameUnion(main tui.Handler) *FrameUnion {
	u := new(FrameUnion)
	u.Init(main)
	return u
}

// Init initializes this FrameUnion with main as its main handler. As
// component.FrameUnion.Init does, it keeps the handlers stacked around
// the main one: start from a zero FrameUnion to drop them.
func (u *FrameUnion) Init(main tui.Handler) {
	u.FrameUnion.Init(main)
	u.main = main
	u.layout = u.children
	u.SetFocus(main)
}

// UnionTop stacks top on top of the main handler. It panics if top is
// nil.
func (u *FrameUnion) UnionTop(top tui.Handler, height int) {
	u.UnionTopFrame(top, height, true)
}

// UnionTopFrame stacks top on top of the main handler, unioning their
// frames if frame and u.Frame are set. It panics if top is nil.
func (u *FrameUnion) UnionTopFrame(top tui.Handler, height int, frame bool) {
	u.FrameUnion.UnionTopFrame(top, height, frame)
	u.top = append(u.top, top)
}

// UnionBottom stacks bottom under the main handler. It panics if bottom
// is nil.
func (u *FrameUnion) UnionBottom(bottom tui.Handler, height int) {
	u.UnionBottomFrame(bottom, height, true)
}

// UnionBottomFrame stacks bottom under the main handler, unioning their
// frames if frame and u.Frame are set. It panics if bottom is nil.
func (u *FrameUnion) UnionBottomFrame(bottom tui.Handler, height int, frame bool) {
	u.FrameUnion.UnionBottomFrame(bottom, height, frame)
	u.bottom = append([]tui.Handler{bottom}, u.bottom...)
}

// UnionLeft stacks left to the left of the main handler. It panics if
// left is nil.
func (u *FrameUnion) UnionLeft(left tui.Handler, width int) {
	u.UnionLeftFrame(left, width, true)
}

// UnionLeftFrame stacks left to the left of the main handler, unioning
// their frames if frame and u.Frame are set. It panics if left is nil.
func (u *FrameUnion) UnionLeftFrame(left tui.Handler, width int, frame bool) {
	u.FrameUnion.UnionLeftFrame(left, width, frame)
	u.left = append(u.left, left)
}

// UnionRight stacks right to the right of the main handler. It panics
// if right is nil.
func (u *FrameUnion) UnionRight(right tui.Handler, width int) {
	u.UnionRightFrame(right, width, true)
}

// UnionRightFrame stacks right to the right of the main handler,
// unioning their frames if frame and u.Frame are set. It panics if
// right is nil.
func (u *FrameUnion) UnionRightFrame(right tui.Handler, width int, frame bool) {
	u.FrameUnion.UnionRightFrame(right, width, frame)
	u.right = append([]tui.Handler{right}, u.right...)
}

// CaptureDrags makes member, which must already be part of this union,
// capture the left presses on it as Child.CaptureDrags describes.
func (u *FrameUnion) CaptureDrags(member tui.Handler) {
	u.capturing = append(u.capturing, member)
}

// Resize satisfies tui.Component.
func (u *FrameUnion) Resize(width, height int) {
	u.width, u.height = width, height
	u.FrameUnion.Resize(width, height)
}

// children lays the members out in the reverse of the order ComponentAt
// looks for them in, as the last child a Layout has at a cell is the one
// there.
func (u *FrameUnion) children() Layout {
	members := slices.Concat([]tui.Handler{u.main}, u.top, u.bottom, u.left, u.right)
	l := Layout{Width: u.width, Height: u.height, Children: make([]Child, 0, len(members))}
	for _, h := range slices.Backward(members) {
		if v, ok := u.Member(h); ok {
			l.Children = append(l.Children, Child{
				Handler: h, Position: v.Position(), Width: v.Width(), Height: v.Height(),
				CaptureDrags: slices.ContainsFunc(u.capturing, func(c tui.Handler) bool { return same(c, h) }),
			})
		}
	}
	return l
}

// Draw satisfies tui.Component.
func (u *FrameUnion) Draw(w term.Writer) {
	u.FrameUnion.Draw(w)
}
