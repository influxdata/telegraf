package models

import (
	"log"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// makemetric is used by both RunningAggregator & RunningInput
// to make metrics.
//   nameOverride: override the name of the measurement being made.
//   namePrefix:   add this prefix to each measurement name.
//   nameSuffix:   add this suffix to each measurement name.
//   pluginTags:   these are tags that are specific to this plugin.
//   daemonTags:   these are daemon-wide global tags, and get applied after pluginTags.
//   filter:       this is a filter to apply to each metric being made.
//   applyFilter:  if false, the above filter is not applied to each metric.
//                 This is used by Aggregators, because aggregators use filters
//                 on incoming metrics instead of on created metrics.
// TODO refactor this to not have such a huge func signature.
func makemetric(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	nameOverride string,
	namePrefix string,
	nameSuffix string,
	pluginTags map[string]string,
	daemonTags map[string]string,
	filter Filter,
	applyFilter bool,
	mType telegraf.ValueType,
	t time.Time,
) telegraf.Metric {
	if len(fields) == 0 || len(measurement) == 0 {
		return nil
	}
	if tags == nil {
		tags = make(map[string]string)
	}

	// Override measurement name if set
	if len(nameOverride) != 0 {
		measurement = nameOverride
	}
	// Apply measurement prefix and suffix if set
	if len(namePrefix) != 0 {
		measurement = namePrefix + measurement
	}
	if len(nameSuffix) != 0 {
		measurement = measurement + nameSuffix
	}

	// Apply plugin-wide tags if set
	for k, v := range pluginTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}
	// Apply daemon-wide tags if set
	for k, v := range daemonTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	// Apply the metric filter(s)
	// for aggregators, the filter does not get applied when the metric is made.
	// instead, the filter is applied to metric incoming into the plugin.
	//   ie, it gets applied in the RunningAggregator.Apply function.
	if applyFilter {
		if ok := filter.Apply(measurement, fields, tags); !ok {
			return nil
		}
	}

	m, err := metric.New(measurement, tags, fields, t, mType)
	if err != nil {
		log.Printf("Error adding point [%s]: %s\n", measurement, err.Error())
		return nil
	}

	return m
}
