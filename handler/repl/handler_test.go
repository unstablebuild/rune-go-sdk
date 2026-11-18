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
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/storagestub"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/component/comptest"
	"github.com/unstablebuild/rune-go-sdk/handler/handlertest"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

var _ tui.Handler = (*Handler)(nil)

type testHandler struct {
	handleFn func(
		context.Context, Command, ProgressWriter,
	) (iterator.Iterator[component.Responsive], error)
	completeFn func(
		context.Context, string, []string,
	) (iterator.Iterator[string], error)
}

func (t *testHandler) HandleCommand(
	ctx context.Context, cmd Command, pw ProgressWriter,
) (iterator.Iterator[component.Responsive], error) {
	if t.handleFn != nil {
		return t.handleFn(ctx, cmd, pw)
	}
	return iterator.FromSlice[component.Responsive](nil), nil
}

func (t *testHandler) Complete(
	ctx context.Context, cmd string, args []string,
) (iterator.Iterator[string], error) {
	if t.completeFn != nil {
		return t.completeFn(ctx, cmd, args)
	}
	return iterator.FromSlice[string](nil), nil
}

func nopSchedule(func()) bool { return false }

type queuedScheduler struct {
	mu      sync.Mutex
	pending []func()
}

type channelIterator[T any] struct {
	once    sync.Once
	started chan struct{}
	ch      <-chan T
}

func newChannelIterator[T any](ch <-chan T) *channelIterator[T] {
	return &channelIterator[T]{
		started: make(chan struct{}),
		ch:      ch,
	}
}

func (it *channelIterator[T]) Next(ctx context.Context) (T, bool) {
	it.once.Do(func() { close(it.started) })
	select {
	case item, ok := <-it.ch:
		return item, ok
	case <-ctx.Done():
		var zero T
		return zero, false
	}
}

func (it *channelIterator[T]) Err() error {
	return nil
}

func (it *channelIterator[T]) Close() error {
	return nil
}

func (s *queuedScheduler) ScheduleNextTick(fn func()) bool {
	s.mu.Lock()
	s.pending = append(s.pending, fn)
	s.mu.Unlock()
	return true
}

func (s *queuedScheduler) Flush() {
	for {
		s.mu.Lock()
		if len(s.pending) == 0 {
			s.mu.Unlock()
			return
		}
		fn := s.pending[0]
		s.pending = s.pending[1:]
		s.mu.Unlock()
		fn()
	}
}

// settle runs the callbacks command goroutines schedule,
// as the event loop would, until cond holds.
func settle(t *testing.T, s *queuedScheduler, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		s.Flush()
		if cond() {
			return
		}
		require.True(t, time.Now().Before(deadline), "condition never held")
		time.Sleep(time.Millisecond)
	}
}

// submit enters line and lets its command end.
func submit(t *testing.T, h *Handler, s *queuedScheduler, line string) {
	t.Helper()
	sendKeys(t, h, line+"<enter>")
	h.Wait()
	s.Flush()
}

// blockingCommands runs every command until its context
// is cancelled or finish releases it, in which case it
// prints "<name> done". A command named sync reports
// 25/100 files first.
type blockingCommands struct {
	started chan string

	mu      sync.Mutex
	ctxs    map[string]context.Context
	release map[string]chan struct{}
}

func newBlockingCommands() *blockingCommands {
	return &blockingCommands{
		started: make(chan string, 16),
		ctxs:    map[string]context.Context{},
		release: map[string]chan struct{}{},
	}
}

func (b *blockingCommands) HandleCommand(
	ctx context.Context, cmd Command, pw ProgressWriter,
) (iterator.Iterator[component.Responsive], error) {
	b.mu.Lock()
	b.ctxs[cmd.Name] = ctx
	release := b.releaseLocked(cmd.Name)
	b.mu.Unlock()
	if cmd.Name == "sync" {
		pw.Progress(25, 100, "files")
	}
	b.started <- cmd.Name
	select {
	case <-release:
		return iterator.FromSlice([]component.Responsive{
			component.NewResponsiveString(
				cmd.Name+" done",
				component.StringResponsiveConfig{},
			),
		}), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (b *blockingCommands) Complete(
	context.Context, string, []string,
) (iterator.Iterator[string], error) {
	return iterator.Empty[string](), nil
}

func (b *blockingCommands) releaseLocked(name string) chan struct{} {
	ch, ok := b.release[name]
	if !ok {
		ch = make(chan struct{})
		b.release[name] = ch
	}
	return ch
}

func (b *blockingCommands) waitStarted(n int) {
	for range n {
		<-b.started
	}
}

func (b *blockingCommands) ctx(name string) context.Context {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.ctxs[name]
}

func (b *blockingCommands) finish(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	close(b.releaseLocked(name))
}

// newBlockingREPL returns a Handler running cmds whose
// spinner is a static "*".
func newBlockingREPL(t *testing.T, cmds *blockingCommands, s *queuedScheduler) *Handler {
	t.Helper()
	h := New(
		cmds, s.ScheduleNextTick, term.NopInterrupter(),
		WithPrompt("$ "),
		WithRunningAnimationFrames([]string{"*"}, []int{0}),
	)
	t.Cleanup(func() { _ = h.Close() })
	return h
}

func sendKeys(t *testing.T, h *Handler, seq string) {
	t.Helper()
	keys, err := term.ParseKeys(seq)
	require.NoError(t, err)
	for _, k := range keys {
		h.Handle(term.Event{
			Type: term.EventKey,
			Ch:   k.Ch,
			Mod:  k.Mod,
			Key:  k.Key,
		})
	}
}

func drawHandler(h *Handler, w, ht int) string {
	return handlertest.DrawHandler(h, w, ht)
}

func wheelUpAt(h *Handler, x, y int) (exit, handled bool) {
	return h.Handle(term.Event{
		Type:   term.EventMouse,
		Key:    term.MouseWheelUp,
		MouseX: x,
		MouseY: y,
	})
}

func TestBasicInput(t *testing.T) {
	h := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
	)
	cases := []handlertest.SequenceTestCase{
		{
			InputSequence: "",
			Expected: "                    \n" +
				"                    \n" +
				"                    \n" +
				"                    \n" +
				"$ ▐                 ",
		},
		{
			InputSequence: "hello",
			Expected: "                    \n" +
				"                    \n" +
				"                    \n" +
				"                    \n" +
				"$ hello▐            ",
		},
	}
	handlertest.RunHandlerSequence(t, h, 20, 5, cases)
}

func TestCtrlDEOF(t *testing.T) {
	h := New(&testHandler{}, nopSchedule, term.NopInterrupter())
	h.Resize(20, 5)
	exit, handled := h.Handle(term.Event{
		Type: term.EventKey,
		Ch:   'd',
		Mod:  term.ModCtrl,
	})
	assert.True(t, exit)
	assert.True(t, handled)
}

func TestCtrlCClearsAndShowsCaret(t *testing.T) {
	h := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
	)
	h.Resize(20, 5)
	sendKeys(t, h, "hello")
	exit, handled := h.Handle(term.Event{
		Type: term.EventKey,
		Ch:   'c',
		Mod:  term.ModCtrl,
	})
	assert.False(t, exit)
	assert.True(t, handled)

	got := drawHandler(h, 20, 5)
	assert.Contains(t, got, "^C")
	assert.Contains(t, got, "$ ")
}

