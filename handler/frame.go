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

// Frame is a proxy handler that simply draws a frame around
// the underlying handler.
type Frame struct {
	component.Frame
	handler tui.Handler
}

// NewFrame allocates storage for a new Frame and initializes it.
func NewFrame(handler tui.Handler) (f *Frame) {
	f = new(Frame)
	f.Init(handler)
	return
}

// Init initializes this Frame with the given underlying handler
// and frame attributes.
func (f *Frame) Init(handler tui.Handler) {
	f.handler = handler
	f.Frame.Init(handler)
}

// Handle delegates the event to the underlying handler.
func (f *Frame) Handle(ev term.Event) (bool, bool) {
	if ev.Type == term.EventMouse {
		content := f.ContentPosition()
		ev.MouseX -= content.X
		ev.MouseY -= content.Y
		if ev.MouseX < 0 {
			ev.MouseX = 0
		}
		if ev.MouseY < 0 {
			ev.MouseY = 0
		}
	}
	return f.handler.Handle(ev)
}

// Cursor returns the underlying handler's cursor position
// with the frame offset.
func (f *Frame) Cursor() (pos term.Coordinates, style term.CursorStyle, show bool) {
	pos, style, show = f.handler.Cursor()
	content := f.ContentPosition()
	pos.X += content.X
	pos.Y += content.Y
	return
}

// Selection returns the underlying handler's selection.
func (f *Frame) Selection() (string, bool) {
	return f.handler.Selection()
}
