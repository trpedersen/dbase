package dbase

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	//"log"
)

type OverflowPage interface {
	Page

	GetPreviousPageID() PageID
	SetPreviousPageID(id PageID)
	GetNextPageID() PageID
	SetNextPageID(id PageID)

	GetSegmentID() int32
	GetSegmentLength() int
	GetSegment(buf []byte) (int, error)
	SetSegment(segmentID int32, buf []byte) error
}

const (
	OVERFLOW_PREVIOUS_ID_OFFSET = 9
	OVERFLOW_NEXT_ID_OFFSET     = 17
	OVERFLOW_SEGMENT_ID_OFFSET  = 25
	OVERFLOW_SEGMENT_LEN_OFFSET = 29
	OVERFLOW_SEGMENT_OFFSET     = PAGE_HEADER_LEN
	OVERFLOW_SEGMENT_LEN        = PAGE_SIZE - OVERFLOW_SEGMENT_OFFSET
	MAX_SEGMENT_LEN             = PAGE_SIZE - PAGE_HEADER_LEN
)

type overflowPage struct {
	l *sync.Mutex

	page

	previousID PageID
	nextID     PageID

	segmentID     int32
	segmentLength int
	segment       []byte
}

// NewOverflowPage returns a new Overflow Page.
func NewOverflowPage() OverflowPage {

	page := &overflowPage{
		page: page{
			id:       0,
			pagetype: PAGE_TYPE_OVERFLOW,
			bytes:    make([]byte, PAGE_SIZE, PAGE_SIZE),
		},
		l:             &sync.Mutex{},
		segmentID:     -1,
		segmentLength: -1,
	}

	page.header = page.bytes[0:PAGE_HEADER_LEN]
	page.segment = page.bytes[OVERFLOW_SEGMENT_OFFSET : OVERFLOW_SEGMENT_OFFSET+OVERFLOW_SEGMENT_LEN]

	return page
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// The page is encoded as a []byte PAGE_SIZE long, ready for serialisation.
func (page *overflowPage) MarshalBinary() ([]byte, error) {

	page.l.Lock()
	defer page.l.Unlock()

	binary.LittleEndian.PutUint64(page.header[PAGE_ID_OFFSET:], uint64(page.id))
	page.header[PAGE_TYPE_OFFSET] = byte(PAGE_TYPE_OVERFLOW)

	// TODO: overflow page fields
	binary.LittleEndian.PutUint64(page.header[OVERFLOW_PREVIOUS_ID_OFFSET:], uint64(page.previousID))
	binary.LittleEndian.PutUint64(page.header[OVERFLOW_NEXT_ID_OFFSET:], uint64(page.nextID))

	binary.LittleEndian.PutUint32(page.header[OVERFLOW_SEGMENT_ID_OFFSET:], uint32(page.segmentID))
	binary.LittleEndian.PutUint16(page.header[OVERFLOW_SEGMENT_LEN_OFFSET:], uint16(page.segmentLength))

	return page.bytes, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// PAGE_SIZE bytes are used to rehydrate the page.
func (page *overflowPage) UnmarshalBinary(buf []byte) error {

	page.l.Lock()
	defer page.l.Unlock()

	if len(buf) != int(PAGE_SIZE) {
		panic("Invalid buffer")
	}
	// check page type, panic if wrong
	pageType := PageType(buf[PAGE_TYPE_OFFSET])
	if pageType != PAGE_TYPE_OVERFLOW {
		panic("Invalid page type")
	}

	copy(page.bytes, buf)

	page.header = page.bytes[0:PAGE_HEADER_LEN]

	page.id = PageID(binary.LittleEndian.Uint64(page.header[PAGE_ID_OFFSET:]))
	page.pagetype = PAGE_TYPE_OVERFLOW

	// TODO: overflow page fields
	page.previousID = PageID(binary.LittleEndian.Uint64(page.header[OVERFLOW_PREVIOUS_ID_OFFSET:]))
	page.nextID = PageID(binary.LittleEndian.Uint64(page.header[OVERFLOW_NEXT_ID_OFFSET:]))

	page.segmentID = int32(binary.LittleEndian.Uint32(page.header[OVERFLOW_SEGMENT_ID_OFFSET:]))
	page.segmentLength = int(binary.LittleEndian.Uint16(page.header[OVERFLOW_SEGMENT_LEN_OFFSET:]))
	page.segment = page.bytes[OVERFLOW_SEGMENT_OFFSET : OVERFLOW_SEGMENT_OFFSET+page.segmentLength]

	return nil
}

func (page *overflowPage) GetPreviousPageID() PageID {
	return page.previousID
}

func (page *overflowPage) SetPreviousPageID(id PageID) {
	page.previousID = id
}

func (page *overflowPage) GetNextPageID() PageID {
	return page.nextID
}

func (page *overflowPage) SetNextPageID(id PageID) {
	page.nextID = id
}

func (page *overflowPage) GetSegmentID() int32 {
	return page.segmentID
}

func (page *overflowPage) GetSegmentLength() int {
	return page.segmentLength
}

func (page *overflowPage) GetSegment(buf []byte) (int, error) {
	n := copy(buf, page.segment)
	return n, nil
}

func (page *overflowPage) SetSegment(segmentID int32, buf []byte) error {
	if len(buf) > int(MAX_SEGMENT_LEN) {
		return errors.New(fmt.Sprintf("Buffer length (%d) exceeds MAX_SEGMENT_LEN (%d)", len(buf), MAX_SEGMENT_LEN))
	}
	page.segmentID = segmentID
	page.segmentLength = len(page.segment)
	page.segment = page.segment[0:len(buf)]
	copy(page.segment, buf)
	return nil
}
