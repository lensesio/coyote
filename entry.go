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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	shellwords "github.com/mattn/go-shellwords"
)

type (
	// Entry contains the test case's entries' structure, see examples for more.
	Entry struct {
		Name        string        `yaml:"name"`
		Description string        `yaml:"description,omitempty"`
		WorkDir     string        `yaml:"workdir"`
		Command     string        `yaml:"command,omitempty"`
		Stdin       string        `yaml:"stdin,omitempty"`
		NoLog       bool          `yaml:"nolog,omitempty"`
		EnvVars     []string      `yaml:"env,omitempty"`
		Timeout     time.Duration `yaml:"timeout,omitempty"`

		// It differs from the `Timeout`,
		// `SleepBefore` will wait for 'x' duration before the execution of this command.
		SleepBefore time.Duration `yaml:"sleep_before,omitempty"`
		// It differs from the `Timeout`,
		// `SleepAfter` will wait for 'x' duration after the execution of this command.
		SleepAfter time.Duration `yaml:"sleep_after,omitempty"`

		// keep for backwards compatibility.
		StdoutExpect    []string `yaml:"stdout_has,omitempty"`
		StdoutNotExpect []string `yaml:"stdout_not_has,omitempty"`
		StderrExpect    []string `yaml:"stderr_has,omitempty"`
		StderrNotExpect []string `yaml:"stderr_not_has,omitempty"`
		// NoRegex if true disables the regex matching which is the default behavior for "stdout_has", "stdout_not_has", "stderr_has", "stderr_not_has".
		// Useful for matching [raw array results]).
		NoRegex bool `yaml:"noregex,omitempty"`

		Stdout OutFilters `yaml:"stdout,omitempty"`
		Stderr OutFilters `yaml:"stderr,omitempty"`

		IgnoreExitCode bool `yaml:"ignore_exit_code,omitempty"`

		// Skip will Skip only if "true".
		// It's type of string instead of bool because it is meant to help with manipulating tests from scripts.
		//
		// See `NoSkip` too.
		Skip string `yaml:"skip,omitempty"`
		// NoSkip will skip if it is set and not "true".
		// It's type of string instead of bool because it is meant to help with manipulating tests from scripts.
		//
		// See `Skip` too.
		NoSkip string `yaml:"noskip,omitempty"`
	}

	// OutFilter describes the stdout and stderr output's search expectation.
	//
	// See `Entry` for more.
	OutFilter struct {
		// Match should match (against regex expression if NoRegex is false, default behavior).
		Match []string `yaml:"match,omitempty"`
		// NotMatch should not match (against regex expression if NoRegex is false, default behavior).
		NotMatch []string `yaml:"not_match,omitempty"`

		/* More options below... */

		// NoRegex if true disables the regex matching, which is the default behavior.
		// Useful for matching [raw array results]).
		NoRegex bool `yaml:"noregex,omitempty"`
		// Partial if true then it passes the test if at least one of the Match/NotMatch entries and their content
		// exist in the command's output.
		// Essentialy is a small helper, it can be done with regex as well.
		Partial bool `yaml:"partial,omitempty"`
	}

	// OutFilters is a set of `OutFilter`.
	OutFilters []OutFilter
)

// GetMatches simply returns the text of all `Match` inside "filters".
func (filters OutFilters) GetMatches() (matches []string) {
	for _, filter := range filters {
		matches = append(matches, filter.Match...)
	}

	return
}

// GetNotMatches simply returns the text of all `NotMatch` inside "filters".
func (filters OutFilters) GetNotMatches() (notMatches []string) {
	for _, filter := range filters {
		notMatches = append(notMatches, filter.NotMatch...)
	}

	return
}

func removeNewLine(s string) string {
	return strings.TrimRightFunc(s, func(c rune) bool {
		return c == '\r' || c == '\n'
	})
}

func canPassAgainstBackwards(against, output string, noregex bool) (bool, error) {
	if noregex {
		against, output = removeNewLine(against), removeNewLine(output)
		return output == against, nil
	}
	return regexp.MatchString(against, output)
}

func canPassAgainst(against, output string, f OutFilter) (bool, error) {
	if f.NoRegex {
		against, output = removeNewLine(against), removeNewLine(output)
		if f.Partial {
			return strings.Contains(output, against), nil
		}
		return output == against, nil
	}

	if f.Partial {
		return strings.Contains(output, against), nil
	}

	return regexp.MatchString(against, output)
}

