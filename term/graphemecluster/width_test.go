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

package graphemecluster

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/rivo/uniseg"
)

func TestWidth(t *testing.T) {
	suite := []struct {
		inputStr string
		expected int
	}{
		{"", 0},
		{"a", 1},
		{"aa", 2},
		{"", 2},
		{"1", 3},
		{"1", 3},
		{"", 4},
		{"中", 2},
		{"\t", 0},
		{"\n", 0},
	}

	for i, test := range suite {
		t.Run(fmt.Sprintf("test case %d: %q", i, test.inputStr), func(t *testing.T) {
			actual := StringWidth(test.inputStr)
			assert.Equal(t, test.expected, actual)
		})
	}
}

// TestStepStringMatchesUniseg pins StepString to uniseg's full segmentation:
// identical clusters and rest, and identical widths except for the nerd-font
// overrides applied by graphemeClusterWidth.
func TestStepStringMatchesUniseg(t *testing.T) {
	suite := []string{
		"",
		"hello world",
		"héllo wörld",
		"e\u0301e\u0301",
		"🚀 rocket",
		"👨‍👩‍👧 family",
		"🇺🇸🇯🇵 flags",
		"❤️ vs16 ❤ plain",
		"中文字符 mixed ascii",
		"\t\n controls \x00\x07",
		"nerd 󰗠 icons ",
		"한글 hangul",
	}

	for i, input := range suite {
		t.Run(fmt.Sprintf("test case %d: %q", i, input), func(t *testing.T) {
			gotStr, wantStr := input, input
			gotState, wantState := -1, -1
			for len(wantStr) > 0 {
				var gotCluster, wantCluster string
				var gotWidth uint8
				var boundaries int
				gotCluster, gotStr, gotWidth, gotState = StepString(gotStr, gotState)
				wantCluster, wantStr, boundaries, wantState = uniseg.StepString(wantStr, wantState)
				assert.Equal(t, wantCluster, gotCluster)
				assert.Equal(t, wantStr, gotStr)
				wantWidth := boundaries >> uniseg.ShiftWidth
				if gotWidth != uint8(wantWidth) {
					// The only allowed deviation is the nerd-font override
					// promoting a width-1 icon to 2 cells.
					assert.Equal(t, 1, wantWidth, "cluster %q", wantCluster)
					assert.Equal(t, uint8(2), gotWidth, "cluster %q", wantCluster)
				}
			}
		})
	}
}

func BenchmarkStepStringASCII(b *testing.B) {
	s := strings.Repeat("the quick brown fox jumps over the lazy dog ", 20)
	b.SetBytes(int64(len(s)))
	b.ReportAllocs()
	for range b.N {
		state := -1
		str := s
		for len(str) > 0 {
			_, str, _, state = StepString(str, state)
		}
	}
}

func BenchmarkStepStringMixed(b *testing.B) {
	s := strings.Repeat("héllo 🚀 中文 👨‍👩‍👧 word ", 40)
	b.SetBytes(int64(len(s)))
	b.ReportAllocs()
	for range b.N {
		state := -1
		str := s
		for len(str) > 0 {
			_, str, _, state = StepString(str, state)
		}
	}
}

func BenchmarkStringWidthASCII(b *testing.B) {
	s := strings.Repeat("the quick brown fox jumps over the lazy dog ", 20)
	b.SetBytes(int64(len(s)))
	b.ReportAllocs()
	for range b.N {
		StringWidth(s)
	}
}
