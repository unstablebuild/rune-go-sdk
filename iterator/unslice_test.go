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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnslice(t *testing.T) {
	suite := []struct {
		description string
		it          Iterator[[]int]
		expectRes   []int
	}{
		{
			description: "empty iterator returns empty iterator",
			it:          Empty[[]int](),
			expectRes:   nil,
		},
		{
			description: "iterator with single value returns that value",
			it:          FromSlice[[]int]([][]int{{9}}),
			expectRes:   []int{9},
		},
		{
			description: "iterator with multiple slices with different lengths",
			it:          FromSlice[[]int]([][]int{{1, 1}, {1}, {1}, {1, 1, 1, 1}, {1, 1}}),
			expectRes:   []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		},
		{
			description: "continues calling fn until iterator is exhausted",
			it:          FromSlice[[]int]([][]int{{1, 1}, nil, {1, 1}, {1, 1, 1, 1}, nil, nil, {1, 1}}),
			expectRes:   []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		},
	}

	for _, test := range suite {
		t.Run(test.description, func(t *testing.T) {
			actualResIt := Unslice(test.it)
			actualRes, err := Reduce(context.Background(), actualResIt,
				func(ret []int, i int) ([]int, error) {
					return append(ret, i), nil
				})
			require.NoError(t, err)
			assert.Equal(t, test.expectRes, actualRes)
		})
	}
}
