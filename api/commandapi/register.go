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

package commandapi

// CommandRegistry abstracts the ability to register new commands.
type CommandRegistry interface {
	// RegisterCommand registers a command to be dispatched
	// to CommandHandler. Registered commands appear in the
	// command prompt. Use RegisterCommand for editing and
	// live-programming actions that operate on the current
	// file and need no persistent output, such as
	// navigating to a symbol definition, toggling a fold, or
	// reformatting a selection.
	RegisterCommand(CommandManual, CommandHandler) error

	// RegisterREPLCommand registers a REPL command to be
	// dispatched to REPLHandler. Registered commands appear
	// in the IDE's shell. Use RegisterREPLCommand for
	// configuration, monitoring and troubleshooting commands
	// that produce inspectable output the user wants to
	// review, such as a debugger, a log viewer, or a status
	// dashboard.
	RegisterREPLCommand(CommandManual, REPLHandler) error
}
