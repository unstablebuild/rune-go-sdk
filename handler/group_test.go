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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

func TestGroupHandle(t *testing.T) {
	focus := func(i int) func(*grouped, *rec) {
		return func(g *grouped, _ *rec) { g.SetFocus(g.layout, i) }
	}
	runHandleCases(t, newGrouped, []handleCase[*grouped]{
		{
			name:    "keys with no child in focus are left unhandled",
			events:  []term.Event{key('x'), special(term.KeyTab)},
			want:    []string{"-", "-"},
			handled: []bool{false, false},
		},
		{
			name:    "keys go to the child in focus",
			setup:   focus(1),
			events:  []term.Event{key('x'), key('n'), key('q')},
			want:    []string{"B key x", "B key n", "B key q"},
			handled: []bool{true, false, true},
			exit:    []bool{false, false, true},
		},
		{
			name:    "events that are neither keys nor mouse events go to the child in focus",
			setup:   focus(0),
			events:  []term.Event{paste()},
			want:    []string{"A paste"},
			handled: []bool{false},
		},
		{
			name:    "a press goes to the child under it, relative to it",
			events:  []term.Event{press(5, 1)},
			want:    []string{"B left 1,1"},
			handled: []bool{true},
		},
		{
			name:    "a press the child leaves unhandled is unhandled",
			events:  []term.Event{press(9, 0)},
			want:    []string{"C left 1,0"},
			handled: []bool{false},
		},
		{
			name:    "a press outside every child reaches none",
			events:  []term.Event{press(10, 1), release(10, 1), press(-1, 0), release(-1, 0)},
			want:    []string{"-", "-", "-", "-"},
			handled: []bool{false, false, false, false},
		},
		{
			name:   "a press leaves the focus where it was",
			setup:  focus(1),
			events: []term.Event{press(1, 1), release(1, 1), key('x')},
			want:   []string{"A left 1,1", "A release 1,1", "B key x"},
		},
		{
			name:    "a capturing child keeps the drags and the release, past its bounds",
			events:  []term.Event{press(1, 1), press(6, 2), release(12, -1)},
			want:    []string{"A left 1,1", "A left 6,2", "A release 12,-1"},
			handled: []bool{true, true, true},
		},
		{
			name:    "drags from a child that does not capture route by position",
			events:  []term.Event{press(5, 1), press(9, 1), release(6, 0)},
			want:    []string{"B left 1,1", "C left 1,1", "B release 2,0"},
			handled: []bool{true, false, true},
		},
		{
			name:    "drags from other children never reach a capturing child",
			events:  []term.Event{press(5, 1), press(1, 1), release(1, 1)},
			want:    []string{"B left 1,1", "-", "-"},
			handled: []bool{true, false, false},
		},
		{
			name:   "drags from outside never reach a capturing child",
			events: []term.Event{press(20, 20), press(1, 1), release(1, 1)},
			want:   []string{"-", "-", "-"},
		},
		{
			name:   "drags and releases off every child reach none",
			events: []term.Event{press(5, 1), press(20, 1), release(20, 1)},
			want:   []string{"B left 1,1", "-", "-"},
		},
		{
			name:   "a release with no press routes by position",
			events: []term.Event{release(5, 1)},
			want:   []string{"B release 1,1"},
		},
		{
			name:   "the child the pointer leaves gets a motion outside it",
			events: []term.Event{move(1, 1), move(2, 2), move(5, 1)},
			want:   []string{"A move 1,1", "A move 2,2", "A move 5,1", "B move 1,1"},
		},
		{
			name:    "the pointer leaving every child",
			events:  []term.Event{move(1, 1), move(20, 1), move(21, 1)},
			want:    []string{"A move 1,1", "A move 20,1", "-"},
			handled: []bool{true, false, false},
		},
		{
			name:   "no leave while a press is held, and one once it is over",
			events: []term.Event{press(1, 1), press(5, 1), release(5, 1), move(5, 1)},
			want:   []string{"A left 1,1", "A left 5,1", "A release 5,1", "A move 5,1", "B move 1,1"},
		},
		{
			name:   "a press on another child is a leave too",
			events: []term.Event{move(1, 1), press(5, 1)},
			want:   []string{"A move 1,1", "A move 5,1", "B left 1,1"},
		},
		{
			name:   "a motion ends a press whose release never came",
			events: []term.Event{press(1, 1), move(5, 1), press(5, 1)},
			want:   []string{"A left 1,1", "A move 5,1", "B move 1,1", "B left 1,1"},
		},
		{
			name:    "the wheel routes by position while a press is held, and moves no hover",
			events:  []term.Event{press(1, 1), wheelDown(5, 1), release(5, 1)},
			want:    []string{"A left 1,1", "B wheeldown 1,1", "A release 5,1"},
			handled: []bool{true, false, true},
		},
		{
			name:   "the wheel moves the hover when no press is held",
			events: []term.Event{move(1, 1), wheelUp(5, 1)},
			want:   []string{"A move 1,1", "A move 5,1", "B wheelup 1,1"},
		},
		{
			name:   "a capturing child gone mid-press takes the rest of the gesture with it",
			events: []term.Event{press(1, 1), press(5, 1), release(5, 1)},
			before: map[int]func(*grouped, *rec){
				1: func(g *grouped, _ *rec) { g.layout.Children = g.layout.Children[1:] },
			},
			want: []string{"A left 1,1", "-", "-"},
		},
		{
			name:   "a hovered child gone gets no leave",
			events: []term.Event{move(1, 1), move(5, 1)},
			before: map[int]func(*grouped, *rec){
				1: func(g *grouped, _ *rec) { g.layout.Children = g.layout.Children[1:] },
			},
			want: []string{"A move 1,1", "B move 1,1"},
		},
		{
			name:   "a child that is a pointer keeps the focus wherever it moves in the layout",
			setup:  focus(0),
			events: []term.Event{key('x'), key('y')},
			before: map[int]func(*grouped, *rec){
				0: func(g *grouped, _ *rec) {
					cs := g.layout.Children
					cs[0], cs[2] = cs[2], cs[0]
				},
				1: func(g *grouped, _ *rec) { g.layout.Children = g.layout.Children[:2] },
			},
			want: []string{"A key x", "-"},
		},
		{
			name:    "a child that is not a pointer keeps the focus by its index",
			fixture: newByValueGrouped,
			setup:   focus(1),
			events:  []term.Event{key('x'), key('y'), key('z')},
			before: map[int]func(*grouped, *rec){
				1: func(g *grouped, _ *rec) {
					cs := g.layout.Children
					cs[0], cs[1] = cs[1], cs[0]
				},
				2: func(g *grouped, _ *rec) { g.layout.Children = g.layout.Children[:1] },
			},
			want: []string{"V1 key x", "V0 key y", "-"},
		},
		{
			name:    "where children overlap, the one laid out last gets the pointer",
			fixture: newOverlapGrouped,
			events:  []term.Event{press(2, 1), release(2, 1), press(1, 1)},
			want:    []string{"O left 0,0", "O release 0,0", "O move -1,0", "A left 1,1"},
		},
		{
			name:    "a leave onto a child over the one left is above and left of it",
			fixture: newOverlapGrouped,
			events:  []term.Event{move(1, 1), move(2, 1)},
			want:    []string{"A move 1,1", "A move -1,-1", "O move 0,0"},
		},
	})
}

