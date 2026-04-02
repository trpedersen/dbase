# Architecture

**Status**: Draft — current state inventory only. Target architecture to be defined via ADRs and specs.

---

## Purpose

This document describes the **current state** of the dbase codebase as observed in April 2026. It is an honest inventory, not a design goal. Use it as a baseline before designing new features.

---

## Packages

The codebase is a single Go module (`github.com/trpedersen/dbase`) with two packages:

| Package | Location | Purpose |
|---------|----------|---------|
| `dbase` | repo root | Storage engine — all page types, stores, heap, scanner, allocation |
| `main` | `cmd/dbase/` | Binary entry point — currently a stub |

There is one external dependency: `github.com/trpedersen/rand` — a seeded PRNG used in tests for generating deterministic test data.

---

## Current State: Type and Interface Inventory

### Storage abstractions

| Type | Kind | State | Notes |
|------|------|-------|-------|
| `Page` | interface | Complete | Core page abstraction: ID, type, marshal/unmarshal |
| `PageStore` | interface | Complete | Read, Write, New, Append, Count, Statistics, Close |
| `FileStore` | interface + impl | Complete | File-backed store; extends PageStore with `Path()` |
| `MemoryStore` | interface + impl | Complete | In-memory store; used in tests |
| `BufferedPageStore` | struct stub | Not started | Empty struct; intended to buffer writes in front of a PageStore |

**Open question**: What is the write-buffering strategy for `BufferedPageStore`? Dirty-page LRU cache? Fixed ring? This needs a spec before implementation.

### Page types

| Type | Kind | State | Notes |
|------|------|-------|-------|
| `page` (unexported) | struct | Complete | Base page: id, type, header bytes, raw bytes |
| `HeapPage` | interface + impl | Complete | Slot-table layout; AddRecord, GetRecord, SetRecord, DeleteRecord, GetFreeSpace, compact |
| `HeapHeaderPage` | interface + impl | Complete | Page 0 of a heap; holds record count + last-page pointer |
| `OverflowPage` | interface + impl | Partial | Linked-chain segments for records > page payload; implemented but **not integrated into Heap** |
| `AllocationPage` | interface + impl | Partial | Wraps AllocationBitMap; implemented but **not connected to any store** |

### Record management

| Type | Kind | State | Notes |
|------|------|-------|-------|
| `Heap` | interface + impl | Mostly complete | Put, Get, Set, Delete, Clear, Count; uses PageStore |
| `HeapScanner` | interface + impl | Complete | State-machine iterator; returns records via `Next(buf)` |
| `RID` | struct | Complete | Record ID = PageID + Slot number |
| `Record` | struct stub | Not started | Empty struct placeholder |

**Open question**: `Record` is a stub. What is the data model? Raw bytes only? Typed fields? Schema? This is a fundamental design decision that needs an ADR before anything above the heap layer is built.

### Allocation management

| Type | Kind | State | Notes |
|------|------|-------|-------|
| `AllocationBitMap` | interface + impl | Complete | 64,000-bit map; Allocate, AllocateExplicit, Deallocate, IsAllocated |
| `AllocationPage` | interface + impl | Partial | Wraps bitmap; serialises to a page — but nothing uses it |
| `PageDirectory` | interface only | Not started | Interface defined: AllocatePage, DeallocatePage, Count, AllocatedCount — **no implementation** |

**Critical gap**: `FileStore.New()` appends a blank page without consulting any allocation structure. There is no free-page recycling. Deleted records leave slots marked deleted; deleted pages are never reclaimed. `AllocationBitMap` and `AllocationPage` exist but are not wired into anything.

### Top-level

| Type | Kind | State | Notes |
|------|------|-------|-------|
| `DB` | interface stub | Not started | Empty interface; intended as the top-level API |

---

## Gaps and Known Issues

### Not integrated
- `OverflowPage` exists but `Heap.Put()` does not use it. A record larger than a heap page payload returns an error. There is no transparent large-record support.
- `AllocationBitMap` / `AllocationPage` / `PageDirectory` are unconnected. Free-space management does not exist at runtime.
- `HeapScanner` is commented out of the `Heap` interface (`//Scanner() HeapScanner`). It exists and appears functional but is not part of the public API.

### Design issues
- `panic` is used in `UnmarshalBinary` implementations and in `NewHeap` for I/O errors. These are runtime conditions, not programming errors, and should return `error`. Changing this cascades to callers — it is a deliberate architectural change, not a cleanup task.
- Statistics counters (`gets`, `sets`, etc.) on `fileStore`, `memoryStore`, and `heap` are updated without synchronisation. They will give incorrect results under concurrent access.
- `fileStore.Write` has `file.Sync()` commented out. Writes are not flushed to disk. There are no durability guarantees.
- `fileStore` does not take a file lock on open (a TODO in the code). Concurrent processes opening the same file is unsafe.

### Missing layers (not yet designed)
- No key-value store abstraction above the heap
- No index structures (B-tree, hash table, etc.)
- No transaction or WAL support
- No HTTP/REST/JSON server
- No schema or data model beyond raw `[]byte`

---

## Package Dependency Graph (current)

```
cmd/dbase/main
    └── (no imports from dbase yet — stub only)

dbase (root package)
    ├── page.go            — Page interface, page struct, PageID, PageType, constants
    ├── page_store.go      — PageStore interface
    ├── file_store.go      — FileStore interface + fileStore implementation
    ├── memory_store.go    — MemoryStore interface + memoryStore implementation
    ├── buffered_page_store.go — BufferedPageStore stub
    ├── heap_page.go       — HeapPage interface + heapPage impl; RID, error types
    ├── heap_header_page.go — HeapHeaderPage interface + impl
    ├── overflow_page.go   — OverflowPage interface + overflowPage impl
    ├── allocation_bitmap.go — AllocationBitMap interface + impl
    ├── allocation_page.go — AllocationPage interface + impl
    ├── page_directory.go  — PageDirectory interface only
    ├── heap.go            — Heap interface + heap impl
    ├── heap_scanner.go    — HeapScanner interface + heapScanner impl
    ├── db.go              — DB interface stub
    ├── record.go          — Record struct stub
    └── doc.go             — package doc comment
```

All production code is in a single flat package. There are no sub-packages.

---

## Open Questions (to resolve via ADRs or specs)

1. **Data model**: What does a "value" look like? Raw `[]byte`? A typed struct? Schema-on-read? Schema-on-write?
2. **Key-value API**: What is the public interface for get/put/delete? Synchronous? Streaming?
3. **Error handling strategy**: Should `UnmarshalBinary` panic or return an error? This affects every page type and their callers.
4. **Durability model**: Is `fsync` on every write acceptable? What is the consistency contract?
5. **Concurrency model**: Single-writer/multiple-reader? Full concurrent access? What is the locking strategy?
6. **Free-space management**: When should `AllocationBitMap` be used? Is `PageDirectory` the right abstraction?
7. **Large records**: Should overflow pages be transparent (handled inside Heap.Put) or explicit?
8. **HTTP layer**: REST? What does the API look like? Auth? Encoding?
9. **Package structure**: Should the key-value layer and HTTP layer live in sub-packages or a new module?
