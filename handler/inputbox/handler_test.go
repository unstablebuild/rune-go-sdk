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

package inputbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/handler/handlertest"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestHandler(t *testing.T) {
	t.Run("basic input", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{InputSequence: "", Expected: "▐         "},
			{InputSequence: "h", Expected: "h▐        "},
			{InputSequence: "e", Expected: "he▐       "},
			{InputSequence: "l", Expected: "hel▐      "},
			{InputSequence: "l", Expected: "hell▐     "},
			{InputSequence: "o", Expected: "hello▐    "},
		}
		handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
	})

	t.Run("input wraps", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{InputSequence: "hello", Expected: "hello\n▐    \n     "},
			{InputSequence: "w", Expected: "hello\nw▐   \n     "},
			{InputSequence: "orld", Expected: "hello\nworld\n▐    "},
			{InputSequence: "!", Expected: "hello\nworld\n!▐   "},
		}
		handlertest.RunHandlerSequence(t, ib, 5, 3, cases)
	})

	t.Run("height with resize", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{InputSequence: "helloworld", Expected: "world\n▐    "},
		}
		handlertest.RunHandlerSequence(t, ib, 5, 2, cases)
		actualHeight := ib.Height(5)
		// 10 chars needs 2 lines, but cursor is wrapped at the next one so Height
		// should return 3; thus we render the text in the first and second
		// line and render the cursor at the last one.
		require.Equal(t, 3, actualHeight)
		cases = []handlertest.SequenceTestCase{
			{InputSequence: "helloworl", Expected: "hello\nworld\nhello\nworl▐"},
		}
		handlertest.RunHandlerSequence(t, ib, 5, 4, cases)
		actualHeight = ib.Height(5)
		require.Equal(t, 4, actualHeight) // 19 chars needs 4 lines
	})
}

func TestHandlerBackspaceAcrossLines(t *testing.T) {
	ib := New()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "helloworld", Expected: "world\n▐    "},
		{
			InputSequence: "<backspace>",
			Expected:      "hello\nworl▐",
		},
		{
			InputSequence: "<backspace><backspace><backspace><backspace>",
			Expected:      "hello\n▐    ",
		},
		{
			InputSequence: "<backspace>",
			Expected:      "hell▐\n     ",
		},
	}
	handlertest.RunHandlerSequence(t, ib, 5, 2, cases)
}

func TestHandlerArrowNavigationWrapped(t *testing.T) {
	ib := New()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "helloworld", Expected: "world\n▐    "},
		{InputSequence: "<left>", Expected: "hello\nworl▐"},
		{
			InputSequence: "<left><left><left><left>",
			Expected:      "hello\n▐orld",
		},
		{InputSequence: "<left>", Expected: "hell▐\nworld"},
		{InputSequence: "<right>", Expected: "hello\n▐orld"},
		{
			InputSequence: "<right><right><right><right>",
			Expected:      "hello\nworl▐",
		},
	}
	handlertest.RunHandlerSequence(t, ib, 5, 2, cases)
}

func TestHandlerHomeEnd(t *testing.T) {
	ib := New()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "hello", Expected: "hello▐    "},
		{InputSequence: "<home>", Expected: "▐ello     "},
		{InputSequence: "<end>", Expected: "hello▐    "},
	}
	handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
}

func TestHandlerEmacsNavigation(t *testing.T) {
	ib := New()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "hello", Expected: "hello▐    "},
		{InputSequence: "<c-a>", Expected: "▐ello     "},
		{InputSequence: "<c-e>", Expected: "hello▐    "},
		{InputSequence: "<c-b>", Expected: "hell▐     "},
		{InputSequence: "<c-f>", Expected: "hello▐    "},
	}
	handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
}

func TestHandlerEmacsEditing(t *testing.T) {
	ib := New()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "hello", Expected: "hello▐    "},
		{InputSequence: "<c-a>", Expected: "▐ello     "},
		{InputSequence: "<c-d>", Expected: "▐llo      "},
		{InputSequence: "<c-e><c-h>", Expected: "ell▐      "},
		{InputSequence: "<c-a><c-k>", Expected: "▐         "},
	}
	handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
}

