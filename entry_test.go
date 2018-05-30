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

func TestOutFilterPartial(t *testing.T) {
	has := `logs-broker\nnullsink\n`

	tests := []struct {
		entry          Entry
		shouldPass     bool
		stdout, stderr string
	}{
		{
			entry: Entry{
				Name: "test when partial is true, while match[second] does not exist",
				Stdout: OutFilters{OutFilter{
					Match:   []string{"nullsink", "not"},
					Partial: true,
				}},
			},
			shouldPass: true,
			stdout:     has,
		},
		{
			// this should fail because "Partial" is per match entry.
			entry: Entry{
				Name: "test when partial is true for the first filter with a single match but second does not exist and partial is false",
				Stdout: OutFilters{
					OutFilter{
						Match:   []string{"nullsink"},
						Partial: true,
					},
					OutFilter{
						Match: []string{"failure"},
					},
				},
			},
			shouldPass: false,
			stdout:     has,
		},
		{
			entry: Entry{
				Name: "test when partial is true but reverse order, first element does not exist but second does",
				Stdout: OutFilters{OutFilter{
					Match:   []string{"not", "logs-broker"},
					Partial: true,
				}},
			},
			shouldPass: true,
			stdout:     has,
		},
	}

	for i, tt := range tests {
		pass, err := tt.entry.Test(tt.stdout, tt.stderr)
		if tt.shouldPass != pass {
			if tt.shouldPass {
				t.Fatalf("[%d] expected to pass but failed for test '%s', error trace: %v", i, tt.entry.Name, err)
			} else {
				t.Fatalf("[%d] expected to not pass but passed for test '%s'", i, tt.entry.Name)
			}
		}
	}
}
