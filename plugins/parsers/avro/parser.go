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
	MetricName       string            `toml:"metric_name"`
	SchemaRegistry   string            `toml:"avro_schema_registry"`
	Schema           string            `toml:"avro_schema"`
	Measurement      string            `toml:"avro_measurement"`
	Tags             []string          `toml:"avro_tags"`
	Fields           []string          `toml:"avro_fields"`
	Timestamp        string            `toml:"avro_timestamp"`
	TimestampFormat  string            `toml:"avro_timestamp_format"`
	DiscardArrays    bool              `toml:"avro_discard_arrays"`
	FieldSeparator   string            `toml:"avro_field_separator"`
	RoundTimestampTo string            `toml:"avro_round_timestamp_to"`
	DefaultTags      map[string]string `toml:"-"`

	Log         telegraf.Logger `toml:"-"`
	registryObj *SchemaRegistry
	createMetric func(interface{}, string) (telegraf.Metric, error)
}

type metricInput struct {
	Name string
	Tags map[string]string
	Fields map[string]interface{}
	Timestamp time.Time
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
	m, err := p.createMetric(native, schema)
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

func (p *Parser) setupMetric(native interface{}, schema string) (metricInput, error) {
	badReturn := metricInput{
		Name: "",
		Tags: nil,
		Fields: nil,
		Timestamp: time.Time{},
	}
			
	deep, ok := native.(map[string]interface{})
	if !ok {
		return badReturn, fmt.Errorf("cannot cast native interface {} to map[string]interface{}")
	}

	timestamp := nestedValue(deep[p.Timestamp])
	if timestamp == nil {
		timestamp = time.Now()
	}

	metricTime, err := internal.ParseTimestamp(p.TimestampFormat, timestamp, "UTC")
	if err != nil {
		return badReturn, err
	}

	if p.RoundTimestampTo != "" {
		// If we're still using this in 2262, it's gonna break.
		nanos := metricTime.UnixNano()
		if p.RoundTimestampTo == "s" {
			nanos = metricTime.Unix() * 1e9
		}
		if p.RoundTimestampTo == "ms" {
			nanos = metricTime.UnixMilli() * 1e6
		}
		if p.RoundTimestampTo == "us" {
			nanos = metricTime.UnixMicro() * 1e3
		}
		metricTime = time.Unix(0, nanos)
	}


	// Tags differ from fields, in that tags are inherently strings.
	// fields can be of any type.
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	for k, v := range p.DefaultTags {
		tags[k] = v
	}
	for _, tag := range p.Tags {
		if value, ok := deep[tag]; ok {
			tags[tag], err = internal.ToString(nestedValue(value))
			if err != nil {
				p.Log.Warnf("Could not convert %v to string", nestedValue(value))
			}
		} else {
			// It wouldn't unpack.  Probably not fatal, but does
			// mean we can't get this column.  This should
			// be very rare, since the tag is a string.
			p.Log.Warnf("tag %s value was %v; not added to tags", tag, value)
		}
	}
	fieldList := make([]string, len(p.Fields), (cap(p.Fields)))
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
			// It wouldn't unpack.  Probably not fatal, but does
			// mean we can't get this column.
			p.Log.Warnf("field %s value was %v; not added to fields", field, value)
		}
	}
	var schemaObj map[string]interface{}

	err = json.Unmarshal([]byte(schema), &schemaObj)
	if err != nil {
		return badReturn, err
	}
	if len(fields) == 0 {
		// A telegraf metric needs at least one field.
		return badReturn, fmt.Errorf("number of fields is 0; unable to create metric")
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
			return badReturn, fmt.Errorf("could not determine name from schema %s", schema)
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
		return badReturn, fmt.Errorf("could not determine measurement name")
	}
	return metricInput{
		Name: name,
		Tags: tags,
		Fields: fields,
		Timestamp: metricTime,
	}, nil
}

func (p *Parser) createScalarMetric(native interface{}, schema string) (telegraf.Metric, error) {
	m, err := p.setupMetric(native, schema)
	if err != nil {
		return nil, err
	}
	return metric.New(m.Name, m.Tags, m.Fields, m.Timestamp), nil
}

func (p *Parser) createComplexMetric(native interface{}, schema string) (telegraf.Metric, error) {
	m, err := p.setupMetric(native, schema)
	if err != nil {
		return nil, err
	}
	// The default (the separator string is empty) is equivalent to
	// what streamreactor does.
	sep := flatten.SeparatorStyle{
		Before: "",
		Middle: p.FieldSeparator,
		After:  "",
	}
	flat, err := flatten.Flatten(m.Fields, "", sep)
	if err != nil {
		return nil, err
	}
	return metric.New(m.Name, m.Tags, flat, m.Timestamp), nil
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
	if (p.Schema == "" && p.SchemaRegistry == "") || (p.Schema != "" && p.SchemaRegistry != "") {
		return fmt.Errorf("exactly one of 'schema_registry' or 'schema' must be specified")
	}
	if p.SchemaRegistry != "" {
		p.registryObj = NewSchemaRegistry(p.SchemaRegistry)
	}
	if p.TimestampFormat == "" {
		return fmt.Errorf("must specify 'timestamp_format'")
	}
	if p.RoundTimestampTo != "" && p.TimestampFormat != "unix" {
		return fmt.Errorf("'round_timestamp_to' can only be used with 'timestamp_format' of 'unix'")
	}
	if p.RoundTimestampTo != "" && (p.RoundTimestampTo != "s" && p.RoundTimestampTo != "ms" && p.RoundTimestampTo != "us") {
		return fmt.Errorf("'round_timestamp_to' must be one of 's', 'ms', or 'us'")
	}

	p.createMetric = p.createComplexMetric
	if p.DiscardArrays {
		p.createMetric = p.createScalarMetric
	}
	return nil
}
