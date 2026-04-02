# Codebase Overview

This repository is an experimental storage engine written in Go. It is not yet a full database. The implemented core is a page-based storage layer with a heap-file abstraction for storing variable-length records.

At a high level, the design is:

1. Fixed-size pages are the unit of storage.
2. A page store persists those pages either in memory or in a file.
3. A heap layer stores records inside heap pages using slots.
4. Tests exercise the storage and heap behavior more than any higher-level database API.

## Core Architecture

The main abstractions are centered on pages and stores:

- `Page` represents a serializable fixed-size page.
- `PageStore` abstracts persistence for pages.
- `Heap` provides a record-oriented interface on top of the page store.

The most important architectural choice is that everything is page-oriented. Records are not written directly to files. Instead, records are placed into in-memory page objects, and those pages are serialized through a `PageStore`.

Key files:

- `page.go`: defines page IDs, page types, the fixed page size, and the base page implementation.
- `page_store.go`: defines the `PageStore` interface used throughout the system.
- `db.go`: defines a `DB` interface, but it is currently empty, which suggests the project is still focused on storage-engine primitives rather than a full database API.

## Storage Backends

There are two concrete storage implementations:

- `file_store.go`: file-backed storage using fixed-size page offsets in a single file.
- `memory_store.go`: in-memory storage used mainly for testing and experimentation.

Both expose the same core operations:

- read a page
- write a page
- append a page
- create a new empty page
- count pages
- report simple statistics

That common interface keeps the heap logic independent from the underlying persistence mechanism.

## Record Storage Model

The most complete subsystem in the repository is the heap implementation:

- `heap.go`: high-level record API with `Put`, `Get`, `Set`, `Delete`, `Count`, `Clear`, and `Statistics`.
- `heap_page.go`: slotted-page implementation for variable-length records.
- `heap_header_page.go`: metadata page tracking record count and the last heap page.
- `heap_scanner.go`: sequential scanner over heap records.

Conceptually, the heap works like this:

1. Page 0 is a heap header page.
2. The header stores metadata such as total record count and the ID of the current last data page.
3. Records are appended into the current last heap page until there is no room.
4. When that page fills, a new heap page is appended and becomes the new last page.
5. Each record is addressed by an RID composed of `PageID` and `Slot`.

The slotted-page implementation in `heap_page.go` is the most database-like part of the project. It stores variable-length records inside an 8 KB page and uses a slot table to track offsets, lengths, and deletion state.

One notable detail is that slot 0 is reserved as a free-space descriptor rather than a user record slot.

## Read/Write/Delete Flow

### Write path

1. `Heap.Put` receives a byte slice.
2. The heap checks free space on the current last heap page.
3. If needed, it appends a new heap page.
4. The record is inserted into the page via the slot table.
5. The page is written back through the `PageStore`.
6. The heap header page is updated with the new record count and last page metadata.

### Read path

1. `Heap.Get` takes an RID.
2. It loads the target heap page from the `PageStore`.
3. It resolves the slot and copies the record bytes into the caller's buffer.

### Delete path

1. `Heap.Delete` loads the target page.
2. The slot is marked deleted.
3. The page is compacted so free space can be reused.

## Implemented vs. Exploratory Components

### Most developed

- file-backed page storage
- in-memory page storage
- heap file abstraction
- slotted heap page layout
- sequential heap scanning
- basic tests for storage and heap behavior

### Present but incomplete or not fully integrated

- `overflow_page.go`: defines overflow page structure for records too large to fit on a heap page, but the heap does not fully use this yet.
- `allocation_bitmap.go` and `allocation_page.go`: bitmap-based allocation tracking exists, but it is not integrated into the main heap flow.
- `page_directory.go`: interface only.
- `buffered_page_store.go`: stub only.
- `db_header_page.go`: commented-out earlier or abandoned direction.
- `record.go`: placeholder only.

This means the repository currently behaves more like a heap-file storage prototype than a full database engine.

## Tests and Usage

The tests are the clearest expression of intended behavior:

- `file_store_test.go`: validates file-backed page creation, reopen behavior, and page read/write.
- `heap_test.go`: exercises heap creation, repeated writes, delete semantics, and sequential scanning.
- `page_test.go`: checks basic page identity and type behavior.
- `memory_store_test.go`: mostly scaffolded and incomplete.

There is also an executable entry point in `main/main.go`, but it acts more like a manual test and profiling harness than a user-facing application.

## Current Scope

The codebase is best understood as a prototype for learning and experimenting with database internals. The current focus is on:

- page layout
- page persistence
- heap storage
- record addressing
- low-level storage mechanics

It is not yet focused on higher-level database features such as:

- schemas
- query processing
- indexing
- transactions
- recovery
- concurrency control beyond basic mutex protection

## Summary

In one sentence: this repository is a Go prototype of a page-based heap-file storage engine, with working file and memory stores, working slotted pages for records, and several future database components scaffolded but not yet integrated.