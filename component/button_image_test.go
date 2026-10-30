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
	"fmt"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestButtonImagesDraw(t *testing.T) {
	idle := testButtonImage("idle")
	hover := testButtonImage("hover")
	selected := testButtonImage("selected")
	pressed := testButtonImage("pressed")
	loading := ButtonImage{
		Left:   &testImage{name: "loading", loading: true},
		Middle: &testImage{name: "loading", loading: true},
		Right:  &testImage{name: "loading", loading: true},
	}
	label := image.Rect(2, 1, 6, 2)

	tests := []struct {
		name    string
		images  *ButtonImages
		label   image.Rectangle
		state   ButtonState
		cannot  bool
		clip    image.Rectangle
		want    []string
		wantRet bool
	}{
		{
			name:   "idle covers the label from end to end",
			images: &ButtonImages{Idle: idle},
			label:  label,
			state:  ButtonIdle,
			want: []string{
				"idle.middle 3,1 1x1", "idle.middle 4,1 1x1",
				"idle.left 2,1 1x1", "idle.right 5,1 1x1",
			},
			wantRet: true,
		},
		{
			name: "outset reaches above, below and outward",
			images: &ButtonImages{
				Idle:   idle,
				Outset: ButtonOutset{Top: 1, Bottom: 2, Sides: 1},
			},
			label: label,
			state: ButtonIdle,
			want: []string{
				"idle.middle 3,0 1x4", "idle.middle 4,0 1x4",
				"idle.left 1,0 2x4", "idle.right 5,0 2x4",
			},
			wantRet: true,
		},
		{
			name: "negative outset is none",
			images: &ButtonImages{
				Idle:   idle,
				Outset: ButtonOutset{Top: -1, Bottom: -2, Sides: -3},
			},
			label: label,
			state: ButtonIdle,
			want: []string{
				"idle.middle 3,1 1x1", "idle.middle 4,1 1x1",
				"idle.left 2,1 1x1", "idle.right 5,1 1x1",
			},
			wantRet: true,
		},
		{
			name:    "one cell label is covered by both ends",
			images:  &ButtonImages{Idle: idle, Outset: ButtonOutset{Sides: 1}},
			label:   image.Rect(2, 1, 3, 2),
			state:   ButtonIdle,
			want:    []string{"idle.left 1,1 2x1", "idle.right 2,1 2x1"},
			wantRet: true,
		},
		{
			name: "each state draws its own images",
			images: &ButtonImages{
				Idle: idle, Selected: selected, Hover: hover, Pressed: pressed,
			},
			label:   image.Rect(0, 0, 2, 1),
			state:   ButtonPressed,
			want:    []string{"pressed.left 0,0 1x1", "pressed.right 1,0 1x1"},
			wantRet: true,
		},
		{
			name:    "selected falls back to hover",
			images:  &ButtonImages{Idle: idle, Hover: hover},
			label:   image.Rect(0, 0, 2, 1),
			state:   ButtonSelected,
			want:    []string{"hover.left 0,0 1x1", "hover.right 1,0 1x1"},
			wantRet: true,
		},
		{
			name:    "selected falls back to idle",
			images:  &ButtonImages{Idle: idle},
			label:   image.Rect(0, 0, 2, 1),
			state:   ButtonSelected,
			want:    []string{"idle.left 0,0 1x1", "idle.right 1,0 1x1"},
			wantRet: true,
		},
		{
			name:    "pressed falls back to hover",
			images:  &ButtonImages{Idle: idle, Hover: hover, Selected: selected},
			label:   image.Rect(0, 0, 2, 1),
			state:   ButtonPressed,
			want:    []string{"hover.left 0,0 1x1", "hover.right 1,0 1x1"},
			wantRet: true,
		},
		{
			name:    "hover falls back to idle",
			images:  &ButtonImages{Idle: idle, Selected: selected},
			label:   image.Rect(0, 0, 2, 1),
			state:   ButtonHover,
			want:    []string{"idle.left 0,0 1x1", "idle.right 1,0 1x1"},
			wantRet: true,
		},
		{
			name: "nil images leave their cells",
			images: &ButtonImages{Idle: ButtonImage{
				Middle: idle.Middle,
			}},
			label:   label,
			state:   ButtonIdle,
			want:    []string{"idle.middle 3,1 1x1", "idle.middle 4,1 1x1"},
			wantRet: true,
		},
		{
			name:    "a writer that cannot draw images",
			images:  &ButtonImages{Idle: idle},
			label:   image.Rect(0, 0, 2, 1),
			state:   ButtonIdle,
			cannot:  true,
			want:    []string{"idle.left 0,0 1x1", "idle.right 1,0 1x1"},
			wantRet: false,
		},
		{
			name:    "images still loading",
			images:  &ButtonImages{Idle: loading},
			label:   label,
			state:   ButtonIdle,
			wantRet: false,
		},
		{
			name:    "no images for the state",
			images:  &ButtonImages{Hover: hover},
			label:   label,
			state:   ButtonSelected,
			want:    []string{"hover.middle 3,1 1x1", "hover.middle 4,1 1x1", "hover.left 2,1 1x1", "hover.right 5,1 1x1"},
			wantRet: true,
		},
		{
			name:    "nothing to fall back to",
			images:  &ButtonImages{Hover: hover},
			label:   label,
			state:   ButtonIdle,
			wantRet: false,
		},
		{
			name:    "nil images",
			label:   label,
			wantRet: false,
		},
		{
			name:    "empty label",
			images:  &ButtonImages{Idle: idle},
			label:   image.Rect(2, 1, 2, 2),
			wantRet: false,
		},
		{
			name: "cut to the writer's bounds",
			images: &ButtonImages{
				Idle:   idle,
				Outset: ButtonOutset{Top: 1, Sides: 1},
			},
			label: image.Rect(0, 0, 3, 1),
			state: ButtonIdle,
			clip:  image.Rect(0, 0, 3, 1),
			want: []string{
				"idle.middle 1,-1 1x2 clip (1,0)-(2,1)",
				"idle.left -1,-1 2x2 clip (0,0)-(1,1)",
				"idle.right 2,-1 2x2 clip (2,0)-(3,1)",
			},
			wantRet: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &imageRecorder{graphic: !tt.cannot}
			var w term.Writer = rec
			if !tt.clip.Empty() {
				w = &VirtualWriter{Writer: rec, Width: tt.clip.Dx(), Height: tt.clip.Dy()}
			}
			got := tt.images.Draw(w, tt.label, tt.state)
			assert.Equal(t, tt.wantRet, got)
			assert.Equal(t, tt.want, describeImages(rec.images))
			for _, img := range rec.images {
				assert.Equal(t, term.ImageLayerBelowText, img.Layer)
			}
		})
	}
}

