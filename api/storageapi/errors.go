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

package storageapi

import "errors"

var (
	// ErrNotFound is returned when the given document was not found in the store.
	ErrNotFound = errors.New("document not found")

	// ErrAlreadyExists is returned when the given document already exists.
	ErrAlreadyExists = errors.New("document already exists")

	// ErrPreconditionFailed is returned when a Precondition to Update
	// is not met by the underlying document.
	ErrPreconditionFailed = errors.New("document pre-condition was not met")

	// ErrPermissionDenied is returned when an operation is denied due to
	// permissions.
	ErrPermissionDenied = errors.New("access to document denied")
)