func TestGroupViews(t *testing.T) {
	focus := func(i int) func(*grouped, *rec) {
		return func(g *grouped, _ *rec) { g.SetFocus(g.layout, i) }
	}
	runViewCases(t, newGrouped, []viewCase[*grouped]{
		{name: "no child in focus shows no cursor and selects nothing"},
		{
			name:      "the cursor of the child in focus, offset by its position",
			setup:     focus(1),
			cursor:    term.Coordinates{X: 5},
			show:      true,
			selection: "B",
		},
		{
			name: "a cursor right of the layout is hidden",
			setup: func(g *grouped, r *rec) {
				r.leaf("B").cur = term.Coordinates{X: 6}
				g.SetFocus(g.layout, 1)
			},
			selection: "B",
		},
		{
			name: "a cursor above the layout is hidden",
			setup: func(g *grouped, r *rec) {
				r.leaf("A").cur = term.Coordinates{Y: -1}
				g.SetFocus(g.layout, 0)
			},
			selection: "A",
		},
		{
			name: "a child that hides its cursor",
			setup: func(g *grouped, r *rec) {
				r.leaf("C").hide = true
				g.SetFocus(g.layout, 2)
			},
			selection: "C",
		},
		{
			name: "an index past the children takes the focus away",
			setup: func(g *grouped, _ *rec) {
				g.SetFocus(g.layout, 1)
				g.SetFocus(g.layout, 3)
			},
		},
		{
			name: "a negative index takes the focus away",
			setup: func(g *grouped, _ *rec) {
				g.SetFocus(g.layout, 1)
				g.SetFocus(g.layout, -1)
			},
		},
	})
}

