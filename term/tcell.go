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

// This file is the only file in the term package that imports
// github.com/unstablebuild/tcell/v3 (or its termbox shim). Every
// helper that bridges between term.* types and tcell/termbox lives
// here so the rest of the package — Color, AttrMask, Style, Cell,
// Screen, ScreenWriter — stays free of tcell types.
//
// Contents, grouped:
//   - Color <-> tcell.Color conversion and tcell-backed Color methods.
//   - AttrMask / Style <-> tcell counterparts.
//   - CursorStyle constants (aliased onto tcell.CursorStyle).
//   - Event/Key/Modifier/InputMode constants (aliased onto termbox).
//   - termbox.Event <-> term.Event translation tables.
//   - TcellScreen, the adapter from tcell.Screen to term.Screen.
//   - Process-wide termbox lifecycle helpers (NewDefaultWriter,
//     PollEvent, PublishEvent, Close, SetInputMode, RingBell, ...).

package term

import (
	"sync/atomic"

	"github.com/unstablebuild/tcell/v3"
	"github.com/unstablebuild/tcell/v3/termbox"
)

// ---------------------------------------------------------------------------
// Color <-> tcell.Color
// ---------------------------------------------------------------------------

// Tcell converts c into a tcell.Color suitable for handing to a
// tcell-backed renderer. The bit layout of tcell.Color differs from
// term.Color (tcell uses bits 32/33/34 for its flags); this method
// performs the translation.
func (c Color) Tcell() tcell.Color {
	if c == ColorDefault {
		return tcell.ColorDefault
	}
	out := tcell.Color(uint64(c) & 0x00FFFFFF)
	if c&ColorValid != 0 {
		out |= tcell.ColorValid
	}
	if c&ColorIsRGB != 0 {
		out |= tcell.ColorIsRGB
	}
	if c&ColorSpecial != 0 {
		out |= tcell.ColorSpecial
	}
	return out
}

// FromTcellColor converts a tcell.Color into the term.Color layout.
func FromTcellColor(c tcell.Color) Color {
	if c == tcell.ColorDefault {
		return ColorDefault
	}
	out := Color(uint32(uint64(c) & 0x00FFFFFF))
	if c&tcell.ColorValid != 0 {
		out |= ColorValid
	}
	if c&tcell.ColorIsRGB != 0 {
		out |= ColorIsRGB
	}
	if c&tcell.ColorSpecial != 0 {
		out |= ColorSpecial
	}
	return out
}

// Hex returns c as a 24-bit RGB hex value, or -1 when c is not RGB.
func (c Color) Hex() int32 {
	if !c.Valid() {
		return -1
	}
	return c.Tcell().Hex()
}

// RGB returns the red, green, blue components of c, each in 0..255,
// or (-1, -1, -1) when c is not valid.
func (c Color) RGB() (int32, int32, int32) {
	if !c.Valid() {
		return -1, -1, -1
	}
	return c.Tcell().RGB()
}

// TrueColor returns c expanded to an RGB color, preserving its value
// when it is already RGB.
func (c Color) TrueColor() Color { return FromTcellColor(c.Tcell().TrueColor()) }

// String returns c as a "#rrggbb" hex literal, or "" when c is
// ColorDefault.
func (c Color) String() string { return c.Tcell().String() }

// CSS returns c as a CSS color string.
func (c Color) CSS() string { return c.Tcell().CSS() }

// Name returns the W3C name for c, or "" when c is unnamed.
func (c Color) Name(css ...bool) string { return c.Tcell().Name(css...) }

// NewColor returns the named Color matching r, g, b if one exists,
// otherwise the RGB Color produced by NewRGBColor.
func NewColor(r, g, b int32) Color { return FromTcellColor(tcell.NewColor(r, g, b)) }

// colorNames mirrors tcell.ColorNames into term.Color space. It is
// rebuilt whenever the underlying tcell palette changes via
// MergeColorValues.
var colorNames = buildColorNames()

func buildColorNames() map[string]Color {
	src := tcell.GetColorNames()
	out := make(map[string]Color, len(src))
	for k, v := range src {
		out[k] = FromTcellColor(v)
	}
	return out
}

// GetColorNames returns a snapshot of the W3C color name table as
// term.Color values.
func GetColorNames() map[string]Color { return buildColorNames() }

// MergeColorValues adds or overwrites entries in the underlying tcell
// palette table and refreshes the term color-name lookup so subsequent
// GetColor calls observe the updates.
func MergeColorValues(m map[Color]int32) {
	tm := make(map[tcell.Color]int32, len(m))
	for k, v := range m {
		tm[k.Tcell()] = v
	}
	tcell.MergeColorValues(tm)
	colorNames = buildColorNames()
}

