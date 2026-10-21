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

// Span is a handler that takes another handler and handles padding and
// alignment. It is the handler counterpart to component.Span.
type Span struct {
	component.Span
	cfg     component.SpanConfig
	handler tui.Handler
	width   int
	height  int
	dimh    int
	dimw    int
}

var _ Responsive = (*Span)(nil)
var _ WithAttributes = (*Span)(nil)
var _ Floating = (*Span)(nil)

// NewSpan returns an initialized Span.
func NewSpan(content tui.Handler, cfg component.SpanConfig) *Span {
	s := new(Span)
	s.handler = content
	s.cfg = cfg
	s.Init(content, cfg)
	return s
}

// Resize satisfies tui.Component
func (s *Span) Resize(width, height int) {
	s.width = width
	s.height = height
	s.Span.Resize(width, height)
	if s.cfg.PadAutoFloating {
		s.dimw, s.dimh = s.Span.Content().(Floating).Dimensions()
	}
}

// Handle satisfies tui.Handler by delegating events to the underlying
// handler. Mouse events have their coordinates adjusted to account for
// the span's content offset, and those on the padding are clamped onto
// the content's edge. A motion beyond the span's top or left edge is not
// clamped: that is how a parent says the pointer left, and clamping it
// would make it a hover on the content's edge.
func (s *Span) Handle(ev term.Event) (bool, bool) {
	if ev.Type == term.EventMouse {
		beyond := ev.Key == 0 && (ev.MouseX < 0 || ev.MouseY < 0)
		offset := s.ContentOffset()
		ev.MouseX -= offset.X
		ev.MouseY -= offset.Y
		if !beyond {
			ev.MouseX = max(ev.MouseX, 0)
			ev.MouseY = max(ev.MouseY, 0)
		}
	}
	exit, handled := s.handler.Handle(ev)
	if handled && s.cfg.PadAutoFloating {
		dimw, dimh := s.Span.Content().(Floating).Dimensions()
		if dimw != s.dimw || dimh != s.dimh {
			s.Resize(s.width, s.height)
		}
	}
	return exit, handled
}

// Cursor satisfies tui.Handler by returning the underlying handler's
// cursor position offset by the span's content position.
func (s *Span) Cursor() (pos term.Coordinates, style term.CursorStyle, show bool) {
	pos, style, show = s.handler.Cursor()
	offset := s.ContentOffset()
	pos.X += offset.X
	pos.Y += offset.Y
	return
}

// Selection satisfies tui.Handler.
func (s *Span) Selection() (string, bool) {
	return s.handler.Selection()
}

// SetContent updates the underlying handler and resizes it
// to conform to this span's width and height.
func (s *Span) SetContent(content tui.Handler) {
	s.handler = content
	s.Span.SetContent(content)
}

// Content returns this Span's underlying handler.
func (s *Span) Content() tui.Handler {
	return s.handler
}
