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
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/component/comptest"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestSpanDimensions(t *testing.T) {
	t.Run("panics if underlying component is not Floating", func(t *testing.T) {
		assert.Panics(t, func() {
			comp := &TestComponent{}
			s := NewSpan(comp, SpanConfig{})
			s.Dimensions()
		})
	})
	t.Run("uses underlying Floating dimensions if padding is 0", func(t *testing.T) {
		comp := StaticFloating(&TestComponent{}, 10, 20)
		s := NewSpan(comp, SpanConfig{})
		width, height := s.Dimensions()
		assert.Equal(t, 10, width)
		assert.Equal(t, 20, height)
	})
	t.Run("adds absolute vertical padding from underlying component's returned Dimensions", func(t *testing.T) {
		comp := StaticFloating(&TestComponent{}, 10, 20)
		s := NewSpan(comp, SpanConfig{PadVertical: 2})
		width, height := s.Dimensions()
		assert.Equal(t, 10, width)
		assert.Equal(t, 22, height)
	})
	t.Run("subtracts absolute horizontal padding from underlying component's call to Dimensions", func(t *testing.T) {
		comp := StaticFloating(&TestComponent{}, 10, 20)
		s := NewSpan(comp, SpanConfig{PadHorizontal: 2})
		width, height := s.Dimensions()
		assert.Equal(t, 12, width)
		assert.Equal(t, 20, height)
	})
	t.Run("adds perc vertical padding from underlying component's returned Dimensions", func(t *testing.T) {
		comp := StaticFloating(&TestComponent{}, 10, 20)
		s := NewSpan(comp, SpanConfig{PadVerticalPerc: 0.2})
		// needs Resize or else it does vertical perc over 0
		s.Resize(10, 10)
		width, height := s.Dimensions()
		assert.Equal(t, 10, width)
		assert.Equal(t, 22, height)
	})
	t.Run("subtracts perc horizontal padding from underlying component's call to Dimensions", func(t *testing.T) {
		comp := StaticFloating(&TestComponent{}, 10, 20)
		s := NewSpan(comp, SpanConfig{PadHorizontalPerc: 0.1})
		width, height := s.Dimensions()
		assert.Equal(t, 11, width)
		assert.Equal(t, 20, height)
	})
	t.Run("adds relative horizontal padding and passes remaining with to underlying component's Dimensions", func(t *testing.T) {
		comp := StaticFloating(&TestComponent{}, 10, 20)
		s := NewSpan(comp, SpanConfig{PadHorizontal: -2})
		s.Resize(20, 20)
		width, height := s.Dimensions()
		assert.Equal(t, 18, width)
		assert.Equal(t, 20, height)
	})
}

func TestSpanHeight(t *testing.T) {
	t.Run("panics if underlying component is not Responsive", func(t *testing.T) {
		assert.Panics(t, func() {
			comp := &TestComponent{}
			s := NewSpan(comp, SpanConfig{})
			s.Height(1)
		})
	})
	t.Run("uses underlying Responsive Height if padding is 0", func(t *testing.T) {
		comp := &TestResponsive{WantHeight: 10}
		s := NewSpan(comp, SpanConfig{})
		assert.Equal(t, 10, s.Height(0))
	})
	t.Run("adds absolute vertical padding from underlying component's returned Height", func(t *testing.T) {
		comp := &TestResponsive{WantHeight: 10}
		s := NewSpan(comp, SpanConfig{PadVertical: 2})
		assert.Equal(t, 12, s.Height(0))
	})
	t.Run("subtracts absolute horizontal padding from underlying component's call to Height", func(t *testing.T) {
		comp := &TestResponsive{WantHeight: 10}
		s := NewSpan(comp, SpanConfig{PadHorizontal: 2})
		require.Equal(t, 10, s.Height(20))
		assert.Equal(t, 18, comp.PassedWidth)
	})
	t.Run("adds perc vertical padding from underlying component's returned Height", func(t *testing.T) {
		comp := &TestResponsive{WantHeight: 10}
		s := NewSpan(comp, SpanConfig{PadVerticalPerc: 0.2})
		// needs Resize or else it does vertical perc over 0
		s.Resize(10, 10)
		assert.Equal(t, 12, s.Height(0))
	})
	t.Run("subtracts perc horizontal padding from underlying component's call to Height", func(t *testing.T) {
		comp := &TestResponsive{WantHeight: 10}
		s := NewSpan(comp, SpanConfig{PadHorizontalPerc: 0.1})
		require.Equal(t, 10, s.Height(20))
		assert.Equal(t, 18, comp.PassedWidth)
	})
	t.Run("ignores relative PadVertical when underlying component is Responsive", func(t *testing.T) {
		comp := &TestResponsive{WantHeight: 10}
		s := NewSpan(comp, SpanConfig{PadVertical: -2})
		s.Resize(10, 10)
		assert.Equal(t, 10, s.Height(0))
	})
	t.Run("adds relative horizontal padding and passes remaining with to underlying component's Height", func(t *testing.T) {
		comp := &TestResponsive{WantHeight: 10}
		s := NewSpan(comp, SpanConfig{PadHorizontal: -2})
		s.Resize(20, 20)
		require.Equal(t, 10, s.Height(20))
		assert.Equal(t, 2, comp.PassedWidth)
	})
}

