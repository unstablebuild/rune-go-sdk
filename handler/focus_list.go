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

// FocusList is the handler counterpart to component.FocusList. Its
// element in focus, which the arrow keys, j and k, Home, End, PgUp,
// PgDn, the wheel and left presses move, gets the events that are not
// mouse events first when it is a handler, and Enter selects it. Mouse
// events go to the elements that are handlers as Group routes them.
//
// Use NewFocusList, or Init, to initialize a FocusList.
type FocusList struct {
	component.FocusList
	// OnFocus, when set, is called when the user moves the focus to
	// node.
	OnFocus func(node component.ListNode)
	// OnSelect, when set, is called with the element in focus on Enter
	// and on a left click on it that no element handles.
	OnSelect func(node component.ListNode)

	group         Group
	width, height int
	pressed       component.ListNode
}

// NewFocusList allocates storage for a new FocusList and initializes it.
func NewFocusList() *FocusList {
	l := new(FocusList)
	l.Init()
	return l
}

// Resize satisfies tui.Component.
func (l *FocusList) Resize(width, height int) {
	l.width, l.height = width, height
	l.FocusList.Resize(width, height)
}

// position returns the position of the element at index i of this list.
func (l *FocusList) position(i int) term.Coordinates {
	return term.Coordinates{Y: (i - l.Offset()) * l.ElementHeight()}
}

func (l *FocusList) children() Layout {
	lay := Layout{Width: l.width, Height: l.height}
	i := 0
	for n, ok := l.Front(); ok; n, ok = n.Next() {
		if h, ok := n.Value().(tui.Handler); ok {
			lay.Children = append(lay.Children, Child{
				Handler: h, Position: l.position(i), Width: l.width, Height: l.ElementHeight(), CaptureDrags: true,
			})
		}
		i++
	}
	return lay
}

// focused returns the element in focus as a Child, or false when there
// is none or it is not a handler.
func (l *FocusList) focused() (Child, bool) {
	n, ok := l.Focus()
	if !ok {
		return Child{}, false
	}
	h, ok := n.Value().(tui.Handler)
	return Child{
		Handler: h, Position: l.position(l.FocusOffset()), Width: l.width, Height: l.ElementHeight(),
	}, ok
}

// Handle satisfies tui.Handler.
func (l *FocusList) Handle(ev term.Event) (exit, handled bool) {
	if ev.Type == term.EventMouse {
		return l.handleMouse(ev)
	}
	if c, ok := l.focused(); ok {
		if exit, handled = c.Handle(ev); exit || handled {
			return
		}
	}
	if ev.Type != term.EventKey {
		return false, false
	}
	switch {
	case ev.Key == term.KeyEnter && ev.Mod == 0:
		return false, l.selectFocus()
	case ev.Key == term.KeyArrowUp, ev.Ch == 'k' && ev.Mod == 0:
		return false, l.move(l.FocusUp)
	case ev.Key == term.KeyArrowDown, ev.Ch == 'j' && ev.Mod == 0:
		return false, l.move(l.FocusDown)
	case ev.Key == term.KeyHome:
		return false, l.move(l.FocusStart)
	case ev.Key == term.KeyEnd:
		return false, l.move(l.FocusEnd)
	case ev.Key == term.KeyPgup:
		return false, l.move(func() bool { return l.page(l.FocusUp) })
	case ev.Key == term.KeyPgdn:
		return false, l.move(func() bool { return l.page(l.FocusDown) })
	}
	return false, false
}

func (l *FocusList) page(move func() bool) (ok bool) {
	for range max(1, l.height/l.ElementHeight()) {
		if !move() {
			break
		}
		ok = true
	}
	return ok
}

// move moves the focus with f, and calls OnFocus if it moved to another
// element. It returns what f does.
func (l *FocusList) move(f func() bool) bool {
	before, _ := l.Focus()
	ok := f()
	if n, focus := l.Focus(); ok && focus && n != before {
		l.notifyFocus(n)
	}
	return ok
}

// notifyFocus calls OnFocus, if set, with n.
func (l *FocusList) notifyFocus(n component.ListNode) {
	if l.OnFocus != nil {
		l.OnFocus(n)
	}
}

func (l *FocusList) selectFocus() bool {
	n, ok := l.Focus()
	if !ok || l.OnSelect == nil {
		return false
	}
	l.OnSelect(n)
	return true
}

func (l *FocusList) handleMouse(ev term.Event) (exit, handled bool) {
	pos := term.Coordinates{X: ev.MouseX, Y: ev.MouseY}
	var at component.ListNode
	inside := pos.X >= 0 && pos.Y >= 0 && pos.X < l.width && pos.Y < l.height
	if inside {
		at, inside = l.ElementAt(pos)
	}
	switch ev.Key {
	case term.MouseLeft:
		if inside {
			if focus, _ := l.Focus(); focus != at {
				l.SetFocus(at)
				l.notifyFocus(at)
			}
			if l.pressed == (component.ListNode{}) {
				l.pressed = at
			}
		}
	case term.MouseRelease, 0:
		defer func() { l.pressed = component.ListNode{} }()
	}
	if exit, handled = l.group.Handle(l.children(), ev); exit || handled {
		return
	}
	switch ev.Key {
	case term.MouseLeft:
		return false, inside
	case term.MouseRelease:
		if inside && at == l.pressed {
			return false, l.selectFocus()
		}
	case term.MouseWheelUp:
		return false, l.move(l.FocusUp)
	case term.MouseWheelDown:
		return false, l.move(l.FocusDown)
	}
	return false, false
}

// Cursor satisfies tui.Handler by returning the cursor of the element
// in focus.
func (l *FocusList) Cursor() (pos term.Coordinates, style term.CursorStyle, show bool) {
	c, ok := l.focused()
	if !ok {
		return
	}
	return c.cursor(l.width, l.height)
}

// Selection satisfies tui.Handler by returning the selection of the
// element in focus.
func (l *FocusList) Selection() (string, bool) {
	if c, ok := l.focused(); ok {
		return c.Selection()
	}
	return "", false
}
