package dbase_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/trpedersen/dbase"
)

// Ensure that a database can be opened without error.
func TestOpen(t *testing.T) {
	path := tempfile()
	db, err := dbase.Open(path, 0666, nil)
	defer os.Remove(db.Path())
	if err != nil {
		t.Fatal(err)
	} else if db == nil {
		t.Fatal("expected db")
	}

	if s := db.Path(); s != path {
		t.Fatalf("unexpected path: %s", s)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

// Ensure that opening a database with a blank path returns an error.
func TestOpen_ErrPathRequired(t *testing.T) {
	_, err := dbase.Open("", 0666, nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

// Ensure that opening a database with a bad path returns an error.
func TestOpen_ErrNotExists(t *testing.T) {
	_, err := dbase.Open(filepath.Join(tempfile(), "bad-path"), 0666, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

//
//func TestCreateDB(t *testing.T) {
//
//	db := NewDB()
//
//	tmpfile, err := ioutil.TempFile("", "DB")
//	if err != nil {
//		t.Fatal(err)
//	}
//	 // clean up
//
//	if err := db.OpenFile(tmpfile); err != nil {
//		t.Fatalf("db.OpenFile, err: %s", err)
//	}
//
//	if pageCount := db.GetTotalPageCount(); pageCount != 1 {
//		t.Errorf("Invalid page count, expect: 1, got: %d", pageCount)
//	}
//
//	db.Close()
//
//
//}

func TestCreateAndReopenDB(t *testing.T) {

	path := tempfile()
	db, err := dbase.Open(path, 0666, nil)
	defer os.Remove(db.Path())
	if err != nil {
		t.Fatal(err)
	} else if db == nil {
		t.Fatal("expected db")
	}

	if pageCount := db.GetTotalPageCount(); pageCount != 1 {
		t.Errorf("Invalid page count, expect: 1, got: %d", pageCount)
	}

	db.Close()

	if db, err = dbase.Open(path, 0666, nil); err != nil {
		t.Fatalf("db.Open, err: %s", err)
	}
	if pageCount := db.GetTotalPageCount(); pageCount != 1 {
		t.Errorf("Invalid page count, expect: 1, got: %d", pageCount)
	}
	db.Close()

}

// tempfile returns a temporary file path.
func tempfile() string {
	f, err := ioutil.TempFile("", "db-")
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
