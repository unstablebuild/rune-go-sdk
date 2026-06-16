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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseKeys(t *testing.T) {
	suite := []KeyComb{
		{Key: KeyF1},
		{Key: KeyF2},
		{Key: KeyF3},
		{Key: KeyF4},
		{Key: KeyF5},
		{Key: KeyF6},
		{Key: KeyF7},
		{Key: KeyF8},
		{Key: KeyF9},
		{Key: KeyF10},
		{Key: KeyF11},
		{Key: KeyF12},
		{Key: KeyInsert},
		{Key: KeyDelete},
		{Key: KeyHome},
		{Key: KeyEnd},
		{Key: KeyPgup},
		{Key: KeyPgdn},
		{Key: KeyArrowUp},
		{Key: KeyArrowDown},
		{Key: KeyArrowLeft},
		{Key: KeyArrowRight},
		{Key: MouseLeft},
		{Key: MouseMiddle},
		{Key: MouseRight},
		{Key: MouseRelease},
		{Key: MouseWheelUp},
		{Key: MouseWheelDown},
		{Key: KeyBackspace},
		{Key: KeyTab},
		{Key: KeyEnter},
		{Key: KeyEsc},
		{Key: KeySpace},
		{Key: KeyCapsLock},
		{Key: KeyNumLock},
		{Key: KeyScrollLock},
		{Key: KeyMenu},

		{Key: KeyF1, Mod: ModAlt},
		{Key: KeyF2, Mod: ModAlt},
		{Key: KeyF3, Mod: ModAlt},
		{Key: KeyF4, Mod: ModAlt},
		{Key: KeyF5, Mod: ModAlt},
		{Key: KeyF6, Mod: ModAlt},
		{Key: KeyF7, Mod: ModAlt},
		{Key: KeyF8, Mod: ModAlt},
		{Key: KeyF9, Mod: ModAlt},
		{Key: KeyF10, Mod: ModAlt},
		{Key: KeyF11, Mod: ModAlt},
		{Key: KeyF12, Mod: ModAlt},
		{Key: KeyInsert, Mod: ModAlt},
		{Key: KeyDelete, Mod: ModAlt},
		{Key: KeyHome, Mod: ModAlt},
		{Key: KeyEnd, Mod: ModAlt},
		{Key: KeyPgup, Mod: ModAlt},
		{Key: KeyPgdn, Mod: ModAlt},
		{Key: KeyArrowUp, Mod: ModAlt},
		{Key: KeyArrowDown, Mod: ModAlt},
		{Key: KeyArrowLeft, Mod: ModAlt},
		{Key: KeyArrowRight, Mod: ModAlt},
		{Key: MouseLeft, Mod: ModAlt},
		{Key: MouseMiddle, Mod: ModAlt},
		{Key: MouseRight, Mod: ModAlt},
		{Key: MouseRelease, Mod: ModAlt},
		{Key: MouseWheelUp, Mod: ModAlt},
		{Key: MouseWheelDown, Mod: ModAlt},
		{Key: KeyBackspace, Mod: ModAlt},
		{Key: KeyTab, Mod: ModAlt},
		{Key: KeyEnter, Mod: ModAlt},
		{Key: KeyEsc, Mod: ModAlt},
		{Key: KeySpace, Mod: ModAlt},

		{Key: KeyF1, Mod: ModCtrl},
		{Key: KeyF2, Mod: ModCtrl},
		{Key: KeyF3, Mod: ModCtrl},
		{Key: KeyF4, Mod: ModCtrl},
		{Key: KeyF5, Mod: ModCtrl},
		{Key: KeyF6, Mod: ModCtrl},
		{Key: KeyF7, Mod: ModCtrl},
		{Key: KeyF8, Mod: ModCtrl},
		{Key: KeyF9, Mod: ModCtrl},
		{Key: KeyF10, Mod: ModCtrl},
		{Key: KeyF11, Mod: ModCtrl},
		{Key: KeyF12, Mod: ModCtrl},
		{Key: KeyInsert, Mod: ModCtrl},
		{Key: KeyDelete, Mod: ModCtrl},
		{Key: KeyHome, Mod: ModCtrl},
		{Key: KeyEnd, Mod: ModCtrl},
		{Key: KeyPgup, Mod: ModCtrl},
		{Key: KeyPgdn, Mod: ModCtrl},
		{Key: KeyArrowUp, Mod: ModCtrl},
		{Key: KeyArrowDown, Mod: ModCtrl},
		{Key: KeyArrowLeft, Mod: ModCtrl},
		{Key: KeyArrowRight, Mod: ModCtrl},
		{Key: MouseLeft, Mod: ModCtrl},
		{Key: MouseMiddle, Mod: ModCtrl},
		{Key: MouseRight, Mod: ModCtrl},
		{Key: MouseRelease, Mod: ModCtrl},
		{Key: MouseWheelUp, Mod: ModCtrl},
		{Key: MouseWheelDown, Mod: ModCtrl},
		{Key: KeyBackspace, Mod: ModCtrl},
		{Key: KeyTab, Mod: ModCtrl},
		{Key: KeyEnter, Mod: ModCtrl},
		{Key: KeyEsc, Mod: ModCtrl},
		{Key: KeySpace, Mod: ModCtrl},

		{Key: KeyF1, Mod: ModShift},
		{Key: KeyF2, Mod: ModShift},
		{Key: KeyF3, Mod: ModShift},
		{Key: KeyF4, Mod: ModShift},
		{Key: KeyF5, Mod: ModShift},
		{Key: KeyF6, Mod: ModShift},
		{Key: KeyF7, Mod: ModShift},
		{Key: KeyF8, Mod: ModShift},
		{Key: KeyF9, Mod: ModShift},
		{Key: KeyF10, Mod: ModShift},
		{Key: KeyF11, Mod: ModShift},
		{Key: KeyF12, Mod: ModShift},
		{Key: KeyInsert, Mod: ModShift},
		{Key: KeyDelete, Mod: ModShift},
		{Key: KeyHome, Mod: ModShift},
		{Key: KeyEnd, Mod: ModShift},
		{Key: KeyPgup, Mod: ModShift},
		{Key: KeyPgdn, Mod: ModShift},
		{Key: KeyArrowUp, Mod: ModShift},
		{Key: KeyArrowDown, Mod: ModShift},
		{Key: KeyArrowLeft, Mod: ModShift},
		{Key: KeyArrowRight, Mod: ModShift},
		{Key: MouseLeft, Mod: ModShift},
		{Key: MouseMiddle, Mod: ModShift},
		{Key: MouseRight, Mod: ModShift},
		{Key: MouseRelease, Mod: ModShift},
		{Key: MouseWheelUp, Mod: ModShift},
		{Key: MouseWheelDown, Mod: ModShift},
		{Key: KeyBackspace, Mod: ModShift},
		{Key: KeyTab, Mod: ModShift},
		{Key: KeyEnter, Mod: ModShift},
		{Key: KeyEsc, Mod: ModShift},
		{Key: KeySpace, Mod: ModShift},

		{Key: KeyF1, Mod: ModMeta},
		{Key: KeyF2, Mod: ModMeta},
		{Key: KeyF3, Mod: ModMeta},
		{Key: KeyF4, Mod: ModMeta},
		{Key: KeyF5, Mod: ModMeta},
		{Key: KeyF6, Mod: ModMeta},
		{Key: KeyF7, Mod: ModMeta},
		{Key: KeyF8, Mod: ModMeta},
		{Key: KeyF9, Mod: ModMeta},
		{Key: KeyF10, Mod: ModMeta},
		{Key: KeyF11, Mod: ModMeta},
		{Key: KeyF12, Mod: ModMeta},
		{Key: KeyInsert, Mod: ModMeta},
		{Key: KeyDelete, Mod: ModMeta},
		{Key: KeyHome, Mod: ModMeta},
		{Key: KeyEnd, Mod: ModMeta},
		{Key: KeyPgup, Mod: ModMeta},
		{Key: KeyPgdn, Mod: ModMeta},
		{Key: KeyArrowUp, Mod: ModMeta},
		{Key: KeyArrowDown, Mod: ModMeta},
		{Key: KeyArrowLeft, Mod: ModMeta},
		{Key: KeyArrowRight, Mod: ModMeta},
		{Key: MouseLeft, Mod: ModMeta},
		{Key: MouseMiddle, Mod: ModMeta},
		{Key: MouseRight, Mod: ModMeta},
		{Key: MouseRelease, Mod: ModMeta},
		{Key: MouseWheelUp, Mod: ModMeta},
		{Key: MouseWheelDown, Mod: ModMeta},
		{Key: KeyBackspace, Mod: ModMeta},
		{Key: KeyTab, Mod: ModMeta},
		{Key: KeyEnter, Mod: ModMeta},
		{Key: KeyEsc, Mod: ModMeta},
		{Key: KeySpace, Mod: ModMeta},

		{Key: KeyF1, Mod: ModCtrlShift},
		{Key: KeyF2, Mod: ModCtrlShift},
		{Key: KeyF3, Mod: ModCtrlShift},
		{Key: KeyF4, Mod: ModCtrlShift},
		{Key: KeyF5, Mod: ModCtrlShift},
		{Key: KeyF6, Mod: ModCtrlShift},
		{Key: KeyF7, Mod: ModCtrlShift},
		{Key: KeyF8, Mod: ModCtrlShift},
		{Key: KeyF9, Mod: ModCtrlShift},
		{Key: KeyF10, Mod: ModCtrlShift},
		{Key: KeyF11, Mod: ModCtrlShift},
		{Key: KeyF12, Mod: ModCtrlShift},
		{Key: KeyInsert, Mod: ModCtrlShift},
		{Key: KeyDelete, Mod: ModCtrlShift},
		{Key: KeyHome, Mod: ModCtrlShift},
		{Key: KeyEnd, Mod: ModCtrlShift},
		{Key: KeyPgup, Mod: ModCtrlShift},
		{Key: KeyPgdn, Mod: ModCtrlShift},
		{Key: KeyArrowUp, Mod: ModCtrlShift},
		{Key: KeyArrowDown, Mod: ModCtrlShift},
		{Key: KeyArrowLeft, Mod: ModCtrlShift},
		{Key: KeyArrowRight, Mod: ModCtrlShift},
		{Key: MouseLeft, Mod: ModCtrlShift},
		{Key: MouseMiddle, Mod: ModCtrlShift},
		{Key: MouseRight, Mod: ModCtrlShift},
		{Key: MouseRelease, Mod: ModCtrlShift},
		{Key: MouseWheelUp, Mod: ModCtrlShift},
		{Key: MouseWheelDown, Mod: ModCtrlShift},
		{Key: KeyBackspace, Mod: ModCtrlShift},
		{Key: KeyTab, Mod: ModCtrlShift},
		{Key: KeyEnter, Mod: ModCtrlShift},
		{Key: KeyEsc, Mod: ModCtrlShift},
		{Key: KeySpace, Mod: ModCtrlShift},

		{Key: KeyF1, Mod: ModCtrlAlt},
		{Key: KeyF2, Mod: ModCtrlAlt},
		{Key: KeyF3, Mod: ModCtrlAlt},
		{Key: KeyF4, Mod: ModCtrlAlt},
		{Key: KeyF5, Mod: ModCtrlAlt},
		{Key: KeyF6, Mod: ModCtrlAlt},
		{Key: KeyF7, Mod: ModCtrlAlt},
		{Key: KeyF8, Mod: ModCtrlAlt},
		{Key: KeyF9, Mod: ModCtrlAlt},
		{Key: KeyF10, Mod: ModCtrlAlt},
		{Key: KeyF11, Mod: ModCtrlAlt},
		{Key: KeyF12, Mod: ModCtrlAlt},
		{Key: KeyInsert, Mod: ModCtrlAlt},
		{Key: KeyDelete, Mod: ModCtrlAlt},
		{Key: KeyHome, Mod: ModCtrlAlt},
		{Key: KeyEnd, Mod: ModCtrlAlt},
		{Key: KeyPgup, Mod: ModCtrlAlt},
		{Key: KeyPgdn, Mod: ModCtrlAlt},
		{Key: KeyArrowUp, Mod: ModCtrlAlt},
		{Key: KeyArrowDown, Mod: ModCtrlAlt},
		{Key: KeyArrowLeft, Mod: ModCtrlAlt},
		{Key: KeyArrowRight, Mod: ModCtrlAlt},
		{Key: MouseLeft, Mod: ModCtrlAlt},
		{Key: MouseMiddle, Mod: ModCtrlAlt},
		{Key: MouseRight, Mod: ModCtrlAlt},
		{Key: MouseRelease, Mod: ModCtrlAlt},
		{Key: MouseWheelUp, Mod: ModCtrlAlt},
		{Key: MouseWheelDown, Mod: ModCtrlAlt},
		{Key: KeyBackspace, Mod: ModCtrlAlt},
		{Key: KeyTab, Mod: ModCtrlAlt},
		{Key: KeyEnter, Mod: ModCtrlAlt},
		{Key: KeyEsc, Mod: ModCtrlAlt},
		{Key: KeySpace, Mod: ModCtrlAlt},

		{Key: KeyF1, Mod: ModCtrlMeta},
		{Key: KeyF2, Mod: ModCtrlMeta},
		{Key: KeyF3, Mod: ModCtrlMeta},
		{Key: KeyF4, Mod: ModCtrlMeta},
		{Key: KeyF5, Mod: ModCtrlMeta},
		{Key: KeyF6, Mod: ModCtrlMeta},
		{Key: KeyF7, Mod: ModCtrlMeta},
		{Key: KeyF8, Mod: ModCtrlMeta},
		{Key: KeyF9, Mod: ModCtrlMeta},
		{Key: KeyF10, Mod: ModCtrlMeta},
		{Key: KeyF11, Mod: ModCtrlMeta},
		{Key: KeyF12, Mod: ModCtrlMeta},
		{Key: KeyInsert, Mod: ModCtrlMeta},
		{Key: KeyDelete, Mod: ModCtrlMeta},
		{Key: KeyHome, Mod: ModCtrlMeta},
		{Key: KeyEnd, Mod: ModCtrlMeta},
		{Key: KeyPgup, Mod: ModCtrlMeta},
		{Key: KeyPgdn, Mod: ModCtrlMeta},
		{Key: KeyArrowUp, Mod: ModCtrlMeta},
		{Key: KeyArrowDown, Mod: ModCtrlMeta},
		{Key: KeyArrowLeft, Mod: ModCtrlMeta},
		{Key: KeyArrowRight, Mod: ModCtrlMeta},
		{Key: MouseLeft, Mod: ModCtrlMeta},
		{Key: MouseMiddle, Mod: ModCtrlMeta},
		{Key: MouseRight, Mod: ModCtrlMeta},
		{Key: MouseRelease, Mod: ModCtrlMeta},
		{Key: MouseWheelUp, Mod: ModCtrlMeta},
		{Key: MouseWheelDown, Mod: ModCtrlMeta},
		{Key: KeyBackspace, Mod: ModCtrlMeta},
		{Key: KeyTab, Mod: ModCtrlMeta},
		{Key: KeyEnter, Mod: ModCtrlMeta},
		{Key: KeyEsc, Mod: ModCtrlMeta},
		{Key: KeySpace, Mod: ModCtrlMeta},

		{Key: KeyF1, Mod: ModCtrlShiftAlt},
		{Key: KeyF2, Mod: ModCtrlShiftAlt},
		{Key: KeyF3, Mod: ModCtrlShiftAlt},
		{Key: KeyF4, Mod: ModCtrlShiftAlt},
		{Key: KeyF5, Mod: ModCtrlShiftAlt},
		{Key: KeyF6, Mod: ModCtrlShiftAlt},
		{Key: KeyF7, Mod: ModCtrlShiftAlt},
		{Key: KeyF8, Mod: ModCtrlShiftAlt},
		{Key: KeyF9, Mod: ModCtrlShiftAlt},
		{Key: KeyF10, Mod: ModCtrlShiftAlt},
		{Key: KeyF11, Mod: ModCtrlShiftAlt},
		{Key: KeyF12, Mod: ModCtrlShiftAlt},
		{Key: KeyInsert, Mod: ModCtrlShiftAlt},
		{Key: KeyDelete, Mod: ModCtrlShiftAlt},
		{Key: KeyHome, Mod: ModCtrlShiftAlt},
		{Key: KeyEnd, Mod: ModCtrlShiftAlt},
		{Key: KeyPgup, Mod: ModCtrlShiftAlt},
		{Key: KeyPgdn, Mod: ModCtrlShiftAlt},
		{Key: KeyArrowUp, Mod: ModCtrlShiftAlt},
		{Key: KeyArrowDown, Mod: ModCtrlShiftAlt},
		{Key: KeyArrowLeft, Mod: ModCtrlShiftAlt},
		{Key: KeyArrowRight, Mod: ModCtrlShiftAlt},
		{Key: MouseLeft, Mod: ModCtrlShiftAlt},
		{Key: MouseMiddle, Mod: ModCtrlShiftAlt},
		{Key: MouseRight, Mod: ModCtrlShiftAlt},
		{Key: MouseRelease, Mod: ModCtrlShiftAlt},
		{Key: MouseWheelUp, Mod: ModCtrlShiftAlt},
		{Key: MouseWheelDown, Mod: ModCtrlShiftAlt},
		{Key: KeyBackspace, Mod: ModCtrlShiftAlt},
		{Key: KeyTab, Mod: ModCtrlShiftAlt},
		{Key: KeyEnter, Mod: ModCtrlShiftAlt},
		{Key: KeyEsc, Mod: ModCtrlShiftAlt},
		{Key: KeySpace, Mod: ModCtrlShiftAlt},

		{Key: KeyF1, Mod: ModCtrlShiftMeta},
		{Key: KeyF2, Mod: ModCtrlShiftMeta},
		{Key: KeyF3, Mod: ModCtrlShiftMeta},
		{Key: KeyF4, Mod: ModCtrlShiftMeta},
		{Key: KeyF5, Mod: ModCtrlShiftMeta},
		{Key: KeyF6, Mod: ModCtrlShiftMeta},
		{Key: KeyF7, Mod: ModCtrlShiftMeta},
		{Key: KeyF8, Mod: ModCtrlShiftMeta},
		{Key: KeyF9, Mod: ModCtrlShiftMeta},
		{Key: KeyF10, Mod: ModCtrlShiftMeta},
		{Key: KeyF11, Mod: ModCtrlShiftMeta},
		{Key: KeyF12, Mod: ModCtrlShiftMeta},
		{Key: KeyInsert, Mod: ModCtrlShiftMeta},
		{Key: KeyDelete, Mod: ModCtrlShiftMeta},
		{Key: KeyHome, Mod: ModCtrlShiftMeta},
		{Key: KeyEnd, Mod: ModCtrlShiftMeta},
		{Key: KeyPgup, Mod: ModCtrlShiftMeta},
		{Key: KeyPgdn, Mod: ModCtrlShiftMeta},
		{Key: KeyArrowUp, Mod: ModCtrlShiftMeta},
		{Key: KeyArrowDown, Mod: ModCtrlShiftMeta},
		{Key: KeyArrowLeft, Mod: ModCtrlShiftMeta},
		{Key: KeyArrowRight, Mod: ModCtrlShiftMeta},
		{Key: MouseLeft, Mod: ModCtrlShiftMeta},
		{Key: MouseMiddle, Mod: ModCtrlShiftMeta},
		{Key: MouseRight, Mod: ModCtrlShiftMeta},
		{Key: MouseRelease, Mod: ModCtrlShiftMeta},
		{Key: MouseWheelUp, Mod: ModCtrlShiftMeta},
		{Key: MouseWheelDown, Mod: ModCtrlShiftMeta},
		{Key: KeyBackspace, Mod: ModCtrlShiftMeta},
		{Key: KeyTab, Mod: ModCtrlShiftMeta},
		{Key: KeyEnter, Mod: ModCtrlShiftMeta},
		{Key: KeyEsc, Mod: ModCtrlShiftMeta},
		{Key: KeySpace, Mod: ModCtrlShiftMeta},

		{Key: KeyF1, Mod: ModCtrlAltMeta},
		{Key: KeyF2, Mod: ModCtrlAltMeta},
		{Key: KeyF3, Mod: ModCtrlAltMeta},
		{Key: KeyF4, Mod: ModCtrlAltMeta},
		{Key: KeyF5, Mod: ModCtrlAltMeta},
		{Key: KeyF6, Mod: ModCtrlAltMeta},
		{Key: KeyF7, Mod: ModCtrlAltMeta},
		{Key: KeyF8, Mod: ModCtrlAltMeta},
		{Key: KeyF9, Mod: ModCtrlAltMeta},
		{Key: KeyF10, Mod: ModCtrlAltMeta},
		{Key: KeyF11, Mod: ModCtrlAltMeta},
		{Key: KeyF12, Mod: ModCtrlAltMeta},
		{Key: KeyInsert, Mod: ModCtrlAltMeta},
		{Key: KeyDelete, Mod: ModCtrlAltMeta},
		{Key: KeyHome, Mod: ModCtrlAltMeta},
		{Key: KeyEnd, Mod: ModCtrlAltMeta},
		{Key: KeyPgup, Mod: ModCtrlAltMeta},
		{Key: KeyPgdn, Mod: ModCtrlAltMeta},
		{Key: KeyArrowUp, Mod: ModCtrlAltMeta},
		{Key: KeyArrowDown, Mod: ModCtrlAltMeta},
		{Key: KeyArrowLeft, Mod: ModCtrlAltMeta},
		{Key: KeyArrowRight, Mod: ModCtrlAltMeta},
		{Key: MouseLeft, Mod: ModCtrlAltMeta},
		{Key: MouseMiddle, Mod: ModCtrlAltMeta},
		{Key: MouseRight, Mod: ModCtrlAltMeta},
		{Key: MouseRelease, Mod: ModCtrlAltMeta},
		{Key: MouseWheelUp, Mod: ModCtrlAltMeta},
		{Key: MouseWheelDown, Mod: ModCtrlAltMeta},
		{Key: KeyBackspace, Mod: ModCtrlAltMeta},
		{Key: KeyTab, Mod: ModCtrlAltMeta},
		{Key: KeyEnter, Mod: ModCtrlAltMeta},
		{Key: KeyEsc, Mod: ModCtrlAltMeta},
		{Key: KeySpace, Mod: ModCtrlAltMeta},

		{Key: KeyF1, Mod: ModShiftMeta},
		{Key: KeyF2, Mod: ModShiftMeta},
		{Key: KeyF3, Mod: ModShiftMeta},
		{Key: KeyF4, Mod: ModShiftMeta},
		{Key: KeyF5, Mod: ModShiftMeta},
		{Key: KeyF6, Mod: ModShiftMeta},
		{Key: KeyF7, Mod: ModShiftMeta},
		{Key: KeyF8, Mod: ModShiftMeta},
		{Key: KeyF9, Mod: ModShiftMeta},
		{Key: KeyF10, Mod: ModShiftMeta},
		{Key: KeyF11, Mod: ModShiftMeta},
		{Key: KeyF12, Mod: ModShiftMeta},
		{Key: KeyInsert, Mod: ModShiftMeta},
		{Key: KeyDelete, Mod: ModShiftMeta},
		{Key: KeyHome, Mod: ModShiftMeta},
		{Key: KeyEnd, Mod: ModShiftMeta},
		{Key: KeyPgup, Mod: ModShiftMeta},
		{Key: KeyPgdn, Mod: ModShiftMeta},
		{Key: KeyArrowUp, Mod: ModShiftMeta},
		{Key: KeyArrowDown, Mod: ModShiftMeta},
		{Key: KeyArrowLeft, Mod: ModShiftMeta},
		{Key: KeyArrowRight, Mod: ModShiftMeta},
		{Key: MouseLeft, Mod: ModShiftMeta},
		{Key: MouseMiddle, Mod: ModShiftMeta},
		{Key: MouseRight, Mod: ModShiftMeta},
		{Key: MouseRelease, Mod: ModShiftMeta},
		{Key: MouseWheelUp, Mod: ModShiftMeta},
		{Key: MouseWheelDown, Mod: ModShiftMeta},
		{Key: KeyBackspace, Mod: ModShiftMeta},
		{Key: KeyTab, Mod: ModShiftMeta},
		{Key: KeyEnter, Mod: ModShiftMeta},
		{Key: KeyEsc, Mod: ModShiftMeta},
		{Key: KeySpace, Mod: ModShiftMeta},

		{Key: KeyF1, Mod: ModAltMeta},
		{Key: KeyF2, Mod: ModAltMeta},
		{Key: KeyF3, Mod: ModAltMeta},
		{Key: KeyF4, Mod: ModAltMeta},
		{Key: KeyF5, Mod: ModAltMeta},
		{Key: KeyF6, Mod: ModAltMeta},
		{Key: KeyF7, Mod: ModAltMeta},
		{Key: KeyF8, Mod: ModAltMeta},
		{Key: KeyF9, Mod: ModAltMeta},
		{Key: KeyF10, Mod: ModAltMeta},
		{Key: KeyF11, Mod: ModAltMeta},
		{Key: KeyF12, Mod: ModAltMeta},
		{Key: KeyInsert, Mod: ModAltMeta},
		{Key: KeyDelete, Mod: ModAltMeta},
		{Key: KeyHome, Mod: ModAltMeta},
		{Key: KeyEnd, Mod: ModAltMeta},
		{Key: KeyPgup, Mod: ModAltMeta},
		{Key: KeyPgdn, Mod: ModAltMeta},
		{Key: KeyArrowUp, Mod: ModAltMeta},
		{Key: KeyArrowDown, Mod: ModAltMeta},
		{Key: KeyArrowLeft, Mod: ModAltMeta},
		{Key: KeyArrowRight, Mod: ModAltMeta},
		{Key: MouseLeft, Mod: ModAltMeta},
		{Key: MouseMiddle, Mod: ModAltMeta},
		{Key: MouseRight, Mod: ModAltMeta},
		{Key: MouseRelease, Mod: ModAltMeta},
		{Key: MouseWheelUp, Mod: ModAltMeta},
		{Key: MouseWheelDown, Mod: ModAltMeta},
		{Key: KeyBackspace, Mod: ModAltMeta},
		{Key: KeyTab, Mod: ModAltMeta},
		{Key: KeyEnter, Mod: ModAltMeta},
		{Key: KeyEsc, Mod: ModAltMeta},
		{Key: KeySpace, Mod: ModAltMeta},

		{Key: KeyF1, Mod: ModAltShiftMeta},
		{Key: KeyF2, Mod: ModAltShiftMeta},
		{Key: KeyF3, Mod: ModAltShiftMeta},
		{Key: KeyF4, Mod: ModAltShiftMeta},
		{Key: KeyF5, Mod: ModAltShiftMeta},
		{Key: KeyF6, Mod: ModAltShiftMeta},
		{Key: KeyF7, Mod: ModAltShiftMeta},
		{Key: KeyF8, Mod: ModAltShiftMeta},
		{Key: KeyF9, Mod: ModAltShiftMeta},
		{Key: KeyF10, Mod: ModAltShiftMeta},
		{Key: KeyF11, Mod: ModAltShiftMeta},
		{Key: KeyF12, Mod: ModAltShiftMeta},
		{Key: KeyInsert, Mod: ModAltShiftMeta},
		{Key: KeyDelete, Mod: ModAltShiftMeta},
		{Key: KeyHome, Mod: ModAltShiftMeta},
		{Key: KeyEnd, Mod: ModAltShiftMeta},
		{Key: KeyPgup, Mod: ModAltShiftMeta},
		{Key: KeyPgdn, Mod: ModAltShiftMeta},
		{Key: KeyArrowUp, Mod: ModAltShiftMeta},
		{Key: KeyArrowDown, Mod: ModAltShiftMeta},
		{Key: KeyArrowLeft, Mod: ModAltShiftMeta},
		{Key: KeyArrowRight, Mod: ModAltShiftMeta},
		{Key: MouseLeft, Mod: ModAltShiftMeta},
		{Key: MouseMiddle, Mod: ModAltShiftMeta},
		{Key: MouseRight, Mod: ModAltShiftMeta},
		{Key: MouseRelease, Mod: ModAltShiftMeta},
		{Key: MouseWheelUp, Mod: ModAltShiftMeta},
		{Key: MouseWheelDown, Mod: ModAltShiftMeta},
		{Key: KeyBackspace, Mod: ModAltShiftMeta},
		{Key: KeyTab, Mod: ModAltShiftMeta},
		{Key: KeyEnter, Mod: ModAltShiftMeta},
		{Key: KeyEsc, Mod: ModAltShiftMeta},
		{Key: KeySpace, Mod: ModAltShiftMeta},

		{Key: KeyF1, Mod: ModAltShift},
		{Key: KeyF2, Mod: ModAltShift},
		{Key: KeyF3, Mod: ModAltShift},
		{Key: KeyF4, Mod: ModAltShift},
		{Key: KeyF5, Mod: ModAltShift},
		{Key: KeyF6, Mod: ModAltShift},
		{Key: KeyF7, Mod: ModAltShift},
		{Key: KeyF8, Mod: ModAltShift},
		{Key: KeyF9, Mod: ModAltShift},
		{Key: KeyF10, Mod: ModAltShift},
		{Key: KeyF11, Mod: ModAltShift},
		{Key: KeyF12, Mod: ModAltShift},
		{Key: KeyInsert, Mod: ModAltShift},
		{Key: KeyDelete, Mod: ModAltShift},
		{Key: KeyHome, Mod: ModAltShift},
		{Key: KeyEnd, Mod: ModAltShift},
		{Key: KeyPgup, Mod: ModAltShift},
		{Key: KeyPgdn, Mod: ModAltShift},
		{Key: KeyArrowUp, Mod: ModAltShift},
		{Key: KeyArrowDown, Mod: ModAltShift},
		{Key: KeyArrowLeft, Mod: ModAltShift},
		{Key: KeyArrowRight, Mod: ModAltShift},
		{Key: MouseLeft, Mod: ModAltShift},
		{Key: MouseMiddle, Mod: ModAltShift},
		{Key: MouseRight, Mod: ModAltShift},
		{Key: MouseRelease, Mod: ModAltShift},
		{Key: MouseWheelUp, Mod: ModAltShift},
		{Key: MouseWheelDown, Mod: ModAltShift},
		{Key: KeyBackspace, Mod: ModAltShift},
		{Key: KeyTab, Mod: ModAltShift},
		{Key: KeyEnter, Mod: ModAltShift},
		{Key: KeyEsc, Mod: ModAltShift},
		{Key: KeySpace, Mod: ModAltShift},
		{Mod: ModAlt},
		{Mod: ModShift},
		{Mod: ModMeta},
		{Mod: ModCtrl},
		{Mod: ModCtrlShift},
		{Mod: ModCtrlAlt},
		{Mod: ModCtrlMeta},
		{Mod: ModCtrlShiftAlt},
		{Mod: ModCtrlShiftMeta},
		{Mod: ModCtrlAltMeta},
		{Mod: ModShiftMeta},
		{Mod: ModAltMeta},
		{Mod: ModAltShiftMeta},
		{Mod: ModAltShift},
	}

	for i, comb := range suite {
		t.Run(fmt.Sprintf("String, ParseKey twice: %d: %#v", i, comb), func(t *testing.T) {
			c, err := ParseKey(comb.String())
			require.NoError(t, err)
			assert.Equal(t, comb, c, comb)

			c, err = ParseKey(c.String())
			require.NoError(t, err)
			assert.Equal(t, comb, c, comb)
		})

		t.Run(fmt.Sprintf("ShortString, ParseKey twice: %d: %#v", i, comb), func(t *testing.T) {
			c, err := ParseKey(comb.ShortString())
			require.NoError(t, err)
			assert.Equal(t, comb, c, comb)

			c, err = ParseKey(c.ShortString())
			require.NoError(t, err)
			assert.Equal(t, comb, c, comb)
		})

		t.Run(fmt.Sprintf("ShortString, ParseKey, String, ParseKey: %d: %#v", i, comb), func(t *testing.T) {
			c, err := ParseKey(comb.ShortString())
			require.NoError(t, err)
			assert.Equal(t, comb, c, comb)

			c, err = ParseKey(c.String())
			require.NoError(t, err)
			assert.Equal(t, comb, c, comb)
		})

		t.Run(fmt.Sprintf("String, ParseKey, ShortString, ParseKey: %d: %#v", i, comb), func(t *testing.T) {
			c, err := ParseKey(comb.String())
			require.NoError(t, err)
			assert.Equal(t, comb, c, comb)

			c, err = ParseKey(c.ShortString())
			require.NoError(t, err)
			assert.Equal(t, comb, c, comb)
		})
	}

	t.Run("combined sequence of keys", func(t *testing.T) {
		var builder strings.Builder
		for _, comb := range suite {
			builder.WriteString(comb.String())
		}
		keys, err := ParseKeys(builder.String())
		require.NoError(t, err)
		assert.ElementsMatch(t, suite, keys)
	})
}

