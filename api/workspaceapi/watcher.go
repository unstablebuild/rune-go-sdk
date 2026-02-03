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
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/unstablebuild/rune-go-sdk/debug"
)

// ChanProcessWatcher returns a ProcessWatcher that simply returns
// ch when Watch is called.
func ChanProcessWatcher(ch chan error) ProcessWatcher {
	return waitCh(ch)
}

// MultiProcessWatcher returns a ProcessWatcher that ensures that
// all the given watchers get notified when the process is done.
func MultiProcessWatcher(watchers ...ProcessWatcher) ProcessWatcher {
	if len(watchers) == 0 {
		panic("watchers cannot be empty")
	}
	return newMultiWatcher(watchers)
}

type waitCh chan error

func (w waitCh) WatchProcess() chan error {
	return w
}

type multiWatcher struct {
	ch chan error
}

func newMultiWatcher(watchers []ProcessWatcher) ProcessWatcher {
	ret := &multiWatcher{ch: make(chan error)}
	go debug.CapturePanicReport(func() {
		const watcherWaitTimeout = 30 * time.Second
		err := <-ret.ch
		ctx, cancel := context.WithTimeout(context.Background(),
			watcherWaitTimeout)
		defer cancel()

		var wg sync.WaitGroup
		for _, watcher := range watchers {
			ch := watcher.WatchProcess()
			if ch == nil {
				continue
			}
			wg.Add(1)
			go debug.CapturePanicReport(func() {
				defer wg.Done()
				select {
				case ch <- err:
				case <-ctx.Done():
					slog.Warn("could not deliver error to watcher chan: " +
						"watcher not ready for too long")
				}
			})
		}
		wg.Wait()
	})
	return ret
}

func (w *multiWatcher) WatchProcess() chan error {
	return w.ch
}
