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

package iterator

import "context"

// Error returns an Iterator of T that never returns a value
// and always returns the given error.
func Error[T any](err error) Iterator[T] {
	return errorIt[T]{err: err}
}

type errorIt[T any] struct {
	err error
}

func (e errorIt[T]) Next(context.Context) (ret T, ok bool) {
	return
}

func (e errorIt[T]) Err() error {
	return e.err
}

func (e errorIt[T]) Close() error {
	return nil
}
