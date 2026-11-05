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

// Reference is the handler counterpart to component.Reference: it
// delegates to a handler that can be swapped, and does nothing while it
// is nil.
type Reference struct {
	component.Reference
	handler tui.Handler
}

// NewReference allocates storage for a new Reference and initializes it.
func NewReference(h tui.Handler) *Reference {
	r := new(Reference)
	r.Init(h)
	return r
}

// Init sets the handler this Reference delegates to.
func (r *Reference) Init(h tui.Handler) {
	r.handler = h
	r.Reference.Init(h)
}

// Content returns the handler this Reference delegates to.
func (r *Reference) Content() tui.Handler { return r.handler }

// Handle satisfies tui.Handler.
func (r *Reference) Handle(ev term.Event) (bool, bool) {
	if r.handler == nil {
		return false, false
	}
	return r.handler.Handle(ev)
}

// Cursor satisfies tui.Handler.
func (r *Reference) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	if r.handler == nil {
		return term.Coordinates{}, 0, false
	}
	return r.handler.Cursor()
}

// Selection satisfies tui.Handler.
func (r *Reference) Selection() (string, bool) {
	if r.handler == nil {
		return "", false
	}
	return r.handler.Selection()
}
