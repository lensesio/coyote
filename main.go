package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	shellwords "github.com/mattn/go-shellwords"
	"gopkg.in/yaml.v2"
)

//go:generate go run template/include_templates.go

var (
	configFile     = flag.String("c", "config.yml", "configuration file")
	defaultTimeout = flag.Duration("timeout", 5*time.Minute, "default timeout for commands (e.g 2h45m, 60s, 300ms)")
	title          = flag.String("title", "Coyote Tests", "title to use for report")
	outputFile     = flag.String("out", "out.html", "filename to save the results under, if exists it will be overwritten")

//	execute        = flag.Bool("execute", true, "whether to actually execute the tests, otherwise mock them")
)

var (
	uniqStrings = make(map[string]string)
	uniqRegexp  = regexp.MustCompile("%UNIQUE_[0-9A-Za-z_-]+%")
)

func init() {
	flag.Parse()
	if *defaultTimeout == 0 {
		*defaultTimeout = time.Duration(365 * 24 * time.Hour)
	}
}

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
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

		log.Printf("Starting processing group: [ %s ]\n", v.Name)
		// For entries in group
		for _, v := range v.Entries {

			// If unique strings are asked, replace the placeholders
			v.Command = replaceUnique(v.Command)
			v.Stdin = replaceUnique(v.Stdin)
			for k2, v2 := range v.EnvVars {
				v.EnvVars[k2] = replaceUnique(v2)
			}

			// If timeout is missing, set the default. If it is <0, set infinite.
			if v.Timeout == 0 {
				v.Timeout = *defaultTimeout
			} else if v.Timeout < 0 {
				v.Timeout = time.Duration(365 * 24 * time.Hour)
			}

			args, err := shellwords.Parse(v.Command)

			if err != nil {
				log.Printf("Error when parsing command [ %s ] for [ %s ]\n", v.Command, v.Name)
			}

			// TODO
			// if (!execute)
			//   cmd := exec.Command("echo", args[0:]...)
			// else
			//   ...
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
				log.Printf("Error, command '%s', test '%s'. Error: %s, Stderr: %s\n", v.Command, v.Name, err.Error(), strconv.Quote(stderr))
			} else if err != nil && !timerLive {
				log.Printf("Timeout, command '%s', test '%s'. Error: %s, Stderr: %s\n", v.Command, v.Name, err.Error(), strconv.Quote(stderr))
			} else if textErr != nil {
				log.Printf("Output Error, command '%s', test '%s'. Error: %s, Stderr: %s\n", v.Command, v.Name, textErr.Error(), strconv.Quote(stdout))
			} else {
				log.Printf("Success, command '%s', test '%s'. Stdout: %s\n", v.Command, v.Name, strconv.Quote(stdout))
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

	rotateColorCanary := 0
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
		"rotateColorCharts": func(ext, int int) string {
			colors := []string{
				"#2383c1",
				"#64a61f",
				"#7b6788",
				"#a05c56",
				"#961919",
				"#d8d239",
				"#e98125",
				"#d0743c",
				"#635122",
				"#6ada6a",
				"#0b6197",
				"#7c9058",
				"#207f32",
				"#44b9af",
				"#bca349",
			}
			v := rotateColorCanary % (len(colors) - 1)
			rotateColorCanary++
			return colors[v]
		},
		"colorStatus": func(s string) string {
			switch s {
			case "error", "timeout":
				return "red"
			case "ok":
				// return "green" // green disabled because status ok should be expected
				return ""
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
	//t, err := template.New("").Funcs(funcMap).ParseFiles("template.html")
	t, err := template.New("output").Funcs(funcMap).Parse(mainTemplate)

	if err != nil {
		log.Println(err)
	} else {
		f, err := os.Create(*outputFile)
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

			// err = t.ExecuteTemplate(h, "template.html", data)
			err = t.Execute(h, data)
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
