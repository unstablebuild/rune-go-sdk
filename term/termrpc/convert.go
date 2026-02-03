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


package termrpc

import (
	"fmt"

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/tcell/v3"
)

// ToModel maps this Event into a term.Event.
func (e *Event) ToModel() (ev term.Event, err error) {
	switch e.Type {
	case Event_TypeKey:
		ev.Type = term.EventKey
	case Event_TypeMouse:
		ev.Type = term.EventMouse
	case Event_TypeNone:
		ev.Type = term.EventNone
	case Event_TypeInterrupt:
		ev.Type = term.EventInterrupt
	case Event_TypePasteStart:
		ev.Type = term.EventPasteStart
	case Event_TypePasteEnd:
		ev.Type = term.EventPasteEnd
	case Event_TypeFocus:
		ev.Type = term.EventFocus
	case Event_TypeUnfocus:
		ev.Type = term.EventUnfocus
	case Event_TypeRaw:
		ev.Type = term.EventRaw
	case Event_TypeResize:
		ev.Type = term.EventResize
	case Event_TypeError:
		ev.Type = term.EventError
	default:
		return term.Event{},
			fmt.Errorf("serialization error: unknown event type: %s", e.Type)
	}

	switch e.Mod {
	case Event_Alt:
		ev.Mod = term.ModAlt
	case Event_Shift:
		ev.Mod = term.ModShift
	case Event_Meta:
		ev.Mod = term.ModMeta
	case Event_Ctrl:
		ev.Mod = term.ModCtrl
	case Event_CtrlShift:
		ev.Mod = term.ModCtrlShift
	case Event_CtrlAlt:
		ev.Mod = term.ModCtrlAlt
	case Event_CtrlMeta:
		ev.Mod = term.ModCtrlMeta
	case Event_CtrlShiftAlt:
		ev.Mod = term.ModCtrlShiftAlt
	case Event_CtrlShiftMeta:
		ev.Mod = term.ModCtrlShiftMeta
	case Event_CtrlAltMeta:
		ev.Mod = term.ModCtrlAltMeta
	case Event_ShiftMeta:
		ev.Mod = term.ModShiftMeta
	case Event_AltMeta:
		ev.Mod = term.ModAltMeta
	case Event_AltShiftMeta:
		ev.Mod = term.ModAltShiftMeta
	case Event_AltShift:
		ev.Mod = term.ModAltShift
	case Event_None:
	default:
		return term.Event{},
			fmt.Errorf("serialization error: unknown event Mod: %s", e.Mod)
	}

	switch e.Key {
	case Event_F1:
		ev.Key = term.KeyF1
	case Event_F2:
		ev.Key = term.KeyF2
	case Event_F3:
		ev.Key = term.KeyF3
	case Event_F4:
		ev.Key = term.KeyF4
	case Event_F5:
		ev.Key = term.KeyF5
	case Event_F6:
		ev.Key = term.KeyF6
	case Event_F7:
		ev.Key = term.KeyF7
	case Event_F8:
		ev.Key = term.KeyF8
	case Event_F9:
		ev.Key = term.KeyF9
	case Event_F10:
		ev.Key = term.KeyF10
	case Event_F11:
		ev.Key = term.KeyF11
	case Event_F12:
		ev.Key = term.KeyF12
	case Event_Insert:
		ev.Key = term.KeyInsert
	case Event_Delete:
		ev.Key = term.KeyDelete
	case Event_Home:
		ev.Key = term.KeyHome
	case Event_End:
		ev.Key = term.KeyEnd
	case Event_Pgup:
		ev.Key = term.KeyPgup
	case Event_Pgdn:
		ev.Key = term.KeyPgdn
	case Event_ArrowUp:
		ev.Key = term.KeyArrowUp
	case Event_ArrowDown:
		ev.Key = term.KeyArrowDown
	case Event_ArrowLeft:
		ev.Key = term.KeyArrowLeft
	case Event_ArrowRight:
		ev.Key = term.KeyArrowRight
	case Event_MouseLeft:
		ev.Key = term.MouseLeft
	case Event_MouseMiddle:
		ev.Key = term.MouseMiddle
	case Event_MouseRight:
		ev.Key = term.MouseRight
	case Event_MouseRelease:
		ev.Key = term.MouseRelease
	case Event_MouseWheelUp:
		ev.Key = term.MouseWheelUp
	case Event_MouseWheelDown:
		ev.Key = term.MouseWheelDown
	case Event_Backspace:
		ev.Key = term.KeyBackspace
	case Event_Tab:
		ev.Key = term.KeyTab
	case Event_Enter:
		ev.Key = term.KeyEnter
	case Event_Esc:
		ev.Key = term.KeyEsc
	case Event_Space:
		ev.Key = term.KeySpace
	case Event_Null:
	default:
		return term.Event{},
			fmt.Errorf("serialization error: unknown event key: %s", e.Key)
	}

	ev.Ch = rune(e.Char)
	ev.MouseX = int(e.GetMouseX())
	ev.MouseY = int(e.GetMouseY())
	ev.Raw = e.GetRaw()

	return ev, nil
}

