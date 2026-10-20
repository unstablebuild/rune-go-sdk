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

// Prompt implements a prompt / question with options component. Its
// options share a row while a cell of air still fits between and
// around them, and wrap onto more rows when it does not, a blank row
// apart. They stay pinned to the bottom edge, so a prompt given less
// room than it needs keeps every option in view: it first packs the
// rows of options together, then scrolls its message, see Scrollable.
type Prompt struct {
	optComp      []WithAttributes
	optVirtuals  []Virtual[tui.Component]
	optMaxWidth  int
	optMaxHeight int
	// opts stacks the rows of options under the message. Resize
	// rebuilds it, as how many options share a row depends on the
	// width.
	opts    Virtual[*Container]
	optRows []promptRowEntry
	message tui.Component
	// messageContent is message without its background, which
	// knows how tall the message is.
	messageContent tui.Component
	cfg            PromptConfig
	width          int
	height         int
	// msgHeight is the rows shown of the message, which is msgFull
	// rows tall and scrolled msgOffset rows down.
	msgHeight    int
	msgFull      int
	msgOffset    int
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
	if attr.Attrs&term.AttrReverse != 0 {
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

var (
	_ Floating   = (*Prompt)(nil)
	_ Responsive = (*Prompt)(nil)
	_ Scrollable = (*Prompt)(nil)
)

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
	spaced := p.Height(width) <= height
	p.layoutOptions(width, spaced)
	optHeight := p.optionsHeight(width, spaced)
	p.msgHeight = max(0, height-optHeight)
	p.msgFull = max(p.msgHeight, p.messageHeight(width))
	p.msgOffset = min(p.msgOffset, p.MaxSeekOffset())
	p.message.Resize(width, p.msgFull)
	p.opts.Move(term.Coordinates{Y: p.msgHeight})
	p.opts.Resize(width, optHeight)
}

// layoutOptions stacks the options in as few rows as width fits,
// spread evenly over them so the last row is not left with a
// straggler. Every row spaces its buttons as a full one would, so
// the rows line up. spaced puts a blank row between rows.
func (p *Prompt) layoutOptions(width int, spaced bool) {
	n := len(p.optVirtuals)
	rows := p.optionRows(width)
	perRow := (n + rows - 1) / rows
	gap := max(0, (width-perRow*p.optMaxWidth)/(perRow+1))
	p.opts.C = NewContainer()
	p.optRows = p.optRows[:0]
	for first := 0; first < n; first += perRow {
		if spaced && first > 0 {
			p.opts.C.AddRow().AddComponent(promptGap{}, MaxCols)
		}
		opts := &promptRow{
			opts:   p.optVirtuals[first:min(first+perRow, n)],
			width:  p.optMaxWidth,
			height: p.optMaxHeight,
			gap:    gap,
		}
		row := p.opts.C.AddRow()
		row.AddComponent(opts, MaxCols)
		p.optRows = append(p.optRows,
			promptRowEntry{row: row, opts: opts, first: first})
	}
}

// optionRows is how many rows the options take at width: as few as
// leave a cell between every pair of buttons and at both edges.
func (p *Prompt) optionRows(width int) int {
	n := len(p.optVirtuals)
	perRow := min(max((width-1)/(p.optMaxWidth+1), 1), n)
	return (n + perRow - 1) / perRow
}

func (p *Prompt) optionsHeight(width int, spaced bool) int {
	rows := p.optionRows(width)
	height := rows * p.optMaxHeight
	if spaced {
		height += rows - 1
	}
	return height
}

// Height satisfies Responsive: the message's height at width, or its
// ideal height when it is not Responsive, over the rows the options
// wrap onto at width and the blank rows between them.
func (p *Prompt) Height(width int) int {
	return p.messageHeight(width) + p.optionsHeight(width, true)
}

func (p *Prompt) messageHeight(width int) int {
	switch m := p.messageContent.(type) {
	case Responsive:
		return m.Height(width)
	case Floating:
		_, height := m.Dimensions()
		return height
	}
	return 0
}

// SeekUp satisfies Scrollable by scrolling the message up a row.
func (p *Prompt) SeekUp() bool {
	if p.msgOffset == 0 {
		return false
	}
	p.msgOffset--
	return true
}

// SeekDown satisfies Scrollable by scrolling the message down a row.
func (p *Prompt) SeekDown() bool {
	if p.msgOffset >= p.MaxSeekOffset() {
		return false
	}
	p.msgOffset++
	return true
}

// SeekOffset satisfies Scrollable.
func (p *Prompt) SeekOffset() int {
	return p.msgOffset
}

// MaxSeekOffset satisfies Scrollable. It is zero unless the last
// Resize left the message less room than it needs.
func (p *Prompt) MaxSeekOffset() int {
	return p.msgFull - p.msgHeight
}

// Draw satisfies tui.Component
func (p *Prompt) Draw(w term.Writer) {
	// Options taller or wider than the prompt are cut to it.
	bounds := VirtualWriter{Writer: w, Height: p.height, Width: p.width}

	// Fill background.
	bgCell := term.NewCell(0, 1, p.cfg.BackgroundAttributes)
	for y := 0; y < p.height; y++ {
		for x := 0; x < p.width; x++ {
			bounds.SetCell(term.Coordinates{X: x, Y: y}, bgCell)
		}
	}

	// The message is drawn scrolled into the rows above the options.
	shown := VirtualWriter{Writer: &bounds, Height: p.msgHeight, Width: p.width}
	scrolled := VirtualWriter{
		Writer: &shown,
		Offset: term.Coordinates{Y: -p.msgOffset},
		Height: p.msgFull,
		Width:  p.width,
	}
	p.message.Draw(&scrolled)
	p.opts.Draw(&bounds)
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
// or -1 if no option is at that position. Coordinates are those of
// the last Resize, and an option cut off by it is not at the cells
// it was cut off from.
func (p *Prompt) OptionAt(x, y int) int {
	if x < 0 || y < 0 || x >= p.width || y >= p.height {
		return -1
	}
	origin := p.opts.Position()
	x, y = x-origin.X, y-origin.Y
	for _, r := range p.optRows {
		top := r.row.Position().Y
		if y < top || y >= top+r.row.height {
			continue
		}
		if i := r.opts.optionAt(x, y-top); i >= 0 {
			return r.first + i
		}
		return -1
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

	p.message = NewBackground(messageComp, term.NewCell(0, 0, p.cfg.BackgroundAttributes))
	p.messageContent = messageComp

	p.initOptions(cfg)

	// The last layout was of the old message and options, so none
	// of it holds until the next Resize.
	p.width, p.height = 0, 0
	p.msgHeight, p.msgFull, p.msgOffset = 0, 0, 0
	p.opts = Virtual[*Container]{C: NewContainer()}
	p.optRows = nil
}

// promptGap is the blank row between two rows of options.
type promptGap struct{}

func (promptGap) Resize(int, int) {}

func (promptGap) Draw(term.Writer) {}

// Height satisfies Responsive.
func (promptGap) Height(int) int { return 1 }

// promptRowEntry is a row of options as the Container holds it and
// as the prompt hit-tests it.
type promptRowEntry struct {
	row   *Row
	opts  *promptRow
	first int
}

// promptRow lays out a row of a prompt's options: buttons all as wide
// as the widest option, gap cells apart, centered as a group.
type promptRow struct {
	// opts shares its backing array with Prompt.optVirtuals, so
	// laying them out here moves the prompt's options.
	opts          []Virtual[tui.Component]
	width, height int
	gap           int
}

// Height satisfies Responsive.
func (r *promptRow) Height(int) int {
	return r.height
}

// Resize satisfies tui.Component.
func (r *promptRow) Resize(width, height int) {
	n := len(r.opts)
	group := n*r.width + (n-1)*r.gap
	// Widen the first gap rather than leave the margins lopsided.
	extraGap := 0
	if n > 1 && (width-group)%2 != 0 {
		extraGap = 1
		group++
	}
	x := max(0, (width-group)/2)
	for i := range r.opts {
		r.opts[i].Move(term.Coordinates{X: x})
		r.opts[i].Resize(max(0, min(r.width, width-x)), height)
		x += r.width + r.gap
		if i == 0 {
			x += extraGap
		}
	}
}

// Draw satisfies tui.Component.
func (r *promptRow) Draw(w term.Writer) {
	for i := range r.opts {
		r.opts[i].Draw(w)
	}
}

// optionAt returns the index within the row of the option at x, y
// in row coordinates, or -1.
func (r *promptRow) optionAt(x, y int) int {
	for i := range r.opts {
		pos := r.opts[i].Position()
		if x >= pos.X && x < pos.X+r.opts[i].Width() &&
			y >= 0 && y < r.opts[i].Height() {
			return i
		}
	}
	return -1
}