// key -> the position of the test case for both stdout and stderr.
// value -> the error(s) produced by each of them.
type filterErrors map[int][]string

var newLineB = []byte("\n")

func (errs filterErrors) String() string {
	b := new(strings.Builder)
	if len(errs) == 0 {
		return ""
	}

	for _, errors := range errs {
		for _, errMsg := range errors {
			if errMsg != "" {
				b.WriteString(errMsg)
				b.Write(newLineB)
			}
		}
	}

	return b.String()
}

func (f OutFilter) check(output string) (bool, error) {
	matchErrors, notMatchErrors := make(filterErrors), make(filterErrors)

	for i, v := range f.Match {
		if v == "" {
			continue
		}

		// check for match.
		pass, errPass := canPassAgainst(v, output, f)

		if errPass != nil {
			matchErrors[i] = append(matchErrors[i], fmt.Sprintf("match: bad regexp: %v.", errPass))
		}

		if !pass {
			errMsg := fmt.Sprintf("match: should expected '%s'.", v)
			if output == "" {
				errMsg += " Output is empty ''."
			}
			matchErrors[i] = append(matchErrors[i], errMsg)
		} else if f.Partial { // we passed at least one case, break.
			// and delete any previous errors for THIS `match` entry.
			for j := 0; j < i; j++ {
				delete(matchErrors, j)
			}
			break
		}
	}

	for i, v := range f.NotMatch {
		if v == "" {
			continue
		}

		// check for not match (too).
		pass, errPass := canPassAgainst(v, output, f)
		if errPass != nil {
			notMatchErrors[i] = append(notMatchErrors[i], fmt.Sprintf("not_match: bad regexp: %v.", errPass))
		}

		if pass {
			notMatchErrors[i] = append(notMatchErrors[i], fmt.Sprintf("not_match: should not expected '%s'.", v))
		} else if errPass == nil {
			pass = true    // we can ignore it because we only check for errMsg != "", it's here for readability.
			if f.Partial { // we passed at least one case, break.
				// and delete any previous errors for THIS `match` entry.
				for j := 0; j < i; j++ {
					delete(notMatchErrors, j)
				}
				break
			}
		}
	}

	if errMsg := matchErrors.String() + notMatchErrors.String(); errMsg != "" {
		return false, errors.New(errMsg)
	}

	return true, nil
}

func (e *Entry) testBackwards(stdout, stderr string) (bool, error) {
	var errMsg string

	for _, v := range e.StdoutExpect {
		if v == "" {
			continue
		}
		toMatch := replaceUnique(v)
		pass, errPass := canPassAgainstBackwards(toMatch, stdout, e.NoRegex)
		if errPass != nil {
			errMsg = fmt.Sprintf("%sStdout_has Bad Regexp: %v. \n", errMsg, errPass)
		}

		if !pass {
			errMsg = fmt.Sprintf("%sStdout_has not matched expected '%s'.\n", errMsg, toMatch)
		}
	}

	for _, v := range e.StdoutNotExpect {
		if v == "" {
			continue
		}
		toMatch := replaceUnique(v)
		pass, errPass := canPassAgainstBackwards(toMatch, stdout, e.NoRegex)
		if errPass != nil {
			errMsg = fmt.Sprintf("%sStdout_not_has Bad Regexp: %v. \n", errMsg, errPass)
		}

		if pass {
			errMsg = fmt.Sprintf("%sStdout_not_has matched not expected '%s'.\n", errMsg, toMatch)
		} else if errPass == nil {
			pass = true // pass the test.
		}
	}

	// if errMsg != "" {
	// 	errMsg += fmt.Sprintf("Output was: '%s'", stdout)
	// }

	for _, v := range e.StderrExpect {
		if v == "" {
			continue
		}
		toMatch := replaceUnique(v)
		pass, errPass := canPassAgainstBackwards(toMatch, stderr, e.NoRegex)
		if errPass != nil {
			errMsg = fmt.Sprintf("%sStderr_has Bad Regexp: %v. \n", errMsg, errPass)
		}

		if !pass {
			errMsg = fmt.Sprintf("%sStderr_has not matched expected '%s'.\n", errMsg, toMatch)
		}
	}

	for _, v := range e.StderrNotExpect {
		if v == "" {
			continue
		}
		toMatch := replaceUnique(v)
		pass, errPass := canPassAgainstBackwards(toMatch, stderr, e.NoRegex)
		if errPass != nil {
			errMsg = fmt.Sprintf("%sStderr_not_has Bad Regexp: %v. \n", errMsg, errPass)
		}

		if pass {
			errMsg = fmt.Sprintf("%sStderr_not_has matched not expected '%s'.\n", errMsg, toMatch)
		} else if errPass == nil {
			pass = true // we can ignore it because we only check for errMsg != "", it's here for readability.
		}
	}

	if errMsg != "" {
		return false, errors.New(errMsg)
	}

	return true, nil
}

