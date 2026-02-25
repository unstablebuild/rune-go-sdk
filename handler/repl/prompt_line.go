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

package repl

import (
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
)

var _ component.Responsive = (*promptLine)(nil)

// promptLine is a Responsive component that draws the prompt
// prefix with one set of attributes and the command text with
// another. This enables per-cell coloring that is not possible
// through the public ResponsiveString API.
type promptLine struct {
	promptRunes []rune
	allRunes    []rune
	promptAttr  term.Attributes
	textAttr    term.Attributes
	width       int
	height      int
}

func newPromptLine(
	prompt, text string,
	promptAttr, textAttr term.Attributes,
) *promptLine {
	pr := []rune(prompt)
	all := append(pr, []rune(text)...)
	return &promptLine{
		promptRunes: pr,
		allRunes:    all,
		promptAttr:  promptAttr,
		textAttr:    textAttr,
	}
}

// Height satisfies component.Responsive.
func (p *promptLine) Height(width int) int {
	if width <= 0 {
		return 0
	}
	n := len(p.allRunes)
	if n == 0 {
		return 1
	}
	return (n + width - 1) / width
}

// Resize satisfies tui.Component.
func (p *promptLine) Resize(width, height int) {
	p.width = width
	p.height = height
}

// Draw satisfies tui.Component.
func (p *promptLine) Draw(w term.Writer) {
	if p.width <= 0 || p.height <= 0 {
		return
	}
	promptLen := len(p.promptRunes)
	x, y := 0, 0
	for i, r := range p.allRunes {
		if y >= p.height {
			break
		}
		attr := p.textAttr
		if i < promptLen {
			attr = p.promptAttr
		}
		w.SetCell(term.Coordinates{X: x, Y: y}, term.Cell{
			Ch:         r,
			Attributes: attr,
			Width:      1,
		})
		x++
		if x >= p.width {
			x = 0
			y++
		}
	}
}

func (p *promptLine) setPromptAttr(attr term.Attributes) {
	p.promptAttr = attr
}
