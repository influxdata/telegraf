package influx_protobuf

import (
	"fmt"

	influx "github.com/influxdata/influxdb-pb-data-protocol/golang"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

type table struct {
	Time   *influx.Column
	Tags   map[string]*influx.Column
	Fields map[string]*influx.Column
	rows   uint32
}

func newTable() table {
	return table{
		Time: &influx.Column{
			ColumnName:   "time",
			SemanticType: influx.Column_SEMANTIC_TYPE_TIME,
			Values:       &influx.Column_Values{},
		},
		Tags:   make(map[string]*influx.Column),
		Fields: make(map[string]*influx.Column),
	}
}

func (t *table) addMetric(metric telegraf.Metric, isiox bool) error {
	// Compute the byte and bit-offset for null-value handling
	offset, bit := (t.rows / 8), (t.rows % 8)

	// Add time (we leave out null-mask as all values are set)
	t.Time.Values.I64Values = append(t.Time.Values.I64Values, metric.Time().UnixNano())

	// Add tag columns
	for _, tag := range metric.TagList() {
		col, found := t.Tags[tag.Key]
		if !found {
			// We get a new tag column, so create a new column and mark
			// all previous entries as null-values
			col = &influx.Column{
				ColumnName:   tag.Key,
				SemanticType: influx.Column_SEMANTIC_TYPE_TAG,
				Values:       &influx.Column_Values{StringValues: make([]string, t.rows)},
				NullMask:     make([]byte, offset+1),
			}
			if isiox {
				col.SemanticType = influx.Column_SEMANTIC_TYPE_IOX
			}
			// Set all null-bits for the previous entries
			for i := uint32(0); i < offset; i++ {
				col.NullMask[i] = ^byte(0)
			}
			col.NullMask[offset] = byte(1<<bit+1) - 1
		}

		missing := int(offset+1) - len(col.NullMask)
		if missing > 0 {
			col.NullMask = append(col.NullMask, make([]byte, missing)...)
		}

		// Add the tag
		col.Values.StringValues = append(col.Values.StringValues, tag.Value)
		col.NullMask[offset] = col.NullMask[offset] & ^byte(1<<bit)
		t.Tags[tag.Key] = col
	}

	// Add field columns
	for _, field := range metric.FieldList() {
		col, found := t.Fields[field.Key]
		if !found {
			// We get a new tag column, so create a new column and mark
			// all previous entries as null-values
			col = &influx.Column{
				ColumnName:   field.Key,
				SemanticType: influx.Column_SEMANTIC_TYPE_FIELD,
				Values:       &influx.Column_Values{},
				NullMask:     make([]byte, offset+1),
			}
			if isiox {
				col.SemanticType = influx.Column_SEMANTIC_TYPE_IOX
			}
			// Set all null-bits for the previous entries
			for i := uint32(0); i < offset; i++ {
				col.NullMask[i] = ^byte(0)
			}
			col.NullMask[offset] = byte(1<<bit+1) - 1
		}

		missing := int(offset+1) - len(col.NullMask)
		if missing > 0 {
			col.NullMask = append(col.NullMask, make([]byte, missing)...)
		}
		// Add the field
		switch v := field.Value.(type) {
		case int, int8, int16, int32, int64:
			n := len(col.Values.I64Values)
			if uint32(n) != t.rows {
				return fmt.Errorf("field %q of type %T has insufficient length (%d != %d)", field.Key, v, n, t.rows)
			}
			x, err := internal.ToInt64(field.Value)
			if err != nil {
				return fmt.Errorf("converting field %q of type %T failed: %v", field.Key, v, err)
			}
			col.Values.I64Values = append(col.Values.I64Values, x)
		case uint, uint8, uint16, uint32, uint64:
			n := len(col.Values.U64Values)
			if uint32(n) != t.rows {
				return fmt.Errorf("field %q of type %T has insufficient length (%d != %d)", field.Key, v, n, t.rows)
			}
			x, err := internal.ToUint64(field.Value)
			if err != nil {
				return fmt.Errorf("converting field %q of type %T failed: %v", field.Key, v, err)
			}
			col.Values.U64Values = append(col.Values.U64Values, x)
		case float32, float64:
			n := len(col.Values.F64Values)
			if uint32(n) != t.rows {
				return fmt.Errorf("field %q of type %T has insufficient length (%d != %d)", field.Key, v, n, t.rows)
			}
			x, err := internal.ToFloat64(field.Value)
			if err != nil {
				return fmt.Errorf("converting field %q of type %T failed: %v", field.Key, v, err)
			}
			col.Values.F64Values = append(col.Values.F64Values, x)
		case string:
			n := len(col.Values.StringValues)
			if uint32(n) != t.rows {
				return fmt.Errorf("field %q of type %T has insufficient length (%d != %d)", field.Key, v, n, t.rows)
			}
			col.Values.StringValues = append(col.Values.StringValues, v)
		case bool:
			n := len(col.Values.BoolValues)
			if uint32(n) != t.rows {
				return fmt.Errorf("field %q of type %T has insufficient length (%d != %d)", field.Key, v, n, t.rows)
			}
			col.Values.BoolValues = append(col.Values.BoolValues, v)
		case []byte:
			n := len(col.Values.BytesValues)
			if uint32(n) != t.rows {
				return fmt.Errorf("field %q of type %T has insufficient length (%d != %d)", field.Key, v, n, t.rows)
			}
			col.Values.BytesValues = append(col.Values.BytesValues, v)
		default:
			return fmt.Errorf("field %q contains unknown type %t", field.Key, v)
		}
		col.NullMask[offset] = col.NullMask[offset] & ^byte(1<<bit)
		t.Fields[field.Key] = col
	}
	t.rows++

	return nil
}
