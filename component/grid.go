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
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

type grid struct {
	width, height int
	matrix        [][]tui.Component
}

// Grid renders matrix as an equaly sized grid of components.
// Note that each []tui.Component in matrix is treated as a row.
func Grid(matrix [][]tui.Component) tui.Component {
	return &grid{matrix: matrix}
}

func (g *grid) Resize(width, height int) {
	g.width, g.height = width, height
}

func (g *grid) Draw(w term.Writer) {
	var v Virtual[tui.Component]

	if len(g.matrix) == 0 {
		return
	}

	heightPerComp := g.height / len(g.matrix)
	for y, row := range g.matrix {
		yComp := y * heightPerComp
		if yComp+heightPerComp > g.height {
			break
		}
		if len(row) == 0 {
			continue
		}
		widthPerComp := g.width / len(row)
		for x, comp := range row {
			if comp == nil {
				continue
			}
			xComp := x * widthPerComp
			if xComp+widthPerComp > g.width {
				break
			}
			v.C = comp
			v.Resize(widthPerComp, heightPerComp)
			v.Move(term.Coordinates{X: xComp, Y: yComp})
			v.Draw(w)
		}
	}
}
