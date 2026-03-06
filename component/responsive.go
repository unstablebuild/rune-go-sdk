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
	"math"

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Responsive components implement a backpressure mechanism (Height) for
// aggregate components to dynamically resize children based on their contents.
// See Height for more details.
type Responsive interface {
	tui.Component
	// Height allows for children components to return a height hint
	// given a width so a parent component can compose accordingly.
	// The returned height can be overridden at the parent's discretion
	// (i.e. there's simply no height left on the screen)
	// so implementers should expect that on calls to Resize.
	Height(width int) int
}

// FloatingResponsive is a Floating that also satisfies Responsive.
type FloatingResponsive interface {
	Responsive
	Floating
}

// StringResponsiveConfig adds responsive-specific configuration to a StringConfig.
type StringResponsiveConfig struct {
	// NoSplitWords instructs the underlying string responsive component
	// to attempt to not split words in half when possible.
	NoSplitWords bool
	StringConfig
}

// NewResponsiveString allocates storage for a new ResponsiveString based on str and cfg.
func NewResponsiveString(str string, cfg StringResponsiveConfig) *ResponsiveString {
	return NewResponsiveStringFromCells(term.StringToCells(str), cfg)
}

// NewResponsiveStringFromCells returns a Responsive implementation for a matrix of cells.
func NewResponsiveStringFromCells(cells [][]term.Cell, cfg StringResponsiveConfig) *ResponsiveString {
	ret := new(ResponsiveString)
	ret.Init(cells, cfg)
	return ret
}

// FuncResponsive wraps a tui.Component that satisfies Responsive's Height
// by calling heightFn.
func FuncResponsive(c tui.Component, heightFn func(width int) int) Responsive {
	return respFn{Component: c, heightFn: heightFn}
}

// ResponsiveString is a String component that also satisfies Responsive.
type ResponsiveString struct {
	cfg    StringResponsiveConfig
	in     [][]term.Cell
	out    floatingWithAttributes
	width  int
	height int
}

var _ WithAttributes = (*ResponsiveString)(nil)
var _ Responsive = (*ResponsiveString)(nil)
var _ Floating = (*ResponsiveString)(nil)
var _ fmt.Stringer = (*ResponsiveString)(nil)

// Init initializes a ResponsiveString with the given cells and cfg.
func (s *ResponsiveString) Init(cells [][]term.Cell, cfg StringResponsiveConfig) {
	s.Reset(cells, cfg)
	s.Resize(0, 0) // initialize ret.out
}

// Reset resets a ResponsiveString with the given cells and cfg.
func (s *ResponsiveString) Reset(cells [][]term.Cell, cfg StringResponsiveConfig) {
	s.cfg = cfg
	s.in = cells
}

// Height satisfies Responsive.
func (s *ResponsiveString) Height(width int) int {
	if width <= 0 {
		return 0
	}
	height := len(s.massageInput(width))
	height += s.cfg.PaddingVertical
	if s.cfg.FrameCharSet != (FrameCharSet{}) {
		height += 2
	}
	return height
}

// Dimensions returns the optimal width and height.
func (s *ResponsiveString) Dimensions() (width, height int) {
	return s.out.Dimensions()
}

// Resize satisfies tui.Component.
func (s *ResponsiveString) Resize(width, height int) {
	s.width = width
	s.height = height
	outRaw := s.massageInput(width)
	s.out = newStringComp(outRaw, s.cfg.Attributes, ' ',
		s.cfg.BackgroundAttributes, s.cfg.FrameCharSet,
		s.cfg.PaddingHorizontal, s.cfg.PaddingVertical, s.cfg.Alignment, s.cfg.MinWidth)
	s.out.Resize(width, height)
}

// Draw satisfies tui.Component.
func (s *ResponsiveString) Draw(w term.Writer) {
	s.out.Draw(w)
}

// SetAttr satisfies WithAttributes.
func (s *ResponsiveString) SetAttr(attr term.Attributes) term.Attributes {
	s.cfg.Attributes = attr
	// do not set s.cfg.BackgroundAttributes
	// as this is not what the user most likely intends.
	return s.out.SetAttr(attr)
}

// String satisfies fmt.Stringer.
func (s *ResponsiveString) String() string {
	return s.out.String()
}

func (s *ResponsiveString) massageInput(width int) [][]term.Cell {
	effectiveWidth := width
	if effectiveWidth > 2 && s.cfg.FrameCharSet != (FrameCharSet{}) {
		effectiveWidth -= 2
	}
	if effectiveWidth > s.cfg.PaddingHorizontal {
		effectiveWidth -= s.cfg.PaddingHorizontal
	}
	var outRaw [][]term.Cell
	for _, col := range s.in {
		if len(col) == 0 {
			outRaw = append(outRaw, col[:])
			continue
		}
		for len(col) > 0 {
			chunkLen := int(math.Min(float64(len(col)), float64(effectiveWidth)))
			if chunkLen == 0 {
				break
			}
			origChunkLen := chunkLen
			// do not split word in half
			for s.cfg.NoSplitWords && origChunkLen != len(col) && chunkLen > 1 && col[chunkLen-1].Ch != ' ' {
				chunkLen--
			}
			// word doesn't fit, split word
			if chunkLen == 1 {
				chunkLen = origChunkLen
			}
			outRaw = append(outRaw, col[:chunkLen])
			col = col[chunkLen:]
		}
	}
	return outRaw
}

func (f nopResponsive) Height(width int) int {
	return 0
}

type respFn struct {
	tui.Component
	heightFn func(int) int
}

func (f respFn) Height(width int) int {
	return f.heightFn(width)
}
