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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	t.Run("the pointer is a finger over the buttons", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			key  term.Key
			x, y int
			want term.PointerShape
		}{
			{"hover over a button", 0, noX, btnY, term.PointerLink},
			{"press on a button", term.MouseLeft, yesX, btnY, term.PointerLink},
			{"hover between the buttons", 0, 9, btnY, term.PointerDefault},
			{"hover over the message", 0, 4, 2, term.PointerDefault},
		} {
			t.Run(tc.name, func(t *testing.T) {
				p, _ := newPrompt(t)
				ptr := new(term.Pointer)
				// A fallback a parent offers must not show through.
				ptr.SetShape(term.PointerText)
				p.Handle(term.Event{
					Type: term.EventMouse, Key: tc.key, MouseX: tc.x, MouseY: tc.y,
					Context: term.ContextWithPointer(context.Background(), ptr),
				})
				shape, _ := ptr.Shape()
				assert.Equal(t, tc.want.String(), shape.String())
			})
		}
	})

	t.Run("click on a wrapped row selects its option", func(t *testing.T) {
		ph := new(testPromptHandler)
		p := NewPrompt(PromptConfig{
			PromptConfig: component.PromptConfig{
				Message: "Ok?",
				Options: []string{"Yay", "Nay", "Say"},
				Frame:   component.FrameCharSetDefault(),
			},
			PromptHandler: ph,
		})
		// Too narrow for three framed buttons: "Say" wraps
		// onto rows 7..9, centered at x=6..12.
		p.Resize(width, height)
		for _, key := range []term.Key{term.MouseLeft, term.MouseRelease} {
			p.Handle(term.Event{
				Type: term.EventMouse, Key: key, MouseX: 9, MouseY: 8,
			})
		}
		assert.True(t, ph.onSelectCalled)
		assert.Equal(t, 2, ph.selectedIdx)
		assert.Equal(t, "Say", ph.selectedOption)
	})
}

func TestPromptMouseWheelScrollsTheMessage(t *testing.T) {
	down, up := term.MouseWheelDown, term.MouseWheelUp
	// A message four rows tall over "Yes" and "No", which share a
	// row from a width of nine, with "Yes" at x=5..7 at a width of
	// 20. Three rows leave the message two, so it scrolls two rows.
	tests := []struct {
		name          string
		width, height int
		unsized       bool
		x, y          int
		keys          []term.Key
		wantOffset    int
		wantHandled   bool
	}{
		{name: "down a row", width: 20, height: 3, keys: []term.Key{down},
			wantOffset: 1, wantHandled: true},
		{name: "down to the last row", width: 20, height: 3, keys: []term.Key{down, down},
			wantOffset: 2, wantHandled: true},
		{name: "no further than the last row", width: 20, height: 3,
			keys: []term.Key{down, down, down}, wantOffset: 2},
		{name: "back up a row", width: 20, height: 3, keys: []term.Key{down, up},
			wantHandled: true},
		{name: "no further than the top", width: 20, height: 3, keys: []term.Key{up}},
		{name: "up and back down", width: 20, height: 3, keys: []term.Key{down, down, up, down},
			wantOffset: 2, wantHandled: true},
		{name: "over an option", width: 20, height: 3, x: 5, y: 2, keys: []term.Key{down},
			wantOffset: 1, wantHandled: true},
		{name: "beyond the prompt", width: 20, height: 3, x: 40, y: 9, keys: []term.Key{down},
			wantOffset: 1, wantHandled: true},
		{name: "a message with room to spare", width: 20, height: 20, keys: []term.Key{down}},
		{name: "a message with exactly the room it asks for", width: 20, height: 5,
			keys: []term.Key{down}},
		{name: "a message the options leave no row", width: 20, height: 1,
			keys: []term.Key{down, down, down, down, down}, wantOffset: 4},
		{name: "a prompt with no rows", width: 20, height: 0, keys: []term.Key{down},
			wantOffset: 1, wantHandled: true},
		{name: "a message over packed wrapped options", width: 8, height: 4,
			keys: []term.Key{down, down, down}, wantOffset: 2},
		{name: "before any Resize", unsized: true, keys: []term.Key{down}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ph := new(testPromptHandler)
			p := NewPrompt(PromptConfig{
				PromptConfig: component.PromptConfig{
					Message: "Ok?",
					Options: []string{"Yes", "No"},
					NewMessage: func(string) component.Floating {
						return &component.TestResponsive{WantHeight: 4}
					},
				},
				PromptHandler: ph,
			})
			if !tt.unsized {
				p.Resize(tt.width, tt.height)
			}
			var handled bool
			for i, key := range tt.keys {
				var exit bool
				exit, handled = p.Handle(term.Event{
					Type: term.EventMouse, Key: key, MouseX: tt.x, MouseY: tt.y,
				})
				assert.False(t, exit, "event %d", i)
			}
			assert.Equal(t, tt.wantHandled, handled)
			assert.Equal(t, tt.wantOffset, p.SeekOffset())
			assert.False(t, ph.onSelectCalled)
		})
	}
}

