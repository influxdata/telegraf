//go:generate ../../../tools/readme_config_includer/generator
package classify

import (
	_ "embed"
	"errors"
	"fmt"
	"regexp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type Classify struct {
	// Selector: which tag/field value determines the group to use.
	// Mutually exclusive; omit both to use DefaultGroup unconditionally.
	SelectorTag   string `toml:"selector_tag"`
	SelectorField string `toml:"selector_field"`

	// Ordered pattern-to-group mapping applied to the selector value.
	SelectorMapping []selectorMapping `toml:"selector_mapping"`

	// Group to fall back to when no selector mapping matches.
	DefaultGroup string `toml:"default_group"`

	// Match: which tag/field value is tested against category regexes.
	// Exactly one must be defined.
	MatchTag   string `toml:"match_tag"`
	MatchField string `toml:"match_field"`

	// DefaultCategory is used when no regex matches. Metrics pass through
	// unclassified if empty.
	DefaultCategory string `toml:"default_category"`

	// DropCategories lists categories whose matched metrics are dropped.
	DropCategories []string `toml:"drop_categories"`

	// Result: where to write the classification outcome.
	// Exactly one must be defined.
	ResultTag   string `toml:"result_tag"`
	ResultField string `toml:"result_field"`

	// Groups maps each group name to its ordered categories.
	Groups map[string][]category `toml:"groups"`

	Log telegraf.Logger `toml:"-"`

	// Internal state derived from config fields during Init.
	selectorRules    []selectorRule
	categoryRules    map[string][]categoryRule
	dropThisCategory map[string]bool
	trace            bool
}

type selectorMapping struct {
	Pattern string `toml:"pattern"`
	Group   string `toml:"group"`
}

type category struct {
	Category string   `toml:"category"`
	Patterns []string `toml:"patterns"`
}

type selectorRule struct {
	re    *regexp.Regexp
	group string
}

type categoryRule struct {
	name    string
	regexes []*regexp.Regexp
}

