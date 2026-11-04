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

package acprpc_test

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	acp "github.com/coder/acp-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/acpapi"
	"github.com/unstablebuild/rune-go-sdk/api/acpapi/acprpc"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type fakeService struct {
	mu       sync.Mutex
	sessions map[string]*fakeSession
	lastNew  acpapi.NewSessionRequest
}

func (f *fakeService) Agents(context.Context) iterator.Iterator[acpapi.AgentInfo] {
	return iterator.FromSlice([]acpapi.AgentInfo{{Name: "claude", Running: true}, {Name: "codex"}})
}

func (f *fakeService) NewSession(_ context.Context, req acpapi.NewSessionRequest) (acpapi.Session, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastNew = req
	if req.Agent != "claude" {
		return nil, acpapi.ErrNotFound
	}
	if req.Name == "" {
		req.Name = "brave-otter"
	}
	if _, ok := f.sessions[req.Name]; ok {
		return nil, acpapi.ErrExists
	}
	s := &fakeSession{info: acpapi.SessionInfo{Name: req.Name, Agent: req.Agent, ID: "id-" + acp.SessionId(req.Name)}, log: acpapi.NewLog()}
	s.log.Started(acpapi.Started{})
	f.sessions[req.Name] = s
	return s, nil
}

func (f *fakeService) Sessions(context.Context) iterator.Iterator[acpapi.SessionInfo] {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []acpapi.SessionInfo
	for _, s := range f.sessions {
		out = append(out, s.info)
	}
	return iterator.FromSlice(out)
}

func (f *fakeService) OpenSession(_ context.Context, name string) (acpapi.Session, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s, ok := f.sessions[name]; ok {
		return s, nil
	}
	return nil, acpapi.ErrNotFound
}

type fakeSession struct {
	info acpapi.SessionInfo
	log  *acpapi.Log

	mu        sync.Mutex
	cancelled bool
	config    acp.SetSessionConfigOptionRequest
}

func (s *fakeSession) Info() acpapi.SessionInfo                { return s.info }
func (s *fakeSession) Events() iterator.Iterator[acpapi.Event] { return s.log.Events() }
func (s *fakeSession) Prompt(_ context.Context, c []acp.ContentBlock) error {
	return s.log.Prompt(c)
}

func (s *fakeSession) Cancel(context.Context) error {
	s.mu.Lock()
	s.cancelled = true
	s.mu.Unlock()
	s.log.Done(acpapi.TurnEnd{StopReason: acp.StopReasonCancelled})
	return nil
}

func (s *fakeSession) Respond(_ context.Context, id acpapi.PermissionID, o acp.RequestPermissionOutcome) error {
	return s.log.Respond(id, o)
}

func (s *fakeSession) SetConfigOption(_ context.Context, req acp.SetSessionConfigOptionRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = req
	return nil
}

func (s *fakeSession) Close() error {
	s.log.Close(nil)
	return nil
}

func serve(t *testing.T, svc acpapi.Service) *acprpc.Client {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	acprpc.RegisterACPServer(srv, acprpc.NewServer(svc))
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	cc, err := grpc.NewClient("passthrough:///bufconn",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cc.Close() })
	return acprpc.NewClient(context.Background(), cc)
}

func newService() *fakeService {
	return &fakeService{sessions: make(map[string]*fakeSession)}
}

func next(t *testing.T, it iterator.Iterator[acpapi.Event]) acpapi.Event {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ev, ok := it.Next(ctx)
	require.True(t, ok, "Next: %v", it.Err())
	return ev
}

