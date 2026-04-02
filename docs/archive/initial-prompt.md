# Prompt for Claude Code — dbase project setup

Paste everything below the line into Claude Code.

---

I need you to set up the project memory and documentation scaffolding for this Go project. Do NOT write any implementation code or tests yet. This is purely about establishing the project's specification, conventions, and development methodology docs.

## Project context

This is `dbase`, a Go project implementing a simple key-value storage engine. It originated ~9 years ago as a learning exercise porting concepts from minibase (a Java teaching database) to Go. The code has been modernised to current Go but is incomplete. We are now resuming development properly.

The owner is an experienced full-stack developer (strong in C#, SQL/T-SQL, software engineering fundamentals) who is rusty with Go. He is NOT interested in vibe-coding. The goal is rigorous, maintainable, well-engineered Go code developed using a disciplined methodology.

## What I need you to create

### 1. CLAUDE.md (project root)

Create a CLAUDE.md that establishes:

**Project identity**
- Project name, one-line purpose, current status (resuming active development)

**Development methodology — this is critical**

We follow a strict spec-first, interface-first, TDD workflow:

1. **Specification**: Before implementing any feature or module, its design must be documented in `docs/`. The spec captures intent, invariants, data structures, error handling, and edge cases. Specs are living documents updated as understanding evolves.

2. **Interface design**: Key abstractions are defined as Go interfaces before implementation begins. Interfaces live in the package they belong to and are the primary contract other packages depend on.

3. **Test-driven development**: Tests are written BEFORE implementation. Tests are the executable specification. Use standard Go `testing` package with table-driven tests as the default pattern. No BDD frameworks.

4. **Implementation**: Code is written to pass existing tests. Follow the spec. If the spec is ambiguous, ASK — don't assume.

5. **Review and refactor**: After tests pass, review for clarity, idiomatic Go, and alignment with the spec. Refactor with tests green.

**The key rule: if there is no spec for it, don't implement it. If the spec is unclear, clarify first.**

**Go conventions for this project**
- Go 1.22+ (or whatever version is in go.mod — check and use the actual version)
- Standard library first. Only add dependencies when genuinely justified.
- `errors` package for error handling, wrap with `fmt.Errorf("context: %w", err)` 
- No `panic` except for genuine programming errors (not runtime conditions)
- Exported types and functions must have doc comments
- Package names: short, lowercase, single-word where possible
- File organisation: one primary type per file where practical
- Test files colocated with source (`foo_test.go` alongside `foo.go`)
- Table-driven tests as the default pattern
- Avoid `init()` functions
- Use `context.Context` for operations that may block or need cancellation

**Code quality rules**
- All code must pass `go vet` and `golangci-lint` (if configured) before being considered complete
- Tests must pass. No "we'll fix this later" test failures.
- No TODO/FIXME without a corresponding issue or doc reference
- Commit messages: imperative mood, concise, reference spec docs where applicable

**Documentation conventions**
- Architecture and design docs live in `docs/`
- Architecture Decision Records live in `docs/decisions/`
- ADR format: number-title.md (e.g., `001-storage-model.md`)
- Specs are markdown. They should be readable by both humans and by you (Claude) as context for implementation work.

**How to work with me (the developer)**
- Before implementing anything non-trivial, summarise your understanding of the relevant spec back to me
- If a spec is missing or ambiguous, say so and propose a clarification — don't fill gaps with assumptions  
- When writing tests, explain what behaviour each test case is verifying and why
- When I say "implement X", check for a spec first. If there isn't one, remind me we need one.
- Prefer small, reviewable changes over large rewrites

### 2. docs/ directory structure

Create the following files with skeleton content. Each should have a clear heading, a "Status: Draft" marker, and placeholder sections that we will fill in together. Don't invent detailed designs — use placeholders and questions that prompt us to make decisions.

```
docs/
  README.md              — index of all docs, how to navigate them
  architecture.md        — high-level module decomposition, package dependency graph, key interfaces
  development-guide.md   — how to build, test, lint; the TDD workflow in practice
  decisions/
    README.md            — what ADRs are, how we use them, index of decisions
    000-template.md      — ADR template (context, decision, consequences)
```

For `architecture.md`, look at the existing code in the repo and produce an honest inventory:
- What packages exist today?
- What do they appear to do?
- What's their current state (complete, partial, stub, dead code)?
- What interfaces or types are already defined?
- What's missing or unclear?

Frame this as a "current state" section, not a target architecture. We will define the target architecture together.

For `development-guide.md`, include practical instructions:
- How to run tests (`go test ./...`)
- How to vet/lint
- The TDD loop: spec → interface → test → implement → refactor
- A note that we use CLAUDE.md as the source of truth for project conventions

### 3. Do NOT do any of the following
- Do not refactor or modify any existing source code
- Do not write any implementation code
- Do not write any test code
- Do not create any Go source files
- Do not make design decisions about the storage engine, data structures, or algorithms — those are for the spec phase

### 4. After creating the files

Give me a summary of:
1. What you created and where
2. Your honest assessment of the current codebase state based on what you read
3. What you think the most important design decisions are that we need to make before writing any new code
4. Suggested next steps (which specs to write first)