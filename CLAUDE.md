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

## Go Code Quality Gate

Before completing ANY task that modifies `.go` files, you MUST:
1. Use the `reviewer` subagent to review all modified Go files
3. Address any violations it identifies
4. Re-run the subagent to confirm compliance
5. Use the `go-idioms` subagent to review all modified Go files
6. Address any violations it identifies
7. Re-run the subagent to confirm compliance

Never mark a task complete without this review step.

## Rune MCP Tools

When the Rune MCP server is connected (`runectl mcp`), prefer
these tools over the built-in alternatives. They provide
precise and semantically accurate results from Rune's
language servers and tree-sitter parser, which are more
precise than text-based search.

### Code Navigation (prefer over Grep)

Instead of using Grep to find where something is defined or
used, use the suite LSP tools which understand the code semantically:

- **`lsp_definition`** — Jump to where a symbol is defined.
  Use instead of grepping for a function or type name.
- **`lsp_references`** — Find all usages of a symbol across
  the workspace. Use instead of grepping for callers or
  consumers.
- **`lsp_declaration`** — Find the interface or forward
  declaration of a symbol.
- **`lsp_type_definition`** — Find the type behind a value
  (e.g., the struct that a variable holds).
- **`lsp_implementation`** — Find concrete types implementing
  an interface. Use instead of grepping for implementors.

### Code Structure (prefer over Glob + Grep)

Instead of globbing for files and grepping for patterns to
understand code structure, use these tools:

- **`lsp_document_symbols`** — List all functions, types, and
  variables in a single file. Use instead of reading an entire
  file to understand its structure.
- **`lsp_workspace_symbols`** — Search for symbols by name
  across the entire workspace. Use instead of Glob + Grep to
  locate a function or type.
- **`syntax_search_node`** — Find all functions, methods,
  types, or variables workspace-wide. Use when you need
  "all functions in the project"
  without knowing exact names.
- **`syntax_query_node`** — Same as above but scoped to a
  single file.

### Code Searching (prefer over regex Grep)

When you need to find variables, functions, methods, references,
namespaces or types, use rune's search node tools:

- **`syntax_search_node`** — Run a tree-sitter query across all
  workspace files. Use for searching variables, functions, methods
  references, namespaces, or types across the workspace.
- **`syntax_query_node`** — Same as above but scoped to a single
  known file.

When you need to search for a custom node you can provide
your own tree-sitter query to find it use rune's search syntax tools:

- **`syntax_search`** — Run a tree-sitter query across all
  workspace files. Use for structural patterns like "all
  function calls with two arguments" or "all composite literals
  of type X". More precise than regex.
- **`syntax_query`** — Same as above but scoped to a single
  known file.

### Understanding Code (prefer over reading whole files)

- **`lsp_hover`** — Get type info and documentation for any
  symbol at a position. Use instead of reading source to
  understand what a symbol is.
- **`lsp_signature_help`** — Get function parameter names and
  types. Use when you need to know a function's signature
  without reading its definition.
- **`lsp_completion`** — Discover available methods and fields
  on a type at a cursor position.

### Error Checking (prefer over go build / go vet)

- **`lsp_diagnostics`** — Get compilation errors, warnings,
  and linter diagnostics for a file. Use instead of running
  `go build` or `go vet` to check for errors.

### Refactoring (prefer over manual find-and-replace)

- **`lsp_rename`** — Safely rename a symbol across the entire
  workspace, updating all references. Use instead of
  Grep + Edit for renaming.
- **`lsp_prepare_rename`** — Check if a rename is valid before
  performing it.
- **`lsp_code_actions`** — Discover available refactorings and
  quick fixes at a position (extract variable, organize
  imports, etc.).
- **`lsp_formatting`** — Format an entire file using the
  language server. Use instead of running `go fmt`.
- **`lsp_range_formatting`** — Format a specific range within
  a file.

### Call and Type Hierarchies

Use these for understanding relationships between functions
and types:

- **`lsp_prepare_call_hierarchy`** then
  **`lsp_incoming_calls`** — Find all callers of a function.
  More accurate than grepping for the function name.
- **`lsp_prepare_call_hierarchy`** then
  **`lsp_outgoing_calls`** — Find all functions called by a
  function.
- **`lsp_prepare_type_hierarchy`** then
  **`lsp_type_supertypes`** / **`lsp_type_subtypes`** — Walk
  inheritance chains up or down.

### Other Useful Tools

- **`lsp_document_highlight`** — See all read/write
  occurrences of a variable within a single file.
- **`lsp_code_lens`** — Discover inline actions like
  "Run test" attached to code ranges.
- **`lsp_execute_command`** — Execute a server-side command
  (e.g., from a code lens or code action).
- **`lsp_folding_range`** — Understand the block structure of
  a file.
- **`lsp_selection_range`** — Get smart selection expansion at
  a cursor position.

## Code Conventions

- **File organization**: Public functions, methods, and types go
  at the top of the file. Private types, functions, and methods
  go at the bottom.

## Go Style Guide

### Key Rules

**Interfaces & Types:** Never use pointers to interfaces. Verify interface compliance at compile time with `var _ Interface = (*Type)(nil)`. Don't embed types in public structs—use explicit delegation instead.
**Mutexes:** Use zero-value mutexes as struct fields (`mu sync.Mutex`), never embed them, never use pointers to them.
**Slices & Maps:** Copy at boundaries (both receiving and returning) to avoid shared-state bugs. `nil` is a valid empty slice—return `nil` not `[]T{}`. Check emptiness with `len(s) == 0`, not `s == nil`.
**Errors:** Handle errors once (don't log-and-return). Wrap with `fmt.Errorf("context: %w", err)`, avoid "failed to" prefixes. Use `Err` prefix for sentinel errors, `Error` suffix for error types. Use `errors.Is`/`errors.As` for matching.
**Channels:** Size of 0 (unbuffered) or 1 only. Anything else needs strong justification.
**Enums:** Start at `iota + 1` unless zero value is a meaningful default.
**Time:** Always use `time.Time` and `time.Duration`. Never bare `int` for time values.
**Goroutines:** Never fire-and-forget. Every goroutine must have a stop signal and a way to wait for exit. No goroutines in `init()`. Use `go.uber.org/goleak` in tests.
**Panics:** Only panic for programmer errors during init. Production code returns errors. Tests use `t.Fatal`.
**Globals:** Avoid mutable globals (use dependency injection). Prefix unexported globals with `_`. Avoid `init()`.
**Exit:** Only call `os.Exit`/`log.Fatal` in `main()`. Use a `run() error` pattern.

### Style

- Soft line limit: 90 chars, only break a line into multiple lines when crossing this boundary.
- Two import groups: stdlib, then everything else
- Use `:=` for local vars, `var` for zero values
- Use field names in struct literals; omit zero-value fields
- Use `make()` for empty maps, literals for fixed sets
- Reduce nesting: handle errors early, return/continue first
- Group similar declarations; public declarations at top of file
- Functions sorted by receiver, then rough call order
- `strconv` over `fmt` for primitive conversions on hot paths
- Pre-allocate slice/map capacity when size is known
- Use raw string literals to avoid escaping
- Use C-style comments for naked bool params: `fn(true /* isLocal */)`

### Testing

- Table-driven tests with subtests, using `tests` slice and `tt` loop var
- Use `give`/`want` prefixes for test case fields
- Keep table tests simple—split complex conditional logic into separate test functions
