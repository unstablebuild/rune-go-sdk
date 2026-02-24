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
	"context"
	"errors"
	"io"
	"slices"
	"strings"
	"sync"

	"github.com/unstablebuild/rune-go-sdk/api/storageapi"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/storagestub"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/debug"
	"github.com/unstablebuild/rune-go-sdk/handler"
	"github.com/unstablebuild/rune-go-sdk/handler/inputbox"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"github.com/unstablebuild/rune-go-sdk/term"
)

// Handler is a read-eval-print loop handler that owns
// an output container (top) and an inputbox handler
// (bottom). It dispatches commands asynchronously via
// a CommandHandler and streams output into the
// container.
type Handler struct {
	output *component.Virtual[*component.Container]
	rl     *handler.Virtual[*inputbox.Handler]
	cmd    CommandHandler
	prompt string

	storage    storageapi.Service
	storageKey string
	history    []string

	tabStyle inputbox.TabStyle
	attr     term.Attributes
	hasAttr  bool

	cmdCtx    context.Context
	cmdCancel context.CancelFunc

	wg sync.WaitGroup

	width, height int
}

// New creates a new Handler with the given command
// handler and options.
func New(cmd CommandHandler, opts ...Option) *Handler {
	h := &Handler{
		output: &component.Virtual[*component.Container]{
			C: component.NewContainer(),
		},
		cmd:    cmd,
		prompt: "> ",
	}
	for _, o := range opts {
		o(h)
	}
	if h.storage == nil {
		h.storage = storagestub.NewInMemoryService()
		h.storageKey = "repl-history"
	}
	h.loadHistory()
	h.cmdCtx, h.cmdCancel = context.WithCancel(context.Background())
	h.rl = h.newInputBox()
	return h
}

// Resize satisfies tui.Component.
func (h *Handler) Resize(width, height int) {
	h.width = width
	h.height = height
	h.layout()
}

// Draw satisfies tui.Component.
func (h *Handler) Draw(w term.Writer) {
	h.output.Draw(w)
	h.rl.Draw(w)
}

// Handle satisfies tui.Handler.
func (h *Handler) Handle(ev term.Event) (exit, handled bool) {
	exit, handled = h.rl.Handle(ev)
	if !exit {
		h.layout()
		return false, handled
	}

	text, err := h.rl.C.Result()
	switch {
	case errors.Is(err, io.EOF):
		return true, true
	case errors.Is(err, inputbox.ErrAborted):
		h.cmdCancel()
		h.cmdCtx, h.cmdCancel = context.WithCancel(context.Background())
		h.addLine("^C")
		h.output.C.ScrollToBottom()
		h.resetInputBox()
		return false, true
	}

	if strings.TrimSpace(text) == "" {
		h.resetInputBox()
		return false, true
	}
	h.history = append(h.history, text)
	h.saveHistory()
	h.addLine(h.prompt + text)
	h.output.C.ScrollToBottom()
	h.dispatchCommand(text)
	h.resetInputBox()
	return false, true
}

// Cursor satisfies tui.Handler.
func (h *Handler) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	return h.rl.Cursor()
}

// Selection satisfies tui.Handler.
func (h *Handler) Selection() (string, bool) {
	return "", false
}

// Wait blocks until all dispatched command goroutines
// have finished.
func (h *Handler) Wait() {
	h.wg.Wait()
}

func (h *Handler) newInputBox() *handler.Virtual[*inputbox.Handler] {
	opts := []inputbox.Option{
		inputbox.WithPrompt(h.prompt),
		inputbox.WithWordCompleter(h.makeCompleter()),
		inputbox.WithTabStyle(h.tabStyle),
		inputbox.WithCtrlCAborts(),
		inputbox.WithHistory(h.history),
	}
	if h.hasAttr {
		opts = append(opts, inputbox.WithAttributes(h.attr))
	}
	ib := inputbox.New(opts...)
	return &handler.Virtual[*inputbox.Handler]{
		Virtual: component.Virtual[*inputbox.Handler]{
			C: ib,
		},
	}
}

func (h *Handler) resetInputBox() {
	h.rl.C.Reset()
	h.rl.C.SetHistory(h.history)
	h.layout()
}

func (h *Handler) loadHistory() {
	var doc historyDoc
	if err := h.storage.Get(context.Background(), h.storageKey, &doc); err != nil {
		return
	}
	h.history = doc.Items
}

func (h *Handler) saveHistory() {
	_ = h.storage.Set(context.Background(), h.storageKey, &historyDoc{Items: h.history})
}

func (h *Handler) addLine(msg string) {
	row := h.output.C.AddRow()
	row.AddComponent(
		component.NewResponsiveString(
			msg,
			component.StringResponsiveConfig{},
		),
		component.MaxCols,
	)
}

func (h *Handler) layout() {
	if h.width == 0 || h.height == 0 {
		return
	}
	rlH := min(h.rl.C.Height(h.width), h.height)
	outH := h.height - rlH

	contentH := h.output.C.Height(h.width)
	if contentH < outH {
		h.output.Resize(h.width, contentH)
		h.output.Move(term.Coordinates{Y: outH - contentH})
	} else {
		h.output.Resize(h.width, outH)
		h.output.Move(term.Coordinates{})
	}

	h.rl.Resize(h.width, rlH)
	h.rl.Move(term.Coordinates{Y: outH})
}

func (h *Handler) dispatchCommand(text string) {
	cmd := parseCommand(text)
	ctx := h.cmdCtx
	// All state mutations are serialized onto the event
	// loop via ScheduleNextTick, so no additional
	// synchronization is needed.
	h.wg.Add(1)
	go debug.CapturePanicReport(func() {
		defer h.wg.Done()
		iter, err := h.cmd.HandleCommand(ctx, cmd)
		if err != nil {
			term.ScheduleNextTick(func() {
				h.addLine(err.Error())
				h.output.C.ScrollToBottom()
				h.layout()
			})
			return
		}
		defer func() { _ = iter.Close() }()
		for {
			item, ok := iter.Next(ctx)
			if !ok {
				break
			}
			term.ScheduleNextTick(func() {
				row := h.output.C.AddRow()
				row.AddComponent(
					item, component.MaxCols,
				)
				h.output.C.ScrollToBottom()
				h.layout()
			})
		}
		if err := iter.Err(); err != nil {
			term.ScheduleNextTick(func() {
				h.addLine(err.Error())
				h.output.C.ScrollToBottom()
				h.layout()
			})
		}
	})
}

func (h *Handler) makeCompleter() inputbox.WordCompleter {
	return func(line string, pos int) (string, []string, string) {
		head := line[:pos]
		tail := line[pos:]
		parts := strings.Fields(head)
		var cmd string
		var args []string
		if len(parts) > 0 {
			cmd = parts[0]
			if len(parts) > 1 {
				args = parts[1:]
			}
		}
		lastSpace := strings.LastIndex(head, " ")
		if lastSpace >= 0 {
			head = head[:lastSpace+1]
		} else {
			head = ""
		}
		iter, err := h.cmd.Complete(h.cmdCtx, cmd, args)
		if err != nil {
			return line[:pos], nil, tail
		}
		defer func() { _ = iter.Close() }()
		candidates, err := iterator.ToSlice(h.cmdCtx, iter)
		if err != nil {
			return line[:pos], nil, tail
		}
		return head, candidates, tail
	}
}

func parseCommand(text string) Command {
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return Command{}
	}
	return Command{
		Name: parts[0],
		Args: slices.Clone(parts[1:]),
	}
}

type historyDoc struct {
	Items []string
}
