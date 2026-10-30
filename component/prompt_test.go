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
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/component/comptest"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func testDrawPrompt(
	t *testing.T, cfg PromptConfig,
	width, height int, expectedOut string,
) {
	t.Helper()
	s := NewPrompt(cfg)
	s.Resize(width, height)
	w := term.NewStringWriter(width, height)
	tests := []comptest.TestCase{
		{Expected: expectedOut},
	}
	comptest.TestComponent(t, s, w, tests)
}

func TestDrawPrompt(t *testing.T) {
	t.Run("no frame", func(t *testing.T) {
		testDrawPrompt(t, PromptConfig{
			Message: "Do you?",
			Options: []string{"Yay", "Nay"},
			// 20 wide, uniform width=3, gap=4, group=10, startX=5
		}, 20, 10,
			"                    \n"+
				"                    \n"+
				"      Do you?       \n"+
				"                    \n"+
				"                    \n"+
				"                    \n"+
				"                    \n"+
				"                    \n"+
				"                    \n"+
				"     Yay    Nay     ",
		)
	})

	t.Run("with frame", func(t *testing.T) {
		testDrawPrompt(t, PromptConfig{
			Message: "Do you?",
			Options: []string{"Yay", "Nay"},
			Frame:   FrameCharSetDefault(),
			// 20 wide, uniform width=7, gap=2, group=16, startX=2
		}, 20, 10,
			"                    \n"+
				"                    \n"+
				"      Do you?       \n"+
				"                    \n"+
				"                    \n"+
				"                    \n"+
				"                    \n"+
				"  ┌─────┐  ┌─────┐  \n"+
				"  │ Yay │  │ Nay │  \n"+
				"  └─────┘  └─────┘  ",
		)
	})

	t.Run("with frame three options", func(t *testing.T) {
		testDrawPrompt(t, PromptConfig{
			Message: "Why?",
			Options: []string{"Yay", "Nay", "Say"},
			Frame:   FrameCharSetDefault(),
			// 30 wide, uniform width=7, gaps=3+2, startX=2
		}, 30, 10,
			"                              \n"+
				"                              \n"+
				"             Why?             \n"+
				"                              \n"+
				"                              \n"+
				"                              \n"+
				"                              \n"+
				"  ┌─────┐   ┌─────┐  ┌─────┐  \n"+
				"  │ Yay │   │ Nay │  │ Say │  \n"+
				"  └─────┘   └─────┘  └─────┘  ",
		)
	})

	t.Run("with frame uniform width centering", func(t *testing.T) {
		testDrawPrompt(t, PromptConfig{
			Message: "Install?",
			Options: []string{"Go", "Skip"},
			Frame:   FrameCharSetDefault(),
		}, 25, 10,
			"                         \n"+
				"                         \n"+
				"        Install?         \n"+
				"                         \n"+
				"                         \n"+
				"                         \n"+
				"                         \n"+
				"   ┌──────┐   ┌──────┐   \n"+
				"   │  Go  │   │ Skip │   \n"+
				"   └──────┘   └──────┘   ",
		)
	})
}

func TestDrawPromptFourButtons(t *testing.T) {
	e80 := "                                                                                "
	// Message centered in 80 columns.
	m80 := "                    File is already open by another process.                    "

	// Uniform width = 15 (widest: "Open rdonly"=11+2pad+2frame).
	// 4 buttons of 15 = 60. Gap = (80-60)/5 = 4. Group = 72.
	// startX = (80-72)/2 = 4. All gaps equal to 4.
	testDrawPrompt(t, PromptConfig{
		Message: "File is already open by another process.",
		Options: []string{
			"Recover", "Open rdonly", "force Edit", "Skip",
		},
		Frame: FrameCharSetDefault(),
	}, 80, 10,
		e80+"\n"+e80+"\n"+m80+"\n"+
			e80+"\n"+e80+"\n"+e80+"\n"+e80+"\n"+
			"    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    \n"+
			"    │   Recover   │    │ Open rdonly │    │ force Edit  │    │    Skip     │    \n"+
			"    └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘    ",
	)
}

func TestDrawPromptWide(t *testing.T) {
	e := "                                                                                                    "
	// Message line: "Sure?" centered in 100 columns.
	m := "                                               Sure?                                                "

	t.Run("single button", func(t *testing.T) {
		// optCols = 12 (all columns), span = 100 wide
		// "Ok" frame (6) centered in 100: x=47
		testDrawPrompt(t, PromptConfig{
			Message: "Sure?",
			Options: []string{"Ok"},
			Frame:   FrameCharSetDefault(),
		}, 100, 10,
			e+"\n"+e+"\n"+m+"\n"+e+"\n"+e+"\n"+e+"\n"+e+"\n"+
				"                                               ┌────┐                                               \n"+
				"                                               │ Ok │                                               \n"+
				"                                               └────┘                                               ",
		)
	})

	t.Run("two buttons", func(t *testing.T) {
		// uniform width=10 (widest: "Cancel"), gap=26,
		// group=46, startX=27
		testDrawPrompt(t, PromptConfig{
			Message: "Sure?",
			Options: []string{"Ok", "Cancel"},
			Frame:   FrameCharSetDefault(),
		}, 100, 10,
			e+"\n"+e+"\n"+m+"\n"+e+"\n"+e+"\n"+e+"\n"+e+"\n"+
				"                           ┌────────┐                          ┌────────┐                           \n"+
				"                           │   Ok   │                          │ Cancel │                           \n"+
				"                           └────────┘                          └────────┘                           ",
		)
	})

	t.Run("three buttons", func(t *testing.T) {
		// uniform width=10 (widest: "Cancel"), gap=17,
		// group=64, startX=18
		testDrawPrompt(t, PromptConfig{
			Message: "Sure?",
			Options: []string{"Ok", "Cancel", "Help"},
			Frame:   FrameCharSetDefault(),
		}, 100, 10,
			e+"\n"+e+"\n"+m+"\n"+e+"\n"+e+"\n"+e+"\n"+e+"\n"+
				"                  ┌────────┐                 ┌────────┐                 ┌────────┐                  \n"+
				"                  │   Ok   │                 │ Cancel │                 │  Help  │                  \n"+
				"                  └────────┘                 └────────┘                 └────────┘                  ",
		)
	})
}

