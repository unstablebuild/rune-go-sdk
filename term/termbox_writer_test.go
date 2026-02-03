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
			{keyComb{Mod: ModCtrl, Ch: 'a'}, termbox.Event{Key: termbox.KeyCtrlA}},
			{keyComb{Mod: ModCtrlAlt, Ch: 'a'}, termbox.Event{Mod: termbox.ModAlt, Key: termbox.KeyCtrlA}},
			{keyComb{Mod: ModCtrl, Ch: 'b'}, termbox.Event{Key: termbox.KeyCtrlB}},
			{keyComb{Mod: ModCtrl, Ch: 'c'}, termbox.Event{Key: termbox.KeyCtrlC}},
			{keyComb{Mod: ModCtrl, Ch: 'd'}, termbox.Event{Key: termbox.KeyCtrlD}},
			{keyComb{Mod: ModCtrl, Ch: 'e'}, termbox.Event{Key: termbox.KeyCtrlE}},
			{keyComb{Mod: ModCtrl, Ch: 'f'}, termbox.Event{Key: termbox.KeyCtrlF}},
			{keyComb{Mod: ModCtrl, Ch: 'g'}, termbox.Event{Key: termbox.KeyCtrlG}},
			{keyComb{Mod: ModCtrl, Ch: 'j'}, termbox.Event{Key: termbox.KeyCtrlJ}},
			{keyComb{Mod: ModCtrl, Ch: 'k'}, termbox.Event{Key: termbox.KeyCtrlK}},
			{keyComb{Mod: ModCtrl, Ch: 'l'}, termbox.Event{Key: termbox.KeyCtrlL}},
			{keyComb{Mod: ModCtrl, Ch: 'n'}, termbox.Event{Key: termbox.KeyCtrlN}},
			{keyComb{Mod: ModCtrl, Ch: 'o'}, termbox.Event{Key: termbox.KeyCtrlO}},
			{keyComb{Mod: ModCtrl, Ch: 'p'}, termbox.Event{Key: termbox.KeyCtrlP}},
			{keyComb{Mod: ModCtrl, Ch: 'q'}, termbox.Event{Key: termbox.KeyCtrlQ}},
			{keyComb{Mod: ModCtrl, Ch: 'r'}, termbox.Event{Key: termbox.KeyCtrlR}},
			{keyComb{Mod: ModCtrl, Ch: 's'}, termbox.Event{Key: termbox.KeyCtrlS}},
			{keyComb{Mod: ModCtrl, Ch: 't'}, termbox.Event{Key: termbox.KeyCtrlT}},
			{keyComb{Mod: ModCtrl, Ch: 'u'}, termbox.Event{Key: termbox.KeyCtrlU}},
			{keyComb{Mod: ModCtrl, Ch: 'v'}, termbox.Event{Key: termbox.KeyCtrlV}},
			{keyComb{Mod: ModCtrl, Ch: 'w'}, termbox.Event{Key: termbox.KeyCtrlW}},
			{keyComb{Mod: ModCtrl, Ch: 'x'}, termbox.Event{Key: termbox.KeyCtrlX}},
			{keyComb{Mod: ModCtrl, Ch: 'y'}, termbox.Event{Key: termbox.KeyCtrlY}},
			{keyComb{Mod: ModCtrl, Ch: 'z'}, termbox.Event{Key: termbox.KeyCtrlZ}},
			{keyComb{Key: KeyBackspace}, termbox.Event{Key: termbox.KeyBackspace2}},
			{keyComb{Key: KeyTab}, termbox.Event{Key: termbox.KeyTab}},
			{keyComb{Key: KeyEnter}, termbox.Event{Key: termbox.KeyEnter}},
			{keyComb{Key: KeyEsc}, termbox.Event{Key: termbox.KeyEsc}},
			{keyComb{Key: KeyPgdn}, termbox.Event{Key: termbox.KeyPgdn}},
			{keyComb{Key: KeyPgup}, termbox.Event{Key: termbox.KeyPgup}},
			{keyComb{Key: KeySpace}, termbox.Event{Key: termbox.KeySpace}},
			{keyComb{Mod: ModAlt, Key: KeySpace}, termbox.Event{Mod: termbox.ModAlt, Key: termbox.KeySpace}},
			{keyComb{Key: KeySpace, Mod: ModCtrl}, termbox.Event{Key: termbox.KeyCtrlSpace}},
			{keyComb{Key: MouseLeft}, termbox.Event{Key: termbox.MouseLeft}},
			{keyComb{Key: MouseRight}, termbox.Event{Key: termbox.MouseRight}},
			{keyComb{Key: MouseMiddle}, termbox.Event{Key: termbox.MouseMiddle}},
			{keyComb{Key: MouseRelease}, termbox.Event{Key: termbox.MouseRelease}},
			{keyComb{Key: MouseWheelUp}, termbox.Event{Key: termbox.MouseWheelUp}},
			{keyComb{Key: MouseWheelDown}, termbox.Event{Key: termbox.MouseWheelDown}},

			// ambiguous keys
			{keyComb{Ch: 'h', Mod: ModCtrl}, termbox.Event{Key: termbox.KeyCtrlH}},
			{keyComb{Mod: ModCtrl, Key: KeySpace}, termbox.Event{Key: termbox.KeyCtrl2}},
			{keyComb{Key: KeyEsc}, termbox.Event{Key: termbox.KeyCtrl3}},
			{keyComb{Ch: '6', Mod: ModCtrl}, termbox.Event{Key: termbox.KeyCtrl6}},
			{keyComb{Mod: ModCtrl, Ch: '/'}, termbox.Event{Key: termbox.KeyCtrl7}},
			{keyComb{Ch: '/', Mod: ModCtrl}, termbox.Event{Key: termbox.KeyCtrlSlash}},
			{keyComb{Ch: '\\', Mod: ModCtrl}, termbox.Event{Key: termbox.KeyCtrlBackslash}},
			{keyComb{Key: KeyEsc}, termbox.Event{Key: termbox.KeyCtrlLsqBracket}},
			{keyComb{Ch: '\\', Mod: ModCtrl}, termbox.Event{Key: termbox.KeyCtrl4}},
			{keyComb{Ch: ']', Mod: ModCtrl}, termbox.Event{Key: termbox.KeyCtrl5}},
			{keyComb{Mod: ModCtrl, Ch: '/'}, termbox.Event{Key: termbox.KeyCtrlUnderscore}},
			{keyComb{Key: KeyBackspace}, termbox.Event{Key: termbox.KeyCtrl8}},
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
