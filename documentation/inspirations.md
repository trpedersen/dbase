# Inspirations

This codebase follows a well-established educational lineage for implementing a storage manager from scratch. The patterns here are not invented — they are the canonical patterns taught in database systems courses and textbooks going back decades.

## Primary Sources

### Ramakrishnan & Gehrke — "Database Management Systems" (3rd ed.)

The foundational textbook behind this codebase. Every major component maps to a chapter:

- **Chapter 9** — Storing Data: Disks and Files — page layout, fixed-size pages, heap files
- **§9.6** — Slotted page format: records grow forward, slot table grows backward, free-space tracking
- **§9.7** — Large records and overflow pages as linked lists
- **§9.4** — Free-space management via allocation bitmaps
- **§9.5** — Page directories tracking free space per page
- **Chapter 10** — Buffer Manager: pin/unpin, dirty pages, replacement policies

The book is sometimes called "the Cow Book" (after its cover). A free first-edition PDF is available from the University of Wisconsin: https://pages.cs.wisc.edu/~dbbook/

### CMU 15-445/645 — Database Systems (Andy Pavlo)

Andy Pavlo's course at Carnegie Mellon uses Ramakrishnan & Gehrke as its textbook and is the best freely available video lecture series on database internals. Course materials and lecture slides are published openly each semester at https://15445.courses.cs.cmu.edu/

The course project sequence — Buffer Pool Manager → Heap File → B+ Tree Index → Query Execution — maps directly onto the architecture being built here.

### Minibase — University of Wisconsin-Madison

Minibase is the companion educational DBMS written to accompany the Ramakrishnan & Gehrke book. Its `HFPage` (Heap File Page) is essentially this project's `heapPage`: slot table, variable-length records, free-space management by the same method. Given that R&G was written by Wisconsin faculty, the connection is direct.

Minibase home page: https://pages.cs.wisc.edu/~dbbook/openAccess/Minibase/minibase.html

### SimpleDB — Edward Sciore, "Database Design and Implementation" (2nd ed., Springer 2020)

A second educational lineage. SimpleDB uses the same RID + slot-directory heap page pattern and the same abstraction of a page store beneath a heap. A good read if you want a different author's take on the same ideas.

## Pattern-by-Pattern Mapping

| This codebase | Canonical source |
|---|---|
| 8 KB fixed-size pages | Standard in R&G, CMU, Minibase |
| 56-byte page header (PageID, PageType) | R&G Chapter 9 storage layout |
| Slot directory — records forward, slots backward | R&G §9.6 · CMU Lecture 3 · Minibase HFPage |
| Slot 0 as free-space sentinel | Minibase / CMU BusTub convention |
| RID = PageID + slot number | R&G, Minibase, SimpleDB — universal |
| Compaction on delete | R&G, Minibase |
| OverflowPage as doubly-linked list with segmentID | R&G §9.7 · PostgreSQL TOAST · InnoDB |
| Allocation bitmap — 64,000 bits per page | R&G §9.4 · SQL Server IAM pages |
| HeapHeaderPage at page 0 | Standard heap file header convention |
| PageStore with FileStore + MemoryStore | R&G storage manager abstraction |

## Further Reading

- **"Architecture of a Database System"** — Hellerstein, Stonebraker & Hamilton (2007). A concise survey of how the major layers of a real DBMS fit together. Free PDF: https://dsf.berkeley.edu/papers/fntdb07-architecture.pdf
- **BuzzDB** (Georgia Tech) — another open educational DBMS in C++: https://github.com/jarulraj/buzzdb
- **CMU BusTub** — the actual C++ teaching DBMS used in 15-445: https://github.com/cmu-db/bustub
