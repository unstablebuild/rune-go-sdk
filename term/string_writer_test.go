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


package term

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// this test is not very thorough because StringWriter is used by
// many tui.Component tests and so it's already indirectly tested.
func TestStringWriter(t *testing.T) {
	width, height := 5, 6
	writer := NewStringWriter(width, height)

	c := 'A'
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			if i > j-1 {
				writer.SetCell(Coordinates{X: j, Y: i}, Cell{Ch: c})
			}
		}
		c++
	}

	// we use StringWriter mostly for tests so it's important that it panics on oob
	assert.Panics(t, func() {
		writer.SetCell(Coordinates{X: width + 1, Y: height + 1}, Cell{Ch: '='})
	})

	if err := writer.Flush(); err != nil {
		t.Fatal(err)
	}

	expected := "A    \nBB   \nCCC  \nDDDD \nEEEEE\n     "
	assert.Equal(t, expected, writer.String())
}
