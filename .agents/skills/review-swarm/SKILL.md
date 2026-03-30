---
name: review-swarm
description: >
    Parallel read-only multi-agent review of git diff or explicit file scope.
    Use when: review swarm, parallel review, diff review, regression review,
    security review, or wants prioritized issues without editing files.
---

# What I do

Parallel diff review with four read-only sub-agents, then synthesize findings.

## Step 1: Determine Scope

Prefer order:
1. Explicit files from user
2. Current git changes
3. Explicit branch/commit/PR
4. Recently modified files

If no clear scope, ask user.

Build intent packet:
- What behavior should change
- What should stay unchanged
- Any constraints (compatibility, security, migration)

## Step 2: Launch 4 Reviewers

Launch in parallel for anything beyond tiny diffs.

Each reviewer gets same scope + intent, is read-only, and returns:
- File + line, issue, why it matters, confidence

### Reviewer 1: Intent & Regression
- Unintended behavior changes
- Broken edge cases
- Contract drift
- Missing adjacent updates

### Reviewer 2: Security & Privacy
- Authn/authz gaps
- Injection risks
- Secret exposure
- Risky defaults

### Reviewer 3: Performance & Reliability
- Redundant work/I/O
- Hot path cost
- Leaks, race conditions
- Brittle failure handling

### Reviewer 4: Contracts & Coverage
- API/schema mismatches
- Compatibility fallout
- Missing tests
- Missing logs/metrics

## Step 3: Aggregate & Filter

Merge findings, drop:
- Duplicates
- Speculative claims
- Minor style nits

Normalize to: file:line, category, severity, why, fix, confidence

## Step 4: Order Output

1. High-severity, high-confidence
2. Medium-severity worth fixing
3. Lower-severity follow-ups

If no issues, say so.

## Step 5: Path Forward

- `fix now` - before merge
- `fix soon` - if time permits
- `optional` - can wait

Output is review only - do not apply fixes.

## Constraints

- NEVER edit files or apply patches
- NEVER skip intent clarification
- NEVER bury user in low-value noise
- Only report material issues