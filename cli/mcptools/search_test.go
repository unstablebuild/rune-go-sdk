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

package mcptools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/searchapi"
)

func TestRegexpSearchLimit(t *testing.T) {
	tests := []struct {
		name     string
		limit    any
		wantJSON string
		wantHint string
	}{
		{
			name: "lines left out", limit: 1,
			wantJSON: `[{"path":"src/main.go","line":6,"text":"func MyFunc() {"}]`,
			wantHint: "Showing the first 1 matching lines; narrow pattern, roots or " +
				"include, or raise limit, to see the rest.",
		},
		{
			name: "every line", limit: 2,
			wantJSON: `[{"path":"src/main.go","line":6,"text":"func MyFunc() {"},` +
				`{"path":"src/types.go","line":31,"text":"func (s MyStruct) MyMethod() {"}]`,
		},
		{
			name: "no limit", limit: 0,
			wantJSON: `[{"path":"src/main.go","line":6,"text":"func MyFunc() {"},` +
				`{"path":"src/types.go","line":31,"text":"func (s MyStruct) MyMethod() {"}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer()
			res := callTool(t, s.MCPServer, "regexp_search",
				map[string]any{"pattern": "func", "limit": tt.limit})
			require.False(t, res.IsError, resultText(t, res))
			require.JSONEq(t, tt.wantJSON, resultText(t, res))
			require.Equal(t, tt.wantHint, resultHint(t, res))
		})
	}
}

func TestSemanticSearchStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   searchapi.SemanticStatus
		wantHint string
	}{
		{name: "complete"},
		{
			name:   "indexing",
			status: searchapi.SemanticStatus{Indexing: true, Message: "12 files left to index"},
			wantHint: "Note: 12 files left to index. Files not indexed yet may be missing; " +
				"regexp_search finds exact text in every file.",
		},
		{
			name:     "words only",
			status:   searchapi.SemanticStatus{WordsOnly: true, Message: "matched by words only"},
			wantHint: "Note: matched by words only.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer()
			s.searcher.status = tt.status
			res := callTool(t, s.MCPServer, "semantic_search", map[string]any{"query": "q"})
			require.False(t, res.IsError, resultText(t, res))
			require.Equal(t, tt.wantHint, resultHint(t, res))
		})
	}
}

func TestSearchUnsupported(t *testing.T) {
	unsupported := fmt.Errorf("no index: %w", searchapi.ErrUnsupported)
	tests := []struct {
		tool    string
		method  string
		wantErr string
	}{
		{"regexp_search", "RegexpSearch", "no index: search: not supported"},
		{"semantic_search", "SemanticSearch",
			"no index: search: not supported; use regexp_search instead"},
	}
	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			s := newTestServer()
			s.searcher.failOn(tt.method, unsupported)
			res := callTool(t, s.MCPServer, tt.tool, map[string]any{"pattern": "x", "query": "x"})
			require.True(t, res.IsError)
			require.Equal(t, tt.wantErr, resultText(t, res))
		})
	}
}

func TestSearchArguments(t *testing.T) {
	tests := []struct {
		tool    string
		args    map[string]any
		wantErr string
	}{
		{"semantic_search", map[string]any{"query": " \t\n"}, `"query" must not be blank`},
		{"regexp_search", map[string]any{"pattern": "x", "limit": math.MaxInt32}, `"limit"`},
		{"regexp_search", map[string]any{"pattern": "x", "max_per_file": int64(math.MaxInt32) + 1}, `"max_per_file"`},
		{"semantic_search", map[string]any{"query": "x", "limit": int64(math.MaxInt32) + 1}, `"limit"`},
		{"regexp_search", map[string]any{"pattern": "x", "limit": -1},
			`"limit" must be a whole number`},
		{"regexp_search", map[string]any{"pattern": "x", "max_per_file": 1.5},
			`"max_per_file" must be a whole number`},
		{"semantic_search", map[string]any{"query": "x", "limit": -2},
			`"limit" must be a whole number`},
	}
	for _, tt := range tests {
		t.Run(tt.tool+"/"+tt.wantErr, func(t *testing.T) {
			s := newTestServer()
			res := callTool(t, s.MCPServer, tt.tool, tt.args)
			require.True(t, res.IsError)
			require.Contains(t, resultText(t, res), tt.wantErr)
			require.Empty(t, s.searcher.methods())
		})
	}
}

func TestSearchNilSearcher(t *testing.T) {
	require.Panics(t, func() { RegisterSearchTools(context.Background(), server.NewMCPServer("test", "1"), nil) })
}

func requireSemanticStructure(t *testing.T, res mcp.CallToolResult, results string) {
	t.Helper()
	data, err := json.Marshal(res.StructuredContent)
	require.NoError(t, err)
	require.JSONEq(t, `{"results":`+results+`,"status":{"indexing":false,"words_only":false,"message":""}}`, string(data))
}

func TestSearchWireLimitBoundaries(t *testing.T) {
	for _, tc := range []struct {
		tool string
		args map[string]any
	}{
		{"regexp_search", map[string]any{"pattern": "x", "limit": math.MaxInt32 - 1, "max_per_file": math.MaxInt32}},
		{"semantic_search", map[string]any{"query": "x", "limit": math.MaxInt32}},
	} {
		s := newTestServer()
		res := callTool(t, s.MCPServer, tc.tool, tc.args)
		require.False(t, res.IsError, resultText(t, res))
	}
}

func TestSemanticSearchStructuredStatus(t *testing.T) {
	for _, empty := range []bool{false, true} {
		s := newTestServer()
		if empty {
			s.searcher.semantic = nil
		}
		s.searcher.status = searchapi.SemanticStatus{Indexing: true, WordsOnly: true, Message: "index unavailable"}
		res := callTool(t, s.MCPServer, "semantic_search", map[string]any{"query": "q"})
		data, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)
		require.JSONEq(t, `{"results":`+resultText(t, res)+`,"status":{"indexing":true,"words_only":true,"message":"index unavailable"}}`, string(data))
		require.Contains(t, resultHint(t, res), "index unavailable")
	}
}

func TestSearchIteratorLifecycle(t *testing.T) {
	for _, tc := range []struct {
		name    string
		iter    *searchIterator
		wantErr bool
		next    int
	}{
		{"cap", &searchIterator{}, false, 2},
		{"iteration error", &searchIterator{err: errors.New("iteration failed")}, true, 1},
		{"close error", &searchIterator{closeErr: errors.New("close failed")}, true, 2},
		{"both errors", &searchIterator{err: errors.New("iteration failed"), closeErr: errors.New("close failed")}, true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestServer()
			s.searcher.regexpIter = tc.iter
			res := callTool(t, s.MCPServer, "regexp_search", map[string]any{"pattern": "x", "limit": 1})
			require.Equal(t, tc.wantErr, res.IsError)
			require.Equal(t, 1, tc.iter.closed)
			require.Equal(t, tc.next, tc.iter.next)
			if tc.iter.err != nil {
				require.Contains(t, resultText(t, res), tc.iter.err.Error())
			}
			if tc.iter.closeErr != nil {
				require.Contains(t, resultText(t, res), tc.iter.closeErr.Error())
			}
		})
	}
}

func TestSearchCallCancellation(t *testing.T) {
	for _, tool := range []string{"regexp_search", "semantic_search"} {
		for _, source := range []string{"pre-call", "call", "workspace"} {
			t.Run(tool+"/"+source, func(t *testing.T) {
				workspace, cancelWorkspace := context.WithCancel(context.WithValue(context.Background(), ctxKey{}, "workspace"))
				defer cancelWorkspace()
				call, cancelCall := context.WithCancel(context.Background())
				defer cancelCall()
				if source == "pre-call" {
					cancelCall()
				}
				f := newFakeSearcher()
				f.onContext = func(ctx context.Context) {
					require.Equal(t, "workspace", ctx.Value(ctxKey{}))
					if source == "workspace" {
						cancelWorkspace()
					} else {
						cancelCall()
					}
					select {
					case <-ctx.Done():
					case <-time.After(time.Second):
						t.Fatal("search context not canceled")
					}
				}
				s := server.NewMCPServer("test", "1")
				RegisterSearchTools(workspace, s, f)
				var req mcp.CallToolRequest
				req.Params.Arguments = map[string]any{"pattern": "x", "query": "x"}
				res, err := s.GetTool(tool).Handler(call, req)
				require.NoError(t, err)
				require.True(t, res.IsError)
			})
		}
	}
}
