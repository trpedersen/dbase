package dbase

import (
	"encoding/binary"
)

type HeapHeaderPage struct {
	page
	lastPageId  PageID // total number of pages in file
	recordCount int64
}

const (
	HEAP_LAST_PAGE_ID_START = 9
	HEAP_LAST_PAGE_ID_END   = 17
	HEAP_RECORD_COUNT_START = 17
	HEAP_RECORD_COUNT_END   = 25
)

func NewHeapHeaderPage() *HeapHeaderPage {
	page := &HeapHeaderPage{
		page: page{
			id:       0,
			pagetype: HEAP_HEADER_PAGE,
			bytes:    make([]byte, PAGE_SIZE, PAGE_SIZE),
		},
		lastPageId:  1,
		recordCount: 0,
	}
	page.header = page.bytes[0:PAGE_HEADER_LEN]
	return page
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// The page is encoded as a []byte PAGE_SIZE long, ready for serialisation.
func (page *HeapHeaderPage) MarshalBinary() ([]byte, error) {
	page.pagetype = HEAP_HEADER_PAGE
	page.page.Marshal()
	binary.LittleEndian.PutUint64(page.header[HEAP_LAST_PAGE_ID_START:], uint64(page.lastPageId))
	binary.LittleEndian.PutUint64(page.header[HEAP_RECORD_COUNT_START:], uint64(page.recordCount))
	return page.bytes, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// PAGE_SIZE bytes are used to rehydrate the page.
func (page *HeapHeaderPage) UnmarshalBinary(buf []byte) error {

	if len(buf) != int(PAGE_SIZE) {
		panic("Invalid buffer")
	}
	// check page type, panic if wrong
	pageType := buf[PAGE_TYPE_BYTE]
	if pageType&HEAP_HEADER_PAGE == 0 {
		panic("Invalid page type")
	}
	page.page.Unmarshal(buf)
	page.lastPageId = PageID(binary.LittleEndian.Uint64(page.header[HEAP_LAST_PAGE_ID_START:]))
	page.recordCount = int64(binary.LittleEndian.Uint64(page.header[HEAP_RECORD_COUNT_START:]))
	return nil
}
