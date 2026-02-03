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

// Reduce combines all the elements in an iterator using a
// binary operation to produce a single value.
func Reduce[T any, V any](
	ctx context.Context,
	it Iterator[T],
	reducer func(V, T) (V, error),
) (ret V, err error) {
	for {
		t, ok := it.Next(ctx)
		if !ok {
			err = it.Err()
			return
		}
		ret, err = reducer(ret, t)
		if err != nil {
			return
		}
	}
}
