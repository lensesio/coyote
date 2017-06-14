// Copyright 2016-2017 Landoop LTD
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

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	shellwords "github.com/mattn/go-shellwords"
	"gopkg.in/yaml.v2"
)

//go:generate go run template-generate/include_templates.go
//go:generate go run version-generate/main.go

var (
	configFile     = flag.String("c", "coyote.yml", "configuration file")
	defaultTimeout = flag.Duration("timeout", 5*time.Minute, "default timeout for commands (e.g 2h45m, 60s, 300ms)")
	title          = flag.String("title", "Coyote Tests", "title to use for report")
	outputFile     = flag.String("out", "coyote.html", "filename to save the results under, if exists it will be overwritten")
	outputJsonFile = flag.String("json-out", "", "filename to save the results JSON array under, if exitst it will be overwritten, if empty, will not be writter")
	version        = flag.Bool("version", false, "print coyote version")
	customTemplate = flag.String("template", "", "override internal golang template with this")
)

var (
	logger      *log.Logger
	uniqStrings = make(map[string]string)
	uniqRegexp  = regexp.MustCompile("%UNIQUE_[0-9A-Za-z_-]+%")
	t           *template.Template
)

func init() {
	logger = log.New(os.Stderr, "", log.Ldate|log.Ltime)
	flag.Parse()

	if *defaultTimeout == 0 {
		*defaultTimeout = time.Duration(365 * 24 * time.Hour)
	}

	var err error
	var templateData = make([]byte, 0)
	if len(*customTemplate) == 0 {
		t, err = template.New("").Delims("<{=(", ")=}>").Parse(mainTemplate)
	} else {
		templateData, err = ioutil.ReadFile(*customTemplate)
		if err == nil {
			t, err = template.New("").Delims("<{=(", ")=}>").Parse(string(templateData))
		}
	}
	if err != nil {
		logger.Printf("Error while trying to load template: %s\n", err)
		os.Exit(255)
	}
}

