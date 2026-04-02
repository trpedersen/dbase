# Planning — Completing dbase

This document proposes a roadmap for completing the storage manager layer of dbase, following the architecture described in Ramakrishnan & Gehrke ("the Cow Book"), taught in CMU 15-445, and demonstrated in Minibase.

The codebase already has a solid foundation: fixed-size pages, a slotted heap page with compaction, overflow pages, an allocation bitmap, a heap layer with Put/Get/Set/Delete, and both a file-backed and in-memory page store. The stubs left behind (`BufferedPageStore`, `PageDirectory`, `HeapDirectory`, `HeapScanner`, `Record`, `DB`) tell the story of what comes next.

---

## Current State

| Component | Status |
|---|---|
| `Page` (base type, marshal/unmarshal) | Done |
| `FileStore` | Done |
| `MemoryStore` | Done |
| `HeapPage` (slot directory, compaction) | Done |
| `HeapHeaderPage` | Done |
| `OverflowPage` (linked list, segments) | Done — not yet wired into Heap |
| `AllocationBitmap` + `AllocationPage` | Done |
| `Heap` (Put/Get/Set/Delete) | Done — uses append-only, no free-space directory |
| `BufferedPageStore` | Stub |
| `PageDirectory` / `HeapDirectory` | Stub |
| `HeapScanner` | Stub |
| `Record` (typed fields) | Stub |
| `DB` (top-level interface) | Stub |

---

## Phase 1 — Buffer Pool Manager

**R&G Chapter 10 · CMU 15-445 Project #1**

This is the most important missing piece. Currently the heap reads and writes pages on every operation — no caching. The `BufferedPageStore` stub is the right place for this.

A buffer pool maintains a fixed pool of in-memory frames. Each frame holds one page. When a page is requested:
- If it is already in a frame (a hit), return it directly.
- If not (a miss), evict a frame using the replacement policy, load the requested page from the underlying store, and return it.

### Key concepts to implement

**Frame** — an in-memory slot holding one page plus metadata:
```
type frame struct {
    page     []byte
    pageID   PageID
    pinCount int
    dirty    bool
}
```

**Pin / Unpin** — a caller pins a frame before using it (preventing eviction) and unpins it when done. Dirty frames are written back to the underlying store on eviction.

**Replacement policy** — determines which unpinned frame to evict. Two standard choices:
- **Clock** — approximates LRU with a single reference bit per frame; cheap and practical. Minibase uses Clock.
- **LRU** — evicts the least-recently-used frame. Simpler to reason about but slightly more expensive.

Start with Clock. It is what Minibase uses and what R&G describes in detail.

### Interface sketch

```go
type BufferPool interface {
    // FetchPage pins and returns the frame for the given page.
    // Loads from the underlying store on a cache miss.
    FetchPage(id PageID) ([]byte, error)

    // UnpinPage unpins a frame. If dirty is true, the frame is
    // marked dirty and will be flushed on eviction.
    UnpinPage(id PageID, dirty bool) error

    // FlushPage writes a dirty page to the underlying store immediately.
    FlushPage(id PageID) error

    // NewPage allocates a new page in the underlying store and pins it.
    NewPage() (PageID, []byte, error)
}
```

The buffer pool wraps a `PageStore` (FileStore) and replaces the direct store calls in `Heap`.

### Acceptance criteria
- All existing heap tests pass with a buffered store substituted for MemoryStore.
- A test that repeatedly reads the same page shows only one underlying store read (cache hit).
- A test with a pool smaller than the working set evicts correctly and data is not lost.

---

## Phase 2 — Page Directory (Free Space Management)

**R&G §9.5 · Minibase `HFBufMgr` / directory pages**

Currently `Heap.Put` always appends to the last page. If a record is deleted, that space is never reused by future inserts. A page directory fixes this.

A page directory maps each heap page to a summary of its free space. This allows `Put` to find a page with enough room before resorting to appending a new one.

### Design

A directory is itself a chain of pages (directory pages), each holding an array of entries:

```go
type directoryEntry struct {
    pageID    PageID
    freeSpace uint16  // bytes free on that page
}
```

On `Put`:
1. Scan the directory for a page with `freeSpace >= len(record)`.
2. If found, insert there and update the directory entry.
3. If not found, allocate a new heap page (via the AllocationBitmap), append a directory entry for it, and insert.

On `Delete`:
1. Delete the record from its heap page (compaction already happens).
2. Update the directory entry for that page to reflect the new free space.

The `HeapDirectory` stub is the right home for this. The `AllocationBitmap` and `AllocationPage` are already in place for tracking which pages exist — they feed into the allocation side of this.

