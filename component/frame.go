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

// FrameCharSet is a struct used to store the set of characters used to
// draw a frame.
//
// Example characters (ASCII 9472-9580):
//
//	'─', '━', '│', '┃', '┄', '┅', '┆', '┇', '┈', '┉', '┊', '┋', '┌', '┍',
//
//	'┎', '┏', '┐', '┑', '┒', '┓', '└', '┕', '┖', '┗', '┘', '┙', '┚', '┛',
//
//	'├', '┝', '┞', '┟', '┠', '┡', '┢', '┣', '┤', '┥', '┦', '┧', '┨', '┩',
//
//	'┪', '┫', '┬', '┭', '┮', '┯', '┰', '┱', '┲', '┳', '┴', '┵', '┶', '┷',
//
//	'┸', '┹', '┺', '┻', '┼', '┽', '┾', '┿', '╀', '╁', '╂', '╃', '╄', '╅',
//
//	'╆', '╇', '╈', '╉', '╊', '╋', '╌', '╍', '╎', '╏', '═', '║', '╒', '╓',
//
//	'╔', '╕', '╖', '╗', '╘', '╙', '╚', '╛', '╜', '╝', '╞', '╟', '╠', '╡',
//
//	'╢', '╣', '╤', '╥', '╦', '╧', '╨', '╩'
type FrameCharSet struct {
	TopLeft, TopRight               rune
	BottomLeft, BottomRight         rune
	HorizontalTop, VerticalLeft     rune
	HorizontalBottom, VerticalRight rune
}

// FrameCharSetDefault returns the default FrameCharSet
// used accross the library. It produces the following frame:
//
//	┌─┐
//	│ │
//	└─┘
func FrameCharSetDefault() FrameCharSet {
	return FrameCharSet{
		HorizontalTop:    '─',
		HorizontalBottom: '─',
		VerticalRight:    '│',
		VerticalLeft:     '│',
		TopLeft:          '┌',
		TopRight:         '┐',
		BottomLeft:       '└',
		BottomRight:      '┘',
	}
}

// FrameCharSetHighlight returns a charset that produces the following frame:
//
//	┏━┓
//	┃ ┃
//	┗━┛
func FrameCharSetHighlight() FrameCharSet {
	return FrameCharSet{
		HorizontalTop:    '━',
		HorizontalBottom: '━',
		VerticalLeft:     '┃',
		VerticalRight:    '┃',
		TopLeft:          '┏',
		TopRight:         '┓',
		BottomLeft:       '┗',
		BottomRight:      '┛',
	}
}

// FrameCharSetStack returns a charset that produces the following frame:
//
//	├─┤
//	│ │
//	├─┤
func FrameCharSetStack() FrameCharSet {
	return FrameCharSet{
		HorizontalTop:    '─',
		HorizontalBottom: '─',
		VerticalLeft:     '│',
		VerticalRight:    '│',
		TopLeft:          '├',
		TopRight:         '┤',
		BottomLeft:       '├',
		BottomRight:      '┤',
	}
}

// FrameCharSetStackHighlight returns a charset that produces the following frame:
//
//	┢━┪
//	┃ ┃
//	┡━┩
func FrameCharSetStackHighlight() FrameCharSet {
	return FrameCharSet{
		HorizontalTop:    '━',
		HorizontalBottom: '━',
		VerticalLeft:     '┃',
		VerticalRight:    '┃',
		TopLeft:          '┢',
		TopRight:         '┪',
		BottomLeft:       '┡',
		BottomRight:      '┩',
	}
}

// FrameCharSetStackHead returns a charset that produces the following frame:
//
//	┌─┐
//	│ │
//	├─┤
func FrameCharSetStackHead() FrameCharSet {
	return FrameCharSet{
		HorizontalTop:    '─',
		HorizontalBottom: '─',
		VerticalLeft:     '│',
		VerticalRight:    '│',
		TopLeft:          '┌',
		TopRight:         '┐',
		BottomLeft:       '├',
		BottomRight:      '┤',
	}
}

