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

// Please remember that when an error occurs in the code (err != nil),
// Coyote should exit with code 255. Exit codes 1-254 are reserved for the tests.

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
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	shellwords "github.com/mattn/go-shellwords"
)

//go:generate go run template-generate/include_templates.go
//go:generate go run version-generate/main.go

const DEFAULT_TITLE = "Coyote Tests"

var (
	configFilesArray configFilesArrayFlag // This is set via flag.Var, so look into the init() function
	defaultTimeout   = flag.Duration("timeout", 5*time.Minute, "default timeout for commands (e.g 2h45m, 60s, 300ms)")
	title            = flag.String("title", DEFAULT_TITLE, "title to use for report")
	outputFile       = flag.String("out", "coyote.html", "filename to save the results under, if exists it will be overwritten")
	outputJsonFile   = flag.String("json-out", "", "filename to save the results JSON array under, if exitst it will be overwritten, if empty, will not be written")
	version          = flag.Bool("version", false, "print coyote version")
	customTemplate   = flag.String("template", "", "override internal golang template with this")
	mergeResults     = flag.Bool("merge-results", false, "merge all trailing json results into one")
	testGroups       = flag.String("run", ".*", "run tests against a particular set of entries by group name (regex). Works in converse of the inline 'skip' YAML option")
)

var (
	logger            *log.Logger
	uniqStrings       = make(map[string]string)
	uniqRegexp        = regexp.MustCompile("%UNIQUE_[0-9A-Za-z_-]+%")
	t                 *template.Template
	acceptableVarName = regexp.MustCompile("^[a-zA-Z0-9_]+$")
	globalVars        = make(map[string]string)
)

type configFilesArrayFlag []string

func (i *configFilesArrayFlag) String() string {
	return strings.Join(*i, ",")
}

func (i *configFilesArrayFlag) Set(value string) error {
	absPath, err := filepath.Abs(value)
	if err != nil {
		return err
	}

	if info, err := os.Stat(absPath); err == nil && info.IsDir() {
		// act directory as a list of yaml files, if valid directory path passed.
		if absPath[len(absPath)-1] != '/' {
			absPath += "/"
		}

		files, err := filepath.Glob(absPath + "*.yml")
		if err != nil {
			return err
		}

		*i = append(*i, files...)
		return nil
	}

	*i = append(*i, absPath)
	return nil
}

func init() {
	logger = log.New(os.Stderr, "", log.Ldate|log.Ltime)

	flag.Var(&configFilesArray, "c", "configuration file(s), may be set more than once (default \"coyote.yml\")")
	flag.Parse()
	if len(configFilesArray) == 0 {
		configFilesArray = append(configFilesArray, "coyote.yml")
	}

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

	if *mergeResults == true {
		if len(flag.Args()) == 0 {
			logger.Printf("Requested to merge results, but no results were passed.")
			os.Exit(255)
		}
		err := doMergeResults()
		if err != nil {
			log.Println(err)
			os.Exit(255)
		}
		os.Exit(0)
	}

	logger.Printf("Starting coyote-tester\n")

	// Set the available loaders to load EntryGroups from.
	var loaders = []ContextLoader{
		// from yaml file(s) configuration.
		FileContextLoader(configFilesArray),
	}

	// Load the set of `EntryGroup` based on the available `EntryLoader`s.
	var entriesGroups []EntryGroup

	// we load all of them.
	for _, loader := range loaders {
		if err := loader.Load(&entriesGroups); err != nil {
			// exit on first error.
			logger.Fatal(err)
		}
	}

	// keep only the user-defined "testGroups", if value not changed don't waste time here.
	if q := *testGroups; q != ".*" {
		testGroupsRegexp := regexp.MustCompile(q)

		for i, v := range entriesGroups {
			if !testGroupsRegexp.MatchString(v.Name) {
				logger.Printf("Skipping processing group: [ %s ]\n", v.Name)
				entriesGroups = append(entriesGroups[:i], entriesGroups[i+1:]...)
			}
		}
	}

	var resultsGroups []ResultGroup
	var passed = 0
	var errors = 0
	var totalTime = 0.0

	// Search for Coyote Groups which Contain Global Configuration
	for _, v := range entriesGroups {
		// Reserved name coyote is used to set the title and global vars
		if v.Name == "coyote" {
			if v.Title != "" {
				*title = v.Title
			}
			if len(v.Vars) != 0 {
				var err error
				globalVars, err = checkVarNames(v.Vars)
				if err != nil {
					log.Fatalln(err)
				}
			}
		}
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
			continue
		}

		// Check for Local Variables
		localVars, err := checkVarNames(v.Vars)
		if err != nil {
			log.Fatalln(err)
		}
		// Replace any variables in title
		v.Title = replaceVars(v.Title, localVars, globalVars)
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

			// If timeout is missing, set the default. If it is <0, set infinite.
			if v.Timeout == 0 {
				v.Timeout = *defaultTimeout
			} else if v.Timeout < 0 {
				v.Timeout = time.Duration(365 * 24 * time.Hour)
			}

			v.MapVars(localVars, globalVars)
			args, err := shellwords.Parse(v.Command)

			if err != nil {
				logger.Printf("Error when parsing command [ %s ] for [ %s ]\n", v.Command, v.Name)
			}

			// TODO
			// if (!execute)
			//   cmd := exec.Command("echo", args[0:]...)
			// else
			//   ...

			if v.SleepBefore > 0 {
				if !v.NoLog {
					logger.Printf("Wait for %d seconds before run the test '%s'\n", int(v.SleepBefore.Seconds()), v.Name)
				}
				time.Sleep(v.SleepBefore)
			}

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
			_, textErr := v.Test(stdout, stderr)

			if err != nil && timerLive && !v.IgnoreExitCode && textErr != nil {
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

			if v.SleepAfter > 0 {
				if !v.NoLog {
					logger.Printf("Wait for %d seconds after the test '%s' ran\n", int(v.SleepAfter.Seconds()), v.Name)
				}
				time.Sleep(v.SleepAfter)
			}
		}
		resultGroup.Total = resultGroup.Passed + resultGroup.Errors
		passed += resultGroup.Passed
		errors += resultGroup.Errors
		totalTime += resultGroup.TotalTime
		resultsGroups = append(resultsGroups, resultGroup)
	}

	data := ExportData{
		resultsGroups,
		errors,
		passed,
		errors + passed,
		totalTime,
		time.Now().UTC().Format("2006 Jan 02, Mon, 15:04 MST"),
		*title,
	}

	if err := writeResults(data); err != nil {
		log.Println(err)
		os.Exit(255)
	}

	if errors == 0 {
		logger.Println("no errors")
	} else {
		logger.Printf("errors were made: %d\n", errors)
		if errors > 254 {
			errors = 254
		}
		// If we had 253 or less errors, the error code indicates the number of errors.
		// If we had 254 or more errors, the error code is 254.
		os.Exit(errors)
	}
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

