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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unstablebuild/rune-go-sdk/component/comptest"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/tcell/v3"
)

func testDrawPrompt(t *testing.T, cfg PromptConfig, expectedOut string) {
	s := NewPrompt(cfg)

	s.Resize(20, 10)

	w := term.NewStringWriter(21, 11)

	tests := []comptest.TestCase{
		{Expected: expectedOut},
	}

	comptest.TestComponent(t, s, w, tests)
}

func TestDrawPrompt(t *testing.T) {
	t.Run("no frame", func(t *testing.T) {
		testDrawPrompt(t, PromptConfig{
			Message: "Do you?",
			Options: []string{"Yay", "Nay"},
		}, `                     
                     
      Do you?        
                     
                     
   Yay       Nay     
                     
                     
                     
                     
                     `,
		)
	})

	t.Run("with frame", func(t *testing.T) {
		testDrawPrompt(t, PromptConfig{
			Message: "Do you?",
			Options: []string{"Yay", "Nay"},
			Frame:   FrameCharSetDefault(),
		}, `                     
                     
      Do you?        
                     
                     
 ┌─────┐   ┌─────┐   
 │ Yay │   │ Nay │   
 └─────┘   └─────┘   
                     
                     
                     `,
		)
	})

	t.Run("with frame overflow options", func(t *testing.T) {
		testDrawPrompt(t, PromptConfig{
			Message: "Why soooooo serious?",
			Options: []string{"Yay", "Nay", "Say", "Wey"},
			Frame:   FrameCharSetDefault(),
		}, `                     
                     
    Why soooooo      
    serious?         
                     
                     
┌───┐┌───┐┌───┐┌───┐ 
│ Y ││ N ││ S ││ W │ 
│ a ││ a ││ a ││ e │ 
└───┘└───┘└───┘└───┘ 
                     `,
		)
	})

	t.Run("with frame overflow options", func(t *testing.T) {
		testDrawPrompt(t, PromptConfig{
			Message: "Why soooooo serious?",
			Options: []string{"Yay", "Nay", "Say", "Wey", "They", "May"},
		}, `                     
                     
    Why soooooo      
    serious?         
                     
                     
 Y  N  S  W  T  M    
 a  a  a  e  h  a    
 y  y  y  y  e  y    
             y       
                     `,
		)
	})
}

func TestPromptSetOptionAttr(t *testing.T) {
	t.Run("does not panic with frame", func(t *testing.T) {
		p := NewPrompt(PromptConfig{
			Message: "Wasup?",
			Options: []string{"Meh", "Bleh"},
			Frame:   FrameCharSetDefault(),
		})
		p.SetOptionAttr(0, term.Attributes{})
		p.SetOptionAttr(1, term.Attributes{})
	})

	t.Run("does not panic without frame", func(t *testing.T) {
		p := NewPrompt(PromptConfig{
			Message: "Wasup?",
			Options: []string{"Meh", "Bleh"},
		})
		p.SetOptionAttr(0, term.Attributes{})
		p.SetOptionAttr(1, term.Attributes{})
	})
}

func TestPromptDefaults(t *testing.T) {
	t.Run("panics on empty message", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = NewPrompt(PromptConfig{
				Message: "", Options: []string{"a"},
			})
		})
	})
	t.Run("panics on empty options", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = NewPrompt(PromptConfig{
				Message: "blah", Options: []string{},
			})
		})
	})
}

func TestPromptInitReset(t *testing.T) {
	var p Prompt
	opts := make(map[string]*TestResponsive)
	makeTestOption := func(opt string, cfg PromptConfig) responsiveWithAttributes {
		t := testResponsive(([]rune)(opt)[0], 10)
		opts[opt] = t
		return t
	}
	p.init(makeTestOption, PromptConfig{
		Message: "?",
		Options: []string{"Y", "N"},
	})
	p.init(makeTestOption, PromptConfig{
		Message: "?!",
		Options: []string{"y", "n"},
	})

	t.Run("Draw", func(t *testing.T) {
		p.Resize(20, 10)
		w := term.NewStringWriter(21, 11)

		tests := []comptest.TestCase{
			{Expected: `
                     
                     
         ?!          
                     
                     
yyyyyyyyyynnnnnnnnnn 
yyyyyyyyyynnnnnnnnnn 
yyyyyyyyyynnnnnnnnnn 
yyyyyyyyyynnnnnnnnnn 
yyyyyyyyyynnnnnnnnnn 
                     `,
			},
		}
		comptest.TestComponent(t, &p, w, tests)
	})

	t.Run("SetAttr", func(t *testing.T) {
		attr := term.Attributes{Fg: tcell.ColorRed, Attrs: tcell.AttrBold}
		p.SetOptionAttr(0, attr)
		assert.Equal(t, attr, opts["y"].Attributes)
	})
}
