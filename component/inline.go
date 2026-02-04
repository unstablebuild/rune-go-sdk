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
)

// Inline renders the given slice of components next to each other,
// occupying the maximum height of all the components but respecting
// each component's width. If there's overflow, then the overflowing
// components will be truncated at the opposite side of the Alignment.
func Inline(components []Floating, alignment Alignment) Floating {
	comps := make([]*Virtual[Floating], len(components))
	for i := range comps {
		comps[i] = &Virtual[Floating]{C: components[i]}
	}
	return inline{alignment: alignment, comps: comps}
}

type inline struct {
	alignment Alignment
	comps     []*Virtual[Floating]
}

func (i inline) Dimensions() (width int, height int) {
	for _, comp := range i.comps {
		cwidth, cheight := comp.C.Dimensions()
		width += cwidth
		if cheight > height {
			height = cheight
		}
	}
	return
}

func (i inline) Resize(width, height int) {
	var offset int
	for _, comp := range i.comps {
		desiredWidth, _ := comp.C.Dimensions()
		if desiredWidth > width {
			desiredWidth = width
		}
		comp.Move(term.Coordinates{X: offset})
		comp.Resize(desiredWidth, height)
		width -= desiredWidth
		offset += desiredWidth
	}
}

func (i inline) Draw(w term.Writer) {
	for _, comp := range i.comps {
		comp.Draw(w)
	}
}
