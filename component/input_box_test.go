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

package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unstablebuild/rune-go-sdk/handler/handlertest"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestInputBox_BasicInput(t *testing.T) {
	ib := NewInputBox()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "", Expected: "▐         "},
		{InputSequence: "h", Expected: "h▐        "},
		{InputSequence: "e", Expected: "he▐       "},
		{InputSequence: "l", Expected: "hel▐      "},
		{InputSequence: "l", Expected: "hell▐     "},
		{InputSequence: "o", Expected: "hello▐    "},
		{InputSequence: "<space>", Expected: "hello ▐   "},
	}
	handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
}

func TestInputBox_Backspace(t *testing.T) {
	ib := NewInputBox()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "hello", Expected: "hello▐    "},
		{
			InputSequence: "<backspace>",
			Expected:      "hell▐     ",
		},
		{
			InputSequence: "<backspace>",
			Expected:      "hel▐      ",
		},
	}
	handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
}

func TestInputBox_Delete(t *testing.T) {
	ib := NewInputBox()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "hello", Expected: "hello▐    "},
		{InputSequence: "<home>", Expected: "▐ello     "},
		{InputSequence: "<delete>", Expected: "▐llo      "},
		{InputSequence: "<delete>", Expected: "▐lo       "},
	}
	handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
}

func TestInputBox_ArrowNavigation(t *testing.T) {
	ib := NewInputBox()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "hello", Expected: "hello▐    "},
		{InputSequence: "<left>", Expected: "hell▐     "},
		{InputSequence: "<left>", Expected: "hel▐o     "},
		{InputSequence: "<right>", Expected: "hell▐     "},
		{InputSequence: "<right>", Expected: "hello▐    "},
	}
	handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
}

func TestInputBox_HomeEnd(t *testing.T) {
	ib := NewInputBox()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "hello", Expected: "hello▐    "},
		{InputSequence: "<home>", Expected: "▐ello     "},
		{InputSequence: "<end>", Expected: "hello▐    "},
	}
	handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
}

func TestInputBox_EmacsNavigation(t *testing.T) {
	ib := NewInputBox()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "hello", Expected: "hello▐    "},
		// Ctrl+A: beginning of line
		{InputSequence: "<c-a>", Expected: "▐ello     "},
		// Ctrl+E: end of line
		{InputSequence: "<c-e>", Expected: "hello▐    "},
		// Ctrl+B: backward char
		{InputSequence: "<c-b>", Expected: "hell▐     "},
		// Ctrl+B again
		{InputSequence: "<c-b>", Expected: "hel▐o     "},
		// Ctrl+F: forward char
		{InputSequence: "<c-f>", Expected: "hell▐     "},
	}
	handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
}

func TestInputBox_EmacsEditing(t *testing.T) {
	ib := NewInputBox()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "hello", Expected: "hello▐    "},
		{InputSequence: "<c-a>", Expected: "▐ello     "},
		// Ctrl+D: delete char at cursor
		{InputSequence: "<c-d>", Expected: "▐llo      "},
		// Ctrl+E then Ctrl+H: backspace
		{InputSequence: "<c-e><c-h>", Expected: "ell▐      "},
		// Ctrl+K: kill to end of line
		{InputSequence: "<c-a><c-k>", Expected: "▐         "},
	}
	handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
}

func TestInputBox_CtrlU_KillLine(t *testing.T) {
	ib := NewInputBox()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "hello", Expected: "hello▐    "},
		// Ctrl+U: kill entire line
		{InputSequence: "<c-u>", Expected: "▐         "},
	}
	handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
}

func TestInputBox_InsertAtCursor(t *testing.T) {
	ib := NewInputBox()
	cases := []handlertest.SequenceTestCase{
		{InputSequence: "helo", Expected: "helo▐     "},
		// Move cursor between e and l
		{InputSequence: "<left><left>", Expected: "he▐o      "},
		// Insert 'l'
		{InputSequence: "l", Expected: "hel▐o     "},
	}
	handlertest.RunHandlerSequence(t, ib, 10, 1, cases)
}

func TestInputBox_API(t *testing.T) {
	ib := NewInputBox()

	// Test Text() and SetText()
	assert.Equal(t, "", ib.Text())
	ib.SetText("hello")
	assert.Equal(t, "hello", ib.Text())
	assert.Equal(t, 5, ib.cursor)

	// Test Clear()
	ib.Clear()
	assert.Equal(t, "", ib.Text())
	assert.Equal(t, 0, ib.cursor)
}

func TestInputBox_Cursor(t *testing.T) {
	ib := NewInputBox()
	ib.SetText("hello")
	ib.Resize(80, 1)

	coords, style, show := ib.Cursor()
	assert.True(t, show)
	assert.Equal(t, term.CursorStyleSteadyBlock, style)
	assert.Equal(t, 5, coords.X)
	assert.Equal(t, 0, coords.Y)
}

func TestInputBox_EnterReturnsTrue(t *testing.T) {
	ib := NewInputBox()
	ib.SetText("hello")

	exit, handled := ib.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})
	assert.True(t, exit, "Enter should signal exit")
	assert.True(t, handled)
}
