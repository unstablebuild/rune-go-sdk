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

package handlerrpc

import (
	"context"
	"testing"

	"github.com/unstablebuild/rune-go-sdk/term"
)

// ---------------------------------------------------------------------------
// Draw response writer benchmark — proves that the cell slab eliminates
// per-cell heap allocations. allocs/op should be constant (~5) regardless
// of screen size.
// ---------------------------------------------------------------------------

func benchDrawResponseWriter(b *testing.B, width, height int) {
	ctx := context.Background()
	cell := term.Cell{
		Ch:    'A',
		Width: 1,
		Bytes: 1,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		w := newDrawResponseWriter(ctx, width, height)
		for y := range height {
			for x := range width {
				w.SetCell(term.Coordinates{X: x, Y: y}, cell)
			}
		}
	}
}

func BenchmarkDrawResponseWriter(b *testing.B) {
	// allocs/op should remain constant across sizes — only the slab
	// allocations in newDrawResponseWriter, not one per cell.
	b.Run("60x15", func(b *testing.B) { benchDrawResponseWriter(b, 60, 15) })
	b.Run("120x30", func(b *testing.B) { benchDrawResponseWriter(b, 120, 30) })
	b.Run("240x60", func(b *testing.B) { benchDrawResponseWriter(b, 240, 60) })
}
