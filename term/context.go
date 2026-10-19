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

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type ctxKey int

const (
	// pKey is the key for sync.Payload values in Contexts. It is
	// unexported; clients use workspace.ContextWithPayload and
	// PayloadFromContext instead of using this key directly.
	pKey ctxKey = iota
	cellPixelSizeKey
)

type cellPixelSize struct {
	width, height float64
}

// ContextWithPayload returns a new Context that holds locker.
func ContextWithPayload(ctx context.Context, payload []byte) context.Context {
	return context.WithValue(ctx, pKey, payload)
}

// PayloadFromContext returns the payload value stored in ctx, if any.
func PayloadFromContext(ctx context.Context) ([]byte, bool) {
	locker, ok := ctx.Value(pKey).([]byte)
	return locker, ok
}

// ContextWithCellPixelSize returns ctx carrying the cell size, in device
// pixels, of the frame drawn with it.
func ContextWithCellPixelSize(ctx context.Context, width, height float64) context.Context {
	return context.WithValue(ctx, cellPixelSizeKey, cellPixelSize{width: width, height: height})
}

// CellPixelSizeFromContext returns the cell size, in device pixels, of the
// frame being drawn, which need not be whole. It is valid during Draw and
// may change on the next one; a handler that needs it outside Draw keeps
// the last one it saw. ok is false if ctx carries no size with positive
// sides. A missing size does not tell whether the writer shows images:
// only Writer.DrawImage does.
func CellPixelSizeFromContext(ctx context.Context) (width, height float64, ok bool) {
	if ctx == nil {
		return 0, 0, false
	}
	size, ok := ctx.Value(cellPixelSizeKey).(cellPixelSize)
	if !ok || !(size.width > 0 && size.height > 0) {
		return 0, 0, false
	}
	return size.width, size.height, true
}
