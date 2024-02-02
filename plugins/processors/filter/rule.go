package filter

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
)

type rule struct {
	Name   []string            `toml:"name"`
	Tags   map[string][]string `toml:"tags"`
	Fields []string            `toml:"fields"`
	Action string              `toml:"action"`

	nameFilter  filter.Filter
	fieldFilter filter.Filter
	tagFilters  map[string]filter.Filter
	pass        bool
}

func (r *rule) init() error {
	// Check the action setting
	switch r.Action {
	case "pass":
		r.pass = true
	case "", "drop":
		// Do nothing, those options are valid
	default:
		return fmt.Errorf("invalid action %q", r.Action)
	}

	// Compile the filters
	var err error
	r.nameFilter, err = filter.Compile(r.Name)
	if err != nil {
		return fmt.Errorf("creating name filter failed: %w", err)
	}

	r.fieldFilter, err = filter.Compile(r.Fields)
	if err != nil {
		return fmt.Errorf("creating fields filter failed: %w", err)
	}

	r.tagFilters = make(map[string]filter.Filter, len(r.Tags))
	for k, values := range r.Tags {
		r.tagFilters[k], err = filter.Compile(values)
		if err != nil {
			return fmt.Errorf("creating tag filter for tag %q failed: %w", k, err)
		}
	}

	return nil
}

func (r *rule) apply(m telegraf.Metric) (pass, applies bool) {
	// Check the metric name
	if r.nameFilter != nil {
		if !r.nameFilter.Match(m.Name()) {
			return true, false
		}
	}

	// Check the tags if given
	tags := m.Tags()
	for k, f := range r.tagFilters {
		if value, found := tags[k]; !found || !f.Match(value) {
			return true, false
		}
	}

	// Check the field names
	if r.fieldFilter != nil {
		var matches bool
		for _, field := range m.FieldList() {
			if r.fieldFilter.Match(field.Key) {
				matches = true
				break
			}
		}
		if !matches {
			return true, false
		}
	}

	return r.pass, true
}
