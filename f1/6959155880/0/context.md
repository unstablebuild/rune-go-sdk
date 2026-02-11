# Session Context

## User Prompts

### Prompt 1

The cli package provides a nice and convenient way to implement CLIs. The problem is that it's not able to parse flags after the arguments. We should update it to allow passing flags in any position after a subcommand. For example if the synopsis of a command is command subcommand <arg1> <arg2>, currently flags can only be parsed before arg1, but not after arg2. We want users to be able to pass command subcommand -F one -F two <arg1> -F three <arg2> -F four. Identify what part of the code has th...

### Prompt 2

Sorry, rebased this branch; we were missing some changes. Sure do this changes. This is a good place to practice TDD, add the tests first, make sure they fail, then try to fix them by adding the correct parsing logic. Add many tests to the testing harness, as this could break in many ways.

### Prompt 3

Do an analysis on how much work it would be to replace our cli package with https://github.com/spf13/cobra.

### Prompt 4

Remove the cli package and cmd/runectl to use cobra. The commands should do exactly the same, same flags (in long form expand the name, i.e. -F is now -F/--format, etc.). I'm interested in the shell completion aspect it, and we definetly don't want to implement that ourselves.

### Prompt 5

This session is being continued from a previous conversation that ran out of context. The summary below covers the earlier portion of the conversation.

Summary:
1. Primary Request and Intent:
   The user had three main requests:
   1. Initially: Investigate the `cli` package's limitation where flags could only be parsed before positional arguments, and propose changes to allow interspersed flags (flags in any position after subcommand)
   2. After rebase: Implement the interspersed flags featur...

