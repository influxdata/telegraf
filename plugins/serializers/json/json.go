package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	jsonata "github.com/blues/jsonata-go"

	"github.com/influxdata/telegraf"
)

type Serializer struct {
	TimestampUnits  time.Duration
	TimestampFormat string

	transformation *jsonata.Expr
}

func NewSerializer(timestampUnits time.Duration, timestampFormat, transform string) (*Serializer, error) {
	s := &Serializer{
		TimestampUnits:  truncateDuration(timestampUnits),
		TimestampFormat: timestampFormat,
	}

	if transform != "" {
		e, err := jsonata.Compile(transform)
		if err != nil {
			return nil, err
		}
		s.transformation = e
	}

	return s, nil
}

func (s *Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	var obj interface{}
	obj = s.createObject(metric)

	if s.transformation != nil {
		var err error
		if obj, err = s.transform(obj); err != nil {
			if errors.Is(err, jsonata.ErrUndefined) {
				return nil, fmt.Errorf("%v (maybe configured for batch mode?)", err)
			}
			return nil, err
		}
	}

	serialized, err := json.Marshal(obj)
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

	var obj interface{}
	obj = map[string]interface{}{
		"metrics": objects,
	}

	if s.transformation != nil {
		var err error
		if obj, err = s.transform(obj); err != nil {
			if errors.Is(err, jsonata.ErrUndefined) {
				return nil, fmt.Errorf("%v (maybe configured for non-batch mode?)", err)
			}
			return nil, err
		}
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

func (s *Serializer) transform(obj interface{}) (interface{}, error) {
	return s.transformation.Eval(obj)
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
