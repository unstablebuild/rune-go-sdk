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

package textapi

import "errors"

var (
	// ErrInvalidSave is returned when trying to save a tab that it's not a file
	// in the file system.
	ErrInvalidSave = errors.New("cannot flush this content")

	// ErrInvalidReload is returned when trying to reload a tab that it's not a file
	// in the file system.
	ErrInvalidReload = errors.New("cannot reload this content")

	// ErrInvalidSplit is returned when attempting to split over a floating window.
	ErrInvalidSplit = errors.New("cannot split this window")

	// ErrInvalidOverwrite is returned when trying to overwrite a tab that it's not a file
	// in the file system.
	ErrInvalidOverwrite = errors.New("cannot overwrite this content")
)
