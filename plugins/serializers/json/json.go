package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/blues/jsonata-go"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type Serializer struct {
	TimestampUnits      config.Duration `toml:"json_timestamp_units"`
	TimestampFormat     string          `toml:"json_timestamp_format"`
	Transformation      string          `toml:"json_transformation"`
	NestedFieldsInclude []string        `toml:"json_nested_fields_include"`
	NestedFieldsExclude []string        `toml:"json_nested_fields_exclude"`

	nestedFields filter.Filter
}

func (s *Serializer) Init() error {
	// Default precision is 1s
	if s.TimestampUnits <= 0 {
		s.TimestampUnits = config.Duration(time.Second)
	}

	// Search for the power of ten less than the duration
	d := time.Nanosecond
	t := time.Duration(s.TimestampUnits)
	for {
		if d*10 > t {
			t = d
			break
		}
		d = d * 10
	}
	s.TimestampUnits = config.Duration(t)

	if len(s.NestedFieldsInclude) > 0 || len(s.NestedFieldsExclude) > 0 {
		f, err := filter.NewIncludeExcludeFilter(s.NestedFieldsInclude, s.NestedFieldsExclude)
		if err != nil {
			return err
		}
		s.nestedFields = f
	}

	return nil
}

func (s *Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	m := s.createMetric(metric)

	var obj any
	if s.Transformation == "" {
		obj = m
	} else {
		var err error
		if obj, err = s.transform(m.toMap()); err != nil {
			if errors.Is(err, jsonata.ErrUndefined) {
				return nil, fmt.Errorf("%w (maybe configured for batch mode?)", err)
			}
			return nil, err
		}
	}

	serialized, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	serialized = append(serialized, '\n')

	return serialized, nil
}

func (s *Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	var obj any
	if s.Transformation == "" {
		batch := jsonBatch{Metrics: make([]jsonMetric, 0, len(metrics))}
		for _, metric := range metrics {
			batch.Metrics = append(batch.Metrics, s.createMetric(metric))
		}
		obj = batch
	} else {
		objects := make([]any, 0, len(metrics))
		for _, metric := range metrics {
			objects = append(objects, s.createMetric(metric).toMap())
		}

		var err error
		if obj, err = s.transform(map[string]any{"metrics": objects}); err != nil {
			if errors.Is(err, jsonata.ErrUndefined) {
				return nil, fmt.Errorf("%w (maybe configured for non-batch mode?)", err)
			}
			return nil, err
		}
	}

	serialized, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	serialized = append(serialized, '\n')

	return serialized, nil
}

// jsonMetric is one metric. Its fields are in the order that encoding/json
// writes the keys of the map from toMap, so both give the same output.
type jsonMetric struct {
	Fields    map[string]any    `json:"fields"`
	Name      string            `json:"name"`
	Tags      map[string]string `json:"tags"`
	Timestamp any               `json:"timestamp"`
}

type jsonBatch struct {
	Metrics []jsonMetric `json:"metrics"`
}

func (s *Serializer) createMetric(metric telegraf.Metric) jsonMetric {
	var m jsonMetric

	tags := make(map[string]string, len(metric.TagList()))
	for _, tag := range metric.TagList() {
		tags[tag.Key] = tag.Value
	}
	m.Tags = tags

	fields := make(map[string]any, len(metric.FieldList()))
	for _, field := range metric.FieldList() {
		val := field.Value
		switch fv := field.Value.(type) {
		case float64:
			// JSON does not support these special values
			if math.IsNaN(fv) || math.IsInf(fv, 0) {
				continue
			}
		case string:
			// Check for nested fields if any
			if s.nestedFields != nil && s.nestedFields.Match(field.Key) {
				bv := []byte(fv)
				if json.Valid(bv) {
					var nested any
					if err := json.Unmarshal(bv, &nested); err == nil {
						val = nested
					}
				}
			}
		}
		fields[field.Key] = val
	}
	m.Fields = fields

	m.Name = metric.Name()
	if s.TimestampFormat == "" {
		m.Timestamp = metric.Time().UnixNano() / int64(s.TimestampUnits)
	} else {
		m.Timestamp = metric.Time().UTC().Format(s.TimestampFormat)
	}
	return m
}

// toMap returns the metric as a map, for the JSONata transformation.
func (m jsonMetric) toMap() map[string]any {
	return map[string]any{
		"tags":      m.Tags,
		"fields":    m.Fields,
		"name":      m.Name,
		"timestamp": m.Timestamp,
	}
}

func (s *Serializer) transform(obj any) (any, error) {
	transformation, err := jsonata.Compile(s.Transformation)
	if err != nil {
		return nil, err
	}

	return transformation.Eval(obj)
}

func init() {
	serializers.Add("json",
		func() telegraf.Serializer {
			return &Serializer{}
		},
	)
}
