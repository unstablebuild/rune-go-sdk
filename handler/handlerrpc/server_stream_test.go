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

package handlerrpc_test

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/browserapi/browserrpc"
	"github.com/unstablebuild/rune-go-sdk/handler/handlerrpc"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/term/termrpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type scriptedStream struct {
	t       *testing.T
	dir     string
	script  []*handlerrpc.ServerMessage
	draws   []*handlerrpc.DrawStreamResponse
	handles []*handlerrpc.HandleStreamResponse
	files   []map[string][]byte
}

func (s *scriptedStream) Header() (metadata.MD, error) { return nil, nil }
func (s *scriptedStream) Trailer() metadata.MD         { return nil }
func (s *scriptedStream) CloseSend() error             { return nil }
func (s *scriptedStream) Context() context.Context     { return context.Background() }

func (s *scriptedStream) SendMsg(m any) error {
	if handle := m.(*browserrpc.SplitWindowMessage).GetHandle(); handle != nil {
		s.handles = append(s.handles, proto.Clone(handle).(*handlerrpc.HandleStreamResponse))
	}
	draw := m.(*browserrpc.SplitWindowMessage).GetDraw()
	if draw == nil {
		return nil
	}
	files := make(map[string][]byte)
	for _, up := range draw.GetUploads() {
		if up.GetFile() == "" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, up.GetFile()))
		require.NoError(s.t, err)
		files[up.GetFile()] = data
	}
	// SendMsg encodes m before it returns, so the sender may reuse what it
	// references.
	s.draws = append(s.draws, proto.Clone(draw).(*handlerrpc.DrawStreamResponse))
	s.files = append(s.files, files)
	return nil
}

func (s *scriptedStream) RecvMsg(m any) error {
	if len(s.script) == 0 {
		return io.EOF
	}
	msg := m.(*handlerrpc.ServerMessage)
	proto.Reset(msg)
	proto.Merge(msg, s.script[0])
	s.script = s.script[1:]
	return nil
}

type rasterHandler struct {
	versions      []uint64
	sizes         []image.Point
	width, height int
	calls         []string
	shown         []bool
	srcs          []*image.RGBA
}

func (h *rasterHandler) Resize(width, height int) {
	h.width, h.height = width, height
	h.calls = append(h.calls, fmt.Sprintf("resize %dx%d", width, height))
}

func (h *rasterHandler) Draw(w term.Writer) {
	if cw, ch, ok := term.CellPixelSizeFromContext(w.Context()); ok {
		h.calls = append(h.calls, fmt.Sprintf("draw %vx%v", cw, ch))
	} else {
		h.calls = append(h.calls, "draw")
	}
	n := len(h.shown)
	version := h.versions[n]
	size := image.Pt(8, 4)
	if n < len(h.sizes) {
		size = h.sizes[n]
	}
	src := image.NewRGBA(image.Rectangle{Max: size})
	for i := range src.Pix {
		src.Pix[i] = byte(i) + byte(version)
	}
	src.SetRGBA(0, 0, color.RGBA{A: 0xff})
	h.srcs = append(h.srcs, src)
	ok := w.DrawImage(term.Image{
		ID: 7, Version: version, Src: src, Width: h.width, Height: h.height,
	})
	h.shown = append(h.shown, ok)
	if ok {
		return
	}
	for y := range h.height {
		for x := range h.width {
			w.SetCell(term.Coordinates{X: x, Y: y}, term.Cell{Ch: 'F', Width: 1})
		}
	}
}

func (h *rasterHandler) Handle(term.Event) (bool, bool) { return false, false }
func (h *rasterHandler) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	return term.Coordinates{}, 0, false
}
func (h *rasterHandler) Selection() (string, bool) { return "", false }
func (h *rasterHandler) Close() error              { return nil }

type pointerHandler struct {
	rasterHandler
}

func (h *pointerHandler) Handle(ev term.Event) (bool, bool) {
	if ev.MouseX == 1 {
		term.SetPointerShape(ev, term.PointerColResize)
	}
	return false, true
}

