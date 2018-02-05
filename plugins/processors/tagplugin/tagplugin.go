package tagplugin


import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type TagPlugin struct {
	ReferenceTagName	string `toml:"reference_tag_name"`
	NewTagName			string `toml:"new_tag_name"`
	NewTagValueMap		map[string][]string `toml:"new_tag_value_map"`
	FlatTagValueMap		map[string]string
	NewTagDefaultValue 	string `toml:"new_tag_default_value"`
}

var sampleConfig = `
  ## Only metrics with a name in this list will be processed by this plugin.
  ## If undefined all metrics will be processed.
  namepass = ["net"]

  ## The reference tag is the existing metric tag that is used to determine the value of the new tag.
  ## A tag will not be added to the metric if reference_tag_name is missing or empty.
  reference_tag_name = "interface"

  ## Name of the new tag given to the metric.
  ## A tag will not be added to the metric if new_tag_name is missing or empty.
  new_tag_name = "category"

  ## If the metric's value of reference_tag_name is not present in the map below,
  ## the metric will be tagged with the default_tag value.
  ## However, the metric will not receive a tag if the default_tag is missing or empty.
  new_tag_default_value = "other"

  ## The keys in this map are the values to use for the new tag, when the reference tag value matches any
  ## of the elements in the corresponding list.
  ## All keys should be strings, and all values should be lists of strings.
  ## If this map is empty or not defined, all metrics passed through this plugin will be tagged with the
  ## default_tag instead.
  ## Do not repeat values in different lists, if this happens a random of the matching keys will be used.
  ## Due to the way TOML is parsed, this map must be defined at the end of the plugin definition,
  ## otherwise subsequent plugin config options will be interpreted as part of this map.
  [processors.tagplugin.new_tag_value_map]
    management = ["en0", "en1"]
    api = ["en2"]
`

func newTagPlugin() *TagPlugin {
	return &TagPlugin{}
}

func (t *TagPlugin) SampleConfig() string {
	return sampleConfig
}

func (t *TagPlugin) Description() string {
	return "Add a new tag to metrics based on the value of an existing tag."
}

func (t *TagPlugin) Apply(in ...telegraf.Metric) []telegraf.Metric {
	if t.ReferenceTagName == "" || t.NewTagName == "" { return in }

	// Late initialization of a flat version of the new_tag_value_map defined in the config
	if t.FlatTagValueMap == nil {
		// This will only be run on the first invocation of this plugin
		t.FlatTagValueMap = map[string]string{}
		for tagName, tagValues := range t.NewTagValueMap {
			for _, tag := range tagValues {
				t.FlatTagValueMap[tag] = tagName
			}
		}
	}

	for _, metric := range in {
		newTagValue := t.FlatTagValueMap[metric.Tags()[t.ReferenceTagName]]
		if newTagValue == "" {
			if t.NewTagDefaultValue == "" { continue }
			newTagValue = t.NewTagDefaultValue
		}
		metric.AddTag(t.NewTagName, newTagValue)
	}

	return in
}

func init() {
	processors.Add("tagplugin", func() telegraf.Processor {
		return &TagPlugin{}
	})
}