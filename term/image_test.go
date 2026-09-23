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

package term

import (
	"context"
	"image"
	"sync"
	"testing"
)

func TestImageBounds(t *testing.T) {
	tests := []struct {
		name     string
		img      Image
		expected image.Rectangle
	}{
		{
			name:     "origin",
			img:      Image{Width: 4, Height: 2},
			expected: image.Rect(0, 0, 4, 2),
		},
		{
			name:     "offset",
			img:      Image{Pos: Coordinates{X: 3, Y: 5}, Width: 4, Height: 2},
			expected: image.Rect(3, 5, 7, 7),
		},
		{
			name:     "clip is ignored",
			img:      Image{Width: 4, Height: 2, Clip: image.Rect(0, 0, 1, 1)},
			expected: image.Rect(0, 0, 4, 2),
		},
		{
			name:     "empty",
			img:      Image{Pos: Coordinates{X: 2, Y: 2}},
			expected: image.Rect(2, 2, 2, 2),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.img.Bounds(); got != test.expected {
				t.Fatalf("Bounds() = %v; expected %v", got, test.expected)
			}
		})
	}
}

func TestImageVisible(t *testing.T) {
	tests := []struct {
		name     string
		img      Image
		expected image.Rectangle
	}{
		{
			name:     "zero clip is unclipped",
			img:      Image{Pos: Coordinates{X: 1, Y: 1}, Width: 4, Height: 4},
			expected: image.Rect(1, 1, 5, 5),
		},
		{
			name: "partial overlap",
			img: Image{
				Width: 4, Height: 4,
				Clip: image.Rect(2, 2, 10, 10),
			},
			expected: image.Rect(2, 2, 4, 4),
		},
		{
			name: "clip larger than bounds",
			img: Image{
				Width: 2, Height: 2,
				Clip: image.Rect(-5, -5, 20, 20),
			},
			expected: image.Rect(0, 0, 2, 2),
		},
		{
			name: "disjoint clip",
			img: Image{
				Width: 2, Height: 2,
				Clip: image.Rect(8, 8, 10, 10),
			},
			expected: image.Rectangle{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.img.Visible()
			if test.expected.Empty() {
				if !got.Empty() {
					t.Fatalf("Visible() = %v; expected empty", got)
				}
				return
			}
			if got != test.expected {
				t.Fatalf("Visible() = %v; expected %v", got, test.expected)
			}
		})
	}
}

func TestImageClipped(t *testing.T) {
	tests := []struct {
		name        string
		img         Image
		clip        image.Rectangle
		expectedOK  bool
		expectedVis image.Rectangle
	}{
		{
			name:        "first clip",
			img:         Image{Width: 4, Height: 4},
			clip:        image.Rect(0, 0, 2, 4),
			expectedOK:  true,
			expectedVis: image.Rect(0, 0, 2, 4),
		},
		{
			name:        "narrows existing clip",
			img:         Image{Width: 8, Height: 8, Clip: image.Rect(0, 0, 4, 8)},
			clip:        image.Rect(2, 0, 8, 3),
			expectedOK:  true,
			expectedVis: image.Rect(2, 0, 4, 3),
		},
		{
			name:       "disjoint from bounds",
			img:        Image{Width: 2, Height: 2},
			clip:       image.Rect(4, 4, 6, 6),
			expectedOK: false,
		},
		{
			name:       "disjoint from existing clip",
			img:        Image{Width: 8, Height: 8, Clip: image.Rect(0, 0, 2, 2)},
			clip:       image.Rect(4, 4, 6, 6),
			expectedOK: false,
		},
		{
			name:       "empty placement",
			img:        Image{Width: 0, Height: 0},
			clip:       image.Rect(0, 0, 4, 4),
			expectedOK: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := test.img.Clipped(test.clip)
			if ok != test.expectedOK {
				t.Fatalf("Clipped() ok = %v; expected %v", ok, test.expectedOK)
			}
			if !ok {
				return
			}
			if vis := got.Visible(); vis != test.expectedVis {
				t.Fatalf("Visible() = %v; expected %v", vis, test.expectedVis)
			}
		})
	}
}

