package prometheus

import (
	"strings"
	"unicode"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/model"

	"github.com/influxdata/telegraf"
)

type table struct {
	first *unicode.RangeTable
	rest  *unicode.RangeTable
}

var metricNameTable = table{
	first: &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x003A, 0x003A, 1}, // :
			{0x0041, 0x005A, 1}, // A-Z
			{0x005F, 0x005F, 1}, // _
			{0x0061, 0x007A, 1}, // a-z
		},
		LatinOffset: 4,
	},
	rest: &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x0030, 0x003A, 1}, // 0-:
			{0x0041, 0x005A, 1}, // A-Z
			{0x005F, 0x005F, 1}, // _
			{0x0061, 0x007A, 1}, // a-z
		},
		LatinOffset: 4,
	},
}

var labelNameTable = table{
	first: &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x0041, 0x005A, 1}, // A-Z
			{0x005F, 0x005F, 1}, // _
			{0x0061, 0x007A, 1}, // a-z
		},
		LatinOffset: 3,
	},
	rest: &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x0030, 0x0039, 1}, // 0-9
			{0x0041, 0x005A, 1}, // A-Z
			{0x005F, 0x005F, 1}, // _
			{0x0061, 0x007A, 1}, // a-z
		},
		LatinOffset: 4,
	},
}

// sanitize checks if the name is valid, according to the table.
// If not, it attempts to replace invalid runes with an underscore to create a valid name.
func sanitize(name string, table table) (string, bool) {
	var b strings.Builder

	for i, r := range name {
		switch i {
		case 0:
			if unicode.In(r, table.first) {
				b.WriteRune(r)
			}
		default:
			if unicode.In(r, table.rest) {
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

// SanitizeMetricName checks if the name is a valid Prometheus metric name.
// If not, it attempts to replace invalid runes with an underscore to create a valid name.
func SanitizeMetricName(name string) (string, bool) {
	if model.LegacyValidation.IsValidMetricName(name) {
		return name, true
	}
	return sanitize(name, metricNameTable)
}

// SanitizeLabelName checks if the name is a valid Prometheus label name.
// If not, it attempts to replace invalid runes with an underscore to create a valid name.
func SanitizeLabelName(name string) (string, bool) {
	if model.LegacyValidation.IsValidLabelName(name) {
		return name, true
	}
	return sanitize(name, labelNameTable)
}

// sanitizeMetricNameUTF8 checks if the name is a valid Prometheus metric name
// using UTF-8 validation. If the name is invalid (including non-UTF-8 input),
// it falls back to legacy ASCII sanitization.
func sanitizeMetricNameUTF8(name string) (string, bool) {
	if model.UTF8Validation.IsValidMetricName(name) {
		return name, true
	}
	return sanitize(name, metricNameTable)
}

// sanitizeLabelNameUTF8 checks if the name is a valid Prometheus label name
// using UTF-8 validation. If the name is invalid (including non-UTF-8 input),
// it falls back to legacy ASCII sanitization.
func sanitizeLabelNameUTF8(name string) (string, bool) {
	if model.UTF8Validation.IsValidLabelName(name) {
		return name, true
	}
	return sanitize(name, labelNameTable)
}

// SanitizeMetricNameByEncoding sanitizes metric names according to the
// configured name sanitization mode.
func SanitizeMetricNameByEncoding(name, mode string) (string, bool) {
	switch mode {
	case "utf8":
		return sanitizeMetricNameUTF8(name)
	default:
		return SanitizeMetricName(name)
	}
}

// SanitizeLabelNameByEncoding sanitizes label names according to the configured
// name sanitization mode.
func SanitizeLabelNameByEncoding(name, mode string) (string, bool) {
	switch mode {
	case "utf8":
		return sanitizeLabelNameUTF8(name)
	default:
		return SanitizeLabelName(name)
	}
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

func metricType(valueType telegraf.ValueType) *dto.MetricType {
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
