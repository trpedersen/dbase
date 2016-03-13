package dbase_test

import (
	"bytes"
	"testing"

	"github.com/trpedersen/dbase"
	randstr "github.com/trpedersen/rand"
)

func TestFillOverflowPage(t *testing.T) {

	page := dbase.NewOverflowPage()

	pageID := dbase.PageID(2321323)
	prevID := dbase.PageID(213245456)
	nextID := dbase.PageID(192328347)

	page.SetID(pageID)
	page.SetPreviousPageID(prevID)
	page.SetNextPageID(nextID)

	segmentID := int32(1234)
	segment1 := make([]byte, dbase.MAX_SEGMENT_LEN)
	copy(segment1, randstr.RandStr(int(dbase.MAX_SEGMENT_LEN), "alphanum"))

	err := page.SetSegment(segmentID, segment1)
	if err != nil {
		t.Errorf("page.SetSegment, err: %s", err)
	}

	var buf []byte
	buf, err = page.MarshalBinary()
	if err != nil {
		t.Fatalf("page.MarshalBinary, err: %s", err)
	}
	err = page.UnmarshalBinary(buf)
	if err != nil {
		t.Fatalf("page.UnmarshalBinary, err: %s", err)
	}

	if pageID != page.GetID() {
		t.Errorf("pageID, expected: %d, got: %d", pageID, page.GetID())
	}

	if prevID != page.GetPreviousPageID() {
		t.Errorf("prevID, expected: %d, got: %d", prevID, page.GetPreviousPageID())
	}
	if nextID != page.GetNextPageID() {
		t.Errorf("nextID, expected: %d, got: %d", nextID, page.GetNextPageID())
	}

	if segmentID != page.GetSegmentID() {
		t.Errorf("segmentID, expected: %d, got: %d", segmentID, page.GetSegmentID())
	}
	if int(dbase.MAX_SEGMENT_LEN) != page.GetSegmentLength() {
		t.Errorf("segmentLength, expected: %d, got: %d", dbase.MAX_SEGMENT_LEN, page.GetSegmentLength())
	}
	segment2 := make([]byte, dbase.MAX_SEGMENT_LEN)
	var n int
	n , err = page.GetSegment(segment2)
	if err != nil {
		t.Fatalf("page.GetSegment, err: %s", err)
	}
	if n != len(segment2) {
		t.Errorf("page.GetSegment bytes read, expected %d, got: %d", len(segment2), n)
	}
	if bytes.Compare(segment1, segment2) != 0 {
		t.Errorf("bytes.Compare, expected: %s, got: %s", segment1, segment2)
	}

}
