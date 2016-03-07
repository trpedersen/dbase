package dbase

import (
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
)

const (
	SLOT_TABLE_ENTRY_LEN = int16(5) // bytes
	SLOT_COUNT_OFFSET    = int16(9)
	SLOT_TABLE_OFFSET    = PAGE_HEADER_LEN
	SLOT_TABLE_LEN       = PAGE_SIZE - SLOT_TABLE_OFFSET
	RECORD_COUNT_END     = 17 // int64
	FREE_POINTER_OFFSET  = 17
	FREE_POINTER_END     = 19 // int16

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
	GetRecord(slot int16, buf []byte) error
	SetRecord(slot int16, buf []byte) error
	DeleteRecord(slot int16) error
	GetFreeSpace() int
	Clear() error
}

type heapPage struct {
	page
	slotCount int16 // 32:36
	//freePointer int16 // 36:40
	//slotTable   []byte
	//slotTable *SlotTable
	slotTable []byte
	slots     []*Slot
	l         *sync.Mutex
}

var bufferPool *sync.Pool = &sync.Pool{
	New: func() interface{} {
		return make([]byte, SLOT_TABLE_LEN, SLOT_TABLE_LEN)
	},
}

var slotPool *sync.Pool = &sync.Pool{
	New: func() interface{} {
		return NewSlot()
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

type Slot struct {
	id     int16
	flags  byte
	offset int16
	length int16
}

type SlotByID []*Slot

func (s SlotByID) Len() int           { return len(s) }
func (s SlotByID) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SlotByID) Less(i, j int) bool { return s[i].id < s[j].id }

type SlotByOffset []*Slot

func (s SlotByOffset) Len() int           { return len(s) }
func (s SlotByOffset) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SlotByOffset) Less(i, j int) bool { return s[i].offset < s[j].offset }

type SlotByLength []*Slot

func (s SlotByLength) Len() int           { return len(s) }
func (s SlotByLength) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SlotByLength) Less(i, j int) bool { return s[i].length < s[j].length }

func NewFreeSpaceSlot() *Slot {
	return &Slot{
		id:     0,
		flags:  RECORD_ON_PAGE,
		offset: 0,
		length: SLOT_TABLE_LEN - (2 * SLOT_TABLE_ENTRY_LEN),
	}
}

func NewSlot() *Slot {
	return &Slot{
		id:     -1,
		flags:  SLOT_UNALLOCATED,
		offset: -1,
		length: -1,
	}
}

type RID struct {
	PageID PageID
	Slot   int16
}

//type SlotTable struct {
//	page      *heapPage
//	freespace Slot
//	slots     []Slot
//}
//
//func NewSlotTable(page *heapPage) *SlotTable {
//	table := &SlotTable{
//		page:  page,
//		slots: make([]Slot, 1, 200),
//	}
//	table.slots[0] = NewFreeSpaceSlot()
//	return table
//}

