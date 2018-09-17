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

	TagDrop []TagFilter
	TagPass []TagFilter

	TagExclude []string
	tagExclude filter.Filter
	TagInclude []string
	tagInclude filter.Filter

	isActive bool
}

// Compile all Filter lists into filter.Filter objects.
func (f *Filter) Compile() error {
	if len(f.NameDrop) == 0 &&
		len(f.NamePass) == 0 &&
		len(f.FieldDrop) == 0 &&
		len(f.FieldPass) == 0 &&
		len(f.TagInclude) == 0 &&
		len(f.TagExclude) == 0 &&
		len(f.TagPass) == 0 &&
		len(f.TagDrop) == 0 {
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

	for i, _ := range f.TagDrop {
		f.TagDrop[i].filter, err = filter.Compile(f.TagDrop[i].Filter)
		if err != nil {
			return fmt.Errorf("Error compiling 'tagdrop', %s", err)
		}
	}
	for i, _ := range f.TagPass {
		f.TagPass[i].filter, err = filter.Compile(f.TagPass[i].Filter)
		if err != nil {
			return fmt.Errorf("Error compiling 'tagpass', %s", err)
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

// IsActive checking if filter is active
func (f *Filter) IsActive() bool {
	return f.isActive
}

// shouldNamePass returns true if the metric should pass, false if should drop
// based on the drop/pass filter parameters
func (f *Filter) shouldNamePass(key string) bool {

	pass := func(f *Filter) bool {
		if f.namePass.Match(key) {
			return true
		}
		return false
	}

	drop := func(f *Filter) bool {
		if f.nameDrop.Match(key) {
			return false
		}
		return true
	}

	if f.namePass != nil && f.nameDrop != nil {
		return pass(f) && drop(f)
	} else if f.namePass != nil {
		return pass(f)
	} else if f.nameDrop != nil {
		return drop(f)
	}

	return true
}

// shouldFieldPass returns true if the metric should pass, false if should drop
// based on the drop/pass filter parameters
func (f *Filter) shouldFieldPass(key string) bool {

	pass := func(f *Filter) bool {
		if f.fieldPass.Match(key) {
			return true
		}
		return false
	}

	drop := func(f *Filter) bool {
		if f.fieldDrop.Match(key) {
			return false
		}
		return true
	}

	if f.fieldPass != nil && f.fieldDrop != nil {
		return pass(f) && drop(f)
	} else if f.fieldPass != nil {
		return pass(f)
	} else if f.fieldDrop != nil {
		return drop(f)
	}

	return true
}

// shouldTagsPass returns true if the metric should pass, false if should drop
// based on the tagdrop/tagpass filter parameters
func (f *Filter) shouldTagsPass(tags map[string]string) bool {

	pass := func(f *Filter) bool {
		for _, pat := range f.TagPass {
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

	drop := func(f *Filter) bool {
		for _, pat := range f.TagDrop {
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

	// Add additional logic in case where both parameters are set.
	// see: https://github.com/influxdata/telegraf/issues/2860
	if f.TagPass != nil && f.TagDrop != nil {
		// return true only in case when tag pass and won't be dropped (true, true).
		// in case when the same tag should be passed and dropped it will be dropped (true, false).
		return pass(f) && drop(f)
	} else if f.TagPass != nil {
		return pass(f)
	} else if f.TagDrop != nil {
		return drop(f)
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
