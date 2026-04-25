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
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/handler/inputbox"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/tcell/v3"
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

// WithMaxHistory sets the maximum number of command history entries
// retained in memory and persisted in storage.
func WithMaxHistory(max int) Option {
	return func(h *Handler) {
		h.maxHistory = max
	}
}

// WithSuccessAttributes sets the prompt prefix color on
// command success. Default: green foreground.
func WithSuccessAttributes(attr term.Attributes) Option {
	return func(h *Handler) {
		h.successAttr = attr
	}
}

// WithErrorAttributes sets the prompt prefix color on
// command failure and the color of error messages.
// Default: red foreground.
func WithErrorAttributes(attr term.Attributes) Option {
	return func(h *Handler) {
		h.errorAttr = attr
	}
}

// WithValidCommandAttributes sets inputbox attributes
// when the typed command is a valid executable.
// Default: base attributes with bold.
func WithValidCommandAttributes(attr term.Attributes) Option {
	return func(h *Handler) {
		h.validCmdAttr = attr
		h.hasValidCmdAttr = true
	}
}

// WithRunningAnimationFrames sets the spinner animation
// shown while a command is running. Default:
// component.ProgressAnimationFrames().
func WithRunningAnimationFrames(frames []string, sequence []int) Option {
	return func(h *Handler) {
		h.animFrames = frames
		h.animSequence = sequence
	}
}

// WithProgressBarCharSet sets the characters used to
// render the progress bar. Default:
// component.DefaultProgressBarCharSet().
func WithProgressBarCharSet(chars component.ProgressBarCharSet) Option {
	return func(h *Handler) {
		h.progressChars = chars
	}
}

// WithExitError configures an error that, when
// returned by HandleCommand (detected via errors.Is),
// causes the REPL to exit on the next call to Handle.
func WithExitError(err error) Option {
	return func(h *Handler) {
		h.exitError = err
	}
}

// defaultSuccessAttr returns the default success attributes.
func defaultSuccessAttr() term.Attributes {
	return term.Attributes{Fg: tcell.ColorGreen}
}

// defaultErrorAttr returns the default error attributes.
func defaultErrorAttr() term.Attributes {
	return term.Attributes{Fg: tcell.ColorRed}
}

// defaultValidCmdAttr returns the default valid command
// attributes (green foreground + bold).
func defaultValidCmdAttr() term.Attributes {
	return term.Attributes{
		Fg:    tcell.ColorGreen,
		Attrs: tcell.AttrBold,
	}
}
