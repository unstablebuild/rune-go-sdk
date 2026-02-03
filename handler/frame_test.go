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
