package dbase

import (
	"bufio"
	"bytes"
	"log"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	randstr "github.com/trpedersen/rand"
)

const (
	heapRuns = 100
)

var store FileStore

func TestMain(m *testing.M) {

	path := tempfile()
	var err error

	store, err = Open(path, 0666, nil)
	//store, err = NewMemoryStore()

	if err != nil {
		panic(err)
	} else if store == nil {
		panic("expected db")
	}
	defer func() {
		store.Close()
		os.Remove(store.Path())
	}()
	m.Run()
}

func Test_CreateHeap(t *testing.T) {

	heap := NewHeap(store)

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

	heap := NewHeap(store)

	heapRuns := 100000

	rand.Seed(2323)
	l := rand.Intn(1000)
	if l == 0 {
		l = 1
	}
	record1 := []byte(randstr.RandStr(l, "alphanum"))
	record2 := make([]byte, len(record1))

	for i := 0; i < heapRuns; i++ {

		rid, err := heap.Put(record1)
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
	count = heap.Count()
	if count != int64(heapRuns) {
		t.Fatalf("Record count, expected: %d, got: %d", heapRuns, count)
	}

}

func Test_HeapDelete(t *testing.T) {

	heap := NewHeap(store)

	record1 := []byte("DELETE ME")
	rid, err := heap.Put(record1)
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
	if _, ok := err.(RecordDeleted); !ok {
		t.Fatalf("Heap delete, expected: RECORD_DELETED, got: %s", err)
	}
}

func Test_FileUploadSequential(t *testing.T) {

	datapath := "d:/algs4-data/leipzig1M.txt"
	//path := "D:/algs4-data/mobydick.txt"
	file, err := os.Open(datapath)
	if err != nil {
		panic(err)
	}

	storepath := tempfile()
	store2, _ := Open(storepath, 0666, nil)

	defer logElapsedTime(time.Now(), "Test_FileUploadSequential")
	defer func() {
		store2.Close()
		os.Remove(store2.Path())
		file.Close()
	}()

	heap := NewHeap(store2)

	var heapWrites int

	var input = make([][]byte, 0, 20000)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) != 0 {
			b := []byte(line)
			heap.Put(b)
			input = append(input, b)
			heapWrites++
		}
	}
	log.Println(heap.Statistics())
	var heapReads int
	var n int
	heapScanner := NewHeapScanner(heap)
	buf := make([]byte, maxRecordLen)
	_, n, err = heapScanner.Next(buf)
	var successfulCompares int
	for err == nil {
		buf := buf[0:n]
		if bytes.Compare(input[heapReads], buf) != 0 {
			t.Fatalf("compare, \nexpecting: %s\n      got: %s", input[heapReads], buf)
		} else {
			successfulCompares++
		}
		heapReads++
		buf = buf[0:maxRecordLen]
		_, n, err = heapScanner.Next(buf)
	}
	if heapReads != heapWrites {
		t.Errorf("File read, expected: %d, got: %d", heapWrites, heapReads)
	}
	log.Println(heapWrites, heapReads, successfulCompares)
}

func Test_FileUploadParallel(t *testing.T) {

	datapath := "d:/algs4-data/leipzig1M.txt"
	//path := "D:/algs4-data/mobydick.txt"
	file, err := os.Open(datapath)
	if err != nil {
		panic(err)
	}
	store1, err := NewMemoryStore()

	defer logElapsedTime(time.Now(), "Test_FileUploadParallel")
	defer func() {
		store.Close()
		os.Remove(store.Path())
		file.Close()
	}()

	heap := NewHeap(store1)
	heap.Clear()

	var heapWrites int32
	var scanCount int
	var lineCount int
	var sends int

	var wg sync.WaitGroup

	type sample struct {
		rid RID
		buf []byte
	}
	samples := make([]*sample, 0, 2000)
	sampleQ := make(chan *sample)
	go func() {
		for sample := range sampleQ {
			samples = append(samples, sample)
		}
	}()

	scan := func() chan string {
		linech := make(chan string)
		go func() {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				scanCount++
				line := scanner.Text()
				if len(line) != 0 {
					lineCount++
					select {
					case linech <- line:
						sends++
						continue
					}
				}
			}
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
					rid, err := heap.Put(b)
					writes++
					if err == nil && (writes%1000 == 0) {
						sampleQ <- &sample{rid: rid, buf: b}
					}
					atomic.AddInt32(&heapWrites, 1)
				} else {
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

	log.Println(heap.Statistics())
	log.Println("scanCount", scanCount, "lineCount", lineCount, "sends", sends)

	var heapReads int32
	heapScanner := NewHeapScanner(heap)

	reader := func(i int) {
		defer wg.Done()
		var n int
		var reads int
		buf := make([]byte, maxRecordLen)

		_, n, err = heapScanner.Next(buf)
		for err == nil {
			reads++
			buf := buf[0:n]
			atomic.AddInt32(&heapReads, 1)
			buf = buf[0:maxRecordLen]
			_, n, err = heapScanner.Next(buf)
		}
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

	buf := make([]byte, maxRecordLen)
	var goodCompares int
	for _, sample := range samples {
		buf = buf[0:maxRecordLen]
		n, err := heap.Get(sample.rid, buf)
		if err != nil {
			t.Fatalf("get record, err: %s", err)
		}
		if bytes.Compare(sample.buf, buf[0:n]) != 0 {
			t.Fatalf("get record, expecting: %s, got: %s", sample.buf, buf[0:n])
		} else {
			goodCompares++
		}
	}
	log.Println("len(samples)", len(samples), "good compares", goodCompares)
}

func logElapsedTime(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
