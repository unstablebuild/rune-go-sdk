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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/browserapi/browserrpc"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/storagerpc/docpb"
	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi/syntaxrpc"
	"github.com/unstablebuild/rune-go-sdk/api/textapi/textrpc"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi/workspacerpc"
	"github.com/unstablebuild/rune-go-sdk/handler/handlerrpc"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/term/termrpc"
	"google.golang.org/grpc"
)

func TestDatadir(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run("datadir")
	require.NoError(t, err)
	require.Equal(t, env.datadir+"\n", out)
}

func TestURI(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run("uri", "/tmp/test.txt")
	require.NoError(t, err)
	require.Contains(t, out, "file:///tmp/test.txt")
}

func TestOpen(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"open", "file:///tmp/test.txt",
	)
	require.NoError(t, err)
	require.Contains(t, out, "file:///tmp/test.txt")
}

func TestNotify(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run("notify", "info", "hello world")
	require.NoError(t, err)
	require.Equal(t, "OK\n", out)
	require.Equal(t, uint32(2), env.notif.lastLevel)
	require.Equal(t, "hello world", env.notif.lastMsg)
}

func TestWMFocus(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run("wm", "focus")
	require.NoError(t, err)
	require.Equal(t, "42\n", out)
}

func TestWMClose(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run("wm", "close", "99")
	require.NoError(t, err)
	require.Equal(t, "OK\n", out)
	require.Equal(t, uint64(99), env.wm.closedID)
}

func TestWMSplit(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"wm", "split", "-o", "right", "42",
		"file:///tmp/test.txt",
	)
	require.NoError(t, err)
	require.Equal(t, "200\n", out)
}

func TestWMSetContent(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"wm", "set-content", "42",
		"file:///tmp/test.txt",
	)
	require.NoError(t, err)
	require.Equal(t, "OK\n", out)
}

func TestEditorPrint(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"editor", "print", "file:///tmp/test.txt",
	)
	require.NoError(t, err)
	require.Contains(t, out, "hello world")
}

func TestEditorColor(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"editor", "color", "file:///tmp/test.txt",
		"blue", "red",
	)
	require.NoError(t, err)
	require.Equal(t, "OK\n", out)
	require.True(t, env.editor.setAttrsCalled)
}

func TestEditorEdit(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"editor", "edit",
		"file:///tmp/test.txt",
		"0", "0", "5", "0", "replacement",
	)
	require.NoError(t, err)
	require.Contains(t, out, "old-text")
}

func TestEditorCursorGet(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"editor", "cursor", "get",
		"file:///tmp/test.txt",
	)
	require.NoError(t, err)
	require.Equal(t, "10 20\n", out)
}

func TestEditorCursorSet(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"editor", "cursor", "set",
		"file:///tmp/test.txt", "5", "10",
	)
	require.NoError(t, err)
	require.Equal(t, "OK\n", out)
	require.True(t, env.editor.setCursorCalled)
}

func TestEditorLocationsSet(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"editor", "locations", "set",
		"file:///tmp/test.txt",
		"warning", "lint",
		`[{"from":{"x":0,"y":0},"to":{"x":5,"y":0},"message":"err"}]`,
	)
	require.NoError(t, err)
	require.Equal(t, "OK\n", out)
	require.True(t, env.editor.setLocCalled)
}

func TestEditorLocationsNext(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"editor", "locations", "next",
		"file:///tmp/test.txt", "lint",
	)
	require.NoError(t, err)
	require.Equal(t, "OK\n", out)
	require.True(t, env.editor.moveNextCalled)
}

func TestEditorLocationsPrev(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"editor", "locations", "prev",
		"file:///tmp/test.txt", "lint",
	)
	require.NoError(t, err)
	require.Equal(t, "OK\n", out)
	require.True(t, env.editor.movePrevCalled)
}

func TestStorageCreateAndGet(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"storage", "create", "doc1",
		`{"name":"test","value":42}`,
	)
	require.NoError(t, err)
	require.Equal(t, "OK\n", out)

	out, err = env.run(
		"storage", "get", "doc1",
	)
	require.NoError(t, err)
	require.NotEmpty(t, out)
}

func TestStorageSet(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"storage", "set", "doc2",
		`{"key":"val"}`,
	)
	require.NoError(t, err)
	require.Equal(t, "OK\n", out)
}

