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

	"github.com/stretchr/testify/assert"
	"github.com/unstablebuild/rune-go-sdk/component/comptest"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestInlineLeft(t *testing.T) {
	t.Run("nil slice doesn't panic", func(t *testing.T) {
		l := Inline(nil, AlignmentLeft)
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
		l := Inline([]Floating{
			&TestResponsive{WantWidth: 10, TestComponent: TestComponent{Ch: 'a'}},
			&TestResponsive{WantWidth: 5, TestComponent: TestComponent{Ch: 'b'}},
			&TestResponsive{WantWidth: 8, TestComponent: TestComponent{Ch: 'c'}},
			&TestResponsive{WantWidth: 1, TestComponent: TestComponent{Ch: 'd'}},
		}, AlignmentLeft)
		actualWidth, actualHeight := l.Dimensions()
		assert.Equal(t, 24, actualWidth)
		assert.Equal(t, 0, actualHeight)
		l.Resize(20, 4)

		w := term.NewStringWriter(20, 9)

		tests := []comptest.TestCase{
			{
				Action: nil, Expected: `
aaaaaaaaaabbbbbccccc
aaaaaaaaaabbbbbccccc
aaaaaaaaaabbbbbccccc
aaaaaaaaaabbbbbccccc
                    
                    
                    
                    
                    `,
			}, {
				Action: func() { l.Resize(11, 9) }, Expected: `
aaaaaaaaaab         
aaaaaaaaaab         
aaaaaaaaaab         
aaaaaaaaaab         
aaaaaaaaaab         
aaaaaaaaaab         
aaaaaaaaaab         
aaaaaaaaaab         
aaaaaaaaaab         `,
			}, {
				Action: func() { l.Resize(2, 2) }, Expected: `
aa                  
aa                  
                    
                    
                    
                    
                    
                    
                    `,
			},
		}
		comptest.TestComponent(t, l, w, tests)
	})
}
