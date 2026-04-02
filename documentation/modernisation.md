# Go Codebase Modernisation

This document describes why the dbase codebase needs modernising, the background knowledge behind each change, and the exact steps to execute. It is written so a future Claude session can carry out all changes without needing extra context.

---

## Background: What Changed in Go

### Go Modules (Go 1.11+, mandatory by ~1.16)

Go originally used `GOPATH` — all source code lived under `$GOPATH/src/`. This made it hard to version dependencies or work outside a single directory tree. Go modules replaced this:

- A `go.mod` file at the repo root declares the module path and minimum Go version.
- A `go.sum` file records cryptographic checksums of all dependencies.
- `go mod tidy` resolves, downloads, and records all dependencies.
- Projects can live anywhere on disk, not just under `$GOPATH`.

The dbase project has no `go.mod` — it was written before or without modules. This must be added first because every other `go` tool command depends on it.

**Commands:**
```bash
go mod init github.com/trpedersen/dbase
# Edit go.mod to set: go 1.21
go mod tidy
```

### `io/ioutil` Deprecated (Go 1.16)

The `ioutil` package was a historical convenience wrapper. In Go 1.16 all its functions were moved to `os` and `io` directly, and `ioutil` was officially deprecated in Go 1.16.

| Old | New |
|-----|-----|
| `ioutil.TempFile(dir, pattern)` | `os.CreateTemp(dir, pattern)` |
| `ioutil.TempDir(dir, pattern)` | `os.MkdirTemp(dir, pattern)` |
| `ioutil.ReadFile(path)` | `os.ReadFile(path)` |
| `ioutil.WriteFile(path, data, perm)` | `os.WriteFile(path, data, perm)` |
| `ioutil.ReadAll(r)` | `io.ReadAll(r)` |
| `ioutil.NopCloser(r)` | `io.NopCloser(r)` |
| `ioutil.Discard` | `io.Discard` |

### `rand.Seed` Deprecated (Go 1.20)

Before Go 1.20, the global random source in `math/rand` used a fixed seed (1) unless explicitly seeded with `rand.Seed()`. This meant programs had deterministic (non-random) output unless seeded.

From Go 1.20, the global source is **automatically seeded with a random value** at program startup. Calling `rand.Seed()` is now a no-op and officially deprecated.

**Old pattern:**
```go
rand.Seed(2323)  // deprecated, no-op in Go 1.20+
l := rand.Intn(1000)
```

**New pattern (if you want a reproducible seed for tests):**
```go
r := rand.New(rand.NewSource(2323))
l := r.Intn(1000)
```

**Or, if you don't need reproducibility, just remove `rand.Seed` entirely.**

### `bytes.Compare` for Equality

`bytes.Compare(a, b)` returns -1, 0, or 1 (like `strcmp`). Using it to test equality (`bytes.Compare(a, b) != 0`) is inefficient — it computes a full ordering when only equality is needed.

`bytes.Equal(a, b)` is the idiomatic replacement. It's clearer and potentially faster.

```go
// Old
if bytes.Compare(record1, record2) != 0 { ... }

// New
if !bytes.Equal(record1, record2) { ... }
```

### `interface{}` → `any` (Go 1.18)

Go 1.18 introduced `any` as a built-in alias for `interface{}`. They are identical; `any` is simply more readable.

```go
// Old
New: func() interface{} { return NewHeapPage() }

// New
New: func() any { return NewHeapPage() }
```

---

## What Is NOT Being Changed

These issues exist but are architectural decisions, not mechanical modernisation:

- **`panic` in `NewHeap` and `UnmarshalBinary`** — Converting to error returns would cascade through all callers. Out of scope.
- **Concurrency / atomic statistics counters** — Design issue.
- **Completing stubs** (`db.go`, `buffered_page_store.go`) — New work.
- **Adding `context.Context`** — New feature.
- **Adding godoc to all exported symbols** — Separate task.

---

## Step-by-Step Execution

### Step 1: Initialize Go Modules

```bash
cd /path/to/dbase
go mod init github.com/trpedersen/dbase
```

Then edit `go.mod` to add the Go version line (it may already be present):
```
module github.com/trpedersen/dbase

go 1.21
```

Then:
```bash
go mod tidy
```

This will fetch `github.com/trpedersen/rand` (used in `heap_test.go`, `overflow_page_test.go`, and `main/main.go`) and generate `go.sum`.

**Files created:** `go.mod`, `go.sum`

---

### Step 2: Replace `ioutil` with `os`/`io`

#### `main/main.go`

1. Remove `"io/ioutil"` from imports (verify `"os"` is already imported).
2. Line ~134: change `ioutil.TempFile("d:/tmp", "db-")` to `os.CreateTemp(os.TempDir(), "db-")`.

#### `file_store_test.go`

1. Remove `"io/ioutil"` from imports.
2. Line ~170: change `ioutil.TempFile("c:/tmp", "db-")` to `os.CreateTemp(os.TempDir(), "db-")`.

