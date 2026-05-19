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
	"sync/atomic"
)

var _ Writer = (*ScreenWriter)(nil)

// ScreenWriter implements term.Writer on top of a term.Screen. The
// writer is tcell-agnostic; tcell-backed screens are wrapped by a
// term.TcellScreen adapter so any color/style/event conversions happen
// at that boundary rather than here.
type ScreenWriter struct {
	ctx    context.Context
	screen Screen
	// interruptPending coalesces payload-less interrupt events posted via
	// this writer so that bursts collapse into at most one outstanding
	// interrupt at a time. The event loop clears the flag when it
	// processes the interrupt.
	interruptPending atomic.Bool
}

// NewScreenWriter allocates storage for a new ScreenWriter. The
// returned writer has no Screen attached yet; it is the caller's
// responsibility to set one via SetScreen before the writer is used.
// term.Init() attaches the process-wide termbox screen to
// term.DefaultWriter, so that writer is ready to use after Init.
func NewScreenWriter(scr Screen) *ScreenWriter {
	if scr == nil {
		panic("term: NewScreenWriter called with nil Screen")
	}
	ret := new(ScreenWriter)
	ret.screen = scr
	ret.ctx = context.Background()
	return ret
}

// SetCell satisfies term.Writer.
func (w *ScreenWriter) SetCell(pos Coordinates, c Cell) {
	w.screen.SetContent(
		pos.X, pos.Y, c.Ch, c.CombiningRunes(),
		c.Width, c.Style(),
	)
}

// UnionAttributes satisfies term.Writer.
func (w *ScreenWriter) UnionAttributes(pos Coordinates, attr Attributes) {
	w.screen.UnionStyle(
		pos.X, pos.Y, attr.Style(),
	)
}

// Flush makes all the content changes made using SetCell and
// UnionAttributes visible on the display.
func (w *ScreenWriter) Flush() error {
	w.screen.Show()
	return nil
}

// Clear fills the screen with the given attributes and empty cells.
func (w *ScreenWriter) Clear(attr Attributes) (err error) {
	w.screen.Fill(' ', attr.Style())
	return
}

// SetCursor displays the terminal cursor at the given location.
func (w *ScreenWriter) SetCursor(pos Coordinates) {
	w.screen.ShowCursor(pos.X, pos.Y)
}

// Size returns the size of this writer's screen.
func (w *ScreenWriter) Size() (int, int) {
	return w.screen.Size()
}

// SetCursorStyle sets the cursor style on this writer's screen.
func (w *ScreenWriter) SetCursorStyle(style CursorStyle) {
	w.screen.SetCursorStyle(style)
}

// Poll returns the underlying event channel of this writer's screen.
func (w *ScreenWriter) Poll() <-chan Event {
	return w.screen.Poll()
}

// Bell rings the bell on this writer's screen.
func (w *ScreenWriter) Bell() {
	w.screen.Bell()
}

// PublishEvent posts the given term.Event onto this writer's screen
// event queue. Returns false if the queue is full. Does not support
// EventMouse, EventRaw or EventPaste events (these are silently ignored
// and true is returned, mirroring termbox.PublishEvent).
func (w *ScreenWriter) PublishEvent(ev Event) bool {
	if ev.Type == EventInterrupt && ev.Raw == nil && ev.UserFunc == nil &&
		!w.interruptPending.CompareAndSwap(false, true) {
		return true
	}
	return w.screen.PostEvent(ev) == nil
}

// ClearInterruptPending clears the coalesced-interrupt flag for this
// writer. The TUI event loop calls this after delivering an interrupt
// to the root handler so that the next payload-less interrupt is
// allowed through.
func (w *ScreenWriter) ClearInterruptPending() {
	w.interruptPending.Store(false)
}

// Context satisfies term.Writer.
func (w *ScreenWriter) Context() context.Context {
	return w.ctx
}

// SetContext sets the context for the next call to Context.
func (w *ScreenWriter) SetContext(ctx context.Context) {
	w.ctx = ctx
}
