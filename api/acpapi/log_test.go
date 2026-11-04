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

package acpapi_test

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	acp "github.com/coder/acp-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/acpapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

func drain(t *testing.T, it iterator.Iterator[acpapi.Event]) []acpapi.Event {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var out []acpapi.Event
	for {
		ev, ok := it.Next(ctx)
		if !ok {
			require.NoError(t, ctx.Err(), "iterator did not end")
			return out
		}
		out = append(out, ev)
	}
}

func next(t *testing.T, it iterator.Iterator[acpapi.Event]) acpapi.Event {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ev, ok := it.Next(ctx)
	require.True(t, ok, "Next: %v", it.Err())
	return ev
}

func waitPending(t *testing.T, l *acpapi.Log, n int) {
	t.Helper()
	require.Eventually(t, func() bool { return len(l.State().Pending) == n }, 5*time.Second, time.Millisecond)
}

func TestLogLateSubscriberReplaysFromZero(t *testing.T) {
	l := acpapi.NewLog()
	l.Started(acpapi.Started{})
	l.Update(acp.UpdateAgentMessageText("hi"))

	live := l.Events()
	defer func() { _ = live.Close() }()
	assert.Equal(t, 0, next(t, live).Seq)
	assert.Equal(t, 1, next(t, live).Seq)

	got := make(chan acpapi.Event)
	go func() { got <- next(t, live) }()
	l.Update(acp.UpdateAgentMessageText("there"))
	assert.Equal(t, 2, (<-got).Seq)

	late := l.Events()
	defer func() { _ = late.Close() }()
	l.Close(nil)
	events := drain(t, late)
	require.Len(t, events, 3)
	for i, ev := range events {
		assert.Equal(t, i, ev.Seq)
	}
	assert.NoError(t, late.Err())
}

func TestLogIndependentIterators(t *testing.T) {
	l := acpapi.NewLog()
	l.Update(acp.UpdateAgentMessageText("a"))
	l.Update(acp.UpdateAgentMessageText("b"))

	a, b := l.Events(), l.Events()
	assert.Equal(t, 0, next(t, a).Seq)
	require.NoError(t, a.Close())
	require.NoError(t, a.Close())
	_, ok := a.Next(context.Background())
	assert.False(t, ok)
	assert.NoError(t, a.Err())

	assert.Equal(t, 0, next(t, b).Seq)
	assert.Equal(t, 1, next(t, b).Seq)
	l.Close(nil)
	assert.Empty(t, drain(t, b))
}

func TestLogCloseUnblocksIterator(t *testing.T) {
	l := acpapi.NewLog()
	it := l.Events()
	done := make(chan bool)
	go func() {
		_, ok := it.Next(context.Background())
		done <- ok
	}()
	require.NoError(t, it.Close())
	assert.False(t, <-done)
}

func TestLogNextContextBoundsWaitOnly(t *testing.T) {
	l := acpapi.NewLog()
	it := l.Events()
	defer func() { _ = it.Close() }()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, ok := it.Next(ctx)
	assert.False(t, ok)
	assert.ErrorIs(t, it.Err(), context.Canceled)

	l.Update(acp.UpdateAgentMessageText("a"))
	assert.Equal(t, 0, next(t, it).Seq)
	assert.NoError(t, it.Err())
}

func TestLogCloseErrorSurfacesAsErr(t *testing.T) {
	cause := errors.New("agent exited: boom")
	l := acpapi.NewLog()
	l.Update(acp.UpdateAgentMessageText("a"))
	l.Close(cause)
	l.Close(errors.New("second close"))
	l.Update(acp.UpdateAgentMessageText("ignored"))

	it := l.Events()
	assert.Len(t, drain(t, it), 1)
	assert.Equal(t, cause, it.Err())
	assert.ErrorIs(t, l.Prompt(nil), acpapi.ErrClosed)
}

func TestLogPromptBusy(t *testing.T) {
	l := acpapi.NewLog()
	require.NoError(t, l.Prompt([]acp.ContentBlock{acp.TextBlock("one")}))
	assert.True(t, l.State().Busy)
	assert.ErrorIs(t, l.Prompt([]acp.ContentBlock{acp.TextBlock("two")}), acpapi.ErrBusy)
	l.Done(acpapi.TurnEnd{StopReason: acp.StopReasonEndTurn})
	assert.False(t, l.State().Busy)
	require.NoError(t, l.Prompt([]acp.ContentBlock{acp.TextBlock("two")}))
}

func TestLogAskRespondFirstAnswerWins(t *testing.T) {
	l := acpapi.NewLog()
	type result struct {
		o   acp.RequestPermissionOutcome
		err error
	}
	res := make(chan result, 1)
	go func() {
		o, err := l.Ask(context.Background(), acp.RequestPermissionRequest{SessionId: "s"})
		res <- result{o, err}
	}()
	waitPending(t, l, 1)
	id := l.State().Pending[0].ID

	allow := acp.NewRequestPermissionOutcomeSelected("allow")
	reject := acp.NewRequestPermissionOutcomeSelected("reject")
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, o := range []acp.RequestPermissionOutcome{allow, reject} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = l.Respond(id, o)
		}()
	}
	wg.Wait()
	r := <-res
	require.NoError(t, r.err)
	if errs[0] == nil {
		assert.ErrorIs(t, errs[1], acpapi.ErrResolved)
		assert.Equal(t, allow, r.o)
	} else {
		assert.ErrorIs(t, errs[0], acpapi.ErrResolved)
		assert.Equal(t, reject, r.o)
	}
	assert.Empty(t, l.State().Pending)
	assert.ErrorIs(t, l.Respond(42, allow), acpapi.ErrNotFound)

	l.Close(nil)
	events := drain(t, l.Events())
	require.Len(t, events, 2)
	require.NotNil(t, events[0].Permission)
	require.NotNil(t, events[1].Resolved)
	assert.Equal(t, id, events[1].Resolved.ID)
	assert.Equal(t, r.o, events[1].Resolved.Outcome)
}

