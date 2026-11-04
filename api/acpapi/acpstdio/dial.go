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

// Package acpstdio speaks ACP v1 over a byte stream, typically an agent
// process's stdio. Dial is the client side and Serve the agent side, both on
// top of acpapi's interfaces.
package acpstdio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"sync"
	"time"

	acp "github.com/coder/acp-go-sdk"
	"github.com/unstablebuild/rune-go-sdk/api/acpapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

// closeTimeout bounds the session/close or session/cancel Session.Close sends.
const closeTimeout = 5 * time.Second

// maxEarlyUpdates caps the updates buffered for a session ID that is not yet
// registered.
const maxEarlyUpdates = 1024

// announceTimeout bounds how long a permission request for an unknown
// session waits for the session to be registered or announced, as requests
// are handled as they arrive while notifications queue.
const announceTimeout = 10 * time.Second

// ErrUnsupported is returned for operations the agent did not advertise.
var ErrUnsupported = errors.New("acp: unsupported by the agent")

// Dial initializes an ACP v1 connection to the agent reading w and writing r.
// ctx bounds the handshake only. A zero req.ProtocolVersion is sent as 1, and
// an agent answering with another version is an error. No fs or terminal
// capabilities are offered, whatever req says. The unstable subagents
// capability is: the subagents an agent announces are logged in the session
// that owns them, nested ones included, with their traffic and permissions
// attributed to them.
//
// The connection ends when r does, which ends every session with an error
// built from r's final read error. Close also closes w when it is an
// io.Closer. On error the caller still owns w and r.
func Dial(ctx context.Context, w io.Writer, r io.Reader, req acp.InitializeRequest) (acpapi.Agent, error) {
	a := &agent{
		r:         &errReader{r: r},
		w:         w,
		sessions:  make(map[acp.SessionId]*session),
		subagents: make(map[acp.SessionId]*session),
		early:     make(map[acp.SessionId][]inbound),
		ended:     make(map[acp.SessionId]bool),
		wake:      make(chan struct{}),
		done:      make(chan struct{}),
	}
	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.conn = acp.NewConnection(a.handle, w, a.r)

	if req.ProtocolVersion == 0 {
		req.ProtocolVersion = acp.ProtocolVersionNumber
	}
	req.ClientCapabilities.Fs = acp.FileSystemCapabilities{}
	req.ClientCapabilities.Terminal = false
	info, err := acp.SendRequest[acp.InitializeResponse](a.conn, ctx, acp.AgentMethodInitialize, initializeRequest{req})
	if err != nil {
		a.cancel()
		return nil, fmt.Errorf("acp: initialize: %w", a.connErr(err))
	}
	if info.ProtocolVersion != acp.ProtocolVersionNumber {
		a.cancel()
		return nil, fmt.Errorf("acp: agent speaks protocol version %d, want %d",
			info.ProtocolVersion, acp.ProtocolVersionNumber)
	}
	a.info = info
	go a.watch()
	return a, nil
}

type agent struct {
	conn   *acp.Connection
	info   acp.InitializeResponse
	r      *errReader
	w      io.Writer
	ctx    context.Context
	cancel context.CancelFunc

	mu       sync.Mutex
	sessions map[acp.SessionId]*session
	// subagents maps each announced subagent to the session that logs it.
	subagents map[acp.SessionId]*session
	// early is the traffic for session IDs not yet registered or announced.
	early map[acp.SessionId][]inbound
	ended map[acp.SessionId]bool
	// wake is closed and replaced whenever a session ID becomes known.
	wake   chan struct{}
	closed bool
	err    error
	done   chan struct{}
}

func (a *agent) Info() acp.InitializeResponse { return a.info }

func (a *agent) NewSession(ctx context.Context, req acp.NewSessionRequest) (acpapi.Session, error) {
	if req.McpServers == nil {
		req.McpServers = []acp.McpServer{}
	}
	resp, err := acp.SendRequest[acp.NewSessionResponse](a.conn, ctx, acp.AgentMethodSessionNew, req)
	if err != nil {
		return nil, fmt.Errorf("acp: session/new: %w", a.connErr(err))
	}
	s := a.newSession(resp.SessionId)
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return nil, a.endErr()
	}
	a.sessions[resp.SessionId] = s
	a.knownLocked(resp.SessionId, s, "")
	a.mu.Unlock()
	s.log.Started(acpapi.Started{ConfigOptions: resp.ConfigOptions, Modes: resp.Modes})
	return s, nil
}

