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
)

func TestIntegrationString(t *testing.T) {
	width, height := 8, 4
	str := "AAAAAAAAAAAA\nBBBBBBBBBBBB\nCCCCCCCCCCCC\nDDDDDDDDDDDD"
	virtualScroll := Virtual[String]{C: NewString(str)}
	virtualScroll.Resize(width, height)

	w := term.NewStringWriter(12, height)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
AAAAAAAA    
BBBBBBBB    
CCCCCCCC    
DDDDDDDD    `,
		},
		{
			Action: func() {
				virtualScroll.Resize(8, 3)
				virtualScroll.Move(term.Coordinates{X: 1, Y: 1})
			}, Expected: `
            
 AAAAAAAA   
 BBBBBBBB   
 CCCCCCCC   `,
		},
		{
			Action: func() {
				virtualScroll.Resize(8, 4)
				virtualScroll.Move(term.Coordinates{X: 3, Y: 0})
			}, Expected: `
   AAAAAAAA 
   BBBBBBBB 
   CCCCCCCC 
   DDDDDDDD `,
		},
	}

	comptest.TestComponent(t, &virtualScroll, w, tests)
}

func TestVirtualDraw(t *testing.T) {
	width, height := 4, 4
	v := Virtual[*TestComponent]{C: &TestComponent{Ch: '$'}}
	v.Resize(width, height)

	w := term.NewStringWriter(width, height)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
$$$$
$$$$
$$$$
$$$$`,
		},
		{
			Action: func() {
				v.Resize(3, 3)
				v.Move(term.Coordinates{X: 1, Y: 1})
			}, Expected: `
    
 $$$
 $$$
 $$$`,
		},
	}
	comptest.TestComponent(t, &v, w, tests)
}
