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

package syntaxapi

import (
	"context"
	"errors"

	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"github.com/unstablebuild/rune-go-sdk/term"
)

// NodeCaptureName represents a known node capture name on an IDE provided
// query. Values can be combined using bitwise OR to search for multiple node types.
type NodeCaptureName uint32

const (
	NodeCaptureScope NodeCaptureName = 1 << iota
	NodeCaptureDefinitionNamespace
	NodeCaptureReference
	NodeCaptureDefinitionFunc
	NodeCaptureDefinitionVar
	NodeCaptureDefinitionMethod
	NodeCaptureDefinitionType
	NodeCaptureNameAll = NodeCaptureScope | NodeCaptureDefinitionNamespace |
		NodeCaptureReference | NodeCaptureDefinitionFunc |
		NodeCaptureDefinitionVar | NodeCaptureDefinitionMethod |
		NodeCaptureDefinitionType
)

// Result represents a single search match.
type Result struct {
	File        workspaceapi.URI
	Text        string
	From        term.Coordinates
	To          term.Coordinates
	CaptureName string
}

// Match represents a resolved symbol location.
type Match struct {
	URI        string
	Pos        term.Coordinates
	Display    string
	ImportPath string
}

// ErrNoDot is returned by ResolveSymbol when the name does not contain a "."
// separating the qualifier (package or module) from the symbol, regardless of
// language, and therefore cannot be resolved to a qualified symbol.
var ErrNoDot = errors.New("name does not contain a qualifier separator")

// Progress reports incremental progress while resolving a symbol.
type Progress interface {
	Report(msg string, found int, step, total int64)
}

// ProgressFunc adapts a function to the Progress interface.
type ProgressFunc func(msg string, found int, step, total int64)

// Report implements Progress.
func (f ProgressFunc) Report(msg string, found int, step, total int64) {
	f(msg, found, step, total)
}

// Parser provides workspace-wide AST-level search and parsing capabilities.
type Parser interface {
	// Search searches for matches in the workspace using the given tree-sitter literal query
	// and a list of capture names that should be returned. An optional set of language names
	// (e.g. "go", "python") restricts the search to files of those languages.
	Search(query string, captureNames []string, languages ...string) (iterator.Iterator[Result], error)
	// SearchNode is implemented with Search by using an internally provided
	// query that is able to capture a known set of tree nodes across programming
	// languages. Multiple node types can be combined using bitwise OR. An
	// optional set of language names (e.g. "go", "python") restricts the search
	// to files of those languages.
	SearchNode(nodeTypes NodeCaptureName, languages ...string) (iterator.Iterator[Result], error)
	// Query searches for matches in a specific file using the given tree-sitter
	// literal query and a list of capture names that should be returned.
	Query(file workspaceapi.URI, query string, captureNames []string) (
		iterator.Iterator[Result], error,
	)
	// QueryNode is implemented with Query by using an internally provided
	// query that is able to capture a known set of tree nodes across
	// programming languages. Multiple node types can be combined using bitwise OR.
	QueryNode(file workspaceapi.URI, nodeTypes NodeCaptureName) (
		iterator.Iterator[Result], error,
	)
	// Highlight returns syntax highlighting locations for the given content,
	// interpreted as belonging to the file identified by uri.
	Highlight(uri workspaceapi.URI, content string) (
		iterator.Iterator[textapi.Location], error,
	)
	// ResolveSymbol resolves a dotted symbol name (e.g. "pkg.Symbol") to the
	// locations that define or reference it across the workspace.
	// Progress, if non-nil, receives incremental updates. ResolveSymbol returns
	// ErrNoDot when the name is not a fully qualified symbol name.
	ResolveSymbol(ctx context.Context, name string, progress Progress) (
		iterator.Iterator[Match], error,
	)
	// ListReferencedSymbols streams the fully qualified names (e.g.
	// "pkg.Symbol") of every package/module-qualified symbol referenced or
	// defined across the workspace, deduplicated by the caller.
	ListReferencedSymbols(ctx context.Context) (iterator.Iterator[string], error)
}
