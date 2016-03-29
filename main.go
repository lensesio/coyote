package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	shellwords "github.com/mattn/go-shellwords"
	"gopkg.in/yaml.v2"
)

var (
	configFile = flag.String("c", "config.yml", "configuration file")
)

func init() {
	flag.Parse()
}

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	logger.Printf("Starting coyote-tester.\n")

	configData, err := ioutil.ReadFile(*configFile)
	if err != nil {
		logger.Fatalln(err)
	}

	var config []Entry
	var results []Result
	var succesful = 0
	var errors = 0

	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		logger.Fatalf("Error reading configuration file: %v", err)
	}

	for _, v := range config {
		args, err := shellwords.Parse(v.Command)
		if err != nil {
			log.Printf("Error when parsing command %s for %s.\n", v.Command, v.Name)
		}

		cmd := exec.Command(args[0], args[1:]...)
		if len(v.WorkDir) > 0 {
			cmd.Dir = v.WorkDir
		}
		if len(v.Stdin) > 0 {
			cmd.Stdin = strings.NewReader(v.Stdin)
		}

		cmdOut := &bytes.Buffer{}
		cmdErr := &bytes.Buffer{}
		cmd.Stdout = cmdOut
		cmd.Stderr = cmdErr

		start := time.Now()
		//out, err := cmd.CombinedOutput()
		err = cmd.Run()
		elapsed := time.Since(start)
		if err != nil {
			log.Printf("Error when running '%s' for %s: %s\n", v.Command, v.Name, string(cmdErr.Bytes()))

		} else {
			log.Printf("Ran command for %s succesfully. Result: %s\n", v.Name, string(cmdOut.Bytes()))
		}

		if v.NoLog == false {
			var t = Result{Name: v.Name, Command: v.Command, Stdout: string(cmdOut.Bytes()), Stderr: string(cmdErr.Bytes())}
			if err == nil {
				t.Status = "ok"
				succesful++
			} else {
				t.Status = "error"
				t.Error = err.Error()
				errors++
			}
			t.Time = elapsed.Seconds()
			results = append(results, t)

		}
	}

	j, err := json.Marshal(results)
	fmt.Println(string(j))

	if errors == 0 {
		fmt.Println("no errors")
	} else {
		fmt.Printf("errors were made: %d\n", errors)
		os.Exit(1)
	}
}

// TreatCommand takes a comma separated string and converts it to a cmd and []args,
// removing commads and whitespace in the process.
// It isn't complete (it will fail for commands containing commas) but it has a bit of
// work to catch this case, due to command being provided by a yml entry.
func TreatCommand(s string) (string, []string) {
	tokens := strings.Split(s, ",")
	command := strings.TrimSpace(tokens[0])
	var args []string
	if len(tokens) > 1 {
		for i := 1; i < len(tokens); i++ {
			args = append(args, strings.TrimSpace(tokens[i]))
		}
	}
	return command, args
}
