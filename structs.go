package main

import "time"

type Entry struct {
	Name            string        `yaml:"name"`
	WorkDir         string        `yaml:"workdir"`
	Command         string        `yaml:"command,omitempty"`
	Stdin           string        `yaml:"stdin,omitempty"`
	NoLog           bool          `yaml:"nolog",omitempty`
	EnvVars         []string      `yaml:"env",omitempty`
	Timeout         time.Duration `yaml:"timeout,omitempty"`
	StdoutExpect    []string      `yaml:"stdout_has,omitempty"`
	StdoutNotExpect []string      `yaml:"stdout_not_has,omitempty"`
	StderrExpect    []string      `yaml:"stderr_has,omitempty"`
	StderrNotExpect []string      `yaml:"stderr_not_has,omitempty"`
	OnlyText        bool          `yaml:"only_text,omitempty"` // NI (Not Implemented)
	IgnoreExitCode  bool          `yaml:"ignore_exit_code,omitempty"`
}

type EntryGroup struct {
	Name    string  `yaml:"name"`
	Entries []Entry `yaml:"entries"`
	Title   string  `yaml:"title,omitempty"`
}

type Result struct {
	Name    string
	Command string
	Status  string
	Time    float64
	Stdout  []string
	Stderr  []string
	Exit    string
}

type ResultGroup struct {
	Name      string
	Results   []Result
	Passed    int
	Errors    int
	Total     int
	TotalTime float64
}
