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

package llmapi

import (
	"errors"
	"fmt"
)

// ErrModelNotFound is returned by Service.GetModel when the requested
// model is not known to the service. Callers can detect it with errors.Is.
var ErrModelNotFound = errors.New("llm: model not found")

// ErrContextWindowExceeded is returned when the number of tokens in a request
// exceeds the context window of a model. Users are encouraged to retry
// with a reduced number of messages.
type ErrContextWindowExceeded struct {
	Count, Max int
}

// Error satisfies the error interface.
func (e *ErrContextWindowExceeded) Error() string {
	return fmt.Sprintf("model context window exceeded (%d, max is %d)", e.Count, e.Max)
}

// Unwrap satisfies the error interface.
func (e *ErrContextWindowExceeded) Unwrap() error {
	return nil
}

// Is satisfies the error interface.
func (e *ErrContextWindowExceeded) Is(target error) bool {
	_, ok := target.(*ErrContextWindowExceeded)
	return ok
}
