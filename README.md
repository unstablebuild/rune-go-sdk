# rune-go-sdk

A Go SDK for extending the [Rune IDE](https://github.com/unstablebuild). Use it to build:

- **Plugins** — text-based applications that run inside Rune's virtual terminal emulator, built from a composable component/handler/TUI framework.
- **Extensions** — programs that interact with Rune over gRPC on a local socket (window management, workspace/file operations, editor control, storage, LLM access, and more).
- **runectl** — a CLI for driving Rune from the shell or from within Rune's plugin commands.

```
import "github.com/unstablebuild/rune-go-sdk"
```

## Installation

```bash
go get github.com/unstablebuild/rune-go-sdk
```

Requires Go 1.25 or newer.

## Quick Start

The example below (from [`examples/repl`](examples/repl/main.go)) builds a small REPL plugin: an interpreter that handles commands and tab-completion, wired into a `repl` handler and run with `tui.Run`.

```go
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
	_ context.Context, cmd repl.Command, _ repl.ProgressWriter,
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
```

See [`examples/`](examples) for complete, runnable programs:

- **`examples/dialogue`** — a chat-like interface built with `inputbox` and `Container`.
- **`examples/repl`** — a read-eval-print loop.

## Architecture

### Component / Handler / TUI model

The plugin framework is built around three core abstractions (defined in `tui/`):

1. **Component** — a basic UI element that can be drawn and resized.
2. **Handler** — a component that can also handle keyboard and mouse events and manage the cursor/selection.
3. **Event loop** — polls terminal events, routes them to handlers, redraws, and flushes to the terminal.

Components compose through small, focused interfaces rather than inheritance. Optional capability interfaces let collection components lay out their children:

- **`Responsive`** — reports a height *hint* for a given width and wraps/scrolls/truncates content to fit the space it is actually given.
- **`Scrollable`** — lets the runtime render a scroll bar for components that scroll their content vertically.
- **`Floating`** — exposes a component's ideal width and height when its content size is known ahead of time.

Most components offer a zero-value-safe constructor (`NewX()`) plus a configurable variant (`NewXWithConfig(cfg)`).

### Packages

| Package | Description |
| --- | --- |
| `tui/` | Core framework interfaces and the event loop. |
| `term/` | Terminal I/O abstraction (writers, event types, cursor styles, colors, attributes). |
| `component/` | UI components — String, Container, Row, Frame, Divider, List, Grid, Overlay, Prompt, Async, and more. |
| `handler/` | Event-handling patterns and component wrappers (Nop, Sync, Virtual, Frame, Floating, Responsive, Scrollable, inputbox, …). |
| `iterator/` | Generic iteration utilities (Filter, Map, Reduce, Aggregate). |
| `api/` | gRPC-based integrations with the Rune host (see below). |
| `cmd/runectl/` | The `runectl` CLI. |

### API packages

The `api/` packages wrap Rune's gRPC services so extensions can drive the host:

- **`browserapi`** — window management, notifications, and tabs.
- **`workspaceapi`** — URI resolution, file operations, and watchers.
- **`textapi`** — editor and text operations.
- **`storageapi`** — document storage and persistence.
- **`extensionapi`** — the extension/plugin system and workspace connection.
- **`schemeapi`** — URI scheme registration.
- **`config`** — configuration management.
- **`debugapi`**, **`llmapi`**, **`semanticapi`**, **`syntaxapi`** — debugging, LLM access, semantic (language-server) operations, and syntax/tree-sitter queries.

Each API package contains an `*rpc/` subdirectory with the `.proto` definitions, generated message and gRPC stubs, and Go code wrapping the RPC layer.

## runectl

`runectl` is a CLI for interacting with Rune. It is typically invoked from within Rune's plugin commands (`!`/`!!`), where the `RUNE_SOCKET` and `RUNE_DATADIR` environment variables are set.

```bash
runectl --help
```

Command groups include:

- `wm` — window management (focus, split, floating, set-content, close).
- `editor` — read/modify editor content, cursors, colors, and locations.
- `storage` — create/get/set/update/delete/list stored documents.
- `lsp` — language-server operations (definition, references, hover, rename, diagnostics, completion, and many more).
- `syntax` — tree-sitter structural search and queries.
- `llm` — list models, inspect model info, count tokens, send messages.
- `mcp` — Model Context Protocol helpers.
- `open`, `notify`, `uri`, `datadir`, `signal`, `exec` — miscellaneous host operations.

## Development

Common Make targets:

```bash
make              # Build examples and executables
make debug        # Build with the race detector

make test         # Run all tests with the race detector
make test-no-race # Run tests without the race detector
make coverage     # Generate an HTML coverage report

make lint         # Run golangci-lint
make format       # Format code with go fmt

make generate     # Regenerate protobuf/gRPC code
make license      # Add Apache 2.0 license headers
```

Run the full pre-commit validation before committing:

```bash
make generate && make lint && make test
```

Tests follow a table-driven approach. Components are tested with the `comptest` package and handlers with the `handlertest` package.

## License

Apache License 2.0. See [LICENSE](LICENSE).
