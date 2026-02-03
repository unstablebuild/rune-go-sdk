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

package workspaceapi

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMultiProcessWatcher(t *testing.T) {
	t.Run("ChanProcessWatcher returns same channel", func(t *testing.T) {
		ch := make(chan error)
		pw := ChanProcessWatcher(ch)

		got := pw.WatchProcess()
		assert.Equal(t, ch, got, "WatchProcess should return the same channel")
	})

	t.Run("Panics on empty watchers", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = MultiProcessWatcher()
		}, "MultiProcessWatcher should panic if no watchers are provided")
	})

	t.Run("Returns non-nil channel", func(t *testing.T) {
		wch := make(chan error, 1)
		pw := ChanProcessWatcher(wch)

		ret := MultiProcessWatcher(pw)
		assert.NotNil(t, ret, "MultiProcessWatcher should not return nil")
		assert.NotNil(t, ret.WatchProcess(), "returned watcher channel should not be nil")
	})

	t.Run("Propagates error to all watchers", func(t *testing.T) {
		w1 := make(chan error, 1)
		w2 := make(chan error, 1)

		pw1 := ChanProcessWatcher(w1)
		pw2 := ChanProcessWatcher(w2)

		ret := MultiProcessWatcher(pw1, pw2)
		retCh := ret.WatchProcess()

		want := errors.New("boom")
		go func() {
			retCh <- want
		}()

		within := 200 * time.Millisecond
		got1 := recvWithin(t, w1, within)
		got2 := recvWithin(t, w2, within)

		assert.ErrorIs(t, got1, want, "watcher 1 should receive the propagated error")
		assert.ErrorIs(t, got2, want, "watcher 2 should receive the propagated error")
	})

	t.Run("Nil watcher channel is skipped", func(t *testing.T) {
		wValid1 := make(chan error, 1)
		wValid2 := make(chan error, 1)

		pwValid1 := ChanProcessWatcher(wValid1)
		pwValid2 := ChanProcessWatcher(wValid2)
		pwNil := nilWatcher{}

		ret := MultiProcessWatcher(pwNil, pwValid1, pwValid2)
		retCh := ret.WatchProcess()

		want := errors.New("should be delivered")
		go func() {
			retCh <- want
		}()

		within := 150 * time.Millisecond
		got1 := recvWithin(t, wValid1, within)
		got2 := recvWithin(t, wValid2, within)
		assert.ErrorIs(t, got1, want)
		assert.ErrorIs(t, got2, want)
	})
}

func recvWithin(t *testing.T, ch <-chan error, d time.Duration) error {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(d):
		assert.Fail(t, "timed out after %s waiting for receive", d.String())
		return nil
	}
}

// A ProcessWatcher that returns a nil channel (to exercise the nil-chan branch).
type nilWatcher struct{}

func (nilWatcher) WatchProcess() chan error { return nil }
