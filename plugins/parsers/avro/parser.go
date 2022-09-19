package avro

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/jeremywohl/flatten"
	"github.com/linkedin/goavro/v2"
	"time"
)

type Parser struct {
	MetricName      string   `toml:"metric_name"`
	SchemaRegistry  string   `toml:"avro_schema_registry"`
	Schema          string   `toml:"avro_schema"`
	Measurement     string   `toml:"avro_measurement"`
	Tags            []string `toml:"avro_tags"`
	Fields          []string `toml:"avro_fields"`
	Timestamp       string   `toml:"avro_timestamp"`
	TimestampFormat string   `toml:"avro_timestamp_format"`
	DiscardArrays   bool     `toml:"avro_discard_arrays"`
	FieldSeparator  string   `toml:"avro_field_separator"`
	DefaultTags     map[string]string
	TimeFunc        func() time.Time

	Log telegraf.Logger `toml:"-"`
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	var schema string

	if len(buf) < 5 {
		err := fmt.Errorf("buf is %d bytes; must be at least 5", len(buf))
		return nil, err
	}
	schemaID := int(binary.BigEndian.Uint32(buf[1:5]))
	binaryData := buf[5:]

	switch {
	case p.SchemaRegistry != "":
		schemaRegistry := NewSchemaRegistry(p.SchemaRegistry)
		retrSchema, err := schemaRegistry.getSchema(schemaID)
		if err != nil {
			return nil, err
		}
		schema = retrSchema
	default:
		schema = p.Schema
	}

	codec, err := goavro.NewCodec(schema)

	if err != nil {
		return nil, err
	}

	native, _, err := codec.NativeFromBinary(binaryData)
	if err != nil {
		return nil, err
	}

	m, err := p.createMetric(native, schema)
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
		return nil, fmt.Errorf("line contains multiple metrics")
	}

	return metrics[0], nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) parseTimestamp(timestamp interface{}) (time.Time, error) {
	if timestamp == nil {
		return p.TimeFunc(), nil
	}

	if p.TimestampFormat == "" {
		return p.TimeFunc(), fmt.Errorf("must specify timestamp format")
	}

	metricTime, err := internal.ParseTimestamp(p.TimestampFormat, timestamp, "UTC")
	if err != nil {
		return p.TimeFunc(), err
	}

	return metricTime, nil
}

func (p *Parser) createMetric(native interface{}, schema string) (telegraf.Metric, error) {
	deep, ok := native.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot cast native interface {} to map[string]interface{}")
	}

	metricTime, err := p.parseTimestamp(nestedValue(deep[p.Timestamp]))
	if err != nil {
		return nil, err
	}

	fields := make(map[string]interface{})
	tags := make(map[string]string)

	for k, v := range p.DefaultTags {
		tags[k] = v
	}

	for _, tag := range p.Tags {
		if value, ok := deep[tag]; ok {
			tags[tag] = fmt.Sprintf("%v", nestedValue(value))
		} else {
			p.Log.Infof("tag %s value was %v; not added to tags", tag, value)
		}
	}
	fieldList := make([]string, len(p.Fields), (cap(p.Fields)+1)*2)
	copy(fieldList, p.Fields)
	if len(fieldList) == 0 { // Get fields from whatever we just unpacked
		for k := range deep {
			if _, ok := tags[k]; !ok {
				fieldList = append(fieldList, k)
			}
		}
	}
	for _, field := range fieldList {
		if value, ok := deep[field]; ok {
			fields[field] = nestedValue(value)
		} else {
			p.Log.Infof("field %s value was %v; not added to fields", field, value)
		}
	}
	var schemaObj map[string]interface{}

	err = json.Unmarshal([]byte(schema), &schemaObj)
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("number of fields is 0; unable to create metric")
	}
	name := ""
	if p.Measurement != "" {
		name = p.Measurement
	} else {
		// get Measurement name from schema
		nsStr, ok := schemaObj["namespace"].(string)
		if !ok {
			return nil, fmt.Errorf("could not determine namespace from schema %s", schema)
		}
		nStr, ok := schemaObj["name"].(string)
		if !ok {
			return nil, fmt.Errorf("could not determine name from schema %s", schema)
		}
		name = nsStr + "." + nStr
	}
	if name == "" {
		name = p.MetricName
	}
	if name == "" {
		return nil, fmt.Errorf("could not determine measurement name")
	}
	if p.DiscardArrays {
		// Any non-scalars end up being a nil field
		return metric.New(name, tags, fields, metricTime), nil
	}
	// But if we do it this way, we flatten any compound structures,
	// including arrays.  Goavro is only going to hand us back
	// arrays, not maps.
	// The default (the separator string is empty) is equivalent to
	// what streamreactor does.
	sep := flatten.SeparatorStyle{
		Before: "",
		Middle: p.FieldSeparator,
		After:  "",
	}
	flat, err := flatten.Flatten(fields, "", sep)
	if err != nil {
		return nil, err
	}
	return metric.New(name, tags, flat, metricTime), nil
}

func nestedValue(deep interface{}) interface{} {
	if m, ok := deep.(map[string]interface{}); ok {
		for _, value := range m {
			return nestedValue(value)
		}
	}
	return deep
}

func init() {
	parsers.Add("avro",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{MetricName: defaultMetricName}
		})
}

func (p *Parser) Init() error {
	p.TimeFunc = time.Now
	if p.Schema == "" && p.SchemaRegistry == "" {
		err := fmt.Errorf("one of SchemaRegistry or Schema must be specified")
		return err
	}
	return nil
}

func (p *Parser) InitFromConfig(config *parsers.Config) error {
	p.MetricName = config.MetricName
	p.SchemaRegistry = config.AvroSchemaRegistry
	p.Schema = config.AvroSchema
	p.Measurement = config.AvroMeasurement
	p.Tags = config.AvroTags
	p.Fields = config.AvroFields
	p.Timestamp = config.AvroTimestamp
	p.TimestampFormat = config.AvroTimestampFormat
	p.FieldSeparator = config.AvroFieldSeparator
	p.DefaultTags = config.DefaultTags

	return p.Init()
}