func TestDrawPromptWraps(t *testing.T) {
	providers := []string{"OpenAI", "Anthropic", "Gemini", "Codex", "Claude"}
	tests := []struct {
		name     string
		cfg      PromptConfig
		width    int
		height   int
		expected string
	}{
		{
			name:  "spreads the options evenly over rows a blank row apart",
			cfg:   PromptConfig{Message: "Pick?", Options: providers},
			width: 40, height: 8,
			expected: "                                        \n" +
				"                                        \n" +
				"                 Pick?                  \n" +
				"                                        \n" +
				"                                        \n" +
				"    OpenAI      Anthropic    Gemini     \n" +
				"                                        \n" +
				"           Codex       Claude           ",
		},
		{
			name:  "packs the rows rather than scroll the message",
			cfg:   PromptConfig{Message: "Pick?", Options: providers},
			width: 40, height: 7,
			expected: "                                        \n" +
				"                                        \n" +
				"                 Pick?                  \n" +
				"                                        \n" +
				"                                        \n" +
				"    OpenAI      Anthropic    Gemini     \n" +
				"           Codex       Claude           ",
		},
		{
			name:  "scrolls the message once the rows are packed",
			cfg:   PromptConfig{Message: "Pick?", Options: providers},
			width: 40, height: 6,
			expected: "                                        \n" +
				"                                        \n" +
				"                 Pick?                  \n" +
				"                                        \n" +
				"    OpenAI      Anthropic    Gemini     \n" +
				"           Codex       Claude           ",
		},
		{
			name:  "takes as many rows as the width needs",
			cfg:   PromptConfig{Message: "Pick?", Options: providers},
			width: 24, height: 6,
			expected: "                        \n" +
				"                        \n" +
				"         Pick?          \n" +
				"   OpenAI    Anthropic  \n" +
				"   Gemini      Codex    \n" +
				"        Claude          ",
		},
		{
			name: "wraps framed options a blank row apart",
			cfg: PromptConfig{
				Message: "Pick?",
				Options: []string{"Yay", "Nay", "Say"},
				Frame:   FrameCharSetDefault(),
			},
			width: 20, height: 12,
			expected: "                    \n" +
				"                    \n" +
				"       Pick?        \n" +
				"                    \n" +
				"                    \n" +
				"  ┌─────┐  ┌─────┐  \n" +
				"  │ Yay │  │ Nay │  \n" +
				"  └─────┘  └─────┘  \n" +
				"                    \n" +
				"      ┌─────┐       \n" +
				"      │ Say │       \n" +
				"      └─────┘       ",
		},
		{
			name: "packs framed options when short of room",
			cfg: PromptConfig{
				Message: "Pick?",
				Options: []string{"Yay", "Nay", "Say"},
				Frame:   FrameCharSetDefault(),
			},
			width: 20, height: 10,
			expected: "                    \n" +
				"                    \n" +
				"       Pick?        \n" +
				"                    \n" +
				"  ┌─────┐  ┌─────┐  \n" +
				"  │ Yay │  │ Nay │  \n" +
				"  └─────┘  └─────┘  \n" +
				"      ┌─────┐       \n" +
				"      │ Say │       \n" +
				"      └─────┘       ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDrawPrompt(t, tt.cfg, tt.width, tt.height, tt.expected)
		})
	}
}

func TestPromptHeight(t *testing.T) {
	yesNo := func(cfg func(...string) PromptConfig) PromptConfig { return cfg("Yes", "No") }
	fixed3 := func(options ...string) PromptConfig { return fixedPrompt(3, options...) }
	tests := []struct {
		name  string
		cfg   PromptConfig
		width int
		want  int
	}{
		{"one row of options under the message", yesNo(fixed3), 12, 3 + 1},
		{"one row at the narrowest width leaving a cell around each option", yesNo(fixed3), 9, 3 + 1},
		{"a row per option a blank row apart a cell narrower", yesNo(fixed3), 8, 3 + 2 + 1},
		{"a row per option when no two fit side by side", yesNo(fixed3), 2, 3 + 2 + 1},
		{"a row per option at zero width", yesNo(fixed3), 0, 3 + 2 + 1},
		{"a row per option at a negative width", yesNo(fixed3), -7, 3 + 2 + 1},
		{"a single option takes no blank row", fixed3("Ok"), 1, 3 + 1},
		{"a single option at zero width", fixed3("Ok"), 0, 3 + 1},
		{"every option on one row", fixed3(providerOptions...), 51, 3 + 1},
		{"four to a row over two rows a cell narrower", fixed3(providerOptions...), 50, 3 + 2 + 1},
		{"three to a row over two rows", fixed3(providerOptions...), 40, 3 + 2 + 1},
		{"two to a row over three rows", fixed3(providerOptions...), 30, 3 + 3 + 2},
		{"one to a row over five rows", fixed3(providerOptions...), 20, 3 + 5 + 4},
		{"framed options on one row", framed(fixed3("Yay", "Nay", "Say")), 25, 3 + 3},
		{"framed options over two rows", framed(fixed3("Yay", "Nay", "Say")), 24, 3 + 2*3 + 1},
		{"framed options a row each", framed(fixed3("Yay", "Nay", "Say")), 16, 3 + 3*3 + 2},
		{"the message's height at the width", linesPrompt([]string{"abcdef"}, "Yes", "No"), 12, 1 + 1},
		{"the message reflowed taller by a narrower width", linesPrompt([]string{"abcdef"}, "Yes", "No"), 3, 2 + 2 + 1},
		{"a message that is only Floating takes its ideal height", yesNo(idealPrompt), 12, 2 + 1},
		{"a message that is only Floating at any width", yesNo(idealPrompt), 2, 2 + 2 + 1},
		{"a message with no height", yesNo(nopPrompt), 12, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewPrompt(tt.cfg).Height(tt.width))
		})
	}

	t.Run("asks the message for its height at the width", func(t *testing.T) {
		for _, width := range []int{-3, 0, 1, 7, 40} {
			msg := &TestResponsive{WantHeight: 3}
			p := NewPrompt(PromptConfig{
				Message:    "?",
				Options:    []string{"Yes", "No"},
				NewMessage: func(string) Floating { return msg },
			})
			p.Height(width)
			assert.Equal(t, width, msg.PassedWidth, "width=%d", width)
		}
	})

	t.Run("is the same whatever the last Resize", func(t *testing.T) {
		widths := []int{-1, 0, 5, 9, 24, 40, 51, 100}
		fresh := make(map[int]int, len(widths))
		for _, width := range widths {
			fresh[width] = NewPrompt(PromptConfig{Message: "Pick?", Options: providerOptions}).Height(width)
		}
		for _, size := range [][2]int{{0, 0}, {-1, -1}, {5, 1}, {24, 6}, {40, 8}, {100, 50}} {
			p := NewPrompt(PromptConfig{Message: "Pick?", Options: providerOptions})
			p.Resize(size[0], size[1])
			for p.SeekDown() {
			}
			for _, width := range widths {
				assert.Equal(t, fresh[width], p.Height(width),
					"width=%d after Resize(%d, %d)", width, size[0], size[1])
			}
		}
	})

	t.Run("leaves the last layout alone", func(t *testing.T) {
		p := NewPrompt(fixedPrompt(3, providerOptions...))
		p.Resize(40, 4)
		p.SeekDown()
		before := optionCells(p, 40, 4)
		offset, maxOffset := p.SeekOffset(), p.MaxSeekOffset()
		require.Positive(t, offset)
		for _, width := range []int{-1, 0, 5, 24, 51, 100} {
			p.Height(width)
		}
		assert.Equal(t, before, optionCells(p, 40, 4))
		assert.Equal(t, offset, p.SeekOffset())
		assert.Equal(t, maxOffset, p.MaxSeekOffset())
	})
}

