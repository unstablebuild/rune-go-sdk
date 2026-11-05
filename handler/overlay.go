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

package handler

import (
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Overlay is the handler counterpart to component.Overlay: it routes
// events to its cover and the background under it as Group does, with
// the cover in focus until SetFocus moves it.
//
// Use NewOverlay, or Init, to initialize an Overlay.
type Overlay struct {
	component.Overlay
	routed
	background, cover tui.Handler
	width, height     int
}

// NewOverlay allocates storage for a new Overlay and initializes it.
func NewOverlay(
	background, cover tui.Handler, backAttr term.Attributes, config component.SpanConfig,
) *Overlay {
	o := new(Overlay)
	o.Init(background, cover, backAttr, config)
	return o
}

// Init initializes this Overlay with cover over background, placed by
// config on a backAttr background.
func (o *Overlay) Init(
	background, cover tui.Handler, backAttr term.Attributes, config component.SpanConfig,
) {
	o.Overlay.Init(background, cover, backAttr, config)
	o.background, o.cover = background, cover
	o.layout = o.children
	o.SetFocus(cover)
}

// Resize satisfies tui.Component.
func (o *Overlay) Resize(width, height int) {
	o.width, o.height = width, height
	o.Overlay.Resize(width, height)
}

func (o *Overlay) children() Layout {
	width, height := o.ContentSize()
	return Layout{Width: o.width, Height: o.height, Children: []Child{
		{Handler: o.background, Width: o.width, Height: o.height, CaptureDrags: true},
		{Handler: o.cover, Position: o.ContentOffset(), Width: width, Height: height, CaptureDrags: true},
	}}
}
