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
	"context"
	"errors"
	"log"
	"strings"

	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/handler/inputbox"
	"github.com/unstablebuild/rune-go-sdk/handler/repl"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

var (
	commands = []string{
		"clear", "echo", "exit",
		"help", "history", "quit",
	}
	errExit = errors.New("exit requested")
)

type interpHandler struct{}

func (interpHandler) HandleCommand(
	_ context.Context, cmd repl.Command,
) (iterator.Iterator[component.Responsive], error) {
	switch cmd.Name {
	case "help":
		return stringIter("Commands: " + strings.Join(commands, ", ")), nil
	case "echo":
		return stringIter(strings.Join(cmd.Args, " ")), nil
	case "history":
		return stringIter("(use Up/Down to browse history)"), nil
	case "quit", "exit":
		return nil, errExit
	default:
		return nil, errors.New("Unknown: " + cmd.Name + ". Try 'help'.")
	}
}

func (interpHandler) Complete(
	_ context.Context, cmd string, arg []string,
) (iterator.Iterator[string], error) {
	if len(arg) >= 1 {
		return iterator.Empty[string](), nil
	}
	prefix := cmd
	var matches []string
	for _, c := range commands {
		if strings.HasPrefix(c, prefix) {
			matches = append(matches, c)
		}
	}
	return iterator.FromSlice(matches), nil
}

func stringIter(s string) iterator.Iterator[component.Responsive] {
	resp := component.NewResponsiveString(s, component.StringResponsiveConfig{})
	return iterator.FromSlice([]component.Responsive{resp})
}

func main() {
	r := repl.New(
		interpHandler{},
		term.ScheduleNextTick,
		term.FuncInterrupter(func(context.Context) error {
			if !term.PublishEvent(term.Event{Type: term.EventInterrupt}) {
				return errors.New("could not publish interrupt")
			}
			return nil
		}),
		repl.WithPrompt("interp> "),
		repl.WithTabStyle(inputbox.TabPrints),
		repl.WithExitError(errExit),
	)
	if err := tui.Run(r); err != nil {
		log.Fatal(err)
	}
}