func TestStorageDelete(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"storage", "delete", "doc1",
	)
	require.NoError(t, err)
	require.Equal(t, "OK\n", out)
}

func TestStorageUpdate(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	_, _ = env.run(
		"storage", "create", "doc3",
		`{"name":"before"}`,
	)
	out, err := env.run(
		"storage", "update", "doc3",
		"name", "after",
	)
	require.NoError(t, err)
	require.Equal(t, "OK\n", out)
}

func TestStorageList(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	_, _ = env.run(
		"storage", "create", "list-doc",
		`{"x":1}`,
	)
	out, err := env.run("storage", "list")
	require.NoError(t, err)
	require.NotEmpty(t, out)
}

func TestNotifyInvalidLevel(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	_, err := env.run("notify", "bad", "msg")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid level")
}

func TestMissingEnvVars(t *testing.T) {
	t.Setenv("RUNE_SOCKET", "")
	t.Setenv("RUNE_DATADIR", "")
	a := &app{}
	_, err := a.getWorkspace()
	require.ErrorIs(t, err, errNotInRune)
}

func TestSyntaxSearch(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"syntax", "search",
		"(function_declaration) @fn",
	)
	require.NoError(t, err)
	require.Contains(t, out, "file:///src/main.go")
	require.Contains(t, out, "func main()")
	require.Contains(t, out, "fn")
}

func TestSyntaxSearchWithCapture(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"syntax", "search",
		"-c", "fn",
		"(function_declaration) @fn",
	)
	require.NoError(t, err)
	require.Contains(t, out, "func main()")
	require.Equal(t,
		"fn", env.syntax.lastCaptures[0],
	)
}

func TestSyntaxSearchNode(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"syntax", "searchnode", "func",
	)
	require.NoError(t, err)
	require.Contains(t, out, "MyFunc")
	require.Equal(t, uint32(8), env.syntax.lastNodeTypes)
}

func TestSyntaxQuery(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	testFile := filepath.Join(env.datadir, "test.go")
	out, err := env.run(
		"syntax", "query",
		testFile, "(package_clause) @pkg",
	)
	require.NoError(t, err)
	require.Contains(t, out, "package main")
	require.Contains(t,
		env.syntax.lastQueryURI,
		"test.go",
	)
	require.Equal(t,
		"(package_clause) @pkg",
		env.syntax.lastQuery,
	)
}

func TestSyntaxQueryWithCapture(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	testFile := filepath.Join(env.datadir, "test.go")
	out, err := env.run(
		"syntax", "query",
		"-c", "pkg",
		testFile, "(package_clause) @pkg",
	)
	require.NoError(t, err)
	require.Contains(t, out, "package main")
	require.Equal(t,
		"pkg", env.syntax.lastCaptures[0],
	)
}

func TestSyntaxQueryNode(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	testFile := filepath.Join(env.datadir, "test.go")
	out, err := env.run(
		"syntax", "querynode",
		testFile, "namespace",
	)
	require.NoError(t, err)
	require.Contains(t, out, "main")
	require.Contains(t,
		env.syntax.lastQueryURI,
		"test.go",
	)
	require.Equal(t, uint32(2), env.syntax.lastNodeTypes)
}

func TestSyntaxNodeTypeErr(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	_, err := env.run(
		"syntax", "searchnode", "invalid",
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid node type")
}

func TestDatadirFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
		check  func(t *testing.T, out string)
	}{
		{
			name:   "j",
			format: "json",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, `"success":true`)
			},
		},
		{
			name:   "t",
			format: "table",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "DATA DIR")
			},
		},
		{
			name:   "g",
			format: "{{.DataDir}}",
			check: func(t *testing.T, out string) {
				require.NotEmpty(t, out)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			defer env.cleanup()
			out, err := env.run(
				"datadir", "-F", tt.format,
			)
			require.NoError(t, err)
			tt.check(t, out)
		})
	}
}

func TestURIFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
		check  func(t *testing.T, out string)
	}{
		{
			name:   "j",
			format: "json",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "file:///")
				require.Contains(t, out, `"success":true`)
			},
		},
		{
			name:   "t",
			format: "table",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "URI")
			},
		},
		{
			name:   "g",
			format: "{{.URI}}",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "file:///")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			defer env.cleanup()
			out, err := env.run(
				"uri", "-F", tt.format,
				"/tmp/test.txt",
			)
			require.NoError(t, err)
			tt.check(t, out)
		})
	}
}

func TestWMFocusFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
		check  func(t *testing.T, out string)
	}{
		{
			name:   "j",
			format: "json",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "42")
				require.Contains(t, out, `"success":true`)
			},
		},
		{
			name:   "t",
			format: "table",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "WINDOW ID")
				require.Contains(t, out, "42")
			},
		},
		{
			name:   "g",
			format: "{{.WindowID}}",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "42")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			defer env.cleanup()
			out, err := env.run(
				"wm", "focus", "-F", tt.format,
			)
			require.NoError(t, err)
			tt.check(t, out)
		})
	}
}

func TestWMSplitFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
		check  func(t *testing.T, out string)
	}{
		{
			name:   "j",
			format: "json",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "200")
				require.Contains(t, out, `"success":true`)
			},
		},
		{
			name:   "t",
			format: "table",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "WINDOW ID")
				require.Contains(t, out, "200")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			defer env.cleanup()
			out, err := env.run(
				"wm", "split", "-F", tt.format,
				"42", "file:///tmp/test.txt",
			)
			require.NoError(t, err)
			tt.check(t, out)
		})
	}
}

func TestEditorEditFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
		check  func(t *testing.T, out string)
	}{
		{
			name:   "j",
			format: "json",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "old-text")
				require.Contains(t, out, `"success":true`)
			},
		},
		{
			name:   "t",
			format: "table",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "OLD")
				require.Contains(t, out, "old-text")
			},
		},
		{
			name:   "g",
			format: "{{.Old}}",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "old-text")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			defer env.cleanup()
			out, err := env.run(
				"editor", "edit", "-F", tt.format,
				"file:///tmp/test.txt",
				"0", "0", "5", "0", "replacement",
			)
			require.NoError(t, err)
			tt.check(t, out)
		})
	}
}

func TestEditorCursorGetFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
		check  func(t *testing.T, out string)
	}{
		{
			name:   "j",
			format: "json",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "10")
				require.Contains(t, out, "20")
				require.Contains(t, out, `"success":true`)
			},
		},
		{
			name:   "t",
			format: "table",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "X")
				require.Contains(t, out, "Y")
			},
		},
		{
			name:   "g",
			format: "{{.X}} {{.Y}}",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "10 20")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			defer env.cleanup()
			out, err := env.run(
				"editor", "cursor", "get",
				"-F", tt.format,
				"file:///tmp/test.txt",
			)
			require.NoError(t, err)
			tt.check(t, out)
		})
	}
}

func TestEditorPrintFormat(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	out, err := env.run(
		"editor", "print", "-F", "json",
		"file:///tmp/test.txt",
	)
	require.NoError(t, err)
	require.Contains(t, out, "hello world")
	require.Contains(t, out, `"success":true`)
}

func TestSyntaxSearchFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
		check  func(t *testing.T, out string)
	}{
		{
			name:   "t",
			format: "table",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "FILE")
				require.Contains(t, out, "TEXT")
			},
		},
		{
			name:   "g",
			format: "{{.File}} {{.Text}}",
			check: func(t *testing.T, out string) {
				require.Contains(t, out, "func main()")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			defer env.cleanup()
			out, err := env.run(
				"syntax", "search",
				"-F", tt.format,
				"(function_declaration) @fn",
			)
			require.NoError(t, err)
			tt.check(t, out)
		})
	}
}

func TestStorageUpdatePipeline(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	_, err := env.run(
		"storage", "create", "pipeline-doc",
		`{"name":{"first":"Alice","last":"Smith"}}`,
	)
	require.NoError(t, err)
	out, err := env.run(
		"storage", "update", "pipeline-doc",
		"name.first", "Bob",
	)
	require.NoError(t, err)
	require.Equal(t, "OK\n", out)
}

