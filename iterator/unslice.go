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

// Unslice converts an iterator of slices of T into an iterator of T.
func Unslice[T any](it Iterator[[]T]) Iterator[T] {
	return &unslice[T]{it: it}
}

type unslice[T any] struct {
	it   Iterator[[]T]
	next []T
}

func (u *unslice[T]) Next(ctx context.Context) (ret T, ok bool) {
	for {
		if len(u.next) > 0 {
			head := u.next[0]
			u.next = u.next[1:]
			return head, true
		}

		u.next, ok = u.it.Next(ctx)
		if !ok {
			return
		}
	}
}

func (u *unslice[T]) Err() error {
	return u.it.Err()
}

func (u *unslice[T]) Close() error {
	return u.it.Close()
}