func TestPromptHover(t *testing.T) {
	tests := []struct {
		name        string
		noHoverAttr bool
		// unsized sends the events before the first Resize.
		unsized     bool
		events      []term.Event
		wantHovered int
		wantYes     term.Attributes
		wantNo      term.Attributes
	}{
		{
			name:        "the keyboard's choice before the pointer arrives",
			wantHovered: -1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name:        "hovering an option",
			events:      []term.Event{hoverTestMove(hoverTestNoX)},
			wantHovered: 1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestHoverAttr,
		},
		{
			name:        "hovering the keyboard's choice",
			events:      []term.Event{hoverTestMove(hoverTestYesX)},
			wantHovered: 0,
			wantYes:     hoverTestHoverAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name:        "the last cell of an option",
			events:      []term.Event{hoverTestMove(hoverTestNoX + 2)},
			wantHovered: 1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestHoverAttr,
		},
		{
			name:        "from one option onto the other",
			events:      []term.Event{hoverTestMove(hoverTestYesX), hoverTestMove(hoverTestNoX)},
			wantHovered: 1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestHoverAttr,
		},
		{
			name:        "moving off the options",
			events:      []term.Event{hoverTestMove(hoverTestNoX), hoverTestMove(9)},
			wantHovered: -1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name: "moving back onto an option",
			events: []term.Event{
				hoverTestMove(hoverTestNoX), hoverTestMove(9), hoverTestMove(hoverTestNoX),
			},
			wantHovered: 1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestHoverAttr,
		},
		{
			name:        "the message",
			events:      []term.Event{hoverTestMouse(0, hoverTestYesX, 2)},
			wantHovered: -1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name:        "left of the prompt",
			events:      []term.Event{hoverTestMove(-1)},
			wantHovered: -1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name:        "right of the prompt",
			events:      []term.Event{hoverTestMove(hoverTestWidth)},
			wantHovered: -1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name:        "below the prompt",
			events:      []term.Event{hoverTestMouse(0, hoverTestNoX, hoverTestHeight)},
			wantHovered: -1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name:        "a press hovers the option it lands on",
			events:      []term.Event{hoverTestMouse(term.MouseLeft, hoverTestNoX, hoverTestBtnY)},
			wantHovered: 1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestHighlightAttr,
		},
		{
			name: "a pressed option keeps the highlight",
			events: []term.Event{
				hoverTestMove(hoverTestNoX),
				hoverTestMouse(term.MouseLeft, hoverTestNoX, hoverTestBtnY),
			},
			wantHovered: 1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestHighlightAttr,
		},
		{
			name: "a press off the options",
			events: []term.Event{
				hoverTestMove(hoverTestNoX), hoverTestMouse(term.MouseLeft, 9, hoverTestBtnY),
			},
			wantHovered: -1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name: "dragging a press onto the other option",
			events: []term.Event{
				hoverTestMouse(term.MouseLeft, hoverTestNoX, hoverTestBtnY),
				hoverTestMove(hoverTestYesX),
			},
			wantHovered: 0,
			wantYes:     hoverTestHoverAttr, wantNo: hoverTestHighlightAttr,
		},
		{
			name: "a cancelled press leaves the hover where it ends",
			events: []term.Event{
				hoverTestMouse(term.MouseLeft, hoverTestNoX, hoverTestBtnY),
				hoverTestMouse(term.MouseRelease, hoverTestYesX, hoverTestBtnY),
			},
			wantHovered: 0,
			wantYes:     hoverTestHoverAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name: "a press released off the options",
			events: []term.Event{
				hoverTestMouse(term.MouseLeft, hoverTestNoX, hoverTestBtnY),
				hoverTestMouse(term.MouseRelease, 9, hoverTestBtnY),
			},
			wantHovered: -1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name: "a selecting release keeps the hover",
			events: []term.Event{
				hoverTestMouse(term.MouseLeft, hoverTestNoX, hoverTestBtnY),
				hoverTestMouse(term.MouseRelease, hoverTestNoX, hoverTestBtnY),
			},
			wantHovered: 1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestHoverAttr,
		},
		{
			name:        "a wheel over an option",
			events:      []term.Event{hoverTestMouse(term.MouseWheelDown, hoverTestNoX, hoverTestBtnY)},
			wantHovered: 1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestHoverAttr,
		},
		{
			name: "the keyboard moving onto the hovered option",
			events: []term.Event{
				hoverTestMove(hoverTestNoX), {Type: term.EventKey, Key: term.KeyArrowRight},
			},
			wantHovered: 1,
			wantYes:     hoverTestOptionAttr, wantNo: hoverTestHoverAttr,
		},
		{
			name: "the keyboard moving off the hovered option",
			events: []term.Event{
				hoverTestMove(hoverTestYesX), {Type: term.EventKey, Key: term.KeyArrowRight},
			},
			wantHovered: 0,
			wantYes:     hoverTestHoverAttr, wantNo: hoverTestHighlightAttr,
		},
		{
			name: "a key the prompt ignores",
			events: []term.Event{
				hoverTestMove(hoverTestNoX), {Type: term.EventKey, Key: term.KeyArrowLeft},
			},
			wantHovered: 1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestHoverAttr,
		},
		{
			name:        "before any Resize",
			unsized:     true,
			events:      []term.Event{hoverTestMove(hoverTestNoX)},
			wantHovered: -1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name:        "no hover attributes",
			noHoverAttr: true,
			events:      []term.Event{hoverTestMove(hoverTestNoX)},
			wantHovered: 1,
			wantYes:     hoverTestOptionAttr, wantNo: hoverTestHighlightAttr,
		},
		{
			name:        "no hover attributes over the keyboard's choice",
			noHoverAttr: true,
			events:      []term.Event{hoverTestMove(hoverTestYesX)},
			wantHovered: 0,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name:        "no hover attributes after the pointer leaves the options",
			noHoverAttr: true,
			events:      []term.Event{hoverTestMove(hoverTestNoX), hoverTestMove(9)},
			wantHovered: -1,
			wantYes:     hoverTestOptionAttr, wantNo: hoverTestHighlightAttr,
		},
		{
			name:        "no hover attributes over the message",
			noHoverAttr: true,
			events:      []term.Event{hoverTestMouse(0, hoverTestNoX, 2)},
			wantHovered: -1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name:        "no hover attributes, the keyboard moving off the hovered option",
			noHoverAttr: true,
			events: []term.Event{
				hoverTestMove(hoverTestNoX), {Type: term.EventKey, Key: term.KeyArrowLeft},
			},
			wantHovered: 1,
			wantYes:     hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name:        "no hover attributes, a cancelled press",
			noHoverAttr: true,
			events: []term.Event{
				hoverTestMouse(term.MouseLeft, hoverTestNoX, hoverTestBtnY),
				hoverTestMouse(term.MouseRelease, 9, hoverTestBtnY),
			},
			wantHovered: -1,
			wantYes:     hoverTestOptionAttr, wantNo: hoverTestHighlightAttr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := newHoverTestPrompt(tt.noHoverAttr)
			if !tt.unsized {
				p.Resize(hoverTestWidth, hoverTestHeight)
			}
			for _, ev := range tt.events {
				p.Handle(ev)
			}
			p.Resize(hoverTestWidth, hoverTestHeight)
			assert.Equal(t, tt.wantHovered, p.Hovered())
			yes, no := hoverTestOptionAttrs(t, p)
			assert.Equal(t, tt.wantYes, yes, "Yes")
			assert.Equal(t, tt.wantNo, no, "No")
		})
	}

	t.Run("motion is handled over an option only", func(t *testing.T) {
		for _, tc := range []struct {
			x, y int
			want bool
		}{
			{hoverTestYesX, hoverTestBtnY, true},
			{hoverTestNoX, hoverTestBtnY, true},
			{9, hoverTestBtnY, false},
			{hoverTestYesX, 2, false},
			{-1, hoverTestBtnY, false},
			{hoverTestNoX, hoverTestHeight, false},
		} {
			p, ph := newHoverTestPrompt(false)
			p.Resize(hoverTestWidth, hoverTestHeight)
			exit, handled := p.Handle(hoverTestMouse(0, tc.x, tc.y))
			assert.False(t, exit, "x=%d y=%d", tc.x, tc.y)
			assert.Equal(t, tc.want, handled, "x=%d y=%d", tc.x, tc.y)
			assert.False(t, ph.onSelectCalled, "x=%d y=%d", tc.x, tc.y)
		}
	})

	t.Run("Init forgets the hover", func(t *testing.T) {
		p, _ := newHoverTestPrompt(false)
		p.Resize(hoverTestWidth, hoverTestHeight)
		p.Handle(hoverTestMove(hoverTestNoX))
		require.Equal(t, 1, p.Hovered())
		p.Init(p.cfg)
		assert.Equal(t, -1, p.Hovered())
		p.Resize(hoverTestWidth, hoverTestHeight)
		yes, no := hoverTestOptionAttrs(t, p)
		assert.Equal(t, hoverTestHighlightAttr, yes)
		assert.Equal(t, hoverTestOptionAttr, no)
	})

	t.Run("enter selects the option the pointer highlighted", func(t *testing.T) {
		for _, tc := range []struct {
			name        string
			noHoverAttr bool
			wantIdx     int
		}{
			{name: "no hover attributes", noHoverAttr: true, wantIdx: 1},
			{name: "hover attributes", wantIdx: 0},
		} {
			t.Run(tc.name, func(t *testing.T) {
				p, ph := newHoverTestPrompt(tc.noHoverAttr)
				p.Resize(hoverTestWidth, hoverTestHeight)
				p.Handle(hoverTestMove(hoverTestNoX))
				exit, handled := p.Handle(term.Event{Type: term.EventKey, Key: term.KeyEnter})
				assert.True(t, exit)
				assert.True(t, handled)
				assert.True(t, ph.onSelectCalled)
				assert.Equal(t, tc.wantIdx, ph.selectedIdx)
			})
		}
	})
}

