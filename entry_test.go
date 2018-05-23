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
