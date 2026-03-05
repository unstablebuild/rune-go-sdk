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
	w := term.NewStringWriter(width+1, height+1)
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
			// 20 wide: each option span = 10 columns
			// "Yay" centered in 10: x=3, "Nay" centered in 10: x=13
		}, 20, 10,
			"                     \n"+
				"                     \n"+
				"      Do you?        \n"+
				"                     \n"+
				"                     \n"+
				"                     \n"+
				"                     \n"+
				"                     \n"+
				"                     \n"+
				"   Yay       Nay     \n"+
				"                     ",
		)
	})

	t.Run("with frame", func(t *testing.T) {
		testDrawPrompt(t, PromptConfig{
			Message: "Do you?",
			Options: []string{"Yay", "Nay"},
			Frame:   FrameCharSetDefault(),
			// 20 wide: each option span = 10 columns
			// "Yay" frame (7) centered in 10: x=1
			// "Nay" frame (7) centered in 10: x=11
		}, 20, 10,
			"                     \n"+
				"                     \n"+
				"      Do you?        \n"+
				"                     \n"+
				"                     \n"+
				"                     \n"+
				"                     \n"+
				" ┌─────┐   ┌─────┐   \n"+
				" │ Yay │   │ Nay │   \n"+
				" └─────┘   └─────┘   \n"+
				"                     ",
		)
	})

	t.Run("with frame three options", func(t *testing.T) {
		testDrawPrompt(t, PromptConfig{
			Message: "Why?",
			Options: []string{"Yay", "Nay", "Say"},
			Frame:   FrameCharSetDefault(),
			// 30 wide: each option span = 10 columns (4 of 12)
			// Each frame (7) centered in 10: offset=1
			// Frames at x=1, x=11, x=21
		}, 30, 10,
			"                               \n"+
				"                               \n"+
				"             Why?              \n"+
				"                               \n"+
				"                               \n"+
				"                               \n"+
				"                               \n"+
				" ┌─────┐   ┌─────┐   ┌─────┐   \n"+
				" │ Yay │   │ Nay │   │ Say │   \n"+
				" └─────┘   └─────┘   └─────┘   \n"+
				"                               ",
		)
	})

	t.Run("with frame even text centering", func(t *testing.T) {
		testDrawPrompt(t, PromptConfig{
			Message: "Install?",
			Options: []string{"Go", "Skip"},
			Frame:   FrameCharSetDefault(),
			// 24 wide: each option span = 12 columns
			// "Go" frame (6) centered in 12: x=3
			// "Skip" frame (8) centered in 12: x=14
		}, 24, 10,
			"                         \n"+
				"                         \n"+
				"        Install?         \n"+
				"                         \n"+
				"                         \n"+
				"                         \n"+
				"                         \n"+
				"   ┌────┐     ┌──────┐   \n"+
				"   │ Go │     │ Skip │   \n"+
				"   └────┘     └──────┘   \n"+
				"                         ",
		)
	})
}

func TestDrawPromptWide(t *testing.T) {
	// 101-char empty line (width=100 + 1 writer column).
	e := "                                                                                                     "
	// Message line: "Sure?" centered in 100 columns.
	m := "                                               Sure?                                                 "

	t.Run("single button", func(t *testing.T) {
		// optCols = 12 (all columns), span = 100 wide
		// "Ok" frame (6) centered in 100: x=47
		testDrawPrompt(t, PromptConfig{
			Message: "Sure?",
			Options: []string{"Ok"},
			Frame:   FrameCharSetDefault(),
		}, 100, 10,
			e+"\n"+e+"\n"+m+"\n"+e+"\n"+e+"\n"+e+"\n"+e+"\n"+
				"                                               ┌────┐                                                \n"+
				"                                               │ Ok │                                                \n"+
				"                                               └────┘                                                \n"+
				e,
		)
	})

	t.Run("two buttons", func(t *testing.T) {
		// optCols = 6, each span = 50 wide
		// "Ok" frame (6) centered in 50: x=22
		// "Cancel" frame (10) centered in 50: x=70
		testDrawPrompt(t, PromptConfig{
			Message: "Sure?",
			Options: []string{"Ok", "Cancel"},
			Frame:   FrameCharSetDefault(),
		}, 100, 10,
			e+"\n"+e+"\n"+m+"\n"+e+"\n"+e+"\n"+e+"\n"+e+"\n"+
				"                      ┌────┐                                          ┌────────┐                     \n"+
				"                      │ Ok │                                          │ Cancel │                     \n"+
				"                      └────┘                                          └────────┘                     \n"+
				e,
		)
	})

	t.Run("three buttons", func(t *testing.T) {
		// optCols = 4, each span = 33 wide (99 of 100 used)
		// "Ok" frame (6) centered in 33: x=13
		// "Cancel" frame (10) centered in 33: x=44
		// "Help" frame (8) centered in 33: x=78
		testDrawPrompt(t, PromptConfig{
			Message: "Sure?",
			Options: []string{"Ok", "Cancel", "Help"},
			Frame:   FrameCharSetDefault(),
		}, 100, 10,
			e+"\n"+e+"\n"+m+"\n"+e+"\n"+e+"\n"+e+"\n"+e+"\n"+
				"             ┌────┐                         ┌────────┐                        ┌──────┐               \n"+
				"             │ Ok │                         │ Cancel │                        │ Help │               \n"+
				"             └────┘                         └────────┘                        └──────┘               \n"+
				e,
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
				"  yyyyyy    nnnnnn   \n" +
				"  yyyyyy    nnnnnn   \n" +
				"  yyyyyy    nnnnnn   \n" +
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

	t.Run("respects MinWidth", func(t *testing.T) {
		p := NewPrompt(PromptConfig{
			Message:  "Ok?",
			Options:  []string{"Y"},
			MinWidth: 50,
		})
		w, _ := p.Dimensions()
		assert.GreaterOrEqual(t, w, 50)
	})
}
