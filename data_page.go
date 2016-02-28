package dbase

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	RECORD_TABLE_ENTRY_LEN = 4 // bytes
	RECORD_COUNT_START     = 9
	RECORD_COUNT_END       = 17 // int64
	FREE_POINTER_START     = 17
	FREE_POINTER_END       = 19 // int16
)

type DataPage struct {
	page
	recordCount uint16 // 32:36
	freePointer uint16 // 36:40

	recordTable []byte
}

//// NewPage returns a new page of size page.PAGE_SIZE
//func NewDataPage() *DataPage {
//	page := &DataPage{
//		Page.bytes: make([]byte, PAGE_SIZE),
//	}
//	page.header = page.bytes[0:HEADER_LEN]
//	page.recordTable = page.bytes[HEADER_LEN:HEADER_LEN]
//	page.freePointer = PAGE_SIZE
//	return page
//}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// The page is encoded as a []byte PAGE_SIZE long, ready for serialisation.
func (page *DataPage) MarshalBinary() ([]byte, error) {

	page.page.MarshalBinary()
	page.header[PAGE_TYPE_BYTE] = DATA_PAGE
	binary.LittleEndian.PutUint16(page.header[RECORD_COUNT_START:RECORD_COUNT_END], page.recordCount)
	binary.LittleEndian.PutUint16(page.header[FREE_POINTER_START:FREE_POINTER_END], page.freePointer)

	return page.bytes, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// PAGE_SIZE bytes are used to rehydrate the page.
func (page *DataPage) UnmarshalBinary(buf []byte) error {

	if len(buf) != int(PAGE_SIZE) {
		panic("Invalid buffer")
	}
	// check page type, panic if wrong
	pageType := buf[8]
	if pageType & DATA_PAGE == 0 {
		panic("Invalid page type")
	}

	page.page.UnmarshalBinary(buf)

	//page.prevId = pageid(binary.LittleEndian.Uint64(page.header[8:16]))
	//page.nextId = pageid(binary.LittleEndian.Uint64(page.header[16:24]))
	page.recordCount = binary.LittleEndian.Uint16(page.header[RECORD_COUNT_START:RECORD_COUNT_END])
	page.freePointer = binary.LittleEndian.Uint16(page.header[FREE_POINTER_START:FREE_POINTER_END])

	page.recordTable = page.bytes[PAGE_HEADER_LEN : PAGE_HEADER_LEN +page.recordCount*4]

	return nil
}

// GetRecordCount returns the number of records held in page.
func (page *DataPage) GetRecordCount() uint16 {
	return page.recordCount
}

func (page *DataPage) setRecordTable(recordNumber uint16, offset uint16, recLen uint16) error {
	tableOffset := recordNumber * RECORD_TABLE_ENTRY_LEN
	// resize recordTable
	len := len(page.recordTable)
	page.recordTable = page.recordTable[0 : len+RECORD_TABLE_ENTRY_LEN] // add two more uint16 == 4 bytes
	binary.LittleEndian.PutUint16(page.recordTable[tableOffset:tableOffset+2], offset)
	binary.LittleEndian.PutUint16(page.recordTable[tableOffset+2:tableOffset+4], recLen)
	return nil
}

// GetFreeSpace return the amount of free space available to store a record (inclusive of any header fields.)
func (page *DataPage) GetFreeSpace() uint16 {
	return uint16(page.freePointer) - PAGE_HEADER_LEN - uint16(page.recordCount*RECORD_TABLE_ENTRY_LEN) - RECORD_TABLE_ENTRY_LEN // free pointer -  bytes header fields - #records * 4 bytes per table entry - another table entry
}

// AddRecord adds record to page, using copy semantics.
// Returns an error if insufficient page free space.
func (page *DataPage) AddRecord(record []byte) (uint16, error) {
	recLen := uint16(len(record))
	if uint16(recLen) > page.GetFreeSpace() {
		return 0, errors.New("Record length exceeds free space")
	}

	offset := page.freePointer - recLen
	copy(page.bytes[offset:page.freePointer], record)
	page.freePointer = offset
	recordNumber := page.recordCount // NB 0-based
	page.recordCount += 1
	page.setRecordTable(recordNumber, offset, recLen)
	return recordNumber, nil
}

// GetRecord returns record specified by recordNumber.
// Note: record numbers are 0 based.
func (page *DataPage) GetRecord(recordNumber uint16) ([]byte, error) {
	// recordNumber is 0 based
	if recordNumber+1 > page.recordCount {
		return nil, errors.New(fmt.Sprintf("Invalid record number: %d, record count: %d", recordNumber, page.recordCount))
	}
	tableOffset := recordNumber * RECORD_TABLE_ENTRY_LEN
	offset := binary.LittleEndian.Uint16(page.recordTable[tableOffset : tableOffset+2])
	len := binary.LittleEndian.Uint16(page.recordTable[tableOffset+2 : tableOffset+4])
	record := make([]byte, len, len)
	copy(record, page.bytes[offset:offset+len])
	return record, nil
}
