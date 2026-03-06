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
	"github.com/unstablebuild/tcell/v3"
)

// Prompt implements a prompt / question with options component.
type Prompt struct {
	optComp      []WithAttributes
	optVirtuals  []Virtual[tui.Component]
	optMaxWidth  int
	optMaxHeight int
	message      tui.Component
	cfg          PromptConfig
	width        int
	height       int
	msgHeight    int
	makeOptionFn func(msg string, cfg PromptConfig) floatingOption
}

// floatingOption is a Floating component that also supports SetAttr.
type floatingOption interface {
	Floating
	WithAttributes
}

// promptOption wraps a floatingOption to fill its background
// via UnionAttributes before drawing content.
type promptOption struct {
	floatingOption
	width, height int
	bg            term.Attributes
}

func (o *promptOption) Resize(width, height int) {
	o.width, o.height = width, height
	o.floatingOption.Resize(width, height)
}

func (o *promptOption) Draw(w term.Writer) {
	for y := range o.height {
		for x := range o.width {
			w.UnionAttributes(term.Coordinates{X: x, Y: y}, o.bg)
		}
	}
	o.floatingOption.Draw(w)
}

func (o *promptOption) SetAttr(
	attr term.Attributes,
) term.Attributes {
	bg := attr.Bg
	if attr.Attrs&tcell.AttrReverse != 0 {
		bg = attr.Fg
	}
	o.bg = term.Attributes{Bg: bg}
	return o.floatingOption.SetAttr(attr)
}

// PromptConfig holds configuration for initializing a Prompt.
type PromptConfig struct {
	Message string
	Options []string
	Frame   FrameCharSet
	// AspectRatio of the floating prompt message. By default
	// DefaultAspectRatio is used. Ignored when NewMessage is set.
	AspectRatio float64
	// NewMessage, when non-nil, is called with cfg.Message to
	// build the prompt's message area. The returned Floating
	// replaces the default ResponsiveString wrapped with
	// NewAspectRatioFloatingResponsive.
	NewMessage           func(msg string) Floating
	BackgroundAttributes term.Attributes
	MinWidth             int
}

var _ Floating = (*Prompt)(nil)

// Init initializes this prompt with cfg. Note that PromptConfig.Options
// must always contain at least one option and PromptConfig.Message must
// not be empty. If one of these two rules is violated this method panics.
func (p *Prompt) Init(cfg PromptConfig) {
	hasFrame := cfg.Frame != (FrameCharSet{})
	p.init(func(msg string, cfg PromptConfig) floatingOption {
		cells := term.StringToCells(msg)
		textAttr := term.Attributes{
			Bg: p.cfg.BackgroundAttributes.Bg,
		}
		text := &stringComp{cells: cells, attr: textAttr}
		if !hasFrame {
			return NewSpan(text, SpanConfig{
				PadAutoFloating:  true,
				ContentAlignment: AlignmentCentered,
			})
		}
		// Inner span provides minimum padding around the text.
		// Outer span auto-centers the padded text when the
		// frame is wider than its natural width (uniform sizing).
		inner := NewSpan(text, SpanConfig{
			PadHorizontal:    2,
			ContentAlignment: AlignmentCentered,
		})
		outer := NewSpan(inner, SpanConfig{
			PadAutoFloating:  true,
			ContentAlignment: AlignmentCentered,
		})
		frame := NewFrame(outer)
		frame.FrameCharSet = cfg.Frame
		frame.Attributes = textAttr
		return frame
	}, cfg)
}

// NewPrompt allocates storage for a new prompt and initializes it.
// See Init for more details.
func NewPrompt(cfg PromptConfig) *Prompt {
	p := new(Prompt)
	p.Init(cfg)
	return p
}

// SetOptionAttr sets the attributes of option at index i.
func (p *Prompt) SetOptionAttr(i int, attr term.Attributes) {
	p.optComp[i].SetAttr(attr)
}

// Resize satisfies tui.Component
func (p *Prompt) Resize(width, height int) {
	p.width, p.height = width, height

	_, optHeight := p.optsDimensions()
	p.msgHeight = max(0, height-optHeight)
	p.message.Resize(width, p.msgHeight)

	n := len(p.optVirtuals)
	if n == 0 {
		return
	}

	innerGap := (width - n*p.optMaxWidth) / (n + 1)
	if innerGap < 0 {
		innerGap = 0
	}
	groupWidth := n*p.optMaxWidth + (n-1)*innerGap

	// Ensure symmetric margins: when (width − groupWidth) is
	// odd, widen one inner gap by 1 so margins split evenly.
	extraGap := 0
	if n > 1 && (width-groupWidth)%2 != 0 {
		extraGap = 1
		groupWidth++
	}
	startX := (width - groupWidth) / 2

	x := startX
	for i := range p.optVirtuals {
		p.optVirtuals[i].Move(
			term.Coordinates{X: x, Y: p.msgHeight})
		p.optVirtuals[i].Resize(p.optMaxWidth, optHeight)
		if i < n-1 {
			gap := innerGap
			if i == 0 && extraGap > 0 {
				gap++
			}
			x += p.optMaxWidth + gap
		}
	}
}

