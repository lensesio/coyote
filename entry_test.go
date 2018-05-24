package main

import (
	"testing"
)

func TestOutFilterNoRegex(t *testing.T) {
	hasAndExpected := `{"categories":{"Infrastructure":[{"id":1,"description":"License is invalid","category":"Infrastructure","enabled":true,"isAvailable":true}]}}\n`

	entry := Entry{
		Stdout: OutFilters{
			OutFilter{
				Match:   []string{hasAndExpected},
				NoRegex: true,
			},
		},
	}

	ok, err := entry.Test(hasAndExpected, "")
	if err != nil {
		if ok {
			t.Fatalf("expected to not be passed if error")
		}
		t.Fatal(err)
	}

	if !ok && err == nil {
		t.Fatalf("expected to be passed if error is nil")
	}
}

func TestOutFilterContains(t *testing.T) {
	has := `logs-broker\nnullsink\n`
	expectingOneOfThem := []string{"nullsink"}

	entry := Entry{
		Stdout: OutFilters{
			OutFilter{
				Match:    expectingOneOfThem,
				Contains: true,
			},
		},
	}

	ok, err := entry.Test(has, "")
	if err != nil {
		if ok {
			t.Fatalf("expected to not be passed if error")
		}
		t.Fatal(err)
	}

	if !ok && err == nil {
		t.Fatalf("expected to be passed if error is nil")
	}

	// this should fail because contains is per match entry.
	entry2 := Entry{
		Stdout: OutFilters{
			OutFilter{
				Match:    expectingOneOfThem,
				Contains: true,
			},
			OutFilter{
				Match: []string{"failure"},
			},
		},
	}

	if _, err := entry2.Test(has, ""); err == nil {
		t.Fatalf("expected to not be passed")
	}

}
