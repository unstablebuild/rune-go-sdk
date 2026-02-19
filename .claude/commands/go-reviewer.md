---
description: Go idiom agent - reviews Go code for idiomatic patterns, style, and best practices.
---

**ACTION REQUIRED: Spawn a subagent using the Task tool.**

Do NOT review code directly. Instead, immediately call the Task tool with:

```
Task(
  subagent_type: "general-purpose",
  description: "Go Enforcer checking [feature]",
  prompt: "
    Read and follow the instructions in .claude/agents/go-idioms.md

    Requirements folder: $ARGUMENTS

    Your task:
    1. Read .claude/agents/go-idioms.md for your role and process
    2. Read $ARGUMENTS/README.md for requirements context
    3. Read any existing go-review-NN.md files to understand previous feedback
    4. Review branch changes: git diff main...HEAD -- .
    5. Systematically check against the review checklist
    6. Write review to $ARGUMENTS/go-review-NN.md (increment NN from last review)
    7. Verdict: APPROVE or REQUEST CHANGES
  "
)
```

Replace `$ARGUMENTS` with: **$ARGUMENTS**

If `$ARGUMENTS` is empty, review all uncommitted changes and write review to current directory.
