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

import "github.com/unstablebuild/rune-go-sdk/term"

// Scrollable abstracts a component that can scroll up or down.
type Scrollable interface {
	SeekUp() bool
	SeekDown() bool
	SeekOffset() int
	MaxSeekOffset() int
}

// ScrollableFloating combines a Scrollable with Floating.
type ScrollableFloating interface {
	Scrollable
	Floating
}

// NopScrollable returns a Scrollable that does nothing.
func NopScrollable() Scrollable {
	return nopScrollableFloating{}
}

// NopScrollableFloating returns a ScrollableFloating that does nothing.
func NopScrollableFloating() ScrollableFloating {
	return nopScrollableFloating{}
}

type nopScrollableFloating struct{}

func (n nopScrollableFloating) SeekUp() bool                        { return false }
func (n nopScrollableFloating) SeekDown() bool                      { return false }
func (n nopScrollableFloating) SeekOffset() int                     { return 0 }
func (n nopScrollableFloating) MaxSeekOffset() int                  { return 0 }
func (n nopScrollableFloating) Dimensions() (width int, height int) { return }
func (n nopScrollableFloating) Resize(width, height int)            {}
func (n nopScrollableFloating) Draw(w term.Writer)                  {}
