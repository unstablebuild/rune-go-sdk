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
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Responsive components implement a backpressure mechanism (Height) for
// aggregate components to dynamically resize children based on their contents.
// See Height for more details.
type Responsive interface {
	tui.Handler
	component.Responsive
}

// FloatingResponsive is a Floating that also satisfies Responsive.
type FloatingResponsive interface {
	Responsive
	Floating
}
