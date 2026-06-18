# snippets

`snippets` is a complete, runnable **extension** example. Unlike the
[`repl`](../repl) and [`dialogue`](../dialogue) examples (which are in-terminal
*plugins* built on the TUI framework), this one is a standalone program that
Rune launches as a child process and talks to over gRPC. It exercises a broad
slice of the extension surface:

- **Command registration** via `Workspace.RegisterCommand` (`snippets
  [insert|edit|copy|delete] <name>`), with tab completion.
- **Storage** (`storageapi.Service`) to persist snippets across sessions,
  auto-partitioned by extension id.
- **Editor** (`textapi.Editor`) to insert a snippet body at the cursor.
- **Editor events** (`textapi.EventHandler`) to capture the last selection and
  to persist scratch buffers when they are flushed or closed.
- **File system** (`workspaceapi.FileSystem`) and the **resource opener**
  (`browserapi.ResourceOpener`) plus the **window manager**
  (`browserapi.WindowManager`) to open a scratch buffer in a tab and focus it.
- **Notifications** (`browserapi.Notifications`) for success and informational
  messages.

## What it does

| Command | Behavior |
| --- | --- |
| `snippets insert <name>` | Insert the stored snippet's body at the cursor. |
| `snippets edit <name>` | Open a scratch buffer seeded with the snippet (empty if new); saving the buffer persists it. |
| `snippets copy <name>` | Open a scratch buffer seeded with the last text selection, saved under `<name>`. |
| `snippets delete <name>` | Remove a stored snippet. |

Scratch buffers live in the OS temp directory and are removed when the buffer
is closed.

## File layout

The example is deliberately split to show a clean separation of concerns:

- **`main.go`** does only negotiation, metadata, dependency wiring, and command
  registration. There is no business logic here.
- **`command_handler.go`** holds all the logic: the `snippets` type that
  implements `textapi.CommandHandler` and `textapi.EventHandler`, plus its
  helpers. Dependencies are typed as SDK interfaces so they can be faked.
- **`command_handler_test.go`** contains black-box tests that drive the type
  through its interfaces, using `storagestub.NewInMemoryService()` for real
  storage behavior and small in-file fakes for the editor, file system,
  resource opener, window manager, and notifications.

## Patterns worth copying

- **Thin `main`, testable handler.** Keep `main`/`run` to wiring; put logic in
  a type that takes SDK interfaces, so tests can substitute fakes.
- **Return wrapped errors; let the host notify.** Command handlers return
  `fmt.Errorf(...)`; Rune surfaces the message to the user. Only call
  `Notifications` directly when you return `nil` but still want to tell the
  user something (a success or an event-handler failure that cannot be
  returned).
- **`run` returns after setup.** `ServeWorkspaceExtension` blocks until
  shutdown on its own; do not block in `run`. The `ctx` is cancelled on
  shutdown for any in-flight work.
- **Declare every permission you use.** A call into a capability you did not
  declare in `Metadata.Permissions` is rejected by the host.

See the [extension SDK guide](https://docs.rune.build/develop/sdk) for the full
extension-authoring walkthrough.
