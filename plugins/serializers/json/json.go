package json

import (
	"encoding/json"
	"math"
	"time"

	"github.com/influxdata/telegraf"
)

type Serializer struct {
	TimestampUnits  time.Duration
	TimestampFormat string
}

func NewSerializer(timestampUnits time.Duration, timestampFormat string) (*Serializer, error) {
	s := &Serializer{
		TimestampUnits:  truncateDuration(timestampUnits),
		TimestampFormat: timestampFormat,
	}
	return s, nil
}

func (s *Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	m := s.createObject(metric)
	serialized, err := json.Marshal(m)
	if err != nil {
		return []byte{}, err
	}
	serialized = append(serialized, '\n')

	return serialized, nil
}

func (s *Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	objects := make([]interface{}, 0, len(metrics))
	for _, metric := range metrics {
		m := s.createObject(metric)
		objects = append(objects, m)
	}

	obj := map[string]interface{}{
		"metrics": objects,
	}

	serialized, err := json.Marshal(obj)
	if err != nil {
		return []byte{}, err
	}
	return serialized, nil
}

func (s *Serializer) createObject(metric telegraf.Metric) map[string]interface{} {
	m := make(map[string]interface{}, 4)

	tags := make(map[string]string, len(metric.TagList()))
	for _, tag := range metric.TagList() {
		tags[tag.Key] = tag.Value
	}
	m["tags"] = tags

	fields := make(map[string]interface{}, len(metric.FieldList()))
	for _, field := range metric.FieldList() {
		if fv, ok := field.Value.(float64); ok {
			// JSON does not support these special values
			if math.IsNaN(fv) || math.IsInf(fv, 0) {
				continue
			}
		}
		fields[field.Key] = field.Value
	}
	m["fields"] = fields

	m["name"] = metric.Name()
	if s.TimestampFormat == "" {
		m["timestamp"] = metric.Time().UnixNano() / int64(s.TimestampUnits)
	} else {
		m["timestamp"] = metric.Time().UTC().Format(s.TimestampFormat)
	}
	return m
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
