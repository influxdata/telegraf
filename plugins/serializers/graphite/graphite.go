package graphite

import (
	"fmt"
	"sort"
	"strings"

	"github.com/influxdata/telegraf"
)

type GraphiteSerializer struct {
	Prefix string
}

func (s *GraphiteSerializer) Serialize(metric telegraf.Metric) ([]string, error) {
	out := []string{}
	// Get name
	name := metric.Name()
	// Convert UnixNano to Unix timestamps
	timestamp := metric.UnixNano() / 1000000000
	tag_str := buildTags(metric)

	for field_name, value := range metric.Fields() {
		// Convert value
		value_str := fmt.Sprintf("%#v", value)
		// Write graphite metric
		var graphitePoint string
		if name == field_name {
			graphitePoint = fmt.Sprintf("%s.%s %s %d",
				tag_str,
				strings.Replace(name, ".", "_", -1),
				value_str,
				timestamp)
		} else {
			graphitePoint = fmt.Sprintf("%s.%s.%s %s %d",
				tag_str,
				strings.Replace(name, ".", "_", -1),
				strings.Replace(field_name, ".", "_", -1),
				value_str,
				timestamp)
		}
		if s.Prefix != "" {
			graphitePoint = fmt.Sprintf("%s.%s", s.Prefix, graphitePoint)
		}
		out = append(out, graphitePoint)
	}
	return out, nil
}

func buildTags(metric telegraf.Metric) string {
	var keys []string
	tags := metric.Tags()
	for k := range tags {
		if k == "host" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var tag_str string
	if host, ok := tags["host"]; ok {
		if len(keys) > 0 {
			tag_str = strings.Replace(host, ".", "_", -1) + "."
		} else {
			tag_str = strings.Replace(host, ".", "_", -1)
		}
	}

	for i, k := range keys {
		tag_value := strings.Replace(tags[k], ".", "_", -1)
		if i == 0 {
			tag_str += tag_value
		} else {
			tag_str += "." + tag_value
		}
	}
	return tag_str
}
