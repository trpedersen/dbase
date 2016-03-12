package main

import (
	"bytes"
	"flag"
	"github.com/trpedersen/dbase"
	randstr "github.com/trpedersen/rand"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
)

func main() {
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	//memoryStore()
	Test_HeapDelete()
}

func memoryStore() {
	store, _ := dbase.NewMemoryStore()
	Test_HeapWrite(store)

}

func fileStore() {
	path := tempfile()
	store, err := dbase.Open(path, 0666, nil)
	defer func() {
		store.Close()
		os.Remove(store.Path())
	}()

	if err != nil {
		log.Fatal(err)
	} else if store == nil {
		log.Fatal("expected db")
	}

	Test_HeapWrite(store)
}

func Test_HeapWrite(store dbase.PageStore) {

	heap := dbase.NewHeap(store)

	heapRuns := 1000000

	rand.Seed(2323)
	l := rand.Intn(1000)
	if l == 0 {
		l = 1
	}
	record1 := []byte(randstr.RandStr(l, "alphanum"))
	record2 := make([]byte, len(record1))

	for i := 0; i < heapRuns; i++ {

		rid, err := heap.Write(record1)
		if err != nil {
			log.Fatalf("Write, err: %s", err)
		}
		if rid.PageID == 0 {
			log.Fatalf("RID zero")
		}
		if _, err = heap.Get(rid, record2); err != nil {
			log.Fatalf("heap.Get, err: %s", err)
		}
		if bytes.Compare(record1, record2) != 0 {
			log.Fatalf("bytes.Compare: expected %t, got %t", record1, record2)
			break
		}
	}
	count := heap.Count()
	if count != int64(heapRuns) {
		log.Fatalf("Record count, expected: %d, got: %d", heapRuns, count)
	}

	log.Println(heap.Statistics())
	log.Println(store.Statistics())

	count = heap.Count()
	if count != int64(heapRuns) {
		log.Fatalf("Record count, expected: %d, got: %d", heapRuns, count)
	}

	log.Println(heap.Statistics())
	log.Println(store.Statistics())
}

func Test_CreateHeap() {

	path := tempfile()
	store, err := dbase.Open(path, 0666, nil)
	defer func() {
		store.Close()
		os.Remove(store.Path())
	}()

	if err != nil {
		log.Fatal(err)
	} else if store == nil {
		log.Fatal("expected db")
	}
	heap := dbase.NewHeap(store)

	count := store.Count()
	if count != 2 {
		log.Fatalf("Page count, expected: 2, got: %d", count)
	}

	count = heap.Count()
	if count != 0 {
		log.Fatalf("Record count, expected: 0, got: %d", count)
	}

}

// tempfile returns a temporary file path.
func tempfile() string {

	f, err := ioutil.TempFile("d:/tmp", "db-")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	return f.Name()
}

func Test_HeapDelete() {
	path := tempfile()
	store, err := dbase.Open(path, 0666, nil)
	defer func() {
		store.Close()
		os.Remove(store.Path())
	}()

	if err != nil {
		log.Fatal(err)
	} else if store == nil {
		log.Fatal("expected db")
	}
	heap := dbase.NewHeap(store)

	record1 := []byte("DELETE ME")
	rid, err := heap.Write(record1)
	if err != nil {
		log.Fatalf("heap.Write, err: %s", err)
	}
	err = heap.Delete(rid)
	if err != nil {
		log.Fatalf("heap.Delete, err: %s", err)
	}
	_, err = heap.Get(rid, record1)
	if err == nil {
		log.Fatalf("Heap delete, expected: RECORD_DELETED, got: nil")
	}
	if _, ok := err.(dbase.RecordDeleted); !ok {
		log.Fatalf("Heap delete, expected: RECORD_DELETED, got: %s", err)
	}
}
