---
name: delegation
description: When to delegate tasks and work to subagents.
---

# What I do

Delegate implementation tasks to subagents to conserve context and tokens.

## When to Delegate

| Delegate | Don't Delegate |
|----------|----------------|
| New class/module | Quick fixes |
| Writing tests | Reading code |
| Utility functions | Planning decisions |
| Building integrations | Git operations |

## How to Delegate

1. **Break work** into discrete, well-scoped tasks
2. **Spawn subagents** with Task tool
3. **Provide context**: relevant docs, interfaces, requirements, file paths
4. **Parallel**: delegate independent tasks together `[{},{},{}]`
5. **Sequential**: delegate dependent tasks one at a time
6. **Review**: verify requirements met, test output, check integration

## Task Template

```
Implement [component] in [file path].

CONTEXT: [relevant background]

REQUIREMENTS:
- [requirement 1]
- [requirement 2]

INTERFACE: [interfaces to implement]

PATTERNS: [reference existing code]

DELIVERABLES:
- Working implementation
- Docstrings for public methods
- Error handling
```

## Constraints

- NEVER delegate: quick fixes, code reviews, planning, git ops
- Always provide file constraints and requirements
- Review subagent output before proceeding