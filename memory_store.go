package dbase

import (
	"errors"
	"fmt"
	"sync"
)

type MemoryStore interface {
	PageStore
}

type memoryStore struct {
	bufferPool *sync.Pool
	lastPageId PageID
	count      int64
	l          sync.Mutex
	pages      [][]byte
	gets       int
	sets       int
	news       int
	appends    int
	wipes      int
}

func NewMemoryStore() (MemoryStore, error) {
	store := &memoryStore{
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, PAGE_SIZE, PAGE_SIZE)
			},
		},
	}
	store.count = 0
	store.lastPageId = -1
	return store, nil
}

// Get returns the page with ID=id. Caller's responsibility to create page.
func (store *memoryStore) Get(id PageID, page Page) error {
	if id > store.lastPageId {
		return errors.New("Invalid page ID")
	}
	//log.Println("get", id)
	buf := store.pages[int(id)]
	store.gets += 1
	return page.UnmarshalBinary(buf)
}

// Set updates the page with id=ID.
func (store *memoryStore) Set(id PageID, page Page) error {

	store.l.Lock()
	defer store.l.Unlock()

	//log.Println("set", id)

	// NB: file.WriteAt will just write at the end of the file if offset is past the end of the file
	// so check total pages to stop this happening
	if id > store.lastPageId {
		return errors.New("Invalid page ID")
	} else if buf, err := page.MarshalBinary(); err != nil {
		return err
	} else {
		copy(store.pages[int(id)], buf)
	}
	store.sets += 1
	return nil //store.file.Sync()
}

// New creates an empty page at the end of the database file.
// Returns the page ID of the new page. Page count & Last page ID will be increased by 1.
func (store *memoryStore) New() (PageID, error) {

	//log.Println("new")

	store.l.Lock()
	defer store.l.Unlock()

	buf := make([]byte, PAGE_SIZE)
	store.pages = append(store.pages, buf)
	store.lastPageId += 1
	store.count += 1
	store.news += 1
	return PageID(store.lastPageId), nil
}

// Append appends the given page at the end of the database file.
// Returns the page ID of the new page. Page count & Last page ID will be increased by 1.
func (store *memoryStore) Append(page Page) (PageID, error) {

	//log.Println("append", page)

	store.l.Lock()
	defer store.l.Unlock()

	buf, err := page.MarshalBinary()
	if err != nil {
		return 0, err
	}
	buf2 := make([]byte, PAGE_SIZE)
	copy(buf2, buf)
	store.pages = append(store.pages, buf2)
	store.lastPageId += 1
	store.count += 1
	store.appends += 1
	return PageID(store.lastPageId), nil
}

// Wipe zeros out the specified page. It does not reduce page count. Use with care!
func (store *memoryStore) Wipe(id PageID) error {

	store.l.Lock()
	defer store.l.Unlock()

	if id > store.lastPageId {
		return errors.New("Invalid page ID")
	}
	buf := store.pages[int(id)]
	for i, _ := range buf {
		buf[i] = 0
	}
	store.wipes += 1
	return nil //store.file.Sync()
}

// Count returns the total number of pages in the store.
func (store *memoryStore) Count() int64 {
	return int64(len(store.pages))
}

func (store *memoryStore) Close() error {
	return nil
}

func (store *memoryStore) Statistics() string {
	return fmt.Sprintf("memory store: gets: %d, sets: %d, news: %d, appends: %d", store.gets, store.sets, store.news, store.appends)
}
