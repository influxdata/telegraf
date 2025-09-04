package config

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/influxdata/telegraf/filter"
)

const (
	selectorSeparator = ";"
)

var (
	reKey   = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	reValue = regexp.MustCompile(`^[a-z0-9\*\?]([-a-z0-9\*\?]*[a-z0-9\*\?])?$`)

	pluginLabelSelector labelSelector
)

// CheckSelectionKeyValuePairs checks the key and value of a selector or
// label pair
func CheckSelectionKeyValuePairs(k, v string) error {
	if !reKey.MatchString(k) {
		return fmt.Errorf("invalid key %q", k)
	}
	if !reValue.MatchString(v) {
		return fmt.Errorf("invalid value %q", v)
	}
	return nil
}

// SetPluginLabelSelections initializes the plugin label selector with
// the given different selection groups. Within a group selectors are
// combined via logical AND. Different selector groups are combined via OR.
func SetPluginLabelSelections(selections []string) error {
	return pluginLabelSelector.setSelections(selections)
}

type labelSelector struct {
	groups []map[string]filter.Filter
}

func (l *labelSelector) setSelections(selections []string) error {
	// Pre-allocate the groups
	l.groups = make([]map[string]filter.Filter, 0, len(selections))

	// Within each group, the selector key-value pairs are separated by selector(semi-colon).
	for _, selection := range selections {
		if err := l.addGroup(strings.Split(selection, selectorSeparator)); err != nil {
			return err
		}
	}

	return nil
}

func (l *labelSelector) addGroup(selection []string) error {
	// Skip empty selection
	// len(selection) can never be 0
	if len(selection) == 1 && selection[0] == "" {
		return nil
	}

	// Parse the key-value pairs and create the corresponding filters
	group := make(map[string]filter.Filter, len(selection))
	for _, s := range selection {
		k, v, found := strings.Cut(s, "=")
		if !found {
			return fmt.Errorf("invalid selector %q: missing equal sign", s)
		}

		k = strings.TrimSpace(k)
		if _, found := group[k]; found {
			return fmt.Errorf("duplicate selector key %q within one statement", k)
		}

		v = strings.TrimSpace(v)
		if err := CheckSelectionKeyValuePairs(k, v); err != nil {
			return fmt.Errorf("invalid selector %q: %w", s, err)
		}

		f, err := filter.Compile([]string{v})
		if err != nil {
			return fmt.Errorf("compiling filter for selector %q failed: %w", s, err)
		}
		group[k] = f
	}

	// Add the new group for logical OR combination
	l.groups = append(l.groups, group)

	return nil
}

func (l *labelSelector) matches(labels map[string]string) bool {
	// Fallback to accepting all plugins without labels or if no select
	// statement specified via command line
	if len(labels) == 0 || len(l.groups) == 0 {
		return true
	}

	// Iterate over the filter groups and combine all filters within a group via
	// logical AND and the different groups via logical OR.
	return slices.ContainsFunc(l.groups, func(group map[string]filter.Filter) bool {
		for k, f := range group {
			if label, found := labels[k]; !found || !f.Match(label) {
				return false
			}
		}
		return true
	})
}
