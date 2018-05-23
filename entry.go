package main

import (
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

		// keep for backwards compability.
		StdoutExpect    []string `yaml:"stdout_has,omitempty"`
		StdoutNotExpect []string `yaml:"stdout_not_has,omitempty"`
		StderrExpect    []string `yaml:"stderr_has,omitempty"`
		StderrNotExpect []string `yaml:"stderr_not_has,omitempty"`
		// NoRegex if true disables the regex matching which is the default behavior for "stdout_has", "stdout_not_has", "stderr_has", "stderr_not_has".
		// Useful for matching [raw array results]).
		NoRegex bool `yaml:"noregex,omitempty"`

		Stdout []OutFilter `yaml:"stdout,omitempty"`
		Stderr []OutFilter `yaml:"stderr,omitempty"`

		IgnoreExitCode bool   `yaml:"ignore_exit_code,omitempty"`
		Skip           string `yaml:"skip,omitempty"`   // Skips only if true
		NoSkip         string `yaml:"noskip,omitempty"` // Skips if it is set and not true
	}

	// OutFilter describes the stdout and stderr output's search expectation.
	//
	// See `Entry` for more.
	OutFilter struct {
		// Match should match (against regex expression if NoRegex is false, default behavior).
		Match string `yaml:"match,omitempty"`
		// NotMatch should not match (against regex expression if NoRegex is false, default behavior).
		NotMatch string `yaml:"not_match,omitempty"`

		/* More options below... */

		// NoRegex if true disables the regex matching, which is the default behavior.
		// Useful for matching [raw array results]).
		NoRegex bool `yaml:"noregex,omitempty"`
	}
)

func canPassAgainst(output, against string, noregex bool) (bool, error) {
	if noregex {
		if strings.Contains(output, "\n") {
			output = strings.Replace(output, "\n", "", -1)
			against = strings.Replace(against, "\n", "", -1)
		}

		return output == against, nil
	}

	return regexp.MatchString(output, against)
}

func (f OutFilter) check(output string) (bool, error) {
	var errMsg string

	if v := f.Match; v != "" {
		// check for match.
		pass, errPass := canPassAgainst(v, output, f.NoRegex)
		if errPass != nil {
			errMsg = fmt.Sprintf("match: bad regexp: %v. \n", errPass)
		}

		if !pass {
			errMsg = fmt.Sprintf("%smatch: should expected '%s'.\n", errMsg, v)
		}
	}

	if v := f.NotMatch; v != "" {
		// check for not match (too).
		pass, errPass := canPassAgainst(v, output, f.NoRegex)
		if errPass != nil {
			errMsg = fmt.Sprintf("%snot_match: bad regexp: %v. \n", errMsg, errPass)
		}

		if pass {
			errMsg = fmt.Sprintf("%snot_match: should not expected '%s'.\n", errMsg, v)
		} else if errPass == nil {
			pass = true // we can ignore it because we only check for errMsg != "", it's here for readability.
		}
	}

	if errMsg != "" {
		return false, fmt.Errorf(errMsg)
	}

	return true, nil
}

func (e Entry) testBackwards(stdout, stderr string) (bool, error) {
	var errMsg string

	for _, v := range e.StdoutExpect {
		if v == "" {
			continue
		}

		pass, errPass := canPassAgainst(v, stdout, e.NoRegex)
		if errPass != nil {
			errMsg = fmt.Sprintf("Stdout_has Bad Regexp: %v. \n", errMsg, errPass)
		}

		if !pass {
			errMsg = fmt.Sprintf("%sStdout_has not matched expected '%s'.\n", errMsg, v)
		}
	}

	for _, v := range e.StdoutNotExpect {
		if v == "" {
			continue
		}

		pass, errPass := canPassAgainst(v, stdout, e.NoRegex)
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

		pass, errPass := canPassAgainst(v, stderr, e.NoRegex)
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

		pass, errPass := canPassAgainst(v, stderr, e.NoRegex)
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
		return false, fmt.Errorf(errMsg)
	}

	return true, nil
}

// Test runs the tests based on the entry's fields and returns false if failed.
// The error is empty if test passed, otherwise it contains the necessary information text that
// the caller should know about the reason of failure of the particular test.
func (e Entry) Test(stdout, stderr string) (bool, error) {
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
