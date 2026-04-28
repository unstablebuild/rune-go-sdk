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
// The placeholder uses the provided StringResponsiveConfig for styling.
func WithPlaceholder(text string, cfg component.StringResponsiveConfig) Option {
	return func(h *Handler) {
		h.placeholder = component.NewResponsiveString(text, cfg)
	}
}

// WithPlaceholderText adds a placeholder with default gray styling.
func WithPlaceholderText(text string) Option {
	cfg := component.StringResponsiveConfig{
		NoSplitWords: true,
		StringConfig: component.StringConfig{
			Attributes: term.Attributes{
				Fg: tcell.ColorGray,
			},
		},
	}
	return WithPlaceholder(text, cfg)
}

// WithPrompt sets the prompt string displayed before
// the input text on the first line.
func WithPrompt(prompt string) Option {
	return func(h *Handler) {
		h.prompt = []rune(prompt)
		h.promptWidth = len(h.prompt)
	}
}

// WithText pre-populates the input text and places
// the cursor at the end.
func WithText(text string) Option {
	return func(h *Handler) {
		h.text = []rune(text)
		h.cursor = len(h.text)
	}
}

// WithHistory pre-populates the history.
func WithHistory(items []string) Option {
	return func(h *Handler) {
		h.SetHistory(items)
	}
}

// WithCompleter sets a line completer for tab
// completion.
func WithCompleter(c Completer) Option {
	return func(h *Handler) {
		h.completer = func(line string, _ int) (string, []string, string) {
			return c(line)
		}
		h.tabStyle = TabCircular
	}
}

// WithWordCompleter sets a word completer for tab
// completion.
func WithWordCompleter(c WordCompleter) Option {
	return func(h *Handler) {
		h.completer = c
		h.tabStyle = TabCircular
	}
}

// WithTabStyle sets the tab completion style.
func WithTabStyle(style TabStyle) Option {
	return func(h *Handler) {
		h.tabStyle = style
	}
}

// WithCtrlCAborts makes Ctrl+C abort the input
// (returning ErrAborted) instead of clearing the line.
func WithCtrlCAborts() Option {
	return func(h *Handler) {
		h.ctrlCAborts = true
	}
}

// WithRedact enables redacted rendering: each character of the buffer is
// drawn as the default redaction glyph ('*'). The underlying buffer is
// kept intact: Text(), Result() and Selection() still return the real
// content. Useful for password / secret prompts.
//
// Use WithRedactRune to override the glyph.
func WithRedact(redact bool) Option {
	return func(h *Handler) {
		h.redact = redact
		if h.redactRune == 0 {
			h.redactRune = defaultRedactRune
		}
	}
}

// WithRedactRune sets the rune used when redact rendering is enabled.
// Implies WithRedact(true). The zero rune resets to the default ('*').
func WithRedactRune(r rune) Option {
	return func(h *Handler) {
		h.redact = true
		if r == 0 {
			r = defaultRedactRune
		}
		h.redactRune = r
	}
}