// ---------------------------------------------------------------------------
// AttrMask / Style <-> tcell
// ---------------------------------------------------------------------------

// tcellAttrMask returns the low 8 bits of m as a tcell.AttrMask,
// dropping the term-specific render-offset bits that tcell does not
// understand.
func (m AttrMask) tcellAttrMask() tcell.AttrMask { return tcell.AttrMask(m & 0xFF) }

// AttrMaskFromTcell converts a tcell.AttrMask into a term.AttrMask.
func AttrMaskFromTcell(m tcell.AttrMask) AttrMask { return AttrMask(m) }

// Tcell returns s converted to a tcell.Style suitable for handing to
// a tcell-backed renderer.
func (s Style) Tcell() tcell.Style {
	return tcell.Style{
		Fg:    s.Fg.Tcell(),
		Bg:    s.Bg.Tcell(),
		Attrs: s.Attrs.tcellAttrMask(),
	}
}

// StyleFromTcell converts a tcell.Style into a term.Style.
func StyleFromTcell(s tcell.Style) Style {
	return Style{
		Fg:    FromTcellColor(s.Fg),
		Bg:    FromTcellColor(s.Bg),
		Attrs: AttrMaskFromTcell(s.Attrs),
	}
}

// styleToTcell is the internal companion to Style.Tcell used by the
// TcellScreen adapter to translate Screen.Fill / SetContent /
// UnionStyle calls.
func styleToTcell(s Style) tcell.Style { return s.Tcell() }

// ErrEventQFull is returned from Screen.PostEvent when the underlying
// tcell event queue is full.
var ErrEventQFull = tcell.ErrEventQFull

// ---------------------------------------------------------------------------
// CursorStyle (mirrors tcell)
// ---------------------------------------------------------------------------

// CursorStyle represents a cursor shape/blink combination. Support
// varies by terminal.
type CursorStyle int

// Supported cursor styles.
const (
	CursorStyleDefault           = CursorStyle(tcell.CursorStyleDefault)
	CursorStyleBlinkingBlock     = CursorStyle(tcell.CursorStyleBlinkingBlock)
	CursorStyleSteadyBlock       = CursorStyle(tcell.CursorStyleSteadyBlock)
	CursorStyleBlinkingUnderline = CursorStyle(tcell.CursorStyleBlinkingUnderline)
	CursorStyleSteadyUnderline   = CursorStyle(tcell.CursorStyleSteadyUnderline)
	CursorStyleBlinkingBar       = CursorStyle(tcell.CursorStyleBlinkingBar)
	CursorStyleSteadyBar         = CursorStyle(tcell.CursorStyleSteadyBar)
)

// SetCursorStyle sets the cursor style on the process-wide termbox
// screen. Has no effect when cursor styles are unsupported.
func SetCursorStyle(style CursorStyle) {
	termbox.Screen().SetCursorStyle(tcell.CursorStyle(style))
}

// ---------------------------------------------------------------------------
// Event / Key / Modifier / InputMode constants (mirror termbox)
// ---------------------------------------------------------------------------

// Event types. See Event.Type.
const (
	EventKey        EventType = EventType(termbox.EventKey)
	EventResize               = EventType(termbox.EventResize)
	EventMouse                = EventType(termbox.EventMouse)
	EventError                = EventType(termbox.EventError)
	EventInterrupt            = EventType(termbox.EventInterrupt)
	EventRaw                  = EventType(termbox.EventRaw)
	EventNone                 = EventType(termbox.EventNone)
	EventPasteStart           = EventType(termbox.EventPasteStart)
	EventPasteEnd             = EventType(termbox.EventPasteEnd)
	EventFocus                = EventType(termbox.EventFocus)
	EventUnfocus              = EventType(termbox.EventUnfocus)
)

