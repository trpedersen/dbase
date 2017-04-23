package dbase

import (
	"encoding/binary"
)

// HeapHeaderPage is the first data page in a store, contains stats and pointers to important pages at known locations.
type HeapHeaderPage interface {
	Page
	GetRecordCount() int64
	SetRecordCount(count int64)
	GetLastPageID() PageID
	SetLastPageID(id PageID)
}

type heapHeaderPage struct {
	page
	lastPageID  PageID // total number of pages in file
	recordCount int64
}

const (
	heapLastPageIDOffset  = 9
	heapRecordCountOffset = 17
)

// NewHeapHeaderPage returns a new heap header page.
func NewHeapHeaderPage() HeapHeaderPage {
	page := &heapHeaderPage{
		page: page{
			id:       0,
			pagetype: pageTypeHeapHeader,
			bytes:    make([]byte, PageSize, PageSize),
		},
		lastPageID:  1,
		recordCount: 0,
	}
	page.header = page.bytes[0:pageHeaderLength]
	return page
}

func (page *heapHeaderPage) GetRecordCount() int64 {
	return page.recordCount
}

func (page *heapHeaderPage) SetRecordCount(count int64) {
	page.recordCount = count
}

func (page *heapHeaderPage) GetLastPageID() PageID {
	return page.lastPageID
}

func (page *heapHeaderPage) SetLastPageID(id PageID) {
	page.lastPageID = id
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// The page is encoded as a []byte PAGE_SIZE long, ready for serialisation.
func (page *heapHeaderPage) MarshalBinary() ([]byte, error) {

	binary.LittleEndian.PutUint64(page.header[pageIDOffset:], uint64(page.id))
	page.header[pageTypeOffset] = byte(pageTypeHeapHeader)

	binary.LittleEndian.PutUint64(page.header[heapLastPageIDOffset:], uint64(page.lastPageID))
	binary.LittleEndian.PutUint64(page.header[heapRecordCountOffset:], uint64(page.recordCount))
	return page.bytes, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// PAGE_SIZE bytes are used to rehydrate the page.
func (page *heapHeaderPage) UnmarshalBinary(buf []byte) error {

	if len(buf) != int(PageSize) {
		panic("Invalid buffer")
	}
	// check page type, panic if wrong
	pageType := PageType(buf[pageTypeOffset])
	if pageType != pageTypeHeapHeader {
		panic("Invalid page type")
	}
	copy(page.bytes, buf)
	page.header = page.bytes[0:pageHeaderLength]

	page.id = PageID(binary.LittleEndian.Uint64(page.header[pageIDOffset:]))
	page.pagetype = pageTypeHeapHeader

	page.lastPageID = PageID(binary.LittleEndian.Uint64(page.header[heapLastPageIDOffset:]))
	page.recordCount = int64(binary.LittleEndian.Uint64(page.header[heapRecordCountOffset:]))

	return nil
}
