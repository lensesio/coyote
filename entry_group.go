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

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type EntryGroup struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Entries     []Entry           `yaml:"entries"`
	Title       string            `yaml:"title,omitempty"`
	Skip        string            `yaml:"skip"`
	NoSkip      string            `yaml:"noskip,omitempty"`
	Type        string            `yaml:"type,omitempty"`
	Vars        map[string]string `yaml:"vars,omitempty"`
}

// mergeEntryGroups appends the entries of the "newGroups" to the "groups".
func mergeEntryGroups(groups *[]EntryGroup, newGroups []EntryGroup) {
	for _, newGroup := range newGroups {
		merged := false
		for i, group := range *groups {
			if newGroup.Name == group.Name {
				// join variables.
				if group.Vars == nil {
					group.Vars = newGroup.Vars
				} else if newGroup.Vars != nil {
					for varName, varValue := range newGroup.Vars {
						group.Vars[varName] = varValue
					}
				}

				// join entries.
				group.Entries = append(group.Entries, newGroup.Entries...)
				(*groups)[i] = group
				merged = true
				break
			}
		}

		if !merged {
			*groups = append(*groups, newGroup)
		}
	}
}

// EntryLoader should be implement by any structure that provides EntryGroup loading functionality.
//
// See `FileEntryLoader` for an example.
type EntryLoader interface {
	Load(groups *[]EntryGroup) error
}

// FileEntryGroupLoader is an implementation of the `EntryLoader`
// which loads a set of `EntryGroup` based on caller-specific yaml files.
type FileEntryGroupLoader []string

// AddFile adds file(s) to be loaded on `Load`.
// Note that it doesn't check for file existence, only `Load` does that.
// func (l *FileEntryLoader) AddFile(files ...string) {
// 	tmp := *l
// 	for _, file := range files {
// 		exists := false
// 		for _, existingFile := range tmp {
// 			if file == existingFile {
// 				exists = true
// 				break
// 			}
// 		}

// 		if !exists {
// 			tmp = append(tmp, file)
// 		}
// 	}

// 	*l = tmp
// }

// Load updates the "groups" based on the contents of the corresponding yaml files.
func (l FileEntryGroupLoader) Load(groups *[]EntryGroup) error {
	for idx, file := range l {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		if err = TextEntryGroupLoader(data).Load(groups); err != nil {
			errMsg := fmt.Sprintf("error reading configuration file(%s): %v", file, err)
			if idx > 0 {
				loadedSuc := l[0:idx]
				errMsg += fmt.Sprintf(".\nLoaded: %s", strings.Join(loadedSuc, ", "))
			}

			if len(l) > idx+1 {
				errMsg += fmt.Sprintf(", remained: %s", strings.Join(l[idx+1:], ", "))
			}

			return fmt.Errorf(errMsg)
		}
	}

	return nil
}

// TextEntryGroupLoader is an implementation of the `EntryLoader`
// which loads a set of `EntryGroup` based on caller-specific yaml raw text contents.
type TextEntryGroupLoader []byte

// Load updates the "groups" based on the contents of the corresponding raw yaml contents.
func (l TextEntryGroupLoader) Load(groups *[]EntryGroup) error {
	var newGroups []EntryGroup
	if err := yaml.Unmarshal(l, &newGroups); err != nil {
		return fmt.Errorf("error reading configuration contents: %v", err)
	}

	mergeEntryGroups(groups, newGroups)
	return nil
}
