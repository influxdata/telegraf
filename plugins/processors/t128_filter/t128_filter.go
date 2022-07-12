package t128_filter

import (
	"fmt"
	"regexp"

	"github.com/gobwas/glob"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
[[processors.t128_filter]]
  ## The conditions that must be met to pass a metric through. This is similar
  ## behavior to a tagpass, but the multiple tags are ANDed
  [[processors.t128_filter.condition]]
	## Mode dictates how to match the condition's tag values
	## Valid values are:
	##  * "exact": exact string comparison
	##  * "glob": go flavored glob comparison (see https://github.com/gobwas/glob)
	##  * "regex": go flavored regex comparison
	# mode = "exact"

  [processors.t128_filter.condition.tags]
	# tag1 = ["value1", "value2"]
	# tag2 = ["value3"]

  [[processors.t128_filter.condition]]
	# mode = "exact"

  [processors.t128_filter.condition.tags]
	# tag1 = ["value3"]
`

type tags map[string][]string
type mode string

const (
	emptyMode mode = ""
	exactMode mode = "exact"
	regexMode mode = "regex"
	globMode  mode = "glob"
)

type Condition struct {
	Mode mode `toml:"mode"`
	Tags tags `toml:"tags"`
}

type T128Filter struct {
	Conditions []Condition `toml:"condition"`

	log     telegraf.Logger `toml:"-"`
	matcher matcher         `toml:"-"`
}

func (r *T128Filter) SampleConfig() string {
	return sampleConfig
}

func (r *T128Filter) Description() string {
	return "Filter metrics from being emitted."
}

type matcher interface {
	Matches(telegraf.Metric) bool
}

type exactMatcher struct {
	tag    string
	values []string
}

func (m exactMatcher) Matches(point telegraf.Metric) bool {
	value, ok := point.GetTag(m.tag)
	if !ok {
		return false
	}

	for _, expectedValue := range m.values {
		if value == expectedValue {
			return true
		}
	}

	return false
}

type regexMatcher struct {
	tag         string
	expressions []*regexp.Regexp
}

func (m regexMatcher) Matches(point telegraf.Metric) bool {
	value, ok := point.GetTag(m.tag)
	if !ok {
		return false
	}

	for _, expression := range m.expressions {
		if expression.MatchString(value) {
			return true
		}
	}

	return false
}

type globMatcher struct {
	tag   string
	globs []glob.Glob
}

func (m globMatcher) Matches(point telegraf.Metric) bool {
	value, ok := point.GetTag(m.tag)
	if !ok {
		return false
	}

	for _, glob := range m.globs {
		if glob.Match(value) {
			return true
		}
	}

	return false
}

type andConjMatcher struct {
	matchers []matcher
}

func (c andConjMatcher) Matches(point telegraf.Metric) bool {
	for _, matcher := range c.matchers {
		if !matcher.Matches(point) {
			return false
		}
	}

	return true
}

type orConjMatcher struct {
	matchers []matcher
}

func (c orConjMatcher) Matches(point telegraf.Metric) bool {
	for _, matcher := range c.matchers {
		if matcher.Matches(point) {
			return true
		}
	}

	return false
}

func createMatcher(conditions []Condition) (matcher, error) {
	conditionMatchers := make([]matcher, len(conditions))
	for i, condition := range conditions {
		tagMatchers, err := getTagMatchers(condition.Tags, condition.Mode)
		if err != nil {
			return nil, err
		}

		conditionMatchers[i] = andConjMatcher{matchers: tagMatchers}
	}

	return orConjMatcher{conditionMatchers}, nil
}

func getTagMatchers(tags tags, mode mode) ([]matcher, error) {
	tagMatchers := make([]matcher, 0, len(tags))

	for tagKey, tagValues := range tags {
		tagMatcher, err := getTagMatcher(mode, tagKey, tagValues)
		if err != nil {
			return nil, err
		}

		tagMatchers = append(tagMatchers, tagMatcher)
	}

	return tagMatchers, nil
}

func getTagMatcher(mode mode, tag string, values []string) (matcher, error) {
	switch mode {
	case exactMode, emptyMode:
		return exactMatcher{tag, values}, nil
	case regexMode:
		expressions, err := compileExpressions(values)
		if err != nil {
			return nil, err
		}

		return regexMatcher{tag, expressions}, nil
	case globMode:
		globs, err := compileGlobs(values)
		if err != nil {
			return nil, err
		}

		return globMatcher{tag, globs}, nil
	}

	return nil, fmt.Errorf("invalid mode: %s", mode)
}

func compileExpressions(values []string) ([]*regexp.Regexp, error) {
	expressions := make([]*regexp.Regexp, len(values))
	for i, value := range values {
		var err error
		expressions[i], err = regexp.Compile(fmt.Sprintf("^%s$", value))
		if err != nil {
			return nil, err
		}
	}

	return expressions, nil
}

func compileGlobs(values []string) ([]glob.Glob, error) {
	globs := make([]glob.Glob, len(values))
	for i, value := range values {
		var err error
		globs[i], err = glob.Compile(value)
		if err != nil {
			return nil, err
		}
	}

	return globs, nil
}

func (r *T128Filter) Apply(in ...telegraf.Metric) []telegraf.Metric {
	filteredPoints := make([]telegraf.Metric, 0)

	for _, point := range in {
		if r.matcher.Matches(point) {
			filteredPoints = append(filteredPoints, point)
		}
	}

	return filteredPoints
}

func (r *T128Filter) Init() error {
	var err error
	r.matcher, err = createMatcher(r.Conditions)
	return err
}

func newFilter() *T128Filter {
	return &T128Filter{
		Conditions: make([]Condition, 0),
	}
}

func init() {
	processors.Add("t128_filter", func() telegraf.Processor {
		return newFilter()
	})
}
