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
	"fmt"

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// StringConfig defines options for StringConfig and StringResponsive
// constructors.
type StringConfig struct {
	Alignment
	term.Attributes
	FrameCharSet
	BackgroundAttributes term.Attributes
	BackgroundRune       rune
	PaddingVertical      int
	PaddingHorizontal    int
	MinWidth             int
}

// String is a tui.Component that draws a string with or without
// a background. It satisfies both Floating and WithAttributes.
// See NewString and NewStringWithConfig for more details.
type String struct {
	cfg StringConfig
	floatingWithAttributes
}

var _ WithAttributes = String{}
var _ Floating = String{}
var _ fmt.Stringer = String{}

// NewStringWithConfig converts a string into a static tui.Compontent with
// background/foreground attributes, content alignment and a frame,
// all configurable through cfg. The returned component is significantly
// slower to Draw and Resize than the component returned by String.
func NewStringWithConfig(str string, cfg StringConfig) String {
	cells := term.StringToCells(str)
	comp := newStringComp(cells, cfg.Attributes, cfg.BackgroundRune,
		cfg.BackgroundAttributes, cfg.FrameCharSet,
		cfg.PaddingHorizontal, cfg.PaddingVertical, cfg.Alignment, cfg.MinWidth)
	return String{
		cfg:                    cfg,
		floatingWithAttributes: comp,
	}
}

// NewString converts a string into top left centered one line tui.Component
// which draws the given string. If the string needs to be centered dynamically,
// or drawn multi-line use StringWithConfig instead.
func NewString(str string) String {
	return NewStringWithConfig(str, StringConfig{
		Alignment: AlignmentLeft,
	})
}

// Config returns this String's StringConfig.
func (s String) Config() StringConfig {
	return s.cfg
}

// LazyBytes is an immutable String component that is allocation free
// until the first call to Draw.
//
// Akin to String, it compacts string into one line and but it
// takes data as a slice of bytes and a set of x positions
// to apply TokenAttributes to. In contrast, SetAttr it's a O(n)
// rather than String's O(1), so do not call it for every string
// after they have been drawn once, otherwise just use String,
// which will offer better features and performance characteristics.
//
// It should be wrapped by a Virtual component or used with
// a term.Writer to handle out of bound calls to SetCell.
//
// It is useful for collections, where not all strings need to
// be drawn and there's a clear performance requirement
// that offsets its limitations.
//
// Note that grapheme clusters are not supported by this component.
type LazyBytes struct {
	Data            []byte
	Tokens          *[]int
	Attributes      term.Attributes
	TokenAttributes term.Attributes
}

// Resize is ignored.
func (l LazyBytes) Resize(width, height int) {
}

// Draw satisfies tui.Component.
func (l LazyBytes) Draw(w term.Writer) {
	for i, r := range l.Data {
		w.SetCell(term.Coordinates{X: i, Y: 0}, term.Cell{
			Ch:         rune(r),
			Attributes: l.Attributes,
			Width:      1, // only support width=1 graphemes
		})
	}
	if l.Tokens != nil {
		for _, x := range *l.Tokens {
			w.UnionAttributes(term.Coordinates{X: x}, l.TokenAttributes)
		}
	}
}

type floatingWithAttributes interface {
	tui.Component
	Floating
	WithAttributes
	fmt.Stringer
}

type stringComp struct {
	width, height int
	cells         [][]term.Cell
	attr          term.Attributes
}

func (s *stringComp) String() string {
	return term.CellsToString(s.cells)
}

func (s *stringComp) Resize(width, height int) {
	s.width, s.height = width, height
}

func (s *stringComp) Draw(w term.Writer) {
	for y, r := range s.cells {
		if y >= s.height {
			break
		}
		var offset int
		for x, c := range r {
			xi := x + offset
			if c.Width > 1 {
				offset += int(c.Width) - 1
			}
			if xi >= s.width {
				break
			}
			w.SetCell(term.Coordinates{X: xi, Y: y},
				term.Cell{
					Attributes: s.attr,
					Ch:         c.Ch,
					Combining:  c.Combining,
					Width:      c.Width,
					Bytes:      c.Bytes,
				})
		}
	}
}

func (s *stringComp) SetAttr(attr term.Attributes) (ret term.Attributes) {
	ret = s.attr
	s.attr = attr
	return
}

func (s *stringComp) Dimensions() (width, height int) {
	height = len(s.cells)
	width = term.CalculateOptimalWidth(s.cells)
	return
}

// used to wrap Background and provide SetAttr to underlying stringComp
type backgroundStrWrapper struct {
	frame bool
	pad   bool
	span  *Span
	*Background
}

func (s backgroundStrWrapper) SetAttr(attr term.Attributes) term.Attributes {
	return s.Background.SetAttr(attr)
}

func (s backgroundStrWrapper) String() string {
	spanContent := s.Background.Content().(*Span).Content()
	if s.frame {
		return spanContent.(stringerFrame).String()
	}
	return spanContent.(*stringComp).String()
}

func (s backgroundStrWrapper) Dimensions() (width, height int) {
	return s.span.Dimensions()
}

func withBackgroundWrapper(
	comp WithAttributes, height, width int, background term.Cell,
	frame, pad bool, alignment Alignment,
) backgroundStrWrapper {
	span := NewSpan(comp, SpanConfig{
		PadVertical:      -height,
		PadHorizontal:    -width,
		ContentAlignment: alignment,
	})
	return backgroundStrWrapper{
		frame:      frame,
		pad:        pad,
		span:       span,
		Background: WithBackground(span, background),
	}
}

func newStringComp(
	cells [][]term.Cell, attr term.Attributes, c rune, battr term.Attributes,
	frameCharSet FrameCharSet, padWidth, padHeight int, alg Alignment,
	minWidth int,
) floatingWithAttributes {
	var comp floatingWithAttributes

	comp = &stringComp{cells: cells, attr: attr}
	shouldFrame := frameCharSet != (FrameCharSet{})

	var height int
	if shouldFrame {
		minWidth -= (2 + padWidth)
	}
	width := max(minWidth, term.CalculateOptimalWidth(cells))

	background := term.Cell{Width: 1, Ch: c, Attributes: battr}
	height = len(cells)
	shouldPad := padWidth != 0 || padHeight != 0

	if shouldFrame {
		// if inner pad is provided, center text
		if shouldPad {
			// background of inner padding looks better if it's the same attr
			// as the text.
			background := term.Cell{Width: 1, Ch: c, Attributes: attr}
			comp = withBackgroundWrapper(comp, height, width, background, false, false, alg)
		}
		width += 2 + padWidth
		height += 2 + padHeight

		frame := NewFrame(comp)
		frame.Attributes = attr
		frame.FrameCharSet = frameCharSet
		comp = stringerFrame{Stringer: comp, Frame: frame}
	}

	return withBackgroundWrapper(comp, height, width, background, shouldFrame, shouldPad, alg)
}

type stringerFrame struct {
	fmt.Stringer
	*Frame
}
