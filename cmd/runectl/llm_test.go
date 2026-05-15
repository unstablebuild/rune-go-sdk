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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/llmapi"
)

func TestParseProviderModel(t *testing.T) {
	tests := []struct {
		in            string
		wantProvider  string
		wantModel     string
		wantErrSubstr string
	}{
		{"openai/gpt-5", "openai", "gpt-5", ""},
		{"anthropic/claude-3", "anthropic", "claude-3", ""},
		{"llamacpp/gemma-3-4b-it-Q4", "llamacpp", "gemma-3-4b-it-Q4", ""},
		{"missingSlash", "", "", "expected provider/model"},
		{"/leading", "", "", "expected provider/model"},
		{"trailing/", "", "", "expected provider/model"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			p, m, err := parseProviderModel(tc.in)
			if tc.wantErrSubstr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrSubstr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantProvider, p)
			assert.Equal(t, tc.wantModel, m)
		})
	}
}

func TestLoadMessageBody_FileVsLiteral(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "msg.txt")
	require.NoError(t, os.WriteFile(path, []byte("hello from file"), 0o644))

	body, err := loadMessageBody(path)
	require.NoError(t, err)
	assert.Equal(t, "hello from file", body)

	// Non-file argument is treated as a literal.
	body, err = loadMessageBody("just a literal string")
	require.NoError(t, err)
	assert.Equal(t, "just a literal string", body)
}

func TestParseResponseFormat(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		rf, err := parseResponseFormat("")
		require.NoError(t, err)
		assert.Nil(t, rf)
	})
	t.Run("text", func(t *testing.T) {
		rf, err := parseResponseFormat("text")
		require.NoError(t, err)
		require.NotNil(t, rf)
		assert.Equal(t, llmapi.ResponseFormatTypeText, rf.Type)
	})
	t.Run("json", func(t *testing.T) {
		rf, err := parseResponseFormat("json")
		require.NoError(t, err)
		require.NotNil(t, rf)
		assert.Equal(t, llmapi.ResponseFormatTypeJSONObject, rf.Type)
	})
	t.Run("json schema with nested schema", func(t *testing.T) {
		raw := `{"name":"city","schema":{"type":"object","properties":{"name":{"type":"string"}}},"strict":true}`
		rf, err := parseResponseFormat(raw)
		require.NoError(t, err)
		require.NotNil(t, rf)
		assert.Equal(t, llmapi.ResponseFormatTypeJSONSchema, rf.Type)
		require.NotNil(t, rf.JSONSchema)
		assert.Equal(t, "city", rf.JSONSchema.Name)
		assert.True(t, rf.JSONSchema.Strict)
		b, err := rf.JSONSchema.Schema.MarshalJSON()
		require.NoError(t, err)
		assert.Contains(t, string(b), `"type":"object"`)
	})
	t.Run("invalid json", func(t *testing.T) {
		_, err := parseResponseFormat("{not json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--response-format")
	})
}
