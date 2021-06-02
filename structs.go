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

package main

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
