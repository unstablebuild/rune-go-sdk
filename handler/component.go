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

package handler

import (
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

type withComponent struct {
	tui.Handler
	c tui.Component
}

// WithComponent wraps h with comp for the methods of h that satisfy tui.Component
// and delegates the remaining tui.Handler methods to h.
func WithComponent(h tui.Handler, comp tui.Component) tui.Handler {
	return withComponent{Handler: h, c: comp}
}

func (c withComponent) Draw(w term.Writer) {
	c.c.Draw(w)
}

func (c withComponent) Resize(width, height int) {
	c.c.Resize(width, height)
}
