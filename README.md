# rune-go-sdk

The Go SDK for building **extensions** for [Rune](https://rune.build).
Extensions are programs that plug into Rune and drive it over gRPC: register
commands, control the editor, react to editor events, manage windows and tabs,
persist data, resolve files, call language servers, talk to LLMs, and more.

> Extensions are the richest way to add new functionality to Rune, but if all
> you need is simple one-shot functionality that can be packaged as a TUI or
> CLI, you can also use this SDK for that. See [Beyond extensions](#beyond-extensions).

```
import "github.com/unstablebuild/rune-go-sdk"
```

## Installation

```bash
go get github.com/unstablebuild/rune-go-sdk
```

Requires Go 1.25 or newer.

## What is an extension?

An extension is a standalone executable that Rune launches as a child process
and connects to over a local gRPC socket. The SDK handles the handshake for you:

1. You describe your extension with `Metadata`: its id, version, and the
   `Permissions` it needs. `ServeWorkspaceExtension` writes that to Rune and
   reads back the connection config.
2. Your setup function receives a `*Workspace`, whose accessors
   (`Storage`, `Editor`, `FileSystem`, `WindowManager`, `Notifications`, `LSP`,
   `LLM`, `Debugger`, and more) are gRPC clients into the host. Every call is
   gated on a `Permission*` you declared, so an undeclared capability is
   rejected.
3. You wire capabilities, register commands and event handlers, and **return**.
   `ServeWorkspaceExtension` owns the lifetime and blocks until shutdown.

## Quick Start

The example below is the wiring half of [`examples/snippets`](examples/snippets),
a complete extension that adds a `snippets` command for storing and reusing
text. `main` does only negotiation, metadata, and wiring. The logic lives in a
handler typed against SDK interfaces so it stays testable.

```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/unstablebuild/rune-go-sdk/api/config"
	"github.com/unstablebuild/rune-go-sdk/api/extensionapi"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
)

func main() {
	meta := extensionapi.Metadata{
		DeveloperID:      "rune-sdk-examples",
		DeveloperKey:     "1234",
		DeveloperEmail:   "your@email.com",
		ExtensionID:      "snippets",
		ExtensionName:    "Snippets",
		ExtensionVersion: "0.1.0",
		Permissions: extensionapi.NewPermissions(
			extensionapi.PermissionStorage,
			extensionapi.PermissionEditor,
			extensionapi.PermissionCommands,
			extensionapi.PermissionFileSystem,
			extensionapi.PermissionBrowserResourceOpener,
			extensionapi.PermissionBrowserWindowManager,
			extensionapi.PermissionNotifications,
		),
	}

	if err := extensionapi.ServeWorkspaceExtension(
		extensionapi.FuncWorkspaceExtension(run), meta,
	); err != nil {
		slog.Error("snippets exited", "error", err)
		os.Exit(1)
	}
}

// run wires capabilities, subscribes to editor events, and registers the
// command, then returns. ServeWorkspaceExtension owns the lifetime.
func run(ctx context.Context, ws *extensionapi.Workspace, cfg config.Config) error {
	s := newSnippets(
		ws.Storage(ctx),        // persist snippets across sessions
		ws.Editor(ctx),         // insert bodies at the cursor
		ws.FileSystem(ctx),     // back scratch buffers with real files
		ws.ResourceOpener(ctx), // open those buffers in a tab
		ws.WindowManager(ctx),  // focus the tab that was opened
		ws.Notifications(ctx),  // report success/failure to the user
	)

	events := []textapi.EventType{
		textapi.EventTypeSelection,
		textapi.EventTypeFlush,
		textapi.EventTypeClose,
	}
	if err := ws.Editor(ctx).SubscribeEvents(events, s); err != nil {
		return fmt.Errorf("subscribe events: %w", err)
	}

	manual := textapi.CommandManual{
		Name:     "snippets",
		Summary:  "Insert, edit, copy, or delete reusable text snippets.",
		Synopsis: "[insert|edit|copy|delete] <name>",
	}
	if err := ws.RegisterCommand(manual, s); err != nil {
		return fmt.Errorf("register command: %w", err)
	}
	return nil
}
```

The handler (`command_handler.go`) implements `textapi.CommandHandler` and
`textapi.EventHandler`: `snippets insert` writes a stored body at the cursor,
`snippets edit`/`copy` open a scratch buffer that is persisted when saved or
closed, and `snippets delete` removes one, all with tab completion over stored
names. See the [`examples/snippets` README](examples/snippets/README.md) for the
full walkthrough and the patterns worth copying, and the
[extension SDK guide](https://docs.rune.build/develop/sdk) for the complete
authoring guide.

### Patterns worth copying

- **Thin `main`, testable handler.** Keep `main`/`run` to wiring; put logic in a
  type that takes SDK interfaces, so tests can substitute fakes
  (`storagestub.NewInMemoryService()` and small in-file fakes for the rest).
- **Return wrapped errors; let the host notify.** Command handlers return
  `fmt.Errorf(...)` and Rune surfaces the message. Call `Notifications` directly
  only when you return `nil` but still want to tell the user something.
- **Declare every permission you use.** A call into a capability you did not
  declare in `Metadata.Permissions` is rejected by the host.
- **`run` returns after setup.** Do not block; `ServeWorkspaceExtension` blocks
  on its own and cancels `ctx` on shutdown for any in-flight work.

## Extension API

The `extensionapi` package handles the handshake and hands you a `*Workspace`.
The rest of the `api/` packages wrap Rune's gRPC services, so each `Workspace`
accessor returns a typed client into the host:

| Package | Capability | Accessor |
| --- | --- | --- |
| `extensionapi` | Handshake, `Workspace`, command registration | (none) |
| `textapi` | Editor and text operations, editor events, commands | `Editor`, `RegisterCommand` |
| `storageapi` | Document storage and persistence | `Storage` |
| `browserapi` | Windows, tabs, resource opening, notifications | `WindowManager`, `ResourceOpener`, `Notifications` |
| `workspaceapi` | URI resolution, file operations, watchers, command execution | `FileSystem`, `Executor` |
| `semanticapi` | Language-server operations (definition, references, rename, diagnostics, and more) | `LSP` |
| `llmapi` | LLM access (models, token counting, messages) | `LLM` |
| `debugapi` | Debug-adapter control | `Debugger` |
| `syntaxapi` | Tree-sitter structural search and queries | `Parser` |
| `config` | Merged host configuration | `Config` |
| `schemeapi` | URI scheme registration | (none) |

Every accessor is gated on the matching `Permission*` you declared in
`Metadata`, so an undeclared capability is rejected by the host. Each API
package contains an `*rpc/` subdirectory with the `.proto` definitions,
generated message and gRPC stubs, and the Go code wrapping the RPC layer.

## Running and debugging your extension

Rune launches an extension as a child process and connects to it over a private
local socket, so you never run your binary directly during development. Build it,
then start it in a workspace from the [Rune console](https://docs.rune.build/learn/console)
with the `extensions` command:

```
extensions start snippets /path/to/snippets --config '{"key":"value"}'
```

`extensions start <id> <cmdAndArgs> [--config <json>]` runs your freshly built
binary as the extension `id`, so you can iterate without publishing a package.
The rest of the lifecycle is managed from the same console command:

- `extensions status` lists every workspace extension with its status, pid, and uptime.
- `extensions info <id>` shows detailed state, including restart counts.
- `extensions logs <id> [--tail <N>]` prints what your extension wrote to stderr, where startup failures and crashes show up. Use `slog`/`log` to write there.
- `extensions restart <id>` relaunches with the current config, the fastest way to pick up a rebuild.
- `extensions stop <id>` stops it.

Permissions are enforced at two layers: a call into a capability you did not
declare in `Metadata.Permissions` is rejected outright, and the first time your
extension uses a declared capability Rune prompts the user to allow or deny it.
If a capability seems unreachable, check `authorizer list` for a lingering
**deny always** decision and clear it with `authorizer revoke <permission>`.
Once you are ready to distribute, ship the extension as a package so users can
`pkg install` it. See the [extensions guide](https://docs.rune.build/develop/extensions)
for the full workflow.

## Beyond extensions

Extensions are the richest way to add new functionality to Rune, but if all you
need is simple one-shot functionality which can be packaged as a TUI or CLI,
you can also use this SDK for that.

### Plugins (in-terminal TUI)

Plugins are text-based applications that run inside Rune's virtual terminal
emulator, built from a composable component/handler/TUI framework in `tui/`,
`term/`, `component/`, `handler/`, and `iterator/`. The framework is built around
three core abstractions (defined in `tui/`):

1. **Component**: a basic UI element that can be drawn and resized.
2. **Handler**: a component that can also handle keyboard and mouse events and manage the cursor/selection.
3. **Event loop**: polls terminal events, routes them to handlers, redraws, and flushes to the terminal.

Components compose through small, focused interfaces rather than inheritance.
Optional capability interfaces (`Responsive`, `Scrollable`, `Floating`) let
collection components lay out their children, and most components offer a
zero-value-safe constructor (`NewX()`) plus a configurable variant
(`NewXWithConfig(cfg)`). See [`examples/repl`](examples/repl/main.go) and
[`examples/dialogue`](examples/dialogue/main.go) for runnable plugins.

### runectl (CLI)

`runectl` is a CLI for driving Rune from the shell. It is typically invoked from
within Rune's plugin commands (`!`/`!!`), which pass authentication to runectl.
It works regardless of the language you write extensions in, so it is worth
reaching for from any of the SDKs.

Most users install it from the [Rune console](https://docs.rune.build/learn/console),
which drops it on your `PATH` inside Rune's terminals and targets the current
workspace automatically:

```
pkg install runectl
```

To build it from this repo instead:

```bash
go install github.com/unstablebuild/rune-go-sdk/cmd/runectl@latest
runectl --help
```

Command groups include:

- `wm`: window management (focus, split, floating, set-content, close).
- `editor`: read/modify editor content, cursors, colors, and locations.
- `storage`: create/get/set/update/delete/list stored documents.
- `lsp`: language-server operations (definition, references, hover, rename, diagnostics, completion, and many more).
- `syntax`: tree-sitter structural search and queries.
- `llm`: list models, inspect model info, count tokens, send messages.
- `mcp`: A full MCP server to expose Rune's tools to other agent harnesses like Claude Code or Opencode.
- `open`, `notify`, `uri`, `datadir`, `signal`, `exec`: miscellaneous host operations.

`runectl` is also the fastest way to **sample the data Rune returns** for each
API while you build an extension. Every command wraps the same gRPC services the
`api/` packages expose, and most accept a `-F`/`--format` flag (`table`, `json`,
or a Go template). Running the equivalent `runectl` command by hand lets you see
the exact shape of a response before you write a line of Go against it, which
makes iterating and debugging an extension much faster:

```bash
# Inspect what the storage API returns for your extension's documents.
runectl storage list --format json

# See the exact LSP payload before coding against semanticapi.
runectl lsp hover term.Writer --format json

# Check diagnostics the way your extension would receive them.
runectl lsp workspace-diagnostics --format json
```

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
