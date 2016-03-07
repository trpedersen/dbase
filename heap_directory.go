package dbase

//
//import (
//	"encoding/binary"
//	"fmt"
//	"io"
//)
//
//const (
//	DIRECTORY_0_PAGE_ID = PageID(1)
//	DIRECTORY_SIZE      = 8000
//	MAX_DIRECTORY       = 100
//
//	DIR_PAGE_ON_DISK      = 0200
//	DIR_PAGE_ALLOCATED    = 0100
//	DIR_PAGE_DELETED_ROWS = 0010
//	DIR_PAGE_EMPTY        = 0000
//	DIR_PAGE_1_50         = 0001
//	DIR_PAGE_51_80        = 0002
//	DIR_PAGE_81_95        = 0003
//	DIR_PAGE_96_100       = 0004
//
//	DIRECTORY_NUMBER_START     = 9
//	DIRECTORY_NUMBER_END       = 11
//	ALLOCATED_PAGE_COUNT_START = 11
//	ALLOCATED_PAGE_COUNT_END   = 13
//	FREE_PAGE_COUNT_START      = 13
//	FREE_PAGE_COUNT_END        = 15
//	PAGE_LIST_START            = PAGE_HEADER_LEN
//	PAGE_LIST_END              = PAGE_LIST_START + DIRECTORY_SIZE
//)
//
//type DBDirectoryPage struct {
//	page
//	directoryNumber    uint16
//	allocatedPageCount uint16
//	freePageCount      uint16
//	pageList           []byte
//}
//
//func NewDBDirectoryPage(directoryNumber uint16) *DBDirectoryPage {
//	page := &DBDirectoryPage{
//		page: page{
//			id:       0,
//			pagetype: DB_DIRECTORY_PAGE,
//			bytes:    make([]byte, PAGE_SIZE, PAGE_SIZE),
//		},
//		allocatedPageCount: 0,
//		freePageCount:      0,
//		directoryNumber:    directoryNumber,
//	}
//	page.header = page.bytes[0:PAGE_HEADER_LEN]
//	page.pageList = page.bytes[PAGE_HEADER_LEN:DIRECTORY_SIZE]
//	if directoryNumber == 0 {
//		// first directory, allocate header page and directory page
//		page.pageList[0] |= DIR_PAGE_ON_DISK | DIR_PAGE_ALLOCATED
//		page.pageList[1] |= DIR_PAGE_ON_DISK | DIR_PAGE_ALLOCATED
//	}
//	return page
//}
//
//// MarshalBinary implements the encoding.BinaryMarshaler interface.
//// The page is encoded as a []byte PAGE_SIZE long, ready for serialisation.
//func (page *DBDirectoryPage) MarshalBinary() ([]byte, error) {
//	page.pagetype = DB_DIRECTORY_PAGE
//	page.page.Marshal()
//	binary.LittleEndian.PutUint16(page.header[DIRECTORY_NUMBER_START:DIRECTORY_NUMBER_END], uint16(page.directoryNumber))
//	binary.LittleEndian.PutUint16(page.header[ALLOCATED_PAGE_COUNT_START:ALLOCATED_PAGE_COUNT_END], uint16(page.allocatedPageCount))
//	binary.LittleEndian.PutUint16(page.header[FREE_PAGE_COUNT_START:FREE_PAGE_COUNT_END], uint16(page.freePageCount))
//	return page.bytes, nil
//}
//
//// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
//// PAGE_SIZE bytes are used to rehydrate the page.
//func (page *DBDirectoryPage) UnmarshalBinary(buf []byte) error {
//
//	if len(buf) != int(PAGE_SIZE) {
//		panic("Invalid buffer")
//	}
//	// check page type, panic if wrong
//	pageType := buf[PAGE_TYPE_BYTE]
//	if pageType&DB_DIRECTORY_PAGE == 0 {
//		panic("Invalid page type")
//	}
//	page.page.Unmarshal(buf)
//	page.directoryNumber = binary.LittleEndian.Uint16(page.header[DIRECTORY_NUMBER_START:DIRECTORY_NUMBER_END])
//	page.allocatedPageCount = binary.LittleEndian.Uint16(page.header[ALLOCATED_PAGE_COUNT_START:ALLOCATED_PAGE_COUNT_END])
//	page.freePageCount = binary.LittleEndian.Uint16(page.header[FREE_PAGE_COUNT_START:FREE_PAGE_COUNT_END])
//	page.pageList = page.bytes[PAGE_LIST_START:PAGE_LIST_END]
//	return nil
//}
//
//type directory struct {
//	pageStore      PageStore
//	directoryPages []*DBDirectoryPage
//}
//
//func NewDBDirectory(pageStore PageStore) PageDirectory {
//	dir := &directory{
//		pageStore:      pageStore,
//		directoryPages: make([]*DBDirectoryPage, 0, MAX_DIRECTORY),
//	}
//
//	page := NewDBDirectoryPage(0)
//	err := dir.pageStore.Get(DIRECTORY_0_PAGE_ID, page)
//	if err != nil {
//		if err == io.EOF {
//			// new store, write first directory page
//			if err := dir.pageStore.Set(DIRECTORY_0_PAGE_ID, page); err != nil {
//				panic(fmt.Sprintf("NewDBDirectory, err: %s", err))
//			}
//		} else {
//			panic(fmt.Sprintf("NewDBDirectory, err: %s", err))
//		}
//	}
//
//	return dir
//}
//
//func (dir *directory) AllocatePage() (PageID, error) {
//	return 0, nil
//}
//
//func (dir *directory) DeallocatePage(id PageID) error {
//	return nil
//}
//
//func (dir *directory) Count() int64 {
//	return 0
//}
//
//func (dir *directory) AllocatedCount() int64 {
//	return 0
//}
