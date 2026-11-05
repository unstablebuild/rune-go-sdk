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
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestFrameUnionHandle(t *testing.T) {
	runHandleCases(t, newTestFrameUnion, []handleCase[*FrameUnion]{
		{
			name:    "keys go to main",
			events:  []term.Event{key('x'), key('n')},
			want:    []string{"M key x", "M key n"},
			handled: []bool{true, false},
		},
		{
			name:   "a click on another member leaves the keys to main",
			events: []term.Event{press(3, 1), release(3, 1), key('x')},
			want:   []string{"T left 3,1", "T release 3,1", "M key x"},
		},
		{
			name:   "keys go where SetFocus moves them",
			setup:  func(u *FrameUnion, r *rec) { require.True(t, u.SetFocus(r.leaf("B"))) },
			events: []term.Event{key('x')},
			want:   []string{"B key x"},
		},
		{
			name:   "SetFocus of a handler that is no member leaves the keys to main",
			setup:  func(u *FrameUnion, _ *rec) { assert.False(t, u.SetFocus(newLeaf(new(rec), "X"))) },
			events: []term.Event{key('x')},
			want:   []string{"M key x"},
		},
		{
			name:   "a click on the capturing member",
			events: []term.Event{press(3, 1), release(3, 1)},
			want:   []string{"T left 3,1", "T release 3,1"},
		},
		{
			name:   "a drag leaving the capturing member keeps unclamped coordinates",
			events: []term.Event{press(3, 1), press(3, 4), press(12, 9), release(2, 7)},
			want:   []string{"T left 3,1", "T left 3,4", "T left 12,9", "T release 2,7"},
		},
		{
			name:   "a drag from main never reaches the capturing member",
			events: []term.Event{press(3, 4), press(3, 1), press(4, 0), release(4, 0), press(4, 0)},
			want:   []string{"M left 3,2", "-", "-", "-", "M move 4,-2", "T left 4,0"},
		},
		{
			name:   "a drag from a member that does not capture routes by position",
			events: []term.Event{press(3, 4), press(3, 7), release(3, 7)},
			want:   []string{"M left 3,2", "B left 3,0", "B release 3,0"},
		},
		{
			name:   "a release ends the gesture",
			events: []term.Event{press(3, 1), release(3, 1), press(3, 4)},
			want:   []string{"T left 3,1", "T release 3,1", "T move 3,4", "M left 3,2"},
		},
		{
			name:   "a motion ends a gesture whose release was missed",
			events: []term.Event{press(3, 1), move(3, 4), press(3, 4)},
			want:   []string{"T left 3,1", "T move 3,4", "M move 3,2", "M left 3,2"},
		},
		{
			name:   "the wheel while a press is held routes by position",
			events: []term.Event{press(3, 1), wheelDown(3, 4), release(3, 4)},
			want:   []string{"T left 3,1", "M wheeldown 3,2", "T release 3,4"},
		},
		{
			name:    "a press outside every member",
			events:  []term.Event{press(20, 20), press(3, 1), release(3, 1)},
			want:    []string{"-", "-", "-"},
			handled: []bool{false, false, false},
		},
		{
			name:   "the pointer onto another member",
			events: []term.Event{move(3, 1), move(3, 4)},
			want:   []string{"T move 3,1", "T move 3,4", "M move 3,2"},
		},
		{
			name:   "the pointer within a member",
			events: []term.Event{move(3, 4), move(5, 5)},
			want:   []string{"M move 3,2", "M move 5,3"},
		},
		{
			name:   "the pointer off every member",
			events: []term.Event{move(3, 4), move(12, 4), move(13, 4)},
			want:   []string{"M move 3,2", "M move 12,2", "-"},
		},
		{
			name:   "the pointer above the union",
			events: []term.Event{move(3, 1), move(3, -1)},
			want:   []string{"T move 3,1", "T move 3,-1"},
		},
		{
			name:   "a parent's leave far off",
			events: []term.Event{move(3, 7), move(-5, -5)},
			want:   []string{"B move 3,0", "B move -5,-12"},
		},
		{
			name:   "no leave while a drag is held",
			events: []term.Event{press(3, 4), press(3, 7), release(3, 7), move(3, 7)},
			want:   []string{"M left 3,2", "B left 3,0", "B release 3,0", "M move 3,5", "B move 3,0"},
		},
		{
			name:   "the leave lands where the member is now, outside it",
			events: []term.Event{move(3, 7), move(3, 7)},
			before: map[int]func(*FrameUnion, *rec){1: func(u *FrameUnion, _ *rec) { u.Resize(10, 12) }},
			want:   []string{"B move 3,0", "B move 3,-4", "M move 3,5"},
		},
		{
			name:    "members on every side",
			fixture: newSidedFrameUnion,
			events: []term.Event{
				press(1, 2), release(1, 2), press(9, 2), release(9, 2), press(4, 2), release(4, 2),
				press(4, 0), release(4, 0), press(4, 4), release(4, 4), press(4, 5), release(4, 5),
			},
			want: []string{
				"L left 1,1", "L release 1,1", "L move 9,1", "R left 1,1", "R release 1,1",
				"R move -4,1", "M left 2,1", "M release 2,1", "M move 2,-1", "T left 4,0", "T release 4,0",
				"T move 4,4", "B2 left 4,0", "B2 release 4,0", "B2 move 4,1", "B1 left 4,0", "B1 release 4,0",
			},
		},
		{
			name:    "where framed members share a row, main gets it, as ComponentAt has it",
			fixture: newFramedFrameUnion,
			events:  []term.Event{press(3, 1), release(3, 1), press(3, 0)},
			want:    []string{"M left 3,0", "M release 3,0", "M move 3,-1", "T left 3,0"},
		},
		{
			name: "Init swaps main and keeps the members around it",
			setup: func(u *FrameUnion, r *rec) {
				u.Init(newLeaf(r, "N"))
				u.Frame = false
				u.Resize(10, 8)
			},
			events: []term.Event{key('x'), press(3, 1), press(3, 4), release(3, 4)},
			want:   []string{"N key x", "T left 3,1", "T left 3,4", "T release 3,4"},
		},
	})
}

