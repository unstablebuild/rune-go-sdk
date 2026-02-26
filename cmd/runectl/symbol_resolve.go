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
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
)

// symbolPosition holds a resolved URI and position for
// a workspace symbol.
type symbolPosition struct {
	URI      string
	Position semanticapi.Position
}

// symbolResolveHint biases kind preference when multiple
// symbols share the same name.
type symbolResolveHint int

const (
	hintDefault        symbolResolveHint = iota
	hintImplementation                   // prefer interfaces
	hintTypeDefinition                   // prefer variables/fields
	hintCallHierarchy                    // prefer functions/methods
	hintTypeHierarchy                    // prefer types
)

// preferredKinds returns the set of symbol kinds to
// prioritize for the given hint.
func preferredKinds(
	hint symbolResolveHint,
) map[semanticapi.SymbolKind]bool {
	switch hint {
	case hintImplementation:
		return map[semanticapi.SymbolKind]bool{
			semanticapi.SymbolKindInterface: true,
			semanticapi.SymbolKindClass:     true,
		}
	case hintTypeDefinition:
		return map[semanticapi.SymbolKind]bool{
			semanticapi.SymbolKindVariable: true,
			semanticapi.SymbolKindField:    true,
			semanticapi.SymbolKindConstant: true,
			semanticapi.SymbolKindProperty: true,
		}
	case hintCallHierarchy:
		return map[semanticapi.SymbolKind]bool{
			semanticapi.SymbolKindFunction: true,
			semanticapi.SymbolKindMethod:   true,
		}
	case hintTypeHierarchy:
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

// resolveSymbol queries WorkspaceSymbol, filters to exact
// name matches, sorts by kind preference, and deduplicates
// by (URI, line, character).
func resolveSymbol(
	ctx context.Context,
	lsp semanticapi.LSP,
	symbol string,
	hint symbolResolveHint,
) ([]symbolPosition, error) {
	syms, err := lsp.WorkspaceSymbol(
		ctx, semanticapi.WorkspaceSymbolParams{
			Query: symbol,
		},
	)
	if err != nil {
		return nil, err
	}
	// Filter to exact name matches. Methods and fields
	// have qualified names (e.g., "Receiver.Method"), so
	// also match the unqualified suffix.
	suffix := "." + symbol
	var exact []semanticapi.SymbolInformation
	for _, s := range syms {
		if s.Name == symbol || strings.HasSuffix(s.Name, suffix) {
			exact = append(exact, s)
		}
	}
	if len(exact) == 0 {
		return nil, fmt.Errorf(
			"symbol %q not found", symbol,
		)
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
	// Convert to symbolPosition and dedup.
	positions := make([]symbolPosition, len(exact))
	for i, s := range exact {
		positions[i] = symbolPosition{
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
	return dedup(positions, func(p symbolPosition) posKey {
		return posKey{
			URI:  p.URI,
			Line: p.Position.Line,
			Char: p.Position.Character,
		}
	}), nil
}

// resolveSymbolBest returns the single best match for a
// symbol name.
func resolveSymbolBest(
	ctx context.Context,
	lsp semanticapi.LSP,
	symbol string,
	hint symbolResolveHint,
) (symbolPosition, error) {
	positions, err := resolveSymbol(
		ctx, lsp, symbol, hint,
	)
	if err != nil {
		return symbolPosition{}, err
	}
	return positions[0], nil
}

// dedup removes duplicates from items based on a key
// function, preserving order.
func dedup[T any, K comparable](
	items []T, key func(T) K,
) []T {
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