### Acceptance criteria
- After inserting and deleting records, a subsequent insert reuses the freed space rather than allocating a new page.
- Page count does not grow monotonically under insert/delete/insert workloads.

---

## Phase 3 — Overflow Page Integration

**R&G §9.7**

`OverflowPage` is implemented but not wired in. Currently `Heap.Put` returns an error if a record exceeds a page's max payload. There is also a TODO in `HeapPage.SetRecord` for this case.

### Design

When a record exceeds `maxRecordLen` (~8 KB minus headers):
1. Allocate as many overflow pages as needed (ceiling of `len(record) / maxSegmentLen`).
2. Write each segment to an overflow page, chaining them via `previousID` / `nextID`.
3. On the heap page, write a stub slot entry flagged `recordOnOverflow` containing the ID of the first overflow page (rather than the record data itself).

On `Get`:
1. Read the heap page slot. If the flag is `recordOnOverflow`, read the chain of overflow pages and reassemble the record.

The slot flags (`recordOnPage`, `recordOnOverflow`, `recordDeleted`) are already defined in `heap_page.go` — the structure is ready for this.

### Acceptance criteria
- A record larger than `maxRecordLen` can be Put, Got, Set, and Deleted without error.
- The overflow chain is correctly freed (all pages deallocated) on Delete.

---

## Phase 4 — Heap Scanner

**R&G §9.5 · Minibase `HeapFileScan`**

A scanner performs a sequential scan of all records in a heap — the foundation of a full table scan.

```go
type HeapScanner interface {
    // Next advances to the next live record. Returns io.EOF when exhausted.
    Next() (RID, error)

    // Get returns the current record into buf.
    Get(buf []byte) (int, error)

    // Close releases resources held by the scanner.
    Close() error
}
```

Implementation:
- Start at the first heap page (page 1, after the header).
- Iterate slots on the current page, skipping deleted slots.
- When a page is exhausted, advance to the next page via the page directory.
- Interacts with the buffer pool: pin the current page, unpin when moving on.

### Acceptance criteria
- A scanner over a heap returns every live record exactly once.
- Scanner correctly skips deleted records.
- Scanner works correctly when records span overflow pages.

---

## Phase 5 — Record / Schema Layer

**R&G Chapter 8**

Currently records are raw `[]byte`. A schema layer adds typed fields and fixed/variable-length record encoding.

### Schema

```go
type FieldType int

const (
    FieldTypeInt    FieldType = iota
    FieldTypeFloat
    FieldTypeString // variable length, up to a max
)

type Field struct {
    Name    string
    Type    FieldType
    MaxLen  int // for strings
}

type Schema struct {
    Fields []Field
}
```

### Record encoding

For a fixed-length schema (all ints and fixed-width strings), records are a flat byte array — straightforward to implement first. Variable-length fields require a small per-record header holding field offsets (the same slot-directory idea, applied at the record level).

```go
type Record struct {
    schema *Schema
    data   []byte
}

func (r *Record) GetInt(field string) (int64, error)
func (r *Record) GetString(field string) (string, error)
func (r *Record) SetInt(field string, v int64) error
func (r *Record) SetString(field string, v string) error
```

### Acceptance criteria
- A record can be constructed from a schema, serialised to `[]byte`, stored in a heap, retrieved, and deserialised with correct field values.

---

## Phase 6 — DB Top-Level Interface

The `DB` interface in `db.go` is currently empty. Once the above layers are in place, a minimal implementation could look like:

```go
type DB interface {
    // CreateHeap creates a new named heap (analogous to a table).
    CreateHeap(name string, schema Schema) (Heap, error)

    // OpenHeap opens an existing named heap.
    OpenHeap(name string) (Heap, error)

    // Close flushes all dirty pages and closes all resources.
    Close() error
}
```

The DB header page (`db_header_page.go`) and DB directory page (`page_directory.go`) would hold the mapping from heap names to their header page IDs.

---

## Suggested Order of Work

1. **Buffer Pool Manager** — everything else benefits from it; unlocks realistic performance testing.
2. **Page Directory** — makes the heap correct under delete/reinsert workloads.
3. **Overflow Page integration** — completes the heap's record size contract.
4. **Heap Scanner** — needed before any query execution can be built.
5. **Record / Schema layer** — moves from raw bytes to structured data.
6. **DB interface** — ties it all together.

---

## Testing Strategy

Follow the existing pattern: use `MemoryStore` (or a buffered wrapper over it) in tests to avoid disk I/O. Each phase should have its own `_test.go` file. For the buffer pool, also add a file-backed test to verify persistence across open/close cycles.
