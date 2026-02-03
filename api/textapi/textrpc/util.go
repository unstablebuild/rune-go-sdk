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


package textrpc

import (
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/term"
	termrpc "github.com/unstablebuild/rune-go-sdk/term/termrpc"
)

// NewURIFromProto maps rpc.URI into a workspaceapi.URI.
func NewURIFromProto(u *URI) (workspaceapi.URI, error) {
	return workspaceapi.ParseURI(u.GetUri())
}

// NewURI maps a workspaceapi.URI into a rpc.URI.
func NewURI(u workspaceapi.URI) *URI {
	return &URI{Uri: u.String()}
}

// RawCellsResponseToBuffer converts an EditRequest into a cell.Buffer
func RawCellsResponseToBuffer(in *RawCellsResponse) [][]term.Cell {
	return rowsToBuffer(in.GetRows())
}

// NewRawCellsResponse converts a buf into an RawCellsResponse.
func NewRawCellsResponse(cells [][]term.Cell) *RawCellsResponse {
	return &RawCellsResponse{Rows: rawCellsToProtoCells(cells)}
}

func rowsToBuffer(in []*termrpc.CellRow) [][]term.Cell {
	var maxWidth int
	for _, row := range in {
		if len(row.Cells) > maxWidth {
			maxWidth = len(row.Cells)
		}
	}

	ret := make([][]term.Cell, len(in))
	for y, rows := range in {
		ret[y] = make([]term.Cell, maxWidth)
		for x, cell := range rows.Cells {
			ret[y][x] = cell.ToModel()
		}
	}
	return ret
}

func rawCellsToProtoCells(cells [][]term.Cell) []*termrpc.CellRow {
	var size int
	for _, row := range cells {
		size += len(row)
	}

	// these slabs reduce allocations from ~N (=num cells)
	// to 4 which reduces this function's ns/op from 60 to 80%
	rows := make([]*termrpc.CellRow, len(cells))
	protoCellRowSlabPtr := make([]termrpc.CellRow, len(cells))
	cellRowSlab := make([]*termrpc.Cell, size)
	cellRowSlabIdx := 0
	cellSlabPtr := make([]termrpc.Cell, size)
	cellSlabPtrIdx := 0

	for y, row := range cells {
		cells := cellRowSlab[cellRowSlabIdx : cellRowSlabIdx+len(row)]
		cellRowSlabIdx += len(row)
		for x, cell := range row {
			c := &cellSlabPtr[cellSlabPtrIdx]
			cellSlabPtrIdx++
			c.FromModel(cell)
			cells[x] = c
		}
		protoCellRowSlabPtr[y].Cells = cells
		rows[y] = &protoCellRowSlabPtr[y]
	}

	return rows
}
