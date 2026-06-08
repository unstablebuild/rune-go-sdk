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

package textapi

import (
	"context"

	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"github.com/unstablebuild/rune-go-sdk/term"
)

// Command represents a command issued by the user.
type Command struct {
	Name string
	Args []string

	// optional. If command is dispatched while non-tab is in focus,
	// then these fields will be zero-valued.
	URI      workspaceapi.URI
	Resource Handler
	Window   browserapi.Window
	Cursor   struct {
		Content term.Coordinates
		Window  term.Coordinates
	}
}

// CommandHandler is a callback interface that wraps the basic
// method Command.
type CommandHandler interface {
	// Handle is called when user issued a command previously
	// registered via SubscribeCommand.
	HandleCommand(context.Context, Command) (err error)

	// Complete takes command args and returns a list of expanded
	// options for them.
	// It also returns a expanded version of the last arg, or an
	// empty string if the last arg could/should not be
	// automatically expanded.
	Complete(ctx context.Context, cmd string, args []string) (
		iterator.Iterator[string], error,
	)
}

// CommandManual represents a command's manual and documentation.
type CommandManual struct {
	Name string

	// Summary is a short 80-100 character description.
	Summary string

	// Synopsis is a single line synopsis of how
	// this CLI is to be used. It should ONLY include
	// the semantic information about how arguments are parsed.
	//
	// Example: [<options>] [<revision-range>] [[--] <path>...]
	Synopsis string

	// Commands is a list of accepted commands or nil
	// if no commands are expected.
	Commands []CommandManual
}

// FuncCommandHandler returns an CommandHandler that calls fn
// every time HandleCommand is invoked and completer when
// Complete is invoked.
func FuncCommandHandler(
	fn func(context.Context, Command) error,
	completer func(
		context.Context, string, []string,
	) (iterator.Iterator[string], error),
) CommandHandler {
	return fnCommandHandler{
		cb:         fn,
		completeFn: completer,
	}
}

// NopCommandCompleter returns an CommandHandler that calls fn
// every time HandleCommand is invoked, but does not have a
// completion function.
func NopCommandCompleter(
	fn func(context.Context, Command) error,
) CommandHandler {
	return fnCommandHandler{
		cb: fn,
	}
}

type fnCommandHandler struct {
	cb         func(context.Context, Command) error
	completeFn func(
		context.Context, string, []string,
	) (iterator.Iterator[string], error)
}

func (f fnCommandHandler) HandleCommand(
	ctx context.Context, c Command,
) error {
	return f.cb(ctx, c)
}

func (f fnCommandHandler) Complete(
	ctx context.Context, cmd string, args []string,
) (iterator.Iterator[string], error) {
	if f.completeFn != nil {
		return f.completeFn(ctx, cmd, args)
	}
	return iterator.FromSlice[string](nil), nil
}
