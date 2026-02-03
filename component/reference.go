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

var _ tui.Component = (*Reference)(nil)

// Reference can be used to dynamically swap a tui.Component.
// If the underlying tui.Component is nil, then Resize and Draw do nothing.
type Reference struct {
	component tui.Component
	dirty     bool
	height    int
	width     int
}

// NewReference allocates storage for a new Reference and initializes it with ref.
func NewReference(ref tui.Component) *Reference {
	ret := new(Reference)
	ret.Init(ref)
	return ret
}

// Init initializes this reference with ref.
// It can be used subsequently to override the underlying tui.Component reference.
func (r *Reference) Init(ref tui.Component) {
	r.component = ref
	r.dirty = true
}

// Resize satisfies tui.Component.
func (r *Reference) Resize(width, height int) {
	r.height, r.width = height, width
	if r.component == nil {
		return
	}
	r.dirty = false
	r.component.Resize(width, height)
}

// Draw satisfies tui.Component.
func (r *Reference) Draw(w term.Writer) {
	if r.component == nil {
		return
	}
	if r.dirty {
		r.Resize(r.width, r.height)
	}
	r.component.Draw(w)
}

// FloatingReference can be used to dynamically swap a component.Floating.
// If the underlying component.Floating is nil, then methods do nothing.
type FloatingReference struct {
	Reference
}

// NewFloatingReference allocates storage for a new FloatingReference and
// initializes it with ref.
func NewFloatingReference(ref Floating) *FloatingReference {
	ret := new(FloatingReference)
	ret.Init(ref)
	return ret
}

// Init initializes this reference with ref.
// It can be used subsequently to override the underlying tui.Component reference.
func (r *FloatingReference) Init(ref Floating) {
	r.Reference.Init(ref)
}

// Dimensions satisfies Floating.
func (r *FloatingReference) Dimensions() (width, height int) {
	if _, ok := r.component.(Floating); !ok {
		return
	}
	return r.Reference.component.(Floating).Dimensions()
}
