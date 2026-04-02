# Test Refactor Plan

This document describes Go testing conventions, the issues in the current test suite, and the exact changes to make. It is written so Claude can execute the changes without additional context.

---

## Background: Go Testing Conventions

### Test file location and package

Test files (`_test.go`) live in the same directory as the code they test. Go supports two package styles:

- `package dbase` — **white-box** tests. Same package as the code. Can access unexported types and functions. Correct for testing internal behaviour (page layout, slot management, bitmap operations).
- `package dbase_test` — **black-box** tests. Separate package. Only exported API visible. Correct for integration and API-level tests.

For this storage engine, **`package dbase` is correct throughout**. The tests legitimately need access to unexported types (`heapPage`, `memoryStore`, constants like `maxRecordLen`, `slotTableLen`). Do not switch to `package dbase_test`.

### No shared global state between tests

Each test must be independently runnable and not depend on the order in which other tests run. A global `var store` that is set up in `TestMain` and shared across tests is an anti-pattern — it makes tests order-dependent and prevents parallelism.

`TestMain` is only appropriate when setup is genuinely impossible per-test (e.g. a Docker container or a network service). Opening a file store is cheap; every test that needs a store should create its own.

### `t.TempDir()` replaces custom temp file helpers

`t.TempDir()` (Go 1.15+) creates a temporary directory that is automatically cleaned up when the test ends. It integrates with `t.Cleanup()` internally.

The idiomatic replacement for the current `tempfile()` pattern:

```go
// Old: create temp file, delete it, return path
func tempfile() string {
    f, _ := os.CreateTemp(os.TempDir(), "db-")
    f.Close()
    os.Remove(f.Name())
    return f.Name()
}

// New: return a path inside t's auto-cleaned temp directory
func tempfile(t *testing.T) string {
    t.Helper()
    return filepath.Join(t.TempDir(), "db")
}
```

The `t.Helper()` call marks the function as a test helper so that failure line numbers point to the caller, not the helper.

### `t.Cleanup()` for deferred teardown

`t.Cleanup(fn)` registers a function to run when the test ends, in the same way as `defer` but integrated with the testing framework. It runs even when `t.Fatal` is called mid-test.

```go
store, _ := Open(tempfile(t), 0666, nil)
t.Cleanup(func() { store.Close() })
```

`os.Remove` is no longer needed when using `t.TempDir()` — the directory and all its contents are cleaned up automatically.

### `t.Fatal` / `t.Fatalf` — not `panic`

Panic in a test goroutine crashes the entire test binary and produces confusing output. Use `t.Fatal` to stop a test and `t.Errorf` to record a failure without stopping. Never call `panic` in test code.

```go
// Bad
if err != nil {
    panic(err)
}

// Good
if err != nil {
    t.Fatalf("Open: %v", err)
}
```

### Table-driven tests with empty cases are noise

A table-driven test scaffold with no test cases (`// TODO: Add test cases.`) and an empty table is dead code. It should either be filled in or replaced with straightforward non-table-driven tests.

---

## Current Problems

### `cmd/dbase/main.go`

Contains test logic disguised as functions: `Test_HeapWrite`, `Test_HeapDelete`, `Test_CreateHeap`, `memoryStore()`, `fileStore()`, `tempfile()`. These all belong in the proper test files (the heap variants are already duplicated in `heap_test.go`). `main()` calls `Test_HeapDelete()` directly.

### `heap_test.go`

- `var store FileStore` — global store shared by `Test_CreateHeap`, `Test_HeapWrite`, `Test_HeapDelete`. Tests are order-dependent.
- `TestMain` — sets up the global store. Should be removed once each test owns its store.
- `const heapRuns = 100` — declared but unused (redefined locally inside `Test_HeapWrite`).
- `panic()` in `TestMain` — goes away when `TestMain` is removed.
- Bug in `Test_FileUploadParallel` line ~198: `store.Close()` closes the global store, not `store1`.
- `tempfile()` is defined here and in `file_store_test.go` — duplicate symbol in the same package; one must be removed.
- `logElapsedTime` is defined here and used by `overflow_page_test.go` — move to shared helpers file.

### `file_store_test.go`

- `tempfile()` uses `panic()` for errors. Needs replacing with `t.TempDir()`-based helper.
- `TestNewPages` is functionally identical to `TestCreateAndReopenDB` — remove the duplicate.

### `memory_store_test.go`

