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
	"strings"

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Divider returns a divider compatible with Responsive collections
// that will simply draw a line that will occupy perc of its width.
func Divider(perc float64, cfg StringConfig) Responsive {
	return &divider{perc: perc, cfg: cfg}
}

type divider struct {
	perc float64
	cfg  StringConfig
	comp Virtual[tui.Component]
}

func (d *divider) Resize(width, height int) {
	totalWidth := int(float64(width) * d.perc)

	var builder strings.Builder
	for i := 0; i < totalWidth; i++ {
		builder.WriteRune(d.cfg.HorizontalTop)
	}

	d.comp.C = NewStringWithConfig(builder.String(), d.cfg)

	offsetX := int((float64(width) - float64(totalWidth)) / 2)
	offsetY := int(float64(height) / 2)
	d.comp.Resize(totalWidth, 1)
	d.comp.Move(term.Coordinates{X: offsetX, Y: offsetY})
}

func (d *divider) Draw(w term.Writer) {
	d.comp.Draw(w)
}

func (d *divider) Height(width int) int {
	return 1
}
