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

func TestReferenceHandle(t *testing.T) {
	runHandleCases(t, newTestReference, []handleCase[*Reference]{
		{
			name:    "every event goes to the handler, as it is",
			events:  []term.Event{key('x'), press(1, 1), wheelUp(0, 0)},
			want:    []string{"X key x", "X left 1,1", "X wheelup 0,0"},
			handled: []bool{true, true, false},
		},
		{
			name:    "with no handler, nothing is handled",
			fixture: func(*rec) *Reference { return NewReference(nil) },
			events:  []term.Event{key('x'), press(1, 1)},
			want:    []string{"-", "-"},
			handled: []bool{false, false},
			then: func(t *testing.T, ref *Reference, _ *rec) {
				assert.Nil(t, ref.Content())
			},
		},
		{
			name:   "Init swaps the handler",
			setup:  func(ref *Reference, r *rec) { ref.Init(newLeaf(r, "Y")) },
			events: []term.Event{key('x')},
			want:   []string{"Y key x"},
			then: func(t *testing.T, ref *Reference, r *rec) {
				assert.Equal(t, tui.Handler(r.leaf("Y")), ref.Content())
			},
		},
	})
}

func TestReferenceViews(t *testing.T) {
	runViewCases(t, newTestReference, []viewCase[*Reference]{
		{
			name:  "the handler is drawn",
			width: 3, height: 1,
			draw:   "X  ",
			sizes:  map[string][2]int{"X": {3, 1}},
			cursor: term.Coordinates{X: 1}, show: true, selection: "X",
		},
		{
			name:    "with no handler, nothing is drawn or shown",
			fixture: func(*rec) *Reference { return NewReference(nil) },
			width:   3, height: 1,
			draw: "   ",
		},
	})
}

func newTestReference(r *rec) *Reference {
	ref := NewReference(newLeaf(r, "X"))
	ref.Resize(3, 1)
	return ref
}
