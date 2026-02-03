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
	"math"

	"github.com/unstablebuild/rune-go-sdk/term"
)

// DefaultAspectRatio returns an aspect ratio that looks like 16:9 in squared pixels
// Note that cells are not square, so this ratio compensates for that.
const DefaultAspectRatio = 16.0 / 9.0 * 5 / 2

// AspectRatioFloatingResponsive wraps a Responsive component and satisfies Floating by
// using its Height method to find a width such that the resulting
// dimensions approximate a given aspect ratio.
type AspectRatioFloatingResponsive struct {
	Responsive
	aspectRatio float64
}

// NewAspectRatioFloatingResponsive allocates storage for a new AspectRatioFloatingResponsive and initializes it.
func NewAspectRatioFloatingResponsive(responsive Responsive, aspectRatio float64) *AspectRatioFloatingResponsive {
	ret := new(AspectRatioFloatingResponsive)
	ret.Init(responsive, aspectRatio)
	return ret
}

// Init initializes this AspectRatioFloatingResponsive with responsive and aspectRatio.
func (r *AspectRatioFloatingResponsive) Init(responsive Responsive, aspectRatio float64) {
	if aspectRatio == 0 {
		panic("aspect ratio cannot be 0")
	}
	r.Responsive = responsive
	r.aspectRatio = aspectRatio
}

var _ Floating = (*AspectRatioFloatingResponsive)(nil)
var _ Responsive = (*AspectRatioFloatingResponsive)(nil)
var _ WithAttributes = (*AspectRatioFloatingResponsive)(nil)

// SetAttr satisfies WithAttributes, if the underlying Responsive satisfies WithAttributes,
// otherwise this method panics.
func (r *AspectRatioFloatingResponsive) SetAttr(attr term.Attributes) term.Attributes {
	return r.Responsive.(WithAttributes).SetAttr(attr)
}

// Dimensions satisfies Floating by using the underlying component's
// Height to find the set of dimensions that respect the instructed
// aspect ratio.
func (r *AspectRatioFloatingResponsive) Dimensions() (width int, height int) {
	const (
		budget            = 10
		initialIncrements = 10
		biggerIncrements  = 50
	)
	var prevWidth, prevHeight int
	var prevActualLoss float64

	var i int
	iterate := func(increments int) {
		for i = 0; i < budget; i++ {
			width += increments * (i + 1)
			height = r.Height(width)

			actualLoss, ok := r.calculateLoss(width, height)
			if ok {
				return
			}
			if actualLoss < 0 {
				// handle first iteration negative loss
				if i == 0 && increments == initialIncrements {
					prevWidth = width / 2
					prevHeight = height / 2
					prevActualLoss = 1
				}
				break
			}
			prevActualLoss = actualLoss
			prevWidth = width
			prevHeight = height
		}
	}

	// first try in small increments, if we don't get a negative actualLoss
	// then try in bigger increments. This is useful for components that need
	// hundreds of cells in width or height.
	iterate(initialIncrements)
	if i == budget {
		iterate(biggerIncrements)
		if i == budget {
			return
		}
	}

	prevActualLoss = math.Abs(prevActualLoss)
	// now that we now the inflection point, increments can be more precise
	increments := int(math.Max(1, (float64(width)-float64(prevWidth))/float64(budget)))
	width = prevWidth
	height = prevHeight
	for i := 0; i < budget; i++ {
		width += increments
		height = r.Height(width)
		actualLoss, ok := r.calculateLoss(width, height)
		if ok {
			return
		}
		actualLoss = math.Abs(actualLoss)
		if actualLoss >= prevActualLoss {
			width = prevWidth
			height = prevHeight
			return
		}
		prevWidth = width
		prevHeight = height
		prevActualLoss = actualLoss
	}

	return
}

func (r *AspectRatioFloatingResponsive) calculateLoss(width, height int) (float64, bool) {
	const (
		acceptedLoss = 0.1
	)
	aspectRatio := float64(width) / float64(height)
	actualLoss := (r.aspectRatio - aspectRatio) / r.aspectRatio
	ok := (actualLoss > 0 && actualLoss < acceptedLoss) ||
		(actualLoss <= 0 && -actualLoss < acceptedLoss)
	return actualLoss, ok
}
