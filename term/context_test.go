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
	"math"
	"testing"
)

func TestCellPixelSizeFromContext(t *testing.T) {
	sized := ContextWithCellPixelSize(context.Background(), 8, 16)
	tests := []struct {
		name                 string
		ctx                  context.Context
		expectedW, expectedH float64
		expectedOK           bool
	}{
		{name: "nil", ctx: nil},
		{name: "none", ctx: context.Background()},
		{name: "whole pixels", ctx: sized, expectedW: 8, expectedH: 16, expectedOK: true},
		{
			name:      "fractional pixels",
			ctx:       ContextWithCellPixelSize(context.Background(), 8.75, 17.5),
			expectedW: 8.75, expectedH: 17.5, expectedOK: true,
		},
		{
			name:      "the innermost size",
			ctx:       ContextWithCellPixelSize(sized, 10, 20),
			expectedW: 10, expectedH: 20, expectedOK: true,
		},
		{
			name:      "kept by contexts derived for other values",
			ctx:       ContextWithPayload(sized, []byte("p")),
			expectedW: 8, expectedH: 16, expectedOK: true,
		},
		{name: "a zero side", ctx: ContextWithCellPixelSize(sized, 0, 16)},
		{name: "a negative side", ctx: ContextWithCellPixelSize(sized, 8, -16)},
		{name: "not a number", ctx: ContextWithCellPixelSize(sized, math.NaN(), 16)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h, ok := CellPixelSizeFromContext(tt.ctx)
			if w != tt.expectedW || h != tt.expectedH || ok != tt.expectedOK {
				t.Fatalf("CellPixelSizeFromContext() = %v, %v, %v; expected %v, %v, %v",
					w, h, ok, tt.expectedW, tt.expectedH, tt.expectedOK)
			}
		})
	}
}
