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

package handlerrpc

import (
	"context"
	"io"
	"sync/atomic"
	"testing"

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/term/termrpc"
	"google.golang.org/grpc/metadata"
)

// ---------------------------------------------------------------------------
// Draw response writer benchmark — proves that the cell slab eliminates
// per-cell heap allocations. allocs/op should be constant (~5) regardless
// of screen size.
// ---------------------------------------------------------------------------

func benchDrawResponseWriter(b *testing.B, width, height int) {
	ctx := context.Background()
	cell := term.Cell{
		Ch:    'A',
		Width: 1,
		Bytes: 1,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		w := newDrawResponseWriter(ctx, width, height)
		for y := range height {
			for x := range width {
				w.SetCell(term.Coordinates{X: x, Y: y}, cell)
			}
		}
	}
}

func BenchmarkDrawResponseWriter(b *testing.B) {
	// allocs/op should remain constant across sizes — only the slab
	// allocations in newDrawResponseWriter, not one per cell.
	b.Run("60x15", func(b *testing.B) { benchDrawResponseWriter(b, 60, 15) })
	b.Run("120x30", func(b *testing.B) { benchDrawResponseWriter(b, 120, 30) })
	b.Run("240x60", func(b *testing.B) { benchDrawResponseWriter(b, 240, 60) })
}

// ---------------------------------------------------------------------------
// Mouse event coalescing benchmark — proves that a burst of N mouse
// events is reduced to 1 actual Handle call.
// ---------------------------------------------------------------------------

// benchMsg is a minimal StreamMessage implementation for benchmarking.
type benchMsg struct {
	handle *HandleStreamResponse
	draw   *DrawStreamResponse
}

func (m *benchMsg) ProtoMessage()                             {}
func (m *benchMsg) GetType() MessageType                      { return 0 }
func (m *benchMsg) GetDraw() *DrawStreamResponse              { return m.draw }
func (m *benchMsg) GetHandle() *HandleStreamResponse          { return m.handle }
func (m *benchMsg) GetClose() *CloseStreamResponse            { return nil }
func (m *benchMsg) GetCursor() *CursorStreamResponse          { return nil }
func (m *benchMsg) GetSelection() *SelectionStreamResponse    { return nil }
func (m *benchMsg) GetDimensions() *DimensionsStreamResponse  { return nil }
func (m *benchMsg) SetDraw(r *DrawStreamResponse)             { m.draw = r }
func (m *benchMsg) SetHandle(r *HandleStreamResponse)         { m.handle = r }
func (m *benchMsg) SetClose(*CloseStreamResponse)             {}
func (m *benchMsg) SetCursor(*CursorStreamResponse)           {}
func (m *benchMsg) SetSelection(*SelectionStreamResponse)     {}
func (m *benchMsg) SetDimensions(*DimensionsStreamResponse)    {}

// nopStream satisfies grpc.ClientStream with no-op methods.
type nopStream struct {
	sends atomic.Int64
}

func (s *nopStream) SendMsg(any) error                    { s.sends.Add(1); return nil }
func (s *nopStream) RecvMsg(any) error                    { return io.EOF }
func (s *nopStream) CloseSend() error                     { return nil }
func (s *nopStream) Context() context.Context             { return context.Background() }
func (s *nopStream) Header() (metadata.MD, error)         { return nil, nil }
func (s *nopStream) Trailer() metadata.MD                 { return nil }

func mouseHandleMsg(x, y int) *ServerMessage {
	return &ServerMessage{
		Type: MessageType_Handle,
		Handle: &HandleStreamRequest{
			Event: &termrpc.Event{
				Type:   termrpc.Event_TypeMouse,
				Key:    termrpc.Event_MouseLeft,
				MouseX: int32(x),
				MouseY: int32(y),
			},
		},
	}
}

func benchCoalesceMouse(b *testing.B, burstSize int) {
	stream := &nopStream{}
	ss := NewServerStream[*benchMsg](stream, nil, func() *benchMsg {
		return new(benchMsg)
	})

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		stream.sends.Store(0)

		ch := make(chan recvResult, burstSize)
		for i := range burstSize {
			ch <- recvResult{msg: mouseHandleMsg(i, 5)}
		}

		first := mouseHandleMsg(0, 5)
		final, pending := ss.coalesceMouse(first, first.Handle.Event, ch)

		if pending != nil {
			b.Fatal("unexpected pending message")
		}
		if final.Handle.Event.MouseX != int32(burstSize-1) {
			b.Fatalf("expected last event x=%d, got %d",
				burstSize-1, final.Handle.Event.MouseX)
		}
		// burstSize no-op responses should have been sent for the
		// coalesced events; only the final event survives.
		if n := stream.sends.Load(); n != int64(burstSize) {
			b.Fatalf("expected %d SendMsg calls, got %d", burstSize, n)
		}
	}
}

