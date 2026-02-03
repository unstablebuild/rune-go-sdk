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
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// WithAttributes represents a tui.Component that can be set attributes.
type WithAttributes interface {
	tui.Handler
	SetAttr(term.Attributes) term.Attributes
}

// WithAttributesFloating is a Floating that is capable of setting attributes.
type WithAttributesFloating interface {
	WithAttributes
	component.Floating
}

// WithAttributesResponsive is a Responsive that is capable of setting attributes.
type WithAttributesResponsive interface {
	WithAttributes
	component.Responsive
}

// WithAttributesResponsiveFloating is a Responsive and Floating
// that is capable of setting attributes.
type WithAttributesResponsiveFloating interface {
	WithAttributes
	component.Responsive
	component.Floating
}
