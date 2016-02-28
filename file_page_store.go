package dbase

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

const ()

type FilePageStore interface {
	Open(path string, mode os.FileMode, options *FileStoreOptions) error
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
	headerPage *DBHeaderPage
	file       *os.File
	bufferPool *sync.Pool
}

func Open(path string, mode os.FileMode, options *FileStoreOptions) (FilePageStore, error) {

	store := &fileStore{
		readOnly:   false,
		path:       path,
		headerPage: NewDBHeaderPage(),
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

	// TODO: lock the file - see BoltDB's implementation

	// Initialise the DB if necessary
	fi, err := store.file.Stat()
	if err != nil {
		panic(fmt.Sprintf("store.file.Stat(), err: %s", err))
	}
	if fi.Size() == 0 {
		// new database
		store.init() // TODO: error handling
	} else {
		// existing database
		buf := make([]byte, PAGE_SIZE, PAGE_SIZE)
		if err := store.readHeaderPage(buf); err != nil {
			panic(fmt.Sprintf("Invalid database, err: %s", err))
		}
		if err := store.headerPage.UnmarshalBinary(buf); err != nil {
			panic(fmt.Sprintf("Invalid database header, err: %s", err))
		}
	}
	return store, err
}

func (store *fileStore) Path() string {
	return store.path
}

func (store *fileStore) init() error {
	store.headerPage = NewDBHeaderPage()
	if err := store.WritePage(0, store.headerPage); err != nil {
		panic(fmt.Sprintf("store.headerPage.MarshalBinary, err: %s", err))
	}
	return nil
}

func (store *fileStore) Close() error {
	return store.file.Close()
}


// Allocate creates an empty page at the end of the database file.
// Returns the page ID of the new page.
func (store *fileStore) Allocate() (pageid, error) {
	buf := make([]byte, PAGE_SIZE, PAGE_SIZE)
	if _, err := store.file.Write(buf); err != nil {
		panic(fmt.Sprintf("store.NewPage, err: %s", err))
	}
	store.headerPage.lastPageId += 1
	store.WritePage(0, store.headerPage)
	return pageid(store.headerPage.lastPageId), nil
}

func (store *fileStore) readHeaderPage(buf []byte) error {
	if len(buf) != int(PAGE_SIZE) {
		panic("Invalid page buffer")
	}
	if _, err := store.file.ReadAt(buf, 0); err != nil {
		panic(fmt.Sprintf("store.ReadPage, err: %s", err))
	}
	return nil
}

func (store *fileStore) Get(id pageid, page Page) error {
	if id > store.headerPage.lastPageId {
		panic("Invalid page ID")
	}

	buf := store.bufferPool.Get().([]byte)
	defer store.bufferPool.Put(buf)

	if _, err := store.file.ReadAt(buf, int64(id)*int64(PAGE_SIZE)); err != nil {
		return err //		panic(fmt.Sprintf("store.ReadPage, err: %s", err))
	} else {
		return page.UnmarshalBinary(buf)
	}

}

func (store *fileStore) Set(id pageid, page Page) error {

	// NB: file.WriteAt will just write at the end of the file if offset is past the end of the file
	// so check total pages to stop this happening
	if id > store.headerPage.lastPageId {
		return errors.New("Invalid page ID")
	} else if buf, err := page.MarshalBinary(); err != nil {
		return err
	} else if _, err := store.file.WriteAt(buf, int64(id)*int64(PAGE_SIZE)); err != nil {
		return err
	}
	return store.file.Sync()
}

func (store *fileStore) Count() int64 {
	return int64(store.headerPage.lastPageId) + 1
}