func TestStorageUpdateInvalidPath(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	_, _ = env.run(
		"storage", "create", "doc-inv",
		`{"a":"b"}`,
	)
	_, err := env.run(
		"storage", "update", "doc-inv",
		".bad", "value",
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty segment")
}

func TestStorageUpdateNested(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	_, _ = env.run(
		"storage", "create", "nested-doc",
		`{"field":{"nested_field":"old"}}`,
	)
	out, err := env.run(
		"storage", "update", "nested-doc",
		"field.nested_field", "value",
	)
	require.NoError(t, err)
	require.Equal(t, "OK\n", out)
	require.NotNil(t, env.docStore.lastUpdate)
	updates := env.docStore.lastUpdate.GetUpdates()
	require.Len(t, updates, 1)
	require.Equal(t,
		[]string{"field", "nested_field"},
		updates[0].GetFieldPath(),
	)
}

func TestStNoTable(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "get",
			args: []string{
				"storage", "get",
				"-F", "table", "doc",
			},
		},
		{
			name: "list",
			args: []string{
				"storage", "list",
				"-F", "table",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			defer env.cleanup()
			_, err := env.run(tt.args...)
			require.Error(t, err)
			require.Contains(t, err.Error(),
				"table format not supported",
			)
		})
	}
}

func TestStorageListFormat(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	_, _ = env.run(
		"storage", "create", "fmt-doc",
		`{"x":1}`,
	)
	out, err := env.run("storage", "list")
	require.NoError(t, err)
	require.NotEmpty(t, out)
}

