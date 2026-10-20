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

package textrpc

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/debug"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"google.golang.org/protobuf/proto"
)

type resourceCompletionHandler func(context.Context, string) (iterator.Iterator[string], error)

func (h resourceCompletionHandler) CompleteResource(
	ctx context.Context, prefix string,
) (iterator.Iterator[string], error) {
	return h(ctx, prefix)
}

func (h resourceCompletionHandler) OpenResource(context.Context, workspaceapi.URI) (browserapi.Handler, error) {
	panic("completion must not open a resource")
}

type resourceCompletionStream struct {
	ctx    context.Context
	recv   chan *ServerResourceOpenerMessage
	sent   chan *ClientResourceOpenerMessage
	closed chan struct{}
}

func (s *resourceCompletionStream) RecvMsg(dst any) error {
	select {
	case msg := <-s.recv:
		if msg == nil {
			return io.EOF
		}
		proto.Merge(dst.(*ServerResourceOpenerMessage), msg)
		return nil
	case <-s.ctx.Done():
		return io.EOF
	}
}

func (s *resourceCompletionStream) Send(msg *ClientResourceOpenerMessage) error {
	select {
	case s.sent <- proto.Clone(msg).(*ClientResourceOpenerMessage):
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

func (s *resourceCompletionStream) CloseSend() error {
	close(s.closed)
	return nil
}

func newResourceCompletionTestStream(
	t *testing.T, h resourceCompletionHandler,
) (*resourceOpenerServerStream, *resourceCompletionStream) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	wire := &resourceCompletionStream{
		ctx: ctx, recv: make(chan *ServerResourceOpenerMessage, 8),
		sent: make(chan *ClientResourceOpenerMessage, 8), closed: make(chan struct{}),
	}
	s := newResourceOpenerServerStream(ctx, wire, nil, h)
	done := make(chan struct{})
	go debug.CapturePanicReport(func() {
		defer close(done)
		s.receiveMessages()
	})
	t.Cleanup(func() {
		cancel()
		awaitResourceCompletion(t, done)
		awaitResourceCompletion(t, wire.closed)
	})
	return s, wire
}

func awaitResourceCompletion[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for resource completion stream")
		var zero T
		return zero
	}
}

func resourceCompletionRequest(id int64, prefix string) *ServerResourceOpenerMessage {
	return &ServerResourceOpenerMessage{
		Type:     ServerResourceOpenerMessage_Complete,
		Complete: &CompleteResourceRequest{Id: id, Prefix: prefix},
	}
}

func resourceCompletionCancel(id int64) *ServerResourceOpenerMessage {
	return &ServerResourceOpenerMessage{
		Type:           ServerResourceOpenerMessage_CompleteCancel,
		CompleteCancel: &RequestCancel{Id: id},
	}
}

// receiveResourceCompletion returns the values the completion id streams
// before it is reported done, and the error it is reported done with.
func receiveResourceCompletion(
	t *testing.T, wire *resourceCompletionStream, id int64,
) ([]string, string) {
	t.Helper()
	var values []string
	for {
		msg := awaitResourceCompletion(t, wire.sent)
		require.Nil(t, msg.GetOpen())
		switch msg.GetType() {
		case ClientResourceOpenerMessage_CompleteValue:
			require.Equal(t, id, msg.GetCompleteValue().GetId())
			values = append(values, msg.GetCompleteValue().GetValue())
		case ClientResourceOpenerMessage_CompleteDone:
			require.Equal(t, id, msg.GetCompleteDone().GetId())
			return values, msg.GetCompleteDone().GetError()
		default:
			t.Fatalf("unexpected message: %v", msg)
		}
	}
}

