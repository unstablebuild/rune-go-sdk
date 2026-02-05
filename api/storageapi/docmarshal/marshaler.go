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

package docmarshal

// Marshaler is a text or binary marshaler that can be used
// by document.Service implementations to abstract document encoding.
type Marshaler interface {
	Marshal(in any) ([]byte, error)
	Unmarshal(data []byte, to any) error
	// DefaultLowerCase should return true if by default
	// struct fields are encoded in lower case.
	DefaultLowerCase() bool
}
