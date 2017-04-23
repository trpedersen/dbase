package dbase

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
)

const (
	slotTableEntryLen = int16(5) // bytes
	slotCountOffset   = int16(9)
	slotTableOffset   = pageHeaderLength
	slotTableLen      = PageSize - slotTableOffset

	slotUnallocated  = 0x00
	recordOnPage     = 0x01
	recordOnOverflow = 0x02
	recordDeleted    = 0x04
	maxRecordLen     = PageSize - slotTableOffset - (2 * slotTableEntryLen)
)

// HeapPage is a page that contains records stored in slots.
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

var bufferPool = &sync.Pool{
	New: func() interface{} {
		return make([]byte, slotTableLen, slotTableLen)
	},
}

// RecordExceedsMaxSize is an error type
type RecordExceedsMaxSize struct {
	PageID PageID
	Slot   int16
	Len    int
}

func (e RecordExceedsMaxSize) Error() string {
	return fmt.Sprintf("Record exceeds maximum size, PageID: %d, Slot: %d, Len: %d", e.PageID, e.Slot, e.Len)
}

// RecordDeleted is an error type
type RecordDeleted struct {
	PageID PageID
	Slot   int16
}

func (e RecordDeleted) Error() string {
	return fmt.Sprintf("Record deleted, PageID: %d, Slot: %d", e.PageID, e.Slot)
}

// InvalidRID is an error type - invalid PageID + slot
type InvalidRID struct {
	PageID PageID
	Slot   int16
}

func (e InvalidRID) Error() string {
	return fmt.Sprintf("Invalid RID, PageID: %d, Slot: %d", e.PageID, e.Slot)
}

// InsufficientPageSpace is an error type - not enough space in the page for this record
type InsufficientPageSpace struct {
	PageID PageID
	Slot   int16
}

func (e InsufficientPageSpace) Error() string {
	return fmt.Sprintf("Insufficient free space on heap page, PageID: %d, Slot: %d", e.PageID, e.Slot)
}

// RID is a record ID = PageID + Slot #
type RID struct {
	PageID PageID
	Slot   int16
}

// NewHeapPage returns a new Heap Page.
func NewHeapPage() HeapPage {

	page := &heapPage{
		page: page{
			id:       0,
			pagetype: pageTypeHeap,
			bytes:    make([]byte, PageSize, PageSize),
		},
		l: &sync.Mutex{},
	}

	page.header = page.bytes[0:pageHeaderLength]
	page.slotTable = page.bytes[slotTableOffset : slotTableOffset+slotTableLen]
	page.setSlotFlags(0, recordOnPage)
	page.setSlotOffset(0, 0)
	page.setSlotLength(0, slotTableLen-(2*slotTableEntryLen))
	page.slotCount = 1

	return page
}

func (page *heapPage) getSlotFlags(slot int16) byte {
	if slot > page.slotCount-1 {
		panic("Invalid slot")
	}
	offset := slotTableLen - ((slot + 1) * slotTableEntryLen)
	return page.slotTable[offset]
}

func (page *heapPage) setSlotFlags(slot int16, flags byte) error {
	//if slot > page.slotCount - 1 {
	//	panic("Invalid slot")
	//}
	offset := slotTableLen - ((slot + 1) * slotTableEntryLen)
	page.slotTable[offset] = flags
	return nil
}

func (page *heapPage) getSlotOffset(slot int16) int16 {
	if slot > page.slotCount-1 {
		panic("Invalid slot")
	}
	offset := slotTableLen - ((slot + 1) * slotTableEntryLen)
	return int16(binary.LittleEndian.Uint16(page.slotTable[offset+1 : offset+3]))
}

func (page *heapPage) setSlotOffset(slot int16, slotOffset int16) error {
	//if slot > page.slotCount - 1 {
	//	panic("Invalid slot")
	//}
	offset := slotTableLen - ((slot + 1) * slotTableEntryLen)
	binary.LittleEndian.PutUint16(page.slotTable[offset+1:offset+3], uint16(slotOffset))
	return nil
}

func (page *heapPage) getSlotLength(slot int16) int16 {
	if slot > page.slotCount-1 {
		panic("Invalid slot")
	}
	offset := slotTableLen - ((slot + 1) * slotTableEntryLen)
	result := int16(binary.LittleEndian.Uint16(page.slotTable[offset+3 : offset+5]))
	return result
}