func TestResourceOpenerStreamComplete(t *testing.T) {
	failure := errors.New("completion failed")
	for _, tc := range []struct {
		name    string
		prefix  string
		values  []string
		openErr error
		iterErr error
	}{
		{name: "empty prefix and no values"},
		{
			name:   "preserves opaque strings and order",
			prefix: " fake://host/a%2Fb?x=é# ",
			values: []string{"z", "fake://host/a%2Fb?x=é# ", "", "z", " a b "},
		},
		{name: "fails to start", prefix: "fake:", openErr: failure},
		{name: "fails after values", prefix: "fake://", values: []string{"fake://a"}, iterErr: failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prefixes := make(chan string, 1)
			closed := make(chan struct{})
			s, wire := newResourceCompletionTestStream(t, func(
				_ context.Context, prefix string,
			) (iterator.Iterator[string], error) {
				prefixes <- prefix
				if tc.openErr != nil {
					close(closed)
					return nil, tc.openErr
				}
				values := tc.values
				return iterator.FromFunc(func(context.Context) (string, bool, error) {
					if len(values) == 0 {
						return "", false, tc.iterErr
					}
					value := values[0]
					values = values[1:]
					return value, true, nil
				}, func() error { close(closed); return nil }), nil
			})
			wire.recv <- resourceCompletionRequest(71, tc.prefix)
			require.Equal(t, tc.prefix, awaitResourceCompletion(t, prefixes))
			values, errStr := receiveResourceCompletion(t, wire, 71)
			require.Equal(t, tc.values, values)
			wantErr := cmp.Or(tc.openErr, tc.iterErr)
			if wantErr == nil {
				require.Empty(t, errStr)
			} else {
				require.Contains(t, errStr, wantErr.Error())
			}
			awaitResourceCompletion(t, closed)
			require.Eventually(t, func() bool {
				_, ok := s.completions.Load(int64(71))
				return !ok
			}, time.Second, time.Millisecond)
		})
	}
}

func TestResourceOpenerStreamCompleteCancellation(t *testing.T) {
	for _, disconnect := range []bool{false, true} {
		t.Run(fmt.Sprintf("disconnect=%t", disconnect), func(t *testing.T) {
			canceled := make(chan struct{})
			closed := make(chan struct{})
			_, wire := newResourceCompletionTestStream(t, func(
				ctx context.Context, prefix string,
			) (iterator.Iterator[string], error) {
				if prefix != "fake://blocked" {
					return iterator.FromSlice([]string{prefix}), nil
				}
				context.AfterFunc(ctx, func() { close(canceled) })
				first := true
				return iterator.FromFunc(func(nextCtx context.Context) (string, bool, error) {
					if first {
						first = false
						return "fake://blocked/first", true, nil
					}
					<-nextCtx.Done()
					return "", false, nextCtx.Err()
				}, func() error { close(closed); return nil }), nil
			})
			wire.recv <- resourceCompletionRequest(11, "fake://blocked")
			value := awaitResourceCompletion(t, wire.sent).GetCompleteValue()
			require.Equal(t, int64(11), value.GetId())
			require.Equal(t, "fake://blocked/first", value.GetValue())

			// Another completion is served while the first one blocks.
			wire.recv <- resourceCompletionRequest(12, "fake://second")
			values, errStr := receiveResourceCompletion(t, wire, 12)
			require.Equal(t, []string{"fake://second"}, values)
			require.Empty(t, errStr)

			if disconnect {
				wire.recv <- nil
			} else {
				wire.recv <- resourceCompletionCancel(11)
			}
			awaitResourceCompletion(t, canceled)
			awaitResourceCompletion(t, closed)
			if disconnect {
				return
			}
			values, errStr = receiveResourceCompletion(t, wire, 11)
			require.Empty(t, values)
			require.Empty(t, errStr, "a canceled completion is done, not failed")

			wire.recv <- resourceCompletionRequest(13, "fake://third")
			values, _ = receiveResourceCompletion(t, wire, 13)
			require.Equal(t, []string{"fake://third"}, values)
		})
	}
}

func TestResourceOpenerStreamCompleteStaleCancel(t *testing.T) {
	for _, id := range []int64{41, 999} {
		t.Run(fmt.Sprint(id), func(t *testing.T) {
			_, wire := newResourceCompletionTestStream(t, func(
				_ context.Context, prefix string,
			) (iterator.Iterator[string], error) {
				return iterator.FromSlice([]string{prefix}), nil
			})
			wire.recv <- resourceCompletionRequest(41, "fake://first")
			receiveResourceCompletion(t, wire, 41)
			wire.recv <- resourceCompletionCancel(id)
			wire.recv <- resourceCompletionRequest(42, "fake://second")
			values, errStr := receiveResourceCompletion(t, wire, 42)
			require.Equal(t, []string{"fake://second"}, values)
			require.Empty(t, errStr)
		})
	}
}