// Keys and mouse-button pseudo-keys.
const (
	KeyF1          Key = Key(termbox.KeyF1)
	KeyF2              = Key(termbox.KeyF2)
	KeyF3              = Key(termbox.KeyF3)
	KeyF4              = Key(termbox.KeyF4)
	KeyF5              = Key(termbox.KeyF5)
	KeyF6              = Key(termbox.KeyF6)
	KeyF7              = Key(termbox.KeyF7)
	KeyF8              = Key(termbox.KeyF8)
	KeyF9              = Key(termbox.KeyF9)
	KeyF10             = Key(termbox.KeyF10)
	KeyF11             = Key(termbox.KeyF11)
	KeyF12             = Key(termbox.KeyF12)
	KeyInsert          = Key(termbox.KeyInsert)
	KeyDelete          = Key(termbox.KeyDelete)
	KeyHome            = Key(termbox.KeyHome)
	KeyEnd             = Key(termbox.KeyEnd)
	KeyPgup            = Key(termbox.KeyPgup)
	KeyPgdn            = Key(termbox.KeyPgdn)
	KeyArrowUp         = Key(termbox.KeyArrowUp)
	KeyArrowDown       = Key(termbox.KeyArrowDown)
	KeyArrowLeft       = Key(termbox.KeyArrowLeft)
	KeyArrowRight      = Key(termbox.KeyArrowRight)
	MouseLeft          = Key(termbox.MouseLeft)
	MouseMiddle        = Key(termbox.MouseMiddle)
	MouseRight         = Key(termbox.MouseRight)
	MouseRelease       = Key(termbox.MouseRelease)
	MouseWheelUp       = Key(termbox.MouseWheelUp)
	MouseWheelDown     = Key(termbox.MouseWheelDown)
	KeyBackspace       = Key(termbox.KeyBackspace2)
	KeyTab             = Key(termbox.KeyTab)
	KeyEnter           = Key(termbox.KeyEnter)
	KeyEsc             = Key(termbox.KeyEsc)
	KeySpace           = Key(termbox.KeySpace)
)

// Input modes (see SetInputMode).
const (
	InputEsc     InputMode = InputMode(termbox.InputEsc)
	InputAlt               = InputMode(termbox.InputAlt)
	InputMouse             = InputMode(termbox.InputMouse)
	InputCurrent           = InputMode(termbox.InputCurrent)
)

// Modifier bits (see Event.Mod and SetInputMode).
const (
	ModAlt Modifier = 1 << iota
	ModShift
	ModMeta
	ModCtrl

	ModCtrlShift     = ModShift | ModCtrl
	ModCtrlAlt       = ModCtrl | ModAlt
	ModCtrlMeta      = ModCtrl | ModMeta
	ModCtrlShiftAlt  = ModShift | ModAlt | ModCtrl
	ModCtrlShiftMeta = ModCtrl | ModShift | ModMeta
	ModCtrlAltMeta   = ModCtrl | ModAlt | ModMeta
	ModShiftMeta     = ModShift | ModMeta
	ModAltMeta       = ModAlt | ModMeta
	ModAltShiftMeta  = ModAlt | ModShift | ModMeta
	ModAltShift      = ModAlt | ModShift
)

// ---------------------------------------------------------------------------
// termbox.Event <-> term.Event translation
// ---------------------------------------------------------------------------

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

func termboxModToMod(mod termbox.Modifier) Modifier { return termboxModToTermTable[mod&0xf] }

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

func modToTermboxMod(mod Modifier) termbox.Modifier { return modToTermboxTable[mod&0xf] }

// ---------------------------------------------------------------------------
// TcellScreen adapter
// ---------------------------------------------------------------------------

// tcellScreen is the subset of tcell.Screen consumed by the
// TcellScreen adapter; matches the methods exposed by both
// tcell.Screen and the termbox shim.
type tcellScreen interface {
	SetContent(x, y int, primary rune, combining []rune, width uint8, style tcell.Style)
	UnionStyle(x, y int, style tcell.Style)
	Fill(ch rune, style tcell.Style)
	ShowCursor(x, y int)
	HideCursor()
	SetCursorStyle(tcell.CursorStyle)
	Size() (int, int)
	Show()
	Poll() <-chan tcell.Event
	PostEvent(tcell.Event) error
	Bell()
}

// TcellScreen adapts any tcell-shaped screen (tcell.Screen, the
// process termbox screen, simulation screens, GUI renderers, ...) to
// term.Screen. Color, style and event conversions happen here so the
// rest of the SDK and its consumers never see tcell types directly.
type TcellScreen struct {
	tcellScreen
	events chan Event
	stop   chan struct{}
}

// NewTcellScreen returns a Screen adapter backed by s. The adapter
// owns a goroutine that translates tcell events into term events; it
// is released when the underlying tcell.Screen channel closes.
func NewTcellScreen(s tcell.Screen) *TcellScreen { return newTcellScreenAdapter(s) }

// NewTermboxScreen returns a Screen adapter backed by the process
// termbox screen. Callers should ensure termbox.Init has been called
// before invoking this.
func NewTermboxScreen() *TcellScreen { return newTcellScreenAdapter(termbox.Screen()) }

func newTcellScreenAdapter(s tcellScreen) *TcellScreen {
	a := &TcellScreen{
		tcellScreen: s,
		events:      make(chan Event, 1),
		stop:        make(chan struct{}),
	}
	go a.translate(s.Poll())
	return a
}

