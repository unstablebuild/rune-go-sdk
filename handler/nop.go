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

package handler

import (
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// NopFromComponent wraps a Component to provide a Handler that does nothing.
func NopFromComponent(c tui.Component) tui.Handler {
	return wrapComponentAsHandler{c, nopHandlerMixin{}}
}

// NopResponsiveFromComponent wraps a Component to provide a Responsive that does nothing.
func NopResponsiveFromComponent(c tui.Component) Responsive {
	return wrapComponentAsResponsive{c, nopHandlerMixin{}, nopResponsiveMixin{}}
}

// NopScrollableFromComponent wraps a Component to provide a Scrollable that does nothing.
func NopScrollableFromComponent(c tui.Component) Scrollable {
	return wrapComponentAsScrollable{c, nopHandlerMixin{}, nopScrollableMixin{}}
}

// NopWithAttributesFromComponent wraps a Component to provide a WithAttributes that does nothing.
func NopWithAttributesFromComponent(c tui.Component) WithAttributes {
	return wrapComponentAsWithAttributes{c, nopHandlerMixin{}, nopWithAttributesMixin{}}
}

// NopFloatingFromComponent wraps a Component to provide a Floating that does nothing.
func NopFloatingFromComponent(c tui.Component) Floating {
	return wrapComponentAsFloating{c, nopHandlerMixin{}, nopFloatingMixin{}}
}

// NopWithAttributesResponsiveFromComponent wraps a Component to provide a WithAttributesResponsive that does nothing.
func NopWithAttributesResponsiveFromComponent(c tui.Component) WithAttributesResponsive {
	return wrapComponentAsWithAttributesResponsive{c, nopHandlerMixin{}, nopWithAttributesMixin{}, nopResponsiveMixin{}}
}

// NopFloatingResponsiveFromComponent wraps a Component to provide a FloatingResponsive that does nothing.
func NopFloatingResponsiveFromComponent(c tui.Component) FloatingResponsive {
	return wrapComponentAsFloatingResponsive{c, nopHandlerMixin{}, nopFloatingMixin{}, nopResponsiveMixin{}}
}

// NopScrollableWithAttributesFromComponent wraps a Component to provide a ScrollableWithAttributes that does nothing.
func NopScrollableWithAttributesFromComponent(c tui.Component) ScrollableWithAttributes {
	return wrapComponentAsScrollableWithAttributes{c, nopHandlerMixin{}, nopScrollableMixin{}, nopWithAttributesMixin{}}
}

// NopScrollableFloatingFromComponent wraps a Component to provide a ScrollableFloating that does nothing.
func NopScrollableFloatingFromComponent(c tui.Component) ScrollableFloating {
	return wrapComponentAsScrollableFloating{c, nopHandlerMixin{}, nopScrollableMixin{}, nopFloatingMixin{}}
}

// NopWithAttributesFloatingFromComponent wraps a Component to provide a WithAttributesFloating that does nothing.
func NopWithAttributesFloatingFromComponent(c tui.Component) WithAttributesFloating {
	return wrapComponentAsWithAttributesFloating{c, nopHandlerMixin{}, nopWithAttributesMixin{}, nopFloatingMixin{}}
}

// NopWithAttributesResponsiveFloatingFromComponent wraps a Component to provide a WithAttributesResponsiveFloating that does nothing.
func NopWithAttributesResponsiveFloatingFromComponent(c tui.Component) WithAttributesResponsiveFloating {
	return wrapComponentAsWithAttributesResponsiveFloating{c, nopHandlerMixin{}, nopWithAttributesMixin{}, nopResponsiveMixin{}, nopFloatingMixin{}}
}

// NopScrollableFloatingWithAttributesFromComponent wraps a Component to provide a ScrollableFloatingWithAttributes that does nothing.
func NopScrollableFloatingWithAttributesFromComponent(c tui.Component) ScrollableFloatingWithAttributes {
	return wrapComponentAsScrollableFloatingWithAttributes{c, nopHandlerMixin{}, nopScrollableMixin{}, nopFloatingMixin{}, nopWithAttributesMixin{}}
}

// Nop returns a Handler that does nothing.
func Nop() tui.Handler { return nopHandler{} }

// NopResponsive returns a Responsive that does nothing.
func NopResponsive() Responsive { return nopResponsive{} }

// NopScrollable returns a Scrollable that does nothing.
func NopScrollable() Scrollable { return nopScrollable{} }

// NopWithAttributes returns a WithAttributes that does nothing.
func NopWithAttributes() WithAttributes { return nopWithAttributes{} }

// NopFloating returns a Floating that does nothing.
func NopFloating() Floating { return nopFloating{} }

// NopWithAttributesResponsive returns a WithAttributesResponsive that does nothing.
func NopWithAttributesResponsive() WithAttributesResponsive {
	return nopWithAttributesResponsive{}
}

// NopFloatingResponsive returns a FloatingResponsive that does nothing.
func NopFloatingResponsive() FloatingResponsive { return nopFloatingResponsive{} }

// NopScrollableWithAttributes returns a ScrollableWithAttributes that does nothing.
func NopScrollableWithAttributes() ScrollableWithAttributes {
	return nopScrollableWithAttributes{}
}

// NopScrollableFloating returns a ScrollableFloating that does nothing.
func NopScrollableFloating() ScrollableFloating { return nopScrollableFloating{} }

// NopWithAttributesFloating returns a WithAttributesFloating that does nothing.
func NopWithAttributesFloating() WithAttributesFloating { return nopWithAttributesFloating{} }

// NopWithAttributesResponsiveFloating returns a WithAttributesResponsiveFloating that does nothing.
func NopWithAttributesResponsiveFloating() WithAttributesResponsiveFloating {
	return nopWithAttributesResponsiveFloating{}
}

// NopScrollableFloatingWithAttributes returns a ScrollableFloatingWithAttributes that does nothing.
func NopScrollableFloatingWithAttributes() ScrollableFloatingWithAttributes {
	return nopScrollableFloatingWithAttributes{}
}

// NopResponsiveFromHandler wraps a Handler to provide a Responsive that does nothing.
func NopResponsiveFromHandler(h tui.Handler) Responsive {
	return wrapHandlerAsResponsive{h, nopResponsiveMixin{}}
}

// NopScrollableFromHandler wraps a Handler to provide a Scrollable that does nothing.
func NopScrollableFromHandler(h tui.Handler) Scrollable {
	return wrapHandlerAsScrollable{h, nopScrollableMixin{}}
}

// NopWithAttributesFromHandler wraps a Handler to provide a WithAttributes that does nothing.
func NopWithAttributesFromHandler(h tui.Handler) WithAttributes {
	return wrapHandlerAsWithAttributes{h, nopWithAttributesMixin{}}
}

// NopFloatingFromHandler wraps a Handler to provide a Floating that does nothing.
func NopFloatingFromHandler(h tui.Handler) Floating {
	return wrapHandlerAsFloating{h, nopFloatingMixin{}}
}

// NopWithAttributesResponsiveFromHandler wraps a Handler to provide a WithAttributesResponsive that does nothing.
func NopWithAttributesResponsiveFromHandler(h tui.Handler) WithAttributesResponsive {
	return wrapHandlerAsWithAttributesResponsive{h, nopWithAttributesMixin{}, nopResponsiveMixin{}}
}

// NopFloatingResponsiveFromHandler wraps a Handler to provide a FloatingResponsive that does nothing.
func NopFloatingResponsiveFromHandler(h tui.Handler) FloatingResponsive {
	return wrapHandlerAsFloatingResponsive{h, nopFloatingMixin{}, nopResponsiveMixin{}}
}

// NopScrollableWithAttributesFromHandler wraps a Handler to provide a ScrollableWithAttributes that does nothing.
func NopScrollableWithAttributesFromHandler(h tui.Handler) ScrollableWithAttributes {
	return wrapHandlerAsScrollableWithAttributes{h, nopScrollableMixin{}, nopWithAttributesMixin{}}
}

// NopScrollableFloatingFromHandler wraps a Handler to provide a ScrollableFloating that does nothing.
func NopScrollableFloatingFromHandler(h tui.Handler) ScrollableFloating {
	return wrapHandlerAsScrollableFloating{h, nopScrollableMixin{}, nopFloatingMixin{}}
}

// NopWithAttributesFloatingFromHandler wraps a Handler to provide a WithAttributesFloating that does nothing.
func NopWithAttributesFloatingFromHandler(h tui.Handler) WithAttributesFloating {
	return wrapHandlerAsWithAttributesFloating{h, nopWithAttributesMixin{}, nopFloatingMixin{}}
}

// NopWithAttributesResponsiveFloatingFromHandler wraps a Handler to provide a WithAttributesResponsiveFloating that does nothing.
func NopWithAttributesResponsiveFloatingFromHandler(h tui.Handler) WithAttributesResponsiveFloating {
	return wrapHandlerAsWithAttributesResponsiveFloating{h, nopWithAttributesMixin{}, nopResponsiveMixin{}, nopFloatingMixin{}}
}

// NopScrollableFloatingWithAttributesFromHandler wraps a Handler to provide a ScrollableFloatingWithAttributes that does nothing.
func NopScrollableFloatingWithAttributesFromHandler(h tui.Handler) ScrollableFloatingWithAttributes {
	return wrapHandlerAsScrollableFloatingWithAttributes{h, nopScrollableMixin{}, nopFloatingMixin{}, nopWithAttributesMixin{}}
}

// NopWithAttributesFromResponsive wraps a Responsive to provide a WithAttributesResponsive that does nothing.
func NopWithAttributesFromResponsive(r Responsive) WithAttributesResponsive {
	return wrapResponsiveAsWithAttributes{r, nopWithAttributesMixin{}}
}

// NopFloatingFromResponsive wraps a Responsive to provide a FloatingResponsive that does nothing.
func NopFloatingFromResponsive(r Responsive) FloatingResponsive {
	return wrapResponsiveAsFloating{r, nopFloatingMixin{}}
}

// NopWithAttributesFloatingFromResponsive wraps a Responsive to provide a WithAttributesResponsiveFloating that does nothing.
func NopWithAttributesFloatingFromResponsive(r Responsive) WithAttributesResponsiveFloating {
	return wrapResponsiveAsWithAttributesFloating{r, nopWithAttributesMixin{}, nopFloatingMixin{}}
}

// NopWithAttributesFromScrollable wraps a Scrollable to provide a ScrollableWithAttributes that does nothing.
func NopWithAttributesFromScrollable(s Scrollable) ScrollableWithAttributes {
	return wrapScrollableAsWithAttributes{s, nopWithAttributesMixin{}}
}

// NopFloatingFromScrollable wraps a Scrollable to provide a ScrollableFloating that does nothing.
func NopFloatingFromScrollable(s Scrollable) ScrollableFloating {
	return wrapScrollableAsFloating{s, nopFloatingMixin{}}
}

// NopWithAttributesFloatingFromScrollable wraps a Scrollable to provide a ScrollableFloatingWithAttributes that does nothing.
func NopWithAttributesFloatingFromScrollable(s Scrollable) ScrollableFloatingWithAttributes {
	return wrapScrollableAsWithAttributesFloating{s, nopWithAttributesMixin{}, nopFloatingMixin{}}
}

// NopResponsiveFromWithAttributes wraps a WithAttributes to provide a WithAttributesResponsive that does nothing.
func NopResponsiveFromWithAttributes(a WithAttributes) WithAttributesResponsive {
	return wrapWithAttributesAsResponsive{a, nopResponsiveMixin{}}
}

// NopScrollableFromWithAttributes wraps a WithAttributes to provide a ScrollableWithAttributes that does nothing.
func NopScrollableFromWithAttributes(a WithAttributes) ScrollableWithAttributes {
	return wrapWithAttributesAsScrollable{a, nopScrollableMixin{}}
}

// NopFloatingFromWithAttributes wraps a WithAttributes to provide a WithAttributesFloating that does nothing.
func NopFloatingFromWithAttributes(a WithAttributes) WithAttributesFloating {
	return wrapWithAttributesAsFloating{a, nopFloatingMixin{}}
}

// NopResponsiveFloatingFromWithAttributes wraps a WithAttributes to provide a WithAttributesResponsiveFloating that does nothing.
func NopResponsiveFloatingFromWithAttributes(a WithAttributes) WithAttributesResponsiveFloating {
	return wrapWithAttributesAsResponsiveFloating{a, nopResponsiveMixin{}, nopFloatingMixin{}}
}

// NopScrollableFloatingFromWithAttributes wraps a WithAttributes to provide a ScrollableFloatingWithAttributes that does nothing.
func NopScrollableFloatingFromWithAttributes(a WithAttributes) ScrollableFloatingWithAttributes {
	return wrapWithAttributesAsScrollableFloating{a, nopScrollableMixin{}, nopFloatingMixin{}}
}

// NopResponsiveFromFloating wraps a Floating to provide a FloatingResponsive that does nothing.
func NopResponsiveFromFloating(f Floating) FloatingResponsive {
	return wrapFloatingAsResponsive{f, nopResponsiveMixin{}}
}

// NopScrollableFromFloating wraps a Floating to provide a ScrollableFloating that does nothing.
func NopScrollableFromFloating(f Floating) ScrollableFloating {
	return wrapFloatingAsScrollable{f, nopScrollableMixin{}}
}

// NopWithAttributesFromFloating wraps a Floating to provide a WithAttributesFloating that does nothing.
func NopWithAttributesFromFloating(f Floating) WithAttributesFloating {
	return wrapFloatingAsWithAttributes{f, nopWithAttributesMixin{}}
}

// NopWithAttributesResponsiveFromFloating wraps a Floating to provide a WithAttributesResponsiveFloating that does nothing.
func NopWithAttributesResponsiveFromFloating(f Floating) WithAttributesResponsiveFloating {
	return wrapFloatingAsWithAttributesResponsive{f, nopWithAttributesMixin{}, nopResponsiveMixin{}}
}

// NopScrollableWithAttributesFromFloating wraps a Floating to provide a ScrollableFloatingWithAttributes that does nothing.
func NopScrollableWithAttributesFromFloating(f Floating) ScrollableFloatingWithAttributes {
	return wrapFloatingAsScrollableWithAttributes{f, nopScrollableMixin{}, nopWithAttributesMixin{}}
}

// NopFloatingFromWithAttributesResponsive wraps a WithAttributesResponsive to provide a WithAttributesResponsiveFloating that does nothing.
func NopFloatingFromWithAttributesResponsive(ar WithAttributesResponsive) WithAttributesResponsiveFloating {
	return wrapWithAttributesResponsiveAsFloating{ar, nopFloatingMixin{}}
}

// NopWithAttributesFromFloatingResponsive wraps a FloatingResponsive to provide a WithAttributesResponsiveFloating that does nothing.
func NopWithAttributesFromFloatingResponsive(fr FloatingResponsive) WithAttributesResponsiveFloating {
	return wrapFloatingResponsiveAsWithAttributes{fr, nopWithAttributesMixin{}}
}

// NopFloatingFromScrollableWithAttributes wraps a ScrollableWithAttributes to provide a ScrollableFloatingWithAttributes that does nothing.
func NopFloatingFromScrollableWithAttributes(sa ScrollableWithAttributes) ScrollableFloatingWithAttributes {
	return wrapScrollableWithAttributesAsFloating{sa, nopFloatingMixin{}}
}

// NopWithAttributesFromScrollableFloating wraps a ScrollableFloating to provide a ScrollableFloatingWithAttributes that does nothing.
func NopWithAttributesFromScrollableFloating(sf ScrollableFloating) ScrollableFloatingWithAttributes {
	return wrapScrollableFloatingAsWithAttributes{sf, nopWithAttributesMixin{}}
}

// NopResponsiveFromWithAttributesFloating wraps a WithAttributesFloating to provide a WithAttributesResponsiveFloating that does nothing.
func NopResponsiveFromWithAttributesFloating(af WithAttributesFloating) WithAttributesResponsiveFloating {
	return wrapWithAttributesFloatingAsResponsive{af, nopResponsiveMixin{}}
}

// NopScrollableFromWithAttributesFloating wraps a WithAttributesFloating to provide a ScrollableFloatingWithAttributes that does nothing.
func NopScrollableFromWithAttributesFloating(af WithAttributesFloating) ScrollableFloatingWithAttributes {
	return wrapWithAttributesFloatingAsScrollable{af, nopScrollableMixin{}}
}

type nopComponent struct{}

func (nopComponent) Resize(int, int)  {}
func (nopComponent) Draw(term.Writer) {}

type nopHandlerMixin struct{}

func (nopHandlerMixin) Handle(term.Event) (bool, bool) { return false, false }
func (nopHandlerMixin) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	return term.Coordinates{}, 0, false
}
func (nopHandlerMixin) Selection() (string, bool) { return "", false }