// FrameCharSetStackHeadHighlight returns a charset that produces the following frame:
//
//	┏━┓
//	┃ ┃
//	┡━┩
func FrameCharSetStackHeadHighlight() FrameCharSet {
	return FrameCharSet{
		HorizontalTop:    '━',
		HorizontalBottom: '━',
		VerticalLeft:     '┃',
		VerticalRight:    '┃',
		TopLeft:          '┏',
		TopRight:         '┓',
		BottomLeft:       '┡',
		BottomRight:      '┩',
	}
}

// FrameCharSetStackTail returns a charset that produces the following frame:
//
//	├─┤
//	│ │
//	└─┘
func FrameCharSetStackTail() FrameCharSet {
	return FrameCharSet{
		HorizontalTop:    '─',
		HorizontalBottom: '─',
		VerticalLeft:     '│',
		VerticalRight:    '│',
		TopLeft:          '├',
		TopRight:         '┤',
		BottomLeft:       '└',
		BottomRight:      '┘',
	}
}

// FrameCharSetStackTailHighlight returns a charset that produces the following frame:
//
//	┢━┪
//	┃ ┃
//	┗━┛
func FrameCharSetStackTailHighlight() FrameCharSet {
	return FrameCharSet{
		HorizontalTop:    '━',
		HorizontalBottom: '━',
		VerticalLeft:     '┃',
		VerticalRight:    '┃',
		TopLeft:          '┢',
		TopRight:         '┪',
		BottomLeft:       '┗',
		BottomRight:      '┛',
	}
}

// Frame is a Component that simply draws a border around a nested component.
// By default, uses FrameCharSetDefault() for the border characters.
//
// Note that this component can achieve other effects (highlight, frame)
// by setting the right cell characters and/or attributes.
type Frame struct {
	FrameCharSet
	term.Attributes
	ScrollBarAttributes term.Attributes
	ScrollBarChar       rune

	content         Virtual[tui.Component]
	bwidth, bheight int
	width, height   int
}

var _ WithAttributes = (*Frame)(nil)
var _ Responsive = (*Frame)(nil)
var _ Floating = (*Frame)(nil)

// NewFrame allocates storage and initializes a new frame with the given
// border attributes and underlying component.
func NewFrame(content tui.Component) (f *Frame) {
	f = new(Frame)
	f.Init(content)
	return
}

// Init initializes this frame with the given Component and border attributes.
func (f *Frame) Init(content tui.Component) {
	f.content.C = content
	f.FrameCharSet = FrameCharSetDefault()
}

// Content returns the underlying Component.
func (f *Frame) Content() tui.Component {
	return f.content.C
}

// Dimensions satisfies Floating. If the underlying component
// is not a Floating component this method panics.
func (f *Frame) Dimensions() (width, height int) {
	width, height = f.content.C.(Floating).Dimensions()
	width += 2
	height += 2
	return
}

// SetContent updates the underlying component and resizes it
// to conform to this frame's width and height.
func (f *Frame) SetContent(content tui.Component) {
	f.SetContentResize(content, true)
}

// SetContentResize allows clients to control whether content
// is resized during this method call.
func (f *Frame) SetContentResize(content tui.Component, resize bool) {
	f.content.C = content
	if resize {
		f.Resize(f.width, f.height)
	}
}

// SetAttr satisfies WithAttributes. This method only changes
// frame attr if underlying content does not satisfy WithAttributes.
func (f *Frame) SetAttr(attr term.Attributes) (ret term.Attributes) {
	ret = f.Attributes
	f.Attributes = attr
	w, ok := f.content.C.(WithAttributes)
	if ok {
		ret = w.SetAttr(attr)
	}
	return
}

// Resize updates this frame with a new width and height. If width or height
// is smaller than 3 cells, the border will not be drawn.
func (f *Frame) Resize(width, height int) {
	// deactivate frame if there's not space for content
	if width < 3 || height < 3 {
		f.bwidth, f.bheight = 0, 0
	} else {
		f.bwidth, f.bheight = 2, 2
	}

	alignContent(&f.content, width, height, f.bwidth, f.bheight, AlignmentCentered)
	f.width, f.height = width, height
}

