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

func TestInlineHandle(t *testing.T) {
	runHandleCases(t, newTestInline, []handleCase[*Inline]{
		{
			name:    "a press goes to the handler under it, relative to it",
			events:  []term.Event{press(2, 0), release(2, 0), press(0, 0)},
			want:    []string{"yz left 1,0", "yz release 1,0", "yz move -1,0", "x left 0,0"},
			handled: []bool{true, true, true},
		},
		{
			name:    "past the handlers is nothing",
			events:  []term.Event{press(3, 0)},
			want:    []string{"-"},
			handled: []bool{false},
		},
		{
			name:   "every handler captures the drags of its presses",
			events: []term.Event{press(0, 0), press(2, 0), release(2, 0)},
			want:   []string{"x left 0,0", "x left 2,0", "x release 2,0"},
		},
		{
			name:    "keys go to the handler in focus",
			setup:   func(in *Inline, r *rec) { in.SetFocus(r.leaf("yz")) },
			events:  []term.Event{key('x'), key('n')},
			want:    []string{"yz key x", "yz key n"},
			handled: []bool{true, false},
		},
		{
			name:    "keys with no handler in focus are left unhandled",
			events:  []term.Event{key('x')},
			want:    []string{"-"},
			handled: []bool{false},
		},
		{
			name:    "handlers that overflow are cut",
			setup:   func(in *Inline, _ *rec) { in.Resize(2, 1) },
			events:  []term.Event{press(1, 0), release(1, 0), press(2, 0)},
			want:    []string{"yz left 0,0", "yz release 0,0", "yz move 1,0"},
			handled: []bool{true, true, false},
		},
	})
}

func TestInlineViews(t *testing.T) {
	runViewCases(t, newTestInline, []viewCase[*Inline]{
		{
			name:  "handlers are laid out at their width",
			width: 4, height: 1,
			draw:  "xyz ",
			sizes: map[string][2]int{"x": {1, 1}, "yz": {2, 1}},
		},
		{
			name:  "handlers that overflow are cut",
			width: 2, height: 1,
			draw:  "xy",
			sizes: map[string][2]int{"x": {1, 1}, "yz": {1, 1}},
		},
		{
			name:      "the cursor of the handler in focus, offset by its position",
			setup:     func(in *Inline, r *rec) { in.SetFocus(r.leaf("yz")) },
			cursor:    term.Coordinates{X: 2},
			show:      true,
			selection: "yz",
		},
	})
	in := newTestInline(new(rec))
	w, h := in.Dimensions()
	assert.Equal(t, [2]int{3, 1}, [2]int{w, h}, "the dimensions of the handlers, side by side")
}

// newTestInline is x and yz at their width, 3 by 1.
func newTestInline(r *rec) *Inline {
	in := NewInline([]Floating{newLeaf(r, "x"), newLeaf(r, "yz")}, component.AlignmentLeft)
	in.Resize(3, 1)
	return in
}