func TestCtrlCCancelsNewestCommand(t *testing.T) {
	scheduler := &queuedScheduler{}
	cmds := newBlockingCommands()
	h := newBlockingREPL(t, cmds, scheduler)
	h.Resize(40, 10)

	sendKeys(t, h, "a<enter>b<enter>")
	cmds.waitStarted(2)

	sendKeys(t, h, "<c-c>")
	settle(t, scheduler, func() bool { return len(h.runs) == 1 })
	assert.ErrorIs(t, cmds.ctx("b").Err(), context.Canceled)
	assert.NoError(t, cmds.ctx("a").Err())

	sendKeys(t, h, "<c-c>")
	settle(t, scheduler, func() bool { return len(h.runs) == 0 })
	assert.ErrorIs(t, cmds.ctx("a").Err(), context.Canceled)

	sendKeys(t, h, "<c-c>")
	assert.Equal(t, 3, strings.Count(drawHandler(h, 40, 10), "^C"))
}

func TestCtrlCWithInputOnlyClears(t *testing.T) {
	scheduler := &queuedScheduler{}
	cmds := newBlockingCommands()
	h := newBlockingREPL(t, cmds, scheduler)
	h.Resize(40, 10)

	sendKeys(t, h, "a<enter>")
	cmds.waitStarted(1)

	sendKeys(t, h, "typed<c-c>")
	scheduler.Flush()
	assert.NoError(t, cmds.ctx("a").Err())
	assert.Len(t, h.runs, 1)
	assert.Empty(t, h.Text())
	assert.Contains(t, drawHandler(h, 40, 10), "^C")
}

func TestEmptyInput(t *testing.T) {
	called := false
	th := &testHandler{
		handleFn: func(
			_ context.Context, _ Command, _ ProgressWriter,
		) (
			iterator.Iterator[component.Responsive],
			error,
		) {
			called = true
			return iterator.FromSlice[component.Responsive](nil), nil
		},
	}
	h := New(th, nopSchedule, term.NopInterrupter(), WithPrompt("$ "))
	h.Resize(20, 5)
	sendKeys(t, h, "<space><space><space>")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})
	assert.False(t, called)
}

func TestCommandDispatch(t *testing.T) {
	ch := make(chan Command, 1)
	th := &testHandler{
		handleFn: func(
			_ context.Context, cmd Command, _ ProgressWriter,
		) (
			iterator.Iterator[component.Responsive],
			error,
		) {
			ch <- cmd
			return iterator.FromSlice[component.Responsive](nil), nil
		},
	}
	h := New(th, nopSchedule, term.NopInterrupter(), WithPrompt("$ "))
	h.Resize(40, 10)
	sendKeys(t, h, "echo<space>hello<space>world")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})

	got := <-ch
	assert.Equal(t, "echo", got.Name)
	assert.Equal(t, []string{"hello", "world"}, got.Args)

	out := drawHandler(h, 40, 10)
	assert.Contains(t, out, "$ echo hello world")
}

func TestHistory(t *testing.T) {
	h := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
	)
	h.Resize(30, 5)

	sendKeys(t, h, "first")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})

	sendKeys(t, h, "second")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})

	// Up arrow should recall "second"
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyArrowUp,
	})
	assert.Equal(t, "second", h.rl.C.Text())

	// Another up should recall "first"
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyArrowUp,
	})
	assert.Equal(t, "first", h.rl.C.Text())
}

func TestHistoryPersistsAcrossSharedStorageHandlers(t *testing.T) {
	store := storagestub.NewInMemoryService()
	const historyKey = "shared-history"

	h1 := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
		WithStorage(historyKey, store),
	)
	h2 := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
		WithStorage(historyKey, store),
	)
	t.Cleanup(func() { _ = h1.Close() })
	t.Cleanup(func() { _ = h2.Close() })

	h1.Resize(30, 5)
	h2.Resize(30, 5)

	sendKeys(t, h1, "first")
	h1.Handle(term.Event{Type: term.EventKey, Key: term.KeyEnter})

	sendKeys(t, h2, "second")
	h2.Handle(term.Event{Type: term.EventKey, Key: term.KeyEnter})

	var doc historyDoc
	require.NoError(t, store.Get(context.Background(), historyKey, &doc))
	assert.Equal(t, []string{"first", "second"}, doc.Items)

	h3 := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
		WithStorage(historyKey, store),
	)
	t.Cleanup(func() { _ = h3.Close() })
	h3.Resize(30, 5)

	h3.Handle(term.Event{Type: term.EventKey, Key: term.KeyArrowUp})
	assert.Equal(t, "second", h3.rl.C.Text())
	h3.Handle(term.Event{Type: term.EventKey, Key: term.KeyArrowUp})
	assert.Equal(t, "first", h3.rl.C.Text())
}

