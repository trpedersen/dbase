package dbase

import (
	"io"
	"sync"
)

// HeapScanner is an iterable
type HeapScanner interface {
	Next(buf []byte) (RID, int, error)
}

type heapScanner struct {
	l      *sync.Mutex
	heap   Heap
	rid    RID
	page   HeapPage
	pageID PageID
	slotID int16
	state  int
}

// Scanner States
const (
	_ReadingPage int = 1 + iota
	_ReadingRecord
	_AtEOF
	_AtBOF
)

// Scanner Events
const (
	_RecordRead int = 1 + iota
	_DeletedRecordRead
	_EndOfPageReached
	_PageRead
	_EOF
)

// NewHeapScanner returns a new heap scanner
func NewHeapScanner(heap Heap) HeapScanner {
	scanner := &heapScanner{
		slotID: 0,
		pageID: 0,
		state:  _AtBOF,
		heap:   heap,
		page:   NewHeapPage(),
		l:      &sync.Mutex{},
	}
	return scanner
}

func (scanner *heapScanner) Next(buf []byte) (RID, int, error) {

	scanner.l.Lock()
	defer scanner.l.Unlock()

	var event int
	for {
		switch scanner.state {

		case _AtBOF:
			//log.Println("AT_BOF")
			scanner.state = _ReadingPage
		case _ReadingRecord:
			//log.Print("READING_RECORD")
			scanner.slotID++
			n, err := scanner.page.GetRecord(scanner.slotID, buf)
			if err != nil {
				if _, ok := err.(RecordDeleted); ok {
					event = _DeletedRecordRead
				} else {
					event = _EndOfPageReached
				}
			} else {
				event = _RecordRead
			}
			//log.Println(event)
			switch event {
			case _RecordRead:
				scanner.state = _ReadingRecord
				return RID{Slot: scanner.slotID, PageID: scanner.pageID}, n, nil
			case _DeletedRecordRead:
				scanner.state = _ReadingRecord
			case _EndOfPageReached:
				scanner.state = _ReadingPage
			}
		case _ReadingPage:
			//log.Print("READING_PAGE")
			scanner.pageID++
			err := scanner.heap.Store().Read(scanner.pageID, scanner.page)
			if err != nil {
				event = _EOF
			} else {
				event = _PageRead
			}
			//log.Println(event)
			switch event {
			case _EOF:
				scanner.state = _AtEOF
				return RID{}, 0, io.EOF
			case _PageRead:
				scanner.slotID = 0
				scanner.state = _ReadingRecord
			}
		case _AtEOF:
			//log.Println("AT_EOF")
			return RID{}, 0, io.EOF
		}
	}
}
