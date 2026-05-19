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

// Screen is the rendering surface used by rune-go-sdk's TUI event loop
// and ScreenWriter. It is intentionally tcell-free: tcell-backed
// renderers (termbox.Screen, gui/Screen, sshshop) adapt to this
// interface via dedicated adapters so callers never see tcell types
// through the SDK.
//
// Screen intentionally omits lifecycle/setup methods (Init, Fini,
// EnablePaste, EnableFocus, EnableMouse, Tty); those are owned by the
// caller (term.Init for the default path, or the SSH/integration layer
// for callers that build their own Screen).
type Screen interface {
	SetContent(x, y int, primary rune, combining []rune, width uint8, style Style)
	UnionStyle(x, y int, style Style)
	Fill(ch rune, style Style)
	ShowCursor(x, y int)
	HideCursor()
	SetCursorStyle(CursorStyle)
	Size() (int, int)
	Show()
	Poll() <-chan Event
	PostEvent(Event) error
	Bell()
}