func TestHistoryRespectsMaxHistory(t *testing.T) {
	h := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
		WithMaxHistory(2),
	)
	h.Resize(30, 5)

	for _, cmd := range []string{"first", "second", "third"} {
		sendKeys(t, h, cmd)
		h.Handle(term.Event{Type: term.EventKey, Key: term.KeyEnter})
	}

	assert.Equal(t, []string{"second", "third"}, h.history)
	h.Handle(term.Event{Type: term.EventKey, Key: term.KeyArrowUp})
	assert.Equal(t, "third", h.rl.C.Text())
	h.Handle(term.Event{Type: term.EventKey, Key: term.KeyArrowUp})
	assert.Equal(t, "second", h.rl.C.Text())
}

func TestHistoryPersistsAcrossSharedStorageHandlersConcurrentCreate(t *testing.T) {
	store := storagestub.NewInMemoryService()
	const historyKey = "shared-history-concurrent-create"

	const total = 16
	var (
		start sync.WaitGroup
		wg    sync.WaitGroup
	)
	start.Add(1)

	for i := range total {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			h := New(
				&testHandler{},
				nopSchedule, term.NopInterrupter(),
				WithPrompt("$ "),
				WithStorage(historyKey, store),
			)
			defer func() { _ = h.Close() }()
			h.Resize(30, 5)
			start.Wait()
			cmd := "cmd-" + strconv.Itoa(i)
			sendKeys(t, h, cmd)
			h.Handle(term.Event{Type: term.EventKey, Key: term.KeyEnter})
		}(i)
	}
	start.Done()
	wg.Wait()

	var doc historyDoc
	require.NoError(t, store.Get(context.Background(), historyKey, &doc))
	require.Len(t, doc.Items, total)
	assert.Equal(t, int64(total), doc.Version)

	got := append([]string(nil), doc.Items...)
	want := make([]string, total)
	for i := range total {
		want[i] = "cmd-" + strconv.Itoa(i)
	}
	assert.ElementsMatch(t, want, got)
}

func TestHistoryPersistsAcrossSharedStorageHandlersConcurrentCreateWithMaxHistory(t *testing.T) {
	store := storagestub.NewInMemoryService()
	const historyKey = "shared-history-concurrent-max"
	const total = 12
	const maxHistory = 5
	var (
		start sync.WaitGroup
		wg    sync.WaitGroup
	)
	start.Add(1)

	for i := range total {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			h := New(
				&testHandler{},
				nopSchedule, term.NopInterrupter(),
				WithPrompt("$ "),
				WithMaxHistory(maxHistory),
				WithStorage(historyKey, store),
			)
			defer func() { _ = h.Close() }()
			h.Resize(30, 5)
			start.Wait()
			cmd := "cmd-" + strconv.Itoa(i)
			sendKeys(t, h, cmd)
			h.Handle(term.Event{Type: term.EventKey, Key: term.KeyEnter})
		}(i)
	}
	start.Done()
	wg.Wait()

	var doc historyDoc
	require.NoError(t, store.Get(context.Background(), historyKey, &doc))
	require.Len(t, doc.Items, maxHistory)
	assert.Equal(t, int64(total), doc.Version)

	seen := make(map[string]struct{}, len(doc.Items))
	for _, item := range doc.Items {
		_, ok := seen[item]
		assert.False(t, ok, "duplicate item persisted: %s", item)
		seen[item] = struct{}{}
		assert.True(t, strings.HasPrefix(item, "cmd-"), "unexpected item: %s", item)
		n, err := strconv.Atoi(strings.TrimPrefix(item, "cmd-"))
		require.NoError(t, err)
		assert.GreaterOrEqual(t, n, 0)
		assert.Less(t, n, total)
	}
}

func TestHistoryPersistsAcrossSharedStorageHandlersConcurrentVersionedDoc(t *testing.T) {
	store := storagestub.NewInMemoryService()
	const historyKey = "versioned-history-concurrent"
	require.NoError(t, store.Create(context.Background(), historyKey, &historyDoc{
		Items:   []string{"seed"},
		Version: 1,
	}))

	const total = 10
	var (
		start sync.WaitGroup
		wg    sync.WaitGroup
	)
	start.Add(1)

	for i := range total {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			h := New(
				&testHandler{},
				nopSchedule, term.NopInterrupter(),
				WithPrompt("$ "),
				WithStorage(historyKey, store),
			)
			defer func() { _ = h.Close() }()
			h.Resize(30, 5)
			start.Wait()
			cmd := "v-" + strconv.Itoa(i)
			sendKeys(t, h, cmd)
			h.Handle(term.Event{Type: term.EventKey, Key: term.KeyEnter})
		}(i)
	}
	start.Done()
	wg.Wait()

	var doc historyDoc
	require.NoError(t, store.Get(context.Background(), historyKey, &doc))
	require.Len(t, doc.Items, total+1)
	assert.Equal(t, int64(total+1), doc.Version)
	assert.Contains(t, doc.Items, "seed")

	got := make([]string, 0, total)
	for _, item := range doc.Items {
		if item != "seed" {
			got = append(got, item)
		}
	}
	want := make([]string, total)
	for i := range total {
		want[i] = "v-" + strconv.Itoa(i)
	}
	assert.ElementsMatch(t, want, got)
}

func TestHistoryLoadRespectsMaxHistory(t *testing.T) {
	store := storagestub.NewInMemoryService()
	const historyKey = "load-max-history"
	require.NoError(t, store.Create(context.Background(), historyKey, &historyDoc{
		Items:   []string{"one", "two", "three", "four"},
		Version: 4,
	}))

	h := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
		WithMaxHistory(2),
		WithStorage(historyKey, store),
	)
	t.Cleanup(func() { _ = h.Close() })
	h.Resize(30, 5)

	assert.Equal(t, []string{"three", "four"}, h.history)
	h.Handle(term.Event{Type: term.EventKey, Key: term.KeyArrowUp})
	assert.Equal(t, "four", h.rl.C.Text())
	h.Handle(term.Event{Type: term.EventKey, Key: term.KeyArrowUp})
	assert.Equal(t, "three", h.rl.C.Text())
}

