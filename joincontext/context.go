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

package joincontext

import (
	"context"
	"sync"
	"time"
)

// New returns a new context that inherits cancellation and
// deadlines from both ctx1 and ctx2. The returned context is
// done when either parent is done or when the returned cancel
// function is called.
//
// Values are looked up in ctx1 first, falling back to ctx2.
func New(ctx1, ctx2 context.Context) (context.Context, context.CancelFunc) {
	jc := &joinedContext{
		ctx1: ctx1,
		ctx2: ctx2,
		done: make(chan struct{}),
	}

	stop1 := context.AfterFunc(ctx1, func() {
		jc.setErr(ctx1.Err())
	})
	stop2 := context.AfterFunc(ctx2, func() {
		jc.setErr(ctx2.Err())
	})

	cancel := func() {
		stop1()
		stop2()
		jc.setErr(context.Canceled)
	}

	return jc, cancel
}

// joinedContext is a context that is done when either of its two
// parent contexts is done, or when its own cancel function is called.
type joinedContext struct {
	ctx1, ctx2 context.Context
	done       chan struct{}
	mu         sync.Mutex
	err        error
}

func (jc *joinedContext) setErr(err error) {
	jc.mu.Lock()
	defer jc.mu.Unlock()

	if jc.err != nil {
		return
	}
	jc.err = err
	close(jc.done)
}

func (jc *joinedContext) Deadline() (time.Time, bool) {
	d1, ok1 := jc.ctx1.Deadline()
	d2, ok2 := jc.ctx2.Deadline()

	switch {
	case ok1 && ok2:
		if d1.Before(d2) {
			return d1, true
		}
		return d2, true
	case ok1:
		return d1, true
	case ok2:
		return d2, true
	default:
		return time.Time{}, false
	}
}

func (jc *joinedContext) Done() <-chan struct{} {
	return jc.done
}

func (jc *joinedContext) Err() error {
	jc.mu.Lock()
	defer jc.mu.Unlock()

	return jc.err
}

func (jc *joinedContext) Value(key any) any {
	if v := jc.ctx1.Value(key); v != nil {
		return v
	}
	return jc.ctx2.Value(key)
}
