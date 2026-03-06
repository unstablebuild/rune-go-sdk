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
	"context"

	"github.com/unstablebuild/tcell/v3"
	"github.com/unstablebuild/tcell/v3/termbox"
)

var _ Writer = (*TermboxWriter)(nil)

// TermboxWriter implements a termbox-like API using tcell/v3.
type TermboxWriter struct {
	ctx context.Context
}

// NewTermboxWriter allocates storage for a new TermboxWriter and initializes it.
func NewTermboxWriter() *TermboxWriter {
	ret := new(TermboxWriter)
	ret.ctx = context.Background()
	return ret
}

// SetCell satisfies term.Writer.
func (w *TermboxWriter) SetCell(pos Coordinates, c Cell) {
	termbox.Screen().SetContent(
		pos.X, pos.Y, c.Ch, c.Combining,
		c.Width, tcell.Style(c.Attributes),
	)
}

// UnionAttributes satisfies term.Writer.
func (w *TermboxWriter) UnionAttributes(pos Coordinates, attr Attributes) {
	termbox.Screen().UnionStyle(
		pos.X, pos.Y, tcell.Style(attr),
	)
}

// Flush makes all the content changes made using SetCell and
// UnionAttributes visible on the display.
func (w *TermboxWriter) Flush() error {
	return termbox.Flush()
}

// Clear fills the screen with the given attributes and empty cells.
func (w *TermboxWriter) Clear(attr Attributes) (err error) {
	termbox.Screen().Fill(' ', tcell.Style(attr))
	return
}

// SetCursor displays the terminal cursor at the given location.
func (w *TermboxWriter) SetCursor(pos Coordinates) {
	termbox.SetCursor(pos.X, pos.Y)
}

// Context satisfies term.Writer.
func (w *TermboxWriter) Context() context.Context {
	return w.ctx
}

// SetContext sets the context for the next call to Context.
func (w *TermboxWriter) SetContext(ctx context.Context) {
	w.ctx = ctx
}