func TestTabCompletion(t *testing.T) {
	th := &testHandler{
		completeFn: func(
			_ context.Context,
			_ string, _ []string,
		) (iterator.Iterator[string], error) {
			candidates := []string{
				"foobar", "foobaz",
			}
			return iterator.FromSlice(candidates), nil
		},
	}
	h := New(th, nopSchedule, term.NopInterrupter(), WithPrompt("$ "))
	h.Resize(30, 5)
	sendKeys(t, h, "foo")
	sendKeys(t, h, "<tab>")
	assert.Equal(t, "foobar", h.rl.C.Text())

	sendKeys(t, h, "<tab>")
	assert.Equal(t, "foobaz", h.rl.C.Text())
}

func TestCtrlCCancelsContext(t *testing.T) {
	scheduler := &queuedScheduler{}
	cmds := newBlockingCommands()
	h := newBlockingREPL(t, cmds, scheduler)
	h.Resize(40, 10)

	sendKeys(t, h, "cmd<enter>")
	cmds.waitStarted(1)

	sendKeys(t, h, "<c-c>")
	h.Wait()
	scheduler.Flush()
	assert.ErrorIs(t, cmds.ctx("cmd").Err(), context.Canceled)
	assert.Empty(t, h.runs)

	sendKeys(t, h, "next<enter>")
	cmds.waitStarted(1)
	assert.NoError(t, cmds.ctx("next").Err())
}

func TestSelection(t *testing.T) {
	h := New(&testHandler{}, nopSchedule, term.NopInterrupter())
	h.Resize(20, 5)
	sel, ok := h.Selection()
	assert.False(t, ok)
	assert.Equal(t, "", sel)
}

func TestLayout(t *testing.T) {
	cases := []struct {
		name     string
		running  string
		expected string
	}{
		{
			name: "idle",
			expected: "                    \n" +
				"                    \n" +
				"                    \n" +
				"                    \n" +
				"$ ▐                 ",
		},
		{
			name:    "rows above the input, oldest first",
			running: "a<enter>b<enter>",
			expected: "$ a                 \n" +
				"$ b                 \n" +
				"a                 * \n" +
				"b                 * \n" +
				"$ ▐                 ",
		},
		{
			name:    "newest rows that fit keep one output row",
			running: "a<enter>b<enter>c<enter>d<enter>",
			expected: "$ d                 \n" +
				"b                 * \n" +
				"c                 * \n" +
				"d                 * \n" +
				"$ ▐                 ",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scheduler := &queuedScheduler{}
			cmds := newBlockingCommands()
			h := newBlockingREPL(t, cmds, scheduler)
			h.Resize(20, 5)

			sendKeys(t, h, tc.running)
			cmds.waitStarted(strings.Count(tc.running, "<enter>"))
			scheduler.Flush()

			pos, _, show := h.Cursor()
			assert.True(t, show)
			assert.Equal(t, term.Coordinates{X: 2, Y: 4}, pos)
			assert.Equal(t, tc.expected, drawHandler(h, 20, 5))
		})
	}
}

func TestHandleCommandError(t *testing.T) {
	th := &testHandler{
		handleFn: func(
			_ context.Context, _ Command, _ ProgressWriter,
		) (
			iterator.Iterator[component.Responsive],
			error,
		) {
			return nil, errors.New("boom")
		},
	}
	h := New(th, nopSchedule, term.NopInterrupter(), WithPrompt("$ "))
	h.Resize(40, 10)

	sendKeys(t, h, "bad")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})

	// The error is dispatched async via
	// ScheduleNextTick. We verify no panic occurs.
}

// TestCompletionErrorSurfacesInline verifies that a tab-completion
// failure is rendered as an inline error line in the output band,
// rather than being silently discarded by the word completer.
func TestCompletionErrorSurfacesInline(t *testing.T) {
	th := &testHandler{
		completeFn: func(
			_ context.Context, _ string, _ []string,
		) (iterator.Iterator[string], error) {
			return nil, errors.New("list packages: forbidden")
		},
	}
	h := New(th, nopSchedule, term.NopInterrupter(), WithPrompt("$ "))
	h.Resize(40, 10)

	sendKeys(t, h, "pkg<space>")
	h.Handle(term.Event{Type: term.EventKey, Key: term.KeyTab})

	got := drawHandler(h, 40, 10)
	assert.Contains(t, got, "list packages: forbidden",
		"completion error must surface inline; got:\n%s", got)
}

func TestProgressBarRendersWhileCommandRunning(t *testing.T) {
	scheduler := &queuedScheduler{}
	cmds := newBlockingCommands()
	h := newBlockingREPL(t, cmds, scheduler)
	h.Resize(30, 5)

	sendKeys(t, h, "sync<enter>")
	cmds.waitStarted(1)
	scheduler.Flush()

	got := drawHandler(h, 30, 5)
	expected := "                              \n" +
		"                              \n" +
		"$ sync                        \n" +
		"sync ╢░_____╟ 25% 25/100 files\n" +
		"$ ▐                           "
	assert.Equal(t, expected, got)
}

func TestProgressBarClearsAfterCommandCompletes(t *testing.T) {
	scheduler := &queuedScheduler{}
	cmds := newBlockingCommands()
	h := newBlockingREPL(t, cmds, scheduler)
	h.Resize(30, 5)

	sendKeys(t, h, "sync<enter>")
	cmds.waitStarted(1)
	scheduler.Flush()
	r := h.runs[0]
	cmds.finish("sync")
	h.Wait()
	scheduler.Flush()
	h.updateProgress(r, 50, 100, "files")

	got := drawHandler(h, 30, 5)
	expected := "                              \n" +
		"                              \n" +
		"$ sync                        \n" +
		"sync done                     \n" +
		"$ ▐                           "
	assert.Equal(t, expected, got)
}

