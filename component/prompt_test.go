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
	"github.com/unstablebuild/rune-go-sdk/component/comptest"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/tcell/v3"
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
			Fg:    tcell.ColorRed,
			Attrs: tcell.AttrBold,
		}
		p.SetOptionAttr(0, attr)
		assert.Equal(t, attr, opts["y"].Attributes)
	})
}

func TestPromptButtonBackgroundAttributes(t *testing.T) {
	focus := term.Attributes{Bg: tcell.ColorBlue}

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
