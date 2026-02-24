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
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

var _ tui.Component = (*Container)(nil)

// Container provides means to position components in a viewport,
// through a set of Rows that fill the available width and
// stack on top of each other. See Row for more details.
//
// A zero-valued container is ready for use.
type Container struct {
	rows          []*Row
	width, height int
	offset        int
}

// NewContainer allocates storage for a new container and initializes it.
func NewContainer() *Container {
	ret := new(Container)
	return ret
}

// AddRow appends a row to this Container.
func (c *Container) AddRow() *Row {
	row := NewRow()
	c.rows = append(c.rows, row)
	return row
}

// Resize satisfies tui.Component.
func (c *Container) Resize(width, height int) {
	c.width, c.height = width, height
	if mo := c.maxOffset(); c.offset > mo {
		c.offset = mo
	}
	offset := -c.offset
	for _, row := range c.rows {
		rowHeight := row.Height(width)
		pos := term.Coordinates{Y: offset}
		row.Move(pos)
		if offset == height {
			// move above is enough to ensure that
			// virtual writer does not draw anything
			continue
		}
		if offset+rowHeight >= height {
			rowHeight = height - offset
		}
		// do not resize if it's before start
		if offset+rowHeight >= 0 && rowHeight > 0 {
			row.Resize(width, rowHeight)
		}
		offset += rowHeight
	}
}

// Draw satisfies tui.Component.
func (c *Container) Draw(w term.Writer) {
	// Height requirements could have changed
	// and that's something that we cannot determine from here
	// so it's better to always resize.
	c.Resize(c.width, c.height)

	// provides additional SetCell clipping for components past height
	vw := VirtualWriter{w, term.Coordinates{}, c.height, c.width}
	for _, row := range c.rows {
		row.Draw(&vw)
	}
}

// ScrollUp scrolls the contents of this container up.
func (c *Container) ScrollUp() bool {
	if c.offset == 0 {
		return false
	}
	c.offset--
	return true
}

// ScrollDown scrolls the contents of this container down.
func (c *Container) ScrollDown() bool {
	if c.offset == c.maxOffset() {
		return false
	}
	c.offset++
	return true
}

// ScrollToBottom scrolls the contents of this container
// to the bottom.
func (c *Container) ScrollToBottom() {
	c.offset = c.maxOffset()
}

// Height satisfies component.Responsive.
func (c *Container) Height(width int) (height int) {
	for _, row := range c.rows {
		height += row.Height(width)
	}
	return
}

// Dimensions satisfies component.Floating. If underlying
// components do not satisfy component.Floating, then this method panics.
func (r *Container) Dimensions() (retWidth int, retHeight int) {
	for _, row := range r.rows {
		width, height := row.Dimensions()
		if width > retWidth {
			retWidth = width
		}
		retHeight += height
	}
	return
}

func (c *Container) maxOffset() int {
	totalHeight := 0
	for _, row := range c.rows {
		totalHeight += row.Height(c.width)
	}
	return max(0, totalHeight-c.height)
}