type nopResponsiveMixin struct{}

func (nopResponsiveMixin) Height(int) int { return 0 }

type nopScrollableMixin struct{}

func (nopScrollableMixin) SeekUp() bool       { return false }
func (nopScrollableMixin) SeekDown() bool     { return false }
func (nopScrollableMixin) SeekOffset() int    { return 0 }
func (nopScrollableMixin) MaxSeekOffset() int { return 0 }

type nopWithAttributesMixin struct{}

func (nopWithAttributesMixin) SetAttr(term.Attributes) term.Attributes { return term.Attributes{} }

type nopFloatingMixin struct{}

func (nopFloatingMixin) Dimensions() (int, int) { return 0, 0 }

type nopHandler struct {
	nopComponent
	nopHandlerMixin
}

type nopResponsive struct {
	nopComponent
	nopHandlerMixin
	nopResponsiveMixin
}

type nopScrollable struct {
	nopComponent
	nopHandlerMixin
	nopScrollableMixin
}

type nopWithAttributes struct {
	nopComponent
	nopHandlerMixin
	nopWithAttributesMixin
}

type nopFloating struct {
	nopComponent
	nopHandlerMixin
	nopFloatingMixin
}

type nopWithAttributesResponsive struct {
	nopComponent
	nopHandlerMixin
	nopWithAttributesMixin
	nopResponsiveMixin
}

