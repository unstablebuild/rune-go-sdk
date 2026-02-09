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

package term

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/tcell/v3/termbox"
)

func TestTermboxEventConvert(t *testing.T) {
	t.Run("fully reversible key combinations", func(t *testing.T) {
		suite := []struct {
			ev  keyComb
			tev termbox.Event
		}{
			{keyComb{Ch: 'a'}, termbox.Event{Ch: 'a'}},
			{keyComb{Ch: 'a', Mod: ModAlt}, termbox.Event{Mod: termbox.ModAlt, Ch: 'a'}},
			{keyComb{Ch: 'A'}, termbox.Event{Ch: 'A'}},
			{keyComb{Ch: '1'}, termbox.Event{Ch: '1'}},
			{keyComb{Ch: '!'}, termbox.Event{Ch: '!'}},
			{keyComb{Ch: '_'}, termbox.Event{Ch: '_'}},
			{keyComb{Ch: '-'}, termbox.Event{Ch: '-'}},
			{keyComb{Ch: '>'}, termbox.Event{Ch: '>'}},
			{keyComb{Ch: '+'}, termbox.Event{Ch: '+'}},
			{keyComb{Ch: '.'}, termbox.Event{Ch: '.'}},
			{keyComb{Ch: '`'}, termbox.Event{Ch: '`'}},
			{keyComb{Ch: '\t'}, termbox.Event{Ch: '\t'}},
			{keyComb{Key: KeyF1}, termbox.Event{Key: termbox.KeyF1}},
			{keyComb{Key: KeyF2}, termbox.Event{Key: termbox.KeyF2}},
			{keyComb{Key: KeyF3}, termbox.Event{Key: termbox.KeyF3}},
			{keyComb{Key: KeyF4}, termbox.Event{Key: termbox.KeyF4}},
			{keyComb{Key: KeyF5}, termbox.Event{Key: termbox.KeyF5}},
			{keyComb{Key: KeyF6}, termbox.Event{Key: termbox.KeyF6}},
			{keyComb{Key: KeyF7}, termbox.Event{Key: termbox.KeyF7}},
			{keyComb{Key: KeyF8}, termbox.Event{Key: termbox.KeyF8}},
			{keyComb{Key: KeyF9}, termbox.Event{Key: termbox.KeyF9}},
			{keyComb{Key: KeyF10}, termbox.Event{Key: termbox.KeyF10}},
			{keyComb{Key: KeyF11}, termbox.Event{Key: termbox.KeyF11}},
			{keyComb{Key: KeyF12}, termbox.Event{Key: termbox.KeyF12}},
			{keyComb{Key: KeyInsert}, termbox.Event{Key: termbox.KeyInsert}},
			{keyComb{Key: KeyDelete}, termbox.Event{Key: termbox.KeyDelete}},
			{keyComb{Key: KeyHome}, termbox.Event{Key: termbox.KeyHome}},
			{keyComb{Key: KeyEnd}, termbox.Event{Key: termbox.KeyEnd}},
			{keyComb{Key: KeyArrowUp}, termbox.Event{Key: termbox.KeyArrowUp}},
			{keyComb{Key: KeyArrowDown}, termbox.Event{Key: termbox.KeyArrowDown}},
			{keyComb{Key: KeyArrowRight}, termbox.Event{Key: termbox.KeyArrowRight}},
			{keyComb{Key: KeyArrowLeft}, termbox.Event{Key: termbox.KeyArrowLeft}},
			{keyComb{Mod: ModCtrl, Ch: 'a'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlA}},
			{keyComb{Mod: ModCtrlAlt, Ch: 'a'}, termbox.Event{Mod: termbox.ModAlt | termbox.ModCtrl, Key: termbox.KeyCtrlA}},
			{keyComb{Mod: ModCtrl, Ch: 'b'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlB}},
			{keyComb{Mod: ModCtrl, Ch: 'c'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlC}},
			{keyComb{Mod: ModCtrl, Ch: 'd'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlD}},
			{keyComb{Mod: ModCtrl, Ch: 'e'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlE}},
			{keyComb{Mod: ModCtrl, Ch: 'f'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlF}},
			{keyComb{Mod: ModCtrl, Ch: 'g'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlG}},
			{keyComb{Mod: ModCtrl, Ch: 'j'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlJ}},
			{keyComb{Mod: ModCtrl, Ch: 'k'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlK}},
			{keyComb{Mod: ModCtrl, Ch: 'l'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlL}},
			{keyComb{Mod: ModCtrl, Ch: 'n'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlN}},
			{keyComb{Mod: ModCtrl, Ch: 'o'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlO}},
			{keyComb{Mod: ModCtrl, Ch: 'p'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlP}},
			{keyComb{Mod: ModCtrl, Ch: 'q'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlQ}},
			{keyComb{Mod: ModCtrl, Ch: 'r'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlR}},
			{keyComb{Mod: ModCtrl, Ch: 's'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlS}},
			{keyComb{Mod: ModCtrl, Ch: 't'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlT}},
			{keyComb{Mod: ModCtrl, Ch: 'u'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlU}},
			{keyComb{Mod: ModCtrl, Ch: 'v'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlV}},
			{keyComb{Mod: ModCtrl, Ch: 'w'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlW}},
			{keyComb{Mod: ModCtrl, Ch: 'x'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlX}},
			{keyComb{Mod: ModCtrl, Ch: 'y'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlY}},
			{keyComb{Mod: ModCtrl, Ch: 'z'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlZ}},
			{keyComb{Key: KeyBackspace}, termbox.Event{Key: termbox.KeyBackspace2}},
			{keyComb{Key: KeyTab}, termbox.Event{Key: termbox.KeyTab}},
			{keyComb{Key: KeyEnter}, termbox.Event{Key: termbox.KeyEnter}},
			{keyComb{Key: KeyEsc}, termbox.Event{Key: termbox.KeyEsc}},
			{keyComb{Key: KeyPgdn}, termbox.Event{Key: termbox.KeyPgdn}},
			{keyComb{Key: KeyPgup}, termbox.Event{Key: termbox.KeyPgup}},
			{keyComb{Key: KeySpace}, termbox.Event{Key: termbox.KeySpace}},
			{keyComb{Mod: ModAlt, Key: KeySpace}, termbox.Event{Mod: termbox.ModAlt, Key: termbox.KeySpace}},
			{keyComb{Key: KeySpace, Mod: ModCtrl}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlSpace}},
			{keyComb{Key: MouseLeft}, termbox.Event{Key: termbox.MouseLeft}},
			{keyComb{Key: MouseRight}, termbox.Event{Key: termbox.MouseRight}},
			{keyComb{Key: MouseMiddle}, termbox.Event{Key: termbox.MouseMiddle}},
			{keyComb{Key: MouseRelease}, termbox.Event{Key: termbox.MouseRelease}},
			{keyComb{Key: MouseWheelUp}, termbox.Event{Key: termbox.MouseWheelUp}},
			{keyComb{Key: MouseWheelDown}, termbox.Event{Key: termbox.MouseWheelDown}},

			// ambiguous keys
			{keyComb{Ch: 'h', Mod: ModCtrl}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlH}},
			{keyComb{Mod: ModCtrl, Key: KeySpace}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrl2}},
			{keyComb{Key: KeyEsc}, termbox.Event{Key: termbox.KeyCtrl3}},
			{keyComb{Ch: '6', Mod: ModCtrl}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrl6}},
			{keyComb{Mod: ModCtrl, Ch: '/'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrl7}},
			{keyComb{Ch: '/', Mod: ModCtrl}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlSlash}},
			{keyComb{Ch: '\\', Mod: ModCtrl}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlBackslash}},
			{keyComb{Key: KeyEsc}, termbox.Event{Key: termbox.KeyCtrlLsqBracket}},
			{keyComb{Ch: '\\', Mod: ModCtrl}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrl4}},
			{keyComb{Ch: ']', Mod: ModCtrl}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrl5}},
			{keyComb{Mod: ModCtrl, Ch: '/'}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyCtrlUnderscore}},
			{keyComb{Key: KeyBackspace}, termbox.Event{Key: termbox.KeyCtrl8}},

			// shift-arrow combinations
			{keyComb{Mod: ModShift, Key: KeyArrowUp}, termbox.Event{Mod: termbox.ModShift, Key: termbox.KeyArrowUp}},
			{keyComb{Mod: ModShift, Key: KeyArrowDown}, termbox.Event{Mod: termbox.ModShift, Key: termbox.KeyArrowDown}},
			{keyComb{Mod: ModShift, Key: KeyArrowLeft}, termbox.Event{Mod: termbox.ModShift, Key: termbox.KeyArrowLeft}},
			{keyComb{Mod: ModShift, Key: KeyArrowRight}, termbox.Event{Mod: termbox.ModShift, Key: termbox.KeyArrowRight}},
		}

		for i, test := range suite {
			t.Run(fmt.Sprintf("event to termbox event first: %d", i), func(t *testing.T) {
				ev := Event{Type: EventKey, Ch: test.ev.Ch, Key: test.ev.Key, Mod: test.ev.Mod}
				actualTev := eventToTermboxEvent(ev)
				test.tev.Type = termbox.EventKey // make test cases easier to spell out
				require.Equal(t, test.tev, actualTev)

				actualEv := termboxEventToEvent(actualTev)
				assert.Equal(t, ev, actualEv)
			})

			t.Run(fmt.Sprintf("termbox event to event: %d", i), func(t *testing.T) {
				test.tev.Type = termbox.EventKey
				actualEv := termboxEventToEvent(test.tev)
				ev := Event{Type: EventKey, Ch: test.ev.Ch, Key: test.ev.Key, Mod: test.ev.Mod}
				assert.Equal(t, ev, actualEv)

				actualTev := eventToTermboxEvent(actualEv)
				require.Equal(t, test.tev, actualTev)
			})
		}
	})

	t.Run("not fully reversible, but compatible combinations", func(t *testing.T) {
		suite := []struct {
			ev  keyComb
			tev termbox.Event
		}{
			{keyComb{Ch: '~'}, termbox.Event{Key: termbox.KeyTilde}},
			{keyComb{Key: KeySpace}, termbox.Event{Ch: ' ', Key: termbox.KeySpace}},
			{keyComb{Key: KeySpace, Mod: ModCtrl}, termbox.Event{Ch: ' ', Key: termbox.KeyCtrlSpace}},
			// unfortunately Key=0 is equivalent to KeyCtrlSpace, so this is an ambiguous case
			{keyComb{Key: KeySpace, Mod: ModCtrl}, termbox.Event{Ch: ' '}},
			{keyComb{Key: KeySpace, Mod: ModCtrlAlt}, termbox.Event{Ch: ' ', Mod: termbox.ModAlt}},
		}
		for i, test := range suite {
			t.Run(fmt.Sprintf("termbox event to event: %d", i), func(t *testing.T) {
				test.tev.Type = termbox.EventKey
				actualEv := termboxEventToEvent(test.tev)
				ev := Event{Type: EventKey, Ch: test.ev.Ch, Key: test.ev.Key, Mod: test.ev.Mod}
				assert.Equal(t, ev, actualEv)
			})
		}
	})
}

