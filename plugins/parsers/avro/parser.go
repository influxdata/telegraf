package avro

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/linkedin/goavro"
)

type Parser struct {
	SchemaRegistry  string
	Measurement     string
	Tags            []string
	Fields          []string
	Timestamp       string
	TimestampFormat string
	DefaultTags     map[string]string
	TimeFunc        func() time.Time
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	schemaRegistry := NewSchemaRegistry(p.SchemaRegistry)

	schemaID := int(binary.BigEndian.Uint32(buf[1:5]))

	schema, err := schemaRegistry.getSchema(schemaID)

	if err != nil {
		log.Printf("E! AvroParser: %s", err)
		return nil, err
	}

	codec, err := goavro.NewCodec(schema)

	if err != nil {
		log.Printf("E! AvroParser: %s", err)
		return nil, err
	}

	native, _, err := codec.NativeFromBinary(buf[5:])
	if err != nil {
		log.Printf("E! AvroParser: %s", err)
		return nil, err
	}

	m, err := p.createMetric(native)
	if err != nil {
		log.Printf("E! AvroParser: %s", err)
		return nil, err
	}

	return []telegraf.Metric{m}, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		log.Printf("E! AvroParser: %s", err)
		return nil, err
	}

	if len(metrics) != 1 {
		log.Printf("E! AvroParser: %s", err)
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
		log.Printf("E! AvroParser: %s", err)
		return p.TimeFunc(), err
	}

	return metricTime, nil
}

func (p *Parser) createMetric(native interface{}) (telegraf.Metric, error) {
	deep, ok := native.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Cannot cast native to map[string]interface{}!")
	}

	metricTime, err := p.parseTimestamp(nestedValue(deep[p.Timestamp]))
	if err != nil {
		log.Printf("E! AvroParser: %s", err)
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
			log.Printf("I! AvroParser: tag %s null", tag)
		}
	}

	for _, field := range p.Fields {
		if value, ok := deep[field]; ok {
			fields[field] = nestedValue(value)
		} else {
			log.Printf("I! AvroParser: field %s null", field)
		}
	}

	if len(fields) == 0 {
		log.Printf("I! AvroParser: number of fields is 0, skipping metric")
		return nil, fmt.Errorf("Number of fields is 0, unable to create metric!")
	} else {
		m, err := metric.New(p.Measurement, tags, fields, metricTime)

		if err != nil {
			log.Printf("E! AvroParser: %s", err)
			return nil, err
		}

		return m, nil
	}
}

func nestedValue(deep interface{}) interface{} {
	if m, ok := deep.(map[string]interface{}); ok {
		for _, value := range m {
			return nestedValue(value)
		}
	}
	return deep
}
