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

// SpanConfig represents the configuration of a Span.
//
// Padding can be configured as an absolute number of cells (PadHorizontal/PadVertical)
// or as a percentage of the available space (PadHorizontalPerc/PadVerticalPerc).
// Either PadHorizontal/PadVertical or PadHorizontalPerc/PadVerticalPerc should be set;
// If both are set, then Horizontal/Vertical take precedence.
//
// Negative padding on PadHorizontal/PadVertical indicates that the padding should be
// automatically calculated based on the available height/width. For instance,
// a Horizontal padding of -1, indicates that the padding needs to be set such
// that the inner component is exactly 1 cell.
type SpanConfig struct {
	PadHorizontal int
	PadVertical   int

	PadHorizontalPerc float64
	PadVerticalPerc   float64

	ContentAlignment Alignment
}

// Span is a component that takes another component and handles padding and
// alignment.
type Span struct {
	content       Virtual[tui.Component]
	width, height int
	cfg           SpanConfig
}

var _ Responsive = (*Span)(nil)
var _ WithAttributes = (*Span)(nil)
var _ Floating = (*Span)(nil)

// DefaultSpanConfig returns the default span configuration wich is no padding,
// and content alignment centered.
func DefaultSpanConfig() (ret SpanConfig) {
	ret.ContentAlignment = AlignmentCentered
	return
}

// NewSpan returns an initialized Span. See Span.Init for more info.
func NewSpan(content tui.Component, cfg SpanConfig) *Span {
	s := new(Span)
	s.Init(content, cfg)
	return s
}

// Init initializes this Span with content.
func (s *Span) Init(content tui.Component, cfg SpanConfig) {
	if cfg.PadHorizontalPerc < 0 || cfg.PadHorizontalPerc > 1 ||
		cfg.PadVerticalPerc < 0 || cfg.PadVerticalPerc > 1 {
		panic("padding percentage must be between range [0, 1]")
	}

	s.cfg = cfg
	s.content.C = content
}

func calculateContentOffset(
	horizontalPadding, verticalPadding int, flags Alignment,
) (offset term.Coordinates) {
	if flags&AlignmentVerticallyCentered != 0 {
		offset.Y = verticalPadding / 2
	} else if flags&AlignmentBottom != 0 {
		offset.Y = verticalPadding
	}

	if flags&AlignmentHorizontallyCentered != 0 {
		offset.X = horizontalPadding / 2
	} else if flags&AlignmentRight != 0 {
		offset.X = horizontalPadding
	}

	return
}

func alignContent(
	content *Virtual[tui.Component],
	width, height int,
	horizontalPadding, verticalPadding int,
	flags Alignment) {

	offset := calculateContentOffset(horizontalPadding, verticalPadding, flags)
	contentWidth := width - horizontalPadding
	contentHeight := height - verticalPadding
	if contentWidth < 0 {
		contentWidth = max(0, width)
		offset.X = 0
	}
	if contentHeight < 0 {
		contentHeight = max(0, height)
		offset.Y = 0
	}

	content.Resize(contentWidth, contentHeight)
	content.Move(offset)
}

func (s *Span) getPadding(width, height int) (int, int) {
	var hPadding, vPadding int

	if s.cfg.PadHorizontal == 0 {
		hPadding = int(s.cfg.PadHorizontalPerc * float64(width))
	} else if s.cfg.PadHorizontal < 0 {
		hPadding = int(math.Max(0, float64(width+s.cfg.PadHorizontal)))
	} else {
		hPadding = s.cfg.PadHorizontal
	}

	if s.cfg.PadVertical == 0 {
		vPadding = int(s.cfg.PadVerticalPerc * float64(height))
	} else if s.cfg.PadVertical < 0 {
		vPadding = int(math.Max(0, float64(height+s.cfg.PadVertical)))
	} else {
		vPadding = s.cfg.PadVertical
	}

	return hPadding, vPadding
}

// Resize satisfies tui.Component
func (s *Span) Resize(width, height int) {
	hPadding, vPadding := s.getPadding(width, height)
	if width < 3 {
		hPadding = 0
	}
	if height < 3 {
		vPadding = 0
	}
	alignContent(&s.content, width, height, hPadding,
		vPadding, s.cfg.ContentAlignment)

	s.width, s.height = width, height
}

// Draw satisfies tui.Component
func (s *Span) Draw(w term.Writer) {
	s.content.Draw(w)
}

// SetContent updates the underlying component and resizes it
// to conform to this frame's width and height.
func (s *Span) SetContent(content tui.Component) {
	s.content.C = content
	s.Resize(s.width, s.height)
}

// Content returns this Span's underlying content.
func (s *Span) Content() tui.Component {
	return s.content.C
}

// ContentOffset returns the offset in term.Coordinates of the content inside the span.
func (s *Span) ContentOffset() term.Coordinates {
	offset := s.content.Position()
	if inner, ok := s.content.C.(*Span); ok {
		inner := inner.ContentOffset()
		return term.Coordinates{
			X: offset.X + inner.X,
			Y: offset.Y + inner.Y,
		}
	}
	return offset
}

// Height satisfies Responsive by returning the underlying component's Height
// with added padding, or panics if the underlying component does
// not satisfy Responsive.
func (s *Span) Height(width int) int {
	r, ok := s.content.C.(Responsive)
	if !ok {
		panic("underlying component does not satisfy Responsive")
	}
	hPadding, vPadding := s.getPadding(width, s.height)

	// it doesn't make sense to specify auto vertical padding
	// if underlying component is Responsive
	if s.cfg.PadVertical < 0 {
		vPadding = 0
	}
	return r.Height(width-hPadding) + vPadding
}

// Dimensions satisfies Floating by returning the underlying component's Dimensions
// with added padding, or panics if the underlying component does not satisfy
// Floating.
func (s *Span) Dimensions() (width, height int) {
	width, height = s.content.C.(Floating).Dimensions()
	hPadding, vPadding := s.getPadding(width, s.height)
	width += hPadding
	height += vPadding
	return
}

// SetAttr satisfies WithAttributes if the underlying component
// satisfies WithAttributes, otherwise it panics.
func (s *Span) SetAttr(attr term.Attributes) term.Attributes {
	return s.content.C.(WithAttributes).SetAttr(attr)
}
