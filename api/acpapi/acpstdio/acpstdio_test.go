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

package acpstdio_test

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	acp "github.com/coder/acp-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/acpapi"
	"github.com/unstablebuild/rune-go-sdk/api/acpapi/acpstdio"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

// fakeAgent is an in-process harness. Prompts are scripts: "hello <x>"
// replies "hi <x>", "ask" requests a permission and reports the outcome,
// "wait" runs until cancelled.
type fakeAgent struct {
	mu       sync.Mutex
	n        int
	sessions map[acp.SessionId]*fakeSession
	done     chan struct{}
}

func newFakeAgent() *fakeAgent {
	return &fakeAgent{sessions: make(map[acp.SessionId]*fakeSession), done: make(chan struct{})}
}

func modelOption(current string) acp.SessionConfigOption {
	opt := acp.NewSessionConfigOptionSelect(acp.SessionConfigValueId(current), acp.SessionConfigSelectOptions{
		Ungrouped: &acp.SessionConfigSelectOptionsUngrouped{{Name: "Small", Value: "small"}, {Name: "Large", Value: "large"}},
	})
	opt.Select.Id = "model"
	opt.Select.Name = "Model"
	return opt
}

func (f *fakeAgent) Info() acp.InitializeResponse {
	return acp.InitializeResponse{
		ProtocolVersion: acp.ProtocolVersionNumber,
		AgentCapabilities: acp.AgentCapabilities{
			LoadSession:         true,
			SessionCapabilities: acp.SessionCapabilities{List: &acp.SessionListCapabilities{}},
		},
		AgentInfo: &acp.Implementation{Name: "fake", Version: "1"},
	}
}

func (f *fakeAgent) add(id acp.SessionId) *fakeSession {
	s := &fakeSession{id: id, log: acpapi.NewLog()}
	f.mu.Lock()
	f.sessions[id] = s
	f.mu.Unlock()
	return s
}

func (f *fakeAgent) session(id acp.SessionId) *fakeSession {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.sessions[id]
}

func (f *fakeAgent) NewSession(_ context.Context, req acp.NewSessionRequest) (acpapi.Session, error) {
	f.mu.Lock()
	f.n++
	id := acp.SessionId(fmt.Sprintf("s%d", f.n))
	f.mu.Unlock()
	s := f.add(id)
	s.cwd = req.Cwd
	s.log.Update(acp.SessionUpdate{AvailableCommandsUpdate: &acp.SessionAvailableCommandsUpdate{
		AvailableCommands: []acp.AvailableCommand{{Name: "review", Description: "Review changes"}},
	}})
	s.log.Started(acpapi.Started{ConfigOptions: []acp.SessionConfigOption{modelOption("small")}})
	return s, nil
}

func (f *fakeAgent) LoadSession(_ context.Context, req acp.LoadSessionRequest) (acpapi.Session, error) {
	s := f.add(req.SessionId)
	if err := s.log.Prompt([]acp.ContentBlock{acp.TextBlock("earlier")}); err != nil {
		return nil, err
	}
	s.log.Update(acp.UpdateAgentMessageText("reply"))
	s.log.Done(acpapi.TurnEnd{StopReason: acp.StopReasonEndTurn})
	s.log.Started(acpapi.Started{ConfigOptions: []acp.SessionConfigOption{modelOption("large")}})
	return s, nil
}

func (f *fakeAgent) ListSessions(context.Context, acp.ListSessionsRequest) iterator.Iterator[acp.SessionInfo] {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []acp.SessionInfo
	for id, s := range f.sessions {
		out = append(out, acp.SessionInfo{SessionId: id, Cwd: s.cwd})
	}
	return iterator.FromSlice(out)
}

func (f *fakeAgent) Done() <-chan struct{} { return f.done }
func (f *fakeAgent) Err() error            { return nil }
func (f *fakeAgent) Close() error          { return nil }

type fakeSession struct {
	id  acp.SessionId
	cwd string
	log *acpapi.Log

	mu     sync.Mutex
	cancel context.CancelFunc
	closed bool
}

func (s *fakeSession) Info() acpapi.SessionInfo                { return acpapi.SessionInfo{ID: s.id} }
func (s *fakeSession) Events() iterator.Iterator[acpapi.Event] { return s.log.Events() }

