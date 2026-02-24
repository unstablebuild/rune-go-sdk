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
	"sync/atomic"

	"github.com/unstablebuild/rune-go-sdk/api/storageapi"
	"github.com/unstablebuild/rune-go-sdk/debug"
)

// FromDocumentIterator maps a storageapi.Iterator to an Iterator of type T.
func FromDocumentIterator[T any](it storageapi.Iterator) Iterator[T] {
	var closed atomic.Bool
	type msg struct {
		data T
		err  error
	}

	ch := make(chan msg)
	quitCh := make(chan struct{})
	go debug.CapturePanicReport(func() {
		defer close(ch) // signal ok = false below
		var err error
		var data T
		for {
			if !it.HasNext() {
				return
			}
			err = it.NextTo(&data)
			select {
			case ch <- msg{data: data, err: err}:
			case <-quitCh:
				return
			}
		}
	})

	return FromFunc(func(ctx context.Context) (ret T, ok bool, err error) {
		var m msg
		select {
		case m, ok = <-ch:
			ret = m.data
			err = m.err
			return
		case <-ctx.Done():
			err = ctx.Err()
			return
		}
	}, func() error {
		if !closed.CompareAndSwap(false, true) {
			return nil
		}

		close(quitCh)
		err := it.Close()
		<-ch // wait for goroutine to be done
		return err
	})
}
