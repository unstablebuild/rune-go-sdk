// Unstable Build LLC ("COMPANY") CONFIDENTIAL
//
// Unpublished Copyright (c) 2017-2024 Unstable Build, All Rights Reserved.
//
// NOTICE: All information contained herein is, and remains the property of COMPANY.
// The intellectual and technical concepts contained herein are proprietary to
// COMPANY and may be covered by U.S. and Foreign Patents, patents in process,
// and are protected by trade secret or copyright law. Dissemination of this information
// or reproduction of this material is strictly forbidden unless prior written permission
// is obtained from COMPANY. Access to the source code contained herein is hereby
// forbidden to anyone except current COMPANY employees, managers or contractors who
// have executed Confidentiality and Non-disclosure agreements explicitly covering such access.
//
// The copyright notice above does not evidence any actual or intended publication or
// disclosure of this source code, which includes information that is confidential and/or
// proprietary, and is a trade secret, of COMPANY. ANY REPRODUCTION, MODIFICATION,
// DISTRIBUTION, PUBLIC  PERFORMANCE, OR PUBLIC DISPLAY OF OR THROUGH USE OF THIS SOURCE CODE
// WITHOUT  THE EXPRESS WRITTEN CONSENT OF COMPANY IS STRICTLY PROHIBITED, AND IN
// VIOLATION OF APPLICABLE LAWS AND INTERNATIONAL TREATIES. THE RECEIPT OR POSSESSION OF
// THIS SOURCE CODE AND/OR RELATED INFORMATION DOES NOT CONVEY OR IMPLY ANY RIGHTS TO
// REPRODUCE, DISCLOSE OR DISTRIBUTE ITS CONTENTS, OR TO MANUFACTURE, USE, OR SELL
// ANYTHING THAT IT MAY DESCRIBE, IN WHOLE OR IN PART.

package handler

import (
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
	"github.com/unstablebuild/tcell/v3"
)

// PromptHandler provides hooks to be called upon Prompt actions.
type PromptHandler interface {
	// OnSelect is run when the user makes a choice from the prompt.
	OnSelect(idx int, option string)
	// OnClose runs when the user dismisses the prompt (usually by closing the
	// component that hosts the Prompt).
	OnClose() error
}

// NopPromptHandler won't do anything on selecting prompt choices nor closing.
func NopPromptHandler() PromptHandler {
	return nopPromptHandler{}
}

// FuncPromptHandler will run the given callbacks upon selecting or closing.
func FuncPromptHandler(
	selectCb func(idx int, option string),
	closeCb func() error,
) PromptHandler {
	return funcPromptHandler{selectCb: selectCb, closeCb: closeCb}
}

// PromptConfig holds configuration for a Prompt.
type PromptConfig struct {
	component.PromptConfig
	PromptHandler

	OptionBindings []term.KeyComb
	HighlightAttr  term.Attributes
	OptionAttr     term.Attributes
}

var _ tui.Handler = (*Prompt)(nil)

// Prompt wraps a component.Prompt to satisfy tui.Handler.
type Prompt struct {
	component.Prompt

	hi       int
	pressed  int
	cfg      PromptConfig
	bindings map[term.KeyComb]int
}

// NewPrompt allocates storage for a new Prompt and initializes it.
// See Init for more details.
func NewPrompt(cfg PromptConfig) (f *Prompt) {
	f = new(Prompt)
	f.Init(cfg)
	return
}

