# TDD with Go

A practical guide for Test-Driven Development in Go, tailored to this project.

---

## The TDD Cycle

TDD follows a tight loop:

1. **Red** — Write a failing test that captures the desired behaviour. It must compile but fail.
2. **Green** — Write the minimum code to make the test pass. Don't over-engineer.
3. **Refactor** — Clean up both code and tests without changing behaviour. Re-run tests to confirm green.

Repeat for every small increment of functionality. The discipline is in keeping the increments small.

---

## Go's Built-in Testing Package

Go ships a first-class test framework in the standard library. No additional dependencies required to get started.

```go
// foo_test.go
package dbase

import "testing"

func TestFoo(t *testing.T) {
    got := foo(42)
    want := "answer"
    if got != want {
        t.Errorf("foo(42) = %q, want %q", got, want)
    }
}
```

Run tests:

```bash
go test ./...          # all packages
go test -v ./...       # verbose output
go test -run TestFoo   # run a specific test by name (regex)
go test -count=1 ./... # disable test caching
```

---

## Test File and Package Naming

Go supports two packaging styles for tests.

### White-box tests (same package)

```go
package dbase   // same package — can access unexported identifiers
```

Use for unit tests that need access to internal state. This is what the existing tests in this project use (e.g. `heap_test.go`, `allocation_bitmap_test.go`).

### Black-box tests (external package)

```go
package dbase_test   // external package — only exported API visible
```

Use for integration or API-level tests. Mirrors what a user of your package would see. Good for HTTP handler tests and key-value store API tests.

You can mix both styles in the same directory. A common pattern is to put unit tests in `package dbase` and HTTP/API tests in `package dbase_test`.

---

## Table-Driven Tests

The dominant Go idiom for tests with multiple cases. Already in use in this project (`allocation_bitmap_test.go`).

```go
func TestKVStore_Get(t *testing.T) {
    tests := []struct {
        name    string
        key     string
        want    string
        wantErr bool
    }{
        {"existing key", "foo", "bar", false},
        {"missing key",  "baz", "",    true},
        {"empty key",    "",    "",    true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            store := newTestStore(t)
            got, err := store.Get(tt.key)
            if (err != nil) != tt.wantErr {
                t.Fatalf("Get(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("Get(%q) = %q, want %q", tt.key, got, tt.want)
            }
        })
    }
}
```

Benefits:
- Adding a case is one line in the table
- `t.Run` gives each case its own name in the output
- Failures are isolated — one case failing doesn't stop others
- Easy to run a single case: `go test -run TestKVStore_Get/missing_key`

---

## TestMain for Suite Setup/Teardown

Use `TestMain` when the whole test binary shares expensive setup (opening a file store, spinning up a server). Already used in `heap_test.go`.

```go
func TestMain(m *testing.M) {
    // setup
    db, cleanup := setupTestDB()
    defer cleanup()

    os.Exit(m.Run())
}
```

Prefer `t.Cleanup` for per-test teardown rather than `defer` inside tests — it runs even when `t.Fatal` is called mid-test.

```go
func TestSomething(t *testing.T) {
    store := openTempStore(t)
    t.Cleanup(func() { store.Close() })
    // ...
}
```

---

## Test Helpers

Extract repeated setup into helpers. The Go convention is to pass `*testing.T` and call `t.Helper()` so that failure line numbers point to the caller, not the helper.

```go
func newTestStore(t *testing.T) *KVStore {
    t.Helper()
    path := t.TempDir() // automatically cleaned up
    store, err := Open(path, 0666, nil)
    if err != nil {
        t.Fatalf("open store: %v", err)
    }
    t.Cleanup(func() { store.Close() })
    return store
}
```

Note: `t.TempDir()` is cleaner than manual `ioutil.TempFile` + `os.Remove`. It creates a unique temp directory that is automatically removed after the test.

---

## Interfaces for Testability

Go interfaces are the primary mechanism for test doubles (mocks/stubs/fakes). Define small interfaces at the point of use rather than large ones at the point of definition.

This project already models this well — `PageStore` is an interface, allowing `MemoryStore` and `FileStore` as interchangeable implementations. Fast unit tests use `MemoryStore`; integration tests use `FileStore`.