func (s *fakeSession) Prompt(_ context.Context, content []acp.ContentBlock) error {
	if err := s.log.Prompt(content); err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.cancel = cancel
	s.mu.Unlock()
	text := content[0].Text.Text
	go func() {
		defer cancel()
		switch {
		case strings.HasPrefix(text, "hello "):
			s.log.Update(acp.UpdateAgentMessageText("hi " + strings.TrimPrefix(text, "hello ")))
		case text == "ask":
			o, _ := s.log.Ask(ctx, acp.RequestPermissionRequest{
				SessionId: s.id,
				ToolCall:  acp.ToolCallUpdate{ToolCallId: "t1", Title: acp.Ptr("rm -rf")},
				Options: []acp.PermissionOption{
					{Kind: acp.PermissionOptionKindAllowOnce, Name: "Allow", OptionId: "allow"},
					{Kind: acp.PermissionOptionKindRejectOnce, Name: "Reject", OptionId: "reject"},
				},
			})
			if o.Selected != nil {
				s.log.Update(acp.UpdateAgentMessageText("allowed:" + string(o.Selected.OptionId)))
			} else {
				s.log.Update(acp.UpdateAgentMessageText("cancelled"))
			}
		case text == "wait":
			s.log.Update(acp.UpdateAgentMessageText("waiting"))
			<-ctx.Done()
		}
		if ctx.Err() != nil {
			s.log.Done(acpapi.TurnEnd{StopReason: acp.StopReasonCancelled})
			return
		}
		s.log.Done(acpapi.TurnEnd{StopReason: acp.StopReasonEndTurn})
	}()
	return nil
}

func (s *fakeSession) Cancel(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

func (s *fakeSession) Respond(_ context.Context, id acpapi.PermissionID, o acp.RequestPermissionOutcome) error {
	return s.log.Respond(id, o)
}

func (s *fakeSession) SetConfigOption(_ context.Context, req acp.SetSessionConfigOptionRequest) error {
	if req.ValueId == nil || req.ValueId.ConfigId != "model" {
		return errors.New("unknown option")
	}
	s.log.Update(acp.SessionUpdate{ConfigOptionUpdate: &acp.SessionConfigOptionUpdate{
		ConfigOptions: []acp.SessionConfigOption{modelOption(string(req.ValueId.Value))},
	}})
	return nil
}

func (s *fakeSession) Close() error {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
	s.log.Close(nil)
	return nil
}

func (s *fakeSession) isClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

type conn struct {
	agent acpapi.Agent
	// exit simulates the agent process dying with err.
	exit  func(err error)
	serve chan error
}

func connect(t *testing.T, f *fakeAgent) *conn {
	t.Helper()
	toAgentR, toAgentW := io.Pipe()
	toClientR, toClientW := io.Pipe()
	c := &conn{serve: make(chan error, 1)}
	ctx, cancel := context.WithCancel(context.Background())
	go func() { c.serve <- acpstdio.Serve(ctx, f, toAgentR, toClientW) }()
	c.exit = func(err error) {
		cancel()
		_ = toClientW.CloseWithError(err)
		_ = toAgentR.CloseWithError(err)
	}
	t.Cleanup(func() { c.exit(io.EOF) })

	dctx, dcancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dcancel()
	a, err := acpstdio.Dial(dctx, toAgentW, toClientR, acp.InitializeRequest{})
	require.NoError(t, err)
	c.agent = a
	t.Cleanup(func() { _ = a.Close() })
	return c
}

func newSession(t *testing.T, a acpapi.Agent) acpapi.Session {
	t.Helper()
	s, err := a.NewSession(context.Background(), acp.NewSessionRequest{Cwd: "/work"})
	require.NoError(t, err)
	return s
}

func text(ev acpapi.Event) string {
	if ev.Update == nil {
		return ""
	}
	switch {
	case ev.Update.AgentMessageChunk != nil && ev.Update.AgentMessageChunk.Content.Text != nil:
		return ev.Update.AgentMessageChunk.Content.Text.Text
	case ev.Update.UserMessageChunk != nil && ev.Update.UserMessageChunk.Content.Text != nil:
		return "user:" + ev.Update.UserMessageChunk.Content.Text.Text
	}
	return ""
}

// waitFor reads it until pred matches, failing after a timeout.
func waitFor(t *testing.T, it iterator.Iterator[acpapi.Event], pred func(acpapi.Event) bool) acpapi.Event {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		ev, ok := it.Next(ctx)
		require.True(t, ok, "events ended: %v", it.Err())
		if pred(ev) {
			return ev
		}
	}
}

