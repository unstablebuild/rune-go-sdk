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

// ToSlice consumes the given iterator and returns a slice with
// all of the elements produced.
func ToSlice[T any](ctx context.Context, it Iterator[T]) ([]T, error) {
	ret := make([]T, 0)
	for {
		t, ok := it.Next(ctx)
		if !ok {
			if err := it.Err(); err != nil {
				return nil, err
			}
			return ret, nil
		}
		ret = append(ret, t)
	}
}

// FromSlice returns an iterator backed by the given slice of elements.
func FromSlice[T any](els []T) Iterator[T] {
	return &sliceIter[T]{els: els}
}

// FromFunc returns an Iterator backed by the provided function.
func FromFunc[T any](
	fn func(context.Context) (T, bool, error),
	close func() error,
) Iterator[T] {
	return &fnIter[T]{fn: fn, close: close}
}

// IsEmpty consumes the first element in i and returns true if it is empty
// or false if not and returns a new iterator that should be used instead of i.
//
// The given iterator is consumed in either case, so its Close method
// is wrapped with the returned iterator, or if the given iterator is empty,
// its Close method is called for the caller, so it's safe to override the variable
// holding the passed iterator with the return value of this function.
func IsEmpty[T any](ctx context.Context, i Iterator[T]) (Iterator[T], bool) {
	el, ok := i.Next(ctx)
	if !ok {
		_ = i.Close()
		return FromSlice[T](nil), true
	}

	return FromFunc(func(ctx context.Context) (T, bool, error) {
		if ok {
			ok = false
			return el, true, nil
		}
		iEl, iOk := i.Next(ctx)
		if !iOk {
			return iEl, false, i.Err()
		}
		return iEl, iOk, nil
	}, i.Close), false
}

type fnIter[T any] struct {
	err   error
	fn    func(context.Context) (T, bool, error)
	close func() error
}

func (f *fnIter[T]) Next(ctx context.Context) (T, bool) {
	t, ok, err := f.fn(ctx)
	if err != nil {
		f.err = multierror.Append(f.err, err)
		return t, false
	}
	return t, ok
}

func (f *fnIter[T]) Err() error {
	return f.err
}

func (f *fnIter[T]) Close() error {
	return f.close()
}

type sliceIter[T any] struct {
	els []T
}

func (i *sliceIter[T]) Next(context.Context) (ret T, ok bool) {
	if len(i.els) == 0 {
		return
	}
	ok = true
	ret = i.els[0]
	i.els = i.els[1:]
	return
}

func (i *sliceIter[T]) Err() error {
	return nil
}

func (i *sliceIter[T]) Close() error {
	return nil
}
