package avro

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/linkedin/goavro/v2"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type Serializer struct {
	Schema           string          `toml:"avro_schema"`
	Format           string          `toml:"avro_format"`
	MeasurementField string          `toml:"avro_measurement_field"`
	Timestamp        string          `toml:"avro_timestamp"`
	TimestampFormat  string          `toml:"avro_timestamp_format"`
	Log              telegraf.Logger `toml:"-"`

	codec      *goavro.Codec
	fieldTypes map[string]interface{}
}

func (s *Serializer) Init() error {
	switch s.Format {
	case "":
		s.Format = "binary"
	case "binary", "json":
		// Valid settings
	default:
		return fmt.Errorf("unknown 'avro_format' %q", s.Format)
	}

	switch s.TimestampFormat {
	case "":
		s.TimestampFormat = "unix"
	case "unix", "unix_ms", "unix_us", "unix_ns":
		// Valid settings
	default:
		return fmt.Errorf("invalid 'avro_timestamp_format' %q", s.TimestampFormat)
	}

	if s.Schema == "" {
		return errors.New("'avro_schema' is required")
	}

	codec, err := goavro.NewCodec(s.Schema)
	if err != nil {
		return fmt.Errorf("parsing schema failed: %w", err)
	}
	s.codec = codec

	// Pull out the declared type of every field in the (record) schema so we
	// can coerce Telegraf's Go values to something goavro accepts.
	types, err := schemaFieldTypes(s.Schema)
	if err != nil {
		return err
	}
	s.fieldTypes = types

	return nil
}

func (s *Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	return s.encode(s.record(metric))
}

func (s *Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	var buf []byte
	for _, m := range metrics {
		b, err := s.encode(s.record(m))
		if err != nil {
			return nil, err
		}
		buf = append(buf, b...)
	}
	return buf, nil
}

func (s *Serializer) encode(record map[string]interface{}) ([]byte, error) {
	switch s.Format {
	case "binary":
		return s.codec.BinaryFromNative(nil, record)
	case "json":
		return s.codec.TextualFromNative(nil, record)
	default:
		return nil, fmt.Errorf("unknown format %q", s.Format)
	}
}

// record flattens a metric into a name-keyed map ready for Avro encoding.
// Tags and fields sharing a name collide; fields win since that matches how the
// parser round-trips them. goavro ignores map keys the schema doesn't mention,
// so we can hand it everything and let the schema select what it wants.
func (s *Serializer) record(metric telegraf.Metric) map[string]interface{} {
	record := make(map[string]interface{}, len(metric.TagList())+len(metric.FieldList())+2)

	for _, tag := range metric.TagList() {
		record[tag.Key] = s.coerce(tag.Key, tag.Value)
	}
	for _, field := range metric.FieldList() {
		record[field.Key] = s.coerce(field.Key, field.Value)
	}

	if s.MeasurementField != "" {
		record[s.MeasurementField] = s.coerce(s.MeasurementField, metric.Name())
	}
	if s.Timestamp != "" {
		t := metric.Time()
		var ts int64
		switch s.TimestampFormat {
		case "unix":
			ts = t.Unix()
		case "unix_ms":
			ts = t.UnixNano() / 1e6
		case "unix_us":
			ts = t.UnixNano() / 1e3
		case "unix_ns":
			ts = t.UnixNano()
		}
		record[s.Timestamp] = s.coerce(s.Timestamp, ts)
	}

	return record
}

// coerce converts a Telegraf value into the concrete Go type goavro wants for
// the schema field of the same name. Fields the schema doesn't declare, or
// types we don't special-case (records, arrays, maps, logical types), are
// passed through untouched.
func (s *Serializer) coerce(name string, value interface{}) interface{} {
	avroType, ok := s.fieldTypes[name]
	if !ok {
		return value
	}
	return coerceTo(avroType, value, s.Log)
}

func coerceTo(avroType, value interface{}, log telegraf.Logger) interface{} {
	switch t := avroType.(type) {
	case string:
		return coercePrimitive(t, value, log)
	case []interface{}:
		// A union, e.g. ["null", "long"]. Encode null as nil and otherwise
		// wrap the value in the first non-null branch we understand.
		if value == nil {
			return nil
		}
		for _, branch := range t {
			name, ok := branch.(string)
			if !ok || name == "null" {
				continue
			}
			return goavro.Union(name, coercePrimitive(name, value, log))
		}
		return value
	default:
		// Complex or logical type: leave it for goavro to handle.
		return value
	}
}

func coercePrimitive(avroType string, value interface{}, log telegraf.Logger) interface{} {
	switch avroType {
	case "string":
		s, err := internal.ToString(value)
		if err != nil {
			if log != nil {
				log.Warnf("Could not convert %v to string: %v", value, err)
			}
			return value
		}
		return s
	case "int":
		v, err := internal.ToInt64(value)
		if err != nil {
			logConvErr(log, value, avroType, err)
			return value
		}
		return int32(v)
	case "long":
		v, err := internal.ToInt64(value)
		if err != nil {
			logConvErr(log, value, avroType, err)
			return value
		}
		return v
	case "float":
		v, err := internal.ToFloat64(value)
		if err != nil {
			logConvErr(log, value, avroType, err)
			return value
		}
		return float32(v)
	case "double":
		v, err := internal.ToFloat64(value)
		if err != nil {
			logConvErr(log, value, avroType, err)
			return value
		}
		return v
	case "boolean":
		v, err := internal.ToBool(value)
		if err != nil {
			logConvErr(log, value, avroType, err)
			return value
		}
		return v
	case "bytes":
		switch v := value.(type) {
		case []byte:
			return v
		case string:
			return []byte(v)
		default:
			s, err := internal.ToString(value)
			if err != nil {
				logConvErr(log, value, avroType, err)
				return value
			}
			return []byte(s)
		}
	default:
		return value
	}
}

func logConvErr(log telegraf.Logger, value, avroType interface{}, err error) {
	if log != nil {
		log.Warnf("Could not convert %v to %v: %v", value, avroType, err)
	}
}

// schemaFieldTypes reads a record schema and returns a map of field name to its
// declared Avro type. Non-record schemas yield an empty map, which just means
// no coercion happens and every value is passed through as-is.
func schemaFieldTypes(schema string) (map[string]interface{}, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &parsed); err != nil {
		return nil, fmt.Errorf("unmarshalling schema failed: %w", err)
	}

	types := make(map[string]interface{})
	fields, ok := parsed["fields"].([]interface{})
	if !ok {
		return types, nil
	}
	for _, f := range fields {
		field, ok := f.(map[string]interface{})
		if !ok {
			continue
		}
		name, ok := field["name"].(string)
		if !ok {
			continue
		}
		types[name] = field["type"]
	}
	return types, nil
}

func init() {
	serializers.Add("avro",
		func() telegraf.Serializer {
			return &Serializer{}
		},
	)
}
