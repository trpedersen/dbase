package dbase

import (
	"fmt"
	"sync"
)

// Uses techniques from go-bitmap: https://github.com/boljen/go-bitmap

const (
	bitMapLen = 8 * 1000 // ties in with AllocationPage.allocationMap size
)

var (
	orMasks = [8]byte{1, 2, 4, 8, 16, 32, 64, 128}
	// {00000001, 00000010, 00000100, 00001000, 00010000, 00100000, 01000000, 10000000}
	andMasks = [8]byte{254, 253, 251, 247, 239, 223, 191, 127}
	// {11111110, 11111101, 11111011, 11110111, 11101111, 11011111, 10111111, 01111111}
)

// AllocationBitMap is a bit map containing 8 * 1000 * 8 = 64,000 allocation bits.
type AllocationBitMap interface {
	// Allocate runlength pages. Function will search for first available space runlength in length, returns starting bit if successful, -1 + err if not
	Allocate(runlength uint16) (bit uint16, err error)
	// AllocateExplicit allocates pages explicitly starting at bit for runlength length.
	AllocateExplicit(bit uint16, runlength uint16) (err error)
	// Deallocate pages starting at bit, for runlength length.
	Deallocate(bit uint16, runlength uint16) (err error)

	// IsAllocated returns true if any bits from bit:bit+runlength are 1
	IsAllocated(bit uint16, runlength uint16) (bool, error)

	// GetBit(int) (bool, error)
	// SetBit(bit int, value bool) error
	// SetBits(bit int, runlength int, value bool) error
	// GetBytes() []byte
}

type allocationBitMap struct {
	sync.Mutex
	bytes  []byte
	maxBit uint16
}

// NewAllocationBitMap returns a new allocation bit map from the provided []byte. Bitmap will be not be initialised - the assumption
// is that bytes contains a pre-existing bit map.
// It is intended that bytes is provided from an AllocationPage
func NewAllocationBitMap(bytes []byte) AllocationBitMap {

	return &allocationBitMap{
		bytes:  bytes, // make([]byte, bitMapLen, bitMapLen),
		maxBit: uint16((len(bytes) * 8) - 1),
		//	l:      &sync.Mutex{},
	}
}

func (bitmap *allocationBitMap) Allocate(runlength uint16) (bit uint16, err error) {

	if (runlength - 1) > bitmap.maxBit {
		return 0, fmt.Errorf("Allocate: Invalid args: runlength %v passes bitmap.maxBit %v (0 based)", runlength, bitmap.maxBit)
	}
	bitmap.Lock()
	defer bitmap.Unlock()

	bit = 0

	var short bool
	allocated := false
	var probe uint16
	for bit = 0; bit < (bitmap.maxBit - runlength); bit++ {
		if !bitmap.getBit(bit) {
			short = false
			for probe = bit; ((probe - bit) < runlength) && (probe < bitmap.maxBit); probe++ {
				if bitmap.getBit(probe) {
					short = true
					break
				}
			}
			if short {
				bit = probe
			} else {
				bitmap.setBits(bit, probe-bit, true)
				allocated = true
				break
			}
		}
	}
	if !allocated {
		return 0, fmt.Errorf("Allocate: unable to allocate %v bits", runlength) // TODO: make this a typed error so can distinguish
	}

	return bit, nil
}

// AllocateExplicit forces allocations for bit:runlength to true
func (bitmap *allocationBitMap) AllocateExplicit(bit uint16, runlength uint16) error {

	if (bit + runlength - 1) > bitmap.maxBit {
		return fmt.Errorf("AllocateExplicit: Invalid args: bit %v + runlength %v passes bitmap.maxBit %v (0 based)", bit, runlength, bitmap.maxBit)
	}

	bitmap.Lock()
	defer bitmap.Unlock()

	bitmap.setBits(bit, runlength, true)

	return nil
}

// Deallocate forces allocations for bit:runlength to false
func (bitmap *allocationBitMap) Deallocate(bit uint16, runlength uint16) (err error) {

	if (bit + runlength - 1) > bitmap.maxBit {
		return fmt.Errorf("Deallocate: Invalid args: bit %v + runlength %v passes bitmap.maxBit %v (0 based)", bit, runlength, bitmap.maxBit)
	}

	bitmap.Lock()
	defer bitmap.Unlock()

	bitmap.setBits(bit, runlength, false)

	return nil
}

// IsAllocated returns true if any of the bits in bit:runlength are true
func (bitmap *allocationBitMap) IsAllocated(bit uint16, runlength uint16) (bool, error) {

	if (bit + runlength - 1) > bitmap.maxBit {
		return false, fmt.Errorf("IsAllocated: Invalid args: bit %v + runlength %v passes bitmap.maxBit %v (0 based)", bit, runlength, bitmap.maxBit)
	}

	bitmap.Lock()
	defer bitmap.Unlock()

	result := false
	for i := uint16(0); i < runlength; i++ {
		result = result || bitmap.getBit(bit)
		bit++
	}
	return result, nil
}

//
// Private functions
//
//

func (bitmap *allocationBitMap) getBytes() []byte {
	return bitmap.bytes
}

func (bitmap *allocationBitMap) getBit(bit uint16) bool {

	// if bit > bitmap.maxBit {
	// 	return false, fmt.Errorf("GetBit: Invalid args: bit %v > bitmap.maxBit %v (0 based)", bit, bitmap.maxBit)
	// }

	return (bitmap.bytes[bit/8]&orMasks[bit%8] != 0)
}

func (bitmap *allocationBitMap) setBit(bit uint16, value bool) {

	// if bit > bitmap.maxBit {
	// 	return fmt.Errorf("SetBit: Invalid args: bit %v > bitmap.maxBit %v (0 based)", bit, bitmap.maxBit)
	// }
	index := bit / 8
	if value {
		bitmap.bytes[index] = bitmap.bytes[index] | orMasks[bit%8]
	} else {
		bitmap.bytes[index] = bitmap.bytes[index] & andMasks[bit%8]
	}
}

func (bitmap *allocationBitMap) setBits(bit uint16, runlength uint16, value bool) {

	// if (bit + runlength - 1) > bitmap.maxBit {
	// 	return fmt.Errorf("SetBits: Invalid args: bit %v + runlength %v >= bitmap.maxBit %v (0 based)", bit, runlength, bitmap.maxBit)
	// }

	for i := uint16(0); i < runlength; i++ {
		bitmap.setBit(bit, value)
		bit++
	}
}
