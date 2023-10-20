package filter

import (
	_ "embed"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/devopsext/utils"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type FilterIf struct {
	Measurement string            `toml:"measurement,omitempty"`
	Field       string            `toml:"field,omitempty"`
	Min         interface{}       `toml:"min,omitempty"`
	Max         interface{}       `toml:"max,omitempty"`
	Disabled    bool              `toml:"disabled"`
	Tags        map[string]string `toml:"tags"`
	measurement *regexp.Regexp
	field       *regexp.Regexp
	min         float64
	max         float64
	tags        map[string]*regexp.Regexp
}

type Filter struct {
	Condition string            `toml:"condition"`
	Ifs       []*FilterIf       `toml:"if"`
	Tags      map[string]string `toml:"tags,omitempty"`
	Log       telegraf.Logger   `toml:"-"`
}

var description = "Advanced filtering for metrics based on tags"

const pluginName = "filter"

// Description will return a short string to explain what the plugin does.
func (*Filter) Description() string {
	return description
}

var sampleConfig = `
#
`

func (*Filter) SampleConfig() string {
	return sampleConfig
}

func (f *Filter) getKeys(arr map[string]string) []string {
	var keys []string
	for k := range arr {
		keys = append(keys, k)
	}
	return keys
}

func (f *Filter) ifCondition(item *FilterIf, metric telegraf.Metric) bool {

	tags := metric.Tags()
	if len(tags) == 0 {
		return false
	}
	keys := f.getKeys(tags)

	flag := true
	for k, r := range item.tags {

		if !utils.Contains(keys, k) {
			continue
		}

		exists := false
		value := tags[k]
		if r != nil {
			exists = r.MatchString(value)
		} else {
			exists = value == item.Tags[k]
		}

		flag = flag && exists
		if !flag {
			return false
		}
	}
	return flag
}

func (f *Filter) skipFields(item *FilterIf, metric telegraf.Metric) bool {

	if item.field == nil {
		return false
	}
	for k := range metric.Fields() {
		if !item.field.MatchString(k) {
			return true
		}
	}
	return false
}

func (f *Filter) skipMinMax(item *FilterIf, metric telegraf.Metric) bool {

	if item.Min == nil && item.Max == nil {
		return false
	}
	for _, field := range metric.Fields() {
		v, err := strconv.ParseFloat(fmt.Sprintf("%v", field), 64)
		if err != nil {
			return true
		}
		if item.Min != nil && v < item.min {
			return true
		}
		if item.Max != nil && v > item.max {
			return true
		}
	}
	return false
}

func (f *Filter) Apply(in ...telegraf.Metric) []telegraf.Metric {

	orAnd := f.Condition != "AND"

	for _, metric := range in {

		measurement := metric.Name()

		flag := len(f.Ifs) > 0
		if orAnd {
			flag = false
		}

		for _, item := range f.Ifs {

			if item.Disabled {
				continue
			}

			if item.measurement != nil && !item.measurement.MatchString(measurement) {
				continue
			}

			if f.skipFields(item, metric) {
				continue
			}

			if f.skipMinMax(item, metric) {
				continue
			}

			exists := f.ifCondition(item, metric)
			if orAnd {
				flag = flag || exists
			} else {
				flag = flag && exists
				if !flag {
					break
				}
			}
		}

		if !flag {
			continue
		}
		for key, value := range f.Tags {
			metric.AddTag(key, value)
		}
	}
	return in
}

func (f *Filter) setTags() {

	for _, item := range f.Ifs {

		if strings.TrimSpace(item.Measurement) != "" {
			item.measurement = regexp.MustCompile(item.Measurement)
		}

		if strings.TrimSpace(item.Field) != "" {
			item.field = regexp.MustCompile(item.Field)
		}

		if item.Min != nil {
			v, err := strconv.ParseFloat(fmt.Sprintf("%v", item.Min), 64)
			if err != nil {
				item.Min = nil
			} else {
				item.min = v
			}
		}

		if item.Max != nil {
			v, err := strconv.ParseFloat(fmt.Sprintf("%v", item.Max), 64)
			if err != nil {
				item.Max = nil
			} else {
				item.max = v
			}
		}

		m := make(map[string]*regexp.Regexp)
		for k, v := range item.Tags {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			m[k] = regexp.MustCompile(v)
		}
		item.tags = m
	}
}

func (f *Filter) Init() error {

	if strings.TrimSpace(f.Condition) == "" {
		f.Condition = "AND"
	}

	if len(f.Ifs) == 0 {
		err := fmt.Errorf("no ifs found")
		f.Log.Error(err)
		return err
	}
	f.setTags()

	return nil
}

func init() {
	processors.Add(pluginName, func() telegraf.Processor {
		return &Filter{}
	})
}
