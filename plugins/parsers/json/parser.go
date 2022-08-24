package json

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/tidwall/gjson"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
)

var (
	utf8BOM      = []byte("\xef\xbb\xbf")
	ErrWrongType = errors.New("must be an object or an array of objects")
)

type Parser struct {
	metricName   string
	tagKeys      filter.Filter
	stringFields filter.Filter
	nameKey      string
	query        string
	timeKey      string
	timeFormat   string
	timezone     string
	defaultTags  map[string]string
	strict       bool

	Log telegraf.Logger `toml:"-"`
}

func New(config *Config) (*Parser, error) {
	stringFilter, err := filter.Compile(config.StringFields)
	if err != nil {
		return nil, err
	}

	tagKeyFilter, err := filter.Compile(config.TagKeys)
	if err != nil {
		return nil, err
	}

	return &Parser{
		metricName:   config.MetricName,
		tagKeys:      tagKeyFilter,
		nameKey:      config.NameKey,
		stringFields: stringFilter,
		query:        config.Query,
		timeKey:      config.TimeKey,
		timeFormat:   config.TimeFormat,
		timezone:     config.Timezone,
		defaultTags:  config.DefaultTags,
		strict:       config.Strict,
	}, nil
}

func (p *Parser) parseArray(data []interface{}, timestamp time.Time) ([]telegraf.Metric, error) {
	results := make([]telegraf.Metric, 0)

	for _, item := range data {
		switch v := item.(type) {
		case map[string]interface{}:
			metrics, err := p.parseObject(v, timestamp)
			if err != nil {
				if p.Strict {
					return nil, err
				}
				continue
			}
			results = append(results, metrics...)
		default:
			return nil, ErrWrongType
		}
	}

	return results, nil
}

func (p *Parser) parseObject(data map[string]interface{}, timestamp time.Time) ([]telegraf.Metric, error) {
	tags := make(map[string]string)
	for k, v := range p.DefaultTags {
		tags[k] = v
	}

	f := JSONFlattener{}
	err := f.FullFlattenJSON("", data, true, true)
	if err != nil {
		return nil, err
	}

	name := p.MetricName

	// checks if json_name_key is set
	if p.nameKey != "" {
		if field, ok := f.Fields[p.nameKey].(string); ok {
			name = field
		}
	}

	// if time key is specified, set timestamp to it
	if p.timeKey != "" {
		if p.timeFormat == "" {
			err := fmt.Errorf("use of 'json_time_key' requires 'json_time_format'")
			return nil, err
		}

		if f.Fields[p.TimeKey] == nil {
			err := fmt.Errorf("JSON time key could not be found")
			return nil, err
		}

		timestamp, err = internal.ParseTimestamp(p.TimeFormat, f.Fields[p.TimeKey], p.Timezone)
		if err != nil {
			return nil, err
		}

		delete(f.Fields, p.TimeKey)

		// if the year is 0, set to current year
		if timestamp.Year() == 0 {
			timestamp = timestamp.AddDate(time.Now().Year(), 0, 0)
		}
	}

	tags, nFields := p.switchFieldToTag(tags, f.Fields)
	m := metric.New(name, tags, nFields, timestamp)

	return []telegraf.Metric{m}, nil
}

// will take in field map with strings and bools,
// search for TagKeys that match fieldnames and add them to tags
// will delete any strings/bools that shouldn't be fields
// assumes that any non-numeric values in TagKeys should be displayed as tags
func (p *Parser) switchFieldToTag(tags map[string]string, fields map[string]interface{}) (map[string]string, map[string]interface{}) {
	for name, value := range fields {
		if p.tagKeys == nil {
			continue
		}
		// skip switch statement if tagkey doesn't match fieldname
		if !p.tagKeys.Match(name) {
			continue
		}
		// switch any fields in tagkeys into tags
		switch t := value.(type) {
		case string:
			tags[name] = t
			delete(fields, name)
		case bool:
			tags[name] = strconv.FormatBool(t)
			delete(fields, name)
		case float64:
			tags[name] = strconv.FormatFloat(t, 'f', -1, 64)
			delete(fields, name)
		default:
			p.Log.Errorf("Unrecognized type %T", value)
		}
	}

	// remove any additional string/bool values from fields
	for fk := range fields {
		switch fields[fk].(type) {
		case string, bool:
			if p.stringFilter != nil && p.stringFilter.Match(fk) {
				continue
			}
			delete(fields, fk)
		}
	}
	return tags, fields
}

func (p *Parser) Init() error {
	var err error

	p.stringFilter, err = filter.Compile(p.StringFields)
	if err != nil {
		return fmt.Errorf("compiling string-fields filter failed: %v", err)
	}

	p.tagFilter, err = filter.Compile(p.TagKeys)
	if err != nil {
		return fmt.Errorf("compiling tag-key filter failed: %v", err)
	}

	return nil
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	if p.Query != "" {
		result := gjson.GetBytes(buf, p.Query)
		buf = []byte(result.Raw)
		if !result.IsArray() && !result.IsObject() && result.Type != gjson.Null {
			err := fmt.Errorf("query path must lead to a JSON object, array of objects or null, but lead to: %v", result.Type)
			return nil, err
		}
		if result.Type == gjson.Null {
			return nil, nil
		}
	}

	buf = bytes.TrimSpace(buf)
	buf = bytes.TrimPrefix(buf, utf8BOM)
	if len(buf) == 0 {
		return make([]telegraf.Metric, 0), nil
	}

	var data interface{}
	err := json.Unmarshal(buf, &data)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().UTC()
	switch v := data.(type) {
	case map[string]interface{}:
		return p.parseObject(v, timestamp)
	case []interface{}:
		return p.parseArray(v, timestamp)
	case nil:
		return nil, nil
	default:
		return nil, ErrWrongType
	}
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line + "\n"))

	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, fmt.Errorf("can not parse the line: %s, for data format: json ", line)
	}

	return metrics[0], nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func init() {
	parsers.Add("json",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{MetricName: defaultMetricName}
		})
}

// FullFlattenJSON flattens nested maps/interfaces into a fields map (including bools and string)
func (f *JSONFlattener) FullFlattenJSON(
	fieldname string,
	v interface{},
	convertString bool,
	convertBool bool,
) error {
	if f.Fields == nil {
		f.Fields = make(map[string]interface{})
	}

	switch t := v.(type) {
	case map[string]interface{}:
		for k, v := range t {
			fieldkey := k
			if fieldname != "" {
				fieldkey = fieldname + "_" + fieldkey
			}

			err := f.FullFlattenJSON(fieldkey, v, convertString, convertBool)
			if err != nil {
				return err
			}
		}
	case []interface{}:
		for i, v := range t {
			fieldkey := strconv.Itoa(i)
			if fieldname != "" {
				fieldkey = fieldname + "_" + fieldkey
			}
			err := f.FullFlattenJSON(fieldkey, v, convertString, convertBool)
			if err != nil {
				return err
			}
		}
	case float64:
		f.Fields[fieldname] = t
	case string:
		if !convertString {
			return nil
		}
		f.Fields[fieldname] = v.(string)
	case bool:
		if !convertBool {
			return nil
		}
		f.Fields[fieldname] = v.(bool)
	case nil:
		return nil
	default:
		return fmt.Errorf("JSON Flattener: got unexpected type %T with value %v (%s)",
			t, t, fieldname)
	}
	return nil
}