func TestRunRowsPerCommand(t *testing.T) {
	scheduler := &queuedScheduler{}
	cmds := newBlockingCommands()
	h := newBlockingREPL(t, cmds, scheduler)
	h.Resize(30, 5)

	sendKeys(t, h, "sync<enter>b<enter>")
	cmds.waitStarted(2)
	scheduler.Flush()

	got := drawHandler(h, 30, 5)
	expected := "$ sync                        \n" +
		"$ b                           \n" +
		"sync ╢░_____╟ 25% 25/100 files\n" +
		"b                           * \n" +
		"$ ▐                           "
	assert.Equal(t, expected, got)

	cmds.finish("sync")
	settle(t, scheduler, func() bool { return len(h.runs) == 1 })

	got = drawHandler(h, 30, 5)
	expected = "$ sync                        \n" +
		"$ b                           \n" +
		"sync done                     \n" +
		"b                           * \n" +
		"$ ▐                           "
	assert.Equal(t, expected, got)
}

func TestRunRow(t *testing.T) {
	cases := []struct {
		name     string
		title    string
		progress bool
		width    int
		expected string
	}{
		{
			name:     "spinner",
			title:    "index",
			width:    12,
			expected: "index     * ",
		},
		{
			name:     "spinner truncates title",
			title:    "search reindex",
			width:    12,
			expected: "search r… * ",
		},
		{
			name:     "bar",
			title:    "sync",
			progress: true,
			width:    30,
			expected: "sync ╢░_____╟ 25% 25/100 files",
		},
		{
			name:     "bar truncates title to half the width",
			title:    "search reindex defs",
			progress: true,
			width:    20,
			expected: "search re… ╢░______╟",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &run{
				title:   tc.title,
				spinner: component.NewAnimation(term.NopInterrupter(), []string{"*"}, []int{0}, 10),
			}
			t.Cleanup(func() { _ = r.spinner.Close() })
			if tc.progress {
				r.bar = component.NewProgressBar(component.DefaultProgressBarCharSet(), term.Attributes{})
				r.bar.SetProgress(25, 100, "files")
			}
			row := &runRow{r: r}
			comptest.TestComponent(t, row, term.NewStringWriter(tc.width, 1), []comptest.TestCase{
				{
					Action:   func() { row.Resize(tc.width, 1) },
					Expected: tc.expected,
				},
			})
		})
	}
}

func TestRunTitleAttributes(t *testing.T) {
	cases := []struct {
		name   string
		opts   []Option
		wantFg term.Color
	}{
		{
			name:   "default gray",
			wantFg: term.ColorGray,
		},
		{
			name:   "configured",
			opts:   []Option{WithRunTitleAttributes(term.Attributes{Fg: term.ColorTeal})},
			wantFg: term.ColorTeal,
		},
		{
			name:   "zero draws default colors",
			opts:   []Option{WithRunTitleAttributes(term.Attributes{})},
			wantFg: term.ColorDefault,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmds := newBlockingCommands()
			opts := append([]Option{
				WithPrompt("$ "),
				WithRunningAnimationFrames([]string{"*"}, []int{0}),
			}, tc.opts...)
			h := New(cmds, (&queuedScheduler{}).ScheduleNextTick, term.NopInterrupter(), opts...)
			t.Cleanup(func() { _ = h.Close() })
			h.Resize(10, 3)

			sendKeys(t, h, "ab<enter>")
			cmds.waitStarted(1)

			w := term.NewStringWriter(10, 3)
			h.Draw(w)
			row := w.Cells()[10 : 2*10]
			assert.Equal(t, 'a', row[0].Ch)
			assert.Equal(t, tc.wantFg, row[0].Fg)
			assert.Equal(t, tc.wantFg, row[1].Fg)
			assert.Equal(t, '*', row[8].Ch)
			assert.Equal(t, term.ColorDefault, row[8].Fg)
		})
	}
}

func TestDispatchLeavesInputAndHistory(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		wantAttr term.Attributes
		wantOut  string
	}{
		{
			name:     "success",
			wantAttr: defaultSuccessAttr(),
		},
		{
			name:     "error",
			err:      errors.New("boom"),
			wantAttr: defaultErrorAttr(),
			wantOut:  "boom",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scheduler := &queuedScheduler{}
			ran := make(chan Command, 1)
			th := &testHandler{
				handleFn: func(
					_ context.Context, cmd Command, _ ProgressWriter,
				) (iterator.Iterator[component.Responsive], error) {
					ran <- cmd
					if tc.err != nil {
						return nil, tc.err
					}
					return iterator.Empty[component.Responsive](), nil
				},
			}
			h := New(th, scheduler.ScheduleNextTick, term.NopInterrupter(), WithPrompt("$ "))
			t.Cleanup(func() { _ = h.Close() })
			h.Resize(40, 10)

			sendKeys(t, h, "typed")
			h.Dispatch("cmd arg")
			assert.Equal(t, Command{Name: "cmd", Args: []string{"arg"}}, <-ran)
			h.Wait()
			scheduler.Flush()

			assert.Equal(t, "typed", h.Text())
			assert.Empty(t, h.history)
			assert.Empty(t, h.runs)
			assert.Equal(t, tc.wantAttr, h.lastPromptLine.promptAttr)
			got := drawHandler(h, 40, 10)
			assert.Contains(t, got, "$ cmd arg")
			assert.Contains(t, got, "$ typed")
			assert.Contains(t, got, tc.wantOut)
		})
	}
}

func TestClearKeepsRunningCommands(t *testing.T) {
	scheduler := &queuedScheduler{}
	cmds := newBlockingCommands()
	h := newBlockingREPL(t, cmds, scheduler)
	h.Resize(20, 4)

	cmds.finish("old")
	submit(t, h, scheduler, "old")
	sendKeys(t, h, "a<enter>")
	cmds.waitStarted(2)

	h.Clear()
	got := drawHandler(h, 20, 4)
	expected := "                    \n" +
		"                    \n" +
		"a                 * \n" +
		"$ ▐                 "
	assert.Equal(t, expected, got)
	assert.NoError(t, cmds.ctx("a").Err())

	cmds.finish("a")
	settle(t, scheduler, func() bool { return len(h.runs) == 0 })
	got = drawHandler(h, 20, 4)
	expected = "                    \n" +
		"                    \n" +
		"a done              \n" +
		"$ ▐                 "
	assert.Equal(t, expected, got)
}

