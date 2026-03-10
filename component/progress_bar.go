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

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

var _ tui.Component = (*ProgressBar)(nil)
var _ WithAttributes = (*ProgressBar)(nil)

// ProgressBarCharSet defines the characters used to
// render a ProgressBar.
type ProgressBarCharSet struct {
	Left, Filled, Empty, Right rune
}

// DefaultProgressBarCharSet returns the default char set
// for ProgressBar.
func DefaultProgressBarCharSet() ProgressBarCharSet {
	return ProgressBarCharSet{
		Left:   '╢',
		Filled: '░',
		Empty:  '_',
		Right:  '╟',
	}
}

// ProgressBar is a component that renders a horizontal
// progress bar. Its height is always 1.
type ProgressBar struct {
	chars           ProgressBarCharSet
	attr            term.Attributes
	progress, total int64
	units           string
	width, height   int
}

// NewProgressBar creates a new ProgressBar with the
// given char set and attributes.
func NewProgressBar(
	chars ProgressBarCharSet, attr term.Attributes,
) *ProgressBar {
	return &ProgressBar{
		chars: chars,
		attr:  attr,
	}
}

// SetProgress updates the bar state.
func (p *ProgressBar) SetProgress(
	progress, total int64, units string,
) {
	p.progress = progress
	p.total = total
	p.units = units
}

// SetAttr satisfies WithAttributes.
func (p *ProgressBar) SetAttr(
	attr term.Attributes,
) term.Attributes {
	old := p.attr
	p.attr = attr
	return old
}

// Resize satisfies tui.Component.
func (p *ProgressBar) Resize(width, height int) {
	p.width = width
	p.height = height
}

// Draw satisfies tui.Component.
func (p *ProgressBar) Draw(w term.Writer) {
	if p.width < 2 || p.height < 1 {
		return
	}

	var pct float64
	if p.total > 0 {
		pct = float64(p.progress) / float64(p.total)
		if pct > 1 {
			pct = 1
		}
	}

	labelRunes := []rune(p.buildLabel(pct))

	// Reserve space for the label after the bar when possible:
	// <left><filled/empty><right><space><label>
	barWidth := p.width
	if len(labelRunes) > 0 {
		reserved := 1 + len(labelRunes)
		if reserved < p.width {
			barWidth = p.width - reserved
		}
	}
	if barWidth < 2 {
		barWidth = 2
	}

	// Inner width excludes the left and right border
	// characters.
	inner := barWidth - 2

	filled := int(pct * float64(inner))
	if filled > inner {
		filled = inner
	}

	// Build the bar runes: filled portion + empty.
	bar := make([]rune, inner)
	for i := range inner {
		if i < filled {
			bar[i] = p.chars.Filled
		} else {
			bar[i] = p.chars.Empty
		}
	}

	for y := range p.height {
		w.SetCell(
			term.Coordinates{X: 0, Y: y},
			term.Cell{
				Ch:         p.chars.Left,
				Attributes: p.attr,
				Width:      1,
			},
		)
		for x, r := range bar {
			w.SetCell(
				term.Coordinates{X: x + 1, Y: y},
				term.Cell{
					Ch:         r,
					Attributes: p.attr,
					Width:      1,
				},
			)
		}
		w.SetCell(
			term.Coordinates{X: barWidth - 1, Y: y},
			term.Cell{
				Ch:         p.chars.Right,
				Attributes: p.attr,
				Width:      1,
			},
		)

		if barWidth < p.width && len(labelRunes) > 0 {
			w.SetCell(
				term.Coordinates{X: barWidth, Y: y},
				term.Cell{
					Ch:         ' ',
					Attributes: p.attr,
					Width:      1,
				},
			)
			for i, r := range labelRunes {
				x := barWidth + 1 + i
				if x >= p.width {
					break
				}
				w.SetCell(
					term.Coordinates{X: x, Y: y},
					term.Cell{
						Ch:         r,
						Attributes: p.attr,
						Width:      1,
					},
				)
			}
		}
	}
}

func (p *ProgressBar) buildLabel(pct float64) string {
	pctInt := int(pct * 100)
	if p.units != "" {
		return fmt.Sprintf(
			"%d%% %d/%d %s",
			pctInt, p.progress, p.total, p.units,
		)
	}
	return fmt.Sprintf(
		"%d%% %d/%d",
		pctInt, p.progress, p.total,
	)
}
