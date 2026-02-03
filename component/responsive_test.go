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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

func TestResponsiveStringDraw(t *testing.T) {
	t.Run("should not panic if not resized yet", func(t *testing.T) {
		assert.NotPanics(t, func() {
			NewResponsiveString("", StringResponsiveConfig{}).Draw(term.NewStringWriter(10, 10))
		})
	})
	t.Run("should behave like String with single line strings and enough space", func(t *testing.T) {
		tcases := []struct {
			in  string
			out string
			cfg StringResponsiveConfig
		}{
			{
				in:  "aaaa",
				out: "aaaa \n     \n     \n     \n     ",
			},
			{
				in:  "X\nX\nX\nX\nX\nX\nX\nX\n",
				out: "X    \nX    \nX    \nX    \nX    ",
			},
			{
				in:  "a",
				out: "a    \n     \n     \n     \n     ",
			},
		}

		for _, tcase := range tcases {
			t.Run("StringResponsive", func(t *testing.T) {
				testString(t, func(in string) tui.Component {
					return NewResponsiveString(in, tcase.cfg)
				}, 5, 5, tcase.in, tcase.out)
			})
		}
	})

	t.Run("should wrap lines around", func(t *testing.T) {
		tcases := []struct {
			in     string
			out    string
			cfg    StringResponsiveConfig
			height int
		}{
			{
				in:  "XXXXXXXXXX\nBBBBBBBBBB\nCCCCCCCCCCCC\nDDDDDDDDDD\nEEEEEEEEEEE\nFFFFFFFFFF\n",
				out: "XXXXX\nXXXXX\nBBBBB\nBBBBB\nCCCCC",
			},
			{
				in:  "XXXXXXXXXXXXXXXX\nYYYYYYYYYnZZZZZZZZZ",
				out: "XXXXX\nXXXXX\nXXXXX\nX    \nYYYYY",
			},
			{
				in:  "X\n1111 \n222222222222222222222222222222222",
				out: "X    \n1111 \n22222\n22222\n22222",
			},
			{
				in: "X\n111\n222222222222222222222222222222222",
				out: `┌───┐
│X  │
│111│
│222│
└───┘`,
				cfg: StringResponsiveConfig{StringConfig: StringConfig{FrameCharSet: FrameCharSetDefault()}},
			},
			{
				in: "XXXX\n111\n22",
				out: `┌───┐
│XXX│
│X  │
│111│
└───┘`,
				cfg: StringResponsiveConfig{StringConfig: StringConfig{FrameCharSet: FrameCharSetDefault()}},
			},
			{
				in: "XXXXXXXXXX\nBBBBBBBBBB\nCCCCCCCCCCCC\nDDDDDDDDDD\nEEEEEEEEEEE\nFFFFFFFFFF",
				cfg: StringResponsiveConfig{
					StringConfig: StringConfig{
						Alignment:         AlignmentCentered,
						FrameCharSet:      FrameCharSetDefault(),
						PaddingVertical:   2,
						PaddingHorizontal: 2,
					},
				},
				height: 67,
				out: `┌───┐
│   │
│ X │
│ X │
│ X │
│ X │
│ X │
│ X │
│ X │
│ X │
│ X │
│ X │
│ B │
│ B │
│ B │
│ B │
│ B │
│ B │
│ B │
│ B │
│ B │
│ B │
│ C │
│ C │
│ C │
│ C │
│ C │
│ C │
│ C │
│ C │
│ C │
│ C │
│ C │
│ C │
│ D │
│ D │
│ D │
│ D │
│ D │
│ D │
│ D │
│ D │
│ D │
│ D │
│ E │
│ E │
│ E │
│ E │
│ E │
│ E │
│ E │
│ E │
│ E │
│ E │
│ E │
│ F │
│ F │
│ F │
│ F │
│ F │
│ F │
│ F │
│ F │
│ F │
│ F │
│   │
└───┘`,
			},
		}

		for _, tcase := range tcases {
			t.Run("StringResponsive", func(t *testing.T) {
				height := 5
				if tcase.height != 0 {
					height = tcase.height
				}
				testString(t, func(in string) tui.Component {
					return NewResponsiveString(in, tcase.cfg)
				}, 5, height, tcase.in, tcase.out)
			})
		}
	})

	t.Run("should not split words in half", func(t *testing.T) {
		tcases := []struct {
			in  string
			out string
			cfg StringResponsiveConfig
		}{
			{
				in:  "XX XX XX\nYYYYY YYYY\nZZZZZZZZZZZZ ZZZZZZZZZZZZ ZZZZZZZZZZZZ",
				out: "XX   \nXX XX\nYYYYY\n YYYY\nZZZZZ",
				cfg: StringResponsiveConfig{NoSplitWords: true},
			},
		}

		for _, tcase := range tcases {
			t.Run("StringResponsive", func(t *testing.T) {
				testString(t, func(in string) tui.Component {
					return NewResponsiveString(in, tcase.cfg)
				}, 5, 5, tcase.in, tcase.out)
			})
		}
	})
}

