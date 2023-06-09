package models

import (
	"github.com/influxdata/telegraf"
)

// makemetric applies new metric plugin and agent measurement and tag
// settings.
func makemetric(
	metric telegraf.Metric,
	nameOverride string,
	namePrefix string,
	nameSuffix string,
	tags map[string]string,
	globalTags map[string]string,
) telegraf.Metric {
	namer := metric.Namer()
	if len(nameOverride) != 0 {
		namer.SetName(nameOverride)
	}
	if len(namePrefix) != 0 {
		namer.SetPrefix(namePrefix)
	}
	if len(nameSuffix) != 0 {
		namer.SetSuffix(nameSuffix)
	}

	// Apply plugin-wide tags
	for k, v := range tags {
		if _, ok := metric.GetTag(k); !ok {
			metric.AddTag(k, v)
		}
	}
	// Apply global tags
	for k, v := range globalTags {
		if _, ok := metric.GetTag(k); !ok {
			metric.AddTag(k, v)
		}
	}

	return metric
}
