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
)

// TestHandler is a handler used to test composite handlers. Each event
// processed by this handler increments the Ch rune to the next rune.
type TestHandler struct {
	component.TestComponent
	CursorPos      term.Coordinates
	CursorStyle    term.CursorStyle
	Exit           bool
	Handled        bool
	HandleOverride func(term.Event) (bool, bool)
}

// NewTestHandler will allocate storage for a new handler and initialize it
func NewTestHandler() (t *TestHandler) {
	t = new(TestHandler)
	t.Ch = 'A'
	return
}

// Handle the next Event
func (t *TestHandler) Handle(ev term.Event) (bool, bool) {
	if t.HandleOverride != nil {
		return t.HandleOverride(ev)
	}
	// signal that we handled the event
	t.Ch++
	return t.Exit, t.Handled
}

// Cursor returns the set CursorPos.
func (t *TestHandler) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	if t.CursorPos == (term.Coordinates{}) {
		return term.Coordinates{X: -1, Y: -1}, 0, false
	}
	return t.CursorPos, t.CursorStyle, true
}

// Selection satisfies tui.Handler.
func (t *TestHandler) Selection() (string, bool) {
	return string(t.Ch), true
}

// TestFloating is a testing Handler.
type TestFloating struct {
	TestHandler
	width, height int
}

// NewTestFloating allocates storage for a new TestHandler and initializes it.
func NewTestFloating(width, height int) *TestFloating {
	ret := new(TestFloating)
	ret.TestHandler = *NewTestHandler()
	ret.width = width
	ret.height = height
	return ret
}

// Dimensions returns the floating component dimensions.
func (t *TestFloating) Dimensions() (width, height int) {
	return t.width, t.height
}
