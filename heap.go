package dbase

import (
	"fmt"
	"sync"
	//"log"
)

type Heap struct {
	store      PageStore
	headerPage *HeapHeaderPage
	lastPage   *DataPage
	pagePool   *sync.Pool
}

func NewHeap(store PageStore) *Heap {
	heap := &Heap{
		store: store,
		//dir:        dir,
		headerPage: NewHeapHeaderPage(),
		lastPage:   NewDataPage(),
		pagePool: &sync.Pool{
			New: func() interface{} {
				return NewDataPage()
			},
		},
	}
	var err error

	if store.Count() == 0 {
		// new store, initialise with header
		_, err = store.Append(heap.headerPage)
		if err != nil {
			panic(fmt.Sprintf("NewHeap init, err: %s", err))
		}
		heap.headerPage.lastPageId, err = store.Append(heap.lastPage)
		if err != nil {
			panic(fmt.Sprintf("NewHeap init, err: %s", err))
		}
		if err := store.Set(0, heap.headerPage); err != nil {
			panic(fmt.Sprintf("NewHeap, err: %s", err))
		}
	}
	// get the header page
	if err := store.Get(0, heap.headerPage); err != nil {
		panic(fmt.Sprintf("NewHeap, err: %s", err))
	}
	// get the last page
	if err := heap.store.Get(heap.headerPage.lastPageId, heap.lastPage); err != nil {
		panic(fmt.Sprintf("NewHeap, err: %s", err))
	}

	return heap
}

// Count returns the number of records in the Heap.
func (heap *Heap) Count() int64 {
	return heap.headerPage.recordCount
}

func (heap *Heap) Write(buf []byte) (RID, error) {
	len := len(buf)
	if len > MAX_PAGE_PAYLOAD {
		buf = buf[0:MAX_PAGE_PAYLOAD]
		len = MAX_PAGE_PAYLOAD
	}

	var err error
	var rid RID
	var slot int16

	free := heap.lastPage.GetFreeSpace()
	if len > int(free) {
		// insufficient space, so make a new page
		heap.lastPage.Clear()
		if heap.headerPage.lastPageId, err = heap.store.Append(heap.lastPage); err != nil {
			return rid, err
		}
		// update header page
		if err = heap.store.Set(0, heap.headerPage); err != nil {
			return rid, err
		}
	}
	if slot, err = heap.lastPage.AddRecord(buf); err != nil {
		return rid, err
	}
	if err := heap.store.Set(heap.headerPage.lastPageId, heap.lastPage); err != nil {
		return rid, err
	}
	heap.headerPage.recordCount += 1
	if err := heap.store.Set(0, heap.headerPage); err != nil {
		return rid, err
	}
	rid.PageID = heap.headerPage.lastPageId
	rid.Slot = slot

	return rid, nil
}

func (heap *Heap) Get(rid RID, buf []byte) error {

	page := heap.pagePool.Get().(*DataPage)
	defer heap.pagePool.Put(page)
	page.Clear()
	if err := heap.store.Get(rid.PageID, page); err != nil {
		return err
	}
	err := page.GetRecord(rid.Slot, buf)
	return err
}

//func (heap *Heap) Write(buf []byte) error {
//	len := len(buf)
//	if len > MAX_PAGE_PAYLOAD {
//		buf = buf[0:MAX_PAGE_PAYLOAD]
//		len = MAX_PAGE_PAYLOAD
//	}
//
//	// find a page with sufficient space
//	percentFullReqd := int(100 - (100 * len / MAX_PAGE_PAYLOAD))
//	var flag byte
//	switch {
//	case percentFullReqd == 0:
//		flag = DIR_PAGE_EMPTY
//	case 1 < percentFullReqd <= 50:
//		flag = DIR_PAGE_1_50
//	case 50 < percentFullReqd <= 80:
//		flag = DIR_PAGE_51_80
//	case 80 < percentFullReqd <= 95:
//		flag = DIR_PAGE_81_95
//	case 95 < percentFullReqd < 100:
//		flag = DIR_PAGE_96_100
//	default:
//		flag = DIR_PAGE_EMPTY
//	}
//	page := heap.dir.
//}
