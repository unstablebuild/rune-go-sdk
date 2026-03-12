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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/handler/handlertest"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/tcell/v3"
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

	t.Run("basic input + unknown modifier does nothing", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{InputSequence: "<m-c>", Expected: "▐         "},
			{InputSequence: "<m-space>", Expected: "▐         "},
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
		ib := New(WithPlaceholderText("Enter text..."))
		cases := []handlertest.SequenceTestCase{
			{InputSequence: "", Expected: "▐nter text...  "},
		}
		handlertest.RunHandlerSequence(t, ib, 15, 1, cases)
	})

	t.Run("placeholder hidden when typing", func(t *testing.T) {
		ib := New(WithPlaceholderText("Enter text..."))
		cases := []handlertest.SequenceTestCase{
			{InputSequence: "", Expected: "▐nter text...       "},
			{InputSequence: "h", Expected: "h▐                  "},
			{InputSequence: "ello", Expected: "hello▐              "},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})

	t.Run("placeholder shown after clear", func(t *testing.T) {
		ib := New(WithPlaceholderText("Enter text..."))
		cases := []handlertest.SequenceTestCase{
			{InputSequence: "hello", Expected: "hello▐              "},
			{InputSequence: "<c-u>", Expected: "▐nter text...       "},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})

	t.Run("placeholder wraps", func(t *testing.T) {
		ib := New(WithPlaceholderText("Enter your message here"))
		cases := []handlertest.SequenceTestCase{
			{InputSequence: "", Expected: "▐nter\n your\n mess\nage  \nhere "},
		}
		handlertest.RunHandlerSequence(t, ib, 5, 5, cases)
	})

	t.Run("placeholder with custom config", func(t *testing.T) {
		ib := New(WithPlaceholder("Test", component.StringResponsiveConfig{
			StringConfig: component.StringConfig{
				Attributes: term.Attributes{
					Fg: tcell.ColorBlue,
				},
			},
		}))
		cases := []handlertest.SequenceTestCase{
			{InputSequence: "", Expected: "▐est      "},
		}
		handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
	})
}

func TestHandlerBackground(t *testing.T) {

	t.Run("background fills entire area with WithAttributes", func(t *testing.T) {
		ib := New(WithAttributes(term.Attributes{
			Fg: tcell.ColorWhite,
			Bg: tcell.ColorBlue,
		}))
		width, height := 10, 3
		ib.Resize(width, height)
		w := term.NewStringWriter(width, height)
		ib.Draw(w)

		cells := w.Cells()
		for y := range height {
			for x := range width {
				idx := y*width + x
				assert.Equal(t, tcell.ColorBlue, cells[idx].Bg,
					"cell at (%d, %d) should have blue background", x, y)
			}
		}
	})

	t.Run("background fills entire area with SetAttr", func(t *testing.T) {
		ib := New()
		ib.SetAttr(term.Attributes{
			Fg: tcell.ColorYellow,
			Bg: tcell.ColorRed,
		})
		width, height := 5, 2
		ib.Resize(width, height)
		w := term.NewStringWriter(width, height)
		ib.Draw(w)

		cells := w.Cells()
		for y := range height {
			for x := range width {
				idx := y*width + x
				assert.Equal(t, tcell.ColorRed, cells[idx].Bg,
					"cell at (%d, %d) should have red background", x, y)
			}
		}
	})

	t.Run("background with text content", func(t *testing.T) {
		ib := New(WithAttributes(term.Attributes{
			Fg: tcell.ColorWhite,
			Bg: tcell.ColorGreen,
		}))
		width, height := 10, 2
		ib.Resize(width, height)
		setText(ib, "hello")
		w := term.NewStringWriter(width, height)
		ib.Draw(w)

		cells := w.Cells()
		for y := range height {
			for x := range width {
				idx := y*width + x
				assert.Equal(t, tcell.ColorGreen, cells[idx].Bg,
					"cell at (%d, %d) should have green background", x, y)
			}
		}

		// First 5 cells on first line should have text
		for x := range 5 {
			assert.NotEqual(t, rune(0), cells[x].Ch,
				"cell at (%d, 0) should have text", x)
		}
	})

	t.Run("background with placeholder", func(t *testing.T) {
		ib := New(
			WithAttributes(term.Attributes{
				Fg: tcell.ColorWhite,
				Bg: tcell.ColorPurple,
			}),
			WithPlaceholderText("placeholder"),
		)
		width, height := 15, 1
		ib.Resize(width, height)
		w := term.NewStringWriter(width, height)
		ib.Draw(w)

		cells := w.Cells()
		for x := range width {
			assert.Equal(t, tcell.ColorPurple, cells[x].Bg,
				"cell at (%d, 0) should have purple background", x)
		}
	})
}

func TestHandlerWordMovement(t *testing.T) {
	t.Run("ctrl+left and ctrl+right", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "hello<space>world",
				Expected:      "hello world▐        ",
			},
			{
				InputSequence: "<ctrl-left>",
				Expected:      "hello ▐orld         ",
			},
			{
				InputSequence: "<ctrl-left>",
				Expected:      "▐ello world         ",
			},
			{
				InputSequence: "<ctrl-right>",
				Expected:      "hello▐world         ",
			},
			{
				InputSequence: "<ctrl-right>",
				Expected:      "hello world▐        ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})

	t.Run("alt+left and alt+right", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "foo<space>bar",
				Expected:      "foo bar▐            ",
			},
			{
				InputSequence: "<alt-left>",
				Expected:      "foo ▐ar             ",
			},
			{
				InputSequence: "<alt-left>",
				Expected:      "▐oo bar             ",
			},
			{
				InputSequence: "<alt-right>",
				Expected:      "foo▐bar             ",
			},
			{
				InputSequence: "<alt-right>",
				Expected:      "foo bar▐            ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})

	t.Run("alt+b and alt+f", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "one<space>two",
				Expected:      "one two▐            ",
			},
			{
				InputSequence: "<a-b>",
				Expected:      "one ▐wo             ",
			},
			{
				InputSequence: "<a-b>",
				Expected:      "▐ne two             ",
			},
			{
				InputSequence: "<a-f>",
				Expected:      "one▐two             ",
			},
			{
				InputSequence: "<a-f>",
				Expected:      "one two▐            ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})

	t.Run("word movement with multiple spaces", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "a<space><space>b",
				Expected:      "a  b▐               ",
			},
			{
				InputSequence: "<ctrl-left>",
				Expected:      "a  ▐                ",
			},
			{
				InputSequence: "<ctrl-left>",
				Expected:      "▐  b                ",
			},
			{
				InputSequence: "<ctrl-right>",
				Expected:      "a▐ b                ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})
}

func TestHandlerWordDeletion(t *testing.T) {
	t.Run("ctrl+backspace", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "hello<space>world",
				Expected:      "hello world▐        ",
			},
			{
				InputSequence: "<ctrl-backspace>",
				Expected:      "hello ▐             ",
			},
			{
				InputSequence: "<ctrl-backspace>",
				Expected:      "▐                   ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})

	t.Run("alt+backspace", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "foo<space>bar",
				Expected:      "foo bar▐            ",
			},
			{
				InputSequence: "<alt-backspace>",
				Expected:      "foo ▐               ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})

	t.Run("ctrl+delete", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "hello<space>world",
				Expected:      "hello world▐        ",
			},
			{
				InputSequence: "<home>",
				Expected:      "▐ello world         ",
			},
			{
				// deletes "hello", leaves " world";
				// cursor at 0 covers the space
				InputSequence: "<ctrl-delete>",
				Expected:      "▐world              ",
			},
			{
				// deletes " world" (space then word)
				InputSequence: "<ctrl-delete>",
				Expected:      "▐                   ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})

	t.Run("alt+delete", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "one<space>two",
				Expected:      "one two▐            ",
			},
			{
				InputSequence: "<home>",
				Expected:      "▐ne two             ",
			},
			{
				// deletes "one", cursor covers the space
				InputSequence: "<alt-delete>",
				Expected:      "▐two                ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})

	t.Run("alt+d delete word forward", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "abc<space>def",
				Expected:      "abc def▐            ",
			},
			{
				InputSequence: "<home>",
				Expected:      "▐bc def             ",
			},
			{
				// deletes "abc", cursor covers the space
				InputSequence: "<a-d>",
				Expected:      "▐def                ",
			},
			{
				// deletes " def" (space then word)
				InputSequence: "<a-d>",
				Expected:      "▐                   ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})

	t.Run("word deletion in middle", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "one<space>two<space>three",
				Expected:      "one two three▐      ",
			},
			{
				InputSequence: "<ctrl-left>",
				Expected:      "one two ▐hree       ",
			},
			{
				InputSequence: "<ctrl-backspace>",
				Expected:      "one ▐hree           ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})
}

func TestHandlerTranspose(t *testing.T) {
	t.Run("basic transpose", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "abc",
				Expected:      "abc▐      ",
			},
			{
				InputSequence: "<c-t>",
				Expected:      "acb▐      ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
	})

	t.Run("transpose at beginning does nothing", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "ab",
				Expected:      "ab▐       ",
			},
			{
				InputSequence: "<home>",
				Expected:      "▐b        ",
			},
			{
				InputSequence: "<c-t>",
				Expected:      "▐b        ",
			},
			{
				InputSequence: "<right>",
				Expected:      "a▐        ",
			},
			{
				InputSequence: "<c-t>",
				Expected:      "a▐        ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
	})

	t.Run("transpose with two chars", func(t *testing.T) {
		ib := New()
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "ab",
				Expected:      "ab▐       ",
			},
			{
				InputSequence: "<c-t>",
				Expected:      "ba▐       ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
	})
}

func TestHandlerShiftArrowSelection(t *testing.T) {
	t.Run("shift+left selects chars backward", func(t *testing.T) {
		ib := New()
		ib.Resize(20, 1)
		setText(ib, "hello")

		sendKeys(t, ib, "<shift-left><shift-left>")
		sel, ok := ib.Selection()
		assert.True(t, ok)
		assert.Equal(t, "lo", sel)
	})

	t.Run("shift+right selects chars forward", func(t *testing.T) {
		ib := New()
		ib.Resize(20, 1)
		setText(ib, "hello")
		sendKeys(t, ib, "<home>")

		sendKeys(t, ib, "<shift-right><shift-right><shift-right>")
		sel, ok := ib.Selection()
		assert.True(t, ok)
		assert.Equal(t, "hel", sel)
	})

	t.Run("shift+home selects to start", func(t *testing.T) {
		ib := New()
		ib.Resize(20, 1)
		setText(ib, "hello")

		sendKeys(t, ib, "<shift-home>")
		sel, ok := ib.Selection()
		assert.True(t, ok)
		assert.Equal(t, "hello", sel)
	})

	t.Run("shift+end selects to end", func(t *testing.T) {
		ib := New()
		ib.Resize(20, 1)
		setText(ib, "hello")
		sendKeys(t, ib, "<home>")

		sendKeys(t, ib, "<shift-end>")
		sel, ok := ib.Selection()
		assert.True(t, ok)
		assert.Equal(t, "hello", sel)
	})

	t.Run("shift+ctrl+left selects word backward", func(t *testing.T) {
		ib := New()
		ib.Resize(20, 1)
		setText(ib, "hello")
		sendKey(ib, ' ')
		setText(ib, "world")

		sendKeys(t, ib, "<ctrl-shift-left>")
		sel, ok := ib.Selection()
		assert.True(t, ok)
		assert.Equal(t, "world", sel)
	})

	t.Run("shift+ctrl+right selects word forward", func(t *testing.T) {
		ib := New()
		ib.Resize(20, 1)
		setText(ib, "hello")
		sendKey(ib, ' ')
		setText(ib, "world")
		sendKeys(t, ib, "<home>")

		sendKeys(t, ib, "<ctrl-shift-right>")
		sel, ok := ib.Selection()
		assert.True(t, ok)
		assert.Equal(t, "hello", sel)
	})

	t.Run("shift+alt+left selects word backward", func(t *testing.T) {
		ib := New()
		ib.Resize(20, 1)
		setText(ib, "foo")
		sendKey(ib, ' ')
		setText(ib, "bar")

		sendKeys(t, ib, "<alt-shift-left>")
		sel, ok := ib.Selection()
		assert.True(t, ok)
		assert.Equal(t, "bar", sel)
	})

	t.Run("shift+alt+right selects word forward", func(t *testing.T) {
		ib := New()
		ib.Resize(20, 1)
		setText(ib, "foo")
		sendKey(ib, ' ')
		setText(ib, "bar")
		sendKeys(t, ib, "<home>")

		sendKeys(t, ib, "<alt-shift-right>")
		sel, ok := ib.Selection()
		assert.True(t, ok)
		assert.Equal(t, "foo", sel)
	})

	t.Run("plain arrow clears selection", func(t *testing.T) {
		ib := New()
		ib.Resize(20, 1)
		setText(ib, "hello")

		sendKeys(t, ib, "<shift-left><shift-left>")
		_, ok := ib.Selection()
		assert.True(t, ok)

		sendKeys(t, ib, "<right>")
		_, ok = ib.Selection()
		assert.False(t, ok)
	})

	t.Run("typing replaces selection", func(t *testing.T) {
		ib := New()
		ib.Resize(20, 1)
		setText(ib, "hello")

		// select "lo"
		sendKeys(t, ib, "<shift-left><shift-left>")
		sel, ok := ib.Selection()
		require.True(t, ok)
		require.Equal(t, "lo", sel)

		// type "p" to replace
		setText(ib, "p")
		assert.Equal(t, "help", ib.Text())
	})

	t.Run("backspace deletes selection", func(t *testing.T) {
		ib := New()
		ib.Resize(20, 1)
		setText(ib, "hello")

		sendKeys(t, ib, "<shift-left><shift-left><shift-left>")
		sel, ok := ib.Selection()
		require.True(t, ok)
		require.Equal(t, "llo", sel)

		sendKeys(t, ib, "<backspace>")
		assert.Equal(t, "he", ib.Text())
		_, ok = ib.Selection()
		assert.False(t, ok)
	})

	t.Run("delete deletes selection", func(t *testing.T) {
		ib := New()
		ib.Resize(20, 1)
		setText(ib, "hello")
		sendKeys(t, ib, "<home>")

		sendKeys(t, ib, "<shift-right><shift-right>")
		sel, ok := ib.Selection()
		require.True(t, ok)
		require.Equal(t, "he", sel)

		sendKeys(t, ib, "<delete>")
		assert.Equal(t, "llo", ib.Text())
	})

	t.Run("extend and shrink selection", func(t *testing.T) {
		ib := New()
		ib.Resize(20, 1)
		setText(ib, "abcdef")

		// select "ef" backward
		sendKeys(t, ib, "<shift-left><shift-left>")
		sel, ok := ib.Selection()
		require.True(t, ok)
		assert.Equal(t, "ef", sel)

		// extend to "def"
		sendKeys(t, ib, "<shift-left>")
		sel, ok = ib.Selection()
		require.True(t, ok)
		assert.Equal(t, "def", sel)

		// shrink back to "ef"
		sendKeys(t, ib, "<shift-right>")
		sel, ok = ib.Selection()
		require.True(t, ok)
		assert.Equal(t, "ef", sel)
	})

	t.Run("selection visible in draw", func(t *testing.T) {
		ib := New()
		width := 10
		ib.Resize(width, 1)
		setText(ib, "hello")
		sendKeys(t, ib, "<shift-left><shift-left>")

		w := term.NewStringWriter(width, 1)
		ib.Draw(w)

		cells := w.Cells()
		// "hel" (indices 0-2) should NOT be reversed
		for i := range 3 {
			assert.Equal(
				t, tcell.AttrNone, cells[i].Attrs,
				"cell %d should not be reversed", i,
			)
		}
		// "lo" (indices 3-4) should be reversed
		for i := 3; i < 5; i++ {
			assert.Equal(
				t, tcell.AttrReverse, cells[i].Attrs,
				"cell %d should be reversed", i,
			)
		}
	})
}

// sendKeys parses a key sequence string and sends the events to the handler.
func sendKeys(t *testing.T, ib *Handler, seq string) {
	t.Helper()
	keys, err := term.ParseKeys(seq)
	require.NoError(t, err)
	for _, k := range keys {
		ib.Handle(term.Event{
			Type: term.EventKey,
			Ch:   k.Ch,
			Mod:  k.Mod,
			Key:  k.Key,
		})
	}
}

// sendKey sends a single character event to the handler.
func sendKey(ib *Handler, ch rune) {
	ib.Handle(term.Event{Type: term.EventKey, Ch: ch})
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

func drawHandler(h *Handler, w, hh int) string {
	return handlertest.DrawHandler(h, w, hh)
}

func TestDimensions(t *testing.T) {
	cases := []struct {
		name   string
		opts   []Option
		input  string
		wantW  int
		wantH  int
	}{
		{
			name:  "empty no prompt",
			wantW: 1,
			wantH: 1,
		},
		{
			name:  "text no prompt",
			input: "hello",
			wantW: 6,
			wantH: 1,
		},
		{
			name:  "empty with prompt",
			opts:  []Option{WithPrompt("> ")},
			wantW: 3,
			wantH: 1,
		},
		{
			name:  "text with prompt",
			opts:  []Option{WithPrompt("> ")},
			input: "hello",
			wantW: 8,
			wantH: 1,
		},
		{
			name:  "long text",
			input: "hello world",
			wantW: 12,
			wantH: 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ib := New(tc.opts...)
			ib.Resize(20, 1)
			setText(ib, tc.input)
			w, h := ib.Dimensions()
			assert.Equal(t, tc.wantW, w)
			assert.Equal(t, tc.wantH, h)
		})
	}
}

func TestPromptDisplay(t *testing.T) {
	t.Run("prompt rendered on first line", func(t *testing.T) {
		ib := New(WithPrompt("> "))
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "",
				Expected:      "> ▐                 ",
			},
			{
				InputSequence: "h",
				Expected:      "> h▐                ",
			},
			{
				InputSequence: "ello",
				Expected:      "> hello▐            ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})

	t.Run("text wraps after prompt", func(t *testing.T) {
		ib := New(WithPrompt("> "))
		// width=5, promptWidth=2, firstLineChars=3
		// "hello" => line0: "> hel", line1: "lo"
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "hello",
				Expected:      "> hel\nlo▐  ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 5, 2, cases)
	})

	t.Run("prompt not shown when scrolled past", func(t *testing.T) {
		// With a long text that scrolls the prompt off-screen,
		// the prompt should not appear.
		ib := New(WithPrompt("> "))
		// width=5, promptWidth=2, firstLineChars=3
		// "abcdefghij" = 10 chars
		// line0: "> abc" (3 chars), line1: "defgh" (5 chars), line2: "ij" (2 chars), line3: cursor
		// With height=2, should show last 2 visible lines.
		ib.Resize(5, 2)
		setText(ib, "abcdefghij")
		got := drawHandler(ib, 5, 2)
		assert.Contains(t, got, "ij")
	})
}

func TestPromptHeight(t *testing.T) {
	t.Run("empty with prompt", func(t *testing.T) {
		ib := New(WithPrompt("> "))
		assert.Equal(t, 1, ib.Height(20))
	})

	t.Run("short text with prompt", func(t *testing.T) {
		ib := New(WithPrompt("> "))
		ib.Resize(20, 1)
		setText(ib, "hello")
		assert.Equal(t, 1, ib.Height(20))
	})

	t.Run("text wraps with prompt", func(t *testing.T) {
		ib := New(WithPrompt("> "))
		// width=5, promptWidth=2, firstLineChars=3
		ib.Resize(5, 3)
		setText(ib, "hello")
		// "hel" on line 0, "lo" on line 1, cursor after 'o'
		assert.Equal(t, 2, ib.Height(5))
	})
}

func TestHistoryUpDown(t *testing.T) {
	ib := New(
		WithHistory([]string{"first", "second", "third"}),
	)
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "", Expected: "▐                   "},
		{InputSequence: "<up>", Expected: "third▐              "},
		{InputSequence: "<up>", Expected: "second▐             "},
		{InputSequence: "<up>", Expected: "first▐              "},
		{InputSequence: "<up>", Expected: "first▐              "},
		{InputSequence: "<down>", Expected: "second▐             "},
		{InputSequence: "<down>", Expected: "third▐              "},
		{InputSequence: "<down>", Expected: "▐                   "},
	}
	handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
}

func TestHistoryPreservesLine(t *testing.T) {
	ib := New(
		WithHistory([]string{"old"}),
	)
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "current", Expected: "current▐            "},
		{InputSequence: "<up>", Expected: "old▐                "},
		{InputSequence: "<down>", Expected: "current▐            "},
	}
	handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
}

func TestHistoryEmacsBindings(t *testing.T) {
	ib := New(
		WithHistory([]string{"first", "second"}),
	)
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "", Expected: "▐                   "},
		{InputSequence: "<c-p>", Expected: "second▐             "},
		{InputSequence: "<c-p>", Expected: "first▐              "},
		{InputSequence: "<c-n>", Expected: "second▐             "},
		{InputSequence: "<c-n>", Expected: "▐                   "},
	}
	handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
}

func TestReverseSearch(t *testing.T) {
	t.Run("basic search", func(t *testing.T) {
		ib := New(
			WithHistory([]string{
				"git commit", "git push", "make test",
			}),
		)
		ib.Resize(40, 1)
		sendKeys(t, ib, "<c-r>")
		assert.True(t, ib.searching)

		sendKeys(t, ib, "git")
		got := drawHandler(ib, 40, 1)
		assert.Contains(t, got, "git")
	})

	t.Run("search cancel restores line", func(t *testing.T) {
		ib := New(
			WithHistory([]string{"old"}),
		)
		ib.Resize(20, 1)
		sendKeys(t, ib, "current")
		sendKeys(t, ib, "<c-r>")
		sendKeys(t, ib, "old")
		// Ctrl+G cancels
		sendKeys(t, ib, "<c-g>")
		assert.False(t, ib.searching)
		assert.Equal(t, "current", ib.Text())
	})

	t.Run("search accept via enter", func(t *testing.T) {
		ib := New(
			WithHistory([]string{
				"first", "second",
			}),
		)
		ib.Resize(20, 1)
		sendKeys(t, ib, "<c-r>")
		sendKeys(t, ib, "first")
		exit, _ := ib.Handle(term.Event{
			Type: term.EventKey,
			Key:  term.KeyEnter,
		})
		assert.True(t, exit)
		assert.Equal(t, "first", ib.Text())
	})

	t.Run("search cycling", func(t *testing.T) {
		ib := New(
			WithHistory([]string{
				"alpha", "beta", "alpha2",
			}),
		)
		ib.Resize(40, 1)
		sendKeys(t, ib, "<c-r>")
		sendKeys(t, ib, "alpha")
		got1 := drawHandler(ib, 40, 1)
		assert.Contains(t, got1, "alpha2")

		sendKeys(t, ib, "<c-r>")
		got2 := drawHandler(ib, 40, 1)
		assert.Contains(t, got2, "alpha")
	})

	t.Run("ctrl-c cancels search", func(t *testing.T) {
		ib := New(
			WithHistory([]string{"old"}),
		)
		ib.Resize(20, 1)
		sendKeys(t, ib, "current")
		sendKeys(t, ib, "<c-r>")
		sendKeys(t, ib, "old")
		assert.True(t, ib.searching)
		sendKeys(t, ib, "<c-c>")
		assert.False(t, ib.searching)
		assert.Equal(t, "current", ib.Text())
	})

	t.Run("tab accepts search without submit", func(t *testing.T) {
		ib := New(
			WithHistory([]string{"make test", "make lint"}),
		)
		ib.Resize(20, 1)
		sendKeys(t, ib, "<c-r>")
		sendKeys(t, ib, "test")
		assert.True(t, ib.searching)
		exit, handled := ib.Handle(term.Event{
			Type: term.EventKey,
			Key:  term.KeyTab,
		})
		assert.False(t, exit, "tab should not exit")
		assert.True(t, handled)
		assert.False(t, ib.searching, "search should end")
		assert.Equal(t, "make test", ib.Text())
	})
}

func TestTabCircular(t *testing.T) {
	t.Run("single match auto-completes", func(t *testing.T) {
		completer := func(line string, pos int) (string, []string, string) {
			return "", []string{"foobar"}, ""
		}
		ib := New(WithWordCompleter(completer))
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "foo",
				Expected:      "foo▐                ",
			},
			{
				InputSequence: "<tab>",
				Expected:      "foobar▐             ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})

	t.Run("multiple matches cycle", func(t *testing.T) {
		completer := func(line string, pos int) (string, []string, string) {
			return "", []string{"abc", "abd", "abe"}, ""
		}
		ib := New(WithWordCompleter(completer))
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "ab",
				Expected:      "ab▐                 ",
			},
			{
				InputSequence: "<tab>",
				Expected:      "abc▐                ",
			},
			{
				InputSequence: "<tab>",
				Expected:      "abd▐                ",
			},
			{
				InputSequence: "<tab>",
				Expected:      "abe▐                ",
			},
			{
				InputSequence: "<tab>",
				Expected:      "abc▐                ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})

	t.Run("escape cancels completion", func(t *testing.T) {
		completer := func(line string, pos int) (string, []string, string) {
			return "", []string{"abc", "abd"}, ""
		}
		ib := New(WithWordCompleter(completer))
		cases := []handlertest.SequenceTestCase{
			{
				InputSequence: "ab",
				Expected:      "ab▐                 ",
			},
			{
				InputSequence: "<tab>",
				Expected:      "abc▐                ",
			},
			{
				InputSequence: "<esc>",
				Expected:      "ab▐                 ",
			},
		}
		handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
	})
}

func TestTabPrints(t *testing.T) {
	completer := func(line string, pos int) (string, []string, string) {
		return "", []string{"abc", "abd", "abe"}, ""
	}
	ib := New(
		WithWordCompleter(completer),
		WithTabStyle(TabPrints),
	)
	ib.Resize(20, 3)

	sendKeys(t, ib, "ab")

	// First tab: show grid without changing text.
	sendKeys(t, ib, "<tab>")
	assert.Equal(t, "ab", ib.Text())
	got := drawHandler(ib, 20, 3)
	assert.Contains(t, got, "abc")
	assert.Contains(t, got, "abd")
	assert.Contains(t, got, "abe")

	// Second tab: apply first candidate.
	sendKeys(t, ib, "<tab>")
	assert.Equal(t, "abc", ib.Text())

	// Third+ tab: cycle through remaining candidates.
	sendKeys(t, ib, "<tab>")
	assert.Equal(t, "abd", ib.Text())

	sendKeys(t, ib, "<tab>")
	assert.Equal(t, "abe", ib.Text())
}

func TestWordCompleter(t *testing.T) {
	completer := func(line string, pos int) (string, []string, string) {
		return "cmd ", []string{"--help", "--version"}, ""
	}
	ib := New(WithWordCompleter(completer))
	cases := []handlertest.SequenceTestCase{
		{
			InputSequence: "cmd<space>--",
			Expected:      "cmd --▐             ",
		},
		{
			InputSequence: "<tab>",
			Expected:      "cmd --help▐         ",
		},
		{
			InputSequence: "<tab>",
			Expected:      "cmd --version▐      ",
		},
	}
	handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
}

func TestCtrlCAbort(t *testing.T) {
	ib := New(WithCtrlCAborts())
	ib.Resize(20, 1)
	sendKeys(t, ib, "hello")
	exit, handled := ib.Handle(term.Event{
		Type: term.EventKey,
		Ch:   'c',
		Mod:  term.ModCtrl,
	})
	assert.True(t, exit)
	assert.True(t, handled)
	_, err := ib.Result()
	assert.Equal(t, ErrAborted, err)
}

func TestCtrlCClear(t *testing.T) {
	ib := New()
	ib.Resize(20, 1)
	sendKeys(t, ib, "hello")
	exit, handled := ib.Handle(term.Event{
		Type: term.EventKey,
		Ch:   'c',
		Mod:  term.ModCtrl,
	})
	assert.False(t, exit)
	assert.True(t, handled)
	assert.Equal(t, "", ib.Text())
}

func TestCtrlDEOF(t *testing.T) {
	ib := New()
	ib.Resize(20, 1)
	exit, handled := ib.Handle(term.Event{
		Type: term.EventKey,
		Ch:   'd',
		Mod:  term.ModCtrl,
	})
	assert.True(t, exit)
	assert.True(t, handled)
	_, err := ib.Result()
	assert.Equal(t, io.EOF, err)
}

func TestCtrlDDeletesWhenNotEmpty(t *testing.T) {
	ib := New()
	cases := []handlertest.SequenceTestCase{
		{
			InputSequence: "abc",
			Expected:      "abc▐                ",
		},
		{
			InputSequence: "<home><c-d>",
			Expected:      "▐c                  ",
		},
	}
	handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
}

func TestResult(t *testing.T) {
	ib := New()
	ib.Resize(20, 1)
	sendKeys(t, ib, "hello")
	exit, handled := ib.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})
	assert.True(t, exit)
	assert.True(t, handled)
	text, err := ib.Result()
	require.NoError(t, err)
	assert.Equal(t, "hello", text)
}

func TestReset(t *testing.T) {
	ib := New(WithHistory([]string{"old"}))
	ib.Resize(20, 1)
	sendKeys(t, ib, "hello")
	ib.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})

	ib.Reset()
	assert.Equal(t, "", ib.Text())
	assert.False(t, ib.done)
	assert.False(t, ib.aborted)
	assert.False(t, ib.eof)
}

func TestCtrlL(t *testing.T) {
	ib := New()
	ib.Resize(20, 1)
	sendKeys(t, ib, "hello")
	_, handled := ib.Handle(term.Event{
		Type: term.EventKey,
		Ch:   'l',
		Mod:  term.ModCtrl,
	})
	assert.True(t, handled)
	assert.Equal(t, "hello", ib.Text())
}

func TestCtrlWDeleteWordBackward(t *testing.T) {
	ib := New()
	cases := []handlertest.SequenceTestCase{
		{
			InputSequence: "hello<space>world",
			Expected:      "hello world▐        ",
		},
		{
			InputSequence: "<c-w>",
			Expected:      "hello ▐             ",
		},
	}
	handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
}

func TestWithText(t *testing.T) {
	ib := New(WithText("prefilled"))
	cases := []handlertest.SequenceTestCase{
		{
			InputSequence: "",
			Expected:      "prefilled▐          ",
		},
	}
	handlertest.RunHandlerSequence(t, ib, 20, 1, cases)
}

func TestNonKeyEventIgnored(t *testing.T) {
	ib := New()
	ib.Resize(20, 1)
	exit, handled := ib.Handle(term.Event{
		Type: term.EventResize,
	})
	assert.False(t, exit)
	assert.False(t, handled)
}

func TestAppendHistory(t *testing.T) {
	ib := New()
	ib.Resize(20, 1)
	ib.AppendHistory("first")
	ib.AppendHistory("second")

	sendKeys(t, ib, "<up>")
	assert.Equal(t, "second", ib.Text())

	sendKeys(t, ib, "<up>")
	assert.Equal(t, "first", ib.Text())
}

func TestSetHistory(t *testing.T) {
	ib := New()
	ib.Resize(20, 1)
	ib.SetHistory([]string{"a", "b", "c"})

	sendKeys(t, ib, "<up>")
	assert.Equal(t, "c", ib.Text())
}

func TestClearHistory(t *testing.T) {
	ib := New(WithHistory([]string{"old"}))
	ib.Resize(20, 1)
	ib.ClearHistory()

	sendKeys(t, ib, "<up>")
	assert.Equal(t, "", ib.Text())
}

func TestMouseDoubleClickSelectsWord(t *testing.T) {
	t.Run("selects word under cursor", func(t *testing.T) {
		ib := New(WithText("hello world"))
		ib.Resize(20, 1)

		// Double-click on "world" (text position 6 = screen X 6).
		clickAt(ib, 6, 0)
		releaseAt(ib, 6, 0)
		clickAt(ib, 6, 0)

		sel, ok := ib.Selection()
		assert.True(t, ok)
		assert.Equal(t, "world", sel)
	})

	t.Run("selects first word", func(t *testing.T) {
		ib := New(WithText("hello world"))
		ib.Resize(20, 1)

		clickAt(ib, 2, 0)
		releaseAt(ib, 2, 0)
		clickAt(ib, 2, 0)

		sel, ok := ib.Selection()
		assert.True(t, ok)
		assert.Equal(t, "hello", sel)
	})

	t.Run("no selection on non-word chars", func(t *testing.T) {
		ib := New(WithText("hello world"))
		ib.Resize(20, 1)

		// Double-click on space (position 5).
		clickAt(ib, 5, 0)
		releaseAt(ib, 5, 0)
		clickAt(ib, 5, 0)

		_, ok := ib.Selection()
		assert.False(t, ok)
	})

	t.Run("selects URL segment", func(t *testing.T) {
		ib := New(WithText("http://127.0.0.1:8080/path"))
		ib.Resize(30, 1)

		// Double-click on '1' of "127" (text position 7).
		clickAt(ib, 7, 0)
		releaseAt(ib, 7, 0)
		clickAt(ib, 7, 0)

		sel, ok := ib.Selection()
		assert.True(t, ok)
		assert.Equal(t, "127", sel)
	})
}

func TestMouseTripleClickSelectsLine(t *testing.T) {
	ib := New(WithText("hello world"))
	ib.Resize(20, 1)

	clickAt(ib, 2, 0)
	releaseAt(ib, 2, 0)
	clickAt(ib, 2, 0)
	releaseAt(ib, 2, 0)
	clickAt(ib, 2, 0)

	sel, ok := ib.Selection()
	assert.True(t, ok)
	assert.Equal(t, "hello world", sel)
}

func TestMouseSingleClickPositionsCursor(t *testing.T) {
	ib := New(WithText("hello world"))
	ib.Resize(20, 1)

	clickAt(ib, 3, 0)
	releaseAt(ib, 3, 0)

	pos, _, show := ib.Cursor()
	assert.True(t, show)
	assert.Equal(t, 3, pos.X)
	assert.Equal(t, 0, pos.Y)
}

func TestMouseDragCreatesSelection(t *testing.T) {
	ib := New(WithText("hello world"))
	ib.Resize(20, 1)

	// Click at position 2, drag to position 7.
	clickAt(ib, 2, 0)
	// Drag (another MouseLeft while pressed).
	ib.Handle(term.Event{
		Type: term.EventMouse, Key: term.MouseLeft,
		MouseX: 7, MouseY: 0,
	})
	releaseAt(ib, 7, 0)

	sel, ok := ib.Selection()
	assert.True(t, ok)
	assert.Equal(t, "llo w", sel)
}

func TestMouseClickWithPrompt(t *testing.T) {
	ib := New(WithPrompt("> "), WithText("hello"))
	ib.Resize(20, 1)

	// Click on 'e' which is at screen X = 2 (promptWidth) + 1 = 3.
	clickAt(ib, 3, 0)
	releaseAt(ib, 3, 0)

	pos, _, show := ib.Cursor()
	assert.True(t, show)
	assert.Equal(t, 3, pos.X)
	assert.Equal(t, 0, pos.Y)
	// Text position should be 1 ('e' in "hello").
	assert.Equal(t, 1, ib.cursor)
}

func TestMouseDoubleClickWithPrompt(t *testing.T) {
	ib := New(WithPrompt("> "), WithText("hello world"))
	ib.Resize(20, 1)

	// Double-click on 'w' at screen X = 2 + 6 = 8.
	clickAt(ib, 8, 0)
	releaseAt(ib, 8, 0)
	clickAt(ib, 8, 0)

	sel, ok := ib.Selection()
	assert.True(t, ok)
	assert.Equal(t, "world", sel)
}

// clickAt sends a left mouse button press at the given screen coords.
func clickAt(ib *Handler, x, y int) {
	ib.Handle(term.Event{
		Type: term.EventMouse, Key: term.MouseLeft,
		MouseX: x, MouseY: y,
	})
}

// releaseAt sends a mouse release at the given screen coords.
func releaseAt(ib *Handler, x, y int) {
	ib.Handle(term.Event{
		Type: term.EventMouse, Key: term.MouseRelease,
		MouseX: x, MouseY: y,
	})
}