func TestFrameUnionViews(t *testing.T) {
	runViewCases(t, newTestFrameUnion, []viewCase[*FrameUnion]{
		{
			name:  "members are stacked around main",
			width: 10, height: 8,
			draw: `
T         
          
M         
          
          
          
          
B         `,
			sizes:     map[string][2]int{"T": {10, 2}, "M": {10, 5}, "B": {10, 1}},
			cursor:    term.Coordinates{X: 1, Y: 2},
			show:      true,
			selection: "M",
		},
		{
			name:    "members on every side",
			fixture: newSidedFrameUnion,
			width:   10, height: 6,
			sizes:     map[string][2]int{"L": {2, 3}, "R": {2, 3}, "T": {10, 1}, "M": {6, 3}, "B1": {10, 1}, "B2": {10, 1}},
			cursor:    term.Coordinates{X: 3, Y: 1},
			show:      true,
			selection: "M",
		},
		{
			name:      "the cursor and selection of the member SetFocus moves the keys to",
			setup:     func(u *FrameUnion, r *rec) { u.SetFocus(r.leaf("B")) },
			cursor:    term.Coordinates{X: 1, Y: 7},
			show:      true,
			selection: "B",
		},
		{
			name: "a cursor past the union is hidden",
			setup: func(u *FrameUnion, r *rec) {
				r.leaf("M").cur.X = 10
			},
			selection: "M",
		},
	})
}

// -- fixtures ---------------------------------------------------------------

// newTestFrameUnion is M with T, 2 rows tall and capturing, on top, and
// B, a row tall, under it, with no frames, 10 by 8. It is Rune's own.
func newTestFrameUnion(r *rec) *FrameUnion {
	u := NewFrameUnion(newLeaf(r, "M"))
	u.UnionTopFrame(newLeaf(r, "T"), 2, false)
	u.UnionBottomFrame(newLeaf(r, "B"), 1, false)
	u.CaptureDrags(r.leaf("T"))
	u.Resize(10, 8)
	return u
}

// newSidedFrameUnion is M with T on top, B1 then B2 under it, and L and
// R, 2 wide, on its sides, with no frames, 10 by 6.
func newSidedFrameUnion(r *rec) *FrameUnion {
	u := NewFrameUnion(newLeaf(r, "M"))
	u.Frame = false
	u.UnionTop(newLeaf(r, "T"), 1)
	u.UnionBottom(newLeaf(r, "B1"), 1)
	u.UnionBottom(newLeaf(r, "B2"), 1)
	u.UnionLeft(newLeaf(r, "L"), 2)
	u.UnionRight(newLeaf(r, "R"), 2)
	u.Resize(10, 6)
	return u
}

// newFramedFrameUnion is M with T, 2 rows tall, on top, their frames
// joined, 10 by 8.
func newFramedFrameUnion(r *rec) *FrameUnion {
	u := NewFrameUnion(newLeaf(r, "M"))
	u.UnionTop(newLeaf(r, "T"), 2)
	u.Resize(10, 8)
	return u
}