func eventToTermboxEvent(ev Event) (tev termbox.Event) {
	tev.Mod = modToTermboxMod(ev.Mod)
	switch ev.Ch {
	case 0:
		switch ev.Key {
		case KeyBackspace:
			tev.Key = termbox.KeyBackspace2
		case KeySpace:
			switch ev.Mod & ModCtrl {
			case ModCtrl:
				tev.Key = termbox.KeyCtrlSpace
			default:
				tev.Key = termbox.Key(ev.Key)
			}
		default:
			tev.Key = termbox.Key(ev.Key)
		}
	default:
		switch ev.Mod & ModCtrl {
		case ModCtrl:
			switch ev.Ch {
			case 'A', 'a':
				tev.Key = termbox.KeyCtrlA
			case 'B', 'b':
				tev.Key = termbox.KeyCtrlB
			case 'C', 'c':
				tev.Key = termbox.KeyCtrlC
			case 'D', 'd':
				tev.Key = termbox.KeyCtrlD
			case 'E', 'e':
				tev.Key = termbox.KeyCtrlE
			case 'F', 'f':
				tev.Key = termbox.KeyCtrlF
			case 'G', 'g':
				tev.Key = termbox.KeyCtrlG
			case 'H', 'h':
				tev.Key = termbox.KeyCtrlH
			case 'I', 'i':
				tev.Key = termbox.KeyCtrlI
			case 'J', 'j':
				tev.Key = termbox.KeyCtrlJ
			case 'K', 'k':
				tev.Key = termbox.KeyCtrlK
			case 'L', 'l':
				tev.Key = termbox.KeyCtrlL
			case 'M', 'm':
				tev.Key = termbox.KeyCtrlM
			case 'N', 'n':
				tev.Key = termbox.KeyCtrlN
			case 'O', 'o':
				tev.Key = termbox.KeyCtrlO
			case 'P', 'p':
				tev.Key = termbox.KeyCtrlP
			case 'Q', 'q':
				tev.Key = termbox.KeyCtrlQ
			case 'R', 'r':
				tev.Key = termbox.KeyCtrlR
			case 'S', 's':
				tev.Key = termbox.KeyCtrlS
			case 'T', 't':
				tev.Key = termbox.KeyCtrlT
			case 'U', 'u':
				tev.Key = termbox.KeyCtrlU
			case 'V', 'v':
				tev.Key = termbox.KeyCtrlV
			case 'W', 'w':
				tev.Key = termbox.KeyCtrlW
			case 'X', 'x':
				tev.Key = termbox.KeyCtrlX
			case 'Y', 'y':
				tev.Key = termbox.KeyCtrlY
			case 'Z', 'z':
				tev.Key = termbox.KeyCtrlZ
			case '_', '-':
				tev.Key = termbox.KeyCtrlUnderscore
			case '~', '`':
				tev.Key = termbox.KeyCtrlTilde
			case ' ':
				tev.Key = termbox.KeyCtrlSpace
			case '2':
				tev.Key = termbox.KeyCtrl2
			case '3':
				tev.Key = termbox.KeyCtrl3
			case '4':
				tev.Key = termbox.KeyCtrl4
			case '5':
				tev.Key = termbox.KeyCtrl5
			case '6':
				tev.Key = termbox.KeyCtrl6
			case '7':
				tev.Key = termbox.KeyCtrl7
			case '8':
				tev.Key = termbox.KeyCtrl8
			case '/':
				tev.Key = termbox.KeyCtrlSlash
			case ']':
				tev.Key = termbox.KeyCtrlRsqBracket
			case '\\':
				tev.Key = termbox.KeyCtrlBackslash
			case '[':
				tev.Key = termbox.KeyCtrlLsqBracket
			default:
				tev.Ch = ev.Ch
			}
		default:
			tev.Ch = ev.Ch
		}
	}
	tev.Type = termbox.EventType(ev.Type)
	tev.Width = ev.Width
	tev.Height = ev.Height
	tev.Err = ev.Err
	tev.MouseX = ev.MouseX
	tev.MouseY = ev.MouseY
	tev.Raw = ev.Raw
	if ev.UserFunc != nil {
		tev.Metadata = ev.UserFunc
	}
	return
}