func TestJSON(t *testing.T) {
	type testCase struct {
		name    string
		setup   func(*testEnv)
		args    []string
		wantErr bool
		raw     bool // iterator output, no success field
	}
	tests := []testCase{
		{
			name: "datadir/ok",
			args: []string{
				"datadir", "-F", "json",
			},
		},
		{
			name: "uri/ok",
			args: []string{
				"uri", "-F", "json",
				"/tmp/test.txt",
			},
		},
		{
			name: "open/ok",
			args: []string{
				"open", "-F", "json",
				"file:///tmp/test.txt",
			},
		},
		{
			name: "notify/ok",
			args: []string{
				"notify", "-F", "json",
				"info", "hello",
			},
		},
		{
			name: "focus/ok",
			args: []string{
				"wm", "focus", "-F", "json",
			},
		},
		{
			name: "split/ok",
			args: []string{
				"wm", "split", "-F", "json",
				"42", "file:///tmp/test.txt",
			},
		},
		{
			name: "close/ok",
			args: []string{
				"wm", "close", "-F", "json",
				"99",
			},
		},
		{
			name: "setcnt/ok",
			args: []string{
				"wm", "set-content",
				"-F", "json",
				"42", "file:///tmp/test.txt",
			},
		},
		{
			name: "eprint/ok",
			args: []string{
				"editor", "print",
				"-F", "json",
				"file:///tmp/test.txt",
			},
		},
		{
			name: "ecolor/ok",
			args: []string{
				"editor", "color",
				"-F", "json",
				"file:///tmp/test.txt", "blue",
			},
		},
		{
			name: "eedit/ok",
			args: []string{
				"editor", "edit",
				"-F", "json",
				"file:///tmp/test.txt",
				"0", "0", "5", "0", "text",
			},
		},
		{
			name: "ecur_get/ok",
			args: []string{
				"editor", "cursor", "get",
				"-F", "json",
				"file:///tmp/test.txt",
			},
		},
		{
			name: "ecur_set/ok",
			args: []string{
				"editor", "cursor", "set",
				"-F", "json",
				"file:///tmp/test.txt",
				"5", "10",
			},
		},
		{
			name: "eloc_set/ok",
			args: []string{
				"editor", "locations", "set",
				"-F", "json",
				"file:///tmp/test.txt",
				"warning", "lint",
				`[{"from":{"x":0,"y":0},` +
					`"to":{"x":5,"y":0},` +
					`"message":"err"}]`,
			},
		},
		{
			name: "eloc_nxt/ok",
			args: []string{
				"editor", "locations", "next",
				"-F", "json",
				"file:///tmp/test.txt", "lint",
			},
		},
		{
			name: "eloc_prv/ok",
			args: []string{
				"editor", "locations", "prev",
				"-F", "json",
				"file:///tmp/test.txt", "lint",
			},
		},
		{
			name: "st_create/ok",
			args: []string{
				"storage", "create",
				"-F", "json",
				"jdoc1", `{"a":"b"}`,
			},
		},
		{
			name: "st_set/ok",
			args: []string{
				"storage", "set",
				"-F", "json",
				"jdoc2", `{"a":"b"}`,
			},
		},
		{
			name: "st_get/ok",
			setup: func(env *testEnv) {
				_, _ = env.run(
					"storage", "create",
					"jdoc3", `{"a":"b"}`,
				)
			},
			args: []string{
				"storage", "get",
				"-F", "json", "jdoc3",
			},
			raw: true,
		},
		{
			name: "st_upd/ok",
			setup: func(env *testEnv) {
				_, _ = env.run(
					"storage", "create",
					"jdoc4", `{"a":"b"}`,
				)
			},
			args: []string{
				"storage", "update",
				"-F", "json",
				"jdoc4", "a", "c",
			},
		},
		{
			name: "st_del/ok",
			args: []string{
				"storage", "delete",
				"-F", "json", "jdoc5",
			},
		},
		{
			name: "st_list/ok",
			setup: func(env *testEnv) {
				_, _ = env.run(
					"storage", "create",
					"jdoc6", `{"x":1}`,
				)
			},
			args: []string{
				"storage", "list",
				"-F", "json",
			},
			raw: true,
		},
		{
			name: "synsearch/ok",
			args: []string{
				"syntax", "search",
				"-F", "json",
				"(function_declaration) @fn",
			},
			raw: true,
		},
		{
			name: "synnode/ok",
			args: []string{
				"syntax", "searchnode",
				"-F", "json", "func",
			},
			raw: true,
		},
		{
			name: "synquery/ok",
			setup: func(env *testEnv) {},
			args: []string{
				"syntax", "query",
				"-F", "json",
				"/tmp/test.go",
				"(package_clause) @pkg",
			},
			raw: true,
		},
		{
			name: "synqnode/ok",
			args: []string{
				"syntax", "querynode",
				"-F", "json",
				"/tmp/test.go", "var",
			},
			raw: true,
		},
		{
			name: "notify/err",
			args: []string{
				"notify", "-F", "json",
				"bad", "msg",
			},
			wantErr: true,
		},
		{
			name: "close/err",
			args: []string{
				"wm", "close", "-F", "json",
				"abc",
			},
			wantErr: true,
		},
		{
			name: "split/err",
			args: []string{
				"wm", "split", "-F", "json",
				"-o", "bad",
				"42", "file:///tmp/test.txt",
			},
			wantErr: true,
		},
		{
			name: "setcnt/err",
			args: []string{
				"wm", "set-content",
				"-F", "json",
				"abc", "file:///tmp/test.txt",
			},
			wantErr: true,
		},
		{
			name: "float/err",
			args: []string{
				"wm", "floating",
				"-F", "json",
				"file:///tmp/test.txt",
			},
			wantErr: true,
		},
		{
			name: "eedit/err",
			args: []string{
				"editor", "edit",
				"-F", "json",
				"file:///tmp/test.txt",
				"a", "0", "5", "0", "text",
			},
			wantErr: true,
		},
		{
			name: "ecur_set/err",
			args: []string{
				"editor", "cursor", "set",
				"-F", "json",
				"file:///tmp/test.txt",
				"a", "0",
			},
			wantErr: true,
		},
		{
			name: "eloc_set/err",
			args: []string{
				"editor", "locations", "set",
				"-F", "json",
				"file:///tmp/test.txt",
				"bad", "lint", "[]",
			},
			wantErr: true,
		},
		{
			name: "st_create/err",
			args: []string{
				"storage", "create",
				"-F", "json",
				"doc", "not-json",
			},
			wantErr: true,
		},
		{
			name: "st_set/err",
			args: []string{
				"storage", "set",
				"-F", "json",
				"doc", "not-json",
			},
			wantErr: true,
		},
		{
			name: "st_upd/err",
			args: []string{
				"storage", "update",
				"-F", "json",
				"doc", ".bad", "value",
			},
			wantErr: true,
		},
		{
			name: "synnode/err",
			args: []string{
				"syntax", "searchnode",
				"-F", "json", "invalid",
			},
			wantErr: true,
		},
		{
			name: "synqnode/err",
			args: []string{
				"syntax", "querynode",
				"-F", "json",
				"/tmp/test.go", "invalid",
			},
			wantErr: true,
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
			require.NoError(t, err)
			out = strings.TrimSpace(out)
			var m map[string]any
			require.NoError(t,
				json.Unmarshal(
					[]byte(out), &m,
				),
				"invalid JSON: %s", out,
			)
			if tt.wantErr {
				require.Equal(t,
					false, m["success"],
				)
				errMsg, ok :=
					m["error"].(string)
				require.True(t, ok,
					"missing error: %s", out,
				)
				require.NotEmpty(t, errMsg)
			} else if !tt.raw {
				require.Equal(t,
					true, m["success"],
				)
			}
		})
	}
}

