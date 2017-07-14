package msgpack

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/vmihailenco/msgpack"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type MsgpackParser struct {
	MetricName  string
	DefaultTags map[string]string
}

func (p *MsgpackParser) parseObject(metrics []telegraf.Metric, msgpackOut map[string]interface{}) ([]telegraf.Metric, error) {
	tags := make(map[string]string)
	for k, v := range p.DefaultTags {
		tags[k] = v
	}

	f := MsgpackFlatterner{}
	err := f.FlattenMsgpack("", msgpackOut)
	if err != nil {
		return nil, err
	}

	metric, err := metric.New(p.MetricName, tags, f.Fields, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	return append(metrics, metric), nil
}

func (p *MsgpackParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 0)
	var msgpackOut map[string]interface{}
	err := msgpack.Unmarshal(buf, &msgpackOut)

	if err != nil {
		err = fmt.Errorf("unable to parse out as Msgpack, %s", err)
		return nil, err
	}
	return p.parseObject(metrics, msgpackOut)
}

func (p *MsgpackParser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line + "\n"))

	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, fmt.Errorf("Can not parse the line: %s, for data format: influx ", line)
	}

	return metrics[0], nil
}

func (p *MsgpackParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

type MsgpackFlatterner struct {
	Fields map[string]interface{}
}

func (f *MsgpackFlatterner) FlattenMsgpack(fieldname string, v interface{}) error {
	if f.Fields == nil {
		f.Fields = make(map[string]interface{})
	}
	return f.FullFlattenMsgpack(fieldname, v, true, true)
}

func (f *MsgpackFlatterner) FullFlattenMsgpack(fieldname string, v interface{}, convertString bool, convertBool bool) error {
	if f.Fields == nil {
		f.Fields = make(map[string]interface{})
	}
	fieldname = strings.Trim(fieldname, "_")
	switch t := v.(type) {
	case map[string]interface{}:
		for k, v := range t {
			err := f.FullFlattenMsgpack(fieldname+"_"+k+"_", v, convertString, convertBool)
			if err != nil {
				return err
			}
		}
	case []interface{}:
		for i, v := range t {
			k := strconv.Itoa(i)
			err := f.FullFlattenMsgpack(fieldname+"_"+k+"_", v, convertString, convertBool)
			if err != nil {
				return nil
			}
		}
	case float32, float64:
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
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		f.Fields[fieldname] = reflect.ValueOf(t)
	case nil:
		// ignored types
		fmt.Println("json parser ignoring " + fieldname)
		return nil
	default:
		return fmt.Errorf("Msgpack Flattener: got unexpected type %T with value %v (%s)", t, t, fieldname)
	}
	return nil
}
