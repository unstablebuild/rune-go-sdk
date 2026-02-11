# Session Context

## User Prompts

### Prompt 1

Update the runectl cli subcommands that take a uri as an argument to allow users to pass a local file, which then can be converted to a uri using extensionapi.Workspace.URI. Basically if the argument doesn't look like a uri and it's a path, we should convert it for users.

### Prompt 2

I refactored the CLI to use cobra, can you re-apply your changes to the new cli? You need to add it to all of the CLIs that take a uri, including btut not limited to: all of the editor subcommands that take a uri, syntax subcommands that take a uri, wm subcommands that take a uri, open subcommand, and in syntax, query and querynode also should be able to take a file uri/local path.

### Prompt 3

Add tests to verify this.. You should always add tests.

