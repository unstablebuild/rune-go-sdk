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
	"os/exec"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/unstablebuild/rune-go-sdk/api/storageapi"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/storagestub"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/debug"
	"github.com/unstablebuild/rune-go-sdk/handler"
	"github.com/unstablebuild/rune-go-sdk/handler/inputbox"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"github.com/unstablebuild/rune-go-sdk/mouse"
	"github.com/unstablebuild/rune-go-sdk/retry"
	"github.com/unstablebuild/rune-go-sdk/term"
)

var historyRetryStrategy = retry.SequentialStrategy(30 * time.Millisecond)

const defaultMaxHistory = 2000

// Handler is a read-eval-print loop handler that owns
// an output container (top) and an inputbox handler
// (bottom). It dispatches commands asynchronously via
// a CommandHandler and streams output into the
// container.
type Handler struct {
	output *component.Virtual[*component.Container]
	rl     *handler.Virtual[*inputbox.Handler]
	m      *mouse.Mouse
	cmd    CommandHandler
	prompt string

	scheduleNextTick func(func()) bool
	interrupter      term.Interrupter

	storage    storageapi.Service
	storageKey string
	history    []string
	maxHistory int

	tabStyle inputbox.TabStyle
	attr     term.Attributes
	hasAttr  bool

	// Prompt coloring.
	successAttr    term.Attributes
	errorAttr      term.Attributes
	lastPromptLine *promptLine

	// Valid command bold input.
	validCmdAttr    term.Attributes
	hasValidCmdAttr bool
	lastCheckedCmd  string
	lastCmdValid    bool

	// Spinner animation.
	spinner      *component.Animation
	spinnerVirt  component.Virtual[*component.Animation]
	animFrames   []string
	animSequence []int

	// Progress bar.
	progressBar     *component.ProgressBar
	progressBarVirt component.Virtual[*component.ProgressBar]
	progressChars   component.ProgressBarCharSet

	ctx    context.Context
	cancel context.CancelFunc

	cmdCtx    context.Context
	cmdCancel context.CancelFunc

	// exitError, when non-nil, is compared with
	// errors.Is against errors returned by
	// HandleCommand. A match causes the REPL to
	// exit on the next call to Handle.
	exitError   error
	exitPending atomic.Bool

	// wg tracks in-flight command goroutines launched by
	// dispatchCommand. Add is only called from Handle
	// (the event-loop goroutine) and Wait is only called
	// after the event loop exits, so Add and Wait
	// should never be called concurrently
	wg sync.WaitGroup

	width, height int
}

type outputMouseDelegate struct {
	h *Handler
}

var _ mouse.Delegate = (*outputMouseDelegate)(nil)

func (d *outputMouseDelegate) OnAction(term.Event, term.Coordinates, mouse.Action) bool {
	return false
}

func (d *outputMouseDelegate) ScrollUp(n int) bool {
	handled := false
	for range n {
		if !d.h.output.C.ScrollUp() {
			break
		}
		handled = true
	}
	return handled
}

func (d *outputMouseDelegate) ScrollDown(n int) bool {
	handled := false
	for range n {
		if !d.h.output.C.ScrollDown() {
			break
		}
		handled = true
	}
	return handled
}

func (d *outputMouseDelegate) SetSelectionEnd(term.Coordinates)   {}
func (d *outputMouseDelegate) SetSelectionStart(term.Coordinates) {}
func (d *outputMouseDelegate) ClearSelection()                    {}
func (d *outputMouseDelegate) SelectWordAt(term.Coordinates)      {}
func (d *outputMouseDelegate) SelectLine(int)                     {}
func (d *outputMouseDelegate) Width() int                         { return d.h.width }

func (d *outputMouseDelegate) Height() int {
	_, outH := d.h.layoutHeights()
	return outH
}

