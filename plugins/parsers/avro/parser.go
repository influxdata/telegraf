package avro

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	MetricName       string            `toml:"metric_name"`
	SchemaRegistry   string            `toml:"avro_schema_registry"`
	CaCertPath       string            `toml:"avro_schema_registry_cert"`
	Schema           string            `toml:"avro_schema"`
	Format           string            `toml:"avro_format"`
	Measurement      string            `toml:"avro_measurement"`
	MeasurementField string            `toml:"avro_measurement_field"`
	Tags             []string          `toml:"avro_tags"`
	Fields           []string          `toml:"avro_fields"`
	Timestamp        string            `toml:"avro_timestamp"`
	TimestampFormat  string            `toml:"avro_timestamp_format"`
	FieldSeparator   string            `toml:"avro_field_separator"`
	UnionMode        string            `toml:"avro_union_mode"`
	IncludeIndexTag  bool              `toml:"avro_include_index_tag"`
	DefaultTags      map[string]string `toml:"tags"`
	Log              telegraf.Logger   `toml:"-"`
	registryObj      *schemaRegistry
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
	switch p.UnionMode {
	case "":
		p.UnionMode = "flatten"
	case "flatten", "nullable", "any":
		// Do nothing as those are valid settings
	default:
		return fmt.Errorf("unknown avro_union_mode %q", p.Format)
	}

	if (p.Schema == "" && p.SchemaRegistry == "") || (p.Schema != "" && p.SchemaRegistry != "") {
		return errors.New("exactly one of 'schema_registry' or 'schema' must be specified")
	}
	switch p.TimestampFormat {
	case "":
		p.TimestampFormat = "unix"
	case "unix", "unix_ns", "unix_us", "unix_ms":
		// Valid values
	default:
		return fmt.Errorf("invalid timestamp format '%v'", p.TimestampFormat)
	}

	if p.SchemaRegistry != "" {
		registry, err := newSchemaRegistry(p.SchemaRegistry, p.CaCertPath)
		if err != nil {
			return fmt.Errorf("error connecting to the schema registry %q: %w", p.SchemaRegistry, err)
		}
		p.registryObj = registry
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

	// Handle array at root level -> produce multiple metrics
	if arrayData, ok := native.([]interface{}); ok {
		var metrics []telegraf.Metric
		for idx, item := range arrayData {
			record, ok := item.(map[string]interface{})
			if !ok {
				// skip non-record elements
				continue
			}
			m, err := p.createMetric(record, schema)
			if err != nil {
				// skip items that cannot be converted
				continue
			}
			if p.IncludeIndexTag {
				tags := m.Tags()
				tags["array_index"] = strconv.Itoa(idx)
				m = metric.New(m.Name(), tags, m.Fields(), m.Time())
			}
			metrics = append(metrics, m)
		}
		return metrics, nil
	}

	// Single record at root
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

func (p *Parser) flattenField(fldName string, fldVal map[string]interface{}) map[string]interface{} {
	// Helper function for the "nullable" and "any" p.UnionModes
	// fldVal is a one-item map of string-to-something
	ret := make(map[string]interface{})
	if p.UnionMode == "nullable" {
		_, ok := fldVal["null"]
		if ok {
			return ret // Return the empty map
		}
	}
	// Otherwise, we just return the value in the fieldname.
	// See README.md for an important warning about "any" and "nullable".
	for _, v := range fldVal {
		ret[fldName] = v
		break // Not really needed, since it's a one-item map
	}
	return ret
}

func (p *Parser) flattenItem(fld string, fldVal interface{}) (map[string]interface{}, error) {
	sep := flatten.SeparatorStyle{
		Before: "",
		Middle: p.FieldSeparator,
		After:  "",
	}
	candidate := make(map[string]interface{})
	candidate[fld] = fldVal

	var flat map[string]interface{}
	var err error
	// Exactly how we flatten is decided by p.UnionMode
	if p.UnionMode == "flatten" {
		flat, err = flatten.Flatten(candidate, "", sep)
		if err != nil {
			return nil, fmt.Errorf("flatten candidate %q failed: %w", candidate, err)
		}
		return flat, nil
	}

	// nullable or any
	if typedVal, ok := candidate[fld].(map[string]interface{}); ok {
		return p.flattenField(fld, typedVal), nil
	}
	flat, err := flatten.Flatten(candidate, "", sep)
	if err != nil {
		return nil, fmt.Errorf("flatten candidate %q failed: %w", candidate, err)
	}
	return flat, nil
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func (p *Parser) createMetric(data map[string]interface{}, schema string) (telegraf.Metric, error) {
	// fields can be any type; tags must be strings
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	// default tags
	for k, v := range p.DefaultTags {
		tags[k] = v
	}

	// Collect tags from configured avro_tags and remove them from fields
	for _, tag := range p.Tags {
		flat, err := p.flattenItem(tag, data[tag])
		if err != nil {
			return nil, fmt.Errorf("flatten tag %q failed: %w", tag, err)
		}
		for k, v := range flat {
			sTag, err := internal.ToString(v)
			if err != nil {
				p.Log.Warnf("Could not convert %v to string for tag %q: %v", data[tag], tag, err)
				continue
			}
			tags[k] = sTag
		}
	}

	// Determine fields list.
	var fieldList []string
	if len(p.Fields) != 0 {
		fieldList = p.Fields
	} else {
		for k := range data {
			// exclude plain tag field names
			if containsString(p.Tags, k) {
				continue
			}
			fieldList = append(fieldList, k)
		}
	}

	// Flatten and collect fields
	for _, fld := range fieldList {
		flat, err := p.flattenItem(fld, data[fld])
		if err != nil {
			return nil, fmt.Errorf("flatten field %q failed: %w", fld, err)
		}
		for k, v := range flat {
			fields[k] = v
		}
	}

	// Must have at least one field
	if len(fields) == 0 {
		return nil, errors.New("number of fields is 0; unable to create metric")
	}

	// Measurement name resolution
	name := ""
	// If MeasurementField specified, use it (and remove its flattened entries from fields/tags)
	if p.MeasurementField != "" {
		if val, ok := data[p.MeasurementField]; ok {
			if sMetric, err := internal.ToString(val); err == nil {
				name = sMetric
			} else {
				p.Log.Warnf("Could not convert %v to string for metric name %q: %s", data[p.MeasurementField], p.MeasurementField, err.Error())
			}
			// remove flattened keys originating from measurement field
			for fk := range fields {
				if fk == p.MeasurementField || strings.HasPrefix(fk, p.MeasurementField+p.FieldSeparator) {
					delete(fields, fk)
				}
			}
			for tk := range tags {
				if tk == p.MeasurementField || strings.HasPrefix(tk, p.MeasurementField+p.FieldSeparator) {
					delete(tags, tk)
				}
			}
		}
	}
	if name == "" {
		name = p.Measurement
	}
	if name == "" {
		// try schema namespace/name
		var schemaObj map[string]interface{}
		if err := json.Unmarshal([]byte(schema), &schemaObj); err == nil {
			nsStr, _ := schemaObj["namespace"].(string)
			nStr, _ := schemaObj["name"].(string)
			sep := "."
			if nsStr == "" {
				sep = ""
			}
			name = nsStr + sep + nStr
		}
	}
	if name == "" {
		name = p.MetricName
	}
	if name == "" {
		return nil, errors.New("could not determine measurement name")
	}

	// Timestamp handling:
	var timestamp time.Time
	if p.Timestamp != "" {
		// If timestamp field is present, parse and use it
		if rawVal, ok := data[p.Timestamp]; ok {
			rawTime := fmt.Sprintf("%v", rawVal)
			var err error
			timestamp, err = internal.ParseTimestamp(p.TimestampFormat, rawTime, nil)
			if err != nil {
				return nil, fmt.Errorf("could not parse '%s' to '%s': %w", rawTime, p.TimestampFormat, err)
			}
			// Retain timestamp field as a normal field (no deletion here)
		} else {
			// timestamp config set but not present in data -> fallback to now
			timestamp = time.Now()
		}
	} else {
		// No timestamp configured -> use now
		timestamp = time.Now()
	}

	return metric.New(name, tags, fields, timestamp), nil
}

func init() {
	parsers.Add("avro", func(defaultMetricName string) telegraf.Parser {
		return &Parser{MetricName: defaultMetricName}
	})
}
