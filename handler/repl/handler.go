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
	"github.com/unstablebuild/rune-go-sdk/tui"
)

var historyRetryStrategy = retry.SequentialStrategy(30 * time.Millisecond)

const defaultMaxHistory = 2000

// Handler is a read-eval-print loop handler that owns
// an output container (top) and an inputbox handler
// (bottom). It dispatches commands asynchronously via
// a CommandHandler and streams output into the
// container.
//
// Commands run concurrently, each under its own
// context. Every running command has one row between
// the output and the input, oldest at the top, showing
// the command line and its progress bar, or a spinner
// until it reports progress. The row goes away when
// the command ends.
//
// Ctrl-C with text in the input clears it. With an
// empty input it cancels the newest running command.
// Either way it prints ^C.
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

	animFrames    []string
	animSequence  []int
	progressChars component.ProgressBarCharSet

	runTitleAttr    term.Attributes
	hasRunTitleAttr bool

	// runs are the commands that have not ended yet,
	// oldest first, and rows draw the newest of them
	// that fit. Only the event loop touches them.
	runs []*run
	rows []component.Virtual[*runRow]

	ctx    context.Context
	cancel context.CancelFunc

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
	if !h.hasRunTitleAttr {
		h.runTitleAttr = defaultRunTitleAttr()
	}
	if h.animFrames == nil {
		h.animFrames, h.animSequence = component.ProgressAnimationFrames()
	}
	if h.progressChars == (component.ProgressBarCharSet{}) {
		h.progressChars = component.DefaultProgressBarCharSet()
	}
	h.ctx, h.cancel = context.WithCancel(context.Background())
	h.loadHistory()
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
	for i := range h.rows {
		h.rows[i].Draw(w)
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
	hadText := h.rl.C.Text() != ""
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
		if !hadText && len(h.runs) > 0 {
			h.runs[len(h.runs)-1].cancel()
		}
		h.addLine("^C")
		h.output.C.ScrollToBottom()
		h.resetInputBox()
		return false, true
	}

	if strings.TrimSpace(text) != "" {
		h.history = append(h.history, text)
		h.saveHistory(text)
	}
	h.start(text)
	h.resetInputBox()
	return false, true
}

// Dispatch runs line as if the user had typed it and
// pressed Enter: it echoes the prompt line, streams the
// command's output and colours the prompt by the
// outcome. Unlike Enter, it leaves the input line as it
// is and does not add line to the history. Dispatch must
// be called on the event loop.
func (h *Handler) Dispatch(line string) {
	h.start(line)
}

// Clear empties the output. Running commands keep their
// rows, and the output they produce after Clear lands in
// the emptied output. Clear must be called on the event
// loop.
func (h *Handler) Clear() {
	h.output.C = component.NewContainer()
	h.m.Init(&outputMouseDelegate{h: h})
	h.layout()
}

// CancelAll cancels the context of every running
// command. Their rows go away once their commands
// return. CancelAll must be called on the event loop.
func (h *Handler) CancelAll() {
	for _, r := range h.runs {
		r.cancel()
	}
}