// ToModel maps this Coordinates into term.Coordinates.
func (c *Coordinates) ToModel() term.Coordinates {
	return term.Coordinates{
		X: int(c.GetX()),
		Y: int(c.GetY()),
	}
}

// FromModel maps this Coordinates from term.Coordinates.
func (c *Coordinates) FromModel(pos term.Coordinates) {
	c.X = int32(pos.X)
	c.Y = int32(pos.Y)
}

// FromModel takes ev and maps it into this Event.
func (e *Event) FromModel(ev term.Event) error {
	switch ev.Type {
	case term.EventKey:
		e.Type = Event_TypeKey
	case term.EventMouse:
		e.Type = Event_TypeMouse
	case term.EventNone:
		e.Type = Event_TypeNone
	case term.EventInterrupt:
		e.Type = Event_TypeInterrupt
	case term.EventPasteStart:
		e.Type = Event_TypePasteStart
	case term.EventPasteEnd:
		e.Type = Event_TypePasteEnd
	case term.EventFocus:
		e.Type = Event_TypeFocus
	case term.EventUnfocus:
		e.Type = Event_TypeUnfocus
	case term.EventError:
		e.Type = Event_TypeError
	case term.EventResize:
		e.Type = Event_TypeResize
	case term.EventRaw:
		e.Type = Event_TypeRaw
	default:
		return fmt.Errorf("serialization error: unknown event type: %d", ev.Type)
	}

	switch ev.Mod {
	case term.ModAlt:
		e.Mod = Event_Alt
	case term.ModShift:
		e.Mod = Event_Shift
	case term.ModMeta:
		e.Mod = Event_Meta
	case term.ModCtrl:
		e.Mod = Event_Ctrl
	case term.ModCtrlShift:
		e.Mod = Event_CtrlShift
	case term.ModCtrlAlt:
		e.Mod = Event_CtrlAlt
	case term.ModCtrlMeta:
		e.Mod = Event_CtrlMeta
	case term.ModCtrlShiftAlt:
		e.Mod = Event_CtrlShiftAlt
	case term.ModCtrlShiftMeta:
		e.Mod = Event_CtrlShiftMeta
	case term.ModCtrlAltMeta:
		e.Mod = Event_CtrlAltMeta
	case term.ModShiftMeta:
		e.Mod = Event_ShiftMeta
	case term.ModAltMeta:
		e.Mod = Event_AltMeta
	case term.ModAltShiftMeta:
		e.Mod = Event_AltShiftMeta
	case term.ModAltShift:
		e.Mod = Event_AltShift
	case term.Modifier(0):
		e.Mod = Event_None
	default:
		return fmt.Errorf("serialization error: unknown event Mod: %+v", ev.Mod)
	}

	switch ev.Key {
	case term.KeyF1:
		e.Key = Event_F1
	case term.KeyF2:
		e.Key = Event_F2
	case term.KeyF3:
		e.Key = Event_F3
	case term.KeyF4:
		e.Key = Event_F4
	case term.KeyF5:
		e.Key = Event_F5
	case term.KeyF6:
		e.Key = Event_F6
	case term.KeyF7:
		e.Key = Event_F7
	case term.KeyF8:
		e.Key = Event_F8
	case term.KeyF9:
		e.Key = Event_F9
	case term.KeyF10:
		e.Key = Event_F10
	case term.KeyF11:
		e.Key = Event_F11
	case term.KeyF12:
		e.Key = Event_F12
	case term.KeyInsert:
		e.Key = Event_Insert
	case term.KeyDelete:
		e.Key = Event_Delete
	case term.KeyHome:
		e.Key = Event_Home
	case term.KeyEnd:
		e.Key = Event_End
	case term.KeyPgup:
		e.Key = Event_Pgup
	case term.KeyPgdn:
		e.Key = Event_Pgdn
	case term.KeyArrowUp:
		e.Key = Event_ArrowUp
	case term.KeyArrowDown:
		e.Key = Event_ArrowDown
	case term.KeyArrowLeft:
		e.Key = Event_ArrowLeft
	case term.KeyArrowRight:
		e.Key = Event_ArrowRight
	case term.MouseLeft:
		e.Key = Event_MouseLeft
	case term.MouseMiddle:
		e.Key = Event_MouseMiddle
	case term.MouseRight:
		e.Key = Event_MouseRight
	case term.MouseRelease:
		e.Key = Event_MouseRelease
	case term.MouseWheelUp:
		e.Key = Event_MouseWheelUp
	case term.MouseWheelDown:
		e.Key = Event_MouseWheelDown
	case term.KeyBackspace:
		e.Key = Event_Backspace
	case term.KeyTab:
		e.Key = Event_Tab
	case term.KeyEnter:
		e.Key = Event_Enter
	case term.KeyEsc:
		e.Key = Event_Esc
	case term.KeySpace:
		e.Key = Event_Space
	case term.Key(0):
		e.Key = Event_Null
	default:
		return fmt.Errorf("serialization error: unknown event key: %+v", ev.Key)
	}

	e.Char = uint32(ev.Ch)
	e.MouseX = int32(ev.MouseX)
	e.MouseY = int32(ev.MouseY)
	e.Raw = ev.Raw

	return nil
}

