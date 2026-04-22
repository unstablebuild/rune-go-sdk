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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
	"github.com/unstablebuild/tcell/v3"
)

// fakeScreen is a minimal term.Screen implementation for tests. It
// avoids tcell.SimulationScreen because the latter's PostEvent
// recurses infinitely in v3.6.3.
type fakeScreen struct {
	mu       sync.Mutex
	width    int
	height   int
	evch     chan tcell.Event
	cursorX  int
	cursorY  int
	cursorOn bool
}

func newFakeScreen(w, h int) *fakeScreen {
	return &fakeScreen{width: w, height: h, evch: make(chan tcell.Event, 16)}
}

func (s *fakeScreen) SetContent(int, int, rune, []rune, uint8, tcell.Style) {}
func (s *fakeScreen) UnionStyle(int, int, tcell.Style)                      {}
func (s *fakeScreen) Fill(rune, tcell.Style)                                {}
func (s *fakeScreen) ShowCursor(x, y int) {
	s.mu.Lock()
	s.cursorX, s.cursorY = x, y
	s.cursorOn = x >= 0 && y >= 0
	s.mu.Unlock()
}
func (s *fakeScreen) HideCursor()                  { s.ShowCursor(-1, -1) }
func (s *fakeScreen) SetCursorStyle(tcell.CursorStyle) {}
func (s *fakeScreen) Size() (int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.width, s.height
}
func (s *fakeScreen) Show()                        {}
func (s *fakeScreen) Poll() <-chan tcell.Event     { return s.evch }
func (s *fakeScreen) PostEvent(ev tcell.Event) error {
	select {
	case s.evch <- ev:
		return nil
	default:
		return tcell.ErrEventQFull
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

func TestRunScreenDrawsAndResizes(t *testing.T) {
	s := newFakeScreen(80, 24)
	h := &recHandler{exitOn: func(ev term.Event) bool { return ev.Key == term.KeyEsc }}

	done := make(chan error, 1)
	go func() { done <- tui.RunScreen(h, s) }()

	// Initial Draw + Resize from screen.Size().
	waitFor(t, func() bool {
		d, r, _ := h.snapshot()
		if d < 1 || len(r) < 1 {
			return false
		}
		return r[0] == [2]int{80, 24}
	}, time.Second)

	// Post a resize event via the writer (which goes through PostEvent).
	s.setSize(40, 12)
	w := term.NewTermboxWriterFromScreen(s)
	require.True(t, w.PublishEvent(term.Event{
		Type: term.EventResize, Width: 40, Height: 12,
	}))
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
		t.Fatal("RunScreen did not exit after Esc")
	}
}

func TestRunScreenHandlesKey(t *testing.T) {
	s := newFakeScreen(80, 24)
	h := &recHandler{exitOn: func(ev term.Event) bool { return ev.Ch == 'q' }}

	done := make(chan error, 1)
	go func() { done <- tui.RunScreen(h, s) }()

	w := term.NewTermboxWriterFromScreen(s)
	require.True(t, w.PublishEvent(term.Event{Type: term.EventKey, Ch: 'q'}))

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("RunScreen did not exit on 'q'")
	}

	_, _, events := h.snapshot()
	require.NotEmpty(t, events)
	assert.Equal(t, rune('q'), events[len(events)-1].Ch)
}

// TestWriterInterruptCoalescing verifies that back-to-back payload-less
// interrupt events posted via (*term.TermboxWriter).PublishEvent coalesce
// to a single pending event. The flag is cleared once the loop delivers
// the interrupt.
func TestWriterInterruptCoalescing(t *testing.T) {
	s := newFakeScreen(80, 24)
	w := term.NewTermboxWriterFromScreen(s)

	// First payload-less interrupt is accepted.
	assert.True(t, w.PublishEvent(term.Event{Type: term.EventInterrupt}))
	// Second one coalesces (returns true but is not enqueued again).
	assert.True(t, w.PublishEvent(term.Event{Type: term.EventInterrupt}))

	// Only one interrupt event should be pending on the screen.
	var got int
	timeout := time.After(100 * time.Millisecond)
loop:
	for {
		select {
		case ev := <-s.Poll():
			if _, ok := ev.(*tcell.EventInterrupt); ok {
				got++
			}
		case <-timeout:
			break loop
		}
	}
	assert.Equal(t, 1, got, "expected exactly one coalesced interrupt")
}

