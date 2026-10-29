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

// Package searchapi searches the files of the workspace an extension is
// bound to, through the indexes the workspace host keeps of them.
package searchapi

import (
	"context"
	"errors"

	"github.com/unstablebuild/rune-go-sdk/api/llmapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

// ErrUnsupported is reported by a Searcher that cannot answer a search:
// the workspace has no index, a root is outside the workspace, or the
// model cannot encode the query.
var ErrUnsupported = errors.New("search: not supported")

// Searcher searches the files of one workspace.
//
// Both searches may report their errors either from the call or from
// the iterator's Err. Closing the iterator early cancels the search.
type Searcher interface {
	// RegexpSearch streams the lines that match a regular expression request, ordered by
	// file in index order and then by line. An invalid pattern or Include glob is an
	// error, and a root outside the workspace is ErrUnsupported. Files the workspace
	// hides from search (hidden, ignored or excluded) are skipped, unless a root names
	// them or a folder that holds them: like ripgrep, filtering applies only below a
	// root.
	RegexpSearch(ctx context.Context, req RegexpRequest) (
		iterator.Iterator[RegexpResult], error,
	)
	// SemanticSearch streams the files that best answer req, best first, ranked by
	// meaning and by words. It reports llmapi.ErrModelNotFound when req.Model is not
	// known, and ErrUnsupported when it cannot encode the query with it or a root is
	// outside the workspace.
	SemanticSearch(ctx context.Context, req SemanticRequest) (
		iterator.Iterator[SemanticResult], error,
	)
}

// RegexpRequest selects the lines a RegexpSearch reports.
type RegexpRequest struct {
	// Pattern is a Go RE2 regular expression matched against each line
	// without its line ending; ^ and $ match at the line's boundaries.
	Pattern string
	// IgnoreCase matches Pattern case-insensitively.
	IgnoreCase bool
	// Roots are the folders or files to search, workspace-relative or
	// absolute inside the workspace. Empty means the whole workspace.
	Roots []string
	// Include keeps only the files whose base name matches one of these
	// path.Match globs, such as "*.go". Empty keeps every file.
	Include []string
	// Limit is the most lines reported; 0 means no limit.
	Limit int
	// MaxPerFile is the most lines reported from any one file; 0 means
	// no limit. 1 lists the files that match.
	MaxPerFile int
}

// RegexpResult is one line that matched a RegexpRequest.
type RegexpResult struct {
	// Path is the file, workspace-relative with slash separators.
	Path string
	// Line is the line number, from 1.
	Line int
	// Text is the line without its line ending.
	Text string
	// Spans are the [start, end) byte offsets of the matches in Text.
	Spans [][2]int
}

// SemanticRequest is a natural-language or identifier search over the
// workspace.
type SemanticRequest struct {
	// Query may carry a leading "code:" prefix, which is stripped.
	Query string
	// Model encodes Query. Only its Provider and Name identify it; a
	// Name without Provider is an alias the host resolves. The zero
	// value means the "search" alias, the configured local embedding
	// model.
	Model llmapi.ModelEntry
	// Roots are workspace-relative folders that limit results; empty
	// means the whole workspace.
	Roots []string
	// Limit is the most files reported; 0 means 20.
	Limit int
	// Meaning ranks by meaning as well as words. False asks for a
	// words-only search, as while typing.
	Meaning bool
	// PrefixLast marks the last query word as still being typed, so it
	// also matches as a prefix.
	PrefixLast bool
}

// SemanticResult is a result for a SemanticRequest.
type SemanticResult struct {
	Path string
	Kind string
	// Score is the fused rank score and is only meaningful for ordering.
	Score float64
	// Z and Cosine describe the best meaning segment
	// (zero when the hit matched by words only)
	Z      float64
	Cosine float64
	// KeywordRank is the 1-based rank among keyword matches (zero when none)
	KeywordRank int
	// Segment is "lines" or "cells" for a stretch of the file, or "name" when only its
	// path matched;
	Segment string
	// Loc/LocEnd are the first and last line (or cell) of the matched stretch, from 1,
	// and zero for "name"
	Loc    int
	LocEnd int
	// Focus is the line to open at (the definition of a looked-up name or the first line
	// with the query's words), zero meaning Loc
	Focus int
	// Excerpt is a one-line summary or the «marked» keyword snippet
	Excerpt string
	// Modified is the file's mtime in Unix seconds, Size its byte size.
	Modified int64
	Size     int64
	// Text is the "lines" hit's line to open at (Focus, else Loc) as written, without its
	// line ending, as grep prints it, empty for "cells" and "name" hits.
	Text string
}
