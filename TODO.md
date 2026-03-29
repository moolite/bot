# Code Diagnostics and Improvement Plan

## Critical Issues

### 1. Test Failures - FTS5 Not Available in Memory Database ⚠️
**Files Affected:**
- `internal/core/handlers_test.go`
- `internal/db/tables_test.go`
- `internal/statistics/statistics_test.go`

**Problem:** Tests use in-memory database (`:memory:`) but FTS5 is not available in memory databases. The FTS5 extension requires file-based databases with the `fts5` build tag.

**Error Message:**
```
err: no such module: fts5 in line 0: CREATE VIRTUAL TABLE media_fts USING fts5(...)
```

**Solution:**
Option 1: Add FTS5 build tag to tests
- Update test files to build with `-tags "fts5"`
- Requires SQLite FTS5 extension to be available

Option 2: Skip FTS5 migration in tests
- Make FTS5 table creation conditional on build tag
- Add fallback logic for tests without FTS5 support

Option 3: Use test database file
- Create temporary SQLite file for testing instead of `:memory:`

**Recommended:** Option 2 - make FTS5 conditional since it's an optional feature.

---

## Code Quality Issues

### 2. Unused Code (8 warnings)
**Files:**
- `cmd/marrano-bot/main.go:31` - `flagSyncMedia` variable declared but never used
- `internal/core/core.go:31` - `thenFunc` method on chain type
- `internal/core/handlers.go:822` - `isMedia` function
- `internal/db/db.go:45` - `prepareNamedStmt` function
- `flake.nix:68` - unused `hardeningOptions` binding

**Impact:** Code clutter, confusion about active functionality

**Solution:** Remove or comment out unused code with explanation.

---

### 3. Nil Dereference Warnings (3 warnings)
**Files:**
- `internal/db/links.go:56` - Scanning into pointer fields without nil check
- `internal/db/export.go:84` - Map update on nil pointer

**Problem:** `rows.Scan(&l.Text, &l.URL, &l.GID)` can fail but code doesn't handle potential nil returns before dereferencing.

**Solution:** Add nil checks after Scan or use pointer pointers.

---

### 4. Modernization Opportunities
**Files:**
- `pkg/tg/bot.go:245,253` - `interface{}` can be replaced with `any`
- `internal/core/handlers.go:540` - Could use `max()` instead of if statement
- `pkg/tg/emoji.go:4` - SA9004: All constants in group should have explicit type

**Impact:** Better Go 1.18+ idioms, clearer code

---

## Security Considerations

### 5. Nil Pointer Dereferences
**Location:** `internal/db/links.go:56`

**Risk:** If Scan fails or returns nil pointers, dereferencing them causes panic.

**Severity:** Medium - could crash the bot

**Recommendation:** Add defensive nil checks.

---

## Performance

### 6. Prepared Statement Cache
**Location:** `internal/db/db.go`

**Observation:** Prepared statements are cached globally in module variables (`stmts`, `nstmts`). This can cause:
- Memory leaks if database connection is never closed
- Statement pooling issues across requests

**Recommendation:** Consider connection-scoped statement caching or use context to manage lifespan.

---

## Test Coverage

### 7. Missing Test Coverage
**Observation:** No tests for:
- `internal/config` package (0 test files)
- `cmd/marrano-bot/main.go` (0 test files)
- Many database CRUD operations

**Impact:** Reduced confidence in code changes

**Recommendation:** Add unit tests for config loading, CLI flags, and DB operations.

---

## Dependencies

### 8. Outdated Dependencies
**Go Version:** 1.23.3 (from go.mod)

**Observation:** No explicit go.mod checks for minimum version

**Recommendation:** Consider adding `go 1.23` minimum version constraint if features require it.

---

## Documentation

### 9. Missing AGENTS.md Content
**Status:** AGENTS.md created but could be enhanced with:
- Example migration patterns
- Detailed API documentation
- Troubleshooting guide

---

## Priority Fixes

### High Priority
1. **Fix FTS5 test failures** - Critical for CI/CD
2. **Fix nil dereferences** - Security/stability

### Medium Priority
3. Remove unused code
4. Add test coverage
5. Modernize interface{} -> any

### Low Priority
6. Refactor statement caching
7. Enhance AGENTS.md
