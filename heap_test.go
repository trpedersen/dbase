package dbase_test

import (
	"github.com/trpedersen/dbase"
	"os"
	"testing"
	//"log"
	"bytes"
	randstr "github.com/trpedersen/rand"

	"math/rand"
)

const (
	HEAP_RUNS = 100
)

func Test_CreateHeap(t *testing.T) {

	path := tempfile()
	store, err := dbase.Open(path, 0666, nil)
	defer func(){
		store.Close()
		os.Remove(store.Path())
	}()

	if err != nil {
		t.Fatal(err)
	} else if store == nil {
		t.Fatal("expected db")
	}
	heap := dbase.NewHeap(store)

	count := store.Count()
	if count != 2 {
		t.Fatalf("Page count, expected: 2, got: %d", count)
	}

	count = heap.Count()
	if count != 0 {
		t.Fatalf("Record count, expected: 0, got: %d", count)
	}

}

func Test_HeapWrite(t *testing.T) {

	path := tempfile()
	store, err := dbase.Open(path, 0666, nil)
	defer func(){
		store.Close()
		os.Remove(store.Path())
	}()

	if err != nil {
		t.Fatal(err)
	} else if store == nil {
		t.Fatal("expected db")
	}
	heap := dbase.NewHeap(store)

	heapRuns := 10000

	rand.Seed(2323)
	for i := 0; i < heapRuns; i++ {
		l := rand.Intn(1023)
		record1 := []byte(randstr.RandStr(l, "alphanum"))
		record2 := make([]byte, len(record1))

		rid, err := heap.Write(record1)
		if err != nil {
			t.Fatalf("Write, err: %s", err)
		}
		if rid.PageID == 0 {
			t.Fatalf("RID zero")
		}
		if err = heap.Get(rid, record2); err != nil {
			t.Fatalf("heap.Get, err: %s", err)
		}
		if bytes.Compare(record1, record2) != 0 {
			t.Fatalf("bytes.Compare: expected %t, got %t", record1, record2)
			break
		}
	}
	count := heap.Count()
	if count != int64(heapRuns) {
		t.Fatalf("Record count, expected: %d, got: %d", heapRuns, count)
	}

	store.Close()
	store, err = dbase.Open(path, 0666, nil)
	heap = dbase.NewHeap(store)

	count = heap.Count()
	if count != int64(heapRuns) {
		t.Fatalf("Record count, expected: %d, got: %d", heapRuns, count)
	}
}
