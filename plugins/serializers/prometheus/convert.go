package prometheus

import (
	"strings"
	"unicode"

	"github.com/influxdata/telegraf"
	dto "github.com/prometheus/client_model/go"
)

var FirstTable = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0x0041, 0x005A, 1}, // A-Z
		{0x005F, 0x005F, 1}, // _
		{0x0061, 0x007A, 1}, // a-z
	},
	LatinOffset: 3,
}

var RestTable = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0x0030, 0x0039, 1}, // 0-9
		{0x0041, 0x005A, 1}, // A-Z
		{0x005F, 0x005F, 1}, // _
		{0x0061, 0x007A, 1}, // a-z
	},
	LatinOffset: 4,
}

func isValid(name string) bool {
	if name == "" {
		return false
	}

	for i, r := range name {
		switch {
		case i == 0:
			if !unicode.In(r, FirstTable) {
				return false
			}
		default:
			if !unicode.In(r, RestTable) {
				return false
			}
		}
	}

	return true
}

// SanitizeName check if the name is a valid Prometheus metric name and label
// name.  If not, it attempts to replaces invalid runes with an underscore to
// create a valid name.  Returns the metric name and true if the name is valid
// to use.
func SanitizeName(name string) (string, bool) {
	if isValid(name) {
		return name, true
	}

	var b strings.Builder

	for i, r := range name {
		switch {
		case i == 0:
			if unicode.In(r, FirstTable) {
				b.WriteRune(r)
			}
		default:
			if unicode.In(r, RestTable) {
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

// MetricName returns the Prometheus metric name.
func MetricName(measurement, fieldKey string, valueType telegraf.ValueType) string {
	switch valueType {
	case telegraf.Histogram, telegraf.Summary:
		switch {
		case strings.HasSuffix(fieldKey, "_bucket"):
			fieldKey = strings.TrimSuffix(fieldKey, "_bucket")
		case strings.HasSuffix(fieldKey, "_sum"):
			fieldKey = strings.TrimSuffix(fieldKey, "_sum")
		case strings.HasSuffix(fieldKey, "_count"):
			fieldKey = strings.TrimSuffix(fieldKey, "_count")
		}
	}

	if measurement == "prometheus" {
		return fieldKey
	}
	return measurement + "_" + fieldKey
}

func MetricType(valueType telegraf.ValueType) *dto.MetricType {
	switch valueType {
	case telegraf.Counter:
		return dto.MetricType_COUNTER.Enum()
	case telegraf.Gauge:
		return dto.MetricType_GAUGE.Enum()
	case telegraf.Summary:
		return dto.MetricType_SUMMARY.Enum()
	case telegraf.Untyped:
		return dto.MetricType_UNTYPED.Enum()
	case telegraf.Histogram:
		return dto.MetricType_HISTOGRAM.Enum()
	default:
		panic("unknown telegraf.ValueType")
	}
}

// SampleValue converts a field value into a value suitable for a simple sample value.
func SampleValue(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case uint64:
		return float64(v), true
	case bool:
		if v {
			return 1.0, true
		}
		return 0.0, true
	default:
		return 0, false
	}
}

// SampleCount converts a field value into a count suitable for a metric family
// of the Histogram or Summary type.
func SampleCount(value interface{}) (uint64, bool) {
	switch v := value.(type) {
	case float64:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int64:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case uint64:
		return v, true
	default:
		return 0, false
	}
}

// SampleSum converts a field value into a sum suitable for a metric family
// of the Histogram or Summary type.
func SampleSum(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}
