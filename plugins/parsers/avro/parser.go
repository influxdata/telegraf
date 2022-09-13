package avro

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/linkedin/goavro/v2"
)

type Parser struct {
	MetricName      string    `toml:"metric_name"`
	SchemaRegistry  string    `toml:"avro_schema_registry"`
	Schema          string    `toml:"avro_schema"`
	Measurement     string    `toml:"avro_measurement"`
	Tags            []string  `toml:"avro_tags"`
	Fields          []string  `toml:"avro_fields"`
	Timestamp       string    `toml:"avro_timestamp"`
	TimestampFormat string    `toml:"avro_timestamp_format"`
	DefaultTags     map[string]string
	TimeFunc        func() time.Time
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	var schema string

	if len(buf) < 5 {
		err := fmt.Errorf("buf is %d bytes; must be at least 5",len(buf))
		return nil, err
	}
	schemaID := int(binary.BigEndian.Uint32(buf[1:5]))
	binary_data := buf[5:]

	switch {
	case p.SchemaRegistry != "":
		schemaRegistry := NewSchemaRegistry(p.SchemaRegistry)
		retr_schema, err := schemaRegistry.getSchema(schemaID)
		if err != nil {
			return nil, err
		}
		schema = retr_schema
	case p.Schema != "":
		schema = p.Schema
	default:
		err := fmt.Errorf("One of SchemaRegistry or Schema must be specified")
		return nil, err
	}

	codec, err := goavro.NewCodec(schema)

	if err != nil {
		return nil, err
	}

	native, _, err := codec.NativeFromBinary(binary_data)
	if err != nil {
		return nil, err
	}

	m, err := p.createMetric(native)
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
		return nil, fmt.Errorf("Line contains multiple metrics")
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
		return p.TimeFunc(), fmt.Errorf("Must specify timestamp format")
	}

	metricTime, err := internal.ParseTimestamp(p.TimestampFormat, timestamp, "UTC")
	if err != nil {
		return p.TimeFunc(), err
	}

	return metricTime, nil
}

func (p *Parser) createMetric(native interface{}) (telegraf.Metric, error) {
	deep, ok := native.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Cannot cast native interface {} to map[string]interface{}!")
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
			log.Printf("I! AvroParser: tag %s value was %v; not added to tags", tag, value)
		}
	}

	for _, field := range p.Fields {
		if value, ok := deep[field]; ok {
			fields[field] = nestedValue(value)
		} else {
			log.Printf("I! AvroParser: field %s value was %v; not added to fields", field, value)
		}
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("Number of fields is 0, unable to create metric!")
	}
	name := p.MetricName
	if p.Measurement != "" {
		name = p.Measurement
	}
	
	m := metric.New(name, tags, fields, metricTime)

	if m == nil {
		err := fmt.Errorf("Could not create metric")
		return nil, err
	}

	return m, nil
}

func nestedValue(deep interface{}) interface{} {
	if m, ok := deep.(map[string]interface{}); ok {
		for _, value := range m {
			return nestedValue(value)
		}
	}
	return deep
}
