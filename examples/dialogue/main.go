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
	"github.com/unstablebuild/rune-go-sdk/handler/inputbox"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// ChatDemo demonstrates a simple chat-like interface using InputBox
type ChatDemo struct {
	input         *handler.Virtual[*inputbox.Handler]
	messages      []string
	container     *component.Container
	rows          []*component.Row
	width, height int
	inputHeight   int
}

func NewChatDemo() *ChatDemo {
	inputBox := inputbox.New(
		inputbox.WithPlaceholderText("Type a message..."),
		inputbox.WithAttributes(term.Attributes{
			Bg: term.ColorGray,
			Fg: term.ColorYellow,
		}),
	)

	cd := &ChatDemo{
		input: &handler.Virtual[*inputbox.Handler]{
			Virtual: component.Virtual[*inputbox.Handler]{
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

	inputHeight := min(height, cd.input.C.Height(width))
	messagesHeight := max(0, height-inputHeight)

	cd.container.Resize(width, messagesHeight)

	cd.input.Resize(width, inputHeight)
	cd.input.Move(term.Coordinates{X: 0, Y: messagesHeight})
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

		msg := fmt.Sprintf("You: %s", text)
		cd.messages = append(cd.messages, msg)

		row := cd.container.AddRow()
		row.AddComponent(
			component.NewResponsiveString(
				msg,
				component.StringResponsiveConfig{},
			),
			component.MaxCols,
		)
		cd.rows = append(cd.rows, row)

		cd.inputHeight = cd.input.C.Height(cd.width)
		cd.Resize(cd.width, cd.height)
		cd.input.C.Clear()
		return false, true
	}

	exit, handled = cd.input.Handle(ev)
	if !handled || exit {
		return
	}

	newHeight := cd.input.C.Height(cd.width)
	if newHeight != cd.inputHeight {
		cd.inputHeight = newHeight
		cd.Resize(cd.width, cd.height)
	}
	return
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
