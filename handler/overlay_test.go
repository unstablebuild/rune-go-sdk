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

	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestOverlayHandle(t *testing.T) {
	runHandleCases(t, newTestOverlay, []handleCase[*Overlay]{
		{
			name:    "keys go to the cover",
			events:  []term.Event{key('x'), key('n')},
			want:    []string{"C key x", "C key n"},
			handled: []bool{true, false},
		},
		{
			name:   "keys go to the background once SetFocus moves them there",
			setup:  func(o *Overlay, r *rec) { require.True(t, o.SetFocus(r.leaf("B"))) },
			events: []term.Event{key('x')},
			want:   []string{"B key x"},
		},
		{
			name:   "a click does not move the keys",
			events: []term.Event{press(0, 0), release(0, 0), key('x')},
			want:   []string{"B left 0,0", "B release 0,0", "C key x"},
		},
		{
			name:   "the cover is over the background",
			events: []term.Event{press(4, 1), release(4, 1), press(5, 1)},
			want:   []string{"C left 0,0", "C release 0,0", "C move 1,0", "B left 5,1"},
		},
		{
			name:   "the background keeps the drags of its presses over the cover",
			events: []term.Event{press(0, 0), press(4, 1), release(4, 1)},
			want:   []string{"B left 0,0", "B left 4,1", "B release 4,1"},
		},
		{
			name:   "the cover keeps the drags of its presses over the background",
			events: []term.Event{press(4, 1), press(0, 0), release(0, 0)},
			want:   []string{"C left 0,0", "C left -4,-1", "C release -4,-1"},
		},
		{
			name:    "outside the overlay is nothing",
			events:  []term.Event{press(10, 0)},
			want:    []string{"-"},
			handled: []bool{false},
		},
	})
}

func TestOverlayViews(t *testing.T) {
	runViewCases(t, newTestOverlay, []viewCase[*Overlay]{
		{
			name:  "the cover is placed over the background",
			width: 10, height: 3,
			draw:   "B         \n    C     \n          ",
			sizes:  map[string][2]int{"B": {10, 3}, "C": {1, 1}},
			cursor: term.Coordinates{X: 5, Y: 1}, show: true, selection: "C",
		},
		{
			name:      "the cursor of the background once SetFocus moves the keys there",
			setup:     func(o *Overlay, r *rec) { o.SetFocus(r.leaf("B")) },
			cursor:    term.Coordinates{X: 1},
			show:      true,
			selection: "B",
		},
	})
}

// newTestOverlay is C, centered at its dimensions, over B, 10 by 3.
func newTestOverlay(r *rec) *Overlay {
	o := NewOverlay(newLeaf(r, "B"), newLeaf(r, "C"), term.Attributes{}, component.SpanConfig{
		ContentAlignment: component.AlignmentCentered,
		PadAutoFloating:  true,
	})
	o.Resize(10, 3)
	return o
}
