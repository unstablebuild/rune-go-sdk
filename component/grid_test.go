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
	"testing"

	"github.com/unstablebuild/rune-go-sdk/component/comptest"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

func TestGridDraw(t *testing.T) {
	t.Run("zero matrix", func(t *testing.T) {
		l := Grid(nil)
		l.Resize(4, 4)
		w := term.NewStringWriter(20, 9)
		tests := []comptest.TestCase{
			{
				Action: nil, Expected: `
                    
                    
                    
                    
                    
                    
                    
                    
                    `,
			},
		}
		comptest.TestComponent(t, l, w, tests)
	})

	t.Run("happy path", func(t *testing.T) {
		l := Grid([][]tui.Component{
			{&TestComponent{Ch: 'a'}, &TestComponent{Ch: 'b'}},
			{&TestComponent{Ch: 'c'}, &TestComponent{Ch: 'd'}},
			{&TestComponent{Ch: 'e'}, &TestComponent{Ch: 'f'}},
		})
		l.Resize(20, 4)

		w := term.NewStringWriter(20, 9)

		tests := []comptest.TestCase{
			{
				Action: nil, Expected: `
aaaaaaaaaabbbbbbbbbb
ccccccccccdddddddddd
eeeeeeeeeeffffffffff
                    
                    
                    
                    
                    
                    `,
			}, {
				Action: func() { l.Resize(20, 9) }, Expected: `
aaaaaaaaaabbbbbbbbbb
aaaaaaaaaabbbbbbbbbb
aaaaaaaaaabbbbbbbbbb
ccccccccccdddddddddd
ccccccccccdddddddddd
ccccccccccdddddddddd
eeeeeeeeeeffffffffff
eeeeeeeeeeffffffffff
eeeeeeeeeeffffffffff`,
			}, {
				Action: func() { l.Resize(2, 2) }, Expected: `
                    
                    
                    
                    
                    
                    
                    
                    
                    `,
			},
		}
		comptest.TestComponent(t, l, w, tests)
	})

	t.Run("zero row", func(t *testing.T) {
		l := Grid([][]tui.Component{
			{&TestComponent{Ch: 'a'}},
			nil,
			{&TestComponent{Ch: 'b'}},
		})
		l.Resize(4, 4)
		w := term.NewStringWriter(20, 9)
		tests := []comptest.TestCase{
			{
				Action: nil, Expected: `
aaaa                
                    
bbbb                
                    
                    
                    
                    
                    
                    `,
			},
		}
		comptest.TestComponent(t, l, w, tests)
	})

}