For the HTTP layer, define interfaces for the storage backend so handlers can be tested without a real store:

```go
type Store interface {
    Get(key string) ([]byte, error)
    Put(key string, value []byte) error
    Delete(key string) error
}
```

---

## Testing HTTP Handlers

Go's `net/http/httptest` package is purpose-built for this. No need to start a real server.

```go
import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHandleGet(t *testing.T) {
    store := &fakeStore{data: map[string][]byte{"foo": []byte("bar")}}
    handler := NewHandler(store)

    req := httptest.NewRequest(http.MethodGet, "/keys/foo", nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
    }
    // assert body...
}
```

This pattern lets you test:
- Status codes
- Response bodies and JSON
- Headers
- Error paths (store returns error → handler returns 404/500)

---

## Fake vs Mock vs Stub

In Go, hand-written fakes are usually preferable to generated mocks:

- **Fake**: A working in-memory implementation of an interface. Simple, readable, no external deps. Prefer this.
- **Stub**: Returns hardcoded values. Fine for simple cases.
- **Mock**: Records calls and asserts on them. Use sparingly. If you need a mock framework, `github.com/stretchr/testify/mock` is the standard choice.

A simple fake store for HTTP handler tests:

```go
type fakeStore struct {
    data map[string][]byte
    err  error // if set, all operations return this error
}

func (f *fakeStore) Get(key string) ([]byte, error) {
    if f.err != nil { return nil, f.err }
    v, ok := f.data[key]
    if !ok { return nil, ErrKeyNotFound }
    return v, nil
}
// Put, Delete omitted for brevity
```

---

## Subtests and Parallel Tests

Run subtests in parallel to speed up large suites:

```go
for _, tt := range tests {
    tt := tt // capture loop variable (required in Go < 1.22)
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // ...
    })
}
```

Note: parallel subtests share the parent test's resources — be careful with shared state. For the store layer, give each parallel subtest its own store instance via `t.TempDir()`.

---

## Benchmarks

Go benchmarks live alongside tests and use the same file conventions:

```go
func BenchmarkKVStore_Put(b *testing.B) {
    store := newTestStoreB(b)
    key := "bench-key"
    val := []byte("bench-value")

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        store.Put(key, val)
    }
}
```

Run:

```bash
go test -bench=. -benchmem ./...
```

Benchmarks are separate from the TDD cycle but valuable for database internals — use them to catch performance regressions when changing page layouts or buffer management.

---

## Test Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out   # opens browser
```

Coverage is a useful signal, not a goal. 100% coverage with weak assertions is worse than 70% with precise ones. Focus coverage on the business logic and error paths; don't chase coverage on boilerplate.

---

## Common Anti-Patterns to Avoid

| Anti-pattern | Better approach |
|---|---|
| Tests that depend on execution order | Each test sets up its own state independently |
| `time.Sleep` in tests | Use channels, sync primitives, or `httptest` |
| Hard-coded file paths (`d:/tmp`) | Use `t.TempDir()` |
| Testing private functions directly | Test through the public API; refactor if private logic is complex |
| `log.Fatal` in tests | Use `t.Fatal` so the test framework cleans up properly |
| Shared global state between tests | Use `TestMain` setup + per-test helpers |
| Large test functions testing many things | One test per behaviour; table-drive variations |

---

## Recommended Libraries

The standard library covers 90% of needs. When it doesn't:

- **`github.com/stretchr/testify`** — `assert` and `require` packages for cleaner assertions. `require` stops the test immediately (like `t.Fatal`); `assert` continues.
- **`github.com/matryer/is`** — lightweight alternative to testify, single import.

For this project, standard library + table-driven tests is a good default. Add testify if the assertion boilerplate becomes noisy.

---

## Applying TDD to the HTTP/KV Layer

Suggested sequence for adding a new endpoint (e.g. `PUT /keys/{key}`):

1. Write a failing handler test using `httptest` with a fake store.
2. Define the `Store` interface method signature.
3. Implement the handler to pass the test.
4. Write a failing integration test against the real store.
5. Implement the store method to pass the integration test.
6. Refactor.

This inside-out approach (handler → interface → store) keeps each layer independently testable and avoids premature coupling.
