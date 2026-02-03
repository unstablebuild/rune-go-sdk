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
)

func TestRowHeight(t *testing.T) {
	t.Run("takes advantage of the full width", func(t *testing.T) {
		row := NewRow()
		row.AddComponent(NewResponsiveString("123456789", StringResponsiveConfig{}), MaxCols)
		assert.Equal(t, 1, row.Height(9))
	})
	t.Run("passes the correct width to components", func(t *testing.T) {
		row := NewRow()
		row.AddComponent(NewResponsiveString("123456789", StringResponsiveConfig{}), MaxCols/2)
		assert.Equal(t, 3, row.Height(9))
	})
	t.Run("returns the max height", func(t *testing.T) {
		row := NewRow()
		row.AddComponent(NewResponsiveString("12", StringResponsiveConfig{}), MaxCols/2)
		row.AddComponent(NewResponsiveString("123456789", StringResponsiveConfig{}), MaxCols/2)
		assert.Equal(t, 3, row.Height(9))
	})
}

func TestRowDimensions(t *testing.T) {
	row := NewRow()
	row.AddComponent(testResponsiveWidth('a', 4, 10), MaxCols/2)
	row.AddComponent(testResponsiveWidth('b', 8, 8), MaxCols/2)
	actualWidth, actualHeight := row.Dimensions()
	assert.Equal(t, 12, actualWidth)
	assert.Equal(t, 10, actualHeight)
}

func testResponsiveWidth(ch rune, wantWidth, wantHeight int) Responsive {
	return &TestResponsive{
		WantWidth:  wantWidth,
		WantHeight: wantHeight,
		TestComponent: TestComponent{
			Ch: ch,
		},
	}
}
