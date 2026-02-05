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

package main

import (
	"fmt"
	"log"

	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/handler"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
	"github.com/unstablebuild/tcell/v3"
)

// ChatDemo demonstrates a simple chat-like interface using InputBox
type ChatDemo struct {
	input         *handler.Virtual[*component.InputBox]
	messages      []string
	container     *component.Container
	rows          []*component.Row
	width, height int
}

func NewChatDemo() *ChatDemo {
	inputBox := component.NewInputBoxWithAttrs(term.Attributes{
		Fg: tcell.ColorWhite,
		Bg: tcell.ColorBlack,
	})

	cd := &ChatDemo{
		input: &handler.Virtual[*component.InputBox]{
			Virtual: component.Virtual[*component.InputBox]{
				C: inputBox,
			},
		},
		messages: []string{
			"Welcome to the chat demo!",
			"Type a message and press Enter to send.",
			"Press Esc to exit.",
			"",
		},
		container: component.NewContainer(),
	}

	// Pre-create rows for initial messages
	for _, msg := range cd.messages {
		row := cd.container.AddRow()
		row.AddComponent(
			component.NewResponsiveString(
				msg,
				component.StringResponsiveConfig{},
			),
			component.MaxCols,
		)
		cd.rows = append(cd.rows, row)
	}

	return cd
}

func (cd *ChatDemo) Resize(width, height int) {
	cd.width = width
	cd.height = height

	// Resize container for messages (all but bottom line)
	cd.container.Resize(width, height-1)

	// Resize and position input box at bottom
	cd.input.Resize(width, 1)
	cd.input.Move(term.Coordinates{X: 0, Y: height - 1})
}

func (cd *ChatDemo) Draw(w term.Writer) {
	cd.container.Draw(w)
	cd.input.Draw(w)
}

func (cd *ChatDemo) Handle(ev term.Event) (exit, handled bool) {
	if ev.Key == term.KeyEnter {
		text := cd.input.C.Text()
		if text == "" {
			return
		}
		
		// Add new message
		msg := fmt.Sprintf("You: %s", text)
		cd.messages = append(cd.messages, msg)

		// Add new row to container
		row := cd.container.AddRow()
		row.AddComponent(
			component.NewResponsiveString(
				msg,
				component.StringResponsiveConfig{},
			),
			component.MaxCols,
		)
		cd.rows = append(cd.rows, row)

		// Re-layout with new message
		cd.container.Resize(cd.width, cd.height-1)

		cd.input.C.Clear()
		return false, true
	}

	return cd.input.Handle(ev)
}

func (cd *ChatDemo) Cursor() (
	term.Coordinates,
	term.CursorStyle,
	bool,
) {
	return cd.input.Cursor()
}

func (cd *ChatDemo) Selection() (string, bool) {
	return cd.input.Selection()
}

func main() {
	demo := NewChatDemo()
	if err := tui.Run(demo); err != nil {
		log.Fatal(err)
	}
}
