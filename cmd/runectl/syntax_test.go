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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSyntax(t *testing.T) {
	tests := []struct {
		name    string
		args    []string // $FILE will be substituted with test file path
		wantErr string
		check   func(*testing.T, *testEnv, string)
	}{
		// search
		{
			name: "search",
			args: []string{"syntax", "search", "(function_declaration) @fn"},
			check: func(t *testing.T, env *testEnv, out string) {
				require.Contains(t, out, "file:///src/main.go")
				require.Contains(t, out, "func main()")
			},
		},
		{
			name: "search/capture",
			args: []string{"syntax", "search", "-c", "fn", "(function_declaration) @fn"},
			check: func(t *testing.T, env *testEnv, out string) {
				require.Equal(t, "fn", env.syntax.lastCaptures[0])
			},
		},
		{
			name: "search/tpl",
			args: []string{"syntax", "search", "-F", "{{.File}} {{.Text}}", "(function_declaration) @fn"},
			check: func(t *testing.T, env *testEnv, out string) {
				require.Equal(t, "file:///src/main.go func main()\n", out)
			},
		},

		// searchnode
		{
			name: "searchnode",
			args: []string{"syntax", "searchnode", "func"},
			check: func(t *testing.T, env *testEnv, out string) {
				require.Contains(t, out, "MyFunc")
				require.Equal(t, uint32(8), env.syntax.lastNodeTypes)
			},
		},
		{
			name: "searchnode/combined",
			args: []string{"syntax", "searchnode", "scope|func"},
			check: func(t *testing.T, env *testEnv, out string) {
				require.Equal(t, uint32(9), env.syntax.lastNodeTypes)
			},
		},
		{
			name:    "searchnode/err",
			args:    []string{"syntax", "searchnode", "invalid"},
			wantErr: "invalid node type",
		},
		{
			name:    "searchnode/err_cmb",
			args:    []string{"syntax", "searchnode", "func|invalid"},
			wantErr: "invalid node type",
		},

		// query
		{
			name: "query",
			args: []string{"syntax", "query", "$FILE", "(package_clause) @pkg"},
			check: func(t *testing.T, env *testEnv, out string) {
				require.Contains(t, out, "package main")
				require.Contains(t, env.syntax.lastQueryURI, "test.go")
				require.Equal(t, "(package_clause) @pkg", env.syntax.lastQuery)
			},
		},
		{
			name: "query/uri",
			args: []string{"syntax", "query", "$URI", "(package_clause) @pkg"},
			check: func(t *testing.T, env *testEnv, out string) {
				require.Contains(t, env.syntax.lastQueryURI, "test.go")
			},
		},
		{
			name: "query/capture",
			args: []string{"syntax", "query", "-c", "pkg", "$FILE", "(package_clause) @pkg"},
			check: func(t *testing.T, env *testEnv, out string) {
				require.Equal(t, "pkg", env.syntax.lastCaptures[0])
			},
		},

		// querynode
		{
			name: "querynode",
			args: []string{"syntax", "querynode", "$FILE", "namespace"},
			check: func(t *testing.T, env *testEnv, out string) {
				require.Contains(t, out, "main")
				require.Contains(t, env.syntax.lastQueryURI, "test.go")
				require.Equal(t, uint32(2), env.syntax.lastNodeTypes)
			},
		},
		{
			name: "querynode/uri",
			args: []string{"syntax", "querynode", "$URI", "namespace"},
			check: func(t *testing.T, env *testEnv, out string) {
				require.Contains(t, env.syntax.lastQueryURI, "test.go")
			},
		},
		{
			name: "querynode/combined",
			args: []string{"syntax", "querynode", "$FILE", "namespace|var|type"},
			check: func(t *testing.T, env *testEnv, out string) {
				require.Equal(t, uint32(82), env.syntax.lastNodeTypes)
			},
		},
		{
			name:    "querynode/err",
			args:    []string{"syntax", "querynode", "$FILE", "invalid"},
			wantErr: "invalid node type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			defer env.cleanup()

			testFile := filepath.Join(env.datadir, "test.go")
			testURI := "file://" + testFile

			// Substitute placeholders
			args := make([]string, len(tt.args))
			for i, arg := range tt.args {
				switch arg {
				case "$FILE":
					args[i] = testFile
				case "$URI":
					args[i] = testURI
				default:
					args[i] = arg
				}
			}

			out, err := env.run(args...)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, env, out)
			}
		})
	}
}