// New creates a new Handler with the given command
// handler and options. The scheduleNextTick function
// schedules a callback on the next event-loop tick.
// The interrupter forces a redraw of the event loop.
func New(
	cmd CommandHandler,
	scheduleNextTick func(func()) bool,
	interrupter term.Interrupter,
	opts ...Option,
) *Handler {
	h := &Handler{
		output: &component.Virtual[*component.Container]{
			C: component.NewContainer(),
		},
		cmd:              cmd,
		scheduleNextTick: scheduleNextTick,
		interrupter:      interrupter,
		prompt:           "> ",
	}
	for _, o := range opts {
		o(h)
	}
	if h.storage == nil {
		h.storage = storagestub.NewInMemoryService()
		h.storageKey = "repl-history"
	}
	if h.maxHistory <= 0 {
		h.maxHistory = defaultMaxHistory
	}
	if h.successAttr == (term.Attributes{}) {
		h.successAttr = defaultSuccessAttr()
	}
	if h.errorAttr == (term.Attributes{}) {
		h.errorAttr = defaultErrorAttr()
	}
	if !h.hasValidCmdAttr {
		h.validCmdAttr = defaultValidCmdAttr()
	}
	if h.animFrames == nil {
		h.animFrames, h.animSequence = component.ProgressAnimationFrames()
	}
	if h.progressChars == (component.ProgressBarCharSet{}) {
		h.progressChars = component.DefaultProgressBarCharSet()
	}
	h.ctx, h.cancel = context.WithCancel(context.Background())
	h.loadHistory()
	h.cmdCtx, h.cmdCancel = context.WithCancel(h.ctx)
	h.rl = h.newInputBox()
	h.m = mouse.New(&outputMouseDelegate{h: h})
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
	if h.progressBar != nil {
		h.progressBarVirt.Draw(w)
	} else if h.spinner != nil {
		h.spinnerVirt.Draw(w)
	}
	h.rl.Draw(w)
}

// Handle satisfies tui.Handler.
func (h *Handler) Handle(ev term.Event) (exit, handled bool) {
	if h.exitPending.Load() {
		return true, true
	}
	if ev.Type == term.EventMouse {
		return h.handleMouse(ev)
	}
	exit, handled = h.rl.Handle(ev)
	if !exit {
		h.updateInputStyle()
		h.layout()
		return false, handled
	}

	text, err := h.rl.C.Result()
	switch {
	case errors.Is(err, io.EOF):
		return true, true
	case errors.Is(err, inputbox.ErrAborted):
		h.cmdCancel()
		h.cmdCtx, h.cmdCancel = context.WithCancel(h.ctx)
		h.addLine("^C")
		h.output.C.ScrollToBottom()
		h.resetInputBox()
		return false, true
	}

	if strings.TrimSpace(text) == "" {
		h.addPromptLine("")
		h.output.C.ScrollToBottom()
		h.resetInputBox()
		return false, true
	}
	h.history = append(h.history, text)
	h.saveHistory(text)
	h.addPromptLine(text)
	h.output.C.ScrollToBottom()
	h.dispatchCommand(text)
	h.resetInputBox()
	return false, true
}

func (h *Handler) handleMouse(ev term.Event) (exit, handled bool) {
	rlH, outH := h.layoutHeights()
	switch {
	case ev.MouseY >= 0 && ev.MouseY < outH:
		return h.m.Handle(ev)
	case ev.MouseY >= outH && ev.MouseY < outH+rlH:
		ev.MouseY -= outH
		return h.rl.Handle(ev)
	default:
		return false, false
	}
}

// Cursor satisfies tui.Handler.
func (h *Handler) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	return h.rl.Cursor()
}

// Selection satisfies tui.Handler.
func (h *Handler) Selection() (string, bool) {
	return "", false
}

// Close cancels all in-flight commands and animations,
// then waits for goroutines to finish.
func (h *Handler) Close() error {
	h.cancel()
	h.stopSpinner()
	h.wg.Wait()
	return nil
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
	h.lastCheckedCmd = ""
	h.lastCmdValid = false
	h.applyInputStyle()
	h.layout()
}

func (h *Handler) loadHistory() {
	var doc historyDoc
	if err := h.storage.Get(context.Background(), h.storageKey, &doc); err != nil {
		return
	}
	h.history = h.boundHistory(doc.Items)
}

