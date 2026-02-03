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

// WithLogging wraps comp to log calls to Resize and Draw using the provided logger.
func WithLogging(comp tui.Component, logger func(string, ...any)) tui.Component {
	if logger == nil {
		panic("logger cannot be nil")
	}
	return withLogging{comp: comp, logger: logger}
}

type withLogging struct {
	comp   tui.Component
	logger func(string, ...any)
}

func (l withLogging) Resize(width, height int) {
	l.logger("Resize(%p): width=%d, height=%d", l.comp, width, height)
	l.comp.Resize(width, height)
}

func (l withLogging) Draw(w term.Writer) {
	l.logger("Draw(%p)", l.comp)
	l.comp.Draw(w)
}
