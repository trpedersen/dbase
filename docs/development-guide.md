# Development Guide

**Status**: Draft

This guide covers how to build, test, and lint the project, and describes the TDD workflow we follow.  
For project conventions, see [CLAUDE.md](../CLAUDE.md) at the repo root — that file is the source of truth.

---

## Prerequisites

- Go 1.26.1 (check `go.mod` for the canonical version)
- `git` and `git-flow`
- Optional: `golangci-lint` for static analysis

---

## Build

```bash
go build ./...
```

This compiles all packages including `cmd/dbase`. No output on success.

---

## Test

```bash
# Run all tests once
go test ./...

# Run with verbose output
go test ./... -v

# Disable test caching (always re-run)
go test ./... -count=1

# Run a specific test
go test -run Test_HeapWrite -v

# Run tests in a specific package
go test github.com/trpedersen/dbase -v
```

Two tests (`Test_FileUploadSequential`, `Test_FileUploadParallel`) require test data at `d:/algs4-data/`. They will **skip** automatically if that path does not exist.

---

## Vet and Lint

```bash
# Vet (always run before committing)
go vet ./...

# Lint (if golangci-lint is installed)
golangci-lint run
```

All code must pass `go vet` before it is considered complete. Lint issues should be resolved, not suppressed.

---

## The TDD Loop

We follow a strict **spec → interface → test → implement → refactor** cycle. Do not skip steps.

### Step 1: Write or update the spec

Before writing any code, there must be a spec in `docs/`. The spec describes:

- **Intent**: What does this feature do? Why does it exist?
- **Interface**: What is the public API? (method signatures, types)
- **Invariants**: What is always true? What can never happen?
- **Data structures**: How is data laid out? What are the size constraints?
- **Error conditions**: What can fail? What error type is returned?
- **Edge cases**: Empty input, zero values, maximum sizes, concurrent access.

If there is no spec, don't implement. If the spec is ambiguous, clarify it first.

### Step 2: Define the interface (if introducing a new abstraction)

Write the Go interface in the appropriate file before writing any implementation code.

```go
// Frob manages frobbing operations on pages.
type Frob interface {
    Frob(id PageID) error
    Unfrob(id PageID) error
}
```

### Step 3: Write a failing test

Write the test before the implementation. The test should fail with a compile error or a clear assertion failure — not a panic.

```go
func TestFrob_Basic(t *testing.T) {
    store := openStore(t)
    f := NewFrob(store)

    if err := f.Frob(0); err != nil {
        t.Fatalf("Frob(0): %v", err)
    }
}
```

Run `go test ./...` — the test should fail (red).

### Step 4: Implement

Write the minimum code needed to make the test pass. Don't gold-plate.

Run `go test ./...` — all tests should pass (green).

### Step 5: Refactor

With tests green, review for:
- Clarity and idiomatic Go
- Alignment with the spec
- Any duplication or complexity that can be simplified

Run `go test ./...` again after refactoring — still green.

### Step 6: Commit

```bash
git add <specific files>
git commit -m "add Frob implementation"
```

---

## Git Workflow

We use git-flow. The main branches are:

| Branch | Purpose |
|--------|---------|
| `main` | Production-ready, tagged releases |
| `develop` | Integration branch — all feature work merges here |
| `feature/*` | One feature per branch, branched from `develop` |

```bash
# Start a feature
git flow feature start my-feature

# Finish a feature (merges to develop, deletes feature branch)
git flow feature finish my-feature

# Push develop to origin
git push origin develop
```

**Rule**: A feature branch is finished only when all tests pass and `go vet` is clean.

---

## Adding a New Feature: Checklist

- [ ] Spec written in `docs/<feature>.md`
- [ ] Significant design choices recorded as an ADR in `docs/decisions/`
- [ ] Interface defined in the appropriate `.go` file
- [ ] Tests written (failing) before implementation
- [ ] Implementation written to pass tests
- [ ] `go vet ./...` passes
- [ ] `go test ./... -count=1` passes
- [ ] Feature branch finished and pushed to `develop`
