package dbase


type PageStore interface {
	Get(id pageid, page Page) error
	Set(id pageid, page Page) error
	Allocate() (pageid, error)
	Deallocate(id pageid) error
	Count() int64
}
