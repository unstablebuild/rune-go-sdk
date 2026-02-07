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
