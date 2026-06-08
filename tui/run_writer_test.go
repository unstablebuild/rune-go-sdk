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

package tui_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// fakeScreen is a minimal term.Screen implementation for tests.
type fakeScreen struct {
	mu       sync.Mutex
	width    int
	height   int
	evch     chan term.Event
	cursorX  int
	cursorY  int
	cursorOn bool
}

func newFakeScreen(w, h int) *fakeScreen {
	return &fakeScreen{width: w, height: h, evch: make(chan term.Event, 16)}
}

func (s *fakeScreen) SetContent(int, int, rune, []rune, uint8, term.Style) {}
func (s *fakeScreen) UnionStyle(int, int, term.Style)                      {}
func (s *fakeScreen) Fill(rune, term.Style)                                {}
func (s *fakeScreen) ShowCursor(x, y int) {
	s.mu.Lock()
	s.cursorX, s.cursorY = x, y
	s.cursorOn = x >= 0 && y >= 0
	s.mu.Unlock()
}
func (s *fakeScreen) HideCursor()                     { s.ShowCursor(-1, -1) }
func (s *fakeScreen) SetCursorStyle(term.CursorStyle) {}
func (s *fakeScreen) Size() (int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.width, s.height
}
func (s *fakeScreen) Show()                   {}
func (s *fakeScreen) Poll() <-chan term.Event { return s.evch }
func (s *fakeScreen) PostEvent(ev term.Event) error {
	select {
	case s.evch <- ev:
		return nil
	default:
		return term.ErrEventQFull
	}
}
func (s *fakeScreen) Bell() {}

func (s *fakeScreen) setSize(w, h int) {
	s.mu.Lock()
	s.width, s.height = w, h
	s.mu.Unlock()
}

// recHandler is a minimal tui.Handler that records the calls it receives.
type recHandler struct {
	mu          sync.Mutex
	draws       int
	resizeCalls [][2]int
	events      []term.Event
	exitOn      func(term.Event) bool
}

func (h *recHandler) Draw(term.Writer) {
	h.mu.Lock()
	h.draws++
	h.mu.Unlock()
}

func (h *recHandler) Resize(w, height int) {
	h.mu.Lock()
	h.resizeCalls = append(h.resizeCalls, [2]int{w, height})
	h.mu.Unlock()
}

func (h *recHandler) Handle(ev term.Event) (exit, handled bool) {
	h.mu.Lock()
	h.events = append(h.events, ev)
	h.mu.Unlock()
	if h.exitOn != nil && h.exitOn(ev) {
		return true, true
	}
	return false, true
}

func (h *recHandler) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	return term.Coordinates{}, term.CursorStyleDefault, false
}

func (h *recHandler) Selection() (string, bool) { return "", false }

func (h *recHandler) snapshot() (draws int, resizes [][2]int, events []term.Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	resizes = append(resizes, h.resizeCalls...)
	events = append(events, h.events...)
	return h.draws, resizes, events
}

func waitFor(t *testing.T, cond func() bool, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("condition not met within %s", timeout)
}

// TestRunWriterDrawsAndResizes verifies the new lower-level primitive
// that drives the event loop from an explicit *term.TermboxWriter
// rather than constructing one internally from a Screen.
func TestRunWriterDrawsAndResizes(t *testing.T) {
	s := newFakeScreen(80, 24)
	w := term.NewScreenWriter(s)
	h := &recHandler{exitOn: func(ev term.Event) bool { return ev.Key == term.KeyEsc }}

	done := make(chan error, 1)
	go func() { done <- tui.RunWriter(h, w) }()

	waitFor(t, func() bool {
		d, r, _ := h.snapshot()
		if d < 1 || len(r) < 1 {
			return false
		}
		return r[0] == [2]int{80, 24}
	}, time.Second)

	s.setSize(40, 12)
	require.True(t, w.PublishEvent(term.Event{Type: term.EventResize, Width: 40, Height: 12}))
	waitFor(t, func() bool {
		_, r, _ := h.snapshot()
		for _, dim := range r {
			if dim[0] == 40 && dim[1] == 12 {
				return true
			}
		}
		return false
	}, time.Second)

	require.True(t, w.PublishEvent(term.Event{Type: term.EventKey, Key: term.KeyEsc}))
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("RunWriter did not exit after Esc")
	}
}

// TestRunWriterInterruptResetsPending verifies that the writer passed
// to RunWriter is the same writer whose interruptPending flag is reset
// by the loop after delivery, so payload-less interrupts can continue
// to drive redraws.
func TestRunWriterInterruptResetsPending(t *testing.T) {
	s := newFakeScreen(80, 24)
	w := term.NewScreenWriter(s)
	delivered := make(chan struct{}, 4)
	h := &recHandler{exitOn: func(ev term.Event) bool { return ev.Key == term.KeyEsc }}

	done := make(chan error, 1)
	go func() { done <- tui.RunWriter(h, w) }()

	waitFor(t, func() bool { d, _, _ := h.snapshot(); return d >= 1 }, time.Second)

	require.True(t, w.PublishEvent(term.Event{Type: term.EventInterrupt, UserFunc: func() { delivered <- struct{}{} }}))
	select {
	case <-delivered:
	case <-time.After(time.Second):
		t.Fatal("interrupt UserFunc never ran")
	}

	require.True(t, w.PublishEvent(term.Event{Type: term.EventInterrupt, UserFunc: func() { delivered <- struct{}{} }}))
	select {
	case <-delivered:
	case <-time.After(time.Second):
		t.Fatal("second interrupt never delivered (pending flag not reset)")
	}

	require.True(t, w.PublishEvent(term.Event{Type: term.EventKey, Key: term.KeyEsc}))
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("RunWriter did not exit after Esc")
	}
}
