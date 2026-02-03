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
	"context"
)

type boundsCheckWriter struct {
	height int
	width  int
	w      Writer
}

// BoundsCheckWriter returns a Writer which wraps w to make sure that
// calls to SetCell and SetCursor will never be out of the bounds defined by height
// or width.
func BoundsCheckWriter(width, height int, w Writer) Writer {
	return boundsCheckWriter{height: height, width: width, w: w}
}

func (p boundsCheckWriter) SetCell(pos Coordinates, c Cell) {
	if p.outOfBounds(pos) {
		return
	}
	p.w.SetCell(pos, c)
}

func (p boundsCheckWriter) UnionAttributes(pos Coordinates, attr Attributes) {
	if p.outOfBounds(pos) {
		return
	}
	p.w.UnionAttributes(pos, attr)
}

func (p boundsCheckWriter) Context() context.Context {
	return p.w.Context()
}

func (p boundsCheckWriter) outOfBounds(pos Coordinates) bool {
	return pos.X >= p.width || pos.Y >= p.height || pos.X < 0 || pos.Y < 0
}
