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
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/storagestub"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/handler/handlertest"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

var _ tui.Handler = (*Handler)(nil)

type testHandler struct {
	handleFn func(
		context.Context, Command,
	) (iterator.Iterator[component.Responsive], error)
	completeFn func(
		context.Context, string, []string,
	) (iterator.Iterator[string], error)
}

func (t *testHandler) HandleCommand(
	ctx context.Context, cmd Command,
) (iterator.Iterator[component.Responsive], error) {
	if t.handleFn != nil {
		return t.handleFn(ctx, cmd)
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

func TestEmptyInput(t *testing.T) {
	called := false
	th := &testHandler{
		handleFn: func(
			_ context.Context, _ Command,
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
			_ context.Context, cmd Command,
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
	var mu sync.Mutex
	var capturedCtx context.Context
	th := &testHandler{
		handleFn: func(
			ctx context.Context, _ Command,
		) (
			iterator.Iterator[component.Responsive],
			error,
		) {
			mu.Lock()
			capturedCtx = ctx
			mu.Unlock()
			return iterator.FromSlice[component.Responsive](nil), nil
		},
	}
	h := New(th, nopSchedule, term.NopInterrupter(), WithPrompt("$ "))
	h.Resize(40, 10)

	// Capture the context before dispatch so
	// we can check it was cancelled.
	preCtx := h.cmdCtx

	sendKeys(t, h, "cmd")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})

	// Ctrl+C should cancel the context
	h.Handle(term.Event{
		Type: term.EventKey,
		Ch:   'c',
		Mod:  term.ModCtrl,
	})

	// Wait for the dispatch goroutine to complete
	// before inspecting capturedCtx.
	h.Wait()

	assert.Error(t, preCtx.Err())

	// New context should be alive
	assert.NoError(t, h.cmdCtx.Err())

	// Just verify handler got called
	mu.Lock()
	assert.NotNil(t, capturedCtx)
	mu.Unlock()
}

func TestSelection(t *testing.T) {
	h := New(&testHandler{}, nopSchedule, term.NopInterrupter())
	h.Resize(20, 5)
	sel, ok := h.Selection()
	assert.False(t, ok)
	assert.Equal(t, "", sel)
}

func TestLayout(t *testing.T) {
	h := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
	)
	h.Resize(20, 5)
	pos, _, show := h.Cursor()
	assert.True(t, show)
	assert.Equal(t, 4, pos.Y)
	assert.Equal(t, 2, pos.X)
}

func TestHandleCommandError(t *testing.T) {
	th := &testHandler{
		handleFn: func(
			_ context.Context, _ Command,
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
			_ context.Context, _ Command,
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

	_, err := ch.HandleCommand(context.Background(), Command{Name: "test"})
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
			_ context.Context, _ Command,
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
	h := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
	)
	h.Resize(20, 5)

	// Submit one command — one output line sits just
	// above the prompt, not at the top.
	sendKeys(t, h, "cmd1")
	h.Handle(term.Event{
		Type: term.EventKey,
		Key:  term.KeyEnter,
	})

	got := drawHandler(h, 20, 5)
	expected := "                    \n" +
		"                    \n" +
		"                    \n" +
		"$ cmd1              \n" +
		"$ ▐                 "
	assert.Equal(t, expected, got)
}

func TestLayoutOverflowContent(t *testing.T) {
	h := New(
		&testHandler{},
		nopSchedule, term.NopInterrupter(),
		WithPrompt("$ "),
	)
	// Small viewport: 3 rows total (2 output + 1
	// prompt).
	h.Resize(20, 3)

	// Submit 4 commands to overflow the output area.
	for _, cmd := range []string{"a", "b", "c", "d"} {
		sendKeys(t, h, cmd)
		h.Handle(term.Event{
			Type: term.EventKey,
			Key:  term.KeyEnter,
		})
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
			_ context.Context, cmd Command,
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
			_ context.Context, _ Command,
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
			_ context.Context, _ Command,
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
	t.Run("cancels command context", func(t *testing.T) {
		h := New(
			&testHandler{},
			nopSchedule, term.NopInterrupter(),
		)
		h.Resize(20, 5)

		assert.NoError(t, h.cmdCtx.Err())
		err := h.Close()
		assert.NoError(t, err)
		assert.Error(t, h.cmdCtx.Err())
	})

	t.Run("stops blocking command", func(t *testing.T) {
		started := make(chan struct{})
		th := &testHandler{
			handleFn: func(
				ctx context.Context, _ Command,
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