func TestPromptSeek(t *testing.T) {
	const (
		opDown = iota
		opUp
		opResize
		opFit
		opInit
	)
	type step struct {
		op            int
		width, height int
		want          bool
		offset, max   int
	}
	down := func(want bool, offset, max int) step {
		return step{op: opDown, want: want, offset: offset, max: max}
	}
	up := func(want bool, offset, max int) step {
		return step{op: opUp, want: want, offset: offset, max: max}
	}
	resize := func(width, height, offset, max int) step {
		return step{op: opResize, width: width, height: height, offset: offset, max: max}
	}
	// fit resizes to the height the prompt asks for at width, plus extra.
	fit := func(width, extra, offset, max int) step {
		return step{op: opFit, width: width, height: extra, offset: offset, max: max}
	}
	reinit := func(offset, max int) step {
		return step{op: opInit, offset: offset, max: max}
	}
	// The four lines take four rows at a width of five or more and six
	// at three. "Yes" and "No" share a row from a width of nine and
	// take a row each below that.
	four := linesPrompt(fourLines, "Yes", "No")

	tests := []struct {
		name  string
		cfg   PromptConfig
		steps []step
	}{
		{
			name:  "nothing to scroll before the first Resize",
			cfg:   four,
			steps: []step{down(false, 0, 0), up(false, 0, 0)},
		},
		{
			name:  "nothing to scroll with room to spare",
			cfg:   four,
			steps: []step{resize(12, 10, 0, 0), down(false, 0, 0), up(false, 0, 0)},
		},
		{
			name:  "nothing to scroll with exactly the room it asks for",
			cfg:   four,
			steps: []step{resize(12, 5, 0, 0), down(false, 0, 0), up(false, 0, 0)},
		},
		{
			name: "a row short scrolls by a row",
			cfg:  four,
			steps: []step{
				resize(12, 4, 0, 1),
				down(true, 1, 1), down(false, 1, 1),
				up(true, 0, 1), up(false, 0, 1),
			},
		},
		{
			name: "a row at a time to the last row and back",
			cfg:  four,
			steps: []step{
				resize(12, 3, 0, 2),
				down(true, 1, 2), down(true, 2, 2), down(false, 2, 2),
				up(true, 1, 2), up(true, 0, 2), up(false, 0, 2),
			},
		},
		{
			name:  "up at the top stays at the top",
			cfg:   four,
			steps: []step{resize(12, 3, 0, 2), up(false, 0, 2), up(false, 0, 2)},
		},
		{
			name:  "wrapped options give up their blank row before the message scrolls",
			cfg:   four,
			steps: []step{resize(8, 6, 0, 0), down(false, 0, 0)},
		},
		{
			name:  "the message scrolls once wrapped options are packed",
			cfg:   four,
			steps: []step{resize(8, 5, 0, 1), down(true, 1, 1), down(false, 1, 1)},
		},
		{
			name: "the whole message scrolls when the options leave it no row",
			cfg:  four,
			steps: []step{
				resize(12, 1, 0, 4),
				down(true, 1, 4), down(true, 2, 4), down(true, 3, 4), down(true, 4, 4),
				down(false, 4, 4),
			},
		},
		{
			name:  "options taller than the prompt leave the message no row",
			cfg:   four,
			steps: []step{resize(8, 1, 0, 4), down(true, 1, 4)},
		},
		{
			name:  "zero height",
			cfg:   four,
			steps: []step{resize(12, 0, 0, 4), down(true, 1, 4)},
		},
		{
			name:  "negative height",
			cfg:   four,
			steps: []step{resize(12, -3, 0, 4), down(true, 1, 4)},
		},
		{
			name:  "zero width",
			cfg:   four,
			steps: []step{resize(0, 3, 0, 3), down(true, 1, 3)},
		},
		{
			name:  "negative width",
			cfg:   four,
			steps: []step{resize(-4, 3, 0, 3), down(true, 1, 3)},
		},
		{
			name: "growing pulls the offset back to the rows still hidden",
			cfg:  four,
			steps: []step{
				resize(12, 3, 0, 2), down(true, 1, 2), down(true, 2, 2),
				resize(12, 4, 1, 1),
				resize(12, 10, 0, 0), down(false, 0, 0), up(false, 0, 0),
			},
		},
		{
			name: "shrinking keeps the offset",
			cfg:  four,
			steps: []step{
				resize(12, 4, 0, 1), down(true, 1, 1),
				resize(12, 3, 1, 2), down(true, 2, 2),
			},
		},
		{
			name: "narrowing reflows the message taller and keeps the offset",
			cfg:  four,
			steps: []step{
				resize(12, 3, 0, 2), down(true, 1, 2),
				resize(3, 3, 1, 5), down(true, 2, 5),
			},
		},
		{
			name: "widening unwraps the options and gives the message their rows",
			cfg:  four,
			steps: []step{
				resize(8, 5, 0, 1), down(true, 1, 1),
				resize(12, 5, 0, 0), up(false, 0, 0),
			},
		},
		{
			name: "Init forgets the scroll until the next Resize",
			cfg:  four,
			steps: []step{
				resize(12, 3, 0, 2), down(true, 1, 2), down(true, 2, 2),
				reinit(0, 0), down(false, 0, 0), up(false, 0, 0),
				resize(12, 3, 0, 2), down(true, 1, 2),
			},
		},
		{
			name: "framed options take three rows each",
			cfg:  framed(four),
			steps: []step{
				resize(20, 7, 0, 0),
				resize(20, 4, 0, 3), down(true, 1, 3),
				resize(12, 10, 0, 0),
				resize(12, 9, 0, 1),
			},
		},
		{
			name: "the default message at the height it asks for",
			cfg:  PromptConfig{Message: "Are you sure?", Options: []string{"Yes", "No"}},
			steps: []step{
				fit(20, 0, 0, 0), down(false, 0, 0),
				fit(20, 5, 0, 0),
				fit(20, -1, 0, 1), down(true, 1, 1), down(false, 1, 1),
			},
		},
		{
			name: "the default message over options wrapped onto three rows",
			cfg:  PromptConfig{Message: "Pick a provider", Options: providerOptions},
			steps: []step{
				fit(24, 0, 0, 0),
				fit(24, -2, 0, 0),
				fit(24, -3, 0, 1), down(true, 1, 1),
			},
		},
		{
			name:  "a message that is only Floating scrolls its ideal height",
			cfg:   idealPrompt("Yes", "No"),
			steps: []step{resize(12, 2, 0, 1), down(true, 1, 1), down(false, 1, 1)},
		},
		{
			name:  "a message with no height never scrolls",
			cfg:   nopPrompt("Yes", "No"),
			steps: []step{resize(12, 0, 0, 0), down(false, 0, 0), up(false, 0, 0)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPrompt(tt.cfg)
			for i, s := range tt.steps {
				switch s.op {
				case opDown:
					assert.Equal(t, s.want, p.SeekDown(), "step %d: SeekDown", i)
				case opUp:
					assert.Equal(t, s.want, p.SeekUp(), "step %d: SeekUp", i)
				case opResize:
					p.Resize(s.width, s.height)
				case opFit:
					p.Resize(s.width, p.Height(s.width)+s.height)
				case opInit:
					p.Init(tt.cfg)
				}
				assert.Equal(t, s.offset, p.SeekOffset(), "step %d: SeekOffset", i)
				assert.Equal(t, s.max, p.MaxSeekOffset(), "step %d: MaxSeekOffset", i)
			}
		})
	}
}

