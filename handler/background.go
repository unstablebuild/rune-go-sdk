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
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Background is the handler counterpart to component.Background: it
// delegates events to the handler it draws on a background.
type Background struct {
	component.Background
	handler tui.Handler
}

// NewBackground allocates storage for a new Background and initializes it.
func NewBackground(h tui.Handler, cell term.Cell) *Background {
	b := new(Background)
	b.Init(h, cell)
	return b
}

// Init initializes this Background with h and cell.
func (b *Background) Init(h tui.Handler, cell term.Cell) {
	b.handler = h
	b.Background.Init(h, cell)
}

// Content returns the handler this Background draws.
func (b *Background) Content() tui.Handler { return b.handler }

// Handle satisfies tui.Handler.
func (b *Background) Handle(ev term.Event) (bool, bool) { return b.handler.Handle(ev) }

// Cursor satisfies tui.Handler.
func (b *Background) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	return b.handler.Cursor()
}

// Selection satisfies tui.Handler.
func (b *Background) Selection() (string, bool) { return b.handler.Selection() }