func (a *agent) LoadSession(ctx context.Context, req acp.LoadSessionRequest) (acpapi.Session, error) {
	if !a.info.AgentCapabilities.LoadSession {
		return nil, fmt.Errorf("acp: session/load: %w", ErrUnsupported)
	}
	if req.McpServers == nil {
		req.McpServers = []acp.McpServer{}
	}
	s := a.newSession(req.SessionId)
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return nil, a.endErr()
	}
	if _, ok := a.sessions[req.SessionId]; ok {
		a.mu.Unlock()
		return nil, fmt.Errorf("acp: session %s: %w", req.SessionId, acpapi.ErrExists)
	}
	delete(a.ended, req.SessionId)
	a.sessions[req.SessionId] = s
	a.knownLocked(req.SessionId, s, "")
	a.mu.Unlock()

	resp, err := acp.SendRequest[acp.LoadSessionResponse](a.conn, ctx, acp.AgentMethodSessionLoad, req)
	if err != nil {
		err = fmt.Errorf("acp: session/load: %w", a.connErr(err))
		a.forget(s)
		s.log.Close(err)
		return nil, err
	}
	s.log.Started(acpapi.Started{ConfigOptions: resp.ConfigOptions, Modes: resp.Modes})
	return s, nil
}

func (a *agent) ListSessions(_ context.Context, req acp.ListSessionsRequest) iterator.Iterator[acp.SessionInfo] {
	if a.info.AgentCapabilities.SessionCapabilities.List == nil {
		return iterator.Error[acp.SessionInfo](fmt.Errorf("acp: session/list: %w", ErrUnsupported))
	}
	return &sessionPager{a: a, req: req}
}

func (a *agent) Done() <-chan struct{} { return a.done }

func (a *agent) Err() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.err
}

func (a *agent) Close() error {
	a.end(nil)
	if c, ok := a.w.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (a *agent) newSession(id acp.SessionId) *session {
	return &session{
		a:    a,
		info: acpapi.SessionInfo{ID: id},
		log:  acpapi.NewLog(),
	}
}

func (a *agent) watch() {
	<-a.conn.Done()
	a.end(a.cause())
}

// cause is why the connection ended, from the reader's final error.
func (a *agent) cause() error {
	err := a.r.Err()
	if err == nil || errors.Is(err, io.EOF) {
		return errors.New("acp: agent closed the connection")
	}
	return err
}

// end closes every session with err and marks the agent done. Only the first
// call has effect.
func (a *agent) end(err error) {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return
	}
	a.closed = true
	a.err = err
	sessions := slices.Collect(maps.Values(a.sessions))
	clear(a.sessions)
	clear(a.subagents)
	clear(a.early)
	close(a.done)
	close(a.wake)
	a.mu.Unlock()
	a.cancel()
	for _, s := range sessions {
		s.log.Close(err)
	}
}

// endErr is the error for operations attempted after the connection ended.
func (a *agent) endErr() error {
	if err := a.Err(); err != nil {
		return err
	}
	return acpapi.ErrClosed
}

// connErr prefers the connection's end cause over the generic error requests
// fail with once the peer is gone.
func (a *agent) connErr(err error) error {
	if err == nil {
		return nil
	}
	select {
	case <-a.conn.Done():
		return a.cause()
	default:
	}
	return err
}

func (a *agent) forget(s *session) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.sessions[s.info.ID] == s {
		delete(a.sessions, s.info.ID)
		a.ended[s.info.ID] = true
	}
	for id, owner := range a.subagents {
		if owner == s {
			delete(a.subagents, id)
			a.ended[id] = true
		}
	}
}

// ownerLocked returns the session that logs the traffic on session ID id,
// and from, the subagent id is, or empty when id is that session's own.
func (a *agent) ownerLocked(id acp.SessionId) (s *session, from acp.SessionId) {
	if s := a.sessions[id]; s != nil {
		return s, ""
	}
	if s := a.subagents[id]; s != nil {
		return s, id
	}
	return nil, ""
}

