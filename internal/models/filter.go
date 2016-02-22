package internal_models

import (
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

// TagFilter is the name of a tag, and the values on which to filter
type TagFilter struct {
	Name   string
	Filter []string
}

// Filter containing drop/pass and tagdrop/tagpass rules
type Filter struct {
	NameDrop []string
	NamePass []string

	FieldDrop []string
	FieldPass []string

	TagDrop []TagFilter
	TagPass []TagFilter

	IsActive bool
}

func (f Filter) ShouldMetricPass(metric telegraf.Metric) bool {
	if f.ShouldNamePass(metric.Name()) && f.ShouldTagsPass(metric.Tags()) {
		return true
	}
	return false
}

// ShouldFieldsPass returns true if the metric should pass, false if should drop
// based on the drop/pass filter parameters
func (f Filter) ShouldNamePass(key string) bool {
	if f.NamePass != nil {
		for _, pat := range f.NamePass {
			// TODO remove HasPrefix check, leaving it for now for legacy support.
			// Cam, 2015-12-07
			if strings.HasPrefix(key, pat) || internal.Glob(pat, key) {
				return true
			}
		}
		return false
	}

	if f.NameDrop != nil {
		for _, pat := range f.NameDrop {
			// TODO remove HasPrefix check, leaving it for now for legacy support.
			// Cam, 2015-12-07
			if strings.HasPrefix(key, pat) || internal.Glob(pat, key) {
				return false
			}
		}

		return true
	}
	return true
}

// ShouldFieldsPass returns true if the metric should pass, false if should drop
// based on the drop/pass filter parameters
func (f Filter) ShouldFieldsPass(key string) bool {
	if f.FieldPass != nil {
		for _, pat := range f.FieldPass {
			// TODO remove HasPrefix check, leaving it for now for legacy support.
			// Cam, 2015-12-07
			if strings.HasPrefix(key, pat) || internal.Glob(pat, key) {
				return true
			}
		}
		return false
	}

	if f.FieldDrop != nil {
		for _, pat := range f.FieldDrop {
			// TODO remove HasPrefix check, leaving it for now for legacy support.
			// Cam, 2015-12-07
			if strings.HasPrefix(key, pat) || internal.Glob(pat, key) {
				return false
			}
		}

		return true
	}
	return true
}

// ShouldTagsPass returns true if the metric should pass, false if should drop
// based on the tagdrop/tagpass filter parameters
func (f Filter) ShouldTagsPass(tags map[string]string) bool {
	if f.TagPass != nil {
		for _, pat := range f.TagPass {
			if tagval, ok := tags[pat.Name]; ok {
				for _, filter := range pat.Filter {
					if internal.Glob(filter, tagval) {
						return true
					}
				}
			}
		}
		return false
	}

	if f.TagDrop != nil {
		for _, pat := range f.TagDrop {
			if tagval, ok := tags[pat.Name]; ok {
				for _, filter := range pat.Filter {
					if internal.Glob(filter, tagval) {
						return false
					}
				}
			}
		}
		return true
	}

	return true
}
