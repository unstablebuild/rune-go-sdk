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
)

// String returns the long string representation of this KeyComb.
func (k KeyComb) String() string {
	if k.Ch != 0 && k.Key != KeySpace {
		ch, mod := processCharacter(k.Ch, k.Mod)
		// if Ch is set, then ignore key for representing this key comb
		switch mod {
		case ModMeta:
			return fmt.Sprintf("<meta-%s>", ch)
		case ModAlt:
			return fmt.Sprintf("<alt-%s>", ch)
		case ModShift:
			return fmt.Sprintf("<shift-%s>", ch)
		case ModCtrl:
			return fmt.Sprintf("<ctrl-%s>", ch)
		case ModCtrlShift:
			return fmt.Sprintf("<ctrl-shift-%s>", ch)
		case ModCtrlAlt:
			return fmt.Sprintf("<ctrl-alt-%s>", ch)
		case ModCtrlMeta:
			return fmt.Sprintf("<ctrl-meta-%s>", ch)
		case ModCtrlShiftAlt:
			return fmt.Sprintf("<ctrl-shift-alt-%s>", ch)
		case ModCtrlShiftMeta:
			return fmt.Sprintf("<ctrl-shift-meta-%s>", ch)
		case ModCtrlAltMeta:
			return fmt.Sprintf("<ctrl-alt-meta-%s>", ch)
		case ModShiftMeta:
			return fmt.Sprintf("<shift-meta-%s>", ch)
		case ModAltMeta:
			return fmt.Sprintf("<alt-meta-%s>", ch)
		case ModAltShiftMeta:
			return fmt.Sprintf("<alt-shift-meta-%s>", ch)
		case ModAltShift:
			return fmt.Sprintf("<alt-shift-%s>", ch)
		default:
			return string(ch)
		}
	}

	switch k {
	case KeyComb{Key: KeyF1}:
		return "<f1>"
	case KeyComb{Key: KeyF2}:
		return "<f2>"
	case KeyComb{Key: KeyF3}:
		return "<f3>"
	case KeyComb{Key: KeyF4}:
		return "<f4>"
	case KeyComb{Key: KeyF5}:
		return "<f5>"
	case KeyComb{Key: KeyF6}:
		return "<f6>"
	case KeyComb{Key: KeyF7}:
		return "<f7>"
	case KeyComb{Key: KeyF8}:
		return "<f8>"
	case KeyComb{Key: KeyF9}:
		return "<f9>"
	case KeyComb{Key: KeyF10}:
		return "<f10>"
	case KeyComb{Key: KeyF11}:
		return "<f11>"
	case KeyComb{Key: KeyF12}:
		return "<f12>"
	case KeyComb{Key: KeyInsert}:
		return "<insert>"
	case KeyComb{Key: KeyDelete}:
		return "<delete>"
	case KeyComb{Key: KeyHome}:
		return "<home>"
	case KeyComb{Key: KeyEnd}:
		return "<end>"
	case KeyComb{Key: KeyPgup}:
		return "<pgup>"
	case KeyComb{Key: KeyPgdn}:
		return "<pgdn>"
	case KeyComb{Key: KeyArrowUp}:
		return "<up>"
	case KeyComb{Key: KeyArrowDown}:
		return "<down>"
	case KeyComb{Key: KeyArrowLeft}:
		return "<left>"
	case KeyComb{Key: KeyArrowRight}:
		return "<right>"
	case KeyComb{Key: MouseLeft}:
		return "<mouse-left>"
	case KeyComb{Key: MouseMiddle}:
		return "<mouse-middle>"
	case KeyComb{Key: MouseRight}:
		return "<mouse-right>"
	case KeyComb{Key: MouseRelease}:
		return "<mouse-release>"
	case KeyComb{Key: MouseWheelUp}:
		return "<mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown}:
		return "<mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace}:
		return "<backspace>"
	case KeyComb{Key: KeyTab}:
		return "<tab>"
	case KeyComb{Key: KeyEnter}:
		return "<enter>"
	case KeyComb{Key: KeyEsc}:
		return "<esc>"
	case KeyComb{Key: KeySpace, Ch: ' '}, KeyComb{Key: KeySpace}:
		return "<space>"
	case KeyComb{Key: KeyCapsLock}:
		return "<capslock>"
	case KeyComb{Key: KeyNumLock}:
		return "<numlock>"
	case KeyComb{Key: KeyScrollLock}:
		return "<scrolllock>"
	case KeyComb{Key: KeyMenu}:
		return "<menu>"

	// meta
	case KeyComb{Key: KeyF1, Mod: ModMeta}:
		return "<meta-f1>"
	case KeyComb{Key: KeyF2, Mod: ModMeta}:
		return "<meta-f2>"
	case KeyComb{Key: KeyF3, Mod: ModMeta}:
		return "<meta-f3>"
	case KeyComb{Key: KeyF4, Mod: ModMeta}:
		return "<meta-f4>"
	case KeyComb{Key: KeyF5, Mod: ModMeta}:
		return "<meta-f5>"
	case KeyComb{Key: KeyF6, Mod: ModMeta}:
		return "<meta-f6>"
	case KeyComb{Key: KeyF7, Mod: ModMeta}:
		return "<meta-f7>"
	case KeyComb{Key: KeyF8, Mod: ModMeta}:
		return "<meta-f8>"
	case KeyComb{Key: KeyF9, Mod: ModMeta}:
		return "<meta-f9>"
	case KeyComb{Key: KeyF10, Mod: ModMeta}:
		return "<meta-f10>"
	case KeyComb{Key: KeyF11, Mod: ModMeta}:
		return "<meta-f11>"
	case KeyComb{Key: KeyF12, Mod: ModMeta}:
		return "<meta-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModMeta}:
		return "<meta-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModMeta}:
		return "<meta-delete>"
	case KeyComb{Key: KeyHome, Mod: ModMeta}:
		return "<meta-home>"
	case KeyComb{Key: KeyEnd, Mod: ModMeta}:
		return "<meta-end>"
	case KeyComb{Key: KeyPgup, Mod: ModMeta}:
		return "<meta-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModMeta}:
		return "<meta-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModMeta}:
		return "<meta-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModMeta}:
		return "<meta-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModMeta}:
		return "<meta-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModMeta}:
		return "<meta-right>"
	case KeyComb{Key: MouseLeft, Mod: ModMeta}:
		return "<meta-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModMeta}:
		return "<meta-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModMeta}:
		return "<meta-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModMeta}:
		return "<meta-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModMeta}:
		return "<meta-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModMeta}:
		return "<meta-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModMeta}:
		return "<meta-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModMeta}:
		return "<meta-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModMeta}:
		return "<meta-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModMeta}:
		return "<meta-esc>"
	case KeyComb{Key: KeySpace, Mod: ModMeta}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModMeta}:
		return "<meta-space>"

	// alt
	case KeyComb{Key: KeyF1, Mod: ModAlt}:
		return "<alt-f1>"
	case KeyComb{Key: KeyF2, Mod: ModAlt}:
		return "<alt-f2>"
	case KeyComb{Key: KeyF3, Mod: ModAlt}:
		return "<alt-f3>"
	case KeyComb{Key: KeyF4, Mod: ModAlt}:
		return "<alt-f4>"
	case KeyComb{Key: KeyF5, Mod: ModAlt}:
		return "<alt-f5>"
	case KeyComb{Key: KeyF6, Mod: ModAlt}:
		return "<alt-f6>"
	case KeyComb{Key: KeyF7, Mod: ModAlt}:
		return "<alt-f7>"
	case KeyComb{Key: KeyF8, Mod: ModAlt}:
		return "<alt-f8>"
	case KeyComb{Key: KeyF9, Mod: ModAlt}:
		return "<alt-f9>"
	case KeyComb{Key: KeyF10, Mod: ModAlt}:
		return "<alt-f10>"
	case KeyComb{Key: KeyF11, Mod: ModAlt}:
		return "<alt-f11>"
	case KeyComb{Key: KeyF12, Mod: ModAlt}:
		return "<alt-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModAlt}:
		return "<alt-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModAlt}:
		return "<alt-delete>"
	case KeyComb{Key: KeyHome, Mod: ModAlt}:
		return "<alt-home>"
	case KeyComb{Key: KeyEnd, Mod: ModAlt}:
		return "<alt-end>"
	case KeyComb{Key: KeyPgup, Mod: ModAlt}:
		return "<alt-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModAlt}:
		return "<alt-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModAlt}:
		return "<alt-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModAlt}:
		return "<alt-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModAlt}:
		return "<alt-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModAlt}:
		return "<alt-right>"
	case KeyComb{Key: MouseLeft, Mod: ModAlt}:
		return "<alt-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModAlt}:
		return "<alt-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModAlt}:
		return "<alt-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModAlt}:
		return "<alt-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModAlt}:
		return "<alt-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModAlt}:
		return "<alt-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModAlt}:
		return "<alt-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModAlt}:
		return "<alt-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModAlt}:
		return "<alt-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModAlt}:
		return "<alt-esc>"
	case KeyComb{Key: KeySpace, Mod: ModAlt}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModAlt}:
		return "<alt-space>"

	// shift
	case KeyComb{Key: KeyF1, Mod: ModShift}:
		return "<shift-f1>"
	case KeyComb{Key: KeyF2, Mod: ModShift}:
		return "<shift-f2>"
	case KeyComb{Key: KeyF3, Mod: ModShift}:
		return "<shift-f3>"
	case KeyComb{Key: KeyF4, Mod: ModShift}:
		return "<shift-f4>"
	case KeyComb{Key: KeyF5, Mod: ModShift}:
		return "<shift-f5>"
	case KeyComb{Key: KeyF6, Mod: ModShift}:
		return "<shift-f6>"
	case KeyComb{Key: KeyF7, Mod: ModShift}:
		return "<shift-f7>"
	case KeyComb{Key: KeyF8, Mod: ModShift}:
		return "<shift-f8>"
	case KeyComb{Key: KeyF9, Mod: ModShift}:
		return "<shift-f9>"
	case KeyComb{Key: KeyF10, Mod: ModShift}:
		return "<shift-f10>"
	case KeyComb{Key: KeyF11, Mod: ModShift}:
		return "<shift-f11>"
	case KeyComb{Key: KeyF12, Mod: ModShift}:
		return "<shift-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModShift}:
		return "<shift-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModShift}:
		return "<shift-delete>"
	case KeyComb{Key: KeyHome, Mod: ModShift}:
		return "<shift-home>"
	case KeyComb{Key: KeyEnd, Mod: ModShift}:
		return "<shift-end>"
	case KeyComb{Key: KeyPgup, Mod: ModShift}:
		return "<shift-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModShift}:
		return "<shift-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModShift}:
		return "<shift-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModShift}:
		return "<shift-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModShift}:
		return "<shift-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModShift}:
		return "<shift-right>"
	case KeyComb{Key: MouseLeft, Mod: ModShift}:
		return "<shift-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModShift}:
		return "<shift-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModShift}:
		return "<shift-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModShift}:
		return "<shift-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModShift}:
		return "<shift-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModShift}:
		return "<shift-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModShift}:
		return "<shift-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModShift}:
		return "<shift-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModShift}:
		return "<shift-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModShift}:
		return "<shift-esc>"
	case KeyComb{Key: KeySpace, Mod: ModShift}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModShift}:
		return "<shift-space>"

	// ctrl
	case KeyComb{Key: KeyF1, Mod: ModCtrl}:
		return "<ctrl-f1>"
	case KeyComb{Key: KeyF2, Mod: ModCtrl}:
		return "<ctrl-f2>"
	case KeyComb{Key: KeyF3, Mod: ModCtrl}:
		return "<ctrl-f3>"
	case KeyComb{Key: KeyF4, Mod: ModCtrl}:
		return "<ctrl-f4>"
	case KeyComb{Key: KeyF5, Mod: ModCtrl}:
		return "<ctrl-f5>"
	case KeyComb{Key: KeyF6, Mod: ModCtrl}:
		return "<ctrl-f6>"
	case KeyComb{Key: KeyF7, Mod: ModCtrl}:
		return "<ctrl-f7>"
	case KeyComb{Key: KeyF8, Mod: ModCtrl}:
		return "<ctrl-f8>"
	case KeyComb{Key: KeyF9, Mod: ModCtrl}:
		return "<ctrl-f9>"
	case KeyComb{Key: KeyF10, Mod: ModCtrl}:
		return "<ctrl-f10>"
	case KeyComb{Key: KeyF11, Mod: ModCtrl}:
		return "<ctrl-f11>"
	case KeyComb{Key: KeyF12, Mod: ModCtrl}:
		return "<ctrl-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModCtrl}:
		return "<ctrl-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModCtrl}:
		return "<ctrl-delete>"
	case KeyComb{Key: KeyHome, Mod: ModCtrl}:
		return "<ctrl-home>"
	case KeyComb{Key: KeyEnd, Mod: ModCtrl}:
		return "<ctrl-end>"
	case KeyComb{Key: KeyPgup, Mod: ModCtrl}:
		return "<ctrl-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModCtrl}:
		return "<ctrl-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModCtrl}:
		return "<ctrl-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModCtrl}:
		return "<ctrl-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModCtrl}:
		return "<ctrl-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModCtrl}:
		return "<ctrl-right>"
	case KeyComb{Key: MouseLeft, Mod: ModCtrl}:
		return "<ctrl-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModCtrl}:
		return "<ctrl-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModCtrl}:
		return "<ctrl-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModCtrl}:
		return "<ctrl-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModCtrl}:
		return "<ctrl-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModCtrl}:
		return "<ctrl-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModCtrl}:
		return "<ctrl-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModCtrl}:
		return "<ctrl-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModCtrl}:
		return "<ctrl-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModCtrl}:
		return "<ctrl-esc>"
	case KeyComb{Key: KeySpace, Mod: ModCtrl}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModCtrl}:
		return "<ctrl-space>"

	// ctrl+shift
	case KeyComb{Key: KeyF1, Mod: ModCtrlShift}:
		return "<ctrl-shift-f1>"
	case KeyComb{Key: KeyF2, Mod: ModCtrlShift}:
		return "<ctrl-shift-f2>"
	case KeyComb{Key: KeyF3, Mod: ModCtrlShift}:
		return "<ctrl-shift-f3>"
	case KeyComb{Key: KeyF4, Mod: ModCtrlShift}:
		return "<ctrl-shift-f4>"
	case KeyComb{Key: KeyF5, Mod: ModCtrlShift}:
		return "<ctrl-shift-f5>"
	case KeyComb{Key: KeyF6, Mod: ModCtrlShift}:
		return "<ctrl-shift-f6>"
	case KeyComb{Key: KeyF7, Mod: ModCtrlShift}:
		return "<ctrl-shift-f7>"
	case KeyComb{Key: KeyF8, Mod: ModCtrlShift}:
		return "<ctrl-shift-f8>"
	case KeyComb{Key: KeyF9, Mod: ModCtrlShift}:
		return "<ctrl-shift-f9>"
	case KeyComb{Key: KeyF10, Mod: ModCtrlShift}:
		return "<ctrl-shift-f10>"
	case KeyComb{Key: KeyF11, Mod: ModCtrlShift}:
		return "<ctrl-shift-f11>"
	case KeyComb{Key: KeyF12, Mod: ModCtrlShift}:
		return "<ctrl-shift-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModCtrlShift}:
		return "<ctrl-shift-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModCtrlShift}:
		return "<ctrl-shift-delete>"
	case KeyComb{Key: KeyHome, Mod: ModCtrlShift}:
		return "<ctrl-shift-home>"
	case KeyComb{Key: KeyEnd, Mod: ModCtrlShift}:
		return "<ctrl-shift-end>"
	case KeyComb{Key: KeyPgup, Mod: ModCtrlShift}:
		return "<ctrl-shift-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModCtrlShift}:
		return "<ctrl-shift-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModCtrlShift}:
		return "<ctrl-shift-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModCtrlShift}:
		return "<ctrl-shift-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModCtrlShift}:
		return "<ctrl-shift-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModCtrlShift}:
		return "<ctrl-shift-right>"
	case KeyComb{Key: MouseLeft, Mod: ModCtrlShift}:
		return "<ctrl-shift-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModCtrlShift}:
		return "<ctrl-shift-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModCtrlShift}:
		return "<ctrl-shift-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModCtrlShift}:
		return "<ctrl-shift-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModCtrlShift}:
		return "<ctrl-shift-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModCtrlShift}:
		return "<ctrl-shift-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModCtrlShift}:
		return "<ctrl-shift-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModCtrlShift}:
		return "<ctrl-shift-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModCtrlShift}:
		return "<ctrl-shift-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModCtrlShift}:
		return "<ctrl-shift-esc>"
	case KeyComb{Key: KeySpace, Mod: ModCtrlShift}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModCtrlShift}:
		return "<ctrl-shift-space>"

	// ctrl+alt
	case KeyComb{Key: KeyF1, Mod: ModCtrlAlt}:
		return "<ctrl-alt-f1>"
	case KeyComb{Key: KeyF2, Mod: ModCtrlAlt}:
		return "<ctrl-alt-f2>"
	case KeyComb{Key: KeyF3, Mod: ModCtrlAlt}:
		return "<ctrl-alt-f3>"
	case KeyComb{Key: KeyF4, Mod: ModCtrlAlt}:
		return "<ctrl-alt-f4>"
	case KeyComb{Key: KeyF5, Mod: ModCtrlAlt}:
		return "<ctrl-alt-f5>"
	case KeyComb{Key: KeyF6, Mod: ModCtrlAlt}:
		return "<ctrl-alt-f6>"
	case KeyComb{Key: KeyF7, Mod: ModCtrlAlt}:
		return "<ctrl-alt-f7>"
	case KeyComb{Key: KeyF8, Mod: ModCtrlAlt}:
		return "<ctrl-alt-f8>"
	case KeyComb{Key: KeyF9, Mod: ModCtrlAlt}:
		return "<ctrl-alt-f9>"
	case KeyComb{Key: KeyF10, Mod: ModCtrlAlt}:
		return "<ctrl-alt-f10>"
	case KeyComb{Key: KeyF11, Mod: ModCtrlAlt}:
		return "<ctrl-alt-f11>"
	case KeyComb{Key: KeyF12, Mod: ModCtrlAlt}:
		return "<ctrl-alt-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModCtrlAlt}:
		return "<ctrl-alt-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModCtrlAlt}:
		return "<ctrl-alt-delete>"
	case KeyComb{Key: KeyHome, Mod: ModCtrlAlt}:
		return "<ctrl-alt-home>"
	case KeyComb{Key: KeyEnd, Mod: ModCtrlAlt}:
		return "<ctrl-alt-end>"
	case KeyComb{Key: KeyPgup, Mod: ModCtrlAlt}:
		return "<ctrl-alt-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModCtrlAlt}:
		return "<ctrl-alt-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModCtrlAlt}:
		return "<ctrl-alt-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModCtrlAlt}:
		return "<ctrl-alt-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModCtrlAlt}:
		return "<ctrl-alt-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModCtrlAlt}:
		return "<ctrl-alt-right>"
	case KeyComb{Key: MouseLeft, Mod: ModCtrlAlt}:
		return "<ctrl-alt-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModCtrlAlt}:
		return "<ctrl-alt-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModCtrlAlt}:
		return "<ctrl-alt-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModCtrlAlt}:
		return "<ctrl-alt-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModCtrlAlt}:
		return "<ctrl-alt-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModCtrlAlt}:
		return "<ctrl-alt-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModCtrlAlt}:
		return "<ctrl-alt-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModCtrlAlt}:
		return "<ctrl-alt-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModCtrlAlt}:
		return "<ctrl-alt-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModCtrlAlt}:
		return "<ctrl-alt-esc>"
	case KeyComb{Key: KeySpace, Mod: ModCtrlAlt}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModCtrlAlt}:
		return "<ctrl-alt-space>"

	// ctrl+meta
	case KeyComb{Key: KeyF1, Mod: ModCtrlMeta}:
		return "<ctrl-meta-f1>"
	case KeyComb{Key: KeyF2, Mod: ModCtrlMeta}:
		return "<ctrl-meta-f2>"
	case KeyComb{Key: KeyF3, Mod: ModCtrlMeta}:
		return "<ctrl-meta-f3>"
	case KeyComb{Key: KeyF4, Mod: ModCtrlMeta}:
		return "<ctrl-meta-f4>"
	case KeyComb{Key: KeyF5, Mod: ModCtrlMeta}:
		return "<ctrl-meta-f5>"
	case KeyComb{Key: KeyF6, Mod: ModCtrlMeta}:
		return "<ctrl-meta-f6>"
	case KeyComb{Key: KeyF7, Mod: ModCtrlMeta}:
		return "<ctrl-meta-f7>"
	case KeyComb{Key: KeyF8, Mod: ModCtrlMeta}:
		return "<ctrl-meta-f8>"
	case KeyComb{Key: KeyF9, Mod: ModCtrlMeta}:
		return "<ctrl-meta-f9>"
	case KeyComb{Key: KeyF10, Mod: ModCtrlMeta}:
		return "<ctrl-meta-f10>"
	case KeyComb{Key: KeyF11, Mod: ModCtrlMeta}:
		return "<ctrl-meta-f11>"
	case KeyComb{Key: KeyF12, Mod: ModCtrlMeta}:
		return "<ctrl-meta-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModCtrlMeta}:
		return "<ctrl-meta-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModCtrlMeta}:
		return "<ctrl-meta-delete>"
	case KeyComb{Key: KeyHome, Mod: ModCtrlMeta}:
		return "<ctrl-meta-home>"
	case KeyComb{Key: KeyEnd, Mod: ModCtrlMeta}:
		return "<ctrl-meta-end>"
	case KeyComb{Key: KeyPgup, Mod: ModCtrlMeta}:
		return "<ctrl-meta-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModCtrlMeta}:
		return "<ctrl-meta-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModCtrlMeta}:
		return "<ctrl-meta-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModCtrlMeta}:
		return "<ctrl-meta-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModCtrlMeta}:
		return "<ctrl-meta-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModCtrlMeta}:
		return "<ctrl-meta-right>"
	case KeyComb{Key: MouseLeft, Mod: ModCtrlMeta}:
		return "<ctrl-meta-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModCtrlMeta}:
		return "<ctrl-meta-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModCtrlMeta}:
		return "<ctrl-meta-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModCtrlMeta}:
		return "<ctrl-meta-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModCtrlMeta}:
		return "<ctrl-meta-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModCtrlMeta}:
		return "<ctrl-meta-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModCtrlMeta}:
		return "<ctrl-meta-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModCtrlMeta}:
		return "<ctrl-meta-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModCtrlMeta}:
		return "<ctrl-meta-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModCtrlMeta}:
		return "<ctrl-meta-esc>"
	case KeyComb{Key: KeySpace, Mod: ModCtrlMeta}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModCtrlMeta}:
		return "<ctrl-meta-space>"

	// ctrl+shift+alt
	case KeyComb{Key: KeyF1, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-f1>"
	case KeyComb{Key: KeyF2, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-f2>"
	case KeyComb{Key: KeyF3, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-f3>"
	case KeyComb{Key: KeyF4, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-f4>"
	case KeyComb{Key: KeyF5, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-f5>"
	case KeyComb{Key: KeyF6, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-f6>"
	case KeyComb{Key: KeyF7, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-f7>"
	case KeyComb{Key: KeyF8, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-f8>"
	case KeyComb{Key: KeyF9, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-f9>"
	case KeyComb{Key: KeyF10, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-f10>"
	case KeyComb{Key: KeyF11, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-f11>"
	case KeyComb{Key: KeyF12, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-delete>"
	case KeyComb{Key: KeyHome, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-home>"
	case KeyComb{Key: KeyEnd, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-end>"
	case KeyComb{Key: KeyPgup, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-right>"
	case KeyComb{Key: MouseLeft, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-esc>"
	case KeyComb{Key: KeySpace, Mod: ModCtrlShiftAlt}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt-space>"

	// ctrl+shift+meta
	case KeyComb{Key: KeyF1, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-f1>"
	case KeyComb{Key: KeyF2, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-f2>"
	case KeyComb{Key: KeyF3, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-f3>"
	case KeyComb{Key: KeyF4, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-f4>"
	case KeyComb{Key: KeyF5, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-f5>"
	case KeyComb{Key: KeyF6, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-f6>"
	case KeyComb{Key: KeyF7, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-f7>"
	case KeyComb{Key: KeyF8, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-f8>"
	case KeyComb{Key: KeyF9, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-f9>"
	case KeyComb{Key: KeyF10, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-f10>"
	case KeyComb{Key: KeyF11, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-f11>"
	case KeyComb{Key: KeyF12, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-delete>"
	case KeyComb{Key: KeyHome, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-home>"
	case KeyComb{Key: KeyEnd, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-end>"
	case KeyComb{Key: KeyPgup, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-right>"
	case KeyComb{Key: MouseLeft, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-esc>"
	case KeyComb{Key: KeySpace, Mod: ModCtrlShiftMeta}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta-space>"

	// ctrl+alt+meta
	case KeyComb{Key: KeyF1, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-f1>"
	case KeyComb{Key: KeyF2, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-f2>"
	case KeyComb{Key: KeyF3, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-f3>"
	case KeyComb{Key: KeyF4, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-f4>"
	case KeyComb{Key: KeyF5, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-f5>"
	case KeyComb{Key: KeyF6, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-f6>"
	case KeyComb{Key: KeyF7, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-f7>"
	case KeyComb{Key: KeyF8, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-f8>"
	case KeyComb{Key: KeyF9, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-f9>"
	case KeyComb{Key: KeyF10, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-f10>"
	case KeyComb{Key: KeyF11, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-f11>"
	case KeyComb{Key: KeyF12, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-delete>"
	case KeyComb{Key: KeyHome, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-home>"
	case KeyComb{Key: KeyEnd, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-end>"
	case KeyComb{Key: KeyPgup, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-right>"
	case KeyComb{Key: MouseLeft, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-esc>"
	case KeyComb{Key: KeySpace, Mod: ModCtrlAltMeta}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta-space>"

	// shift+meta
	case KeyComb{Key: KeyF1, Mod: ModShiftMeta}:
		return "<shift-meta-f1>"
	case KeyComb{Key: KeyF2, Mod: ModShiftMeta}:
		return "<shift-meta-f2>"
	case KeyComb{Key: KeyF3, Mod: ModShiftMeta}:
		return "<shift-meta-f3>"
	case KeyComb{Key: KeyF4, Mod: ModShiftMeta}:
		return "<shift-meta-f4>"
	case KeyComb{Key: KeyF5, Mod: ModShiftMeta}:
		return "<shift-meta-f5>"
	case KeyComb{Key: KeyF6, Mod: ModShiftMeta}:
		return "<shift-meta-f6>"
	case KeyComb{Key: KeyF7, Mod: ModShiftMeta}:
		return "<shift-meta-f7>"
	case KeyComb{Key: KeyF8, Mod: ModShiftMeta}:
		return "<shift-meta-f8>"
	case KeyComb{Key: KeyF9, Mod: ModShiftMeta}:
		return "<shift-meta-f9>"
	case KeyComb{Key: KeyF10, Mod: ModShiftMeta}:
		return "<shift-meta-f10>"
	case KeyComb{Key: KeyF11, Mod: ModShiftMeta}:
		return "<shift-meta-f11>"
	case KeyComb{Key: KeyF12, Mod: ModShiftMeta}:
		return "<shift-meta-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModShiftMeta}:
		return "<shift-meta-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModShiftMeta}:
		return "<shift-meta-delete>"
	case KeyComb{Key: KeyHome, Mod: ModShiftMeta}:
		return "<shift-meta-home>"
	case KeyComb{Key: KeyEnd, Mod: ModShiftMeta}:
		return "<shift-meta-end>"
	case KeyComb{Key: KeyPgup, Mod: ModShiftMeta}:
		return "<shift-meta-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModShiftMeta}:
		return "<shift-meta-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModShiftMeta}:
		return "<shift-meta-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModShiftMeta}:
		return "<shift-meta-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModShiftMeta}:
		return "<shift-meta-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModShiftMeta}:
		return "<shift-meta-right>"
	case KeyComb{Key: MouseLeft, Mod: ModShiftMeta}:
		return "<shift-meta-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModShiftMeta}:
		return "<shift-meta-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModShiftMeta}:
		return "<shift-meta-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModShiftMeta}:
		return "<shift-meta-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModShiftMeta}:
		return "<shift-meta-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModShiftMeta}:
		return "<shift-meta-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModShiftMeta}:
		return "<shift-meta-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModShiftMeta}:
		return "<shift-meta-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModShiftMeta}:
		return "<shift-meta-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModShiftMeta}:
		return "<shift-meta-esc>"
	case KeyComb{Key: KeySpace, Mod: ModShiftMeta}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModShiftMeta}:
		return "<shift-meta-space>"

	// alt+meta
	case KeyComb{Key: KeyF1, Mod: ModAltMeta}:
		return "<alt-meta-f1>"
	case KeyComb{Key: KeyF2, Mod: ModAltMeta}:
		return "<alt-meta-f2>"
	case KeyComb{Key: KeyF3, Mod: ModAltMeta}:
		return "<alt-meta-f3>"
	case KeyComb{Key: KeyF4, Mod: ModAltMeta}:
		return "<alt-meta-f4>"
	case KeyComb{Key: KeyF5, Mod: ModAltMeta}:
		return "<alt-meta-f5>"
	case KeyComb{Key: KeyF6, Mod: ModAltMeta}:
		return "<alt-meta-f6>"
	case KeyComb{Key: KeyF7, Mod: ModAltMeta}:
		return "<alt-meta-f7>"
	case KeyComb{Key: KeyF8, Mod: ModAltMeta}:
		return "<alt-meta-f8>"
	case KeyComb{Key: KeyF9, Mod: ModAltMeta}:
		return "<alt-meta-f9>"
	case KeyComb{Key: KeyF10, Mod: ModAltMeta}:
		return "<alt-meta-f10>"
	case KeyComb{Key: KeyF11, Mod: ModAltMeta}:
		return "<alt-meta-f11>"
	case KeyComb{Key: KeyF12, Mod: ModAltMeta}:
		return "<alt-meta-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModAltMeta}:
		return "<alt-meta-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModAltMeta}:
		return "<alt-meta-delete>"
	case KeyComb{Key: KeyHome, Mod: ModAltMeta}:
		return "<alt-meta-home>"
	case KeyComb{Key: KeyEnd, Mod: ModAltMeta}:
		return "<alt-meta-end>"
	case KeyComb{Key: KeyPgup, Mod: ModAltMeta}:
		return "<alt-meta-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModAltMeta}:
		return "<alt-meta-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModAltMeta}:
		return "<alt-meta-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModAltMeta}:
		return "<alt-meta-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModAltMeta}:
		return "<alt-meta-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModAltMeta}:
		return "<alt-meta-right>"
	case KeyComb{Key: MouseLeft, Mod: ModAltMeta}:
		return "<alt-meta-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModAltMeta}:
		return "<alt-meta-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModAltMeta}:
		return "<alt-meta-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModAltMeta}:
		return "<alt-meta-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModAltMeta}:
		return "<alt-meta-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModAltMeta}:
		return "<alt-meta-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModAltMeta}:
		return "<alt-meta-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModAltMeta}:
		return "<alt-meta-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModAltMeta}:
		return "<alt-meta-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModAltMeta}:
		return "<alt-meta-esc>"
	case KeyComb{Key: KeySpace, Mod: ModAltMeta}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModAltMeta}:
		return "<alt-meta-space>"

	// alt+shift+meta
	case KeyComb{Key: KeyF1, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-f1>"
	case KeyComb{Key: KeyF2, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-f2>"
	case KeyComb{Key: KeyF3, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-f3>"
	case KeyComb{Key: KeyF4, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-f4>"
	case KeyComb{Key: KeyF5, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-f5>"
	case KeyComb{Key: KeyF6, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-f6>"
	case KeyComb{Key: KeyF7, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-f7>"
	case KeyComb{Key: KeyF8, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-f8>"
	case KeyComb{Key: KeyF9, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-f9>"
	case KeyComb{Key: KeyF10, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-f10>"
	case KeyComb{Key: KeyF11, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-f11>"
	case KeyComb{Key: KeyF12, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-delete>"
	case KeyComb{Key: KeyHome, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-home>"
	case KeyComb{Key: KeyEnd, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-end>"
	case KeyComb{Key: KeyPgup, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-right>"
	case KeyComb{Key: MouseLeft, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-esc>"
	case KeyComb{Key: KeySpace, Mod: ModAltShiftMeta}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModAltShiftMeta}:
		return "<alt-shift-meta-space>"

	// alt+shift
	case KeyComb{Key: KeyF1, Mod: ModAltShift}:
		return "<alt-shift-f1>"
	case KeyComb{Key: KeyF2, Mod: ModAltShift}:
		return "<alt-shift-f2>"
	case KeyComb{Key: KeyF3, Mod: ModAltShift}:
		return "<alt-shift-f3>"
	case KeyComb{Key: KeyF4, Mod: ModAltShift}:
		return "<alt-shift-f4>"
	case KeyComb{Key: KeyF5, Mod: ModAltShift}:
		return "<alt-shift-f5>"
	case KeyComb{Key: KeyF6, Mod: ModAltShift}:
		return "<alt-shift-f6>"
	case KeyComb{Key: KeyF7, Mod: ModAltShift}:
		return "<alt-shift-f7>"
	case KeyComb{Key: KeyF8, Mod: ModAltShift}:
		return "<alt-shift-f8>"
	case KeyComb{Key: KeyF9, Mod: ModAltShift}:
		return "<alt-shift-f9>"
	case KeyComb{Key: KeyF10, Mod: ModAltShift}:
		return "<alt-shift-f10>"
	case KeyComb{Key: KeyF11, Mod: ModAltShift}:
		return "<alt-shift-f11>"
	case KeyComb{Key: KeyF12, Mod: ModAltShift}:
		return "<alt-shift-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModAltShift}:
		return "<alt-shift-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModAltShift}:
		return "<alt-shift-delete>"
	case KeyComb{Key: KeyHome, Mod: ModAltShift}:
		return "<alt-shift-home>"
	case KeyComb{Key: KeyEnd, Mod: ModAltShift}:
		return "<alt-shift-end>"
	case KeyComb{Key: KeyPgup, Mod: ModAltShift}:
		return "<alt-shift-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModAltShift}:
		return "<alt-shift-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModAltShift}:
		return "<alt-shift-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModAltShift}:
		return "<alt-shift-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModAltShift}:
		return "<alt-shift-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModAltShift}:
		return "<alt-shift-right>"
	case KeyComb{Key: MouseLeft, Mod: ModAltShift}:
		return "<alt-shift-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModAltShift}:
		return "<alt-shift-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModAltShift}:
		return "<alt-shift-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModAltShift}:
		return "<alt-shift-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModAltShift}:
		return "<alt-shift-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModAltShift}:
		return "<alt-shift-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModAltShift}:
		return "<alt-shift-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModAltShift}:
		return "<alt-shift-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModAltShift}:
		return "<alt-shift-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModAltShift}:
		return "<alt-shift-esc>"
	case KeyComb{Key: KeySpace, Mod: ModAltShift}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModAltShift}:
		return "<alt-shift-space>"

	// modifiers only
	case KeyComb{Mod: ModAlt}:
		return "<alt>"
	case KeyComb{Mod: ModShift}:
		return "<shift>"
	case KeyComb{Mod: ModMeta}:
		return "<meta>"
	case KeyComb{Mod: ModCtrl}:
		return "<ctrl>"
	case KeyComb{Mod: ModCtrlShift}:
		return "<ctrl-shift>"
	case KeyComb{Mod: ModCtrlAlt}:
		return "<ctrl-alt>"
	case KeyComb{Mod: ModCtrlMeta}:
		return "<ctrl-meta>"
	case KeyComb{Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt>"
	case KeyComb{Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta>"
	case KeyComb{Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta>"
	case KeyComb{Mod: ModShiftMeta}:
		return "<shift-meta>"
	case KeyComb{Mod: ModAltMeta}:
		return "<alt-meta>"
	case KeyComb{Mod: ModAltShiftMeta}:
		return "<alt-shift-meta>"
	case KeyComb{Mod: ModAltShift}:
		return "<alt-shift>"

	default:
		return "<INVALID>"
	}
}

// ShortString returns the shorthand of the long string representation
// of this KeyComb.
func (k KeyComb) ShortString() string {
	if k.Ch != 0 && k.Key != KeySpace {
		ch, mod := processCharacter(k.Ch, k.Mod)
		// if Ch is set, then ignore key for representing this key comb
		switch mod {
		case ModMeta:
			return fmt.Sprintf("<m-%s>", ch)
		case ModAlt:
			return fmt.Sprintf("<a-%s>", ch)
		case ModShift:
			return fmt.Sprintf("<s-%s>", ch)
		case ModCtrl:
			return fmt.Sprintf("<c-%s>", ch)
		case ModCtrlShift:
			return fmt.Sprintf("<c-s-%s>", ch)
		case ModCtrlAlt:
			return fmt.Sprintf("<c-a-%s>", ch)
		case ModCtrlMeta:
			return fmt.Sprintf("<c-m-%s>", ch)
		case ModCtrlShiftAlt:
			return fmt.Sprintf("<c-s-a-%s>", ch)
		case ModCtrlShiftMeta:
			return fmt.Sprintf("<c-s-m-%s>", ch)
		case ModCtrlAltMeta:
			return fmt.Sprintf("<c-a-m-%s>", ch)
		case ModShiftMeta:
			return fmt.Sprintf("<s-m-%s>", ch)
		case ModAltMeta:
			return fmt.Sprintf("<a-m-%s>", ch)
		case ModAltShiftMeta:
			return fmt.Sprintf("<a-s-m-%s>", ch)
		case ModAltShift:
			return fmt.Sprintf("<a-s-%s>", ch)
		default:
			return string(ch)
		}
	}

	switch k {
	case KeyComb{Key: KeyF1}:
		return "<f1>"
	case KeyComb{Key: KeyF2}:
		return "<f2>"
	case KeyComb{Key: KeyF3}:
		return "<f3>"
	case KeyComb{Key: KeyF4}:
		return "<f4>"
	case KeyComb{Key: KeyF5}:
		return "<f5>"
	case KeyComb{Key: KeyF6}:
		return "<f6>"
	case KeyComb{Key: KeyF7}:
		return "<f7>"
	case KeyComb{Key: KeyF8}:
		return "<f8>"
	case KeyComb{Key: KeyF9}:
		return "<f9>"
	case KeyComb{Key: KeyF10}:
		return "<f10>"
	case KeyComb{Key: KeyF11}:
		return "<f11>"
	case KeyComb{Key: KeyF12}:
		return "<f12>"
	case KeyComb{Key: KeyInsert}:
		return "<insert>"
	case KeyComb{Key: KeyDelete}:
		return "<delete>"
	case KeyComb{Key: KeyHome}:
		return "<home>"
	case KeyComb{Key: KeyEnd}:
		return "<end>"
	case KeyComb{Key: KeyPgup}:
		return "<pgup>"
	case KeyComb{Key: KeyPgdn}:
		return "<pgdn>"
	case KeyComb{Key: KeyArrowUp}:
		return "<up>"
	case KeyComb{Key: KeyArrowDown}:
		return "<down>"
	case KeyComb{Key: KeyArrowLeft}:
		return "<left>"
	case KeyComb{Key: KeyArrowRight}:
		return "<right>"
	case KeyComb{Key: MouseLeft}:
		return "<mouse-left>"
	case KeyComb{Key: MouseMiddle}:
		return "<mouse-middle>"
	case KeyComb{Key: MouseRight}:
		return "<mouse-right>"
	case KeyComb{Key: MouseRelease}:
		return "<mouse-release>"
	case KeyComb{Key: MouseWheelUp}:
		return "<mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown}:
		return "<mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace}:
		return "<backspace>"
	case KeyComb{Key: KeyTab}:
		return "<tab>"
	case KeyComb{Key: KeyEnter}:
		return "<enter>"
	case KeyComb{Key: KeyEsc}:
		return "<esc>"
	case KeyComb{Key: KeySpace, Ch: ' '}, KeyComb{Key: KeySpace}:
		return "<space>"
	case KeyComb{Key: KeyCapsLock}:
		return "<capslock>"
	case KeyComb{Key: KeyNumLock}:
		return "<numlock>"
	case KeyComb{Key: KeyScrollLock}:
		return "<scrolllock>"
	case KeyComb{Key: KeyMenu}:
		return "<menu>"

	// meta
	case KeyComb{Key: KeyF1, Mod: ModMeta}:
		return "<m-f1>"
	case KeyComb{Key: KeyF2, Mod: ModMeta}:
		return "<m-f2>"
	case KeyComb{Key: KeyF3, Mod: ModMeta}:
		return "<m-f3>"
	case KeyComb{Key: KeyF4, Mod: ModMeta}:
		return "<m-f4>"
	case KeyComb{Key: KeyF5, Mod: ModMeta}:
		return "<m-f5>"
	case KeyComb{Key: KeyF6, Mod: ModMeta}:
		return "<m-f6>"
	case KeyComb{Key: KeyF7, Mod: ModMeta}:
		return "<m-f7>"
	case KeyComb{Key: KeyF8, Mod: ModMeta}:
		return "<m-f8>"
	case KeyComb{Key: KeyF9, Mod: ModMeta}:
		return "<m-f9>"
	case KeyComb{Key: KeyF10, Mod: ModMeta}:
		return "<m-f10>"
	case KeyComb{Key: KeyF11, Mod: ModMeta}:
		return "<m-f11>"
	case KeyComb{Key: KeyF12, Mod: ModMeta}:
		return "<m-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModMeta}:
		return "<m-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModMeta}:
		return "<m-delete>"
	case KeyComb{Key: KeyHome, Mod: ModMeta}:
		return "<m-home>"
	case KeyComb{Key: KeyEnd, Mod: ModMeta}:
		return "<m-end>"
	case KeyComb{Key: KeyPgup, Mod: ModMeta}:
		return "<m-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModMeta}:
		return "<m-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModMeta}:
		return "<m-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModMeta}:
		return "<m-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModMeta}:
		return "<m-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModMeta}:
		return "<m-right>"
	case KeyComb{Key: MouseLeft, Mod: ModMeta}:
		return "<m-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModMeta}:
		return "<m-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModMeta}:
		return "<m-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModMeta}:
		return "<m-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModMeta}:
		return "<m-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModMeta}:
		return "<m-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModMeta}:
		return "<m-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModMeta}:
		return "<m-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModMeta}:
		return "<m-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModMeta}:
		return "<m-esc>"
	case KeyComb{Key: KeySpace, Mod: ModMeta}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModMeta}:
		return "<m-space>"

	// alt
	case KeyComb{Key: KeyF1, Mod: ModAlt}:
		return "<a-f1>"
	case KeyComb{Key: KeyF2, Mod: ModAlt}:
		return "<a-f2>"
	case KeyComb{Key: KeyF3, Mod: ModAlt}:
		return "<a-f3>"
	case KeyComb{Key: KeyF4, Mod: ModAlt}:
		return "<a-f4>"
	case KeyComb{Key: KeyF5, Mod: ModAlt}:
		return "<a-f5>"
	case KeyComb{Key: KeyF6, Mod: ModAlt}:
		return "<a-f6>"
	case KeyComb{Key: KeyF7, Mod: ModAlt}:
		return "<a-f7>"
	case KeyComb{Key: KeyF8, Mod: ModAlt}:
		return "<a-f8>"
	case KeyComb{Key: KeyF9, Mod: ModAlt}:
		return "<a-f9>"
	case KeyComb{Key: KeyF10, Mod: ModAlt}:
		return "<a-f10>"
	case KeyComb{Key: KeyF11, Mod: ModAlt}:
		return "<a-f11>"
	case KeyComb{Key: KeyF12, Mod: ModAlt}:
		return "<a-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModAlt}:
		return "<a-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModAlt}:
		return "<a-delete>"
	case KeyComb{Key: KeyHome, Mod: ModAlt}:
		return "<a-home>"
	case KeyComb{Key: KeyEnd, Mod: ModAlt}:
		return "<a-end>"
	case KeyComb{Key: KeyPgup, Mod: ModAlt}:
		return "<a-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModAlt}:
		return "<a-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModAlt}:
		return "<a-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModAlt}:
		return "<a-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModAlt}:
		return "<a-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModAlt}:
		return "<a-right>"
	case KeyComb{Key: MouseLeft, Mod: ModAlt}:
		return "<a-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModAlt}:
		return "<a-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModAlt}:
		return "<a-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModAlt}:
		return "<a-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModAlt}:
		return "<a-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModAlt}:
		return "<a-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModAlt}:
		return "<a-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModAlt}:
		return "<a-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModAlt}:
		return "<a-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModAlt}:
		return "<a-esc>"
	case KeyComb{Key: KeySpace, Mod: ModAlt}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModAlt}:
		return "<a-space>"

	// shift
	case KeyComb{Key: KeyF1, Mod: ModShift}:
		return "<s-f1>"
	case KeyComb{Key: KeyF2, Mod: ModShift}:
		return "<s-f2>"
	case KeyComb{Key: KeyF3, Mod: ModShift}:
		return "<s-f3>"
	case KeyComb{Key: KeyF4, Mod: ModShift}:
		return "<s-f4>"
	case KeyComb{Key: KeyF5, Mod: ModShift}:
		return "<s-f5>"
	case KeyComb{Key: KeyF6, Mod: ModShift}:
		return "<s-f6>"
	case KeyComb{Key: KeyF7, Mod: ModShift}:
		return "<s-f7>"
	case KeyComb{Key: KeyF8, Mod: ModShift}:
		return "<s-f8>"
	case KeyComb{Key: KeyF9, Mod: ModShift}:
		return "<s-f9>"
	case KeyComb{Key: KeyF10, Mod: ModShift}:
		return "<s-f10>"
	case KeyComb{Key: KeyF11, Mod: ModShift}:
		return "<s-f11>"
	case KeyComb{Key: KeyF12, Mod: ModShift}:
		return "<s-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModShift}:
		return "<s-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModShift}:
		return "<s-delete>"
	case KeyComb{Key: KeyHome, Mod: ModShift}:
		return "<s-home>"
	case KeyComb{Key: KeyEnd, Mod: ModShift}:
		return "<s-end>"
	case KeyComb{Key: KeyPgup, Mod: ModShift}:
		return "<s-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModShift}:
		return "<s-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModShift}:
		return "<s-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModShift}:
		return "<s-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModShift}:
		return "<s-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModShift}:
		return "<s-right>"
	case KeyComb{Key: MouseLeft, Mod: ModShift}:
		return "<s-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModShift}:
		return "<s-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModShift}:
		return "<s-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModShift}:
		return "<s-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModShift}:
		return "<s-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModShift}:
		return "<s-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModShift}:
		return "<s-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModShift}:
		return "<s-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModShift}:
		return "<s-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModShift}:
		return "<s-esc>"
	case KeyComb{Key: KeySpace, Mod: ModShift}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModShift}:
		return "<s-space>"

	// ctrl
	case KeyComb{Key: KeyF1, Mod: ModCtrl}:
		return "<c-f1>"
	case KeyComb{Key: KeyF2, Mod: ModCtrl}:
		return "<c-f2>"
	case KeyComb{Key: KeyF3, Mod: ModCtrl}:
		return "<c-f3>"
	case KeyComb{Key: KeyF4, Mod: ModCtrl}:
		return "<c-f4>"
	case KeyComb{Key: KeyF5, Mod: ModCtrl}:
		return "<c-f5>"
	case KeyComb{Key: KeyF6, Mod: ModCtrl}:
		return "<c-f6>"
	case KeyComb{Key: KeyF7, Mod: ModCtrl}:
		return "<c-f7>"
	case KeyComb{Key: KeyF8, Mod: ModCtrl}:
		return "<c-f8>"
	case KeyComb{Key: KeyF9, Mod: ModCtrl}:
		return "<c-f9>"
	case KeyComb{Key: KeyF10, Mod: ModCtrl}:
		return "<c-f10>"
	case KeyComb{Key: KeyF11, Mod: ModCtrl}:
		return "<c-f11>"
	case KeyComb{Key: KeyF12, Mod: ModCtrl}:
		return "<c-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModCtrl}:
		return "<c-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModCtrl}:
		return "<c-delete>"
	case KeyComb{Key: KeyHome, Mod: ModCtrl}:
		return "<c-home>"
	case KeyComb{Key: KeyEnd, Mod: ModCtrl}:
		return "<c-end>"
	case KeyComb{Key: KeyPgup, Mod: ModCtrl}:
		return "<c-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModCtrl}:
		return "<c-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModCtrl}:
		return "<c-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModCtrl}:
		return "<c-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModCtrl}:
		return "<c-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModCtrl}:
		return "<c-right>"
	case KeyComb{Key: MouseLeft, Mod: ModCtrl}:
		return "<c-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModCtrl}:
		return "<c-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModCtrl}:
		return "<c-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModCtrl}:
		return "<c-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModCtrl}:
		return "<c-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModCtrl}:
		return "<c-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModCtrl}:
		return "<c-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModCtrl}:
		return "<c-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModCtrl}:
		return "<c-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModCtrl}:
		return "<c-esc>"
	case KeyComb{Key: KeySpace, Mod: ModCtrl}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModCtrl}:
		return "<c-space>"

	// ctrl+shift
	case KeyComb{Key: KeyF1, Mod: ModCtrlShift}:
		return "<c-s-f1>"
	case KeyComb{Key: KeyF2, Mod: ModCtrlShift}:
		return "<c-s-f2>"
	case KeyComb{Key: KeyF3, Mod: ModCtrlShift}:
		return "<c-s-f3>"
	case KeyComb{Key: KeyF4, Mod: ModCtrlShift}:
		return "<c-s-f4>"
	case KeyComb{Key: KeyF5, Mod: ModCtrlShift}:
		return "<c-s-f5>"
	case KeyComb{Key: KeyF6, Mod: ModCtrlShift}:
		return "<c-s-f6>"
	case KeyComb{Key: KeyF7, Mod: ModCtrlShift}:
		return "<c-s-f7>"
	case KeyComb{Key: KeyF8, Mod: ModCtrlShift}:
		return "<c-s-f8>"
	case KeyComb{Key: KeyF9, Mod: ModCtrlShift}:
		return "<c-s-f9>"
	case KeyComb{Key: KeyF10, Mod: ModCtrlShift}:
		return "<c-s-f10>"
	case KeyComb{Key: KeyF11, Mod: ModCtrlShift}:
		return "<c-s-f11>"
	case KeyComb{Key: KeyF12, Mod: ModCtrlShift}:
		return "<c-s-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModCtrlShift}:
		return "<c-s-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModCtrlShift}:
		return "<c-s-delete>"
	case KeyComb{Key: KeyHome, Mod: ModCtrlShift}:
		return "<c-s-home>"
	case KeyComb{Key: KeyEnd, Mod: ModCtrlShift}:
		return "<c-s-end>"
	case KeyComb{Key: KeyPgup, Mod: ModCtrlShift}:
		return "<c-s-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModCtrlShift}:
		return "<c-s-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModCtrlShift}:
		return "<c-s-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModCtrlShift}:
		return "<c-s-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModCtrlShift}:
		return "<c-s-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModCtrlShift}:
		return "<c-s-right>"
	case KeyComb{Key: MouseLeft, Mod: ModCtrlShift}:
		return "<c-s-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModCtrlShift}:
		return "<c-s-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModCtrlShift}:
		return "<c-s-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModCtrlShift}:
		return "<c-s-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModCtrlShift}:
		return "<c-s-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModCtrlShift}:
		return "<c-s-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModCtrlShift}:
		return "<c-s-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModCtrlShift}:
		return "<c-s-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModCtrlShift}:
		return "<c-s-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModCtrlShift}:
		return "<c-s-esc>"
	case KeyComb{Key: KeySpace, Mod: ModCtrlShift}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModCtrlShift}:
		return "<c-s-space>"

	// ctrl+alt
	case KeyComb{Key: KeyF1, Mod: ModCtrlAlt}:
		return "<c-a-f1>"
	case KeyComb{Key: KeyF2, Mod: ModCtrlAlt}:
		return "<c-a-f2>"
	case KeyComb{Key: KeyF3, Mod: ModCtrlAlt}:
		return "<c-a-f3>"
	case KeyComb{Key: KeyF4, Mod: ModCtrlAlt}:
		return "<c-a-f4>"
	case KeyComb{Key: KeyF5, Mod: ModCtrlAlt}:
		return "<c-a-f5>"
	case KeyComb{Key: KeyF6, Mod: ModCtrlAlt}:
		return "<c-a-f6>"
	case KeyComb{Key: KeyF7, Mod: ModCtrlAlt}:
		return "<c-a-f7>"
	case KeyComb{Key: KeyF8, Mod: ModCtrlAlt}:
		return "<c-a-f8>"
	case KeyComb{Key: KeyF9, Mod: ModCtrlAlt}:
		return "<c-a-f9>"
	case KeyComb{Key: KeyF10, Mod: ModCtrlAlt}:
		return "<c-a-f10>"
	case KeyComb{Key: KeyF11, Mod: ModCtrlAlt}:
		return "<c-a-f11>"
	case KeyComb{Key: KeyF12, Mod: ModCtrlAlt}:
		return "<c-a-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModCtrlAlt}:
		return "<c-a-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModCtrlAlt}:
		return "<c-a-delete>"
	case KeyComb{Key: KeyHome, Mod: ModCtrlAlt}:
		return "<c-a-home>"
	case KeyComb{Key: KeyEnd, Mod: ModCtrlAlt}:
		return "<c-a-end>"
	case KeyComb{Key: KeyPgup, Mod: ModCtrlAlt}:
		return "<c-a-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModCtrlAlt}:
		return "<c-a-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModCtrlAlt}:
		return "<c-a-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModCtrlAlt}:
		return "<c-a-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModCtrlAlt}:
		return "<c-a-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModCtrlAlt}:
		return "<c-a-right>"
	case KeyComb{Key: MouseLeft, Mod: ModCtrlAlt}:
		return "<c-a-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModCtrlAlt}:
		return "<c-a-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModCtrlAlt}:
		return "<c-a-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModCtrlAlt}:
		return "<c-a-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModCtrlAlt}:
		return "<c-a-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModCtrlAlt}:
		return "<c-a-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModCtrlAlt}:
		return "<c-a-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModCtrlAlt}:
		return "<c-a-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModCtrlAlt}:
		return "<c-a-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModCtrlAlt}:
		return "<c-a-esc>"
	case KeyComb{Key: KeySpace, Mod: ModCtrlAlt}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModCtrlAlt}:
		return "<c-a-space>"

	// ctrl+meta
	case KeyComb{Key: KeyF1, Mod: ModCtrlMeta}:
		return "<c-m-f1>"
	case KeyComb{Key: KeyF2, Mod: ModCtrlMeta}:
		return "<c-m-f2>"
	case KeyComb{Key: KeyF3, Mod: ModCtrlMeta}:
		return "<c-m-f3>"
	case KeyComb{Key: KeyF4, Mod: ModCtrlMeta}:
		return "<c-m-f4>"
	case KeyComb{Key: KeyF5, Mod: ModCtrlMeta}:
		return "<c-m-f5>"
	case KeyComb{Key: KeyF6, Mod: ModCtrlMeta}:
		return "<c-m-f6>"
	case KeyComb{Key: KeyF7, Mod: ModCtrlMeta}:
		return "<c-m-f7>"
	case KeyComb{Key: KeyF8, Mod: ModCtrlMeta}:
		return "<c-m-f8>"
	case KeyComb{Key: KeyF9, Mod: ModCtrlMeta}:
		return "<c-m-f9>"
	case KeyComb{Key: KeyF10, Mod: ModCtrlMeta}:
		return "<c-m-f10>"
	case KeyComb{Key: KeyF11, Mod: ModCtrlMeta}:
		return "<c-m-f11>"
	case KeyComb{Key: KeyF12, Mod: ModCtrlMeta}:
		return "<c-m-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModCtrlMeta}:
		return "<c-m-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModCtrlMeta}:
		return "<c-m-delete>"
	case KeyComb{Key: KeyHome, Mod: ModCtrlMeta}:
		return "<c-m-home>"
	case KeyComb{Key: KeyEnd, Mod: ModCtrlMeta}:
		return "<c-m-end>"
	case KeyComb{Key: KeyPgup, Mod: ModCtrlMeta}:
		return "<c-m-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModCtrlMeta}:
		return "<c-m-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModCtrlMeta}:
		return "<c-m-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModCtrlMeta}:
		return "<c-m-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModCtrlMeta}:
		return "<c-m-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModCtrlMeta}:
		return "<c-m-right>"
	case KeyComb{Key: MouseLeft, Mod: ModCtrlMeta}:
		return "<c-m-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModCtrlMeta}:
		return "<c-m-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModCtrlMeta}:
		return "<c-m-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModCtrlMeta}:
		return "<c-m-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModCtrlMeta}:
		return "<c-m-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModCtrlMeta}:
		return "<c-m-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModCtrlMeta}:
		return "<c-m-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModCtrlMeta}:
		return "<c-m-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModCtrlMeta}:
		return "<c-m-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModCtrlMeta}:
		return "<c-m-esc>"
	case KeyComb{Key: KeySpace, Mod: ModCtrlMeta}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModCtrlMeta}:
		return "<c-m-space>"

	// ctrl+shift+alt
	case KeyComb{Key: KeyF1, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-f1>"
	case KeyComb{Key: KeyF2, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-f2>"
	case KeyComb{Key: KeyF3, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-f3>"
	case KeyComb{Key: KeyF4, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-f4>"
	case KeyComb{Key: KeyF5, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-f5>"
	case KeyComb{Key: KeyF6, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-f6>"
	case KeyComb{Key: KeyF7, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-f7>"
	case KeyComb{Key: KeyF8, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-f8>"
	case KeyComb{Key: KeyF9, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-f9>"
	case KeyComb{Key: KeyF10, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-f10>"
	case KeyComb{Key: KeyF11, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-f11>"
	case KeyComb{Key: KeyF12, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-delete>"
	case KeyComb{Key: KeyHome, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-home>"
	case KeyComb{Key: KeyEnd, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-end>"
	case KeyComb{Key: KeyPgup, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-right>"
	case KeyComb{Key: MouseLeft, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-esc>"
	case KeyComb{Key: KeySpace, Mod: ModCtrlShiftAlt}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModCtrlShiftAlt}:
		return "<c-s-a-space>"

	// ctrl+shift+meta
	case KeyComb{Key: KeyF1, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-f1>"
	case KeyComb{Key: KeyF2, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-f2>"
	case KeyComb{Key: KeyF3, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-f3>"
	case KeyComb{Key: KeyF4, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-f4>"
	case KeyComb{Key: KeyF5, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-f5>"
	case KeyComb{Key: KeyF6, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-f6>"
	case KeyComb{Key: KeyF7, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-f7>"
	case KeyComb{Key: KeyF8, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-f8>"
	case KeyComb{Key: KeyF9, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-f9>"
	case KeyComb{Key: KeyF10, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-f10>"
	case KeyComb{Key: KeyF11, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-f11>"
	case KeyComb{Key: KeyF12, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-delete>"
	case KeyComb{Key: KeyHome, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-home>"
	case KeyComb{Key: KeyEnd, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-end>"
	case KeyComb{Key: KeyPgup, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-right>"
	case KeyComb{Key: MouseLeft, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-esc>"
	case KeyComb{Key: KeySpace, Mod: ModCtrlShiftMeta}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModCtrlShiftMeta}:
		return "<c-s-m-space>"

	// ctrl+alt+meta
	case KeyComb{Key: KeyF1, Mod: ModCtrlAltMeta}:
		return "<c-a-m-f1>"
	case KeyComb{Key: KeyF2, Mod: ModCtrlAltMeta}:
		return "<c-a-m-f2>"
	case KeyComb{Key: KeyF3, Mod: ModCtrlAltMeta}:
		return "<c-a-m-f3>"
	case KeyComb{Key: KeyF4, Mod: ModCtrlAltMeta}:
		return "<c-a-m-f4>"
	case KeyComb{Key: KeyF5, Mod: ModCtrlAltMeta}:
		return "<c-a-m-f5>"
	case KeyComb{Key: KeyF6, Mod: ModCtrlAltMeta}:
		return "<c-a-m-f6>"
	case KeyComb{Key: KeyF7, Mod: ModCtrlAltMeta}:
		return "<c-a-m-f7>"
	case KeyComb{Key: KeyF8, Mod: ModCtrlAltMeta}:
		return "<c-a-m-f8>"
	case KeyComb{Key: KeyF9, Mod: ModCtrlAltMeta}:
		return "<c-a-m-f9>"
	case KeyComb{Key: KeyF10, Mod: ModCtrlAltMeta}:
		return "<c-a-m-f10>"
	case KeyComb{Key: KeyF11, Mod: ModCtrlAltMeta}:
		return "<c-a-m-f11>"
	case KeyComb{Key: KeyF12, Mod: ModCtrlAltMeta}:
		return "<c-a-m-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModCtrlAltMeta}:
		return "<c-a-m-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModCtrlAltMeta}:
		return "<c-a-m-delete>"
	case KeyComb{Key: KeyHome, Mod: ModCtrlAltMeta}:
		return "<c-a-m-home>"
	case KeyComb{Key: KeyEnd, Mod: ModCtrlAltMeta}:
		return "<c-a-m-end>"
	case KeyComb{Key: KeyPgup, Mod: ModCtrlAltMeta}:
		return "<c-a-m-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModCtrlAltMeta}:
		return "<c-a-m-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModCtrlAltMeta}:
		return "<c-a-m-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModCtrlAltMeta}:
		return "<c-a-m-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModCtrlAltMeta}:
		return "<c-a-m-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModCtrlAltMeta}:
		return "<c-a-m-right>"
	case KeyComb{Key: MouseLeft, Mod: ModCtrlAltMeta}:
		return "<c-a-m-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModCtrlAltMeta}:
		return "<c-a-m-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModCtrlAltMeta}:
		return "<c-a-m-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModCtrlAltMeta}:
		return "<c-a-m-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModCtrlAltMeta}:
		return "<c-a-m-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModCtrlAltMeta}:
		return "<c-a-m-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModCtrlAltMeta}:
		return "<c-a-m-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModCtrlAltMeta}:
		return "<c-a-m-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModCtrlAltMeta}:
		return "<c-a-m-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModCtrlAltMeta}:
		return "<c-a-m-esc>"
	case KeyComb{Key: KeySpace, Mod: ModCtrlAltMeta}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModCtrlAltMeta}:
		return "<c-a-m-space>"

	// shift+meta
	case KeyComb{Key: KeyF1, Mod: ModShiftMeta}:
		return "<s-m-f1>"
	case KeyComb{Key: KeyF2, Mod: ModShiftMeta}:
		return "<s-m-f2>"
	case KeyComb{Key: KeyF3, Mod: ModShiftMeta}:
		return "<s-m-f3>"
	case KeyComb{Key: KeyF4, Mod: ModShiftMeta}:
		return "<s-m-f4>"
	case KeyComb{Key: KeyF5, Mod: ModShiftMeta}:
		return "<s-m-f5>"
	case KeyComb{Key: KeyF6, Mod: ModShiftMeta}:
		return "<s-m-f6>"
	case KeyComb{Key: KeyF7, Mod: ModShiftMeta}:
		return "<s-m-f7>"
	case KeyComb{Key: KeyF8, Mod: ModShiftMeta}:
		return "<s-m-f8>"
	case KeyComb{Key: KeyF9, Mod: ModShiftMeta}:
		return "<s-m-f9>"
	case KeyComb{Key: KeyF10, Mod: ModShiftMeta}:
		return "<s-m-f10>"
	case KeyComb{Key: KeyF11, Mod: ModShiftMeta}:
		return "<s-m-f11>"
	case KeyComb{Key: KeyF12, Mod: ModShiftMeta}:
		return "<s-m-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModShiftMeta}:
		return "<s-m-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModShiftMeta}:
		return "<s-m-delete>"
	case KeyComb{Key: KeyHome, Mod: ModShiftMeta}:
		return "<s-m-home>"
	case KeyComb{Key: KeyEnd, Mod: ModShiftMeta}:
		return "<s-m-end>"
	case KeyComb{Key: KeyPgup, Mod: ModShiftMeta}:
		return "<s-m-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModShiftMeta}:
		return "<s-m-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModShiftMeta}:
		return "<s-m-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModShiftMeta}:
		return "<s-m-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModShiftMeta}:
		return "<s-m-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModShiftMeta}:
		return "<s-m-right>"
	case KeyComb{Key: MouseLeft, Mod: ModShiftMeta}:
		return "<s-m-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModShiftMeta}:
		return "<s-m-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModShiftMeta}:
		return "<s-m-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModShiftMeta}:
		return "<s-m-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModShiftMeta}:
		return "<s-m-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModShiftMeta}:
		return "<s-m-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModShiftMeta}:
		return "<s-m-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModShiftMeta}:
		return "<s-m-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModShiftMeta}:
		return "<s-m-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModShiftMeta}:
		return "<s-m-esc>"
	case KeyComb{Key: KeySpace, Mod: ModShiftMeta}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModShiftMeta}:
		return "<s-m-space>"

	// alt+meta
	case KeyComb{Key: KeyF1, Mod: ModAltMeta}:
		return "<a-m-f1>"
	case KeyComb{Key: KeyF2, Mod: ModAltMeta}:
		return "<a-m-f2>"
	case KeyComb{Key: KeyF3, Mod: ModAltMeta}:
		return "<a-m-f3>"
	case KeyComb{Key: KeyF4, Mod: ModAltMeta}:
		return "<a-m-f4>"
	case KeyComb{Key: KeyF5, Mod: ModAltMeta}:
		return "<a-m-f5>"
	case KeyComb{Key: KeyF6, Mod: ModAltMeta}:
		return "<a-m-f6>"
	case KeyComb{Key: KeyF7, Mod: ModAltMeta}:
		return "<a-m-f7>"
	case KeyComb{Key: KeyF8, Mod: ModAltMeta}:
		return "<a-m-f8>"
	case KeyComb{Key: KeyF9, Mod: ModAltMeta}:
		return "<a-m-f9>"
	case KeyComb{Key: KeyF10, Mod: ModAltMeta}:
		return "<a-m-f10>"
	case KeyComb{Key: KeyF11, Mod: ModAltMeta}:
		return "<a-m-f11>"
	case KeyComb{Key: KeyF12, Mod: ModAltMeta}:
		return "<a-m-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModAltMeta}:
		return "<a-m-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModAltMeta}:
		return "<a-m-delete>"
	case KeyComb{Key: KeyHome, Mod: ModAltMeta}:
		return "<a-m-home>"
	case KeyComb{Key: KeyEnd, Mod: ModAltMeta}:
		return "<a-m-end>"
	case KeyComb{Key: KeyPgup, Mod: ModAltMeta}:
		return "<a-m-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModAltMeta}:
		return "<a-m-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModAltMeta}:
		return "<a-m-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModAltMeta}:
		return "<a-m-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModAltMeta}:
		return "<a-m-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModAltMeta}:
		return "<a-m-right>"
	case KeyComb{Key: MouseLeft, Mod: ModAltMeta}:
		return "<a-m-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModAltMeta}:
		return "<a-m-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModAltMeta}:
		return "<a-m-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModAltMeta}:
		return "<a-m-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModAltMeta}:
		return "<a-m-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModAltMeta}:
		return "<a-m-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModAltMeta}:
		return "<a-m-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModAltMeta}:
		return "<a-m-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModAltMeta}:
		return "<a-m-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModAltMeta}:
		return "<a-m-esc>"
	case KeyComb{Key: KeySpace, Mod: ModAltMeta}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModAltMeta}:
		return "<a-m-space>"

	// alt+shift+meta
	case KeyComb{Key: KeyF1, Mod: ModAltShiftMeta}:
		return "<a-s-m-f1>"
	case KeyComb{Key: KeyF2, Mod: ModAltShiftMeta}:
		return "<a-s-m-f2>"
	case KeyComb{Key: KeyF3, Mod: ModAltShiftMeta}:
		return "<a-s-m-f3>"
	case KeyComb{Key: KeyF4, Mod: ModAltShiftMeta}:
		return "<a-s-m-f4>"
	case KeyComb{Key: KeyF5, Mod: ModAltShiftMeta}:
		return "<a-s-m-f5>"
	case KeyComb{Key: KeyF6, Mod: ModAltShiftMeta}:
		return "<a-s-m-f6>"
	case KeyComb{Key: KeyF7, Mod: ModAltShiftMeta}:
		return "<a-s-m-f7>"
	case KeyComb{Key: KeyF8, Mod: ModAltShiftMeta}:
		return "<a-s-m-f8>"
	case KeyComb{Key: KeyF9, Mod: ModAltShiftMeta}:
		return "<a-s-m-f9>"
	case KeyComb{Key: KeyF10, Mod: ModAltShiftMeta}:
		return "<a-s-m-f10>"
	case KeyComb{Key: KeyF11, Mod: ModAltShiftMeta}:
		return "<a-s-m-f11>"
	case KeyComb{Key: KeyF12, Mod: ModAltShiftMeta}:
		return "<a-s-m-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModAltShiftMeta}:
		return "<a-s-m-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModAltShiftMeta}:
		return "<a-s-m-delete>"
	case KeyComb{Key: KeyHome, Mod: ModAltShiftMeta}:
		return "<a-s-m-home>"
	case KeyComb{Key: KeyEnd, Mod: ModAltShiftMeta}:
		return "<a-s-m-end>"
	case KeyComb{Key: KeyPgup, Mod: ModAltShiftMeta}:
		return "<a-s-m-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModAltShiftMeta}:
		return "<a-s-m-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModAltShiftMeta}:
		return "<a-s-m-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModAltShiftMeta}:
		return "<a-s-m-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModAltShiftMeta}:
		return "<a-s-m-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModAltShiftMeta}:
		return "<a-s-m-right>"
	case KeyComb{Key: MouseLeft, Mod: ModAltShiftMeta}:
		return "<a-s-m-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModAltShiftMeta}:
		return "<a-s-m-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModAltShiftMeta}:
		return "<a-s-m-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModAltShiftMeta}:
		return "<a-s-m-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModAltShiftMeta}:
		return "<a-s-m-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModAltShiftMeta}:
		return "<a-s-m-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModAltShiftMeta}:
		return "<a-s-m-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModAltShiftMeta}:
		return "<a-s-m-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModAltShiftMeta}:
		return "<a-s-m-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModAltShiftMeta}:
		return "<a-s-m-esc>"
	case KeyComb{Key: KeySpace, Mod: ModAltShiftMeta}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModAltShiftMeta}:
		return "<a-s-m-space>"

	// alt+shift
	case KeyComb{Key: KeyF1, Mod: ModAltShift}:
		return "<a-s-f1>"
	case KeyComb{Key: KeyF2, Mod: ModAltShift}:
		return "<a-s-f2>"
	case KeyComb{Key: KeyF3, Mod: ModAltShift}:
		return "<a-s-f3>"
	case KeyComb{Key: KeyF4, Mod: ModAltShift}:
		return "<a-s-f4>"
	case KeyComb{Key: KeyF5, Mod: ModAltShift}:
		return "<a-s-f5>"
	case KeyComb{Key: KeyF6, Mod: ModAltShift}:
		return "<a-s-f6>"
	case KeyComb{Key: KeyF7, Mod: ModAltShift}:
		return "<a-s-f7>"
	case KeyComb{Key: KeyF8, Mod: ModAltShift}:
		return "<a-s-f8>"
	case KeyComb{Key: KeyF9, Mod: ModAltShift}:
		return "<a-s-f9>"
	case KeyComb{Key: KeyF10, Mod: ModAltShift}:
		return "<a-s-f10>"
	case KeyComb{Key: KeyF11, Mod: ModAltShift}:
		return "<a-s-f11>"
	case KeyComb{Key: KeyF12, Mod: ModAltShift}:
		return "<a-s-f12>"
	case KeyComb{Key: KeyInsert, Mod: ModAltShift}:
		return "<a-s-insert>"
	case KeyComb{Key: KeyDelete, Mod: ModAltShift}:
		return "<a-s-delete>"
	case KeyComb{Key: KeyHome, Mod: ModAltShift}:
		return "<a-s-home>"
	case KeyComb{Key: KeyEnd, Mod: ModAltShift}:
		return "<a-s-end>"
	case KeyComb{Key: KeyPgup, Mod: ModAltShift}:
		return "<a-s-pgup>"
	case KeyComb{Key: KeyPgdn, Mod: ModAltShift}:
		return "<a-s-pgdn>"
	case KeyComb{Key: KeyArrowUp, Mod: ModAltShift}:
		return "<a-s-up>"
	case KeyComb{Key: KeyArrowDown, Mod: ModAltShift}:
		return "<a-s-down>"
	case KeyComb{Key: KeyArrowLeft, Mod: ModAltShift}:
		return "<a-s-left>"
	case KeyComb{Key: KeyArrowRight, Mod: ModAltShift}:
		return "<a-s-right>"
	case KeyComb{Key: MouseLeft, Mod: ModAltShift}:
		return "<a-s-mouse-left>"
	case KeyComb{Key: MouseMiddle, Mod: ModAltShift}:
		return "<a-s-mouse-middle>"
	case KeyComb{Key: MouseRight, Mod: ModAltShift}:
		return "<a-s-mouse-right>"
	case KeyComb{Key: MouseRelease, Mod: ModAltShift}:
		return "<a-s-mouse-release>"
	case KeyComb{Key: MouseWheelUp, Mod: ModAltShift}:
		return "<a-s-mouse-wheel-up>"
	case KeyComb{Key: MouseWheelDown, Mod: ModAltShift}:
		return "<a-s-mouse-wheel-down>"
	case KeyComb{Key: KeyBackspace, Mod: ModAltShift}:
		return "<a-s-backspace>"
	case KeyComb{Key: KeyTab, Mod: ModAltShift}:
		return "<a-s-tab>"
	case KeyComb{Key: KeyEnter, Mod: ModAltShift}:
		return "<a-s-enter>"
	case KeyComb{Key: KeyEsc, Mod: ModAltShift}:
		return "<a-s-esc>"
	case KeyComb{Key: KeySpace, Mod: ModAltShift}, KeyComb{Ch: ' ', Key: KeySpace, Mod: ModAltShift}:
		return "<a-s-space>"

	// modifiers only
	case KeyComb{Mod: ModAlt}:
		return "<alt>"
	case KeyComb{Mod: ModShift}:
		return "<shift>"
	case KeyComb{Mod: ModMeta}:
		return "<meta>"
	case KeyComb{Mod: ModCtrl}:
		return "<ctrl>"
	case KeyComb{Mod: ModCtrlShift}:
		return "<ctrl-shift>"
	case KeyComb{Mod: ModCtrlAlt}:
		return "<ctrl-alt>"
	case KeyComb{Mod: ModCtrlMeta}:
		return "<ctrl-meta>"
	case KeyComb{Mod: ModCtrlShiftAlt}:
		return "<ctrl-shift-alt>"
	case KeyComb{Mod: ModCtrlShiftMeta}:
		return "<ctrl-shift-meta>"
	case KeyComb{Mod: ModCtrlAltMeta}:
		return "<ctrl-alt-meta>"
	case KeyComb{Mod: ModShiftMeta}:
		return "<shift-meta>"
	case KeyComb{Mod: ModAltMeta}:
		return "<alt-meta>"
	case KeyComb{Mod: ModAltShiftMeta}:
		return "<alt-shift-meta>"
	case KeyComb{Mod: ModAltShift}:
		return "<alt-shift>"
	default:
		return "<INVALID>"
	}
}

func processCharacter(ch rune, mod Modifier) (unshifCh string, shiftedMod Modifier) {
	switch mod {
	case ModAlt:
		shiftedMod = ModAltShift
	case ModMeta:
		shiftedMod = ModShiftMeta
	case ModCtrl:
		shiftedMod = ModCtrlShift
	case ModCtrlAlt:
		shiftedMod = ModCtrlShiftAlt
	case ModCtrlMeta:
		shiftedMod = ModCtrlShiftMeta
	case ModCtrlAltMeta:
		shiftedMod = ModCtrlShiftMeta // cannot use all the modifiers at once, drop alt
	case ModAltMeta:
		shiftedMod = ModAltShiftMeta
	case ModShift, ModCtrlShift,
		ModCtrlShiftAlt, ModCtrlShiftMeta,
		ModShiftMeta, ModAltShiftMeta,
		ModAltShift:
		shiftedMod = mod
	case 0:
		shiftedMod = ModShift
	default:
		shiftedMod = mod
	}

	switch ch {
	case 'A':
		return "a", shiftedMod
	case 'B':
		return "b", shiftedMod
	case 'C':
		return "c", shiftedMod
	case 'D':
		return "d", shiftedMod
	case 'E':
		return "e", shiftedMod
	case 'F':
		return "f", shiftedMod
	case 'G':
		return "g", shiftedMod
	case 'H':
		return "h", shiftedMod
	case 'I':
		return "i", shiftedMod
	case 'J':
		return "j", shiftedMod
	case 'K':
		return "k", shiftedMod
	case 'L':
		return "l", shiftedMod
	case 'M':
		return "m", shiftedMod
	case 'N':
		return "n", shiftedMod
	case 'O':
		return "o", shiftedMod
	case 'P':
		return "p", shiftedMod
	case 'Q':
		return "q", shiftedMod
	case 'R':
		return "r", shiftedMod
	case 'S':
		return "s", shiftedMod
	case 'T':
		return "t", shiftedMod
	case 'U':
		return "u", shiftedMod
	case 'V':
		return "v", shiftedMod
	case 'W':
		return "w", shiftedMod
	case 'X':
		return "x", shiftedMod
	case 'Y':
		return "y", shiftedMod
	case 'Z':
		return "z", shiftedMod
	case '_':
		return "-", shiftedMod
	case ')':
		return "0", shiftedMod
	case '!':
		return "1", shiftedMod
	case '@':
		return "2", shiftedMod
	case '#':
		return "3", shiftedMod
	case '$':
		return "4", shiftedMod
	case '%':
		return "5", shiftedMod
	case '^':
		return "6", shiftedMod
	case '&':
		return "7", shiftedMod
	case '*':
		return "8", shiftedMod
	case '(':
		return "9", shiftedMod
	case '+':
		return "=", shiftedMod
	case '<':
		return ",", shiftedMod
	case '{':
		return "[", shiftedMod
	case '}':
		return "]", shiftedMod
	case '~':
		return "`", shiftedMod
	case '?':
		return "/", shiftedMod
	case '|':
		return "\\\\", shiftedMod
	case '>':
		return ".", shiftedMod
	case '"':
		return "'", shiftedMod
	case ':':
		return ";", shiftedMod
	case '\\':
		return "\\\\", mod
	default:
		return string(ch), mod
	}
}