// Init initializes this Prompt with the given PromptConfig.
// Note that if OptionBindings is defined, it should be of the same
// length as Options.
func (f *Prompt) Init(cfg PromptConfig) {
	f.Prompt.Init(cfg.PromptConfig)
	f.hi = 0       // allow for Init to be used as reset
	f.pressed = -1 // no button pressed

	if len(cfg.OptionBindings) != 0 &&
		len(cfg.OptionBindings) != len(cfg.Options) {
		panic("invalid OptionBindings; length should match of Options")
	}

	if cfg.HighlightAttr == (term.Attributes{}) {
		cfg.HighlightAttr = term.Attributes{
			Bg:    cfg.OptionAttr.Bg,
			Fg:    cfg.OptionAttr.Fg,
			Attrs: cfg.OptionAttr.Attrs | tcell.AttrReverse,
		}
	}

	if cfg.PromptHandler == nil {
		cfg.PromptHandler = NopPromptHandler()
	}

	f.cfg = cfg
	f.bindings = make(map[term.KeyComb]int)
	for i, ev := range f.cfg.OptionBindings {
		f.bindings[ev] = i
	}

	// Prompt guarantees that there's at least one option
	f.highlightOption()
}

func (f *Prompt) highlightOption() {
	for j := range f.cfg.Options {
		f.SetOptionAttr(j, f.cfg.OptionAttr)
	}
	f.SetOptionAttr(f.hi, f.cfg.HighlightAttr)
}

// Handle satisfies tui.Handler.
func (f *Prompt) Handle(ev term.Event) (exit, handled bool) {
	if ev.Type == term.EventMouse {
		return f.handleMouse(ev)
	}
	if ev.Type != term.EventKey {
		return
	}

	if i, ok := f.bindings[ev.KeyComb()]; ok {
		f.cfg.OnSelect(i, f.cfg.Options[i])
		exit = true
		handled = true
		return
	}

	if ev.Mod == term.ModCtrl {
		switch ev.Ch {
		case 'h':
			handled = f.optionLeft()
		case 'l':
			handled = f.optionRight()
		}
		if handled {
			return
		}
	}

	if ev.Mod != 0 {
		return
	}

	switch ev.Key {
	case term.KeyArrowLeft:
		handled = f.optionLeft()
	case term.KeyArrowRight:
		handled = f.optionRight()
	case term.KeyEnter:
		f.cfg.OnSelect(f.hi, f.cfg.Options[f.hi])
		exit = true
		handled = true
	case term.KeyEsc:
		exit = true
		handled = true
	}
	return
}

func (f *Prompt) handleMouse(
	ev term.Event,
) (exit, handled bool) {
	switch ev.Key {
	case term.MouseLeft:
		i := f.OptionAt(ev.MouseX, ev.MouseY)
		if i < 0 {
			return
		}
		f.pressed = i
		f.SetOptionAttr(i, f.cfg.HighlightAttr)
		handled = true
	case term.MouseRelease:
		if f.pressed < 0 {
			return
		}
		pressed := f.pressed
		f.pressed = -1
		f.highlightOption()
		handled = true
		if f.OptionAt(ev.MouseX, ev.MouseY) == pressed {
			f.cfg.OnSelect(pressed, f.cfg.Options[pressed])
			exit = true
		}
	}
	return
}

func (f *Prompt) optionRight() (handled bool) {
	if f.hi < len(f.cfg.Options)-1 {
		f.hi++
		handled = true
		f.highlightOption()
	}
	return
}

func (f *Prompt) optionLeft() (handled bool) {
	if f.hi > 0 {
		f.hi--
		handled = true
		f.highlightOption()
	}
	return
}

// Selection satisfies tui.Handler but always returns false.
func (f *Prompt) Selection() (string, bool) {
	return "", false
}

// Close satisfies tui.Handler running the configured close callback.
func (f *Prompt) Close() error {
	return f.cfg.OnClose()
}

// Cursor satisfies tui.Handler.
func (f *Prompt) Cursor() (pos term.Coordinates, style term.CursorStyle, show bool) {
	return
}

type nopPromptHandler struct{}

func (h nopPromptHandler) OnSelect(idx int, option string) {}

func (h nopPromptHandler) OnClose() error {
	return nil
}

type funcPromptHandler struct {
	selectCb func(idx int, option string)
	closeCb  func() error
}

func (h funcPromptHandler) OnSelect(idx int, option string) {
	h.selectCb(idx, option)
}

func (h funcPromptHandler) OnClose() error {
	return h.closeCb()
}
