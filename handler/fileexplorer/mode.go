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

package fileexplorer

// Mode represents the current interaction mode of the file explorer handler.
type Mode int

const (
	// ModeView is the default mode for navigation and expand/collapse.
	ModeView Mode = iota
	// ModeEdit is the mode for inline filename editing.
	ModeEdit
	// ModeConfirm is the mode for confirming pending operations.
	ModeConfirm
)

// String returns the string representation of the mode.
func (m Mode) String() string {
	switch m {
	case ModeView:
		return "view"
	case ModeEdit:
		return "edit"
	case ModeConfirm:
		return "confirm"
	default:
		return "unknown"
	}
}
