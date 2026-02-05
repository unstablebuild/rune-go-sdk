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
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestInputBox_Basic(t *testing.T) {
	ib := NewInputBox()
	assert.NotNil(t, ib)
	assert.Equal(t, "", ib.Text())
	assert.Equal(t, 0, ib.cursor)
}

func TestInputBox_CharacterInput(t *testing.T) {
	ib := NewInputBox()

	// Type "hello"
	for _, ch := range "hello" {
		exit, handled := ib.Handle(term.Event{
			Type: term.EventKey,
			Ch:   ch,
		})
		assert.False(t, exit)
		assert.True(t, handled)
	}

	assert.Equal(t, "hello", ib.Text())
	assert.Equal(t, 5, ib.cursor)
}

func TestInputBox_Backspace(t *testing.T) {
	ib := NewInputBox()
	ib.SetText("hello")

	// Backspace once
	exit, handled := ib.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyBackspace,
	})
	assert.False(t, exit)
	assert.True(t, handled)
	assert.Equal(t, "hell", ib.Text())
	assert.Equal(t, 4, ib.cursor)
}

func TestInputBox_Delete(t *testing.T) {
	ib := NewInputBox()
	ib.SetText("hello")
	ib.cursor = 0 // move to start

	// Delete once
	exit, handled := ib.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyDelete,
	})
	assert.False(t, exit)
	assert.True(t, handled)
	assert.Equal(t, "ello", ib.Text())
	assert.Equal(t, 0, ib.cursor)
}

func TestInputBox_ArrowNavigation(t *testing.T) {
	ib := NewInputBox()
	ib.SetText("hello")

	// Arrow left twice
	ib.Handle(term.Event{Type: term.EventKey, Key: term.KeyArrowLeft})
	ib.Handle(term.Event{Type: term.EventKey, Key: term.KeyArrowLeft})
	assert.Equal(t, 3, ib.cursor)

	// Arrow right once
	ib.Handle(term.Event{Type: term.EventKey, Key: term.KeyArrowRight})
	assert.Equal(t, 4, ib.cursor)
}

func TestInputBox_HomeEnd(t *testing.T) {
	ib := NewInputBox()
	ib.SetText("hello")

	// Home
	ib.Handle(term.Event{Type: term.EventKey, Key: term.KeyHome})
	assert.Equal(t, 0, ib.cursor)

	// End
	ib.Handle(term.Event{Type: term.EventKey, Key: term.KeyEnd})
	assert.Equal(t, 5, ib.cursor)
}

func TestInputBox_CtrlA_SelectAll(t *testing.T) {
	ib := NewInputBox()
	ib.SetText("hello")

	// Ctrl+A
	exit, handled := ib.Handle(term.Event{
		Type: term.EventKey,
		Ch:   'a',
		Mod:  term.ModCtrl,
	})
	assert.False(t, exit)
	assert.True(t, handled)

	// Check selection
	text, selected := ib.Selection()
	assert.True(t, selected)
	assert.Equal(t, "hello", text)
}

func TestInputBox_CtrlU_ClearLine(t *testing.T) {
	ib := NewInputBox()
	ib.SetText("hello")

	// Ctrl+U
	exit, handled := ib.Handle(term.Event{
		Type: term.EventKey,
		Ch:   'u',
		Mod:  term.ModCtrl,
	})
	assert.False(t, exit)
	assert.True(t, handled)
	assert.Equal(t, "", ib.Text())
	assert.Equal(t, 0, ib.cursor)
}

func TestInputBox_Enter(t *testing.T) {
	ib := NewInputBox()
	ib.SetText("hello")

	// Enter should signal exit
	exit, handled := ib.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})
	assert.True(t, exit)
	assert.True(t, handled)
}

func TestInputBox_InsertAtCursor(t *testing.T) {
	ib := NewInputBox()
	ib.SetText("helo")
	ib.cursor = 2 // position between 'e' and 'l'

	// Insert 'l'
	ib.Handle(term.Event{
		Type: term.EventKey,
		Ch:   'l',
	})
	assert.Equal(t, "hello", ib.Text())
	assert.Equal(t, 3, ib.cursor)
}

func TestInputBox_CursorPosition(t *testing.T) {
	ib := NewInputBox()
	ib.SetText("hello")
	ib.Resize(80, 1)

	coords, style, show := ib.Cursor()
	assert.True(t, show)
	assert.Equal(t, term.CursorStyleSteadyBlock, style)
	assert.Equal(t, 5, coords.X)
	assert.Equal(t, 0, coords.Y)
}