// ToModel maps this Cell into a term.Cell.
func (c *Cell) ToModel() term.Cell {
	if c == nil {
		return term.Cell{}
	}
	var combining []rune
	// prefer nil combining rather than a slice of length 0
	if c.Combining != nil {
		combining = make([]rune, len(c.Combining))
		for i, cell := range c.Combining {
			combining[i] = rune(cell)
		}
	}
	return term.Cell{
		Attributes: term.Attributes{
			Bg:    tcell.Color(c.Background),
			Fg:    tcell.Color(c.Foreground),
			Attrs: tcell.AttrMask(c.Attrs),
		},
		Ch:        rune(c.Character),
		Combining: combining,
		Width:     uint8(c.Width),
		Bytes:     uint8(c.Bytes),
	}
}

// ToModel maps this Attributes into the corresponding term.Attributes.
func (a *Attributes) ToModel() term.Attributes {
	return term.Attributes{
		Bg:    tcell.Color(a.Background),
		Fg:    tcell.Color(a.Foreground),
		Attrs: tcell.AttrMask(a.Attrs),
	}
}

// FromModel sets this Attributes from attr term.Attributes.
func (a *Attributes) FromModel(attr term.Attributes) {
	a.Background = uint64(attr.Bg)
	a.Foreground = uint64(attr.Fg)
	a.Attrs = int64(attr.Attrs)
}

// FromModel takes cc and maps it into this Cell.
func (c *Cell) FromModel(cc term.Cell) {
	c.Background = uint64(cc.Bg)
	c.Foreground = uint64(cc.Fg)
	c.Attrs = int64(cc.Attrs)
	c.Character = uint32(cc.Ch)
	c.Width = uint32(c.Width)
	c.Bytes = uint32(c.Bytes)
	var combining []uint32
	if cc.Combining != nil {
		combining = make([]uint32, len(cc.Combining))
		for i, cell := range cc.Combining {
			combining[i] = uint32(cell)
		}
	}
	c.Combining = combining
}
