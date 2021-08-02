package serializer

import (
	"bytes"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs/sematext/tags"
	"math"
	"sort"
	"strconv"
	"strings"
)

type SerializationConfig struct {
	notSerializable map[string]bool
}

func NewSerializationConfig() *SerializationConfig {
	return &SerializationConfig{
		notSerializable: map[string]bool{tags.SematextProcessedTag: true},
	}
}

// MetricSerializer is an interface implemented by different metric serialization implementations.
type MetricSerializer interface {
	// Write serializes metrics from metrics parameter and returns the result in []byte
	Write(metrics []telegraf.Metric) []byte
}

// LinePerMetricSerializer provides simple implementation which writes each metric into a new line. The logic is simpler
// and lighter, but the resulting output will be bigger.
type LinePerMetricSerializer struct {
	log    telegraf.Logger
	config *SerializationConfig
}

// NewLinePerMetricSerializer creates an instance of LinePerMetricSerializer
func NewLinePerMetricSerializer(log telegraf.Logger) *LinePerMetricSerializer {
	return &LinePerMetricSerializer{
		log:    log,
		config: NewSerializationConfig(),
	}
}

// NewMetricSerializer creates and instance serializer which should be used to produce Sematext metrics format
func NewMetricSerializer(log telegraf.Logger) MetricSerializer {
	return NewCompactMetricSerializer(log)
}

// Write serializes input metrics array according to Sematext variant of influx line protocol. The output is returned
// as []byte which can be empty if there were no metrics or metrics couldn't be serialized.
func (s *LinePerMetricSerializer) Write(metrics []telegraf.Metric) []byte {
	var output bytes.Buffer

	// sematext format is based on influx line protocol: namespace,tags metrics timestamp
	for _, metric := range metrics {
		if len(metric.Fields()) == 0 {
			s.log.Debugf("Skipping the serialization of metric %s without fields ", metric.Name())
			continue
		}

		serializedTags := serializeTags(metric.Tags(), s.config.notSerializable)
		serializedMetrics := serializeMetric(metric)
		serializedTimestamp := strconv.FormatInt(metric.Time().UnixNano(), 10)

		if serializedMetrics == "" {
			continue
		}

		_, _ = output.WriteString(nameEscape(metric.Name()))
		if serializedTags != "" {
			_, _ = output.WriteString(",")
			_, _ = output.WriteString(serializedTags)
		}
		_, _ = output.WriteString(" ")
		_, _ = output.WriteString(serializedMetrics)
		_, _ = output.WriteString(" ")
		_, _ = output.WriteString(serializedTimestamp)
		// has to end with a newline
		_, _ = output.WriteString("\n")
	}

	return output.Bytes()
}

func serializeTags(tagsMap map[string]string, notSerializable map[string]bool) string {
	var serializedTags strings.Builder

	// make tag order sorted
	sortedTagKeys := make([]string, 0, len(tagsMap))
	for t := range tagsMap {
		sortedTagKeys = append(sortedTagKeys, t)
	}
	sort.Strings(sortedTagKeys)

	for _, tagKey := range sortedTagKeys {
		if _, set := notSerializable[tagKey]; set {
			continue
		}

		tagValue := tagsMap[tagKey]
		if serializedTags.Len() > 0 {
			_, _ = serializedTags.WriteString(",")
		}

		_, _ = serializedTags.WriteString(escape(tagKey))
		_, _ = serializedTags.WriteString("=")
		_, _ = serializedTags.WriteString(escape(tagValue))
	}
	return serializedTags.String()
}

func serializeMetric(metric telegraf.Metric) string {
	var serializedMetrics strings.Builder

	// make the field order sorted
	sort.Slice(metric.FieldList(), func(i, j int) bool {
		return metric.FieldList()[i].Key < metric.FieldList()[j].Key
	})

	var countAdded = 0
	for _, field := range metric.FieldList() {
		var serializedMetric = serializeMetricField(field.Key, field.Value)

		if serializedMetric == "" {
			continue
		}

		if countAdded > 0 {
			_, _ = serializedMetrics.WriteString(",")
		}
		_, _ = serializedMetrics.WriteString(serializedMetric)
		countAdded++
	}

	return serializedMetrics.String()
}

