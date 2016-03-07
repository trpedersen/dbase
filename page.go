// Package page provides low-level types and functions for managing data pages and records.
// Writing pages to disk is left to other packages.
package dbase

import ()

const (
	PAGE_SIZE        = int16(8192)                 // PAGE_SIZE is typically the same as the filesystem blocksize
	PAGE_HEADER_LEN  = 56                          // bytes
	MAX_PAGE_PAYLOAD = PAGE_SIZE - PAGE_HEADER_LEN // Maximum page payload bytes

	// Page buffer offsets for page fields
	PAGE_ID_OFFSET   = 0
	PAGE_TYPE_OFFSET = 8 // single bytes

	// Page types:
	DB_HEADER_PAGE        = PageType(0x01)
	DB_DIRECTORY_PAGE     = PageType(0x02)
	PAGE_TYPE_HEAP        = PageType(0x03)
	PAGE_TYPE_HEAP_HEADER = PageType(0x04)
)

type PageID int64
type PageType byte

type Page interface {
	GetID() PageID
	SetID(id PageID) error
	GetType() PageType
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(buf []byte) error
}

type page struct {
	id       PageID   // 0:8
	pagetype PageType // 8
	header   []byte
	bytes    []byte
}

func (page *page) GetID() PageID {
	return page.id
}

func (page *page) SetID(id PageID) error {
	page.id = id
	return nil
}

func (page *page) GetType() PageType {
	return page.pagetype
}
