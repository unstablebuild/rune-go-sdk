# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: `idelsp` Package

## Context

Create an `idelsp/` package at the repository root that implements `idelsp.Manager` — a multi-language LSP server manager. The Manager implements `semanticapi.LSP` (routing calls to underlying per-language LSP servers) and `textapi.EventHandler` (receiving file events and mapping them to LSP notifications). It manages the full lifecycle of LSP servers: discovery via `PkgManager`, process start via `workspaceapi.Executor`, JSO...

