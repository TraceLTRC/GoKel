package gokel

import (
	"testing"
)

// Getting the title of a work
func TestGetWorkURL(t *testing.T) {
	expectedWorkURL := "https://archiveofourown.org/works/35103526?view_adult=true&view_full_work=true"

	workId := "35103526"
	workURL := GetWorkURL(workId)

	if workURL != expectedWorkURL {
		t.Fatalf(`GetWorkURL resulted in: %s. Expected %s`, workURL, expectedWorkURL)
	}
}

func TestGetWorkValid(t *testing.T) {
	workId := "35103526"
	expectedWorkTitle := "I'm sorry that nothing happened"
	expectedArchiveWarning := ArchiveGraphicViolence

	w, err := GetWork(workId)
	if err != nil {
		t.Fatalf("An error occured! %v", err)
	}

	if w.WorkTitle != expectedWorkTitle {
		t.Fatalf("Resulting title: %s | Expected title: %s", w.WorkTitle, expectedWorkTitle)
	}

	if w.WorkWarnings&ArchiveGraphicViolence != expectedArchiveWarning {
		t.Fatalf("Resulting warning bitmask: %d | Expected warning bitmask: %d", w.WorkWarnings, expectedArchiveWarning)
	}
}