// knownLocked delivers the early traffic on session ID id, which s now logs
// as from's, and wakes the requests waiting for it.
func (a *agent) knownLocked(id acp.SessionId, s *session, from acp.SessionId) {
	early := a.early[id]
	delete(a.early, id)
	for _, in := range early {
		a.deliverLocked(s, from, in)
	}
	close(a.wake)
	a.wake = make(chan struct{})
}

// inbound is a session/update, held until its session ID is known. Exactly
// one field is set.
type inbound struct {
	update   *acp.SessionUpdate
	subagent *acpapi.SubagentUpdate
	message  *acpapi.SessionMessage
	chunk    *acpapi.SessionMessageChunk
}

// route logs in, which arrived on session ID id, or holds it until id is
// known. Notifications are handled one at a time, so routing under mu keeps
// them in order across sessions.
func (a *agent) route(id acp.SessionId, in inbound) {
	a.mu.Lock()
	defer a.mu.Unlock()
	s, from := a.ownerLocked(id)
	if s == nil {
		if !a.closed && !a.ended[id] && len(a.early[id]) < maxEarlyUpdates {
			a.early[id] = append(a.early[id], in)
		}
		return
	}
	a.deliverLocked(s, from, in)
}

// deliverLocked logs in on s as from's traffic.
func (a *agent) deliverLocked(s *session, from acp.SessionId, in inbound) {
	switch {
	case in.update != nil && from == "":
		s.log.Update(*in.update)
	case in.update != nil:
		s.log.UpdateSubagent(from, *in.update)
	case in.subagent != nil:
		s.log.Subagent(from, *in.subagent)
		child := in.subagent.SessionID
		if a.sessions[child] == nil && a.subagents[child] == nil && !a.ended[child] {
			a.subagents[child] = s
			a.knownLocked(child, s, child)
		}
	case in.message != nil:
		s.log.Message(from, *in.message)
	case in.chunk != nil:
		s.log.MessageChunk(from, *in.chunk)
	}
}

// awaitOwner returns the session that logs session ID id once it is known,
// or nil when it is not known within announceTimeout or ends first.
func (a *agent) awaitOwner(ctx context.Context, id acp.SessionId) *session {
	timer := time.NewTimer(announceTimeout)
	defer timer.Stop()
	for {
		a.mu.Lock()
		s, _ := a.ownerLocked(id)
		if s != nil || a.closed || a.ended[id] {
			a.mu.Unlock()
			return s
		}
		wake := a.wake
		a.mu.Unlock()
		select {
		case <-wake:
		case <-ctx.Done():
			return nil
		case <-timer.C:
			return nil
		}
	}
}

type session struct {
	a    *agent
	info acpapi.SessionInfo
	log  *acpapi.Log
}

func (s *session) Info() acpapi.SessionInfo { return s.info }

func (s *session) Events() iterator.Iterator[acpapi.Event] { return s.log.Events() }

func (s *session) Prompt(_ context.Context, content []acp.ContentBlock) error {
	if err := s.log.Prompt(content); err != nil {
		return err
	}
	go func() {
		resp, err := acp.SendRequest[acp.PromptResponse](s.a.conn, s.a.ctx, acp.AgentMethodSessionPrompt,
			acp.PromptRequest{SessionId: s.info.ID, Prompt: content})
		if err != nil {
			s.log.Done(acpapi.TurnEnd{Error: ErrorText(s.a.connErr(err))})
			return
		}
		s.log.Done(acpapi.TurnEnd{StopReason: resp.StopReason})
	}()
	return nil
}

// Cancel resolves the session's own pending permissions Cancelled. The
// subagents' stay pending until the agent reports their work cancelled.
func (s *session) Cancel(ctx context.Context) error {
	cancelled := acp.NewRequestPermissionOutcomeCancelled()
	for _, p := range s.log.State().Pending {
		if p.Session == "" {
			_ = s.log.Respond(p.ID, cancelled)
		}
	}
	return s.a.conn.SendNotification(ctx, acp.AgentMethodSessionCancel, acp.CancelNotification{SessionId: s.info.ID})
}