func end(t *testing.T, it iterator.Iterator[acpapi.Event]) error {
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

func TestAgentsAndSessions(t *testing.T) {
	svc := newService()
	c := serve(t, svc)
	agents, err := iterator.ToSlice(context.Background(), c.Agents(context.Background()))
	require.NoError(t, err)
	assert.Equal(t, []acpapi.AgentInfo{{Name: "claude", Running: true}, {Name: "codex"}}, agents)

	s, err := c.NewSession(context.Background(), acpapi.NewSessionRequest{
		Agent:   "claude",
		Session: acp.NewSessionRequest{Cwd: "/work", Meta: map[string]any{"k": "v"}},
	})
	require.NoError(t, err)
	want := acpapi.SessionInfo{Name: "brave-otter", Agent: "claude", ID: "id-brave-otter"}
	assert.Equal(t, want, s.Info())
	assert.Equal(t, "/work", svc.lastNew.Session.Cwd)
	assert.Equal(t, map[string]any{"k": "v"}, svc.lastNew.Session.Meta)

	sessions, err := iterator.ToSlice(context.Background(), c.Sessions(context.Background()))
	require.NoError(t, err)
	assert.Equal(t, []acpapi.SessionInfo{want}, sessions)

	opened, err := c.OpenSession(context.Background(), "brave-otter")
	require.NoError(t, err)
	assert.Equal(t, want, opened.Info())
}

func TestErrors(t *testing.T) {
	c := serve(t, newService())
	_, err := c.NewSession(context.Background(), acpapi.NewSessionRequest{Agent: "nope"})
	assert.ErrorIs(t, err, acpapi.ErrNotFound)
	_, err = c.NewSession(context.Background(), acpapi.NewSessionRequest{Agent: "claude", Name: "x"})
	require.NoError(t, err)
	_, err = c.NewSession(context.Background(), acpapi.NewSessionRequest{Agent: "claude", Name: "x"})
	assert.ErrorIs(t, err, acpapi.ErrExists)
	_, err = c.OpenSession(context.Background(), "missing")
	assert.ErrorIs(t, err, acpapi.ErrNotFound)
}

func TestEventsPromptRespond(t *testing.T) {
	svc := newService()
	c := serve(t, svc)
	s, err := c.NewSession(context.Background(), acpapi.NewSessionRequest{Agent: "claude", Name: "x"})
	require.NoError(t, err)
	fs := svc.sessions["x"]

	it := s.Events()
	defer func() { _ = it.Close() }()
	assert.NotNil(t, next(t, it).Started)

	require.NoError(t, s.Prompt(context.Background(), []acp.ContentBlock{acp.TextBlock("hi"), acp.ResourceLinkBlock("a.go", "file:///a.go")}))
	ev := next(t, it)
	require.NotNil(t, ev.Prompt)
	assert.Equal(t, "hi", ev.Prompt.Content[0].Text.Text)
	assert.Equal(t, "file:///a.go", ev.Prompt.Content[1].ResourceLink.Uri)
	assert.ErrorIs(t, s.Prompt(context.Background(), []acp.ContentBlock{acp.TextBlock("again")}), acpapi.ErrBusy)

	go func() {
		_, _ = fs.log.Ask(context.Background(), acp.RequestPermissionRequest{SessionId: "id-x"})
	}()
	perm := next(t, it)
	require.NotNil(t, perm.Permission)
	require.NoError(t, s.Respond(context.Background(), perm.Permission.ID, acp.NewRequestPermissionOutcomeSelected("allow")))
	res := next(t, it)
	require.NotNil(t, res.Resolved)
	assert.Equal(t, acp.PermissionOptionId("allow"), res.Resolved.Outcome.Selected.OptionId)
	assert.ErrorIs(t, s.Respond(context.Background(), perm.Permission.ID, acp.NewRequestPermissionOutcomeCancelled()), acpapi.ErrResolved)

	require.NoError(t, s.Cancel(context.Background()))
	ev = next(t, it)
	require.NotNil(t, ev.Done)
	assert.Equal(t, acp.StopReasonCancelled, ev.Done.StopReason)

	require.NoError(t, s.SetConfigOption(context.Background(), acp.SetSessionConfigOptionRequest{
		ValueId: &acp.SetSessionConfigOptionValueId{ConfigId: "model", Value: "large"},
	}))
	fs.mu.Lock()
	assert.Equal(t, acp.SessionConfigValueId("large"), fs.config.ValueId.Value)
	fs.mu.Unlock()

	late := s.Events()
	require.NoError(t, s.Close())
	assert.NoError(t, end(t, it))
	assert.NoError(t, end(t, late))
	assert.ErrorIs(t, s.Prompt(context.Background(), []acp.ContentBlock{acp.TextBlock("closed")}), acpapi.ErrClosed)
}

func TestEventsEndWithCause(t *testing.T) {
	svc := newService()
	c := serve(t, svc)
	s, err := c.NewSession(context.Background(), acpapi.NewSessionRequest{Agent: "claude", Name: "x"})
	require.NoError(t, err)
	svc.sessions["x"].log.Close(errors.New("agent exited: boom"))
	err = end(t, s.Events())
	require.Error(t, err)
	assert.Equal(t, "agent exited: boom", err.Error())
}

func TestEventsNextContextAndClose(t *testing.T) {
	c := serve(t, newService())
	s, err := c.NewSession(context.Background(), acpapi.NewSessionRequest{Agent: "claude", Name: "x"})
	require.NoError(t, err)
	it := s.Events()
	assert.NotNil(t, next(t, it).Started)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, ok := it.Next(ctx)
	assert.False(t, ok)
	assert.ErrorIs(t, it.Err(), context.DeadlineExceeded)

	done := make(chan bool)
	go func() {
		_, ok := it.Next(context.Background())
		done <- ok
	}()
	require.NoError(t, it.Close())
	select {
	case ok := <-done:
		assert.False(t, ok)
	case <-time.After(5 * time.Second):
		t.Fatal("Close did not unblock Next")
	}
	assert.NoError(t, it.Err())
}
