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

// Parser provides AST-level search and parsing capabilities.
type Parser interface {
	// Search searches for matches in the workspace using the given tree-sitter literal query
	// and a list of capture names that should be returned.
	Search(query string, captureNames []string) (iterator.Iterator[Result], error)
	// SearchNode is implemented with Search by using an internally provided
	// query that is able to capture a known set of tree nodes across programming
	// languages. Multiple node types can be combined using bitwise OR.
	SearchNode(nodeTypes NodeCaptureName) (iterator.Iterator[Result], error)
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
}
