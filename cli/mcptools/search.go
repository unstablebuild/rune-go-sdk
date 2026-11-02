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
	"cmp"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/unstablebuild/rune-go-sdk/api/searchapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

// defaultRegexpLimit is the most lines regexp_search returns unless asked for more: enough
// to see how a pattern is used, few enough to stay readable.
const defaultRegexpLimit = 200

// RegisterSearchTools registers regexp_search and semantic_search, backed by searcher, on
// s. Search calls retain ctx's values and lifetime and also honor tool-call cancellation.
// A nil searcher panics.
func RegisterSearchTools(
	ctx context.Context,
	s *server.MCPServer,
	searcher searchapi.Searcher,
	opts ...Option,
) {
	if searcher == nil {
		panic("mcptools: nil searcher")
	}
	r := newRegistrar(s, opts)
	registerRegexpSearch(r, searcher, ctx)
	registerSemanticSearch(r, searcher, ctx)
}

var mcpRoots = mcp.WithArray("roots",
	mcp.WithStringItems(),
	mcp.Description(
		"Folders or files to search, workspace-relative or absolute inside the "+
			"workspace. Default the whole workspace. Example: [\"internal\", \"cmd/main.go\"]",
	),
)

// regexpMatch is the JSON shape of a line regexp_search found.
type regexpMatch struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

func registerRegexpSearch(r *registrar, searcher searchapi.Searcher, bgCtx context.Context) {
	r.add(
		mcp.NewTool("regexp_search",
			mcp.WithDescription(
				"Searches the lines of workspace files for a regular expression, "+
					"through the workspace's index.\n\n"+
					"Use this tool when:\n"+
					"- Finding exact text: a string literal, an error message, a comment\n"+
					"- Finding every line that matches a pattern\n\n"+
					"Do NOT use this tool when:\n"+
					"- You know a symbol's name — use lsp_definition or lsp_references\n"+
					"- You can describe what the code does but not its text — use "+
					"semantic_search\n\n"+
					"Hidden, ignored and excluded files are skipped unless a root names "+
					"them or a folder that holds them.\n\n"+
					"Returns [{path, line, text}] by file, then line: path is "+
					"workspace-relative and line is 1-based.",
			),
			mcp.WithString("pattern", mcp.Required(),
				mcp.Description(
					"Go RE2 regular expression, matched against each line; ^ and $ "+
						"match at the line's boundaries. Example: \"func \\\\w+Handler\\\\(\"",
				),
			),
			mcp.WithBoolean("ignore_case",
				mcp.Description("Match pattern case-insensitively. Default false.")),
			mcpRoots,
			mcp.WithArray("include",
				mcp.WithStringItems(),
				mcp.Description(
					"Only search files whose base name matches one of these globs. "+
						"Example: [\"*.go\", \"*.md\"]",
				),
			),
			mcp.WithNumber("limit",
				mcp.Description(fmt.Sprintf(
					"Most lines returned; 0 means no limit. Default %d.", defaultRegexpLimit)),
			),
			mcp.WithNumber("max_per_file",
				mcp.Description(
					"Most lines returned from any one file; 0 means no limit. "+
						"1 lists the files that match. Default 0.",
				),
			),
		),
		func(callCtx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			ctx, cancel := searchContext(bgCtx, callCtx)
			defer cancel()
			pattern, err := req.RequireString("pattern")
			if err != nil {
				return mcpErr(err), nil
			}
			limit, err := searchLimit(req, "limit", defaultRegexpLimit, math.MaxInt32-1)
			if err != nil {
				return mcpErr(err), nil
			}
			perFile, err := searchLimit(req, "max_per_file", 0, math.MaxInt32)
			if err != nil {
				return mcpErr(err), nil
			}
			search := searchapi.RegexpRequest{
				Pattern:    pattern,
				IgnoreCase: req.GetBool("ignore_case", false),
				Roots:      req.GetStringSlice("roots", nil),
				Include:    req.GetStringSlice("include", nil),
				MaxPerFile: int(perFile),
			}
			if limit > 0 {
				// One more line than asked tells whether lines were left out.
				search.Limit = int(limit) + 1
			}
			results, err := collectSearch(ctx, search.Limit, func() (iterator.Iterator[searchapi.RegexpResult], error) {
				return searcher.RegexpSearch(ctx, search)
			})
			if err != nil {
				return mcpErr(err), nil
			}
			hint := ""
			if limit > 0 && len(results) > int(limit) {
				results = results[:limit]
				hint = fmt.Sprintf("Showing the first %d matching lines; narrow pattern, "+
					"roots or include, or raise limit, to see the rest.", limit)
			}
			out := make([]regexpMatch, len(results))
			for i, m := range results {
				out[i] = regexpMatch{Path: m.Path, Line: m.Line, Text: m.Text}
			}
			return mcpJSONHint(out, hint)
		},
	)
}

// semanticMatch is the JSON shape of a file semantic_search found.
type semanticMatch struct {
	Path string `json:"path"`
	// Line is the line to read from: the definition of a looked-up name, the first line
	// with the query's words, or the start of the matched stretch.
	Line    int    `json:"line"`
	Segment string `json:"segment"`
	Start   int    `json:"start,omitempty"`
	End     int    `json:"end,omitempty"`
	Excerpt string `json:"excerpt,omitempty"`
	Text    string `json:"text,omitempty"`
}

