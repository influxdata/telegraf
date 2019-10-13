package binmetric

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/toml"
)

// Field is ...
type Field struct {
	Name   string
	Type   string
	Offset int
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

var data string = `
[BinMetric]
	metric_name = "drone_status"
	endiannes = "be"
	fields = [
		{name="version",type="uint16",offset=0},
		{name="time",type="int32",offset=2},
		{name="location_latitude",type="float64",offset=6},
		{name="location_longitude",type="float64",offset=14},
		{name="location_altitude",type="float32",offset=22},
		{name="orientation_heading",type="float32",offset=26},
		{name="orientation_elevation",type="float32",offset=30},
		{name="orientation_bank",type="float32",offset=34},
		{name="speed_ground",type="float32",offset=38},
		{name="speed_air",type="float32",offset=42},
		{name="is_healthy",type="bool",offset=46},
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
}

// Parse is ...
func (p *BinMetric) Parse(BinMetric []byte) ([]telegraf.Metric, error) {
	var config Config
	var _ = toml.Unmarshal([]byte(data), &config)

	endiannes := strings.ToLower(config.BinMetric.Endiannes)
	fields := make(map[string]interface{})

	byteReader := bytes.NewReader(BinMetric)

	for _, field := range config.BinMetric.Fields {
		fieldType := fieldTypes[field.Type]
		fieldValue := reflect.New(fieldType)
		if endiannes == "be" {
			binary.Read(byteReader, binary.BigEndian, fieldValue.Interface())
		} else {
			binary.Read(byteReader, binary.LittleEndian, fieldValue.Interface())
		}
		fields[field.Name] = fieldValue.Elem().Interface()
	}

	timeValue := fields["time"]
	metricTime := time.Now().UTC()
	if timeValue != nil {
		metricTime = time.Unix(int64(timeValue.(int32)), 0).UTC()
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
