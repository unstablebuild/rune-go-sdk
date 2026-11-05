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

func TestResponsiveListHandle(t *testing.T) {
	runHandleCases(t, newTestResponsiveList, []handleCase[*ResponsiveList]{
		{
			name:   "events go to the elements at their height",
			events: []term.Event{press(2, 0), release(2, 0), press(2, 2), release(2, 2)},
			want:   []string{"a left 2,0", "a release 2,0", "a move 2,2", "b left 2,1", "b release 2,1"},
		},
		{
			name:    "a wheel event no element takes seeks",
			events:  []term.Event{wheelDown(0, 0), press(3, 0)},
			want:    []string{"a wheeldown 0,0", "a move 3,1", "b left 3,0"},
			handled: []bool{true, true},
		},
		{
			name:    "a wheel event that seeks back up",
			events:  []term.Event{wheelDown(0, 0), wheelUp(0, 0)},
			handled: []bool{true, true},
		},
		{
			name:    "a wheel event past the top is left to the parent",
			events:  []term.Event{wheelUp(0, 0)},
			handled: []bool{false},
		},
		{
			name:    "a wheel event an element takes does not seek",
			fixture: func(r *rec) *ResponsiveList { return responsiveListOf(r, 2, wheelLeaf(r, "a"), newLeaf(r, "b")) },
			events:  []term.Event{wheelDown(0, 0)},
			handled: []bool{true},
			then: func(t *testing.T, l *ResponsiveList, _ *rec) {
				assert.Equal(t, 0, l.Offset())
			},
		},
		{
			name:    "keys go to the element in focus",
			setup:   func(l *ResponsiveList, r *rec) { l.SetFocus(r.leaf("b")) },
			events:  []term.Event{key('x')},
			want:    []string{"b key x"},
			handled: []bool{true},
		},
		{
			name: "a bottom aligned list lays its end at the bottom",
			fixture: func(r *rec) *ResponsiveList {
				l := NewResponsiveList()
				l.Alignment = component.AlignmentBottom
				for _, name := range []string{"a", "b", "c"} {
					l.PushBack(newLeaf(r, name))
				}
				l.Resize(5, 2)
				return l
			},
			events: []term.Event{press(0, 0), release(0, 0), press(0, 1)},
			want:   []string{"b left 0,0", "b release 0,0", "b move 0,1", "c left 0,0"},
		},
	})
}

func TestResponsiveListViews(t *testing.T) {
	runViewCases(t, newTestResponsiveList, []viewCase[*ResponsiveList]{
		{
			name:  "the elements that show are drawn",
			width: 5, height: 3,
			draw: "a    \nb    \n     ",
		},
		{
			name:      "the cursor of the element in focus",
			setup:     func(l *ResponsiveList, r *rec) { l.SetFocus(r.leaf("b")) },
			cursor:    term.Coordinates{X: 1, Y: 1},
			show:      true,
			selection: "b",
		},
	})
}

// newTestResponsiveList is a, b, 2 rows tall, and c, 5 by 3.
func newTestResponsiveList(r *rec) *ResponsiveList {
	return responsiveListOf(r, 3, newLeaf(r, "a"), tallLeaf(r, "b", 2), newLeaf(r, "c"))
}
