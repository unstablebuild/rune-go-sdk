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

// ResponsiveList is the handler counterpart to component.ResponsiveList:
// the elements pushed with its methods that are handlers get events as
// Group routes them. A wheel event no element takes seeks it.
//
// Use NewResponsiveList, or Init, to initialize a ResponsiveList.
type ResponsiveList struct {
	component.ResponsiveList
	routed
}

var _ Responsive = (*ResponsiveList)(nil)

// NewResponsiveList allocates storage for a new ResponsiveList and
// initializes it.
func NewResponsiveList() *ResponsiveList {
	l := new(ResponsiveList)
	l.Init()
	return l
}

// Init initializes this ResponsiveList. It can be used to reset its
// internal state.
func (l *ResponsiveList) Init() {
	l.ResponsiveList.Init()
	l.layout = l.children
}

func (l *ResponsiveList) children() Layout {
	lay := Layout{Width: l.SizeWidth(), Height: l.SizeHeight()}
	y := -l.Offset()
	if l.Alignment == component.AlignmentBottom {
		y = l.Offset() - l.MaxOffset()
	}
	for n, ok := l.Front(); ok; n, ok = n.Next() {
		r := n.Value().(component.Responsive)
		height := r.Height(lay.Width)
		if h, ok := r.(tui.Handler); ok {
			lay.Children = append(lay.Children, Child{
				Handler: h, Position: term.Coordinates{Y: y}, Width: lay.Width, Height: height, CaptureDrags: true,
			})
		}
		y += height
	}
	return lay
}

// Handle satisfies tui.Handler.
func (l *ResponsiveList) Handle(ev term.Event) (exit, handled bool) {
	exit, handled = l.routed.Handle(ev)
	if exit || handled || ev.Type != term.EventMouse {
		return
	}
	switch ev.Key {
	case term.MouseWheelUp:
		handled = l.SeekUp()
	case term.MouseWheelDown:
		handled = l.SeekDown()
	}
	return
}
