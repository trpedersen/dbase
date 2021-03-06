package dbase

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

// FileStore is a file-backed page store
type FileStore interface {
	PageStore
	Path() string
}

// FileStoreOptions is used to control a filestore
type FileStoreOptions struct {
	ReadOnly bool
}

// DefaultOptions - read/write
var DefaultOptions = &FileStoreOptions{
	ReadOnly: false,
}

// fileStore is the concrete implementation for FileStore - internal use only
type fileStore struct {
	readOnly   bool
	path       string
	file       *os.File
	bufferPool *sync.Pool
	lastPageID PageID
	count      int64
	l          sync.Mutex
	gets       int
	sets       int
	news       int
	appends    int
	wipes      int
}

// Open opens a FileStore. Created fresh if necessary.
func Open(path string, mode os.FileMode, options *FileStoreOptions) (FileStore, error) {

	store := &fileStore{
		readOnly: false,
		path:     path,
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, PageSize, PageSize)
			},
		},
	}

	if options == nil {
		options = DefaultOptions
	}

	flag := os.O_RDWR
	if options.ReadOnly {
		flag = os.O_RDONLY
		store.readOnly = true
	}

	var err error
	// open the file; create a new file if it doesn't exist

	if store.file, err = os.OpenFile(store.path, flag|os.O_CREATE, mode); err != nil {
		store.Close()
		return nil, err
	}
	var fi os.FileInfo
	if fi, err = store.file.Stat(); err != nil {
		return nil, err
	}
	size := fi.Size()
	if size != 0 {
		store.count = int64(size / int64(PageSize))
		store.lastPageID = PageID(store.count - 1)
	} else {
		store.count = 0
		store.lastPageID = -1
	}

	// TODO: lock the file - see BoltDB's implementation

	return store, nil
}

// Path returns the filestore path.
func (store *fileStore) Path() string {
	return store.path
}

// Close closes the filestore.
func (store *fileStore) Close() error {
	return store.file.Close()
}

// Read returns the page with ID=id. Caller's responsibility to create page.
func (store *fileStore) Read(id PageID, page Page) error {
	if id > store.lastPageID {
		return errors.New("Invalid page ID")
	}

	buf := store.bufferPool.Get().([]byte)
	defer store.bufferPool.Put(buf)

	store.gets++
	if n, err := store.file.ReadAt(buf, int64(id)*int64(PageSize)); err != nil || n != int(PageSize) {
		return err
	}
	return page.UnmarshalBinary(buf)

}

// Write updates the page with id=ID.
func (store *fileStore) Write(id PageID, page Page) error {

	store.l.Lock()
	defer store.l.Unlock()

	// NB: file.WriteAt will just write at the end of the file if offset is past the end of the file
	// so check total pages to stop this happening
	if id > store.lastPageID {
		return errors.New("Invalid page ID")
	} else if buf, err := page.MarshalBinary(); err != nil {
		return err
	} else if _, err := store.file.WriteAt(buf, int64(id)*int64(PageSize)); err != nil {
		return err
	}
	store.sets++
	// TODO: filestore.Set - implement algorithm to do a file.Sync at set intervals
	return nil //store.file.Sync()
}

// New creates an empty page at the end of the database file.
// Returns the page ID of the new page. Page count & Last page ID will be increased by 1.
func (store *fileStore) New() (PageID, error) {

	store.l.Lock()
	defer store.l.Unlock()

	buf := store.bufferPool.Get().([]byte)
	defer store.bufferPool.Put(buf)
	for i := range buf {
		buf[i] = 0
	}
	if _, err := store.file.Write(buf); err != nil {
		return 0, err
	}
	store.lastPageID++
	store.count++
	store.news++
	return PageID(store.lastPageID), nil
}

// Append appends the given page at the end of the database file.
// Returns the page ID of the new page. Page count & Last page ID will be increased by 1.
func (store *fileStore) Append(page Page) (PageID, error) {

	store.l.Lock()
	defer store.l.Unlock()

	buf, err := page.MarshalBinary()
	if err != nil {
		return 0, err
	}
	if _, err = store.file.Write(buf); err != nil {
		return 0, err
	}
	store.lastPageID++
	store.count++
	//store.file.Sync()
	store.appends++
	return PageID(store.lastPageID), nil
}

// Wipe zeros out the specified page. It does not reduce page count. Use with care!
func (store *fileStore) Wipe(id PageID) error {

	store.l.Lock()
	defer store.l.Unlock()

	if id > store.lastPageID {
		return errors.New("Invalid page ID")
	}
	buf := store.bufferPool.Get().([]byte)
	defer store.bufferPool.Put(buf)
	for i := range buf {
		buf[i] = 0
	}

	if _, err := store.file.WriteAt(buf, int64(id)*int64(PageSize)); err != nil {
		return err
	}
	store.wipes++
	return nil //store.file.Sync()
}

// Count returns the total number of pages in the store.
func (store *fileStore) Count() int64 {
	return store.count
}

// Statistics returns a string with get/set/new/append counts.
func (store *fileStore) Statistics() string {
	return fmt.Sprintf("file store: gets: %d, sets: %d, news: %d, appends: %d", store.gets, store.sets, store.news, store.appends)
}
