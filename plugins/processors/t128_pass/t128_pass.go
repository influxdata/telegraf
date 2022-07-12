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

	## Whether to ignore if any tag or field keys are missing.
	# ignore_missing_keys = false

  [processors.t128_pass.condition.tags]
	# tag1 = ["value1", "value2"]
	# tag2 = ["value3"]

  ## Fields work the same was a fields and can be included in the same condition.
  ## Only string values are accepted and the non-string field values in this section
  ## will be converted to strings before comparison.
  [processors.t128_pass.condition.fields.string]
	# field1 = ["value1", "value2"]
	# field2 = ["value3"]

  [[processors.t128_pass.condition]]
	# mode = "exact"

  [processors.t128_pass.condition.tags]
	# tag1 = ["value3"]
`

type leaves map[string][]string
type mode string

type valueGetter func(string, telegraf.Metric) (string, error)

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
	Mode              mode       `toml:"mode"`
	Operation         operation  `toml:"operation"`
	Invert            bool       `toml:"invert"`
	IgnoreMissingKeys bool       `toml:"ignore_missing_keys"`
	Tags              leaves     `toml:"tags"`
	Fields            fieldTypes `toml:"fields"`
}

type fieldTypes struct {
	String leaves `toml:"string"`
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

func tagGetter(key string, point telegraf.Metric) (string, error) {
	value, ok := point.GetTag(key)
	if !ok {
		return "", fmt.Errorf("unable to find tag key: %s in metric %+v", key, point)
	}

	return value, nil
}

func fieldGetter(key string, point telegraf.Metric) (string, error) {
	value, ok := point.GetField(key)
	if !ok {
		return "", fmt.Errorf("unable to find field key: %s in metric %+v", key, point)
	}

	valueStr := fmt.Sprintf("%+v", value)

	return valueStr, nil
}

type matcher interface {
	Matches(telegraf.Metric) bool
}

type baseLeafMatcher struct {
	leafKey           string
	valueGetter       valueGetter
	ignoreMissingKeys bool
}

type exactMatcher struct {
	baseLeafMatcher
	values []string
}

func (m exactMatcher) Matches(point telegraf.Metric) bool {
	value, err := m.valueGetter(m.leafKey, point)
	if err != nil {
		return m.ignoreMissingKeys
	}

	for _, expectedValue := range m.values {
		if value == expectedValue {
			return true
		}
	}

	return false
}

type regexMatcher struct {
	baseLeafMatcher
	expressions []*regexp.Regexp
}

func (m regexMatcher) Matches(point telegraf.Metric) bool {
	value, err := m.valueGetter(m.leafKey, point)
	if err != nil {
		return m.ignoreMissingKeys
	}

	for _, expression := range m.expressions {
		if expression.MatchString(value) {
			return true
		}
	}

	return false
}

type globMatcher struct {
	baseLeafMatcher
	globs []glob.Glob
}

func (m globMatcher) Matches(point telegraf.Metric) bool {
	value, err := m.valueGetter(m.leafKey, point)
	if err != nil {
		return m.ignoreMissingKeys
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
		tagMatchers, err := getLeafMatchers(condition.Tags, condition.Mode, tagGetter, condition.IgnoreMissingKeys)
		if err != nil {
			return nil, err
		}

		fieldMatchers, err := getLeafMatchers(condition.Fields.String, condition.Mode, fieldGetter, condition.IgnoreMissingKeys)
		if err != nil {
			return nil, err
		}

		conditionMatcher, err := getConditionMatcher(append(tagMatchers, fieldMatchers...), condition.Operation)
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

func getLeafMatchers(leaves leaves, mode mode, valueGetter valueGetter, ignoreMissingKeys bool) ([]matcher, error) {
	leafMatchers := make([]matcher, 0, len(leaves))

	for leafKey, leafValues := range leaves {
		leafMatcher, err := getLeafMatcher(mode, leafKey, leafValues, valueGetter, ignoreMissingKeys)
		if err != nil {
			return nil, err
		}

		leafMatchers = append(leafMatchers, leafMatcher)
	}

	return leafMatchers, nil
}

func getLeafMatcher(mode mode, leafKey string, leafValues []string, valueGetter valueGetter, ignoreMissingKeys bool) (matcher, error) {
	switch mode {
	case exactMode, emptyMode:
		return exactMatcher{baseLeafMatcher{leafKey, valueGetter, ignoreMissingKeys}, leafValues}, nil
	case regexMode:
		expressions, err := compileExpressions(leafValues)
		if err != nil {
			return nil, err
		}

		return regexMatcher{baseLeafMatcher{leafKey, valueGetter, ignoreMissingKeys}, expressions}, nil
	case globMode:
		globs, err := compileGlobs(leafValues)
		if err != nil {
			return nil, err
		}

		return globMatcher{baseLeafMatcher{leafKey, valueGetter, ignoreMissingKeys}, globs}, nil
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
