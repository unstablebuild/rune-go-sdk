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

import "github.com/unstablebuild/rune-go-sdk/tui"

// Floating components are not in principle confined to a predetermined
// space and so are allowed certain degree of freedom. Dimensions should be used
// to indicate to parent components the desired dimensions, although it shouldn't be
// assumed that it will be respected.
type Floating interface {
	tui.Component
	Dimensions() (width int, height int)
}

// StaticFloating wraps a tui.Component and returns a Floating that always
// return the same Dimensions values.
func StaticFloating(c tui.Component, width, height int) Floating {
	return &staticFloating{width: width, height: height, Component: c}
}

// PaddedFloating wraps a Floating component and adds a pre-determined
// amount of x axis and y axis padding.
func PaddedFloating(f Floating, padx, pady int) Floating {
	return paddedFloating{padx: padx, pady: pady, Floating: f}
}

type staticFloating struct {
	tui.Component
	width, height int
}

func (s *staticFloating) Dimensions() (int, int) {
	return s.width, s.height
}

type paddedFloating struct {
	Floating
	padx, pady int
}

func (p paddedFloating) Dimensions() (width, height int) {
	width, height = p.Floating.Dimensions()
	width += p.padx
	height += p.pady
	return
}
