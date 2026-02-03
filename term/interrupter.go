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
	"time"
)

// Interrupter wraps the basic method Interrupt, which
// sends an interrupt event to the main loop, forcing a redraw
// of all components.
//
// The given context is piped back into the next loop iteration
// so callers can use it to distinguish between an interrupt-driven
// call to Draw or just the next tick.
type Interrupter interface {
	Interrupt(context.Context) error
}

// NopInterrupter is an interrupter that does nothing when Interrupt
// is called.
func NopInterrupter() Interrupter {
	return FuncInterrupter(func(context.Context) error { return nil })
}

// FuncInterrupter returns an Interrupter that calls fn
// every time Interrupt is called.
func FuncInterrupter(fn func(context.Context) error) Interrupter {
	return fnInterrupter{fn: fn}
}

type fnInterrupter struct {
	fn func(context.Context) error
}

func (i fnInterrupter) Interrupt(ctx context.Context) error {
	return i.fn(ctx)
}

// InterruptAt interrupts the main event loop at the given fps, using the
// given interrupter. This function only returns when context is canceled.
func InterruptAt(ctx context.Context, interrupter Interrupter, fps int) {
	cadence := time.Duration(int(time.Second) / fps)
	ticker := time.NewTicker(cadence)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = interrupter.Interrupt(ctx)
		case <-ctx.Done():
			return
		}
	}
}