func TestPromptPointerLeaves(t *testing.T) {
	tests := []struct {
		name        string
		noHoverAttr bool
		x, y        int
		wantYes     term.Attributes
		wantNo      term.Attributes
	}{
		{
			name: "to the left", x: -1, y: hoverTestBtnY,
			wantYes: hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name: "to the right", x: hoverTestWidth, y: hoverTestBtnY,
			wantYes: hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name: "above", x: hoverTestNoX, y: -1,
			wantYes: hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name: "below", x: hoverTestNoX, y: hoverTestHeight,
			wantYes: hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name: "far off", x: -100, y: -100,
			wantYes: hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name:        "no hover attributes keep the highlight where it went",
			noHoverAttr: true, x: -1, y: hoverTestBtnY,
			wantYes: hoverTestOptionAttr, wantNo: hoverTestHighlightAttr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := newHoverTestPrompt(tt.noHoverAttr)
			p.Resize(hoverTestWidth, hoverTestHeight)
			p.Handle(hoverTestMove(hoverTestNoX))
			require.Equal(t, 1, p.Hovered())
			exit, handled := p.Handle(hoverTestMouse(0, tt.x, tt.y))
			assert.False(t, exit)
			assert.False(t, handled)
			assert.Equal(t, -1, p.Hovered())
			yes, no := hoverTestOptionAttrs(t, p)
			assert.Equal(t, tt.wantYes, yes, "Yes")
			assert.Equal(t, tt.wantNo, no, "No")
		})
	}
}

