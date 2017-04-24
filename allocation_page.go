package dbase

import (
	"encoding/binary"
	"sync"
)

const (
	allocationBitMapLen    = 8 * 1000 // allocation table can hold 8 * 1000 * 8 = 64,000 allocation bits
	allocationBitMapOffset = pageHeaderLength
)

// AllocationPage tracks page allocations in a []byte map
type AllocationPage interface {
	Page
	GetAllocationBitMap() AllocationBitMap
}
type allocationPage struct {
	l *sync.Mutex
	page
	allocationBitMapBytes []byte
}

//NewAllocationPage returns a new allocation page with no allocation bits set
func NewAllocationPage() AllocationPage {
	page := &allocationPage{
		page: page{
			id:       0,
			pagetype: pageTypeAllocationMap,
			bytes:    make([]byte, PageSize, PageSize),
		},
		l: &sync.Mutex{},
	}
	page.header = page.bytes[0:pageHeaderLength]
	page.allocationBitMapBytes = page.bytes[allocationBitMapOffset : allocationBitMapOffset+allocationBitMapLen]
	for i := range page.allocationBitMapBytes {
		page.allocationBitMapBytes[i] = 0
	}
	return page
}

func (page *allocationPage) GetAllocationBitMap() AllocationBitMap {
	return NewAllocationBitMap(page.allocationBitMapBytes)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// The page is encoded as a []byte PAGE_SIZE long, ready for serialisation.
func (page *allocationPage) MarshalBinary() ([]byte, error) {
	page.l.Lock()
	defer page.l.Unlock()
	binary.LittleEndian.PutUint64(page.header[pageIDOffset:], uint64(page.id))
	page.header[pageTypeOffset] = byte(pageTypeAllocationMap)
	return page.bytes, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// PAGE_SIZE bytes are used to rehydrate the page.
func (page *allocationPage) UnmarshalBinary(buf []byte) error {
	page.l.Lock()
	defer page.l.Unlock()
	if len(buf) != int(PageSize) {
		panic("Invalid buffer")
	}
	// check page type, panic if wrong
	pageType := PageType(buf[pageTypeOffset])
	if pageType != pageTypeAllocationMap {
		panic("Invalid page type")
	}
	copy(page.bytes, buf)
	page.header = page.bytes[0:pageHeaderLength]
	page.id = PageID(binary.LittleEndian.Uint64(page.header[pageIDOffset:]))
	page.pagetype = pageTypeAllocationMap
	page.allocationBitMapBytes = page.bytes[allocationBitMapOffset : allocationBitMapOffset+allocationBitMapLen]
	return nil
}