func mapVars(localVars, globalVars map[string]string, lists ...*[]string) {
	for _, items := range lists {
		tmp := *items
		for i, item := range tmp {
			result := replaceVars(replaceUnique(item), localVars, globalVars)
			tmp[i] = result
		}

		*items = tmp
	}
}

// MapVars maps the local and global vars to the name, command, stdin, env_vars and (not) expected stdout and stderr.
func (e *Entry) MapVars(localVars, globalVars map[string]string) { // note that local vars have priority over global vars.
	// If unique strings are asked, replace the placeholders
	// Also replace local and global vars.
	e.Name = replaceVars(e.Name, localVars, globalVars)
	e.Command = replaceVars(replaceUnique(e.Command), localVars, globalVars)
	e.Stdin = replaceVars(replaceUnique(e.Stdin), localVars, globalVars)
	mapVars(localVars, globalVars, &e.EnvVars)

	shouldFirstCheckForOld := len(e.StderrExpect) > 0 || len(e.StderrNotExpect) > 0
	if shouldFirstCheckForOld {
		mapVars(localVars, globalVars, &e.StdoutExpect, &e.StdoutNotExpect, &e.StderrExpect, &e.StderrNotExpect)
	}

	for _, filter := range e.Stdout {
		mapVars(localVars, globalVars, &filter.Match, &filter.NotMatch)
	}

	for _, filter := range e.Stderr {
		mapVars(localVars, globalVars, &filter.Match, &filter.NotMatch)
	}

}

// Test runs the tests based on the entry's fields and returns false if failed.
// The error is empty if test passed, otherwise it contains the necessary information text that
// the caller should know about the reason of failure of the particular test.
//
// Call of `MapVars` is required if local or/and global variables declared.
func (e *Entry) Test(stdout, stderr string) (bool, error) {
	// here we can mix the old and new syntax,
	// first check if with the old syntax passed, if passed and has new syntax is there, check that as well, otherwise fail.
	shouldFirstCheckForOld := len(e.StdoutExpect)+len(e.StdoutNotExpect)+len(e.StderrExpect)+len(e.StderrNotExpect) > 0
	if shouldFirstCheckForOld {
		if _, err := e.testBackwards(stdout, stderr); err != nil {
			return false, err
		}
	}

	for i, filter := range e.Stdout {
		if _, err := filter.check(stdout); err != nil {
			err = fmt.Errorf("stdout[%d]: %s", i, err.Error())
			return false, err
		}
	}

	for i, filter := range e.Stderr {
		if _, err := filter.check(stderr); err != nil {
			err = fmt.Errorf("stderr[%d]: %s", i, err.Error())
			return false, err
		}
	}

	return true, nil
}

// TestCommand will test against the entry's command's output result.
func (e *Entry) TestCommand() (bool, error) {
	args, err := shellwords.Parse(e.Command)
	if err != nil {
		return false, err
	}

	if len(args) == 0 { // Empty command?
		return false, fmt.Errorf("test '%s' is missing the command field", e.Name)
	}

	cmd := exec.Command(args[0], args[1:]...)

	if len(e.WorkDir) > 0 {
		cmd.Dir = e.WorkDir
	}
	if len(e.Stdin) > 0 {
		cmd.Stdin = strings.NewReader(e.Stdin)
	}

	cmd.Env = os.Environ()
	if len(e.EnvVars) > 0 {
		for _, v := range e.EnvVars {
			cmd.Env = append(cmd.Env, v)
		}
	}

	cmdOut, cmdErr := new(strings.Builder), new(strings.Builder)
	cmd.Stdout, cmd.Stderr = cmdOut, cmdErr

	if err = cmd.Run(); err != nil {
		return false, err
	}

	return e.Test(cmdOut.String(), cmdErr.String())
}
