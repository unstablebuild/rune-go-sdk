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

package acpapi

import (
	"context"
	"maps"
	"slices"
	"sync"

	acp "github.com/coder/acp-go-sdk"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

// Log is the append-only event log Session implementations are built on. It
// is safe for concurrent use. Once closed, appends are ignored.
type Log struct {
	mu      sync.Mutex
	events  []Event
	state   SessionState
	wake    chan struct{}
	closed  bool
	err     error
	pending map[PermissionID]chan acp.RequestPermissionOutcome
}

// NewLog returns an empty, open Log.
func NewLog() *Log {
	return &Log{
		wake:    make(chan struct{}),
		pending: make(map[PermissionID]chan acp.RequestPermissionOutcome),
	}
}

// Started logs the state returned by session/new or session/load.
func (l *Log) Started(s Started) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.appendLocked(Event{Started: &s})
}

// Update logs a session/update the agent sent on the session itself.
func (l *Log) Update(u acp.SessionUpdate) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.appendLocked(Event{Update: &u})
}

// UpdateSubagent logs a session/update the agent sent on subagent id.
func (l *Log) UpdateSubagent(id acp.SessionId, u acp.SessionUpdate) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.appendLocked(Event{Session: id, Update: &u})
}

// Subagent logs a subagent_update the agent sent on parent, the session
// itself when parent is empty. A subagent reported idle because it was
// cancelled has its pending permissions resolved Cancelled, as nothing will
// use their answers.
func (l *Log) Subagent(parent acp.SessionId, u SubagentUpdate) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.appendLocked(Event{Session: parent, Subagent: &u})
	if st := u.State.Value; st != nil && st.State == SubagentIdle && st.StopReason == acp.StopReasonCancelled {
		l.cancelPendingLocked(func(p *Permission) bool { return p.Session == u.SessionID })
	}
}

// Message logs a session_message the agent sent on session id, the session
// itself when id is empty.
func (l *Log) Message(id acp.SessionId, m SessionMessage) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.appendLocked(Event{Session: id, Message: &m})
}

// MessageChunk logs a session_message_chunk the agent sent on session id,
// the session itself when id is empty.
func (l *Log) MessageChunk(id acp.SessionId, c SessionMessageChunk) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.appendLocked(Event{Session: id, MessageChunk: &c})
}

// Prompt logs the start of a turn. It returns ErrBusy while a turn runs and
// ErrClosed after Close.
func (l *Log) Prompt(content []acp.ContentBlock) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return ErrClosed
	}
	if l.state.Busy {
		return ErrBusy
	}
	l.appendLocked(Event{Prompt: &Prompt{Content: content}})
	return nil
}

// Ask logs a Permission and blocks until it is resolved by Respond, Done,
// Close or ctx. When ctx ends first, the permission resolves Cancelled and
// Ask returns ctx's error. A request whose SessionId names an announced
// subagent is that subagent's, which Done does not resolve.
func (l *Log) Ask(ctx context.Context, req acp.RequestPermissionRequest) (acp.RequestPermissionOutcome, error) {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return acp.NewRequestPermissionOutcomeCancelled(), ErrClosed
	}
	id := PermissionID(len(l.events))
	ch := make(chan acp.RequestPermissionOutcome, 1)
	l.pending[id] = ch
	p := &Permission{ID: id, Request: req}
	if _, ok := l.state.Subagents[req.SessionId]; ok {
		p.Session = req.SessionId
	}
	l.appendLocked(Event{Permission: p})
	l.mu.Unlock()

	select {
	case o := <-ch:
		return o, nil
	case <-ctx.Done():
	}
	l.mu.Lock()
	_, withdrawn := l.pending[id]
	if withdrawn {
		l.resolveLocked(id, acp.NewRequestPermissionOutcomeCancelled())
	}
	l.mu.Unlock()
	o := <-ch
	if withdrawn {
		return o, ctx.Err()
	}
	return o, nil
}

// Respond answers a pending permission. The first answer wins; later ones
// return ErrResolved. An id that never named a permission is ErrNotFound.
func (l *Log) Respond(id PermissionID, o acp.RequestPermissionOutcome) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.pending[id]; ok {
		l.resolveLocked(id, o)
		return nil
	}
	if int(id) >= 0 && int(id) < len(l.events) && l.events[id].Permission != nil {
		return ErrResolved
	}
	return ErrNotFound
}

// Done ends the running turn. The session's own pending permissions resolve
// Cancelled first; the subagents' stay pending.
func (l *Log) Done(end TurnEnd) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return
	}
	l.cancelPendingLocked(func(p *Permission) bool { return p.Session == "" })
	l.appendLocked(Event{Done: &end})
}

// Events returns a new iterator over every event from Seq 0 that then follows
// live. It ends when the log is closed and drained; Err then returns the
// error given to Close. Next's ctx bounds the wait only: after it ends, Next
// returns false with Err set to ctx's error, and a later Next resumes.
func (l *Log) Events() iterator.Iterator[Event] {
	return &logIterator{l: l, done: make(chan struct{})}
}

// State returns the state folded from every event so far.
func (l *Log) State() SessionState {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.state
}

// Close ends the log. Pending permissions resolve Cancelled and iterators end
// after draining with err. Only the first Close has effect.
func (l *Log) Close(err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return
	}
	l.cancelPendingLocked(func(*Permission) bool { return true })
	l.closed = true
	l.err = err
	close(l.wake)
}

func (l *Log) appendLocked(ev Event) {
	if l.closed {
		return
	}
	ev.Seq = len(l.events)
	l.events = append(l.events, ev)
	Apply(&l.state, ev)
	close(l.wake)
	l.wake = make(chan struct{})
}

func (l *Log) resolveLocked(id PermissionID, o acp.RequestPermissionOutcome) {
	ch := l.pending[id]
	delete(l.pending, id)
	l.appendLocked(Event{Resolved: &Resolution{ID: id, Outcome: o}})
	ch <- o
}

// cancelPendingLocked resolves the pending permissions that match Cancelled,
// oldest first.
func (l *Log) cancelPendingLocked(match func(*Permission) bool) {
	for _, id := range slices.Sorted(maps.Keys(l.pending)) {
		if match(l.events[id].Permission) {
			l.resolveLocked(id, acp.NewRequestPermissionOutcomeCancelled())
		}
	}
}

type logIterator struct {
	l    *Log
	next int
	err  error
	done chan struct{}
	once sync.Once
}

func (it *logIterator) Next(ctx context.Context) (Event, bool) {
	for {
		select {
		case <-it.done:
			it.err = nil
			return Event{}, false
		default:
		}
		it.l.mu.Lock()
		if it.next < len(it.l.events) {
			ev := it.l.events[it.next]
			it.next++
			it.l.mu.Unlock()
			it.err = nil
			return ev, true
		}
		if it.l.closed {
			it.err = it.l.err
			it.l.mu.Unlock()
			return Event{}, false
		}
		wake := it.l.wake
		it.l.mu.Unlock()
		select {
		case <-wake:
		case <-it.done:
		case <-ctx.Done():
			it.err = ctx.Err()
			return Event{}, false
		}
	}
}

func (it *logIterator) Err() error { return it.err }

func (it *logIterator) Close() error {
	it.once.Do(func() { close(it.done) })
	return nil
}
