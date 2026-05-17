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

	"github.com/ernestrc/go-multierror"
)

// Aggregate combines multiple iterators of T into one single iterator
// of T. If its is nil or empty, this method safely returns an empty
// iterator. The returned iterator is a wrapper: its Close delegates
// to each input iterator's Close, and inputs already drained during
// iteration are also closed at that point. The caller is responsible
// for closing the returned iterator. See the Iterator type for the
// wrapper/terminal contract.
func Aggregate[T any](its ...Iterator[T]) Iterator[T] {
	return &aggregate[T]{its: its}
}

type aggregate[T any] struct {
	its []Iterator[T]
	err error
}

func (a *aggregate[T]) Next(ctx context.Context) (ret T, ok bool) {
	for {
		if len(a.its) == 0 {
			return
		}
		ret, ok = a.its[0].Next(ctx)
		if ok {
			return
		}
		if ierr := a.its[0].Err(); ierr != nil {
			a.err = multierror.Append(a.err, ierr)
		}
		if cerr := a.its[0].Close(); cerr != nil {
			a.err = multierror.Append(a.err, cerr)
		}
		if a.err != nil {
			return
		}
		a.its = a.its[1:]
	}
}

func (a *aggregate[T]) Err() error {
	return a.err
}

func (a *aggregate[T]) Close() (ret error) {
	for _, it := range a.its {
		if err := it.Close(); err != nil {
			ret = multierror.Append(ret, err)
		}
	}
	return
}
