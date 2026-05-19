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
	"time"

	"github.com/unstablebuild/rune-go-sdk/term"
)

// Run initializes the process-wide default terminal, then runs the TUI
// event loop with root as the top-level Handler.
//
// It configures the terminal (default attributes and input mode) from
// opts, polls term.Events, and dispatches them to root:
//   - EventResize calls root.Resize.
//   - EventInterrupt updates the iteration context and invokes any
//     attached UserFunc (under the handler lock) without advancing the
//     iteration id.
//   - All other events are forwarded to root.Handle; if Handle returns
//     exit=true, the loop stops.
//
// SIGINT is intercepted: a single Ctrl-C is delivered as usual, but two
// within one second cause Run to return. The terminal is restored via
// term.Close before Run returns.
//
// Run returns nil on a clean exit, or the first error encountered while
// drawing, flushing, or polling events. To run against a caller-owned
// Screen/writer, use RunWriter instead.
func Run(root Handler, opts ...Option) (err error) {
	if err != nil {
		return fmt.Errorf("initialize term: %w", err)
	}
	cfg := defaultConfig()
	for _, o := range opts {
		o(&cfg)
	}
	writer, err := term.NewDefaultWriter()
	term.SetAttr(cfg.defAttr)
	term.SetInputMode(cfg.inputMode)
	defer term.Close()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	defer signal.Stop(sigs)

	return run(root, cfg.locker, writer, cfg.defAttr, sigs)
}

// RunWriter runs the TUI event loop with root as the top-level Handler,
// rendering through the caller-owned ScreenWriter.
//
// Unlike Run, RunWriter does not initialize or finalize the terminal and
// does not install a SIGINT handler; the caller owns the Screen lifecycle.
// Specifically, the caller is responsible for:
//   - Constructing the writer with a Screen, and managing its lifecycle.
//   - Enabling paste/focus/mouse and setting any input modes as desired.
//   - Terminating the loop by closing the underlying Screen (which closes
//     the event channel returned by writer.Poll) or by having a Handler
//     return exit=true.
//
// Options configure the default attributes and the locker used to serialize
// Handler calls. Event dispatch (Resize, Interrupt, Handle) matches Run.
//
// RunWriter panics if writer is nil. It returns nil on clean exit, or the
// first error encountered while drawing, flushing, or polling events.
func RunWriter(root Handler, writer *term.ScreenWriter, opts ...Option) error {
	if writer == nil {
		panic("tui: RunWriter called with nil writer")
	}
	cfg := defaultConfig()
	for _, o := range opts {
		o(&cfg)
	}
	return run(root, cfg.locker, writer, cfg.defAttr, nil)
}

const exitSignalDuration = 1 * time.Second

func redraw(
	root Handler, lock sync.Locker, termw *term.ScreenWriter,
	prevCursor term.CursorStyle, defAttr term.Attributes,
) (term.CursorStyle, error) {
	if err := termw.Clear(defAttr); err != nil {
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
		termw.SetCursorStyle(style)
	}

	if err := termw.Flush(); err != nil {
		return 0, err
	}

	return style, nil
}

func drain(evs <-chan term.Event) {
	for {
		select {
		case <-evs:
		default:
			return
		}
	}
}

// PublishEvent publishes the given event to the event loop.
// It targets the process-wide default writer; callers that run the
// loop via RunWriter should use (*term.TermboxWriter).PublishEvent.
func PublishEvent(ev term.Event) bool {
	return term.PublishEvent(ev)
}

func handleInterruptSignal(
	lastSignalAt *time.Time, exit *bool, evs <-chan term.Event,
) {
	now := time.Now()
	shouldExit := now.Sub(*lastSignalAt) < exitSignalDuration
	*exit = shouldExit
	*lastSignalAt = now
	if shouldExit {
		drain(evs)
	}
}

func run(
	root Handler, lock sync.Locker, termw *term.ScreenWriter,
	defAttr term.Attributes, sigs <-chan os.Signal,
) (err error) {
	ctx := context.Background()
	width, height := termw.Size()

	lock.Lock()
	root.Resize(width, height)
	lock.Unlock()

	evs := termw.Poll()

	var exit bool
	var lastSignalAt time.Time
	var prevCursor term.CursorStyle
	var i int64
	termw.SetContext(ContextWithIteration(ctx, i))

loop:
	for !exit && err == nil {
		if prevCursor, err = redraw(root, lock, termw, prevCursor, defAttr); err != nil {
			return
		}

		for {
			select {
			case <-sigs:
				handleInterruptSignal(&lastSignalAt, &exit, evs)
			case ev, ok := <-evs:
				if !ok {
					exit = true
					break
				}
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
						lock.Lock()
						ev.UserFunc()
						lock.Unlock()
					}
					termw.ClearInterruptPending()
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
