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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/docmarshal/doctoml"
	"go.uber.org/goleak"
)

func TestDocumentIterator(t *testing.T) {
	t.Run("double Close is a no-op", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		it := FromDocumentIterator[string](&errorDocumentIterator{})
		_, ok := it.Next(context.Background())
		assert.False(t, ok)
		require.NoError(t, it.Close())
		require.NoError(t, it.Close())
	})

	t.Run("unblocks Next if context is canceled", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()
		dit := blockingDocumentIterator{quitCh: make(chan struct{})}
		it := FromDocumentIterator[string](dit)
		_, ok := it.Next(ctx)
		assert.False(t, ok)
		assert.Error(t, it.Err())
		require.NoError(t, it.Close())
	})

	t.Run("exhausts underlying storageapi iterator", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		type bob struct {
			X string
		}
		a, b, c := bob{X: "a"}, bob{X: "b"}, bob{X: "c"}
		dit := storageapi.NewListIterator(doctoml.Marshaler(), a, b, c)
		it := FromDocumentIterator[bob](dit)
		actual, err := Reduce(context.Background(),
			it, func(ret []bob, t bob) ([]bob, error) {
				return append(ret, t), nil
			})
		require.NoError(t, err)
		assert.EqualValues(t, []bob{a, b, c}, actual)
		require.NoError(t, it.Close())
	})

	t.Run("bubbles up underlying storageapi iterator errors", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		it := FromDocumentIterator[string](&errorDocumentIterator{})
		_, ok := it.Next(context.Background())
		assert.False(t, ok)
		assert.EqualError(t, it.Err(), "1 error occurred: kaboom")
		require.NoError(t, it.Close())
	})

	// Regression: the goroutine in FromDocumentIterator reused a single
	// `var data T` across iterations. After sending data[i] on the channel the
	// goroutine immediately called NextTo(&data) for data[i+1], which caused the
	// TOML decoder to overwrite the slice backing arrays of the already-returned
	// value while the consumer was still reading them.
	t.Run("no data race on slice fields across successive Next calls", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		type record struct {
			Tags []string
		}
		a := record{Tags: []string{"go", "sdk"}}
		b := record{Tags: []string{"fix", "race"}}
		dit := storageapi.NewListIterator(doctoml.Marshaler(), a, b)
		it := FromDocumentIterator[record](dit)

		r0, ok := it.Next(context.Background())
		require.True(t, ok)

		// Read r0.Tags in a goroutine while the iterator fetches r1.
		// With the old shared-variable implementation the background goroutine
		// would decode r1 into the same Tags backing array concurrently.
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, tag := range r0.Tags {
				_ = tag
			}
		}()

		_, _ = it.Next(context.Background())
		wg.Wait()
		require.NoError(t, it.Close())
	})

	t.Run("does not leak goroutine if data was read, "+
		"but Close was called before Next", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		type bob struct {
			X string
		}
		a := bob{X: "a"}
		dit := storageapi.NewListIterator(doctoml.Marshaler(), a)
		it := FromDocumentIterator[bob](dit)
		require.NoError(t, it.Close())
	})
}

type blockingDocumentIterator struct {
	quitCh chan struct{}
}

func (b blockingDocumentIterator) HasNext() bool {
	return true
}

func (b blockingDocumentIterator) NextTo(doc any) error {
	<-b.quitCh
	return nil
}

func (b blockingDocumentIterator) Close() error {
	close(b.quitCh)
	return nil
}

type errorDocumentIterator struct {
	closed bool
}

func (b *errorDocumentIterator) HasNext() bool {
	return true
}

func (b *errorDocumentIterator) NextTo(doc any) error {
	return errors.New("kaboom")
}

func (b *errorDocumentIterator) Close() error {
	b.closed = true
	return nil
}
