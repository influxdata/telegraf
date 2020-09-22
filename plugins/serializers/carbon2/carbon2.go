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
	Carbon2FormatFieldSeparate       string = "field_separate"
	Carbon2FormatMetricIncludesField string = "metric_includes_field"

	formatFieldSeparate       = format(Carbon2FormatFieldSeparate)
	formatMetricIncludesField = format(Carbon2FormatMetricIncludesField)
)

var formats = map[string]format{
	// Field separate is the default when no format specified.
	"":                               formatFieldSeparate,
	Carbon2FormatFieldSeparate:       formatFieldSeparate,
	Carbon2FormatMetricIncludesField: formatMetricIncludesField,
}

type Serializer struct {
	metricsFormat format
}

func NewSerializer(f string) (*Serializer, error) {
	var (
		ok            bool
		metricsFormat format
	)
	if metricsFormat, ok = formats[f]; !ok {
		return nil, fmt.Errorf("unknown carbon2 format: %s", f)
	}

	return &Serializer{
		metricsFormat: metricsFormat,
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
	for fieldName, fieldValue := range metric.Fields() {
		if !isNumeric(fieldValue) {
			continue
		}

		switch s.metricsFormat {
		case formatFieldSeparate:
			m.WriteString(serializeMetricFieldSeparate(
				metric.Name(), fieldName,
			))
		case formatMetricIncludesField:
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