// NewHeapPage returns a new Heap Page.
func NewHeapPage() HeapPage {

	page := &heapPage{
		page: page{
			id:       0,
			pagetype: PAGE_TYPE_HEAP,
			bytes:    make([]byte, PAGE_SIZE, PAGE_SIZE),
		},
		slots: make([]*Slot, 1, 200),
		l:     &sync.Mutex{},
	}

	page.header = page.bytes[0:PAGE_HEADER_LEN]
	page.slotTable = page.bytes[SLOT_TABLE_OFFSET : SLOT_TABLE_OFFSET+SLOT_TABLE_LEN]
	page.slots[0] = NewFreeSpaceSlot()
	page.slotCount = 1

	return page
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// The page is encoded as a []byte PAGE_SIZE long, ready for serialisation.
func (page *heapPage) MarshalBinary() ([]byte, error) {

	page.l.Lock()
	defer page.l.Unlock()

	binary.LittleEndian.PutUint64(page.header[PAGE_ID_OFFSET:], uint64(page.id))
	page.header[PAGE_TYPE_OFFSET] = byte(PAGE_TYPE_HEAP)

	binary.LittleEndian.PutUint16(page.header[SLOT_COUNT_OFFSET:], uint16(len(page.slots)))

	sort.Sort(SlotByID(page.slots))
	for i, slot := range page.slots {
		offset := SLOT_TABLE_LEN - ((int16(i) + 1) * SLOT_TABLE_ENTRY_LEN)
		page.slotTable[offset] = slot.flags
		offset += 1
		binary.LittleEndian.PutUint16(page.slotTable[offset:offset+2], uint16(slot.offset))
		offset += 2
		binary.LittleEndian.PutUint16(page.slotTable[offset:offset+2], uint16(slot.length))
	}
	//log.Println("ma", page.slotTable)

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
	//log.Println("unm", page.slotTable)

	page.id = PageID(binary.LittleEndian.Uint64(page.header[PAGE_ID_OFFSET:]))
	page.pagetype = PAGE_TYPE_HEAP

	page.slotCount = int16(binary.LittleEndian.Uint16(page.header[SLOT_COUNT_OFFSET:]))

	page.slots = page.slots[0:0]
	for i := int16(0); i < page.GetSlotCount(); i++ {
		offset := SLOT_TABLE_LEN - ((i + 1) * SLOT_TABLE_ENTRY_LEN)
		slot := &Slot{
			id:     i,
			flags:  page.slotTable[offset],
			offset: int16(binary.LittleEndian.Uint16(page.slotTable[offset+1 : offset+3])),
			length: int16(binary.LittleEndian.Uint16(page.slotTable[offset+3 : offset+5])),
		}
		page.slots = append(page.slots, slot) // page.slots now sorted by slot.id
	}

	return nil
}

// GetSlotCount returns the number of record slots held in page.
func (page *heapPage) GetSlotCount() int16 {
	return page.slotCount // int16(len(page.slots))
}

// GetFreeSpace return the amount of free space available to store a record (inclusive of any header fields.)
func (page *heapPage) GetFreeSpace() int {
	return int(page.slots[0].length)
}

// AddRecord adds record to page, using copy semantics. Returns record number for added record.
// Returns an error if insufficient page free space.
func (page *heapPage) AddRecord(record []byte) (int16, error) {

	page.l.Lock()
	defer page.l.Unlock()

	if len(record) > int(page.slots[0].length) {
		return 0, InsufficientPageSpace{page.id, 0}
	}

	recordLength := int16(len(record))
	recordOffset := page.slots[0].offset
	page.slots = append(page.slots, &Slot{
		id:     int16(len(page.slots)),
		flags:  RECORD_ON_PAGE,
		offset: recordOffset,
		length: int16(recordLength),
	})
	page.slotCount += 1
	copy(page.slotTable[recordOffset:recordOffset+recordLength], record)
	page.slots[0].offset += int16(recordLength)
	page.slots[0].length = SLOT_TABLE_LEN - page.slots[0].offset - (int16(len(page.slots)+1) * SLOT_TABLE_ENTRY_LEN)
	if page.slots[0].length < 0 {
		page.slots[0].length = 0
	}
	slot := page.slotCount - 1 // slots are 0-based
	return slot, nil
}

// GetRecord returns record specified by recordNumber.
// Note: slot numbers are 0 based. Slot 0 is freespace slot.
func (page *heapPage) GetRecord(slotNumber int16, buf []byte) error {

	page.l.Lock()
	defer page.l.Unlock()

	// slots are 0 based
	if slotNumber > page.slotCount-1 {
		return InvalidRID{page.id, slotNumber}
	}
	slot := page.slots[slotNumber]
	if slot.flags == RECORD_DELETED {
		return RecordDeleted{page.id, slotNumber}
	}
	copy(buf, page.slotTable[slot.offset:slot.offset+slot.length])
	return nil
}

// SetRecord updates record specified by slot.
// Note: record numbers are 0 based.
func (page *heapPage) SetRecord(slotNumber int16, buf []byte) error {

	page.l.Lock()
	defer page.l.Unlock()

	// recordNumber is 0 based
	if slotNumber > page.slotCount-1 {
		return InvalidRID{page.id, slotNumber}
	}
	slot := page.slots[slotNumber]
	recordLength := len(buf)

	switch {

	case recordLength == int(slot.length): // record same length
		// just update the slot
		copy(page.slotTable[slot.offset:slot.offset+slot.length], buf)
	case recordLength < int(slot.length): // record smaller than original length
		slot.length = int16(recordLength)
		copy(page.slotTable[slot.offset:slot.offset+slot.length], buf)
		page.compact()
	case recordLength < int(slot.length+page.slots[0].length): // still space enough in the page
		if err := page.reallocateSlot(slot, int16(recordLength)); err != nil {
			panic(err) // TODO: replace panic once fully debugged
		}
		copy(page.slotTable[slot.offset:slot.offset+slot.length], buf)
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
	if slotNumber > page.slotCount-1 {
		return InvalidRID{page.id, slotNumber}
	}
	slot := page.slots[slotNumber]
	if slot.flags == RECORD_DELETED {
		return nil // delete is idempotent
	}
	slot.flags = RECORD_DELETED

	return page.compact() // TODO: compact later?
}

// TODO: implement
func (page *heapPage) reallocateSlot(slot *Slot, requestedLength int16) error {
	return nil
}

// TODO: implement
func (page *heapPage) compact() error {
	//log.Println("compact A", page.slots[0].offset, page.slots[0].length)

	var slot *Slot

	if len(page.slots) == 1 {
		// reset free space
		page.slots[0] = NewFreeSpaceSlot()
		return nil
	}

	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)

	for i := 0; i < len(buf); i++ {
		buf[i] = 0
	}

	var offset int16

	for i := 1; i < len(page.slots); i++ {
		slot = page.slots[i]

		switch slot.flags {
		case RECORD_DELETED:
			slot.offset = -1
			slot.length = -1
			continue
		case RECORD_ON_PAGE:
			copy(buf[offset:offset+slot.length], page.bytes[slot.offset:slot.offset+slot.length])
			slot.offset = offset
			offset += slot.length
		}
	}
	page.slots[0].offset = offset
	page.slots[0].length = SLOT_TABLE_LEN - offset - (int16(len(page.slots)+1) * SLOT_TABLE_ENTRY_LEN)
	copy(page.slotTable, buf)

	//log.Println("compact B", page.slots[0].offset, page.slots[0].length)
	return nil
}

// Clear resets the page, removing any records. Record count is set to 0
func (page *heapPage) Clear() error {

	page.l.Lock()
	defer page.l.Unlock()

	//page.slotCount = 0
	page.slots = page.slots[0:0]
	page.slots = append(page.slots, NewFreeSpaceSlot())
	page.slotCount = 1
	//log.Println("clear", len(page.slots), page.slots)
	return nil
}