func TestDrawPromptScrolled(t *testing.T) {
	four := linesPrompt(fourLines, "Yes", "No")
	tests := []struct {
		name          string
		cfg           PromptConfig
		width, height int
		down          int
		expected      string
	}{
		{
			name:  "the first rows of the message over the options",
			cfg:   four,
			width: 12, height: 3,
			expected: "one         \n" +
				"two         \n" +
				"  Yes  No   ",
		},
		{
			name:  "a row down",
			cfg:   four,
			width: 12, height: 3, down: 1,
			expected: "two         \n" +
				"three       \n" +
				"  Yes  No   ",
		},
		{
			name:  "the last rows",
			cfg:   four,
			width: 12, height: 3, down: 2,
			expected: "three       \n" +
				"four        \n" +
				"  Yes  No   ",
		},
		{
			name:  "never past the last row",
			cfg:   four,
			width: 12, height: 3, down: 9,
			expected: "three       \n" +
				"four        \n" +
				"  Yes  No   ",
		},
		{
			name:     "the options alone when they leave the message no row",
			cfg:      four,
			width:    12,
			height:   1,
			expected: "  Yes  No   ",
		},
		{
			name:     "the options alone however far the message is scrolled",
			cfg:      four,
			width:    12,
			height:   1,
			down:     2,
			expected: "  Yes  No   ",
		},
		{
			name:  "packed wrapped options under a scrolled message",
			cfg:   four,
			width: 8, height: 4, down: 1,
			expected: "two     \n" +
				"three   \n" +
				"  Yes   \n" +
				"  No    ",
		},
		{
			name:  "a reflowed message scrolled over wrapped options",
			cfg:   four,
			width: 3, height: 3, down: 2,
			expected: "thr\n" +
				"Yes\n" +
				"No ",
		},
		{
			name:  "framed options under a scrolled message",
			cfg:   framed(four),
			width: 20, height: 4, down: 1,
			expected: "two                 \n" +
				"  ┌─────┐  ┌─────┐  \n" +
				"  │ Yes │  │ No  │  \n" +
				"  └─────┘  └─────┘  ",
		},
		{
			name:  "framed options cut to a prompt shorter than them",
			cfg:   framed(four),
			width: 20, height: 2,
			expected: "  ┌─────┐  ┌─────┐  \n" +
				"  │ Yes │  │ No  │  ",
		},
		{
			name:     "the rows of options a short prompt has room for",
			cfg:      four,
			width:    8,
			height:   1,
			expected: "  Yes   ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPrompt(tt.cfg)
			p.Resize(tt.width, tt.height)
			for range tt.down {
				p.SeekDown()
			}
			comptest.TestComponent(t, p, term.NewStringWriter(tt.width, tt.height),
				[]comptest.TestCase{{Expected: tt.expected}})
		})
	}
}

func TestPromptDrawsWithinItsBounds(t *testing.T) {
	configs := []struct {
		name string
		cfg  PromptConfig
	}{
		{"default message", PromptConfig{Message: "Are you sure?", Options: []string{"Yes", "No"}}},
		{"default message, framed", framed(PromptConfig{Message: "Are you sure?", Options: []string{"Yes", "No"}})},
		{"default message, wrapping options", PromptConfig{Message: "Pick?", Options: providerOptions}},
		{"lines, framed options", framed(linesPrompt(fourLines, "Yay", "Nay", "Say"))},
		{"lines", linesPrompt(fourLines, "Yes", "No")},
		{"only Floating message", idealPrompt("Yes", "No")},
		{"message with no height", nopPrompt("Yes", "No")},
	}
	sizes := [][2]int{
		{-1, -1}, {0, 0}, {0, 5}, {5, 0}, {1, 1}, {2, 3}, {8, 1},
		{12, 0}, {12, 1}, {20, 2}, {20, 4}, {24, 6}, {40, 8}, {80, 20},
	}
	for _, c := range configs {
		t.Run(c.name+", before any Resize", func(t *testing.T) {
			w := &outOfBoundsWriter{}
			assert.NotPanics(t, func() { NewPrompt(c.cfg).Draw(w) })
			assert.Empty(t, w.outside)
		})
		for _, size := range sizes {
			for _, scrolled := range []bool{false, true} {
				name := fmt.Sprintf("%s, %dx%d, scrolled=%t", c.name, size[0], size[1], scrolled)
				t.Run(name, func(t *testing.T) {
					p := NewPrompt(c.cfg)
					p.Resize(size[0], size[1])
					for scrolled && p.SeekDown() {
					}
					w := &outOfBoundsWriter{width: size[0], height: size[1]}
					p.Draw(w)
					assert.Empty(t, w.outside)
				})
			}
		}
	}
}

