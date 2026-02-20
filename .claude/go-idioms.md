# Go Style Guide

## Guidelines
## Go Style Guide

### Key Rules
**Interfaces & Types:** Never use pointers to interfaces. Verify interface compliance at compile time with `var _ Interface = (*Type)(nil)`. Don't embed types in public structs—use explicit delegation instead.
**Mutexes:** Use zero-value mutexes as struct fields (`mu sync.Mutex`), never embed them, never use pointers to them.
**Slices & Maps:** Copy at boundaries (both receiving and returning) to avoid shared-state bugs. `nil` is a valid empty slice—return `nil` not `[]T{}`. Check emptiness with `len(s) == 0`, not `s == nil`.
**Errors:** Handle errors once (don't log-and-return). Wrap with `fmt.Errorf("context: %w", err)`, avoid "failed to" prefixes. Use `Err` prefix for sentinel errors, `Error` suffix for error types. Use `errors.Is`/`errors.As` for matching.
**Channels:** Size of 0 (unbuffered) or 1 only. Anything else needs strong justification.
**Enums:** Start at `iota + 1` unless zero value is a meaningful default.
**Time:** Always use `time.Time` and `time.Duration`. Never bare `int` for time values.
**Goroutines:** Never fire-and-forget. Every goroutine must have a stop signal and a way to wait for exit. No goroutines in `init()`.
**Panics:** Only panic for programmer errors during init. Production code returns errors. Tests use `t.Fatal`.
**Globals:** Avoid mutable globals (use dependency injection). Prefix unexported globals with `_`. Avoid `init()`.
**Exit:** Only call `os.Exit`/`log.Fatal` in `main()`. Use a `run() error` pattern.

### Style
- Soft line limit: 90 characters. Lines should be as long as possible but not longer than limit.
- Use `:=` for local vars, `var` for zero values
- Use field names in struct literals; omit zero-value fields
- Use `make()` for empty maps, literals for fixed sets
- Reduce nesting: handle errors early, return/continue first.
- Do not use if with short stamenets if the statement is multiline or doesn't fit in the soft line limit.
- Group similar declarations; public declarations at top of file
- Functions sorted by receiver, then rough call order
- `strconv` over `fmt` for primitive conversions on hot paths
- Pre-allocate slice/map capacity when size is known
- Use raw string literals to avoid escaping (`str`)
- Use C-style comments for naked bool params: `fn(true /* isLocal */)`

### Line wrapping

When wrapping long Go function signatures, **do NOT** use the style "one parameter per line"
unless there is a strong readability reason. This is considered noisy and makes
small edits create large diffs.

#### Bad (avoid)
- Don’t wrap like this (one param per line):
```go
func HoverHandler(
    lsp semanticapi.LSP,
    wm browserapi.WindowManager,
    newFloating func(string) component.Floating,
) (textapi.CommandHandler, error) {
```

#### Good (preferred)
Prefer packing multiple parameters per line up to the line limit; wrap at commas:

```go
func HoverHandler(
    lsp semanticapi.LSP, wm browserapi.WindowManager,
    newFloating func(string) component.Floating,
) (textapi.CommandHandler, error) {
```

If the signature is still too long, break at semantic boundaries:
- keep each name type pair together
- keep func(...) ReturnType together when possible
- keep return types together (don’t scatter textapi.CommandHandler and error)
- avoid leaving lone ) on its own line

### Testing
- Table-driven tests with subtests, using `tests` slice and `test` in loop var
- Use `give`/`want` prefixes for test case fields
- Keep table tests simple — split complex conditional logic into separate test functions

### Patterns
- Functional options with `Option` interface for extensible constructors (not closures)

### Linting
Run make lint.
