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
	"strings"
	"sync/atomic"
)

// PointerShape is the shape of the mouse pointer.
type PointerShape int32

// Pointer shapes. Renderers without a dedicated cursor for a shape show
// the closest one they have.
const (
	PointerDefault PointerShape = iota
	PointerText
	PointerLink
	PointerMove
	PointerResizeEW
	PointerResizeNS
	PointerResizeNESW
	PointerResizeNWSE
	PointerColResize
	PointerRowResize
	PointerNotAllowed
)

var pointerShapeNames = [...]string{
	PointerDefault:    "default",
	PointerText:       "text",
	PointerLink:       "pointer",
	PointerMove:       "move",
	PointerResizeEW:   "ew-resize",
	PointerResizeNS:   "ns-resize",
	PointerResizeNESW: "nesw-resize",
	PointerResizeNWSE: "nwse-resize",
	PointerColResize:  "col-resize",
	PointerRowResize:  "row-resize",
	PointerNotAllowed: "not-allowed",
}

var pointerShapeAliases = map[string]PointerShape{
	"auto":              PointerDefault,
	"left_ptr":          PointerDefault,
	"xterm":             PointerText,
	"ibeam":             PointerText,
	"hand":              PointerLink,
	"hand1":             PointerLink,
	"hand2":             PointerLink,
	"fleur":             PointerMove,
	"e-resize":          PointerResizeEW,
	"w-resize":          PointerResizeEW,
	"sb_h_double_arrow": PointerResizeEW,
	"n-resize":          PointerResizeNS,
	"s-resize":          PointerResizeNS,
	"sb_v_double_arrow": PointerResizeNS,
	"ne-resize":         PointerResizeNESW,
	"sw-resize":         PointerResizeNESW,
	"nw-resize":         PointerResizeNWSE,
	"se-resize":         PointerResizeNWSE,
	"no-drop":           PointerNotAllowed,
}

// String returns the CSS cursor name of s, which is also the name OSC 22
// takes. Unknown shapes are "default".
func (s PointerShape) String() string {
	if s < 0 || int(s) >= len(pointerShapeNames) {
		return pointerShapeNames[PointerDefault]
	}
	return pointerShapeNames[s]
}

// ParsePointerShape returns the shape named by a CSS cursor name, case
// insensitively. It also accepts the CSS single-edge resize names and the
// X11 cursor font names terminal programs commonly send in OSC 22. ok is
// false for names without a matching shape.
func ParsePointerShape(name string) (shape PointerShape, ok bool) {
	name = strings.ToLower(name)
	for s, n := range pointerShapeNames {
		if n == name {
			return PointerShape(s), true
		}
	}
	shape, ok = pointerShapeAliases[name]
	return shape, ok
}

// Pointer is the pointer shape slot of one mouse Event. A runtime that can
// change the pointer shape attaches a fresh Pointer to the Context of every
// mouse event it dispatches and shows the shape of the most recent one,
// or PointerDefault if unset.
//
// The zero value is an unset Pointer. Pointer is safe for concurrent use.
type Pointer struct {
	// zero value means unset
	shape atomic.Int32
}

// SetShape sets the pointer shape.
func (p *Pointer) SetShape(s PointerShape) {
	p.shape.Store(int32(s) + 1)
}

// Shape returns the shape last set and whether any was. An unset Pointer
// returns PointerDefault.
func (p *Pointer) Shape() (shape PointerShape, set bool) {
	v := p.shape.Load()
	if v == 0 {
		return PointerDefault, false
	}
	return PointerShape(v - 1), true
}

// ContextWithPointer returns ctx carrying p.
func ContextWithPointer(ctx context.Context, p *Pointer) context.Context {
	return context.WithValue(ctx, pointerKey, p)
}

// PointerFromContext returns the Pointer carried by ctx, if any.
func PointerFromContext(ctx context.Context) (*Pointer, bool) {
	if ctx == nil {
		return nil, false
	}
	p, ok := ctx.Value(pointerKey).(*Pointer)
	return p, ok && p != nil
}

// SetPointerShape sets the pointer shape of the mouse event ev. It is a
// no-op when ev carries no Pointer, as is the case for events that are not
// mouse events or are dispatched by a runtime that cannot change the
// pointer shape.
func SetPointerShape(ev Event, s PointerShape) {
	if p, ok := PointerFromContext(ev.Context); ok {
		p.SetShape(s)
	}
}
