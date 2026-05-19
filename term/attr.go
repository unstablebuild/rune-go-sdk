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

package term

// AttrMask is a bitmask of text-rendering attributes (bold, italic,
// underline, ...). Bits 0..7 mirror the standard SGR attributes; the
// higher bits carry term-specific render flags.
type AttrMask uint16

// Standard SGR attribute bits (matching tcell's layout in the low 8
// bits so cells round-trip through tcell-backed renderers cleanly).
const (
	AttrBold          AttrMask = 1 << iota // bit 0
	AttrBlink                              // bit 1
	AttrReverse                            // bit 2
	AttrUnderline                          // bit 3
	AttrDim                                // bit 4
	AttrItalic                             // bit 5
	AttrStrikeThrough                      // bit 6
	AttrInvalid                            // bit 7 (sentinel)
)

const (
	// AttrNone is the empty attribute set.
	AttrNone AttrMask = 0

	// AttrVerticalRenderOffset instructs the renderer to render the
	// cell offset by +height/2.
	AttrVerticalRenderOffset AttrMask = AttrInvalid << 1 // bit 8

	// AttrNegativeVerticalRenderOffset instructs the renderer to render
	// the cell offset by -height/2.
	AttrNegativeVerticalRenderOffset AttrMask = AttrInvalid << 2 // bit 9
)

// Style is a foreground/background color pair plus an attribute mask.
// It is the renderer-facing companion of term.Attributes (which
// carries the same fields on a Cell).
type Style struct {
	Fg    Color
	Bg    Color
	Attrs AttrMask
}

// StyleDefault is the zero-valued Style.
var StyleDefault Style