func (r categoryRule) matches(s string) bool {
	for _, re := range r.regexes {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

func (*Classify) SampleConfig() string {
	return sampleConfig
}

// Init validates config and compiles all regular expressions.
func (cl *Classify) Init() error {
	if cl.SelectorTag != "" && cl.SelectorField != "" {
		return errors.New("selector_tag and selector_field cannot both be defined")
	}
	hasSelector := cl.SelectorTag != "" || cl.SelectorField != ""
	if hasSelector && len(cl.SelectorMapping) == 0 {
		return errors.New("selector_mapping must be set when selector_tag or selector_field is defined")
	}
	if !hasSelector && len(cl.SelectorMapping) > 0 {
		return errors.New("selector_mapping requires selector_tag or selector_field")
	}
	if cl.MatchTag == "" && cl.MatchField == "" {
		return errors.New("either match_tag or match_field must be defined")
	}
	if cl.MatchTag != "" && cl.MatchField != "" {
		return errors.New("match_tag and match_field cannot both be defined")
	}
	if cl.ResultTag == "" && cl.ResultField == "" {
		return errors.New("either result_tag or result_field must be defined")
	}
	if cl.ResultTag != "" && cl.ResultField != "" {
		return errors.New("result_tag and result_field cannot both be defined")
	}

	known := make(map[string]bool)
	cl.categoryRules = make(map[string][]categoryRule, len(cl.Groups))
	for group, categories := range cl.Groups {
		if group == "" {
			return errors.New("empty group name in groups")
		}
		if len(categories) == 0 {
			return fmt.Errorf("group %q has no categories", group)
		}
		seen := make(map[string]bool, len(categories))
		for _, c := range categories {
			if c.Category == "" {
				return fmt.Errorf("empty category name in group %q", group)
			}
			if seen[c.Category] {
				return fmt.Errorf("duplicate category %q in group %q", c.Category, group)
			}
			seen[c.Category] = true
			if len(c.Patterns) == 0 {
				return fmt.Errorf("group %q category %q has no patterns", group, c.Category)
			}
			rule := categoryRule{name: c.Category}
			for _, pattern := range c.Patterns {
				if pattern == "" {
					return fmt.Errorf("empty pattern in group %q category %q", group, c.Category)
				}
				re, err := regexp.Compile(pattern)
				if err != nil {
					return fmt.Errorf("invalid pattern %q in group %q category %q: %w", pattern, group, c.Category, err)
				}
				rule.regexes = append(rule.regexes, re)
			}
			known[c.Category] = true
			cl.categoryRules[group] = append(cl.categoryRules[group], rule)
		}
	}
	if len(cl.categoryRules) == 0 {
		return errors.New("no groups defined")
	}

	if cl.DefaultGroup != "" {
		if _, ok := cl.categoryRules[cl.DefaultGroup]; !ok {
			return fmt.Errorf("default_group %q is not a defined group", cl.DefaultGroup)
		}
	} else if !hasSelector {
		return errors.New("default_group must be set when neither selector_tag nor selector_field is defined")
	}

	seenPattern := make(map[string]bool, len(cl.SelectorMapping))
	cl.selectorRules = make([]selectorRule, 0, len(cl.SelectorMapping))
	for _, m := range cl.SelectorMapping {
		if m.Pattern == "" {
			return fmt.Errorf("empty pattern in selector_mapping for group %q", m.Group)
		}
		if seenPattern[m.Pattern] {
			return fmt.Errorf("duplicate selector_mapping pattern %q", m.Pattern)
		}
		seenPattern[m.Pattern] = true
		// "*" passes the selector value through as the group name and ""
		// drops the metric, so only other names must be defined groups.
		if _, ok := cl.categoryRules[m.Group]; !ok && m.Group != "*" && m.Group != "" {
			return fmt.Errorf("selector_mapping group %q is not a defined group", m.Group)
		}
		re, err := regexp.Compile(m.Pattern)
		if err != nil {
			return fmt.Errorf("invalid selector_mapping pattern %q for group %q: %w", m.Pattern, m.Group, err)
		}
		cl.selectorRules = append(cl.selectorRules, selectorRule{re: re, group: m.Group})
	}

	cl.dropThisCategory = make(map[string]bool, len(cl.DropCategories))
	for _, c := range cl.DropCategories {
		if c == "" {
			return errors.New("drop_categories contains an empty string")
		}
		if !known[c] && c != cl.DefaultCategory {
			return fmt.Errorf("%q in drop_categories is not a known category or default_category", c)
		}
		cl.dropThisCategory[c] = true
	}

	cl.trace = cl.Log.Level().Includes(telegraf.Trace)
	return nil
}

// Apply classifies the metrics and returns the ones that are kept.
func (cl *Classify) Apply(in ...telegraf.Metric) []telegraf.Metric {
	out := make([]telegraf.Metric, 0, len(in))
	for _, metric := range in {
		category, drop := cl.classify(metric)
		if drop {
			metric.Drop()
			continue
		}
		if category != "" {
			if cl.ResultTag != "" {
				metric.AddTag(cl.ResultTag, category)
			} else {
				metric.AddField(cl.ResultField, category)
			}
		}
		out = append(out, metric)
	}
	return out
}

// classify resolves the category of a metric and reports whether to drop it.
// Only an empty selector_mapping group and drop_categories drop a metric;
// a metric that cannot be classified passes through with an empty category.
func (cl *Classify) classify(metric telegraf.Metric) (category string, drop bool) {
	group := cl.DefaultGroup
	if cl.SelectorTag != "" || cl.SelectorField != "" {
		selector, ok := cl.lookupString(metric, cl.SelectorTag, cl.SelectorField, "selector")
		if !ok {
			return "", false
		}
		if group, drop = cl.selectGroup(selector); drop {
			cl.tracef("Dropping point (selector item value %q maps to an empty group)", selector)
			return "", true
		}
		if group == "" {
			cl.tracef("Passing point through (selector item value %q maps to no known group)", selector)
			return "", false
		}
	}

	value, ok := cl.lookupString(metric, cl.MatchTag, cl.MatchField, "match")
	if !ok {
		return "", false
	}

	category = cl.DefaultCategory
	for _, rule := range cl.categoryRules[group] {
		if rule.matches(value) {
			category = rule.name
			break
		}
	}
	cl.tracef("Group %q classified the point as %q", group, category)

	if category == "" {
		cl.tracef("Passing point through unclassified (no match and default_category is not set)")
		return "", false
	}
	if cl.dropThisCategory[category] {
		cl.tracef("Dropping point (category %q is in drop_categories)", category)
		return "", true
	}
	return category, false
}

// selectGroup maps a selector value to a known group, or to default_group
// when no selector_mapping entry matches or a "*" entry yields an unknown
// group. The first matching entry wins, and an entry with the group ""
// reports that the metric is dropped. The returned group is empty when
// default_group is not set and no known group applies.
func (cl *Classify) selectGroup(selector string) (group string, drop bool) {
	for _, rule := range cl.selectorRules {
		if !rule.re.MatchString(selector) {
			continue
		}
		switch rule.group {
		case "":
			return "", true
		case "*":
			if _, ok := cl.categoryRules[selector]; ok {
				return selector, false
			}
			cl.tracef("Selector item value %q is not a known group", selector)
			return cl.DefaultGroup, false
		default:
			return rule.group, false
		}
	}
	cl.tracef("Selector item value %q does not match anything in selector_mapping", selector)
	return cl.DefaultGroup, false
}

// tracef logs at trace level; the check keeps formatting off the per-metric
// path when trace logging is disabled.
func (cl *Classify) tracef(format string, args ...any) {
	if cl.trace {
		cl.Log.Tracef(format, args...)
	}
}

// lookupString reads a string value from the given tag, or else from the
// given field, logging why the point passes through when it is unavailable.
func (cl *Classify) lookupString(metric telegraf.Metric, tag, field, role string) (string, bool) {
	if tag != "" {
		value, ok := metric.GetTag(tag)
		if !ok {
			cl.tracef("Passing point through (%s tag %q is missing)", role, tag)
		}
		return value, ok
	}
	raw, ok := metric.GetField(field)
	if !ok {
		cl.tracef("Passing point through (%s field %q is missing)", role, field)
		return "", false
	}
	value, ok := raw.(string)
	if !ok {
		cl.tracef("Passing point through (%s field %q is not a string)", role, field)
	}
	return value, ok
}

func init() {
	processors.Add("classify", func() telegraf.Processor {
		return &Classify{}
	})
}
