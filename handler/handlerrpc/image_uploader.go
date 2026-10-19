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
	"fmt"
	"image"
	"image/draw"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/unstablebuild/rune-go-sdk/term"
)

var (
	errUploadBusy       = errors.New("both upload files are waiting for the host")
	errUploadOverBudget = errors.New("upload exceeds the host's byte budget")
)

// maxStreamedBytes stays well under the 4 MiB message limit gRPC applies by default.
const maxStreamedBytes = 1 << 20

const defaultMaxBytes = 64 << 20

type imageUploader struct {
	log      *slog.Logger
	send     func(*DrawStreamResponse) error
	on       bool
	dir      string
	maxBytes uint64
	ack, seq uint64
	// lastName is never reused, so a new image cannot overwrite the files of a
	// released one the host may still be reading.
	lastName uint64
	bytes    uint64
	images   map[term.ImageID]*uploadedImage
	garbage  []uploadFile
	scratch  []byte
	warned   bool

	uploads    []*ImageUpload
	placements []*ImagePlacement
	released   []uint64
	redraw     bool
}

type uploadedImage struct {
	name          uint64
	version       uint64
	width, height int
	uploaded      bool
	files         [2]uploadFile
	placed        bool
}

type uploadFile struct {
	f     *os.File
	path  string
	seq   uint64
	bytes uint64
}

func (u *imageUploader) configure(on bool, dir string, maxBytes uint64) {
	if maxBytes == 0 {
		maxBytes = defaultMaxBytes
	}
	u.maxBytes = maxBytes
	if on == u.on && dir == u.dir {
		return
	}
	for id, img := range u.images {
		u.release(id, img)
	}
	u.on, u.dir = on, dir
}

func (u *imageUploader) ackUploads(seq uint64) {
	seq = min(seq, u.seq)
	if seq <= u.ack {
		return
	}
	u.ack = seq
	live := u.garbage[:0]
	for _, f := range u.garbage {
		if f.seq <= seq {
			u.remove(f.path)
			continue
		}
		live = append(live, f)
	}
	clear(u.garbage[len(live):])
	u.garbage = live
}

func (u *imageUploader) place(img term.Image, width, height int) bool {
	if !u.on {
		return false
	}
	if !img.Overflow {
		var visible bool
		if img, visible = img.Clipped(image.Rect(0, 0, width, height)); !visible {
			return true
		}
	}
	if img.Src == nil {
		return true
	}
	bounds := img.Src.Bounds()
	if bounds.Empty() {
		return true
	}
	state := u.images[img.ID]
	if state == nil {
		if u.images == nil {
			u.images = make(map[term.ImageID]*uploadedImage)
		}
		u.lastName++
		state = &uploadedImage{name: u.lastName}
		u.images[img.ID] = state
	}
	state.placed = true
	if !state.holds(img.Version, bounds.Dx(), bounds.Dy()) {
		err := u.upload(img.ID, state, img.Version, img.Src)
		switch {
		case err == nil:
		case errors.Is(err, errUploadBusy):
			img.Version = state.version
			u.redraw = true
		case errors.Is(err, errUploadOverBudget):
			u.log.Debug("drop image placement", "id", img.ID, "error", err)
			return false
		default:
			if !u.warned {
				u.warned = true
				u.log.Warn("drop image placement", "id", img.ID, "error", err)
			}
			return false
		}
	}
	if img.Crop.Empty() {
		img.Crop = image.Rectangle{}
	} else {
		img.Crop = img.Crop.Sub(bounds.Min)
	}
	u.placements = append(u.placements, NewImagePlacement(img))
	return true
}

func (u *imageUploader) upload(
	id term.ImageID, state *uploadedImage, version uint64, src image.Image,
) error {
	file := &state.files[state.older()]
	if u.dir != "" && file.seq > u.ack {
		return errUploadBusy
	}
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	size := 4 * uint64(width) * uint64(height)
	if u.bytes-file.bytes+size > u.maxBytes {
		return errUploadOverBudget
	}
	// A failed upload still consumes its seq, so that the host never takes the
	// records of the next one for the rest of it.
	u.seq++
	up := &ImageUpload{
		Id:      uint64(id),
		Version: version,
		Seq:     u.seq,
		Width:   int32(width),
		Height:  int32(height),
	}
	pix := u.pixels(src, bounds)
	if u.dir == "" {
		if err := u.stream(up, pix); err != nil {
			return err
		}
	} else {
		if err := u.write(state, file, pix); err != nil {
			return err
		}
		up.File = filepath.Base(file.path)
		u.uploads = append(u.uploads, up)
	}
	u.bytes += size - file.bytes
	file.seq, file.bytes = u.seq, size
	state.version, state.width, state.height = version, width, height
	state.uploaded = true
	return nil
}

