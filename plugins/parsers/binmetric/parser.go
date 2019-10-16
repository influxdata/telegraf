package binmetric

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/toml"
)

// Field is ...
type Field struct {
	Name   string
	Type   string
	Offset int
	Size   int
}

// BinMetric is ...
type BinMetric struct {
	MetricName string
	// TagKeys    []string
	TimeFormat string
	Timezone   string
	Endiannes  string
	Fields     []Field
}

// Config is ...
type Config struct {
	BinMetric BinMetric
}

const timeKey = "time"
const timezone = "UTC"

var data string = `
[binmetric]
	metric_name = "drone_status"
	endiannes = "be"
	time_format = "unix"
	fields = [
		{name="version",type="uint16",offset=0,size=2},
		{name="time",type="int32",offset=2,size=4},
		{name="location_latitude",type="float64",offset=6,size=8},
		{name="location_longitude",type="float64",offset=14,size=8},
		{name="location_altitude",type="float32",offset=22,size=4},
		{name="orientation_heading",type="float32",offset=26,size=4},
		{name="orientation_elevation",type="float32",offset=30,size=4},
		{name="orientation_bank",type="float32",offset=34,size=4},
		{name="speed_ground",type="float32",offset=38,size=4},
		{name="speed_air",type="float32",offset=42,size=4},
		{name="is_healthy",type="bool",offset=46,size=1},
		{name="state",type="string",offset=47,size=8},
	]
`

var fieldTypes = map[string]reflect.Type{
	"bool":    reflect.TypeOf((*bool)(nil)).Elem(),
	"uint8":   reflect.TypeOf((*uint8)(nil)).Elem(),
	"int8":    reflect.TypeOf((*int8)(nil)).Elem(),
	"uint16":  reflect.TypeOf((*uint16)(nil)).Elem(),
	"int16":   reflect.TypeOf((*int16)(nil)).Elem(),
	"uint32":  reflect.TypeOf((*uint32)(nil)).Elem(),
	"int32":   reflect.TypeOf((*int32)(nil)).Elem(),
	"uint64":  reflect.TypeOf((*uint64)(nil)).Elem(),
	"int64":   reflect.TypeOf((*int64)(nil)).Elem(),
	"float32": reflect.TypeOf((*float32)(nil)).Elem(),
	"float64": reflect.TypeOf((*float64)(nil)).Elem(),
	"string":  reflect.TypeOf((*string)(nil)).Elem(),
}

// Parse is ...
func (p *BinMetric) Parse(binMetric []byte) ([]telegraf.Metric, error) {
	var config Config
	var err = toml.Unmarshal([]byte(data), &config)
	if err != nil {
		return nil, err
	}

	endiannes := strings.ToLower(config.BinMetric.Endiannes)
	if endiannes != "be" && endiannes != "le" {
		return nil, fmt.Errorf(`invalid endiannes "%s""`, endiannes)
	}

	fields := make(map[string]interface{})
	s := io.NewSectionReader(bytes.NewReader(binMetric), 0, int64(len(binMetric)))

	for _, field := range config.BinMetric.Fields {
		fieldBuffer := make([]byte, field.Size)
		_, err := s.ReadAt(fieldBuffer, int64(field.Offset))
		if err != nil {
			return nil, err
		}

		fieldType := fieldTypes[field.Type]
		if fieldType == nil {
			return nil, fmt.Errorf(`invalid field type "%s""`, field.Type)
		}

		if fieldType.Name() == "string" {
			fields[field.Name] = string(fieldBuffer)
		} else {
			fieldValue := reflect.New(fieldType)
			byteReader := bytes.NewReader(fieldBuffer)
			if endiannes == "be" {
				binary.Read(byteReader, binary.BigEndian, fieldValue.Interface())
			} else {
				binary.Read(byteReader, binary.LittleEndian, fieldValue.Interface())
			}
			fields[field.Name] = fieldValue.Elem().Interface()
		}
	}

	metricTime := time.Now().UTC()
	timeValue := fields[timeKey]
	if timeValue != nil {
		if config.BinMetric.TimeFormat == "" {
			return nil, fmt.Errorf(`use of "%s" field requires "binmetric_time_format"`, timeKey)
		}

		var timeValueInt64 int64 = int64(timeValue.(int32))
		metricTime, err = internal.ParseTimestamp(config.BinMetric.TimeFormat, timeValueInt64, timezone)
		if err != nil {
			return nil, err
		}

		delete(fields, timeKey)
	}

	// metricTags := make(map[string]string)
	// for _, tagKey := range config.BinMetric.TagKeys {
	// 	metricTags[tagKey] = ""
	// }

	metric, err := metric.New(config.BinMetric.MetricName, nil, fields, metricTime)
	if err != nil {
		return nil, err
	}
	return []telegraf.Metric{metric}, err
}
