package main

import (
	"log"
	"os"
	"github.com/trpedersen/dbase"
	"io/ioutil"
)

func main(){
	Test_CreateHeap()
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