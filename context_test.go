package main

import (
	"bytes"
	"runtime"
	"testing"
)

func TestConvertToEntryGroups(t *testing.T) {

	yamlContents := []byte(`
describe: Tests

constants:
  TEST_VAR: test

before:
  - command: echo "Set up everything before all"

before_each:
  - command: echo "Set up everything before each"

after_each:
  - command: echo "Tear down everything after each"

after:
  - command: echo "Tear down everything after all"

specs:
  - name: First Test should do whatever
    entries:
      - name: Stdout_not_has | must pass
        command: echo "hello world"
        stdout_not_has: ["bye" ]
      - name: Stdout_has | must pass
        command: echo "hello world"
        stdout_has: ["hello", "world"]

  - name: Second Test should do whatever
    entries:
      - name: Stdout_has | must pass
        command: echo "bye world"
        stdout_has: ["bye", "world"]`)
	if runtime.GOOS == "windows" {
		yamlContents = bytes.Replace(yamlContents, []byte("echo"), []byte("cmd /C echo"), -1)
	}

	var groups []EntryGroup
	if err := TextContextLoader(yamlContents).Load(&groups); err != nil {
		t.Fatal(err)
	}

	if len(groups) != 5 {
		t.Fatal("Context must be converted to 5 EntryGroups")
	}
}

func TestVersionsCompatibility(t *testing.T) {
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
	if err := TextContextLoader(yamlContents).Load(&groups); err != nil {
		t.Fatal(err)
	}

	if len(groups) != 1 {
		t.Fatal("Context must be converted to 1 EntryGroup")
	}
}
