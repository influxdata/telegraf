package vngcloud_vmonitor

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/influxdata/telegraf"
)

type serializer struct {
	TimestampUnits time.Duration
}

type Table struct {
	First           *unicode.RangeTable
	Rest            *unicode.RangeTable
	DimensionValues *unicode.RangeTable
}

const (
	MaxCharDmsValue   = 255
	MinCharDmsValue   = 1
	MaxCharMetricName = 255
	MinCharMetricName = 1
)

var LabelNameTable = Table{
	First: &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x0041, 0x005A, 1}, // A-Z
			{0x005F, 0x005F, 1}, // _
			{0x0061, 0x007A, 1}, // a-z
		},
		LatinOffset: 3,
	},
	Rest: &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x002D, 0x0039, 1}, // - . / and 0-9
			{0x0041, 0x005A, 1}, // A-Z
			{0x005F, 0x005F, 1}, // _
			{0x0061, 0x007A, 1}, // a-z
		},
		LatinOffset: 4,
	},
}

var LabelValueTable = Table{
	// black list char
	DimensionValues: &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x0022, 0x0022, 1}, // "
			{0x0026, 0x0026, 1}, // &
			{0x003B, 0x003B, 1}, // ;
			{0x003C, 0x003E, 1}, // < = >
			{0x005C, 0x005C, 1}, // \
			{0x007B, 0x007D, 1}, // { | }
		},
		LatinOffset: 6,
	},
}

var LabelValueWhitelistTable = Table{
	// white list char
	DimensionValues: &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x002D, 0x0039, 1}, // - . / and 0-9
			{0x0041, 0x005A, 1}, // A-Z
			{0x005F, 0x005F, 1}, // _
			{0x0061, 0x007A, 1}, // a-z
		},
		LatinOffset: 4,
	},
}

var MetricNameTable = Table{
	First: &unicode.RangeTable{
		R16: []unicode.Range16{
			// {0x003A, 0x003A, 1}, // :
			{0x0041, 0x005A, 1}, // A-Z
			{0x005F, 0x005F, 1}, // _
			{0x0061, 0x007A, 1}, // a-z
		},
		LatinOffset: 3,
	},
	Rest: &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x002D, 0x0039, 1}, // - . / and 0-9
			{0x0041, 0x005A, 1}, // A-Z
			{0x005F, 0x005F, 1}, // _
			{0x0061, 0x007A, 1}, // a-z
		},
		LatinOffset: 4,
	},
}

func isValid(name string, table Table) bool {
	if name == "" {
		return false
	}

	for i, r := range name {
		switch {
		case i == 0:
			if !unicode.In(r, table.First) {
				return false
			}
		default:
			if !unicode.In(r, table.Rest) {
				return false
			}
		}
	}

	return true
}

func isWhitelistDimensionsValue(name string, table Table) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if !unicode.In(r, table.DimensionValues) {
			return false
		}
	}
	return true
}

func isValidDimensionsValue(name string, table Table) bool {
	if name == "" {
		return false
	}

	for _, r := range name {
		if unicode.In(r, table.DimensionValues) {
			log.Printf("[vMonitor] Valid str: %s", name)
			return false
		}
	}

	return true
}

func sanitize(name string, table Table) (string, bool) {
	if len(name) > MaxCharMetricName || len(name) < MinCharMetricName {
		log.Printf("[vMonitor] Metric or Dimension name higher than max character length (%d): %s", len(name), name)
		return name, false
	}

	if isValid(name, table) {
		return name, true
	}

	var b strings.Builder

	for i, r := range name {
		switch {
		case i == 0:
			if unicode.In(r, table.First) {
				b.WriteRune(r)
			}
		default:
			if unicode.In(r, table.Rest) {
				b.WriteRune(r)
			} else {
				b.WriteString("_")
			}
		}
	}

	name = strings.Trim(b.String(), "_")
	if name == "" {
		return "", false
	}

	return name, true
}

func sanitizeWhitelistString(name string, table Table) (string, bool) {

	if len(name) > MaxCharDmsValue || len(name) < MinCharDmsValue {
		log.Printf("[vMonitor] Dimension value higher than max character length (%d): %s", len(name), name)
		return name, false
	}

	if isWhitelistDimensionsValue(name, table) {
		return name, true
	}

	// sanitize name
	var b strings.Builder

	for _, r := range name {
		if unicode.In(r, table.DimensionValues) {
			b.WriteRune(r)
		} else {
			b.WriteString("_")
		}
	}
	name = strings.Trim(b.String(), "_")
	if name == "" {
		return "", false
	}

	return name, true
}

