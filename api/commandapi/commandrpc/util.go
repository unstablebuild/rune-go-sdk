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

import "github.com/unstablebuild/rune-go-sdk/api/workspaceapi"

// NewURIFromProto maps rpc.URI into a workspaceapi.URI.
func NewURIFromProto(u *URI) (workspaceapi.URI, error) {
	return workspaceapi.ParseURI(u.GetUri())
}

// NewURI maps a workspaceapi.URI into a rpc.URI.
func NewURI(u workspaceapi.URI) *URI {
	return &URI{Uri: u.String()}
}
