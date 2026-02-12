# Session Context

## User Prompts

### Prompt 1

We're building a state of the art sandbox implementation for our extensions and plugins. We want to support 4 platforms: linux x86_64, linux arm64, macos x86_64, mscos arm64, with a caveat: macos 86_64 is not a priority so might drop it. There's a project at github.com/jinkkaihe/matchlock that looks promising, there's also github.com/apple/container as well, which we could combine with just one of the million container virtualization frameworks for linux. Agent Sandbox Security: Essential Featur...

### Prompt 2

This needs to be retrofitted into a workspaceapi.Executor, that's the integration point: clients pass a workspaceapi.Cmd and we need to run that inside a sandbox and manage everything else, keeping the sandboxing transparent for the calling client. How does this affect your evaluation of vendors?

### Prompt 3

[Request interrupted by user for tool use]

