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

// List is the handler counterpart to component.List: the elements
// pushed with its methods that are handlers get events as Group routes
// them. A wheel event no element takes seeks it, and it seeks to show
// the element in focus when SetFocus moves the focus or a key reaches
// it.
//
// Use NewList, or Init, to initialize a List.
type List struct {
	component.List
	routed
}

// NewList allocates storage for a new List and initializes it.
func NewList(elementHeight int) *List {
	l := new(List)
	l.Init(elementHeight)
	return l
}

// Init initializes this List.
func (l *List) Init(elementHeight int) {
	l.List.Init(elementHeight)
	l.layout = l.children
}

func (l *List) children() Layout {
	lay := Layout{Width: l.Width(), Height: l.Height()}
	eh, y := l.ElementHeight(), -l.Offset()*l.ElementHeight()
	for n, ok := l.Front(); ok; n, ok = n.Next() {
		if h, ok := n.Value().(tui.Handler); ok {
			lay.Children = append(lay.Children, Child{
				Handler: h, Position: term.Coordinates{Y: y}, Width: lay.Width, Height: eh, CaptureDrags: true,
			})
		}
		y += eh
	}
	return lay
}

// Handle satisfies tui.Handler.
func (l *List) Handle(ev term.Event) (exit, handled bool) {
	exit, handled = l.routed.Handle(ev)
	switch {
	case exit:
	case ev.Type != term.EventMouse:
		if handled {
			l.reveal()
		}
	case handled:
	case ev.Key == term.MouseWheelUp:
		handled = l.SeekUp()
	case ev.Key == term.MouseWheelDown:
		handled = l.SeekDown()
	}
	return
}

// SetFocus moves the focus to h, an element of this List, and returns
// true, or returns false, leaving the focus where it was, when h is not
// one.
func (l *List) SetFocus(h tui.Handler) bool {
	ok := l.routed.SetFocus(h)
	if ok {
		l.reveal()
	}
	return ok
}

// reveal seeks this list so the element in focus shows.
func (l *List) reveal() {
	lay := l.children()
	i := l.group.Focus(lay)
	if i < 0 {
		return
	}
	c := lay.Children[i]
	// n is the index of the element in focus among those that show.
	n, visible := c.Position.Y/c.Height, lay.Height/c.Height
	switch {
	case n < 0:
		for range -n {
			l.SeekUp()
		}
	case visible > 0 && n >= visible:
		for range n - visible + 1 {
			l.SeekDown()
		}
	}
}
