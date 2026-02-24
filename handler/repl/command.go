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

package repl

import (
	"context"

	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

// Command represents a user-issued command.
type Command struct {
	Name string
	Args []string
}

// CommandHandler dispatches commands and provides
// completions.
type CommandHandler interface {
	// HandleCommand is called when the user submits
	// a command. It returns an iterator of responsive
	// components to display as output.
	HandleCommand(ctx context.Context, cmd Command) (
		iterator.Iterator[component.Responsive], error,
	)

	// Complete returns tab completion candidates for
	// the given command and arguments.
	Complete(ctx context.Context, cmd string, args []string) (
		iterator.Iterator[string], error,
	)
}

// FuncCommandHandler returns a CommandHandler that
// calls fn every time HandleCommand is invoked and
// completer when Complete is invoked.
func FuncCommandHandler(
	fn func(context.Context, Command) (iterator.Iterator[component.Responsive], error),
	completer func(context.Context, string, []string) (iterator.Iterator[string], error),
) CommandHandler {
	return fnCommandHandler{
		cb:         fn,
		completeFn: completer,
	}
}

// NopCommandCompleter returns a CommandHandler that
// calls fn every time HandleCommand is invoked, but
// does not have a completion function.
func NopCommandCompleter(
	fn func(context.Context, Command) (iterator.Iterator[component.Responsive], error),
) CommandHandler {
	return fnCommandHandler{cb: fn}
}

type fnCommandHandler struct {
	cb         func(context.Context, Command) (iterator.Iterator[component.Responsive], error)
	completeFn func(context.Context, string, []string) (iterator.Iterator[string], error)
}

func (f fnCommandHandler) HandleCommand(
	ctx context.Context, c Command,
) (iterator.Iterator[component.Responsive], error) {
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