func termboxEventToEvent(tev termbox.Event) (ev Event) {
	switch tev.Ch {
	case 0:
		switch tev.Type {
		case termbox.EventKey:
			switch tev.Mod & termbox.ModAlt {
			case termbox.ModAlt:
				switch tev.Key {
				case termbox.KeyCtrlA:
					ev.Ch = 'a'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlB:
					ev.Ch = 'b'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlC:
					ev.Ch = 'c'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlD:
					ev.Ch = 'd'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlE:
					ev.Ch = 'e'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlF:
					ev.Ch = 'f'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlG:
					ev.Ch = 'g'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlH: /* termbox.KeyBackspace */
					ev.Ch = 'h'
					ev.Mod = ModCtrlAlt
				case termbox.KeyTab: /* termbox.KeyCtrlI: */
					ev.Key = KeyTab
					ev.Mod = ModAlt
				case termbox.KeyCtrlJ:
					ev.Ch = 'j'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlK:
					ev.Ch = 'k'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlL:
					ev.Ch = 'l'
					ev.Mod = ModCtrlAlt
				case termbox.KeyEnter: /* termbox.KeyCtrlM: */
					ev.Key = KeyEnter
					ev.Mod = ModAlt
				case termbox.KeyCtrlN:
					ev.Ch = 'n'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlO:
					ev.Ch = 'o'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlP:
					ev.Ch = 'p'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlQ:
					ev.Ch = 'q'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlR:
					ev.Ch = 'r'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlS:
					ev.Ch = 's'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlT:
					ev.Ch = 't'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlU:
					ev.Ch = 'u'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlV:
					ev.Ch = 'v'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlW:
					ev.Ch = 'w'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlX:
					ev.Ch = 'x'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlY:
					ev.Ch = 'y'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlZ:
					ev.Ch = 'z'
					ev.Mod = ModCtrlAlt
				case termbox.KeyEsc: /* termbox.KeyCtrl3: */
					ev.Key = KeyEsc
					ev.Mod = ModAlt
				case termbox.KeyTilde:
					ev.Ch = '~'
					ev.Mod = ModAlt
				case termbox.KeyCtrlBackslash: /* termbox.KeyCtrl4: */
					ev.Ch = '\\'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlRsqBracket: /* termbox.KeyCtrl5: */
					ev.Ch = ']'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrl6:
					ev.Ch = '6'
					ev.Mod = ModCtrlAlt
				case termbox.KeyCtrlSlash: /* KeyCtrl7, KeyCtrlUnderscore */
					ev.Ch = '/'
					ev.Mod = ModCtrlAlt
				case termbox.KeyBackspace2: /* termbox.KeyCtrl8 */
					ev.Key = KeyBackspace
					ev.Mod = ModAlt
				case termbox.KeyCtrlSpace: /* KeyCtrl2, KeyCtrTilde */
					ev.Key = KeySpace
					ev.Mod = ModCtrlAlt
				case termbox.KeySpace: // special case, termbox sends .Ch = ' '
					ev.Key = KeySpace
					ev.Mod = ModAlt
				default:
					ev.Ch = tev.Ch
					ev.Mod = ModAlt
					ev.Key = Key(tev.Key)
				}
			default:
				switch tev.Key {
				case termbox.KeyCtrlA:
					ev.Ch = 'a'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlB:
					ev.Ch = 'b'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlC:
					ev.Ch = 'c'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlD:
					ev.Ch = 'd'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlE:
					ev.Ch = 'e'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlF:
					ev.Ch = 'f'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlG:
					ev.Ch = 'g'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlH: /* termbox.KeyBackspace */
					ev.Ch = 'h'
					ev.Mod = ModCtrl
				case termbox.KeyTab: /* termbox.KeyCtrlI: */
					ev.Key = KeyTab
				case termbox.KeyCtrlJ:
					ev.Ch = 'j'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlK:
					ev.Ch = 'k'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlL:
					ev.Ch = 'l'
					ev.Mod = ModCtrl
				case termbox.KeyEnter: /* termbox.KeyCtrlM: */
					ev.Key = KeyEnter
				case termbox.KeyCtrlN:
					ev.Ch = 'n'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlO:
					ev.Ch = 'o'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlP:
					ev.Ch = 'p'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlQ:
					ev.Ch = 'q'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlR:
					ev.Ch = 'r'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlS:
					ev.Ch = 's'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlT:
					ev.Ch = 't'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlU:
					ev.Ch = 'u'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlV:
					ev.Ch = 'v'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlW:
					ev.Ch = 'w'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlX:
					ev.Ch = 'x'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlY:
					ev.Ch = 'y'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlZ:
					ev.Ch = 'z'
					ev.Mod = ModCtrl
				case termbox.KeyEsc: /* termbox.KeyCtrl3: */
					ev.Key = KeyEsc
				case termbox.KeyTilde:
					ev.Ch = '~'
				case termbox.KeyCtrlBackslash: /* termbox.KeyCtrl4: */
					ev.Ch = '\\'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlRsqBracket: /* termbox.KeyCtrl5: */
					ev.Ch = ']'
					ev.Mod = ModCtrl
				case termbox.KeyCtrl6:
					ev.Ch = '6'
					ev.Mod = ModCtrl
				case termbox.KeyCtrlSlash: /* KeyCtrl7, KeyCtrlUnderscore */
					ev.Ch = '/'
					ev.Mod = ModCtrl
				case termbox.KeyBackspace2: /* termbox.KeyCtrl8 */
					ev.Key = KeyBackspace
				case termbox.KeyCtrlSpace: /* KeyCtrl2, KeyCtrTilde */
					ev.Key = KeySpace
					ev.Mod = ModCtrl
				case termbox.KeySpace: // special case, termbox sends .Ch = ' '
					ev.Key = KeySpace
				default:
					ev.Ch = tev.Ch
					ev.Key = Key(tev.Key)
				}
			}
		default:
			ev.Key = Key(tev.Key)
		}
	case ' ':
		// handle all space and ctrl space quirks
		switch tev.Mod & termbox.ModAlt {
		case termbox.ModAlt:
			switch tev.Key {
			case termbox.KeySpace:
				ev.Mod = ModAlt
			case termbox.KeyCtrlSpace:
				ev.Mod = ModCtrlAlt
			}
		default:
			switch tev.Key {
			case termbox.KeyCtrlSpace:
				ev.Mod = ModCtrl
			}
		}
		ev.Key = KeySpace
	default:
		ev.Ch = tev.Ch
		ev.Key = Key(tev.Key)
	}
	ev.Mod |= termboxModToMod(tev.Mod)
	ev.Type = EventType(tev.Type)
	ev.Width = tev.Width
	ev.Height = tev.Height
	ev.Err = tev.Err
	ev.MouseX = tev.MouseX
	ev.MouseY = tev.MouseY
	ev.Raw = tev.Raw
	if tev.Metadata != nil {
		ev.UserFunc = tev.Metadata.(func())
	}
	return
}