func serializeMetricField(key string, value interface{}) string {
	var metricValue string
	switch v := value.(type) {
	case string:
		// temporarily made string values ignorable (until Sematext backend starts supporting them)
		// metricValue = fmt.Sprintf("\"%s\"", stringFieldEscape(v))
		return ""
	case bool:
		metricValue = strconv.FormatBool(v)
	case float64:
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			metricValue = strconv.FormatFloat(v, 'f', -1, 64)
		}
	case uint64:
		metricValue = strconv.FormatUint(v, 10) + "i"
	case int64:
		metricValue = strconv.FormatInt(v, 10) + "i"
	default:
		return ""
	}

	return fmt.Sprint(key, "=", metricValue)
}

// CompactMetricSerializer can be used to output metrics in compact format. Compact format squeezes as many metrics
// as possible in a single output line, based on tags and timestamp of each metric. When multiple metrics share the
// same tags and the same timestamp, we can write them in a single line to reduce the total bulk size by not
// repeating the same tags+timestamp multiple times.
type CompactMetricSerializer struct {
	log    telegraf.Logger
	config *SerializationConfig
}

// NewCompactMetricSerializer creates an instance of CompactMetricSerializer
func NewCompactMetricSerializer(log telegraf.Logger) *CompactMetricSerializer {
	return &CompactMetricSerializer{
		log:    log,
		config: NewSerializationConfig(),
	}
}

// Write serializes input metrics array according to Sematext variant of influx line protocol. The output is returned
// as []byte which can be empty if there were no metrics or metrics couldn't be serialized.
func (s *CompactMetricSerializer) Write(metrics []telegraf.Metric) []byte {
	var output bytes.Buffer
	idToMetrics := make(map[string][]telegraf.Metric)

	// first group the metrics that share the same identification
	for _, m := range metrics {
		id := buildID(m)
		idToMetrics[id] = append(idToMetrics[id], m)
	}

	// sort the keys to keep the order fixed
	sortedIds := make([]string, 0, len(idToMetrics))
	for i := range idToMetrics {
		sortedIds = append(sortedIds, i)
	}
	sort.Strings(sortedIds)

	// then create 1 metrics line for each of created groups
	for _, groupID := range sortedIds {
		metrics := idToMetrics[groupID]
		serializedTags := serializeTags(metrics[0].Tags(), s.config.notSerializable)
		serializedMetrics := serializeMetrics(metrics)
		serializedTimestamp := strconv.FormatInt(metrics[0].Time().UnixNano(), 10)

		if serializedMetrics == "" {
			continue
		}

		_, _ = output.WriteString(nameEscape(metrics[0].Name()))
		if serializedTags != "" {
			_, _ = output.WriteString(",")
			_, _ = output.WriteString(serializedTags)
		}
		_, _ = output.WriteString(" ")
		_, _ = output.WriteString(serializedMetrics)
		_, _ = output.WriteString(" ")
		_, _ = output.WriteString(serializedTimestamp)
		// has to end with a newline
		_, _ = output.WriteString("\n")
	}

	return output.Bytes()
}

func serializeMetrics(metrics []telegraf.Metric) string {
	var serializedMetrics strings.Builder
	fieldList := make([]*telegraf.Field, 0)

	for _, metric := range metrics {
		fieldList = append(fieldList, metric.FieldList()...)
	}

	// make the field order sorted
	sort.Slice(fieldList, func(i, j int) bool {
		return fieldList[i].Key < fieldList[j].Key
	})

	countAdded := 0
	for _, field := range fieldList {
		var serializedMetric = serializeMetricField(field.Key, field.Value)

		if serializedMetric == "" {
			continue
		}

		if countAdded > 0 {
			_, _ = serializedMetrics.WriteString(",")
		}
		_, _ = serializedMetrics.WriteString(serializedMetric)
		countAdded++
	}

	return serializedMetrics.String()
}

func buildID(metric telegraf.Metric) string {
	return strconv.FormatInt(metric.Time().UnixNano(), 10) + "-" + metric.Name() + "-" + tags.GetTagsKey(metric.Tags())
}