func isDone(ev acpapi.Event) bool { return ev.Done != nil }

func hasText(s string) func(acpapi.Event) bool {
	return func(ev acpapi.Event) bool { return text(ev) == s }
}

// waitEnd reads it until it ends and returns its error.
func waitEnd(t *testing.T, it iterator.Iterator[acpapi.Event]) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		if _, ok := it.Next(ctx); !ok {
			require.NoError(t, ctx.Err(), "events did not end")
			return it.Err()
		}
	}
}

func TestTwoConcurrentSessions(t *testing.T) {
	c := connect(t, newFakeAgent())
	sessions := make([]acpapi.Session, 2)
	var wg sync.WaitGroup
	for i := range sessions {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sessions[i] = newSession(t, c.agent)
		}()
	}
	wg.Wait()
	require.NotEqual(t, sessions[0].Info().ID, sessions[1].Info().ID)

	its := make([]iterator.Iterator[acpapi.Event], 2)
	for i, s := range sessions {
		its[i] = s.Events()
		require.NoError(t, s.Prompt(context.Background(), []acp.ContentBlock{acp.TextBlock(fmt.Sprintf("hello %d", i))}))
	}
	for i, it := range its {
		waitFor(t, it, hasText(fmt.Sprintf("hi %d", i)))
		ev := waitFor(t, it, isDone)
		assert.Equal(t, acp.StopReasonEndTurn, ev.Done.StopReason)
	}
}

func TestUpdatesBeforeNewSessionResponse(t *testing.T) {
	c := connect(t, newFakeAgent())
	s := newSession(t, c.agent)
	it := s.Events()
	first := waitFor(t, it, func(acpapi.Event) bool { return true })
	require.NotNil(t, first.Update)
	require.NotNil(t, first.Update.AvailableCommandsUpdate)
	assert.Equal(t, "review", first.Update.AvailableCommandsUpdate.AvailableCommands[0].Name)
	second := waitFor(t, it, func(acpapi.Event) bool { return true })
	require.NotNil(t, second.Started)
	assert.Equal(t, []acp.SessionConfigOption{modelOption("small")}, second.Started.ConfigOptions)
}

func TestPromptBusy(t *testing.T) {
	c := connect(t, newFakeAgent())
	s := newSession(t, c.agent)
	require.NoError(t, s.Prompt(context.Background(), []acp.ContentBlock{acp.TextBlock("wait")}))
	assert.ErrorIs(t, s.Prompt(context.Background(), []acp.ContentBlock{acp.TextBlock("wait")}), acpapi.ErrBusy)
}

func TestPermission(t *testing.T) {
	f := newFakeAgent()
	c := connect(t, f)
	s := newSession(t, c.agent)
	it := s.Events()
	require.NoError(t, s.Prompt(context.Background(), []acp.ContentBlock{acp.TextBlock("ask")}))
	ev := waitFor(t, it, func(ev acpapi.Event) bool { return ev.Permission != nil })
	assert.Equal(t, "rm -rf", *ev.Permission.Request.ToolCall.Title)
	require.Len(t, ev.Permission.Request.Options, 2)

	require.NoError(t, s.Respond(context.Background(), ev.Permission.ID, acp.NewRequestPermissionOutcomeSelected("allow")))
	assert.ErrorIs(t, s.Respond(context.Background(), ev.Permission.ID, acp.NewRequestPermissionOutcomeSelected("reject")), acpapi.ErrResolved)
	waitFor(t, it, hasText("allowed:allow"))
	waitFor(t, it, isDone)
}

func TestPermissionResolvedOnAgentSide(t *testing.T) {
	f := newFakeAgent()
	c := connect(t, f)
	s := newSession(t, c.agent)
	it := s.Events()
	require.NoError(t, s.Prompt(context.Background(), []acp.ContentBlock{acp.TextBlock("ask")}))
	ev := waitFor(t, it, func(ev acpapi.Event) bool { return ev.Permission != nil })

	fs := f.session(s.Info().ID)
	pending := fs.log.State().Pending
	require.Len(t, pending, 1)
	require.NoError(t, fs.log.Respond(pending[0].ID, acp.NewRequestPermissionOutcomeSelected("reject")))

	res := waitFor(t, it, func(ev acpapi.Event) bool { return ev.Resolved != nil })
	assert.Equal(t, ev.Permission.ID, res.Resolved.ID)
	waitFor(t, it, hasText("allowed:reject"))
	waitFor(t, it, isDone)
}

