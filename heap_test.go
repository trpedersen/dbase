package dbase_test

import (
	"testing"

	"github.com/trpedersen/dbase"
	randstr "github.com/trpedersen/rand"

	"bufio"
	"bytes"
	"log"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	HEAP_RUNS = 100
)

var store dbase.FileStore

func TestMain(m *testing.M) {
	path := tempfile()
	store, _ = dbase.Open(path, 0666, nil)

	//var err error
	//store, err = dbase.NewMemoryStore()
	//if err != nil {
	//	panic(err)
	//} else if store == nil {
	//	panic("expected db")
	//}
	defer func() {
		store.Close()
		os.Remove(store.Path())
	}()
	m.Run()
}

func Test_CreateHeap(t *testing.T) {

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
	//
	////path := tempfile()
	////store, err := dbase.Open(path, 0666, nil)
	////defer func() {
	////	store.Close()
	////	os.Remove(store.Path())
	////}()
	//store, err := dbase.NewMemoryStore()

	//if err != nil {
	//	t.Fatal(err)
	//} else if store == nil {
	//	t.Fatal("expected db")
	//}
	heap := dbase.NewHeap(store)

	heapRuns := 100000

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
			t.Fatalf("Write, err: %s", err)
		}
		if rid.PageID == 0 {
			t.Fatalf("RID zero")
		}
		if _, err = heap.Get(rid, record2); err != nil {
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

	//log.Println(heap.Statistics())
	//log.Println(store.Statistics())

	//store.Close()
	//store, err = dbase.Open(path, 0666, nil)
	//heap = dbase.NewHeap(store)

	count = heap.Count()
	if count != int64(heapRuns) {
		t.Fatalf("Record count, expected: %d, got: %d", heapRuns, count)
	}

	//log.Println(heap.Statistics())
	//log.Println(store.Statistics())
}

func Test_HeapDelete(t *testing.T) {
	//path := tempfile()
	//store, err := dbase.Open(path, 0666, nil)
	//defer func() {
	//	store.Close()
	//	os.Remove(store.Path())
	////}()
	//
	//if err != nil {
	//	t.Fatal(err)
	//} else if store == nil {
	//	t.Fatal("expected db")
	//}
	heap := dbase.NewHeap(store)

	record1 := []byte("DELETE ME")
	rid, err := heap.Write(record1)
	if err != nil {
		t.Fatalf("heap.Write, err: %s", err)
	}
	err = heap.Delete(rid)
	if err != nil {
		t.Fatalf("heap.Delete, err: %s", err)
	}
	_, err = heap.Get(rid, record1)
	if err == nil {
		t.Fatalf("Heap delete, expected: RECORD_DELETED, got: nil")
	}
	if _, ok := err.(dbase.RecordDeleted); !ok {
		t.Fatalf("Heap delete, expected: RECORD_DELETED, got: %s", err)
	}
}

func Test_FileUpload(t *testing.T) {

	datapath := "D:/algs4-data/leipzig1M.txt"
	//path := "D:/algs4-data/mobydick.txt"
	file, err := os.Open(datapath)
	if err != nil {
		panic(err)
	}

	storepath := tempfile()
	store, _ = dbase.Open(storepath, 0666, nil)

	defer func() {
		store.Close()
		os.Remove(store.Path())
		logElapsedTime(time.Now(), "Test_FileUpload")
		file.Close()
	}()


	heap := dbase.NewHeap(store)
	//heap.Clear()

	var heapWrites int

	var input = make([][]byte, 0, 20000)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) != 0 {
			b := []byte(line)
			heap.Write(b)
			//b := make([]byte, len(line))
			//copy(b, []byte(line))
			input = append(input, b)
			heapWrites += 1
		}
		//if heapWrites > 10 {
		//	break
		//}
	}
	log.Println(heap.Statistics())
	var heapReads int
	var n int
	heapScanner := dbase.NewHeapScanner(heap)
	buf := make([]byte, dbase.MAX_RECORD_LEN)
	_, n, err = heapScanner.Next(buf)
	//log.Println(rid, err)
	var successfulCompares int
	for err == nil {
		buf := buf[0:n]
		if bytes.Compare(input[heapReads], buf) != 0 {
			t.Fatalf("compare, \nexpecting: %s\n      got: %s", input[heapReads], buf)
		} else {
			successfulCompares += 1
		}
		heapReads += 1
		buf = buf[0:dbase.MAX_RECORD_LEN]
		_, n, err = heapScanner.Next(buf)
	}
	if heapReads != heapWrites {
		t.Errorf("File read, expected: %d, got: %d", heapWrites, heapReads)
	}
	log.Println(heapWrites, heapReads, successfulCompares)
}

