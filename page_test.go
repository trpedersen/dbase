package dbase

import (
	"testing"
)

func TestPageSetGetIDType(t *testing.T) {

	page := NewHeapPage().(Page)

	err := page.SetID(17)
	if err != nil {
		t.Fatalf("Page set ID, expected: nil, got: %s", err)
	}

	id := page.GetID()
	if id != 17 {
		t.Fatalf("Page get/set ID, expected: 17, got: %d", id)
	}

	pageType := page.GetType()
	if pageType != pageTypeHeap {
		t.Fatalf("Page type, expected: %d, got: %d", pageTypeHeap, pageType)
	}

}