func TestCancel(t *testing.T) {
	c := connect(t, newFakeAgent())
	s := newSession(t, c.agent)
	it := s.Events()

	require.NoError(t, s.Prompt(context.Background(), []acp.ContentBlock{acp.TextBlock("wait")}))
	waitFor(t, it, hasText("waiting"))
	require.NoError(t, s.Cancel(context.Background()))
	ev := waitFor(t, it, isDone)
	assert.Equal(t, acp.StopReasonCancelled, ev.Done.StopReason)

	require.NoError(t, s.Prompt(context.Background(), []acp.ContentBlock{acp.TextBlock("ask")}))
	perm := waitFor(t, it, func(ev acpapi.Event) bool { return ev.Permission != nil })
	require.NoError(t, s.Cancel(context.Background()))
	res := waitFor(t, it, func(ev acpapi.Event) bool { return ev.Resolved != nil })
	assert.Equal(t, perm.Permission.ID, res.Resolved.ID)
	assert.NotNil(t, res.Resolved.Outcome.Cancelled)
	ev = waitFor(t, it, isDone)
	assert.Equal(t, acp.StopReasonCancelled, ev.Done.StopReason)
}

func TestLoadSessionReplaysHistory(t *testing.T) {
	c := connect(t, newFakeAgent())
	s, err := c.agent.LoadSession(context.Background(), acp.LoadSessionRequest{SessionId: "old", Cwd: "/work"})
	require.NoError(t, err)
	assert.Equal(t, acp.SessionId("old"), s.Info().ID)

	require.NoError(t, s.Close())
	var got []string
	it := s.Events()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		ev, ok := it.Next(ctx)
		if !ok {
			break
		}
		switch {
		case ev.Started != nil:
			assert.Equal(t, []acp.SessionConfigOption{modelOption("large")}, ev.Started.ConfigOptions)
			got = append(got, "started")
		case ev.Update != nil:
			got = append(got, text(ev))
		default:
			t.Errorf("unexpected event %+v", ev)
		}
	}
	require.NoError(t, it.Err())
	assert.Equal(t, []string{"user:earlier", "reply", "started"}, got)
}

func TestCloseOneSessionLeavesOthers(t *testing.T) {
	f := newFakeAgent()
	c := connect(t, f)
	s1, s2 := newSession(t, c.agent), newSession(t, c.agent)
	it1 := s1.Events()
	require.NoError(t, s1.Close())
	assert.NoError(t, waitEnd(t, it1))
	assert.ErrorIs(t, s1.Prompt(context.Background(), []acp.ContentBlock{acp.TextBlock("hello")}), acpapi.ErrClosed)
	require.Eventually(t, func() bool { return f.session(s1.Info().ID).isClosed() }, 5*time.Second, time.Millisecond)

	it2 := s2.Events()
	require.NoError(t, s2.Prompt(context.Background(), []acp.ContentBlock{acp.TextBlock("hello two")}))
	waitFor(t, it2, hasText("hi two"))
	assert.False(t, f.session(s2.Info().ID).isClosed())
}

func TestAgentExitEndsAllSessions(t *testing.T) {
	c := connect(t, newFakeAgent())
	s1, s2 := newSession(t, c.agent), newSession(t, c.agent)
	it1, it2 := s1.Events(), s2.Events()
	c.exit(errors.New("agent exited: boom"))

	for _, it := range []iterator.Iterator[acpapi.Event]{it1, it2} {
		err := waitEnd(t, it)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "boom")
	}
	select {
	case <-c.agent.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("agent not done")
	}
	require.Error(t, c.agent.Err())
	_, err := c.agent.NewSession(context.Background(), acp.NewSessionRequest{Cwd: "/work"})
	assert.Error(t, err)
}