// Draw satisfies tui.Component
func (p *Prompt) Draw(w term.Writer) {
	// Fill background.
	bgCell := term.Cell{
		Attributes: p.cfg.BackgroundAttributes,
		Width:      1,
	}
	for y := 0; y < p.height; y++ {
		for x := 0; x < p.width; x++ {
			w.SetCell(term.Coordinates{X: x, Y: y}, bgCell)
		}
	}

	p.message.Draw(w)
	for i := range p.optVirtuals {
		p.optVirtuals[i].Draw(w)
	}
}

// Dimensions satisfies Floating.
func (p *Prompt) Dimensions() (int, int) {
	msgWidth, msgHeight := p.message.(Floating).Dimensions()
	optWidth, optHeight := p.optsDimensions()
	width := max(msgWidth, optWidth, p.cfg.MinWidth)

	// Ensure symmetric centering: the remainder
	// (width − n·maxW) mod (n+1) must be even so that
	// (width − groupWidth) divides evenly by 2.
	n := len(p.optComp)
	if n > 0 && (width-n*p.optMaxWidth)%(n+1)%2 != 0 {
		width++
	}

	return width, msgHeight + optHeight
}

func (p *Prompt) optsDimensions() (int, int) {
	n := len(p.optComp)
	// Minimum width: buttons + 2-cell gap on each edge
	// and between each pair of buttons.
	return n*p.optMaxWidth + 2*(n+1), p.optMaxHeight
}

// OptionAt returns the index of the option at the given coordinates,
// or -1 if no option is at that position.
func (p *Prompt) OptionAt(x, y int) int {
	for i := range p.optVirtuals {
		pos := p.optVirtuals[i].Position()
		h := p.optVirtuals[i].Height()
		if x >= pos.X && x < pos.X+p.optMaxWidth &&
			y >= pos.Y && y < pos.Y+h {
			return i
		}
	}
	return -1
}

func (p *Prompt) initOptions(cfg PromptConfig) {
	if len(cfg.Options) == 0 {
		panic("Options should be greater than zero")
	}

	n := len(cfg.Options)
	p.optComp = make([]WithAttributes, 0, n)
	p.optVirtuals = make([]Virtual[tui.Component], n)

	bg := term.Attributes{Bg: cfg.BackgroundAttributes.Bg}
	for i, opt := range cfg.Options {
		compi := p.makeOptionFn(opt, cfg)
		btn := &promptOption{
			floatingOption: compi,
			bg:             bg,
		}
		p.optComp = append(p.optComp, btn)
		p.optVirtuals[i].C = btn
	}

	// Compute uniform button dimensions.
	p.optMaxWidth = 0
	p.optMaxHeight = 0
	for _, opt := range p.optComp {
		w, h := opt.(Floating).Dimensions()
		if w > p.optMaxWidth {
			p.optMaxWidth = w
		}
		if h > p.optMaxHeight {
			p.optMaxHeight = h
		}
	}
}

func (p *Prompt) init(
	makeOptionFn func(msg string, cfg PromptConfig) floatingOption,
	cfg PromptConfig,
) {
	if cfg.Message == "" {
		panic("Message cannot be empty")
	}
	if cfg.AspectRatio == 0 {
		cfg.AspectRatio = DefaultAspectRatio
	}
	p.cfg = cfg
	p.makeOptionFn = makeOptionFn

	var messageComp tui.Component
	if cfg.NewMessage != nil {
		messageComp = cfg.NewMessage(cfg.Message)
	} else {
		container := NewContainer()
		messageResponsive := NewResponsiveString(cfg.Message, StringResponsiveConfig{
			NoSplitWords: true,
			StringConfig: StringConfig{
				PaddingVertical:      4,
				PaddingHorizontal:    4,
				Alignment:            AlignmentCentered,
				BackgroundAttributes: p.cfg.BackgroundAttributes,
				Attributes: term.Attributes{
					Bg: p.cfg.BackgroundAttributes.Bg,
				},
			},
		})
		msg := NewAspectRatioFloatingResponsive(messageResponsive, cfg.AspectRatio)
		row := container.AddRow()
		row.AddComponent(msg, MaxCols)
		messageComp = container
	}

	p.message = NewBackground(messageComp, term.Cell{
		Attributes: p.cfg.BackgroundAttributes,
	})

	p.initOptions(cfg)
}
