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

package handlertest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// SequenceTestCase represents an input sequence and
// the result expected draw string representation.
type SequenceTestCase struct {
	InputSequence string
	Expected      string
}

// SingleTestCase represents an event and the result
// expected draw string representation.
type SingleTestCase struct {
	Event    term.Event
	Expected string
}

// DrawHandler is a helper function that renders the handler into a string.
//
// It can be beneficial for print-debugging your tests:
//
// fmt.Printf("\n%s", handlertest.DrawHandler(vi, 20, 20))
func DrawHandler(handler tui.Handler, width, height int) string {
	w := term.NewStringWriter(width, height)
	handler.Draw(w)
	cursor, _, ok := handler.Cursor()
	if ok {
		w.SetCursor(cursor)
	}
	_ = w.Flush()
	out := w.String()
	return out
}

// RunHandlerIsolated is a helper function that runs a set of test cases,
// by calling fn for every test case, resizing the given handler with
// the given width and height and comparing the results of Draw against it.
func RunHandlerIsolated(
	t *testing.T, fn func(t *testing.T) tui.Handler, width, height int,
	cases []SequenceTestCase,
) {
	writer := term.NewStringWriter(width, height)

	for i, tcase := range cases {
		t.Run(fmt.Sprintf("test case %d (input: %s)", i, tcase.InputSequence),
			func(t *testing.T) {
				handler := fn(t)
				handler.Resize(width, height)
				runTestCase(t, i, writer, handler, tcase)
			})
	}
}

// RunHandlerSequenceWriter is a helper function that runs a set of test cases
// against the given handler in sequence.
func RunHandlerSequenceWriter(
	t *testing.T, writer *term.StringWriter, handler tui.Handler, width, height int,
	cases []SequenceTestCase,
) {
	handler.Resize(width, height)
	for i, tcase := range cases {
		runTestCase(t, i, writer, handler, tcase)
	}
}

// RunHandlerSequence is a helper function that runs a set of test cases
// against the given handler in sequence, with a StringWriter set with the given
// width and height.
func RunHandlerSequence(
	t *testing.T, handler tui.Handler, width, height int,
	cases []SequenceTestCase,
) {
	writer := term.NewStringWriter(width, height)
	RunHandlerSequenceWriter(t, writer, handler, width, height, cases)
}

func runTestCase(
	t *testing.T, i int, w *term.StringWriter,
	h tui.Handler, tcase SequenceTestCase,
) {
	err := w.Clear(term.Attributes{Fg: 0, Bg: 0})
	require.NoError(t, err)

	keys, err := term.ParseKeys(tcase.InputSequence)
	require.NoError(t, err)

	for _, key := range keys {
		evType := term.EventKey
		if isMouseKey(key.Key) {
			evType = term.EventMouse
		}
		h.Handle(term.Event{Ch: key.Ch, Mod: key.Mod, Key: key.Key, Type: evType})
	}
	h.Draw(w)

	cursor, _, ok := h.Cursor()
	if ok {
		w.SetCursor(cursor)
	}

	err = w.Flush()
	require.NoError(t, err)

	out := w.String()
	assert.Equal(t, tcase.Expected, out, "test case %d (input: %s)", i, tcase.InputSequence)
}

func isMouseKey(k term.Key) bool {
	switch k {
	case term.MouseLeft, term.MouseMiddle, term.MouseRight,
		term.MouseRelease, term.MouseWheelUp, term.MouseWheelDown:
		return true
	}
	return false
}
