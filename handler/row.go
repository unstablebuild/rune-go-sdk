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

// Row is the handler counterpart to component.Row: it stacks the
// handlers added with AddHandler horizontally and routes events to them
// as Group does, keys to the one SetFocus put in focus.
// Components added with AddComponent are only drawn.
//
// Use NewRow, or Init, to initialize a Row.
type Row struct {
	component.Row
	routed
	handlers []rowHandler
}

var _ Responsive = (*Row)(nil)

type rowHandler struct {
	h tui.Handler
	v *component.Virtual[component.Responsive]
}

// NewRow allocates storage for a new Row and initializes it.
func NewRow() *Row {
	r := new(Row)
	r.Init()
	return r
}

// Init initializes this Row.
func (r *Row) Init() {
	r.layout = r.children
}

// AddHandler adds h to this Row with the given columns over
// component.MaxCols as its width.
func (r *Row) AddHandler(h Responsive, cols int) *component.Virtual[component.Responsive] {
	v := r.AddComponent(h, cols)
	r.handlers = append(r.handlers, rowHandler{h: h, v: v})
	return v
}

func (r *Row) children() Layout {
	width, height := r.Size()
	l := Layout{Width: width, Height: height, Children: make([]Child, len(r.handlers))}
	for i, rh := range r.handlers {
		l.Children[i] = Child{
			Handler: rh.h, Position: rh.v.Position(), Width: rh.v.Width(), Height: rh.v.Height(), CaptureDrags: true,
		}
	}
	return l
}