Grep to confirm no other `ioutil` usage remains:
```bash
grep -r "ioutil" .
```

---

### Step 3: Remove `rand.Seed`

#### `main/main.go`

Remove line: `rand.Seed(2323)`

The random number generation in that function is used to pick a random record length for a write benchmark. Removing the seed means the length will vary between runs, which is fine for a benchmark. If reproducibility is ever needed, replace with `r := rand.New(rand.NewSource(2323))` and use `r.Intn(...)`.

After removing, check whether the `"math/rand"` import is still needed elsewhere in the file. If `rand.Intn` is still used, keep the import but remove just the `rand.Seed` call.

---

### Step 4: Replace `bytes.Compare` with `bytes.Equal`

Grep for all occurrences:
```bash
grep -rn "bytes.Compare" .
```

Expected locations:
- `main/main.go` — line ~81
- `heap_test.go` — line ~85
- `heap_page_test.go` — multiple
- `overflow_page_test.go` — ~line 70

For each occurrence:
```go
// Old
if bytes.Compare(a, b) != 0 { ... }

// New
if !bytes.Equal(a, b) { ... }
```

After replacing, verify `"bytes"` import is still used (it will be, for `bytes.Equal`).

---

### Step 5: Fix Hardcoded Paths

#### `main/main.go` — `tempfile()` function

```go
// Old
f, err := ioutil.TempFile("d:/tmp", "db-")

// New (after ioutil fix above)
f, err := os.CreateTemp(os.TempDir(), "db-")
```

#### `file_store_test.go` — `tempfile()` function

```go
// Old
f, err := ioutil.TempFile("c:/tmp", "db-")

// New
f, err := os.CreateTemp(os.TempDir(), "db-")
```

#### `heap_test.go` — Add `t.Skip` guard for data-file tests

`Test_FileUploadSequential` and `Test_FileUploadParallel` both hardcode `d:/algs4-data/leipzig1M.txt`. They panic if the file doesn't exist. Add a skip guard immediately after the `datapath` assignment in each function:

```go
datapath := "d:/algs4-data/leipzig1M.txt"
if _, err := os.Stat(datapath); os.IsNotExist(err) {
    t.Skip("test data not available:", datapath)
}
```

---

### Step 6: Fix Variable Shadowing

#### `heap.go` — `Put` method (~line 138)

```go
// Old — shadows built-in len
len := len(buf)
free := heap.lastPage.GetFreeSpace()
if len > int(free) {

// New
bufLen := len(buf)
free := heap.lastPage.GetFreeSpace()
if bufLen > int(free) {
```

Only one reference to `len` inside that function — update it to `bufLen`.

---

### Step 7: Replace `interface{}` with `any`

#### `heap.go` — `sync.Pool` `New` function (~line 43)

```go
// Old
New: func() interface{} {
    return NewHeapPage()
},

// New
New: func() any {
    return NewHeapPage()
},
```

Grep for any other `interface{}` in the codebase:
```bash
grep -rn "interface{}" .
```

Replace any remaining `interface{}` usages with `any` where they are not part of an interface definition (e.g. `type Foo interface { Bar() interface{} }` — those should also be updated).

---

### Step 8: Delete Dead Code Files

The following files contain only a package declaration plus entirely commented-out code. Delete them:

- `db_header_page.go`
- `heap_directory.go`

```bash
rm db_header_page.go heap_directory.go
```

---

### Step 9: Remove Commented-Out Code Blocks

#### `heap.go`

Remove the large commented-out `Write` method at the bottom of the file (lines ~239–264):

```go
//func (heap *Heap) Write(buf []byte) error {
//  len := len(buf)
//  ...
//}
```

Delete this entire block.

#### `heap_page.go`

Read the file and identify any large commented-out blocks. Remove them.

---

### Step 10: Verify

```bash
go build ./...
go vet ./...
go test ./...
```

Expected results:
- `go build` — no errors
- `go vet` — no warnings
- `go test` — all tests pass; `Test_FileUploadSequential` and `Test_FileUploadParallel` report SKIP (not FAIL or panic)

If `go mod tidy` reports missing dependencies, check that `github.com/trpedersen/rand` is publicly accessible on GitHub.

---

## Files Changed Summary

| File | Action | Reason |
|------|--------|--------|
| `go.mod` | Create | Add Go modules |
| `go.sum` | Create | Dependency checksums |
| `main/main.go` | Edit | Remove ioutil, rand.Seed; fix path; bytes.Equal |
| `heap.go` | Edit | interface{}→any; len shadowing; remove dead comment block |
| `heap_test.go` | Edit | bytes.Equal; t.Skip guards |
| `heap_page_test.go` | Edit | bytes.Equal |
| `overflow_page_test.go` | Edit | bytes.Equal |
| `file_store_test.go` | Edit | Remove ioutil; fix path |
| `db_header_page.go` | Delete | 100% commented-out dead code |
| `heap_directory.go` | Delete | 100% commented-out dead code |
