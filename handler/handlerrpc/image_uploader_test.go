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
	"errors"
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func gradient(w, h int, seed byte) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.SetRGBA(x, y, color.RGBA{
				R: seed, G: byte(x * 16), B: byte(y * 16), A: 0xff,
			})
		}
	}
	return img
}

func rgbaBytes(src image.Image) []byte {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Rect, src, b.Min, draw.Src)
	return dst.Pix
}

type wantUpload struct {
	ID, Version, Seq uint64
	File             string
}

type wantPlacement struct {
	ID, Version uint64
	Crop        image.Rectangle
}

type uploadStep struct {
	ack   uint64
	place []term.Image

	shown    []bool
	uploads  []wantUpload
	images   []wantPlacement
	released []uint64
	redraw   bool
	files    []string
}

func TestImageUploader(t *testing.T) {
	const width, height = 10, 5
	frame := func(seed byte) *image.RGBA { return gradient(2, 2, seed) }
	a1, a2, a3 := frame(1), frame(2), frame(3)
	b1 := frame(9)
	at := func(id term.ImageID, version uint64, src image.Image) term.Image {
		return term.Image{ID: id, Version: version, Src: src, Width: 2, Height: 1}
	}
	wide := gradient(6, 4, 7)
	offset := wide.SubImage(image.Rect(1, 1, 4, 3))

	tests := []struct {
		name     string
		noArena  bool
		maxBytes uint64
		steps    []uploadStep
	}{
		{
			name:    "no graphics",
			noArena: true,
			steps: []uploadStep{{
				place: []term.Image{at(1, 1, a1)},
				shown: []bool{false},
			}},
		},
		{
			name: "the first placement uploads",
			steps: []uploadStep{{
				place:   []term.Image{at(1, 1, a1)},
				shown:   []bool{true},
				uploads: []wantUpload{{1, 1, 1, "1-0"}},
				images:  []wantPlacement{{ID: 1, Version: 1}},
				files:   []string{"1-0"},
			}},
		},
		{
			name: "the same version is not uploaded again",
			steps: []uploadStep{{
				place:   []term.Image{at(1, 1, a1), at(1, 1, a1)},
				shown:   []bool{true, true},
				uploads: []wantUpload{{1, 1, 1, "1-0"}},
				images:  []wantPlacement{{ID: 1, Version: 1}, {ID: 1, Version: 1}},
				files:   []string{"1-0"},
			}, {
				place:  []term.Image{at(1, 1, a1)},
				shown:  []bool{true},
				images: []wantPlacement{{ID: 1, Version: 1}},
				files:  []string{"1-0"},
			}},
		},
		{
			name: "a new version uses the other file while the first is not acked",
			steps: []uploadStep{{
				place:   []term.Image{at(1, 1, a1)},
				shown:   []bool{true},
				uploads: []wantUpload{{1, 1, 1, "1-0"}},
				images:  []wantPlacement{{ID: 1, Version: 1}},
				files:   []string{"1-0"},
			}, {
				place:   []term.Image{at(1, 2, a2)},
				shown:   []bool{true},
				uploads: []wantUpload{{1, 2, 2, "1-1"}},
				images:  []wantPlacement{{ID: 1, Version: 2}},
				files:   []string{"1-0", "1-1"},
			}},
		},
		{
			name: "both files busy places the last version uploaded",
			steps: []uploadStep{{
				place:   []term.Image{at(1, 1, a1)},
				shown:   []bool{true},
				uploads: []wantUpload{{1, 1, 1, "1-0"}},
				images:  []wantPlacement{{ID: 1, Version: 1}},
				files:   []string{"1-0"},
			}, {
				place:   []term.Image{at(1, 2, a2)},
				shown:   []bool{true},
				uploads: []wantUpload{{1, 2, 2, "1-1"}},
				images:  []wantPlacement{{ID: 1, Version: 2}},
				files:   []string{"1-0", "1-1"},
			}, {
				place:  []term.Image{at(1, 3, a3)},
				shown:  []bool{true},
				images: []wantPlacement{{ID: 1, Version: 2}},
				redraw: true,
				files:  []string{"1-0", "1-1"},
			}},
		},
		{
			name: "an ack frees the file it covers",
			steps: []uploadStep{{
				place:   []term.Image{at(1, 1, a1)},
				shown:   []bool{true},
				uploads: []wantUpload{{1, 1, 1, "1-0"}},
				images:  []wantPlacement{{ID: 1, Version: 1}},
				files:   []string{"1-0"},
			}, {
				place:   []term.Image{at(1, 2, a2)},
				shown:   []bool{true},
				uploads: []wantUpload{{1, 2, 2, "1-1"}},
				images:  []wantPlacement{{ID: 1, Version: 2}},
				files:   []string{"1-0", "1-1"},
			}, {
				ack:     1,
				place:   []term.Image{at(1, 3, a3)},
				shown:   []bool{true},
				uploads: []wantUpload{{1, 3, 3, "1-0"}},
				images:  []wantPlacement{{ID: 1, Version: 3}},
				files:   []string{"1-0", "1-1"},
			}},
		},
		{
			name: "an image not placed is released and its acked files deleted",
			steps: []uploadStep{{
				place:   []term.Image{at(1, 1, a1), at(2, 1, b1)},
				shown:   []bool{true, true},
				uploads: []wantUpload{{1, 1, 1, "1-0"}, {2, 1, 2, "2-0"}},
				images:  []wantPlacement{{ID: 1, Version: 1}, {ID: 2, Version: 1}},
				files:   []string{"1-0", "2-0"},
			}, {
				ack:      2,
				place:    []term.Image{at(1, 1, a1)},
				shown:    []bool{true},
				images:   []wantPlacement{{ID: 1, Version: 1}},
				released: []uint64{2},
				files:    []string{"1-0"},
			}},
		},
		{
			name: "a released file is kept until the host acks it",
			steps: []uploadStep{{
				place:   []term.Image{at(1, 1, a1), at(2, 1, b1)},
				shown:   []bool{true, true},
				uploads: []wantUpload{{1, 1, 1, "1-0"}, {2, 1, 2, "2-0"}},
				images:  []wantPlacement{{ID: 1, Version: 1}, {ID: 2, Version: 1}},
				files:   []string{"1-0", "2-0"},
			}, {
				ack:      1,
				place:    []term.Image{at(1, 1, a1)},
				shown:    []bool{true},
				images:   []wantPlacement{{ID: 1, Version: 1}},
				released: []uint64{2},
				files:    []string{"1-0", "2-0"},
			}, {
				ack:    2,
				place:  []term.Image{at(1, 1, a1)},
				shown:  []bool{true},
				images: []wantPlacement{{ID: 1, Version: 1}},
				files:  []string{"1-0"},
			}},
		},
		{
			name: "an image placed again after its release is uploaded to new files",
			steps: []uploadStep{{
				place:   []term.Image{at(1, 1, a1)},
				shown:   []bool{true},
				uploads: []wantUpload{{1, 1, 1, "1-0"}},
				images:  []wantPlacement{{ID: 1, Version: 1}},
				files:   []string{"1-0"},
			}, {
				released: []uint64{1},
				files:    []string{"1-0"},
			}, {
				place:   []term.Image{at(1, 1, a1)},
				shown:   []bool{true},
				uploads: []wantUpload{{1, 1, 2, "2-0"}},
				images:  []wantPlacement{{ID: 1, Version: 1}},
				files:   []string{"1-0", "2-0"},
			}},
		},
		{
			name:     "a new image over budget is not shown",
			maxBytes: 40,
			steps: []uploadStep{{
				place:   []term.Image{at(1, 1, a1), at(1, 1, a1)},
				shown:   []bool{true, true},
				uploads: []wantUpload{{1, 1, 1, "1-0"}},
				images:  []wantPlacement{{ID: 1, Version: 1}, {ID: 1, Version: 1}},
				files:   []string{"1-0"},
			}, {
				place:   []term.Image{at(1, 2, a2), at(2, 1, b1)},
				shown:   []bool{true, false},
				uploads: []wantUpload{{1, 2, 2, "1-1"}},
				images:  []wantPlacement{{ID: 1, Version: 2}},
				files:   []string{"1-0", "1-1"},
			}},
		},
		{
			name:     "a new version over budget is not shown",
			maxBytes: 24,
			steps: []uploadStep{{
				place:   []term.Image{at(1, 1, a1)},
				shown:   []bool{true},
				uploads: []wantUpload{{1, 1, 1, "1-0"}},
				images:  []wantPlacement{{ID: 1, Version: 1}},
				files:   []string{"1-0"},
			}, {
				ack:   1,
				place: []term.Image{at(1, 2, a2)},
				shown: []bool{false},
				files: []string{"1-0"},
			}},
		},
		{
			name: "a placement outside the frame is dropped without an upload",
			steps: []uploadStep{{
				place: []term.Image{{
					ID: 1, Version: 1, Src: a1, Width: 2, Height: 1,
					Pos: term.Coordinates{X: width, Y: 0},
				}},
				shown: []bool{true},
			}},
		},
		{
			name: "an overflowing placement outside the frame is kept",
			steps: []uploadStep{{
				place: []term.Image{{
					ID: 1, Version: 1, Src: a1, Width: 2, Height: 1,
					Pos: term.Coordinates{X: width, Y: 0}, Overflow: true,
				}},
				shown:   []bool{true},
				uploads: []wantUpload{{1, 1, 1, "1-0"}},
				images:  []wantPlacement{{ID: 1, Version: 1}},
				files:   []string{"1-0"},
			}},
		},
		{
			name: "a sub-image uploads its own pixels and a crop relative to them",
			steps: []uploadStep{{
				place: []term.Image{{
					ID: 1, Version: 1, Src: offset, Width: 2, Height: 1,
					Crop: image.Rect(2, 1, 4, 3),
				}},
				shown:   []bool{true},
				uploads: []wantUpload{{1, 1, 1, "1-0"}},
				images: []wantPlacement{{
					ID: 1, Version: 1, Crop: image.Rect(1, 0, 3, 2),
				}},
				files: []string{"1-0"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			u := imageUploader{log: slog.Default()}
			if !tt.noArena {
				u.configure(true, dir, tt.maxBytes)
			}
			defer u.close()
			for i, step := range tt.steps {
				u.ackUploads(step.ack)
				var shown []bool
				for _, img := range step.place {
					shown = append(shown, u.place(img, width, height))
				}
				var resp DrawStreamResponse
				u.finish(&resp)

				assert.Equal(t, step.shown, shown, "step %d: shown", i)
				var uploads []wantUpload
				for _, up := range resp.GetUploads() {
					uploads = append(uploads, wantUpload{
						up.GetId(), up.GetVersion(), up.GetSeq(), up.GetFile(),
					})
				}
				assert.Equal(t, step.uploads, uploads, "step %d: uploads", i)
				var images []wantPlacement
				for _, p := range resp.GetImages() {
					images = append(images, wantPlacement{
						p.GetId(), p.GetVersion(), p.GetCrop().ToModel(),
					})
				}
				assert.Equal(t, step.images, images, "step %d: images", i)
				assert.Equal(t, step.released, resp.GetReleased(), "step %d: released", i)
				assert.Equal(t, step.redraw, resp.GetRedraw(), "step %d: redraw", i)
				assert.Equal(t, step.files, arenaFiles(t, dir), "step %d: files", i)

				for _, up := range resp.GetUploads() {
					var src image.Image
					for _, img := range step.place {
						if uint64(img.ID) == up.GetId() && img.Version == up.GetVersion() {
							src = img.Src
						}
					}
					require.NotNil(t, src)
					assert.Equal(t, int32(src.Bounds().Dx()), up.GetWidth())
					assert.Equal(t, int32(src.Bounds().Dy()), up.GetHeight())
					got, err := os.ReadFile(filepath.Join(dir, up.GetFile()))
					require.NoError(t, err)
					assert.Equal(t, rgbaBytes(src), got, "step %d: %s", i, up.GetFile())
				}
			}
			u.close()
			assert.Empty(t, arenaFiles(t, dir), "close must delete every upload file")
		})
	}
}

type streamStep struct {
	place   []term.Image
	sendErr error

	shown    []bool
	streamed []wantUpload
	records  int
	images   []wantPlacement
	released []uint64
}

func TestImageUploaderStreamsPixels(t *testing.T) {
	const width, height = 10, 5
	frame := func(seed byte) *image.RGBA { return gradient(2, 2, seed) }
	a1, a2, a3 := frame(1), frame(2), frame(3)
	b1 := frame(9)
	at := func(id term.ImageID, version uint64, src image.Image) term.Image {
		return term.Image{ID: id, Version: version, Src: src, Width: 2, Height: 1}
	}
	big := gradient(maxStreamedBytes/4/512+1, 512, 5)
	offset := gradient(6, 4, 7).SubImage(image.Rect(1, 1, 4, 3))

	tests := []struct {
		name     string
		maxBytes uint64
		steps    []streamStep
	}{
		{
			name: "the first placement streams its pixels",
			steps: []streamStep{{
				place:    []term.Image{at(1, 1, a1)},
				shown:    []bool{true},
				streamed: []wantUpload{{ID: 1, Version: 1, Seq: 1}},
				records:  1,
				images:   []wantPlacement{{ID: 1, Version: 1}},
			}},
		},
		{
			name: "pixels larger than a response are split across responses",
			steps: []streamStep{{
				place:    []term.Image{at(1, 1, big), at(2, 1, b1)},
				shown:    []bool{true, true},
				streamed: []wantUpload{{ID: 1, Version: 1, Seq: 1}, {ID: 2, Version: 1, Seq: 2}},
				records:  3,
				images:   []wantPlacement{{ID: 1, Version: 1}, {ID: 2, Version: 1}},
			}},
		},
		{
			name: "the same version is not streamed again",
			steps: []streamStep{{
				place:    []term.Image{at(1, 1, a1), at(1, 1, a1)},
				shown:    []bool{true, true},
				streamed: []wantUpload{{ID: 1, Version: 1, Seq: 1}},
				records:  1,
				images:   []wantPlacement{{ID: 1, Version: 1}, {ID: 1, Version: 1}},
			}, {
				place:  []term.Image{at(1, 1, a1)},
				shown:  []bool{true},
				images: []wantPlacement{{ID: 1, Version: 1}},
			}},
		},
		{
			name: "a new version streams without waiting for an ack",
			steps: []streamStep{{
				place:    []term.Image{at(1, 1, a1)},
				shown:    []bool{true},
				streamed: []wantUpload{{ID: 1, Version: 1, Seq: 1}},
				records:  1,
				images:   []wantPlacement{{ID: 1, Version: 1}},
			}, {
				place:    []term.Image{at(1, 2, a2)},
				shown:    []bool{true},
				streamed: []wantUpload{{ID: 1, Version: 2, Seq: 2}},
				records:  1,
				images:   []wantPlacement{{ID: 1, Version: 2}},
			}, {
				place:    []term.Image{at(1, 3, a3)},
				shown:    []bool{true},
				streamed: []wantUpload{{ID: 1, Version: 3, Seq: 3}},
				records:  1,
				images:   []wantPlacement{{ID: 1, Version: 3}},
			}},
		},
		{
			name: "an image not placed is released",
			steps: []streamStep{{
				place:    []term.Image{at(1, 1, a1), at(2, 1, b1)},
				shown:    []bool{true, true},
				streamed: []wantUpload{{ID: 1, Version: 1, Seq: 1}, {ID: 2, Version: 1, Seq: 2}},
				records:  2,
				images:   []wantPlacement{{ID: 1, Version: 1}, {ID: 2, Version: 1}},
			}, {
				place:    []term.Image{at(1, 1, a1)},
				shown:    []bool{true},
				images:   []wantPlacement{{ID: 1, Version: 1}},
				released: []uint64{2},
			}},
		},
		{
			name:     "a new image over budget is not shown",
			maxBytes: 40,
			steps: []streamStep{{
				place:    []term.Image{at(1, 1, a1)},
				shown:    []bool{true},
				streamed: []wantUpload{{ID: 1, Version: 1, Seq: 1}},
				records:  1,
				images:   []wantPlacement{{ID: 1, Version: 1}},
			}, {
				place:    []term.Image{at(1, 2, a2), at(2, 1, b1)},
				shown:    []bool{true, false},
				streamed: []wantUpload{{ID: 1, Version: 2, Seq: 2}},
				records:  1,
				images:   []wantPlacement{{ID: 1, Version: 2}},
			}},
		},
		{
			name: "a sub-image streams its own pixels",
			steps: []streamStep{{
				place:    []term.Image{at(1, 1, offset)},
				shown:    []bool{true},
				streamed: []wantUpload{{ID: 1, Version: 1, Seq: 1}},
				records:  1,
				images:   []wantPlacement{{ID: 1, Version: 1}},
			}},
		},
		{
			name: "a placement whose pixels fail to send is not shown",
			steps: []streamStep{{
				place:   []term.Image{at(1, 1, a1)},
				sendErr: errors.New("stream closed"),
				shown:   []bool{false},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sent []*DrawStreamResponse
			var sendErr error
			u := imageUploader{log: slog.Default()}
			u.send = func(resp *DrawStreamResponse) error {
				if sendErr != nil {
					return sendErr
				}
				sent = append(sent, resp)
				return nil
			}
			u.configure(true, "", tt.maxBytes)
			defer u.close()
			for i, step := range tt.steps {
				sent, sendErr = nil, step.sendErr
				var shown []bool
				for _, img := range step.place {
					shown = append(shown, u.place(img, width, height))
				}
				var resp DrawStreamResponse
				u.finish(&resp)

				assert.Equal(t, step.shown, shown, "step %d: shown", i)
				assert.Empty(t, resp.GetUploads(), "step %d: streamed pixels precede the response", i)
				assert.False(t, resp.GetRedraw(), "step %d: streaming never waits for the host", i)
				var images []wantPlacement
				for _, p := range resp.GetImages() {
					images = append(images, wantPlacement{
						p.GetId(), p.GetVersion(), p.GetCrop().ToModel(),
					})
				}
				assert.Equal(t, step.images, images, "step %d: images", i)
				assert.Equal(t, step.released, resp.GetReleased(), "step %d: released", i)

				assert.Len(t, sent, step.records, "step %d: responses streamed", i)
				uploads, data := reassemble(t, sent)
				assert.Equal(t, step.streamed, uploads, "step %d: streamed", i)
				for j, up := range uploads {
					var src image.Image
					for _, img := range step.place {
						if uint64(img.ID) == up.ID && img.Version == up.Version {
							src = img.Src
						}
					}
					require.NotNil(t, src)
					assert.Equal(t, rgbaBytes(src), data[j], "step %d: pixels of %d", i, up.ID)
				}
			}
		})
	}
}

func reassemble(t *testing.T, sent []*DrawStreamResponse) ([]wantUpload, [][]byte) {
	t.Helper()
	var uploads []wantUpload
	var data [][]byte
	for _, resp := range sent {
		require.True(t, resp.GetMore(), "a streamed response must announce the draw's")
		require.Len(t, resp.GetUploads(), 1)
		rec := resp.GetUploads()[0]
		assert.Empty(t, rec.GetFile())
		assert.LessOrEqual(t, len(rec.GetData()), maxStreamedBytes)
		if n := len(uploads); n > 0 && uploads[n-1].Seq == rec.GetSeq() {
			require.EqualValues(t, len(data[n-1]), rec.GetOffset())
			data[n-1] = append(data[n-1], rec.GetData()...)
			continue
		}
		require.Zero(t, rec.GetOffset())
		uploads = append(uploads, wantUpload{ID: rec.GetId(), Version: rec.GetVersion(), Seq: rec.GetSeq()})
		data = append(data, append([]byte(nil), rec.GetData()...))
	}
	return uploads, data
}

func TestImageUploaderConfigureReleasesOnTransportChange(t *testing.T) {
	dir := t.TempDir()
	u := imageUploader{log: slog.Default(), send: func(*DrawStreamResponse) error { return nil }}
	defer u.close()
	img := term.Image{ID: 1, Version: 1, Src: gradient(2, 2, 1), Width: 2, Height: 1}

	u.configure(true, dir, 0)
	require.True(t, u.place(img, 10, 5))
	var resp DrawStreamResponse
	u.finish(&resp)
	require.Len(t, resp.GetUploads(), 1)

	u.configure(true, "", 0)
	require.True(t, u.place(img, 10, 5))
	resp = DrawStreamResponse{}
	u.finish(&resp)
	assert.Equal(t, []uint64{1}, resp.GetReleased(),
		"the host must drop pixels it read from another arena")
	assert.Empty(t, resp.GetUploads())
	assert.Equal(t, uint64(16), u.bytes, "only the streamed upload must be accounted")
	u.ackUploads(1)
	assert.Empty(t, arenaFiles(t, dir), "the arena's files must go once acked")
}

func arenaFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}