func (s *session) Respond(_ context.Context, id acpapi.PermissionID, o acp.RequestPermissionOutcome) error {
	return s.log.Respond(id, o)
}

func (s *session) SetConfigOption(ctx context.Context, req acp.SetSessionConfigOptionRequest) error {
	switch {
	case req.ValueId != nil:
		v := *req.ValueId
		v.SessionId = s.info.ID
		req.ValueId = &v
	case req.Boolean != nil:
		v := *req.Boolean
		v.SessionId = s.info.ID
		req.Boolean = &v
	default:
		return errors.New("acp: empty config option request")
	}
	resp, err := acp.SendRequest[acp.SetSessionConfigOptionResponse](s.a.conn, ctx,
		acp.AgentMethodSessionSetConfigOption, req)
	if err != nil {
		return fmt.Errorf("acp: session/set_config_option: %w", s.a.connErr(err))
	}
	s.log.Update(acp.SessionUpdate{ConfigOptionUpdate: &acp.SessionConfigOptionUpdate{
		ConfigOptions: resp.ConfigOptions,
	}})
	return nil
}

// Close sends session/close when the agent supports it, else session/cancel.
func (s *session) Close() error {
	s.a.mu.Lock()
	live := s.a.sessions[s.info.ID] == s
	s.a.mu.Unlock()
	s.a.forget(s)
	s.log.Close(nil)
	if !live {
		return nil
	}
	ctx, cancel := context.WithTimeout(s.a.ctx, closeTimeout)
	defer cancel()
	if s.a.info.AgentCapabilities.SessionCapabilities.Close != nil {
		_, err := acp.SendRequest[acp.CloseSessionResponse](s.a.conn, ctx, acp.AgentMethodSessionClose,
			acp.CloseSessionRequest{SessionId: s.info.ID})
		return s.a.connErr(err)
	}
	return s.a.connErr(s.a.conn.SendNotification(ctx, acp.AgentMethodSessionCancel,
		acp.CancelNotification{SessionId: s.info.ID}))
}

// handle serves the agent's requests and notifications. It decodes them
// itself because the acp import does not model the subagent updates yet.
// The fs and terminal methods are not offered, and extension notifications
// are ignored.
func (a *agent) handle(ctx context.Context, method string, params json.RawMessage) (any, *acp.RequestError) {
	switch method {
	case acp.ClientMethodSessionUpdate:
		return nil, a.sessionUpdate(params)
	case acp.ClientMethodSessionRequestPermission:
		var req acp.RequestPermissionRequest
		if err := decode(params, &req); err != nil {
			return nil, err
		}
		return a.requestPermission(ctx, req)
	default:
		return nil, acp.NewMethodNotFound(method)
	}
}

func (a *agent) sessionUpdate(params json.RawMessage) *acp.RequestError {
	var n struct {
		SessionID acp.SessionId `json:"sessionId"`
		Update    struct {
			Kind string `json:"sessionUpdate"`
		} `json:"update"`
	}
	if err := json.Unmarshal(params, &n); err != nil {
		return invalidParams(err)
	}
	var in inbound
	var err *acp.RequestError
	switch n.Update.Kind {
	case "subagent_update":
		in.subagent, err = decodeUpdate[acpapi.SubagentUpdate](params)
		if err == nil && in.subagent.SessionID == "" {
			err = invalidParams(errors.New("subagent_update without sessionId"))
		}
	case "session_message":
		in.message, err = decodeUpdate[acpapi.SessionMessage](params)
	case "session_message_chunk":
		in.chunk, err = decodeUpdate[acpapi.SessionMessageChunk](params)
	default:
		var sn acp.SessionNotification
		if err := decode(params, &sn); err != nil {
			return err
		}
		in.update = &sn.Update
	}
	if err != nil {
		return err
	}
	if n.SessionID == "" {
		return invalidParams(errors.New("session/update without sessionId"))
	}
	a.route(n.SessionID, in)
	return nil
}

