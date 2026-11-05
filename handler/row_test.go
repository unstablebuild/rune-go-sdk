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
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestRowHandle(t *testing.T) {
	focus := func(name string) func(*Row, *rec) {
		return func(row *Row, r *rec) { row.SetFocus(r.leaf(name)) }
	}
	runHandleCases(t, newTestRow, []handleCase[*Row]{
		{
			name:    "keys with no handler in focus are left unhandled",
			events:  []term.Event{key('x')},
			want:    []string{"-"},
			handled: []bool{false},
		},
		{
			name:    "keys go to the handler SetFocus put in focus",
			setup:   focus("B"),
			events:  []term.Event{key('x'), key('n'), special(term.KeyTab)},
			want:    []string{"B key x", "B key n", "B key tab"},
			handled: []bool{true, false, false},
		},
		{
			name:  "SetFocus of a handler the row does not hold leaves the focus",
			setup: focus("B"),
			then: func(t *testing.T, row *Row, r *rec) {
				assert.False(t, row.SetFocus(newLeaf(new(rec), "X")))
				assert.Equal(t, r.leaf("B"), focused(row.Focus()))
			},
		},
		{
			name:    "a press goes to the handler under it, relative to it",
			events:  []term.Event{press(5, 0), release(5, 0)},
			want:    []string{"B left 1,0", "B release 1,0"},
			handled: []bool{true, true},
		},
		{
			name:    "every handler captures the drags of its presses",
			events:  []term.Event{press(5, 0), press(1, 0), release(1, 0)},
			want:    []string{"B left 1,0", "B left -3,0", "B release -3,0"},
			handled: []bool{true, true, true},
		},
		{
			name:    "what the handler leaves is unhandled",
			events:  []term.Event{press(9, 0), wheelDown(1, 0)},
			want:    []string{"C left 1,0", "A wheeldown 1,0"},
			handled: []bool{false, false},
		},
		{
			name:   "the handler the pointer leaves gets a motion outside it",
			events: []term.Event{move(1, 0), move(5, 0)},
			want:   []string{"A move 1,0", "A move 5,0", "B move 1,0"},
		},
		{
			name:    "components are only drawn",
			fixture: newMixedRow,
			events:  []term.Event{press(1, 0), release(1, 0), press(7, 0)},
			want:    []string{"-", "-", "B left 1,0"},
		},
		{
			name:    "a zero Row holds nothing",
			fixture: func(*rec) *Row { return new(Row) },
			events:  []term.Event{press(0, 0), key('x')},
			want:    []string{"-", "-"},
			handled: []bool{false, false},
			then: func(t *testing.T, row *Row, _ *rec) {
				_, ok := row.Focus()
				assert.False(t, ok)
			},
		},
		{
			name:    "a handler == panics on is never found",
			fixture: newUncomparableRow,
			then: func(t *testing.T, row *Row, r *rec) {
				assert.False(t, row.SetFocus(uncomparable{leaf: r.leaf("U")}))
				assert.True(t, row.SetFocus(r.leaf("A")))
			},
		},
	})
}

func TestRowViews(t *testing.T) {
	runViewCases(t, newTestRow, []viewCase[*Row]{
		{
			name:  "handlers are laid out by their columns",
			width: 12, height: 1,
			draw:  "A   B   C   ",
			sizes: map[string][2]int{"A": {4, 1}, "B": {4, 1}, "C": {4, 1}},
		},
		{name: "no handler in focus shows no cursor"},
		{
			name:      "the cursor of the handler in focus, offset by its position",
			setup:     func(row *Row, r *rec) { row.SetFocus(r.leaf("C")) },
			cursor:    term.Coordinates{X: 9},
			show:      true,
			selection: "C",
		},
		{
			name: "a cursor past the row is hidden",
			setup: func(row *Row, r *rec) {
				r.leaf("C").cur.X = 4
				row.SetFocus(r.leaf("C"))
			},
			selection: "C",
		},
		{
			name:    "components are drawn with the handlers",
			fixture: newMixedRow,
			width:   12, height: 1,
			draw: "text  B     ",
		},
	})
}

// newTestRow is A, B and C, inert, 4 columns each, 12 by 1.
func newTestRow(r *rec) *Row {
	row := NewRow()
	row.AddHandler(newLeaf(r, "A"), 4)
	row.AddHandler(newLeaf(r, "B"), 4)
	row.AddHandler(inertLeaf(r, "C"), 4)
	row.Resize(12, 1)
	return row
}

// newMixedRow is a component, 6 columns, and B, 6 columns, 12 by 1.
func newMixedRow(r *rec) *Row {
	row := NewRow()
	row.AddComponent(NopResponsiveFromComponent(component.NewString("text")), 6)
	row.AddHandler(newLeaf(r, "B"), 6)
	row.Resize(12, 1)
	return row
}

func newUncomparableRow(r *rec) *Row {
	row := NewRow()
	row.AddHandler(uncomparable{leaf: newLeaf(r, "U")}, 6)
	row.AddHandler(newLeaf(r, "A"), 6)
	row.Resize(12, 1)
	return row
}
