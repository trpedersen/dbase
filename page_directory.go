package dbase

// PageDirectory maintains a directory of allocated & deallocated pages
type PageDirectory interface {
	AllocatePage() (PageID, error)
	DeallocatePage(id PageID) error
	Count() int64
	AllocatedCount() int64
}