- Six test functions with empty test case tables (`// TODO: Add test cases.`). Loops never execute. Tests are useless.
- Replace with real tests using `NewMemoryStore()` as the factory.
- `reflect` import becomes unused — remove.

### `overflow_page_test.go`

- Calls `tempfile()` (no args) — needs updating to `tempfile(t)` once the helper signature changes.
- Uses `logElapsedTime` from `heap_test.go` — will resolve from shared helpers file.

### `heap_page_test.go` and `page_test.go`

Structurally sound. No changes needed.

### `allocation_bitmap_test.go`

`Test_AllocationBitMap_AllocateExplicit` and `Test_AllocationBitMap_Deallocate` mutate a shared bitmap across subtests intentionally (sequential state transitions). Leave the test logic intact — no structural change needed.

---

## Changes to Make

### Step 1: Create git flow feature branch

```bash
git flow feature start test-refactor
```

---

### Step 2: Simplify `cmd/dbase/main.go`

Replace the entire file content with a minimal entry point. Remove all test functions, helpers, profiling code, and unused imports.

New content:
```go
package main

import "log"

func main() {
	log.Println("dbase: starting")
	// TODO: initialise server and start listening
	log.Println("dbase: stopped")
}
```

---

### Step 3: Create `testhelpers_test.go`

**File to create:** `testhelpers_test.go` at the repo root.
**Package:** `package dbase`

Consolidates `tempfile`, `openStore`, and `logElapsedTime` in one place, removing duplication across `heap_test.go` and `file_store_test.go`.

```go
package dbase

import (
	"log"
	"path/filepath"
	"testing"
	"time"
)

// tempfile returns a path for a temporary database file inside t's temp
// directory. The directory is automatically removed when the test ends.
func tempfile(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "db")
}

// openStore opens a FileStore at a temporary path and registers cleanup.
func openStore(t *testing.T) FileStore {
	t.Helper()
	store, err := Open(tempfile(t), 0666, nil)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func logElapsedTime(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
```

---

### Step 4: Refactor `heap_test.go`

**Remove these from the file:**
- `import "log"` and `import "time"` (no longer needed)
- `const heapRuns = 100`
- `var store FileStore`
- `func TestMain(m *testing.M) { ... }` — entire function
- `func logElapsedTime(...)` — moved to `testhelpers_test.go`
- `func tempfile()` — moved to `testhelpers_test.go`

**Update `Test_CreateHeap`:**
```go
func Test_CreateHeap(t *testing.T) {
	store := openStore(t)
	heap := NewHeap(store)

	if count := store.Count(); count != 2 {
		t.Fatalf("page count: got %d, want 2", count)
	}
	if count := heap.Count(); count != 0 {
		t.Fatalf("record count: got %d, want 0", count)
	}
}
```

**Update `Test_HeapWrite`:** replace `store` (global) with `store := openStore(t)` at top. Keep the heapRuns local variable (value `100000`).

**Update `Test_HeapDelete`:** replace `store` (global) with `store := openStore(t)` at top.

**Update `Test_FileUploadSequential`:** replace manual `Open`/defer with `store2 := openStore(t)`.

**Update `Test_FileUploadParallel`:**
- `store1` is created with `NewMemoryStore()` — keep that, but remove the broken defer that calls `store.Close()` and `os.Remove(store.Path())` (which references the now-removed global `store`).
- The defer should be removed entirely; `store1` (a MemoryStore) needs no file cleanup.

---

### Step 5: Refactor `file_store_test.go`

**Remove from the file:**
- `func tempfile() string { ... }` — moved to `testhelpers_test.go`
- `TestNewPages` — duplicate of `TestCreateAndReopenDB`
- `"path/filepath"` import — now in `testhelpers_test.go`

**Update all tests** to use `openStore(t)` instead of manual open/defer/remove. Example pattern:

Before:
```go
path := tempfile()
store, err := Open(path, 0666, nil)
defer func() {
    store.Close()
    os.Remove(store.Path())
}()
if err != nil { t.Fatal(err) }
```

After:
```go
store := openStore(t)
```

Apply to: `TestOpen`, `TestCreateAndReopenDB`, `TestSetGet`.

For `TestOpen_ErrNotExists` — uses `tempfile()` to construct a bad path. Update to `tempfile(t)`.

`TestOpen_ErrPathRequired` — no store opened, no changes needed.

