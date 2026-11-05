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
	"testing"

	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestFocusListHandle(t *testing.T) {
	runHandleCases(t, newTestFocusList, []handleCase[*FocusList]{
		{
			name:    "the element in focus gets keys first",
			events:  []term.Event{key('x'), key('q')},
			want:    []string{"a key x", "a key q"},
			handled: []bool{true, true},
			exit:    []bool{false, true},
		},
		{
			name:    "j and k move the focus the element leaves them",
			events:  []term.Event{key('j'), key('j'), key('j'), key('k')},
			want:    []string{"a key j", "focus b", "b key j", "focus c", "c key j", "c key k", "focus b"},
			handled: []bool{true, true, false, true},
		},
		{
			name:    "k at the top is left to the parent",
			events:  []term.Event{key('k')},
			want:    []string{"a key k"},
			handled: []bool{false},
		},
		{
			name:    "the arrows move the focus",
			events:  []term.Event{special(term.KeyArrowDown), special(term.KeyArrowUp)},
			want:    []string{"a key down", "focus b", "b key up", "focus a"},
			handled: []bool{true, true},
		},
		{
			name:    "Home and End move the focus to the ends, and notify only when it moves",
			events:  []term.Event{special(term.KeyEnd), special(term.KeyEnd), special(term.KeyHome)},
			want:    []string{"a key end", "focus c", "c key end", "c key home", "focus a"},
			handled: []bool{true, true, true},
		},
		{
			name:    "PgDn and PgUp move the focus a page",
			fixture: func(r *rec) *FocusList { return focusListOf(r, 5, 2, "a", "b", "c", "d", "e") },
			events:  []term.Event{special(term.KeyPgdn), special(term.KeyPgdn), special(term.KeyPgup)},
			want:    []string{"a key pgdn", "focus c", "c key pgdn", "focus e", "e key pgup", "focus c"},
			handled: []bool{true, true, true},
		},
		{
			name:    "PgUp at the top is left to the parent",
			events:  []term.Event{special(term.KeyPgup)},
			want:    []string{"a key pgup"},
			handled: []bool{false},
		},
		{
			name:    "Enter selects the element in focus",
			events:  []term.Event{special(term.KeyArrowDown), special(term.KeyEnter)},
			want:    []string{"a key down", "focus b", "b key enter", "select b"},
			handled: []bool{true, true},
		},
		{
			name: "with no callbacks, moves are handled and Enter is not",
			fixture: func(r *rec) *FocusList {
				l := focusListOf(r, 5, 3, "a", "b")
				l.OnFocus, l.OnSelect = nil, nil
				return l
			},
			events:  []term.Event{special(term.KeyArrowDown), special(term.KeyEnter)},
			want:    []string{"a key down", "b key enter"},
			handled: []bool{true, false},
		},
		{
			name:    "other keys are left to the parent",
			events:  []term.Event{special(term.KeyTab), key('n')},
			want:    []string{"a key tab", "a key n"},
			handled: []bool{false, false},
		},
		{
			name:    "events that are neither keys nor mouse events go to the element in focus only",
			events:  []term.Event{paste()},
			want:    []string{"a paste"},
			handled: []bool{false},
		},
		{
			name:    "an empty list leaves everything to the parent",
			fixture: func(r *rec) *FocusList { return focusListOf(r, 5, 3) },
			events:  []term.Event{key('x'), special(term.KeyArrowDown), special(term.KeyEnter), press(1, 0)},
			want:    []string{"-", "-", "-", "-"},
			handled: []bool{false, false, false, false},
		},
		{
			name: "elements that are not handlers get nothing, but the list moves and selects",
			fixture: func(r *rec) *FocusList {
				l := focusListOf(r, 5, 3)
				l.PushBack(component.NewString("s"))
				l.PushBack(newLeaf(r, "b"))
				return l
			},
			events:  []term.Event{key('x'), key('j'), key('x')},
			want:    []string{"-", "focus b", "b key x"},
			handled: []bool{false, true, true},
		},
		{
			name:    "a press moves the focus to the element, which gets it",
			events:  []term.Event{press(1, 2), release(1, 2), key('x')},
			want:    []string{"focus c", "c left 1,0", "c release 1,0", "c key x"},
			handled: []bool{true, true, true},
		},
		{
			name:    "a press on the element in focus does not move it",
			events:  []term.Event{press(1, 0)},
			want:    []string{"a left 1,0"},
			handled: []bool{true},
		},
		{
			name:    "a click no element takes selects",
			fixture: newInertFocusList,
			events:  []term.Event{press(1, 1), release(1, 1)},
			want:    []string{"focus b", "b left 1,0", "b release 1,0", "select b"},
			handled: []bool{true, true},
		},
		{
			name:    "a release over another element than the press is no click",
			fixture: newInertFocusList,
			events:  []term.Event{press(1, 1), release(1, 2)},
			want:    []string{"focus b", "b left 1,0", "b release 1,1"},
			handled: []bool{true, false},
		},
		{
			name:    "a release with no press is no click",
			fixture: newInertFocusList,
			events:  []term.Event{release(1, 1)},
			want:    []string{"b release 1,0"},
			handled: []bool{false},
		},
		{
			name:    "a motion forgets the press",
			fixture: newInertFocusList,
			events:  []term.Event{press(1, 1), move(1, 1), release(1, 1)},
			want:    []string{"focus b", "b left 1,0", "b move 1,0", "b release 1,0"},
			handled: []bool{true, false, false},
		},
		{
			name:    "a press outside the list or past its elements",
			fixture: func(r *rec) *FocusList { return focusListOf(r, 5, 4, "a", "b") },
			events:  []term.Event{press(7, 0), release(7, 0), press(1, 3), release(1, 3)},
			want:    []string{"-", "-", "-", "-"},
			handled: []bool{false, false, false, false},
		},
		{
			name:    "the wheel moves the focus when the element leaves it",
			events:  []term.Event{wheelDown(1, 0), wheelUp(1, 1), wheelUp(1, 0)},
			want:    []string{"a wheeldown 1,0", "focus b", "a move 1,1", "b wheelup 1,0", "focus a", "b move 1,-1", "a wheelup 1,0"},
			handled: []bool{true, true, false},
		},
		{
			name:    "the wheel an element takes",
			fixture: func(r *rec) *FocusList { l := focusListOf(r, 5, 3); l.PushBack(wheelLeaf(r, "w")); return l },
			events:  []term.Event{wheelDown(1, 0)},
			want:    []string{"w wheeldown 1,0"},
			handled: []bool{true},
		},
	})
}

