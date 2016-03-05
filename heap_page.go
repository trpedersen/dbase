package dbase

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/lukehoban/ident"
	"sort"
)

const (
	SLOT_TABLE_ENTRY_LEN = 4 // bytes
	SLOT_COUNT_OFFSET    = 9
	SLOT_TABLE_OFFSET    = PAGE_HEADER_LEN
	RECORD_COUNT_END     = 17 // int64
	FREE_POINTER_OFFSET  = 17
	FREE_POINTER_END     = 19 // int16
)

type HeapPage interface {
	Page
	//PreviousPageID() PageID
	//NextPageID() PageID
	GetSlotCount() int16
	AddRecord(buf []byte) (int16, error)
	GetRecord(slot int16, buf []byte) error
	SetRecord(slot int16, buf []byte) error
	GetFreeSpace() int
	Clear() error
}

type heapPage struct {
	page
	slotCount   int16 // 32:36
	freePointer int16 // 36:40
	slotTable   []byte
	slots       []Slot
}

type Slot struct {
	id     int16
	offset int16
	length int16
}

type SlotByID []Slot

func (s SlotByID) Len() int      { return len(s) }
func (s SlotByID) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SlotByID) Less(i, j int) { return s[i].id < s[j].id }

type SlotByOffset []Slot

func (s SlotByOffset) Len() int      { return len(s) }
func (s SlotByOffset) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SlotByOffset) Less(i, j int) { return s[i].offset < s[j].offset }

type SlotByLength []Slot

func (s SlotByLength) Len() int      { return len(s) }
func (s SlotByLength) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SlotByLength) Less(i, j int) { return s[i].length < s[j].length }

type RID struct {
	PageID PageID
	Slot   int16
}

// NewHeapPage returns a new Heap Page.
func NewHeapPage() HeapPage {

	page := &heapPage{
		page: page{
			id:       0,
			pagetype: PAGE_TYPE_HEAP,
			bytes:    make([]byte, PAGE_SIZE, PAGE_SIZE),
		},
		slots: make([]Slot, 0, 200),
	}

	page.header = page.bytes[0:PAGE_HEADER_LEN]
	page.slotTable = page.bytes[SLOT_TABLE_OFFSET:SLOT_TABLE_OFFSET]
	// Slot 0 is the free space slot
	page.slots = append(page.slots, Slot{
		id: 0,
		offset:PAGE_SIZE,
		len: PAGE_SIZE - SLOT_TABLE_OFFSET - SLOT_TABLE_ENTRY_LEN,
	})
	//log.Println("len(page.slotTable)", len(page.slotTable))
	page.freePointer = PAGE_SIZE

	return page
}

