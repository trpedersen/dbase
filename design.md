
// DB is the main interface for accessing data
type DB interface {
}

// PageStore is the primary interface for types that store pages.
type PageStore interface {
	Read(id PageID, page Page) error
	Write(id PageID, page Page) error
	Allocate(runSize int) (PageID, error)
	Deallocate(id PageID, runSize int)
	Count() int64
	Statistics() string
}

type PageMap interface {
	Page
}

// FileStore is a file-backed page store
type FileStore interface {
	PageStore
	Path() string
}

// MemoryStore is an in-memory implementation of a page store - useful for testing.
type MemoryStore interface {
	PageStore
}


// Page is the main abstraction that other page types inherit from
type Page interface {
	// GetID returns the page ID for this page
	GetID() PageID
	// SetID sets the page ID for this page
	SetID(id PageID) error
	// GetType returns the type of the page
	GetType() PageType
	// MarshalBinary returns a byte array representing the page
	MarshalBinary() ([]byte, error)
	// UnmarshalBinary rehydrates a page from a byte array
	UnmarshalBinary(buf []byte) error

	frame #, pageid, pin_count, dirty
}

// HeapPage is a page that contains records stored in slots.
type HeapPage interface {
	Page
	//PreviousPageID() PageID
	//NextPageID() PageID
	GetSlotCount() int16
	AddRecord(buf []byte) (int16, error)
	GetRecord(slot int16, buf []byte) (int, error)
	GetRecordLength(slot int16) (int, error)
	SetRecord(slot int16, buf []byte) error
	DeleteRecord(slot int16) error
	GetFreeSpace() int
	Clear() error
}