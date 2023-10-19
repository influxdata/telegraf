package filter

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/devopsext/utils"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type FilterIf struct {
	Disabled bool              `toml:"disabled"`
	Tags     map[string]string `toml:"tags"`
	tags     map[string]*regexp.Regexp
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

func (f *Filter) Apply(in ...telegraf.Metric) []telegraf.Metric {

	orAnd := f.Condition != "AND"

	for _, metric := range in {

		flag := len(f.Ifs) > 0
		for _, item := range f.Ifs {

			if item.Disabled {
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
		err := fmt.Errorf("no metrics found")
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
