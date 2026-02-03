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

func TestReferenceDraw(t *testing.T) {
	s := NewReference(nil)

	// valid sequence of calls
	s.Resize(8, 4)

	s.Init(&TestComponent{Ch: 'X'})

	w := term.NewStringWriter(9, 5)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
XXXXXXXX 
XXXXXXXX 
XXXXXXXX 
XXXXXXXX 
         `,
		},
	}
	comptest.TestComponent(t, s, w, tests)

	s.Init(&TestComponent{Ch: 'Y'})

	tests = []comptest.TestCase{
		{
			Action: nil, Expected: `
YYYYYYYY 
YYYYYYYY 
YYYYYYYY 
YYYYYYYY 
         `,
		},
	}
	comptest.TestComponent(t, s, w, tests)
}
