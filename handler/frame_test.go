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

func TestFrameProxyCursor(t *testing.T) {
	handler := &TestHandler{
		CursorPos:   term.Coordinates{X: 1},
		CursorStyle: term.CursorStyleBlinkingBar,
	}
	proxy := NewFrame(handler)
	proxy.Resize(4, 4)
	offsetCursor, _, ok := handler.Cursor()
	require.True(t, ok)
	offsetCursor.X++
	offsetCursor.Y++

	newCursor, style, ok := proxy.Cursor()
	require.True(t, ok)
	assert.Equal(t, offsetCursor, newCursor)
	assert.Equal(t, term.CursorStyleBlinkingBar, style)
}

func TestFrameHandleMouse(t *testing.T) {
	tests := []struct {
		name         string
		key          term.Key
		x, y         int
		wantX, wantY int
	}{
		{name: "motion over the content", x: 2, y: 2, wantX: 1, wantY: 1},
		{name: "motion on the border clamps onto the content", x: 0, y: 2, wantX: 0, wantY: 1},
		{name: "motion left of the frame is left outside the content", x: -1, y: 2, wantX: -2, wantY: 1},
		{name: "motion above the frame is left outside the content", x: 2, y: -1, wantX: 1, wantY: -2},
		{name: "motion below the frame is left outside the content", x: 2, y: 5, wantX: 1, wantY: 4},
		{name: "a drag left of the frame clamps onto the content", key: term.MouseLeft, x: -1, y: 2, wantX: 0, wantY: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got term.Event
			h := &TestHandler{HandleOverride: func(ev term.Event) (bool, bool) {
				got = ev
				return false, true
			}}
			f := NewFrame(h)
			f.Resize(6, 5)
			f.Handle(term.Event{Type: term.EventMouse, Key: tc.key, MouseX: tc.x, MouseY: tc.y})
			assert.Equal(t, tc.wantX, got.MouseX, "x")
			assert.Equal(t, tc.wantY, got.MouseY, "y")
		})
	}
}