func TestPromptClickedAttr(t *testing.T) {
	press := func(x int) term.Event { return hoverTestMouse(term.MouseLeft, x, hoverTestBtnY) }
	release := func(x int) term.Event { return hoverTestMouse(term.MouseRelease, x, hoverTestBtnY) }
	tests := []struct {
		name          string
		noHoverAttr   bool
		noClickedAttr bool
		events        []term.Event
		wantYes       term.Attributes
		wantNo        term.Attributes
	}{
		{
			name:    "a press on an option",
			events:  []term.Event{press(hoverTestNoX)},
			wantYes: hoverTestHighlightAttr, wantNo: hoverTestClickedAttr,
		},
		{
			name:    "a press on the keyboard's choice",
			events:  []term.Event{press(hoverTestYesX)},
			wantYes: hoverTestClickedAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name:    "a press off the options",
			events:  []term.Event{press(9)},
			wantYes: hoverTestHighlightAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name:    "a press dragged onto the other option",
			events:  []term.Event{press(hoverTestNoX), hoverTestMove(hoverTestYesX)},
			wantYes: hoverTestHoverAttr, wantNo: hoverTestClickedAttr,
		},
		{
			name:    "a press dragged off the prompt",
			events:  []term.Event{press(hoverTestNoX), press(-1)},
			wantYes: hoverTestHighlightAttr, wantNo: hoverTestClickedAttr,
		},
		{
			name:    "a cancelled press",
			events:  []term.Event{press(hoverTestNoX), release(hoverTestYesX)},
			wantYes: hoverTestHoverAttr, wantNo: hoverTestOptionAttr,
		},
		{
			name:    "a selecting release",
			events:  []term.Event{press(hoverTestNoX), release(hoverTestNoX)},
			wantYes: hoverTestHighlightAttr, wantNo: hoverTestHoverAttr,
		},
		{
			name:        "no hover attributes",
			noHoverAttr: true,
			events:      []term.Event{press(hoverTestNoX)},
			wantYes:     hoverTestOptionAttr, wantNo: hoverTestClickedAttr,
		},
		{
			name:        "no hover attributes, a selecting release",
			noHoverAttr: true,
			events:      []term.Event{press(hoverTestNoX), release(hoverTestNoX)},
			wantYes:     hoverTestOptionAttr, wantNo: hoverTestHighlightAttr,
		},
		{
			name:          "no clicked attributes",
			noClickedAttr: true,
			events:        []term.Event{press(hoverTestNoX)},
			wantYes:       hoverTestHighlightAttr, wantNo: hoverTestHighlightAttr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _ := newHoverTestConfig(tt.noHoverAttr)
			if !tt.noClickedAttr {
				cfg.ClickedAttr = hoverTestClickedAttr
			}
			p := NewPrompt(cfg)
			p.Resize(hoverTestWidth, hoverTestHeight)
			for _, ev := range tt.events {
				p.Handle(ev)
			}
			yes, no := hoverTestOptionAttrs(t, p)
			assert.Equal(t, tt.wantYes, yes, "Yes")
			assert.Equal(t, tt.wantNo, no, "No")
		})
	}
}

