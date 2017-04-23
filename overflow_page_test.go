package dbase

import (
	"bytes"
	"testing"

	"os"
	"time"

	randstr "github.com/trpedersen/rand"
)

func Test_FillOverflowPageAndMarshal(t *testing.T) {

	page := NewOverflowPage()

	pageID := PageID(2321323)
	prevID := PageID(213245456)
	nextID := PageID(192328347)

	page.SetID(pageID)
	page.SetPreviousPageID(prevID)
	page.SetNextPageID(nextID)

	segmentID := int32(1234)
	segment1 := make([]byte, maxSegmentLen)
	copy(segment1, randstr.RandStr(int(maxSegmentLen), "alphanum"))

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
	if int(maxSegmentLen) != page.GetSegmentLength() {
		t.Errorf("segmentLength, expected: %d, got: %d", maxSegmentLen, page.GetSegmentLength())
	}
	segment2 := make([]byte, maxSegmentLen)
	var n int
	n, err = page.GetSegment(segment2)
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

func Test_MultipleOverflowPages(t *testing.T) {

	//store, _ := NewMemoryStore()
	storepath := tempfile()
	store, _ := Open(storepath, 0666, nil)
	defer logElapsedTime(time.Now(), "Test_MultipleOverflowPages")
	defer func() {
		store.Close()
		os.Remove(store.Path())
	}()

	segmentCount := int32(200000)

	record1 := []byte(randstr.RandStr(int(segmentCount*int32(maxSegmentLen)), "alphanum"))

	var prevPage OverflowPage

	page := NewOverflowPage()
	firstPage := page
	for segmentID := int32(0); segmentID < segmentCount; segmentID++ {
		offset := segmentID * int32(maxSegmentLen)
		page.SetSegment(segmentID, record1[offset:offset+int32(maxSegmentLen)])
		pageID, _ := store.New()
		page.SetID(pageID)
		if prevPage != nil {
			page.SetPreviousPageID(prevPage.GetID())
			prevPage.SetNextPageID(page.GetID())
			store.Write(prevPage.GetID(), prevPage)
		}
		store.Write(page.GetID(), page)
		prevPage = page
		page = NewOverflowPage()
	}
	//log.Println(store, firstPage)
	record2 := make([]byte, segmentCount*int32(maxSegmentLen))
	page = firstPage
	pageID := page.GetID()
	for {
		err := store.Read(pageID, page)
		if err != nil {
			t.Fatalf("store.Get, err: %s", err)
		}
		offset := page.GetSegmentID() * int32(maxSegmentLen)
		page.GetSegment(record2[offset : offset+int32(maxSegmentLen)])
		pageID = page.GetNextPageID()
		if pageID <= 0 {
			break
		}
	}

	if bytes.Compare(record1, record2) != 0 {
		t.Errorf("bytes.Compare\n"+
			"expecting: %s\n"+
			"      got: %s", record1, record2)
	}
}
