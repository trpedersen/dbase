# dbase

`dbase` is an experimental Go codebase for learning and prototyping database storage internals.

The current implementation is centered on a page-based storage engine with a heap-file layer for variable-length records. It is not yet a full database system, but it already contains the core mechanics for:

- fixed-size page serialization
- file-backed and in-memory page stores
- slotted heap pages for variable-length records
- record addressing by page ID and slot
- sequential heap scanning

## What It Is

At a high level, the codebase is organized like this:

1. `Page` is the fixed-size storage unit.
2. `PageStore` persists pages to either memory or disk.
3. `Heap` provides record-oriented operations on top of pages.
4. Tests and the small `main` program act as the primary usage examples.

## Current Focus

The repository is currently focused on low-level storage concerns:

- page layout
- page serialization
- heap-file record storage
- page allocation concepts
- storage engine experimentation

It is not yet focused on higher-level database features such as SQL, schemas, indexing, transactions, or recovery.

## Main Components

- `page.go`: shared page definitions, page size, page IDs, and page types
- `page_store.go`: storage abstraction used by higher-level structures
- `file_store.go`: file-backed implementation of `PageStore`
- `memory_store.go`: in-memory implementation of `PageStore`
- `heap.go`: record-oriented heap API
- `heap_page.go`: slotted-page implementation for storing records
- `heap_header_page.go`: heap metadata page
- `heap_scanner.go`: sequential heap record scanner

## Docs

- `overview.md`: high-level overview of the codebase
- `design.md`: more technical design notes and architecture diagrams
- `appendix.md`: file-by-file reference for the repository

## Project Status

This is best understood as a storage-engine prototype. Some subsystems are working and exercised by tests, while others are present only as scaffolding for future work.

Implemented or mostly implemented:

- file and memory page stores
- heap storage
- slotted heap pages
- record reads, writes, updates, deletes
- sequential scanning

Partially implemented or exploratory:

- overflow pages for large records
- allocation bitmap and allocation page machinery
- page directory concepts
- buffered page store
- higher-level database API

## Running Tests

Use the standard Go test command from the repository root:

```bash
go test ./...
```

## Intent

The long-term direction appears to be broader than a simple heap file. The code and notes suggest this repository is a stepping stone toward a more ambitious database design, with the current emphasis on getting the storage fundamentals correct first.