func TestServerStreamHandleReportsThePointerShape(t *testing.T) {
	handle := func(ev term.Event) *handlerrpc.ServerMessage {
		var rpcEv termrpc.Event
		require.NoError(t, rpcEv.FromModel(ev))
		return &handlerrpc.ServerMessage{
			Type:   handlerrpc.MessageType_Handle,
			Handle: &handlerrpc.HandleStreamRequest{Event: &rpcEv},
		}
	}
	stream := &scriptedStream{t: t, script: []*handlerrpc.ServerMessage{
		handle(term.Event{Type: term.EventMouse, MouseX: 1}),
		handle(term.Event{Type: term.EventMouse, MouseX: 2}),
		handle(term.Event{Type: term.EventKey, Ch: 'a', MouseX: 1}),
	}}
	handlerrpc.NewServerStream(stream, &pointerHandler{}, func() *browserrpc.SplitWindowMessage {
		return new(browserrpc.SplitWindowMessage)
	}).ReceiveMessages()

	require.Len(t, stream.handles, 3)
	require.NotNil(t, stream.handles[0].PointerShape)
	assert.Equal(t, int32(term.PointerColResize), stream.handles[0].GetPointerShape())
	assert.Nil(t, stream.handles[1].PointerShape, "a handler that sets no shape reports none")
	assert.Nil(t, stream.handles[2].PointerShape, "key events carry no Pointer")
}

func TestServerStreamImages(t *testing.T) {
	dir := t.TempDir()
	resize := func(g *handlerrpc.Graphics) *handlerrpc.ServerMessage {
		return &handlerrpc.ServerMessage{
			Type:   handlerrpc.MessageType_Resize,
			Resize: &handlerrpc.ResizeStreamRequest{Width: 4, Height: 2, Graphics: g},
		}
	}
	draw := func(ack uint64) *handlerrpc.ServerMessage {
		return &handlerrpc.ServerMessage{
			Type: handlerrpc.MessageType_Draw,
			Draw: &handlerrpc.DrawStreamRequest{
				PackedOk: true, CellsUnchangedOk: true, ImageAck: ack,
			},
		}
	}
	graphics := func(cellWidth float64) *handlerrpc.Graphics {
		return &handlerrpc.Graphics{
			CellWidth: cellWidth, CellHeight: 2 * cellWidth,
			ArenaDir: dir, MaxBytes: 1 << 20,
		}
	}
	stream := &scriptedStream{t: t, dir: dir, script: []*handlerrpc.ServerMessage{
		resize(graphics(8)),
		draw(0),
		draw(1),
		draw(1),
		resize(graphics(10)),
		resize(nil),
		draw(2),
	}}
	h := &rasterHandler{versions: []uint64{1, 1, 2, 3}}
	handlerrpc.NewServerStream(stream, h, func() *browserrpc.SplitWindowMessage {
		return new(browserrpc.SplitWindowMessage)
	}).ReceiveMessages()

	assert.Equal(t, []bool{true, true, true, false}, h.shown)

	type upload struct {
		ID, Version, Seq uint64
		File             string
	}
	type frame struct {
		uploads        []upload
		versions       []uint64
		released       []uint64
		cellsUnchanged bool
	}
	want := []frame{
		{uploads: []upload{{7, 1, 1, "1-0"}}, versions: []uint64{1}},
		{versions: []uint64{1}, cellsUnchanged: true},
		{uploads: []upload{{7, 2, 2, "1-1"}}, versions: []uint64{2}, cellsUnchanged: true},
		{released: []uint64{7}},
	}
	require.Len(t, stream.draws, len(want))
	for i, resp := range stream.draws {
		var got frame
		for _, up := range resp.GetUploads() {
			got.uploads = append(got.uploads,
				upload{up.GetId(), up.GetVersion(), up.GetSeq(), up.GetFile()})
			assert.Equal(t, h.srcs[i].Pix, stream.files[i][up.GetFile()],
				"draw %d: the upload file must hold the handler's pixels", i)
			assert.Equal(t, int32(8), up.GetWidth())
			assert.Equal(t, int32(4), up.GetHeight())
		}
		for _, p := range resp.GetImages() {
			got.versions = append(got.versions, p.GetVersion())
			assert.Equal(t, term.Image{
				ID: 7, Version: p.GetVersion(), Width: 4, Height: 2,
				Clip: image.Rect(0, 0, 4, 2),
			}, p.ToModel(), "draw %d: the placement must be confined to the frame", i)
		}
		got.released = resp.GetReleased()
		got.cellsUnchanged = resp.GetCellsUnchanged()
		assert.Equal(t, want[i], got, "draw %d", i)
		assert.Equal(t, got.cellsUnchanged, resp.GetPacked() == nil,
			"draw %d: cells are sent unless unchanged", i)
	}

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Empty(t, entries, "the end of the stream must delete every upload file")
}

