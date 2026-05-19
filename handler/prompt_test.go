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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestPromptDefaults(t *testing.T) {
	t.Run("sets default callback", func(t *testing.T) {
		p := NewPrompt(PromptConfig{
			PromptConfig: component.PromptConfig{
				Message: "blah", Options: []string{"a"},
			}})
		exit, handled := p.Handle(term.Event{
			Type: term.EventKey, Key: term.KeyEnter})
		assert.True(t, exit)
		assert.True(t, handled)
	})
	t.Run("sets default highlight attrs", func(t *testing.T) {
		p := NewPrompt(PromptConfig{
			PromptConfig: component.PromptConfig{
				Message: "blah", Options: []string{"a"},
			}})
		attr := term.Attributes{Attrs: term.AttrReverse}
		assert.Equal(t, attr, p.cfg.HighlightAttr)
	})
}

func TestPromptHandle(t *testing.T) {
	opt0 := "Say what?"
	opt1 := "Yes"

	var calledI int
	var calledOpt string
	cfg := PromptConfig{
		PromptConfig: component.PromptConfig{
			Message: "What's for supper?",
			Options: []string{opt0, opt1},
			Frame:   component.FrameCharSetDefault(),
		},
		PromptHandler: FuncPromptHandler(
			func(i int, opt string) {
				calledI = i
				calledOpt = opt
			},
			func() error { return nil }),
		OptionBindings: []term.KeyComb{

			{Ch: 'W'},
			{Mod: term.ModAlt, Ch: 'Y'},
		},
	}

	h := NewPrompt(cfg)

	resetStub := func() {
		calledI = -1
		calledOpt = "-1"
		h.Init(cfg)
	}
	resetStub()

	t.Run("enter after init calls first option", func(t *testing.T) {
		defer resetStub()

		exit, handled := h.Handle(term.Event{
			Type: term.EventKey, Key: term.KeyEnter})
		assert.True(t, exit)
		assert.True(t, handled)

		assert.Equal(t, 0, calledI)
		assert.Equal(t, opt0, calledOpt)
	})

	t.Run("arrow key right allow for moving right", func(t *testing.T) {
		defer resetStub()

		exit, handled := h.Handle(term.Event{
			Type: term.EventKey, Key: term.KeyArrowRight})
		assert.False(t, exit)
		assert.True(t, handled)

		exit, handled = h.Handle(term.Event{
			Type: term.EventKey, Key: term.KeyArrowRight})
		assert.False(t, exit)
		assert.False(t, handled)

		exit, handled = h.Handle(term.Event{
			Type: term.EventKey, Key: term.KeyEnter})
		assert.True(t, exit)
		assert.True(t, handled)

		assert.Equal(t, 1, calledI)
		assert.Equal(t, opt1, calledOpt)
	})

	t.Run("key ctrl l allow for moving right", func(t *testing.T) {
		defer resetStub()

		exit, handled := h.Handle(term.Event{
			Type: term.EventKey, Mod: term.ModCtrl, Ch: 'l'})
		assert.False(t, exit)
		assert.True(t, handled)

		exit, handled = h.Handle(term.Event{
			Type: term.EventKey, Key: term.KeyEnter})
		assert.True(t, exit)
		assert.True(t, handled)

		assert.Equal(t, 1, calledI)
		assert.Equal(t, opt1, calledOpt)
	})

	t.Run("arrow key left allow for moving left", func(t *testing.T) {
		defer resetStub()

		exit, handled := h.Handle(term.Event{
			Type: term.EventKey, Key: term.KeyArrowRight})
		assert.False(t, exit)
		assert.True(t, handled)

		exit, handled = h.Handle(term.Event{
			Type: term.EventKey, Key: term.KeyArrowLeft})
		assert.False(t, exit)
		assert.True(t, handled)

		exit, handled = h.Handle(term.Event{
			Type: term.EventKey, Key: term.KeyArrowLeft})
		assert.False(t, exit)
		assert.False(t, handled)

		exit, handled = h.Handle(term.Event{
			Type: term.EventKey, Key: term.KeyEnter})
		assert.True(t, exit)
		assert.True(t, handled)

		assert.Equal(t, 0, calledI)
		assert.Equal(t, opt0, calledOpt)
	})

	t.Run("ctrl h allow for moving left", func(t *testing.T) {
		defer resetStub()

		exit, handled := h.Handle(term.Event{
			Type: term.EventKey, Key: term.KeyArrowRight})
		assert.False(t, exit)
		assert.True(t, handled)

		exit, handled = h.Handle(term.Event{
			Type: term.EventKey, Mod: term.ModCtrl, Ch: 'h'})
		assert.False(t, exit)
		assert.True(t, handled)

		exit, handled = h.Handle(term.Event{
			Type: term.EventKey, Key: term.KeyEnter})
		assert.True(t, exit)
		assert.True(t, handled)

		assert.Equal(t, 0, calledI)
		assert.Equal(t, opt0, calledOpt)
	})

	t.Run("valid auto key binding no conflict", func(t *testing.T) {
		defer resetStub()

		exit, handled := h.Handle(term.Event{
			Type: term.EventKey, Mod: term.ModAlt, Ch: 'Y'})
		assert.True(t, exit)
		assert.True(t, handled)

		assert.Equal(t, 1, calledI)
		assert.Equal(t, opt1, calledOpt)
	})

	t.Run("invalid key binding", func(t *testing.T) {
		defer resetStub()

		exit, handled := h.Handle(term.Event{Type: term.EventKey, Ch: ' '})
		assert.False(t, exit)
		assert.False(t, handled)

		assert.Equal(t, -1, calledI)
		assert.Equal(t, "-1", calledOpt)
	})
}

