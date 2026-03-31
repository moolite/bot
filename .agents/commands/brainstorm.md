---
name: brainstorm
description: >
    Design-first ideation workflow — explore user intent, clarify constraints,
    propose approaches, and produce an approved design document before
    planning. Use when (1) starting a new feature, (2) exploring architectural
    changes, (3) user wants design before implementation.
---

# Brainstorm Workflow

**Response language follows `language` in `.agents/config/user-preferences.yaml` if configured.**

**Do NOT write any code.** This workflow produces a design document, not implementation.

## Communication Protocol

**Ask questions proactively.** Use the `question` tool to present options or gather critical decisions. Do not assume — ask.

**Iterate with the user.** Go back and forth between exploration and clarification. Present findings, ask questions, refine understanding. Repeat until the approach is solid.

**Use thinking tools when stuck:**
- `sequential-thinking` to break down complex problems
- `memory` tools to recall past decisions or similar designs

---

## Step 1: Explore Project Context

Use code analysis tools to understand the current codebase:
- `glob`, `grep`, `list` for project structure and patterns
- Identify relevant modules, conventions, and constraints
- Summarize what exists and what the user's idea would affect

---

## Step 2: Ask Clarifying Questions

Use the `question` tool. Prefer multiple-choice options when possible.

Key areas:
- **Intent**: What problem are they solving? Who is the target user?
- **Scope**: Must-have vs nice-to-have features
- **Constraints**: Tech stack, timeline, existing integrations
- **Success criteria**: How will they know it's done?

Do NOT proceed to Step 3 until you have clear understanding.

---

## Step 3: Propose Approaches

Present **2-3 distinct approaches**:
- Summary, pros, cons, effort estimate (S/M/L)
- Highlight the **recommended approach** with rationale
- Trade-off comparison matrix

**Get user confirmation before proceeding to Step 4.**

---

## Step 4: Present Design

Present **section by section**, getting feedback at each:
- Architecture overview (components, data flow)
- Key interfaces and contracts
- Integration points with existing code
- Edge cases and error handling

Each section requires explicit approval before moving forward.

---

## Step 5: Save Design

Write to `docs/plans/<feature-name>-design.md`

---

## Step 6: Transition to Planning

Inform the user:
> "Design approved. Run `/plan` to decompose this into actionable tasks."

The design document will be loaded by the planning workflow as context.