type testPromptHandler struct {
	onCloseCalled  bool
	onSelectCalled bool
	selectedIdx    int
	selectedOption string
}

func TestPromptButtonImagesState(t *testing.T) {
	tests := []struct {
		name        string
		noHoverAttr bool
		events      []term.Event
		wantYes     string
		wantNo      string
		// wantAttrs is how Yes and No look where their images are not
		// drawn.
		wantAttrs [2]term.Attributes
	}{
		{
			name:    "the keyboard's choice",
			wantYes: "selected", wantNo: "idle",
			wantAttrs: [2]term.Attributes{hoverTestHighlightAttr, hoverTestOptionAttr},
		},
		{
			name:    "hovering an option",
			events:  []term.Event{hoverTestMove(hoverTestNoX)},
			wantYes: "selected", wantNo: "hover",
			wantAttrs: [2]term.Attributes{hoverTestHighlightAttr, hoverTestHoverAttr},
		},
		{
			name:    "hovering the keyboard's choice",
			events:  []term.Event{hoverTestMove(hoverTestYesX)},
			wantYes: "hover", wantNo: "idle",
			wantAttrs: [2]term.Attributes{hoverTestHoverAttr, hoverTestOptionAttr},
		},
		{
			name: "pressing an option",
			events: []term.Event{
				hoverTestMouse(term.MouseLeft, hoverTestNoX, hoverTestBtnY),
			},
			wantYes: "selected", wantNo: "pressed",
			wantAttrs: [2]term.Attributes{hoverTestHighlightAttr, hoverTestClickedAttr},
		},
		{
			name:        "hovering without a HoverAttr",
			noHoverAttr: true,
			events:      []term.Event{hoverTestMove(hoverTestNoX)},
			wantYes:     "idle", wantNo: "hover",
			wantAttrs: [2]term.Attributes{hoverTestOptionAttr, hoverTestHighlightAttr},
		},
		{
			name:        "leaving without a HoverAttr keeps the keyboard's choice",
			noHoverAttr: true,
			events:      []term.Event{hoverTestMove(hoverTestNoX), hoverTestMove(9)},
			wantYes:     "idle", wantNo: "selected",
			wantAttrs: [2]term.Attributes{hoverTestOptionAttr, hoverTestHighlightAttr},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _ := newHoverTestConfig(tt.noHoverAttr)
			cfg.ClickedAttr = hoverTestClickedAttr
			states := map[term.ImageID]string{}
			image := func(name string) component.ButtonImage {
				img := &stateImage{id: term.NewImageID()}
				states[img.id] = name
				return component.ButtonImage{Left: img}
			}
			cfg.ButtonImages = &component.ButtonImages{
				Idle:     image("idle"),
				Selected: image("selected"),
				Hover:    image("hover"),
				Pressed:  image("pressed"),
			}
			p := NewPrompt(cfg)
			p.Resize(hoverTestWidth, hoverTestHeight)
			for _, ev := range tt.events {
				p.Handle(ev)
			}

			rec := &stateRecorder{}
			p.Draw(rec)
			got := map[int]string{}
			for _, img := range rec.images {
				got[img.Pos.X] = states[img.ID]
			}
			assert.Equal(t, map[int]string{
				hoverTestYesX: tt.wantYes, hoverTestNoX: tt.wantNo,
			}, got)

			yes, no := hoverTestOptionAttrs(t, p)
			assert.Equal(t, tt.wantAttrs, [2]term.Attributes{yes, no})
		})
	}
}

