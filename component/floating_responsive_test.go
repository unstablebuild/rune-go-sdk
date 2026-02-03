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

func TestFloatingResponsiveDraw(t *testing.T) {
	l := NewAspectRatioFloatingResponsive(NewResponsiveString("mixtral-8x7b instruct-v0.1.Q4_K_M",
		StringResponsiveConfig{
			NoSplitWords: true,
			StringConfig: StringConfig{
				Alignment: AlignmentCentered,
			},
		}), 4.0/3.0)

	l.Resize(20, 4)

	w := term.NewStringWriter(20, 9)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
                    
mixtral-8x7b        
instruct-v0.1.Q4_K_M
                    
                    
                    
                    
                    
                    `,
		},
		{
			Action: func() {
				width, height := l.Dimensions()
				assert.Equal(t, 7, width)
				assert.Equal(t, 5, height)
				l.Resize(7, 5)
			}, Expected: `
mixtral             
-8x7b               
instruc             
t-v0.1.             
Q4_K_M              
                    
                    
                    
                    `,
		},
	}
	comptest.TestComponent(t, l, w, tests)

}

func TestFloatingResponsiveLoss(t *testing.T) {
	suite := []struct {
		inAspectRatio float64
		inWidth       int
		inHeight      int
		expectedLoss  float64
		expectedOk    bool
	}{
		{4.0 / 3.0, 400, 300, 0, true},
		{16.0 / 9.0, 400, 300, 0.25, false},
	}
	for _, test := range suite {
		l := NewAspectRatioFloatingResponsive(testResponsive('a', 10 /* nop */), test.inAspectRatio)
		actualLoss, ok := l.calculateLoss(test.inWidth, test.inHeight)
		assert.Equal(t, test.expectedLoss, actualLoss)
		assert.Equal(t, test.expectedOk, ok)
	}
}

func TestFloatingResponsiveDimensionsStatic(t *testing.T) {
	suite := []struct {
		inAspectRatio  float64
		inWidth        int
		inHeight       int
		expectedWidth  int
		expectedHeight int
	}{
		{4.0 / 3.0, 400, 300, 360, 300},
		{16.0 / 9.0, 400, 300, 480, 300},
		{4.0 / 3.0, 9, 3, 6, 3},
	}
	for _, test := range suite {
		root := testResponsiveWidth('a', test.inWidth, test.inHeight)
		l := NewAspectRatioFloatingResponsive(root, test.inAspectRatio)
		actualWidth, actualHeight := l.Dimensions()
		assert.Equal(t, test.expectedWidth, actualWidth)
		assert.Equal(t, test.expectedHeight, actualHeight)
	}
}

func TestFloatingResponsiveDimensionsDynamic(t *testing.T) {
	suite := []struct {
		inAspectRatio  float64
		expectedWidth  int
		expectedHeight int
		inStr          string
	}{
		{4.0 / 3.0, 7, 5, "mixtral-8x7b instruct-v0.1.Q4_K_M"},
		{4.0 / 3.0, 8, 6, "aaaaaaaaa\naaaaaaaaa\naaaaaaaaa"},
		{4.0 / 3.0, 6, 3, "a\na\na"},
		{4.0 / 3.0, 16, 12, "a\na\na\na\na\na\na\na\na\na\na\na"},
		{4.0 / 3.0, 14, 12, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		{4.0 / 3.0, 36, 28, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	}
	for _, test := range suite {
		root := NewResponsiveString(test.inStr, StringResponsiveConfig{
			NoSplitWords: true,
			StringConfig: StringConfig{
				Alignment: AlignmentCentered,
			},
		})
		l := NewAspectRatioFloatingResponsive(root, test.inAspectRatio)
		actualWidth, actualHeight := l.Dimensions()
		assert.Equal(t, test.expectedWidth, actualWidth)
		assert.Equal(t, test.expectedHeight, actualHeight)
	}
}

func TestFloatingResponsiveIntegrationStringWithPadding(t *testing.T) {
	messageResponsive := NewResponsiveString("Do you want to restore the previous session?",
		StringResponsiveConfig{
			NoSplitWords: true,
			StringConfig: StringConfig{
				PaddingVertical:   4,
				PaddingHorizontal: 2,
				Alignment:         AlignmentCentered,
			}})
	l := NewAspectRatioFloatingResponsive(messageResponsive, DefaultAspectRatio)
	actualWidth, actualHeight := l.Dimensions()
	assert.Equal(t, 26, actualWidth)
	assert.Equal(t, 6, actualHeight)
}

func TestFloatingResponsiveEdgeCases(t *testing.T) {

	t.Run("zero aspect ratio panics", func(t *testing.T) {
		assert.Panics(t, func() {
			NewAspectRatioFloatingResponsive(testResponsive('a', 10), 0)
		})
	})

	t.Run("inner component wants 0", func(t *testing.T) {
		l := NewAspectRatioFloatingResponsive(testResponsive('a', 0), 4.0/3.0)
		l.Resize(20, 4)
		assert.NotPanics(t, func() {
			l.Dimensions()
		})
	})

	t.Run("inner component wants large number", func(t *testing.T) {
		l := NewAspectRatioFloatingResponsive(testResponsive('a', 1000), 4.0/3.0)
		l.Resize(20, 4)
		width, height := l.Dimensions()
		assert.Equal(t, 1200, width)
		assert.Equal(t, 1000, height)
	})
}
