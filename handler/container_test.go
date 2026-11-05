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

func TestContainerHandle(t *testing.T) {
	focus := func(name string) func(*Container, *rec) {
		return func(c *Container, r *rec) { c.SetFocus(r.leaf(name)) }
	}
	offset := func(want int) func(*testing.T, *Container, *rec) {
		return func(t *testing.T, c *Container, _ *rec) { assert.Equal(t, want, c.Offset()) }
	}
	runHandleCases(t, newTestContainer, []handleCase[*Container]{
		{
			name:    "a wheel event no row takes scrolls",
			events:  []term.Event{wheelDown(0, 0), wheelDown(0, 0)},
			want:    []string{"a wheeldown 0,0", "a move 0,1", "b wheeldown 0,0"},
			handled: []bool{true, true},
			then:    offset(2),
		},
		{
			name:    "a wheel event past the top is left to the parent",
			events:  []term.Event{wheelUp(0, 0)},
			want:    []string{"a wheelup 0,0"},
			handled: []bool{false},
			then:    offset(0),
		},
		{
			name:    "a wheel event past the bottom is left to the parent",
			setup:   focus("e"),
			events:  []term.Event{wheelDown(0, 1)},
			want:    []string{"e wheeldown 0,0"},
			handled: []bool{false},
			then:    offset(3),
		},
		{
			name:    "a wheel event a row takes does not scroll",
			fixture: newWheelContainer,
			events:  []term.Event{wheelDown(0, 0)},
			want:    []string{"a wheeldown 0,0"},
			handled: []bool{true},
			then:    offset(0),
		},
		{
			name:   "events reach the rows where they show after a scroll",
			events: []term.Event{wheelDown(0, 0), press(0, 0)},
			want:   []string{"a wheeldown 0,0", "a move 0,1", "b left 0,0"},
		},
		{
			name:    "keys with no row in focus are left unhandled",
			events:  []term.Event{key('x')},
			want:    []string{"-"},
			handled: []bool{false},
		},
		{
			name:    "keys go to the handler in focus, and a key it takes scrolls it into view",
			setup:   focus("e"),
			events:  []term.Event{wheelUp(0, 0), wheelUp(0, 0), key('x')},
			handled: []bool{true, true, true},
			then:    offset(3),
		},
		{
			name:    "a key the handler in focus leaves does not scroll",
			setup:   focus("e"),
			events:  []term.Event{wheelUp(0, 0), key('n')},
			handled: []bool{true, false},
			then:    offset(2),
		},
		{
			name:    "a key the handler in focus exits on",
			setup:   focus("a"),
			events:  []term.Event{key('q')},
			want:    []string{"a key q"},
			handled: []bool{true},
			exit:    []bool{true},
		},
		{
			name:  "SetFocus scrolls the row in focus into view",
			setup: focus("e"),
			then: func(t *testing.T, c *Container, r *rec) {
				assert.Equal(t, 3, c.Offset())
				assert.Equal(t, r.leaf("e"), focused(c.Focus()))
				assert.True(t, c.SetFocus(r.leaf("a")))
				assert.Equal(t, 0, c.Offset())
				assert.True(t, c.SetFocus(r.leaf("b")))
				assert.Equal(t, 0, c.Offset(), "a row that shows needs no scroll")
				assert.False(t, c.SetFocus(newLeaf(new(rec), "x")))
				assert.Equal(t, r.leaf("b"), focused(c.Focus()))
			},
		},
		{
			name:    "a row taller than the container scrolls to its top",
			fixture: newTallContainer,
			setup:   focus("tall"),
			then:    offset(1),
		},
		{
			name: "no row in focus",
			then: func(t *testing.T, c *Container, _ *rec) {
				_, ok := c.Focus()
				assert.False(t, ok)
			},
		},
	})
}

func TestContainerViews(t *testing.T) {
	runViewCases(t, newTestContainer, []viewCase[*Container]{
		{
			name:  "the rows that show are drawn",
			width: 3, height: 2,
			draw:  "a  \nb  ",
			sizes: map[string][2]int{"a": {3, 1}, "b": {3, 1}},
		},
		{name: "no row in focus shows no cursor"},
		{
			name:      "the cursor of the handler in focus, where its row shows",
			setup:     func(c *Container, r *rec) { c.SetFocus(r.leaf("e")) },
			cursor:    term.Coordinates{X: 1, Y: 1},
			show:      true,
			selection: "e",
		},
		{
			name:  "the rows that show after a scroll are drawn",
			width: 3, height: 2,
			setup: func(c *Container, r *rec) { c.SetFocus(r.leaf("e")) },
			draw:  "d  \ne  ", cursor: term.Coordinates{X: 1, Y: 1}, show: true, selection: "e",
		},
	})
}

// newTestContainer is a row for each of a to e, 12 by 2.
func newTestContainer(r *rec) *Container {
	c := NewContainer()
	for _, name := range []string{"a", "b", "c", "d", "e"} {
		c.AddRow().AddHandler(newLeaf(r, name), component.MaxCols)
	}
	c.Resize(12, 2)
	return c
}

func newWheelContainer(r *rec) *Container {
	c := NewContainer()
	c.AddRow().AddHandler(wheelLeaf(r, "a"), component.MaxCols)
	c.AddRow().AddHandler(newLeaf(r, "b"), component.MaxCols)
	c.AddRow().AddHandler(newLeaf(r, "c"), component.MaxCols)
	c.Resize(12, 2)
	return c
}

// newTallContainer is a, then tall, 3 rows tall, 12 by 2.
func newTallContainer(r *rec) *Container {
	c := NewContainer()
	c.AddRow().AddHandler(newLeaf(r, "a"), component.MaxCols)
	c.AddRow().AddHandler(tallLeaf(r, "tall", 3), component.MaxCols)
	c.Resize(12, 2)
	return c
}
