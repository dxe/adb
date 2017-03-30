package main

import (
	"testing"

	"github.com/directactioneverywhere/adb/model"
)

func TestAutocompleteActivistsHandler(t *testing.T) {
	db := model.NewDB(":memory:")

	_, err := GetOrCreateUser(db, "User One")
	if err != nil {
		t.Fatal(err)
	}

	_, err = GetOrCreateUser(db, "User Two")
	if err != nil {
		t.Fatal(err)
	}

	gotNames := getAutocompleteNames(db)
	wantNames := []string{"User One", "User Two"}
	if len(gotNames) != len(wantNames) {
		t.Fatalf("gotNames and wantNames must have the same length.")
	}
	for i := range wantNames {
		if gotNames[i] != wantNames[i] {
			t.Fatalf("gotNames and wantNames must be equal")
		}
	}
}
