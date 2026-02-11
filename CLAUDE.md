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

**Testing**: Table-driven tests common in component package. Mock RPC types available in `*test/` packages.

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

### Code Quality Requirements

- Race detector enabled by default in tests
- Pre-commit hooks enforce: YAML validation, protobuf regeneration, license headers
- All tests must pass with `-race -timeout 120s`
- License headers required on all Go files (Apache 2.0)

## Code Conventions

- **Line length**: Maximum 90 columns; wrap lines accordingly.
- **File organization**: Public functions, methods, and types go
  at the top of the file. Private types, functions, and methods
  go at the bottom.
- **Testing**: Always use a table-driven approach.
  - `tui.Component` (and derived) implementations are tested
    using the `comptest` package. See `component/*_test.go`
    for examples.
  - `tui.Handler` (and derived) implementations are tested
    using the `handlertest` package. See `handler/inputbox/`
    and `handler/` for examples.
- **Bugs**: If we find a bug or are fixing one, we must practice
TDD and first add a test that reproduces it, then fix it and verify
that the test passes.
