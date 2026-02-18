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

func TestDrawBackgroundNoZero(t *testing.T) {
	u := &TestComponent{Ch: 'X'}
	s := NewSpan(u, SpanConfig{
		PadHorizontalPerc: 0.2,
		PadVerticalPerc:   0.2,
	})
	b := WithBackground(s, term.Cell{Ch: 'O'})

	b.Resize(9, 5)

	w := term.NewStringWriter(9, 5)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
XXXXXXXXO
XXXXXXXXO
XXXXXXXXO
XXXXXXXXO
OOOOOOOOO`,
		},
	}

	comptest.TestComponent(t, b, w, tests)
}

func TestResizeBackground(t *testing.T) {
	t.Run("zero width does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			b := WithBackground(&TestComponent{}, term.Cell{})
			b.Resize(0, 1)
		})
	})
	t.Run("zero height does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			b := WithBackground(&TestComponent{}, term.Cell{})
			b.Resize(1, 0)
		})
	})
	t.Run("negative width does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			b := WithBackground(&TestComponent{}, term.Cell{})
			b.Resize(-1, 5)
		})
	})
	t.Run("negative height does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			b := WithBackground(&TestComponent{}, term.Cell{})
			b.Resize(5, -1)
		})
	})
	t.Run("both negative does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			b := WithBackground(&TestComponent{}, term.Cell{})
			b.Resize(-1, -1)
		})
	})
}

func TestDrawBackground(t *testing.T) {
	u := &TestComponent{Ch: 'X'}
	s := NewSpan(u, SpanConfig{
		PadHorizontalPerc: 0.2,
		PadVerticalPerc:   0.2,
	})
	b := WithBackground(s, term.Cell{Ch: 0})

	b.Resize(9, 5)

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

	comptest.TestComponent(t, b, w, tests)
}