func TestCancelAll(t *testing.T) {
	scheduler := &queuedScheduler{}
	cmds := newBlockingCommands()
	h := newBlockingREPL(t, cmds, scheduler)
	h.Resize(40, 10)

	sendKeys(t, h, "a<enter>b<enter>")
	cmds.waitStarted(2)

	h.CancelAll()
	settle(t, scheduler, func() bool { return len(h.runs) == 0 })
	assert.ErrorIs(t, cmds.ctx("a").Err(), context.Canceled)
	assert.ErrorIs(t, cmds.ctx("b").Err(), context.Canceled)

	sendKeys(t, h, "c<enter>")
	cmds.waitStarted(1)
	assert.NoError(t, cmds.ctx("c").Err())
}

func TestMultipleCommands(t *testing.T) {
	h := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
	)
	h.Resize(40, 10)

	sendKeys(t, h, "cmd1")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})
	sendKeys(t, h, "cmd2")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})

	got := drawHandler(h, 40, 10)
	assert.Contains(t, got, "$ cmd1")
	assert.Contains(t, got, "$ cmd2")
}

func TestOutputAccumulates(t *testing.T) {
	h := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
	)
	h.Resize(40, 10)

	// Submit cmd1
	sendKeys(t, h, "cmd1")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})
	got := drawHandler(h, 40, 10)
	assert.Contains(t, got, "$ cmd1")

	// Submit cmd2 — both commands must remain
	sendKeys(t, h, "cmd2")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})
	got = drawHandler(h, 40, 10)
	assert.Contains(t, got, "$ cmd1")
	assert.Contains(t, got, "$ cmd2")

	// Submit cmd3 — all three must remain
	sendKeys(t, h, "cmd3")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})
	got = drawHandler(h, 40, 10)
	assert.Contains(t, got, "$ cmd1")
	assert.Contains(t, got, "$ cmd2")
	assert.Contains(t, got, "$ cmd3")
}

func TestCtrlDWithTextDeletesForward(t *testing.T) {
	h := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
	)
	h.Resize(20, 5)
	sendKeys(t, h, "abc")
	sendKeys(t, h, "<home>")
	exit, _ := h.Handle(term.Event{
		Type: term.EventKey,
		Ch:   'd',
		Mod:  term.ModCtrl,
	})
	assert.False(t, exit)
	assert.Equal(t, "bc", h.rl.C.Text())
}

