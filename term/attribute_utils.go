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

import (
	"github.com/unstablebuild/tcell/v3"
)

// AttributesDifference computes the set difference between a and b,
// that is it returns a set of attributes that contain
// all the bit flags set in a but not set in b, and returns
// ColorDefault if a's color is equal to b's color or returns
// the color set in a.
func AttributesDifference(a, b Attributes) Attributes {
	retFgColor := a.Fg
	retBgColor := a.Bg
	if a.Fg == b.Fg {
		retFgColor = tcell.ColorDefault
	}
	if a.Bg == b.Bg {
		retBgColor = tcell.ColorDefault
	}
	return Attributes{Fg: retFgColor, Bg: retBgColor, Attrs: (a.Attrs &^ b.Attrs)}
}

// AttributesUnion computes the set union between a and b,
// that is it returns a set of attributes that contain
// all the bit flags set in a, b or both, and uses the color
// defined in b or if not set, uses the color in a.
func AttributesUnion(a, b Attributes) Attributes {
	retFgColor := b.Fg
	retBgColor := b.Bg
	if b.Fg == tcell.ColorDefault {
		retFgColor = a.Fg
	}
	if b.Bg == tcell.ColorDefault {
		retBgColor = a.Bg
	}
	return Attributes{Fg: retFgColor, Bg: retBgColor, Attrs: (a.Attrs | b.Attrs)}
}
