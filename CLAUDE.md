# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Go SDK for extending the Rune IDE by building plugins (text-based applications that run inside the Rune IDE's virtual terminal emulator),
or extensions which are programs that interact with Rune via grpc over a local socket.

## Common Commands

```bash
# Build
make              # Build examples and executables
make debug        # Build with race detection

# Testing
make test         # Run all tests with race detector (120s timeout)
make test-no-race # Run tests without race detection
make coverage     # Generate HTML coverage report

# Code Quality
make lint         # Run golangci-lint (600s timeout)
make format       # Format code with go fmt

# Code Generation & License
make generate     # Regenerate protobuf files
make license      # Add Apache 2.0 license headers to Go files

# Pre-commit validation (run before committing)
make generate && make lint && make test
```

## High-Level Architecture

### Component-Handler-TUI Model

The SDK is built around three core abstractions defined in `tui/tui.go`:

1. **Component** - Basic UI element that can be drawn and resized
2. **Handler** - Component that can handle events (keyboard, mouse) and manage cursor/selection
3. **Event Loop** (`tui/run.go`) - Polls tcell events, routes to handlers, redraws, and flushes to terminal

#### component.Responsive and handler.Responsive
Designed to allow collection components (List, FrameUnion, Container, ResponsiveList, etc.)
to vertically compose their (Responsive) children,
given the collection's width during a call to Resize.
- `Height(width int) int` must return a **hint** — the number
  of lines needed to render content at the given width.
- The component must **not assume** that `Resize` will be called
  with the same height returned by `Height()`.
- If the component can receive less vertical space than it
  needs, it must implement vertical scrolling or truncation.
- Text wrapping must happen at the width boundary (no
  horizontal scrolling for text content).

#### component.Scrollable and handler.Scrollable
Designed to allow the TUI runtime to show a scroll bar next to
the handler/component. It should be implemented by components/handlers
that vertically scroll their content, or collections of Responsive
component/handlers.

#### component.Floating and handler.Floating
Designed to allow components that are inherently capable of calculating
their ideal width and height, because their content size and shape is known
ahead of time via Dimensions, which should not return the width/height passed
in the last Resize, but rather the ideal width and height for this component to
render the entire content. This should be implemented by all components/handlers with
easy to calculate dimensions or when they're collections of Floating components/handlers.

### Package Organization

- **`tui/`** - Core framework interfaces and event loop
- **`term/`** - Terminal I/O abstraction layer (TermboxWriter, VirtualWriter, event types, cursor styles)
- **`component/`** - 48 UI components organized by complexity:
  - Basic: String, Container, Row, Frame, Divider
  - Layout: ResponsiveList, List, FocusList, Grid, FloatingResponsive
  - Behavioral: Background, Overlay, Prompt, Async
- **`handler/`** - Event handling patterns and component wrappers (Nop, Sync, Virtual, Frame, Floating, Responsive, Scrollable)
- **`iterator/`** - Generic iteration utilities with Filter, Map, Reduce, Aggregate operations
- **`api/`** - gRPC-based external integrations:
  - `browserapi/` - Window management, notifications, tabs
  - `workspaceapi/` - URI resolution, file operations, watchers
  - `textapi/` - Editor/text operations
  - `config/` - Configuration management
  - `storageapi/` - Document storage/persistence (largest at ~4000 LOC)
  - `extensionapi/` - Plugin system
  - `schemeapi/` - URI scheme registration

### Key Patterns

**Interface Composition**: Small, focused interfaces combined through composition rather than inheritance.

**Protobuf/gRPC Integration**: API packages contain `*rpc/` subdirectories with:
- `*.proto` - Service definitions
- `*.pb.go` - Generated messages
- `*_grpc.pb.go` - Generated gRPC stubs
- Custom Go implementations wrapping the RPC layer

**Builder Pattern**: Components support zero-value safe constructors (`NewX()`) and full customization (`NewXWithConfig(cfg)`).

**Testing**: Always use a table-driven approach.
  - `tui.Component` (and derived) implementations are tested
    using the `comptest` package. See `component/*_test.go`
    for examples.
  - `tui.Handler` (and derived) implementations are tested
    using the `handlertest` package. See `handler/inputbox/`
    and `handler/` for examples.
* `handlertest.SequenceTestCase.InputSequence` is a **concatenation of tokens** (no separators). Each token becomes one `KeyComb`.
* Token forms:
  1. **Single character**: any rune except `'<', '>', '\\', ' '`
  2. **Escapes** (literal special chars): `\\` → `\`, `\>` → `>`, `\<` → `<`
  3. **Named key**: `<name>` (case-insensitive), where `name` is:
     * Keys: `f1..f12`, `insert`, `delete`, `home`, `end`, `pgup`, `pgdn`, `up`, `down`, `left`, `right`, `tab`, `enter`, `esc`, `space`, `backspace`
     * Mouse: `mouse-left`, `mouse-middle`, `mouse-right`, `mouse-release`, `mouse-wheel-up`, `mouse-wheel-down`
  4. **Modified key**: `<mods-key>` where `mods` uses short/long aliases:
     * `c|ctrl`, `s|shift`, `a|alt`, `m|meta` (order/combos supported per list)
     * Examples: `<c-f1>`, `<m-left>`, `<a-enter>`, `<c-s-a-tab>`, etc.
     * Shifted chars are often encoded directly: `<s-a>` → `A`, `<s-1>` → `!`, `<s-.>` → `>`, etc.
  5. **Modifier-only tokens**: `<ctrl> <shift> <alt> <meta>` and combos:
     `<ctrl-shift> <ctrl-alt> <ctrl-meta> <ctrl-shift-alt> <ctrl-shift-meta> <ctrl-alt-meta> <shift-meta> <alt-meta> <alt-shift> <alt-shift-meta>`
* Invalid:
  * literal space (must be `<space>`)
  * raw `>`; raw `<` without matching `>`
  * raw `\` or `\` followed by anything other than `\`, `<`, `>`

**Bugs**: If we find a bug or are fixing one, we must practice
TDD and first add a test that reproduces it, then fix it and verify
that the test passes.

## Development Workflow

### Adding New Components

1. Create file in `component/`
2. Implement `tui.Component` interface (and optionally `Responsive`, `Floating`, `WithAttributes`)
3. Provide nop wrapper functions for common handler implementations
4. Add table-driven tests
5. Run `make license` to add Apache 2.0 headers

### Working with APIs

1. Define/modify `.proto` files in `api/<name>/<name>rpc/`
2. Run `make generate` to regenerate Go code
3. Implement service logic in `api/<name>/`
4. Add test helpers in `api/<name>/<name>test/`

## Builtin Skills

Six skills are available for semantic code navigation via `runectl`.
They use the language server and tree-sitter parser — prefer them over
text-based alternatives when applicable.

- **`code-navigation`** — Jump to definitions, find all references,
  locate declarations, resolve type definitions, find implementations.
  Prefer over Grep when navigating to where a symbol is defined or used.
- **`code-structure`** — List all symbols in a file or search for
  functions, types, and methods across the workspace. Prefer over
  reading an entire file to understand its structure.
- **`code-search`** — Structural search using tree-sitter queries.
  Use for AST patterns (e.g. "all composite literals of type X").
  More precise than regex for structural patterns.
- **`code-understanding`** — Get type info, documentation, and
  function signatures for any symbol without reading source files.
- **`code-diagnostics`** — Get compilation errors and linter
  diagnostics from the language server. Prefer over `go vet`/`go build`.
- **`code-refactoring`** — Rename symbols across the workspace,
  discover code actions, format files. Prefer over find-and-replace
  or `go fmt`.
