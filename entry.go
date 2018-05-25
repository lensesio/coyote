package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type (
	// Entry contains the test case's entries' structure, see examples for more.
	Entry struct {
		Name    string        `yaml:"name"`
		WorkDir string        `yaml:"workdir"`
		Command string        `yaml:"command,omitempty"`
		Stdin   string        `yaml:"stdin,omitempty"`
		NoLog   bool          `yaml:"nolog,omitempty"`
		EnvVars []string      `yaml:"env,omitempty"`
		Timeout time.Duration `yaml:"timeout,omitempty"`

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

		IgnoreExitCode bool   `yaml:"ignore_exit_code,omitempty"`
		Skip           string `yaml:"skip,omitempty"`   // Skips only if true
		NoSkip         string `yaml:"noskip,omitempty"` // Skips if it is set and not true
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
		// Contains if true then it passes the test if at least one of the Match/NotMatch entries and their content
		// exist in the command's output.
		// Essentialy is a small helper, it can be done with regex as well.
		Contains bool `yaml:"contains,omitempty"`
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
		if f.Contains {
			return strings.Contains(output, against), nil
		}
		return output == against, nil
	}

	if f.Contains {
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
			matchErrors[i] = append(matchErrors[i], fmt.Sprintf("match: should expected '%s' but got '%s'.", v, output))
		} else if f.Contains { // we passed at least one case, break.
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
			pass = true     // we can ignore it because we only check for errMsg != "", it's here for readability.
			if f.Contains { // we passed at least one case, break.
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

		pass, errPass := canPassAgainstBackwards(v, stdout, e.NoRegex)
		if errPass != nil {
			errMsg = fmt.Sprintf("Stdout_has Bad Regexp: %v. \n", errPass)
		}

		if !pass {
			errMsg = fmt.Sprintf("%sStdout_has not matched expected '%s'.\n", errMsg, v)
		}
	}

	for _, v := range e.StdoutNotExpect {
		if v == "" {
			continue
		}

		pass, errPass := canPassAgainstBackwards(v, stdout, e.NoRegex)
		if errPass != nil {
			errMsg = fmt.Sprintf("%sStdout_not_has Bad Regexp: %v. \n", errMsg, errPass)
		}

		if pass {
			errMsg = fmt.Sprintf("%sStdout_not_has not matched expected '%s'.\n", errMsg, v)
		} else if errPass == nil {
			pass = true // pass the test.
		}
	}

	for _, v := range e.StderrExpect {
		if v == "" {
			continue
		}

		pass, errPass := canPassAgainstBackwards(v, stderr, e.NoRegex)
		if errPass != nil {
			errMsg = fmt.Sprintf("%sStderr_has Bad Regexp: %v. \n", errMsg, errPass)
		}

		if !pass {
			errMsg = fmt.Sprintf("%sStderr_has not matched expected '%s'.\n", errMsg, v)
		}
	}

	for _, v := range e.StderrNotExpect {
		if v == "" {
			continue
		}

		pass, errPass := canPassAgainstBackwards(v, stderr, e.NoRegex)
		if errPass != nil {
			errMsg = fmt.Sprintf("%sStderr_has Bad Regexp: %v. \n", errMsg, errPass)
		}

		if pass {
			errMsg = fmt.Sprintf("%sStderr_has not matched expected '%s'.\n", errMsg, v)
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
	shouldFirstCheckForOld := len(e.StderrExpect) > 0 || len(e.StderrNotExpect) > 0
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
