package dbase

import (
	"encoding/binary"
	"errors"
	"fmt"
	//"log"
)

const (
	SLOT_TABLE_ENTRY_LEN = 4 // bytes
	RECORD_COUNT_START     = 9
	RECORD_COUNT_END       = 17 // int64
	FREE_POINTER_START     = 17
	FREE_POINTER_END       = 19 // int16
)

type HeapPage struct {
	page
	recordCount int16 // 32:36
	freePointer int16 // 36:40

	slotTable   []byte
}

type Slot struct {
	offset int16
	length int16
}

type RID struct {
	PageID PageID
	Slot   int16
}

// NewPage returns a new page of size page.PAGE_SIZE
func NewDataPage() *HeapPage {

	page := &HeapPage{
		page: page{
			id:       0,
			pagetype: HEAP_PAGE,
			bytes:    make([]byte, PAGE_SIZE, PAGE_SIZE),
		},
	}
	page.header = page.bytes[0:PAGE_HEADER_LEN]
	page.slotTable = page.bytes[PAGE_HEADER_LEN:PAGE_HEADER_LEN]
	page.freePointer = PAGE_SIZE

	return page
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// The page is encoded as a []byte PAGE_SIZE long, ready for serialisation.
func (page *HeapPage) MarshalBinary() ([]byte, error) {

	page.page.Marshal()
	page.header[PAGE_TYPE_BYTE] = HEAP_PAGE
	binary.LittleEndian.PutUint16(page.header[RECORD_COUNT_START:], uint16(page.recordCount))
	binary.LittleEndian.PutUint16(page.header[FREE_POINTER_START:], uint16(page.freePointer))

	return page.bytes, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// PAGE_SIZE bytes are used to rehydrate the page.
func (page *HeapPage) UnmarshalBinary(buf []byte) error {

	if len(buf) != int(PAGE_SIZE) {
		panic("Invalid buffer")
	}
	// check page type, panic if wrong
	pageType := buf[8]
	if pageType& HEAP_PAGE == 0 {
		panic("Invalid page type")
	}

	page.page.Unmarshal(buf)

	page.recordCount = int16(binary.LittleEndian.Uint16(page.header[RECORD_COUNT_START:]))
	page.freePointer = int16(binary.LittleEndian.Uint16(page.header[FREE_POINTER_START:]))

	page.slotTable = page.bytes[PAGE_HEADER_LEN : PAGE_HEADER_LEN+page.recordCount*4]

	return nil
}

// GetRecordCount returns the number of records held in page.
func (page *HeapPage) GetRecordCount() int16 {
	return page.recordCount
}

func (page *HeapPage) setRecordTable(recordNumber int16, offset int16, recLen int16) error {
	tableOffset := recordNumber * SLOT_TABLE_ENTRY_LEN
	// resize recordTable
	len := len(page.slotTable)
	page.slotTable = page.slotTable[0 : len+ SLOT_TABLE_ENTRY_LEN] // add two more uint16 == 4 bytes
	binary.LittleEndian.PutUint16(page.slotTable[tableOffset:tableOffset+2], uint16(offset))
	binary.LittleEndian.PutUint16(page.slotTable[tableOffset+2:tableOffset+4], uint16(recLen))
	return nil
}

// GetFreeSpace return the amount of free space available to store a record (inclusive of any header fields.)
func (page *HeapPage) GetFreeSpace() int {
	result := int(page.freePointer) - (PAGE_HEADER_LEN + int(page.recordCount* SLOT_TABLE_ENTRY_LEN) + SLOT_TABLE_ENTRY_LEN) - 1 // free pointer -  bytes header fields - #records * 4 bytes per table entry - another table entry
	if result < 0 {
		result = 0
	}
	return result
}

// AddRecord adds record to page, using copy semantics. Returns record number for added record.
// Returns an error if insufficient page free space.
func (page *HeapPage) AddRecord(record []byte) (int16, error) {
	recLen := len(record)
	if recLen > page.GetFreeSpace() {
		return 0, errors.New("Record length exceeds free space")
	}

	offset := page.freePointer - int16(recLen)
	copy(page.bytes[offset:page.freePointer], record)
	page.freePointer = offset
	recordNumber := page.recordCount // NB 0-based
	page.recordCount += 1
	page.setRecordTable(recordNumber, offset, int16(recLen))
	return recordNumber, nil
}

// GetRecord returns record specified by recordNumber.
// Note: record numbers are 0 based.
func (page *HeapPage) GetRecord(recordNumber int16, buf []byte) error {
	// recordNumber is 0 based
	if recordNumber+1 > page.recordCount {
		return errors.New(fmt.Sprintf("Invalid record number: %d, record count: %d", recordNumber, page.recordCount))
	}
	tableOffset := recordNumber * SLOT_TABLE_ENTRY_LEN
	offset := binary.LittleEndian.Uint16(page.slotTable[tableOffset : tableOffset+2])
	len := binary.LittleEndian.Uint16(page.slotTable[tableOffset+2 : tableOffset+4])
	copy(buf, page.bytes[offset:offset+len])
	return nil
}

// SetRecord updates record specified by recordNumber.
// Note: record numbers are 0 based.
func (page *HeapPage) SetRecord(recordNumber int16, buf []byte) error {
	// recordNumber is 0 based
	if recordNumber+1 > page.recordCount {
		return errors.New(fmt.Sprintf("Invalid record number: %d, record count: %d", recordNumber, page.recordCount))
	}
	tableOffset := recordNumber * SLOT_TABLE_ENTRY_LEN
	offset := binary.LittleEndian.Uint16(page.slotTable[tableOffset : tableOffset+2])
	len := binary.LittleEndian.Uint16(page.slotTable[tableOffset+2 : tableOffset+4])
	copy(page.bytes[offset:offset+len], buf)

	// cases
	/*
		len < current_len
		len == current_len
			copy over the top
		len > current_len

	 */


	return nil
}

// Clear resets the page, removing any records. Record count is set to 0
func (page *HeapPage) Clear() error {
	page.slotTable = page.bytes[PAGE_HEADER_LEN:PAGE_HEADER_LEN]
	page.freePointer = PAGE_SIZE
	page.recordCount = 0
	return nil
}