func TestPromptOptionAt(t *testing.T) {
	four := linesPrompt(fourLines, "Yes", "No")
	at := func(width, height int) func(*Prompt) {
		return func(p *Prompt) { p.Resize(width, height) }
	}
	tests := []struct {
		name    string
		cfg     PromptConfig
		prepare func(*Prompt)
		x, y    int
		want    int
	}{
		// "Yes" at x=2..4 and "No" at x=7..9 on the bottom row.
		{"the first cell of an option", four, at(12, 5), 2, 4, 0},
		{"the last cell of an option", four, at(12, 5), 4, 4, 0},
		{"the gap after an option", four, at(12, 5), 5, 4, -1},
		{"the first cell of the next option", four, at(12, 5), 7, 4, 1},
		{"the cell a short label leaves blank", four, at(12, 5), 9, 4, 1},
		{"the margin right of the options", four, at(12, 5), 10, 4, -1},
		{"the margin left of the options", four, at(12, 5), 1, 4, -1},
		{"the message above an option", four, at(12, 5), 2, 3, -1},
		{"below the prompt", four, at(12, 5), 2, 5, -1},
		{"right of the prompt", four, at(12, 5), 12, 4, -1},
		{"a negative column", four, at(12, 5), -1, 4, -1},
		{"a negative row", four, at(12, 5), 2, -1, -1},
		// A row each, a blank row apart: "Yes" on row 4, "No" on 6.
		{"the first of rows a blank row apart", four, at(8, 7), 2, 4, 0},
		{"the blank row between rows", four, at(8, 7), 2, 5, -1},
		{"the second of rows a blank row apart", four, at(8, 7), 2, 6, 1},
		{"the cell right of a wrapped option", four, at(8, 7), 5, 6, -1},
		// Packed: "Yes" on row 4, "No" on 5.
		{"the first of packed rows", four, at(8, 6), 2, 4, 0},
		{"the second of packed rows", four, at(8, 6), 2, 5, 1},
		{"the row of options a short prompt shows", four, at(8, 1), 2, 0, 0},
		{"a row of options the prompt is too short for", four, at(8, 1), 2, 1, -1},
		{"a zero height prompt", four, at(12, 0), 2, 0, -1},
		// Framed options three rows tall, cut to the prompt's two.
		{"the top row of framed options", framed(four), at(20, 2), 3, 0, 0},
		{"the second row of framed options", framed(four), at(20, 2), 3, 1, 0},
		{"the row of framed options past the prompt", framed(four), at(20, 2), 3, 2, -1},
		// Nine-cell options cut to a five-cell prompt, a row each.
		{"the last cell of an option cut to the width", nopPrompt(providerOptions...), at(5, 9), 4, 0, 0},
		{"past the width an option is cut to", nopPrompt(providerOptions...), at(5, 9), 5, 0, -1},
		{"the last of five rows", nopPrompt(providerOptions...), at(5, 9), 0, 8, 4},
		{"the blank row above the last", nopPrompt(providerOptions...), at(5, 9), 0, 7, -1},
		// Three to a row and two centered under them.
		{"first row, first option", nopPrompt(providerOptions...), at(40, 2), 3, 0, 0},
		{"first row, last cell of an option", nopPrompt(providerOptions...), at(40, 2), 24, 0, 1},
		{"first row, last option", nopPrompt(providerOptions...), at(40, 2), 28, 0, 2},
		{"first row, gap between options", nopPrompt(providerOptions...), at(40, 2), 13, 0, -1},
		{"second row, first option", nopPrompt(providerOptions...), at(40, 2), 9, 1, 3},
		{"second row, last option", nopPrompt(providerOptions...), at(40, 2), 30, 1, 4},
		{"second row, margin of a centered row", nopPrompt(providerOptions...), at(40, 2), 3, 1, -1},
		{"before any Resize", four, func(*Prompt) {}, 0, 0, -1},
		{"where an option was before Init", four, func(p *Prompt) {
			p.Resize(12, 5)
			p.Init(four)
		}, 2, 4, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPrompt(tt.cfg)
			tt.prepare(p)
			assert.Equal(t, tt.want, p.OptionAt(tt.x, tt.y))
		})
	}
}

func TestPromptInitForgetsTheLastLayout(t *testing.T) {
	four := linesPrompt(fourLines, "Yes", "No")
	tests := []struct {
		name    string
		prepare func(*Prompt)
		reinit  PromptConfig
		// options are those OptionAt finds once resized again.
		options []int
	}{
		{"laid out", func(p *Prompt) { p.Resize(12, 5) }, four, []int{0, 1}},
		{"scrolled", func(p *Prompt) { p.Resize(12, 3); p.SeekDown() }, four, []int{0, 1}},
		{"wrapped", func(p *Prompt) { p.Resize(8, 7) }, four, []int{0, 1}},
		{"with more options", func(p *Prompt) { p.Resize(12, 5) },
			linesPrompt(fourLines, "A", "B", "C"), []int{0, 1, 2}},
		{"with fewer options", func(p *Prompt) { p.Resize(12, 5) },
			linesPrompt(fourLines, "Ok"), []int{0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPrompt(four)
			tt.prepare(p)
			p.Init(tt.reinit)

			assert.Zero(t, p.SeekOffset())
			assert.Zero(t, p.MaxSeekOffset())
			assert.Empty(t, optionCells(p, 12, 7), "no option is hit before the next Resize")
			w := term.NewStringWriter(12, 7)
			assert.NotPanics(t, func() { p.Draw(w) })
			require.NoError(t, w.Flush())
			assert.Equal(t,
				strings.TrimSuffix(strings.Repeat(strings.Repeat(" ", 12)+"\n", 7), "\n"),
				w.String(), "nothing of the old layout is drawn")

			p.Resize(12, 5)
			assert.Equal(t, tt.options, optionsHit(p, 12, 5))
		})
	}
}

func TestPromptSetOptionAttr(t *testing.T) {
	t.Run("does not panic with frame", func(t *testing.T) {
		p := NewPrompt(PromptConfig{
			Message: "Wasup?",
			Options: []string{"Meh", "Bleh"},
			Frame:   FrameCharSetDefault(),
		})
		p.SetOptionAttr(0, term.Attributes{})
		p.SetOptionAttr(1, term.Attributes{})
	})

	t.Run("does not panic without frame", func(t *testing.T) {
		p := NewPrompt(PromptConfig{
			Message: "Wasup?",
			Options: []string{"Meh", "Bleh"},
		})
		p.SetOptionAttr(0, term.Attributes{})
		p.SetOptionAttr(1, term.Attributes{})
	})
}

func TestPromptDefaults(t *testing.T) {
	t.Run("panics on empty message", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = NewPrompt(PromptConfig{
				Message: "", Options: []string{"a"},
			})
		})
	})
	t.Run("panics on empty options", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = NewPrompt(PromptConfig{
				Message: "blah", Options: []string{},
			})
		})
	})
}