func TestGroupFocus(t *testing.T) {
	for _, tc := range []struct {
		name string
		set  []int
		want int
	}{
		{name: "none", want: -1},
		{name: "set", set: []int{2}, want: 2},
		{name: "moved", set: []int{2, 0}, want: 0},
		{name: "taken away", set: []int{2, 5}, want: -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newGrouped(new(rec))
			for _, i := range tc.set {
				g.SetFocus(g.layout, i)
			}
			assert.Equal(t, tc.want, g.Focus(g.layout))
		})
	}
}

// -- fixtures ---------------------------------------------------------------

// rec records what the leaves of a fixture get, and finds them by name.
type rec struct {
	log    []string
	leaves map[string]*leaf
}

func (r *rec) add(l *leaf) *leaf {
	if r.leaves == nil {
		r.leaves = make(map[string]*leaf)
	}
	r.leaves[l.name] = l
	l.rec = r
	return l
}

// leaf returns the leaf named name, which must exist.
func (r *rec) leaf(name string) *leaf {
	l, ok := r.leaves[name]
	if !ok {
		panic("no leaf " + name)
	}
	return l
}

func (r *rec) note(s string) { r.log = append(r.log, s) }

// leaf is a handler that records what it gets. It handles mouse events
// but the wheel, and printable keys but n, j and k, exiting on q,
// unless it is inert, which handles nothing, or wheel, which handles the
// wheel too. It draws its name, is h rows tall and shows its cursor at
// cur unless hide is set.
type leaf struct {
	name          string
	rec           *rec
	inert, wheel  bool
	hide          bool
	h             int
	cur           term.Coordinates
	width, height int
}

func newLeaf(r *rec, name string) *leaf {
	return r.add(&leaf{name: name, h: 1, cur: term.Coordinates{X: 1}})
}

func inertLeaf(r *rec, name string) *leaf {
	l := newLeaf(r, name)
	l.inert = true
	return l
}

func wheelLeaf(r *rec, name string) *leaf {
	l := newLeaf(r, name)
	l.wheel = true
	return l
}

func tallLeaf(r *rec, name string, h int) *leaf {
	l := newLeaf(r, name)
	l.h = h
	return l
}

func (l *leaf) Resize(width, height int) { l.width, l.height = width, height }
func (l *leaf) Height(int) int           { return l.h }
func (l *leaf) Dimensions() (int, int)   { return len(l.name), l.h }

func (l *leaf) Draw(w term.Writer) {
	if l.height <= 0 {
		return
	}
	for x, ch := range l.name {
		if x >= l.width {
			break
		}
		w.SetCell(term.Coordinates{X: x}, term.NewCell(ch, 1, term.Attributes{}))
	}
}

func (l *leaf) Selection() (string, bool) { return l.name, true }

func (l *leaf) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	return l.cur, term.CursorStyleBlinkingBar, !l.hide
}

func (l *leaf) Handle(ev term.Event) (exit, handled bool) {
	if l.rec != nil {
		l.rec.note(l.name + " " + describe(ev))
	}
	if l.inert {
		return false, false
	}
	switch ev.Type {
	case term.EventKey:
		if ev.Ch == 'q' {
			return true, true
		}
		return false, ev.Ch != 0 && !strings.ContainsRune("njk", ev.Ch)
	case term.EventMouse:
		wheel := ev.Key == term.MouseWheelUp || ev.Key == term.MouseWheelDown
		return false, !wheel || l.wheel
	}
	return false, false
}