func TestSetConfigOption(t *testing.T) {
	c := connect(t, newFakeAgent())
	s := newSession(t, c.agent)
	it := s.Events()
	require.NoError(t, s.SetConfigOption(context.Background(), acp.SetSessionConfigOptionRequest{
		ValueId: &acp.SetSessionConfigOptionValueId{ConfigId: "model", Value: "large"},
	}))
	ev := waitFor(t, it, func(ev acpapi.Event) bool {
		return ev.Update != nil && ev.Update.ConfigOptionUpdate != nil
	})
	assert.Equal(t, []acp.SessionConfigOption{modelOption("large")}, ev.Update.ConfigOptionUpdate.ConfigOptions)
	assert.Error(t, s.SetConfigOption(context.Background(), acp.SetSessionConfigOptionRequest{
		ValueId: &acp.SetSessionConfigOptionValueId{ConfigId: "nope", Value: "x"},
	}))
}

func TestListSessions(t *testing.T) {
	c := connect(t, newFakeAgent())
	s := newSession(t, c.agent)
	got, err := iterator.ToSlice(context.Background(), c.agent.ListSessions(context.Background(), acp.ListSessionsRequest{}))
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, s.Info().ID, got[0].SessionId)
	assert.Equal(t, "/work", got[0].Cwd)
}

func TestServeReturnsWhenClientCloses(t *testing.T) {
	f := newFakeAgent()
	c := connect(t, f)
	s := newSession(t, c.agent)
	require.NoError(t, c.agent.Close())
	select {
	case err := <-c.serve:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return")
	}
	assert.True(t, f.session(s.Info().ID).isClosed())
}

func TestDialVersionMismatch(t *testing.T) {
	toAgentR, toAgentW := io.Pipe()
	toClientR, toClientW := io.Pipe()
	defer func() { _ = toClientW.Close() }()
	go func() {
		sc := bufio.NewScanner(toAgentR)
		if !sc.Scan() {
			return
		}
		var req struct {
			ID json.RawMessage `json:"id"`
		}
		_ = json.Unmarshal(sc.Bytes(), &req)
		_, _ = fmt.Fprintf(toClientW, `{"jsonrpc":"2.0","id":%s,"result":{"protocolVersion":2,"authMethods":[]}}`+"\n", req.ID)
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := acpstdio.Dial(ctx, toAgentW, toClientR, acp.InitializeRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "protocol version 2")
}

// rawAgent writes JSON-RPC lines itself, for traffic the acp import cannot
// send, and hands over what the client sends.
type rawAgent struct {
	t   *testing.T
	out io.Writer
	in  chan map[string]any
}

// dialRaw dials a rawAgent, returning it, the client and the initialize
// params the client sent.
func dialRaw(t *testing.T) (*rawAgent, acpapi.Agent, map[string]any) {
	t.Helper()
	toAgentR, toAgentW := io.Pipe()
	toClientR, toClientW := io.Pipe()
	ra := &rawAgent{t: t, out: toClientW, in: make(chan map[string]any, 64)}
	go func() {
		defer close(ra.in)
		sc := bufio.NewScanner(toAgentR)
		for sc.Scan() {
			var m map[string]any
			if json.Unmarshal(sc.Bytes(), &m) == nil {
				ra.in <- m
			}
		}
		// The pipe closing ends the scan; its error says nothing.
		_ = sc.Err()
	}()
	t.Cleanup(func() {
		_ = toClientW.Close()
		_ = toAgentR.Close()
	})
	type dialed struct {
		a   acpapi.Agent
		err error
	}
	ch := make(chan dialed, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		a, err := acpstdio.Dial(ctx, toAgentW, toClientR, acp.InitializeRequest{})
		ch <- dialed{a, err}
	}()
	init := ra.next()
	ra.reply(init, `{"protocolVersion":1,"agentCapabilities":{},"authMethods":[]}`)
	d := <-ch
	require.NoError(t, d.err)
	t.Cleanup(func() { _ = d.a.Close() })
	return ra, d.a, init["params"].(map[string]any)
}

func (ra *rawAgent) next() map[string]any {
	ra.t.Helper()
	select {
	case m, ok := <-ra.in:
		require.True(ra.t, ok, "client closed the connection")
		return m
	case <-time.After(5 * time.Second):
		ra.t.Fatal("client sent nothing")
		return nil
	}
}

func (ra *rawAgent) write(format string, args ...any) {
	_, err := fmt.Fprintf(ra.out, format+"\n", args...)
	require.NoError(ra.t, err)
}

func (ra *rawAgent) reply(req map[string]any, result string) {
	id, err := json.Marshal(req["id"])
	require.NoError(ra.t, err)
	ra.write(`{"jsonrpc":"2.0","id":%s,"result":%s}`, id, result)
}

func (ra *rawAgent) update(session acp.SessionId, update string) {
	ra.write(`{"jsonrpc":"2.0","method":"session/update","params":{"sessionId":%q,"update":%s}}`, session, update)
}

func (ra *rawAgent) askPermission(id int, session acp.SessionId) {
	ra.write(`{"jsonrpc":"2.0","id":%d,"method":"session/request_permission","params":{"sessionId":%q,`+
		`"toolCall":{"toolCallId":"t1"},"options":[{"optionId":"allow","name":"Allow","kind":"allow_once"}]}}`, id, session)
}

// outcome reads the client's answer to the permission request id.
func (ra *rawAgent) outcome(id int) map[string]any {
	ra.t.Helper()
	m := ra.next()
	require.Equal(ra.t, float64(id), m["id"], "%v", m)
	return m["result"].(map[string]any)["outcome"].(map[string]any)
}

func newRawSession(t *testing.T, ra *rawAgent, a acpapi.Agent) acpapi.Session {
	t.Helper()
	ch := make(chan acpapi.Session, 1)
	go func() { ch <- newSession(t, a) }()
	req := ra.next()
	require.Equal(t, "session/new", req["method"])
	ra.reply(req, `{"sessionId":"root"}`)
	return <-ch
}

// describe renders ev as "[session] kind detail".
func describe(ev acpapi.Event) string {
	switch {
	case ev.Started != nil:
		return "started"
	case ev.Subagent != nil:
		return fmt.Sprintf("[%s] subagent %s", ev.Session, ev.Subagent.SessionID)
	case ev.Message != nil:
		return fmt.Sprintf("[%s] message %s", ev.Session, ev.Message.MessageID)
	case ev.MessageChunk != nil:
		return fmt.Sprintf("[%s] chunk %s", ev.Session, ev.MessageChunk.MessageID)
	case ev.Update != nil && ev.Update.ToolCall != nil:
		return fmt.Sprintf("[%s] tool_call %s", ev.Session, ev.Update.ToolCall.ToolCallId)
	case ev.Update != nil && ev.Update.AgentMessageChunk != nil:
		return fmt.Sprintf("[%s] text %s", ev.Session, ev.Update.AgentMessageChunk.Content.Text.Text)
	case ev.Permission != nil:
		return fmt.Sprintf("[%s] permission", ev.Permission.Session)
	case ev.Resolved != nil:
		return "resolved"
	}
	return fmt.Sprintf("%+v", ev)
}

// events reads n events from it, described.
func events(t *testing.T, it iterator.Iterator[acpapi.Event], n int) []string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var got []string
	for len(got) < n {
		ev, ok := it.Next(ctx)
		require.True(t, ok, "events ended after %v: %v", got, it.Err())
		got = append(got, describe(ev))
	}
	return got
}

