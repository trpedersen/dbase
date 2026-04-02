# 001: Error Handling — Return errors, do not panic

**Status**: Accepted  
**Date**: 2026-04-02

---

## Context

The current codebase uses `panic` in two distinct situations that need to be separated:

**1. Runtime I/O and store errors** — `NewHeap` panics if `store.Append`, `store.Write`, or `store.Read` returns an error. These are ordinary runtime failures (disk full, file closed, permission denied). Panicking on them crashes the calling program with a confusing stack trace and makes the library unusable in any production context.

**2. Deserialisation precondition violations** — `UnmarshalBinary` on every page type (`heapPage`, `heapHeaderPage`, `overflowPage`, `allocationPage`) panics if the buffer is the wrong length or the page type byte doesn't match. These panics fire when a caller passes a buffer that was read from a corrupt or incompatible file.

**3. Internal invariant violations** — `heapPage.getSlotFlags/Offset/Length` panic on out-of-range slot numbers. `heapPage.SetRecord` panics via `reallocateSlot` on certain code paths marked "TODO: replace panic once fully debugged".

The current panic-based approach:
- Makes `NewHeap` impossible to use safely — callers cannot handle store failures
- Makes page deserialisation fatal on any corrupt or mismatched file data
- Prevents any meaningful error recovery or graceful shutdown
- Is inconsistent with the rest of the codebase, which already returns `error` everywhere else (all `PageStore` methods, `Heap` methods, etc.)

The `CLAUDE.md` convention states: *"No `panic` except for genuine programming errors (impossible states, violated invariants). Never for runtime conditions."*

---

## Decision

**Return `error` throughout. Eliminate all `panic` calls except those that represent genuine, unrecoverable programming errors.**

Specifically:

### `NewHeap` — change signature to return `error`

```go
// Before
func NewHeap(store PageStore) Heap

// After
func NewHeap(store PageStore) (Heap, error)
```

All `panic(fmt.Sprintf("NewHeap ...", err))` calls become `return nil, fmt.Errorf("NewHeap: %w", err)`. The same change applies to `heap.Clear()` which has the same pattern.

### `UnmarshalBinary` — return `error` instead of `panic`

All four page types (`heapPage`, `heapHeaderPage`, `overflowPage`, `allocationPage`) currently panic on invalid buffer length or wrong page type. These become:

```go
// Before
if len(buf) != int(PageSize) {
    panic("Invalid buffer")
}

// After
if len(buf) != int(PageSize) {
    return fmt.Errorf("UnmarshalBinary: buffer length %d, want %d", len(buf), PageSize)
}
```

`UnmarshalBinary` already returns `error` in its signature — it is just not being used.

### `heapPage` internal slot accessors — return `error`

`getSlotFlags`, `getSlotOffset`, `getSlotLength` panic on out-of-range slot. These are called only from within the package. Change them to return `(value, error)` and propagate up to the public methods (`GetRecord`, `SetRecord`, `DeleteRecord`, etc.) which already return `error`.

### `heapPage.SetRecord` / `reallocateSlot` — remove `panic`

The two `panic(err)` calls marked "TODO: replace panic once fully debugged" become `return err`.

### `panic` that remains acceptable

| Location | Reason |
|----------|--------|
| A future `panic("unreachable")` after an exhaustive switch | Genuine programming error — impossible at runtime if the code is correct |

No other `panic` use is acceptable in library code.

---

## Consequences

### Positive
- `NewHeap` can be called safely; callers can handle store failures and report them cleanly
- Corrupt or mismatched files produce error messages, not crashes
- The codebase is consistent: error handling follows the same pattern everywhere
- The library becomes usable from any context (server, CLI, tests) without crashing the host process on I/O errors

### Negative
- **Cascading change**: Every call site of `NewHeap` must be updated to handle the returned error. This includes `heap_test.go` (multiple tests), `overflow_page_test.go`, and any future callers.
- The slot accessor methods become slightly more verbose internally.

### Follow-on work
- The test files (`heap_test.go`, `file_store_test.go`) also contain `panic` calls — those will be addressed as part of the test refactor described in `plans/test-refactor-plan.md`, not as part of this change.
- `cmd/dbase/main.go` contains `panic` calls in test helper code that will be removed entirely in the test refactor.

---

## Alternatives Considered

| Alternative | Why rejected |
|-------------|-------------|
| Keep `panic` in `UnmarshalBinary`, treat as "impossible" | Corrupt files are not impossible. A file written by a different version, a partial write, or an OS error can all produce mismatched page type bytes. These must be recoverable. |
| Wrap `NewHeap` in a `recover()` at call sites | Papering over the problem. Does not make the error handleable in any meaningful way and is un-idiomatic Go. |
| Leave as-is, document the panics | The panics prevent the library from being used safely. Not acceptable. |