type nopFloatingResponsive struct {
	nopComponent
	nopHandlerMixin
	nopFloatingMixin
	nopResponsiveMixin
}

type nopScrollableWithAttributes struct {
	nopComponent
	nopHandlerMixin
	nopScrollableMixin
	nopWithAttributesMixin
}

type nopScrollableFloating struct {
	nopComponent
	nopHandlerMixin
	nopScrollableMixin
	nopFloatingMixin
}

type nopWithAttributesFloating struct {
	nopComponent
	nopHandlerMixin
	nopWithAttributesMixin
	nopFloatingMixin
}

type nopWithAttributesResponsiveFloating struct {
	nopComponent
	nopHandlerMixin
	nopWithAttributesMixin
	nopResponsiveMixin
	nopFloatingMixin
}

type nopScrollableFloatingWithAttributes struct {
	nopComponent
	nopHandlerMixin
	nopScrollableMixin
	nopFloatingMixin
	nopWithAttributesMixin
}

type wrapHandlerAsResponsive struct {
	tui.Handler
	nopResponsiveMixin
}

type wrapHandlerAsScrollable struct {
	tui.Handler
	nopScrollableMixin
}

type wrapHandlerAsWithAttributes struct {
	tui.Handler
	nopWithAttributesMixin
}

type wrapHandlerAsFloating struct {
	tui.Handler
	nopFloatingMixin
}

