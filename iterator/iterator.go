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

import (
	"context"
	"io"
)

// Iterator provides a convenient interface for iterating over
// chunks of structured or unstructured data such as
// a file of newline-delimited lines of text or a set of datastore documents.
type Iterator[T any] interface {
	// Next returns the next element and true or nil and false
	// if there's no more elements in this Iterator. Err should be
	// checked for any errors incurred during the lifecycle of this Iterator.
	// Note that if an error is found, it's up to the implementation as to
	// whether to return false and stop iteration or aggregate errors
	// and return at the end.
	//
	// This method blocks until data is available. Implementations should
	// use the given context's Done channel to know when data is no longer
	// required and so the call should return.
	Next(context.Context) (T, bool)
	// Err returns the first error or an aggreation of the errors
	// encountered by the Iterator.
	Err() error

	io.Closer
}

// Empty returns an empty iterator.
func Empty[T any]() Iterator[T] {
	return FromSlice[T](nil)
}