func (a *TcellScreen) translate(src <-chan tcell.Event) {
	defer close(a.events)
	for tev := range src {
		select {
		case <-a.stop:
			return
		default:
		}
		ev := termboxEventToEvent(termbox.NewEvent(tev))
		select {
		case a.events <- ev:
		case <-a.stop:
			return
		}
	}
}

// Close stops the event-translation goroutine. The underlying tcell
// screen is not finalised.
func (a *TcellScreen) Close() {
	select {
	case <-a.stop:
	default:
		close(a.stop)
	}
}

// SetContent implements Screen.
func (a *TcellScreen) SetContent(x, y int, primary rune, combining []rune, width uint8, style Style) {
	a.tcellScreen.SetContent(x, y, primary, combining, width, styleToTcell(style))
}

// UnionStyle implements Screen.
func (a *TcellScreen) UnionStyle(x, y int, style Style) {
	a.tcellScreen.UnionStyle(x, y, styleToTcell(style))
}

// Fill implements Screen.
func (a *TcellScreen) Fill(ch rune, style Style) {
	a.tcellScreen.Fill(ch, styleToTcell(style))
}

// SetCursorStyle implements Screen.
func (a *TcellScreen) SetCursorStyle(s CursorStyle) {
	a.tcellScreen.SetCursorStyle(tcell.CursorStyle(s))
}

// Poll implements Screen.
func (a *TcellScreen) Poll() <-chan Event { return a.events }

// PostEvent implements Screen. The event is translated to a tcell
// event and pushed onto the underlying screen's event queue.
func (a *TcellScreen) PostEvent(ev Event) error {
	tev := eventToTermboxEvent(ev)
	var out tcell.Event
	switch tev.Type {
	case termbox.EventNone:
		out = tcell.NewEventInterrupt(termbox.EventNone, tev.Metadata)
	case termbox.EventKey:
		mod := tcell.ModMask(tev.Mod)
		k := tcell.Key(tev.Key)
		if tev.Ch != 0 {
			k = tcell.KeyRune
		}
		out = tcell.NewEventKey(k, tev.Ch, mod, tev.Raw)
	case termbox.EventResize:
		out = tcell.NewEventResize(tev.Width, tev.Height)
	case termbox.EventInterrupt:
		out = tcell.NewEventInterrupt(tev.Raw, tev.Metadata)
	case termbox.EventError:
		out = tcell.NewEventError(tev.Err)
	default:
		// silently ignore unsupported events (mouse, raw, paste)
		return nil
	}
	return a.tcellScreen.PostEvent(out)
}

// ---------------------------------------------------------------------------
// Process-wide termbox lifecycle helpers
// ---------------------------------------------------------------------------

var (
	defaultAttr  = Attributes{Fg: ColorDefault, Bg: ColorDefault}
	publishEvent atomic.Value
)

func init() {
	publishEvent.Store(func(termbox.Event) bool { return false })
}

// SetInputMode sets termbox input mode. See InputMode constants for
// the available modes.
func SetInputMode(mode InputMode) InputMode {
	return InputMode(termbox.SetInputMode(termbox.InputMode(mode)))
}

// SetAttr sets the global foreground and background attributes.
func SetAttr(newattr Attributes) { defaultAttr = newattr }

// Attr returns the global foreground and background attributes.
func Attr() Attributes { return defaultAttr }

// NewDefaultWriter initializes the underlying terminal client with
// the default screen and writer.
func NewDefaultWriter() (*ScreenWriter, error) {
	if err := termbox.Init(); err != nil {
		return nil, err
	}
	tscreen := termbox.Screen()
	tscreen.EnablePaste()
	tscreen.EnableFocus()
	publishEvent.Store(termbox.PublishEvent)
	return NewScreenWriter(NewTcellScreen(tscreen)), nil
}

// Size returns the size of the terminal window.
func Size() (width int, height int) { return termbox.Size() }

// PollEvent waits for an event and returns it. Blocking.
func PollEvent() (ev Event) { return termboxEventToEvent(termbox.PollEvent()) }

// Close releases the process-wide termbox screen.
func Close() { termbox.Close() }

// PublishEvent sends a synthetic event to the process-wide event
// poller. Returns false when the queue is full.
func PublishEvent(ev Event) bool {
	return publishEvent.Load().(func(termbox.Event) bool)(eventToTermboxEvent(ev))
}

// RingBell makes an audible noise. Must be synchronised against other
// accesses to the term.Writer's screen buffer.
func RingBell() { termbox.Screen().Bell() }

// PublishBell schedules RingBell on the next event-loop iteration.
func PublishBell() { ScheduleNextTick(RingBell) }

// ScheduleNextTick schedules running fn on the next event-loop iteration.
func ScheduleNextTick(fn func()) bool {
	return PublishEvent(Event{Type: EventInterrupt, UserFunc: fn})
}
