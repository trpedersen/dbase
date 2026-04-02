# Go Code Organisation Plan

This document describes Go conventions for code organisation, the current state of the dbase repository, and the exact changes to make. It is written so Claude can execute the changes without additional context.

---

## Background: Go Package and Directory Conventions

### The Central Rule

In Go, the directory structure **is** the package structure. One directory = one package. The package name should match the directory name (except for `main` packages). This is enforced by the toolchain.

### What Makes a Good Package

Go takes an unusually conservative view on splitting code into packages:

- **Split when there is a reason**, not for organisational tidiness. Go's standard library has large, cohesive packages (e.g. `net/http`, `encoding/json`).
- **Don't create `util`, `common`, or `helper` packages.** These accumulate unrelated things and have no useful API contract.
- **Package names are the API.** Callers write `heap.New()`, `store.Open()` etc. A bad package name makes the call site ugly.
- **Tightly coupled code belongs in one package.** If package A can't function without package B's internals, they should be the same package.

For a storage engine like dbase — where `Heap` needs `HeapPage`, `HeapHeaderPage`, `PageStore`, and `RID` all at once — keeping them in a single `dbase` package is correct Go idiom.

### Key Conventions

#### Executables live in `cmd/<name>/`

Any `package main` that produces a binary belongs in a subdirectory of `cmd/`:

```
cmd/
  dbase/
    main.go   ← the server/CLI binary
```

This is the most universally observed Go convention. The directory `main/` is not idiomatic. A project can have multiple binaries in `cmd/`.

#### `internal/` restricts imports

A package at `internal/` (or any subdirectory of it) can only be imported by code rooted at the parent of the `internal/` directory. Use this when you want to prevent external consumers from importing implementation details.

For this project, `internal/` is not needed yet — the storage engine is legitimately a public library. Use it in the future if implementation packages need to be hidden from API consumers.

#### The root package is the primary library

For a module named `github.com/trpedersen/dbase`, the root package (`package dbase`) is what users import. It should contain the most important public-facing API. This project does this correctly — the storage engine types (`Heap`, `PageStore`, `FileStore`, etc.) live at the root.

#### `doc.go` for package-level documentation

A file named `doc.go` is the conventional place for a package-level doc comment and nothing else:

```go
// Package dbase implements a page-based storage engine.
// ...
package dbase
```

This keeps the doc comment separate from implementation files and makes it easy to find.

#### Test files alongside source

Go test files (`_test.go`) belong in the same directory as the code they test. They are already doing this correctly.

#### No deep nesting without reason

A flat structure with a few well-named packages is better than a deep hierarchy. Adding sub-packages too early creates import cycles and coupling problems.

### Planned Future Structure (after HTTP layer is added)

```
github.com/trpedersen/dbase/        ← storage engine library (package dbase)
  *.go
  cmd/
    dbase/
      main.go                       ← HTTP server entry point
  server/
    server.go                       ← HTTP handler setup and routing
    handlers.go                     ← individual route handlers
    handlers_test.go
  documentation/
    *.md
```

The `server/` package is where the HTTP/REST/JSON layer will live. It imports `package dbase` for storage operations. Keeping it separate from the storage engine maintains clean separation — the engine has no knowledge of HTTP.

---

## Current State

### Module
- Module path: `github.com/trpedersen/dbase`
- Go version: 1.26.1
- One external dependency: `github.com/trpedersen/rand`

### Package layout
Everything is in `package dbase` at the repository root. This is correct for the storage engine code.

### Problems to fix

#### 1. `main/` should be `cmd/dbase/`

`main/main.go` is a `package main` binary. By Go convention it belongs in `cmd/dbase/main.go`.

The import path for the library (`github.com/trpedersen/dbase`) does not change — only the directory containing the binary changes.

#### 2. No `doc.go`

The root package has no package-level documentation comment. There should be a `doc.go`.

#### 3. Stub files are confusing

Four files exist as pure placeholders with no implementation:
- `db.go` — `type DB interface{}`
- `record.go` — `type Record struct{}`
- `buffered_page_store.go` — `type BufferedPageStore struct{}` with a truncated comment
- `page_directory.go` — `PageDirectory` interface, no implementation

These are legitimate forward-design placeholders, but they should carry a short comment explaining their intent so future Claude sessions and readers understand they are not dead code but planned extension points.

#### 4. `main/main.go` is a test harness, not a real entry point

