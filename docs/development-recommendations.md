# Development Recommendations

Guidelines for working with Claude Code on this project using TDD and Go.

---

## CLAUDE.md — Project Instructions for Claude

Create a `CLAUDE.md` file at the project root. Claude reads this automatically at the start of every session. It's the single best way to avoid repeating yourself.

Recommended content for this project:

```markdown
# dbase

Go project. Page-based key-value store with HTTP/REST/JSON API. Learning project.

## Commands
- Run tests: `go test ./...`
- Run tests verbose: `go test -v ./...`
- Run specific test: `go test -run TestName ./...`
- Run benchmarks: `go test -bench=. -benchmem ./...`
- Lint: `go vet ./...`
- Format: `gofmt -w .` or `goimports -w .`

## TDD Workflow
1. Write failing test first.
2. Implement minimum code to pass.
3. Refactor.
4. Run `go test ./...` after every change.

## Conventions
- Table-driven tests for multiple cases.
- Use `t.TempDir()` for temp files in tests, not hardcoded paths.
- Use `t.Cleanup()` for per-test teardown.
- Pass `*testing.T` to helpers and call `t.Helper()`.
- Prefer fake implementations over mocks.
- Package `dbase` for white-box tests; `dbase_test` for black-box/API tests.

## Do Not
- Add features beyond what was asked.
- Add comments unless logic is non-obvious.
- Use hardcoded paths like `d:/tmp` in tests.
```

---

## Working with Claude in a TDD Workflow

### The core pattern: you write the test, Claude writes the implementation

This is the most effective way to use Claude for TDD. You stay in control of the specification; Claude fills in the implementation.

1. Write a failing test that captures the exact behaviour you want.
2. Ask Claude: *"Make this test pass"* or *"Implement `Put` so this test passes."*
3. Review the implementation.
4. Run tests yourself to confirm green.
5. If refactoring is needed, ask Claude specifically: *"Refactor this without changing behaviour. Tests must still pass."*

### Alternatively: describe the behaviour, Claude writes test + implementation

When you're unsure how to structure the test:

> "I want `GET /keys/{key}` to return 404 with a JSON error body when the key doesn't exist. Write a table-driven test for the handler and then implement it."

Claude will scaffold both. Review the test before accepting the implementation — the test is the specification.

### Provide context with the request

Claude doesn't retain context between sessions unless it's in CLAUDE.md or memory. Be explicit:

- Reference the interface: *"Implement `Store.Get` — the interface is in `store.go`."*
- Reference existing patterns: *"Follow the same pattern as `Test_AllocationBitMap_Allocate` in `allocation_bitmap_test.go`."*
- Reference constraints: *"Use `t.TempDir()` not hardcoded paths."*

---

## Agentic Coding — Letting Claude Drive Larger Tasks

Claude can handle multi-step tasks autonomously (write test, implement, run tests, fix failures). This works best when:

- The scope is clearly bounded: one endpoint, one method, one struct.
- You give Claude permission to run `go test ./...` after each change.
- You review diffs before accepting, especially for anything touching storage or page layout.

### Good prompts for agentic tasks

> "Using TDD, add a `DELETE /keys/{key}` endpoint. Write a handler test first using httptest and a fake store. Then implement the handler. Then write an integration test against the real store. Then implement the store method. Run `go test ./...` after each step and fix any failures."

> "Refactor the heap tests to use `t.TempDir()` instead of the hardcoded `d:/tmp` path. Run tests after to confirm nothing broke."

> "Add a benchmark for `Heap.Put` following the same pattern as existing benchmarks. Run `go test -bench=BenchmarkHeap -benchmem` to verify it works."

### What to review after agentic tasks

- Does the test actually test the right thing, or does it just pass trivially?
- Are error paths tested (key not found, store error, malformed JSON)?
- Are any hardcoded values or file paths introduced?
- Does the implementation stay within the scope requested?

---

## Project Layout for the HTTP Layer

When adding the HTTP/REST layer, keep it in a separate sub-package to maintain clean separation:

```
dbase/
  *.go              # storage engine (existing)
  *_test.go         # storage tests (existing)
  server/
    server.go       # http.Handler setup, routing
    handlers.go     # individual route handlers
    handlers_test.go
  cmd/
    dbase/
      main.go       # entry point: parse flags, start server
```

This keeps the storage engine independently importable and testable without HTTP concerns.

---

## Running Tests — Practical Notes

The existing `heap_test.go` requires a file at `d:/algs4-data/leipzig1M.txt` for `Test_FileUploadSequential` and `Test_FileUploadParallel`. These tests will panic or fail on machines without that file.

Consider guarding those tests:

```go
func TestFileUpload(t *testing.T) {
    datapath := "d:/algs4-data/leipzig1M.txt"
    if _, err := os.Stat(datapath); os.IsNotExist(err) {
        t.Skip("test data not available")
    }
    // ...
}
```

Or use a build tag to separate them from the standard test run.

---

## Git Workflow

- Commit after each Red-Green-Refactor cycle, not in big batches.
- Commit message pattern: `feat: add GET /keys/{key} handler` or `test: add table-driven tests for KVStore.Put`.
- Keep failing tests off main — only commit when green.

---

## Useful Tools

| Tool | Purpose | Install |
|---|---|---|
| `go test ./...` | Run all tests | built-in |
| `go vet ./...` | Static analysis | built-in |
| `goimports` | Format + manage imports | `go install golang.org/x/tools/cmd/goimports@latest` |
| `golangci-lint` | Comprehensive linting | see golangci-lint.run |
| `dlv` (Delve) | Debugger | `go install github.com/go-delve/delve/cmd/dlv@latest` |

---

## Summary: The Recommended Loop

```
Write failing test
       ↓
Ask Claude to implement / implement yourself
       ↓
go test ./...  →  red?  →  diagnose, fix
       ↓
green
       ↓
Refactor (optional, keep tests green)
       ↓
git commit
       ↓
repeat
```

Keep iterations short. The value of TDD is fast feedback — if a cycle takes more than 10-15 minutes, the increment is too large.
