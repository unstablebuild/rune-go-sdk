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

// Container is the handler counterpart to component.Container: its
// rows are Rows, so it routes events to the handlers in them as Group
// does. A wheel event no handler takes scrolls it, and it scrolls to
// show the row in focus when SetFocus moves the focus or a key reaches
// it.
//
// Use NewContainer, or Init, to initialize a Container.
type Container struct {
	component.Container
	routed
	rows []*Row
}

var _ Responsive = (*Container)(nil)

// NewContainer allocates storage for a new Container and initializes it.
func NewContainer() *Container {
	c := new(Container)
	c.Init()
	return c
}

// Init initializes this Container.
func (c *Container) Init() {
	c.layout = c.children
}

// AddRow appends a new Row to this Container.
func (c *Container) AddRow() *Row {
	r := NewRow()
	c.AppendRow(&r.Row)
	c.rows = append(c.rows, r)
	return r
}

func (c *Container) children() Layout {
	width, height := c.Size()
	l := Layout{Width: width, Height: height, Children: make([]Child, len(c.rows))}
	y := -c.Offset()
	for i, r := range c.rows {
		rowHeight := r.Height(width)
		l.Children[i] = Child{
			Handler: r, Position: term.Coordinates{Y: y}, Width: width, Height: rowHeight, CaptureDrags: true,
		}
		y += rowHeight
	}
	return l
}

// Handle satisfies tui.Handler.
func (c *Container) Handle(ev term.Event) (exit, handled bool) {
	exit, handled = c.routed.Handle(ev)
	switch {
	case exit:
	case ev.Type != term.EventMouse:
		if handled {
			c.reveal()
		}
	case handled:
	case ev.Key == term.MouseWheelUp:
		handled = c.ScrollUp() && c.relayout()
	case ev.Key == term.MouseWheelDown:
		handled = c.ScrollDown() && c.relayout()
	}
	return
}

// relayout lays the rows out again after a scroll, so they route the
// events that come before the next Draw. It returns true.
func (c *Container) relayout() bool {
	c.Resize(c.Size())
	return true
}

// Focus returns the handler in focus, the one in focus in the row in
// focus, or false for none.
func (c *Container) Focus() (tui.Handler, bool) {
	row, ok := c.routed.Focus()
	if !ok {
		return nil, false
	}
	return row.(*Row).Focus()
}

// SetFocus moves the focus to h, a handler added to one of the rows of
// this Container, and returns true, or returns false, leaving the focus
// where it was, when no row has h.
func (c *Container) SetFocus(h tui.Handler) bool {
	for _, r := range c.rows {
		if r.SetFocus(h) {
			c.routed.SetFocus(r)
			c.reveal()
			return true
		}
	}
	return false
}

// reveal scrolls this container so the row in focus shows, or its top
// if it is taller than the container.
func (c *Container) reveal() {
	l := c.children()
	i := c.group.Focus(l)
	if i < 0 {
		return
	}
	row := l.Children[i]
	offset := c.Offset()
	switch {
	case row.Position.Y < 0:
		offset += row.Position.Y
	case row.Position.Y+row.Height > l.Height:
		offset += min(row.Position.Y, row.Position.Y+row.Height-l.Height)
	default:
		return
	}
	c.ScrollTo(offset)
	c.relayout()
}