func (page *heapPage) setSlotLength(slot int16, length int16) error {
	//if slot > page.slotCount - 1 {
	//	panic("Invalid slot")
	//}
	offset := slotTableLen - ((slot + 1) * slotTableEntryLen)
	binary.LittleEndian.PutUint16(page.slotTable[offset+3:offset+5], uint16(length))
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// The page is encoded as a []byte PAGE_SIZE long, ready for serialisation.
func (page *heapPage) MarshalBinary() ([]byte, error) {

	page.l.Lock()
	defer page.l.Unlock()

	binary.LittleEndian.PutUint64(page.header[pageIDOffset:], uint64(page.id))
	page.header[pageTypeOffset] = byte(pageTypeHeap)

	binary.LittleEndian.PutUint16(page.header[slotCountOffset:], uint16(page.slotCount))

	return page.bytes, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// PAGE_SIZE bytes are used to rehydrate the page.
func (page *heapPage) UnmarshalBinary(buf []byte) error {

	page.l.Lock()
	defer page.l.Unlock()

	if len(buf) != int(PageSize) {
		panic("Invalid buffer")
	}
	// check page type, panic if wrong
	pageType := PageType(buf[pageTypeOffset])
	if pageType != pageTypeHeap {
		panic("Invalid page type")
	}

	copy(page.bytes, buf)

	page.header = page.bytes[0:pageHeaderLength]
	page.slotTable = page.bytes[slotTableOffset : slotTableOffset+slotTableLen]

	page.id = PageID(binary.LittleEndian.Uint64(page.header[pageIDOffset:]))
	page.pagetype = pageTypeHeap

	page.slotCount = int16(binary.LittleEndian.Uint16(page.header[slotCountOffset:]))

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
	page.setSlotFlags(page.slotCount, recordOnPage)
	page.setSlotOffset(page.slotCount, recordOffset)
	page.setSlotLength(page.slotCount, int16(recordLength))
	page.slotCount++
	copy(page.slotTable[recordOffset:recordOffset+recordLength], record)
	page.setSlotOffset(0, page.getSlotOffset(0)+int16(recordLength))
	page.setSlotLength(0, slotTableLen-page.getSlotOffset(0)-(int16(page.slotCount+1)*slotTableEntryLen))
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

	if page.getSlotFlags(slotNumber) == recordDeleted {
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
	if page.getSlotFlags(slotNumber) == recordDeleted {
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
		if err := page.reallocateSlot(slotNumber, int16(recordLength)); err != nil {
			panic(err) // TODO: replace panic once fully debugged
		}
		copy(page.slotTable[slotOffset:int(slotOffset)+recordLength], buf)
		slotLength = int16(recordLength)
	case recordLength < int(slotLength+freeLength): // still space enough in the page
		if err := page.reallocateSlot(slotNumber, int16(recordLength)); err != nil {
			panic(err) // TODO: replace panic once fully debugged
		}
		copy(page.slotTable[slotOffset:int(slotOffset)+recordLength], buf)
	case recordLength < int(maxRecordLen):
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
	if page.getSlotFlags(slotNumber) == recordDeleted {
		return nil // delete is idempotent
	}
	page.setSlotFlags(slotNumber, recordDeleted)
	return page.compact() // TODO: compact later?
}

// TODO: implement
func (page *heapPage) reallocateSlot(slot int16, requestedLength int16) error {
	// take a copy of the record
	// temporarily delete then compact
	// allocate new record length from freespace

	flags := page.getSlotFlags(slot)
	if flags != recordOnPage {
		return errors.New("Invalid record flag")
	}

	buf := bufferPool.Get().([]byte)
	buf = buf[0:requestedLength]
	defer bufferPool.Put(buf)

	offset := page.getSlotOffset(slot)
	length := page.getSlotLength(slot)
	copy(buf, page.slotTable[offset:offset+length]) // NB: requested length could be shorter than original

	page.setSlotFlags(slot, recordDeleted)
	page.compact()
	page.setSlotFlags(slot, recordOnPage)

	offset = page.getSlotOffset(0)
	// make a new slot table entry
	page.setSlotOffset(slot, offset)
	page.setSlotLength(slot, requestedLength)
	copy(page.slotTable[offset:offset+requestedLength], buf)
	page.setSlotOffset(0, page.getSlotOffset(0)+requestedLength)
	page.setSlotLength(0, slotTableLen-page.getSlotOffset(0)-(int16(page.slotCount+1)*slotTableEntryLen))
	if page.getSlotLength(0) < 0 {
		page.setSlotLength(0, 0)
	}

	return nil
}

func (page *heapPage) compact() error {

	if page.slotCount == 1 {
		// reset free space
		page.setSlotFlags(0, recordOnPage)
		page.setSlotOffset(0, 0)
		page.setSlotLength(0, slotTableLen-(2*slotTableEntryLen))
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
		case recordDeleted:
			page.setSlotOffset(i, -1)
			page.setSlotLength(i, -1)
		case recordOnPage:
			copy(buf[offset:offset+slotLength], page.bytes[slotOffset:slotOffset+slotLength])
			page.setSlotOffset(i, offset)
			offset += slotLength
		}
	}
	copy(buf[slotTableLen-((page.slotCount+1)*slotTableEntryLen):slotTableLen], page.slotTable[slotTableLen-((page.slotCount+1)*slotTableEntryLen):slotTableLen])
	copy(page.slotTable, buf)
	page.setSlotOffset(0, offset)
	page.setSlotLength(0, slotTableLen-offset-((page.slotCount+1)*slotTableEntryLen))
	return nil
}

// Clear resets the page, removing any records. Record count is set to 0
func (page *heapPage) Clear() error {

	page.l.Lock()
	defer page.l.Unlock()

	page.setSlotFlags(0, recordOnPage)
	page.setSlotOffset(0, 0)
	page.setSlotLength(0, slotTableLen-(2*slotTableEntryLen))
	page.slotCount = 1

	return nil
}