func (h *Handler) start(text string) {
	h.addPromptLine(text)
	if strings.TrimSpace(text) != "" {
		h.dispatchCommand(text)
	}
	h.output.C.ScrollToBottom()
	h.layout()
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

// Text returns the current text in the inputbox, i.e. whatever the
// user has typed at the current prompt but not yet submitted.
func (h *Handler) Text() string {
	return h.rl.C.Text()
}

// Close cancels all in-flight commands and animations,
// then waits for goroutines to finish.
func (h *Handler) Close() error {
	h.cancel()
	for _, r := range h.runs {
		_ = r.spinner.Close()
	}
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
	n := h.visibleRuns(outH)
	outputH := outH - n

	contentH := h.output.C.Height(h.width)
	if contentH < outputH {
		h.output.Resize(h.width, contentH)
		h.output.Move(term.Coordinates{Y: outputH - contentH})
	} else {
		h.output.Resize(h.width, outputH)
		h.output.Move(term.Coordinates{})
	}

	h.rows = h.rows[:0]
	for i, r := range h.runs[len(h.runs)-n:] {
		row := component.Virtual[*runRow]{C: &runRow{r: r, attr: h.runTitleAttr}}
		row.Resize(h.width, 1)
		row.Move(term.Coordinates{Y: outputH + i})
		h.rows = append(h.rows, row)
	}

	h.rl.Resize(h.width, rlH)
	h.rl.Move(term.Coordinates{Y: outH})
}

// layoutHeights returns the height of the input and of
// everything above it, which holds the output and the
// rows of running commands.
func (h *Handler) layoutHeights() (rlH, outH int) {
	if h.width == 0 || h.height == 0 {
		return 0, 0
	}
	rlH = min(h.rl.C.Height(h.width), h.height)
	outH = h.height - rlH
	return rlH, outH
}

// visibleRuns returns how many of the newest runs get a
// row above the input, leaving at least one row to the
// output.
func (h *Handler) visibleRuns(outH int) int {
	return max(min(len(h.runs), outH-1), 0)
}

// LayoutHeights computes the inputbox and output layout heights The first return value is
// the height allocated to the inputbox (rl), and the second is the height allocated to
// previous command output and the rows of running commands, which sit at its bottom.
func (h *Handler) LayoutHeights() (rlH, outH int) {
	return h.layoutHeights()
}

// run is a dispatched command that has not ended yet.
type run struct {
	title   string
	cancel  context.CancelFunc
	spinner *component.Animation
	bar     *component.ProgressBar
}

// spinnerCols is the width a run's spinner takes at the
// end of its row.
const spinnerCols = 2

// runRow draws a run on one row: its command line on the
// left, truncated to fit, and either its progress bar in
// the remaining width or its spinner in the last columns.
type runRow struct {
	r      *run
	attr   term.Attributes
	title  []rune
	status component.Virtual[tui.Component]
}

// Resize satisfies tui.Component.
func (rr *runRow) Resize(width, height int) {
	if rr.r.bar != nil {
		rr.title = truncate([]rune(rr.r.title), width/2)
		x := len(rr.title) + 1
		rr.status.C = rr.r.bar
		rr.status.Resize(max(width-x, 0), height)
		rr.status.Move(term.Coordinates{X: x})
		return
	}
	rr.title = truncate([]rune(rr.r.title), width-spinnerCols-1)
	rr.status.C = rr.r.spinner
	rr.status.Resize(1, height)
	rr.status.Move(term.Coordinates{X: width - spinnerCols})
}

// Draw satisfies tui.Component.
func (rr *runRow) Draw(w term.Writer) {
	for x, c := range rr.title {
		w.SetCell(term.Coordinates{X: x}, term.NewCell(c, 1, rr.attr))
	}
	rr.status.Draw(w)
}

// truncate shortens s to at most n runes, marking a cut
// with an ellipsis.
func truncate(s []rune, n int) []rune {
	switch {
	case len(s) <= n:
		return s
	case n <= 0:
		return nil
	}
	return append(s[:n-1:n-1], '…')
}

// endRun removes r once its command has returned.
func (h *Handler) endRun(r *run) {
	h.runs = slices.DeleteFunc(h.runs, func(x *run) bool { return x == r })
	r.cancel()
	_ = r.spinner.Close()
}

type progressWriter struct {
	h *Handler
	r *run
}

func (pw *progressWriter) Progress(
	progress, total int64, units string,
) {
	pw.h.scheduleNextTick(func() {
		pw.h.updateProgress(pw.r, progress, total, units)
	})
}

func (h *Handler) updateProgress(
	r *run, progress, total int64, units string,
) {
	if !slices.Contains(h.runs, r) {
		return
	}
	if r.bar == nil {
		r.bar = component.NewProgressBar(h.progressChars, term.Attributes{})
	}
	r.bar.SetProgress(progress, total, units)
	h.layout()
}

func (h *Handler) dispatchCommand(text string) {
	cmd := parseCommand(text)
	ctx, cancel := context.WithCancel(h.ctx)
	r := &run{
		title:  text,
		cancel: cancel,
		spinner: component.NewAnimation(
			h.interrupter,
			h.animFrames, h.animSequence, 10,
		),
	}
	h.runs = append(h.runs, r)
	promptLine := h.lastPromptLine
	h.wg.Add(1)
	go debug.CapturePanicReport(func() {
		defer h.wg.Done()
		pw := &progressWriter{h: h, r: r}
		iter, err := h.cmd.HandleCommand(ctx, cmd, pw)
		if err != nil {
			isExit := h.exitError != nil && errors.Is(err, h.exitError)
			if isExit {
				h.exitPending.Store(true)
			}
			h.scheduleNextTick(func() {
				h.endRun(r)
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

			h.endRun(r)
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
		iter, err := h.cmd.Complete(h.ctx, cmd, args)
		if err != nil {
			h.showCompletionError(err)
			return line[:pos], nil, tail
		}
		defer func() { _ = iter.Close() }()
		candidates, err := iterator.ToSlice(h.ctx, iter)
		if err != nil {
			h.showCompletionError(err)
			return line[:pos], nil, tail
		}
		return head, candidates, tail
	}
}

// showCompletionError surfaces a tab-completion failure inline in the
// output band, the same way a failed command does, instead of silently
// dropping it. The word completer runs synchronously on the event-loop
// goroutine inside Handle, so the output mutation is safe without
// scheduleNextTick.
func (h *Handler) showCompletionError(err error) {
	h.addErrorLine(err.Error())
	h.output.C.ScrollToBottom()
	h.layout()
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
