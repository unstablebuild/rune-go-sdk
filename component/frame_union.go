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
)

// FrameUnionCharSet configures the characters used to draw the frame union
// between the top and bottom components.
type FrameUnionCharSet struct {
	FrameCharSet
	Left   rune
	Right  rune
	Top    rune
	Bottom rune
}

// DefaultFrameUnionCharSet returns the default FrameUnionCharSet used.
func DefaultFrameUnionCharSet() (ret FrameUnionCharSet) {
	ret.FrameCharSet = FrameCharSetDefault()
	ret.Left = '├'
	ret.Right = '┤'
	ret.Top = '┬'
	ret.Bottom = '┴'
	return
}

// FrameUnion is a tui.Component which Draws a main Component
// at the max available width and height, but otherwise depending on
// how many other statically sized components are stacked either left, right
// top or bottom.
//
// It respects the stacked components height if stacked on top or bottom
// and it respects the stacked components width if stacked left or right.
// The API uses Virtual instead of tui.Component to determine what's the desired
// height or width.
type FrameUnion struct {
	main          Virtual[tui.Component]
	top, bottom   []*frameVirtual
	left, right   []*frameVirtual
	height, width int

	// Frame determines whether underlying components are an instance of Frame
	// and so FrameUnion should stitch them together with FrameUnionCharSet.
	// Default is true.
	Frame bool

	// Attributes to be set to FrameUnionCharSet.
	term.Attributes

	FrameUnionCharSet
}

// NewFrameUnion allocates storage for a new FrameUnion and initializes it.
func NewFrameUnion(main tui.Component) *FrameUnion {
	ret := new(FrameUnion)
	ret.Init(main)
	return ret
}

// Init initializes this frame union with main as the main compontent.
func (u *FrameUnion) Init(main tui.Component) {
	u.FrameUnionCharSet = DefaultFrameUnionCharSet()
	u.main.C = main
	u.Frame = true
}

// UnionTop stacks top on top of the main component. This
// method panics if top is nil.
func (u *FrameUnion) UnionTop(top tui.Component, height int) {
	u.UnionTopFrame(top, height, true)
}

// UnionTopFrame stacks top on top of the main component, and if
// u.Frame is set to true and the given frame argument too, it will
// union the frames of the adjacent components with the configured
// union charset. This method panics if top is nil.
func (u *FrameUnion) UnionTopFrame(top tui.Component, height int, frame bool) {
	if top == nil {
		panic("invalid componen.Virtual")
	}
	u.top = append(u.top, &frameVirtual{
		Virtual: Virtual[tui.Component]{C: top},
		size:    height,
		frame:   frame,
	})
	u.Resize(u.width, u.height)
}

// UnionBottom stacks bottom under of the main component. This
// method panics if bottom is nil.
func (u *FrameUnion) UnionBottom(bottom tui.Component, height int) {
	u.UnionBottomFrame(bottom, height, true)
}

// UnionBottomFrame stacks bottom under of the main component, and if
// u.Frame is set to true and the given frame argument too, it will
// union the frames of the adjacent components with the configured
// union charset. This method panics if bottom is nil.
func (u *FrameUnion) UnionBottomFrame(bottom tui.Component, height int, frame bool) {
	if bottom == nil {
		panic("invalid componen.Virtual")
	}
	head := []*frameVirtual{{
		Virtual: Virtual[tui.Component]{C: bottom},
		size:    height,
		frame:   frame,
	}}
	u.bottom = append(head, u.bottom...)
	u.Resize(u.width, u.height)
}

// UnionLeft stacks left to the left of the main component. This
// method panics if left is nil.
func (u *FrameUnion) UnionLeft(left tui.Component, width int) {
	u.UnionLeftFrame(left, width, true)
}

// UnionLeftFrame stacks left to the left of the main component, and if
// u.Frame is set to true and the given frame argument too, it will
// union the frames of the adjacent components with the configured
// union charset. This method panics if left is nil.
func (u *FrameUnion) UnionLeftFrame(left tui.Component, width int, frame bool) {
	if left == nil {
		panic("invalid componen.Virtual")
	}
	u.left = append(u.left, &frameVirtual{
		Virtual: Virtual[tui.Component]{C: left},
		size:    width,
		frame:   frame,
	})
	u.Resize(u.width, u.height)
}

// UnionRight stacks right to the right of the main component. This
// method panics if right is nil.
func (u *FrameUnion) UnionRight(right tui.Component, width int) {
	u.UnionRightFrame(right, width, true)
}

// UnionRightFrame stacks right to the right of the main component, and if
// u.Frame is set to true and the given frame argument too, it will
// union the frames of the adjacent components with the configured
// union charset. This method panics if right is nil.
func (u *FrameUnion) UnionRightFrame(right tui.Component, width int, frame bool) {
	if right == nil {
		panic("invalid componen.Virtual")
	}
	head := []*frameVirtual{{
		Virtual: Virtual[tui.Component]{C: right},
		frame:   frame,
		size:    width,
	}}
	u.right = append(head, u.right...)
	u.Resize(u.width, u.height)
}