func TestParseCommand(t *testing.T) {
	cases := []struct {
		name string
		text string
		want Command
	}{
		{
			name: "empty",
			text: "",
			want: Command{},
		},
		{
			name: "name only",
			text: "help",
			want: Command{
				Name: "help",
				Args: []string{},
			},
		},
		{
			name: "name and args",
			text: "echo hello world",
			want: Command{
				Name: "echo",
				Args: []string{
					"hello", "world",
				},
			},
		},
		{
			name: "extra whitespace",
			text: "  cmd   arg1   arg2  ",
			want: Command{
				Name: "cmd",
				Args: []string{"arg1", "arg2"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseCommand(tc.text)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFuncCommandHandlerComplete(t *testing.T) {
	called := false
	ch := FuncCommandHandler(
		func(
			_ context.Context, _ Command, _ ProgressWriter,
		) (
			iterator.Iterator[component.Responsive],
			error,
		) {
			called = true
			return iterator.FromSlice[component.Responsive](nil), nil
		},
		func(_ context.Context, _ string, _ []string) (iterator.Iterator[string], error) {
			return iterator.FromSlice([]string{"a", "b"}), nil
		},
	)

	_, err := ch.HandleCommand(context.Background(), Command{Name: "test"}, NopProgressWriter())
	assert.NoError(t, err)
	assert.True(t, called)

	iter, err := ch.Complete(context.Background(), "t", nil)
	assert.NoError(t, err)
	defer func() { _ = iter.Close() }()
	got, err := iterator.ToSlice(context.Background(), iter)
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, got)
}

func TestNopCommandCompleterComplete(t *testing.T) {
	ch := NopCommandCompleter(
		func(
			_ context.Context, _ Command, _ ProgressWriter,
		) (
			iterator.Iterator[component.Responsive],
			error,
		) {
			return iterator.FromSlice[component.Responsive](nil), nil
		},
	)
	iter, err := ch.Complete(context.Background(), "", nil)
	assert.NoError(t, err)
	defer func() { _ = iter.Close() }()
	got, err := iterator.ToSlice(context.Background(), iter)
	assert.NoError(t, err)
	assert.Empty(t, got)
}

func TestResultEOF(t *testing.T) {
	h := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
	)
	h.Resize(20, 5)
	exit, handled := h.Handle(term.Event{
		Type: term.EventKey,
		Ch:   'd',
		Mod:  term.ModCtrl,
	})
	assert.True(t, exit)
	assert.True(t, handled)

	_, err := h.rl.C.Result()
	assert.Equal(t, io.EOF, err)
}

func TestLayoutShortContent(t *testing.T) {
	scheduler := &queuedScheduler{}
	h := New(
		&testHandler{},
		scheduler.ScheduleNextTick, term.NopInterrupter(),
		WithPrompt("$ "),
	)
	h.Resize(20, 5)

	// Submit one command — one output line sits just
	// above the prompt, not at the top.
	submit(t, h, scheduler, "cmd1")

	got := drawHandler(h, 20, 5)
	expected := "                    \n" +
		"                    \n" +
		"                    \n" +
		"$ cmd1              \n" +
		"$ ▐                 "
	assert.Equal(t, expected, got)
}

func TestLayoutOverflowContent(t *testing.T) {
	scheduler := &queuedScheduler{}
	h := New(
		&testHandler{},
		scheduler.ScheduleNextTick, term.NopInterrupter(),
		WithPrompt("$ "),
	)
	// Small viewport: 3 rows total (2 output + 1
	// prompt).
	h.Resize(20, 3)

	// Submit 4 commands to overflow the output area.
	for _, cmd := range []string{"a", "b", "c", "d"} {
		submit(t, h, scheduler, cmd)
	}

	got := drawHandler(h, 20, 3)
	// The two most recent echoed commands fill the
	// output area from the top, scroll keeping the
	// latest visible.
	expected := "$ c                 \n" +
		"$ d                 \n" +
		"$ ▐                 "
	assert.Equal(t, expected, got)
}

func TestMouseWheelScrollsOutputPane(t *testing.T) {
	scheduler := &queuedScheduler{}
	h := New(
		&testHandler{},
		scheduler.ScheduleNextTick, term.NopInterrupter(),
		WithPrompt("$ "),
	)
	h.Resize(20, 3)

	for _, cmd := range []string{"a", "b", "c", "d"} {
		submit(t, h, scheduler, cmd)
	}

	exit, handled := wheelUpAt(h, 0, 1)
	assert.False(t, exit)
	assert.True(t, handled)

	got := drawHandler(h, 20, 3)
	expected := "$ b                 \n" +
		"$ c                 \n" +
		"$ ▐                 "
	assert.Equal(t, expected, got)
}

func TestMouseWheelInInputRegionDoesNotScrollOutput(t *testing.T) {
	scheduler := &queuedScheduler{}
	h := New(
		&testHandler{},
		scheduler.ScheduleNextTick, term.NopInterrupter(),
		WithPrompt("$ "),
	)
	h.Resize(20, 5)

	for _, cmd := range []string{"a", "b", "c", "d", "e"} {
		submit(t, h, scheduler, cmd)
	}
	sendKeys(t, h, "abcdefghijklmnopqrst")

	before := drawHandler(h, 20, 5)
	exit, handled := wheelUpAt(h, 0, 3)
	assert.False(t, exit)
	assert.False(t, handled)
	assert.Equal(t, before, drawHandler(h, 20, 5))
	assert.Contains(t, before, "$ b")
	assert.Contains(t, before, "$ c")
	assert.Contains(t, before, "$ d")
}

func TestMouseWheelAtPaneBoundaryTargetsCorrectPane(t *testing.T) {
	newHandler := func() *Handler {
		scheduler := &queuedScheduler{}
		h := New(
			&testHandler{},
			scheduler.ScheduleNextTick, term.NopInterrupter(),
			WithPrompt("$ "),
		)
		h.Resize(20, 5)
		for _, cmd := range []string{"a", "b", "c", "d", "e"} {
			submit(t, h, scheduler, cmd)
		}
		sendKeys(t, h, "abcdefghijklmnopqrst")
		return h
	}

	t.Run("last output row scrolls output", func(t *testing.T) {
		h := newHandler()
		exit, handled := wheelUpAt(h, 0, 2)
		assert.False(t, exit)
		assert.True(t, handled)

		got := drawHandler(h, 20, 5)
		assert.Contains(t, got, "$ a")
		assert.Contains(t, got, "$ b")
		assert.Contains(t, got, "$ c")
		assert.NotContains(t, got, "$ d")
	})

	t.Run("first input row does not scroll output", func(t *testing.T) {
		h := newHandler()
		before := drawHandler(h, 20, 5)
		exit, handled := wheelUpAt(h, 0, 3)
		assert.False(t, exit)
		assert.False(t, handled)
		assert.Equal(t, before, drawHandler(h, 20, 5))
	})
}

func TestMouseWheelScrollsOutputViaHandleSequence(t *testing.T) {
	scheduler := &queuedScheduler{}
	h := New(
		&testHandler{},
		scheduler.ScheduleNextTick, term.NopInterrupter(),
		WithPrompt("$ "),
	)
	h.Resize(20, 3)
	for _, cmd := range []string{"a", "b", "c", "d"} {
		submit(t, h, scheduler, cmd)
	}
	cases := []handlertest.SequenceTestCase{
		{
			InputSequence: "",
			Expected: "$ c                 \n" +
				"$ d                 \n" +
				"$ ▐                 ",
		},
		{
			InputSequence: "<mouse-wheel-up>",
			Expected: "$ b                 \n" +
				"$ c                 \n" +
				"$ ▐                 ",
		},
		{
			InputSequence: "<mouse-wheel-up>",
			Expected: "$ a                 \n" +
				"$ b                 \n" +
				"$ ▐                 ",
		},
		{
			InputSequence: "<mouse-wheel-up>",
			Expected: "$ a                 \n" +
				"$ b                 \n" +
				"$ ▐                 ",
		},
		{
			InputSequence: "<mouse-wheel-down>",
			Expected: "$ b                 \n" +
				"$ c                 \n" +
				"$ ▐                 ",
		},
		{
			InputSequence: "<mouse-wheel-down><mouse-wheel-down>",
			Expected: "$ c                 \n" +
				"$ d                 \n" +
				"$ ▐                 ",
		},
	}
	handlertest.RunHandlerSequence(t, h, 20, 3, cases)
}

func TestNewOutputSnapsToBottomAfterManualScroll(t *testing.T) {
	scheduler := &queuedScheduler{}
	items := make(chan component.Responsive, 2)
	var closeItems sync.Once
	stream := newChannelIterator(items)
	th := &testHandler{
		handleFn: func(
			_ context.Context, cmd Command, _ ProgressWriter,
		) (
			iterator.Iterator[component.Responsive],
			error,
		) {
			if cmd.Name == "stream" {
				return stream, nil
			}
			return iterator.Empty[component.Responsive](), nil
		},
	}
	h := New(
		th, scheduler.ScheduleNextTick, term.NopInterrupter(),
		WithPrompt("$ "),
		WithRunningAnimationFrames([]string{}, []int{}),
	)
	t.Cleanup(func() {
		closeItems.Do(func() { close(items) })
		_ = h.Close()
	})
	h.Resize(20, 4)

	submit(t, h, scheduler, "prev")

	items <- component.NewResponsiveString(
		"first",
		component.StringResponsiveConfig{},
	)
	sendKeys(t, h, "stream")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})

	<-stream.started
	settle(t, scheduler, func() bool {
		return strings.Contains(drawHandler(h, 20, 4), "first")
	})

	got := drawHandler(h, 20, 4)
	assert.Contains(t, got, "$ stream")
	assert.Contains(t, got, "first")

	_, handled := wheelUpAt(h, 0, 1)
	assert.True(t, handled)
	got = drawHandler(h, 20, 4)
	assert.Contains(t, got, "$ prev")
	assert.Contains(t, got, "$ stream")
	assert.NotContains(t, got, "second")

	items <- component.NewResponsiveString(
		"second",
		component.StringResponsiveConfig{},
	)
	closeItems.Do(func() { close(items) })
	h.Wait()
	scheduler.Flush()

	got = drawHandler(h, 20, 4)
	expected := "$ stream            \n" +
		"first               \n" +
		"second              \n" +
		"$ ▐                 "
	assert.Equal(t, expected, got)
}

func TestHistoryPersistence(t *testing.T) {
	svc := storagestub.NewInMemoryService()
	h := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
		WithStorage("repl-history", svc),
	)
	h.Resize(30, 5)

	sendKeys(t, h, "cmd1")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})

	sendKeys(t, h, "cmd2")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})

	var doc historyDoc
	err := svc.Get(
		context.Background(), "repl-history", &doc,
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"cmd1", "cmd2"}, doc.Items)
}