// focused returns the handler a Focus method returns, or nil.
func focused[H any](h H, ok bool) any {
	if !ok {
		return nil
	}
	return h
}

// byValue is a leaf handler that is not a pointer.
type byValue struct{ *leaf }

// uncomparable is a leaf handler == panics on: a func field makes it
// uncomparable.
type uncomparable struct {
	*leaf
	_ func()
}

// grouped routes events to fixed children with a Group.
type grouped struct {
	Group
	layout Layout
}

func (g *grouped) Resize(int, int)  {}
func (g *grouped) Draw(term.Writer) {}

func (g *grouped) Handle(ev term.Event) (bool, bool) { return g.Group.Handle(g.layout, ev) }

func (g *grouped) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	return g.Group.Cursor(g.layout)
}

func (g *grouped) Selection() (string, bool) { return g.Group.Selection(g.layout) }

// newGrouped lays out, 10 by 3: A, which captures drags, 4 wide; B, 4
// wide; and C, inert, 2 wide.
func newGrouped(r *rec) *grouped {
	return &grouped{layout: Layout{Width: 10, Height: 3, Children: []Child{
		{Handler: newLeaf(r, "A"), Width: 4, Height: 3, CaptureDrags: true},
		{Handler: newLeaf(r, "B"), Position: term.Coordinates{X: 4}, Width: 4, Height: 3},
		{Handler: inertLeaf(r, "C"), Position: term.Coordinates{X: 8}, Width: 2, Height: 3},
	}}}
}

// newByValueGrouped lays out V0 and V1, which are not pointers, 2 wide.
func newByValueGrouped(r *rec) *grouped {
	return &grouped{layout: Layout{Width: 4, Height: 1, Children: []Child{
		{Handler: byValue{newLeaf(r, "V0")}, Width: 2, Height: 1},
		{Handler: byValue{newLeaf(r, "V1")}, Position: term.Coordinates{X: 2}, Width: 2, Height: 1},
	}}}
}

// newOverlapGrouped lays out A, 4 by 3 and capturing, and O over it at
// 2,1, 2 by 1.
func newOverlapGrouped(r *rec) *grouped {
	return &grouped{layout: Layout{Width: 4, Height: 3, Children: []Child{
		{Handler: newLeaf(r, "A"), Width: 4, Height: 3, CaptureDrags: true},
		{Handler: newLeaf(r, "O"), Position: term.Coordinates{X: 2, Y: 1}, Width: 2, Height: 1},
	}}}
}

// handleCase is events sent to a fixture and what its leaves got.
type handleCase[H tui.Handler] struct {
	name string
	// fixture, if set, replaces the fixture of the table.
	fixture func(r *rec) H
	setup   func(h H, r *rec)
	events  []term.Event
	// before runs a function before the event at its index.
	before map[int]func(h H, r *rec)
	// want is what the leaves got, and what the callbacks of the fixture
	// noted, in order, with "-" for an event nothing got.
	want []string
	// handled and exit, if set, are what Handle returned for each event.
	handled, exit []bool
	then          func(t *testing.T, h H, r *rec)
}

func runHandleCases[H tui.Handler](t *testing.T, fixture func(r *rec) H, cases []handleCase[H]) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			build := fixture
			if tc.fixture != nil {
				build = tc.fixture
			}
			r := new(rec)
			h := build(r)
			if tc.setup != nil {
				tc.setup(h, r)
			}
			r.log = nil
			var got []string
			var handled, exit []bool
			for i, ev := range tc.events {
				if f := tc.before[i]; f != nil {
					f(h, r)
				}
				n := len(r.log)
				x, ok := h.Handle(ev)
				handled, exit = append(handled, ok), append(exit, x)
				if len(r.log) == n {
					got = append(got, "-")
				}
				got = append(got, r.log[n:]...)
			}
			if tc.want != nil {
				assert.Equal(t, tc.want, got, "what the leaves got")
			}
			if tc.handled != nil {
				assert.Equal(t, tc.handled, handled, "handled")
			}
			if tc.exit != nil {
				assert.Equal(t, tc.exit, exit, "exit")
			}
			if tc.then != nil {
				tc.then(t, h, r)
			}
		})
	}
}

