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
	"context"
	"fmt"
	"sync/atomic"

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Async runs the given constructor in a separate goroutine and renders
// a placeholder until it returns and the returned component is installed.
//
// Note that if Async's type parameter is an uknown flavor of tui.Component,
// it will panic. Only Reponsive, Floating, WithAttributes are allowed.
// This design choice might feel un-ergonomic at first, but it's better to
// panic at the constructor than when SetAttr/Dimensions/Height panics at
// runtime because the underlying component doesn't implement them.
func Async[T tui.Component](
	interrupter term.Interrupter, placeholder T, fn func() (T, error),
) T {
	ret := &async[T]{fn: fn, placeholder: placeholder}
	go ret.callFunc(interrupter)
	// if this panics, then it's a developer error, since
	// we are expecting async to implement an interface that it doesn't
	return (any)(ret).(T)
}

// async is a flavor of tui.Component that runs a constructor asynchronously.
type async[T tui.Component] struct {
	fn          func() (T, error)
	placeholder T
	done        atomic.Value
	attr        atomic.Value
	width       atomic.Int32
	height      atomic.Int32
}

// Dimensions satisfies Floating.
func (a *async[T]) Dimensions() (width int, height int) {
	if done := a.done.Load(); done != nil {
		// Dimensions can only be called if T does indeed satisfy it
		return done.(Floating).Dimensions()
	}
	return (any)(a.placeholder).(Floating).Dimensions()
}

// SetAttr satisfies WithAttributes.
func (a *async[T]) SetAttr(attr term.Attributes) term.Attributes {
	if done := a.done.Load(); done != nil {
		return done.(WithAttributes).SetAttr(attr)
	}
	a.attr.Store(&attr)
	return (any)(a.placeholder).(WithAttributes).SetAttr(attr)
}

// Height satisfies Responsive.
func (a *async[T]) Height(width int) int {
	if done := a.done.Load(); done != nil {
		return done.(Responsive).Height(width)
	}
	return (any)(a.placeholder).(Responsive).Height(width)
}

// Resize satisfies tui.Component.
func (a *async[T]) Resize(width, height int) {
	if done := a.done.Load(); done != nil {
		done.(tui.Component).Resize(width, height)
		return
	}
	a.width.Store(int32(width))
	a.height.Store(int32(height))
	a.placeholder.Resize(width, height)
}

// Draw satisfies tui.Component.
func (a *async[T]) Draw(w term.Writer) {
	if done := a.done.Load(); done != nil {
		done.(tui.Component).Draw(w)
		return
	}
	a.placeholder.Draw(w)
}

// IsAsyncContext determines if context belongs to an Async component interrupting.
func IsAsyncContext(ctx context.Context) bool {
	val, ok := ctx.Value(pKey).(string)
	return ok && val == asyncContextValue
}

const (
	smtgWrongCopy = `
             
          ___
         /___/\_               
        _\   \/_/\__           
      __\       \/_/\          
      \   __    __ \ \         
     __\  \_\   \_\ \ \   __   
    /_/\\   __   __  \ \_/_/\  
    \_\/_\__\/\__\/\__\/_\_\/  
       \_\/_/\       /_\_\/    
          \_\/       \_\/      


%v
`
)

type ctxKey int

var pKey ctxKey

func asyncContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, pKey, asyncContextValue)
}

var asyncContextValue string = "__async"

func (a *async[T]) callFunc(interrupter term.Interrupter) {
	defer interrupter.Interrupt(asyncContext(context.Background())) //nolint:errcheck
	t, err := a.fn()
	width, height := int(a.width.Load()), int(a.height.Load())
	if err != nil {
		errComp := makeErrorComponent(err)
		errComp.Resize(width, height)
		a.done.Store(errComp)
		return
	}
	var val tui.Component = t
	val.Resize(width, height)
	if attr := a.attr.Load(); attr != nil {
		val.(WithAttributes).SetAttr(*attr.(*term.Attributes))
	}
	a.done.Store(val)
}

func makeErrorComponent(err error) tui.Component {
	cfg := StringResponsiveConfig{
		NoSplitWords: true,
		StringConfig: StringConfig{
			PaddingVertical:   4,
			PaddingHorizontal: 4,
			Alignment:         AlignmentCentered,
			//BackgroundAttributes: p.cfg.BackgroundAttributes,
			//Attributes:           term.Attributes{Bg: p.cfg.BackgroundAttributes.Bg},
		},
	}
	return NewResponsiveString(fmt.Sprintf(smtgWrongCopy, err), cfg)
}
