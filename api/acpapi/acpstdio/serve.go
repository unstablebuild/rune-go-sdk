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

package acpstdio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"sync"

	acp "github.com/coder/acp-go-sdk"
	"github.com/unstablebuild/rune-go-sdk/api/acpapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

// Serve serves a as an ACP v1 agent to the client reading w and writing r,
// until r ends or ctx is done. It returns nil when the client closed the
// connection. Sessions opened through the connection are closed when Serve
// returns; a is not.
//
// The events a session has logged when a's NewSession or LoadSession returns
// are sent before the response; for session/load they are the history, with
// Prompt events as user_message_chunk updates and permissions and turn ends
// left out. Serve reads them with an expired context, so Events must return
// already-logged events from Next regardless of its context, as Log does.
// Later events follow live: updates as session/update, permissions as
// session/request_permission and turn ends as the session/prompt response.
// Prompt events are not sent live. Subagents are not exposed: their events
// are left out, and their permissions are asked on the session.
func Serve(ctx context.Context, a acpapi.Agent, r io.Reader, w io.Writer) error {
	parent := ctx
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sv := &server{
		a:        a,
		ctx:      ctx,
		sessions: make(map[acp.SessionId]*served),
	}
	ready := make(chan struct{})
	er := &errReader{r: r, ready: ready}
	sv.conn = acp.NewAgentSideConnection((*handler)(sv), w, er)
	close(ready)
	select {
	case <-sv.conn.Done():
	case <-ctx.Done():
	}
	cancel()

	sv.mu.Lock()
	sessions := slices.Collect(maps.Values(sv.sessions))
	clear(sv.sessions)
	sv.mu.Unlock()
	for _, s := range sessions {
		_ = s.s.Close()
	}
	if err := parent.Err(); err != nil {
		return err
	}
	if err := er.Err(); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

type server struct {
	a    acpapi.Agent
	conn *acp.AgentSideConnection
	ctx  context.Context

	mu       sync.Mutex
	sessions map[acp.SessionId]*served
}

// served is a session opened through the connection.
type served struct {
	sv *server
	s  acpapi.Session
	id acp.SessionId

	mu    sync.Mutex
	snap  iterator.Iterator[acpapi.Event]
	state acpapi.SessionState
	turn  *turnWaiter
	asks  map[acpapi.PermissionID]context.CancelFunc
	done  chan struct{}
	err   error
}

// turnWaiter is a session/prompt waiting for its turn to end. It is armed by
// the next Prompt event so a lagging forwarder cannot hand it a previous
// turn's end.
type turnWaiter struct {
	armed bool
	ch    chan acpapi.TurnEnd
}

func (sv *server) open(ctx context.Context, s acpapi.Session, history bool) (*served, error) {
	ss := &served{
		sv:   sv,
		s:    s,
		id:   s.Info().ID,
		snap: s.Events(),
		asks: make(map[acpapi.PermissionID]context.CancelFunc),
		done: make(chan struct{}),
	}
	it := s.Events()
	for _, ev := range logged(it) {
		var updates []acp.SessionUpdate
		switch {
		case ev.Update != nil && ev.Session == "":
			updates = []acp.SessionUpdate{*ev.Update}
		case ev.Prompt != nil && history:
			for _, b := range ev.Prompt.Content {
				updates = append(updates, acp.UpdateUserMessage(b))
			}
		}
		for _, u := range updates {
			if err := sv.conn.SessionUpdate(ctx, acp.SessionNotification{SessionId: ss.id, Update: u}); err != nil {
				_ = it.Close()
				_ = ss.snap.Close()
				_ = s.Close()
				return nil, err
			}
		}
	}
	ss.refresh()

	sv.mu.Lock()
	if prev := sv.sessions[ss.id]; prev != nil {
		sv.mu.Unlock()
		_ = it.Close()
		_ = ss.snap.Close()
		return nil, fmt.Errorf("acp: session %s: %w", ss.id, acpapi.ErrExists)
	}
	sv.sessions[ss.id] = ss
	sv.mu.Unlock()
	go ss.forward(it)
	return ss, nil
}

// logged returns the events already in it without waiting for more.
func logged(it iterator.Iterator[acpapi.Event]) []acpapi.Event {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out []acpapi.Event
	for {
		ev, ok := it.Next(ctx)
		if !ok {
			return out
		}
		out = append(out, ev)
	}
}

// refresh folds the events logged since the last refresh into state.
func (ss *served) refresh() acpapi.SessionState {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	for _, ev := range logged(ss.snap) {
		acpapi.Apply(&ss.state, ev)
	}
	return ss.state
}

func (ss *served) forward(it iterator.Iterator[acpapi.Event]) {
	defer func() { _ = it.Close() }()
	for {
		ev, ok := it.Next(ss.sv.ctx)
		if !ok {
			break
		}
		switch {
		case ev.Update != nil && ev.Session == "":
			_ = ss.sv.conn.SessionUpdate(ss.sv.ctx, acp.SessionNotification{SessionId: ss.id, Update: *ev.Update})
		case ev.Prompt != nil:
			ss.mu.Lock()
			if ss.turn != nil {
				ss.turn.armed = true
			}
			ss.mu.Unlock()
		case ev.Permission != nil:
			ss.ask(*ev.Permission)
		case ev.Resolved != nil:
			ss.mu.Lock()
			if cancel := ss.asks[ev.Resolved.ID]; cancel != nil {
				cancel()
				delete(ss.asks, ev.Resolved.ID)
			}
			ss.mu.Unlock()
		case ev.Done != nil:
			ss.mu.Lock()
			if ss.turn != nil && ss.turn.armed {
				ss.turn.ch <- *ev.Done
				ss.turn = nil
			}
			ss.mu.Unlock()
		}
	}

	ss.mu.Lock()
	ss.err = it.Err()
	for _, cancel := range ss.asks {
		cancel()
	}
	clear(ss.asks)
	close(ss.done)
	ss.mu.Unlock()
	_ = ss.snap.Close()
	ss.sv.mu.Lock()
	if ss.sv.sessions[ss.id] == ss {
		delete(ss.sv.sessions, ss.id)
	}
	ss.sv.mu.Unlock()
}

// ask relays a permission to the client until it answers or the permission
// is resolved elsewhere.
func (ss *served) ask(p acpapi.Permission) {
	ctx, cancel := context.WithCancel(ss.sv.ctx)
	ss.mu.Lock()
	ss.asks[p.ID] = cancel
	ss.mu.Unlock()
	req := p.Request
	req.SessionId = ss.id
	go func() {
		defer cancel()
		o := acp.NewRequestPermissionOutcomeCancelled()
		if resp, err := ss.sv.conn.RequestPermission(ctx, req); err == nil {
			o = resp.Outcome
		}
		_ = ss.s.Respond(ss.sv.ctx, p.ID, o)
	}()
}

// endErr is why the session's events ended.
func (ss *served) endErr() error {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if ss.err != nil {
		return ss.err
	}
	return acpapi.ErrClosed
}

// handler is the acp.Agent the client calls into.
type handler server

func (h *handler) sv() *server { return (*server)(h) }

func (h *handler) lookup(id acp.SessionId) (*served, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if ss := h.sessions[id]; ss != nil {
		return ss, nil
	}
	return nil, acp.NewInvalidParams(map[string]any{"error": fmt.Sprintf("unknown session %s", id)})
}

func (h *handler) Initialize(context.Context, acp.InitializeRequest) (acp.InitializeResponse, error) {
	info := h.a.Info()
	info.ProtocolVersion = acp.ProtocolVersionNumber
	if info.AuthMethods == nil {
		info.AuthMethods = []acp.AuthMethod{}
	}
	if info.AgentCapabilities.SessionCapabilities.Close == nil {
		info.AgentCapabilities.SessionCapabilities.Close = &acp.SessionCloseCapabilities{}
	}
	return info, nil
}

func (h *handler) Authenticate(context.Context, acp.AuthenticateRequest) (acp.AuthenticateResponse, error) {
	return acp.AuthenticateResponse{}, acp.NewMethodNotFound(acp.AgentMethodAuthenticate)
}

func (h *handler) Logout(context.Context, acp.LogoutRequest) (acp.LogoutResponse, error) {
	return acp.LogoutResponse{}, acp.NewMethodNotFound(acp.AgentMethodLogout)
}

func (h *handler) NewSession(ctx context.Context, req acp.NewSessionRequest) (acp.NewSessionResponse, error) {
	s, err := h.a.NewSession(ctx, req)
	if err != nil {
		return acp.NewSessionResponse{}, err
	}
	ss, err := h.sv().open(ctx, s, false)
	if err != nil {
		return acp.NewSessionResponse{}, err
	}
	st := ss.refresh()
	return acp.NewSessionResponse{SessionId: ss.id, ConfigOptions: st.ConfigOptions, Modes: st.Modes}, nil
}

func (h *handler) LoadSession(ctx context.Context, req acp.LoadSessionRequest) (acp.LoadSessionResponse, error) {
	if !h.a.Info().AgentCapabilities.LoadSession {
		return acp.LoadSessionResponse{}, acp.NewMethodNotFound(acp.AgentMethodSessionLoad)
	}
	s, err := h.a.LoadSession(ctx, req)
	if err != nil {
		return acp.LoadSessionResponse{}, err
	}
	ss, err := h.sv().open(ctx, s, true)
	if err != nil {
		return acp.LoadSessionResponse{}, err
	}
	st := ss.refresh()
	return acp.LoadSessionResponse{ConfigOptions: st.ConfigOptions, Modes: st.Modes}, nil
}

func (h *handler) ListSessions(ctx context.Context, req acp.ListSessionsRequest) (acp.ListSessionsResponse, error) {
	sessions, err := iterator.ToSlice(ctx, h.a.ListSessions(ctx, acp.ListSessionsRequest{Cwd: req.Cwd}))
	if err != nil {
		return acp.ListSessionsResponse{}, err
	}
	if sessions == nil {
		sessions = []acp.SessionInfo{}
	}
	return acp.ListSessionsResponse{Sessions: sessions}, nil
}

func (h *handler) Prompt(ctx context.Context, req acp.PromptRequest) (acp.PromptResponse, error) {
	ss, err := h.lookup(req.SessionId)
	if err != nil {
		return acp.PromptResponse{}, err
	}
	w := &turnWaiter{ch: make(chan acpapi.TurnEnd, 1)}
	ss.mu.Lock()
	if ss.turn != nil {
		ss.mu.Unlock()
		return acp.PromptResponse{}, acpapi.ErrBusy
	}
	ss.turn = w
	ss.mu.Unlock()
	if err := ss.s.Prompt(ctx, req.Prompt); err != nil {
		ss.mu.Lock()
		if ss.turn == w {
			ss.turn = nil
		}
		ss.mu.Unlock()
		return acp.PromptResponse{}, err
	}

	cancelled := ctx.Done()
	for {
		select {
		case end := <-w.ch:
			if end.Error != "" {
				return acp.PromptResponse{}, acp.NewInternalError(map[string]any{"error": end.Error})
			}
			if end.StopReason == "" {
				end.StopReason = acp.StopReasonEndTurn
			}
			return acp.PromptResponse{StopReason: end.StopReason}, nil
		case <-ss.done:
			return acp.PromptResponse{}, ss.endErr()
		case <-cancelled:
			cancelled = nil
			_ = ss.s.Cancel(h.ctx)
		}
	}
}

func (h *handler) Cancel(ctx context.Context, n acp.CancelNotification) error {
	ss, err := h.lookup(n.SessionId)
	if err != nil {
		return nil
	}
	return ss.s.Cancel(ctx)
}

func (h *handler) CloseSession(_ context.Context, req acp.CloseSessionRequest) (acp.CloseSessionResponse, error) {
	ss, err := h.lookup(req.SessionId)
	if err != nil {
		return acp.CloseSessionResponse{}, err
	}
	h.mu.Lock()
	delete(h.sessions, req.SessionId)
	h.mu.Unlock()
	return acp.CloseSessionResponse{}, ss.s.Close()
}

func (h *handler) SetSessionConfigOption(ctx context.Context, req acp.SetSessionConfigOptionRequest) (acp.SetSessionConfigOptionResponse, error) {
	var id acp.SessionId
	switch {
	case req.ValueId != nil:
		id = req.ValueId.SessionId
	case req.Boolean != nil:
		id = req.Boolean.SessionId
	}
	ss, err := h.lookup(id)
	if err != nil {
		return acp.SetSessionConfigOptionResponse{}, err
	}
	if err := ss.s.SetConfigOption(ctx, req); err != nil {
		return acp.SetSessionConfigOptionResponse{}, err
	}
	opts := ss.refresh().ConfigOptions
	if opts == nil {
		opts = []acp.SessionConfigOption{}
	}
	return acp.SetSessionConfigOptionResponse{ConfigOptions: opts}, nil
}

func (h *handler) SetSessionMode(context.Context, acp.SetSessionModeRequest) (acp.SetSessionModeResponse, error) {
	return acp.SetSessionModeResponse{}, acp.NewMethodNotFound(acp.AgentMethodSessionSetMode)
}

func (h *handler) ResumeSession(context.Context, acp.ResumeSessionRequest) (acp.ResumeSessionResponse, error) {
	return acp.ResumeSessionResponse{}, acp.NewMethodNotFound(acp.AgentMethodSessionResume)
}
