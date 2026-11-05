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

func TestBackgroundHandle(t *testing.T) {
	runHandleCases(t, newTestBackground, []handleCase[*Background]{
		{
			name:    "every event goes to the handler, as it is",
			events:  []term.Event{key('x'), key('q'), press(1, 1), wheelDown(1, 1), paste()},
			want:    []string{"X key x", "X key q", "X left 1,1", "X wheeldown 1,1", "X paste"},
			handled: []bool{true, true, true, false, false},
			exit:    []bool{false, true, false, false, false},
		},
		{
			name: "Content is the handler",
			then: func(t *testing.T, b *Background, r *rec) {
				assert.Equal(t, tui.Handler(r.leaf("X")), b.Content())
			},
		},
	})
}

func TestBackgroundViews(t *testing.T) {
	runViewCases(t, newTestBackground, []viewCase[*Background]{
		{
			name:  "the handler is drawn over the background",
			width: 4, height: 2,
			draw:   "X...\n....",
			sizes:  map[string][2]int{"X": {4, 2}},
			cursor: term.Coordinates{X: 1}, show: true, selection: "X",
		},
	})
}

// newTestBackground is X on a background of dots, 4 by 2.
func newTestBackground(r *rec) *Background {
	b := NewBackground(newLeaf(r, "X"), term.NewCell('.', 1, term.Attributes{}))
	b.Resize(4, 2)
	return b
}
