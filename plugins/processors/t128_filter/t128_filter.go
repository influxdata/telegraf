package t128_filter

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
[[processors.t128_filter]]
  ## The conditions that must be met to pass a metric through. This is similar
  ## behavior to a tagpass, but the multiple tags are ANDed
  [[processors.t128_filter.condition]]

  [processors.t128_filter.condition.tags]
     #tag1 = ["value1", "value2"]
	 #tag2 = ["value3"]

  [[processors.t128_filter.condition]]

  [processors.t128_filter.condition.tags]
     #tag1 = ["value3"]
`

type tags map[string][]string

type Condition struct {
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
		tagMatchers := getTagMatchers(condition.Tags)

		conditionMatchers[i] = andConjMatcher{matchers: tagMatchers}
	}

	return orConjMatcher{conditionMatchers}, nil
}

func getTagMatchers(tags tags) []matcher {
	tagMatchers := make([]matcher, len(tags))

	j := 0
	for tagKey, tagValues := range tags {
		tagMatchers[j] = exactMatcher{tag: tagKey, values: tagValues}
		j++
	}

	return tagMatchers
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
