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
	"image"
	"sync/atomic"
)

// ImageID identifies the pixel state a writer retains for a picture.
type ImageID uint64

var lastImageID atomic.Uint64

// NewImageID returns a process-unique id. IDs key a writer's retained
// pixel state, so a caller allocates one per picture and keeps it for
// the picture's lifetime.
func NewImageID() ImageID {
	return ImageID(lastImageID.Add(1))
}

// ImageFit selects how a raster is scaled into its cell rectangle.
type ImageFit uint8

const (
	// ImageFitFill stretches the raster to the cell rectangle.
	ImageFitFill ImageFit = iota
	// ImageFitContain scales the raster uniformly and centres it in the
	// cell rectangle.
	ImageFitContain
)

// Image is a raster placed over a rectangle of cells.
type Image struct {
	// Src holds the pixels. It is only read between DrawImage and the
	// writer's next Clear. An *image.RGBA with a zero Min and a tight
	// stride is the zero-copy upload path.
	Src image.Image
	// ID identifies the retained pixel state for this picture.
	ID ImageID
	// Version must change whenever Src's pixels change. Writers
	// re-upload only when it does.
	Version uint64
	// Crop is a sub-rectangle of Src in pixels. The zero value selects
	// all of Src.
	Crop image.Rectangle
	// Pos is the top-left cell of the placement.
	Pos Coordinates
	// Width and Height are the placement size in cells.
	Width, Height int
	// Fit selects how Src is scaled into the cell rectangle.
	Fit ImageFit
	// Clip is an absolute, right-exclusive cell rectangle the placement
	// is confined to. The zero value is unclipped.
	Clip image.Rectangle
}

// Bounds returns the placement's cell rectangle, ignoring Clip.
func (img Image) Bounds() image.Rectangle {
	return image.Rect(
		img.Pos.X, img.Pos.Y,
		img.Pos.X+img.Width, img.Pos.Y+img.Height,
	)
}

// Visible returns the cell rectangle the placement actually covers.
func (img Image) Visible() image.Rectangle {
	b := img.Bounds()
	if img.Clip.Empty() {
		return b
	}
	return b.Intersect(img.Clip)
}

// Clipped narrows the placement's clip rectangle to r, reporting false
// when nothing remains visible. A placement that is no longer visible is
// returned with an empty cell rectangle, because the zero Clip means
// unclipped and so cannot express an empty intersection.
func (img Image) Clipped(r image.Rectangle) (Image, bool) {
	visible := img.Visible().Intersect(r)
	if visible.Empty() {
		img.Width, img.Height = 0, 0
		img.Clip = image.Rectangle{}
		return img, false
	}
	img.Clip = visible
	return img, true
}

// Translated moves the placement and its clip rectangle by d cells.
func (img Image) Translated(d Coordinates) Image {
	img.Pos.X += d.X
	img.Pos.Y += d.Y
	if !img.Clip.Empty() {
		img.Clip = img.Clip.Add(image.Pt(d.X, d.Y))
	}
	return img
}
