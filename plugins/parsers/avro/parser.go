package avro

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jeremywohl/flatten/v2"
	"github.com/linkedin/goavro/v2"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
)

// If SchemaRegistry is set, we assume that our input will be in
// Confluent Wire Format
// (https://docs.confluent.io/platform/current/schema-registry/serdes-develop/index.html#wire-format) and we will load the schema from the registry.

// If Schema is set, we assume the input will be Avro binary format, without
// an attached schema or schema fingerprint

type Parser struct {
	MetricName      string            `toml:"metric_name"`
	SchemaRegistry  string            `toml:"avro_schema_registry"`
	Schema          string            `toml:"avro_schema"`
	Format          string            `toml:"avro_format"`
	Measurement     string            `toml:"avro_measurement"`
	Tags            []string          `toml:"avro_tags"`
	Fields          []string          `toml:"avro_fields"`
	Timestamp       string            `toml:"avro_timestamp"`
	TimestampFormat string            `toml:"avro_timestamp_format"`
	FieldSeparator  string            `toml:"avro_field_separator"`
	DefaultTags     map[string]string `toml:"tags"`

	Log         telegraf.Logger `toml:"-"`
	registryObj *schemaRegistry
}

func (p *Parser) Init() error {
	switch p.Format {
	case "":
		p.Format = "binary"
	case "binary", "json":
		// Do nothing as those are valid settings
	default:
		return fmt.Errorf("unknown 'avro_format' %q", p.Format)
	}

	if (p.Schema == "" && p.SchemaRegistry == "") || (p.Schema != "" && p.SchemaRegistry != "") {
		return errors.New("exactly one of 'schema_registry' or 'schema' must be specified")
	}
	if p.TimestampFormat == "" {
		if p.Timestamp != "" {
			return errors.New("if 'timestamp' field is specified, 'timestamp_format' must be as well")
		}
		if p.TimestampFormat != "unix" && p.TimestampFormat != "unix_us" && p.TimestampFormat != "unix_ms" && p.TimestampFormat != "unix_ns" {
			return fmt.Errorf("invalid timestamp format '%v'", p.TimestampFormat)
		}
	}
	if p.SchemaRegistry != "" {
		p.registryObj = newSchemaRegistry(p.SchemaRegistry)
	}

	return nil
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	var schema string
	var codec *goavro.Codec
	var err error
	var message []byte
	message = buf[:]

	if p.registryObj != nil {
		// The input must be Confluent Wire Protocol
		if buf[0] != 0 {
			return nil, errors.New("first byte is not 0: not Confluent Wire Protocol")
		}
		schemaID := int(binary.BigEndian.Uint32(buf[1:5]))
		schemastruct, err := p.registryObj.getSchemaAndCodec(schemaID)
		if err != nil {
			return nil, err
		}
		schema = schemastruct.Schema
		codec = schemastruct.Codec
		message = buf[5:]
	} else {
		// Check for single-object encoding
		magicBytes := int(binary.BigEndian.Uint16(buf[:2]))
		expectedMagic := int(binary.BigEndian.Uint16([]byte("c301")))
		if magicBytes == expectedMagic {
			message = buf[10:]
			// We could in theory validate the fingerprint against
			// the schema.  Maybe later.
			// We would get the fingerprint as int(binary.LittleEndian.Uint64(buf[2:10]))
		} // Otherwise we assume bare Avro binary
		schema = p.Schema
		codec, err = goavro.NewCodec(schema)
		if err != nil {
			return nil, err
		}
	}

	var native interface{}
	switch p.Format {
	case "binary":
		native, _, err = codec.NativeFromBinary(message)
	case "json":
		native, _, err = codec.NativeFromTextual(message)
	default:
		return nil, fmt.Errorf("unknown format %q", p.Format)
	}
	if err != nil {
		return nil, err
	}
	// Cast to string-to-interface
	codecSchema, ok := native.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("native is of unsupported type %T", native)
	}
	m, err := p.createMetric(codecSchema, schema)
	if err != nil {
		return nil, err
	}

	return []telegraf.Metric{m}, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	if len(metrics) != 1 {
		return nil, errors.New("line contains multiple metrics")
	}

	return metrics[0], nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) createMetric(data map[string]interface{}, schema string) (telegraf.Metric, error) {
	// Tags differ from fields, in that tags are inherently strings.
	// fields can be of any type.
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	// Set default tag values
	for k, v := range p.DefaultTags {
		tags[k] = v
	}
	// Avro doesn't have a Tag/Field distinction, so we have to tell
	// Telegraf which items are our tags.
	for _, tag := range p.Tags {
		sTag, err := internal.ToString(data[tag])
		if err != nil {
			p.Log.Warnf("Could not convert %v to string for tag %q: %v", data[tag], tag, err)
			continue
		}
		tags[tag] = sTag
	}
	var fieldList []string
	if len(p.Fields) != 0 {
		// If you have specified your fields in the config, you
		// get what you asked for.
		fieldList = p.Fields

		// Except...if you specify the timestamp field, and it's
		// not listed in your fields, you'll get it anyway.
		// This will randomize your field ordering, which isn't
		// ideal.  If you care, list the timestamp field.
		if p.Timestamp != "" {
			// quick list-to-set-to-list implementation
			fieldSet := make(map[string]bool)
			for k := range fieldList {
				fieldSet[fieldList[k]] = true
			}
			fieldSet[p.Timestamp] = true
			var newList []string
			for s := range fieldSet {
				newList = append(newList, s)
			}
			fieldList = newList
		}
	} else {
		for k := range data {
			// Otherwise, that which is not a tag is a field
			if _, ok := tags[k]; !ok {
				fieldList = append(fieldList, k)
			}
		}
	}
	// We need to flatten out our fields.  The default (the separator
	// string is empty) is equivalent to what streamreactor does.
	sep := flatten.SeparatorStyle{
		Before: "",
		Middle: p.FieldSeparator,
		After:  "",
	}
	for _, fld := range fieldList {
		candidate := make(map[string]interface{})
		candidate[fld] = data[fld] // 1-item map
		flat, err := flatten.Flatten(candidate, "", sep)
		if err != nil {
			return nil, fmt.Errorf("flatten field %q failed: %w", fld, err)
		}
		for k, v := range flat {
			fields[k] = v
		}
	}

	var schemaObj map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &schemaObj); err != nil {
		return nil, fmt.Errorf("unmarshaling schema failed: %w", err)
	}
	if len(fields) == 0 {
		// A telegraf metric needs at least one field.
		return nil, errors.New("number of fields is 0; unable to create metric")
	}
	// Now some fancy stuff to extract the measurement.
	// If it's set in the configuration, use that.
	name := p.Measurement
	separator := "."
	if name == "" {
		// Try using the namespace defined in the schema. In case there
		// is none, just use the schema's name definition.
		nsStr, ok := schemaObj["namespace"].(string)
		// namespace is optional
		if !ok {
			separator = ""
		}

		nStr, ok := schemaObj["name"].(string)
		if !ok {
			return nil, fmt.Errorf("could not determine name from schema %s", schema)
		}
		name = nsStr + separator + nStr
	}
	// Still don't have a name?  Guess we should use the metric name if
	// it's set.
	if name == "" {
		name = p.MetricName
	}
	// Nothing?  Give up.
	if name == "" {
		return nil, errors.New("could not determine measurement name")
	}
	var timestamp time.Time
	if p.Timestamp != "" {
		rawTime := fmt.Sprintf("%v", fields[p.Timestamp])
		var err error
		timestamp, err = internal.ParseTimestamp(p.TimestampFormat, rawTime, nil)
		if err != nil {
			return nil, fmt.Errorf("could not parse '%s' to '%s'", rawTime, p.TimestampFormat)
		}
	} else {
		timestamp = time.Now()
	}
	return metric.New(name, tags, fields, timestamp), nil
}

func init() {
	parsers.Add("avro",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{MetricName: defaultMetricName}
		})
}
