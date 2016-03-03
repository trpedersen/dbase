package dbase

type PageStore interface {
	Get(id PageID, page Page) error
	Set(id PageID, page Page) error
	New() (PageID, error)
	Append(page Page) (PageID, error)
	Count() int64
}