// ComponentAt returns the component at pos or false if there's no component at pos.
func (u *FrameUnion) ComponentAt(pos term.Coordinates) (Virtual[tui.Component], bool) {
	frameVirtualMain := frameVirtual{Virtual: u.main}
	main := [1]*frameVirtual{&frameVirtualMain}
	c, ok := u.componentAt(main[:], pos)
	if ok {
		return c, true
	}
	c, ok = u.componentAt(u.top, pos)
	if ok {
		return c, true
	}
	c, ok = u.componentAt(u.bottom, pos)
	if ok {
		return c, true
	}
	c, ok = u.componentAt(u.left, pos)
	if ok {
		return c, true
	}
	return u.componentAt(u.right, pos)
}

// MainPosition returns the main component's offset from the top-left corner.
func (u *FrameUnion) MainPosition() (offset term.Coordinates) {
	return u.main.Position()
}

// MainWidth returns the main component's width.
func (u *FrameUnion) MainWidth() int {
	return u.main.Width()
}

// MainHeight returns the main component's height.
func (u *FrameUnion) MainHeight() int {
	return u.main.Height()
}

// Resize satisfies tui.Component
func (u *FrameUnion) Resize(width, height int) {
	u.height, u.width = height, width

	mainHeight, topOffset := u.resizeTopBottom(width, height)
	mainWidth, leftOffset := u.resizeLeftRight(width, mainHeight, topOffset)

	u.main.Resize(mainWidth, mainHeight)
	u.main.Move(term.Coordinates{Y: topOffset, X: leftOffset})
}

// Draw satisfies tui.Component
func (u *FrameUnion) Draw(w term.Writer) {
	for _, top := range u.top {
		top.Draw(w)
	}
	for _, bottom := range u.bottom {
		bottom.Draw(w)
	}
	for _, left := range u.left {
		left.Draw(w)
	}
	for _, right := range u.right {
		right.Draw(w)
	}

	u.main.Draw(w)

	if u.main.Height() < 2 || u.main.Width() < 2 || !u.Frame {
		return
	}

	u.drawUnionCells(w)
}

type frameVirtual struct {
	size int
	Virtual[tui.Component]
	frame bool
}

func (u *FrameUnion) drawUnionCells(w term.Writer) {
	var prevStamp bool
	for i, left := range u.left {
		if !left.frame || (i < len(u.left)-1 && !u.left[i+1].frame) {
			prevStamp = false
			continue
		}
		topCh := u.Top
		bottomCh := u.Bottom
		pos := left.Position()

		if !prevStamp && i != 0 {
			u.setHorizontalUnionFrameCells(w, topCh, bottomCh, left, u.main.Height(), pos.X, pos.Y)
		}
		prevStamp = true
		u.setHorizontalUnionFrameCells(w, topCh, bottomCh, left, u.main.Height(), pos.X+left.Width()-1, pos.Y)
	}

	prevStamp = false
	for i := len(u.right) - 1; i >= 0; i-- {
		right := u.right[i]
		if !right.frame || (i > 0 && !u.right[i-1].frame) {
			prevStamp = false
			continue
		}
		pos := right.Position()
		topCh := u.Top
		bottomCh := u.Bottom

		if !prevStamp && i != len(u.right)-1 {
			u.setHorizontalUnionFrameCells(w, topCh, bottomCh, right, u.main.Height(), pos.X+right.Width()-1, pos.Y)
		}
		prevStamp = true
		u.setHorizontalUnionFrameCells(w, topCh, bottomCh, right, u.main.Height(), pos.X, pos.Y)
	}

	for i, top := range u.top {
		if !top.frame || (i < len(u.top)-1 && !u.top[i+1].frame) {
			continue
		}
		leftCh := u.Left
		rightCh := u.Right
		if i == len(u.top)-1 && len(u.left) != 0 && !u.left[0].frame {
			leftCh = u.BottomLeft
		}
		if i == len(u.top)-1 && len(u.right) != 0 && !u.right[len(u.right)-1].frame {
			rightCh = u.BottomRight
		}
		u.setVerticalUnionFrameCells(w, leftCh, rightCh, top, top.Position().Y+top.Height()-1)
	}

	for i, bottom := range u.bottom {
		if !bottom.frame || (i > 0 && !u.bottom[i-1].frame) {
			continue
		}
		leftCh := u.Left
		rightCh := u.Right
		if i == 0 && len(u.left) != 0 && !u.left[0].frame {
			leftCh = u.TopLeft
		}
		if i == 0 && len(u.right) != 0 && !u.right[len(u.right)-1].frame {
			rightCh = u.TopRight
		}
		u.setVerticalUnionFrameCells(w, leftCh, rightCh, bottom, bottom.Position().Y)
	}
}

