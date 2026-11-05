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
	"reflect"

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Child is a handler laid out at Position, in the coordinates of the
// handler that contains it, with Width and Height.
type Child struct {
	tui.Handler
	Position      term.Coordinates
	Width, Height int
	// CaptureDrags makes the child own the whole gesture of a left
	// press on it: its drags and its release reach the child wherever
	// the pointer goes, with coordinates past its bounds once the
	// pointer leaves it, and drags that start on other children never
	// reach it. Children that tell clicks from drags need it.
	CaptureDrags bool
}

func (c Child) contains(pos term.Coordinates) bool {
	pos = term.CoordinatesDiff(pos, c.Position)
	return pos.X >= 0 && pos.Y >= 0 && pos.X < c.Width && pos.Y < c.Height
}

// Layout is the children of a handler that is Width wide and Height
// tall, in the order they are drawn. Children may lie partly or wholly
// outside of it, as those scrolled out of view do.
type Layout struct {
	Width, Height int
	Children      []Child
}

// at returns the index of the child drawn last at pos, or -1 when pos is
// outside of the layout or no child is there.
func (l Layout) at(pos term.Coordinates) int {
	if pos.X < 0 || pos.Y < 0 || pos.X >= l.Width || pos.Y >= l.Height {
		return -1
	}
	for i := len(l.Children) - 1; i >= 0; i-- {
		if c := l.Children[i]; c.Handler != nil && c.contains(pos) {
			return i
		}
	}
	return -1
}

// Group routes the events of a handler to its children and keeps one of
// them in focus, for handlers that contain several. Its methods take
// the children as laid out at the time they are called, so a handler
// that embeds a Group need not tell it when they change. A child keeps
// the focus while it is in the layout, which for children that are not
// pointers means while it keeps its index.
//
// The zero value is ready to use.
type Group struct {
	focus, pressed, hovered member
	// held reports that a left press came and no release or motion has
	// since; captured, that the child it came on, pressed, captures it.
	held, captured bool
}

// member tracks a child of a Group through layouts.
type member struct {
	h  tui.Handler
	i  int
	ok bool
}

func (m *member) set(cs []Child, i int) {
	*m = member{h: cs[i].Handler, i: i, ok: true}
}

// find returns the index of m in cs, or -1, forgetting m, when it is no
// longer there.
func (m *member) find(cs []Child) int {
	if !m.ok {
		return -1
	}
	if !isPointer(m.h) {
		if m.i < len(cs) {
			return m.i
		}
	} else {
		if m.i < len(cs) && cs[m.i].Handler == m.h {
			return m.i
		}
		for i, c := range cs {
			if c.Handler == m.h {
				m.i = i
				return i
			}
		}
	}
	*m = member{}
	return -1
}

func isPointer(h tui.Handler) bool {
	return h != nil && reflect.TypeOf(h).Kind() == reflect.Pointer
}

// same reports whether a and b are equal handlers, without the panic ==
// has for handlers that cannot be compared.
func same(a, b tui.Handler) bool {
	return a != nil && b != nil && reflect.TypeOf(a) == reflect.TypeOf(b) &&
		reflect.ValueOf(a).Comparable() && a == b
}

// Handle routes ev to a child of l.
//
// Events that are not mouse events go to the child in focus, which only
// SetFocus moves, and are left unhandled when there is none.
//
// Mouse events go to the child under the pointer, with coordinates
// relative to it, except for the drags and releases of the left presses
// on children with CaptureDrags set. When the pointer leaves a child,
// the child receives a motion outside it so it drops whatever hover it
// shows; not while a left press is held.
func (g *Group) Handle(l Layout, ev term.Event) (exit, handled bool) {
	if ev.Type == term.EventMouse {
		return g.handleMouse(l, ev)
	}
	i := g.focus.find(l.Children)
	if i < 0 {
		return false, false
	}
	return l.Children[i].Handle(ev)
}

func (g *Group) handleMouse(l Layout, ev term.Event) (exit, handled bool) {
	pos := term.Coordinates{X: ev.MouseX, Y: ev.MouseY}
	hit := l.at(pos)
	switch ev.Key {
	case term.MouseLeft:
		if g.held {
			return g.drag(l, hit, ev)
		}
		g.held = true
		g.pointerTo(l, hit, pos)
		g.pressed, g.captured = member{}, false
		if hit >= 0 {
			g.pressed.set(l.Children, hit)
			g.captured = l.Children[hit].CaptureDrags
		}
	case term.MouseRelease:
		if g.held {
			g.held = false
			exit, handled = g.drag(l, hit, ev)
			g.pressed, g.captured = member{}, false
			return exit, handled
		}
		g.pointerTo(l, hit, pos)
	case 0:
		// A motion means no button is held, so a release that never
		// reached this group cannot hold on to the next press.
		g.held = false
		g.pressed, g.captured = member{}, false
		g.pointerTo(l, hit, pos)
	default:
		if !g.held {
			g.pointerTo(l, hit, pos)
		}
	}
	if hit < 0 {
		return false, false
	}
	return deliver(l.Children[hit], ev)
}

