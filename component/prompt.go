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

// Prompt implements a prompt / question with options component. Its
// options share a row while a cell of air still fits between and
// around them, and wrap onto more rows when it does not, a blank row
// apart. They stay pinned to the bottom edge, so a prompt given less
// room than it needs keeps every option in view: it first packs the
// rows of options together, then scrolls its message, see Scrollable.
type Prompt struct {
	optComp      []*promptOption
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
// via UnionAttributes before drawing content. Over images, it is drawn
// on the prompt's background instead, for the images to show behind.
type promptOption struct {
	floatingOption
	width, height int
	attr          term.Attributes
	state         ButtonState
	onImages      bool
	promptBg      term.Color
}

func (o *promptOption) Resize(width, height int) {
	o.width, o.height = width, height
	o.floatingOption.Resize(width, height)
}

func (o *promptOption) Draw(w term.Writer) {
	if o.onImages {
		o.floatingOption.Draw(w)
		return
	}
	o.drawSolid(w)
}

// drawSolid draws the option as it looks without images.
func (o *promptOption) drawSolid(w term.Writer) {
	if o.onImages {
		o.SetAttr(o.attr)
		defer o.SetAttr(o.labelAttr())
	}
	bg := term.Attributes{Bg: o.attr.Bg}
	if o.attr.Attrs&term.AttrReverse != 0 {
		bg.Bg = o.attr.Fg
	}
	for y := range o.height {
		for x := range o.width {
			w.UnionAttributes(term.Coordinates{X: x, Y: y}, bg)
		}
	}
	o.floatingOption.Draw(w)
}

func (o *promptOption) setState(st ButtonState, attr term.Attributes) {
	o.state, o.attr = st, attr
	if o.onImages {
		o.SetAttr(o.labelAttr())
		return
	}
	o.SetAttr(attr)
}

// labelAttr is how the option is drawn over its images: in the
// foreground and attributes of its state, but on the prompt's
// background and never reversed, which would paint the images over.
func (o *promptOption) labelAttr() term.Attributes {
	return term.Attributes{
		Fg:    o.attr.Fg,
		Bg:    o.promptBg,
		Attrs: o.attr.Attrs &^ term.AttrReverse,
	}
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
	ButtonImages         *ButtonImages
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

// SetOptionAttr sets the attributes of option at index i, which it
// draws in ButtonIdle.
func (p *Prompt) SetOptionAttr(i int, attr term.Attributes) {
	p.SetOptionState(i, ButtonIdle, attr)
}

// SetOptionState sets the state option at index i is drawn in, which
// picks its ButtonImages, and its attributes.
func (p *Prompt) SetOptionState(i int, st ButtonState, attr term.Attributes) {
	p.optComp[i].setState(st, attr)
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
// the rows line up. spaced puts blank rows between rows, as many as
// the outsets of their images take, or one.
func (p *Prompt) layoutOptions(width int, spaced bool) {
	n := len(p.optVirtuals)
	rows := p.optionRows(width)
	perRow := (n + rows - 1) / rows
	_, between := p.optionAir()
	gap := max(between, (width-perRow*p.optMaxWidth)/(perRow+1))
	o := p.outset()
	p.opts.C = NewContainer()
	p.optRows = p.optRows[:0]
	if o.Top > 0 {
		p.opts.C.AddRow().AddComponent(promptGap{rows: o.Top}, MaxCols)
	}
	for first := 0; first < n; first += perRow {
		if spaced && first > 0 {
			p.opts.C.AddRow().AddComponent(promptGap{rows: p.rowGap()}, MaxCols)
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
	if o.Bottom > 0 {
		p.opts.C.AddRow().AddComponent(promptGap{rows: o.Bottom}, MaxCols)
	}
}

// optionRows is how many rows the options take at width: as few as
// leave the air of optionAir between every pair of buttons and at both
// edges.
func (p *Prompt) optionRows(width int) int {
	n := len(p.optVirtuals)
	edge, between := p.optionAir()
	perRow := min(max((width-2*edge+between)/(p.optMaxWidth+between), 1), n)
	return (n + perRow - 1) / perRow
}

// optionAir returns the fewest cells an option leaves to the prompt's
// edges and to the next option: one, or as many as keep the outsets of
// their images apart.
func (p *Prompt) optionAir() (edge, between int) {
	s := p.outset().Sides
	return max(1, s), max(1, 2*s)
}

// outset is how far the images of the options reach past them.
func (p *Prompt) outset() ButtonOutset {
	if p.cfg.ButtonImages == nil {
		return ButtonOutset{}
	}
	return p.cfg.ButtonImages.Outset.clamped()
}

// rowGap is the blank rows between two spaced rows of options.
func (p *Prompt) rowGap() int {
	o := p.outset()
	return max(1, o.Top+o.Bottom)
}

func (p *Prompt) optionsHeight(width int, spaced bool) int {
	rows := p.optionRows(width)
	o := p.outset()
	height := o.Top + rows*p.optMaxHeight + o.Bottom
	if spaced {
		height += (rows - 1) * p.rowGap()
	}
	return height
}

// Height satisfies Responsive: the message's height at width, or its
// ideal height when it is not Responsive, over the rows the options
// wrap onto at width, the blank rows between them and the rows their
// images reach above and below them.
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
	p.drawButtonImages(&bounds)
}

// drawButtonImages draws the images of every option under it, and
// draws again as it looks without images every option whose images w
// did not draw.
func (p *Prompt) drawButtonImages(w term.Writer) {
	if p.cfg.ButtonImages == nil {
		return
	}
	origin := p.opts.Position()
	for _, r := range p.optRows {
		top := origin.Y + r.row.Position().Y
		for i := range r.opts.opts {
			v := &r.opts.opts[i]
			pos := term.Coordinates{X: origin.X + v.Position().X, Y: top + v.Position().Y}
			label := image.Rect(pos.X, pos.Y, pos.X+v.Width(), pos.Y+v.Height())
			opt := p.optComp[r.first+i]
			if p.cfg.ButtonImages.Draw(w, label, opt.state) {
				continue
			}
			opt.drawSolid(&VirtualWriter{
				Writer: w, Offset: pos, Width: label.Dx(), Height: label.Dy(),
			})
		}
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
	// Ideal width: buttons + a 2-cell gap on each edge and between each
	// pair of buttons, or as many as keep their images' outsets apart.
	o := p.outset()
	edge, between := max(2, o.Sides), max(2, 2*o.Sides)
	return n*p.optMaxWidth + 2*edge + (n-1)*between,
		o.Top + p.optMaxHeight + o.Bottom
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
	p.optComp = make([]*promptOption, 0, n)
	p.optVirtuals = make([]Virtual[tui.Component], n)

	for i, opt := range cfg.Options {
		compi := p.makeOptionFn(opt, cfg)
		btn := &promptOption{
			floatingOption: compi,
			attr:           term.Attributes{Bg: cfg.BackgroundAttributes.Bg},
			onImages:       cfg.ButtonImages != nil,
			promptBg:       cfg.BackgroundAttributes.Bg,
		}
		p.optComp = append(p.optComp, btn)
		p.optVirtuals[i].C = btn
	}

	// Compute uniform button dimensions.
	p.optMaxWidth = 0
	p.optMaxHeight = 0
	for _, opt := range p.optComp {
		w, h := opt.Dimensions()
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

// promptGap is blank rows between two rows of options, or between them
// and the edges of the rows of options.
type promptGap struct{ rows int }

func (promptGap) Resize(int, int) {}

func (promptGap) Draw(term.Writer) {}

// Height satisfies Responsive.
func (g promptGap) Height(int) int { return g.rows }

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
