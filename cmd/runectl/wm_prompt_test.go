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

package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/browserapi/browserrpc"
	"github.com/unstablebuild/rune-go-sdk/term"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestWMPrompt(t *testing.T) {
	opened := &browserrpc.PromptEvent{Event: &browserrpc.PromptEvent_Opened_{
		Opened: &browserrpc.PromptEvent_Opened{WindowId: 7},
	}}
	selected := &browserrpc.PromptEvent{Event: &browserrpc.PromptEvent_Selected_{
		Selected: &browserrpc.PromptEvent_Selected{Index: 1, Option: "No"},
	}}
	closed := &browserrpc.PromptEvent{Event: &browserrpc.PromptEvent_Closed_{
		Closed: &browserrpc.PromptEvent_Closed{},
	}}
	answered := []*browserrpc.PromptEvent{opened, selected, closed}

	tests := []struct {
		name      string
		args      []string
		events    []*browserrpc.PromptEvent
		streamErr error
		wantErr   string
		wantOut   string
		// want is the request the host receives.
		want *browserrpc.PromptRequest
		// wantKeys are the bindings of the request the host receives.
		wantKeys []term.KeyComb
	}{
		{
			name:    "answered",
			args:    []string{"wm", "prompt", "Save?", "Yes", "No"},
			events:  answered,
			wantOut: "No\n",
			want:    &browserrpc.PromptRequest{Message: "Save?", Options: []string{"Yes", "No"}},
		},
		{
			name:    "answered/j",
			args:    []string{"wm", "prompt", "-F", "json", "Save?", "Yes", "No"},
			events:  answered,
			wantOut: "{\"Index\":1,\"Option\":\"No\",\"success\":true}\n",
		},
		{
			name:    "answered/tpl",
			args:    []string{"wm", "prompt", "-F", "{{.Index}}", "Save?", "Yes", "No"},
			events:  answered,
			wantOut: "1\n",
		},
		{
			name:    "keys and width",
			args:    []string{"wm", "prompt", "--keys", "y<c-n>", "--max-width", "40", "Save?", "Yes", "No"},
			events:  answered,
			wantOut: "No\n",
			want: &browserrpc.PromptRequest{
				Message: "Save?", Options: []string{"Yes", "No"}, MaxWidth: 40,
			},
			wantKeys: []term.KeyComb{{Ch: 'y'}, {Mod: term.ModCtrl, Ch: 'n'}},
		},
		{
			name:    "dismissed",
			args:    []string{"wm", "prompt", "Save?", "Yes", "No"},
			events:  []*browserrpc.PromptEvent{opened, closed},
			wantErr: "prompt dismissed",
		},
		{
			name:    "dismissed/j",
			args:    []string{"wm", "prompt", "-F", "json", "Save?", "Yes", "No"},
			events:  []*browserrpc.PromptEvent{opened, closed},
			wantOut: "{\"error\":\"prompt dismissed\",\"success\":false}\n",
		},
		{
			name:    "already open",
			args:    []string{"wm", "prompt", "Save?", "Yes", "No"},
			events:  []*browserrpc.PromptEvent{opened},
			wantErr: "a prompt with this message is already open",
		},
		{
			name:      "connection breaks",
			args:      []string{"wm", "prompt", "Save?", "Yes", "No"},
			events:    []*browserrpc.PromptEvent{opened},
			streamErr: status.Error(codes.Unavailable, "gone"),
			wantErr:   "gone",
		},
		{
			name:      "host refuses",
			args:      []string{"wm", "prompt", "Save?", "Yes", "No"},
			streamErr: status.Error(codes.InvalidArgument, "bad prompt"),
			wantErr:   "bad prompt",
		},
		{
			name:    "few keys",
			args:    []string{"wm", "prompt", "--keys", "y", "Save?", "Yes", "No"},
			wantErr: "different number of bindings and options",
		},
		{
			name:    "bad keys",
			args:    []string{"wm", "prompt", "--keys", "<bogus>", "Save?", "Yes", "No"},
			wantErr: "invalid --keys",
		},
		{
			name:    "no options",
			args:    []string{"wm", "prompt", "Save?"},
			wantErr: "requires at least 2 arg(s)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			defer env.cleanup()
			env.wm.promptEvents = tt.events
			env.wm.promptErr = tt.streamErr

			out, err := env.run(tt.args...)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantOut, out)
			if tt.want == nil {
				return
			}
			rpcs := env.rpcs.take()
			require.Len(t, rpcs, 1)
			require.Equal(t, "WindowManager/Prompt", rpcs[0].method)
			req, ok := rpcs[0].req.(*browserrpc.PromptRequest)
			require.True(t, ok)
			require.Equal(t, tt.want.GetMessage(), req.GetMessage())
			require.Equal(t, tt.want.GetOptions(), req.GetOptions())
			require.Equal(t, tt.want.GetMaxWidth(), req.GetMaxWidth())
			var keys []term.KeyComb
			for _, b := range req.GetBindings() {
				ev, err := b.ToModel()
				require.NoError(t, err)
				require.Equal(t, term.EventKey, ev.Type)
				keys = append(keys, ev.KeyComb())
			}
			require.Equal(t, tt.wantKeys, keys)
		})
	}
}