func TestLogDoneCancelsPending(t *testing.T) {
	l := acpapi.NewLog()
	require.NoError(t, l.Prompt(nil))
	outcomes := make(chan acp.RequestPermissionOutcome, 2)
	for range 2 {
		go func() {
			o, err := l.Ask(context.Background(), acp.RequestPermissionRequest{})
			assert.NoError(t, err)
			outcomes <- o
		}()
	}
	waitPending(t, l, 2)
	l.Done(acpapi.TurnEnd{StopReason: acp.StopReasonCancelled})
	for range 2 {
		assert.NotNil(t, (<-outcomes).Cancelled)
	}
	l.Close(nil)

	events := drain(t, l.Events())
	require.Len(t, events, 6)
	assert.NotNil(t, events[3].Resolved)
	assert.NotNil(t, events[4].Resolved)
	require.NotNil(t, events[5].Done)
	assert.Equal(t, acp.StopReasonCancelled, events[5].Done.StopReason)
	assert.Empty(t, l.State().Pending)
}

// ask asks l for permission on session in the background, returning the
// outcome once resolved.
func ask(t *testing.T, l *acpapi.Log, session acp.SessionId) <-chan acp.RequestPermissionOutcome {
	t.Helper()
	n := len(l.State().Pending)
	res := make(chan acp.RequestPermissionOutcome, 1)
	go func() {
		o, err := l.Ask(context.Background(), acp.RequestPermissionRequest{SessionId: session})
		assert.NoError(t, err)
		res <- o
	}()
	waitPending(t, l, n+1)
	return res
}

func TestLogSubagentPermissions(t *testing.T) {
	cancelled := acpapi.SubagentState{State: acpapi.SubagentIdle, StopReason: acp.StopReasonCancelled}
	finished := acpapi.SubagentState{State: acpapi.SubagentIdle, StopReason: acp.StopReasonEndTurn}
	cases := []struct {
		name string
		// end ends the work that made the permissions pending.
		end func(*acpapi.Log)
		// resolved are the sessions whose permission end resolves.
		resolved []acp.SessionId
	}{
		{
			name:     "turn end resolves only the session's own",
			end:      func(l *acpapi.Log) { l.Done(acpapi.TurnEnd{StopReason: acp.StopReasonEndTurn}) },
			resolved: []acp.SessionId{"s"},
		},
		{
			name: "cancelled subagent resolves its own",
			end: func(l *acpapi.Log) {
				l.Subagent("", acpapi.SubagentUpdate{SessionID: "a", State: acpapi.Some(cancelled)})
			},
			resolved: []acp.SessionId{"a"},
		},
		{
			name: "finished subagent keeps them",
			end: func(l *acpapi.Log) {
				l.Subagent("", acpapi.SubagentUpdate{SessionID: "a", State: acpapi.Some(finished)})
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := acpapi.NewLog()
			l.Subagent("", acpapi.SubagentUpdate{SessionID: "a"})
			l.Subagent("a", acpapi.SubagentUpdate{SessionID: "b"})
			require.NoError(t, l.Prompt(nil))
			outcomes := map[acp.SessionId]<-chan acp.RequestPermissionOutcome{}
			for _, s := range []acp.SessionId{"s", "a", "b"} {
				outcomes[s] = ask(t, l, s)
			}
			var asked []acp.SessionId
			for _, p := range l.State().Pending {
				asked = append(asked, p.Session)
			}
			assert.Equal(t, []acp.SessionId{"", "a", "b"}, asked)

			tc.end(l)
			var pending []acp.SessionId
			for _, p := range l.State().Pending {
				pending = append(pending, p.Request.SessionId)
			}
			for s, ch := range outcomes {
				if slices.Contains(tc.resolved, s) {
					assert.NotNil(t, (<-ch).Cancelled, s)
					assert.NotContains(t, pending, s)
				} else {
					assert.Contains(t, pending, s)
				}
			}
			l.Close(nil)
		})
	}
}

func TestLogAskContextWithdraws(t *testing.T) {
	l := acpapi.NewLog()
	ctx, cancel := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() {
		o, err := l.Ask(ctx, acp.RequestPermissionRequest{})
		assert.NotNil(t, o.Cancelled)
		errc <- err
	}()
	waitPending(t, l, 1)
	id := l.State().Pending[0].ID
	cancel()
	assert.ErrorIs(t, <-errc, context.Canceled)
	assert.ErrorIs(t, l.Respond(id, acp.NewRequestPermissionOutcomeSelected("x")), acpapi.ErrResolved)
}

func TestLogCloseCancelsPending(t *testing.T) {
	l := acpapi.NewLog()
	res := make(chan acp.RequestPermissionOutcome, 1)
	go func() {
		o, _ := l.Ask(context.Background(), acp.RequestPermissionRequest{})
		res <- o
	}()
	waitPending(t, l, 1)
	l.Close(nil)
	assert.NotNil(t, (<-res).Cancelled)
	_, err := l.Ask(context.Background(), acp.RequestPermissionRequest{})
	assert.ErrorIs(t, err, acpapi.ErrClosed)
}