// stateImage places an image over its cells.
type stateImage struct {
	id            term.ImageID
	width, height int
}

func (c *stateImage) Resize(width, height int) { c.width, c.height = width, height }

func (c *stateImage) Draw(w term.Writer) {
	w.DrawImage(term.Image{ID: c.id, Width: c.width, Height: c.height})
}

// stateRecorder draws every image, and records them.
type stateRecorder struct {
	term.NoopWriter
	images []term.Image
}

func (w *stateRecorder) DrawImage(img term.Image) bool {
	w.images = append(w.images, img)
	return true
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

// A hover test prompt is 20x10 with frameless "Yes" and "No" sharing
// the bottom row: "Yes" at x=5..7 and "No" at x=12..14.
const (
	hoverTestWidth  = 20
	hoverTestHeight = 10
	hoverTestYesX   = 5
	hoverTestNoX    = 12
	hoverTestBtnY   = hoverTestHeight - 1
)

var (
	hoverTestOptionAttr    = term.Attributes{Fg: term.ColorWhite}
	hoverTestHighlightAttr = term.Attributes{Bg: term.ColorYellow}
	hoverTestHoverAttr     = term.Attributes{Fg: term.ColorBlack, Bg: term.ColorSilver}
	hoverTestClickedAttr   = term.Attributes{Fg: term.ColorWhite, Bg: term.ColorFuchsia}
)

func newHoverTestPrompt(noHoverAttr bool) (*Prompt, *testPromptHandler) {
	cfg, ph := newHoverTestConfig(noHoverAttr)
	return NewPrompt(cfg), ph
}

func newHoverTestConfig(noHoverAttr bool) (PromptConfig, *testPromptHandler) {
	ph := new(testPromptHandler)
	cfg := PromptConfig{
		PromptConfig: component.PromptConfig{
			Message: "Ok?",
			Options: []string{"Yes", "No"},
		},
		PromptHandler: ph,
		OptionAttr:    hoverTestOptionAttr,
		HighlightAttr: hoverTestHighlightAttr,
		HoverAttr:     hoverTestHoverAttr,
	}
	if noHoverAttr {
		cfg.HoverAttr = term.Attributes{}
	}
	return cfg, ph
}

// hoverTestMove moves the pointer to x on the row of options.
func hoverTestMove(x int) term.Event {
	return hoverTestMouse(0, x, hoverTestBtnY)
}

func hoverTestMouse(key term.Key, x, y int) term.Event {
	return term.Event{Type: term.EventMouse, Key: key, MouseX: x, MouseY: y}
}

// hoverTestOptionAttrs draws p and returns how "Yes" and "No" look.
func hoverTestOptionAttrs(t *testing.T, p *Prompt) (yes, no term.Attributes) {
	t.Helper()
	w := term.NewStringWriter(hoverTestWidth, hoverTestHeight)
	p.Draw(w)
	cells := w.Cells()
	yesCell := cells[hoverTestBtnY*hoverTestWidth+hoverTestYesX]
	noCell := cells[hoverTestBtnY*hoverTestWidth+hoverTestNoX]
	require.Equal(t, "Y", string(yesCell.Ch))
	require.Equal(t, "N", string(noCell.Ch))
	return yesCell.Attributes(), noCell.Attributes()
}