func (a *agent) requestPermission(ctx context.Context, req acp.RequestPermissionRequest) (any, *acp.RequestError) {
	s := a.awaitOwner(ctx, req.SessionId)
	if s == nil {
		return acp.RequestPermissionResponse{Outcome: acp.NewRequestPermissionOutcomeCancelled()}, nil
	}
	o, err := s.log.Ask(ctx, req)
	if err != nil && !errors.Is(err, acpapi.ErrClosed) {
		return nil, acp.NewInternalError(map[string]any{"error": err.Error()})
	}
	return acp.RequestPermissionResponse{Outcome: o}, nil
}

// decode unmarshals params into v and validates it.
func decode[T interface{ Validate() error }](params json.RawMessage, v T) *acp.RequestError {
	if err := json.Unmarshal(params, v); err != nil {
		return invalidParams(err)
	}
	if err := v.Validate(); err != nil {
		return invalidParams(err)
	}
	return nil
}

// decodeUpdate unmarshals the update of a session/update.
func decodeUpdate[T any](params json.RawMessage) (*T, *acp.RequestError) {
	var n struct {
		Update T `json:"update"`
	}
	if err := json.Unmarshal(params, &n); err != nil {
		return nil, invalidParams(err)
	}
	return &n.Update, nil
}

func invalidParams(err error) *acp.RequestError {
	return acp.NewInvalidParams(map[string]any{"error": err.Error()})
}

// initializeRequest marshals as its request with the subagents client
// capability, which acp.ClientCapabilities does not model yet.
type initializeRequest struct {
	req acp.InitializeRequest
}

func (r initializeRequest) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(r.req)
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return nil, err
	}
	caps := map[string]json.RawMessage{}
	if raw, ok := fields["clientCapabilities"]; ok {
		if err := json.Unmarshal(raw, &caps); err != nil {
			return nil, err
		}
	}
	caps["subagents"] = json.RawMessage(`{}`)
	if fields["clientCapabilities"], err = json.Marshal(caps); err != nil {
		return nil, err
	}
	return json.Marshal(fields)
}

type sessionPager struct {
	a    *agent
	req  acp.ListSessionsRequest
	page []acp.SessionInfo
	last bool
	err  error
}

func (p *sessionPager) Next(ctx context.Context) (acp.SessionInfo, bool) {
	for len(p.page) == 0 {
		if p.last || p.err != nil {
			return acp.SessionInfo{}, false
		}
		resp, err := acp.SendRequest[acp.ListSessionsResponse](p.a.conn, ctx, acp.AgentMethodSessionList, p.req)
		if err != nil {
			p.err = fmt.Errorf("acp: session/list: %w", p.a.connErr(err))
			return acp.SessionInfo{}, false
		}
		p.page = resp.Sessions
		p.req.Cursor = resp.NextCursor
		p.last = resp.NextCursor == nil
	}
	s := p.page[0]
	p.page = p.page[1:]
	return s, true
}

func (p *sessionPager) Err() error   { return p.err }
func (p *sessionPager) Close() error { return nil }

// errReader records the first error its reader returns. When ready is set,
// reads wait for it to close so handlers never run before their connection
// is assigned.
type errReader struct {
	r     io.Reader
	ready <-chan struct{}
	mu    sync.Mutex
	err   error
}

func (e *errReader) Read(p []byte) (int, error) {
	if e.ready != nil {
		<-e.ready
	}
	n, err := e.r.Read(p)
	if err != nil {
		e.mu.Lock()
		if e.err == nil {
			e.err = err
		}
		e.mu.Unlock()
	}
	return n, err
}

func (e *errReader) Err() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.err
}

// ErrorText renders err for users, unwrapping the JSON-RPC error envelope
// agents answer with.
func ErrorText(err error) string {
	var re *acp.RequestError
	if !errors.As(err, &re) {
		return err.Error()
	}
	if data, ok := re.Data.(map[string]any); ok {
		for _, k := range []string{"error", "details", "message"} {
			if s, ok := data[k].(string); ok && s != "" {
				return s
			}
		}
	}
	if re.Data != nil {
		if b, err := json.Marshal(re.Data); err == nil {
			return re.Message + ": " + string(b)
		}
	}
	return re.Message
}
