//go:generate ../../../tools/readme_config_includer/generator
package classify

import (
	_ "embed"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type Classify struct {
	// Selector: which tag/field value determines the regex group to use.
	// Mutually exclusive; omit both to use DefaultRegexGroup unconditionally.
	SelectorTag   string `toml:"selector_tag"`
	SelectorField string `toml:"selector_field"`

	// Ordered regex-to-group mapping applied to the selector value.
	// Each element must contain exactly one key.
	SelectorMapping []map[string]string `toml:"selector_mapping"`

	// Regex group to fall back to when no selector mapping matches.
	DefaultRegexGroup string `toml:"default_regex_group"`

	// Match: which tag/field value is tested against category regexes.
	// Exactly one must be defined.
	MatchTag   string `toml:"match_tag"`
	MatchField string `toml:"match_field"`

	// DefaultCategory is used when no regex matches. Metrics are dropped if empty.
	DefaultCategory string `toml:"default_category"`

	// DropCategories lists categories whose matched metrics are dropped.
	// Accepts a single string or an array of strings in TOML.
	DropCategories any `toml:"drop_categories"`

	// Result: where to write the classification outcome.
	// Exactly one must be defined.
	ResultTag   string `toml:"result_tag"`
	ResultField string `toml:"result_field"`

	// MappedSelectorRegexes maps each group name to an ordered list of
	// category→regex definitions. Regex values may be a single string,
	// a multi-line string (one regex per line), or an array of strings.
	MappedSelectorRegexes map[string][]map[string]any `toml:"mapped_selector_regexes"`

	// Aggregation options. All are optional; aggregation is enabled by setting
	// AggregationPeriod together with at least one complete summary/group/selector
	// set, and then requires AggregationMeasurement.
	AggregationPeriod         config.Duration `toml:"aggregation_period"`
	AggregationMeasurement    string          `toml:"aggregation_measurement"`
	AggregationDroppedField   string          `toml:"aggregation_dropped_field"`
	AggregationTotalField     string          `toml:"aggregation_total_field"`
	AggregationSummaryTag     string          `toml:"aggregation_summary_tag"`
	AggregationSummaryValue   string          `toml:"aggregation_summary_value"`
	AggregationSummaryFields  []string        `toml:"aggregation_summary_fields"`
	AggregationGroupTag       string          `toml:"aggregation_group_tag"`
	AggregationGroupFields    []string        `toml:"aggregation_group_fields"`
	AggregationSelectorTag    string          `toml:"aggregation_selector_tag"`
	AggregationSelectorFields []string        `toml:"aggregation_selector_fields"`
	AggregationIncludesZeroes bool            `toml:"aggregation_includes_zeroes"`

	Log telegraf.Logger `toml:"-"`

	// Internal state derived from config fields during Init.
	acc              telegraf.Accumulator
	selectorRules    []selectorRule
	categoryRules    map[string][]categoryRule
	dropThisCategory map[string]bool
	aggregators      []*aggregator

	// mu guards the counts of all aggregators, which are updated by Add()
	// and drained by the aggregation goroutine.
	mu   sync.Mutex
	wg   sync.WaitGroup
	stop chan struct{}
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

// aggregationKey selects which value an aggregator bins its counts by.
type aggregationKey int

const (
	bySummary aggregationKey = iota
	byGroup
	bySelector
)

// aggregator counts classification outcomes per tag value and emits the
// configured fields once per aggregation period.
type aggregator struct {
	key    aggregationKey
	tag    string
	fields []string
	counts map[string]map[string]int // tag value → field → count
}

func (*Classify) SampleConfig() string {
	return sampleConfig
}

// Init validates config and compiles all regular expressions.
// It may be called multiple times on the same instance (e.g. in tests);
// all derived state is rebuilt each time.
func (cl *Classify) Init() error {
	known, err := cl.initClassification()
	if err != nil {
		return err
	}
	return cl.initAggregation(known)
}

// Start stores the accumulator and launches the aggregation goroutine if needed.
func (cl *Classify) Start(acc telegraf.Accumulator) error {
	cl.acc = acc
	if len(cl.aggregators) > 0 {
		cl.stop = make(chan struct{})
		cl.wg.Add(1)
		go cl.runAggregation()
	}
	return nil
}

// Add classifies one metric and either passes it downstream or drops it.
func (cl *Classify) Add(metric telegraf.Metric, _ telegraf.Accumulator) error {
	result := cl.classify(metric)
	if result.keep {
		if cl.ResultTag != "" {
			metric.AddTag(cl.ResultTag, result.category)
		} else {
			metric.AddField(cl.ResultField, result.category)
		}
		cl.acc.AddMetric(metric)
	} else {
		metric.Drop()
	}
	if len(cl.aggregators) > 0 {
		cl.count(result)
	}
	return nil
}

// Stop signals the aggregation goroutine to finish and waits for it to exit.
func (cl *Classify) Stop() {
	if cl.stop != nil {
		close(cl.stop)
		cl.wg.Wait()
		cl.stop = nil
	}
}

// outcome is the result of classifying one metric: as much as was resolved,
// for aggregation, and whether the metric is kept.
type outcome struct {
	selector, group, category string
	keep                      bool
}

// classify resolves the regex group and category for a metric.
func (cl *Classify) classify(metric telegraf.Metric) outcome {
	var selector string
	group := cl.DefaultRegexGroup
	if cl.SelectorTag != "" || cl.SelectorField != "" {
		var ok bool
		if selector, ok = cl.lookupString(metric, cl.SelectorTag, cl.SelectorField, "selector"); !ok {
			return outcome{}
		}
		group = cl.selectGroup(selector)
	}
	if group == "" {
		cl.Log.Debugf("Dropping point (selector item value %q maps to an empty regex group)", selector)
		return outcome{selector: selector}
	}

	rules, ok := cl.categoryRules[group]
	if !ok {
		// Init guarantees that a non-empty default_regex_group is a known group.
		if cl.DefaultRegexGroup == "" {
			cl.Log.Debugf("Dropping point (selector %q maps to %q, which is not a known regex group)", selector, group)
			return outcome{selector: selector}
		}
		rules = cl.categoryRules[cl.DefaultRegexGroup]
		group = cl.DefaultRegexGroup
	}

	value, ok := cl.lookupString(metric, cl.MatchTag, cl.MatchField, "match")
	if !ok {
		return outcome{selector: selector, group: group}
	}

	category := cl.DefaultCategory
	for _, rule := range rules {
		if rule.matches(value) {
			category = rule.name
			break
		}
	}
	cl.Log.Debugf("Regex group %q classified the point as %q", group, category)

	if category == "" {
		cl.Log.Debug("Dropping point (no match and default_category is not set)")
		return outcome{selector: selector, group: group}
	}
	if cl.dropThisCategory[category] {
		cl.Log.Debugf("Dropping point (category %q is in drop_categories)", category)
		return outcome{selector: selector, group: group, category: category}
	}
	return outcome{selector: selector, group: group, category: category, keep: true}
}

// selectGroup maps a selector value to a regex group; the first matching
// selector_mapping entry wins.
func (cl *Classify) selectGroup(selector string) string {
	for _, rule := range cl.selectorRules {
		if rule.re.MatchString(selector) {
			if rule.group == "*" {
				return selector
			}
			return rule.group
		}
	}
	cl.Log.Debugf("Selector item value %q does not match anything in selector_mapping", selector)
	return cl.DefaultRegexGroup
}

// lookupString reads a string value from the given tag, or else from the
// given field, logging why the point is dropped when it is unavailable.
func (cl *Classify) lookupString(metric telegraf.Metric, tag, field, role string) (string, bool) {
	if tag != "" {
		value, ok := metric.GetTag(tag)
		if !ok {
			cl.Log.Debugf("Dropping point (%s tag %q is missing)", role, tag)
		}
		return value, ok
	}
	raw, ok := metric.GetField(field)
	if !ok {
		cl.Log.Debugf("Dropping point (%s field %q is missing)", role, field)
		return "", false
	}
	value, ok := raw.(string)
	if !ok {
		cl.Log.Debugf("Dropping point (%s field %q is not a string)", role, field)
	}
	return value, ok
}

// count records one classification outcome in every aggregator.
func (cl *Classify) count(result outcome) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	for _, a := range cl.aggregators {
		var value string
		switch a.key {
		case bySummary:
			value = cl.AggregationSummaryValue
		case byGroup:
			value = result.group
		case bySelector:
			value = result.selector
		}
		if value == "" {
			continue
		}
		counts := a.counts[value]
		if counts == nil {
			counts = make(map[string]int)
			a.counts[value] = counts
		}
		if result.category != "" {
			counts[result.category]++
		}
		if !result.keep && cl.AggregationDroppedField != "" {
			counts[cl.AggregationDroppedField]++
		}
		if cl.AggregationTotalField != "" {
			counts[cl.AggregationTotalField]++
		}
	}
}

func (cl *Classify) initClassification() (map[string]bool, error) {
	if cl.SelectorTag != "" && cl.SelectorField != "" {
		return nil, errors.New("selector_tag and selector_field cannot both be defined")
	}
	if cl.MatchTag == "" && cl.MatchField == "" {
		return nil, errors.New("either match_tag or match_field must be defined")
	}
	if cl.MatchTag != "" && cl.MatchField != "" {
		return nil, errors.New("match_tag and match_field cannot both be defined")
	}
	if cl.ResultTag == "" && cl.ResultField == "" {
		return nil, errors.New("either result_tag or result_field must be defined")
	}
	if cl.ResultTag != "" && cl.ResultField != "" {
		return nil, errors.New("result_tag and result_field cannot both be defined")
	}

	seenSelectorRegex := make(map[string]bool)
	cl.selectorRules = make([]selectorRule, 0, len(cl.SelectorMapping))
	for _, mapping := range cl.SelectorMapping {
		if len(mapping) > 1 {
			return nil, errors.New("selector_mapping element contains more than one key")
		}
		for regex, group := range mapping {
			if regex == "" {
				return nil, fmt.Errorf("empty regex in selector_mapping for group %q", group)
			}
			if seenSelectorRegex[regex] {
				return nil, fmt.Errorf("duplicate selector_mapping regex %q", regex)
			}
			seenSelectorRegex[regex] = true
			re, err := regexp.Compile(regex)
			if err != nil {
				return nil, fmt.Errorf("invalid selector_mapping regex %q for group %q: %w", regex, group, err)
			}
			cl.selectorRules = append(cl.selectorRules, selectorRule{re: re, group: group})
		}
	}

	known := make(map[string]bool)
	cl.categoryRules = make(map[string][]categoryRule)
	for group, entries := range cl.MappedSelectorRegexes {
		seenCategory := make(map[string]bool)
		for _, entry := range entries {
			if len(entry) > 1 {
				return nil, fmt.Errorf("mapped_selector_regexes group %q element contains more than one key", group)
			}
			for category, value := range entry {
				if category == "" {
					return nil, fmt.Errorf("empty category name in mapped_selector_regexes group %q", group)
				}
				if seenCategory[category] {
					return nil, fmt.Errorf("duplicate category %q in mapped_selector_regexes group %q", category, group)
				}
				seenCategory[category] = true
				rule, err := compileCategory(category, value)
				if err != nil {
					return nil, fmt.Errorf("mapped_selector_regexes group %q category %q: %w", group, category, err)
				}
				// A category without regexes never matches, but its name stays
				// known so drop_categories and aggregation fields may refer to it.
				known[category] = true
				if len(rule.regexes) == 0 {
					cl.Log.Debugf("mapped_selector_regexes group %q category %q has no regexes", group, category)
					continue
				}
				cl.categoryRules[group] = append(cl.categoryRules[group], rule)
			}
		}
	}
	if len(cl.categoryRules) == 0 {
		return nil, errors.New("mapped_selector_regexes has no groups with category regexes defined")
	}

	if cl.DefaultRegexGroup != "" {
		if _, ok := cl.categoryRules[cl.DefaultRegexGroup]; !ok {
			return nil, fmt.Errorf("default_regex_group %q is not a group with category regexes in mapped_selector_regexes", cl.DefaultRegexGroup)
		}
	} else if cl.SelectorTag == "" && cl.SelectorField == "" {
		return nil, errors.New("default_regex_group must be set when neither selector_tag nor selector_field is defined")
	}
	for _, rule := range cl.selectorRules {
		if _, ok := cl.categoryRules[rule.group]; !ok && rule.group != "*" && rule.group != "" {
			cl.Log.Warnf("selector_mapping group %q is not a group with category regexes in mapped_selector_regexes; "+
				"its points fall back to default_regex_group or are dropped",
				rule.group)
		}
	}

	cl.dropThisCategory = make(map[string]bool)
	var dropCategories []string
	switch v := cl.DropCategories.(type) {
	case nil:
	case string:
		dropCategories = []string{v}
	default:
		var ok bool
		if dropCategories, ok = toStrings(v); !ok {
			return nil, errors.New("drop_categories must be a string or array of strings")
		}
	}
	for _, category := range dropCategories {
		if category == "" {
			return nil, errors.New("drop_categories contains an empty string")
		}
		if !known[category] && category != cl.DefaultCategory {
			return nil, fmt.Errorf("%q in drop_categories is not a known regex category or default_category", category)
		}
		cl.dropThisCategory[category] = true
	}

	return known, nil
}

// compileCategory compiles a category's regexes from a single string, a
// multi-line string (one regex per non-blank, trimmed line), or a string array.
func compileCategory(name string, value any) (categoryRule, error) {
	var patterns []string
	switch v := value.(type) {
	case string:
		if !strings.Contains(v, "\n") {
			patterns = []string{v}
			break
		}
		for line := range strings.SplitSeq(v, "\n") {
			if s := strings.TrimSpace(line); s != "" {
				patterns = append(patterns, s)
			}
		}
	default:
		var ok bool
		if patterns, ok = toStrings(v); !ok {
			return categoryRule{}, fmt.Errorf("regex value must be a string or array of strings, got %T", value)
		}
	}

	rule := categoryRule{name: name}
	for _, pattern := range patterns {
		if pattern == "" {
			return categoryRule{}, errors.New("empty regex")
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return categoryRule{}, fmt.Errorf("invalid regex %q: %w", pattern, err)
		}
		rule.regexes = append(rule.regexes, re)
	}
	return rule, nil
}

// toStrings converts a string array, which arrives from TOML as []any
// and from Go code as []string.
func toStrings(value any) ([]string, bool) {
	switch v := value.(type) {
	case []string:
		return v, true
	case []any:
		strs := make([]string, 0, len(v))
		for _, elem := range v {
			s, ok := elem.(string)
			if !ok {
				return nil, false
			}
			strs = append(strs, s)
		}
		return strs, true
	}
	return nil, false
}

func (cl *Classify) initAggregation(known map[string]bool) error {
	cl.aggregators = nil

	period := time.Duration(cl.AggregationPeriod)
	if period != 0 && period < time.Second {
		return errors.New("aggregation_period must be at least one second")
	}

	legal := maps.Clone(known)
	if cl.DefaultCategory != "" {
		legal[cl.DefaultCategory] = true
	}

	if cl.AggregationDroppedField != "" {
		if legal[cl.AggregationDroppedField] {
			return fmt.Errorf("aggregation_dropped_field %q conflicts with a regex category or default_category", cl.AggregationDroppedField)
		}
		legal[cl.AggregationDroppedField] = true
	}

	if cl.AggregationTotalField != "" {
		if legal[cl.AggregationTotalField] {
			return fmt.Errorf("aggregation_total_field %q conflicts with a regex category, default_category or aggregation_dropped_field",
				cl.AggregationTotalField)
		}
		legal[cl.AggregationTotalField] = true
	}

	if (cl.AggregationSummaryTag == "") != (cl.AggregationSummaryValue == "") {
		return errors.New("aggregation_summary_tag and aggregation_summary_value must both be set or both be empty")
	}

	sets := []struct {
		name   string
		key    aggregationKey
		tag    string
		fields []string
	}{
		{"summary", bySummary, cl.AggregationSummaryTag, cl.AggregationSummaryFields},
		{"group", byGroup, cl.AggregationGroupTag, cl.AggregationGroupFields},
		{"selector", bySelector, cl.AggregationSelectorTag, cl.AggregationSelectorFields},
	}
	var aggregators []*aggregator
	for _, set := range sets {
		for _, field := range set.fields {
			if field == "" {
				return fmt.Errorf("aggregation_%s_fields contains an empty string", set.name)
			}
			if !legal[field] {
				return fmt.Errorf("aggregation_%s_fields: %q is not a known category, default_category, "+
					"aggregation_dropped_field, or aggregation_total_field", set.name, field)
			}
		}
		if (set.tag == "") != (len(set.fields) == 0) {
			return fmt.Errorf("aggregation_%s_tag and aggregation_%s_fields must both be set or both be empty", set.name, set.name)
		}
		if set.key == bySelector && set.tag != "" && cl.SelectorTag == "" && cl.SelectorField == "" {
			return errors.New("aggregation_selector_tag requires selector_tag or selector_field")
		}
		if set.tag != "" {
			aggregators = append(aggregators, &aggregator{
				key:    set.key,
				tag:    set.tag,
				fields: set.fields,
				counts: make(map[string]map[string]int),
			})
		}
	}

	// A missing aggregation_period leaves aggregation disabled, so a fully
	// described aggregation setup can be switched off by omitting it alone.
	if period == 0 || len(aggregators) == 0 {
		return nil
	}
	if cl.AggregationMeasurement == "" {
		return errors.New("aggregation_measurement must be set when aggregation_period is set")
	}
	cl.aggregators = aggregators
	return nil
}

// runAggregation is the goroutine that periodically emits aggregation metrics.
func (cl *Classify) runAggregation() {
	defer cl.wg.Done()

	// Phase the first tick to align with natural period boundaries
	// (e.g. 00:05:00, 00:10:00 for a 5-minute period), then tick every period.
	period := time.Duration(cl.AggregationPeriod)
	now := time.Now()
	timer := time.NewTimer(now.Add(period).Truncate(period).Sub(now))
	defer timer.Stop()
	ticker := time.NewTicker(period)
	ticker.Stop()
	defer ticker.Stop()

	for {
		var t time.Time
		select {
		case t = <-timer.C:
			ticker.Reset(period)
		case t = <-ticker.C:
		case <-cl.stop:
			cl.flush(time.Now())
			return
		}
		cl.flush(t)
	}
}

// flush emits the aggregation data, recovering from a panic so the goroutine
// keeps draining the counts; otherwise they would grow without bound.
// The counts of the failed period are lost.
func (cl *Classify) flush(ts time.Time) {
	defer func() {
		if p := recover(); p != nil {
			cl.Log.Errorf("Panic emitting aggregation data: %v\n%s", p, debug.Stack())
		}
	}()
	cl.outputAggregationData(ts)
}

// outputAggregationData emits and resets the counts of every aggregator.
// The counts are swapped out under the lock and emitted after releasing it,
// so Add() is never blocked on the accumulator.
func (cl *Classify) outputAggregationData(ts time.Time) {
	drained := make([]map[string]map[string]int, len(cl.aggregators))
	cl.mu.Lock()
	for i, a := range cl.aggregators {
		drained[i] = a.counts
		a.counts = make(map[string]map[string]int)
	}
	cl.mu.Unlock()

	for i, a := range cl.aggregators {
		for value, counts := range drained[i] {
			cl.emit(a, value, counts, ts)
		}
	}
}

// emit outputs one aggregation metric with the aggregator's configured fields.
// A metric whose fields are all zero is suppressed; otherwise zero-valued
// fields are included only when aggregation_includes_zeroes is set.
func (cl *Classify) emit(a *aggregator, value string, counts map[string]int, ts time.Time) {
	fields := make(map[string]any, len(a.fields))
	haveNonzero := false
	for _, field := range a.fields {
		count := counts[field]
		if count > 0 {
			haveNonzero = true
		} else if !cl.AggregationIncludesZeroes {
			continue
		}
		fields[field] = count
	}
	if haveNonzero {
		cl.acc.AddCounter(cl.AggregationMeasurement, fields, map[string]string{a.tag: value}, ts)
	}
}

func init() {
	processors.AddStreaming("classify", func() telegraf.StreamingProcessor {
		return &Classify{}
	})
}
