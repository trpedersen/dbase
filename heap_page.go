package dbase

import (
	"encoding/binary"
	"fmt"
	"sync"
)

const (
	SLOT_TABLE_ENTRY_LEN = int16(5) // bytes
	SLOT_COUNT_OFFSET    = int16(9)
	SLOT_TABLE_OFFSET    = PAGE_HEADER_LEN
	SLOT_TABLE_LEN       = PAGE_SIZE - SLOT_TABLE_OFFSET

	SLOT_UNALLOCATED   = 0x00
	RECORD_ON_PAGE     = 0x01
	RECORD_ON_OVERFLOW = 0x02
	RECORD_DELETED     = 0x04
	MAX_RECORD_LEN     = PAGE_SIZE - SLOT_TABLE_OFFSET - (2 * SLOT_TABLE_ENTRY_LEN)
)

type HeapPage interface {
	Page
	//PreviousPageID() PageID
	//NextPageID() PageID
	GetSlotCount() int16
	AddRecord(buf []byte) (int16, error)
	GetRecord(slot int16, buf []byte) (int, error)
	GetRecordLength(slot int16) (int, error)
	SetRecord(slot int16, buf []byte) error
	DeleteRecord(slot int16) error
	GetFreeSpace() int
	Clear() error
}

type heapPage struct {
	l *sync.Mutex
	page
	slotCount int16 // 32:36
	slotTable []byte
}

var bufferPool *sync.Pool = &sync.Pool{
	New: func() interface{} {
		return make([]byte, SLOT_TABLE_LEN, SLOT_TABLE_LEN)
	},
}

type RecordExceedsMaxSize struct {
	PageID PageID
	Slot   int16
	Len    int
}

func (e RecordExceedsMaxSize) Error() string {
	return fmt.Sprintf("Record exceeds maximum size, PageID: %d, Slot: %d, Len: %d", e.PageID, e.Slot, e.Len)
}

type RecordDeleted struct {
	PageID PageID
	Slot   int16
}

func (e RecordDeleted) Error() string {
	return fmt.Sprintf("Record deleted, PageID: %d, Slot: %d", e.PageID, e.Slot)
}

type InvalidRID struct {
	PageID PageID
	Slot   int16
}

func (e InvalidRID) Error() string {
	return fmt.Sprintf("Invalid RID, PageID: %d, Slot: %d", e.PageID, e.Slot)
}

type InsufficientPageSpace struct {
	PageID PageID
	Slot   int16
}

func (e InsufficientPageSpace) Error() string {
	return fmt.Sprintf("Insufficient free space on heap page, PageID: %d, Slot: %d", e.PageID, e.Slot)
}

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
		l: &sync.Mutex{},
	}

	page.header = page.bytes[0:PAGE_HEADER_LEN]
	page.slotTable = page.bytes[SLOT_TABLE_OFFSET : SLOT_TABLE_OFFSET+SLOT_TABLE_LEN]
	page.setSlotFlags(0, RECORD_ON_PAGE)
	page.setSlotOffset(0, 0)
	page.setSlotLength(0, SLOT_TABLE_LEN-(2*SLOT_TABLE_ENTRY_LEN))
	page.slotCount = 1

	return page
}

func (page *heapPage) getSlotFlags(slot int16) byte {
	if slot > page.slotCount-1 {
		panic("Invalid slot")
	}
	offset := SLOT_TABLE_LEN - ((slot + 1) * SLOT_TABLE_ENTRY_LEN)
	return page.slotTable[offset]
}

func (page *heapPage) setSlotFlags(slot int16, flags byte) error {
	//if slot > page.slotCount - 1 {
	//	panic("Invalid slot")
	//}
	offset := SLOT_TABLE_LEN - ((slot + 1) * SLOT_TABLE_ENTRY_LEN)
	page.slotTable[offset] = flags
	return nil
}

func (page *heapPage) getSlotOffset(slot int16) int16 {
	if slot > page.slotCount-1 {
		panic("Invalid slot")
	}
	offset := SLOT_TABLE_LEN - ((slot + 1) * SLOT_TABLE_ENTRY_LEN)
	return int16(binary.LittleEndian.Uint16(page.slotTable[offset+1 : offset+3]))
}

func (page *heapPage) setSlotOffset(slot int16, slotOffset int16) error {
	//if slot > page.slotCount - 1 {
	//	panic("Invalid slot")
	//}
	offset := SLOT_TABLE_LEN - ((slot + 1) * SLOT_TABLE_ENTRY_LEN)
	binary.LittleEndian.PutUint16(page.slotTable[offset+1:offset+3], uint16(slotOffset))
	return nil
}

