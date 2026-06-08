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
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMisc(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
		wantOut string // use $DATADIR for substitution
		check   func(*testing.T, *testEnv, string)
	}{
		// datadir
		{
			name:    "datadir",
			args:    []string{"datadir"},
			wantOut: "$DATADIR\n",
		},
		{
			name: "datadir/j",
			args: []string{"datadir", "-F", "json"},
			check: func(t *testing.T, env *testEnv, out string) {
				require.Contains(t, out, `"success":true`)
				require.Contains(t, out, env.datadir)
			},
		},
		{
			name:    "datadir/tpl",
			args:    []string{"datadir", "-F", "{{.DataDir}}"},
			wantOut: "$DATADIR\n",
		},

		// uri
		{
			name:    "uri",
			args:    []string{"uri", "/tmp/test.txt"},
			wantOut: "file:///tmp/test.txt\n",
		},
		{
			name:    "uri/j",
			args:    []string{"uri", "-F", "json", "/tmp/test.txt"},
			wantOut: "{\"URI\":\"file:///tmp/test.txt\",\"success\":true}\n",
		},
		{
			name:    "uri/tpl",
			args:    []string{"uri", "-F", "{{.URI}}", "/tmp/test.txt"},
			wantOut: "file:///tmp/test.txt\n",
		},

		// open
		{
			name:    "open",
			args:    []string{"open", "file:///tmp/test.txt"},
			wantOut: "file:///tmp/test.txt\n",
		},
		{
			name:    "open/path",
			args:    []string{"open", "/tmp/test.txt"},
			wantOut: "file:///tmp/test.txt\n",
		},
		{
			name:    "open/j",
			args:    []string{"open", "-F", "json", "file:///tmp/test.txt"},
			wantOut: "{\"Resource\":\"file:///tmp/test.txt\",\"success\":true}\n",
		},

		// notify
		{
			name:    "notify",
			args:    []string{"notify", "info", "hello world"},
			wantOut: "OK\n",
			check: func(t *testing.T, env *testEnv, out string) {
				require.Equal(t, uint32(2), env.notif.lastLevel)
				require.Equal(t, "hello world", env.notif.lastMsg)
			},
		},
		{
			name:    "notify/j",
			args:    []string{"notify", "-F", "json", "info", "hello"},
			wantOut: "{\"success\":true}\n",
		},
		{
			name:    "notify/err",
			args:    []string{"notify", "bad", "msg"},
			wantErr: "invalid level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			defer env.cleanup()

			out, err := env.run(tt.args...)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)

			wantOut := tt.wantOut
			if wantOut != "" {
				wantOut = strings.ReplaceAll(wantOut, "$DATADIR", env.datadir)
				require.Equal(t, wantOut, out)
			}
			if tt.check != nil {
				tt.check(t, env, out)
			}
		})
	}
}

func TestMissingEnvVars(t *testing.T) {
	t.Setenv("RUNE_SOCKET", "")
	t.Setenv("RUNE_DATADIR", "")
	a := &app{}
	_, err := a.getWorkspace()
	require.ErrorIs(t, err, errNotInRune)
}

func TestLooksLikeURI(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{name: "file_uri", arg: "file:///tmp/test.txt", want: true},
		{name: "http_uri", arg: "http://example.com/path", want: true},
		{name: "https_uri", arg: "https://example.com/path", want: true},
		{name: "custom_scheme", arg: "rune://workspace/file.txt", want: true},
		{name: "absolute_path", arg: "/tmp/test.txt", want: false},
		{name: "relative_path", arg: "./test.txt", want: false},
		{name: "relative_no_dot", arg: "test.txt", want: false},
		{name: "windows_path", arg: "C:\\Users\\test.txt", want: false},
		{name: "empty_string", arg: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := looksLikeURI(tt.arg)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestResolveURIArg(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		wantURI string
		wantErr bool
	}{
		{name: "file_uri", arg: "file:///tmp/test.txt", wantURI: "file:///tmp/test.txt"},
		{name: "rune_uri", arg: "rune://workspace/src/main.go", wantURI: "rune://workspace/src/main.go"},
		{name: "uri_query", arg: "file:///tmp/test.txt?line=10", wantURI: "file:///tmp/test.txt?line=10"},
		{name: "invalid_uri", arg: "://invalid", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{}
			got, err := a.resolveURIArg(context.Background(), tt.arg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantURI, got.String())
		})
	}
}
