// Package page provides low-level types and functions for managing data pages and records.
// Writing pages to disk is left to other packages.
package dbase

import (
	"encoding/binary"
)

const (
	PAGE_SIZE        = int16(8192) // PAGE_SIZE is typically the same as the filesystem blocksize
	PAGE_HEADER_LEN  = 192         // bytes
	MAX_PAGE_PAYLOAD = 8000        // Maximum page payload bytes

	// Page buffer offsets for page fields
	PAGE_ID_START  = 0
	PAGE_ID_END    = 8 // int64
	PAGE_TYPE_BYTE = 8 // single bytes

	// Page types:
	DB_HEADER_PAGE    = 0x01
	DB_DIRECTORY_PAGE = 0x02
	DATA_PAGE         = 0x03
	HEAP_HEADER_PAGE  = 0x04
)

type PageID int64

type Page interface {
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(buf []byte) error
}

type page struct {
	id       PageID // 0:8

	pinCount int
	dirty    bool

	pagetype byte   // 8
	header   []byte
	bytes    []byte // bytes contains all the data in the page, including header fields, free space and the records
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// The page is encoded as a []byte PAGE_SIZE long, ready for serialisation.
func (page *page) Marshal() ([]byte, error) {
	binary.LittleEndian.PutUint64(page.header[0:8], uint64(page.id))
	page.header[PAGE_TYPE_BYTE] = page.pagetype
	return page.bytes, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// PAGE_SIZE bytes are used to rehydrate the page.
func (page *page) Unmarshal(buf []byte) error {

	page.bytes = make([]byte, PAGE_SIZE, PAGE_SIZE)
	copy(page.bytes, buf)

	page.header = page.bytes[0:PAGE_HEADER_LEN]
	page.id = PageID(binary.LittleEndian.Uint64(page.header[PAGE_ID_START:PAGE_ID_END]))
	page.pagetype = page.header[PAGE_TYPE_BYTE]

	return nil
}
