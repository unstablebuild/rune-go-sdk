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

package component

import (
	"context"
	"image"
	"testing"

	"github.com/unstablebuild/rune-go-sdk/component/comptest"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestIntegrationString(t *testing.T) {
	width, height := 8, 4
	str := "AAAAAAAAAAAA\nBBBBBBBBBBBB\nCCCCCCCCCCCC\nDDDDDDDDDDDD"
	virtualScroll := Virtual[String]{C: NewString(str)}
	virtualScroll.Resize(width, height)

	w := term.NewStringWriter(12, height)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
AAAAAAAA    
BBBBBBBB    
CCCCCCCC    
DDDDDDDD    `,
		},
		{
			Action: func() {
				virtualScroll.Resize(8, 3)
				virtualScroll.Move(term.Coordinates{X: 1, Y: 1})
			}, Expected: `
            
 AAAAAAAA   
 BBBBBBBB   
 CCCCCCCC   `,
		},
		{
			Action: func() {
				virtualScroll.Resize(8, 4)
				virtualScroll.Move(term.Coordinates{X: 3, Y: 0})
			}, Expected: `
   AAAAAAAA 
   BBBBBBBB 
   CCCCCCCC 
   DDDDDDDD `,
		},
	}

	comptest.TestComponent(t, &virtualScroll, w, tests)
}

func TestVirtualDraw(t *testing.T) {
	width, height := 4, 4
	v := Virtual[*TestComponent]{C: &TestComponent{Ch: '$'}}
	v.Resize(width, height)

	w := term.NewStringWriter(width, height)

	tests := []comptest.TestCase{
		{
			Action: nil, Expected: `
$$$$
$$$$
$$$$
$$$$`,
		},
		{
			Action: func() {
				v.Resize(3, 3)
				v.Move(term.Coordinates{X: 1, Y: 1})
			}, Expected: `
    
 $$$
 $$$
 $$$`,
		},
	}
	comptest.TestComponent(t, &v, w, tests)
}

// imageRecorder records the placements forwarded to it.
type imageRecorder struct {
	term.NoopWriter
	images  []term.Image
	graphic bool
}

func (w *imageRecorder) DrawImage(img term.Image) bool {
	w.images = append(w.images, img)
	return w.graphic
}

func (w *imageRecorder) Context() context.Context { return context.Background() }

func TestVirtualWriterDrawImage(t *testing.T) {
	tests := []struct {
		name          string
		offset        term.Coordinates
		img           term.Image
		graphic       bool
		expectedOK    bool
		expectedPos   term.Coordinates
		expectedVis   image.Rectangle
		expectDropped bool
	}{
		{
			name:        "translates by offset",
			offset:      term.Coordinates{X: 3, Y: 2},
			img:         term.Image{Width: 2, Height: 2},
			graphic:     true,
			expectedOK:  true,
			expectedPos: term.Coordinates{X: 3, Y: 2},
			expectedVis: image.Rect(3, 2, 5, 4),
		},
		{
			name:        "clips to viewport before translating",
			offset:      term.Coordinates{X: 10, Y: 10},
			img:         term.Image{Pos: term.Coordinates{X: 2, Y: 2}, Width: 8, Height: 8},
			graphic:     true,
			expectedOK:  true,
			expectedPos: term.Coordinates{X: 12, Y: 12},
			expectedVis: image.Rect(12, 12, 14, 14),
		},
		{
			name:          "outside the viewport is dropped",
			offset:        term.Coordinates{X: 5, Y: 5},
			img:           term.Image{Pos: term.Coordinates{X: 9, Y: 9}, Width: 2, Height: 2},
			graphic:       true,
			expectedOK:    true,
			expectDropped: true,
		},
		{
			name:        "forwards false from wrapped writer",
			img:         term.Image{Width: 2, Height: 2},
			graphic:     false,
			expectedOK:  false,
			expectedPos: term.Coordinates{},
			expectedVis: image.Rect(0, 0, 2, 2),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := &imageRecorder{graphic: test.graphic}
			w := &VirtualWriter{
				Writer: rec,
				Offset: test.offset,
				Width:  4, Height: 4,
			}

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
			got := rec.images[0]
			if got.Pos != test.expectedPos {
				t.Fatalf("Pos = %v; expected %v", got.Pos, test.expectedPos)
			}
			if vis := got.Visible(); vis != test.expectedVis {
				t.Fatalf("Visible() = %v; expected %v", vis, test.expectedVis)
			}
		})
	}
}
