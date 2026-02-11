# Session Context

## User Prompts

### Prompt 1

Update the runectl cli subcommands that take a uri as an argument to allow users to pass a local file, which then can be converted to a uri using extensionapi.Workspace.URI. Basically if the argument doesn't look like a uri and it's a path, we should convert it for users.

