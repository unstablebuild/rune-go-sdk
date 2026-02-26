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

// Package debugger implements a repl.CommandHandler that drives
// a DAP-compatible debugger through the debugapi.Debugger interface.
package debugger

import (
	"context"
	"strings"

	"github.com/google/go-dap"
	"github.com/unstablebuild/rune-go-sdk/api/debugapi"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/handler/repl"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

var _ repl.CommandHandler = (*Handler)(nil)

// trackedBreakpoint stores a DAP breakpoint alongside the
// user-supplied condition, which DAP reports separately from
// Breakpoint.Message.
type trackedBreakpoint struct {
	dap.Breakpoint
	condition string
}

// Handler implements repl.CommandHandler by dispatching
// commands to a debugapi.Debugger.
type Handler struct {
	dbg          debugapi.Debugger
	capabilities *dap.Capabilities

	threadID   int
	frameID    int
	frameIndex int

	// breakpoints tracks set breakpoints keyed by source path.
	breakpoints map[string][]trackedBreakpoint

	commands []command
}

// New creates a new Handler wrapping the given debugger
// with the capabilities returned from Initialize.
func New(dbg debugapi.Debugger, caps *dap.Capabilities) *Handler {
	h := &Handler{
		dbg:          dbg,
		capabilities: caps,
		breakpoints:  make(map[string][]trackedBreakpoint),
	}
	h.commands = h.buildCommands()
	return h
}

// SelectFirstThread queries the adapter for available threads
// and selects the first one. Call this after Launch/Attach +
// ConfigurationDone so that execution commands use a valid
// thread ID from the start.
func (h *Handler) SelectFirstThread(ctx context.Context) error {
	threads, err := h.dbg.Threads(ctx)
	if err != nil {
		return err
	}
	if len(threads) > 0 {
		h.threadID = threads[0].Id
	}
	return nil
}

// HandleCommand dispatches a REPL command to the
// underlying debugger.
func (h *Handler) HandleCommand(
	ctx context.Context, cmd repl.Command,
) (iterator.Iterator[component.Responsive], error) {
	if cmd.Name == "" {
		return iterator.Empty[component.Responsive](), nil
	}
	c, ok := h.findCommand(cmd.Name)
	if !ok {
		return nil, errUnknownCommand(cmd.Name)
	}
	return c.fn(ctx, cmd.Args)
}

// Complete returns tab-completion candidates.
func (h *Handler) Complete(
	ctx context.Context, cmd string, args []string,
) (iterator.Iterator[string], error) {
	if len(args) == 0 {
		return h.completeCommand(cmd), nil
	}
	return iterator.Empty[string](), nil
}

func (h *Handler) findCommand(name string) (command, bool) {
	name = strings.ToLower(name)
	for _, c := range h.commands {
		if c.name == name {
			return c, true
		}
		for _, a := range c.aliases {
			if a == name {
				return c, true
			}
		}
	}
	return command{}, false
}

func (h *Handler) completeCommand(prefix string) iterator.Iterator[string] {
	prefix = strings.ToLower(prefix)
	var matches []string
	for _, c := range h.commands {
		if strings.HasPrefix(c.name, prefix) {
			matches = append(matches, c.name)
		}
	}
	return iterator.FromSlice(matches)
}
