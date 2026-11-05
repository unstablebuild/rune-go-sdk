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
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

func TestGridHandle(t *testing.T) {
	runHandleCases(t, newTestGrid, []handleCase[*Grid]{
		{
			name:    "a press goes to the cell under it, relative to it",
			events:  []term.Event{press(3, 0), release(3, 0), press(1, 1)},
			want:    []string{"b left 1,0", "b release 1,0", "b move -1,1", "c left 1,0"},
			handled: []bool{true, true, true},
		},
		{
			name:    "an empty cell takes nothing",
			events:  []term.Event{press(3, 1), release(3, 1)},
			want:    []string{"-", "-"},
			handled: []bool{false, false},
		},
		{
			name:   "every cell captures the drags of its presses",
			events: []term.Event{press(1, 0), press(3, 1), release(3, 1)},
			want:   []string{"a left 1,0", "a left 3,1", "a release 3,1"},
		},
		{
			name:   "the cell the pointer leaves gets a motion outside it",
			events: []term.Event{move(1, 0), move(1, 1)},
			want:   []string{"a move 1,0", "a move 1,1", "c move 1,0"},
		},
		{
			name:    "keys with no cell in focus are left unhandled",
			events:  []term.Event{key('x')},
			want:    []string{"-"},
			handled: []bool{false},
		},
		{
			name:    "keys go to the cell in focus",
			setup:   func(g *Grid, r *rec) { g.SetFocus(r.leaf("c")) },
			events:  []term.Event{key('x')},
			want:    []string{"c key x"},
			handled: []bool{true},
		},
		{
			name: "nothing takes the focus before Resize lays the grid out",
			fixture: func(r *rec) *Grid {
				return NewGrid([][]tui.Handler{{newLeaf(r, "a")}})
			},
			events: []term.Event{press(0, 0)},
			want:   []string{"-"},
			then: func(t *testing.T, g *Grid, r *rec) {
				assert.False(t, g.SetFocus(r.leaf("a")))
				g.Resize(1, 1)
				assert.True(t, g.SetFocus(r.leaf("a")))
				assert.Equal(t, r.leaf("a"), focused(g.Focus()))
			},
		},
		{
			name:    "an empty grid takes nothing",
			fixture: func(*rec) *Grid { g := NewGrid(nil); g.Resize(4, 2); return g },
			events:  []term.Event{press(0, 0), key('x')},
			want:    []string{"-", "-"},
		},
		{
			name: "an empty row leaves its cells empty",
			fixture: func(r *rec) *Grid {
				g := NewGrid([][]tui.Handler{{}, {newLeaf(r, "a")}})
				g.Resize(2, 2)
				return g
			},
			events: []term.Event{press(0, 0), release(0, 0), press(0, 1)},
			want:   []string{"-", "-", "a left 0,0"},
		},
	})
}

func TestGridViews(t *testing.T) {
	runViewCases(t, newTestGrid, []viewCase[*Grid]{
		{
			name:  "cells are laid out in equal sizes",
			width: 4, height: 2,
			draw:  "a b \nc   ",
			sizes: map[string][2]int{"a": {2, 1}, "b": {2, 1}, "c": {2, 1}},
		},
		{
			name:  "cells are cut where the room does not divide",
			width: 5, height: 3,
			draw:  "a b  \nc    \n     ",
			sizes: map[string][2]int{"a": {2, 1}, "c": {2, 1}},
		},
		{name: "no cell in focus shows no cursor"},
		{
			name:      "the cursor of the cell in focus, offset by its position",
			setup:     func(g *Grid, r *rec) { g.SetFocus(r.leaf("b")) },
			cursor:    term.Coordinates{X: 3},
			show:      true,
			selection: "b",
		},
	})
}

// newTestGrid is a and b over c and an empty cell, 4 by 2.
func newTestGrid(r *rec) *Grid {
	g := NewGrid([][]tui.Handler{{newLeaf(r, "a"), newLeaf(r, "b")}, {newLeaf(r, "c"), nil}})
	g.Resize(4, 2)
	return g
}