type mockNotifications struct {
	browserrpc.UnimplementedNotificationsServer
	lastLevel uint32
	lastMsg   string
}

func (m *mockNotifications) Notify(
	_ context.Context,
	req *browserrpc.NotifyRequest,
) (*browserrpc.NotifyResponse, error) {
	m.lastLevel = req.GetLevel()
	m.lastMsg = req.GetMsg()
	return &browserrpc.NotifyResponse{
		Id: "notif-1",
	}, nil
}

type mockWindowManager struct {
	browserrpc.UnimplementedWindowManagerServer
	focusWindowID uint64
	closedID      uint64
}

func (m *mockWindowManager) Focus(
	_ context.Context, _ *browserrpc.FocusRequest,
) (*browserrpc.FocusResponse, error) {
	return &browserrpc.FocusResponse{
		WindowId: m.focusWindowID,
	}, nil
}

func (m *mockWindowManager) CloseWindow(
	_ context.Context,
	req *browserrpc.WindowCloseRequest,
) (*browserrpc.WindowCloseResponse, error) {
	m.closedID = req.GetWindowId()
	return &browserrpc.WindowCloseResponse{}, nil
}

func (m *mockWindowManager) Split(
	stream grpc.BidiStreamingServer[browserrpc.SplitWindowMessage, handlerrpc.ServerMessage],
) error {
	_, err := stream.Recv()
	if err != nil {
		return err
	}
	resp := &handlerrpc.ServerMessage{
		Type: handlerrpc.MessageType_Response,
		Response: &handlerrpc.InstallResourceResponse{
			WindowId: 200,
		},
	}
	return stream.Send(resp)
}

func (m *mockWindowManager) SetContent(
	stream grpc.BidiStreamingServer[browserrpc.WindowSetContentMessage, handlerrpc.ServerMessage],
) error {
	_, err := stream.Recv()
	if err != nil {
		return err
	}
	resp := &handlerrpc.ServerMessage{
		Type: handlerrpc.MessageType_Response,
		Response: &handlerrpc.InstallResourceResponse{
			WindowId: 300,
		},
	}
	return stream.Send(resp)
}

type mockResourceOpener struct {
	browserrpc.UnimplementedResourceOpenerServer
}

func (m *mockResourceOpener) Open(
	_ context.Context,
	req *browserrpc.OpenResourceRequest,
) (*browserrpc.OpenResourceResponse, error) {
	return &browserrpc.OpenResourceResponse{
		Uri: req.GetResource(),
	}, nil
}

type mockScheme struct {
	workspacerpc.UnimplementedSchemeServer
}

func (m *mockScheme) URI(
	_ context.Context,
	req *workspacerpc.URIRequest,
) (*workspacerpc.URIResponse, error) {
	uri := fmt.Sprintf(
		"file://%s", req.GetPath(),
	)
	return &workspacerpc.URIResponse{Uri: uri}, nil
}

type mockEditor struct {
	textrpc.UnimplementedEditorServer
	cursorX         int32
	cursorY         int32
	setCursorCalled bool
	setAttrsCalled  bool
	setLocCalled    bool
	moveNextCalled  bool
	movePrevCalled  bool
}

func (m *mockEditor) Editor(
	_ context.Context, _ *textrpc.EditorRequest,
) (*textrpc.EditorResponse, error) {
	return &textrpc.EditorResponse{}, nil
}

func (m *mockEditor) Cursor(
	_ context.Context, _ *textrpc.CursorRequest,
) (*textrpc.CursorResponse, error) {
	return &textrpc.CursorResponse{
		Pos: &termrpc.Coordinates{
			X: m.cursorX, Y: m.cursorY,
		},
	}, nil
}

func (m *mockEditor) SetCursor(
	_ context.Context,
	req *textrpc.SetCursorRequest,
) (*textrpc.SetCursorResponse, error) {
	m.setCursorCalled = true
	m.cursorX = req.GetPos().GetX()
	m.cursorY = req.GetPos().GetY()
	return &textrpc.SetCursorResponse{}, nil
}

