package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/tidwall/gjson"
)

var (
	utf8BOM = []byte("\xef\xbb\xbf")
)

type JSONParser struct {
	MetricName     string
	TagKeys        []string
	StringFields   []string
	JSONNameKey    string
	JSONQuery      string
	JSONTimeKey    string
	JSONTimeFormat string
	DefaultTags    map[string]string
}

func (p *JSONParser) parseArray(buf []byte) ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 0)

	var jsonOut []map[string]interface{}
	err := json.Unmarshal(buf, &jsonOut)
	if err != nil {
		err = fmt.Errorf("unable to parse out as JSON Array, %s", err)
		return nil, err
	}
	for _, item := range jsonOut {
		metrics, err = p.parseObject(metrics, item)
		if err != nil {
			return nil, err
		}
	}
	return metrics, nil
}

func (p *JSONParser) parseObject(metrics []telegraf.Metric, jsonOut map[string]interface{}) ([]telegraf.Metric, error) {

	tags := make(map[string]string)
	for k, v := range p.DefaultTags {
		tags[k] = v
	}

	f := JSONFlattener{}
	err := f.FullFlattenJSON("", jsonOut, true, true)
	if err != nil {
		return nil, err
	}

	//checks if json_name_key is set
	if p.JSONNameKey != "" {
		p.MetricName = f.Fields[p.JSONNameKey].(string)
	}

	//if time key is specified, set it to nTime
	nTime := time.Now().UTC()
	if p.JSONTimeKey != "" {
		if p.JSONTimeFormat == "" {
			err := fmt.Errorf("use of 'json_time_key' requires 'json_time_format'")
			return nil, err
		}

		if f.Fields[p.JSONTimeKey] == nil {
			err := fmt.Errorf("JSON time key could not be found")
			return nil, err
		}

		timeStr, ok := f.Fields[p.JSONTimeKey].(string)
		if !ok {
			err := fmt.Errorf("time: %v could not be converted to string", f.Fields[p.JSONTimeKey])
			return nil, err
		}
		nTime, err = time.Parse(p.JSONTimeFormat, timeStr)
		if err != nil {
			return nil, err
		}

		//if the year is 0, set to current year
		if nTime.Year() == 0 {
			nTime = nTime.AddDate(time.Now().Year(), 0, 0)
		}
	}

	tags, nFields := p.switchFieldToTag(tags, f.Fields)
	metric, err := metric.New(p.MetricName, tags, nFields, nTime)
	if err != nil {
		return nil, err
	}
	return append(metrics, metric), nil
}

//will take in field map with strings and bools,
//search for TagKeys that match fieldnames and add them to tags
//will delete any strings/bools that shouldn't be fields
//assumes that any non-numeric values in TagKeys should be displayed as tags
func (p *JSONParser) switchFieldToTag(tags map[string]string, fields map[string]interface{}) (map[string]string, map[string]interface{}) {
	for _, name := range p.TagKeys {
		//switch any fields in tagkeys into tags
		if fields[name] == nil {
			continue
		}
		switch value := fields[name].(type) {
		case string:
			tags[name] = value
			delete(fields, name)
		case bool:
			tags[name] = strconv.FormatBool(value)
			delete(fields, name)
		case float64:
			tags[name] = strconv.FormatFloat(value, 'f', -1, 64)
			delete(fields, name)
		default:
			log.Printf("E! [parsers.json] Unrecognized type %T", value)
		}
	}

	//remove any additional string/bool values from fields
	for k := range fields {
		//check if field is in StringFields
		sField := false
		for _, v := range p.StringFields {
			if v == k {
				sField = true
			}
		}
		if sField {
			continue
		}

		switch fields[k].(type) {
		case string:
			delete(fields, k)
		case bool:
			delete(fields, k)
		}
	}
	return tags, fields
}

func (p *JSONParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	if p.JSONQuery != "" {
		result := gjson.GetBytes(buf, p.JSONQuery)
		buf = []byte(result.Raw)
		if !result.IsArray() && !result.IsObject() {
			err := fmt.Errorf("E! Query path must lead to a JSON object or array of objects, but lead to: %v", result.Type)
			return nil, err
		}
	}

	buf = bytes.TrimSpace(buf)
	buf = bytes.TrimPrefix(buf, utf8BOM)
	if len(buf) == 0 {
		return make([]telegraf.Metric, 0), nil
	}

	if !isarray(buf) {
		metrics := make([]telegraf.Metric, 0)
		var jsonOut map[string]interface{}
		err := json.Unmarshal(buf, &jsonOut)
		if err != nil {
			err = fmt.Errorf("unable to parse out as JSON, %s", err)
			return nil, err
		}
		return p.parseObject(metrics, jsonOut)
	}
	return p.parseArray(buf)
}

func (p *JSONParser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line + "\n"))

	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, fmt.Errorf("can not parse the line: %s, for data format: json ", line)
	}

	return metrics[0], nil
}

func (p *JSONParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

type JSONFlattener struct {
	Fields map[string]interface{}
}

// FlattenJSON flattens nested maps/interfaces into a fields map (ignoring bools and string)
func (f *JSONFlattener) FlattenJSON(
	fieldname string,
	v interface{}) error {
	if f.Fields == nil {
		f.Fields = make(map[string]interface{})
	}

	return f.FullFlattenJSON(fieldname, v, false, false)
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
	fieldname = strings.Trim(fieldname, "_")
	switch t := v.(type) {
	case map[string]interface{}:
		for k, v := range t {
			err := f.FullFlattenJSON(fieldname+"_"+k+"_", v, convertString, convertBool)
			if err != nil {
				return err
			}
		}
	case []interface{}:
		for i, v := range t {
			k := strconv.Itoa(i)
			err := f.FullFlattenJSON(fieldname+"_"+k+"_", v, convertString, convertBool)
			if err != nil {
				return nil
			}
		}
	case float64:
		f.Fields[fieldname] = t
	case string:
		if convertString {
			f.Fields[fieldname] = v.(string)
		} else {
			return nil
		}
	case bool:
		if convertBool {
			f.Fields[fieldname] = v.(bool)
		} else {
			return nil
		}
	case nil:
		return nil
	default:
		return fmt.Errorf("JSON Flattener: got unexpected type %T with value %v (%s)",
			t, t, fieldname)
	}
	return nil
}

func isarray(buf []byte) bool {
	ia := bytes.IndexByte(buf, '[')
	ib := bytes.IndexByte(buf, '{')
	if ia > -1 && ia < ib {
		return true
	} else {
		return false
	}
}
