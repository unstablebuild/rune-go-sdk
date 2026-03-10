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

	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/handler/repl"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

// REPLHandler extends repl.CommandHandler with a Help
// method that returns documentation for a command.
type REPLHandler interface {
	repl.CommandHandler

	// Help returns responsive components that describe
	// the command's usage for the given arguments.
	Help(ctx context.Context, args []string) (
		iterator.Iterator[component.Responsive], error,
	)
}
