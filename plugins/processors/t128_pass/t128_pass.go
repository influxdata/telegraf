package t128_pass

import (
	"fmt"
	"regexp"

	"github.com/gobwas/glob"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
[[processors.t128_pass]]
  ## The conditions that must be met to pass a metric through. This is similar
  ## behavior to a tagpass, but the multiple tags are ANDed
  [[processors.t128_pass.condition]]
	## Mode dictates how to match the condition's tag values
	## Valid values are:
	##  * "exact": exact string comparison
	##  * "glob": go flavored glob comparison (see https://github.com/gobwas/glob)
	##  * "regex": go flavored regex comparison
	# mode = "exact"

	## Operation dictates how to combine the condition's tag matching
	## Valid values are:
	##  * "and": logical and the results together
	##  * "or": logical or the results together
	# operation = "and"

	## Invert dictates whether to invert the final result of the condition
	# invert = false

  [processors.t128_pass.condition.tags]
	# tag1 = ["value1", "value2"]
	# tag2 = ["value3"]

  [[processors.t128_pass.condition]]
	# mode = "exact"

  [processors.t128_pass.condition.tags]
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

type operation string

const (
	emptyOperation operation = ""
	andOperation   operation = "and"
	orOperation    operation = "or"
)

type Condition struct {
	Mode      mode      `toml:"mode"`
	Operation operation `toml:"operation"`
	Invert    bool      `toml:"invert"`
	Tags      tags      `toml:"tags"`
}

type T128Pass struct {
	Conditions []Condition `toml:"condition"`

	log     telegraf.Logger `toml:"-"`
	matcher matcher         `toml:"-"`
}

func (r *T128Pass) SampleConfig() string {
	return sampleConfig
}

func (r *T128Pass) Description() string {
	return "Passes metrics through when conditions are met."
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

type inversionMatcher struct {
	matcher matcher
}

func (m inversionMatcher) Matches(point telegraf.Metric) bool {
	return !m.matcher.Matches(point)
}

func createMatcher(conditions []Condition) (matcher, error) {
	conditionMatchers := make([]matcher, len(conditions))
	for i, condition := range conditions {
		tagMatchers, err := getTagMatchers(condition.Tags, condition.Mode)
		if err != nil {
			return nil, err
		}

		conditionMatcher, err := getConditionMatcher(tagMatchers, condition.Operation)
		if err != nil {
			return nil, err
		}

		if condition.Invert {
			conditionMatcher = inversionMatcher{matcher: conditionMatcher}
		}

		conditionMatchers[i] = conditionMatcher
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

func getConditionMatcher(matchers []matcher, operation operation) (matcher, error) {
	switch operation {
	case emptyOperation, andOperation:
		return andConjMatcher{matchers: matchers}, nil
	case orOperation:
		return orConjMatcher{matchers: matchers}, nil
	}

	return nil, fmt.Errorf("invalid operation: %s", operation)
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

func (r *T128Pass) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, point := range in {
		if !r.matcher.Matches(point) {
			// copying so that don't hit seg fault
			fields := make([]*telegraf.Field, len(point.FieldList()))
			copy(fields, point.FieldList())

			// removing all fields will have telegraf drop the metric
			for _, field := range fields {
				point.RemoveField(field.Key)
			}
		}
	}

	return in
}

func (r *T128Pass) Init() error {
	var err error
	r.matcher, err = createMatcher(r.Conditions)
	return err
}

func newPass() *T128Pass {
	return &T128Pass{
		Conditions: make([]Condition, 0),
	}
}

func init() {
	processors.Add("t128_pass", func() telegraf.Processor {
		return newPass()
	})
}
