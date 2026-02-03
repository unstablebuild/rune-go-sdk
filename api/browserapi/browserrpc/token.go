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

import (
	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
	"github.com/unstablebuild/rune-go-sdk/term"
)

const errMsg = "this Handler is a token handler that cannot be used directly"

var _ browserapi.Handler = Token{}

// Token is a handler used to indicate which of the remote handlers
// to set as content to a remote server. It satisfies tui.Handler so that clients
// can take the result of an browser.Open type of requests and pass it to Split* or SetContent
// type of responses.
type Token struct {
	URI string
}

// Handle panics if called. This tui.Handler implementation is symbolic.
func (h Token) Handle(term.Event) (exit, handled bool) {
	panic(errMsg)
}

// Cursor panics if called. This tui.Handler implementation is symbolic.
func (h Token) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	panic(errMsg)
}

// Selection panics if called. This tui.Handler implementation is symbolic.
func (h Token) Selection() (string, bool) {
	panic(errMsg)
}

// Resize panics if called. This tui.Handler implementation is symbolic.
func (h Token) Resize(width, height int) {
	panic(errMsg)
}

// Draw panics if called. This tui.Handler implementation is symbolic.
func (h Token) Draw(w term.Writer) {
	panic(errMsg)
}

// Dimensions panics if called. This tui.Handler implementation is symbolic.
func (h Token) Dimensions() (width, height int) {
	panic(errMsg)
}

// Close does nothing if called.
func (h Token) Close() error {
	return nil
}
