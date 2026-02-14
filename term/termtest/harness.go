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

// Package termtest provides a test harness for running tui.Handler
// implementations in a non-interactive environment.
package termtest

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Harness runs a tui.Handler in a non-interactive loop,
// reading key sequences from an io.Reader and writing
// rendered output to an io.Writer.
type Harness struct {
	handler       tui.Handler
	writer        *term.StringWriter
	width, height int
}

// NewHarness creates a harness for the given handler.
func NewHarness(h tui.Handler, width, height int) *Harness {
	w := term.NewStringWriter(width, height)
	h.Resize(width, height)
	return &Harness{
		handler: h,
		writer:  w,
		width:   width,
		height:  height,
	}
}

// Run reads key sequences (one per line, term.ParseKeys format)
// from input, sends them to the handler, and writes rendered
// output to output after each sequence.
// Each output is separated by "---\n".
func (h *Harness) Run(input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)
	first := true

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Skip comment lines
		if strings.HasPrefix(line, "#") {
			continue
		}

		out, err := h.Step(line)
		if err != nil {
			return fmt.Errorf("step %q: %w", line, err)
		}

		if !first {
			if _, err := fmt.Fprint(output, "---\n"); err != nil {
				return err
			}
		}
		first = false

		if _, err := fmt.Fprint(output, out); err != nil {
			return err
		}
		if _, err := fmt.Fprint(output, "\n"); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// Step processes a single key sequence and returns the rendered output.
func (h *Harness) Step(sequence string) (string, error) {
	keys, err := term.ParseKeys(sequence)
	if err != nil {
		return "", fmt.Errorf("parse keys: %w", err)
	}

	for _, key := range keys {
		exit, _ := h.handler.Handle(term.Event{
			Type: term.EventKey,
			Ch:   key.Ch,
			Mod:  key.Mod,
			Key:  key.Key,
		})
		if exit {
			break
		}
	}

	return h.Render(), nil
}

// Render renders the current handler state and returns it as a string.
func (h *Harness) Render() string {
	if err := h.writer.Clear(term.Attributes{}); err != nil {
		return ""
	}

	h.handler.Draw(h.writer)

	cursor, _, show := h.handler.Cursor()
	if show {
		h.writer.SetCursor(cursor)
	}

	if err := h.writer.Flush(); err != nil {
		return ""
	}

	return h.writer.String()
}

// Handler returns the underlying handler.
func (h *Harness) Handler() tui.Handler {
	return h.handler
}

// Resize resizes the handler and internal writer.
func (h *Harness) Resize(width, height int) {
	h.width, h.height = width, height
	h.writer.Resize(width, height)
	h.handler.Resize(width, height)
}
