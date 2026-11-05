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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestListHandle(t *testing.T) {
	focus := func(name string) func(*List, *rec) {
		return func(l *List, r *rec) { l.SetFocus(r.leaf(name)) }
	}
	offset := func(want int) func(*testing.T, *List, *rec) {
		return func(t *testing.T, l *List, _ *rec) { assert.Equal(t, want, l.Offset()) }
	}
	runHandleCases(t, newTestList, []handleCase[*List]{
		{
			name:    "a wheel event no element takes seeks",
			events:  []term.Event{wheelDown(0, 0)},
			want:    []string{"a wheeldown 0,0"},
			handled: []bool{true},
			then:    offset(1),
		},
		{
			name:    "a wheel event past the top is left to the parent",
			events:  []term.Event{wheelUp(0, 0)},
			want:    []string{"a wheelup 0,0"},
			handled: []bool{false},
			then:    offset(0),
		},
		{
			name:    "a wheel event that seeks back up",
			events:  []term.Event{wheelDown(0, 0), wheelUp(0, 0)},
			handled: []bool{true, true},
			then:    offset(0),
		},
		{
			name:    "a wheel event an element takes does not seek",
			fixture: newWheelList,
			events:  []term.Event{wheelDown(0, 0)},
			want:    []string{"a wheeldown 0,0"},
			handled: []bool{true},
			then:    offset(0),
		},
		{
			name:    "events reach the elements where they show after a seek",
			events:  []term.Event{wheelDown(0, 0), press(0, 1)},
			want:    []string{"a wheeldown 0,0", "a move 0,2", "c left 0,0"},
			handled: []bool{true, true},
		},
		{
			name:    "components are only drawn, and a list that shows everything does not seek",
			setup:   func(l *List, _ *rec) { l.Resize(5, 6) },
			events:  []term.Event{press(0, 5), wheelDown(0, 5)},
			want:    []string{"-", "-"},
			handled: []bool{false, false},
		},
		{
			name:    "keys with no element in focus are left unhandled",
			events:  []term.Event{key('x')},
			want:    []string{"-"},
			handled: []bool{false},
		},
		{
			name:    "a key the element in focus takes seeks to show it",
			setup:   focus("e"),
			events:  []term.Event{wheelUp(0, 0), wheelUp(0, 0), key('x')},
			handled: []bool{true, true, true},
			then:    offset(3),
		},
		{
			name:    "a key the element in focus leaves does not seek",
			setup:   focus("e"),
			events:  []term.Event{wheelUp(0, 0), key('n')},
			handled: []bool{true, false},
			then:    offset(2),
		},
		{
			name:    "a key the element in focus exits on",
			setup:   focus("b"),
			events:  []term.Event{key('q')},
			want:    []string{"b key q"},
			handled: []bool{true},
			exit:    []bool{true},
		},
		{
			name:  "SetFocus seeks to show the element",
			setup: focus("e"),
			then: func(t *testing.T, l *List, r *rec) {
				assert.Equal(t, 3, l.Offset())
				require.True(t, l.SetFocus(r.leaf("b")))
				assert.Equal(t, 1, l.Offset())
				require.True(t, l.SetFocus(r.leaf("c")))
				assert.Equal(t, 1, l.Offset(), "an element that shows needs no seek")
				assert.False(t, l.SetFocus(newLeaf(new(rec), "x")))
				assert.Equal(t, r.leaf("c"), focused(l.Focus()))
			},
		},
	})
}

func TestListViews(t *testing.T) {
	runViewCases(t, newTestList, []viewCase[*List]{
		{
			name:  "the elements that show are drawn",
			width: 5, height: 2,
			draw: "a    \nb    ",
		},
		{name: "no element in focus shows no cursor"},
		{
			name:      "the cursor of the element in focus, where it shows",
			setup:     func(l *List, r *rec) { l.SetFocus(r.leaf("e")) },
			cursor:    term.Coordinates{X: 1, Y: 1},
			show:      true,
			selection: "e",
		},
	})
}

// newTestList is a to e and a component, a row each, 5 by 2.
func newTestList(r *rec) *List {
	l := NewList(1)
	for _, name := range []string{"a", "b", "c", "d", "e"} {
		l.PushBack(newLeaf(r, name))
	}
	l.PushBack(component.NewString("text"))
	l.Resize(5, 2)
	return l
}

func newWheelList(r *rec) *List {
	l := NewList(1)
	l.PushBack(wheelLeaf(r, "a"))
	l.PushBack(newLeaf(r, "b"))
	l.PushBack(newLeaf(r, "c"))
	l.Resize(5, 2)
	return l
}

func responsiveListOf(r *rec, height int, elems ...component.Responsive) *ResponsiveList {
	l := NewResponsiveList()
	for _, e := range elems {
		l.PushBack(e)
	}
	l.Resize(5, height)
	return l
}

// focusListOf is a FocusList of leaves named names, width by height,
// that notes its OnFocus and OnSelect calls.
func focusListOf(r *rec, width, height int, names ...string) *FocusList {
	l := NewFocusList()
	for _, name := range names {
		l.PushBack(newLeaf(r, name))
	}
	name := func(n component.ListNode) string {
		if h, ok := n.Value().(*leaf); ok {
			return h.name
		}
		return "?"
	}
	l.OnFocus = func(n component.ListNode) { r.note("focus " + name(n)) }
	l.OnSelect = func(n component.ListNode) { r.note("select " + name(n)) }
	l.Resize(width, height)
	return l
}
