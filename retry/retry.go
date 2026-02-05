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
	"math"
	"time"
)

// Strategy represents a retry strategy. See Retry for more details.
type Strategy func(count uint) (sleep time.Duration, stop bool)

// DefaultStrategy returns an exponential backoff strategy with a limit of 10, a min sleep
// time of 1ms and a max sleep time of 1024ms. At most, this strategy yields a retry mechanism
// that can last approximately 2 seconds.
var DefaultStrategy Strategy = CombinedStrategy(
	LimitStrategy(10), ExponentialStrategy(1*time.Millisecond, 1024*time.Millisecond))

// LimitStrategy returns a retry strategy that stops after limit number of retries
// has been reached.
func LimitStrategy(limit uint) Strategy {
	return func(count uint) (sleep time.Duration, stop bool) {
		return 0, count >= limit
	}
}

// ExponentialStrategy returns a retry strategy that never stops but increments
// the sleep time from min to max in power of 2 increments.
func ExponentialStrategy(min, max time.Duration) Strategy {
	return func(count uint) (sleep time.Duration, stop bool) {
		sleep = time.Duration(math.Pow(2, float64(count-1))) * min
		if sleep > max {
			sleep = max
		}
		return
	}
}

// SequentialStrategy returns a retry strategy that never stops and retries
// always after 'every' duration.
func SequentialStrategy(every time.Duration) Strategy {
	return func(count uint) (sleep time.Duration, stop bool) {
		sleep = every
		return
	}
}

// CombinedStrategy returns a retry strategy that combines all the given strategies
// using the following rules:
//   - If any returns stop=true, then stop=true is returned.
//   - If multiple return a sleep time that is non-zero, then the biggest sleep value is used.
func CombinedStrategy(i Strategy, n ...Strategy) Strategy {
	all := append([]Strategy{}, i)
	all = append(all, n...)
	return func(count uint) (sleep time.Duration, stop bool) {
		var max time.Duration
		for _, strategy := range all {
			sleep, stop = strategy(count)
			if stop {
				return
			}
			if sleep > max {
				max = sleep
			}
		}
		return max, false
	}
}

// Retry retries the given function with the given Strategy, until function returns
// no error, strategy returns stop=true, function returns error and retry=false,
// or ctx deadline is exceeded.
func Retry(
	ctx context.Context, strategy Strategy,
	fn func(ctx context.Context) (shouldRetry bool, err error),
) error {
	var retryCount uint
	var result error
	for {
		retry, err := fn(ctx)
		if err == nil {
			return nil
		}
		if !retry {
			// if we never allowed retries, pass error as is
			if result == nil {
				return err
			}
			return errors.Join(result, err)
		}
		result = errors.Join(result, err)

		retryCount++
		sleep, stop := strategy(retryCount)
		if stop {
			return result
		}
		// short-circuit if context is cancel
		// proceed to wait if it's not
		select {
		case <-ctx.Done():
			// don't pollute the results with extra context errors
			if !errors.Is(result, context.DeadlineExceeded) &&
				!errors.Is(result, context.Canceled) {
				result = errors.Join(result, ctx.Err())
			}
			return result
		default:
		}
		select {
		case <-ctx.Done():
			result = errors.Join(result, ctx.Err())
			return result
		case <-time.After(sleep):
		}
	}
}
