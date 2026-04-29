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
	"sync"

	"github.com/unstablebuild/rune-go-sdk/term"
)

// Option is an option to be passed to Run.
type Option func(*config)

// WithDefaultAttributes sets the default foreground and background attributes.
func WithDefaultAttributes(attr term.Attributes) Option {
	return func(cfg *config) {
		cfg.defAttr = attr
	}
}

// WithLocker configures the event loop to use the given locker
// before calling any of the methods of the root tui.Handler passed
// to Run.
func WithLocker(locker sync.Locker) Option {
	return func(cfg *config) {
		cfg.locker = locker
	}
}

// WithInputMode sets the keyboard input mode.
func WithInputMode(inputMode term.InputMode) Option {
	return func(cfg *config) {
		cfg.inputMode = inputMode
	}
}

type config struct {
	defAttr   term.Attributes
	locker    sync.Locker
	inputMode term.InputMode
}

func defaultConfig() config {
	return config{
		locker: nopLocker{},
	}
}

type nopLocker struct{}

func (l nopLocker) Lock() {
}
func (l nopLocker) Unlock() {
}