func (page *heapPage) getSlotLength(slot int16) int16 {
	if slot > page.slotCount-1 {
		panic("Invalid slot")
	}
	offset := SLOT_TABLE_LEN - ((slot + 1) * SLOT_TABLE_ENTRY_LEN)
	result := int16(binary.LittleEndian.Uint16(page.slotTable[offset+3 : offset+5]))
	return result
}

func (page *heapPage) setSlotLength(slot int16, length int16) error {
	//if slot > page.slotCount - 1 {
	//	panic("Invalid slot")
	//}
	offset := SLOT_TABLE_LEN - ((slot + 1) * SLOT_TABLE_ENTRY_LEN)
	binary.LittleEndian.PutUint16(page.slotTable[offset+3:offset+5], uint16(length))
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// The page is encoded as a []byte PAGE_SIZE long, ready for serialisation.
func (page *heapPage) MarshalBinary() ([]byte, error) {

	page.l.Lock()
	defer page.l.Unlock()

	binary.LittleEndian.PutUint64(page.header[PAGE_ID_OFFSET:], uint64(page.id))
	page.header[PAGE_TYPE_OFFSET] = byte(PAGE_TYPE_HEAP)

	binary.LittleEndian.PutUint16(page.header[SLOT_COUNT_OFFSET:], uint16(page.slotCount))

	return page.bytes, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// PAGE_SIZE bytes are used to rehydrate the page.
func (page *heapPage) UnmarshalBinary(buf []byte) error {

	page.l.Lock()
	defer page.l.Unlock()

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
	page.slotTable = page.bytes[SLOT_TABLE_OFFSET : SLOT_TABLE_OFFSET+SLOT_TABLE_LEN]

	page.id = PageID(binary.LittleEndian.Uint64(page.header[PAGE_ID_OFFSET:]))
	page.pagetype = PAGE_TYPE_HEAP

	page.slotCount = int16(binary.LittleEndian.Uint16(page.header[SLOT_COUNT_OFFSET:]))

	return nil
}

// GetSlotCount returns the number of record slots held in page.
func (page *heapPage) GetSlotCount() int16 {
	return page.slotCount // int16(len(page.slots))
}

// GetFreeSpace return the amount of free space available to store a record (inclusive of any header fields.)
func (page *heapPage) GetFreeSpace() int {
	return int(page.getSlotLength(0))
}

// AddRecord adds record to page, using copy semantics. Returns record number for added record.
// Returns an error if insufficient page free space.
func (page *heapPage) AddRecord(record []byte) (int16, error) {

	page.l.Lock()
	defer page.l.Unlock()

	if len(record) > int(page.getSlotLength(0)) {
		return 0, InsufficientPageSpace{page.id, 0}
	}

	recordLength := int16(len(record))
	recordOffset := page.getSlotOffset(0)
	// make a new slot table entry
	page.setSlotFlags(page.slotCount, RECORD_ON_PAGE)
	page.setSlotOffset(page.slotCount, recordOffset)
	page.setSlotLength(page.slotCount, int16(recordLength))
	page.slotCount += 1
	copy(page.slotTable[recordOffset:recordOffset+recordLength], record)
	page.setSlotOffset(0, page.getSlotOffset(0)+int16(recordLength))
	page.setSlotLength(0, SLOT_TABLE_LEN-page.getSlotOffset(0)-(int16(page.slotCount+1)*SLOT_TABLE_ENTRY_LEN))
	if page.getSlotLength(0) < 0 {
		page.setSlotLength(0, 0)
	}
	return page.slotCount - 1, nil // slots are 0-based
}

// GetRecordLength returns length of record specified by recordNumber.
// Note: slot numbers are 0 based. Slot 0 is freespace slot.
func (page *heapPage) GetRecordLength(slotNumber int16) (int, error) {

	// slots are 0 based
	if slotNumber > page.slotCount-1 {
		return 0, InvalidRID{page.id, slotNumber}
	}

	if page.getSlotFlags(slotNumber) == RECORD_DELETED {
		return 0, RecordDeleted{page.id, slotNumber}
	}
	length := int(page.getSlotLength(slotNumber))
	return length, nil
}

// GetRecord returns record specified by recordNumber.
// Note: slot numbers are 0 based. Slot 0 is freespace slot.
func (page *heapPage) GetRecord(slotNumber int16, buf []byte) (int, error) {

	// slots are 0 based
	if slotNumber == 0 || slotNumber > page.slotCount-1 {
		return 0, InvalidRID{page.id, slotNumber}
	}
	if page.getSlotFlags(slotNumber) == RECORD_DELETED {
		return 0, RecordDeleted{page.id, slotNumber}
	}
	offset := int(page.getSlotOffset(slotNumber))
	length := int(page.getSlotLength(slotNumber))
	copy(buf, page.slotTable[offset:offset+length])
	return length, nil
}

// SetRecord updates record specified by slot.
// Note: record numbers are 0 based.
func (page *heapPage) SetRecord(slotNumber int16, buf []byte) error {

	page.l.Lock()
	defer page.l.Unlock()

	// recordNumber is 0 based
	if slotNumber == 0 || slotNumber > page.slotCount-1 {
		return InvalidRID{page.id, slotNumber}
	}
	slotOffset := page.getSlotOffset(slotNumber)
	slotLength := page.getSlotLength(slotNumber)
	freeLength := page.getSlotLength(0)
	recordLength := len(buf)

	switch {

	case recordLength == int(slotLength): // record same length
		// just update the slot
		copy(page.slotTable[slotOffset:slotOffset+slotLength], buf)
	case recordLength < int(slotLength): // record smaller than original length
		slotLength = int16(recordLength)
		copy(page.slotTable[slotOffset:slotOffset+slotLength], buf)
		page.compact()
	case recordLength < int(slotLength+freeLength): // still space enough in the page
		if err := page.reallocateSlot(slotNumber, int16(recordLength)); err != nil {
			panic(err) // TODO: replace panic once fully debugged
		}
		copy(page.slotTable[slotOffset:slotOffset+slotLength], buf)
	case recordLength < int(MAX_RECORD_LEN):
		return InsufficientPageSpace{PageID: page.id, Slot: slotNumber}
	default:
		// the record is too big to fit on a page, so move it on to an overflow page
		// but for now throw an error
		// TODO: implement overflow pages
		return RecordExceedsMaxSize{page.id, slotNumber, recordLength}
	}

	return nil
}

func (page *heapPage) DeleteRecord(slotNumber int16) error {
	page.l.Lock()
	defer page.l.Unlock()

	// recordNumber is 0 based
	if slotNumber == 0 || slotNumber > page.slotCount-1 {
		return InvalidRID{page.id, slotNumber}
	}
	//slot := page.slots[slotNumber]
	if page.getSlotFlags(slotNumber) == RECORD_DELETED {
		return nil // delete is idempotent
	}
	page.setSlotFlags(slotNumber, RECORD_DELETED)
	return page.compact() // TODO: compact later?
}

// TODO: implement
func (page *heapPage) reallocateSlot(slot int16, requestedLength int16) error {
	return nil
}

func (page *heapPage) compact() error {

	if page.slotCount == 1 {
		// reset free space
		page.setSlotFlags(0, RECORD_ON_PAGE)
		page.setSlotOffset(0, 0)
		page.setSlotLength(0, SLOT_TABLE_LEN-(2*SLOT_TABLE_ENTRY_LEN))
		return nil
	}

	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)

	for i := 0; i < len(buf); i++ {
		buf[i] = 0
	}

	var offset int16

	for i := int16(1); i < page.slotCount; i++ {
		slotOffset := page.getSlotOffset(i)
		slotLength := page.getSlotLength(i)

		switch page.getSlotFlags(i) {
		case RECORD_DELETED:
			page.setSlotOffset(i, -1)
			page.setSlotLength(i, -1)
		case RECORD_ON_PAGE:
			copy(buf[offset:offset+slotLength], page.bytes[slotOffset:slotOffset+slotLength])
			page.setSlotOffset(i, offset)
			offset += slotLength
		}
	}
	copy(buf[SLOT_TABLE_LEN-((page.slotCount+1)*SLOT_TABLE_ENTRY_LEN):SLOT_TABLE_LEN], page.slotTable[SLOT_TABLE_LEN-((page.slotCount+1)*SLOT_TABLE_ENTRY_LEN):SLOT_TABLE_LEN])
	copy(page.slotTable, buf)
	page.setSlotOffset(0, offset)
	page.setSlotLength(0, SLOT_TABLE_LEN-offset-((page.slotCount+1)*SLOT_TABLE_ENTRY_LEN))
	return nil
}

// Clear resets the page, removing any records. Record count is set to 0
func (page *heapPage) Clear() error {

	page.l.Lock()
	defer page.l.Unlock()

	page.setSlotFlags(0, RECORD_ON_PAGE)
	page.setSlotOffset(0, 0)
	page.setSlotLength(0, SLOT_TABLE_LEN-(2*SLOT_TABLE_ENTRY_LEN))
	page.slotCount = 1

	return nil
}
