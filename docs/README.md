# dbase Documentation

This directory contains architecture, design, and decision records for the dbase project.

---

## Index

| File | Contents |
|------|----------|
| [architecture.md](architecture.md) | Current codebase inventory; module decomposition; key interfaces; gaps and open questions |
| [development-guide.md](development-guide.md) | How to build, test, and lint; the TDD workflow in practice |
| [decisions/](decisions/) | Architecture Decision Records (ADRs) |

---

## How to Navigate

- **Starting out?** Read [architecture.md](architecture.md) for an honest picture of where the code currently stands.
- **Setting up?** Read [development-guide.md](development-guide.md) for build/test instructions and the TDD loop.
- **Making a significant decision?** Write an ADR in [decisions/](decisions/) before implementing. See [decisions/README.md](decisions/README.md).
- **Writing a spec?** Create a new `.md` file in `docs/` describing the feature. Follow the same structure as existing specs: intent, invariants, data structures, error handling, edge cases.

---

## Existing Context Files

The following files were written during earlier exploration and planning phases. They are retained as historical context but are **not authoritative** — the specs and ADRs that we write going forward are the source of truth.

| File | Notes |
|------|-------|
| `design.md` | Early design notes |
| `overview.md` | High-level overview draft |
| `planning.md` | Early planning notes |
| `architecture.md` | **Current state inventory** (authoritative — maintained going forward) |
| `modernisation.md` | Notes from the Go modernisation pass |
| `code-organisation-plan.md` | Notes from the package reorganisation |
| `tdd-with-go.md` | Reference on Go TDD conventions |
| `development-recommendations.md` | Earlier workflow recommendations |
