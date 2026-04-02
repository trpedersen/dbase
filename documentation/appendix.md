# File Appendix

This appendix gives a file-by-file summary of the repository so it is easier to navigate the codebase quickly.

## Core Abstractions

### `db.go`

Defines the top-level `DB` interface, which is currently empty. This is more of a placeholder for future database-level behavior than an active part of the system.

### `page.go`

Defines the core page model:

- fixed page size
- page IDs
- page types
- the base `Page` interface
- the shared `page` implementation used by concrete page types

This is the foundational storage file in the codebase.

### `page_store.go`

Defines the `PageStore` interface used by higher-level code. This is the abstraction that decouples logical storage structures from their persistence backend.

## Storage Backends

### `file_store.go`

Implements a file-backed `PageStore`. Pages are stored at fixed offsets in a single file. This is the main durable storage backend.

### `memory_store.go`

Implements an in-memory `PageStore` for tests and experiments. It mirrors the behavior of the file store without using disk.

### `buffered_page_store.go`

Placeholder for a future buffered page store. It is currently only a stub.

## Heap Storage

### `heap.go`

Defines the `Heap` interface and its main implementation. This is the primary record-oriented abstraction in the repository.

### `heap_page.go`

Implements slotted heap pages used to store variable-length records. This is one of the most important files in the project.

### `heap_header_page.go`

Implements the heap metadata page that tracks total record count and the current last heap page.

### `heap_scanner.go`

Implements sequential iteration over heap records by walking heap pages and slots.

### `record.go`

Placeholder for a structured record abstraction. It is currently empty.

## Page Allocation and Space Management

### `allocation_bitmap.go`

Implements a bitmap abstraction for tracking allocation state. It appears to be intended for future page allocation management.

### `allocation_page.go`

Defines a page type that stores allocation bitmap bytes. Present, but not integrated into the main storage flow.

### `page_directory.go`

Defines a `PageDirectory` interface for page allocation and deallocation. Interface only; no concrete implementation currently exists.

## Overflow and Alternate Page Types

### `overflow_page.go`

Defines an overflow page structure for large records that do not fit within a heap page. The type exists, but the heap does not yet fully use overflow pages in practice.

### `db_header_page.go`

Contains commented-out code for a possible database header page design. This appears to be an earlier experiment or abandoned direction.

## Tests

### `file_store_test.go`

Tests file store creation, reopening, page creation, and basic page read/write behavior.

### `memory_store_test.go`

Contains some scaffolding for memory store tests, but much of it remains unimplemented.

### `page_test.go`

Tests basic page identity and type behavior.

### `heap_page_test.go`

Tests heap page behavior directly, including record storage mechanics.

### `heap_test.go`

Exercises heap creation, inserts, deletes, scans, and some higher-level heap workflows.

### `allocation_bitmap_test.go`

Tests the bitmap allocation logic.

### `overflow_page_test.go`

Tests the overflow page type.

### `file_store_test.go` and `memory_store_test.go`

Validate the two page store implementations.

## Executable Harness

### `main/main.go`

Provides a small executable used as a manual test and profiling harness rather than a production CLI.

## Existing Docs

### `README.md`

Short project landing page and orientation document.

### `overview.md`

High-level codebase overview.

### `design.md`

Technical design notes and architecture walkthrough.

### `appendix.md`

This file. A file-by-file reference.

## Suggested Reading Order

If you are trying to understand the codebase efficiently, read in this order:

1. `README.md`
2. `overview.md`
3. `design.md`
4. `page.go`
5. `page_store.go`
6. `file_store.go`
7. `heap.go`
8. `heap_page.go`
9. `heap_scanner.go`
10. relevant tests