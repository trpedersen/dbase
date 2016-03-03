package dbase

import "encoding/binary"

type DBHeaderPage struct {
	page
	lastPageId PageID // total number of pages in file
}

const (
	LAST_PAGE_ID_START = 9
	LAST_PAGE_ID_END   = 17
)

func NewDBHeaderPage() *DBHeaderPage {
	page := &DBHeaderPage{
		page: page{
			id:       0,
			pagetype: DB_HEADER_PAGE,
			bytes:    make([]byte, PAGE_SIZE, PAGE_SIZE),
		},
		lastPageId: 0,
	}
	page.header = page.bytes[0:PAGE_HEADER_LEN]
	return page
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// The page is encoded as a []byte PAGE_SIZE long, ready for serialisation.
func (page *DBHeaderPage) MarshalPage() ([]byte, error) {
	page.pagetype = DB_HEADER_PAGE
	page.page.Marshal()
	binary.LittleEndian.PutUint64(page.header[LAST_PAGE_ID_START:LAST_PAGE_ID_END], uint64(page.lastPageId))
	return page.bytes, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// PAGE_SIZE bytes are used to rehydrate the page.
func (page *DBHeaderPage) UnmarshalBinary(buf []byte) error {

	if len(buf) != int(PAGE_SIZE) {
		panic("Invalid buffer")
	}
	// check page type, panic if wrong
	pageType := buf[PAGE_TYPE_BYTE]
	if pageType&DB_HEADER_PAGE == 0 {
		panic("Invalid page type")
	}
	page.page.Unmarshal(buf)
	page.lastPageId = PageID(binary.LittleEndian.Uint64(page.header[LAST_PAGE_ID_START:LAST_PAGE_ID_END]))
	return nil
}
