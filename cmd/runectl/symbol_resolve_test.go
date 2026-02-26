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
	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
)

func TestResolveSym(t *testing.T) {
	tests := []struct {
		name    string
		symbol  string
		hint    symbolResolveHint
		wantN   int
		wantURI string
		wantErr string
	}{
		{
			name:    "func",
			symbol:  "MyFunc",
			hint:    hintDefault,
			wantN:   1,
			wantURI: "file:///src/main.go",
		},
		{
			name:    "iface",
			symbol:  "MyInterface",
			hint:    hintDefault,
			wantN:   1,
			wantURI: "file:///src/types.go",
		},
		{
			name:    "miss",
			symbol:  "NonExistent",
			hint:    hintDefault,
			wantErr: `symbol "NonExistent" not found`,
		},
		{
			name:    "prefix",
			symbol:  "My",
			hint:    hintDefault,
			wantErr: `symbol "My" not found`,
		},
		{
			name:    "var",
			symbol:  "myVar",
			hint:    hintDefault,
			wantN:   1,
			wantURI: "file:///src/main.go",
		},
		{
			name:    "method",
			symbol:  "MyMethod",
			hint:    hintDefault,
			wantN:   1,
			wantURI: "file:///src/types.go",
		},
		{
			name:    "qual",
			symbol:  "MyStruct.MyMethod",
			hint:    hintDefault,
			wantN:   1,
			wantURI: "file:///src/types.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			defer env.cleanup()

			lsp, err := (&app{}).getLSP(t.Context())
			require.NoError(t, err)

			positions, err := resolveSymbol(
				t.Context(), lsp, tt.symbol, tt.hint,
			)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, positions, tt.wantN)
			require.Equal(t, tt.wantURI, positions[0].URI)
		})
	}
}

func TestResolveSymBest(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()

	lsp, err := (&app{}).getLSP(t.Context())
	require.NoError(t, err)

	pos, err := resolveSymbolBest(
		t.Context(), lsp, "MyFunc", hintDefault,
	)
	require.NoError(t, err)
	require.Equal(t, "file:///src/main.go", pos.URI)
	require.Equal(t, uint32(5), pos.Position.Line)
	require.Equal(t, uint32(0), pos.Position.Character)
}

func TestDedup(t *testing.T) {
	items := []symbolPosition{
		{URI: "a", Position: semanticapi.Position{Line: 1}},
		{URI: "a", Position: semanticapi.Position{Line: 1}},
		{URI: "b", Position: semanticapi.Position{Line: 2}},
		{URI: "a", Position: semanticapi.Position{Line: 1}},
	}
	type key struct {
		URI  string
		Line uint32
	}
	result := dedup(items, func(p symbolPosition) key {
		return key{URI: p.URI, Line: p.Position.Line}
	})
	require.Len(t, result, 2)
	require.Equal(t, "a", result[0].URI)
	require.Equal(t, "b", result[1].URI)
}