The current `main.go` contains heap write/delete test functions, not a server or useful CLI. It should be replaced with a proper entry point when the HTTP layer is built. For now, rename it to `cmd/dbase/main.go` without changes.

---

## Changes to Make

### Change 1: Rename `main/` to `cmd/dbase/`

**What:** Move `main/main.go` to `cmd/dbase/main.go`.

**Why:** `cmd/<name>/` is the universal Go convention for binary entry points.

**Steps:**
1. Create directory `cmd/dbase/`
2. Move `main/main.go` to `cmd/dbase/main.go`
3. Delete the now-empty `main/` directory
4. No import changes needed — the file imports `github.com/trpedersen/dbase` which is unchanged

**Verify:** `go build ./...` still compiles with no errors.

---

### Change 2: Add `doc.go`

**What:** Create `doc.go` at the repository root with a package-level doc comment.

**Why:** Convention; makes `go doc github.com/trpedersen/dbase` useful.

**Content:**
```go
// Package dbase implements a page-based storage engine.
//
// The storage model is built around fixed-size pages (8 KB). Pages are
// persisted through a [PageStore], which can be file-backed ([FileStore])
// or in-memory ([MemoryStore]).
//
// Records are stored in a [Heap], which manages a sequence of [HeapPage]
// instances. Each record is identified by a [RID] (record ID) combining
// a page ID and slot number. The [HeapScanner] provides sequential
// iteration over all stored records.
//
// Large records that exceed the heap page payload are stored as linked
// chains of [OverflowPage] instances.
//
// Page allocation within a store is tracked by [AllocationBitMap] and
// [AllocationPage].
package dbase
```

**File to create:** `doc.go` at repo root.

---

### Change 3: Add intent comments to stub files

**What:** Add a one-line `// TODO:` comment to each stub file explaining its intended purpose.

**Why:** Stubs without comments look like forgotten dead code. A comment makes them deliberate extension points.

**Changes:**

`db.go` — add comment above the interface:
```go
// DB is the top-level interface for a dbase database instance.
// TODO: define once the key-value store API is designed.
type DB interface {
}
```

`record.go` — add comment above the struct:
```go
// Record represents a structured database record.
// TODO: define fields once the key-value data model is settled.
type Record struct {
}
```

`buffered_page_store.go` — replace the truncated comment with a clear one:
```go
// BufferedPageStore is a PageStore implementation that adds a write buffer
// in front of an underlying store to reduce I/O.
// TODO: implement.
type BufferedPageStore struct {
}
```

`page_directory.go` — the interface already exists; add a file-level comment if missing.

---

### Change 4: Create `cmd/dbase/` scaffold for future HTTP server (optional, do last)

**What:** After moving `main/main.go`, add a comment at the top of `cmd/dbase/main.go` noting it is a temporary test harness and will become the HTTP server entry point.

**Why:** Documents intent without code changes.

---

## Execution Order

1. `go build ./...` — confirm clean baseline before changes
2. Create `cmd/dbase/` directory
3. Move `main/main.go` → `cmd/dbase/main.go`
4. Delete empty `main/` directory
5. `go build ./...` — confirm still compiles
6. Create `doc.go`
7. `go build ./...` — confirm still compiles
8. Edit `db.go`, `record.go`, `buffered_page_store.go` to add intent comments
9. Final verification: `go build ./...` + `go vet ./...` + `go test ./...`

---

## What Is Explicitly NOT Being Changed

- The root `package dbase` — correct as-is for a storage engine library
- File names within the root package — no Go convention requires splitting heap.go into sub-packages
- Test file locations — already correct
- The `documentation/` directory — not a Go convention issue
- Import paths — the module path `github.com/trpedersen/dbase` does not change

---

## Future: Adding the HTTP Layer

When building the HTTP/REST/JSON key-value API, add:

```
server/
  server.go        ← package server; http.Handler setup, routing
  handlers.go      ← individual route handlers
  handlers_test.go ← httptest-based handler tests
```

And update `cmd/dbase/main.go` to be a real server entry point that:
1. Parses config flags (port, data file path)
2. Opens a `dbase.FileStore`
3. Creates a `dbase.Heap`
4. Starts the HTTP server from `server.NewServer(...)`

The `server` package imports `package dbase`. The `cmd/dbase` package imports both. This keeps the storage engine and the HTTP layer independently testable.

---

## Verification Commands

```bash
go build ./...        # must produce no errors
go vet ./...          # must produce no warnings
go test ./...         # all tests pass; data-file tests skip gracefully
```
