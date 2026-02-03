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

// Virtual wraps another tui.Handler to provide cursor event coordinates.
type Virtual[T tui.Handler] struct {
	component.Virtual[T]
}

// Handle satisfies tui.Handler.
func (v *Virtual[T]) Handle(ev term.Event) (bool, bool) {
	if ev.Type == term.EventMouse {
		offset := v.Position()
		ev.MouseX -= offset.X
		ev.MouseY -= offset.Y
	}
	return v.C.Handle(ev)
}

// Cursor satisfies tui.Handler.
func (v *Virtual[T]) Cursor() (pos term.Coordinates, style term.CursorStyle, show bool) {
	pos, style, show = v.C.Cursor()
	offset := v.Position()
	pos.X += offset.X
	pos.Y += offset.Y
	return
}

// Selection satisfies tui.Handler.
func (v *Virtual[T]) Selection() (string, bool) {
	return v.C.Selection()
}
