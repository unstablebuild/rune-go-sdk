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

var _ Responsive = (*Row)(nil)

const (
	// MaxCols is the maximum number of available columns in a row.
	MaxCols = 12
)

// Row provides means to horizontally stack a set of components.
// It exposes AddComponent which enables clients to preemptively divide
// the available width in 12 columns. A component that's installed
// with 6 columns for instance, will use exactly 50% of the available width.
//
// Row satisfies component.Responsive by returning the biggest height
// of any of its children.
//
// A zero-valued row is ready to be used.
type Row struct {
	content       []*rowComp
	height, width int
	dirty         bool
	Virtual[tui.Component]
}

type rowComp struct {
	cols int
	*Virtual[Responsive]
}

// NewRow allocates storage for a new Row and initializes it.
func NewRow() *Row {
	ret := new(Row)
	return ret
}

// AddComponent adds a component to this Row with the given columns
// over MaxCols as the pre-defined width.
func (r *Row) AddComponent(c Responsive, cols int) *Virtual[Responsive] {
	virt := new(Virtual[Responsive])
	virt.C = c
	r.content = append(r.content, &rowComp{
		cols:    cols,
		Virtual: virt,
	})
	r.dirty = true
	return virt
}

// Draw satisfies tui.Component.
func (r *Row) Draw(w term.Writer) {
	if r.dirty {
		r.Resize(r.width, r.height)
	}

	// use Virtual position, set by Container
	vw := VirtualWriter{w, r.Position(), r.height, r.width}
	for _, comp := range r.content {
		comp.Draw(&vw)
	}
}

// Resize satisfies tui.Component.
func (r *Row) Resize(width, height int) {
	r.width, r.height = width, height
	colWidth := rowColWidth(width)
	var offset int
	for _, comp := range r.content {
		compWidth := rowCompWidth(colWidth, comp)
		pos := term.Coordinates{X: offset}
		comp.Move(pos)
		if offset == width {
			continue
		}
		if offset+compWidth >= width {
			compWidth = width - offset
		}
		comp.Resize(compWidth, height)
		offset += compWidth
	}
	r.dirty = false
}

// Height satisfies component.Responsive.
func (r *Row) Height(width int) (ret int) {
	colWidth := rowColWidth(width)
	for _, content := range r.content {
		compWidth := rowCompWidth(colWidth, content)
		height := content.C.Height(compWidth)
		if height > ret {
			ret = height
		}
	}
	return
}

// Dimensions satisfies component.Floating. If underlying
// components do not satisfy component.Floating, then this method panics.
func (r *Row) Dimensions() (retWidth int, retHeight int) {
	for _, content := range r.content {
		width, height := content.Virtual.C.(Floating).Dimensions()
		if height > retHeight {
			retHeight = height
		}
		retWidth += width
	}
	return
}

func rowColWidth(width int) float64 {
	return float64(width) / float64(MaxCols)
}

func rowCompWidth(colWidth float64, content *rowComp) int {
	return int(float64(content.cols) * colWidth)
}