type wrapHandlerAsWithAttributesResponsive struct {
	tui.Handler
	nopWithAttributesMixin
	nopResponsiveMixin
}

type wrapHandlerAsFloatingResponsive struct {
	tui.Handler
	nopFloatingMixin
	nopResponsiveMixin
}

type wrapHandlerAsScrollableWithAttributes struct {
	tui.Handler
	nopScrollableMixin
	nopWithAttributesMixin
}

type wrapHandlerAsScrollableFloating struct {
	tui.Handler
	nopScrollableMixin
	nopFloatingMixin
}

type wrapHandlerAsWithAttributesFloating struct {
	tui.Handler
	nopWithAttributesMixin
	nopFloatingMixin
}

type wrapHandlerAsWithAttributesResponsiveFloating struct {
	tui.Handler
	nopWithAttributesMixin
	nopResponsiveMixin
	nopFloatingMixin
}

type wrapHandlerAsScrollableFloatingWithAttributes struct {
	tui.Handler
	nopScrollableMixin
	nopFloatingMixin
	nopWithAttributesMixin
}

type wrapResponsiveAsWithAttributes struct {
	Responsive
	nopWithAttributesMixin
}

type wrapResponsiveAsFloating struct {
	Responsive
	nopFloatingMixin
}

type wrapResponsiveAsWithAttributesFloating struct {
	Responsive
	nopWithAttributesMixin
	nopFloatingMixin
}

