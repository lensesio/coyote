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