// TestParseSyntheticPhysicalKeys covers the GUI-only physical keys that have no
// terminal keysym (CapsLock/NumLock/ScrollLock/Menu) but are parseable so they
// can be used as key-mapping sources/targets.
func TestParseSyntheticPhysicalKeys(t *testing.T) {
	cases := []struct {
		str  string
		comb KeyComb
	}{
		{"<capslock>", KeyComb{Key: KeyCapsLock}},
		{"<numlock>", KeyComb{Key: KeyNumLock}},
		{"<scrolllock>", KeyComb{Key: KeyScrollLock}},
		{"<menu>", KeyComb{Key: KeyMenu}},
	}
	for _, tc := range cases {
		t.Run(tc.str, func(t *testing.T) {
			got, err := ParseKey(tc.str)
			require.NoError(t, err)
			assert.Equal(t, tc.comb, got)
			assert.Equal(t, tc.str, got.String())
		})
	}
}

// test chars separately as equivalence must be tested via KeyComb.String()
// because Mod shift with unshifted characters is not emitted by tui or gui
// and so it's "invalid", but we should still test that it converts to the right
// KeyComb.
func TestParseCharKey(t *testing.T) {
	modifiers := []Modifier{0, ModAlt, ModShift, ModMeta,
		ModCtrl, ModCtrlShift, ModCtrlAlt, ModCtrlMeta,
		ModCtrlShiftAlt, ModCtrlShiftMeta,
		/* ModCtrlAltMeta skip since it cannot be fully upgraded/downgraded with shift */
		ModShiftMeta, ModAltMeta, ModAltShiftMeta, ModAltShift}
	charKeys := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQR" +
		"STUVWXYZ1234567890-=`[];',./~!@#$%^&*()_+{}\":?<>\\|"

	var suite []KeyComb
	for _, ch := range charKeys {
		for _, mod := range modifiers {
			suite = append(suite, KeyComb{Ch: ch, Mod: mod})
		}
	}
	for i, comb := range suite {
		t.Run(fmt.Sprintf("String, ParseKey twice: %d: %#v", i, comb), func(t *testing.T) {
			c, err := ParseKey(comb.String())
			require.NoError(t, err)
			assert.Equal(t, comb.String(), c.String())

			c, err = ParseKey(c.String())
			require.NoError(t, err)
			assert.Equal(t, comb.String(), c.String(), comb)
		})

		t.Run(fmt.Sprintf("ShortString, ParseKey twice: %d: %#v", i, comb), func(t *testing.T) {
			c, err := ParseKey(comb.ShortString())
			require.NoError(t, err)
			assert.Equal(t, comb.String(), c.String(), comb)

			c, err = ParseKey(c.ShortString())
			require.NoError(t, err)
			assert.Equal(t, comb.String(), c.String(), comb)
		})

		t.Run(fmt.Sprintf("ShortString, ParseKey, String, ParseKey: %d: %#v", i, comb), func(t *testing.T) {
			c, err := ParseKey(comb.ShortString())
			require.NoError(t, err)
			assert.Equal(t, comb.String(), c.String(), comb)

			c, err = ParseKey(c.String())
			require.NoError(t, err)
			assert.Equal(t, comb.String(), c.String(), comb)
		})

		t.Run(fmt.Sprintf("String, ParseKey, ShortString, ParseKey: %d: %#v", i, comb), func(t *testing.T) {
			c, err := ParseKey(comb.String())
			require.NoError(t, err)
			assert.Equal(t, comb.String(), c.String(), comb)

			c, err = ParseKey(c.ShortString())
			require.NoError(t, err)
			assert.Equal(t, comb.String(), c.String(), comb)
		})

	}

	t.Run("combined sequence of characters", func(t *testing.T) {
		var builder strings.Builder
		for _, comb := range suite {
			builder.WriteString(comb.String())
		}
		seq := builder.String()
		keys, err := ParseKeys(seq)
		require.NoError(t, err)

		builder.Reset()
		for _, comb := range keys {
			builder.WriteString(comb.String())
		}
		assert.Equal(t, seq, builder.String())
	})

	t.Run("ParseKeyswith with mixed sequences", func(t *testing.T) {
		suite := []struct {
			in          string
			expectedOut []KeyComb
			expectedErr bool
		}{
			{
				in: "<esc>:edit<space>hello<enter>",
				expectedOut: []KeyComb{
					{Key: KeyEsc},
					{Ch: ':'},
					{Ch: 'e'},
					{Ch: 'd'},
					{Ch: 'i'},
					{Ch: 't'},
					{Key: KeySpace},
					{Ch: 'h'},
					{Ch: 'e'},
					{Ch: 'l'},
					{Ch: 'l'},
					{Ch: 'o'},
					{Key: KeyEnter},
				},
			},
			{
				in:          "<esc<:edit<space>",
				expectedErr: true,
			},
		}

		for _, test := range suite {
			actualOut, actualErr := ParseKeys(test.in)
			if test.expectedErr {
				require.Error(t, actualErr)
			} else {
				require.NoError(t, actualErr)
				assert.Equal(t, test.expectedOut, actualOut)
			}
		}
	})
}

