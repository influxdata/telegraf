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
	Ifs    []*FilterIf       `toml:"if"`
	Fields []string          `toml:"fields,omitempty"`
	Tags   map[string]string `toml:"tags,omitempty"`
	Log    telegraf.Logger   `toml:"-"`
	rAll   *regexp.Regexp
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

func (f *Filter) ifMinMax(item *FilterIf, fields map[string]interface{}, name string) bool {

	if item.Min == nil && item.Max == nil {
		return true
	}
	value := fields[name]
	if value == nil {
		return true
	}
	v, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
	if err != nil {
		return false
	}
	if item.Min != nil && v < item.min {
		return false
	}
	if item.Max != nil && v > item.max {
		return false
	}
	return true
}

func (f *Filter) ifTags(item *FilterIf, tags map[string]string) bool {

	if len(tags) == 0 {
		return false
	}

	flag := true
	for k, v := range item.Tags {

		exists := false
		value := tags[k]
		r := item.tags[k]
		if r != nil {
			exists = r.MatchString(value)
		} else {
			exists = value == v
		}

		flag = flag && exists
		if !flag {
			return false
		}
	}
	return flag
}

func (f *Filter) findIfs(measurement, field string) []*FilterIf {

	var r []*FilterIf
	for _, item := range f.Ifs {

		if item.Disabled {
			continue
		}

		if item.measurement != nil && !item.measurement.MatchString(measurement) {
			continue
		}

		if item.field != nil {
			if !item.field.MatchString(field) {
				continue
			}
		}
		r = append(r, item)
	}
	return r
}

func (f *Filter) validate(fif *FilterIf, fields map[string]interface{}, tags map[string]string, name string) bool {

	valid := f.ifMinMax(fif, fields, name)
	if valid {
		valid = f.ifTags(fif, tags)
	}
	return valid
}

func (f *Filter) Apply(metrics ...telegraf.Metric) []telegraf.Metric {

	var only []telegraf.Metric

	for _, metric := range metrics {

		fields := metric.Fields()
		tags := metric.Tags()
		valids := []string{}

		for k := range fields {

			if !utils.Contains(f.Fields, k) {
				continue
			}

			ifs := f.findIfs(metric.Name(), k)
			if len(ifs) > 0 {
				for _, item := range ifs {
					if f.validate(item, fields, tags, k) {
						valids = append(valids, k)
						break
					}
				}
			}
		}

		if len(valids) == 0 {

			only = append(only, metric)

		} else if len(valids) == len(fields) {

			for key, value := range f.Tags {
				metric.AddTag(key, value)
			}
			only = append(only, metric)

		} else {

			m := metric.Copy()

			for k := range fields {
				for _, k1 := range valids {
					if k == k1 {
						metric.RemoveField(k)
					}
				}
			}
			if len(metric.FieldList()) > 0 {
				only = append(only, metric)
			} else {
				metric.Drop()
			}

			for k := range fields {
				for _, k1 := range valids {
					if k != k1 {
						m.RemoveField(k)
					}
				}
			}
			if len(m.FieldList()) > 0 {
				for key, value := range f.Tags {
					m.AddTag(key, value)
				}
				only = append(only, m)
			} else {
				m.Drop()
			}
		}
	}
	return only
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

	if len(f.Ifs) == 0 {
		err := fmt.Errorf("no ifs found")
		f.Log.Error(err)
		return err
	}

	f.rAll = regexp.MustCompile(".*")
	f.setTags()

	return nil
}

func init() {
	processors.Add(pluginName, func() telegraf.Processor {
		return &Filter{}
	})
}
