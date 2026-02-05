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

import (
	"time"

	"github.com/unstablebuild/rune-go-sdk/term"
)

// Mouse handles mouse events to provide text selection and scrolling
// functionality to a text tui.Handler. It can be extended by
// providing an OnAction callback. See MouseDelegate for more details.
type Mouse struct {
	delegate           Delegate
	mousePos           term.Coordinates
	mousePressedLeft   bool
	mousePressedMiddle bool
	mousePressedRight  bool
	lastClick          time.Time
	clickCount         int
}

// TODO should be defaults. add config or options
// time allowed between mouse clicks to chain them into e.g. double-click
const clickChainWindow = 500 * time.Millisecond

// New allocates storage for a Mouse and initializes it.
func New(d Delegate) *Mouse {
	ret := new(Mouse)
	ret.Init(d)
	return ret
}

// Init initializes this Mouse with d and resets all its internal state.
func (m *Mouse) Init(d Delegate) {
	m.delegate = d
	m.mousePos = term.Coordinates{}
	m.mousePressedLeft = false
	m.mousePressedRight = false
	m.mousePressedMiddle = false
	m.lastClick = time.Time{}
	m.clickCount = 0
}

// Handle satisfies tui.Handler.
func (h *Mouse) Handle(ev term.Event) (exit, handled bool) {
	defer func() {
		h.mousePressedLeft = ev.Key == term.MouseLeft
		h.mousePressedMiddle = ev.Key == term.MouseMiddle
		h.mousePressedRight = ev.Key == term.MouseRight
	}()

	if ev.Type != term.EventMouse || ev.Mod != 0 {
		return
	}

	pos := term.Coordinates{X: ev.MouseX, Y: ev.MouseY}
	pressedLeft := ev.Key == term.MouseLeft && !h.mousePressedLeft
	pressedMiddle := ev.Key == term.MouseMiddle && !h.mousePressedMiddle
	pressedRight := ev.Key == term.MouseRight && !h.mousePressedRight
	wheelUp := ev.Key == term.MouseWheelUp
	wheelDown := ev.Key == term.MouseWheelDown
	released := ev.Key == term.MouseRelease

	var action Action
	switch true {
	case pressedLeft:
		action = LeftClick
	case pressedMiddle:
		action = MiddleClick
	case pressedRight:
		action = RightClick
	case released:
		action = Release
	case wheelUp:
		action = WheelUp
	case wheelDown:
		action = WheelDown
	default:
		action = none
	}

	// we do not want to dispatch on drags with mouseNone
	// because it could confuse clients to think that we
	// are able to dispatch simply on mouse moves.
	if action != none {
		handled = h.delegate.OnAction(ev, pos, action)
		if handled {
			return
		}
	}

	switch ev.Key {
	case term.MouseWheelUp:
		handled = h.delegate.ScrollUp(1)
	case term.MouseWheelDown:
		handled = h.delegate.ScrollDown(1)
	case term.MouseLeft:
		handled = h.handleLeftClickSelect(pos)
	}
	return
}

func (h *Mouse) handleLeftClickSelect(pos term.Coordinates) (handled bool) {
	if h.mousePressedLeft { /* drag */
		handled = true
		h.delegate.SetSelectionEnd(pos)
		if pos.Y < 4 {
			h.delegate.ScrollUp(1)
		} else if pos.Y > h.delegate.Height()-4 {
			h.delegate.ScrollDown(1)
		}
		return
	}

	if h.clickCount == 0 || time.Since(h.lastClick) < clickChainWindow {
		h.clickCount++
	} else {
		h.clickCount = 1
	}

	h.lastClick = time.Now()
	handled = true
	switch h.clickCount {
	case 1:
		h.delegate.ClearSelection()
		h.delegate.SetSelectionStart(pos)
	case 2:
		h.delegate.SelectWordAt(pos)
	default:
		h.delegate.SelectLine(pos.Y)
	}
	return
}
