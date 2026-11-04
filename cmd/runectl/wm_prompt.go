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
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
	"github.com/unstablebuild/rune-go-sdk/api/browserapi/browserrpc"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/term/termrpc"
)

var (
	errPromptDismissed = errors.New("prompt dismissed")
	errPromptOpen      = errors.New("a prompt with this message is already open")
)

// promptAnswer is the option the user picked.
type promptAnswer struct {
	Index  int
	Option string
}

func newWMPromptCmd(a *app) *cobra.Command {
	var format string
	var keys string
	var maxWidth int

	cmd := &cobra.Command{
		Use:   "prompt <message> <option>...",
		Short: "Ask the user to pick an option",
		Long: "Ask the user to pick an option in a prompt window, and print the one they pick.\n" +
			"The message is markdown. The command waits for the answer, and fails if the\n" +
			"user dismisses the prompt or one with the same message is already open.\n" +
			"Stopping the command closes the prompt.",
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			cfg := browserapi.PromptConfig{
				Message:  args[0],
				Options:  args[1:],
				MaxWidth: maxWidth,
			}
			if keys != "" {
				bindings, err := term.ParseKeys(keys)
				if err != nil {
					return fmt.Errorf("invalid --keys %q: %w", keys, err)
				}
				cfg.Bindings = bindings
			}
			w, err := a.getWorkspace()
			if err != nil {
				return err
			}
			wm := browserrpc.NewWindowManagerClient(w.RawConn())
			answer, err := prompt(cmd.Context(), wm, cfg)
			if err != nil {
				return err
			}
			return printResult(
				cmd.Context(), format, answer,
				func(v promptAnswer) { fmt.Println(v.Option) },
				[]string{"Index", "Option"},
			)
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)
	cmd.Flags().StringVarP(
		&keys, "keys", "k", "",
		"Keys that pick each option in order, such as yn or <enter><esc>",
	)
	cmd.Flags().IntVarP(
		&maxWidth, "max-width", "w", 0,
		"Size the message to its widest line, wrapping at this many cells",
	)

	return cmd
}

// prompt opens cfg and waits for the user to answer it. The SDK's
// Prompter can't tell a prompt it opened from one already open for the
// same message, so it drives the stream itself.
func prompt(
	ctx context.Context, wm browserrpc.WindowManagerClient, cfg browserapi.PromptConfig,
) (promptAnswer, error) {
	if err := cfg.Validate(); err != nil {
		return promptAnswer{}, err
	}
	req := browserrpc.PromptRequest{
		Message:  cfg.Message,
		Options:  cfg.Options,
		MaxWidth: uint32(cfg.MaxWidth),
	}
	for _, k := range cfg.Bindings {
		ev := new(termrpc.Event)
		if err := ev.FromModel(term.Event{
			Type: term.EventKey, Mod: k.Mod, Key: k.Key, Ch: k.Ch,
		}); err != nil {
			return promptAnswer{}, fmt.Errorf("prompt binding %s: %w", k, err)
		}
		req.Bindings = append(req.Bindings, ev)
	}

	stream, err := wm.Prompt(ctx, &req)
	if err != nil {
		return promptAnswer{}, fmt.Errorf("prompt: %w", err)
	}
	var answer *promptAnswer
	for {
		ev, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return promptAnswer{}, errPromptOpen
		}
		if err != nil {
			return promptAnswer{}, fmt.Errorf("prompt: %w", err)
		}
		if sel := ev.GetSelected(); sel != nil {
			answer = &promptAnswer{Index: int(sel.GetIndex()), Option: sel.GetOption()}
		}
		if ev.GetClosed() != nil {
			if answer == nil {
				return promptAnswer{}, errPromptDismissed
			}
			return *answer, nil
		}
	}
}
