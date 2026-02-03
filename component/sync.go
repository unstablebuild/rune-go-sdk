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
	"sync"

	"github.com/unstablebuild/rune-go-sdk/tui"
	"github.com/unstablebuild/rune-go-sdk/term"
)

type csync struct {
	mu sync.Locker
	c  tui.Component
}

// Sync wraps a tui.Component to provide access synchronization with mu.
func Sync(mu sync.Locker, c tui.Component) tui.Component {
	return csync{mu: mu, c: c}
}

func (s csync) Resize(width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.c.Resize(width, height)
}

func (s csync) Draw(w term.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.c.Draw(w)
}
