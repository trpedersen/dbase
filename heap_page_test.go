package dbase_test

import (
	"bytes"
	"github.com/trpedersen/dbase"
	"testing"
	"log"
)

func TestFillPage(t *testing.T) {
	page := dbase.NewHeapPage()
	recLen := 12
	for i := 0; page.GetFreeSpace() > recLen; i++ {
		record1 := make([]byte, recLen)
		record2 := make([]byte, recLen)
		for j := int16(0); j < int16(recLen); j++ {
			record1[j] = byte(i)
		}
		if recordNumber, err := page.AddRecord(record1); err != nil {
			t.Fatalf("page.AddRecord, err: %s", err)
		} else if err := page.GetRecord(recordNumber, record2); err != nil {
			t.Fatalf("page.GetRecord, err: %s", err)
		} else if bytes.Compare(record1, record2) != 0 {
			log.Println("TestFillPage", recordNumber, page)
			t.Errorf("bytes.Compare: expected %t, got %t", record1, record2)
			break
		}
	}
}

func TestMarshalBinary(t *testing.T) {
	page1 := dbase.NewHeapPage()
	recLen := 10
	for i := 0; page1.GetFreeSpace() > recLen; i++ {
		record := make([]byte, recLen)
		for j := 0; j < recLen; j++ {
			record[j] = byte(i)
		}
		if _, err := page1.AddRecord(record); err != nil {
			t.Fatalf("page.AddRecord, err: %s", err)
		}
	}

	pageBytes, err := page1.MarshalBinary()
	if err != nil {
		t.Fatalf("page1.MarshalBinary, err: %s", err)
	}

	page2 := dbase.NewHeapPage()

	err = page2.UnmarshalBinary(pageBytes)
	if err != nil {
		t.Fatalf("page2.UnmarshalBinary, err: %s", err)
	}

	if page1.GetSlotCount() != page2.GetSlotCount() {
		t.Errorf(".GetRecordCount, expecting: %d, got: %d", page1.GetSlotCount(), page2.GetSlotCount())
	}
	record1 := make([]byte, recLen)
	record2 := make([]byte, recLen)
	for i := int16(1); i < page1.GetSlotCount(); i++ {
		err := page1.GetRecord(i, record1)
		if err != nil {
			t.Fatalf("page1.GetRecord, err: %s", err)
		}
		err = page2.GetRecord(i, record2)
		if err != nil {
			t.Fatalf("page2.GetRecord, err: %s", err)
		}
		if bytes.Compare(record1, record2) != 0 {
			t.Errorf("bytes.Compare, expecting: %s, got: %s", record1, record2)
		}
	}

}

func BenchmarkFillPage(b *testing.B) {
	page := dbase.NewHeapPage()
	recLen := 10
	record1 := make([]byte, recLen)
	record2 := make([]byte, recLen)
	for k := 0; k < b.N; k++ {
		for i := int16(0); page.GetFreeSpace() > recLen; i++ {
			record := make([]byte, recLen)
			for j := 0; j < recLen; j++ {
				record[j] = byte(i)
			}
			if recordNumber, err := page.AddRecord(record1); err != nil {
				b.Fatalf("page.AddRecord, err: %s", err)
			} else if err := page.GetRecord(recordNumber, record2); err != nil {
				b.Fatalf("page.GetRecord, err: %s", err)
			} else if bytes.Compare(record1, record2) != 0 {
				b.Errorf("bytes.Compare: expected %t, got %t", record1, record2)
				break
			}

		}
	}
}