func TestPromptHandler(t *testing.T) {
	t.Run("OnSelect is called upon Select", func(t *testing.T) {
		promptHandler := new(testPromptHandler)
		p := NewPrompt(PromptConfig{
			PromptConfig: component.PromptConfig{
				Message: "blah", Options: []string{"a", "b"},
			},
			PromptHandler: promptHandler,
			OptionBindings: []term.KeyComb{
				{Ch: 'a'},
				{Ch: 'b'},
			}})
		p.Handle(term.Event{
			Type: term.EventKey, Ch: 'b'})
		assert.True(t, promptHandler.onSelectCalled)
		assert.False(t, promptHandler.onCloseCalled)
		assert.Equal(t, promptHandler.selectedIdx, 1)
		assert.Equal(t, promptHandler.selectedOption, "b")
	})

	t.Run("OnClose is called upon Close", func(t *testing.T) {
		promptHandler := new(testPromptHandler)
		p := NewPrompt(PromptConfig{
			PromptConfig: component.PromptConfig{
				Message: "blah", Options: []string{"a", "b"},
			},
			PromptHandler: promptHandler,
			OptionBindings: []term.KeyComb{
				{Ch: 'a'},
				{Ch: 'b'},
			}})
		assert.NoError(t, p.Close())
		assert.False(t, promptHandler.onSelectCalled)
		assert.True(t, promptHandler.onCloseCalled)
	})
}

