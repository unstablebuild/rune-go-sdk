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

package tui

import (
	"github.com/unstablebuild/rune-go-sdk/term"
)

// Component represents a visual element that can be drawn
// in a text-based user interface. It wraps the basic Draw and Resize methods.
type Component interface {
	// Resize is used by clients to indicate the virtual space available
	// for this component to be drawn in subsequent calls to Draw.
	// When a component is initialized, its width and height is 0 until Resize is
	// called to set the appropriate dimensions.
	Resize(width, height int)
	// Draw draws this component to the underlying Writer. It returns non-nil error
	// if something went wrong in the process of writing or the writer returned
	// an error.
	Draw(term.Writer)
}

// Handler builds upon Component to add event-handling behavior.
// It wraps the basic Handle, Cursor and Man methods.
type Handler interface {
	Component
	// Handle abstracts the ability to handle input events. These events could
	// be key presses or other types of events. See tcell's documentation for more
	// information. Handle returns true if a handler is done processing events.
	// Clients can have multiple handlers in the same interface so
	// this method will be called only when a handler is in focus.
	Handle(term.Event) (exit, handled bool)
	// Cursor should return the cursor coordinates, style and whether it should be shown at all.
	Cursor() (c term.Coordinates, s term.CursorStyle, show bool)

	// Selection returns the selected text and true if there's currently any, or
	// an empty string and false if there's no text selected.
	Selection() (string, bool)
}