// drag routes a drag or the release of the press held to the child
// that captures it, or to the child at index hit unless that child
// captures the presses on it.
func (g *Group) drag(l Layout, hit int, ev term.Event) (exit, handled bool) {
	if g.captured {
		if i := g.pressed.find(l.Children); i >= 0 {
			return deliver(l.Children[i], ev)
		}
		return false, false
	}
	if hit < 0 || l.Children[hit].CaptureDrags {
		return false, false
	}
	return deliver(l.Children[hit], ev)
}

// pointerTo records that the pointer at pos is over the child at index
// next, -1 for none, and sends the child it was over before a motion
// outside it. The motion is at the pointer, or above and left of the
// child where children overlap.
func (g *Group) pointerTo(l Layout, next int, pos term.Coordinates) {
	prev := g.hovered.find(l.Children)
	if next < 0 {
		g.hovered = member{}
	} else {
		g.hovered.set(l.Children, next)
	}
	if prev < 0 || prev == next {
		return
	}
	c := l.Children[prev]
	if c.contains(pos) {
		pos = term.Coordinates{X: c.Position.X - 1, Y: c.Position.Y - 1}
	}
	_, _ = deliver(c, term.Event{Type: term.EventMouse, MouseX: pos.X, MouseY: pos.Y})
}

func deliver(c Child, ev term.Event) (exit, handled bool) {
	ev.MouseX -= c.Position.X
	ev.MouseY -= c.Position.Y
	return c.Handle(ev)
}

// Cursor returns the cursor of the child of l in focus, offset by its
// position, and hides it when it falls outside of l.
func (g *Group) Cursor(l Layout) (pos term.Coordinates, style term.CursorStyle, show bool) {
	i := g.focus.find(l.Children)
	if i < 0 {
		return
	}
	return l.Children[i].cursor(l.Width, l.Height)
}

// cursor returns the cursor of c offset by its position, hidden when it
// falls outside of the width by height handler that contains c.
func (c Child) cursor(width, height int) (pos term.Coordinates, style term.CursorStyle, show bool) {
	pos, style, show = c.Cursor()
	pos.X += c.Position.X
	pos.Y += c.Position.Y
	if pos.X < 0 || pos.Y < 0 || pos.X >= width || pos.Y >= height {
		show = false
	}
	return
}

// Selection returns the selection of the child of l in focus.
func (g *Group) Selection(l Layout) (string, bool) {
	i := g.focus.find(l.Children)
	if i < 0 {
		return "", false
	}
	return l.Children[i].Selection()
}

// Focus returns the index in l of the child in focus, or -1 for none.
func (g *Group) Focus(l Layout) int {
	return g.focus.find(l.Children)
}

// SetFocus moves the focus to the child of l at index i, or takes it
// from the child that has it when i is out of range.
func (g *Group) SetFocus(l Layout, i int) {
	if i < 0 || i >= len(l.Children) {
		g.focus = member{}
		return
	}
	g.focus.set(l.Children, i)
}

// routed implements the event methods of tui.Handler, and Focus and
// SetFocus, for a handler whose children layout returns, routing them
// with a Group.
type routed struct {
	group  Group
	layout func() Layout
}

func (r *routed) lay() Layout {
	if r.layout == nil {
		return Layout{}
	}
	return r.layout()
}

// Handle satisfies tui.Handler. See Group.Handle.
func (r *routed) Handle(ev term.Event) (exit, handled bool) {
	return r.group.Handle(r.lay(), ev)
}

// Cursor satisfies tui.Handler.
func (r *routed) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	return r.group.Cursor(r.lay())
}

// Selection satisfies tui.Handler.
func (r *routed) Selection() (string, bool) { return r.group.Selection(r.lay()) }

// Focus returns the child in focus, which gets the events that are not
// mouse events, or false for none.
func (r *routed) Focus() (tui.Handler, bool) {
	l := r.lay()
	i := r.group.Focus(l)
	if i < 0 {
		return nil, false
	}
	return l.Children[i].Handler, true
}

// SetFocus moves the focus to h and returns true, or returns false,
// leaving the focus where it was, when h is not a child laid out.
func (r *routed) SetFocus(h tui.Handler) bool {
	l := r.lay()
	for i, c := range l.Children {
		if same(c.Handler, h) {
			r.group.SetFocus(l, i)
			return true
		}
	}
	return false
}
