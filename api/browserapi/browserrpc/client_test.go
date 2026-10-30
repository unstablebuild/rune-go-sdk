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
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
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
			srv := &promptServer{events: tt.events, err: tt.err}
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
	srv := &promptServer{events: []*PromptEvent{{Event: &PromptEvent_Opened_{
		Opened: &PromptEvent_Opened{WindowId: 1},
	}}}}
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
			srv := &promptServer{}
			c := newTestClient(t, srv)
			_, err := c.Prompt(tt.cfg, nil)
			require.Error(t, err)
			assert.Nil(t, srv.request(), "nothing is sent to the host")
		})
	}
}

// promptServer opens prompts by streaming events, then ending the
// stream with err.
type promptServer struct {
	UnimplementedWindowManagerServer
	events []*PromptEvent
	err    error
	reqs   chan *PromptRequest
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

func newTestClient(t *testing.T, srv *promptServer) *Client {
	t.Helper()
	srv.reqs = make(chan *PromptRequest, 1)
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
