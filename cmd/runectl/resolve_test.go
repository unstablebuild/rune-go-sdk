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

func TestResolve(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
		wantOut string
		wantDir string
	}{
		{
			name:    "found",
			args:    []string{"resolve", "go"},
			wantOut: "/usr/local/go/bin/go\n",
		},
		{
			name:    "dir",
			args:    []string{"resolve", "--dir", "/tmp", "./run.sh"},
			wantOut: "/tmp/run.sh\n",
			wantDir: "/tmp",
		},
		{
			name:    "json",
			args:    []string{"resolve", "-F", "json", "go"},
			wantOut: "{\"Path\":\"/usr/local/go/bin/go\",\"success\":true}\n",
		},
		{
			name:    "err/not found",
			args:    []string{"resolve", "nope"},
			wantErr: "file does not exist",
		},
		{
			name:    "err/args",
			args:    []string{"resolve"},
			wantErr: "accepts 1 arg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			defer env.cleanup()
			env.executor.paths = map[string]string{
				"go":       "/usr/local/go/bin/go",
				"./run.sh": "/tmp/run.sh",
			}

			out, err := env.run(tt.args...)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantOut, out)
			require.Equal(t, tt.wantDir, env.executor.lastLookPath.GetDir())
		})
	}
}
