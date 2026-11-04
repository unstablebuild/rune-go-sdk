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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestWM(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		setup   func(*testEnv)
		wantErr string
		wantOut string
		// req is the protojson of the last RPC's request.
		req   string
		check func(*testing.T, *testEnv)
	}{
		// focus
		{
			name:    "focus",
			args:    []string{"wm", "focus"},
			wantOut: "42\n",
		},
		{
			name:    "focus/j",
			args:    []string{"wm", "focus", "-F", "json"},
			wantOut: "{\"WindowID\":42,\"success\":true}\n",
		},
		{
			name:    "focus/tpl",
			args:    []string{"wm", "focus", "-F", "{{.WindowID}}"},
			wantOut: "42\n",
		},

		// close
		{
			name:    "close",
			args:    []string{"wm", "close", "99"},
			wantOut: "OK\n",
			check: func(t *testing.T, env *testEnv) {
				require.Equal(t, uint64(99), env.wm.closedID)
			},
		},
		{
			name:    "close/j",
			args:    []string{"wm", "close", "-F", "json", "99"},
			wantOut: "{\"success\":true}\n",
		},
		{
			name:    "close/err",
			args:    []string{"wm", "close", "abc"},
			wantErr: "invalid window ID",
		},

		// split
		{
			name:    "split",
			args:    []string{"wm", "split", "-o", "right", "42", "file:///tmp/test.txt"},
			wantOut: "200\n",
		},
		{
			name:    "split/path",
			args:    []string{"wm", "split", "-o", "right", "42", "/tmp/test.txt"},
			wantOut: "200\n",
		},
		{
			name:    "split/j",
			args:    []string{"wm", "split", "-F", "json", "42", "file:///tmp/test.txt"},
			wantOut: "{\"WindowID\":200,\"success\":true}\n",
		},
		{
			name:    "split/err_orient",
			args:    []string{"wm", "split", "-o", "bad", "42", "file:///tmp/test.txt"},
			wantErr: "invalid orientation",
		},

		// set-content
		{
			name:    "set-content",
			args:    []string{"wm", "set-content", "42", "file:///tmp/test.txt"},
			wantOut: "OK\n",
		},
		{
			name:    "set-content/path",
			args:    []string{"wm", "set-content", "42", "/tmp/test.txt"},
			wantOut: "OK\n",
		},
		{
			name:    "set-content/j",
			args:    []string{"wm", "set-content", "-F", "json", "42", "file:///tmp/test.txt"},
			wantOut: "{\"success\":true}\n",
		},
		{
			name:    "set-content/err",
			args:    []string{"wm", "set-content", "abc", "file:///tmp/test.txt"},
			wantErr: "invalid window ID",
		},

		// floating
		{
			name:    "floating",
			args:    []string{"wm", "floating", "--title", "Preview", "file:///tmp/test.txt"},
			wantOut: "400\n",
			req: `{"type":"Request","request":{"offset":{},"title":"Preview",` +
				`"uri":"file:///tmp/test.txt"}}`,
		},
		{
			name:    "floating/path",
			args:    []string{"wm", "floating", "/tmp/test.txt"},
			wantOut: "400\n",
			req:     `{"type":"Request","request":{"offset":{},"uri":"file:///tmp/test.txt"}}`,
		},
		{
			name:    "floating/j",
			args:    []string{"wm", "floating", "-F", "json", "file:///tmp/test.txt"},
			wantOut: "{\"WindowID\":400,\"success\":true}\n",
		},
		{
			name: "floating/shown",
			args: []string{"wm", "floating", "file:///tmp/test.txt"},
			setup: func(env *testEnv) {
				env.wm.floatingErr = status.Error(codes.Unknown,
					"new floating window: tab already rendered in Window")
			},
			wantErr: "file:///tmp/test.txt is already shown in a window",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			defer env.cleanup()
			if tt.setup != nil {
				tt.setup(env)
			}

			out, err := env.run(tt.args...)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantOut, out)
			if tt.req != "" {
				rpcs := env.rpcs.take()
				require.NotEmpty(t, rpcs)
				req, err := protojson.Marshal(rpcs[len(rpcs)-1].req)
				require.NoError(t, err)
				require.JSONEq(t, tt.req, string(req))
			}
			if tt.check != nil {
				tt.check(t, env)
			}
		})
	}
}
