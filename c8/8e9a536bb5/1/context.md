# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Minimal Darwin Sandbox with Embedded Images

## Context

The current sandbox implementation depends on:
- **Matchlock** - Go wrapper around Apple's Virtualization.framework
- **e2fsprogs** - Required at runtime for ext4 image manipulation
- **Downloaded binaries** - guest-agent and guest-fused from GitHub

These external dependencies complicate deployment and make the package harder to use. The goal is to create a minimal implementation that:
1. Uses `github.com/...

