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

package fileexplorer

import (
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/term"
)

// Option configures a Handler.
type Option func(*Handler)

// WithAttributes sets the default attributes for the file explorer.
func WithAttributes(attr term.Attributes) Option {
	return func(h *Handler) {
		h.attrs = attr
	}
}

// WithCursorAttributes sets the attributes for the cursor line.
func WithCursorAttributes(attr term.Attributes) Option {
	return func(h *Handler) {
		h.cursorAttrs = attr
	}
}

// WithOnOpen sets a callback that is invoked when a file is opened.
func WithOnOpen(fn func(uri workspaceapi.URI)) Option {
	return func(h *Handler) {
		h.onOpen = fn
	}
}

// WithOnExit sets a callback that is invoked when the handler exits.
func WithOnExit(fn func()) Option {
	return func(h *Handler) {
		h.onExit = fn
	}
}

// WithOnApply sets a callback that is invoked after changes are applied.
func WithOnApply(fn func(error)) Option {
	return func(h *Handler) {
		h.onApply = fn
	}
}
