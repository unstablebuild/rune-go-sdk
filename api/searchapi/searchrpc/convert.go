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

package searchrpc

import (
	"github.com/unstablebuild/rune-go-sdk/api/llmapi"
	"github.com/unstablebuild/rune-go-sdk/api/searchapi"
)

// ToProtoRegexpRequest converts a searchapi.RegexpRequest to its wire form.
func ToProtoRegexpRequest(r searchapi.RegexpRequest) *RegexpRequest {
	return &RegexpRequest{
		Pattern:    r.Pattern,
		IgnoreCase: r.IgnoreCase,
		Roots:      r.Roots,
		Include:    r.Include,
		Limit:      int32(r.Limit),
		MaxPerFile: int32(r.MaxPerFile),
	}
}

// FromProtoRegexpRequest converts a wire RegexpRequest to
// searchapi.RegexpRequest.
func FromProtoRegexpRequest(p *RegexpRequest) searchapi.RegexpRequest {
	return searchapi.RegexpRequest{
		Pattern:    p.GetPattern(),
		IgnoreCase: p.GetIgnoreCase(),
		Roots:      p.GetRoots(),
		Include:    p.GetInclude(),
		Limit:      int(p.GetLimit()),
		MaxPerFile: int(p.GetMaxPerFile()),
	}
}

// ToProtoRegexpResult converts a searchapi.RegexpResult to its wire form.
func ToProtoRegexpResult(r searchapi.RegexpResult) *RegexpResult {
	var spans []*Span
	if r.Spans != nil {
		spans = make([]*Span, len(r.Spans))
		for i, s := range r.Spans {
			spans[i] = &Span{Start: int32(s[0]), End: int32(s[1])}
		}
	}
	return &RegexpResult{Path: []byte(r.Path), Line: int32(r.Line), Text: []byte(r.Text), Spans: spans}
}

// FromProtoRegexpResult converts a wire RegexpResult to
// searchapi.RegexpResult.
func FromProtoRegexpResult(p *RegexpResult) searchapi.RegexpResult {
	var spans [][2]int
	if ps := p.GetSpans(); len(ps) > 0 {
		spans = make([][2]int, len(ps))
		for i, s := range ps {
			spans[i] = [2]int{int(s.GetStart()), int(s.GetEnd())}
		}
	}
	return searchapi.RegexpResult{
		Path:  string(p.GetPath()),
		Line:  int(p.GetLine()),
		Text:  string(p.GetText()),
		Spans: spans,
	}
}

// ToProtoSemanticRequest converts a searchapi.SemanticRequest to its wire
// form. Only the Provider and Name of its Model travel.
func ToProtoSemanticRequest(r searchapi.SemanticRequest) *SemanticRequest {
	return &SemanticRequest{
		Query:         r.Query,
		ModelProvider: r.Model.Provider,
		ModelName:     r.Model.Name,
		Roots:         r.Roots,
		Limit:         int32(r.Limit),
		Meaning:       r.Meaning,
		PrefixLast:    r.PrefixLast,
	}
}

// FromProtoSemanticRequest converts a wire SemanticRequest to
// searchapi.SemanticRequest.
func FromProtoSemanticRequest(p *SemanticRequest) searchapi.SemanticRequest {
	return searchapi.SemanticRequest{
		Query:      p.GetQuery(),
		Model:      llmapi.ModelEntry{Provider: p.GetModelProvider(), Name: p.GetModelName()},
		Roots:      p.GetRoots(),
		Limit:      int(p.GetLimit()),
		Meaning:    p.GetMeaning(),
		PrefixLast: p.GetPrefixLast(),
	}
}

// ToProtoSemanticResult converts a searchapi.SemanticResult to its wire
// form.
func ToProtoSemanticResult(r searchapi.SemanticResult) *SemanticResult {
	return &SemanticResult{
		Path:        r.Path,
		Kind:        r.Kind,
		Score:       r.Score,
		Z:           r.Z,
		Cosine:      r.Cosine,
		KeywordRank: int32(r.KeywordRank),
		Segment:     r.Segment,
		Loc:         int32(r.Loc),
		LocEnd:      int32(r.LocEnd),
		Focus:       int32(r.Focus),
		Excerpt:     r.Excerpt,
		Modified:    r.Modified,
		Size:        r.Size,
		Text:        r.Text,
	}
}

// FromProtoSemanticResult converts a wire SemanticResult to
// searchapi.SemanticResult.
func FromProtoSemanticResult(p *SemanticResult) searchapi.SemanticResult {
	return searchapi.SemanticResult{
		Path:        p.GetPath(),
		Kind:        p.GetKind(),
		Score:       p.GetScore(),
		Z:           p.GetZ(),
		Cosine:      p.GetCosine(),
		KeywordRank: int(p.GetKeywordRank()),
		Segment:     p.GetSegment(),
		Loc:         int(p.GetLoc()),
		LocEnd:      int(p.GetLocEnd()),
		Focus:       int(p.GetFocus()),
		Excerpt:     p.GetExcerpt(),
		Modified:    p.GetModified(),
		Size:        p.GetSize(),
		Text:        p.GetText(),
	}
}
