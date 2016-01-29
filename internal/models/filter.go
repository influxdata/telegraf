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
	Drop []string
	Pass []string

	TagDrop []TagFilter
	TagPass []TagFilter

	IsActive bool
}

func (f Filter) ShouldMetricPass(metric telegraf.Metric) bool {
	if f.ShouldPass(metric.Name()) && f.ShouldTagsPass(metric.Tags()) {
		return true
	}
	return false
}

// ShouldPass returns true if the metric should pass, false if should drop
// based on the drop/pass filter parameters
func (f Filter) ShouldPass(key string) bool {
	if f.Pass != nil {
		for _, pat := range f.Pass {
			// TODO remove HasPrefix check, leaving it for now for legacy support.
			// Cam, 2015-12-07
			if strings.HasPrefix(key, pat) || internal.Glob(pat, key) {
				return true
			}
		}
		return false
	}

	if f.Drop != nil {
		for _, pat := range f.Drop {
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