type keyComb struct {
	Ch  rune
	Key Key
	Mod Modifier
}

func TestModConversion(t *testing.T) {
	t.Run("individual modifiers", func(t *testing.T) {
		suite := []struct {
			name string
			term Modifier
			tb   termbox.Modifier
		}{
			{"zero", 0, 0},
			{"Alt", ModAlt, termbox.ModAlt},
			{"Shift", ModShift, termbox.ModShift},
			{"Meta", ModMeta, termbox.ModMeta},
			{"Ctrl", ModCtrl, termbox.ModCtrl},
		}
		for _, test := range suite {
			t.Run(test.name, func(t *testing.T) {
				assert.Equal(t, test.tb, modToTermboxMod(test.term))
				assert.Equal(t, test.term, termboxModToMod(test.tb))
			})
		}
	})

	t.Run("combined modifiers", func(t *testing.T) {
		suite := []struct {
			name string
			term Modifier
			tb   termbox.Modifier
		}{
			{"CtrlAlt", ModCtrlAlt, termbox.ModCtrl | termbox.ModAlt},
			{"CtrlShift", ModCtrlShift, termbox.ModCtrl | termbox.ModShift},
			{"AltShift", ModAltShift, termbox.ModAlt | termbox.ModShift},
			{"ShiftMeta", ModShiftMeta, termbox.ModShift | termbox.ModMeta},
			{"CtrlMeta", ModCtrlMeta, termbox.ModCtrl | termbox.ModMeta},
			{"AltMeta", ModAltMeta, termbox.ModAlt | termbox.ModMeta},
			{"CtrlShiftAlt", ModCtrlShiftAlt, termbox.ModCtrl | termbox.ModShift | termbox.ModAlt},
			{"CtrlShiftMeta", ModCtrlShiftMeta, termbox.ModCtrl | termbox.ModShift | termbox.ModMeta},
			{"CtrlAltMeta", ModCtrlAltMeta, termbox.ModCtrl | termbox.ModAlt | termbox.ModMeta},
			{"AltShiftMeta", ModAltShiftMeta, termbox.ModAlt | termbox.ModShift | termbox.ModMeta},
			{"all four", ModCtrl | ModAlt | ModShift | ModMeta, termbox.ModCtrl | termbox.ModAlt | termbox.ModShift | termbox.ModMeta},
		}
		for _, test := range suite {
			t.Run(test.name, func(t *testing.T) {
				assert.Equal(t, test.tb, modToTermboxMod(test.term))
				assert.Equal(t, test.term, termboxModToMod(test.tb))
			})
		}
	})

	t.Run("exhaustive round-trip", func(t *testing.T) {
		for i := Modifier(0); i < 16; i++ {
			got := termboxModToMod(modToTermboxMod(i))
			assert.Equal(t, i, got, "term→termbox→term failed for %d", i)
		}
		for i := termbox.Modifier(0); i < 16; i++ {
			got := modToTermboxMod(termboxModToMod(i))
			assert.Equal(t, i, got, "termbox→term→termbox failed for %d", i)
		}
	})
}