func TestHandlerCtrlU_KillLine(t *testing.T) {
	ib := New()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "hello", Expected: "hello▐    "},
		{InputSequence: "<c-u>", Expected: "▐         "},
	}
	handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
}

func TestHandlerVerticalScroll(t *testing.T) {
	ib := New()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "hello", Expected: "hello\n▐    "},
		{InputSequence: "world", Expected: "world\n▐    "},
		{InputSequence: "!!!!!", Expected: "!!!!!\n▐    "},
		{InputSequence: "x", Expected: "!!!!!\nx▐   "},
		{InputSequence: "<home>", Expected: "▐ello\nworld"},
		{InputSequence: "<end>", Expected: "!!!!!\nx▐   "},
	}
	handlertest.RunHandlerSequence(t, ib, 5, 2, cases)
}

func TestHandlerText(t *testing.T) {
	ib := New()
	ib.Resize(10, 1)

	assert.Equal(t, "", ib.Text())
	setText(ib, "hello")
	assert.Equal(t, "hello", ib.Text())

	ib.Clear()
	assert.Equal(t, "", ib.Text())
}

func TestHandlerEnterReturnsTrue(t *testing.T) {
	ib := New()
	ib.Resize(10, 1)
	setText(ib, "hello")

	exit, handled := ib.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})
	assert.True(t, exit, "Enter should signal exit")
	assert.True(t, handled)
}

func TestHandlerCursor(t *testing.T) {
	ib := New()
	ib.Resize(10, 1)
	setText(ib, "helloworld")
	ib.Resize(5, 3)

	coords, style, show := ib.Cursor()
	assert.True(t, show)
	assert.Equal(t, term.CursorStyleSteadyBar, style)
	assert.Equal(t, 0, coords.X)
	assert.Equal(t, 2, coords.Y)

	ib.Resize(5, 2)
	ib.Draw(term.NewStringWriter(5, 2))
	coords, style, show = ib.Cursor()
	assert.True(t, show)
	assert.Equal(t, term.CursorStyleSteadyBar, style)
	assert.Equal(t, 0, coords.X)
	assert.Equal(t, 1, coords.Y)
}

func TestHandlerInsertInMiddle(t *testing.T) {
	ib := New()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "helloworld", Expected: "world\n▐    "},
		{InputSequence: "<home>", Expected: "▐ello\nworld"},
		{
			InputSequence: "<right><right><right>",
			Expected:      "hel▐o\nworld",
		},
		{InputSequence: "X", Expected: "helX▐\noworl"},
	}
	handlertest.RunHandlerSequence(t, ib, 5, 2, cases)
}

func TestHandlerPlaceholder(t *testing.T) {
	t.Run("placeholder shown when empty", func(t *testing.T) {
		ib := New(WithPlaceholder("Enter text..."))
		cases := []handlertest.SequenceTestCase{
			{InputSequence: "", Expected: "▐nter text...  "},
		}
		handlertest.RunHandlerSequence(t, ib, 15, 1, cases)
	})

	t.Run("placeholder hidden when typing", func(t *testing.T) {
		ib := New(WithPlaceholder("Enter text..."))
		cases := []handlertest.SequenceTestCase{
			{InputSequence: "", Expected: "▐nter text...       "},
			{InputSequence: "h", Expected: "h▐                  "},
			{InputSequence: "ello", Expected: "hello▐              "},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})

	t.Run("placeholder shown after clear", func(t *testing.T) {
		ib := New(WithPlaceholder("Enter text..."))
		cases := []handlertest.SequenceTestCase{
			{InputSequence: "hello", Expected: "hello▐              "},
			{InputSequence: "<c-u>", Expected: "▐nter text...       "},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})

	t.Run("placeholder wraps", func(t *testing.T) {
		ib := New(WithPlaceholder("Enter your message here"))
		cases := []handlertest.SequenceTestCase{
			{InputSequence: "", Expected: "▐nter\n your\n mess\nage  \nhere "},
		}
		handlertest.RunHandlerSequence(t, ib, 5, 5, cases)
	})
}

// setText is a test helper that types text into the input box
func setText(ib *Handler, s string) {
	for _, ch := range s {
		ib.Handle(term.Event{
			Type: term.EventKey,
			Ch:   ch,
		})
	}
}