var termboxModToTermTable = [16]Modifier{
	0:                                  0,
	termbox.ModShift:                   ModShift,
	termbox.ModCtrl:                    ModCtrl,
	termbox.ModShift | termbox.ModCtrl: ModCtrlShift,
	termbox.ModAlt:                     ModAlt,
	termbox.ModShift | termbox.ModAlt:  ModAltShift,
	termbox.ModCtrl | termbox.ModAlt:   ModCtrlAlt,
	termbox.ModShift | termbox.ModCtrl | termbox.ModAlt: ModCtrlShiftAlt,
	termbox.ModMeta:                                                       ModMeta,
	termbox.ModShift | termbox.ModMeta:                                    ModShiftMeta,
	termbox.ModCtrl | termbox.ModMeta:                                     ModCtrlMeta,
	termbox.ModShift | termbox.ModCtrl | termbox.ModMeta:                  ModCtrlShiftMeta,
	termbox.ModAlt | termbox.ModMeta:                                      ModAltMeta,
	termbox.ModShift | termbox.ModAlt | termbox.ModMeta:                   ModAltShiftMeta,
	termbox.ModCtrl | termbox.ModAlt | termbox.ModMeta:                    ModCtrlAltMeta,
	termbox.ModShift | termbox.ModCtrl | termbox.ModAlt | termbox.ModMeta: ModCtrl | ModAlt | ModShift | ModMeta,
}

func termboxModToMod(mod termbox.Modifier) Modifier {
	return termboxModToTermTable[mod&0xf]
}

var modToTermboxTable = [16]termbox.Modifier{
	0:                                     0,
	ModAlt:                                termbox.ModAlt,
	ModShift:                              termbox.ModShift,
	ModAltShift:                           termbox.ModAlt | termbox.ModShift,
	ModMeta:                               termbox.ModMeta,
	ModAltMeta:                            termbox.ModAlt | termbox.ModMeta,
	ModShiftMeta:                          termbox.ModShift | termbox.ModMeta,
	ModAltShiftMeta:                       termbox.ModAlt | termbox.ModShift | termbox.ModMeta,
	ModCtrl:                               termbox.ModCtrl,
	ModCtrlAlt:                            termbox.ModCtrl | termbox.ModAlt,
	ModCtrlShift:                          termbox.ModCtrl | termbox.ModShift,
	ModCtrlShiftAlt:                       termbox.ModCtrl | termbox.ModShift | termbox.ModAlt,
	ModCtrlMeta:                           termbox.ModCtrl | termbox.ModMeta,
	ModCtrlAltMeta:                        termbox.ModCtrl | termbox.ModAlt | termbox.ModMeta,
	ModCtrlShiftMeta:                      termbox.ModCtrl | termbox.ModShift | termbox.ModMeta,
	ModCtrl | ModAlt | ModShift | ModMeta: termbox.ModCtrl | termbox.ModAlt | termbox.ModShift | termbox.ModMeta,
}

func modToTermboxMod(mod Modifier) termbox.Modifier {
	return modToTermboxTable[mod&0xf]
}
