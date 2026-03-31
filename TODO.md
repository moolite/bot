# Code Diagnostics and Improvement Plan

## Current State (Updated: 2026-03-31)

### ✅ RESOLVED ISSUES (Removed)

1. **FTS5 test failures** - Fixed with graceful handling. Tests now skip FTS5 migrations when unavailable and limit to version 10.
2. **Unused code warnings** - Most items resolved:
   - `flagSyncMediaFolder` - Actually in use at main.go:148-151
   - `isMedia` function - Does not exist in codebase
   - `prepareNamedStmt` function - Does not exist in codebase
   - `hardeningOptions` - Does not exist in flake.nix
3. **max() modernization** - Already implemented at handlers.go:544
4. **Nil dereference warnings** - Links.go:56 code is correct (Scan return is checked with error handling)
5. **Go version documentation** - go.mod specifies `go 1.24.1` (not 1.23.3)
6. **AGENTS.md enhancement** - ✅ Completed with comprehensive documentation

---

## Active Issues

### Medium Priority

### 1. Unused Method - thenFunc
**Location:** `internal/core/core.go:31`

**Status:** Marked with `// nolint:unused // kept for future use`

**Issue:** The `thenFunc` method on the `chain` type is unused but has a comment indicating it's kept for future use.

**Recommendation:** Either:
- Remove if no plans to use it
- Remove the nolint comment and actually use it if there's a use case

---

### 2. Modernize interface{} to any
**Location:** `pkg/tg/client.go:15`

**Issue:** One occurrence of `interface{}` instead of the modern `any` alias (Go 1.18+).

**Code:**
```go
Result interface{} `json:"result"`
```

**Recommendation:** Replace with `any` for Go 1.18+ compatibility.

---

### 3. Constants Type Inconsistency ✅ RESOLVED
**Location:** `pkg/tg/types.go:6`

**Status:** Fixed - All constants now have explicit `string` type

**Solution Applied:** Added explicit `string` type to all 18 constants for consistency with Go best practices.

---

### Low Priority

### 1. Missing Test Coverage
**Locations:**
- `internal/config/` package - 0 test files
- `cmd/marrano-bot/main.go` - 0 test files

**Impact:** Reduced confidence in code changes for config loading and CLI functionality.

**Recommendation:** Add unit tests for:
- Config file parsing (TOML)
- Environment variable overrides
- CLI flag handling
- Main function initialization

---

### 2. Statement Cache Architecture
**Location:** `internal/db/db.go`

**Issue:** Prepared statements are cached globally in module-level map (`stmts`).

**Current Implementation:**
```go
var stmts map[string]*sqlx.Stmt = make(map[string]*sqlx.Stmt)

func prepareStmt(stmt string) (*sqlx.Stmt, error) {
    if prepared, ok := stmts[stmt]; ok {
        return prepared, nil
    }
    // ... prepare and cache
}
```

**Potential Issues:**
- Statements not scoped to connection lifetime
- Memory if connection is never closed properly
- Potential concurrency issues with concurrent access

**Current Mitigation:** Connection is typically singleton for bot lifecycle, and `Close()` resets the cache.

**Recommendation:** Consider connection-scoped statement caching or use database connection pooling with proper statement lifecycle management.

---

## Priority Fixes

### High Priority
None - All critical issues resolved.

### Medium Priority
1. Modernize `interface{}` → `any` (single occurrence)
2. Remove or use `thenFunc` method

### Low Priority
1. Add test coverage for config and main.go
2. Refactor statement cache architecture (optional, current design works for use case)