func (u *imageUploader) write(state *uploadedImage, file *uploadFile, pix []byte) error {
	if file.f == nil {
		slot := 0
		if file == &state.files[1] {
			slot = 1
		}
		path := filepath.Join(u.dir, fmt.Sprintf("%d-%d", state.name, slot))
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			return fmt.Errorf("create upload file: %w", err)
		}
		file.f, file.path = f, path
	}
	if _, err := file.f.WriteAt(pix, 0); err != nil {
		u.discard(file)
		return fmt.Errorf("write upload file: %w", err)
	}
	if uint64(len(pix)) < file.bytes {
		if err := file.f.Truncate(int64(len(pix))); err != nil {
			u.discard(file)
			return fmt.Errorf("truncate upload file: %w", err)
		}
	}
	return nil
}

// The records alias pix, which send has encoded by the time it returns.
func (u *imageUploader) stream(up *ImageUpload, pix []byte) error {
	if u.send == nil {
		return errors.New("no stream to send pixels on")
	}
	for off := 0; off < len(pix); off += maxStreamedBytes {
		end := min(off+maxStreamedBytes, len(pix))
		err := u.send(&DrawStreamResponse{More: true, Uploads: []*ImageUpload{{
			Id:      up.GetId(),
			Version: up.GetVersion(),
			Seq:     up.GetSeq(),
			Width:   up.GetWidth(),
			Height:  up.GetHeight(),
			Data:    pix[off:end],
			Offset:  uint64(off),
		}}})
		if err != nil {
			return fmt.Errorf("stream upload: %w", err)
		}
	}
	return nil
}

func (u *imageUploader) pixels(src image.Image, bounds image.Rectangle) []byte {
	width, height := bounds.Dx(), bounds.Dy()
	size := 4 * width * height
	if rgba, ok := src.(*image.RGBA); ok && rgba.Stride == 4*width {
		return rgba.Pix[:size]
	}
	if cap(u.scratch) < size {
		u.scratch = make([]byte, size)
	}
	dst := &image.RGBA{
		Pix:    u.scratch[:size],
		Stride: 4 * width,
		Rect:   image.Rect(0, 0, width, height),
	}
	draw.Draw(dst, dst.Rect, src, bounds.Min, draw.Src)
	return dst.Pix
}

func (u *imageUploader) finish(resp *DrawStreamResponse) {
	for id, img := range u.images {
		if img.placed && img.uploaded {
			img.placed = false
			continue
		}
		u.release(id, img)
	}
	resp.Uploads = u.uploads
	resp.Images = u.placements
	resp.Released = u.released
	resp.Redraw = u.redraw
	u.uploads, u.placements, u.released, u.redraw = nil, nil, nil, false
}

func (u *imageUploader) release(id term.ImageID, img *uploadedImage) {
	for i := range img.files {
		u.discard(&img.files[i])
	}
	delete(u.images, id)
	if img.uploaded {
		u.released = append(u.released, uint64(id))
	}
}

func (u *imageUploader) discard(file *uploadFile) {
	u.bytes -= file.bytes
	if file.f != nil {
		_ = file.f.Close()
		if file.seq <= u.ack {
			u.remove(file.path)
		} else {
			u.garbage = append(u.garbage, uploadFile{path: file.path, seq: file.seq})
		}
	}
	*file = uploadFile{}
}

func (u *imageUploader) remove(path string) {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		u.log.Debug("remove upload file", "path", path, "error", err)
	}
}

func (u *imageUploader) close() {
	for _, img := range u.images {
		for _, f := range img.files {
			if f.f != nil {
				_ = f.f.Close()
				u.remove(f.path)
			}
		}
	}
	for _, f := range u.garbage {
		u.remove(f.path)
	}
	u.images, u.garbage, u.bytes = nil, nil, 0
}

func (img *uploadedImage) holds(version uint64, width, height int) bool {
	return img.uploaded && img.version == version &&
		img.width == width && img.height == height
}

// older is the only slot that can be free, as the host reads uploads in
// seq order.
func (img *uploadedImage) older() int {
	if img.files[0].seq <= img.files[1].seq {
		return 0
	}
	return 1
}
