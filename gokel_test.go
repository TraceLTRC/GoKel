package gokel

import (
	"encoding/json"
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
	workId := "29994336"
	expectedWorkTitle := "This Might As Well Happen (discontinued)"
	expectedArchiveWarning := ArchiveNoWarnings

	w, warns, err := GetWork(workId)
	if err != nil {
		t.Fatalf("An error occured! %v", err)
	}
	if len(warns) != 0 {
		t.Fatalf("Warns detected! %v", warns)
	}

	if w.WorkTitle != expectedWorkTitle {
		t.Fatalf("Resulting title: %s | Expected title: %s", w.WorkTitle, expectedWorkTitle)
	}

	if w.WorkWarnings&expectedArchiveWarning != expectedArchiveWarning {
		t.Fatalf("Resulting warning bitmask: %d | Expected warning bitmask: %d", w.WorkWarnings, expectedArchiveWarning)
	}
}

func TestJson(t *testing.T) {
	workId := "31107491"

	w, _, err := GetWork(workId)
	if err != nil {
		t.Fatalf("An error occured! %v", err)
	}

	marshalledJson, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("An error occured! %v", err)
	}

	unmarshalledJson := &Work{}
	err = json.Unmarshal(marshalledJson, unmarshalledJson)
	if err != nil {
		t.Fatalf("An error occured! %v", err)
	}
}
