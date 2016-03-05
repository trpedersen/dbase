package dbase

import (
	"encoding/binary"
)

type HeapHeaderPage interface {
	Page
	GetRecordCount() int64
	SetRecordCount(count int64)
	GetLastPageID() PageID
	SetLastPageID(id PageID)
}

type heapHeaderPage struct {
	page
	lastPageId  PageID // total number of pages in file
	recordCount int64
}

const (
	HEAP_LAST_PAGE_ID_OFFSET = 9
	HEAP_LAST_PAGE_ID_END    = 17
	HEAP_RECORD_COUNT_OFFSET = 17
	HEAP_RECORD_COUNT_END    = 25
)

func NewHeapHeaderPage() HeapHeaderPage {
	page := &heapHeaderPage{
		page: page{
			id:       0,
			pagetype: PAGE_TYPE_HEAP_HEADER,
			bytes:    make([]byte, PAGE_SIZE, PAGE_SIZE),
		},
		lastPageId:  1,
		recordCount: 0,
	}
	page.header = page.bytes[0:PAGE_HEADER_LEN]
	return page
}

func (page *heapHeaderPage) GetRecordCount() int64 {
	return page.recordCount
}

func (page *heapHeaderPage) SetRecordCount(count int64){
	page.recordCount = count
}

func (page *heapHeaderPage) GetLastPageID() PageID {
	return page.lastPageId
}

func (page *heapHeaderPage) SetLastPageID(id PageID) {
	page.lastPageId = id
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// The page is encoded as a []byte PAGE_SIZE long, ready for serialisation.
func (page *heapHeaderPage) MarshalBinary() ([]byte, error) {

	binary.LittleEndian.PutUint64(page.header[PAGE_ID_OFFSET:], uint64(page.id))
	page.header[PAGE_TYPE_OFFSET] = byte(PAGE_TYPE_HEAP_HEADER)

	binary.LittleEndian.PutUint64(page.header[HEAP_LAST_PAGE_ID_OFFSET:], uint64(page.lastPageId))
	binary.LittleEndian.PutUint64(page.header[HEAP_RECORD_COUNT_OFFSET:], uint64(page.recordCount))
	return page.bytes, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// PAGE_SIZE bytes are used to rehydrate the page.
func (page *heapHeaderPage) UnmarshalBinary(buf []byte) error {

	if len(buf) != int(PAGE_SIZE) {
		panic("Invalid buffer")
	}
	// check page type, panic if wrong
	pageType := PageType(buf[PAGE_TYPE_OFFSET])
	if pageType != PAGE_TYPE_HEAP_HEADER  {
		panic("Invalid page type")
	}
	copy(page.bytes, buf)
	page.header = page.bytes[0:PAGE_HEADER_LEN]

	page.id = PageID(binary.LittleEndian.Uint64(page.header[PAGE_ID_OFFSET:]))
	page.pagetype = PAGE_TYPE_HEAP_HEADER

	page.lastPageId = PageID(binary.LittleEndian.Uint64(page.header[HEAP_LAST_PAGE_ID_OFFSET:]))
	page.recordCount = int64(binary.LittleEndian.Uint64(page.header[HEAP_RECORD_COUNT_OFFSET:]))

	return nil
}
