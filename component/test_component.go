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
)

var _ WithAttributes = (*TestComponent)(nil)

// TestComponent draws rune Ch, and attributes Bg, Fg on every cell
// available. This component is used for testing or debugging.
type TestComponent struct {
	Ch rune
	term.Attributes
	width, height int
}

// Resize satisfies tui.Component.
func (t *TestComponent) Resize(width, height int) {
	t.width, t.height = width, height
}

// Draw satisfies tui.Component.
func (t *TestComponent) Draw(w term.Writer) {
	if t.Ch == 0 {
		return
	}
	for tx := t.width - 1; tx >= 0; tx-- {
		for ty := 0 + t.height - 1; ty >= 0; ty-- {
			w.SetCell(term.Coordinates{X: tx, Y: ty},
				term.Cell{Width: 1, Ch: t.Ch, Attributes: t.Attributes})
		}
	}
}

var _ Responsive = (*TestResponsive)(nil)
var _ Floating = (*TestResponsive)(nil)

// SetAttr satisfies WithAttributes
func (t *TestComponent) SetAttr(attr term.Attributes) (ret term.Attributes) {
	ret = t.Attributes
	t.Attributes = attr
	return
}

// TestResponsive is a Responsive and Floating for testing.
type TestResponsive struct {
	TestComponent
	PassedWidth int
	WantHeight  int
	WantWidth   int
}

// Height satisfies Responsive.
func (t *TestResponsive) Height(width int) int {
	t.PassedWidth = width
	return t.WantHeight
}

// Dimensions satisfies Floating.
func (t *TestResponsive) Dimensions() (width, height int) {
	return t.WantWidth, t.WantHeight
}