func (u *FrameUnion) setVerticalUnionFrameCells(
	w term.Writer, left, right rune, v *frameVirtual, y int,
) {
	if v.Height() == 0 || v.Width() == 0 {
		return
	}
	w.SetCell(term.Coordinates{Y: y},
		term.Cell{Width: 1, Ch: left, Attributes: u.Attributes})
	if u.width > 0 {
		w.SetCell(term.Coordinates{X: u.width - 1, Y: y},
			term.Cell{Width: 1, Ch: right, Attributes: u.Attributes})
	}
}

func (u *FrameUnion) componentAt(
	components []*frameVirtual, pos term.Coordinates,
) (Virtual[tui.Component], bool) {
	for _, t := range components {
		tpos := t.Position()
		twidth := t.Width()
		theight := t.Height()
		if pos.X >= tpos.X && pos.Y >= tpos.Y && pos.X < tpos.X+twidth && pos.Y < tpos.Y+theight {
			return t.Virtual, true
		}
	}
	return Virtual[tui.Component]{}, false
}

func (u *FrameUnion) setHorizontalUnionFrameCells(
	w term.Writer, top, bottom rune, v *frameVirtual, height, x, y int,
) {
	if v.Height() == 0 || v.Width() == 0 {
		return
	}
	w.SetCell(term.Coordinates{X: x, Y: y},
		term.Cell{Width: 1, Ch: top, Attributes: u.Attributes})
	if height > 0 {
		w.SetCell(term.Coordinates{X: x, Y: y + height - 1},
			term.Cell{Width: 1, Ch: bottom, Attributes: u.Attributes})
	}
}

func (u *FrameUnion) resizeTopBottom(totalWidth, totalHeight int) (int, int) {
	var topOffset int
	var prevOverlap bool
	for _, top := range u.top {
		if u.Frame && top.frame && prevOverlap {
			topOffset--
		}
		prevOverlap = top.frame
		coords := term.Coordinates{Y: topOffset}
		top.Move(coords)

		height := top.size
		if height > 0 {
			topOffset += height
			top.Resize(totalWidth, height)
		} else {
			top.Resize(0, 0)
		}
	}

	if prevOverlap && u.Frame && len(u.top) != 0 {
		topOffset--
	}

	prevOverlap = u.Frame
	var bottomHeight int
	for _, bottom := range u.bottom {
		if u.Frame && bottom.frame && prevOverlap {
			bottomHeight--
		}
		prevOverlap = bottom.frame
		height := bottom.size
		if height > 0 {
			bottomHeight += height
		}
	}

	// if there's too many union components for available height
	// do not draw them.
	mainHeight := totalHeight - topOffset - bottomHeight
	if (u.Frame && mainHeight < 2) || mainHeight < 1 {
		for _, top := range u.top {
			top.Resize(0, 0)
		}
		for _, bottom := range u.bottom {
			bottom.Resize(0, 0)
		}
		return max(0, totalHeight), 0
	}

	bottomOffset := totalHeight - bottomHeight
	prevOverlap = u.Frame
	for _, bottom := range u.bottom {
		if u.Frame && bottom.frame && prevOverlap {
			bottomOffset--
		}
		prevOverlap = bottom.frame
		coords := term.Coordinates{Y: bottomOffset}
		bottom.Move(coords)
		height := bottom.size
		if height > 0 {
			bottom.Resize(totalWidth, height)
			bottomOffset += height
		} else {
			bottom.Resize(0, 0)
		}
	}

	return mainHeight, topOffset
}

func (u *FrameUnion) resizeLeftRight(totalWidth, totalHeight, topOffset int) (int, int) {
	var leftOffset int
	var prevOverlap bool
	for _, left := range u.left {
		if u.Frame && left.frame && prevOverlap {
			leftOffset--
		}
		prevOverlap = left.frame
		left.Move(term.Coordinates{X: leftOffset, Y: topOffset})

		width := left.size
		if width > 0 {
			leftOffset += width
			left.Resize(width, totalHeight)
		} else {
			left.Resize(0, 0)
		}
	}
	if prevOverlap && u.Frame && len(u.left) != 0 {
		leftOffset--
	}

	prevOverlap = u.Frame
	var rightWidth int
	for _, right := range u.right {
		if u.Frame && right.frame && prevOverlap {
			rightWidth--
		}
		prevOverlap = right.frame
		width := right.size
		if width > 0 {
			rightWidth += width
		}
	}

	mainWidth := totalWidth - leftOffset - rightWidth
	if (u.Frame && mainWidth < 2) || mainWidth < 1 {
		for _, left := range u.left {
			left.Resize(0, 0)
		}
		for _, right := range u.right {
			right.Resize(0, 0)
		}
		return max(0, totalWidth), 0
	}

	rightOffset := totalWidth - rightWidth
	prevOverlap = u.Frame
	for _, right := range u.right {
		if u.Frame && right.frame && prevOverlap {
			rightOffset--
		}
		prevOverlap = right.frame
		right.Move(term.Coordinates{Y: topOffset, X: rightOffset})
		width := right.size
		if width > 0 {
			right.Resize(width, totalHeight)
			rightOffset += width
		} else {
			right.Resize(0, 0)
		}
	}

	return mainWidth, leftOffset
}