func TestServerStreamDrawsCarryTheCellPixelSize(t *testing.T) {
	resize := func(g *handlerrpc.Graphics) *handlerrpc.ServerMessage {
		return &handlerrpc.ServerMessage{
			Type:   handlerrpc.MessageType_Resize,
			Resize: &handlerrpc.ResizeStreamRequest{Width: 4, Height: 2, Graphics: g},
		}
	}
	draw := &handlerrpc.ServerMessage{
		Type: handlerrpc.MessageType_Draw,
		Draw: &handlerrpc.DrawStreamRequest{PackedOk: true},
	}
	sized := func(width, height float64) *handlerrpc.Graphics {
		return &handlerrpc.Graphics{CellWidth: width, CellHeight: height}
	}

	tests := []struct {
		name   string
		script []*handlerrpc.ServerMessage
		calls  []string
		shown  []bool
	}{
		{
			name:   "a draw sees the size of the last resize",
			script: []*handlerrpc.ServerMessage{resize(sized(8.75, 17.5)), draw, draw},
			calls:  []string{"resize 4x2", "draw 8.75x17.5", "draw 8.75x17.5"},
			shown:  []bool{true, true},
		},
		{
			name:   "a draw sees no size before the host sends graphics",
			script: []*handlerrpc.ServerMessage{resize(nil), draw},
			calls:  []string{"resize 4x2", "draw"},
			shown:  []bool{false},
		},
		{
			name: "a draw sees no size once the host stops sending graphics",
			script: []*handlerrpc.ServerMessage{
				resize(sized(8, 16)), draw, resize(nil), draw,
			},
			calls: []string{"resize 4x2", "draw 8x16", "resize 4x2", "draw"},
			shown: []bool{true, false},
		},
		{
			name: "a draw sees a new pixel size at the same size in cells",
			script: []*handlerrpc.ServerMessage{
				resize(sized(8, 16)), draw, resize(sized(10.5, 21)), draw,
			},
			calls: []string{"resize 4x2", "draw 8x16", "resize 4x2", "draw 10.5x21"},
			shown: []bool{true, true},
		},
		{
			name:   "a size with a zero side is no size",
			script: []*handlerrpc.ServerMessage{resize(sized(0, 16)), draw},
			calls:  []string{"resize 4x2", "draw"},
			shown:  []bool{false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := &scriptedStream{t: t, script: tt.script}
			h := &rasterHandler{versions: make([]uint64, len(tt.shown))}
			handlerrpc.NewServerStream(stream, h, func() *browserrpc.SplitWindowMessage {
				return new(browserrpc.SplitWindowMessage)
			}).ReceiveMessages()

			assert.Equal(t, tt.calls, h.calls)
			assert.Equal(t, tt.shown, h.shown,
				"images must be shown exactly when the host sent graphics")
		})
	}
}

func TestServerStreamShowsAnImageOverBudgetOnceItFits(t *testing.T) {
	draw := &handlerrpc.ServerMessage{
		Type: handlerrpc.MessageType_Draw,
		Draw: &handlerrpc.DrawStreamRequest{PackedOk: true},
	}
	stream := &scriptedStream{t: t, dir: t.TempDir(), script: []*handlerrpc.ServerMessage{{
		Type: handlerrpc.MessageType_Resize,
		Resize: &handlerrpc.ResizeStreamRequest{Width: 4, Height: 2, Graphics: &handlerrpc.Graphics{
			CellWidth: 8, CellHeight: 16, MaxBytes: 4 * 8 * 4,
		}},
	}, draw, draw}}
	stream.script[0].Resize.Graphics.ArenaDir = stream.dir
	h := &rasterHandler{
		versions: []uint64{1, 2},
		sizes:    []image.Point{image.Pt(16, 8), image.Pt(8, 4)},
	}
	handlerrpc.NewServerStream(stream, h, func() *browserrpc.SplitWindowMessage {
		return new(browserrpc.SplitWindowMessage)
	}).ReceiveMessages()

	assert.Equal(t, []bool{false, true}, h.shown)
	require.Len(t, stream.draws, 2)
	over, fits := stream.draws[0], stream.draws[1]
	assert.Empty(t, over.GetUploads(), "pixels over budget must not be uploaded")
	assert.Empty(t, over.GetImages())
	assert.Equal(t, uint32('F'), over.GetPacked().GetChars()[0],
		"the handler must draw its fallback for an image not shown")
	require.Len(t, fits.GetUploads(), 1)
	assert.EqualValues(t, 2, fits.GetUploads()[0].GetVersion())
	require.Len(t, fits.GetImages(), 1)
	assert.EqualValues(t, 2, fits.GetImages()[0].GetVersion())
}