// Height satisfies component.Responsive. It panics if underlying component is
// not component.Responsive.
func (f *Frame) Height(width int) int {
	if width < 3 {
		return f.content.C.(Responsive).Height(width)
	}
	ret := f.content.C.(Responsive).Height(width - 2)
	// Ensure that Height is consistent with Resize
	// behaviour on height or width < 3.
	// This forces return height to at least be >= 3
	if ret == 0 {
		ret = 1
	}
	ret += 2
	return ret
}

// Draw draws this frame's border and contents to the given Writer.
func (f *Frame) Draw(w term.Writer) {
	if f.bwidth == 0 || f.bheight == 0 {
		f.content.Draw(w)
		return
	}
	limitX, limitY := f.width-1, f.height-1
	DrawFrame(w, f.FrameCharSet, f.Attributes, limitX, limitY)
	f.content.Draw(w)

	offset, height, ok := f.ScrollBar()
	if !ok {
		return
	}

	ch := f.ScrollBarChar
	if ch == 0 {
		ch = f.VerticalRight
	}

	for i := offset.Y; i < offset.Y+height; i++ {
		w.SetCell(term.Coordinates{X: limitX, Y: i},
			term.Cell{
				Width:      1,
				Ch:         ch,
				Attributes: f.ScrollBarAttributes,
			})
	}
}

// ContentPosition returns the position of the content inside this frame.
func (f *Frame) ContentPosition() term.Coordinates {
	return calculateContentOffset(f.bwidth, f.bheight, AlignmentCentered)
}

// ContentSize returns the size and width of the inner content.
func (f *Frame) ContentSize() (int, int) {
	return f.content.Width(), f.content.Height()
}

// ScrollBar returns the position and height of the scrollbar, if a scroll bar
// is to be drawn, otherwise it returns false.
func (f *Frame) ScrollBar() (pos term.Coordinates, height int, ok bool) {
	scrollable, ok := f.content.C.(Scrollable)
	if !ok {
		return
	}

	maxOffset := scrollable.MaxSeekOffset()
	if maxOffset == 0 {
		return
	}

	offset, height := calculateScrollBar(scrollable.SeekOffset(), maxOffset, f.height)
	return term.Coordinates{Y: offset, X: f.width - 1}, height, true
}

func calculateScrollBar(offset, maxOffset, height int) (
	proportionalOffset, proportionalHeight int,
) {
	effectiveHeight := height - 2
	rows := maxOffset + effectiveHeight
	proportionalOffset = 1 + int(
		math.Floor(float64((offset*effectiveHeight))/float64(rows)))
	proportionalHeight = int(math.Max(1,
		math.Ceil(float64(effectiveHeight*effectiveHeight)/float64(rows))))
	return
}

// DrawFrame draws a frame at 0, 0, limitX, limitY, with the given FrameCharSet.
func DrawFrame(
	w term.Writer, f FrameCharSet, attrs term.Attributes, limitX, limitY int,
) {
	for i := 0; i < limitX; i++ {
		w.SetCell(term.Coordinates{X: i, Y: 0},
			term.Cell{Width: 1, Ch: f.HorizontalTop, Attributes: attrs})
		w.SetCell(term.Coordinates{X: i, Y: limitY},
			term.Cell{Width: 1, Ch: f.HorizontalBottom, Attributes: attrs})
	}

	for i := 0; i < limitY; i++ {
		w.SetCell(term.Coordinates{X: 0, Y: i},
			term.Cell{Width: 1, Ch: f.VerticalLeft, Attributes: attrs})
		w.SetCell(term.Coordinates{X: limitX, Y: i},
			term.Cell{Width: 1, Ch: f.VerticalRight, Attributes: attrs})
	}

	w.SetCell(term.Coordinates{X: 0, Y: 0},
		term.Cell{Width: 1, Ch: f.TopLeft, Attributes: attrs})

	w.SetCell(term.Coordinates{X: limitX, Y: 0},
		term.Cell{Width: 1, Ch: f.TopRight, Attributes: attrs})

	w.SetCell(term.Coordinates{X: 0, Y: limitY},
		term.Cell{Width: 1, Ch: f.BottomLeft, Attributes: attrs})

	w.SetCell(term.Coordinates{X: limitX, Y: limitY},
		term.Cell{Width: 1, Ch: f.BottomRight, Attributes: attrs})
}