func TestResponsiveHeight(t *testing.T) {
	tcases := []struct {
		in    string
		width int
		out   int
		cfg   StringResponsiveConfig
	}{
		{
			in:    "XXXXXXXXXX\nBBBBBBBBBB\nCCCCCCCCCCCC\nDDDDDDDDDD\nEEEEEEEEEEE\nFFFFFFFFFF\n",
			width: 5,
			out:   15,
		},
		{
			in:    "X",
			width: 5,
			out:   1,
		},
		{
			in:    "X\nX\nX\nX\n",
			width: 10,
			out:   5,
		},
		{
			in:    "XXXXXXXXXX",
			width: 10,
			out:   1,
		},
		{
			in:    "XXXXXXXXXX",
			width: -1,
			out:   0,
		},
		{
			in:    "XXXXXXXXXX",
			width: 0, // could trigger division by zero
			out:   0,
		},
		{
			in:    "X\nX\nX\nX\n",
			width: 10,
			out:   7,
			cfg:   StringResponsiveConfig{StringConfig: StringConfig{FrameCharSet: FrameCharSetDefault()}},
		},
		{
			in:    "X",
			width: 5,
			out:   3,
			cfg:   StringResponsiveConfig{StringConfig: StringConfig{FrameCharSet: FrameCharSetDefault()}},
		},
		{
			in:    "XXXXXXXXXX",
			width: 10,
			out:   4,
			cfg:   StringResponsiveConfig{StringConfig: StringConfig{FrameCharSet: FrameCharSetDefault()}},
		},
		{
			in:    "XXXXXXXXXX\nBBBBBBBBBB\nCCCCCCCCCCCC\nDDDDDDDDDD\nEEEEEEEEEEE\nFFFFFFFFFF\n",
			width: 5,
			out:   27,
			cfg:   StringResponsiveConfig{StringConfig: StringConfig{FrameCharSet: FrameCharSetDefault()}},
		},
		{
			in:    "XXXXXXXXXX\nBBBBBBBBBB\nCCCCCCCCCCCC\nDDDDDDDDDD\nEEEEEEEEEEE\nFFFFFFFFFF\n",
			width: 5,
			out:   68,
			cfg: StringResponsiveConfig{
				StringConfig: StringConfig{
					Alignment:         AlignmentCentered,
					FrameCharSet:      FrameCharSetDefault(),
					PaddingVertical:   2,
					PaddingHorizontal: 2,
				},
			},
		},
	}

	for _, tcase := range tcases {
		t.Run("StringResponsive", func(t *testing.T) {
			s := NewResponsiveString(tcase.in, tcase.cfg)
			out := s.Height(tcase.width)
			assert.Equal(t, tcase.out, out)
		})
	}
}

func TestNopResponsive(t *testing.T) {
	t.Run("does not panic", func(t *testing.T) {
		w := term.NewStringWriter(5, 5)
		s := NopResponsive()
		s.Resize(4, 4)

		s.Draw(w)
		require.NoError(t, w.Flush())
		assert.Equal(t, "     \n     \n     \n     \n     ", w.String())
	})
}