func TestPromptInitReset(t *testing.T) {
	var p Prompt
	opts := make(map[string]*TestResponsive)
	makeTestOption := func(
		opt string, cfg PromptConfig,
	) floatingOption {
		tr := &TestResponsive{
			TestComponent: TestComponent{
				Ch: ([]rune)(opt)[0],
			},
			WantHeight: 3,
			WantWidth:  6,
		}
		opts[opt] = tr
		return tr
	}
	p.init(makeTestOption, PromptConfig{
		Message: "?",
		Options: []string{"Y", "N"},
	})
	p.init(makeTestOption, PromptConfig{
		Message: "?!",
		Options: []string{"y", "n"},
	})

	t.Run("Draw", func(t *testing.T) {
		p.Resize(20, 10)
		w := term.NewStringWriter(21, 11)

		tests := []comptest.TestCase{
			{Expected: "                     \n" +
				"                     \n" +
				"         ?!          \n" +
				"                     \n" +
				"                     \n" +
				"                     \n" +
				"                     \n" +
				"   yyyyyy  nnnnnn    \n" +
				"   yyyyyy  nnnnnn    \n" +
				"   yyyyyy  nnnnnn    \n" +
				"                     ",
			},
		}
		comptest.TestComponent(t, &p, w, tests)
	})

	t.Run("SetAttr", func(t *testing.T) {
		attr := term.Attributes{
			Fg:    term.ColorRed,
			Attrs: term.AttrBold,
		}
		p.SetOptionAttr(0, attr)
		assert.Equal(t, attr, opts["y"].Attributes)
	})
}

func TestPromptButtonBackgroundAttributes(t *testing.T) {
	focus := term.Attributes{Bg: term.ColorBlue}

	tests := []struct {
		name     string
		cfg      PromptConfig
		width    int
		height   int
		focus    int
		expected string
	}{
		{
			name: "framed button interior filled on focus",
			cfg: PromptConfig{
				Message: "Do you?",
				Options: []string{"Yay", "Nay"},
				Frame:   FrameCharSetDefault(),
			},
			width: 20, height: 10,
			focus: 0,
			expected: "                    \n" +
				"                    \n" +
				"      Do you?       \n" +
				"                    \n" +
				"                    \n" +
				"                    \n" +
				"                    \n" +
				"  ███████  ┌─────┐  \n" +
				"  ███████  │ Nay │  \n" +
				"  ███████  └─────┘  ",
		},
		{
			name: "frameless button interior filled on focus",
			cfg: PromptConfig{
				Message: "Do you?",
				Options: []string{"Go", "Skip"},
			},
			width: 20, height: 10,
			focus: 0,
			expected: "                    \n" +
				"                    \n" +
				"      Do you?       \n" +
				"                    \n" +
				"                    \n" +
				"                    \n" +
				"                    \n" +
				"                    \n" +
				"                    \n" +
				"    ████    Skip    ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPrompt(tt.cfg)
			p.Resize(tt.width, tt.height)
			w := term.NewStringWriter(tt.width, tt.height)
			w.BackgroundCh = '█'
			comptest.TestComponent(t, p, w, []comptest.TestCase{
				{
					Action: func() {
						p.SetOptionAttr(tt.focus, focus)
					},
					Expected: tt.expected,
				},
			})
		})
	}

	e20 := "                    "
	msg := "      Do you?       "

	t.Run("switching highlight between framed options",
		func(t *testing.T) {
			width, height := 20, 10
			normal := term.Attributes{}
			p := NewPrompt(PromptConfig{
				Message: "Do you?",
				Options: []string{"Yay", "Nay"},
				Frame:   FrameCharSetDefault(),
			})
			p.Resize(width, height)
			w := term.NewStringWriter(width, height)
			w.BackgroundCh = '█'
			comptest.TestComponent(t, p, w, []comptest.TestCase{
				{
					Action: func() {
						p.SetOptionAttr(0, focus)
						p.SetOptionAttr(1, normal)
					},
					Expected: e20 + "\n" + e20 + "\n" +
						msg + "\n" +
						e20 + "\n" + e20 + "\n" +
						e20 + "\n" + e20 + "\n" +
						"  ███████  ┌─────┐  \n" +
						"  ███████  │ Nay │  \n" +
						"  ███████  └─────┘  ",
				},
				{
					Action: func() {
						p.SetOptionAttr(0, normal)
						p.SetOptionAttr(1, focus)
					},
					Expected: e20 + "\n" + e20 + "\n" +
						msg + "\n" +
						e20 + "\n" + e20 + "\n" +
						e20 + "\n" + e20 + "\n" +
						"  ┌─────┐  ███████  \n" +
						"  │ Yay │  ███████  \n" +
						"  └─────┘  ███████  ",
				},
			})
		})

	t.Run("AttrReverse highlight uses Fg for interior",
		func(t *testing.T) {
			width, height := 20, 10
			optionAttr := term.Attributes{}
			highlightAttr := term.Attributes{
				Fg:    term.ColorWhite,
				Attrs: term.AttrReverse,
			}
			p := NewPrompt(PromptConfig{
				Message: "Do you?",
				Options: []string{"Yay", "Nay"},
				Frame:   FrameCharSetDefault(),
			})
			p.Resize(width, height)
			w := term.NewStringWriter(width, height)
			w.BackgroundCh = '█'
			// Frame borders and text are drawn with SetCell
			// (Bg=0, AttrReverse), overriding the UnionAttributes
			// fill. Only the padding cells between the frame and
			// text keep the Bg from UnionAttributes.
			comptest.TestComponent(t, p, w, []comptest.TestCase{
				{
					Action: func() {
						p.SetOptionAttr(0, highlightAttr)
						p.SetOptionAttr(1, optionAttr)
					},
					Expected: e20 + "\n" + e20 + "\n" +
						msg + "\n" +
						e20 + "\n" + e20 + "\n" +
						e20 + "\n" + e20 + "\n" +
						"  ┌─────┐  ┌─────┐  \n" +
						"  │█Yay█│  │ Nay │  \n" +
						"  └─────┘  └─────┘  ",
				},
				{
					Action: func() {
						p.SetOptionAttr(0, optionAttr)
						p.SetOptionAttr(1, highlightAttr)
					},
					Expected: e20 + "\n" + e20 + "\n" +
						msg + "\n" +
						e20 + "\n" + e20 + "\n" +
						e20 + "\n" + e20 + "\n" +
						"  ┌─────┐  ┌─────┐  \n" +
						"  │ Yay │  │█Nay█│  \n" +
						"  └─────┘  └─────┘  ",
				},
			})
		})
}

