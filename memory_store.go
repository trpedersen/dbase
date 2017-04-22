package dbase

import (
	"errors"
	"fmt"
	"sync"
)

// MemoryStore is an in-memory implementation of a page store - useful for testing.
type MemoryStore interface {
	PageStore
}

type memoryStore struct {
	bufferPool *sync.Pool
	lastPageID PageID
	count      int64
	l          sync.Mutex
	pages      [][]byte
	gets       int
	sets       int
	news       int
	appends    int
	wipes      int
}

// NewMemoryStore returns a new MemoryStore. Under the covers
// it just uses sync.Pool to manage a buffer of page-size bytes
func NewMemoryStore() (MemoryStore, error) {
	store := &memoryStore{
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, PAGE_SIZE, PAGE_SIZE)
			},
		},
	}
	store.count = 0
	store.lastPageID = -1
	return store, nil
}

// Read returns the page with ID=id. Caller's responsibility to create page.
func (store *memoryStore) Read(id PageID, page Page) error {
	if id > store.lastPageID {
		return errors.New("Invalid page ID")
	}
	buf := store.pages[int(id)]
	store.gets++
	return page.UnmarshalBinary(buf)
}

// Write updates the page with id=ID.
func (store *memoryStore) Write(id PageID, page Page) error {

	store.l.Lock()
	defer store.l.Unlock()

	if id > store.lastPageID {
		return errors.New("Invalid page ID")
	} else if buf, err := page.MarshalBinary(); err != nil {
		return err
	} else {
		copy(store.pages[int(id)], buf)
	}
	store.sets++
	return nil
}

// New creates an empty page at the end of the memory store.
// Returns the page ID of the new page. Page count & Last page ID will be increased by 1.
func (store *memoryStore) New() (PageID, error) {

	store.l.Lock()
	defer store.l.Unlock()

	buf := make([]byte, PAGE_SIZE)
	store.pages = append(store.pages, buf)
	store.lastPageID++
	store.count++
	store.news++
	return PageID(store.lastPageID), nil
}

// Append appends the given page at the end of the memory store.
// Returns the page ID of the new page. Page count & Last page ID will be increased by 1.
func (store *memoryStore) Append(page Page) (PageID, error) {

	store.l.Lock()
	defer store.l.Unlock()

	buf, err := page.MarshalBinary()
	if err != nil {
		return 0, err
	}
	buf2 := make([]byte, PAGE_SIZE)
	copy(buf2, buf)
	store.pages = append(store.pages, buf2)
	store.lastPageID++
	store.count++
	store.appends++
	return PageID(store.lastPageID), nil
}

// Wipe zeros out the specified page. It does not reduce page count. Use with care!
func (store *memoryStore) Wipe(id PageID) error {

	store.l.Lock()
	defer store.l.Unlock()

	if id > store.lastPageID {
		return errors.New("Invalid page ID")
	}
	buf := store.pages[int(id)]
	for i := range buf {
		buf[i] = 0
	}
	store.wipes++
	return nil
}

// Count returns the total number of pages in the store.
func (store *memoryStore) Count() int64 {
	return int64(len(store.pages))
}

// Close closes the store.
func (store *memoryStore) Close() error {
	return nil
}

//Statistics returns a string containing stats for gets, sets, news, appends.
func (store *memoryStore) Statistics() string {
	return fmt.Sprintf("memory store: gets: %d, sets: %d, news: %d, appends: %d", store.gets, store.sets, store.news, store.appends)
}
