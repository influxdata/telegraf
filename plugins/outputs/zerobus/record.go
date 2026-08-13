package zerobus

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/influxdata/telegraf"
)

func metricToTableSchemaJSON(metric telegraf.Metric, timestampColumn, measurementColumn string, columns map[string]struct{}) ([]byte, error) {
	values := make(map[string]interface{}, len(metric.TagList())+len(metric.FieldList())+2)
	if timestampColumn != "" && keepColumn(timestampColumn, columns) {
		values[timestampColumn] = metric.Time().UnixMicro()
	}
	if measurementColumn != "" && keepColumn(measurementColumn, columns) {
		values[measurementColumn] = metric.Name()
	}

	// Tags and fields become columns of the destination table, so their names must not collide.
	for _, tag := range metric.TagList() {
		if tag == nil {
			return nil, errors.New("metric contains a nil tag")
		}
		if !keepColumn(tag.Key, columns) {
			continue
		}
		if _, found := values[tag.Key]; found {
			return nil, fmt.Errorf("tag %q conflicts with another table column", tag.Key)
		}
		values[tag.Key] = tag.Value
	}

	for _, field := range metric.FieldList() {
		if field == nil {
			return nil, errors.New("metric contains a nil field")
		}
		if !keepColumn(field.Key, columns) {
			continue
		}
		if _, found := values[field.Key]; found {
			return nil, fmt.Errorf("field %q conflicts with another table column", field.Key)
		}
		switch value := field.Value.(type) {
		case int64, bool, string:
			values[field.Key] = value
		case uint64:
			if value > math.MaxInt64 {
				return nil, fmt.Errorf("field %q contains uint64 value %d exceeding Delta BIGINT maximum %d", field.Key, value, int64(math.MaxInt64))
			}
			values[field.Key] = int64(value)
		case float64:
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return nil, fmt.Errorf("field %q contains a non-finite float", field.Key)
			}
			values[field.Key] = value
		default:
			return nil, fmt.Errorf("field %q has unsupported type %T", field.Key, field.Value)
		}
	}

	if columns != nil && len(values) == 0 {
		return nil, errors.New("metric has no columns matching the table")
	}

	record, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("marshaling JSON record failed: %w", err)
	}
	return record, nil
}

func keepColumn(name string, columns map[string]struct{}) bool {
	if columns == nil {
		return true
	}
	_, ok := columns[name]
	return ok
}