func (m *mockEditor) EditCell(
	_ context.Context,
	req *textrpc.EditCellRequest,
) (*textrpc.EditCellResponse, error) {
	return &textrpc.EditCellResponse{
		From: &termrpc.Coordinates{
			X: req.GetStart().GetX(),
			Y: req.GetStart().GetY(),
		},
		To: &termrpc.Coordinates{
			X: req.GetEnd().GetX(),
			Y: req.GetEnd().GetY(),
		},
		Old: "old-text",
	}, nil
}

func (m *mockEditor) RawCells(
	_ context.Context,
	_ *textrpc.RawCellsRequest,
) (*textrpc.RawCellsResponse, error) {
	cells := term.StringToCells("hello world")
	return textrpc.NewRawCellsResponse(cells), nil
}

func (m *mockEditor) SetDefaultAttributes(
	_ context.Context,
	_ *textrpc.SetDefaultAttributesRequest,
) (*textrpc.SetDefaultAttributesResponse, error) {
	m.setAttrsCalled = true
	return &textrpc.SetDefaultAttributesResponse{}, nil
}

func (m *mockEditor) SetLocationList(
	_ context.Context,
	_ *textrpc.SetLocationListRequest,
) (*textrpc.SetLocationListResponse, error) {
	m.setLocCalled = true
	return &textrpc.SetLocationListResponse{}, nil
}

func (m *mockEditor) MoveToNextLocation(
	_ context.Context,
	_ *textrpc.MoveToLocationRequest,
) (*textrpc.MoveToLocationResponse, error) {
	m.moveNextCalled = true
	return &textrpc.MoveToLocationResponse{}, nil
}

func (m *mockEditor) MoveToPrevLocation(
	_ context.Context,
	_ *textrpc.MoveToLocationRequest,
) (*textrpc.MoveToLocationResponse, error) {
	m.movePrevCalled = true
	return &textrpc.MoveToLocationResponse{}, nil
}

type mockDocStore struct {
	docpb.UnimplementedDocumentStoreServer
	docs       map[string][]byte
	lastUpdate *docpb.UpdateDocumentRequest
}

func newMockDocStore() *mockDocStore {
	return &mockDocStore{
		docs: make(map[string][]byte),
	}
}

func (m *mockDocStore) Create(
	_ context.Context,
	req *docpb.CreateDocumentRequest,
) (*docpb.CreateDocumentResponse, error) {
	m.docs[req.GetId()] = req.GetData()
	return &docpb.CreateDocumentResponse{}, nil
}

func (m *mockDocStore) Set(
	_ context.Context,
	req *docpb.SetDocumentRequest,
) (*docpb.DocumentResponse, error) {
	m.docs[req.GetId()] = req.GetData()
	return &docpb.DocumentResponse{}, nil
}

func (m *mockDocStore) Get(
	_ context.Context,
	req *docpb.GetDocumentRequest,
) (*docpb.GetDocumentResponse, error) {
	data, ok := m.docs[req.GetId()]
	if !ok {
		return &docpb.GetDocumentResponse{}, nil
	}
	return &docpb.GetDocumentResponse{
		Data: data,
	}, nil
}

func (m *mockDocStore) Delete(
	_ context.Context,
	req *docpb.DeleteDocumentRequest,
) (*docpb.DocumentResponse, error) {
	delete(m.docs, req.GetId())
	return &docpb.DocumentResponse{}, nil
}

func (m *mockDocStore) Update(
	_ context.Context,
	req *docpb.UpdateDocumentRequest,
) (*docpb.UpdateDocumentResponse, error) {
	m.lastUpdate = req
	return &docpb.UpdateDocumentResponse{}, nil
}

