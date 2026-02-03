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

package browserapi

import (
	"sync"

	"github.com/unstablebuild/rune-go-sdk/handler"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// FuncHandler returns a Handler by wrapping a tui.Handler
// with an Close callback.
func FuncHandler(h tui.Handler, doClose func() error) Handler {
	return &closeHandler{Handler: h, doClose: doClose}
}

// NopHandler returns a Handler by wrapping a tui.Handler
// with an nop Close callback.
func NopHandler(h tui.Handler) Handler {
	return &closeHandler{Handler: h, doClose: func() error { return nil }}
}

// SyncHandler wraps the given handler and adds synchronized access.
func SyncHandler(locker sync.Locker, h Handler) Handler {
	return &syncHandler{handler: h, locker: locker}
}

// StaticFloating wraps a Handler and returns a Floating that always
// return the same Dimensions values.
func StaticFloating(h Handler, width, height int) Floating {
	return staticFloating{width: width, height: height, Handler: h}
}

// NopFloatingHandler wraps a handler.Floating and returns a Floating
// that does nothing when Close is called.
func NopFloatingHandler(h handler.Floating) Floating {
	return nopFloating{Floating: h}
}

// FuncFloatingHandler wraps a handler.Floating and returns a Floating
// that calls calls closeFn when Close is called.
func FuncFloatingHandler(h handler.Floating, closeFn func() error) Floating {
	return funcFloatingHandler{Floating: h, fn: closeFn}
}

// FuncFloating wraps a Handler and returns a Floating that
// calls dimFn when Dimensions is called.
func FuncFloating(h Handler, dimFn func() (int, int)) Floating {
	return funcFloating{Handler: h, fn: dimFn}
}

type funcFloatingHandler struct {
	handler.Floating
	fn func() error
}

func (f funcFloatingHandler) Close() error {
	return f.fn()
}

type funcFloating struct {
	Handler
	fn func() (int, int)
}

func (f funcFloating) Dimensions() (width, height int) {
	return f.fn()
}

type closeHandler struct {
	tui.Handler
	doClose func() error
}

func (h *closeHandler) Close() error {
	return h.doClose()
}

type staticFloating struct {
	Handler
	width, height int
}

func (s staticFloating) Dimensions() (int, int) {
	return s.width, s.height
}

type nopFloating struct {
	handler.Floating
}

func (n nopFloating) Close() error {
	return nil

}

type syncHandler struct {
	locker  sync.Locker
	handler Handler
}

func (s syncHandler) Resize(width, height int) {
	s.locker.Lock()
	defer s.locker.Unlock()
	s.handler.Resize(width, height)
}

func (s syncHandler) Draw(w term.Writer) {
	s.locker.Lock()
	defer s.locker.Unlock()
	s.handler.Draw(w)
}

func (s syncHandler) Handle(ev term.Event) (exit, handled bool) {
	s.locker.Lock()
	defer s.locker.Unlock()
	return s.handler.Handle(ev)
}

func (s syncHandler) Cursor() (c term.Coordinates, style term.CursorStyle, show bool) {
	s.locker.Lock()
	defer s.locker.Unlock()
	return s.handler.Cursor()
}

func (s syncHandler) Selection() (string, bool) {
	s.locker.Lock()
	defer s.locker.Unlock()
	return s.handler.Selection()
}

func (s syncHandler) Close() error {
	s.locker.Lock()
	defer s.locker.Unlock()
	return s.handler.Close()
}
