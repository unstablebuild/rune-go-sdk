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

import "strconv"

// Color represents a color packed into a 32-bit value. The low 24 bits
// hold the payload (palette index 0..255 or 24-bit RGB), the high byte
// holds the validity / kind flags ColorValid, ColorIsRGB, ColorSpecial.
//
// The renderer boundary (term.Screen implementations) is responsible
// for translating Color into whatever the underlying display library
// expects. Helpers that bridge to the tcell library live in
// tcell.go alongside the TcellScreen adapter.
type Color uint32

// Flag bits live in the upper byte so the low 24 bits remain a usable
// palette/RGB payload.
const (
	// ColorDefault is the zero value; instructs the renderer to use
	// whatever color the underlying terminal considers default.
	ColorDefault Color = 0

	// ColorValid marks a Color as initialised; without it Color is
	// treated as ColorDefault.
	ColorValid Color = 1 << 24

	// ColorIsRGB indicates that the low 24 bits hold a 24-bit RGB value
	// rather than a palette index.
	ColorIsRGB Color = 1 << 25

	// ColorSpecial flags a Color whose payload lives outside the normal
	// palette/RGB color space.
	ColorSpecial Color = 1 << 26
)

// Named palette colors. The low byte matches the ECMA-48 / XTerm
// palette index so that NewColor / GetColor lookups round-trip
// against the color-name table.
const (
	ColorBlack = ColorValid + Color(iota)
	ColorMaroon
	ColorGreen
	ColorOlive
	ColorNavy
	ColorPurple
	ColorTeal
	ColorSilver
	ColorGray
	ColorRed
	ColorLime
	ColorYellow
	ColorBlue
	ColorFuchsia
	ColorAqua
	ColorWhite
)

// PaletteColor returns the color for the given ECMA-48 / XTerm palette
// index (0..255).
func PaletteColor(index int) Color {
	return Color(index)&0x00FFFFFF | ColorValid
}

// NewHexColor returns a Color whose payload is the given 24-bit RGB
// hex value.
func NewHexColor(v int32) Color {
	return ColorIsRGB | ColorValid | Color(uint32(v)&0x00FFFFFF)
}

// NewRGBColor returns a Color from r, g, b component values in 0..255.
func NewRGBColor(r, g, b int32) Color {
	return NewHexColor(((r & 0xff) << 16) | ((g & 0xff) << 8) | (b & 0xff))
}

// GetColor returns the Color for the given W3C name, or a hex literal
// of the form "#rrggbb". Returns ColorDefault if the name is unknown.
func GetColor(name string) Color {
	if c, ok := colorNames[name]; ok {
		return c
	}
	if len(name) == 7 && name[0] == '#' {
		if v, err := strconv.ParseInt(name[1:], 16, 32); err == nil {
			return NewHexColor(int32(v))
		}
	}
	return ColorDefault
}

// Valid reports whether c was initialised (i.e. has ColorValid set).
func (c Color) Valid() bool { return c&ColorValid != 0 }

// IsRGB reports whether c carries an RGB payload.
func (c Color) IsRGB() bool {
	return c&(ColorValid|ColorIsRGB) == (ColorValid | ColorIsRGB)
}

// ColorNamesMatching returns a snapshot of the color-name table
// filtered by an optional predicate. When pred is nil all names are
// returned.
func ColorNamesMatching(pred func(name string, c Color) bool) map[string]Color {
	src := GetColorNames()
	if pred == nil {
		return src
	}
	out := make(map[string]Color, len(src))
	for k, v := range src {
		if pred(k, v) {
			out[k] = v
		}
	}
	return out
}