func toSemanticMatch(r searchapi.SemanticResult) semanticMatch {
	line := 1
	if r.Segment == "lines" {
		line = cmp.Or(r.Focus, r.Loc, 1)
	}
	return semanticMatch{
		Path: r.Path, Line: line, Segment: r.Segment,
		Start: r.Loc, End: r.LocEnd, Excerpt: r.Excerpt, Text: r.Text,
	}
}

func registerSemanticSearch(r *registrar, searcher searchapi.Searcher, bgCtx context.Context) {
	r.add(
		mcp.NewTool("semantic_search",
			mcp.WithDescription(
				"Finds the workspace files that best answer a question, ranked by "+
					"meaning and by words, best first.\n\n"+
					"Use this tool when:\n"+
					"- You can describe what the code does but not the names it uses "+
					"(\"where are retries backed off\")\n"+
					"- Exploring an unfamiliar codebase for where a concept lives\n\n"+
					"Do NOT use this tool when:\n"+
					"- You know the exact text — use regexp_search\n"+
					"- You know a symbol's name — use lsp_definition or lsp_references\n\n"+
					"Returns [{path, line, segment, start, end, excerpt, text}]: path is "+
					"workspace-relative; line is the 1-based line to read from; segment "+
					"is \"lines\" or \"cells\" for the matched stretch from start to end, "+
					"or \"name\" when only the path matched; excerpt summarizes the "+
					"match, «marking» the query's words; text is the line at line. A "+
					"note follows when the index was still being built or could only "+
					"rank by words. Structured content includes results and status "+
					"{indexing, words_only, message}, even when results are empty.",
			),
			mcp.WithString("query", mcp.Required(),
				mcp.Description(
					"What to find, in natural language or as an identifier. Example: "+
						"\"where is the invoice total rounded\"",
				),
			),
			mcpRoots,
			mcp.WithNumber("limit",
				mcp.Description("Most files returned. Default 20.")),
		),
		func(callCtx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			ctx, cancel := searchContext(bgCtx, callCtx)
			defer cancel()
			query, err := req.RequireString("query")
			if err != nil {
				return mcpErr(err), nil
			}
			if strings.TrimSpace(query) == "" {
				return mcpErr(errors.New(`argument "query" must not be blank`)), nil
			}
			limit, err := searchLimit(req, "limit", 20, math.MaxInt32)
			if err != nil {
				return mcpErr(err), nil
			}
			var status searchapi.SemanticStatus
			results, err := collectSearch(ctx, 0, func() (iterator.Iterator[searchapi.SemanticResult], error) {
				it, st, err := searcher.SemanticSearch(ctx, searchapi.SemanticRequest{
					Query:   query,
					Roots:   req.GetStringSlice("roots", nil),
					Limit:   int(limit),
					Meaning: true,
				})
				status = st
				return it, err
			})
			if err != nil {
				if errors.Is(err, searchapi.ErrUnsupported) {
					return mcpErr(fmt.Errorf("%w; use regexp_search instead", err)), nil
				}
				return mcpErr(err), nil
			}
			out := make([]semanticMatch, len(results))
			for i, m := range results {
				out[i] = toSemanticMatch(m)
			}
			res, err := mcpJSONHint(out, semanticNote(status))
			if err != nil {
				return nil, err
			}
			res.StructuredContent = map[string]any{
				"results": out,
				"status": map[string]any{
					"indexing":   status.Indexing,
					"words_only": status.WordsOnly,
					"message":    status.Message,
				},
			}
			return res, nil
		},
	)
}

// semanticNote tells the caller why results with status may be missing files or ranked
// worse, if they may.
func semanticNote(status searchapi.SemanticStatus) string {
	if status.Message == "" {
		return ""
	}
	note := "Note: " + status.Message + "."
	if status.Indexing {
		note += " Files not indexed yet may be missing; regexp_search finds exact text " +
			"in every file."
	}
	return note
}

func searchLimit(req mcp.CallToolRequest, name string, def, max uint32) (uint32, error) {
	n, err := optionalUint(req, name, def)
	if err != nil {
		return 0, err
	}
	if n > max {
		return 0, fmt.Errorf("argument %q must be a whole number from 0 to %d, got %d", name, max, n)
	}
	return n, nil
}

func searchContext(workspace, call context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(workspace)
	stop := context.AfterFunc(call, cancel)
	// AfterFunc runs asynchronously, so handle an already canceled call before searching.
	if call.Err() != nil {
		cancel()
	}
	return ctx, func() { stop(); cancel() }
}

// collectSearch closes the iterator on every exit, including reaching a positive cap.
func collectSearch[T any](
	ctx context.Context, cap int, search func() (iterator.Iterator[T], error),
) (results []T, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	it, err := search()
	if it != nil {
		defer func() { err = errors.Join(err, it.Close()) }()
	}
	if err != nil {
		return nil, err
	}
	results = []T{}
	for cap == 0 || len(results) < cap {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		v, ok := it.Next(ctx)
		if !ok {
			break
		}
		results = append(results, v)
	}
	return results, errors.Join(ctx.Err(), it.Err())
}
