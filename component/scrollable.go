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
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Scrollable abstracts a component that can scroll up or down.
type Scrollable interface {
	tui.Component

	SeekUp() bool
	SeekDown() bool
	SeekOffset() int
	MaxSeekOffset() int
}

// ScrollableFloating is a Floating that is capable of scrolling content.
type ScrollableFloating interface {
	Scrollable
	Floating
}

// ScrollableWithAttributes is a WithAttributes that is capable of scrolling content.
type ScrollableWithAttributes interface {
	Scrollable
	WithAttributes
}

// ScrollableFloatingWithAttributes is a FloatingWithAttributes that is
// capable of scrolling content.
type ScrollableFloatingWithAttributes interface {
	Scrollable
	Floating
	WithAttributes
}
