package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type EntryGroup struct {
	Name    string            `yaml:"name"`
	Entries []Entry           `yaml:"entries"`
	Title   string            `yaml:"title,omitempty"`
	Skip    string            `yaml:"skip"`
	NoSkip  string            `yaml:"noskip,omitempty"`
	Type    string            `yaml:"type,omitempty"`
	Vars    map[string]string `yaml:"vars,omitempty"`
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

// FileEntryLoader is an implementation of the `EntryLoader`
// which loads a set of `EntryGroup` based on caller-specific yaml files.
type FileEntryLoader []string

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
func (l FileEntryLoader) Load(groups *[]EntryGroup) error {
	for idx, file := range l {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		var newGroups []EntryGroup
		if err = yaml.Unmarshal(data, &newGroups); err != nil {
			errMsg := fmt.Sprintf("error reading configuration file(%s): %v", file, err)
			if idx > 0 {
				loadedSuc := append(l[idx-1:idx], l[idx+1:]...)
				errMsg += fmt.Sprintf(".\nLoaded: %s", strings.Join(loadedSuc, ", "))
			}

			if len(l) > idx+1 {
				errMsg += fmt.Sprintf(", remained: %s", strings.Join(l[idx+1:], ", "))
			}

			return fmt.Errorf(errMsg)
		}

		mergeEntryGroups(groups, newGroups)
	}

	return nil
}