func TestSpanResizeNegative(t *testing.T) {
	t.Run("negative width does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			s := NewSpan(
				NewBackground(&TestComponent{}, term.Cell{}),
				DefaultSpanConfig(),
			)
			s.Resize(-1, 5)
		})
	})
	t.Run("negative height does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			s := NewSpan(
				NewBackground(&TestComponent{}, term.Cell{}),
				DefaultSpanConfig(),
			)
			s.Resize(5, -1)
		})
	})
}

func TestSpanPadAutoFloating(t *testing.T) {
	t.Run("panics if inner component does not satisfy Floating",
		func(t *testing.T) {
			assert.Panics(t, func() {
				NewSpan(&TestComponent{},
					SpanConfig{PadAutoFloating: true})
			})
		})
	t.Run("does not panic if inner component satisfies Floating",
		func(t *testing.T) {
			assert.NotPanics(t, func() {
				comp := StaticFloating(
					&TestComponent{Ch: '*'}, 4, 2)
				NewSpan(comp,
					SpanConfig{PadAutoFloating: true})
			})
		})
}

func TestDrawSpanPadAutoFloating(t *testing.T) {
	inner := StaticFloating(&TestComponent{Ch: '*'}, 4, 2)
	cfg := DefaultSpanConfig()
	cfg.PadAutoFloating = true
	s := NewSpan(inner, cfg)
	s.Resize(8, 4)

	w := term.NewStringWriter(9, 5)

	tests := []comptest.TestCase{
		{
			// inner (4x2) smaller than available (8x4)
			Action: nil, Expected: "\n" +
				"         \n" +
				"  ****   \n" +
				"  ****   \n" +
				"         \n" +
				"         ",
		}, {
			// alignment works: right+bottom
			Action: func() {
				s.cfg.ContentAlignment =
					AlignmentRight | AlignmentBottom
				s.Resize(8, 4)
			}, Expected: "\n" +
				"         \n" +
				"         \n" +
				"    **** \n" +
				"    **** \n" +
				"         ",
		}, {
			// inner (4x2) equal to available (4x2)
			Action: func() {
				s.cfg.ContentAlignment = AlignmentCentered
				s.Resize(4, 2)
			}, Expected: "\n" +
				"****     \n" +
				"****     \n" +
				"         \n" +
				"         \n" +
				"         ",
		}, {
			// inner (4x2) bigger than available (3x1)
			Action: func() {
				s.Resize(3, 1)
			}, Expected: "\n" +
				"***      \n" +
				"         \n" +
				"         \n" +
				"         \n" +
				"         ",
		}, {
			// width bigger, height smaller: only h-pad
			Action: func() {
				s.Resize(8, 1)
			}, Expected: "\n" +
				"  ****   \n" +
				"         \n" +
				"         \n" +
				"         \n" +
				"         ",
		},
	}

	comptest.TestComponent(t, s, w, tests)
}

