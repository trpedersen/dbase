# dbase

A page-based key-value storage engine in Go.  
**Status**: Resuming active development — storage layer partially implemented, key-value and HTTP layers not yet started.

---

## Commands

```bash
# Build
go build ./...

# Test
go test ./...
go test ./... -count=1 -v

# Vet
go vet ./...

# Lint (if golangci-lint is installed)
golangci-lint run
```

---

## Development Methodology

We follow a strict **spec-first → interface-first → TDD** workflow. This is not optional.

### The cycle

1. **Specification** — Before implementing any feature or module, its design is documented in `docs/`. The spec captures intent, invariants, data structures, error handling, and edge cases. Specs are living documents; update them as understanding evolves.

2. **Interface design** — Key abstractions are defined as Go interfaces before implementation begins. Interfaces live in the package they belong to and are the primary contract other packages depend on.

3. **Test-driven development** — Tests are written **before** implementation. Tests are the executable specification. Use the standard `testing` package. Table-driven tests are the default pattern. No BDD frameworks.

4. **Implementation** — Code is written to make existing tests pass. Follow the spec. If the spec is ambiguous, **ask — don't assume**.

5. **Review and refactor** — After tests pass, review for clarity, idiomatic Go, and alignment with the spec. Refactor with tests green.

**The key rule: if there is no spec for it, don't implement it. If the spec is unclear, clarify first.**

---

## Go Conventions

- **Go version**: 1.26.1 (see `go.mod`)
- Standard library first. Only add dependencies when genuinely justified.
- Error handling: `errors` package, wrap with `fmt.Errorf("context: %w", err)`
- No `panic` except for genuine programming errors (impossible states, violated invariants). Never for runtime conditions (I/O errors, bad input).
- Exported types and functions must have doc comments.
- Package names: short, lowercase, single-word where possible.
- File organisation: one primary type per file where practical.
- Test files colocated with source (`foo_test.go` alongside `foo.go`).
- Table-driven tests as the default pattern.
- Avoid `init()` functions.
- Use `context.Context` for operations that may block or need cancellation.

---

## Code Quality Rules

- All code must pass `go vet` before being considered complete.
- Tests must pass. No "we'll fix this later" test failures.
- No TODO/FIXME without a corresponding note in a `docs/` spec or a referenced decision.
- Commit messages: imperative mood, concise (e.g., `add AllocationPage serialisation`).

---

## Documentation Conventions

- Architecture and design docs live in `docs/`.
- Architecture Decision Records live in `docs/decisions/`.
- ADR format: `NNN-title.md` (e.g., `001-storage-model.md`).
- Specs are Markdown. Write them to be readable by both humans and Claude as implementation context.

---

## How to Work With Me (Claude)

- Before implementing anything non-trivial, summarise your understanding of the relevant spec back to me.
- If a spec is missing or ambiguous, say so and propose a clarification — don't fill gaps with assumptions.
- When writing tests, explain what behaviour each test case is verifying and why.
- When I say "implement X", check for a spec first. If there isn't one, remind me we need one.
- Prefer small, reviewable changes over large rewrites.
- This file (`CLAUDE.md`) is the source of truth for project conventions.
