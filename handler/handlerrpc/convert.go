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
	"image"

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/term/termrpc"
)

// NewImagePlacement returns img without its Src, so img.Crop must be
// relative to the top-left of the uploaded pixels.
func NewImagePlacement(img term.Image) *ImagePlacement {
	p := &ImagePlacement{
		Id:                           uint64(img.ID),
		Version:                      img.Version,
		Width:                        int32(img.Width),
		Height:                       int32(img.Height),
		Fit:                          ImageFit(img.Fit),
		Layer:                        ImageLayer(img.Layer),
		Overflow:                     img.Overflow,
		VerticalRenderOffset:         img.VerticalRenderOffset,
		NegativeVerticalRenderOffset: img.NegativeVerticalRenderOffset,
	}
	if img.Crop != (image.Rectangle{}) {
		p.Crop = termrpc.NewRect(img.Crop)
	}
	if img.Pos != (term.Coordinates{}) {
		p.Pos = new(termrpc.Coordinates)
		p.Pos.FromModel(img.Pos)
	}
	if img.Offset != (image.Point{}) {
		p.Offset = termrpc.NewPoint(img.Offset)
	}
	if img.RasterOffset != (image.Point{}) {
		p.RasterOffset = termrpc.NewPoint(img.RasterOffset)
	}
	if img.Clip != (image.Rectangle{}) {
		p.Clip = termrpc.NewRect(img.Clip)
	}
	return p
}

// ToModel returns the placement as a term.Image with a nil Src.
func (p *ImagePlacement) ToModel() term.Image {
	return term.Image{
		ID:                           term.ImageID(p.GetId()),
		Version:                      p.GetVersion(),
		Crop:                         p.GetCrop().ToModel(),
		Pos:                          p.GetPos().ToModel(),
		Offset:                       p.GetOffset().ToModel(),
		RasterOffset:                 p.GetRasterOffset().ToModel(),
		Width:                        int(p.GetWidth()),
		Height:                       int(p.GetHeight()),
		Fit:                          term.ImageFit(p.GetFit()),
		Layer:                        term.ImageLayer(p.GetLayer()),
		Clip:                         p.GetClip().ToModel(),
		Overflow:                     p.GetOverflow(),
		VerticalRenderOffset:         p.GetVerticalRenderOffset(),
		NegativeVerticalRenderOffset: p.GetNegativeVerticalRenderOffset(),
	}
}