func TestExitError(t *testing.T) {
	exitErr := errors.New("exit")
	th := &testHandler{
		handleFn: func(
			_ context.Context, cmd Command, _ ProgressWriter,
		) (
			iterator.Iterator[component.Responsive],
			error,
		) {
			if cmd.Name == "quit" {
				return nil, exitErr
			}
			return iterator.FromSlice[component.Responsive](nil), nil
		},
	}
	h := New(
		th, nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
		WithExitError(exitErr),
	)
	h.Resize(40, 10)

	sendKeys(t, h, "quit")
	exit, handled := h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})
	// The command dispatches asynchronously; exit is
	// not immediate on the same Handle call.
	assert.False(t, exit)
	assert.True(t, handled)

	// Wait for the dispatch goroutine to set
	// exitPending.
	h.Wait()

	// The next Handle call should return exit.
	exit, handled = h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})
	assert.True(t, exit)
	assert.True(t, handled)
}

func TestExitErrorWrapped(t *testing.T) {
	exitErr := errors.New("exit")
	th := &testHandler{
		handleFn: func(
			_ context.Context, _ Command, _ ProgressWriter,
		) (
			iterator.Iterator[component.Responsive],
			error,
		) {
			return nil, fmt.Errorf("quitting: %w", exitErr)
		},
	}
	h := New(
		th, nopSchedule, term.NopInterrupter(),
		WithExitError(exitErr),
	)
	h.Resize(40, 10)

	sendKeys(t, h, "quit")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})
	h.Wait()

	exit, _ := h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})
	assert.True(t, exit)
}

func TestExitErrorNotConfigured(t *testing.T) {
	th := &testHandler{
		handleFn: func(
			_ context.Context, _ Command, _ ProgressWriter,
		) (
			iterator.Iterator[component.Responsive],
			error,
		) {
			return nil, errors.New("exit")
		},
	}
	h := New(
		th, nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
	)
	h.Resize(40, 10)

	sendKeys(t, h, "quit")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})
	h.Wait()

	// Without WithExitError, the error is just
	// displayed; Handle does not return exit.
	exit, _ := h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})
	assert.False(t, exit)
}

func TestClose(t *testing.T) {
	t.Run("cancels every running command", func(t *testing.T) {
		cmds := newBlockingCommands()
		h := newBlockingREPL(t, cmds, &queuedScheduler{})
		h.Resize(20, 5)

		sendKeys(t, h, "a<enter>b<enter>")
		cmds.waitStarted(2)

		assert.NoError(t, h.Close())
		assert.ErrorIs(t, cmds.ctx("a").Err(), context.Canceled)
		assert.ErrorIs(t, cmds.ctx("b").Err(), context.Canceled)
	})

	t.Run("stops blocking command", func(t *testing.T) {
		started := make(chan struct{})
		th := &testHandler{
			handleFn: func(
				ctx context.Context, _ Command, _ ProgressWriter,
			) (
				iterator.Iterator[component.Responsive],
				error,
			) {
				close(started)
				<-ctx.Done()
				return nil, ctx.Err()
			},
		}
		h := New(
			th, nopSchedule, term.NopInterrupter(),
			WithPrompt("$ "),
		)
		h.Resize(40, 10)

		sendKeys(t, h, "slow")
		h.Handle(term.Event{
			Type: term.EventKey,
			Key:  term.KeyEnter,
		})

		// Wait for the command goroutine to start.
		<-started

		// Close unblocks the command and waits
		// for it to finish.
		err := h.Close()
		assert.NoError(t, err)
	})

	t.Run("idempotent", func(t *testing.T) {
		h := New(
			&testHandler{},
			nopSchedule, term.NopInterrupter(),
		)
		h.Resize(20, 5)

		assert.NoError(t, h.Close())
		assert.NoError(t, h.Close())
	})
}

func TestHistoryPreloaded(t *testing.T) {
	svc := storagestub.NewInMemoryService()
	doc := historyDoc{Items: []string{"old1", "old2"}}
	err := svc.Set(
		context.Background(), "repl-history", &doc,
	)
	require.NoError(t, err)

	h := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
		WithStorage("repl-history", svc),
	)
	h.Resize(30, 5)

	// Up arrow should recall "old2"
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyArrowUp,
	})
	assert.Equal(t, "old2", h.rl.C.Text())

	// Another up should recall "old1"
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyArrowUp,
	})
	assert.Equal(t, "old1", h.rl.C.Text())
}