func sanitizeString(name string, table Table) (string, bool) {

	if len(name) > MaxCharDmsValue || len(name) < MinCharDmsValue {
		log.Printf("[vMonitor] Valid max(%d): %s", len(name), name)
		return name, false
	}

	if isValidDimensionsValue(name, table) {
		return name, true
	}

	// sanitize name
	var b strings.Builder

	for _, r := range name {
		if unicode.In(r, table.DimensionValues) {
			b.WriteString("_")
		} else {
			b.WriteRune(r)
		}
	}
	name = strings.Trim(b.String(), "_")
	if name == "" {
		return "", false
	}

	return name, true
}

// SanitizeMetricName checks if the name is a valid Prometheus metric name.  If
// not, it attempts to replaces invalid runes with an underscore to create a
// valid name.
func SanitizeMetricName(name string) (string, bool) {
	return sanitize(name, MetricNameTable)
}

// SanitizeLabelName checks if the name is a valid Prometheus label name.  If
// not, it attempts to replaces invalid runes with an underscore to create a
// valid name.
func SanitizeLabelName(name string) (string, bool) {
	return sanitize(name, LabelNameTable)
}

func SanitizeLabelValue(name string) (string, bool) {
	return sanitizeString(name, LabelValueTable)
}

func SanitizeWhitelistLabelValue(name string) (string, bool) {
	return sanitizeWhitelistString(name, LabelValueWhitelistTable)
}

func NewSerializer(timestampUnits time.Duration) (*serializer, error) {
	s := &serializer{
		TimestampUnits: truncateDuration(timestampUnits),
	}
	return s, nil
}

func (s *serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	m, err := s.createObject(metric)
	if err != nil {
		return []byte{}, err
	}
	serialized, err := json.Marshal(m)
	if err != nil {
		return []byte{}, err
	}
	serialized = append(serialized, '\n')

	return serialized, nil
}

func (s *serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	var objects []interface{}

	for _, metric := range metrics {
		m, err := s.createObject(metric)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, v := range m {
			//log.Print(v)
			objects = append(objects, v)
		}
	}

	if len(objects) == 0 {
		return []byte{}, fmt.Errorf("invalid all metrics name")
	}
	serialized, err := json.Marshal(objects)

	if err != nil {
		return []byte{}, err
	}
	log.Printf("[vMonitor] Serialized batch %d metrics to %d objects", len(metrics), len(objects))
	return serialized, nil
}

// func (s *serializer) isNumeric(str string) bool {
// 	_, err := strconv.ParseFloat(str, 64)
// 	return err == nil
// }

func (s *serializer) convertValueToFloat(v interface{}, name string) (float64, bool) {

	invalidLog := func() {
		log.Printf("[vMonitor] Metric_value invalid with value: %s -> %v(%T)", name, v, v)
	}

	switch fv := v.(type) {
	case int64:
		return float64(fv), true
	case uint64:
		return float64(fv), true
	case bool:
		if fv {
			return 1.0, true
		} else {
			return 0.0, true
		}
	case float64:
		return fv, true
	case string:
		metricValue, err := strconv.ParseFloat(fv, 64)
		if err != nil {
			invalidLog()
			return 0, false
		}
		return metricValue, true
	default:
		invalidLog()
		return 0, false
	}
}

func (s *serializer) createObject(metric telegraf.Metric) ([]map[string]interface{}, error) {
	var metrics []map[string]interface{}
	metricNamePrefix, ok := SanitizeMetricName(metric.Name())
	if !ok {
		return nil, fmt.Errorf("invalid metric name %s", metric.Name())
	}
	tags := make(map[string]string, len(metric.TagList()))

	for _, tag := range metric.TagList() {
		name, ok := SanitizeLabelName(tag.Key)
		if !ok || tag.Value == "" {
			continue
		}
		// valueTag, ok := SanitizeLabelValue(tag.Value)
		valueTag, ok := SanitizeWhitelistLabelValue(tag.Value)
		if !ok {
			continue
		}
		tags[name] = valueTag
	}

	for _, v := range metric.FieldList() {

		valueTag, ok := SanitizeMetricName(v.Key)
		if !ok {
			continue
		}
		metricName := fmt.Sprintf("%s.%s", metricNamePrefix, valueTag)
		metricValue, ok := s.convertValueToFloat(v.Value, metricName)
		if !ok {
			continue
		}

		m := make(map[string]interface{}, 4)
		m["dimensions"] = tags
		m["name"] = metricName
		m["value"] = metricValue
		m["timestamp"] = metric.Time().UnixNano() / int64(time.Millisecond)
		m["value_meta"] = make(map[string]interface{})
		metrics = append(metrics, m)
	}

	return metrics, nil
}

func truncateDuration(units time.Duration) time.Duration {
	// Default precision is 1s
	if units <= 0 {
		return time.Second
	}

	// Search for the power of ten less than the duration
	d := time.Nanosecond
	for {
		if d*10 > units {
			return d
		}
		d = d * 10
	}
}
