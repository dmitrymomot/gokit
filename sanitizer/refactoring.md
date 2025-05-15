# Sanitizer Package Refactoring Plan

> **Goal:** Align the `sanitizer` package architecture, API and coding style with the more advanced design already implemented in the `validator` package, embracing breaking API changes as needed.

---

## 1. Motivation

* `validator` has evolved into an instance-based, highly configurable component with clear separation of global and per-instance state, option pattern configuration, customizable separators, and robust error handling.
* `sanitizer` still relies on **package-level** variables & functions, lacks configuration flexibility, and exposes a minimal API.
* Refactoring will improve testability, re-usability, concurrency safety and user experience, and will bring API consistency across `gokit` packages.

## 2. High-Level Objectives

| # | Objective | Outcome |
|---|-----------|---------|
| 1 | Introduce `Sanitizer` struct mirroring `validator.Validator` | Allows isolated instances with their own registered sanitizers & options |
| 2 | Provide *options pattern* (`Option` funcs) for configuration | Users can customise tag name, rule / param separators etc. |
| 3 | Migrate from package-level state to an **instance-based design** | Cleaner API; wrappers or globals become optional convenience helpers |
| 4 | Improve rule parsing (parameter list separator, trimming, regex special case) | Consistent behaviour with `validator.parseRule` |
| 5 | Add contextual error handling helpers (e.g. unknown rule) | Easier debugging for developers |
| 6 | Review built-in sanitizers, split into focused files and add missing ones (e.g. `trimspace`, `slug`, `lowercamel`) | Richer standard library & easier maintenance |
| 7 | Enhance concurrency safety (instance mutex + sync.Map where appropriate) | No global locks in hot paths |
| 8 | Update tests to follow **one-test-case-per-function** style (no table tests) | Align with project testing guidelines |
| 9 | Extend documentation (README, Godoc, examples in README only) | Better DX, no `examples.go` files (user preference) |

## 3. Architectural Changes

### 3.1 `Sanitizer` struct

```go
// Sanitizer holds instance-specific configuration and sanitizer map.
type Sanitizer struct {
    ruleSeparator      string            // default ","
    paramSeparator     string            // default ":"
    paramListSeparator string            // default "," (for nested params)
    fieldNameTag       string            // e.g. "json", defaults to struct field name

    sanitizers     map[string]SanitizeFunc
    sanitizersMu   sync.RWMutex
}
```

### 3.2 Options pattern

Adopt the same approach as `validator.Option`:

```go
type Option func(*Sanitizer) error
```

Built-in options:

* `WithRuleSeparator(string)`
* `WithParamSeparator(string)`
* `WithParamListSeparator(string)`
* `WithFieldNameTag(string)`
* `WithSanitizers(map[string]SanitizeFunc)` (advanced)

### 3.3 Global default instance

* `var DefaultSanitizer = NewMust()`
* Optional: provide package-level helpers (e.g., `RegisterSanitizer`, `SanitizeStruct`) that proxy to the default instance for developer convenience, not for compatibility.

### 3.4 Built-in sanitizer registration

* At `init()` call `resetBuiltInSanitizers()` and register into `DefaultSanitizer` *and* provide a helper to copy into new instances.
* Move each sanitizer into its own small file or grouped logically (e.g. `string_sanitizers.go`, `case_sanitizers.go`).

### 3.5 Rule parsing

Replicate `validator.parseRule` logic, including:

* Trimming spaces.
* Handling `paramListSeparator`.
* Treating certain rules as *raw* (e.g. `regex` equivalent if ever added).

### 3.6 Error handling

Introduce `errors.go` with:

```go
var (
    ErrInvalidSanitizerConfiguration = errors.New("invalid sanitizer configuration")
)
```

Return these from option application or `RegisterSanitizer` when mis-used.

### 3.7 Concurrency

* Each instance keeps its own mutex; global default instance reuses it.
* Hot-path sanitization acquires **read** lock only when retrieving function.
* Consider pre-resolving rules per-type and caching (stretch goal).

### 3.8 Delimiters & Parameters rationale

Using the same delimiters as the `validator` package brings coherence across the toolkit and enables advanced sanitizers that accept arguments:

* **`ruleSeparator`** (default `";"`) — separates multiple sanitizers in a single tag, e.g. `sanitize:"trim;lower;replace:foo,bar"`.
* **`paramSeparator`** (default `":"`) — splits the sanitizer name from its parameter block, e.g. `replace:foo,bar`.
* **`paramListSeparator`** (default `","`) — divides multiple parameters within that block when necessary.

Parameters are essential for rules such as:

* `replace:old,new` – needs two parameters (`old`, `new`).
* `truncate:30` – needs one numeric parameter specifying maximum length.
* Future examples like `pad:left,10` or `slug:dash`.

Reusing these delimiters keeps the learning curve shallow for developers already familiar with the `validator` API and simplifies internal parsing logic.

## 4. API Changes (Public)

| Existing | New equivalent | Notes |
|----------|----------------|-------|
| `RegisterSanitizer(tag, fn)` | `(*Sanitizer).RegisterSanitizer(tag, fn)` + proxy | Returns error on bad input |
| `SanitizeStruct(ptr)` | `(*Sanitizer).SanitizeStruct(ptr)` + proxy | Same behaviour, but configurable |
| – | `New(options ...Option) (*Sanitizer, error)` | Factory |
| – | `MustNew(options ...Option) *Sanitizer` | Helper for pain-free init |

## 5. Behaviour Improvements

1. **`omitempty`-like skip** — support an `omitempty` rule to skip sanitizers on zero values.
2. **Nested struct traversal** — optionally recurse into embedded structs like validator.
3. **Field label tag** — borrow `label` tag for future richer error messages.
4. **Chain optimisation** — stop early when value unchanged (low priority).

## 6. Additional Sanitizers to Consider

* `slug` – convert to URL-friendly slug (uses `utils.Slug`)
* `email` – lower-case & trim spaces
* `uuid` – normalise UUID strings (lowercase, remove braces)
* `bool` – coercion of truthy/falsey strings

## 7. Testing Strategy

* Ensure **100 % coverage of exported functions**.
* One sub-test per rule (current style OK).
* Add tests for:
  * Custom rule registration error cases.
  * Concurrency (race detector).
  * Option parsing edge cases.
* Benchmarks for heavy sanitization workloads (like webhook benchmarks).

## 8. Documentation & Examples

* Update `README.md` to reflect new instance-based API.
* Provide usage examples inside README (no `examples.go`).

## 9. Implementation Milestones

| Phase | Scope |
|-------|-------|
| 1 | Introduce `Sanitizer` struct, options, and refactor existing code (legacy API can be removed) |
| 2 | Migrate tests, ensure green build & race detection |
| 3 | Update README and finalize documentation |
| 4 | Implement optional enhancements (omitempty, nested structs, new sanitizers) |
| 5 | Benchmark & optimise hot paths |

---

**Estimated effort:** 1–2 days of focused work (~500 LOC change + docs + tests).

---

_End of plan._
