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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReduce(t *testing.T) {
	suite := []struct {
		description string
		it          Iterator[int]
		fn          func(int, int) (int, error)
		expectRes   int
		expectErr   error
	}{
		{
			description: "empty iterator returns zero value",
			it:          Empty[int](),
			fn:          func(int, int) (int, error) { return 0, nil },
			expectRes:   0,
		},
		{
			description: "iterator with single value returns that value",
			it:          FromSlice[int]([]int{9}),
			fn:          func(ret int, i int) (int, error) { return ret + i, nil },
			expectRes:   9,
		},
		{
			description: "returns fn error",
			it:          FromSlice[int]([]int{9}),
			fn:          func(ret int, i int) (int, error) { return 0, errors.New("oops") },
			expectErr:   errors.New("oops"),
		},
		{
			description: "continues calling fn until iterator is exhausted",
			it:          FromSlice[int]([]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1}),
			fn:          func(ret int, i int) (int, error) { return ret + i, nil },
			expectRes:   10,
		},
	}

	for _, test := range suite {
		t.Run(test.description, func(t *testing.T) {
			actualRes, actualErr := Reduce(context.Background(), test.it, test.fn)
			assert.Equal(t, test.expectRes, actualRes)
			assert.Equal(t, test.expectErr, actualErr)
		})
	}
}

// TestReduceClosesIterator asserts that Reduce closes its input
// iterator after consuming it. Reduce is a terminal operation: callers
// have no handle to close the iterator themselves once Reduce returns,
// so leaving it open leaks any resources (goroutines, file descriptors,
// network connections) that the iterator owns.
func TestReduceClosesIterator(t *testing.T) {
	for _, tc := range []struct {
		desc string
		in   []int
		fn   func(int, int) (int, error)
	}{
		{"empty", nil, func(int, int) (int, error) { return 0, nil }},
		{"happy path", []int{1, 2, 3}, func(a, b int) (int, error) { return a + b, nil }},
		{"reducer error", []int{1}, func(int, int) (int, error) { return 0, errors.New("oops") }},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			it := newClosableIter(tc.in)
			_, _ = Reduce(context.Background(), it, tc.fn)
			assert.Equal(t, 1, it.closeCount, "Reduce must close its input iterator exactly once")
		})
	}
}

// TestReduceReportsCloseError asserts that Reduce surfaces Close
// errors, joined with any error from iteration, so that callers do not
// silently lose failures from the underlying resource.
func TestReduceReportsCloseError(t *testing.T) {
	closeErr := errors.New("close failed")
	it := newClosableIter([]int{1, 2})
	it.closeErr = closeErr

	_, err := Reduce(context.Background(), it,
		func(acc, v int) (int, error) { return acc + v, nil })
	require.Error(t, err)
	assert.ErrorIs(t, err, closeErr)
}
