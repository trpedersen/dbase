package dbase

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"bytes"
)

// Ensure that a file store can be opened without error.
func TestOpen(t *testing.T) {
	path := tempfile()
	store, err := Open(path, 0666, nil)
	defer func() {
		store.Close()
		os.Remove(store.Path())
	}()

	if err != nil {
		t.Fatal(err)
	} else if store == nil {
		t.Fatal("expected store")
	}

	if s := store.Path(); s != path {
		t.Fatalf("unexpected path: %s", s)
	}

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}

// Ensure that opening a database with a blank path returns an error.
func TestOpen_ErrPathRequired(t *testing.T) {
	_, err := Open("", 0666, nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

// Ensure that opening a database with a bad path returns an error.
func TestOpen_ErrNotExists(t *testing.T) {
	_, err := Open(filepath.Join(tempfile(), "bad-path"), 0666, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateAndReopenDB(t *testing.T) {

	path := tempfile()
	store, err := Open(path, 0666, nil)
	defer func() {
		store.Close()
		os.Remove(store.Path())
	}()

	if err != nil {
		t.Fatal(err)
	} else if store == nil {
		t.Fatal("expected db")
	}

	if pageCount := store.Count(); pageCount != 0 {
		t.Errorf("Invalid page count, expect: 0, got: %d", pageCount)
	}

	const pageRun = 1024

	for i := 0; i < pageRun; i++ {
		store.New()
	}

	store.Close()

	if store, err = Open(path, 0666, nil); err != nil {
		t.Fatalf("db.Open, err: %s", err)
	}
	if pageCount := store.Count(); pageCount != pageRun {
		t.Errorf("Invalid page count, expect: 0, got: %d", pageCount)
	}
	store.Close()

}

func TestNewPages(t *testing.T) {

	path := tempfile()
	store, err := Open(path, 0666, nil)
	defer func() {
		store.Close()
		os.Remove(store.Path())
	}()

	if err != nil {
		t.Fatal(err)
	} else if store == nil {
		t.Fatal("expected db")
	}

	const pageRun = 1024

	for i := 0; i < pageRun; i++ {
		store.New()
	}

	store.Close()

	if store, err = Open(path, 0666, nil); err != nil {
		t.Fatalf("db.Open, err: %s", err)
	}
	if pageCount := store.Count(); pageCount != pageRun {
		t.Errorf("Invalid page count, expect: 0, got: %d", pageCount)
	}
	store.Close()

}

func TestSetGet(t *testing.T) {

	path := tempfile()
	store, err := Open(path, 0666, nil)

	defer func() {
		store.Close()
		os.Remove(store.Path())
	}()

	if err != nil {
		t.Fatal(err)
	} else if store == nil {
		t.Fatal("expected db")
	}

	page := NewHeapPage()

	record1 := []byte("TESTING")
	var id PageID
	var slot int16
	if id, err = store.Append(page); err != nil {
		t.Fatalf("store.Append, err: %s", err)
	}
	page.SetID(id)
	if slot, err = page.AddRecord(record1); err != nil {
		t.Fatalf("page, err: %s", err)
	}

	if err = store.Write(id, page); err != nil {
		t.Fatalf("store.Set, err: %s", err)
	}
	page.Clear()
	if err = store.Read(id, page); err != nil {
		t.Fatalf("store.Get, err: %s", err)
	}
	record2 := make([]byte, len(record1))
	if _, err = page.GetRecord(slot, record2); err != nil {
		t.Fatalf("page.GetRecord, err: %s", err)
	}
	if bytes.Compare(record1, record2) != 0 {
		t.Fatalf("page.GetRecord, expected: %s, got: %s", record1, record2)
	}

}

// tempfile returns a temporary file path.
func tempfile() string {
	f, err := ioutil.TempFile("c:/tmp", "db-")
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
