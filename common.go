/*
Package common includes structs and common functions used by
landoop-box-agent.
*/
package main

type Entry struct {
	Name    string   `yaml:"name"`
	WorkDir string   `yaml:"workdir"`
	Command string   `yaml:"command,omitempty"`
	Stdin   string   `yaml:"stdin,omitempty"`
	Expect  string   `yaml:"expect,omitempty"`
	NoLog   bool     `yaml:"nolog",omitempty`
	EnvVars []string `yaml:"env",omitempty`
}

type Result struct {
	Name    string
	Command string
	Status  string
	Time    float64
	Stdout  []string
	Stderr  []string
	Error   string
}
