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

package retry

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetryErrorValue(t *testing.T) {
	t.Run("returns error as is if retry was set to false since the start", func(t *testing.T) {
		origErr := errors.New("bla")
		err := Retry(context.Background(), LimitStrategy(2), func(ctx context.Context) (bool, error) {
			return false, origErr
		})
		assert.Equal(t, origErr, err)
	})
}

func TestRetry(t *testing.T) {
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	<-canceledCtx.Done()

	tsuite := []struct {
		msg       string
		ctx       context.Context
		strategy  Strategy
		wantCalls int
		wantErr   string
		fn        func(ctx context.Context) (bool, error)
	}{
		{"succeeds on first attempt calls fn only once",
			context.Background(), LimitStrategy(2), 1, "", succeedAfter(0)},
		{"succeeds on last attempt calls fn twice",
			context.Background(), LimitStrategy(2), 2, "", succeedAfter(1)},
		{"fails on last strategy attempt calls fn 3",
			context.Background(), LimitStrategy(2), 2,
			"whoopsie\nwhoopsie", succeedAfter(2)},
		{"fails on last function attempt calls fn 3",
			context.Background(), LimitStrategy(3), 3,
			"whoopsie\nwhoopsie\nlast one", failAfter(2)},
		{"fails on context deadline error calls fn 1",
			canceledCtx, LimitStrategy(3), 1,
			"whoopsie\ncontext canceled", succeedAfter(2)},
	}

	for _, tcase := range tsuite {
		t.Run(tcase.msg, func(t *testing.T) {
			var calls int
			err := Retry(tcase.ctx, tcase.strategy, func(ctx context.Context) (bool, error) {
				calls++
				return tcase.fn(ctx)
			})
			if tcase.wantErr != "" {
				assert.EqualError(t, err, tcase.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tcase.wantCalls, calls)
		})
	}
}

func succeedAfter(n int) func(context.Context) (bool, error) {
	return func(context.Context) (bool, error) {
		if n == 0 {
			return false, nil
		}
		n--
		return true, errors.New("whoopsie")
	}
}

func failAfter(n int) func(context.Context) (bool, error) {
	return func(context.Context) (bool, error) {
		if n == 0 {
			return false, errors.New("last one")
		}
		n--
		return true, errors.New("whoopsie")
	}
}
