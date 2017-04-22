package dbase_test

import (
	"testing"

	"github.com/trpedersen/dbase"
)

func TestPageSetGetIDType(t *testing.T) {

	page := dbase.NewHeapPage().(dbase.Page)

	err := page.SetID(17)
	if err != nil {
		t.Fatalf("Page set ID, expected: nil, got: %s", err)
	}

	id := page.GetID()
	if id != 17 {
		t.Fatalf("Page get/set ID, expected: 17, got: %d", id)
	}

	pageType := page.GetType()
	if pageType != dbase.PAGE_TYPE_HEAP {
		t.Fatalf("Page type, expected: %d, got: %d", dbase.PAGE_TYPE_HEAP, pageType)
	}

}
