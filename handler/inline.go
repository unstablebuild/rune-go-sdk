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
)

// Inline is the handler counterpart to component.Inline: it lays
// handlers out next to each other, each as wide as its Dimensions, and
// routes events to them as Group does.
type Inline struct {
	routed
	alignment     component.Alignment
	handlers      []*component.Virtual[Floating]
	width, height int
}

var _ Floating = (*Inline)(nil)

// NewInline returns an Inline of handlers. Those that overflow it are
// truncated at the side opposite to alignment.
func NewInline(handlers []Floating, alignment component.Alignment) *Inline {
	i := &Inline{alignment: alignment, handlers: make([]*component.Virtual[Floating], len(handlers))}
	for n, h := range handlers {
		i.handlers[n] = &component.Virtual[Floating]{C: h}
	}
	i.layout = i.children
	return i
}

func (i *Inline) children() Layout {
	l := Layout{Width: i.width, Height: i.height, Children: make([]Child, len(i.handlers))}
	for n, v := range i.handlers {
		l.Children[n] = Child{
			Handler: v.C, Position: v.Position(), Width: v.Width(), Height: v.Height(), CaptureDrags: true,
		}
	}
	return l
}

// Dimensions satisfies component.Floating.
func (i *Inline) Dimensions() (width, height int) {
	for _, v := range i.handlers {
		w, h := v.C.Dimensions()
		width += w
		height = max(height, h)
	}
	return
}

// Resize satisfies tui.Component.
func (i *Inline) Resize(width, height int) {
	i.width, i.height = width, height
	var offset int
	for _, v := range i.handlers {
		w, _ := v.C.Dimensions()
		w = min(w, width)
		v.Move(term.Coordinates{X: offset})
		v.Resize(w, height)
		width -= w
		offset += w
	}
}

// Draw satisfies tui.Component.
func (i *Inline) Draw(w term.Writer) {
	for _, v := range i.handlers {
		v.Draw(w)
	}
}