func TestImageTranslated(t *testing.T) {
	tests := []struct {
		name         string
		img          Image
		delta        Coordinates
		expectedPos  Coordinates
		expectedClip image.Rectangle
	}{
		{
			name:         "zero clip stays zero",
			img:          Image{Width: 2, Height: 2},
			delta:        Coordinates{X: 3, Y: 4},
			expectedPos:  Coordinates{X: 3, Y: 4},
			expectedClip: image.Rectangle{},
		},
		{
			name:         "clip moves with pos",
			img:          Image{Width: 2, Height: 2, Clip: image.Rect(0, 0, 1, 1)},
			delta:        Coordinates{X: 2, Y: -1},
			expectedPos:  Coordinates{X: 2, Y: -1},
			expectedClip: image.Rect(2, -1, 3, 0),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.img.Translated(test.delta)
			if got.Pos != test.expectedPos {
				t.Fatalf("Pos = %v; expected %v", got.Pos, test.expectedPos)
			}
			if got.Clip != test.expectedClip {
				t.Fatalf("Clip = %v; expected %v", got.Clip, test.expectedClip)
			}
		})
	}
}

func TestNewImageIDUnique(t *testing.T) {
	const goroutines, perGoroutine = 8, 256

	var (
		mu  sync.Mutex
		ids = make(map[ImageID]struct{}, goroutines*perGoroutine)
		wg  sync.WaitGroup
	)
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local := make([]ImageID, 0, perGoroutine)
			for range perGoroutine {
				local = append(local, NewImageID())
			}
			mu.Lock()
			defer mu.Unlock()
			for _, id := range local {
				ids[id] = struct{}{}
			}
		}()
	}
	wg.Wait()

	if len(ids) != goroutines*perGoroutine {
		t.Fatalf("got %d unique ids; expected %d",
			len(ids), goroutines*perGoroutine)
	}
	if _, ok := ids[0]; ok {
		t.Fatal("NewImageID returned the zero id")
	}
}

// imageRecorder records the placements forwarded to it.
type imageRecorder struct {
	NoopWriter
	images  []Image
	graphic bool
}

func (w *imageRecorder) DrawImage(img Image) bool {
	w.images = append(w.images, img)
	return w.graphic
}

func (w *imageRecorder) Context() context.Context { return context.Background() }

func TestBoundsCheckWriterDrawImage(t *testing.T) {
	tests := []struct {
		name          string
		img           Image
		graphic       bool
		expectedOK    bool
		expectedVis   image.Rectangle
		expectDropped bool
	}{
		{
			name:        "inside bounds forwards unclipped",
			img:         Image{Width: 2, Height: 2},
			graphic:     true,
			expectedOK:  true,
			expectedVis: image.Rect(0, 0, 2, 2),
		},
		{
			name:        "overflow is clipped to bounds",
			img:         Image{Pos: Coordinates{X: 6, Y: 3}, Width: 8, Height: 8},
			graphic:     true,
			expectedOK:  true,
			expectedVis: image.Rect(6, 3, 8, 4),
		},
		{
			name:          "fully outside is dropped",
			img:           Image{Pos: Coordinates{X: 20, Y: 20}, Width: 2, Height: 2},
			graphic:       true,
			expectedOK:    true,
			expectDropped: true,
		},
		{
			name:        "forwards false from wrapped writer",
			img:         Image{Width: 2, Height: 2},
			graphic:     false,
			expectedOK:  false,
			expectedVis: image.Rect(0, 0, 2, 2),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := &imageRecorder{graphic: test.graphic}
			w := BoundsCheckWriter(8, 4, rec)

			if got := w.DrawImage(test.img); got != test.expectedOK {
				t.Fatalf("DrawImage() = %v; expected %v", got, test.expectedOK)
			}
			if test.expectDropped {
				if len(rec.images) != 0 {
					t.Fatalf("forwarded %d placements; expected none", len(rec.images))
				}
				return
			}
			if len(rec.images) != 1 {
				t.Fatalf("forwarded %d placements; expected 1", len(rec.images))
			}
			if vis := rec.images[0].Visible(); vis != test.expectedVis {
				t.Fatalf("Visible() = %v; expected %v", vis, test.expectedVis)
			}
		})
	}
}
