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
	optComp      []WithAttributes
	effective    tui.Component
	messageRow   *Row
	cfg          PromptConfig
	makeOptionFn func(msg string, cfg PromptConfig) responsiveWithAttributes
}

// PromptConfig holds configuration for initializing a Prompt.
type PromptConfig struct {
	Message string
	Options []string
	Frame   FrameCharSet
	// AspectRatio of the floating prompt. By default DefaultAspectRatio is used.
	AspectRatio          float64
	BackgroundAttributes term.Attributes
	MinWidth             int
}

type responsiveWithAttributes interface {
	Responsive
	WithAttributes
}

// Init initializes this prompt with cfg. Note that PromptConfig.Options must
// always contain at least one option and PromptConfig.Message must not be empty.
// If one of these two rules is violated this method panics.
func (p *Prompt) Init(cfg PromptConfig) {
	p.init(func(msg string, cfg PromptConfig) responsiveWithAttributes {
		return NewResponsiveString(msg, StringResponsiveConfig{
			NoSplitWords: true,
			StringConfig: StringConfig{
				Alignment:            AlignmentCentered,
				FrameCharSet:         cfg.Frame,
				BackgroundAttributes: p.cfg.BackgroundAttributes,
				Attributes:           term.Attributes{Bg: p.cfg.BackgroundAttributes.Bg},
				PaddingHorizontal:    2,
			},
		})
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
	p.effective.Resize(width, height)
}

// Draw satisfies tui.Component
func (p *Prompt) Draw(w term.Writer) {
	p.effective.Draw(w)
}

// Dimensions satisfies Floating.
func (p *Prompt) Dimensions() (int, int) {
	width, height := p.effective.(Floating).Dimensions()
	if width < p.cfg.MinWidth {
		width = p.cfg.MinWidth
	}
	return width, height
}

func (p *Prompt) initOptions(cfg PromptConfig, row *Row) {
	if len(cfg.Options) == 0 {
		panic("Options should be greater than zero")
	}

	// we want to divide the space evently between options
	// but longer options should take more space, so paddings
	// are not off
	var totalLength, totalColumns int
	for _, opt := range cfg.Options {
		totalLength += len(opt)
	}

	columns := make([]int, len(cfg.Options))
	for i := range cfg.Options {
		columns[i] = int(float64(MaxCols) / float64(len(cfg.Options)))
		totalColumns += columns[i]
	}

	// allow for Init to be used as reset
	p.optComp = make([]WithAttributes, 0, len(cfg.Options))

	remainder := MaxCols - totalColumns
	for i, opt := range cfg.Options {
		compi := p.makeOptionFn(opt, cfg)
		p.optComp = append(p.optComp, compi)
		effectiveCols := columns[i]
		if remainder > 0 {
			effectiveCols++
			remainder--
		}
		row.AddComponent(NewAspectRatioFloatingResponsive(compi, cfg.AspectRatio), effectiveCols)
	}
}

func (p *Prompt) init(
	makeOptionFn func(msg string, cfg PromptConfig) responsiveWithAttributes,
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

	messageResponsive := NewResponsiveString(cfg.Message, StringResponsiveConfig{
		NoSplitWords: true,
		StringConfig: StringConfig{
			PaddingVertical:      4,
			PaddingHorizontal:    4,
			Alignment:            AlignmentCentered,
			BackgroundAttributes: p.cfg.BackgroundAttributes,
			Attributes:           term.Attributes{Bg: p.cfg.BackgroundAttributes.Bg},
		}})

	p.messageRow = container.AddRow()
	p.messageRow.AddComponent(
		NewAspectRatioFloatingResponsive(messageResponsive, cfg.AspectRatio), MaxCols)

	optionsRow := container.AddRow()

	p.initOptions(cfg, optionsRow)

	p.effective = container

	p.effective = NewBackground(container, term.Cell{
		Attributes: p.cfg.BackgroundAttributes,
	})
}
