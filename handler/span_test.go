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
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestSpanCursor(t *testing.T) {
	tests := []struct {
		name      string
		cfg       component.SpanConfig
		cursorPos term.Coordinates
		spanW     int
		spanH     int
		wantPos   term.Coordinates
	}{
		{
			name: "centered padding offsets cursor",
			cfg: component.SpanConfig{
				PadHorizontal:    4,
				PadVertical:      2,
				ContentAlignment: component.AlignmentCentered,
			},
			cursorPos: term.Coordinates{X: 1, Y: 0},
			spanW:     20,
			spanH:     10,
			wantPos:   term.Coordinates{X: 3, Y: 1},
		},
		{
			name: "top-left alignment no offset",
			cfg: component.SpanConfig{
				PadHorizontal:    4,
				PadVertical:      2,
				ContentAlignment: 0,
			},
			cursorPos: term.Coordinates{X: 1, Y: 0},
			spanW:     20,
			spanH:     10,
			wantPos:   term.Coordinates{X: 1, Y: 0},
		},
		{
			name: "bottom-right alignment full offset",
			cfg: component.SpanConfig{
				PadHorizontal:    4,
				PadVertical:      2,
				ContentAlignment: component.AlignmentBottom |
					component.AlignmentRight,
			},
			cursorPos: term.Coordinates{X: 1, Y: 1},
			spanW:     20,
			spanH:     10,
			wantPos:   term.Coordinates{X: 5, Y: 3},
		},
		{
			name: "no padding no offset",
			cfg: component.SpanConfig{
				ContentAlignment: component.AlignmentCentered,
			},
			cursorPos: term.Coordinates{X: 3, Y: 2},
			spanW:     20,
			spanH:     10,
			wantPos:   term.Coordinates{X: 3, Y: 2},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &TestHandler{
				CursorPos:   tc.cursorPos,
				CursorStyle: term.CursorStyleBlinkingBar,
			}
			s := NewSpan(h, tc.cfg)
			s.Resize(tc.spanW, tc.spanH)

			pos, style, show := s.Cursor()
			require.True(t, show)
			assert.Equal(t, tc.wantPos, pos)
			assert.Equal(t, term.CursorStyleBlinkingBar, style)
		})
	}
}

func TestSpanHandle(t *testing.T) {
	tests := []struct {
		name    string
		cfg     component.SpanConfig
		mouseX  int
		mouseY  int
		spanW   int
		spanH   int
		wantX   int
		wantY   int
	}{
		{
			name: "centered padding adjusts mouse",
			cfg: component.SpanConfig{
				PadHorizontal:    4,
				PadVertical:      2,
				ContentAlignment: component.AlignmentCentered,
			},
			mouseX: 5,
			mouseY: 3,
			spanW:  20,
			spanH:  10,
			wantX:  3,
			wantY:  2,
		},
		{
			name: "top-left alignment no adjustment",
			cfg: component.SpanConfig{
				PadHorizontal:    4,
				PadVertical:      2,
				ContentAlignment: 0,
			},
			mouseX: 5,
			mouseY: 3,
			spanW:  20,
			spanH:  10,
			wantX:  5,
			wantY:  3,
		},
		{
			name: "mouse in padding clamps to zero",
			cfg: component.SpanConfig{
				PadHorizontal:    4,
				PadVertical:      2,
				ContentAlignment: component.AlignmentCentered,
			},
			mouseX: 1,
			mouseY: 0,
			spanW:  20,
			spanH:  10,
			wantX:  0,
			wantY:  0,
		},
		{
			name: "bottom-right alignment full offset",
			cfg: component.SpanConfig{
				PadHorizontal:    4,
				PadVertical:      2,
				ContentAlignment: component.AlignmentBottom |
					component.AlignmentRight,
			},
			mouseX: 6,
			mouseY: 3,
			spanW:  20,
			spanH:  10,
			wantX:  2,
			wantY:  1,
		},
		{
			name: "non-mouse event passes through unchanged",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "non-mouse event passes through unchanged" {
				h := NewTestHandler()
				s := NewSpan(h, component.SpanConfig{
					PadHorizontal:    4,
					PadVertical:      2,
					ContentAlignment: component.AlignmentCentered,
				})
				s.Resize(20, 10)

				ev := term.Event{Type: term.EventKey, Ch: 'a'}
				s.Handle(ev)
				// TestHandler increments Ch on Handle;
				// verify it was called.
				assert.Equal(t, 'B', h.Ch)
				return
			}

			var got term.Event
			h := &TestHandler{}
			h.HandleOverride = func(ev term.Event) (bool, bool) {
				got = ev
				return false, true
			}
			s := NewSpan(h, tc.cfg)
			s.Resize(tc.spanW, tc.spanH)

			ev := term.Event{
				Type:   term.EventMouse,
				MouseX: tc.mouseX,
				MouseY: tc.mouseY,
			}
			s.Handle(ev)

			assert.Equal(t, tc.wantX, got.MouseX)
			assert.Equal(t, tc.wantY, got.MouseY)
		})
	}
}
