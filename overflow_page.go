package dbase

import (
	"encoding/binary"
	"fmt"
	"sync"
	//"log"
)

// OverflowPage is used to store records that exceed maximum record length on a heap page.
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
	overflowPreviousIDOffset = 9
	overflowNextIDOffset     = 17
	overflowSegmentIDOffset  = 25
	overflowSegmentLenOffset = 29
	overflowSegmentOffset    = pageHeaderLength
	overflowSegmentLen       = PageSize - overflowSegmentOffset
	maxSegmentLen            = PageSize - pageHeaderLength
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
			pagetype: pageTypeOverflow,
			bytes:    make([]byte, PageSize, PageSize),
		},
		l:             &sync.Mutex{},
		segmentID:     -1,
		segmentLength: -1,
	}

	page.header = page.bytes[0:pageHeaderLength]
	page.segment = page.bytes[overflowSegmentOffset : overflowSegmentOffset+overflowSegmentLen]

	return page
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// The page is encoded as a []byte PAGE_SIZE long, ready for serialisation.
func (page *overflowPage) MarshalBinary() ([]byte, error) {

	page.l.Lock()
	defer page.l.Unlock()

	binary.LittleEndian.PutUint64(page.header[pageIDOffset:], uint64(page.id))
	page.header[pageTypeOffset] = byte(pageTypeOverflow)

	// TODO: overflow page fields
	binary.LittleEndian.PutUint64(page.header[overflowPreviousIDOffset:], uint64(page.previousID))
	binary.LittleEndian.PutUint64(page.header[overflowNextIDOffset:], uint64(page.nextID))

	binary.LittleEndian.PutUint32(page.header[overflowSegmentIDOffset:], uint32(page.segmentID))
	binary.LittleEndian.PutUint16(page.header[overflowSegmentLenOffset:], uint16(page.segmentLength))

	return page.bytes, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// PAGE_SIZE bytes are used to rehydrate the page.
func (page *overflowPage) UnmarshalBinary(buf []byte) error {

	page.l.Lock()
	defer page.l.Unlock()

	if len(buf) != int(PageSize) {
		panic("Invalid buffer")
	}
	// check page type, panic if wrong
	pageType := PageType(buf[pageTypeOffset])
	if pageType != pageTypeOverflow {
		panic("Invalid page type")
	}

	copy(page.bytes, buf)

	page.header = page.bytes[0:pageHeaderLength]

	page.id = PageID(binary.LittleEndian.Uint64(page.header[pageIDOffset:]))
	page.pagetype = pageTypeOverflow

	// TODO: overflow page fields
	page.previousID = PageID(binary.LittleEndian.Uint64(page.header[overflowPreviousIDOffset:]))
	page.nextID = PageID(binary.LittleEndian.Uint64(page.header[overflowNextIDOffset:]))

	page.segmentID = int32(binary.LittleEndian.Uint32(page.header[overflowSegmentIDOffset:]))
	page.segmentLength = int(binary.LittleEndian.Uint16(page.header[overflowSegmentLenOffset:]))
	page.segment = page.bytes[overflowSegmentOffset : overflowSegmentOffset+page.segmentLength]

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
	if len(buf) > int(maxSegmentLen) {
		return fmt.Errorf("Buffer length (%d) exceeds MAX_SEGMENT_LEN (%d)", len(buf), maxSegmentLen)
	}
	page.segmentID = segmentID
	page.segmentLength = len(page.segment)
	page.segment = page.segment[0:len(buf)]
	copy(page.segment, buf)
	return nil
}
