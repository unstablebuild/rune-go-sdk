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
	"bufio"
	"bytes"
	"io"
	"math"
	"strings"

	"github.com/unstablebuild/rune-go-sdk/term/graphemecluster"
)

// CellsToString returns the string representation of the given cell matrix.
func CellsToString(cells [][]Cell) string {
	builder := strings.Builder{}
	CellsToStringBuilder(&builder, cells)
	return builder.String()
}

// CellsToStringBuilder copies the string representation of the given cell matrix
// to the supplied builder.
//
// Caller is responsible for resetting builder prior to this call if necessary.
func CellsToStringBuilder(builder *strings.Builder, cells [][]Cell) {
	copyToBuilder(builder, cells)
}

// CellsToBytesBuffer copies the bytes representation of the given cell matrix
// to the supplied buffer.
//
// Caller is responsible for resetting buffer prior to this call if necessary.
func CellsToBytesBuffer(buffer *bytes.Buffer, cells [][]Cell) {
	copyToBuffer(buffer, cells)
}

// StringToCells returns the cell matrix representation of the given string.
func StringToCells(str string) (cells [][]Cell) {
	var builder rawCells
	builder.init()
	_, _ = builder.ReadFrom(strings.NewReader(str))
	return builder.cells
}

// CloneCells returns a deep clone of in.
func CloneCells(in [][]Cell) [][]Cell {
	ret := make([][]Cell, len(in))
	for i, r := range in {
		ret[i] = make([]Cell, len(r))
		copy(ret[i], r)
	}
	return ret
}

// CopyCells copies src into dst, re-using dst's capacity when possible.
func CopyCells(dst [][]Cell, src [][]Cell) [][]Cell {
	if len(dst) > len(src) {
		dst = dst[:len(src)]
	} else if len(dst) < len(src) {
		for i := len(dst); i < len(src); i++ {
			dst = append(dst, nil) // will be replaced with a proper row below
		}
	}

	for i := range len(src) {
		if cap(dst[i]) < len(src[i]) {
			dst[i] = make([]Cell, len(src[i]))
		} else if dst[i] == nil { // cap=0, len=0; set to nil above
			dst[i] = make([]Cell, 0)
		} else {
			dst[i] = dst[i][:len(src[i])]
		}
		copy(dst[i], src[i])
	}
	return dst
}

// CalculateOptimalWidth calculates the width of this cells,
// such that nothing is truncated if rendered.
func CalculateOptimalWidth(cells [][]Cell) (max int) {
	for _, row := range cells {
		var rowcount int
		for _, c := range row {
			rowcount += int(c.Width)
		}
		if rowcount > max {
			max = rowcount
		}
	}
	return
}

func nextWrite(c [][]Cell) Coordinates {
	y := len(c) - 1
	x := len(c[y])
	return Coordinates{X: x, Y: y}
}

func copyToBuilder(builder *strings.Builder, cells [][]Cell) {
	for i, r := range cells {
		if i != 0 {
			builder.WriteByte('\n')
		}
		copyRowToBuilder(builder, r)
	}
}

func copyRowToBuilder(builder *strings.Builder, cells []Cell) {
	builder.Grow(len(cells)) // almost every time this is exact
	for _, c := range cells {
		builder.WriteRune(c.Ch)
		if c.Combining != nil {
			for _, comb := range *c.Combining {
				builder.WriteRune(comb)
			}
		}
	}
}

func copyToBuffer(builder *bytes.Buffer, cells [][]Cell) {
	for i, r := range cells {
		if i != 0 {
			builder.WriteByte('\n')
		}
		copyRowToBuffer(builder, r)
	}
}

func copyRowToBuffer(builder *bytes.Buffer, cells []Cell) {
	builder.Grow(len(cells)) // almost every time this is exact
	for _, c := range cells {
		builder.WriteRune(c.Ch)
		if c.Combining != nil {
			for _, comb := range *c.Combining {
				builder.WriteRune(comb)
			}
		}
	}
}

const (
	defColumnCap int = 64
	defRowCap    int = 64
)

// rawCells is a matrix of term.Cell.
type rawCells struct {
	columnCap int
	rowCap    int
	cells     [][]Cell
	zwj       bool
	zwjPos    Coordinates
}

// init initializes this rawCells with the given tabspaces config and resets its contents.
func (c *rawCells) init() {
	c.resetWithCap(defRowCap, defColumnCap)
}

func (c *rawCells) resetWithCap(rowCap, columnCap int) {
	c.columnCap = int(math.Max(float64(columnCap), float64(defColumnCap)))
	c.rowCap = int(math.Max(float64(rowCap), float64(defRowCap)))
	c.cells = make([][]Cell, 1, c.rowCap)
	c.cells[0] = makeNewRow(0, c.columnCap)
	c.zwj = false
	c.zwjPos = Coordinates{}
}

func makeNewRow(length, capacity int) (row []Cell) {
	capacity = int(math.Max(float64(length), float64(capacity)))
	row = make([]Cell, length, capacity)
	return
}

func (c *rawCells) ReadFrom(r io.Reader) (int64, error) {
	rowY := nextWrite(c.cells).Y
	reader := bufio.NewReader(r)
	n := int64(0)
	for {
		str, err := reader.ReadString('\n')
		state := -1
		var cluster string
		var width, byteCount uint8
		n += int64(len([]byte(str)))
		for len(str) > 0 {
			cluster, str, width, state = graphemecluster.StepString(str, state)
			r := []rune(cluster)
			byteCount = uint8(len([]byte(cluster)))
			switch r[0] {
			case '\n':
				c.cells = append(c.cells, makeNewRow(0, c.columnCap))
				rowY++
			default:
				cell := Cell{
					Ch:    r[0],
					Width: width,
					Bytes: byteCount,
				}
				if len(r) > 1 {
					comb := r[1:]
					cell.Combining = &comb
				}
				c.cells[rowY] = append(c.cells[rowY], cell)
			}
		}
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return n, err
		}
	}
}