type wrapScrollableAsWithAttributes struct {
	Scrollable
	nopWithAttributesMixin
}

type wrapScrollableAsFloating struct {
	Scrollable
	nopFloatingMixin
}

type wrapScrollableAsWithAttributesFloating struct {
	Scrollable
	nopWithAttributesMixin
	nopFloatingMixin
}

type wrapWithAttributesAsResponsive struct {
	WithAttributes
	nopResponsiveMixin
}

type wrapWithAttributesAsScrollable struct {
	WithAttributes
	nopScrollableMixin
}

type wrapWithAttributesAsFloating struct {
	WithAttributes
	nopFloatingMixin
}

type wrapWithAttributesAsResponsiveFloating struct {
	WithAttributes
	nopResponsiveMixin
	nopFloatingMixin
}

type wrapWithAttributesAsScrollableFloating struct {
	WithAttributes
	nopScrollableMixin
	nopFloatingMixin
}

type wrapFloatingAsResponsive struct {
	Floating
	nopResponsiveMixin
}

type wrapFloatingAsScrollable struct {
	Floating
	nopScrollableMixin
}

type wrapFloatingAsWithAttributes struct {
	Floating
	nopWithAttributesMixin
}

type wrapFloatingAsWithAttributesResponsive struct {
	Floating
	nopWithAttributesMixin
	nopResponsiveMixin
}