Remove `"os"` import if no longer used after removing manual cleanup.

---

### Step 6: Rewrite `memory_store_test.go`

Replace the entire file content with real tests. Remove the `reflect` import.

New file content:
```go
package dbase

import "testing"

func TestNewMemoryStore(t *testing.T) {
	store, err := NewMemoryStore()
	if err != nil {
		t.Fatalf("NewMemoryStore: %v", err)
	}
	if count := store.Count(); count != 0 {
		t.Fatalf("initial count: got %d, want 0", count)
	}
}

func TestMemoryStore_AppendAndRead(t *testing.T) {
	store, err := NewMemoryStore()
	if err != nil {
		t.Fatalf("NewMemoryStore: %v", err)
	}

	page := NewHeapPage()
	id, err := store.Append(page)
	if err != nil {
		t.Fatalf("Append: %v", err)
	}
	if count := store.Count(); count != 1 {
		t.Fatalf("count after append: got %d, want 1", count)
	}

	page2 := NewHeapPage()
	if err := store.Read(id, page2); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if page2.GetID() != id {
		t.Fatalf("page ID after read: got %d, want %d", page2.GetID(), id)
	}
}

func TestMemoryStore_WriteAndRead(t *testing.T) {
	store, err := NewMemoryStore()
	if err != nil {
		t.Fatalf("NewMemoryStore: %v", err)
	}

	page := NewHeapPage()
	id, err := store.Append(page)
	if err != nil {
		t.Fatalf("Append: %v", err)
	}

	record := []byte("hello store")
	if _, err := page.AddRecord(record); err != nil {
		t.Fatalf("AddRecord: %v", err)
	}
	if err := store.Write(id, page); err != nil {
		t.Fatalf("Write: %v", err)
	}

	page2 := NewHeapPage()
	if err := store.Read(id, page2); err != nil {
		t.Fatalf("Read: %v", err)
	}
	buf := make([]byte, len(record))
	if _, err := page2.GetRecord(1, buf); err != nil {
		t.Fatalf("GetRecord: %v", err)
	}
	if string(buf) != string(record) {
		t.Fatalf("record: got %q, want %q", buf, record)
	}
}

func TestMemoryStore_Count(t *testing.T) {
	store, err := NewMemoryStore()
	if err != nil {
		t.Fatalf("NewMemoryStore: %v", err)
	}
	for i := 0; i < 5; i++ {
		if _, err := store.Append(NewHeapPage()); err != nil {
			t.Fatalf("Append %d: %v", i, err)
		}
	}
	if count := store.Count(); count != 5 {
		t.Fatalf("count: got %d, want 5", count)
	}
}

func TestMemoryStore_Close(t *testing.T) {
	store, err := NewMemoryStore()
	if err != nil {
		t.Fatalf("NewMemoryStore: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}
```

---

### Step 7: Update `overflow_page_test.go`

- Replace manual `Open`/defer with `store := openStore(t)` in `Test_MultipleOverflowPages`
- Remove `"os"` import if no longer used
- Keep `"time"` import — `time.Now()` is still called inline

---

### Step 8: Verify and finish

```bash
go build ./...
go vet ./...
go test ./... -count=1 -v 2>&1 | tail -40
```

All tests must pass. The two `d:/algs4-data` tests must SKIP. No panics.

```bash
git flow feature finish test-refactor
git push origin develop
```

---

## File Change Summary

| File | Action |
|------|--------|
| `cmd/dbase/main.go` | Replace with minimal 3-line entry point |
| `testhelpers_test.go` | **Create** — `tempfile(t)`, `openStore(t)`, `logElapsedTime` |
| `heap_test.go` | Remove `TestMain`, global `store`, unused const, duplicate helpers; update tests |
| `file_store_test.go` | Remove `tempfile()`, remove `TestNewPages`; update tests to `openStore(t)` |
| `memory_store_test.go` | Replace empty stubs with real tests |
| `overflow_page_test.go` | Update to use `openStore(t)` |
| `heap_page_test.go` | No changes |
| `page_test.go` | No changes |
| `allocation_bitmap_test.go` | No changes |

## What Is NOT Being Changed

- Package declarations — all test files stay as `package dbase`
- Test logic in `allocation_bitmap_test.go` — sequential mutation is intentional
- `heap_page_test.go` and `page_test.go` — already correct
- The `d:/algs4-data` skip guards — leave as-is
