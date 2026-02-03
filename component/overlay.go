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

// Overlay overlays one component over another one, with potentially padding
// and/or alignment, specified by SpanConfig.
type Overlay struct {
	background tui.Component
	Span
}

// NewOverlay allocates storage for a new overlay and initializes it.
func NewOverlay(
	background, cover tui.Component, backAttr term.Attributes, config SpanConfig,
) *Overlay {
	ret := new(Overlay)
	ret.Init(background, cover, backAttr, config)
	return ret
}

// Init initializes this overlay with the given background, cover and configuration.
func (o *Overlay) Init(
	background, cover tui.Component, backAttr term.Attributes, config SpanConfig,
) {
	// clean cells before drawing on top
	cover = WithBackground(cover, term.Cell{Attributes: backAttr})

	o.Span.Init(cover, config)
	o.background = background
}

// Resize satisfies tui.Component.
func (o *Overlay) Resize(width, height int) {
	o.background.Resize(width, height)
	o.Span.Resize(width, height)
}

// Draw satisfies tui.Component.
func (o *Overlay) Draw(w term.Writer) {
	o.background.Draw(w)
	o.Span.Draw(w)
}
