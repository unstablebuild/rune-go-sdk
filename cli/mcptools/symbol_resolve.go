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
	"fmt"
	"sort"
	"strings"

	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi"
)

// MaxSites caps the sites ResolveSymbol returns. A symbol tool asks the language server at
// each, so the cap bounds the requests one call makes.
const MaxSites = 16

// SymbolPosition is a site of a symbol: a place that defines or references it, where an
// LSP request about the symbol can run.
type SymbolPosition struct {
	URI      string
	Position semanticapi.Position
}

// Resolution is what ResolveSymbol found for a name.
type Resolution struct {
	// Sites are where to ask the language server about the symbol, at most MaxSites.
	Sites []SymbolPosition
	// Hint, when not empty, tells the caller how the answer may fall short: the name was
	// bare, so dependencies were not searched, or it had more than MaxSites sites.
	Hint string
}

// SymbolResolveHint biases kind preference when multiple symbols share the same name.
type SymbolResolveHint int

const (
	HintDefault        SymbolResolveHint = iota
	HintImplementation                   // prefer interfaces
	HintTypeDefinition                   // prefer variables/fields
	HintCallHierarchy                    // prefer functions/methods
	HintTypeHierarchy                    // prefer types
)

// preferredKinds returns the set of symbol kinds to prioritize for the given hint.
func preferredKinds(hint SymbolResolveHint) map[semanticapi.SymbolKind]bool {
	switch hint {
	case HintImplementation:
		return map[semanticapi.SymbolKind]bool{
			semanticapi.SymbolKindInterface: true,
			semanticapi.SymbolKindClass:     true,
		}
	case HintTypeDefinition:
		return map[semanticapi.SymbolKind]bool{
			semanticapi.SymbolKindVariable: true,
			semanticapi.SymbolKindField:    true,
			semanticapi.SymbolKindConstant: true,
			semanticapi.SymbolKindProperty: true,
		}
	case HintCallHierarchy:
		return map[semanticapi.SymbolKind]bool{
			semanticapi.SymbolKindFunction: true,
			semanticapi.SymbolKindMethod:   true,
		}
	case HintTypeHierarchy:
		return map[semanticapi.SymbolKind]bool{
			semanticapi.SymbolKindInterface: true,
			semanticapi.SymbolKindStruct:    true,
			semanticapi.SymbolKindClass:     true,
		}
	default:
		return map[semanticapi.SymbolKind]bool{
			semanticapi.SymbolKindFunction:  true,
			semanticapi.SymbolKindMethod:    true,
			semanticapi.SymbolKindInterface: true,
			semanticapi.SymbolKindStruct:    true,
			semanticapi.SymbolKindClass:     true,
		}
	}
}

// ResolveSymbol returns up to MaxSites sites of symbol, deduplicated by (URI, line,
// character).
//
// It asks parser's symbol index first, as Rune's go-to does. The index resolves a
// qualified name, such as "pkg.Symbol" or "Type.Method", to places that define or
// reference it, including symbols of dependencies that a language server's workspace
// search does not cover. When parser is nil, or the index fails, even partway, or has no
// match, ResolveSymbol falls back to WorkspaceSymbol: it keeps exact name matches and
// sorts them so the kinds hint prefers come first.
//
// The error for a bare name that is not found, and the Resolution's Hint for one that
// is, suggest qualifying it.
func ResolveSymbol(
	ctx context.Context,
	lsp semanticapi.LSP,
	parser syntaxapi.Parser,
	symbol string,
	hint SymbolResolveHint,
) (Resolution, error) {
	var sites []SymbolPosition
	if parser != nil {
		sites = indexSites(ctx, parser, symbol)
	}
	if len(sites) == 0 {
		var err error
		if sites, err = workspaceSymbolSites(ctx, lsp, symbol, hint); err != nil {
			return Resolution{}, err
		}
	}
	bare := !strings.Contains(symbol, ".")
	if len(sites) == 0 {
		if bare {
			return Resolution{}, fmt.Errorf("symbol %q not found. %s", symbol, bareHint(symbol))
		}
		return Resolution{}, fmt.Errorf("symbol %q not found", symbol)
	}
	var hints []string
	if len(sites) > MaxSites {
		sites = sites[:MaxSites]
		hints = append(hints, fmt.Sprintf("Only the first %d places that define or "+
			"reference %q were looked at; there are more, so this answer may be "+
			"incomplete. Qualify the name further, as pkg.Type.Method, to narrow it.",
			MaxSites, symbol))
	}
	if bare {
		hints = append(hints, bareHint(symbol))
	}
	return Resolution{Sites: sites, Hint: strings.Join(hints, " ")}, nil
}

