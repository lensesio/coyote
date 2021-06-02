// Copyright 2016-2021, Lenses.io Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"runtime"
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

func TestBackwardsCompatibility(t *testing.T) {
	yamlContents := []byte(`
- name: Tests
  entries:
   - name: Stdout_has | must pass
     command: echo "hello world"
     stdout_has: ["hello", "world" ]

   - name: Stdout_has | must not pass
     command: echo "hello world"
     stdout_has: ["hello"]
     stdout_not_has: ["apple"]

   - name: Stdout_not_has | must pass
     command: echo "hello world"
     stdout_not_has: ["orange", "apple"]

   - name: Stdout_not_has | must not pass
     command: echo "hello world"
     stdout_not_has: ["apple"]`)

	if runtime.GOOS == "windows" {
		yamlContents = bytes.Replace(yamlContents, []byte("echo"), []byte("cmd /C echo"), -1)
	}

	var groups []EntryGroup
	if err := TextEntryGroupLoader(yamlContents).Load(&groups); err != nil {
		t.Fatal(err)
	}

	for i, group := range groups {
		for j, entry := range group.Entries {
			if _, err := entry.TestCommand(); err != nil {
				t.Fatalf("[%d:%d] test '%s' failed: %v", i, j, entry.Name, err)
			}
		}
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