func (h *Handler) saveHistory(text string) {
	ctx := context.Background()
	bounded := h.boundHistory(h.history)
	for {
		doc := historyDoc{}
		err := storageapi.ConsistentUpdate(
			ctx, h.storage, h.storageKey, &doc, historyRetryStrategy,
			func() ([]storageapi.Update, []storageapi.Precondition) {
				items := h.boundHistory(append(slices.Clone(doc.Items), text))
				updates := []storageapi.Update{
					{FieldPath: []string{"Items"}, Value: items},
					{FieldPath: []string{"Version"}, Value: effectiveHistoryVersion(doc) + 1},
				}

				preconds := make([]storageapi.Precondition, 0, 1)
				if doc.Version > 0 {
					preconds = append(preconds, storageapi.Precondition{
						FieldPath: []string{"Version"},
						Value:     doc.Version,
					})
				}
				return updates, preconds
			},
		)
		if err == nil {
			h.history = doc.Items
			return
		}
		if errors.Is(err, storageapi.ErrNotFound) {
			err = h.storage.Create(ctx, h.storageKey, &historyDoc{
				Items:   bounded,
				Version: 1,
			})
			if err == nil {
				h.history = bounded
				return
			}
			if errors.Is(err, storageapi.ErrAlreadyExists) {
				continue
			}
			return
		}
		return
	}
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

func (h *Handler) addPromptLine(text string) {
	pl := newPromptLine(
		h.prompt, text,
		term.Attributes{}, term.Attributes{},
	)
	h.lastPromptLine = pl
	row := h.output.C.AddRow()
	row.AddComponent(pl, component.MaxCols)
}

func (h *Handler) addErrorLine(msg string) {
	row := h.output.C.AddRow()
	row.AddComponent(
		component.NewResponsiveString(
			msg,
			component.StringResponsiveConfig{
				StringConfig: component.StringConfig{
					Attributes: h.errorAttr,
				},
			},
		),
		component.MaxCols,
	)
}

func (h *Handler) layout() {
	if h.width == 0 || h.height == 0 {
		return
	}
	rlH, outH := h.layoutHeights()

	contentH := h.output.C.Height(h.width)
	if contentH < outH {
		h.output.Resize(h.width, contentH)
		h.output.Move(term.Coordinates{Y: outH - contentH})
	} else {
		h.output.Resize(h.width, outH)
		h.output.Move(term.Coordinates{})
	}

	if h.spinner != nil {
		h.spinnerVirt.Resize(1, 1)
		h.spinnerVirt.Move(term.Coordinates{
			X: h.width - 2,
			Y: outH - 1,
		})
	}
	if h.progressBar != nil {
		h.progressBarVirt.Resize(h.width, 1)
		h.progressBarVirt.Move(term.Coordinates{
			X: 0,
			Y: outH - 1,
		})
	}

	h.rl.Resize(h.width, rlH)
	h.rl.Move(term.Coordinates{Y: outH})
}

func (h *Handler) layoutHeights() (rlH, outH int) {
	if h.width == 0 || h.height == 0 {
		return 0, 0
	}
	rlH = min(h.rl.C.Height(h.width), h.height)
	outH = h.height - rlH
	return rlH, outH
}

// LayoutHeights computes the inputbox and output layout heights The first return value is
// the height allocated to the inputbox (rl), and the second is the height allocated to
// previous command output.
func (h *Handler) LayoutHeights() (rlH, outH int) {
	return h.layoutHeights()
}

func (h *Handler) startSpinner() {
	h.spinner = component.NewAnimation(
		h.interrupter,
		h.animFrames, h.animSequence, 10,
	)
	h.spinnerVirt.C = h.spinner
	h.layout()
}

func (h *Handler) stopSpinner() {
	if h.spinner != nil {
		_ = h.spinner.Close()
		h.spinner = nil
		h.spinnerVirt.C = nil
	}
	if h.progressBar != nil {
		h.progressBar = nil
		h.progressBarVirt.C = nil
	}
}

type progressWriter struct {
	h *Handler
}

func (pw *progressWriter) Progress(
	progress, total int64, units string,
) {
	pw.h.scheduleNextTick(func() {
		pw.h.updateProgress(progress, total, units)
	})
}

func (h *Handler) updateProgress(
	progress, total int64, units string,
) {
	if h.progressBar == nil {
		h.progressBar = component.NewProgressBar(
			h.progressChars, term.Attributes{},
		)
		h.progressBarVirt.C = h.progressBar
	}
	h.progressBar.SetProgress(progress, total, units)
	h.layout()
}

func (h *Handler) dispatchCommand(text string) {
	cmd := parseCommand(text)
	ctx := h.cmdCtx
	promptLine := h.lastPromptLine
	h.wg.Add(1)
	go debug.CapturePanicReport(func() {
		defer h.wg.Done()
		h.scheduleNextTick(func() {
			h.startSpinner()
		})
		pw := &progressWriter{h: h}
		iter, err := h.cmd.HandleCommand(ctx, cmd, pw)
		if err != nil {
			isExit := h.exitError != nil && errors.Is(err, h.exitError)
			if isExit {
				h.exitPending.Store(true)
			}
			h.scheduleNextTick(func() {
				h.stopSpinner()
				if !isExit {
					promptLine.setPromptAttr(h.errorAttr)
					h.addErrorLine(err.Error())
				}
				h.output.C.ScrollToBottom()
				h.layout()
			})
			return
		}
		defer func() { _ = iter.Close() }()

		// Batch output items so we schedule at most one
		// drain tick at a time, keeping the event queue
		// shallow and ctrl-c responsive.
		var mu sync.Mutex
		var pending []component.Responsive
		drainScheduled := false

		var drain func()
		drain = func() {
			mu.Lock()
			batch := pending
			pending = nil
			drainScheduled = false
			mu.Unlock()

			for _, it := range batch {
				row := h.output.C.AddRow()
				row.AddComponent(it, component.MaxCols)
			}
			// New output always snaps the REPL viewport back to the bottom,
			// even after manual mouse-wheel scrolling.
			h.output.C.ScrollToBottom()
			h.layout()

			mu.Lock()
			if len(pending) > 0 && !drainScheduled {
				drainScheduled = true
				h.scheduleNextTick(drain)
			}
			mu.Unlock()
		}

		for {
			item, ok := iter.Next(ctx)
			if !ok {
				break
			}
			mu.Lock()
			pending = append(pending, item)
			if !drainScheduled {
				drainScheduled = true
				h.scheduleNextTick(drain)
			}
			mu.Unlock()
		}

		iterErr := iter.Err()
		isExit := h.exitError != nil &&
			iterErr != nil &&
			errors.Is(iterErr, h.exitError)
		if isExit {
			h.exitPending.Store(true)
		}
		h.scheduleNextTick(func() {
			// Flush remaining items that arrived after
			// the last drain tick.
			mu.Lock()
			batch := pending
			pending = nil
			mu.Unlock()
			for _, it := range batch {
				row := h.output.C.AddRow()
				row.AddComponent(it, component.MaxCols)
			}

			h.stopSpinner()
			if iterErr != nil && !isExit {
				promptLine.setPromptAttr(h.errorAttr)
				h.addErrorLine(iterErr.Error())
			} else if iterErr == nil {
				promptLine.setPromptAttr(h.successAttr)
			}
			// Final command output follows the same snap-to-bottom behavior as
			// streamed output above.
			h.output.C.ScrollToBottom()
			h.layout()
		})
	})
}

func (h *Handler) updateInputStyle() {
	text := h.rl.C.Text()
	cmd := firstWord(text)
	if cmd == h.lastCheckedCmd {
		return
	}
	h.lastCheckedCmd = cmd
	h.lastCmdValid = false
	if cmd != "" {
		_, err := exec.LookPath(cmd)
		h.lastCmdValid = err == nil
	}
	h.applyInputStyle()
}

func (h *Handler) applyInputStyle() {
	if h.lastCmdValid {
		cmdLen := len([]rune(h.lastCheckedCmd))
		h.rl.C.SetHighlight(0, cmdLen, h.validCmdAttr)
	} else {
		h.rl.C.ClearHighlight()
	}
}

func firstWord(s string) string {
	if i := strings.IndexByte(s, ' '); i >= 0 {
		return s[:i]
	}
	return s
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
			if len(parts) > 1 || (len(parts) == 1 && strings.HasSuffix(head, " ")) {
				if len(parts) > 1 {
					args = parts[1:]
				}
				// If cursor is right after a space, append
				// empty string so the last arg is the
				// (empty) prefix.
				if strings.HasSuffix(head, " ") {
					args = append(args, "")
				}
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
	Items   []string
	Version int64
}

func (h *Handler) boundHistory(items []string) []string {
	if h.maxHistory > 0 && len(items) > h.maxHistory {
		items = items[len(items)-h.maxHistory:]
	}
	return slices.Clone(items)
}

func effectiveHistoryVersion(doc historyDoc) int64 {
	if doc.Version > 0 {
		return doc.Version
	}
	return 0
}