func main() {
	if *version == true {
		fmt.Printf("This is Landoop's Coyote %s.\n", vgVersion)
		os.Exit(0)
	}

	logger.Printf("Starting coyote-tester\n")

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
			Type:      v.Type,
			Results:   results,
			Passed:    0,
			Errors:    0,
			Total:     0,
			TotalTime: 0.0,
		}

		// Reserved name coyote is used to set the title (and maybe other global vars in the future).
		if v.Name == "coyote" {
			*title = v.Title
			continue
		}

		// Skip test if asked
		if strings.ToLower(v.Skip) == "true" {
			logger.Printf("Skipping processing group: [ %s ]\n", v.Name)
			continue
		}
		// Don't skip test if asked
		if len(v.NoSkip) > 0 && strings.ToLower(v.NoSkip) != "true" {
			logger.Printf("Skipping processing group: [ %s ]\n", v.Name)
			continue
		}

		logger.Printf("Starting processing group: [ %s ]\n", v.Name)
		// For entries in group
		for _, v := range v.Entries {
			// Skip command if asked
			if strings.ToLower(v.Skip) == "true" {
				continue
			}
			// Don't skip command if asked
			if len(v.NoSkip) > 0 && strings.ToLower(v.NoSkip) != "true" {
				continue
			}

			// If unique strings are asked, replace the placeholders
			v.Command = replaceUnique(v.Command)
			v.Stdin = replaceUnique(v.Stdin)
			var replaceInArrays = [][]string{
				v.EnvVars,
				v.StderrExpect,
				v.StderrNotExpect,
				v.StdoutExpect,
				v.StdoutNotExpect,
			}
			for _, v2 := range replaceInArrays {
				for k3, v3 := range v2 {
					v2[k3] = replaceUnique(v3)
				}
			}

			// If timeout is missing, set the default. If it is <0, set infinite.
			if v.Timeout == 0 {
				v.Timeout = *defaultTimeout
			} else if v.Timeout < 0 {
				v.Timeout = time.Duration(365 * 24 * time.Hour)
			}

			args, err := shellwords.Parse(v.Command)

			if err != nil {
				logger.Printf("Error when parsing command [ %s ] for [ %s ]\n", v.Command, v.Name)
			}

			// TODO
			// if (!execute)
			//   cmd := exec.Command("echo", args[0:]...)
			// else
			//   ...
			var cmd *exec.Cmd
			if len(args) == 0 { // Empty command?
				logger.Printf("Entry %s is missing the command field. Aborting.\n", v.Name)
				os.Exit(255)
			} else {
				cmd = exec.Command(args[0], args[1:]...)
			}

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
			timer := time.AfterFunc(v.Timeout, func() {
				cmd.Process.Kill()
			})
			//out, err := cmd.CombinedOutput()
			err = cmd.Run()
			timerLive := timer.Stop() // If command already exited, the timer is still live.
			elapsed := time.Since(start)

			stdout := string(cmdOut.Bytes())
			stderr := string(cmdErr.Bytes())

			// Perform a textTest on outputs.
			textErr := textTest(v, stdout, stderr)

			if err != nil && timerLive && !v.IgnoreExitCode {
				logger.Printf("Error, command '%s', test '%s'. Error: %s, Stderr: %s\n", v.Command, v.Name, err.Error(), strconv.Quote(stderr))
			} else if err != nil && !timerLive {
				logger.Printf("Timeout, command '%s', test '%s'. Error: %s, Stderr: %s\n", v.Command, v.Name, err.Error(), strconv.Quote(stderr))
			} else if textErr != nil {
				logger.Printf("Output Error, command '%s', test '%s'. Error: %s, Stderr: %s\n", v.Command, v.Name, textErr.Error(), strconv.Quote(stdout))
			} else {
				logger.Printf("Success, command '%s', test '%s'. Stdout: %s\n", v.Command, v.Name, strconv.Quote(stdout))
			}

			if v.NoLog == false {
				var t = Result{Name: v.Name, Command: v.Command, Stdout: strings.Split(stdout, "\n"), Stderr: strings.Split(stderr, "\n")}

				if (err == nil || v.IgnoreExitCode) && textErr == nil {
					t.Status = "ok"
					if err != nil { // Here we have ignore_exit_code
						t.Exit = "(ignore) " + strings.Replace(err.Error(), "exit status ", "", 1)
					} else { // Here we exited normally
						t.Exit = "0"
					}
					resultGroup.Passed++
					//succesful++
				} else {
					t.Status = "error"
					if err != nil && !v.IgnoreExitCode {
						t.Exit = strings.Replace(err.Error(), "exit status ", "", 1)
					} else {
						t.Exit = "text"
						t.Stderr = append(t.Stderr, strings.Split(textErr.Error(), "\n")...)
					}
					resultGroup.Errors++
					//errors++
					if !timerLive {
						t.Status = "timeout"
						t.Exit = "(timeout) " + t.Exit
					}
				}
				t.Time = elapsed.Seconds()
				// Clean Recursively Empty Top Lines from Output
				t.Stdout = recurseClean(t.Stdout)
				t.Stderr = recurseClean(t.Stderr)

				t.Test = v
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

	if err != nil {
		logger.Println(err)
	} else {
		f, err := os.Create(*outputFile)
		defer f.Close()
		if err != nil {
			logger.Println(err)

		} else {
			h := &bytes.Buffer{}
			data := struct {
				Results    []ResultGroup
				Errors     int
				Successful int
				TotalTests int
				TotalTime  float64
				Date       string
				Title      string
			}{
				resultsGroups,
				errors,
				passed,
				errors + passed,
				totalTime,
				time.Now().UTC().Format("2006 Jan 02, Mon, 15:04 MST"),
				*title,
			}

			jsonData, err := json.Marshal(data)
			if err != nil {
				logger.Println("Coyote error when creating json.")
				os.Exit(255)
			}

			// Write json file if asked
			if *outputJsonFile != "" {
				fj, err := os.Create(*outputJsonFile)
				defer fj.Close()
				if err != nil {
					logger.Println(err)
				} else {
					fj.Write(jsonData)
				}
			}

			templateVars := struct {
				Data    template.JS
				Version template.HTML
			}{
				template.JS(jsonData),
				template.HTML("<!-- Generated by Coyote " + vgVersion + ". -->"),
			}

			// err = t.ExecuteTemplate(h, "template.html", data)
			err = t.Execute(h, templateVars)
			if err != nil {
				logger.Println(err)
			}
			f.Write(h.Bytes())

		}

	}

	if errors == 0 {
		logger.Println("no errors")
	} else {
		logger.Printf("errors were made: %d\n", errors)
		if errors > 255 {
			errors = 255
		}
		// If we had 254 or less errors, the error code indicates the number of errors.
		// If we had 255 or more errors, the error code is 255.
		os.Exit(errors)
	}
}

func textTest(t Entry, stdout, stderr string) error {
	var pass = true
	var msg = ""

	for _, v := range t.StdoutExpect {
		if len(v) > 0 {
			matched, err := regexp.MatchString(v, stdout)
			if err != nil {
				pass = false
				msg = fmt.Sprintf("%sStdout_has Bad Regexp. \n", msg)
			} else if !matched {
				pass = false
				msg = fmt.Sprintf("%sStdout_has not matched expected '%s'.\n", msg, v)
			}
		}
	}
	for _, v := range t.StdoutNotExpect {
		if len(v) > 0 {
			matched, err := regexp.MatchString(v, stdout)
			if err != nil {
				pass = false
				msg = fmt.Sprintf("%sStdout_not_has Bad Regexp.\n", msg)
			} else if matched {
				pass = false
				msg = fmt.Sprintf("%sStdout_not_has matched not expected '%s'.\n", msg, v)
			}
		}
	}
	for _, v := range t.StderrExpect {
		if len(v) > 0 {
			matched, err := regexp.MatchString(v, stderr)
			if err != nil {
				pass = false
				msg = fmt.Sprintf("%sStderr_has Bad Regexp.\n", msg)
			} else if !matched {
				pass = false
				msg = fmt.Sprintf("%sStderr_has not matched expected '%s'.\n", msg, v)
			}
		}
	}
	for _, v := range t.StderrNotExpect {
		if len(v) > 0 {
			matched, err := regexp.MatchString(v, stderr)
			if err != nil {
				pass = false
				msg = fmt.Sprintf("%sStderr_not_has Bad Regexp.\n", msg)
			} else if matched {
				pass = false
				msg = fmt.Sprintf("%sStderr_not_has matched not expected '%s'.\n", msg, v)
			}
		}
	}
	if !pass {
		return errors.New(msg)
	}
	return nil
}

// recurseClean cleans a []string from one or more empty entries at the start of the array.
func recurseClean(t []string) []string {
	if len(t) > 0 {
		if len(t[0]) == 0 {
			return recurseClean(t[1:])
		}
	}
	return t
}

// replaces instances of %UNIQUE% with a unique string based on current microsecond.
func replaceUnique(s string) (result string) {
	var contain = true
	result = s
	for {
		switch contain {
		case true:
			if strings.Contains(result, "%UNIQUE%") { // Single use unique var
				t := time.Now().UnixNano()
				t = t / 1e6 // Keep millisecond
				time.Sleep(time.Millisecond)
				uniqueText := fmt.Sprintf("%d", t)
				result = strings.Replace(result, "%UNIQUE%", uniqueText, 1)
			} else if uniqRegexp.MatchString(result) { // Multi use unique var
				stringsToReplace := uniqRegexp.FindAllString(result, -1)
				assignMultiUseUniques(stringsToReplace)
				for _, v := range stringsToReplace { // This may run more times than needed but it doesn't affect run times.
					result = strings.Replace(result, v, uniqStrings[v], -1)
				}
			} else {
				contain = false
			}
		case false:
			return
		}
	}
}

func assignMultiUseUniques(matches []string) {
	for _, v := range matches {
		if _, exists := uniqStrings[v]; !exists {
			t := time.Now().UnixNano()
			t = t / 1e6 // Keep millisecond
			time.Sleep(time.Millisecond)
			uniqueText := fmt.Sprintf("%d", t)
			uniqStrings[v] = uniqueText
		}
	}
}