// TestRunScreenInterruptResetsPending runs the loop and verifies
// that after an interrupt is processed the writer's coalescing flag
// is reset so new payload-less interrupts can be published.
func TestRunScreenInterruptResetsPending(t *testing.T) {
	s := newFakeScreen(80, 24)

	delivered := make(chan struct{}, 4)
	h := &recHandler{exitOn: func(ev term.Event) bool { return ev.Key == term.KeyEsc }}

	extW := term.NewTermboxWriterFromScreen(s)

	done := make(chan error, 1)
	go func() { done <- tui.RunScreen(h, s) }()

	// Let the loop start.
	waitFor(t, func() bool { d, _, _ := h.snapshot(); return d >= 1 }, time.Second)

	// Post an interrupt carrying a UserFunc so we observe delivery.
	require.True(t, extW.PublishEvent(term.Event{
		Type: term.EventInterrupt,
		UserFunc: func() {
			delivered <- struct{}{}
		},
	}))

	select {
	case <-delivered:
	case <-time.After(time.Second):
		t.Fatal("interrupt UserFunc never ran")
	}

	// After delivery, a second interrupt with a UserFunc must still run.
	require.True(t, extW.PublishEvent(term.Event{
		Type: term.EventInterrupt,
		UserFunc: func() {
			delivered <- struct{}{}
		},
	}))

	select {
	case <-delivered:
	case <-time.After(time.Second):
		t.Fatal("second interrupt never delivered (pending flag not reset)")
	}

	require.True(t, extW.PublishEvent(term.Event{Type: term.EventKey, Key: term.KeyEsc}))
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("RunScreen did not exit after Esc")
	}
}

// TestRunScreenConcurrentIsolation runs two loops against two screens
// concurrently and verifies events posted to one screen never reach
// the other loop's handler.
func TestRunScreenConcurrentIsolation(t *testing.T) {
	s1 := newFakeScreen(80, 24)
	s2 := newFakeScreen(80, 24)
	w1 := term.NewTermboxWriterFromScreen(s1)
	w2 := term.NewTermboxWriterFromScreen(s2)

	exitKey := func(ev term.Event) bool { return ev.Key == term.KeyEsc }
	h1 := &recHandler{exitOn: exitKey}
	h2 := &recHandler{exitOn: exitKey}

	d1 := make(chan error, 1)
	d2 := make(chan error, 1)
	go func() { d1 <- tui.RunScreen(h1, s1) }()
	go func() { d2 <- tui.RunScreen(h2, s2) }()

	waitFor(t, func() bool {
		a, _, _ := h1.snapshot()
		b, _, _ := h2.snapshot()
		return a >= 1 && b >= 1
	}, time.Second)

	require.True(t, w1.PublishEvent(term.Event{Type: term.EventKey, Ch: 'a'}))
	require.True(t, w2.PublishEvent(term.Event{Type: term.EventKey, Ch: 'b'}))

	waitFor(t, func() bool {
		_, _, e1 := h1.snapshot()
		_, _, e2 := h2.snapshot()
		haveA := false
		for _, ev := range e1 {
			if ev.Ch == 'a' {
				haveA = true
			}
		}
		haveB := false
		for _, ev := range e2 {
			if ev.Ch == 'b' {
				haveB = true
			}
		}
		return haveA && haveB
	}, time.Second)

	_, _, e1 := h1.snapshot()
	_, _, e2 := h2.snapshot()
	for _, ev := range e1 {
		assert.NotEqual(t, rune('b'), ev.Ch, "loop1 saw an event meant for loop2")
	}
	for _, ev := range e2 {
		assert.NotEqual(t, rune('a'), ev.Ch, "loop2 saw an event meant for loop1")
	}

	require.True(t, w1.PublishEvent(term.Event{Type: term.EventKey, Key: term.KeyEsc}))
	require.True(t, w2.PublishEvent(term.Event{Type: term.EventKey, Key: term.KeyEsc}))

	for _, done := range []<-chan error{d1, d2} {
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("a RunScreen loop did not exit")
		}
	}
}