func TestPromptDimensions(t *testing.T) {
	t.Run("includes option width with frame", func(t *testing.T) {
		p := NewPrompt(PromptConfig{
			Message: "Ok?",
			Options: []string{"Yes", "No"},
			Frame:   FrameCharSetDefault(),
		})
		w, h := p.Dimensions()
		// "Yes" = 3+2pad+2frame = 7
		// "No" = 2+2pad+2frame = 6
		// total = 7+6 = 13
		assert.GreaterOrEqual(t, w, 13)
		assert.Greater(t, h, 0)
	})

	t.Run("equal span width for different-width buttons", func(t *testing.T) {
		p := NewPrompt(PromptConfig{
			Message: "Recover?",
			Options: []string{"Recover", "Open rdonly", "force Edit", "Skip"},
			Frame:   FrameCharSetDefault(),
		})
		w, _ := p.Dimensions()
		// Widest button is "Open rdonly" = 11+2pad+2frame = 15.
		// 4 buttons × 15 = 60 + 2-cell gaps × 5 = 10 → 70.
		// Parity: (70-60) % 5 = 0 (even), no adjustment.
		assert.Equal(t, 70, w)
	})

	t.Run("respects MinWidth", func(t *testing.T) {
		p := NewPrompt(PromptConfig{
			Message:  "Ok?",
			Options:  []string{"Y"},
			MinWidth: 50,
		})
		w, _ := p.Dimensions()
		// n=1, maxW=1. optWidth = 1+2*2 = 5. MinWidth=50 wins.
		// Parity: (50-1) % 2 = 1 (odd), bumped to 51.
		assert.Equal(t, 51, w)
	})

	t.Run("parity adjustment for symmetric centering", func(t *testing.T) {
		p := NewPrompt(PromptConfig{
			Message: "?",
			Options: []string{"Aa", "Bb", "Cc"},
			Frame:   FrameCharSetDefault(),
		})
		w, _ := p.Dimensions()
		// n=3, maxW=6 (2+2pad+2frame). optWidth=18+8=26.
		// Parity: (26-18) % 4 = 0 (even), no adjustment.
		assert.Equal(t, 26, w)
		// Verify centering is symmetric at this width:
		// innerGap=(26-18)/4=2, group=18+4=22, margin=(26-22)/2=2.
		// left == right == 2. ✓
	})
}

var (
	fourLines       = []string{"one", "two", "three", "four"}
	providerOptions = []string{"OpenAI", "Anthropic", "Gemini", "Codex", "Claude"}
)

func TestPromptButtonImagesLayout(t *testing.T) {
	images := &ButtonImages{
		Idle:   testButtonImage("idle"),
		Outset: ButtonOutset{Top: 1, Bottom: 1, Sides: 2},
	}
	tests := []struct {
		name       string
		images     *ButtonImages
		width      int
		wantHeight int
	}{
		// Message 2, one row of options 1.
		{name: "without images", width: 20, wantHeight: 3},
		// Message 2, top 1, one row of options 1, bottom 1.
		{name: "outset rows around the options", images: images, width: 20, wantHeight: 5},
		// Two 4-wide options need 2+4+4+4+2 cells to share a row.
		{name: "sides fit one row", images: images, width: 16, wantHeight: 5},
		// Message 2, top 1, two rows of 1 with 2 between them, bottom 1.
		{name: "sides wrap the options", images: images, width: 15, wantHeight: 8},
		// Without images a cell of air at each edge and between fits.
		{name: "without images at the same width", width: 15, wantHeight: 3},
		// Message 2, top 1, two rows of 1 with 1 between them.
		{
			name: "at least a row between rows",
			images: &ButtonImages{
				Idle: testButtonImage("idle"), Outset: ButtonOutset{Top: 1},
			},
			width: 9, wantHeight: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := fixedPrompt(2, "Go", "Skip")
			cfg.ButtonImages = tt.images
			p := NewPrompt(cfg)
			assert.Equal(t, tt.wantHeight, p.Height(tt.width))
		})
	}

	t.Run("Dimensions", func(t *testing.T) {
		cfg := fixedPrompt(2, "Go", "Skip")
		cfg.ButtonImages = images
		width, height := NewPrompt(cfg).Dimensions()
		// 2 options of 4 cells, 2 to each edge and 4 between them.
		assert.Equal(t, 16, width)
		assert.Equal(t, 5, height)
	})

	t.Run("Dimensions leave two cells of air", func(t *testing.T) {
		cfg := fixedPrompt(2, "Go", "Skip")
		cfg.ButtonImages = &ButtonImages{Idle: testButtonImage("idle")}
		width, height := NewPrompt(cfg).Dimensions()
		assert.Equal(t, 14, width)
		assert.Equal(t, 3, height)
	})

	t.Run("OptionAt hits only the labels", func(t *testing.T) {
		cfg := fixedPrompt(2, "Go", "Skip")
		cfg.ButtonImages = images
		p := NewPrompt(cfg)
		p.Resize(16, 5)
		want := make(map[term.Coordinates]int)
		for x := range 4 {
			want[term.Coordinates{X: 2 + x, Y: 3}] = 0
			want[term.Coordinates{X: 10 + x, Y: 3}] = 1
		}
		assert.Equal(t, want, optionCells(p, 16, 5))
	})
}

func TestPromptButtonImagesDraw(t *testing.T) {
	images := &ButtonImages{
		Idle:     testButtonImage("idle"),
		Selected: testButtonImage("selected"),
		Pressed:  testButtonImage("pressed"),
		Outset:   ButtonOutset{Top: 1, Bottom: 1, Sides: 2},
	}
	// At 16x5 the message takes rows 0-1, the top outset row 2, the
	// options row 3 at x 2-5 and 10-13, and the bottom outset row 4.
	tests := []struct {
		name   string
		states [2]ButtonState
		want   []string
	}{
		{
			name:   "idle",
			states: [2]ButtonState{ButtonIdle, ButtonIdle},
			want: []string{
				"idle.middle 3,2 1x3", "idle.middle 4,2 1x3",
				"idle.left 0,2 3x3", "idle.right 5,2 3x3",
				"idle.middle 11,2 1x3", "idle.middle 12,2 1x3",
				"idle.left 8,2 3x3", "idle.right 13,2 3x3",
			},
		},
		{
			name:   "per state",
			states: [2]ButtonState{ButtonSelected, ButtonPressed},
			want: []string{
				"selected.middle 3,2 1x3", "selected.middle 4,2 1x3",
				"selected.left 0,2 3x3", "selected.right 5,2 3x3",
				"pressed.middle 11,2 1x3", "pressed.middle 12,2 1x3",
				"pressed.left 8,2 3x3", "pressed.right 13,2 3x3",
			},
		},
		{
			name:   "hover falls back to idle",
			states: [2]ButtonState{ButtonHover, ButtonIdle},
			want: []string{
				"idle.middle 3,2 1x3", "idle.middle 4,2 1x3",
				"idle.left 0,2 3x3", "idle.right 5,2 3x3",
				"idle.middle 11,2 1x3", "idle.middle 12,2 1x3",
				"idle.left 8,2 3x3", "idle.right 13,2 3x3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := fixedPrompt(2, "Go", "Skip")
			cfg.ButtonImages = images
			p := NewPrompt(cfg)
			p.Resize(16, 5)
			for i, st := range tt.states {
				p.SetOptionState(i, st, term.Attributes{})
			}
			rec := &imageRecorder{graphic: true}
			p.Draw(rec)
			assert.Equal(t, tt.want, describeImages(rec.images))
		})
	}
}

