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

package browserrpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
	"github.com/unstablebuild/rune-go-sdk/handler"
	"github.com/unstablebuild/rune-go-sdk/handler/handlerrpc"
	"github.com/unstablebuild/rune-go-sdk/term"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestClientPrompt(t *testing.T) {
	opened := &PromptEvent{Event: &PromptEvent_Opened_{
		Opened: &PromptEvent_Opened{WindowId: 7},
	}}
	selected := &PromptEvent{Event: &PromptEvent_Selected_{
		Selected: &PromptEvent_Selected{Index: 1, Option: "No"},
	}}
	closed := &PromptEvent{Event: &PromptEvent_Closed_{Closed: &PromptEvent_Closed{}}}
	broken := status.Error(codes.Unavailable, "gone")

	tests := []struct {
		name    string
		events  []*PromptEvent
		err     error
		wantErr bool
		want    []string
	}{
		{
			name:   "selected then closed",
			events: []*PromptEvent{opened, selected, closed},
			want:   []string{"select 1 No", "close"},
		},
		{
			name:   "dismissed",
			events: []*PromptEvent{opened, closed},
			want:   []string{"close"},
		},
		{
			name:   "already open for another caller",
			events: []*PromptEvent{opened},
		},
		{
			name:   "connection breaks while open",
			events: []*PromptEvent{opened},
			err:    broken,
			want:   []string{"close"},
		},
		{
			name:    "fails before it opens",
			err:     status.Error(codes.InvalidArgument, "no"),
			wantErr: true,
		},
		{
			name:    "closed before it opens",
			events:  []*PromptEvent{closed},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newPromptServer(tt.events, tt.err)
			c := newTestClient(t, srv)
			h := newRecordingPromptHandler()

			win, err := c.Prompt(browserapi.PromptConfig{
				Message: "Sure?", Options: []string{"Yes", "No"},
			}, h)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, win)
				assert.Empty(t, h.wait(t, 0))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, uint64(7), win.WindowID())
			assert.Equal(t, tt.want, h.wait(t, len(tt.want)))
		})
	}
}

func TestClientPromptRequest(t *testing.T) {
	srv := newPromptServer([]*PromptEvent{{Event: &PromptEvent_Opened_{
		Opened: &PromptEvent_Opened{WindowId: 1},
	}}}, nil)
	c := newTestClient(t, srv)
	_, err := c.Prompt(browserapi.PromptConfig{
		Message:  "Sure?",
		Options:  []string{"Yes", "No"},
		Bindings: []term.KeyComb{{Ch: 'y'}, {Mod: term.ModCtrl, Ch: 'n'}},
		MaxWidth: 80,
	}, nil)
	require.NoError(t, err)

	req := srv.request()
	assert.Equal(t, "Sure?", req.GetMessage())
	assert.Equal(t, []string{"Yes", "No"}, req.GetOptions())
	assert.Equal(t, uint32(80), req.GetMaxWidth())
	var keys []term.KeyComb
	for _, b := range req.GetBindings() {
		ev, err := b.ToModel()
		require.NoError(t, err)
		assert.Equal(t, term.EventKey, ev.Type)
		keys = append(keys, ev.KeyComb())
	}
	assert.Equal(t, []term.KeyComb{{Ch: 'y'}, {Mod: term.ModCtrl, Ch: 'n'}}, keys)
}

