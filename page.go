// Package page provides low-level types and functions for managing data pages and records.
// Writing pages to disk is left to other packages.
// Page is an abstract type

package dbase

const (
	// PageSize is typically the same as the filesystem blocksize
	PageSize = int16(8192)

	pageHeaderLength = 56                          // bytes
	maxPagePayload   = PageSize - pageHeaderLength // Maximum page payload bytes

	// Page buffer offsets for page fields
	pageIDOffset   = 0
	pageTypeOffset = 8 // single bytes

	// Page types:
	dbHeaderPage          = PageType(0x01)
	dbDirectoryPage       = PageType(0x02)
	pageTypeHeap          = PageType(0x03)
	pageTypeHeapHeader    = PageType(0x04)
	pageTypeOverflow      = PageType(0x05)
	pageTypeAllocationMap = PageType(0x06)
)

// PageID is (usually) the same as the block number on disk
type PageID int64

// PageType is one of
// DB_HEADER_PAGE        = PageType(0x01)
// DB_DIRECTORY_PAGE     = PageType(0x02)
// PAGE_TYPE_HEAP        = PageType(0x03)
// PAGE_TYPE_HEAP_HEADER = PageType(0x04)
// PAGE_TYPE_OVERFLOW    = PageType(0x05)
type PageType byte

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
}

// Concrete implementation of a page
type page struct {
	id       PageID   // 0:8
	pagetype PageType // 8
	header   []byte
	bytes    []byte
}

// GetID returns the page ID for this page
func (page *page) GetID() PageID {
	return page.id
}

// SetID sets the page ID for this page
func (page *page) SetID(id PageID) error {
	page.id = id
	return nil
}

// GetType returns the page type, e.g. PAGE_TYPE_HEAP
func (page *page) GetType() PageType {
	return page.pagetype
}