func TestPromptMouseClick(t *testing.T) {
	const width, height = 20, 10

	newPrompt := func(t *testing.T) (
		*Prompt, *testPromptHandler,
	) {
		t.Helper()
		ph := new(testPromptHandler)
		p := NewPrompt(PromptConfig{
			PromptConfig: component.PromptConfig{
				Message: "Ok?",
				Options: []string{"Yes", "No"},
				Frame:   component.FrameCharSetDefault(),
			},
			PromptHandler: ph,
			HighlightAttr: term.Attributes{
				Bg: term.ColorYellow,
			},
		})
		p.Resize(width, height)
		return p, ph
	}

	// Helper: draw and return string with BackgroundCh='B'
	// so any cell with Bg != 0 renders as 'B'.
	drawBg := func(t *testing.T, p *Prompt) string {
		t.Helper()
		w := term.NewStringWriter(width+1, height+1)
		w.BackgroundCh = 'B'
		_ = w.Clear(term.Attributes{})
		p.Draw(w)
		_ = w.Flush()
		return w.String()
	}

	// Each option span = 10 columns (half of 20).
	// "Yes" frame (7) centered in 10: x=[1,7], y=7..9
	// "No" frame (6) centered in 10: x=[12,17], y=7..9
	const (
		yesX = 4 // inside "Yes" button
		noX  = 13
		btnY = 8 // middle row of the button
	)

	t.Run("click highlights button", func(t *testing.T) {
		p, _ := newPrompt(t)

		// Before click: only keyboard-focused "Yes"
		// has HighlightAttr (Bg=Yellow).
		before := drawBg(t, p)

		// Click on "No" button.
		exit, handled := p.Handle(term.Event{
			Type:   term.EventMouse,
			Key:    term.MouseLeft,
			MouseX: noX,
			MouseY: btnY,
		})
		assert.False(t, exit)
		assert.True(t, handled)

		after := drawBg(t, p)
		assert.NotEqual(t, before, after,
			"clicking should change button attributes")

		// The "No" button area should now have 'B' chars
		// (HighlightAttr has Bg=Yellow).
		assert.Contains(t, after, "BBBBB",
			"clicked button should show background attr")
	})

	t.Run("release restores and selects", func(t *testing.T) {
		p, ph := newPrompt(t)

		// Click on "No" button.
		p.Handle(term.Event{
			Type:   term.EventMouse,
			Key:    term.MouseLeft,
			MouseX: noX,
			MouseY: btnY,
		})

		// Release on "No" button.
		exit, handled := p.Handle(term.Event{
			Type:   term.EventMouse,
			Key:    term.MouseRelease,
			MouseX: noX,
			MouseY: btnY,
		})
		assert.True(t, exit)
		assert.True(t, handled)
		assert.True(t, ph.onSelectCalled)
		assert.Equal(t, 1, ph.selectedIdx)
		assert.Equal(t, "No", ph.selectedOption)
	})

	t.Run("release outside cancels", func(t *testing.T) {
		p, ph := newPrompt(t)
		before := drawBg(t, p)

		// Click on "Yes" button.
		p.Handle(term.Event{
			Type:   term.EventMouse,
			Key:    term.MouseLeft,
			MouseX: yesX,
			MouseY: btnY,
		})

		// Release far away (not on any button).
		exit, handled := p.Handle(term.Event{
			Type:   term.EventMouse,
			Key:    term.MouseRelease,
			MouseX: 0,
			MouseY: 0,
		})
		assert.False(t, exit)
		assert.True(t, handled)
		assert.False(t, ph.onSelectCalled)

		// Visual state should be restored.
		after := drawBg(t, p)
		assert.Equal(t, before, after,
			"release outside should restore attrs")
	})

	t.Run("click outside button is not handled", func(t *testing.T) {
		p, _ := newPrompt(t)

		exit, handled := p.Handle(term.Event{
			Type:   term.EventMouse,
			Key:    term.MouseLeft,
			MouseX: 0,
			MouseY: 0,
		})
		assert.False(t, exit)
		assert.False(t, handled)
	})

	t.Run("release without press is not handled", func(t *testing.T) {
		p, _ := newPrompt(t)

		exit, handled := p.Handle(term.Event{
			Type:   term.EventMouse,
			Key:    term.MouseRelease,
			MouseX: noX,
			MouseY: btnY,
		})
		assert.False(t, exit)
		assert.False(t, handled)
	})
}

type testPromptHandler struct {
	onCloseCalled  bool
	onSelectCalled bool
	selectedIdx    int
	selectedOption string
}

func (h *testPromptHandler) OnClose() error {
	h.onCloseCalled = true
	return nil
}

func (h *testPromptHandler) OnSelect(idx int, option string) {
	h.onSelectCalled = true
	h.selectedIdx = idx
	h.selectedOption = option
}