func TestDialAdvertisesSubagents(t *testing.T) {
	_, _, params := dialRaw(t)
	caps := params["clientCapabilities"].(map[string]any)
	assert.Equal(t, map[string]any{}, caps["subagents"])
	assert.Equal(t, map[string]any{}, caps["fs"])
	assert.NotContains(t, caps, "terminal")
}

func TestSubagentTraffic(t *testing.T) {
	ra, a, _ := dialRaw(t)
	s := newRawSession(t, ra, a)
	it := s.Events()
	defer func() { _ = it.Close() }()

	ra.update("root", `{"sessionUpdate":"subagent_update","sessionId":"a","title":"Explore",`+
		`"capabilities":{"cancel":{}},"state":{"state":"running"}}`)
	ra.update("root", `{"sessionUpdate":"session_message","messageId":"m1","senderSessionId":"root",`+
		`"recipientSessionId":"a","content":[{"type":"text","text":"look"}]}`)
	ra.update("a", `{"sessionUpdate":"tool_call","toolCallId":"t1","title":"Run ls","kind":"execute","status":"pending"}`)
	ra.update("a", `{"sessionUpdate":"subagent_update","sessionId":"b"}`)
	ra.update("b", `{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"deep"}}`)
	ra.update("a", `{"sessionUpdate":"session_message_chunk","messageId":"m2","content":{"type":"text","text":"found"}}`)
	ra.write(`{"jsonrpc":"2.0","method":"_vendor/ping","params":{}}`)
	ra.update("c", `{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"early"}}`)
	ra.update("root", `{"sessionUpdate":"subagent_update","sessionId":"c"}`)
	ra.update("root", `{"sessionUpdate":"subagent_update","sessionId":"a","state":{"state":"idle","stopReason":"end_turn"}}`)
	ra.update("root", `{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"done"}}`)

	assert.Equal(t, []string{
		"started",
		"[] subagent a",
		"[] message m1",
		"[a] tool_call t1",
		"[a] subagent b",
		"[b] text deep",
		"[a] chunk m2",
		"[] subagent c",
		"[c] text early",
		"[] subagent a",
		"[] text done",
	}, events(t, it, 11))

	idle := acpapi.SubagentState{State: acpapi.SubagentIdle, StopReason: acp.StopReasonEndTurn}
	var got acpapi.SessionState
	all := s.Events()
	defer func() { _ = all.Close() }()
	for range 11 {
		ev, _ := all.Next(context.Background())
		acpapi.Apply(&got, ev)
	}
	assert.Equal(t, map[acp.SessionId]acpapi.Subagent{
		"a": {Title: "Explore", Capabilities: &acpapi.SubagentCapabilities{Cancel: &acpapi.Capability{}}, State: &idle},
		"b": {Parent: "a"},
		"c": {},
	}, got.Subagents)
}