func TestServerStreamStreamsPixels(t *testing.T) {
	draw := &handlerrpc.ServerMessage{
		Type: handlerrpc.MessageType_Draw,
		Draw: &handlerrpc.DrawStreamRequest{PackedOk: true, CellsUnchangedOk: true},
	}
	stream := &scriptedStream{t: t, script: []*handlerrpc.ServerMessage{{
		Type: handlerrpc.MessageType_Resize,
		Resize: &handlerrpc.ResizeStreamRequest{Width: 4, Height: 2, Graphics: &handlerrpc.Graphics{
			CellWidth: 8, CellHeight: 16, MaxBytes: 1 << 20,
		}},
	}, draw, draw, draw}}
	h := &rasterHandler{versions: []uint64{1, 1, 2}}
	handlerrpc.NewServerStream(stream, h, func() *browserrpc.SplitWindowMessage {
		return new(browserrpc.SplitWindowMessage)
	}).ReceiveMessages()

	assert.Equal(t, []bool{true, true, true}, h.shown,
		"a host without an arena must still show images")
	type sent struct {
		more     bool
		uploads  []uint64
		versions []uint64
	}
	var got []sent
	frame := 0
	for _, resp := range stream.draws {
		var s sent
		s.more = resp.GetMore()
		for _, up := range resp.GetUploads() {
			s.uploads = append(s.uploads, up.GetVersion())
			assert.Empty(t, up.GetFile())
			assert.Zero(t, up.GetOffset())
			assert.Equal(t, h.srcs[frame].Pix, up.GetData(),
				"draw %d: the streamed record must hold the handler's pixels", frame)
		}
		for _, p := range resp.GetImages() {
			s.versions = append(s.versions, p.GetVersion())
		}
		if !s.more {
			frame++
		}
		got = append(got, s)
	}
	assert.Equal(t, []sent{
		{more: true, uploads: []uint64{1}},
		{versions: []uint64{1}},
		{versions: []uint64{1}},
		{more: true, uploads: []uint64{2}},
		{versions: []uint64{2}},
	}, got, "pixels must be streamed ahead of the response that places them")
}

func TestServerStreamCellsUnchangedNeedsTheHostsConsent(t *testing.T) {
	draw := &handlerrpc.ServerMessage{
		Type: handlerrpc.MessageType_Draw,
		Draw: &handlerrpc.DrawStreamRequest{PackedOk: true},
	}
	stream := &scriptedStream{t: t, script: []*handlerrpc.ServerMessage{{
		Type:   handlerrpc.MessageType_Resize,
		Resize: &handlerrpc.ResizeStreamRequest{Width: 4, Height: 2},
	}, draw, draw}}
	h := &rasterHandler{versions: []uint64{1, 1}}
	handlerrpc.NewServerStream(stream, h, func() *browserrpc.SplitWindowMessage {
		return new(browserrpc.SplitWindowMessage)
	}).ReceiveMessages()

	assert.Equal(t, []string{"resize 4x2", "draw", "draw"}, h.calls,
		"a host without graphics must not tell draws a cell size")
	assert.Equal(t, []bool{false, false}, h.shown)
	require.Len(t, stream.draws, 2)
	for _, resp := range stream.draws {
		assert.False(t, resp.GetCellsUnchanged())
		assert.NotNil(t, resp.GetPacked())
		assert.Empty(t, resp.GetImages())
		assert.Empty(t, resp.GetUploads())
	}
}
