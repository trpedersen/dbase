# Architecture Decision Records

**Status**: Active

This directory contains Architecture Decision Records (ADRs) for the dbase project.

---

## What is an ADR?

An ADR is a short document that captures a significant architectural or design decision, the context that motivated it, and the consequences of making it. ADRs are written *before* implementation, not after.

We write ADRs when:
- A decision is hard to reverse (data formats, public interfaces, concurrency model)
- A decision will affect multiple parts of the system
- There were real alternatives considered and the choice isn't obvious
- We want a record of *why* the code is the way it is, not just *what* it does

We do **not** write ADRs for routine implementation choices.

---

## Format

Files are named `NNN-short-title.md` where NNN is a zero-padded sequence number starting at 001. Use `000-template.md` as the starting point.

Statuses:
- **Draft** — being written or discussed
- **Accepted** — decided, implementation may begin
- **Superseded by NNN** — replaced by a later decision
- **Rejected** — considered but not adopted (keep for reference)

---

## Index

| # | Title | Status |
|---|-------|--------|
| [000](000-template.md) | Template | — |
| [001](001-error-handling.md) | Error handling — return errors, do not panic | Accepted |

*(Add new ADRs to this table as they are created)*
