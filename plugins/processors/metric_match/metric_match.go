package metric_match

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

// SensorPathKey is the key for sensor path
const SensorPathKey = "sensor_path"

// TelemetryKey is the key for telemetry
const TelemetryKey = "telemetry"

// PointConfig is the delimiter for nested fields
const PointConfig = "."

const sampleConfig = ``

type MetricMatch struct {
	Tag         map[string][]string `toml:"tag"`
	FieldFilter map[string][]string `toml:"field_filter"`
	Approach    map[string]string   `toml:"approach"`
	Log         telegraf.Logger
}

func (*MetricMatch) SampleConfig() string {
	return sampleConfig
}

// Description returns the description of the processor
func (*MetricMatch) Description() string {
	return "metric match"
}

func (m *MetricMatch) Apply(in ...telegraf.Metric) []telegraf.Metric {
	// get telemetry header field_filter and tag
	res := make([]telegraf.Metric, 0, len(in))
	headerFilter := m.FieldFilter[TelemetryKey]
	headerTag := m.Tag[TelemetryKey]
	headerWay := m.Approach
	var approach string
	for _, v := range headerWay {
		approach = v
	}
	if approach == "include" {
		// include filter field
		for _, eachMetric := range in {
			sensorPath, ok := eachMetric.GetField(SensorPathKey)
			if ok {
				fieldFilters := m.FieldFilter[sensorPath.(string)]
				if len(fieldFilters) == 0 {
					m.Log.Warnf("the %s 's field filters is empty...", sensorPath)
				}
				allKeys := make([]string, 0)
				needKeys := make([]string, 0)
				for _, v := range eachMetric.FieldList() {
					if !strings.Contains(v.Key, PointConfig) {
						needKeys = append(needKeys, v.Key)
					}
					allKeys = append(allKeys, v.Key)
				}

				fieldFilters = append(fieldFilters, headerFilter...)
				for _, filter := range fieldFilters {
					if ok, matchKeys := matchField(filter, eachMetric.FieldList()); ok {
						needKeys = append(needKeys, matchKeys...)
					}
				}

				// 额外保留将要转为标签的字段（防止在 include 阶段被误删）
				tagsToKeep := m.Tag[sensorPath.(string)]
				tagsToKeep = append(tagsToKeep, headerTag...)
				for _, tagKey := range tagsToKeep {
					if ok, matchKeys := matchField(tagKey, eachMetric.FieldList()); ok {
						needKeys = append(needKeys, matchKeys...)
					}
				}

				for _, needKeysV := range needKeys {
					for k, allKeysV := range allKeys {
						if allKeysV == needKeysV {
							allKeys = append(allKeys[:k], allKeys[k+1:]...)
						}
					}
				}

				for _, v := range allKeys {
					eachMetric.RemoveField(v)
				}

			}
		}
	} else {
		// exclude filter field
		for _, eachMetric := range in {
			sensorPath, ok := eachMetric.GetField(SensorPathKey)
			if ok {
				fieldFilters := m.FieldFilter[sensorPath.(string)]
				if len(fieldFilters) == 0 {
					m.Log.Warnf("the %s 's field filters is empty...", sensorPath)
				}
				fieldFilters = append(fieldFilters, headerFilter...)
				for _, filter := range fieldFilters {
					if ok, matchKeys := matchField(filter, eachMetric.FieldList()); ok {
						for _, realKey := range matchKeys {
							eachMetric.RemoveField(realKey)
						}
					}
				}
			}
		}
	}

	// field to tag
	for _, eachMetric := range in {
		sensorPath, ok := eachMetric.GetField(SensorPathKey)
		if ok {
			tags := m.Tag[sensorPath.(string)]
			if len(tags) == 0 {
				m.Log.Warnf("the %s 's tag is empty...", sensorPath)
			}
			tags = append(tags, headerTag...)
			for _, tag := range tags {
				if ok, matchKeys := matchField(tag, eachMetric.FieldList()); ok {
					for _, realKey := range matchKeys {
						value, ok := eachMetric.GetField(realKey)
						if ok {
							typeOfV := reflect.TypeOf(value)
							if typeOfV.Name() != "string" {
								if typeOfV.Name() != "int64" {
									m.Log.Errorf("wrong with metric tag [%s %s], it's type is %s", sensorPath.(string), tag, typeOfV.Name())
									m.Log.Error("telegraf stopped because of error")
									return nil
								}
								value = strconv.FormatInt(value.(int64), 10)
							} else {
								// 默认标签占位：空字符串时填充为 "N/A"，避免被序列化时丢弃
								if strings.TrimSpace(value.(string)) == "" {
									value = "N/A"
								}
							}
							eachMetric.AddTag(realKey, value.(string))
							// 注意要删除真实字段键，而不是配置项名，否则字段会残留
							eachMetric.RemoveField(realKey)
						}
					}
				}
			}
		}
		res = append(res, eachMetric)
	}

	return res
}

func matchField(key string, fields []*telegraf.Field) (bool, []string) {
	var matches []string
	for _, field := range fields {
		if field == nil {
			continue
		}
		// 按字段名“结尾匹配”，例如匹配 *.receiveByte
		if strings.HasSuffix(field.Key, key) {
			matches = append(matches, field.Key)
		}
	}
	if len(matches) > 0 {
		return true, matches
	}
	return false, matches
}

func init() {
	processors.Add("metric_match", func() telegraf.Processor {
		return &MetricMatch{}
	})
}
