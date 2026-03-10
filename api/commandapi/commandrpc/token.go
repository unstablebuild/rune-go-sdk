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

package commandrpc

import (
	"github.com/unstablebuild/rune-go-sdk/api/browserapi/browserrpc"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/term"
)

var _ textapi.Handler = Token{}

// Token wraps a browser.Token to satisfy textapi.Handler.
type Token struct {
	browserrpc.Token
	workspaceapi.URI
}

// Resource satisfies textapi.Handler.
func (t Token) Resource() workspaceapi.URI {
	return t.URI
}

// SetWrap satisfies textapi.Handler.
func (t Token) SetWrap(wrap bool) {
}

// SetCursorAtScroll satisfies textapi.Handler.
func (t Token) SetCursorAtScroll(term.Coordinates) bool {
	return false
}

// ShowCommandBar satisfies textapi.Handler.
func (t Token) ShowCommandBar(show bool) {
}

// SeekUp satisfies textapi.Handler.
func (t Token) SeekUp() bool {
	return false
}

// SeekDown satisfies textapi.Handler.
func (t Token) SeekDown() bool {
	return false
}

// SeekOffset satisfies textapi.Handler.
func (t Token) SeekOffset() int {
	return 0
}

// MaxSeekOffset satisfies textapi.Handler.
func (t Token) MaxSeekOffset() int {
	return 0
}