func BenchmarkCoalesceMouse(b *testing.B) {
	b.Run("burst_10", func(b *testing.B) { benchCoalesceMouse(b, 10) })
	b.Run("burst_50", func(b *testing.B) { benchCoalesceMouse(b, 50) })
	b.Run("burst_100", func(b *testing.B) { benchCoalesceMouse(b, 100) })
}

// ---------------------------------------------------------------------------
// Full ReceiveMessages benchmark — end-to-end with mock handler.
// Compares burst (coalesced) vs interleaved (uncoalesced) mouse events
// to show real-world improvement.
// ---------------------------------------------------------------------------

// benchHandler records Handle and Draw call counts.
type benchHandler struct {
	handleCalls atomic.Int64
	drawCalls   atomic.Int64
	width       int
	height      int
}

func (h *benchHandler) Handle(term.Event) (bool, bool) {
	h.handleCalls.Add(1)
	return false, true
}

func (h *benchHandler) Draw(w term.Writer) {
	h.drawCalls.Add(1)
	for y := range h.height {
		for x := range h.width {
			w.SetCell(term.Coordinates{X: x, Y: y}, term.Cell{
				Ch: 'X', Width: 1, Bytes: 1,
			})
		}
	}
}

func (h *benchHandler) Resize(w, ht int) { h.width = w; h.height = ht }

func (h *benchHandler) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	return term.Coordinates{}, 0, false
}

func (h *benchHandler) Selection() (string, bool) { return "", false }
func (h *benchHandler) Close() error              { return nil }

// feedStream is a grpc.ClientStream backed by a pre-built message list.
type feedStream struct {
	msgs []*ServerMessage
	idx  atomic.Int64
	nopStream
}

func (s *feedStream) RecvMsg(m any) error {
	i := int(s.idx.Add(1)) - 1
	if i >= len(s.msgs) {
		return io.EOF
	}
	dst := m.(*ServerMessage)
	src := s.msgs[i]
	dst.Type = src.Type
	dst.Handle = src.Handle
	dst.Draw = src.Draw
	return nil
}

func BenchmarkReceiveMessages(b *testing.B) {
	const (
		numEvents = 100
		width     = 120
		height    = 30
	)

	// burst: 100 mouse Handles then 1 Draw.
	// With coalescing most Handles are skipped.
	b.Run("burst", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			msgs := make([]*ServerMessage, 0, numEvents+1)
			for i := range numEvents {
				msgs = append(msgs, mouseHandleMsg(i%width, i/width))
			}
			msgs = append(msgs, &ServerMessage{Type: MessageType_Draw})

			handler := &benchHandler{width: width, height: height}
			stream := &feedStream{msgs: msgs}
			ss := NewServerStream[*benchMsg](stream, handler, func() *benchMsg {
				return new(benchMsg)
			})
			ss.width.Store(width)
			ss.height.Store(height)

			ss.ReceiveMessages()

			if h := handler.handleCalls.Load(); h >= int64(numEvents) {
				b.Fatalf("coalescing failed: %d handle calls for %d events", h, numEvents)
			}
			if d := handler.drawCalls.Load(); d != 1 {
				b.Fatalf("expected 1 draw call, got %d", d)
			}
		}
	})

	// interleaved: Handle then Draw, repeated — no coalescing possible.
	b.Run("interleaved", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			msgs := make([]*ServerMessage, 0, numEvents*2)
			for i := range numEvents {
				msgs = append(msgs,
					mouseHandleMsg(i%width, i/width),
					&ServerMessage{Type: MessageType_Draw},
				)
			}

			handler := &benchHandler{width: width, height: height}
			stream := &feedStream{msgs: msgs}
			ss := NewServerStream[*benchMsg](stream, handler, func() *benchMsg {
				return new(benchMsg)
			})
			ss.width.Store(width)
			ss.height.Store(height)

			ss.ReceiveMessages()

			if h := handler.handleCalls.Load(); h != int64(numEvents) {
				b.Fatalf("expected %d handle calls, got %d", numEvents, h)
			}
			if d := handler.drawCalls.Load(); d != int64(numEvents) {
				b.Fatalf("expected %d draw calls, got %d", numEvents, d)
			}
		}
	})
}