// viewCase checks the methods of a fixture that are not Handle.
type viewCase[H tui.Handler] struct {
	name    string
	fixture func(r *rec) H
	// width and height, if set, resize the fixture before setup.
	width, height int
	setup         func(h H, r *rec)
	cursor        term.Coordinates
	show          bool
	// selection is what Selection returns, "" for nothing.
	selection string
	// draw, if set, is what the fixture draws width by height.
	draw string
	// sizes, if set, are the widths and heights of leaves by name.
	sizes map[string][2]int
}

func runViewCases[H tui.Handler](t *testing.T, fixture func(r *rec) H, cases []viewCase[H]) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			build := fixture
			if tc.fixture != nil {
				build = tc.fixture
			}
			r := new(rec)
			h := build(r)
			if tc.width > 0 {
				h.Resize(tc.width, tc.height)
			}
			if tc.setup != nil {
				tc.setup(h, r)
			}
			pos, _, show := h.Cursor()
			assert.Equal(t, tc.show, show, "cursor shows")
			if tc.show {
				assert.Equal(t, tc.cursor, pos, "cursor")
			}
			sel, ok := h.Selection()
			assert.Equal(t, tc.selection != "", ok, "selection")
			assert.Equal(t, tc.selection, sel)
			for name, size := range tc.sizes {
				l := r.leaf(name)
				assert.Equal(t, size, [2]int{l.width, l.height}, "size of %s", name)
			}
			if tc.draw != "" {
				require.Positive(t, tc.width, "draw needs a size")
				assert.Equal(t, strings.TrimPrefix(tc.draw, "\n"), draw(h, tc.width, tc.height))
			}
		})
	}
}

// draw returns what c draws width by height.
func draw(c tui.Component, width, height int) string {
	w := term.NewStringWriter(width, height)
	c.Draw(w)
	_ = w.Flush()
	return w.String()
}

var mouseNames = map[term.Key]string{
	term.MouseLeft: "left", term.MouseRelease: "release", 0: "move",
	term.MouseWheelUp: "wheelup", term.MouseWheelDown: "wheeldown",
}

var keyNames = map[term.Key]string{
	term.KeyTab: "tab", term.KeyEnter: "enter", term.KeyArrowUp: "up", term.KeyArrowDown: "down",
	term.KeyHome: "home", term.KeyEnd: "end", term.KeyPgup: "pgup", term.KeyPgdn: "pgdn",
}

// describe names ev as leaves log it.
func describe(ev term.Event) string {
	switch ev.Type {
	case term.EventMouse:
		return fmt.Sprintf("%s %d,%d", mouseNames[ev.Key], ev.MouseX, ev.MouseY)
	case term.EventKey:
		if ev.Ch != 0 {
			return "key " + string(ev.Ch)
		}
		return "key " + keyNames[ev.Key]
	case term.EventPasteStart:
		return "paste"
	}
	return fmt.Sprintf("event %d", ev.Type)
}

func key(ch rune) term.Event        { return term.Event{Type: term.EventKey, Ch: ch} }
func special(k term.Key) term.Event { return term.Event{Type: term.EventKey, Key: k} }
func paste() term.Event             { return term.Event{Type: term.EventPasteStart} }
func mouse(k term.Key, x, y int) term.Event {
	return term.Event{Type: term.EventMouse, Key: k, MouseX: x, MouseY: y}
}
func press(x, y int) term.Event     { return mouse(term.MouseLeft, x, y) }
func release(x, y int) term.Event   { return mouse(term.MouseRelease, x, y) }
func move(x, y int) term.Event      { return mouse(0, x, y) }
func wheelUp(x, y int) term.Event   { return mouse(term.MouseWheelUp, x, y) }
func wheelDown(x, y int) term.Event { return mouse(term.MouseWheelDown, x, y) }
