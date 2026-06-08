// Copyright 2026 Unstable Build, LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package mouse

import "github.com/unstablebuild/rune-go-sdk/term"

// Action represents a mouse action.
type Action uint16

// List of mouse actions.
const (
	LeftClick Action = iota
	RightClick
	MiddleClick
	WheelUp
	WheelDown
	Release
	none
)

// Delegate is an interface that wraps callbacks to enable
// scrolling and text selection.
//
// OnAction is provided to extend to other features. It takes precedence
// over builtin features so if it returns true, Mouse won't call
// any other callbacks.
type Delegate interface {
	OnAction(ev term.Event, pos term.Coordinates, action Action) bool
	ScrollUp(n int) bool
	ScrollDown(n int) bool
	SetSelectionEnd(pos term.Coordinates)
	SetSelectionStart(pos term.Coordinates)
	ClearSelection()
	SelectWordAt(pos term.Coordinates)
	SelectLine(y int)
	Width() int
	Height() int
}