func Test_FileUploadParallel(t *testing.T) {

	datapath := "D:/algs4-data/leipzig1M.txt"
	//path := "D:/algs4-data/mobydick.txt"
	file, err := os.Open(datapath)
	if err != nil {
		panic(err)
	}
	//
	//storepath := tempfile()
	//store, _ = dbase.Open(storepath, 0666, nil)
	store1, err := dbase.NewMemoryStore()

	defer logElapsedTime(time.Now(), "Test_FileUploadParallel")
	defer func() {
		store.Close()
		os.Remove(store.Path())
		file.Close()
	}()

	heap := dbase.NewHeap(store1)
	heap.Clear()

	var heapWrites int32
	var scanCount int
	var lineCount int
	var sends int

	var wg sync.WaitGroup

	type sample struct {
		rid dbase.RID
		buf []byte
	}
	samples := make([]*sample, 0, 2000)
	sampleQ := make(chan *sample)
	go func(){
		for sample := range sampleQ {
			samples = append(samples, sample)
		}
	}()


	scan := func() chan string {
		linech := make(chan string)
		go func() {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				scanCount += 1
				line := scanner.Text()
				if len(line) != 0 {
					lineCount += 1
					select {
					case linech <- line:
						sends += 1
						continue
					}
				}
			}
			//log.Println("scan finished, closing linech")
			close(linech)
		}()
		return linech
	}
	linech := scan()

	writer := func(i int) {
		var writes int
		defer wg.Done()
		for {
			select {
			case line, ok := <-linech:
				if ok {
					b := []byte(line)
					rid, err := heap.Write(b)
					writes += 1
					if err == nil && (writes % 1000 == 0) {
						sampleQ <- &sample{rid:rid, buf:b}
					}
					atomic.AddInt32(&heapWrites, 1)
				} else {
					//log.Println("writer quit", i, writes)
					return
				}
			}
		}
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go writer(i)
	}
	wg.Wait()

	close(sampleQ)
	//log.Println(samples)

	log.Println(heap.Statistics())
	log.Println("scanCount", scanCount, "lineCount", lineCount, "sends", sends)

	var heapReads int32
	heapScanner := dbase.NewHeapScanner(heap)

	reader := func(i int) {
		defer wg.Done()
		var n int
		var reads int
		buf := make([]byte, dbase.MAX_RECORD_LEN)

		_, n, err = heapScanner.Next(buf)
		for err == nil {
			reads += 1
			buf := buf[0:n]
			atomic.AddInt32(&heapReads, 1)
			buf = buf[0:dbase.MAX_RECORD_LEN]
			_, n, err = heapScanner.Next(buf)
		}
		//log.Println("reader quit", i, reads)
		return
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go reader(i)
	}
	wg.Wait()

	if heapReads != heapWrites {
		t.Errorf("File parallel write/read, expected: %d, got: %d", heapWrites, heapReads)
	}
	log.Println("heapWrites", heapWrites, "heapReads", heapReads)

	buf := make([]byte, dbase.MAX_RECORD_LEN)
	var goodCompares int
	for _, sample := range samples {
		//log.Println(sample.rid, sample.buf)
		buf = buf[0:dbase.MAX_RECORD_LEN]
		n, err := heap.Get(sample.rid, buf)
		if err != nil {
			t.Fatalf("get record, err: %s", err)
		}
		if bytes.Compare(sample.buf, buf[0:n]) != 0 {
			t.Fatalf("get record, expecting: %s, got: %s", sample.buf, buf[0:n])
		} else {
			goodCompares += 1
		}
	}
	log.Println("len(samples)", len(samples), "good compares", goodCompares)
}

func logElapsedTime(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
