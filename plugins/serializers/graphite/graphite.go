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

	// Convert UnixNano to Unix timestamps
	timestamp := metric.UnixNano() / 1000000000

	for field_name, value := range metric.Fields() {
		// Convert value
		value_str := fmt.Sprintf("%#v", value)
		// Write graphite metric
		var graphitePoint string
		graphitePoint = fmt.Sprintf("%s %s %d",
			s.SerializeBucketName(metric, field_name),
			value_str,
			timestamp)
		out = append(out, graphitePoint)
	}
	return out, nil
}

func (s *GraphiteSerializer) SerializeBucketName(metric telegraf.Metric, field_name string) string {
	// Get the metric name
	name := metric.Name()

	// Convert UnixNano to Unix timestamps
	tag_str := buildTags(metric)

	// Write graphite metric
	var serializedBucketName string
	if name == field_name {
		serializedBucketName = fmt.Sprintf("%s.%s",
			tag_str,
			strings.Replace(name, ".", "_", -1))
	} else {
		serializedBucketName = fmt.Sprintf("%s.%s.%s",
			tag_str,
			strings.Replace(name, ".", "_", -1),
			strings.Replace(field_name, ".", "_", -1))
	}
	if s.Prefix != "" {
		serializedBucketName = fmt.Sprintf("%s.%s", s.Prefix, serializedBucketName)
	}
	return serializedBucketName
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