// bareHint suggests qualifying symbol, a name without a dot.
func bareHint(symbol string) string {
	return fmt.Sprintf("%q has no qualifier, so only symbols defined in the workspace "+
		"were searched. Qualify it, as \"pkg.%s\" or \"Type.%s\", to also find symbols "+
		"of dependencies.", symbol, symbol, symbol)
}

// indexSites returns up to MaxSites+1 distinct sites of symbol from parser's symbol
// index, one more than ResolveSymbol keeps so it can tell there are more. It returns none
// when the index cannot resolve symbol or fails while it is read.
func indexSites(
	ctx context.Context,
	parser syntaxapi.Parser,
	symbol string,
) []SymbolPosition {
	it, err := parser.ResolveSymbol(ctx, symbol, nil)
	if err != nil {
		return nil
	}
	defer func() { _ = it.Close() }()
	seen := make(map[SymbolPosition]bool)
	var sites []SymbolPosition
	for len(sites) <= MaxSites {
		m, ok := it.Next(ctx)
		if !ok {
			if it.Err() != nil {
				return nil
			}
			return sites
		}
		sp := SymbolPosition{URI: m.URI, Position: semanticapi.Position{
			Line:      uint32(m.Pos.Y),
			Character: uint32(m.Pos.X),
		}}
		if !seen[sp] {
			seen[sp] = true
			sites = append(sites, sp)
		}
	}
	return sites
}

// workspaceSymbolSites queries WorkspaceSymbol, filters to exact name matches, sorts by
// kind preference, and deduplicates by (URI, line, character). It finds no sites, without
// an error, when no symbol matches.
func workspaceSymbolSites(
	ctx context.Context,
	lsp semanticapi.LSP,
	symbol string,
	hint SymbolResolveHint,
) ([]SymbolPosition, error) {
	syms, err := lsp.WorkspaceSymbol(ctx,
		semanticapi.WorkspaceSymbolParams{Query: symbol})
	if err != nil {
		return nil, err
	}
	// Filter to exact name matches. Methods and fields have qualified names (e.g.,
	// "Receiver.Method"), so also match the unqualified suffix.
	suffix := "." + symbol
	var exact []semanticapi.SymbolInformation
	for _, s := range syms {
		if s.Name == symbol || strings.HasSuffix(s.Name, suffix) {
			exact = append(exact, s)
		}
	}
	// Stable sort: preferred kinds first.
	pref := preferredKinds(hint)
	sort.SliceStable(exact, func(i, j int) bool {
		pi := pref[exact[i].Kind]
		pj := pref[exact[j].Kind]
		if pi != pj {
			return pi
		}
		return false
	})
	// Convert to SymbolPosition and dedup.
	positions := make([]SymbolPosition, len(exact))
	for i, s := range exact {
		positions[i] = SymbolPosition{
			URI: s.Location.URI,
			Position: semanticapi.Position{
				Line:      s.Location.Range.Start.Line,
				Character: s.Location.Range.Start.Character,
			},
		}
	}
	type posKey struct {
		URI  string
		Line uint32
		Char uint32
	}
	return dedup(positions, func(p SymbolPosition) posKey {
		return posKey{
			URI:  p.URI,
			Line: p.Position.Line,
			Char: p.Position.Character,
		}
	}), nil
}

// ResolveSymbolBest returns the first site ResolveSymbol finds.
func ResolveSymbolBest(
	ctx context.Context,
	lsp semanticapi.LSP,
	parser syntaxapi.Parser,
	symbol string,
	hint SymbolResolveHint,
) (SymbolPosition, error) {
	res, err := ResolveSymbol(ctx, lsp, parser, symbol, hint)
	if err != nil {
		return SymbolPosition{}, err
	}
	return res.Sites[0], nil
}

// dedup removes duplicates from items based on a key function, preserving order.
func dedup[T any, K comparable](items []T, key func(T) K) []T {
	seen := make(map[K]struct{}, len(items))
	out := make([]T, 0, len(items))
	for _, item := range items {
		k := key(item)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, item)
	}
	return out
}
