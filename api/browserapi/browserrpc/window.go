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

package browserrpc

import "github.com/unstablebuild/rune-go-sdk/api/browserapi"

// NewWindow returns a browserapi.Window that represents the window
// with the given ID.
func NewWindow(windowID uint64) browserapi.Window {
	return newWindowClient(windowID)
}

var _ browserapi.Window = (*windowClientImpl)(nil)

type windowClientImpl struct {
	windowID uint64
}

func newWindowClient(windowID uint64) *windowClientImpl {
	ret := new(windowClientImpl)
	ret.windowID = windowID
	return ret
}

func (w *windowClientImpl) WindowID() uint64 {
	return w.windowID
}