func TestButtonImagesDrawRefusedOne(t *testing.T) {
	images := &ButtonImages{Idle: testButtonImage("idle")}
	w := &refusingWriter{refuse: images.Idle.Right.(*testImage).id}
	assert.False(t, images.Draw(w, image.Rect(0, 0, 3, 1), ButtonIdle),
		"a button is drawn without images when any one of them is refused")
}

// testButtonImage returns images named name.left, name.middle and
// name.right.
func testButtonImage(name string) ButtonImage {
	return ButtonImage{
		Left:   newTestImage(name + ".left"),
		Middle: newTestImage(name + ".middle"),
		Right:  newTestImage(name + ".right"),
	}
}

var testImageNames = map[term.ImageID]string{}

func newTestImage(name string) *testImage {
	img := &testImage{name: name, id: term.NewImageID()}
	testImageNames[img.id] = name
	return img
}

// testImage places an image over all of its cells, as an image that has
// loaded does, in the layer above the text. A loading image places
// none.
type testImage struct {
	name          string
	id            term.ImageID
	width, height int
	loading       bool
}

func (c *testImage) Resize(width, height int) { c.width, c.height = width, height }

func (c *testImage) Draw(w term.Writer) {
	if c.loading {
		return
	}
	w.DrawImage(term.Image{
		ID: c.id, Width: c.width, Height: c.height, Layer: term.ImageLayerAboveText,
	})
}

func describeImages(images []term.Image) []string {
	var ret []string
	for _, img := range images {
		s := fmt.Sprintf("%s %d,%d %dx%d",
			testImageNames[img.ID], img.Pos.X, img.Pos.Y, img.Width, img.Height)
		if !img.Clip.Empty() && img.Clip != img.Bounds() {
			s += " clip " + img.Clip.String()
		}
		ret = append(ret, s)
	}
	return ret
}

// refusingWriter draws every image but refuse.
type refusingWriter struct {
	term.NoopWriter
	refuse term.ImageID
}

func (w *refusingWriter) DrawImage(img term.Image) bool { return img.ID != w.refuse }
