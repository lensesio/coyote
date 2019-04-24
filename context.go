package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

type (
	ContextSpec struct {
		EntryGroup `yaml:",inline"`
		Before     []Entry `yaml:"before,omitempty"`
		After      []Entry `yaml:"after,omitempty"`
	}

	Context struct {
		Describe   string            `yaml:"describe"`
		Constants  map[string]string `yaml:"constants,omitempty"`
		Before     []Entry           `yaml:"before,omitempty"`
		After      []Entry           `yaml:"after,omitempty"`
		BeforeEach []Entry           `yaml:"before_each,omitempty"`
		AfterEach  []Entry           `yaml:"after_each,omitempty"`
		Specs      []ContextSpec     `yaml:"specs"`
	}
)

func (c *Context) toEntryGroup() []EntryGroup {
	var entryGroups []EntryGroup
	var mainEntryGroup EntryGroup

	mainEntryGroup.Name = "coyote"
	mainEntryGroup.Title = c.Describe
	mainEntryGroup.Vars = c.Constants

	entryGroups = append(entryGroups, mainEntryGroup)

	if len(c.Before) > 0 {
		var beforeEntryGroup EntryGroup
		beforeEntryGroup.Name = c.Describe + " | Before"
		beforeEntryGroup.Entries = c.Before
		entryGroups = append(entryGroups, beforeEntryGroup)
	}

	for _, v := range c.Specs {
		var entryGroup EntryGroup

		entryGroup.Name = c.Describe + " | " + v.Name
		entryGroup.Description = v.Description
		entryGroup.Title = v.Title
		entryGroup.Skip = v.Skip
		entryGroup.NoSkip = v.NoSkip
		entryGroup.Type = v.Type
		entryGroup.Vars = v.Vars
		entryGroup.Entries = append(c.BeforeEach, append(append(v.Before, append(v.Entries, v.After...)...), c.AfterEach...)...)

		entryGroups = append(entryGroups, entryGroup)
	}

	if len(c.After) > 0 {
		var afterEntryGroup EntryGroup
		afterEntryGroup.Name = c.Describe + " | After"
		afterEntryGroup.Entries = c.After
		entryGroups = append(entryGroups, afterEntryGroup)
	}

	return entryGroups
}

type ContextLoader interface {
	Load(groups *[]EntryGroup) error
}

// FileContextLoader is an implementation of the `ContextLoader`
// which loads a `Context` based on caller-specific yaml files.
type FileContextLoader []string

// Load updates the "context" based on the contents of the corresponding yaml files.
func (l FileContextLoader) Load(groups *[]EntryGroup) error {
	for idx, file := range l {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		if err = TextContextLoader(data).Load(groups); err != nil {
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

// TextContextLoader is an implementation of the `ContextLoader`
// which loads a `Context` based on caller-specific yaml raw text contents.
type TextContextLoader []byte

// Load updates the "context" based on the contents of the corresponding raw yaml contents.
func (l TextContextLoader) Load(groups *[]EntryGroup) error {
	var context Context
	if err := yaml.Unmarshal(l, &context); err != nil {
		return TextEntryGroupLoader(l).Load(groups)
	}

	mergeEntryGroups(groups, context.toEntryGroup())

	return nil
}