func TestParseNonKeyChar(t *testing.T) {
	charKeys := "œ∑´®†¥¨ˆøπ“‘æ…¬˚∆˙©ƒ∂ßå≈ç√∫˜µ≤≥÷¡™£¢∞§¶•ªº–≠"

	var suite []KeyComb
	for _, ch := range charKeys {
		suite = append(suite, KeyComb{Ch: ch})
	}
	for i, comb := range suite {
		t.Run(fmt.Sprintf("String, ParseKey twice: %d: %#v", i, comb), func(t *testing.T) {
			c, err := ParseKey(comb.String())
			require.NoError(t, err)
			assert.Equal(t, comb, c)

			c, err = ParseKey(c.String())
			require.NoError(t, err)
			assert.Equal(t, comb, c, comb)
		})

		t.Run(fmt.Sprintf("ShortString, ParseKey twice: %d: %#v", i, comb), func(t *testing.T) {
			c, err := ParseKey(comb.ShortString())
			require.NoError(t, err)
			assert.Equal(t, comb, c, comb)

			c, err = ParseKey(c.ShortString())
			require.NoError(t, err)
			assert.Equal(t, comb, c, comb)
		})

		t.Run(fmt.Sprintf("ShortString, ParseKey, String, ParseKey: %d: %#v", i, comb), func(t *testing.T) {
			c, err := ParseKey(comb.ShortString())
			require.NoError(t, err)
			assert.Equal(t, comb, c, comb)

			c, err = ParseKey(c.String())
			require.NoError(t, err)
			assert.Equal(t, comb, c, comb)
		})

		t.Run(fmt.Sprintf("String, ParseKey, ShortString, ParseKey: %d: %#v", i, comb), func(t *testing.T) {
			c, err := ParseKey(comb.String())
			require.NoError(t, err)
			assert.Equal(t, comb, c, comb)

			c, err = ParseKey(c.ShortString())
			require.NoError(t, err)
			assert.Equal(t, comb, c, comb)
		})

	}

	t.Run("combined sequence of characters", func(t *testing.T) {
		var builder strings.Builder
		for _, comb := range suite {
			builder.WriteString(comb.String())
		}
		keys, err := ParseKeys(builder.String())
		require.NoError(t, err)
		assert.ElementsMatch(t, suite, keys)
	})
}
