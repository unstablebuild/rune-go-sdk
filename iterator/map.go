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

// Map maps an iterator of type T and returns another iterator that will
// apply fn to each of the elements produced.
func Map[T any, V any](it Iterator[T], fn func(T) V) Iterator[V] {
	return FromFunc(func(ctx context.Context) (ret V, ok bool, err error) {
		var t T
		t, ok = it.Next(ctx)
		if !ok {
			err = it.Err()
			return
		}
		ret = fn(t)
		return
	}, it.Close)
}
