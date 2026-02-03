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
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregate(t *testing.T) {
	suite := []struct {
		description string
		it          []Iterator[int]
		expectRes   []int
		expectErr   string
	}{
		{
			description: "no iterators returns empty iterator",
			it:          []Iterator[int]{},
			expectRes:   nil,
		},
		{
			description: "one empty iterator returns empty iterator",
			it:          []Iterator[int]{Empty[int]()},
			expectRes:   nil,
		},
		{
			description: "iterator with single value returns that value",
			it:          []Iterator[int]{FromSlice[int]([]int{9})},
			expectRes:   []int{9},
		},
		{
			description: "returns first iterator error",
			it: []Iterator[int]{
				Error[int](errors.New("oops")),
				FromSlice[int]([]int{9}),
			},
			expectErr: "oops",
		},
		{
			description: "returns last iterator error",
			it: []Iterator[int]{
				FromSlice[int]([]int{9}),
				Error[int](errors.New("oops")),
			},
			expectRes: []int{9},
			expectErr: "oops",
		},
		{
			description: "continues calling fn until iterator is exhausted",
			it: []Iterator[int]{
				FromSlice[int]([]int{2, 2}),
				Empty[int](),
				FromSlice[int]([]int{1}),
				FromSlice[int](nil),
			},
			expectRes: []int{2, 2, 1},
		},
	}

	for _, test := range suite {
		t.Run(test.description, func(t *testing.T) {
			actualResIt := Aggregate[int](test.it...)
			actualRes, actualErr := Reduce(context.Background(), actualResIt,
				func(ret []int, i int) ([]int, error) {
					return append(ret, i), nil
				})
			assert.Equal(t, test.expectRes, actualRes)
			if test.expectErr != "" {
				require.Error(t, actualErr)
				assert.True(t, strings.Contains(actualErr.Error(), test.expectErr))
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}

	t.Run("Close is called as iterators are consumed", func(t *testing.T) {
		var (
			it1Closed bool
			it2Closed bool
		)
		var i int
		it1 := FromFunc[string](func(context.Context) (string, bool, error) {
			return "1", i < 1, nil
		}, func() error {
			it1Closed = true
			return nil
		})
		it2 := FromFunc[string](func(context.Context) (string, bool, error) {
			return "2", i < 2, nil
		}, func() error {
			it2Closed = true
			return nil
		})
		actualResIt := Aggregate[string](it1, it2)
		for ; i < 2; i++ {
			n, ok := actualResIt.Next(context.Background())
			require.True(t, ok, i)
			assert.Equal(t, strconv.Itoa(i+1), n, i)
		}

		_, ok := actualResIt.Next(context.Background())
		require.False(t, ok)

		assert.True(t, it1Closed)
		assert.True(t, it2Closed)

		it1Closed = false
		it2Closed = false
		require.NoError(t, actualResIt.Close())

		assert.False(t, it1Closed)
		assert.False(t, it2Closed)
	})

	t.Run("Close calls Close on all all iterators", func(t *testing.T) {
		var (
			it1Closed bool
			it2Closed bool
		)
		it1 := FromFunc[string](func(context.Context) (string, bool, error) {
			return "1", true, nil
		}, func() error {
			it1Closed = true
			return nil
		})
		it2 := FromFunc[string](func(context.Context) (string, bool, error) {
			return "2", true, nil
		}, func() error {
			it2Closed = true
			return nil
		})
		actualResIt := Aggregate[string](it1, it2)

		require.NoError(t, actualResIt.Close())

		assert.True(t, it1Closed)
		assert.True(t, it2Closed)
	})
}
