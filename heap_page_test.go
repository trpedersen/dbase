package dbase

import (
	"bytes"
	"testing"
)

func Test_FillPage(t *testing.T) {
	page := NewHeapPage()
	recLen := 107
	for i := 0; page.GetFreeSpace() > recLen; i++ {
		record1 := make([]byte, recLen)
		record2 := make([]byte, recLen)
		for j := int16(0); j < int16(recLen); j++ {
			record1[j] = byte(i)
		}
		if recordNumber, err := page.AddRecord(record1); err != nil {
			t.Fatalf("page.AddRecord, err: %s", err)
		} else if _, err := page.GetRecord(recordNumber, record2); err != nil {
			t.Fatalf("page.GetRecord, err: %s", err)
		} else if bytes.Compare(record1, record2) != 0 {
			t.Errorf("bytes.Compare: expected %t, got %t", record1, record2)
			break
		}
	}
}

func Test_MarshalBinary(t *testing.T) {
	page1 := NewHeapPage()
	recLen := 107
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

	page2 := NewHeapPage()

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
		_, err := page1.GetRecord(i, record1)
		if err != nil {
			t.Fatalf("page1.GetRecord, err: %s", err)
		}
		_, err = page2.GetRecord(i, record2)
		if err != nil {
			t.Fatalf("page2.GetRecord, err: %s", err)
		}
		if bytes.Compare(record1, record2) != 0 {
			t.Errorf("bytes.Compare, expecting: %s, got: %s", record1, record2)
		}
	}

}

func Test_DeleteRecords(t *testing.T) {
	page := NewHeapPage()
	recLen := 97
	for i := 0; page.GetFreeSpace() > recLen; i++ {
		record1 := make([]byte, recLen)
		record2 := make([]byte, recLen)
		for j := int16(0); j < int16(recLen); j++ {
			record1[j] = byte(i)
		}
		if recordNumber, err := page.AddRecord(record1); err != nil {
			t.Fatalf("page.AddRecord, err: %s", err)
		} else if _, err := page.GetRecord(recordNumber, record2); err != nil {
			t.Fatalf("page.GetRecord, err: %s", err)
		} else if bytes.Compare(record1, record2) != 0 {
			t.Errorf("bytes.Compare: expected %t, got %t", record1, record2)
			break
		}
	}
	slotCount := page.GetSlotCount()

	buf := make([]byte, recLen, recLen)
	for i := int16(1); i < page.GetSlotCount(); i++ {
		page.DeleteRecord(i)
		_, err := page.GetRecord(i, buf)
		if err == nil {
			t.Errorf("page.DeleteRecord, expecting: record deleted, got: nil")
		}
	}

	expectedFreeSpace := int(slotTableLen - (page.GetSlotCount()+1)*slotTableEntryLen)

	if page.GetFreeSpace() != expectedFreeSpace {
		t.Errorf("freespace, expected: %d, got: %d", expectedFreeSpace, page.GetFreeSpace())
	}
	if slotCount != page.GetSlotCount() {
		t.Errorf("slotcount, expected: %d, got: %d", slotCount, page.GetSlotCount())
	}
}

func Test_ResizeRecords(t *testing.T) {
	page := NewHeapPage()

	count := 50
	recLen := page.GetFreeSpace() / count / 3 // need to take into account slots

	for i := 0; i < count; i++ {
		record1 := make([]byte, recLen)
		for j := int16(0); j < int16(recLen); j++ {
			record1[j] = byte(i)
		}
		if _, err := page.AddRecord(record1); err != nil {
			t.Fatalf("page.AddRecord, err: %s", err)
		}
	}

	slotCount := page.GetSlotCount()
	freespace1 := page.GetFreeSpace()

	// double the record size & resave
	recLen *= 2
	record2 := make([]byte, recLen)
	for i := int16(1); i < slotCount; i++ {
		for j := int16(0); j < int16(recLen); j++ {
			record2[j] = byte(i)
		}
		err := page.SetRecord(i, record2)
		if err != nil {
			t.Fatalf("page.SetRecord after resize, err: %s", err)
		}
	}

	freespace2 := page.GetFreeSpace()

	if freespace2 >= freespace1 {
		t.Errorf("freespace after resize, freespace before: %d, freespace after: %d", freespace1, freespace2)
	}

	//log.Println("freespace", freespace1, freespace2)

	// reset the record size & resave
	recLen /= 2
	record3 := make([]byte, recLen)
	for i := int16(1); i < slotCount; i++ {
		for j := int16(0); j < int16(recLen); j++ {
			record3[j] = byte(i)
		}
		err := page.SetRecord(i, record3)
		if err != nil {
			t.Fatalf("page.SetRecord after resize, err: %s", err)
		}
	}
	freespace3 := page.GetFreeSpace()
	if freespace1 != freespace3 {
		t.Errorf("freespace after re-resize, expecting: %d, got: %d", freespace1, freespace3)
	}
	//log.Println("freespace", freespace1, freespace3)

}

func BenchmarkFillPage(b *testing.B) {
	page := NewHeapPage()
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
			} else if _, err := page.GetRecord(recordNumber, record2); err != nil {
				b.Fatalf("page.GetRecord, err: %s", err)
			} else if bytes.Compare(record1, record2) != 0 {
				b.Errorf("bytes.Compare: expected %t, got %t", record1, record2)
				break
			}

		}
	}
}