func TestMultiModifierEvents(t *testing.T) {
	suite := []struct {
		name string
		ev   keyComb
		tev  termbox.Event
	}{
		// ctrl+arrow
		{"Ctrl+Up", keyComb{Mod: ModCtrl, Key: KeyArrowUp}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyArrowUp}},
		{"Ctrl+Down", keyComb{Mod: ModCtrl, Key: KeyArrowDown}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyArrowDown}},
		{"Ctrl+Left", keyComb{Mod: ModCtrl, Key: KeyArrowLeft}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyArrowLeft}},
		{"Ctrl+Right", keyComb{Mod: ModCtrl, Key: KeyArrowRight}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyArrowRight}},

		// alt+arrow
		{"Alt+Up", keyComb{Mod: ModAlt, Key: KeyArrowUp}, termbox.Event{Mod: termbox.ModAlt, Key: termbox.KeyArrowUp}},
		{"Alt+Down", keyComb{Mod: ModAlt, Key: KeyArrowDown}, termbox.Event{Mod: termbox.ModAlt, Key: termbox.KeyArrowDown}},
		{"Alt+Left", keyComb{Mod: ModAlt, Key: KeyArrowLeft}, termbox.Event{Mod: termbox.ModAlt, Key: termbox.KeyArrowLeft}},
		{"Alt+Right", keyComb{Mod: ModAlt, Key: KeyArrowRight}, termbox.Event{Mod: termbox.ModAlt, Key: termbox.KeyArrowRight}},

		// ctrl+shift+arrow
		{"Ctrl+Shift+Up", keyComb{Mod: ModCtrlShift, Key: KeyArrowUp}, termbox.Event{Mod: termbox.ModCtrl | termbox.ModShift, Key: termbox.KeyArrowUp}},
		{"Ctrl+Shift+Down", keyComb{Mod: ModCtrlShift, Key: KeyArrowDown}, termbox.Event{Mod: termbox.ModCtrl | termbox.ModShift, Key: termbox.KeyArrowDown}},
		{"Ctrl+Shift+Left", keyComb{Mod: ModCtrlShift, Key: KeyArrowLeft}, termbox.Event{Mod: termbox.ModCtrl | termbox.ModShift, Key: termbox.KeyArrowLeft}},
		{"Ctrl+Shift+Right", keyComb{Mod: ModCtrlShift, Key: KeyArrowRight}, termbox.Event{Mod: termbox.ModCtrl | termbox.ModShift, Key: termbox.KeyArrowRight}},

		// alt+shift+arrow
		{"Alt+Shift+Up", keyComb{Mod: ModAltShift, Key: KeyArrowUp}, termbox.Event{Mod: termbox.ModAlt | termbox.ModShift, Key: termbox.KeyArrowUp}},

		// ctrl+alt+shift+arrow
		{"Ctrl+Alt+Shift+Up", keyComb{Mod: ModCtrlShiftAlt, Key: KeyArrowUp}, termbox.Event{Mod: termbox.ModCtrl | termbox.ModAlt | termbox.ModShift, Key: termbox.KeyArrowUp}},

		// shift+function keys
		{"Shift+F1", keyComb{Mod: ModShift, Key: KeyF1}, termbox.Event{Mod: termbox.ModShift, Key: termbox.KeyF1}},
		{"Shift+F5", keyComb{Mod: ModShift, Key: KeyF5}, termbox.Event{Mod: termbox.ModShift, Key: termbox.KeyF5}},
		{"Shift+F12", keyComb{Mod: ModShift, Key: KeyF12}, termbox.Event{Mod: termbox.ModShift, Key: termbox.KeyF12}},
		{"Ctrl+F1", keyComb{Mod: ModCtrl, Key: KeyF1}, termbox.Event{Mod: termbox.ModCtrl, Key: termbox.KeyF1}},

		// shift+navigation
		{"Shift+Home", keyComb{Mod: ModShift, Key: KeyHome}, termbox.Event{Mod: termbox.ModShift, Key: termbox.KeyHome}},
		{"Shift+End", keyComb{Mod: ModShift, Key: KeyEnd}, termbox.Event{Mod: termbox.ModShift, Key: termbox.KeyEnd}},
		{"Shift+PgUp", keyComb{Mod: ModShift, Key: KeyPgup}, termbox.Event{Mod: termbox.ModShift, Key: termbox.KeyPgup}},
		{"Shift+PgDn", keyComb{Mod: ModShift, Key: KeyPgdn}, termbox.Event{Mod: termbox.ModShift, Key: termbox.KeyPgdn}},
		{"Shift+Insert", keyComb{Mod: ModShift, Key: KeyInsert}, termbox.Event{Mod: termbox.ModShift, Key: termbox.KeyInsert}},
		{"Shift+Delete", keyComb{Mod: ModShift, Key: KeyDelete}, termbox.Event{Mod: termbox.ModShift, Key: termbox.KeyDelete}},

		// alt+special keys
		{"Alt+Enter", keyComb{Mod: ModAlt, Key: KeyEnter}, termbox.Event{Mod: termbox.ModAlt, Key: termbox.KeyEnter}},
		{"Alt+Esc", keyComb{Mod: ModAlt, Key: KeyEsc}, termbox.Event{Mod: termbox.ModAlt, Key: termbox.KeyEsc}},
		{"Alt+Backspace", keyComb{Mod: ModAlt, Key: KeyBackspace}, termbox.Event{Mod: termbox.ModAlt, Key: termbox.KeyBackspace2}},
		{"Alt+Tab", keyComb{Mod: ModAlt, Key: KeyTab}, termbox.Event{Mod: termbox.ModAlt, Key: termbox.KeyTab}},

		// ctrl+shift+char (shift carries through)
		{"Ctrl+Shift+a", keyComb{Mod: ModCtrlShift, Ch: 'a'}, termbox.Event{Mod: termbox.ModCtrl | termbox.ModShift, Key: termbox.KeyCtrlA}},
		{"Ctrl+Shift+z", keyComb{Mod: ModCtrlShift, Ch: 'z'}, termbox.Event{Mod: termbox.ModCtrl | termbox.ModShift, Key: termbox.KeyCtrlZ}},

		// ctrl+alt+shift+char
		{"Ctrl+Alt+Shift+a", keyComb{Mod: ModCtrlShiftAlt, Ch: 'a'}, termbox.Event{Mod: termbox.ModCtrl | termbox.ModAlt | termbox.ModShift, Key: termbox.KeyCtrlA}},

		// meta combinations
		{"Meta+Up", keyComb{Mod: ModMeta, Key: KeyArrowUp}, termbox.Event{Mod: termbox.ModMeta, Key: termbox.KeyArrowUp}},
		{"Ctrl+Meta+a", keyComb{Mod: ModCtrlMeta, Ch: 'a'}, termbox.Event{Mod: termbox.ModCtrl | termbox.ModMeta, Key: termbox.KeyCtrlA}},

		// shift+char (no ctrl, shift carries through modifier)
		{"Shift+a", keyComb{Mod: ModShift, Ch: 'a'}, termbox.Event{Mod: termbox.ModShift, Ch: 'a'}},
		{"Alt+Shift+a", keyComb{Mod: ModAltShift, Ch: 'a'}, termbox.Event{Mod: termbox.ModAlt | termbox.ModShift, Ch: 'a'}},
	}

	for _, test := range suite {
		t.Run(test.name+"/forward then reverse", func(t *testing.T) {
			ev := Event{Type: EventKey, Ch: test.ev.Ch, Key: test.ev.Key, Mod: test.ev.Mod}
			actualTev := eventToTermboxEvent(ev)
			test.tev.Type = termbox.EventKey
			require.Equal(t, test.tev, actualTev)

			actualEv := termboxEventToEvent(actualTev)
			assert.Equal(t, ev, actualEv)
		})

		t.Run(test.name+"/reverse then forward", func(t *testing.T) {
			test.tev.Type = termbox.EventKey
			actualEv := termboxEventToEvent(test.tev)
			ev := Event{Type: EventKey, Ch: test.ev.Ch, Key: test.ev.Key, Mod: test.ev.Mod}
			assert.Equal(t, ev, actualEv)

			actualTev := eventToTermboxEvent(actualEv)
			require.Equal(t, test.tev, actualTev)
		})
	}
}

