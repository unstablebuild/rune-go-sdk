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

// Prompt implements a prompt / question with options component.
type Prompt struct {
	optComp       []WithAttributes
	opts          []*Virtual[tui.Component]
	message       tui.Component
	cfg           PromptConfig
	width, height int
	makeOptionFn  func(msg string, cfg PromptConfig) floatingOption
}

// floatingOption is a Floating component that also supports SetAttr.
type floatingOption interface {
	Floating
	WithAttributes
}

// PromptConfig holds configuration for initializing a Prompt.
type PromptConfig struct {
	Message string
	Options []string
	Frame   FrameCharSet
	// AspectRatio of the floating prompt message. By default
	// DefaultAspectRatio is used.
	AspectRatio          float64
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
			return text
		}
		span := NewSpan(text, SpanConfig{
			PadHorizontal:    2,
			ContentAlignment: AlignmentCentered,
		})
		frame := NewFrame(span)
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
	msgHeight := max(0, height-optHeight)
	p.message.Resize(width, msgHeight)

	// Calculate each option's natural dimensions.
	type optDim struct{ w, h int }
	dims := make([]optDim, len(p.opts))
	var totalOptWidth int
	for i, opt := range p.opts {
		w, h := opt.C.(Floating).Dimensions()
		dims[i] = optDim{w, h}
		totalOptWidth += w
	}

	gap := p.optionGap()
	if len(p.opts) > 1 {
		totalOptWidth += gap * (len(p.opts) - 1)
	}

	// Center the group of options horizontally.
	startX := max(0, (width-totalOptWidth+1)/2)

	x := startX
	for i, opt := range p.opts {
		opt.Move(term.Coordinates{X: x, Y: msgHeight})
		opt.Resize(dims[i].w, dims[i].h)
		x += dims[i].w + gap
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
	for _, opt := range p.opts {
		opt.Draw(w)
	}
}

// Dimensions satisfies Floating.
func (p *Prompt) Dimensions() (int, int) {
	msgWidth, msgHeight := p.message.(Floating).Dimensions()
	optWidth, optHeight := p.optsDimensions()
	width := max(msgWidth, optWidth, p.cfg.MinWidth)
	return width, msgHeight + optHeight
}

func (p *Prompt) optsDimensions() (totalWidth, maxHeight int) {
	for _, opt := range p.opts {
		w, h := opt.C.(Floating).Dimensions()
		totalWidth += w
		if h > maxHeight {
			maxHeight = h
		}
	}
	if len(p.opts) > 1 {
		totalWidth += p.optionGap() * (len(p.opts) - 1)
	}
	return
}

// OptionAt returns the index of the option at the given coordinates,
// or -1 if no option is at that position.
func (p *Prompt) OptionAt(x, y int) int {
	for i, opt := range p.opts {
		pos := opt.Position()
		if x >= pos.X && x < pos.X+opt.Width() &&
			y >= pos.Y && y < pos.Y+opt.Height() {
			return i
		}
	}
	return -1
}

func (p *Prompt) optionGap() int {
	return 2
}

func (p *Prompt) initOptions(cfg PromptConfig) {
	if len(cfg.Options) == 0 {
		panic("Options should be greater than zero")
	}

	// allow for Init to be used as reset
	p.optComp = make([]WithAttributes, 0, len(cfg.Options))
	p.opts = make([]*Virtual[tui.Component], 0, len(cfg.Options))

	for _, opt := range cfg.Options {
		compi := p.makeOptionFn(opt, cfg)
		p.optComp = append(p.optComp, compi)
		virt := &Virtual[tui.Component]{C: compi}
		p.opts = append(p.opts, virt)
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

	container := NewContainer()

	messageResponsive := NewResponsiveString(cfg.Message,
		StringResponsiveConfig{
			NoSplitWords: true,
			StringConfig: StringConfig{
				PaddingVertical:   4,
				PaddingHorizontal: 4,
				Alignment:         AlignmentCentered,
				BackgroundAttributes: p.cfg.BackgroundAttributes,
				Attributes: term.Attributes{
					Bg: p.cfg.BackgroundAttributes.Bg,
				},
			},
		})

	row := container.AddRow()
	row.AddComponent(
		NewAspectRatioFloatingResponsive(
			messageResponsive, cfg.AspectRatio),
		MaxCols)

	p.message = NewBackground(container, term.Cell{
		Attributes: p.cfg.BackgroundAttributes,
	})

	p.initOptions(cfg)
}
