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
	"strconv"
	"strings"
	"text/template"
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

	// Open yml configuration
	configData, err := ioutil.ReadFile(*configFile)
	if err != nil {
		logger.Fatalln(err)
	}

	var entriesGroups []EntryGroup
	var resultsGroups []ResultGroup
	var passed = 0
	var errors = 0
	var totalTime = 0.0

	// Read yml configuration
	err = yaml.Unmarshal(configData, &entriesGroups)
	if err != nil {
		logger.Fatalf("Error reading configuration file: %v", err)
	}

	// For groups in configuration
	for _, v := range entriesGroups {
		var results []Result
		var resultGroup = ResultGroup{
			Name:      v.Name,
			Results:   results,
			Passed:    0,
			Errors:    0,
			Total:     0,
			TotalTime: 0.0,
		}

		log.Printf("Starting processing group '%s'.\n", v.Name)
		// For entries in group
		for _, v := range v.Entries {
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
			cmd.Env = os.Environ()
			if len(v.EnvVars) > 0 {
				for _, v := range v.EnvVars {
					cmd.Env = append(cmd.Env, v)
				}
			}
			cmdOut := &bytes.Buffer{}
			cmdErr := &bytes.Buffer{}
			cmd.Stdout = cmdOut
			cmd.Stderr = cmdErr

			start := time.Now()
			//out, err := cmd.CombinedOutput()
			err = cmd.Run()
			elapsed := time.Since(start)

			stdout := string(cmdOut.Bytes())
			stderr := string(cmdErr.Bytes())
			if err != nil {
				log.Printf("Error, command '%s', test '%s'. Error: %s, Stderr: %s\n", v.Command, v.Name, err.Error(), strconv.Quote(stderr))

			} else {
				log.Printf("Success, command '%s', test '%s'. Stdout: %s\n", v.Command, v.Name, strconv.Quote(stdout))
			}

			if v.NoLog == false {
				var t = Result{Name: v.Name, Command: v.Command, Stdout: strings.Split(stdout, "\n"), Stderr: strings.Split(stderr, "\n")}
				if err == nil {
					t.Status = "ok"
					t.Exit = "0"
					resultGroup.Passed++
					//succesful++
				} else {
					t.Status = "error"
					t.Exit = err.Error()
					resultGroup.Errors++
					//errors++
				}
				t.Time = elapsed.Seconds()
				resultGroup.Results = append(resultGroup.Results, t)
				resultGroup.TotalTime += t.Time
			}
		}
		resultGroup.Total = resultGroup.Passed + resultGroup.Errors
		passed += resultGroup.Passed
		errors += resultGroup.Errors
		totalTime += resultGroup.TotalTime
		resultsGroups = append(resultsGroups, resultGroup)
	}

	funcMap := template.FuncMap{
		"isEven": func(i int) bool {
			if i%2 == 0 {
				return true
			}
			return false
		},
		"showmore": func(s []string) bool {
			if len(s) > 1 {
				return true
			}
			return false
		},
		"splitString": func(s string) []string {
			if s == "" {
				return []string{""}
			}
			return strings.Split(s, "\n")
		},
		"rotateColor": func(i int) string {
			v := i % 3
			switch v {
			case 1:
				return "row header green"
			case 2:
				return "row header blue"
			case 0:
				return "row header"
			}
			return "row header"
		},
		"colorStatus": func(s string) string {
			switch s {
			case "error":
				return "red"
			case "ok":
				return "green"
			default:
				return ""
			}
			return ""
		},
		"returnFirstLine": func(s []string) string {
			r := strings.Split(s[0], "<br>")
			r = strings.Split(r[0], "\n")
			return r[0]
		},
	}
	t, err := template.New("").Funcs(funcMap).ParseFiles("template.html")
	//t, err := template.New("output").Funcs(funcMap).Parse("")

	if err != nil {
		log.Println(err)
	} else {
		f, err := os.Create("out.html")
		defer f.Close()
		if err != nil {
			log.Println(err)

		} else {
			h := &bytes.Buffer{}
			data := struct {
				Results    []ResultGroup
				Errors     int
				Successful int
				TotalTests int
				TotalTime  float64
			}{
				resultsGroups,
				errors,
				passed,
				errors + passed,
				totalTime,
			}

			err = t.ExecuteTemplate(h, "template.html", data)
			// err = t.Execute(h, Results)
			if err != nil {
				log.Println(err)
			}
			f.Write(h.Bytes())
		}

	}

	j, err := json.Marshal(resultsGroups)
	fmt.Println(string(j))
	if errors == 0 {
		fmt.Println("no errors")
	} else {
		fmt.Printf("errors were made: %d\n", errors)
		os.Exit(1)
	}
}
