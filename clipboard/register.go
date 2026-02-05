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

package clipboard

var (
	// DefaultRegisterID represents the program's default clipboard register to be used with
	// Clipboard's Copy/Paste operations.
	DefaultRegisterID string = " "

	// UnnamedRegisterID is an alias of DefaultRegisterID.
	UnnamedRegisterID = DefaultRegisterID
)

// Register is the interface that wraps the basic Copy, Paste short-term data storage
// methods for editors to use multiple storage registers. Implementations must be
// gourountine-safe.
type Register interface {
	Paste(registerID string) (Data, error)
	Copy(registerID string, data Data) error
}

// Data is the structure used for short-term data storage and/or data transfer
// via Clipboard's Copy/Paste operations.
type Data struct {
	Text     string
	Metadata interface{}
}