func (m *mockDocStore) List(
	_ *docpb.ListDocumentRequest,
	stream grpc.ServerStreamingServer[docpb.ListDocumentResponse],
) error {
	for _, data := range m.docs {
		err := stream.Send(
			&docpb.ListDocumentResponse{Data: data},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

type mockSyntax struct {
	syntaxrpc.UnimplementedSyntaxServer
	lastCaptures  []string
	lastNodeTypes uint32
	lastQueryURI  string
	lastQuery     string
}

func (m *mockSyntax) Search(
	req *syntaxrpc.SearchRequest,
	stream grpc.ServerStreamingServer[syntaxrpc.SearchResponse],
) error {
	m.lastCaptures = req.GetCaptureNames()
	return stream.Send(&syntaxrpc.SearchResponse{
		Uri:         "file:///src/main.go",
		Text:        "func main()",
		From:        &termrpc.Coordinates{X: 0, Y: 5},
		To:          &termrpc.Coordinates{X: 12, Y: 5},
		CaptureName: "fn",
	})
}

func (m *mockSyntax) SearchNode(
	req *syntaxrpc.SearchNodeRequest,
	stream grpc.ServerStreamingServer[syntaxrpc.SearchResponse],
) error {
	m.lastNodeTypes = req.GetNodeTypes()
	return stream.Send(&syntaxrpc.SearchResponse{
		Uri:         "file:///src/main.go",
		Text:        "MyFunc",
		From:        &termrpc.Coordinates{X: 5, Y: 10},
		To:          &termrpc.Coordinates{X: 11, Y: 10},
		CaptureName: "definition.func",
	})
}

func (m *mockSyntax) Query(
	req *syntaxrpc.QueryRequest,
	stream grpc.ServerStreamingServer[syntaxrpc.SearchResponse],
) error {
	m.lastQueryURI = req.GetUri()
	m.lastQuery = req.GetQuery()
	m.lastCaptures = req.GetCaptureNames()
	return stream.Send(&syntaxrpc.SearchResponse{
		Uri:         req.GetUri(),
		Text:        "package main",
		From:        &termrpc.Coordinates{X: 0, Y: 0},
		To:          &termrpc.Coordinates{X: 12, Y: 0},
		CaptureName: "pkg",
	})
}

func (m *mockSyntax) QueryNode(
	req *syntaxrpc.QueryNodeRequest,
	stream grpc.ServerStreamingServer[syntaxrpc.SearchResponse],
) error {
	m.lastQueryURI = req.GetUri()
	m.lastNodeTypes = req.GetNodeTypes()
	return stream.Send(&syntaxrpc.SearchResponse{
		Uri:         req.GetUri(),
		Text:        "main",
		From:        &termrpc.Coordinates{X: 8, Y: 0},
		To:          &termrpc.Coordinates{X: 12, Y: 0},
		CaptureName: "definition.namespace",
	})
}

type testEnv struct {
	t        *testing.T
	srv      *grpc.Server
	socket   string
	datadir  string
	notif    *mockNotifications
	wm       *mockWindowManager
	opener   *mockResourceOpener
	scheme   *mockScheme
	editor   *mockEditor
	docStore *mockDocStore
	syntax   *mockSyntax
	cleanup  func()
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	datadir := t.TempDir()
	sockPath := filepath.Join(
		datadir, "rune.sock",
	)

	lis, err := net.Listen("unix", sockPath)
	require.NoError(t, err)

	srv := grpc.NewServer()

	notif := &mockNotifications{}
	wm := &mockWindowManager{focusWindowID: 42}
	opener := &mockResourceOpener{}
	scheme := &mockScheme{}
	ed := &mockEditor{cursorX: 10, cursorY: 20}
	ds := newMockDocStore()
	syn := &mockSyntax{}

	browserrpc.RegisterNotificationsServer(
		srv, notif,
	)
	browserrpc.RegisterWindowManagerServer(srv, wm)
	browserrpc.RegisterResourceOpenerServer(
		srv, opener,
	)
	workspacerpc.RegisterSchemeServer(srv, scheme)
	textrpc.RegisterEditorServer(srv, ed)
	docpb.RegisterDocumentStoreServer(srv, ds)
	syntaxrpc.RegisterSyntaxServer(srv, syn)

	go func() { _ = srv.Serve(lis) }()

	t.Setenv("RUNE_SOCKET", sockPath)
	t.Setenv("RUNE_DATADIR", datadir)

	return &testEnv{
		t: t, srv: srv, socket: sockPath,
		datadir: datadir, notif: notif, wm: wm,
		opener: opener, scheme: scheme,
		editor: ed, docStore: ds, syntax: syn,
		cleanup: func() { srv.Stop() },
	}
}

func (e *testEnv) run(
	args ...string,
) (string, error) {
	e.t.Helper()

	root := newRootCLI()

	old := os.Stdout
	r, ww, err := os.Pipe()
	require.NoError(e.t, err)
	os.Stdout = ww

	runErr := root.Run(
		context.Background(), args,
	)

	_ = ww.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	_ = r.Close()

	return buf.String(), runErr
}
