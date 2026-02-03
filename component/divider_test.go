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

func TestDrawDivider(t *testing.T) {
	f := Divider(0.5, StringConfig{
		Alignment:    AlignmentCentered,
		FrameCharSet: FrameCharSetDefault(),
	})

	f.Resize(8, 4)

	w := term.NewStringWriter(9, 5)

	tests := []comptest.TestCase{
		{
			Action: func() { f.Resize(4, 4) }, Expected: `
         
         
 ──      
         
         `,
		}, {
			Action: func() { f.Resize(2, 2) }, Expected: `
         
─        
         
         
         `,
		}, {
			Action: func() { f.Resize(8, 5) }, Expected: `
         
         
  ────   
         
         `,
		},
	}

	comptest.TestComponent(t, f, w, tests)
}

func TestDrawDividerNonUnit(t *testing.T) {
	f := Divider(0.5, StringConfig{
		Alignment:    AlignmentCentered,
		FrameCharSet: FrameCharSetDefault(),
	})

	f.Resize(8, 4)

	w := term.NewStringWriter(9, 5)

	tests := []comptest.TestCase{
		{
			Action: func() { f.Resize(4, 4) }, Expected: `
         
         
 ──      
         
         `,
		}, {
			Action: func() { f.Resize(2, 2) }, Expected: `
         
─        
         
         
         `,
		}, {
			Action: func() { f.Resize(8, 5) }, Expected: `
         
         
  ────   
         
         `,
		},
	}

	comptest.TestComponent(t, f, w, tests)
}