func TestClientPromptFields(t *testing.T) {
	opened := &PromptEvent{Event: &PromptEvent_Opened_{
		Opened: &PromptEvent_Opened{WindowId: 7, SupportsFields: true},
	}}
	// openedOld is how a host that predates fields opens a prompt,
	// without them.
	openedOld := &PromptEvent{Event: &PromptEvent_Opened_{
		Opened: &PromptEvent_Opened{WindowId: 7},
	}}
	answered := &PromptEvent{Event: &PromptEvent_Selected_{
		Selected: &PromptEvent_Selected{Index: 1, Option: "No", Answers: []string{"me", "pw"}},
	}}
	selected := &PromptEvent{Event: &PromptEvent_Selected_{
		Selected: &PromptEvent_Selected{Index: 1, Option: "No"},
	}}
	closed := &PromptEvent{Event: &PromptEvent_Closed_{Closed: &PromptEvent_Closed{}}}
	fields := []browserapi.Field{
		{Question: "User", Answer: "me", Placeholder: "you@example.com"},
		{Question: "Password", Secret: true, Mandatory: true},
	}
	wantFields := []*PromptField{
		{Question: "User", Answer: "me", Placeholder: "you@example.com"},
		{Question: "Password", Secret: true, Mandatory: true},
	}

	tests := []struct {
		name    string
		fields  []browserapi.Field
		events  []*PromptEvent
		answers bool
		// hold keeps the stream open after the events, until the
		// client ends it.
		hold        bool
		wantFields  []*PromptField
		wantErr     error
		wantHandler []string
	}{
		{
			name:        "without fields",
			events:      []*PromptEvent{opened, selected, closed},
			wantHandler: []string{"select 1 No", "close"},
		},
		{
			name:        "with fields",
			fields:      fields,
			events:      []*PromptEvent{opened, answered, closed},
			answers:     true,
			wantFields:  wantFields,
			wantHandler: []string{"answer 1 No [me pw]", "close"},
		},
		{
			name:        "a PromptHandler gets OnSelect without the answers",
			fields:      fields,
			events:      []*PromptEvent{opened, answered, closed},
			wantFields:  wantFields,
			wantHandler: []string{"select 1 No", "close"},
		},
		{
			name:    "a host that predates fields is unsupported and the prompt closed",
			fields:  fields,
			events:  []*PromptEvent{openedOld},
			hold:    true,
			answers: true,
			wantErr: errors.ErrUnsupported,
		},
		{
			name:        "a host that predates fields still opens prompts without",
			events:      []*PromptEvent{openedOld, selected, closed},
			wantHandler: []string{"select 1 No", "close"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newPromptServer(tt.events, nil)
			srv.hold = tt.hold
			c := newTestClient(t, srv)
			rec := newRecordingPromptHandler()
			var h handler.PromptHandler = rec
			if tt.answers {
				h = recordingAnswerHandler{rec}
			}

			win, err := c.Prompt(browserapi.PromptConfig{
				Message: "Sure?", Options: []string{"Yes", "No"}, Fields: tt.fields,
			}, h)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, win)
				select {
				case <-srv.ended:
				case <-time.After(5 * time.Second):
					require.FailNow(t, "the client kept the prompt open")
				}
				assert.Empty(t, rec.wait(t, 0))
				return
			}
			require.NoError(t, err)
			req := srv.request()
			require.Len(t, req.GetFields(), len(tt.wantFields))
			for i, want := range tt.wantFields {
				got := req.GetFields()[i]
				assert.Equal(t, want.GetQuestion(), got.GetQuestion())
				assert.Equal(t, want.GetAnswer(), got.GetAnswer())
				assert.Equal(t, want.GetSecret(), got.GetSecret())
				assert.Equal(t, want.GetMandatory(), got.GetMandatory())
				assert.Equal(t, want.GetPlaceholder(), got.GetPlaceholder())
			}
			assert.Equal(t, tt.wantHandler, rec.wait(t, len(tt.wantHandler)))
		})
	}
}

func TestClientPromptValidates(t *testing.T) {
	tests := []struct {
		name string
		cfg  browserapi.PromptConfig
	}{
		{name: "no message", cfg: browserapi.PromptConfig{Options: []string{"Ok"}}},
		{name: "no options", cfg: browserapi.PromptConfig{Message: "?"}},
		{
			name: "bindings for some options",
			cfg: browserapi.PromptConfig{
				Message: "?", Options: []string{"Yes", "No"},
				Bindings: []term.KeyComb{{Ch: 'y'}},
			},
		},
		{
			name: "negative max width",
			cfg:  browserapi.PromptConfig{Message: "?", Options: []string{"Ok"}, MaxWidth: -1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newPromptServer(nil, nil)
			c := newTestClient(t, srv)
			_, err := c.Prompt(tt.cfg, nil)
			require.Error(t, err)
			assert.Nil(t, srv.request(), "nothing is sent to the host")
		})
	}
}

