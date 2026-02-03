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
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Wrap wraps a tui.Handler with a fn that gets called
// instead of Handle called on h. The rest of tui.Handler
// methods are delegated directly to h, so if h.Handle needs to be
// called, it's the client's responsibility to do so.
//
// Additionally to the usual term.Events, term.EventResize events
// are also dispatched as events to fn, after Resize has been called on h.
func Wrap(h tui.Handler, fn func(term.Event) (bool, bool)) tui.Handler {
	return wrapHandler{h: h, fn: fn}
}

type wrapHandler struct {
	h  tui.Handler
	fn func(term.Event) (bool, bool)
}

func (n wrapHandler) Resize(width, height int) {
	n.h.Resize(width, height)
	n.fn(term.Event{Type: term.EventResize, Width: width, Height: height})
}

func (n wrapHandler) Draw(w term.Writer) {
	n.h.Draw(w)
}

func (n wrapHandler) Handle(ev term.Event) (exit, handled bool) {
	return n.fn(ev)
}

func (n wrapHandler) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	return n.h.Cursor()
}

func (n wrapHandler) Selection() (string, bool) {
	return n.h.Selection()
}