func TestDrawDefaultSpan(t *testing.T) {
	u := &TestComponent{Ch: 'X'}
	cfg := DefaultSpanConfig()
	s := NewSpan(u, cfg)

	s.Resize(8, 4)

	w := term.NewStringWriter(9, 5)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
XXXXXXXX 
XXXXXXXX 
XXXXXXXX 
XXXXXXXX 
         `,
		}, {
			Action: func() { s.SetContent(&TestComponent{Ch: '*'}) }, Expected: `
******** 
******** 
******** 
******** 
         `,
		}, {
			Action: func() { s.cfg.PadHorizontalPerc, s.cfg.PadVerticalPerc = 0.5, 0.5; s.Resize(8, 4) }, Expected: `
         
  ****   
  ****   
         
         `,
		}, {
			Action: func() { s.Resize(4, 4) }, Expected: `
         
 **      
 **      
         
         `,
		}, {
			Action: func() { s.cfg.ContentAlignment = AlignmentRight; s.Resize(8, 4) }, Expected: `
    **** 
    **** 
         
         
         `,
		}, {
			Action: func() { s.cfg.ContentAlignment |= AlignmentBottom; s.Resize(8, 4) }, Expected: `
         
         
    **** 
    **** 
         `,
		}, {
			Action: func() { s.cfg.ContentAlignment = AlignmentLeft | AlignmentBottom; s.Resize(8, 4) }, Expected: `
         
         
****     
****     
         `,
		}, {
			Action: func() {
				s.cfg.ContentAlignment = AlignmentHorizontallyCentered | AlignmentBottom
				s.Resize(8, 4)
			}, Expected: `
         
         
  ****   
  ****   
         `,
		}, {
			Action: func() {
				s.cfg.ContentAlignment = AlignmentHorizontallyCentered | AlignmentTop
				s.Resize(8, 4)
			}, Expected: `
  ****   
  ****   
         
         
         `,
		}, {
			Action: func() {
				s.cfg.ContentAlignment = AlignmentHorizontallyCentered | AlignmentVerticallyCentered
				s.Resize(8, 4)
			}, Expected: `
         
  ****   
  ****   
         
         `,
		}, {
			Action: func() {
				s.cfg.ContentAlignment = AlignmentLeft | AlignmentVerticallyCentered
				s.Resize(8, 4)
			}, Expected: `
         
****     
****     
         
         `,
		}, {
			Action: func() {
				s.cfg.ContentAlignment = AlignmentRight | AlignmentVerticallyCentered
				s.Resize(8, 4)
			}, Expected: `
         
    **** 
    **** 
         
         `,
		}, {
			/* padding is removed if dimensions are < 3 */
			Action: func() {
				s.cfg.ContentAlignment = AlignmentRight | AlignmentVerticallyCentered
				s.Resize(2, 2)
			}, Expected: `
**       
**       
         
         
         `,
		}, {
			Action: func() {
				s.cfg.ContentAlignment = AlignmentCentered

				// Vertical/Horizontal override Perc
				s.cfg.PadVertical = 2
				s.cfg.PadHorizontal = 4

				s.Resize(8, 4)
			}, Expected: `
         
  ****   
  ****   
         
         `,
		}, {
			Action: func() {
				s.cfg.ContentAlignment = AlignmentCentered
				// Vertical/Horizontal greater than size
				s.cfg.PadVertical = 9
				s.cfg.PadHorizontal = 5
				s.Resize(8, 4)
			}, Expected: `
  ***    
  ***    
  ***    
  ***    
         `,
		}, {
			Action: func() {
				s.cfg.ContentAlignment = AlignmentCentered

				// Vertical/Horizontal negatie is used
				// as effective content width/height
				s.cfg.PadVertical = -4
				s.cfg.PadHorizontal = -6

				s.Resize(8, 4)
			}, Expected: `
 ******  
 ******  
 ******  
 ******  
         `,
		}, {
			Action: func() {
				s.cfg.ContentAlignment = AlignmentCentered
				// bigger than available
				s.cfg.PadVertical = -5
				s.cfg.PadHorizontal = -9

				s.Resize(8, 4)
			}, Expected: `
******** 
******** 
******** 
******** 
         `,
		},
	}

	comptest.TestComponent(t, s, w, tests)
}
