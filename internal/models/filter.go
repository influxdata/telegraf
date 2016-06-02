package internal_models

import (
	"fmt"

	"github.com/influxdata/telegraf"
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

	IsActive bool
}

// Compile all Filter lists into filter.Filter objects.
func (f *Filter) CompileFilter() error {
	var err error
	f.nameDrop, err = filter.CompileFilter(f.NameDrop)
	if err != nil {
		return fmt.Errorf("Error compiling 'namedrop', %s", err)
	}
	f.namePass, err = filter.CompileFilter(f.NamePass)
	if err != nil {
		return fmt.Errorf("Error compiling 'namepass', %s", err)
	}

	f.fieldDrop, err = filter.CompileFilter(f.FieldDrop)
	if err != nil {
		return fmt.Errorf("Error compiling 'fielddrop', %s", err)
	}
	f.fieldPass, err = filter.CompileFilter(f.FieldPass)
	if err != nil {
		return fmt.Errorf("Error compiling 'fieldpass', %s", err)
	}

	f.tagExclude, err = filter.CompileFilter(f.TagExclude)
	if err != nil {
		return fmt.Errorf("Error compiling 'tagexclude', %s", err)
	}
	f.tagInclude, err = filter.CompileFilter(f.TagInclude)
	if err != nil {
		return fmt.Errorf("Error compiling 'taginclude', %s", err)
	}

	for i, _ := range f.TagDrop {
		f.TagDrop[i].filter, err = filter.CompileFilter(f.TagDrop[i].Filter)
		if err != nil {
			return fmt.Errorf("Error compiling 'tagdrop', %s", err)
		}
	}
	for i, _ := range f.TagPass {
		f.TagPass[i].filter, err = filter.CompileFilter(f.TagPass[i].Filter)
		if err != nil {
			return fmt.Errorf("Error compiling 'tagpass', %s", err)
		}
	}
	return nil
}

func (f *Filter) ShouldMetricPass(metric telegraf.Metric) bool {
	if f.ShouldNamePass(metric.Name()) && f.ShouldTagsPass(metric.Tags()) {
		return true
	}
	return false
}

// ShouldFieldsPass returns true if the metric should pass, false if should drop
// based on the drop/pass filter parameters
func (f *Filter) ShouldNamePass(key string) bool {
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

// ShouldFieldsPass returns true if the metric should pass, false if should drop
// based on the drop/pass filter parameters
func (f *Filter) ShouldFieldsPass(key string) bool {
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

// ShouldTagsPass returns true if the metric should pass, false if should drop
// based on the tagdrop/tagpass filter parameters
func (f *Filter) ShouldTagsPass(tags map[string]string) bool {
	if f.TagPass != nil {
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

	if f.TagDrop != nil {
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

	return true
}

// Apply TagInclude and TagExclude filters.
// modifies the tags map in-place.
func (f *Filter) FilterTags(tags map[string]string) {
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
