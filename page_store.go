package dbase

// PageStore is the primary interface for types that store pages.
type PageStore interface {
	Read(id PageID, page Page) error
	Write(id PageID, page Page) error
	New() (PageID, error)
	Append(page Page) (PageID, error)
	Count() int64
	Statistics() string
	Close() error
}
