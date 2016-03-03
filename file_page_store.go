package dbase

import (
	"errors"
	"os"
	"sync"
)

type FileStore interface {
	PageStore
	Close() error
	Path() string
}

type FileStoreOptions struct {
	ReadOnly bool
}

var DefaultOptions = &FileStoreOptions{
	ReadOnly: false,
}

type fileStore struct {
	readOnly   bool
	path       string
	file       *os.File
	bufferPool *sync.Pool
	lastPageId PageID
	count      int64
	l          sync.Mutex
}

func Open(path string, mode os.FileMode, options *FileStoreOptions) (FileStore, error) {

	store := &fileStore{
		readOnly: false,
		path:     path,
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, PAGE_SIZE, PAGE_SIZE)
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
	if fi, err := store.file.Stat(); err != nil {
		return nil, err
	} else {
		size := fi.Size()
		if size != 0 {
			store.count = int64(size / int64(PAGE_SIZE))
			store.lastPageId = PageID(store.count - 1)
		} else {
			store.count = 0
			store.lastPageId = -1
		}
	}

	// TODO: lock the file - see BoltDB's implementation

	return store, nil
}

func (store *fileStore) Path() string {
	return store.path
}

func (store *fileStore) Close() error {
	return store.file.Close()
}

// Get returns the page with ID=id. Caller's responsibility to create page.
func (store *fileStore) Get(id PageID, page Page) error {
	if id > store.lastPageId {
		return errors.New("Invalid page ID")
	}

	buf := store.bufferPool.Get().([]byte)
	defer store.bufferPool.Put(buf)

	if n, err := store.file.ReadAt(buf, int64(id)*int64(PAGE_SIZE)); err != nil || n != int(PAGE_SIZE) {
		return err
	} else {
		return page.UnmarshalBinary(buf)
	}
}

// Set updates the page with id=ID.
func (store *fileStore) Set(id PageID, page Page) error {

	store.l.Lock()
	defer store.l.Unlock()

	// NB: file.WriteAt will just write at the end of the file if offset is past the end of the file
	// so check total pages to stop this happening
	if id > store.lastPageId {
		return errors.New("Invalid page ID")
	} else if buf, err := page.MarshalBinary(); err != nil {
		return err
	} else if _, err := store.file.WriteAt(buf, int64(id)*int64(PAGE_SIZE)); err != nil {
		return err
	}
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
	for i, _ := range buf {
		buf[i] = 0
	}
	if _, err := store.file.Write(buf); err != nil {
		return 0, err
	}
	store.lastPageId += 1
	store.count += 1
	return PageID(store.lastPageId), nil
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
	store.lastPageId += 1
	store.count += 1
	//store.file.Sync()
	return PageID(store.lastPageId), nil
}

// Wipe zeros out the specified page. It does not reduce page count. Use with care!
func (store *fileStore) Wipe(id PageID) error {

	store.l.Lock()
	defer store.l.Unlock()

	if id > store.lastPageId {
		return errors.New("Invalid page ID")
	}
	buf := store.bufferPool.Get().([]byte)
	defer store.bufferPool.Put(buf)
	for i, _ := range buf {
		buf[i] = 0
	}

	if _, err := store.file.WriteAt(buf, int64(id)*int64(PAGE_SIZE)); err != nil {
		return err
	}
	return nil //store.file.Sync()
}

// Count returns the total number of pages in the store.
func (store *fileStore) Count() int64 {
	return store.count
}