func TestEventFieldPassthrough(t *testing.T) {
	t.Run("term to termbox", func(t *testing.T) {
		called := false
		fn := func() { called = true }
		ev := Event{
			Type:   EventKey,
			Ch:     'x',
			Width:  80,
			Height: 24,
			MouseX: 10,
			MouseY: 5,
			Raw:    []byte{0x78},
			Err:    fmt.Errorf("test error"),
			UserFunc: fn,
		}
		tev := eventToTermboxEvent(ev)

		assert.Equal(t, termbox.EventKey, tev.Type)
		assert.Equal(t, 80, tev.Width)
		assert.Equal(t, 24, tev.Height)
		assert.Equal(t, 10, tev.MouseX)
		assert.Equal(t, 5, tev.MouseY)
		assert.Equal(t, []byte{0x78}, tev.Raw)
		assert.EqualError(t, tev.Err, "test error")
		require.NotNil(t, tev.Metadata)
		tev.Metadata.(func())()
		assert.True(t, called)
	})

	t.Run("termbox to term", func(t *testing.T) {
		called := false
		fn := func() { called = true }
		tev := termbox.Event{
			Type:     termbox.EventKey,
			Ch:       'x',
			Width:    120,
			Height:   40,
			MouseX:   15,
			MouseY:   8,
			Raw:      []byte{0x78},
			Err:      fmt.Errorf("test error"),
			Metadata: fn,
		}
		ev := termboxEventToEvent(tev)

		assert.Equal(t, EventKey, ev.Type)
		assert.Equal(t, 120, ev.Width)
		assert.Equal(t, 40, ev.Height)
		assert.Equal(t, 15, ev.MouseX)
		assert.Equal(t, 8, ev.MouseY)
		assert.Equal(t, []byte{0x78}, ev.Raw)
		assert.EqualError(t, ev.Err, "test error")
		require.NotNil(t, ev.UserFunc)
		ev.UserFunc()
		assert.True(t, called)
	})

	t.Run("nil metadata does not set UserFunc", func(t *testing.T) {
		tev := termbox.Event{Type: termbox.EventKey, Key: termbox.KeyF1}
		ev := termboxEventToEvent(tev)
		assert.Nil(t, ev.UserFunc)
	})

	t.Run("nil UserFunc does not set metadata", func(t *testing.T) {
		ev := Event{Type: EventKey, Key: KeyF1}
		tev := eventToTermboxEvent(ev)
		assert.Nil(t, tev.Metadata)
	})
}

func TestNonKeyEventPassthrough(t *testing.T) {
	t.Run("resize event round-trips type and dimensions", func(t *testing.T) {
		ev := Event{Type: EventResize, Width: 132, Height: 43}
		tev := eventToTermboxEvent(ev)
		assert.Equal(t, termbox.EventType(EventResize), tev.Type)
		assert.Equal(t, 132, tev.Width)
		assert.Equal(t, 43, tev.Height)

		back := termboxEventToEvent(tev)
		assert.Equal(t, EventResize, back.Type)
		assert.Equal(t, 132, back.Width)
		assert.Equal(t, 43, back.Height)
	})

	t.Run("error event round-trips", func(t *testing.T) {
		ev := Event{Type: EventError, Err: fmt.Errorf("io timeout")}
		tev := eventToTermboxEvent(ev)
		assert.Equal(t, termbox.EventType(EventError), tev.Type)
		assert.EqualError(t, tev.Err, "io timeout")

		back := termboxEventToEvent(tev)
		assert.Equal(t, EventError, back.Type)
		assert.EqualError(t, back.Err, "io timeout")
	})
}
