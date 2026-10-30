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
	"image"

	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// ButtonState is what the user is doing with a button, which picks the
// images it is drawn with.
type ButtonState uint8

const (
	// ButtonIdle is a button the user is not interacting with.
	ButtonIdle ButtonState = iota
	// ButtonSelected is the button the keyboard is on.
	ButtonSelected
	// ButtonHover is a button under the mouse pointer.
	ButtonHover
	// ButtonPressed is a button the mouse holds down.
	ButtonPressed
)

// ButtonImage is the images of a button in one state: Left covers the
// first cell of its label, Right its last and Middle each cell between.
// A nil image leaves its cells as the label drew them. A label one cell
// wide is covered by both Left and Right.
type ButtonImage struct {
	Left, Middle, Right tui.Component
}

// Empty reports whether i has no image.
func (i ButtonImage) Empty() bool {
	return i.Left == nil && i.Middle == nil && i.Right == nil
}

// ButtonOutset is how many cells the images of a button reach past its
// label: every image Top rows above it and Bottom rows below it, and
// Left and Right also Sides cells outward. Negative values are 0.
type ButtonOutset struct {
	Top, Bottom, Sides int
}

func (o ButtonOutset) clamped() ButtonOutset {
	return ButtonOutset{Top: max(0, o.Top), Bottom: max(0, o.Bottom), Sides: max(0, o.Sides)}
}

// ButtonImages are the images a button is drawn over in each of its
// states. A state without images is drawn with those of another:
// ButtonSelected and ButtonPressed with ButtonHover's, and ButtonHover
// with ButtonIdle's.
type ButtonImages struct {
	Idle, Selected, Hover, Pressed ButtonImage
	Outset                         ButtonOutset
}

// Image returns the images st is drawn with.
func (b *ButtonImages) Image(st ButtonState) ButtonImage {
	for _, img := range b.fallbacks(st) {
		if !img.Empty() {
			return img
		}
	}
	return ButtonImage{}
}

func (b *ButtonImages) fallbacks(st ButtonState) []ButtonImage {
	switch st {
	case ButtonSelected:
		return []ButtonImage{b.Selected, b.Hover, b.Idle}
	case ButtonHover:
		return []ButtonImage{b.Hover, b.Idle}
	case ButtonPressed:
		return []ButtonImage{b.Pressed, b.Hover, b.Idle}
	}
	return []ButtonImage{b.Idle}
}

// Draw draws the images of st over label, the cells of w a button's
// label is drawn in, and over the cells of the outset around it. They
// go under the text of every cell they cover, and w cuts them to its
// bounds as it cuts any image. Draw reports whether w drew them: it is
// false for a nil b, when st has no images, when none was drawn, as
// while one is still loading or after it failed to, and when w refused
// one, as a writer that cannot draw images does. A caller then draws the
// button as it looks without images, over whatever images w did draw.
func (b *ButtonImages) Draw(w term.Writer, label image.Rectangle, st ButtonState) bool {
	if b == nil || label.Empty() {
		return false
	}
	img := b.Image(st)
	o := b.Outset.clamped()
	bw := buttonImageWriter{Writer: w}
	draw := func(c tui.Component, area image.Rectangle) {
		if c == nil {
			return
		}
		c.Resize(area.Dx(), area.Dy())
		c.Draw(&VirtualWriter{
			Writer: &bw,
			Offset: term.Coordinates{X: area.Min.X, Y: area.Min.Y},
			Width:  area.Dx(),
			Height: area.Dy(),
		})
	}
	top, bottom := label.Min.Y-o.Top, label.Max.Y+o.Bottom
	first, last := label.Min.X, label.Max.X-1
	for x := first + 1; x < last; x++ {
		draw(img.Middle, image.Rect(x, top, x+1, bottom))
	}
	draw(img.Left, image.Rect(first-o.Sides, top, first+1, bottom))
	draw(img.Right, image.Rect(last, top, last+1+o.Sides, bottom))
	return bw.placed && !bw.refused
}

// buttonImageWriter draws the images of a button under the text of the
// cells they cover, and records whether its Writer drew them.
type buttonImageWriter struct {
	term.Writer
	placed, refused bool
}

// DrawImage satisfies term.Writer.
func (w *buttonImageWriter) DrawImage(img term.Image) bool {
	img.Layer = term.ImageLayerBelowText
	ok := w.Writer.DrawImage(img)
	w.placed = true
	w.refused = w.refused || !ok
	return ok
}
