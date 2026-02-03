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

//go:build !js

package term

import (
	"github.com/unstablebuild/tcell/v3/termbox"
)

// Event type. See Event.Type field.
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

// List of keys.
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

// Input mode. See SetInputMode function.
const (
	InputEsc     InputMode = InputMode(termbox.InputEsc)
	InputAlt               = InputMode(termbox.InputAlt)
	InputMouse             = InputMode(termbox.InputMouse)
	InputCurrent           = InputMode(termbox.InputCurrent)
)

// Alt modifier constant, see Event.Mod field and SetInputMode function.
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
