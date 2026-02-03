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

package handler

import (
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Scrollable is a tui.Handler that is capable of scrolling content.
// See component.Scrollable for more details.
type Scrollable interface {
	tui.Handler
	component.Scrollable
}

// ScrollableFloating is a Floating that is capable of scrolling content.
// See component.Scrollable for more details.
type ScrollableFloating interface {
	Scrollable
	component.Floating
}

// ScrollableWithAttributes is a WithAttributes that is capable of scrolling content.
// See component.Scrollable for more details.
type ScrollableWithAttributes interface {
	Scrollable
	component.WithAttributes
}

// ScrollableFloatingWithAttributes is a FloatingWithAttributes that is
// capable of scrolling content. See component.Scrollable for more details.
type ScrollableFloatingWithAttributes interface {
	Scrollable
	component.Floating
	component.WithAttributes
}
