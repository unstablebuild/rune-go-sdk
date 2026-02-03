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

func TestDrawOverlay(t *testing.T) {
	cfg := SpanConfig{
		PadVertical:      -5,
		PadHorizontal:    -10,
		ContentAlignment: AlignmentCentered,
	}
	background := &TestComponent{Ch: '*'}
	cover := NewFrame(NewStringWithConfig("a", StringConfig{Alignment: AlignmentCentered}))

	o := NewOverlay(background, cover, term.Attributes{}, cfg)

	w := term.NewStringWriter(16, 9)

	tests := []comptest.TestCase{
		{
			Action: func() { o.Resize(16, 9) }, Expected: `
****************
****************
***┌────────┐***
***│        │***
***│   a    │***
***│        │***
***└────────┘***
****************
****************`,
		}, {
			Action: func() { o.Resize(4, 4) }, Expected: `
┌──┐            
│a │            
│  │            
└──┘            
                
                
                
                
                `,
		}, {
			Action: func() { o.Resize(2, 2) }, Expected: `
a               
                
                
                
                
                
                
                
                `,
		},
	}

	comptest.TestComponent(t, o, w, tests)
}
