package dbase

import (
	"fmt"
	"sync"
	"errors"
)
//
type Heap interface {
	Count() int64 // Count of records in heap
	Write(buf []byte) (RID, error)
	Get(rid RID, buf []byte) error
	Set(rid RID, buf []byte) error
	Delete(rid RID) error
}

type heap struct {
	store      PageStore
	headerPage HeapHeaderPage
	lastPage   HeapPage
	pagePool   *sync.Pool
}

func NewHeap(store PageStore) Heap {
	heap := &heap{
		store: store,
		//dir:        dir,
		headerPage: NewHeapHeaderPage(),
		lastPage:   NewHeapPage(),
		pagePool: &sync.Pool{
			New: func() interface{} {
				return NewHeapPage()
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
		var id PageID
		id, err = store.Append(heap.lastPage)
		if err != nil {
			panic(fmt.Sprintf("NewHeap init, err: %s", err))
		}
		heap.headerPage.SetLastPageID(id)
		if err := store.Set(0, heap.headerPage); err != nil {
			panic(fmt.Sprintf("NewHeap, err: %s", err))
		}
	}
	// get the header page
	if err := store.Get(0, heap.headerPage); err != nil {
		panic(fmt.Sprintf("NewHeap, err: %s", err))
	}
	// get the last page
	if err := heap.store.Get(heap.headerPage.GetLastPageID(), heap.lastPage); err != nil {
		panic(fmt.Sprintf("NewHeap, err: %s", err))
	}

	return heap
}

// Count returns the number of records in the Heap.
func (heap *heap) Count() int64 {
	return heap.headerPage.GetRecordCount()
}

func (heap *heap) Write(buf []byte) (RID, error) {

	var err error
	var rid RID
	var slot int16

	if len(buf) == 0 {
		return rid, errors.New("Zero length record")
	}

	len := len(buf)

	free := heap.lastPage.GetFreeSpace()
	if len > int(free) {
		// insufficient space, so make a new page
		heap.lastPage.Clear()
		var id PageID
		if id, err = heap.store.Append(heap.lastPage); err != nil {
			return rid, err
		}
		heap.headerPage.SetLastPageID(id)
		// update header page
		if err = heap.store.Set(0, heap.headerPage); err != nil {
			return rid, err
		}
	}
	if slot, err = heap.lastPage.AddRecord(buf); err != nil {
		return rid, err
	}
	if err := heap.store.Set(heap.headerPage.GetLastPageID(), heap.lastPage); err != nil {
		return rid, err
	}
	heap.headerPage.SetRecordCount( heap.headerPage.GetRecordCount() + 1)

	if err := heap.store.Set(0, heap.headerPage); err != nil {
		return rid, err
	}
	rid.PageID = heap.headerPage.GetLastPageID()
	rid.Slot = slot

	return rid, nil
}

func (heap *heap) Get(rid RID, buf []byte) error {

	page := heap.pagePool.Get().(HeapPage)
	defer heap.pagePool.Put(page)
	page.Clear()
	if err := heap.store.Get(rid.PageID, page); err != nil {
		return err
	}
	err := page.GetRecord(rid.Slot, buf)
	return err
}

func (heap *heap) Set(rid RID, buf []byte) error {

	page := heap.pagePool.Get().(HeapPage)
	defer heap.pagePool.Put(page)
	page.Clear()
	if err := heap.store.Get(rid.PageID, page); err != nil {
		return err
	}
	err := page.SetRecord(rid.Slot, buf)
	if err != nil {
		return err
	}
	err = heap.store.Set(rid.PageID, page)
	return err
}

func (heap *heap) Delete(rid RID) error {
	page := heap.pagePool.Get().(HeapPage)
	defer heap.pagePool.Put(page)
	page.Clear()
	if err := heap.store.Get(rid.PageID, page); err != nil {
		return err
	}
	err := page.DeleteRecord(rid.Slot)
	if err != nil {
		return err
	}
	err = heap.store.Set(rid.PageID, page)
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
