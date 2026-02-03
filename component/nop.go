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
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Nop returns a Component that does nothing.
func Nop() tui.Component { return nop{} }

// NopResponsive returns a Responsive that does nothing.
func NopResponsive() Responsive { return nopResponsive{} }

// NopScrollable returns a Scrollable that does nothing.
func NopScrollable() Scrollable { return nopScrollable{} }

// NopWithAttributes returns a WithAttributes that does nothing.
func NopWithAttributes() WithAttributes { return nopWithAttributes{} }

// NopFloating returns a Floating that does nothing.
func NopFloating() Floating { return nopFloating{} }

// NopWithAttributesResponsive returns a WithAttributesResponsive that does nothing.
func NopWithAttributesResponsive() WithAttributesResponsive { return nopWithAttributesResponsive{} }

// NopFloatingResponsive returns a FloatingResponsive that does nothing.
func NopFloatingResponsive() FloatingResponsive { return nopFloatingResponsive{} }

// NopScrollableWithAttributes returns a ScrollableWithAttributes that does nothing.
func NopScrollableWithAttributes() ScrollableWithAttributes { return nopScrollableWithAttributes{} }

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

// NopResponsiveFromComponent wraps a Component to provide a Responsive that does nothing.
func NopResponsiveFromComponent(c tui.Component) Responsive {
	return wrapComponentAsResponsive{c, nopResponsiveMixin{}}
}

// NopScrollableFromComponent wraps a Component to provide a Scrollable that does nothing.
func NopScrollableFromComponent(c tui.Component) Scrollable {
	return wrapComponentAsScrollable{c, nopScrollableMixin{}}
}

// NopWithAttributesFromComponent wraps a Component to provide a WithAttributes that does nothing.
func NopWithAttributesFromComponent(c tui.Component) WithAttributes {
	return wrapComponentAsWithAttributes{c, nopWithAttributesMixin{}}
}

// NopFloatingFromComponent wraps a Component to provide a Floating that does nothing.
func NopFloatingFromComponent(c tui.Component) Floating {
	return wrapComponentAsFloating{c, nopFloatingMixin{}}
}

// NopWithAttributesResponsiveFromComponent wraps a Component to provide a WithAttributesResponsive that does nothing.
func NopWithAttributesResponsiveFromComponent(c tui.Component) WithAttributesResponsive {
	return wrapComponentAsWithAttributesResponsive{c, nopWithAttributesMixin{}, nopResponsiveMixin{}}
}

// NopFloatingResponsiveFromComponent wraps a Component to provide a FloatingResponsive that does nothing.
func NopFloatingResponsiveFromComponent(c tui.Component) FloatingResponsive {
	return wrapComponentAsFloatingResponsive{c, nopFloatingMixin{}, nopResponsiveMixin{}}
}

// NopScrollableWithAttributesFromComponent wraps a Component to provide a ScrollableWithAttributes that does nothing.
func NopScrollableWithAttributesFromComponent(c tui.Component) ScrollableWithAttributes {
	return wrapComponentAsScrollableWithAttributes{c, nopScrollableMixin{}, nopWithAttributesMixin{}}
}

// NopScrollableFloatingFromComponent wraps a Component to provide a ScrollableFloating that does nothing.
func NopScrollableFloatingFromComponent(c tui.Component) ScrollableFloating {
	return wrapComponentAsScrollableFloating{c, nopScrollableMixin{}, nopFloatingMixin{}}
}

// NopWithAttributesFloatingFromComponent wraps a Component to provide a WithAttributesFloating that does nothing.
func NopWithAttributesFloatingFromComponent(c tui.Component) WithAttributesFloating {
	return wrapComponentAsWithAttributesFloating{c, nopWithAttributesMixin{}, nopFloatingMixin{}}
}

// NopWithAttributesResponsiveFloatingFromComponent wraps a Component to provide a WithAttributesResponsiveFloating that does nothing.
func NopWithAttributesResponsiveFloatingFromComponent(c tui.Component) WithAttributesResponsiveFloating {
	return wrapComponentAsWithAttributesResponsiveFloating{c, nopWithAttributesMixin{}, nopResponsiveMixin{}, nopFloatingMixin{}}
}

// NopScrollableFloatingWithAttributesFromComponent wraps a Component to provide a ScrollableFloatingWithAttributes that does nothing.
func NopScrollableFloatingWithAttributesFromComponent(c tui.Component) ScrollableFloatingWithAttributes {
	return wrapComponentAsScrollableFloatingWithAttributes{c, nopScrollableMixin{}, nopFloatingMixin{}, nopWithAttributesMixin{}}
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

type nop struct{ nopComponent }

type nopResponsive struct {
	nopComponent
}

type nopScrollable struct {
	nopComponent
	nopScrollableMixin
}

type nopWithAttributes struct {
	nopComponent
	nopWithAttributesMixin
}

type nopFloating struct {
	nopComponent
	nopFloatingMixin
}

type nopWithAttributesResponsive struct {
	nopComponent
	nopWithAttributesMixin
	nopResponsiveMixin
}

type nopFloatingResponsive struct {
	nopComponent
	nopFloatingMixin
	nopResponsiveMixin
}

type nopScrollableWithAttributes struct {
	nopComponent
	nopScrollableMixin
	nopWithAttributesMixin
}

type nopScrollableFloating struct {
	nopComponent
	nopScrollableMixin
	nopFloatingMixin
}

type nopWithAttributesFloating struct {
	nopComponent
	nopWithAttributesMixin
	nopFloatingMixin
}

type nopWithAttributesResponsiveFloating struct {
	nopComponent
	nopWithAttributesMixin
	nopResponsiveMixin
	nopFloatingMixin
}

type nopScrollableFloatingWithAttributes struct {
	nopComponent
	nopScrollableMixin
	nopFloatingMixin
	nopWithAttributesMixin
}

type wrapComponentAsResponsive struct {
	tui.Component
	nopResponsiveMixin
}

type wrapComponentAsScrollable struct {
	tui.Component
	nopScrollableMixin
}

type wrapComponentAsWithAttributes struct {
	tui.Component
	nopWithAttributesMixin
}

type wrapComponentAsFloating struct {
	tui.Component
	nopFloatingMixin
}

type wrapComponentAsWithAttributesResponsive struct {
	tui.Component
	nopWithAttributesMixin
	nopResponsiveMixin
}

type wrapComponentAsFloatingResponsive struct {
	tui.Component
	nopFloatingMixin
	nopResponsiveMixin
}

type wrapComponentAsScrollableWithAttributes struct {
	tui.Component
	nopScrollableMixin
	nopWithAttributesMixin
}

type wrapComponentAsScrollableFloating struct {
	tui.Component
	nopScrollableMixin
	nopFloatingMixin
}

type wrapComponentAsWithAttributesFloating struct {
	tui.Component
	nopWithAttributesMixin
	nopFloatingMixin
}

type wrapComponentAsWithAttributesResponsiveFloating struct {
	tui.Component
	nopWithAttributesMixin
	nopResponsiveMixin
	nopFloatingMixin
}

type wrapComponentAsScrollableFloatingWithAttributes struct {
	tui.Component
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