//func (page *heapPage) GetID() PageID {
//	return page.id
//}
//
//func (page *heapPage) GetType() PageType {
//	return page.pagetype
//}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// The page is encoded as a []byte PAGE_SIZE long, ready for serialisation.
func (page *heapPage) MarshalBinary() ([]byte, error) {

	binary.LittleEndian.PutUint64(page.header[PAGE_ID_OFFSET:], uint64(page.id))
	page.header[PAGE_TYPE_OFFSET] = byte(PAGE_TYPE_HEAP)

	binary.LittleEndian.PutUint16(page.header[SLOT_COUNT_OFFSET:], uint16(page.slotCount))
	binary.LittleEndian.PutUint16(page.header[FREE_POINTER_OFFSET:], uint16(page.freePointer))

	sort.Sort(SlotByID(page.slots))
	page.slotTable = page.bytes[SLOT_TABLE_OFFSET : SLOT_TABLE_OFFSET+len(page.slots)*SLOT_TABLE_ENTRY_LEN]
	for i, slot := range page.slots {
		offset := i * SLOT_TABLE_ENTRY_LEN
		binary.LittleEndian.PutUint16(page.slotTable[offset:], slot.offset)
		binary.LittleEndian.PutUint16(page.slotTable[offset+2:], slot.length)
	}

	return page.bytes, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// PAGE_SIZE bytes are used to rehydrate the page.
func (page *heapPage) UnmarshalBinary(buf []byte) error {

	if len(buf) != int(PAGE_SIZE) {
		panic("Invalid buffer")
	}
	// check page type, panic if wrong
	pageType := PageType(buf[PAGE_TYPE_OFFSET])
	if pageType != PAGE_TYPE_HEAP {
		panic("Invalid page type")
	}

	copy(page.bytes, buf)

	page.header = page.bytes[0:PAGE_HEADER_LEN]
	page.id = PageID(binary.LittleEndian.Uint64(page.header[PAGE_ID_OFFSET:]))
	page.pagetype = PAGE_TYPE_HEAP

	page.slotCount = int16(binary.LittleEndian.Uint16(page.header[SLOT_COUNT_OFFSET:]))
	page.freePointer = int16(binary.LittleEndian.Uint16(page.header[FREE_POINTER_OFFSET:]))

	page.slotTable = page.bytes[SLOT_TABLE_OFFSET:SLOT_TABLE_OFFSET + page.slotCount * SLOT_TABLE_ENTRY_LEN]
	for i := 0; i < page.slotCount; i++ {
		offset := i * SLOT_TABLE_ENTRY_LEN
		slot := Slot{
			id: i,
			offset: binary.LittleEndian.PutUint16(page.slotTable[offset:offset+2]),
			length: binary.LittleEndian.PutUint16(page.slotTable[offset+2:offset+4]),
		}
		page.slots = append(page.slots, slot) // page.slots now sorted by slot.id
	}

	return nil
}

// GetSlotCount returns the number of record slots held in page.
func (page *heapPage) GetSlotCount() int16 {
	return page.slotCount
}

func (page *heapPage) setSlot(slotNumber int16, recordOffset int16, recordLength int16) error {
	//log.Println("setSlot", slotNumber, recordOffset, recordLength)
	//slotOffset := slotNumber * SLOT_TABLE_ENTRY_LEN
	//// resize slotTable
	//slotTableLen := len(page.slotTable)
	////log.Println( len(page.bytes), PAGE_HEADER_LEN, slotTableLen, PAGE_HEADER_LEN + slotTableLen + SLOT_TABLE_ENTRY_LEN )
	//page.slotTable = page.bytes[SLOT_TABLE_OFFSET : SLOT_TABLE_OFFSET+slotTableLen+SLOT_TABLE_ENTRY_LEN] // add two more uint16 == 4 bytes
	//binary.LittleEndian.PutUint16(page.slotTable[slotOffset:], uint16(recordOffset))
	//slotOffset += 2
	//binary.LittleEndian.PutUint16(page.slotTable[slotOffset:], uint16(recordLength))
	////log.Println("setSlot end", len(page.slotTable))
	if len(page.slots) < slotNumber {
		return errors.New(fmt.Sprintf("Invalid slotNumber, slot count: %d", len(page.slots)))
	}
	return nil
}

// GetFreeSpace return the amount of free space available to store a record (inclusive of any header fields.)
func (page *heapPage) GetFreeSpace() int {
	//result := int(page.freePointer) - (PAGE_HEADER_LEN + int(page.slotCount*SLOT_TABLE_ENTRY_LEN) + SLOT_TABLE_ENTRY_LEN) - 1 // free pointer -  bytes header fields - #records * 4 bytes per table entry - another table entry
	//if result < 0 {
	//	result = 0
	//}
	return page.slots[0].length
	//return result
}

// AddRecord adds record to page, using copy semantics. Returns record number for added record.
// Returns an error if insufficient page free space.
func (page *heapPage) AddRecord(record []byte) (int16, error) {
	recordLength := len(record)
	if recordLength > page.GetFreeSpace() {
		return 0, errors.New("Record length exceeds free space")
	}

	recordOffset := page.slots[0].offset - int16(recordLength)
	copy(page.bytes[recordOffset:page.freePointer], record)
	page.freePointer = recordOffset
	slot := page.slotCount // NB 0-based
	page.slotCount += 1
	page.setSlot(slot, recordOffset, int16(recordLength))
	return slot, nil
}

// GetRecord returns record specified by recordNumber.
// Note: record numbers are 0 based.
func (page *heapPage) GetRecord(slot int16, buf []byte) error {
	// recordNumber is 0 based
	if slot+1 > page.slotCount {
		return errors.New(fmt.Sprintf("Invalid record number: %d, slot count: %d", slot, page.slotCount))
	}
	slotOffset := slot * SLOT_TABLE_ENTRY_LEN
	//log.Println("slotOffset", slotOffset)
	offset := binary.LittleEndian.Uint16(page.slotTable[slotOffset : slotOffset+2])
	len := binary.LittleEndian.Uint16(page.slotTable[slotOffset+2 : slotOffset+4])
	copy(buf, page.bytes[offset:offset+len])
	return nil
}

// SetRecord updates record specified by recordNumber.
// Note: record numbers are 0 based.
func (page *heapPage) SetRecord(recordNumber int16, buf []byte) error {
	// recordNumber is 0 based
	if recordNumber+1 > page.slotCount {
		return errors.New(fmt.Sprintf("Invalid record number: %d, record count: %d", recordNumber, page.slotCount))
	}
	slotOffset := recordNumber * SLOT_TABLE_ENTRY_LEN
	offset := binary.LittleEndian.Uint16(page.slotTable[slotOffset:])
	len := binary.LittleEndian.Uint16(page.slotTable[slotOffset+2:])
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
func (page *heapPage) Clear() error {
	page.slotTable = page.bytes[SLOT_TABLE_OFFSET:SLOT_TABLE_OFFSET]
	page.freePointer = PAGE_SIZE
	page.slotCount = 0
	return nil
}
