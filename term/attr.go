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

	// AttrKeepBackground makes Writer.SetCell keep the background colour
	// of the cell it replaces instead of the written cell's. Writers that
	// hold their cells store the result without the flag, and writers
	// that forward cells pass it on. A writer that cannot read back the
	// cell it replaces, such as ScreenWriter, draws the written cell's
	// background. UnionAttributes ignores the flag, since a default
	// colour there already keeps the existing one.
	AttrKeepBackground AttrMask = AttrInvalid << 3 // bit 10
)

// Style is a foreground/background color pair plus an attribute mask.
// It is the renderer-facing companion of term.Attributes (which
// carries the same fields on a Cell).
type Style struct {
	Fg    Color
	Bg    Color
	Attrs AttrMask
	// Underline colours an AttrUnderline stroke; when not valid the
	// stroke takes Fg.
	Underline Color
}

// StyleDefault is the zero-valued Style.
var StyleDefault Style
