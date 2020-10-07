package carbon2

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
)

type format string

const (
	carbon2FormatFieldEmpty          = format("")
	Carbon2FormatFieldSeparate       = format("field_separate")
	Carbon2FormatMetricIncludesField = format("metric_includes_field")
)

var formats = map[format]struct{}{
	carbon2FormatFieldEmpty:          {},
	Carbon2FormatFieldSeparate:       {},
	Carbon2FormatMetricIncludesField: {},
}

type Serializer struct {
	metricsFormat format
}

func NewSerializer(metricsFormat string) (*Serializer, error) {
	var f = format(metricsFormat)

	if _, ok := formats[f]; !ok {
		return nil, fmt.Errorf("unknown carbon2 format: %s", f)
	}

	// When unset, default to field separate.
	if f == carbon2FormatFieldEmpty {
		f = Carbon2FormatFieldSeparate
	}

	return &Serializer{
		metricsFormat: f,
	}, nil
}

func (s *Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	return s.createObject(metric), nil
}

func (s *Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	var batch bytes.Buffer
	for _, metric := range metrics {
		batch.Write(s.createObject(metric))
	}
	return batch.Bytes(), nil
}

func (s *Serializer) createObject(metric telegraf.Metric) []byte {
	var m bytes.Buffer
	metricsFormat := s.getMetricsFormat()

	for fieldName, fieldValue := range metric.Fields() {
		if !isNumeric(fieldValue) {
			continue
		}

		switch metricsFormat {
		case Carbon2FormatFieldSeparate:
			m.WriteString(serializeMetricFieldSeparate(
				metric.Name(), fieldName,
			))

		case Carbon2FormatMetricIncludesField:
			m.WriteString(serializeMetricIncludeField(
				metric.Name(), fieldName,
			))
		}

		for _, tag := range metric.TagList() {
			m.WriteString(strings.Replace(tag.Key, " ", "_", -1))
			m.WriteString("=")
			value := tag.Value
			if len(value) == 0 {
				value = "null"
			}
			m.WriteString(strings.Replace(value, " ", "_", -1))
			m.WriteString(" ")
		}
		m.WriteString(" ")
		m.WriteString(fmt.Sprintf("%v", fieldValue))
		m.WriteString(" ")
		m.WriteString(strconv.FormatInt(metric.Time().Unix(), 10))
		m.WriteString("\n")
	}
	return m.Bytes()
}

func (s *Serializer) SetMetricsFormat(f format) {
	s.metricsFormat = f
}

func (s *Serializer) getMetricsFormat() format {
	return s.metricsFormat
}

func (s *Serializer) IsMetricsFormatUnset() bool {
	return s.metricsFormat == carbon2FormatFieldEmpty
}

func serializeMetricFieldSeparate(name, fieldName string) string {
	return fmt.Sprintf("metric=%s field=%s ",
		strings.Replace(name, " ", "_", -1),
		strings.Replace(fieldName, " ", "_", -1),
	)
}

func serializeMetricIncludeField(name, fieldName string) string {
	return fmt.Sprintf("metric=%s_%s ",
		strings.Replace(name, " ", "_", -1),
		strings.Replace(fieldName, " ", "_", -1),
	)
}

func isNumeric(v interface{}) bool {
	switch v.(type) {
	case string:
		return false
	default:
		return true
	}
}
