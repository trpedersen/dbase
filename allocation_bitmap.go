package dbase

import (
	"fmt"
)

// Uses techniques from go-bitmap: https://github.com/boljen/go-bitmap

const (
	bitMapLen = 8 * 1000 // ties in with AllocationPage.allocationMap size
)

var (
	tA = [8]byte{1, 2, 4, 8, 16, 32, 64, 128}
	// {00000001, 00000010, 00000100, 00001000, 00010000, 00100000, 01000000, 10000000}
	tB = [8]byte{254, 253, 251, 247, 239, 223, 191, 127}
	// {11111110, 11111101, 11111011, 11110111, 11101111, 11011111, 10111111, 01111111}
)

// AllocationBitMap is a bit map containing 8 * 1000 * 8 = 64,000 allocation bits.
type AllocationBitMap interface {
	// Allocate runsize pages. Function will search for first available space runsize in length, returns starting offset if successful, -1 + err if not
	Allocate(runsize int) (offset int, err error)
	// Deallocate pages starting at offset, for runsize length.
	Deallocate(offset int, runsize int) (err error)
	// AllocateExplicit allocates pages explicitly starting at offset for runsize length.
	AllocateExplicit(offset int, runsize int) (err error)
	// IsAllocated returns true if any bits from offset:offset+runsize are 1
	IsAllocated(offset int, runsize int) (bool, error)

	GetBytes() []byte
}

type allocationBitMap struct {
	bytes []byte
}

// NewAllocationBitMap returns a new allocation bit map from the provided []byte. Bitmap will be not be initialised - the assumption
// is that bytes contains a pre-existing bit map.
// It is intended that bytes is provided from an AllocationPage
func NewAllocationBitMap(bytes []byte) AllocationBitMap {

	return &allocationBitMap{
		bytes: bytes, // make([]byte, bitMapLen, bitMapLen),
	}
}

func (bitmap *allocationBitMap) GetBytes() []byte {
	return bitmap.bytes
}

func (bitmap *allocationBitMap) Allocate(runsize int) (offset int, err error) {
	return -1, nil
}

// AllocateExplicit forces allocations for offset:runsize to true
func (bitmap *allocationBitMap) AllocateExplicit(offset int, runsize int) error {
	return bitmap.setbits(offset, runsize, true)
}

// Deallocate forces allocations for offset:runsize to false
func (bitmap *allocationBitMap) Deallocate(offset int, runsize int) (err error) {
	return bitmap.setbits(offset, runsize, false)
}

// IsAllocated returns true if any of the bits in offset:runsize are true
func (bitmap *allocationBitMap) IsAllocated(offset int, runsize int) (bool, error) {

	if (offset+runsize)/8 > len(bitmap.bytes) {
		return false, fmt.Errorf("Invalid args: offset+runsize > bitmap length")
	}

	result := false

	var bit int
	for i := 0; i < runsize; i++ {
		bit = (offset + i)
		result = result || (bitmap.bytes[bit/8]&tA[bit%8] != 0)
	}
	return result, nil
}

func (bitmap *allocationBitMap) setbits(offset int, runsize int, value bool) error {

	if (offset+runsize)/8 > len(bitmap.bytes) {
		return fmt.Errorf("Invalid args: offset+runsize > bitmap length")
	}

	var bit int
	var index int
	for i := 0; i < runsize; i++ {
		bit = (offset + i)
		index = bit / 8
		if value {
			bitmap.bytes[index] = bitmap.bytes[index] | tA[bit%8]
		} else {
			bitmap.bytes[index] = bitmap.bytes[index] & tB[bit%8]
		}
	}
	return nil
}
