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
	"github.com/unstablebuild/rune-go-sdk/term"
)

func drawBar(p *ProgressBar, width int) string {
	w := term.NewStringWriter(width, 1)
	p.Resize(width, 1)
	p.Draw(w)
	_ = w.Flush()
	return w.String()
}

func TestProgressBar(t *testing.T) {
	chars := DefaultProgressBarCharSet()

	tests := []struct {
		name     string
		progress int64
		total    int64
		units    string
		width    int
		expected string
	}{
		{
			name:     "empty bar",
			progress: 0,
			total:    100,
			units:    "bytes",
			width:    30,
			expected: "╢_____________╟ 0% 0/100 bytes",
		},
		{
			name:     "half bar",
			progress: 50,
			total:    100,
			units:    "bytes",
			width:    30,
			expected: "╢░░░░░______╟ 50% 50/100 bytes",
		},
		{
			name:     "full bar",
			progress: 100,
			total:    100,
			units:    "bytes",
			width:    30,
			expected: "╢░░░░░░░░░╟ 100% 100/100 bytes",
		},
		{
			name:     "zero total",
			progress: 0,
			total:    0,
			units:    "",
			width:    20,
			expected: "╢___________╟ 0% 0/0",
		},
		{
			name:     "label with units",
			progress: 50,
			total:    100,
			units:    "bytes",
			width:    30,
			expected: "╢░░░░░______╟ 50% 50/100 bytes",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := NewProgressBar(chars, term.Attributes{})
			p.SetProgress(tc.progress, tc.total, tc.units)
			got := drawBar(p, tc.width)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestProgressBarHalfFilled(t *testing.T) {
	p := NewProgressBar(DefaultProgressBarCharSet(), term.Attributes{})
	p.SetProgress(50, 100, "")
	got := drawBar(p, 22)
	assert.Equal(t, "╢░░░░_____╟ 50% 50/100", got)
}

func TestProgressBarSetAttr(t *testing.T) {
	p := NewProgressBar(
		DefaultProgressBarCharSet(), term.Attributes{},
	)
	newAttr := term.Attributes{Fg: 0xFF}
	old := p.SetAttr(newAttr)
	assert.Equal(t, term.Attributes{}, old)
	old = p.SetAttr(term.Attributes{})
	assert.Equal(t, newAttr, old)
}
