package dbase

import (
	//"errors"
	"io"
	//"log"
	"sync"
)

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

const (
	READING_PAGE int = 1 + iota
	READING_RECORD
	AT_EOF
	AT_BOF
)

const (
	RECORD_READ int = 1 + iota
	DELETED_RECORD_READ
	END_OF_PAGE_REACHED
	PAGE_READ
	EOF
)

func NewHeapScanner(heap Heap) HeapScanner {
	scanner := &heapScanner{
		slotID: 0,
		pageID: 0,
		state:  AT_BOF,
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

		case AT_BOF:
			//log.Println("AT_BOF")
			scanner.state = READING_PAGE
		case READING_RECORD:
			//log.Print("READING_RECORD")
			scanner.slotID += 1
			n, err := scanner.page.GetRecord(scanner.slotID, buf)
			if err != nil {
				if _, ok := err.(RecordDeleted); ok {
					event = DELETED_RECORD_READ
				} else {
					event = END_OF_PAGE_REACHED
				}
			} else {
				event = RECORD_READ
			}
			//log.Println(event)
			switch event {
			case RECORD_READ:
				scanner.state = READING_RECORD
				return RID{Slot: scanner.slotID, PageID: scanner.pageID}, n, nil
			case DELETED_RECORD_READ:
				scanner.state = READING_RECORD
			case END_OF_PAGE_REACHED:
				scanner.state = READING_PAGE
			}
		case READING_PAGE:
			//log.Print("READING_PAGE")
			scanner.pageID += 1
			err := scanner.heap.Store().Get(scanner.pageID, scanner.page)
			if err != nil {
				event = EOF
			} else {
				event = PAGE_READ
			}
			//log.Println(event)
			switch event {
			case EOF:
				scanner.state = AT_EOF
				return RID{}, 0, io.EOF
			case PAGE_READ:
				scanner.slotID = 0
				scanner.state = READING_RECORD
			}
		case AT_EOF:
			//log.Println("AT_EOF")
			return RID{}, 0, io.EOF
		}
	}
}

/*



	with(scanner, func() {

		for ;;
		record, error := scanner.Next()
		print record
		return
	})

	func with(o Quitable, f func()) {
		f()
		o.Quit
	}

*/
