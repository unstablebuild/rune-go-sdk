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

package component

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

func testString(t *testing.T,
	fn func(string) tui.Component,
	width, height int, in, out string,
) {
	w := term.NewStringWriter(width, height)
	comp := fn(in)
	comp.Resize(width, height)
	comp.Draw(w)
	require.NoError(t, w.Flush())
	assert.Equal(t, out, w.String())
}

func TestStringCentered(t *testing.T) {
	tcases := []struct {
		in  string
		out string
	}{
		{
			in:  "aaaa",
			out: "     \n     \naaaa \n     \n     ",
		},
		{
			in:  "XXXXXXXXXXXXXXXX\nXXXXXXXXXX\nXXXXXXXXXX",
			out: "     \nXXXXX\nXXXXX\nXXXXX\n     ",
		},
		{
			in:  "X\nX\nX\nX\nX\nX\nX\nX\n",
			out: "  X  \n  X  \n  X  \n  X  \n  X  ",
		},
		{
			in:  "X\nX\nX\nX\nX\nX\nX\nX\n",
			out: "  X  \n  X  \n  X  \n  X  \n  X  ",
		},
		{
			in:  "XXXXXXXXXX\nXXXXXXXXXXX\nXXXXXXXXXXX\nXXXXXXXXXXX\nXXXXXXXXXXX\nXXXXXXXXXXX\nXXXXXXXXXXX\nXXXXXXXXXXX\n",
			out: "XXXXX\nXXXXX\nXXXXX\nXXXXX\nXXXXX",
		},
		{
			in:  "a",
			out: "     \n     \n  a  \n     \n     ",
		},
	}

	for _, tcase := range tcases {
		testString(t, func(str string) tui.Component {
			return NewStringWithConfig(str, StringConfig{Alignment: AlignmentCentered})
		}, 5, 5, tcase.in, tcase.out)
	}
}

func TestStringWithConfigDimensions(t *testing.T) {
	tcases := []struct {
		in             string
		expectedWidth  int
		expectedHeight int
	}{
		{
			in:             "XXXXXXXXXX\nBBBBBBBBBB\nCCCCCCCCCCCC\nDDDDDDDDDD\nEEEEEEEEEEE\nFFFFFFFFFF",
			expectedWidth:  12,
			expectedHeight: 6,
		},
		{
			in:             "X",
			expectedWidth:  1,
			expectedHeight: 1,
		},
		{
			in:             "X\nX\nX\nX\n",
			expectedWidth:  1,
			expectedHeight: 5,
		},
		{
			in:             "XXXXXX\nXXXX\nXXX\nXX\n",
			expectedWidth:  6,
			expectedHeight: 5,
		},
		{
			in:             "XXXXXXX\nXXXX\nXX\nXXXXXXXXXX",
			expectedWidth:  10,
			expectedHeight: 4,
		},
		{
			in:             "  2",
			expectedWidth:  5,
			expectedHeight: 1,
		},
	}

	for i, tcase := range tcases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			s := NewStringWithConfig(tcase.in, StringConfig{})
			actualWidth, actualHeight := s.Dimensions()
			assert.Equal(t, tcase.expectedWidth, actualWidth)
			assert.Equal(t, tcase.expectedHeight, actualHeight)
		})
	}
}

var fortune = `Love in your heart wasn't put there to stay.
Love isn't love 'til you give it_away.
		-- Oscar Hammerstein ⌘⌘
`

func benchmarkString(b *testing.B, fortunes int) {
	var builder strings.Builder
	for i := 0; i < fortunes; i++ {
		_, _ = builder.WriteString(fortune)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewString(builder.String())
	}
}

func BenchmarkString1(b *testing.B) {
	benchmarkString(b, 1)
}
func BenchmarkString10(b *testing.B) {
	benchmarkString(b, 10)
}
func BenchmarkString100(b *testing.B) {
	benchmarkString(b, 100)
}
func BenchmarkString1000(b *testing.B) {
	benchmarkString(b, 1000)
}
func BenchmarkString10000(b *testing.B) {
	benchmarkString(b, 10000)
}
