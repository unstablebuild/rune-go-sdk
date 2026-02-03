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
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unstablebuild/rune-go-sdk/component/comptest"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
	"github.com/unstablebuild/tcell/v3"
)

func TestAsync(t *testing.T) {
	t.Run("draws placeholder if contstructor is not done", func(t *testing.T) {
		interrupt := term.NopInterrupter()
		a := Async[tui.Component](interrupt, &TestComponent{Ch: 'a'},
			func() (tui.Component, error) {
				<-t.Context().Done()
				return &TestComponent{Ch: 'x'}, nil
			})
		a.Resize(20, 9)
		w := term.NewStringWriter(20, 9)
		tests := []comptest.TestCase{
			{
				Action: nil, Expected: `
aaaaaaaaaaaaaaaaaaaa
aaaaaaaaaaaaaaaaaaaa
aaaaaaaaaaaaaaaaaaaa
aaaaaaaaaaaaaaaaaaaa
aaaaaaaaaaaaaaaaaaaa
aaaaaaaaaaaaaaaaaaaa
aaaaaaaaaaaaaaaaaaaa
aaaaaaaaaaaaaaaaaaaa
aaaaaaaaaaaaaaaaaaaa`,
			},
		}
		comptest.TestComponent(t, a, w, tests)
	})

	t.Run("draws constructed component if done", func(t *testing.T) {
		var sema sync.Mutex
		sema.Lock()
		interrupt := term.FuncInterrupter(func(context.Context) error {
			sema.Unlock()
			return nil
		})
		a := Async[tui.Component](interrupt, &TestComponent{Ch: 'a'},
			func() (tui.Component, error) {
				return &TestComponent{Ch: 'x'}, nil
			})
		a.Resize(8, 8)
		w := term.NewStringWriter(20, 9)
		tests := []comptest.TestCase{
			{
				Action: nil, Expected: `
xxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxx`,
			},
		}
		sema.Lock()
		a.Resize(20, 9)
		comptest.TestComponent(t, a, w, tests)
	})

	t.Run("draws error when component constructor errors", func(t *testing.T) {
		var sema sync.Mutex
		sema.Lock()
		interrupt := term.FuncInterrupter(func(context.Context) error {
			sema.Unlock()
			return nil
		})
		a := Async[tui.Component](interrupt, &TestComponent{Ch: 'a'},
			func() (tui.Component, error) {
				return nil, errors.New("boom")
			})
		w := term.NewStringWriter(20, 9)
		tests := []comptest.TestCase{
			{
				Action: nil, Expected: `
                    
                    
            ___     
                    
  /___/\_           
                    
          _\        
  \/_/\__           
                    `,
			},
		}
		sema.Lock()
		a.Resize(20, 9)
		comptest.TestComponent(t, a, w, tests)
	})

	t.Run("Height", func(t *testing.T) {
		var sema sync.Mutex
		sema.Lock()
		interrupt := term.FuncInterrupter(func(context.Context) error {
			sema.Unlock()
			return nil
		})
		a := Async[Responsive](interrupt, &TestResponsive{},
			func() (Responsive, error) {
				return &TestResponsive{WantHeight: 10}, nil
			})
		sema.Lock()
		assert.Equal(t, 10, a.Height(99))
	})

	t.Run("Dimensions", func(t *testing.T) {
		var sema sync.Mutex
		sema.Lock()
		interrupt := term.FuncInterrupter(func(context.Context) error {
			sema.Unlock()
			return nil
		})
		a := Async[Floating](interrupt, &TestResponsive{},
			func() (Floating, error) {
				return &TestResponsive{WantWidth: 10, WantHeight: 10}, nil
			})
		sema.Lock()
		width, height := a.Dimensions()
		assert.Equal(t, 10, width)
		assert.Equal(t, 10, height)
	})

	t.Run("SetAttr", func(t *testing.T) {
		var sema sync.Mutex
		sema.Lock()
		interrupt := term.FuncInterrupter(func(context.Context) error {
			sema.Unlock()
			return nil
		})
		attrs := term.Attributes{Fg: tcell.ColorYellow, Bg: tcell.ColorBlue}
		a := Async[WithAttributes](interrupt, &TestComponent{},
			func() (WithAttributes, error) {
				return &TestComponent{Attributes: attrs}, nil
			})
		sema.Lock()
		repv := a.SetAttr(term.Attributes{})
		assert.Equal(t, tcell.ColorBlue, repv.Bg)
		assert.Equal(t, tcell.ColorYellow, repv.Fg)
	})

	t.Run("SetAttr is called on constructed after constructor returns", func(t *testing.T) {
		var sema1, sema2 sync.Mutex
		sema1.Lock()
		sema2.Lock()
		interrupt := term.FuncInterrupter(func(context.Context) error {
			sema2.Unlock()
			return nil
		})
		tc := &TestComponent{}
		a := Async[WithAttributes](interrupt, &TestComponent{},
			func() (WithAttributes, error) {
				sema1.Lock()
				return tc, nil
			})
		attrs := term.Attributes{Fg: tcell.ColorYellow, Bg: tcell.ColorBlue}
		a.SetAttr(attrs)
		sema1.Unlock()
		sema2.Lock()
		assert.Equal(t, attrs, tc.Attributes)
	})
}
