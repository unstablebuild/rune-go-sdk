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
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestCellSize(t *testing.T) {
	assert.Equal(t, uintptr(24), unsafe.Sizeof(Cell{}))
}

func TestCellAttributesRoundTrip(t *testing.T) {
	attr := Attributes{Fg: ColorRed, Bg: ColorYellow, Attrs: AttrBold}
	var c Cell
	c.SetAttributes(attr)
	assert.Equal(t, attr, c.Attributes())
	assert.Equal(t, Style(attr), c.Style())

	c2 := NewCell('x', 1, attr)
	assert.Equal(t, attr, c2.Attributes())
	assert.Equal(t, 'x', c2.Ch)
	assert.Equal(t, uint8(1), c2.Width)
}
