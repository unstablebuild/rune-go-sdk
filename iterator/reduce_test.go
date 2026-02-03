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
