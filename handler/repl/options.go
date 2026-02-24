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

package repl

import (
	"github.com/unstablebuild/rune-go-sdk/api/storageapi"
	"github.com/unstablebuild/rune-go-sdk/handler/inputbox"
	"github.com/unstablebuild/rune-go-sdk/term"
)

// Option configures a Handler.
type Option func(*Handler)

// WithPrompt sets the prompt string displayed
// before each input line.
func WithPrompt(prompt string) Option {
	return func(h *Handler) {
		h.prompt = prompt
	}
}

// WithTabStyle sets the tab completion style.
func WithTabStyle(style inputbox.TabStyle) Option {
	return func(h *Handler) {
		h.tabStyle = style
	}
}

// WithAttributes sets the display attributes.
func WithAttributes(attr term.Attributes) Option {
	return func(h *Handler) {
		h.attr = attr
		h.hasAttr = true
	}
}

// WithStorage sets the backing store for command
// history. Defaults to storagestub.NewInMemoryService().
func WithStorage(key string, s storageapi.Service) Option {
	return func(h *Handler) {
		h.storageKey = key
		h.storage = s
	}
}
