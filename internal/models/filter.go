package models

import (
	"fmt"

	"github.com/influxdata/telegraf/filter"
)

// TagFilter is the name of a tag, and the values on which to filter
type TagFilter struct {
	Name   string
	Filter []string
	filter filter.Filter
}

// Filter containing drop/pass and tagdrop/tagpass rules
type Filter struct {
	NameDrop []string
	nameDrop filter.Filter
	NamePass []string
	namePass filter.Filter

	FieldDrop []string
	fieldDrop filter.Filter
	FieldPass []string
	fieldPass filter.Filter

	TagDrop     []TagFilter
	TagPass     []TagFilter
	TagDropAny  []TagFilter
	TagDropAll  []TagFilter
	TagPassAny  []TagFilter
	TagPassAll  []TagFilter

	TagExclude []string
	tagExclude filter.Filter
	TagInclude []string
	tagInclude filter.Filter

	isActive bool
}

// Compile all Filter lists into filter.Filter objects.
func (f *Filter) Compile() error {
	if len(f.TagDropAny) == 0 && len(f.TagDrop) > 0 {
		f.TagDropAny = f.TagDrop
	}
	if len(f.TagPassAny) == 0 && len(f.TagPass) > 0 {
		f.TagPassAny = f.TagPass
	}
	if len(f.NameDrop) == 0 &&
		len(f.NamePass) == 0 &&
		len(f.FieldDrop) == 0 &&
		len(f.FieldPass) == 0 &&
		len(f.TagInclude) == 0 &&
		len(f.TagExclude) == 0 &&
		len(f.TagPassAny) == 0 &&
		len(f.TagPassAll) == 0 &&
		len(f.TagDropAny) == 0 &&
		len(f.TagDropAll) == 0 {
		return nil
	}

	f.isActive = true
	var err error
	f.nameDrop, err = filter.Compile(f.NameDrop)
	if err != nil {
		return fmt.Errorf("Error compiling 'namedrop', %s", err)
	}
	f.namePass, err = filter.Compile(f.NamePass)
	if err != nil {
		return fmt.Errorf("Error compiling 'namepass', %s", err)
	}

	f.fieldDrop, err = filter.Compile(f.FieldDrop)
	if err != nil {
		return fmt.Errorf("Error compiling 'fielddrop', %s", err)
	}
	f.fieldPass, err = filter.Compile(f.FieldPass)
	if err != nil {
		return fmt.Errorf("Error compiling 'fieldpass', %s", err)
	}

	f.tagExclude, err = filter.Compile(f.TagExclude)
	if err != nil {
		return fmt.Errorf("Error compiling 'tagexclude', %s", err)
	}
	f.tagInclude, err = filter.Compile(f.TagInclude)
	if err != nil {
		return fmt.Errorf("Error compiling 'taginclude', %s", err)
	}

	for i, _ := range f.TagDropAny {
		f.TagDropAny[i].filter, err = filter.Compile(f.TagDropAny[i].Filter)
		if err != nil {
			return fmt.Errorf("Error compiling 'tagdrop (any)', %s", err)
		}
	}
	for i, _ := range f.TagPassAny {
		f.TagPassAny[i].filter, err = filter.Compile(f.TagPassAny[i].Filter)
		if err != nil {
			return fmt.Errorf("Error compiling 'tagpass (any)', %s", err)
		}
	}

	for i, _ := range f.TagDropAll {
		f.TagDropAll[i].filter, err = filter.Compile(f.TagDropAll[i].Filter)
		if err != nil {
			return fmt.Errorf("Error compiling 'tagdrop (all)', %s", err)
		}
	}
	for i, _ := range f.TagPassAll {
		f.TagPassAll[i].filter, err = filter.Compile(f.TagPassAll[i].Filter)
		if err != nil {
			return fmt.Errorf("Error compiling 'tagpass (all)', %s", err)
		}
	}
	return nil
}

// Apply applies the filter to the given measurement name, fields map, and
// tags map. It will return false if the metric should be "filtered out", and
// true if the metric should "pass".
// It will modify tags & fields in-place if they need to be deleted.
func (f *Filter) Apply(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
) bool {
	if !f.isActive {
		return true
	}

	// check if the measurement name should pass
	if !f.shouldNamePass(measurement) {
		return false
	}

	// check if the tags should pass
	if !f.shouldTagsPass(tags) {
		return false
	}

	// filter fields
	for fieldkey, _ := range fields {
		if !f.shouldFieldPass(fieldkey) {
			delete(fields, fieldkey)
		}
	}
	if len(fields) == 0 {
		return false
	}

	// filter tags
	f.filterTags(tags)

	return true
}

func (f *Filter) IsActive() bool {
	return f.isActive
}

// shouldNamePass returns true if the metric should pass, false if should drop
// based on the drop/pass filter parameters
func (f *Filter) shouldNamePass(key string) bool {
	if f.namePass != nil {
		if f.namePass.Match(key) {
			return true
		}
		return false
	}

	if f.nameDrop != nil {
		if f.nameDrop.Match(key) {
			return false
		}
	}
	return true
}

// shouldFieldPass returns true if the metric should pass, false if should drop
// based on the drop/pass filter parameters
func (f *Filter) shouldFieldPass(key string) bool {
	if f.fieldPass != nil {
		if f.fieldPass.Match(key) {
			return true
		}
		return false
	}

	if f.fieldDrop != nil {
		if f.fieldDrop.Match(key) {
			return false
		}
	}
	return true
}

// shouldTagsPass returns true if the metric should pass, false if should drop
// based on the tagdrop/tagpass filter parameters
func (f *Filter) shouldTagsPass(tags map[string]string) bool {
	if f.TagPassAny != nil {
		for _, pat := range f.TagPassAny {
			if pat.filter == nil {
				continue
			}
			if tagval, ok := tags[pat.Name]; ok {
				if pat.filter.Match(tagval) {
					return true
				}
			}
		}
		return false
	}

	if f.TagPassAll != nil {
		for _, pat := range f.TagPassAll {
			if pat.filter == nil {
				continue
			}
			if tagval, ok := tags[pat.Name]; ok {
				if !pat.filter.Match(tagval) {
					return false
				}
			} else {
				return false
			}
		}
		return true
	}


	if f.TagDropAny != nil {
		for _, pat := range f.TagDropAny {
			if pat.filter == nil {
				continue
			}
			if tagval, ok := tags[pat.Name]; ok {
				if pat.filter.Match(tagval) {
					return false
				}
			}
		}
		return true
	}

	if f.TagDropAll != nil {
		for _, pat := range f.TagDropAll {
			if pat.filter == nil {
				continue
			}
			if tagval, ok := tags[pat.Name]; ok {
				if !pat.filter.Match(tagval) {
					return true
				}
			} else {
				return true
			}
		}
		return false
	}

	return true
}

// Apply TagInclude and TagExclude filters.
// modifies the tags map in-place.
func (f *Filter) filterTags(tags map[string]string) {
	if f.tagInclude != nil {
		for k, _ := range tags {
			if !f.tagInclude.Match(k) {
				delete(tags, k)
			}
		}
	}

	if f.tagExclude != nil {
		for k, _ := range tags {
			if f.tagExclude.Match(k) {
				delete(tags, k)
			}
		}
	}
}
