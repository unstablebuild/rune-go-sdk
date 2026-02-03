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

package term

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyCloneCells(t *testing.T) {
	reversible := []string{
		"",
		"a",
		"",
		"a\n",
		"a",
		"\n\n",
		"\n",
		"a\nb\nc\nd",
		"a\nb\nc\n",
		"",
		"a",
		"\ta",
		"a\n",
		"\x00",
		"\x00a",
		"a\x00",
		"\t\x00",
		"\x00\t",
		"\n\x00",
		"\x00\n",
	}

	t.Run("CopyCells", func(t *testing.T) {
		var dst [][]Cell
		for i, test := range reversible {
			dst = CopyCells(dst, StringToCells(test))
			assert.Equal(t, test, CellsToString(dst), i)
		}
	})
	t.Run("CloneCells", func(t *testing.T) {
		for i, test := range reversible {
			dst := CloneCells(StringToCells(test))
			assert.Equal(t, test, CellsToString(dst), i)
		}
	})
}

func benchmarkCellToString(b *testing.B, n int) {
	c := make([][]Cell, n)
	for i := 0; i < n; i++ {
		c[i] = make([]Cell, n)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CellsToString(c)
	}
}

func BenchmarkCellToString10(b *testing.B) {
	benchmarkCellToString(b, 10)
}
func BenchmarkCellToString100(b *testing.B) {
	benchmarkCellToString(b, 100)
}
func BenchmarkCellToString1000(b *testing.B) {
	benchmarkCellToString(b, 1000)
}
func BenchmarkCellToString10000(b *testing.B) {
	benchmarkCellToString(b, 10000)
}