func TestClientFloatingResource(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{name: "opened"},
		{
			name:    "already on screen",
			err:     status.Error(codes.Unknown, browserapi.ErrTabNotFree.Error()),
			wantErr: browserapi.ErrTabNotFree,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &floatingServer{err: tt.err, reqs: make(chan *FloatingWindowRequest, 1)}
			c := newTestClient(t, srv)

			win, err := c.Floating(Token{URI: "file:///tmp/a.go"}, browserapi.FloatingConfig{
				Title: "a.go",
			})
			req := srv.request()
			assert.Equal(t, "file:///tmp/a.go", req.GetUri())
			assert.Equal(t, "a.go", req.GetTitle())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, win)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, uint64(9), win.WindowID())
		})
	}
}

// floatingServer opens floating window 9 unless err is set, and asks
// nothing of its content.
type floatingServer struct {
	UnimplementedWindowManagerServer
	err  error
	reqs chan *FloatingWindowRequest
}

func (s *floatingServer) Floating(
	srv grpc.BidiStreamingServer[FloatingWindowMessage, handlerrpc.ServerMessage],
) error {
	msg, err := srv.Recv()
	if err != nil {
		return err
	}
	s.reqs <- msg.GetRequest()
	if s.err != nil {
		return s.err
	}
	if err := srv.Send(&handlerrpc.ServerMessage{
		Type:     handlerrpc.MessageType_Response,
		Response: &handlerrpc.InstallResourceResponse{WindowId: 9},
	}); err != nil {
		return err
	}
	<-srv.Context().Done()
	return nil
}

func (s *floatingServer) request() *FloatingWindowRequest {
	select {
	case req := <-s.reqs:
		return req
	default:
		return nil
	}
}

// promptServer opens prompts by streaming events, then ending the
// stream with err. When hold, it keeps the stream open instead, until
// the client ends it, which closes ended.
type promptServer struct {
	UnimplementedWindowManagerServer
	events []*PromptEvent
	err    error
	hold   bool
	reqs   chan *PromptRequest
	ended  chan struct{}
}

func newPromptServer(events []*PromptEvent, err error) *promptServer {
	return &promptServer{
		events: events, err: err,
		reqs:  make(chan *PromptRequest, 1),
		ended: make(chan struct{}),
	}
}

func (s *promptServer) Prompt(
	req *PromptRequest, srv grpc.ServerStreamingServer[PromptEvent],
) error {
	s.reqs <- req
	for _, ev := range s.events {
		if err := srv.Send(ev); err != nil {
			return err
		}
	}
	if s.hold {
		<-srv.Context().Done()
		close(s.ended)
		return srv.Context().Err()
	}
	return s.err
}

func (s *promptServer) request() *PromptRequest {
	select {
	case req := <-s.reqs:
		return req
	default:
		return nil
	}
}

func newTestClient(t *testing.T, srv WindowManagerServer) *Client {
	t.Helper()
	lis := bufconn.Listen(1 << 16)
	gs := grpc.NewServer()
	RegisterWindowManagerServer(gs, srv)
	go func() { _ = gs.Serve(lis) }()
	t.Cleanup(gs.Stop)

	cc, err := grpc.NewClient("passthrough:///bufconn",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	c := NewClient(context.Background(), cc)
	t.Cleanup(func() { _ = c.Close() })
	return c
}

// recordingPromptHandler records the calls a prompt makes to it.
type recordingPromptHandler struct {
	calls chan string
}

func newRecordingPromptHandler() *recordingPromptHandler {
	return &recordingPromptHandler{calls: make(chan string, 8)}
}

func (h *recordingPromptHandler) OnSelect(idx int, option string) {
	h.calls <- fmt.Sprintf("select %d %s", idx, option)
}

func (h *recordingPromptHandler) OnClose() error {
	h.calls <- "close"
	return nil
}

// recordingAnswerHandler records the answers too.
type recordingAnswerHandler struct{ *recordingPromptHandler }

func (h recordingAnswerHandler) OnAnswer(idx int, option string, answers []string) {
	h.calls <- fmt.Sprintf("answer %d %s %v", idx, option, answers)
}

// wait returns the n calls h is expected to receive, and any more that
// arrive shortly after.
func (h *recordingPromptHandler) wait(t *testing.T, n int) []string {
	t.Helper()
	var got []string
	for range n {
		select {
		case c := <-h.calls:
			got = append(got, c)
		case <-time.After(5 * time.Second):
			t.Fatalf("got %v, want %d calls", got, n)
		}
	}
	select {
	case c := <-h.calls:
		got = append(got, c)
	case <-time.After(50 * time.Millisecond):
	}
	return got
}
