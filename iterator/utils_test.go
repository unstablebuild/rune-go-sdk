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

func TestIsEmpty(t *testing.T) {
	tsuite := []struct {
		desc          string
		inSlice       []testStruct
		expectedOutOk bool
	}{
		{"empty returns true and empty iterator", nil, true},
		{"one item returns false and and same item iterator", []testStruct{{"1", 1}}, false},
		{"multiple items returns false and and same iterator", []testStruct{{"1", 1}, {"2", 2}}, false},
	}

	ctx := context.Background()

	for _, tcase := range tsuite {
		t.Run(tcase.desc, func(t *testing.T) {
			actualOutIt, actualOutOk := IsEmpty(ctx, FromSlice(tcase.inSlice))
			assert.Equal(t, tcase.expectedOutOk, actualOutOk)
			actualOutSlice, err := ToSlice(ctx, actualOutIt)
			require.NoError(t, err)
			assert.Equal(t, append([]testStruct{}, tcase.inSlice...), actualOutSlice)
		})
	}

	t.Run("iterator returned Empty is empty", func(t *testing.T) {
		next, empty := IsEmpty(ctx, Empty[string]())
		require.True(t, empty)

		_, empty = IsEmpty(ctx, next)
		require.True(t, empty)
	})

	t.Run("Close is called if empty", func(t *testing.T) {
		var called bool
		_, empty := IsEmpty(ctx, FromFunc[string](func(context.Context) (string, bool, error) {
			return "", false, nil
		}, func() error {
			called = true
			return nil
		}))
		require.True(t, empty)
		assert.True(t, called)
	})

	t.Run("Close is called if non empty, when the returned iterator's Close is called", func(t *testing.T) {
		var called bool
		var i int
		next, empty := IsEmpty(ctx, FromFunc[string](func(context.Context) (string, bool, error) {
			i++
			return "a", i == 1, nil
		}, func() error {
			called = true
			return nil
		}))
		require.False(t, empty)
		assert.False(t, called)

		val, ok := next.Next(ctx)
		require.True(t, ok)
		assert.Equal(t, "a", val)
		assert.False(t, called)

		_, ok = next.Next(ctx)
		require.False(t, ok)
		require.NoError(t, next.Close())
		assert.True(t, called)
	})
}
