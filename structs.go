// Copyright 2016-2018 Landoop LTD
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import "time"

type Entry struct {
	Name            string        `yaml:"name"`
	WorkDir         string        `yaml:"workdir"`
	Command         string        `yaml:"command,omitempty"`
	Stdin           string        `yaml:"stdin,omitempty"`
	NoLog           bool          `yaml:"nolog,omitempty"`
	EnvVars         []string      `yaml:"env,omitempty"`
	Timeout         time.Duration `yaml:"timeout,omitempty"`
	NoRegex         bool          `yaml:"noregex",omitempty` // disables regex matching for stdout_has, stdout_not_has, stderr_has and stderr_not_has (useful for matching [raw array results]).
	StdoutExpect    []string      `yaml:"stdout_has,omitempty"`
	StdoutNotExpect []string      `yaml:"stdout_not_has,omitempty"`
	StderrExpect    []string      `yaml:"stderr_has,omitempty"`
	StderrNotExpect []string      `yaml:"stderr_not_has,omitempty"`
	OnlyText        bool          `yaml:"only_text,omitempty"` // NI (Not Implemented)
	IgnoreExitCode  bool          `yaml:"ignore_exit_code,omitempty"`
	Skip            string        `yaml:"skip,omitempty"`   // Skips only if true
	NoSkip          string        `yaml:"noskip,omitempty"` // Skips if it is set and not true
}

type EntryGroup struct {
	Name    string            `yaml:"name"`
	Entries []Entry           `yaml:"entries"`
	Title   string            `yaml:"title,omitempty"`
	Skip    string            `yaml:"skip"`
	NoSkip  string            `yaml:"noskip,omitempty"`
	Type    string            `yaml:"type,omitempty"`
	Vars    map[string]string `yaml:"vars,omitempty"`
}

type Result struct {
	Name    string
	Command string
	Status  string
	Time    float64
	Stdout  []string
	Stderr  []string
	Exit    string
	Test    Entry
}

type ResultGroup struct {
	Name      string
	Type      string
	Results   []Result
	Passed    int
	Errors    int
	Total     int
	TotalTime float64
}

type ExportData struct {
	Results    []ResultGroup
	Errors     int
	Successful int
	TotalTests int
	TotalTime  float64
	Date       string
	Title      string
}