func TestSubagentPermissions(t *testing.T) {
	ra, a, _ := dialRaw(t)
	s := newRawSession(t, ra, a)
	it := s.Events()
	defer func() { _ = it.Close() }()
	ra.update("root", `{"sessionUpdate":"subagent_update","sessionId":"a"}`)

	// A request is handled as it arrives, so it can overtake the
	// announcement of its subagent.
	ra.askPermission(1, "d")
	ra.update("root", `{"sessionUpdate":"subagent_update","sessionId":"d"}`)
	assert.Equal(t, []string{"started", "[] subagent a", "[] subagent d", "[d] permission"}, events(t, it, 4))
	pending := waitPermissions(t, s, 1)
	require.NoError(t, s.Respond(context.Background(), pending[0].ID, acp.NewRequestPermissionOutcomeSelected("allow")))
	assert.Equal(t, map[string]any{"outcome": "selected", "optionId": "allow"}, ra.outcome(1))

	ra.askPermission(2, "root")
	waitPermissions(t, s, 1)
	ra.askPermission(3, "a")
	waitPermissions(t, s, 2)
	require.NoError(t, s.Cancel(context.Background()))
	var got []string
	for range 2 {
		m := ra.next()
		if m["method"] != nil {
			got = append(got, fmt.Sprintf("%v %v", m["method"], m["params"].(map[string]any)["sessionId"]))
		} else {
			got = append(got, fmt.Sprintf("%v %v", m["id"], m["result"].(map[string]any)["outcome"]))
		}
	}
	assert.ElementsMatch(t, []string{"session/cancel root", "2 map[outcome:cancelled]"}, got)
	pending = waitPermissions(t, s, 1)
	assert.Equal(t, acp.SessionId("a"), pending[0].Session)

	ra.update("root", `{"sessionUpdate":"subagent_update","sessionId":"a","state":{"state":"idle","stopReason":"cancelled"}}`)
	assert.Equal(t, map[string]any{"outcome": "cancelled"}, ra.outcome(3))
	waitPermissions(t, s, 0)
}

// waitPermissions waits for s to have n pending permissions and returns
// them.
func waitPermissions(t *testing.T, s acpapi.Session, n int) []acpapi.Permission {
	t.Helper()
	var state acpapi.SessionState
	it := s.Events()
	defer func() { _ = it.Close() }()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		ev, ok := it.Next(ctx)
		cancel()
		if ok {
			acpapi.Apply(&state, ev)
			continue
		}
		if len(state.Pending) == n {
			return state.Pending
		}
	}
	t.Fatalf("want %d pending permissions, have %+v", n, state.Pending)
	return nil
}

func TestErrorText(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{errors.New("plain"), "plain"},
		{acp.NewInternalError(map[string]any{"error": "boom"}), "boom"},
		{fmt.Errorf("wrap: %w", acp.NewInternalError(map[string]any{"details": "why"})), "why"},
		{&acp.RequestError{Code: 1, Message: "Bad"}, "Bad"},
		{&acp.RequestError{Code: 1, Message: "Bad", Data: []int{1}}, "Bad: [1]"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, acpstdio.ErrorText(tc.err))
	}
}
