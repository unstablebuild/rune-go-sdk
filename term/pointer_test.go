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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPointerShapeStringRoundTrip(t *testing.T) {
	for s := PointerDefault; s <= PointerNotAllowed; s++ {
		got, ok := ParsePointerShape(s.String())
		require.True(t, ok, s.String())
		require.Equal(t, s, got)
	}
	require.Equal(t, "default", PointerShape(-1).String())
	require.Equal(t, "default", PointerShape(1000).String())
}

func TestParsePointerShape(t *testing.T) {
	tests := []struct {
		name     string
		expected PointerShape
		ok       bool
	}{
		{name: "col-resize", expected: PointerColResize, ok: true},
		{name: "Pointer", expected: PointerLink, ok: true},
		{name: "xterm", expected: PointerText, ok: true},
		{name: "hand2", expected: PointerLink, ok: true},
		{name: "w-resize", expected: PointerResizeEW, ok: true},
		{name: "se-resize", expected: PointerResizeNWSE, ok: true},
		{name: "wait"},
		{name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParsePointerShape(tt.name)
			require.Equal(t, tt.ok, ok)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestPointerSetShape(t *testing.T) {
	var p Pointer

	shape, set := p.Shape()
	require.False(t, set)
	require.Equal(t, PointerDefault, shape)

	p.SetShape(PointerDefault)
	shape, set = p.Shape()
	require.True(t, set)
	require.Equal(t, PointerDefault, shape)

	p.SetShape(PointerLink)
	shape, _ = p.Shape()
	require.Equal(t, PointerLink, shape)
}

func TestSetPointerShape(t *testing.T) {
	p := new(Pointer)
	ev := Event{Type: EventMouse, Context: ContextWithPointer(context.Background(), p)}
	SetPointerShape(ev, PointerMove)
	shape, _ := p.Shape()
	require.Equal(t, PointerMove, shape)

	SetPointerShape(Event{Type: EventMouse}, PointerMove)
	SetPointerShape(Event{Type: EventMouse, Context: context.Background()}, PointerMove)

	_, ok := PointerFromContext(nil) //nolint:staticcheck // nil is part of the contract
	require.False(t, ok)
}
