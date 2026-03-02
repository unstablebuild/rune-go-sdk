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

package inputbox

import (
	"github.com/unstablebuild/rune-go-sdk/mouse"
	"github.com/unstablebuild/rune-go-sdk/term"
)

// mouseDelegate adapts Handler to satisfy mouse.Delegate.
// A separate type is needed because Handler already has a
// Height(width int) int method (from Responsive) which
// conflicts with the parameterless Height() int required
// by mouse.Delegate.
type mouseDelegate struct {
	ib *Handler
}

var _ mouse.Delegate = (*mouseDelegate)(nil)

func (d *mouseDelegate) OnAction(term.Event, term.Coordinates, mouse.Action) bool {
	return false
}

func (d *mouseDelegate) SetSelectionStart(pos term.Coordinates) {
	idx := d.ib.coordToTextPos(pos)
	d.ib.selAnchor = idx
	d.ib.cursor = idx
	d.ib.updateVoffset()
}

func (d *mouseDelegate) SetSelectionEnd(pos term.Coordinates) {
	idx := d.ib.coordToTextPos(pos)
	d.ib.cursor = idx
	d.ib.updateVoffset()
}

func (d *mouseDelegate) ClearSelection() {
	d.ib.clearSelection()
}

func (d *mouseDelegate) SelectWordAt(pos term.Coordinates) {
	d.ib.selectWordAt(pos)
}

func (d *mouseDelegate) SelectLine(y int) {
	d.ib.selectLine(y)
}

func (d *mouseDelegate) ScrollUp(n int) bool {
	return d.ib.mouseScrollUp(n)
}

func (d *mouseDelegate) ScrollDown(n int) bool {
	return d.ib.mouseScrollDown(n)
}

func (d *mouseDelegate) Width() int  { return d.ib.width }
func (d *mouseDelegate) Height() int { return d.ib.height }

// coordToTextPos converts screen coordinates to a text index,
// accounting for voffset and prompt width. It is the reverse
// of textPosToCoords.
func (ib *Handler) coordToTextPos(pos term.Coordinates) int {
	if ib.width <= 0 {
		return 0
	}
	y := pos.Y + ib.voffset
	x := pos.X
	var idx int
	if ib.promptWidth == 0 {
		idx = y*ib.width + x
	} else {
		flc := ib.firstLineChars()
		if y == 0 {
			idx = x - ib.promptWidth
		} else {
			idx = flc + (y-1)*ib.width + x
		}
	}
	return max(0, min(idx, len(ib.text)))
}

func (ib *Handler) selectWordAt(pos term.Coordinates) {
	idx := ib.coordToTextPos(pos)
	if idx >= len(ib.text) || !isWordChar(ib.text[idx]) {
		return
	}
	start, end := idx, idx+1
	for start > 0 && isWordChar(ib.text[start-1]) {
		start--
	}
	for end < len(ib.text) && isWordChar(ib.text[end]) {
		end++
	}
	ib.selAnchor = start
	ib.cursor = end
	ib.updateVoffset()
}

func (ib *Handler) selectLine(y int) {
	if ib.width <= 0 {
		return
	}
	absY := y + ib.voffset
	var start, end int
	if ib.promptWidth == 0 {
		start = absY * ib.width
		end = start + ib.width
	} else {
		flc := ib.firstLineChars()
		if absY == 0 {
			start = 0
			end = flc
		} else {
			start = flc + (absY-1)*ib.width
			end = start + ib.width
		}
	}
	start = max(0, min(start, len(ib.text)))
	end = max(start, min(end, len(ib.text)))
	ib.selAnchor = start
	ib.cursor = end
	ib.updateVoffset()
}

func (ib *Handler) mouseScrollUp(n int) bool {
	if ib.voffset <= 0 {
		return false
	}
	ib.voffset = max(0, ib.voffset-n)
	return true
}

func (ib *Handler) mouseScrollDown(n int) bool {
	maxOffset := max(0, ib.totalTextLines() - ib.height)
	if ib.voffset >= maxOffset {
		return false
	}
	ib.voffset = min(maxOffset, ib.voffset+n)
	return true
}
