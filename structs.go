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

type EntryGroup struct {
	Name    string  `yaml:"name"`
	Entries []Entry `yaml:"entries"`
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