func TestPromptButtonImagesLabel(t *testing.T) {
	images := &ButtonImages{
		Idle:   testButtonImage("idle"),
		Outset: ButtonOutset{Top: 1, Bottom: 1, Sides: 2},
	}
	selected := term.Attributes{
		Fg: term.ColorRed, Bg: term.ColorBlue, Attrs: term.AttrBold | term.AttrReverse,
	}
	promptBg := term.Attributes{Bg: term.ColorGreen}
	empty := strings.Repeat(" ", 16)

	tests := []struct {
		name    string
		graphic bool
		want    string
		// wantLabel is the attributes of the G of Go.
		wantLabel term.Attributes
	}{
		{
			name:    "over images the label drops its background",
			graphic: true,
			want: empty + "\n" + empty + "\n" + empty + "\n" +
				"   Go     Skip  \n" + empty,
			wantLabel: term.Attributes{
				Fg: term.ColorRed, Bg: term.ColorGreen, Attrs: term.AttrBold,
			},
		},
		{
			name:    "without images the label is drawn solid",
			graphic: false,
			want: empty + "\n" + empty + "\n" + empty + "\n" +
				"  ████    Skip  \n" + empty,
			wantLabel: selected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := fixedPrompt(2, "Go", "Skip")
			cfg.ButtonImages = images
			p := NewPrompt(cfg)
			p.Resize(16, 5)
			p.SetOptionState(0, ButtonSelected, selected)
			p.SetOptionState(1, ButtonIdle, term.Attributes{})

			sw := term.NewStringWriter(16, 5)
			sw.BackgroundCh = '█'
			w := &graphicStringWriter{StringWriter: sw, graphic: tt.graphic}
			p.Draw(w)
			require.NoError(t, sw.Flush())
			assert.Equal(t, tt.want, sw.String())

			cfg.BackgroundAttributes = promptBg
			p = NewPrompt(cfg)
			p.Resize(16, 5)
			p.SetOptionState(0, ButtonSelected, selected)
			sw = term.NewStringWriter(16, 5)
			p.Draw(&graphicStringWriter{StringWriter: sw, graphic: tt.graphic})
			assert.Equal(t, tt.wantLabel, sw.Cells()[3*16+3].Attributes())
		})
	}
}

// fixedPrompt asks options under a message height rows tall at any
// width.
func fixedPrompt(height int, options ...string) PromptConfig {
	return PromptConfig{
		Message: "?",
		Options: options,
		NewMessage: func(string) Floating {
			return &TestResponsive{WantHeight: height}
		},
	}
}

// linesPrompt asks options under a linesMessage of lines.
func linesPrompt(lines []string, options ...string) PromptConfig {
	return PromptConfig{
		Message: "?",
		Options: options,
		NewMessage: func(string) Floating {
			return &linesMessage{lines: lines}
		},
	}
}

// idealPrompt asks options under a message two rows tall that is
// Floating but not Responsive.
func idealPrompt(options ...string) PromptConfig {
	return PromptConfig{
		Message:    "?",
		Options:    options,
		NewMessage: func(string) Floating { return idealMessage{} },
	}
}

// nopPrompt asks options under a message that takes no rows.
func nopPrompt(options ...string) PromptConfig {
	return PromptConfig{
		Message:    "?",
		Options:    options,
		NewMessage: func(string) Floating { return NopFloating() },
	}
}

func framed(cfg PromptConfig) PromptConfig {
	cfg.Frame = FrameCharSetDefault()
	return cfg
}

// linesMessage is a prompt message drawing a line per row, so a test
// can tell which of its rows are on show. Lines longer than the width
// wrap onto the next row.
type linesMessage struct {
	lines []string
	width int
}

func (m *linesMessage) rows(width int) []string {
	if width <= 0 {
		return m.lines
	}
	var rows []string
	for _, line := range m.lines {
		for len(line) > width {
			rows = append(rows, line[:width])
			line = line[width:]
		}
		rows = append(rows, line)
	}
	return rows
}

func (m *linesMessage) Resize(width, _ int) { m.width = width }

func (m *linesMessage) Draw(w term.Writer) {
	for y, line := range m.rows(m.width) {
		for x, ch := range line {
			w.SetCell(term.Coordinates{X: x, Y: y}, term.NewCell(ch, 1, term.Attributes{}))
		}
	}
}

func (m *linesMessage) Height(width int) int { return len(m.rows(width)) }

func (m *linesMessage) Dimensions() (int, int) { return 10, len(m.lines) }

type idealMessage struct{}

func (idealMessage) Resize(int, int) {}

func (idealMessage) Draw(term.Writer) {}

func (idealMessage) Dimensions() (int, int) { return 10, 2 }

// optionCells is the option OptionAt finds at every cell of a width
// by height prompt that has one.
func optionCells(p *Prompt, width, height int) map[term.Coordinates]int {
	cells := make(map[term.Coordinates]int)
	for y := range height {
		for x := range width {
			if i := p.OptionAt(x, y); i >= 0 {
				cells[term.Coordinates{X: x, Y: y}] = i
			}
		}
	}
	return cells
}

// optionsHit is the options OptionAt finds anywhere in a width by
// height prompt, in order.
func optionsHit(p *Prompt, width, height int) []int {
	var hit []int
	for _, i := range optionCells(p, width, height) {
		if !slices.Contains(hit, i) {
			hit = append(hit, i)
		}
	}
	slices.Sort(hit)
	return hit
}

// outOfBoundsWriter records the cells written outside of width by
// height.
type outOfBoundsWriter struct {
	width, height int
	outside       []term.Coordinates
}

func (w *outOfBoundsWriter) Context() context.Context { return context.Background() }

func (w *outOfBoundsWriter) SetCell(pos term.Coordinates, _ term.Cell) { w.check(pos) }

func (w *outOfBoundsWriter) UnionAttributes(pos term.Coordinates, _ term.Attributes) {
	w.check(pos)
}

func (w *outOfBoundsWriter) DrawImage(term.Image) bool { return false }

func (w *outOfBoundsWriter) check(pos term.Coordinates) {
	if pos.X < 0 || pos.Y < 0 || pos.X >= w.width || pos.Y >= w.height {
		w.outside = append(w.outside, pos)
	}
}

// graphicStringWriter is a StringWriter that draws images when graphic
// is set.
type graphicStringWriter struct {
	*term.StringWriter
	graphic bool
}

func (w *graphicStringWriter) DrawImage(term.Image) bool { return w.graphic }
