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
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSignal(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
		wantOut string
		check   func(*testing.T, *testEnv)
	}{
		{
			name:    "sigterm",
			args:    []string{"signal", "12345", "SIGTERM"},
			wantOut: "OK\n",
			check: func(t *testing.T, env *testEnv) {
				require.True(t, env.executor.signalCalled)
				require.Equal(t, int64(12345), env.executor.lastPid)
				require.Equal(t, int32(15), env.executor.lastSignal)
			},
		},
		{
			name:    "number",
			args:    []string{"signal", "9999", "9"},
			wantOut: "OK\n",
			check: func(t *testing.T, env *testEnv) {
				require.Equal(t, int64(9999), env.executor.lastPid)
				require.Equal(t, int32(9), env.executor.lastSignal)
			},
		},
		{
			name:    "json",
			args:    []string{"signal", "-F", "json", "12345", "TERM"},
			wantOut: "{\"success\":true}\n",
		},
		{
			name:    "err/pid",
			args:    []string{"signal", "abc", "SIGTERM"},
			wantErr: "invalid pid",
		},
		{
			name:    "err/signal",
			args:    []string{"signal", "12345", "INVALID"},
			wantErr: "unknown signal",
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
			require.Equal(t, tt.wantOut, out)
			if tt.check != nil {
				tt.check(t, env)
			}
		})
	}
}

func TestParseSignal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    syscall.Signal
		wantErr bool
	}{
		{name: "numeric", input: "15", want: syscall.Signal(15)},
		{name: "sigterm_full", input: "SIGTERM", want: syscall.SIGTERM},
		{name: "term_short", input: "TERM", want: syscall.SIGTERM},
		{name: "lowercase", input: "term", want: syscall.SIGTERM},
		{name: "sigkill", input: "SIGKILL", want: syscall.SIGKILL},
		{name: "kill", input: "KILL", want: syscall.SIGKILL},
		{name: "sigint", input: "SIGINT", want: syscall.SIGINT},
		{name: "int", input: "INT", want: syscall.SIGINT},
		{name: "sighup", input: "SIGHUP", want: syscall.SIGHUP},
		{name: "sigusr1", input: "SIGUSR1", want: syscall.SIGUSR1},
		{name: "sigusr2", input: "USR2", want: syscall.SIGUSR2},
		{name: "sigstop", input: "STOP", want: syscall.SIGSTOP},
		{name: "sigcont", input: "CONT", want: syscall.SIGCONT},
		{name: "invalid", input: "INVALID", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSignal(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