func TestFocusListViews(t *testing.T) {
	runViewCases(t, newTestFocusList, []viewCase[*FocusList]{
		{
			name:      "the element in focus shows its cursor and selection",
			cursor:    term.Coordinates{X: 1},
			show:      true,
			selection: "a",
		},
		{
			name: "where it is in the list",
			setup: func(l *FocusList, _ *rec) {
				l.Handle(special(term.KeyEnd))
			},
			cursor:    term.Coordinates{X: 1, Y: 2},
			show:      true,
			selection: "c",
		},
		{
			name: "a cursor past the list is hidden",
			setup: func(l *FocusList, r *rec) {
				r.leaf("a").cur.X = 5
			},
			selection: "a",
		},
		{
			name:    "an empty list shows nothing",
			fixture: func(r *rec) *FocusList { return focusListOf(r, 5, 3) },
		},
		{
			name: "an element that is not a handler shows nothing",
			fixture: func(r *rec) *FocusList {
				l := focusListOf(r, 5, 3)
				l.PushBack(component.NewString("s"))
				return l
			},
			width: 5, height: 3,
			draw: "s    \n     \n     ",
		},
		{
			name:  "the elements are drawn",
			width: 5, height: 3,
			draw:   "a    \nb    \nc    ",
			cursor: term.Coordinates{X: 1}, show: true, selection: "a",
		},
	})
}

// newTestFocusList is a, b and c, 5 by 3.
func newTestFocusList(r *rec) *FocusList { return focusListOf(r, 5, 3, "a", "b", "c") }

// newInertFocusList is a, b and c, inert, 5 by 3.
func newInertFocusList(r *rec) *FocusList {
	l := newTestFocusList(r)
	for _, name := range []string{"a", "b", "c"} {
		r.leaf(name).inert = true
	}
	return l
}
