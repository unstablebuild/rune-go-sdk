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

import "github.com/unstablebuild/tcell/v3"

// Screen is the subset of tcell.Screen that rune-go-sdk's TUI event loop
// and TermboxWriter depend on. Any type that satisfies tcell.Screen
// (including the process default returned by termbox.Screen() and
// tcell.NewSimulationScreen()) satisfies Screen.
//
// Screen intentionally omits lifecycle/setup methods (Init, Fini,
// EnablePaste, EnableFocus, EnableMouse, Tty); those are owned by the
// caller (term.Init for the default path, or the SSH/integration layer
// for callers that build their own Screen).
type Screen interface {
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
