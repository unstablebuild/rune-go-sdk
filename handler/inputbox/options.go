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

package inputbox

import (
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/tcell/v3"
)

// Option configures a Handler.
type Option func(*Handler)

// WithAttributes defines the default input text attributes.
func WithAttributes(attr term.Attributes) Option {
	return func(h *Handler) {
		h.attrs = attr
	}
}

// WithPlaceholder adds a placeholder when there's no text in the input box.
func WithPlaceholder(text string) Option {
	cfg := component.StringResponsiveConfig{
		NoSplitWords: true,
		StringConfig: component.StringConfig{
			Attributes: term.Attributes{
				Fg: tcell.ColorGray,
			},
		},
	}
	return func(h *Handler) {
		h.placeholder = component.NewResponsiveString(text, cfg)
	}
}
