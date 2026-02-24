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

package joincontext_test

import (
	"context"
	"testing"
	"time"

	"github.com/unstablebuild/rune-go-sdk/joincontext"
)

func TestJoinContexts_Cancellation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func() (context.Context, context.Context, context.CancelFunc)
		cancelJoin bool
		wantErr    error
	}{
		{
			name: "ctx1 canceled first",
			setup: func() (context.Context, context.Context, context.CancelFunc) {
				ctx1, cancel1 := context.WithCancel(context.Background())
				ctx2 := context.Background()
				cancel1()
				return ctx1, ctx2, nil
			},
			wantErr: context.Canceled,
		},
		{
			name: "ctx2 canceled first",
			setup: func() (context.Context, context.Context, context.CancelFunc) {
				ctx1 := context.Background()
				ctx2, cancel2 := context.WithCancel(context.Background())
				cancel2()
				return ctx1, ctx2, nil
			},
			wantErr: context.Canceled,
		},
		{
			name: "explicit cancel",
			setup: func() (context.Context, context.Context, context.CancelFunc) {
				return context.Background(), context.Background(), nil
			},
			cancelJoin: true,
			wantErr:    context.Canceled,
		},
		{
			name: "ctx1 deadline exceeded",
			setup: func() (context.Context, context.Context, context.CancelFunc) {
				ctx1, cancel1 := context.WithDeadline(
					context.Background(),
					time.Now().Add(-time.Second),
				)
				ctx2 := context.Background()
				return ctx1, ctx2, cancel1
			},
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "both contexts already canceled",
			setup: func() (context.Context, context.Context, context.CancelFunc) {
				ctx1, cancel1 := context.WithCancel(context.Background())
				ctx2, cancel2 := context.WithCancel(context.Background())
				cancel1()
				cancel2()
				return ctx1, ctx2, nil
			},
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx1, ctx2, cleanup := tt.setup()
			if cleanup != nil {
				defer cleanup()
			}

			joined, cancel := joincontext.New(ctx1, ctx2)
			defer cancel()

			if tt.cancelJoin {
				cancel()
			}

			select {
			case <-joined.Done():
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for Done()")
			}

			if got := joined.Err(); got != tt.wantErr {
				t.Errorf("Err() = %v, want %v", got, tt.wantErr)
			}
		})
	}
}

func TestJoinContexts_Deadline(t *testing.T) {
	t.Parallel()

	now := time.Now()
	earlier := now.Add(time.Second)
	later := now.Add(time.Minute)

	tests := []struct {
		name     string
		ctx1     context.Context
		ctx2     context.Context
		wantTime time.Time
		wantOK   bool
	}{
		{
			name:   "no deadlines",
			ctx1:   context.Background(),
			ctx2:   context.Background(),
			wantOK: false,
		},
		{
			name:     "only ctx1 has deadline",
			ctx1:     deadlineCtx(t, earlier),
			ctx2:     context.Background(),
			wantTime: earlier,
			wantOK:   true,
		},
		{
			name:     "only ctx2 has deadline",
			ctx1:     context.Background(),
			ctx2:     deadlineCtx(t, later),
			wantTime: later,
			wantOK:   true,
		},
		{
			name:     "ctx1 earlier",
			ctx1:     deadlineCtx(t, earlier),
			ctx2:     deadlineCtx(t, later),
			wantTime: earlier,
			wantOK:   true,
		},
		{
			name:     "ctx2 earlier",
			ctx1:     deadlineCtx(t, later),
			ctx2:     deadlineCtx(t, earlier),
			wantTime: earlier,
			wantOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			joined, cancel := joincontext.New(tt.ctx1, tt.ctx2)
			defer cancel()

			got, ok := joined.Deadline()
			if ok != tt.wantOK {
				t.Fatalf("Deadline() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && !got.Equal(tt.wantTime) {
				t.Errorf(
					"Deadline() = %v, want %v",
					got, tt.wantTime,
				)
			}
		})
	}
}

func TestJoinContexts_Value(t *testing.T) {
	t.Parallel()

	type ctxKey string

	tests := []struct {
		name string
		ctx1 context.Context
		ctx2 context.Context
		key  ctxKey
		want any
	}{
		{
			name: "value from ctx1",
			ctx1: context.WithValue(
				context.Background(), ctxKey("k"), "v1",
			),
			ctx2: context.Background(),
			key:  "k",
			want: "v1",
		},
		{
			name: "fallback to ctx2",
			ctx1: context.Background(),
			ctx2: context.WithValue(
				context.Background(), ctxKey("k"), "v2",
			),
			key:  "k",
			want: "v2",
		},
		{
			name: "ctx1 preferred over ctx2",
			ctx1: context.WithValue(
				context.Background(), ctxKey("k"), "v1",
			),
			ctx2: context.WithValue(
				context.Background(), ctxKey("k"), "v2",
			),
			key:  "k",
			want: "v1",
		},
		{
			name: "not found returns nil",
			ctx1: context.Background(),
			ctx2: context.Background(),
			key:  "missing",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			joined, cancel := joincontext.New(tt.ctx1, tt.ctx2)
			defer cancel()

			if got := joined.Value(tt.key); got != tt.want {
				t.Errorf("Value(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestJoinContexts_CancelIdempotent(t *testing.T) {
	t.Parallel()

	joined, cancel := joincontext.New(
		context.Background(), context.Background(),
	)
	cancel()
	cancel() // must not panic

	select {
	case <-joined.Done():
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Done()")
	}
}

func deadlineCtx(t *testing.T, dl time.Time) context.Context {
	t.Helper()

	ctx, cancel := context.WithDeadline(context.Background(), dl)
	t.Cleanup(cancel)
	return ctx
}
