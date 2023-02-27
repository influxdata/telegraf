package avro

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jeremywohl/flatten/v2"
	"github.com/linkedin/goavro/v2"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Parser struct {
	MetricName      string            `toml:"metric_name"`
	SchemaRegistry  string            `toml:"avro_schema_registry"`
	Schema          string            `toml:"avro_schema"`
	Measurement     string            `toml:"avro_measurement"`
	Tags            []string          `toml:"avro_tags"`
	Fields          []string          `toml:"avro_fields"`
	Timestamp       string            `toml:"avro_timestamp"`
	TimestampFormat string            `toml:"avro_timestamp_format"`
	FieldSeparator  string            `toml:"avro_field_separator"`
	DefaultTags     map[string]string `toml:"-"`

	Log         telegraf.Logger `toml:"-"`
	registryObj *SchemaRegistry
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	schemaID, binaryData, err := p.extractSchemaAndMessage(buf)
	if err != nil {
		return nil, err
	}

	var schema string
	var codec *goavro.Codec

	if p.SchemaRegistry != "" {
		schemastruct, err := p.registryObj.getSchemaAndCodec(schemaID)
		if err != nil {
			return nil, err
		}
		schema = schemastruct.Schema
		codec = schemastruct.Codec
	} else {
		schema = p.Schema
		codec, err = goavro.NewCodec(schema)
		if err != nil {
			return nil, err
		}
	}
	native, _, err := codec.NativeFromBinary(binaryData)
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

func (p *Parser) extractSchemaAndMessage(buf []byte) (int, []byte, error) {
	if len(buf) < 5 {
		err := fmt.Errorf("buf is %d bytes; must be at least 5", len(buf))
		return 0, nil, err
	}
	schemaID := int(binary.BigEndian.Uint32(buf[1:5]))
	binaryData := buf[5:]
	return schemaID, binaryData, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	if len(metrics) != 1 {
		return nil, fmt.Errorf("line contains multiple metrics")
	}

	return metrics[0], nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) createMetric(data map[string]interface{}, schema string) (telegraf.Metric, error) {
	now := time.Now()

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
		tags[tag], err = internal.ToString(data[tag])
		if err != nil {
			p.Log.Warnf("Could not convert %v to string for tag %q: %v", data[tag], tag, err)
		}
	}
	var fieldList []string
	if len(p.Fields) != 0 {
		// If you have specified your fields in the config, you
		// get what you asked for.
		fieldList = p.Fields
	} else {
		for k := range data {
			// Otherwise, that which is not a tag is a field
			if _, ok := tags[k]; !ok {
				fieldList = append(fieldList, k)
			}
		}
	}
	flatFields := make(map[string]interface{})
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
			return nil, fmt.Errorf("Failed to flatten field %s: %v", fld, err)
		}
		for k, v := range flat {
			flatFields[k] = v
		}
	}
	for fieldName, field := range flatFields {
		fields[fieldName] = field
	}

	var schemaObj map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &schemaObj); err != nil {
		return nil, fmt.Errorf("unmarshaling schema failed: %w", err)
	}
	if len(fields) == 0 {
		// A telegraf metric needs at least one field.
		return nil, fmt.Errorf("number of fields is 0; unable to create metric")
	}
	// Now some fancy stuff to extract the measurement.
	// If it's set in the configuration, use that.
	name := ""
	separator := "."
	if p.Measurement != "" {
		name = p.Measurement
	} else {
		// get Measurement name from schema.  Try using the
		// namespace concatenated to the name, but if no namespace,
		// just use the name.
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
		return nil, fmt.Errorf("could not determine measurement name")
	}
	return metric.New(name, tags, fields, timestamp), nil
}

func init() {
	parsers.Add("avro",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{MetricName: defaultMetricName}
		})
}

func (p *Parser) Init() error {
	if (p.Schema == "" && p.SchemaRegistry == "") || (p.Schema != "" && p.SchemaRegistry != "") {
		return fmt.Errorf("exactly one of 'schema_registry' or 'schema' must be specified")
	}
	if p.SchemaRegistry != "" {
		p.registryObj = NewSchemaRegistry(p.SchemaRegistry)
	}
	if p.TimestampFormat == "" {
		return fmt.Errorf("must specify 'timestamp_format'")
	}
	return nil
}
