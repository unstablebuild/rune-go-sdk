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

package tui

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/tcell/v3"
)

// Run takes the given root handler, renders it full-screen,
// and starts feeding it with term.Events.
// Error is non-nil if there were any errors.
func Run(root Handler, opts ...Option) (err error) {
	err = term.Init()
	if err != nil {
		return fmt.Errorf("initialize term: %w", err)
	}
	cfg := defaultConfig()
	for _, o := range opts {
		o(&cfg)
	}
	term.SetAttr(cfg.defAttr)
	term.SetInputMode(cfg.inputMode)
	defer term.Close()
	return run(root, cfg.locker, cfg.writer)
}

const exitSignalDuration = 1 * time.Second

func redraw(
	root Handler, lock sync.Locker, termw *term.TermboxWriter,
	prevCursor term.CursorStyle,
) (term.CursorStyle, error) {
	if err := termw.Clear(term.Attr()); err != nil {
		return 0, err
	}

	lock.Lock()
	root.Draw(termw)
	cursor, style, show := root.Cursor()
	lock.Unlock()

	if show {
		termw.SetCursor(cursor)
	} else {
		termw.SetCursor(term.Coordinates{X: -1, Y: -1})
	}
	if style != prevCursor {
		term.SetCursorStyle(style)
	}

	if err := termw.Flush(); err != nil {
		return 0, err
	}

	return style, nil
}

func drain(evs <-chan tcell.Event) {
	for {
		select {
		case <-evs:
		default:
			return
		}
	}
}

// batch interrupt events such that we deliver exactly one more
// after every call to publish interrupt.
// This serves as a pressure valve when something is abusing
// the interrupt mechanism.
var interruptPending atomic.Bool

// PublishEvent publishes the given event to the event loop.
func PublishEvent(ev term.Event) bool {
	// only conflate interrupts that have no payload
	if ev.Type == term.EventInterrupt && ev.Raw == nil &&
		ev.UserFunc == nil && !interruptPending.CompareAndSwap(false, true) {
		return true
	}
	return term.PublishEvent(ev)
}

func handleInterruptSignal(
	lastSignalAt *time.Time, exit *bool, evs <-chan tcell.Event,
) {
	now := time.Now()
	shouldExit := now.Sub(*lastSignalAt) < exitSignalDuration
	*exit = shouldExit
	*lastSignalAt = now
	if shouldExit {
		drain(evs)
	}
}

func run(root Handler, lock sync.Locker, termw *term.TermboxWriter) (err error) {
	ctx := context.Background()
	width, height := term.Size()

	lock.Lock()
	root.Resize(width, height)
	lock.Unlock()

	evs := term.Poll()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	defer signal.Stop(sigs)

	var exit bool
	var lastSignalAt time.Time
	var prevCursor term.CursorStyle
	var i int64
	termw.SetContext(ContextWithIteration(ctx, i))

loop:
	for !exit && err == nil {
		if prevCursor, err = redraw(root, lock, termw, prevCursor); err != nil {
			return
		}

		for {
			select {
			case <-sigs:
				handleInterruptSignal(&lastSignalAt, &exit, evs)
			case tev := <-evs:
				ev := term.FromTcellEvent(tev)
				switch ev.Type {
				case term.EventInterrupt:
					if ev.Context != nil {
						if id, ok := IterationFromContext(ev.Context); ok {
							termw.SetContext(ContextWithIteration(ctx, id))
						} else {
							termw.SetContext(ev.Context)
						}
					} else if ev.Raw != nil {
						termw.SetContext(term.ContextWithPayload(ctx, ev.Raw))
					} else {
						// do not set a new iteration id, instead reset to nil
						// so clients can differentiate between an interrupt
						// and a regular iteration loop.
						termw.SetContext(ctx)
					}
					if ev.UserFunc != nil {
						ev.UserFunc()
					}
					interruptPending.Store(false)
					// ensure that i is not incremented
					// and context is not overwritten
					continue loop
				case term.EventError:
					err = ev.Err
				case term.EventResize:
					width, height := ev.Width, ev.Height
					lock.Lock()
					root.Resize(width, height)
					lock.Unlock()
				default:
					lock.Lock()
					exit, _ = root.Handle(ev)
					lock.Unlock()
				}
			}
			if exit || len(evs) == 0 {
				i++
				termw.SetContext(ContextWithIteration(ctx, i))
				break
			}
		}
	}

	return err
}
