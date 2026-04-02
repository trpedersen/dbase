// Package dbase implements a page-based storage engine.
//
// The storage model is built around fixed-size pages (8 KB). Pages are
// persisted through a [PageStore], which can be file-backed ([FileStore])
// or in-memory ([MemoryStore]).
//
// Records are stored in a [Heap], which manages a sequence of [HeapPage]
// instances. Each record is identified by a [RID] (record ID) combining
// a page ID and slot number. The [HeapScanner] provides sequential
// iteration over all stored records.
//
// Large records that exceed the heap page payload are stored as linked
// chains of [OverflowPage] instances.
//
// Page allocation within a store is tracked by [AllocationBitMap] and
// [AllocationPage].
package dbase
