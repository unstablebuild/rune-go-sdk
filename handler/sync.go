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

package handler

import (
	"sync"

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Sync wraps a tui.Handler to provide access synchronization with mu.
func Sync(mu sync.Locker, h tui.Handler) tui.Handler {
	return hsync{mu: mu, h: h}
}

type hsync struct {
	mu sync.Locker
	h  tui.Handler
}

func (s hsync) Resize(width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.h.Resize(width, height)
}

func (s hsync) Draw(w term.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.h.Draw(w)
}

func (s hsync) Handle(ev term.Event) (exit, handled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.h.Handle(ev)
}

func (s hsync) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.h.Cursor()
}

func (s hsync) Selection() (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.h.Selection()
}