type wrapFloatingAsScrollableWithAttributes struct {
	Floating
	nopScrollableMixin
	nopWithAttributesMixin
}

type wrapWithAttributesResponsiveAsFloating struct {
	WithAttributesResponsive
	nopFloatingMixin
}

type wrapFloatingResponsiveAsWithAttributes struct {
	FloatingResponsive
	nopWithAttributesMixin
}

type wrapScrollableWithAttributesAsFloating struct {
	ScrollableWithAttributes
	nopFloatingMixin
}

type wrapScrollableFloatingAsWithAttributes struct {
	ScrollableFloating
	nopWithAttributesMixin
}

type wrapWithAttributesFloatingAsResponsive struct {
	WithAttributesFloating
	nopResponsiveMixin
}

type wrapWithAttributesFloatingAsScrollable struct {
	WithAttributesFloating
	nopScrollableMixin
}

type wrapComponentAsHandler struct {
	tui.Component
	nopHandlerMixin
}

type wrapComponentAsResponsive struct {
	tui.Component
	nopHandlerMixin
	nopResponsiveMixin
}

type wrapComponentAsScrollable struct {
	tui.Component
	nopHandlerMixin
	nopScrollableMixin
}

type wrapComponentAsWithAttributes struct {
	tui.Component
	nopHandlerMixin
	nopWithAttributesMixin
}

type wrapComponentAsFloating struct {
	tui.Component
	nopHandlerMixin
	nopFloatingMixin
}

type wrapComponentAsWithAttributesResponsive struct {
	tui.Component
	nopHandlerMixin
	nopWithAttributesMixin
	nopResponsiveMixin
}

type wrapComponentAsFloatingResponsive struct {
	tui.Component
	nopHandlerMixin
	nopFloatingMixin
	nopResponsiveMixin
}

type wrapComponentAsScrollableWithAttributes struct {
	tui.Component
	nopHandlerMixin
	nopScrollableMixin
	nopWithAttributesMixin
}

type wrapComponentAsScrollableFloating struct {
	tui.Component
	nopHandlerMixin
	nopScrollableMixin
	nopFloatingMixin
}

type wrapComponentAsWithAttributesFloating struct {
	tui.Component
	nopHandlerMixin
	nopWithAttributesMixin
	nopFloatingMixin
}

type wrapComponentAsWithAttributesResponsiveFloating struct {
	tui.Component
	nopHandlerMixin
	nopWithAttributesMixin
	nopResponsiveMixin
	nopFloatingMixin
}

type wrapComponentAsScrollableFloatingWithAttributes struct {
	tui.Component
	nopHandlerMixin
	nopScrollableMixin
	nopFloatingMixin
	nopWithAttributesMixin
}
