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
	"sync/atomic"

	"github.com/unstablebuild/tcell/v3"
	"github.com/unstablebuild/tcell/v3/termbox"
)

var (
	defaultAttr  = Attributes{Fg: tcell.ColorDefault, Bg: tcell.ColorDefault}
	publishEvent atomic.Value

	// DefaultWriter returns the default global Writer.
	DefaultWriter *TermboxWriter = NewTermboxWriter()
)

func init() {
	publishEvent.Store(func(termbox.Event) bool { return false })
}

// SetInputMode sets termbox input mode. Termbox has two input modes:
//
// 1. Esc input mode. When ESC sequence is in the buffer and it doesn't match
// any known sequence. ESC means KeyEsc. This is the default input mode.
//
// 2. Alt input mode. When ESC sequence is in the buffer and it doesn't match
// any known sequence. ESC enables ModAlt modifier for the next keyboard event.
//
// Both input modes can be OR'ed with Mouse mode. Setting Mouse mode bit up will
// enable mouse button press/release and drag events.
//
// If 'mode' is InputCurrent, returns the current input mode. See also Input*
// constants.
func SetInputMode(mode InputMode) InputMode {
	return InputMode(termbox.SetInputMode(termbox.InputMode(mode)))
}

// SetAttr sets the global foreground and background attributes.
func SetAttr(newattr Attributes) {
	defaultAttr = newattr
}

// Attr returns the global foreground and background attributes.
func Attr() Attributes {
	return defaultAttr
}

// Init initializes the underlying vte client.
func Init() error {
	err := termbox.Init()
	if err != nil {
		return err
	}
	screen := termbox.Screen()
	screen.EnablePaste()
	screen.EnableFocus()
	publishEvent.Store(termbox.PublishEvent)
	return nil
}

// Size returns the size of the terminal window.
func Size() (width int, height int) {
	return termbox.Size()
}

// PollEvent waits for an event and returns it.
// This is a blocking function call.
func PollEvent() (ev Event) {
	tev := termbox.PollEvent()
	return termboxEventToEvent(tev)
}

// Close writer; should be called after successful initialization
// when termbox's functionality isn't required anymore.
func Close() {
	termbox.Close()
}

// PublishEvent sends a synthetic event to the event poller.
// If the event queue is full then this method does not
// publish the event and returns false.
func PublishEvent(ev Event) bool {
	return publishEvent.Load().(func(termbox.Event) bool)(eventToTermboxEvent(ev))
}

// RingBell makes an audible noise. This must be synchronized
// against other accesses to the term.Writer's screen buffer.
func RingBell() {
	termbox.Screen().Bell()
}

// PublishBell It's a shorthand for `ScheduleNextTick(RingBell)`.
func PublishBell() {
	ScheduleNextTick(RingBell)
}

// ScheduleNextTick schedules running fn on the next event-loop iteration.
func ScheduleNextTick(fn func()) bool {
	return PublishEvent(Event{Type: EventInterrupt, UserFunc: fn})
}

// Poll gives access to the underlying tcell.Event channel.
func Poll() <-chan tcell.Event {
	return termbox.Screen().Poll()
}

// FromTcellEvent converts a tcell.Event into a term.Event.
func FromTcellEvent(tev tcell.Event) Event {
	return termboxEventToEvent(termbox.NewEvent(tev))
}

// CursorStyle represents a given cursor style, which can include the shape and
// whether the cursor blinks or is solid.  Support for changing this is not universal.
type CursorStyle int

// List of cursor styles.
const (
	CursorStyleDefault           = CursorStyle(tcell.CursorStyleDefault)
	CursorStyleBlinkingBlock     = CursorStyle(tcell.CursorStyleBlinkingBlock)
	CursorStyleSteadyBlock       = CursorStyle(tcell.CursorStyleSteadyBlock)
	CursorStyleBlinkingUnderline = CursorStyle(tcell.CursorStyleBlinkingUnderline)
	CursorStyleSteadyUnderline   = CursorStyle(tcell.CursorStyleSteadyUnderline)
	CursorStyleBlinkingBar       = CursorStyle(tcell.CursorStyleBlinkingBar)
	CursorStyleSteadyBar         = CursorStyle(tcell.CursorStyleSteadyBar)
)

// SetCursorStyle is used to set the cursor style. If the style
// is not supported (or cursor styles are not supported at all),
// then this will have no effect.
func SetCursorStyle(style CursorStyle) {
	termbox.Screen().SetCursorStyle(tcell.CursorStyle(style))
}