// checkVarNames verifies that variable names are within acceptable criteria
// and returns a new map where keys are enclosed within ampersands
// so we can check for %VARNAME% entries.
func checkVarNames(vars map[string]string) (map[string]string, error) {
	r := make(map[string]string)
	for k, v := range vars {
		test := acceptableVarName.MatchString(k)
		if test != true {
			return r, errors.New("Variable name '" + k + "' contains illegal characters. Only alphanumerics and underscore are permitted for var names.")
		}
		test = uniqRegexp.MatchString("%" + k + "%")
		if test == true {
			return r, errors.New("Variable name '" + k + "' matches UNIQUE keyword for autogenerated values.")
		}
		if strings.Compare(k, "UNIQUE") == 0 {
			return r, errors.New("Variable name 'UNIQUE' is reserved.")
		}
		r["%"+k+"%"] = v
	}
	return r, nil
}

// replaceVars searches and replaces local and global variables in a string
// localVars have precedence over globalVars
func replaceVars(text string, localVars map[string]string, globalVars map[string]string) string {
	for k, v := range localVars {
		text = strings.Replace(text, k, v, -1)
	}
	for k, v := range globalVars {
		text = strings.Replace(text, k, v, -1)
	}
	return text
}

// doMergeResults is called when we are asked to take old results
// in json format and merge them and it does exactly that
func doMergeResults() error {
	inputFiles := flag.Args()
	var inputData []ExportData
	for _, v := range inputFiles {
		fh, err := os.Open(v)
		if err != nil {
			return err
		}
		defer fh.Close()
		b, _ := ioutil.ReadAll(fh)
		var ed ExportData
		json.Unmarshal(b, &ed)
		inputData = append(inputData, ed)
	}
	var outData ExportData
	for k, v := range inputData {
		outData.Results = append(outData.Results, v.Results...)
		outData.Errors += v.Errors
		outData.Successful += v.Successful
		outData.TotalTests += v.TotalTests
		outData.TotalTime += v.TotalTime
		if k == 0 {
			if *title == DEFAULT_TITLE {
				outData.Title = v.Title
			} else {
				outData.Title = *title
			}
			outData.Date = v.Date
		}
	}

	err := writeResults(outData)
	return err
}

// writeResults create the htlm report file and optionally (if asked)
// a json output file
func writeResults(data ExportData) error {
	f, err := os.Create(*outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	h := &bytes.Buffer{}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.New("Coyote error when creating json.")
	}

	// Write json file if asked. We don't return error if this fails.
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
		return err
	}
	f.Write(h.Bytes())
	return nil
}
