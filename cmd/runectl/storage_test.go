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
)

func TestStorage(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testEnv)
		args    []string
		wantErr string
		wantOut string
		check   func(*testing.T, *testEnv, string)
	}{
		// create
		{
			name:    "create",
			args:    []string{"storage", "create", "doc1", `{"name":"test","value":42}`},
			wantOut: "OK\n",
		},
		{
			name:    "create/j",
			args:    []string{"storage", "create", "-F", "json", "doc1", `{"name":"test"}`},
			wantOut: "{\"success\":true}\n",
		},
		{
			name:    "create/err",
			args:    []string{"storage", "create", "doc", "not-json"},
			wantErr: "invalid JSON",
		},

		// set
		{
			name:    "set",
			args:    []string{"storage", "set", "doc2", `{"key":"val"}`},
			wantOut: "OK\n",
		},
		{
			name:    "set/j",
			args:    []string{"storage", "set", "-F", "json", "doc2", `{"key":"val"}`},
			wantOut: "{\"success\":true}\n",
		},
		{
			name:    "set/err",
			args:    []string{"storage", "set", "doc", "not-json"},
			wantErr: "invalid JSON",
		},

		// delete
		{
			name:    "delete",
			args:    []string{"storage", "delete", "doc1"},
			wantOut: "OK\n",
		},
		{
			name:    "delete/j",
			args:    []string{"storage", "delete", "-F", "json", "doc1"},
			wantOut: "{\"success\":true}\n",
		},

		// get
		{
			name: "get",
			setup: func(env *testEnv) {
				_, _ = env.run("storage", "create", "doc3", `{"a":"b"}`)
			},
			args: []string{"storage", "get", "doc3"},
			check: func(t *testing.T, env *testEnv, out string) {
				require.Contains(t, out, `"a"`)
				require.Contains(t, out, `"b"`)
			},
		},
		{
			name:    "get/err_table",
			args:    []string{"storage", "get", "-F", "table", "doc"},
			wantErr: "table format not supported",
		},

		// update
		{
			name: "update",
			setup: func(env *testEnv) {
				_, _ = env.run("storage", "create", "doc4", `{"name":"before"}`)
			},
			args:    []string{"storage", "update", "doc4", "name", "after"},
			wantOut: "OK\n",
		},
		{
			name: "update/j",
			setup: func(env *testEnv) {
				_, _ = env.run("storage", "create", "doc5", `{"name":"before"}`)
			},
			args:    []string{"storage", "update", "-F", "json", "doc5", "name", "after"},
			wantOut: "{\"success\":true}\n",
		},
		{
			name: "update/nested",
			setup: func(env *testEnv) {
				_, _ = env.run("storage", "create", "nested-doc", `{"field":{"nested_field":"old"}}`)
			},
			args:    []string{"storage", "update", "nested-doc", "field.nested_field", "value"},
			wantOut: "OK\n",
			check: func(t *testing.T, env *testEnv, out string) {
				require.NotNil(t, env.docStore.lastUpdate)
				updates := env.docStore.lastUpdate.GetUpdates()
				require.Len(t, updates, 1)
				require.Equal(t, []string{"field", "nested_field"}, updates[0].GetFieldPath())
			},
		},
		{
			name: "update/err_path",
			setup: func(env *testEnv) {
				_, _ = env.run("storage", "create", "doc-inv", `{"a":"b"}`)
			},
			args:    []string{"storage", "update", "doc-inv", ".bad", "value"},
			wantErr: "empty segment",
		},

		// list
		{
			name: "list",
			setup: func(env *testEnv) {
				_, _ = env.run("storage", "create", "list-doc", `{"x":1}`)
			},
			args: []string{"storage", "list"},
			check: func(t *testing.T, env *testEnv, out string) {
				require.NotEmpty(t, out)
			},
		},
		{
			name:    "list/err_table",
			args:    []string{"storage", "list", "-F", "table"},
			wantErr: "table format not supported",
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
			if tt.wantOut != "" {
				require.Equal(t, tt.wantOut, out)
			}
			if tt.check != nil {
				tt.check(t, env, out)
			}
		})
	}
}
