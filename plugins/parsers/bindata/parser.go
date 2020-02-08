package bindata

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
)

const timeKey = "time"
const timezone = "UTC"
const stringEncoding = "UTF-8"

// Field is ...
type Field struct {
	Name string
	Type string
	Size uint
}

// BinData is ...
type BinData struct {
	MetricName     string
	TimeFormat     string
	Endiannes      string
	StringEncoding string
	Fields         []Field
	DefaultTags    map[string]string
}

// Parse is ...
func (binData *BinData) Parse(data []byte) ([]telegraf.Metric, error) {

	endiannes, err := binData.getEndiannes()
	if err != nil {
		return nil, err
	}

	// Validate
	err = binData.validate()
	if err != nil {
		return nil, err
	}

	fields := make(map[string]interface{})
	reader := io.NewSectionReader(bytes.NewReader(data), 0, int64(len(data)))
	var offset int64 = 0

	for _, field := range binData.Fields {
		if field.Type != "padding" {
			fieldBuffer := make([]byte, field.Size)

			if _, err := reader.ReadAt(fieldBuffer, offset); err != nil {
				return nil, err
			}

			fieldType, ok := fieldTypes[strings.ToLower(field.Type)]
			if !ok {
				return nil, fmt.Errorf(`invalid field type %s`, field.Type)
			}

			switch fieldType.Name() {
			case "string":
				fields[field.Name] = string(fieldBuffer)
			default:
				fieldValue := reflect.New(fieldType)
				byteReader := bytes.NewReader(fieldBuffer)
				binary.Read(byteReader, endiannes, fieldValue.Interface())
				fields[field.Name] = fieldValue.Elem().Interface()
			}
		}
		offset += int64(field.Size)
	}

	metricTime, err := binData.getTime(fields)
	if err != nil {
		return nil, err
	}

	metric, err := metric.New(binData.MetricName, binData.DefaultTags,
		fields, metricTime)
	if err != nil {
		return nil, err
	}

	return []telegraf.Metric{metric}, err
}

// ParseLine is ...
func (binData *BinData) ParseLine(line string) (telegraf.Metric, error) {
	return nil, fmt.Errorf("BinData.ParseLine() not supported")
}

// SetDefaultTags is ...
func (binData *BinData) SetDefaultTags(tags map[string]string) {
	binData.DefaultTags = tags
}

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
	"padding": reflect.TypeOf((*[]byte)(nil)).Elem(),
}

func (binData *BinData) validate() error {
	if binData.StringEncoding == "" {
		binData.StringEncoding = stringEncoding
	}
	binData.StringEncoding = strings.ToUpper(binData.StringEncoding)
	if binData.StringEncoding != stringEncoding {
		return fmt.Errorf(`invalid string encoding %s`, binData.StringEncoding)
	}

	for i := 0; i < len(binData.Fields); i++ {
		fieldType, ok := fieldTypes[strings.ToLower(binData.Fields[i].Type)]
		if !ok {
			return fmt.Errorf(`invalid field type %s`, binData.Fields[i].Type)
		}
		// Overwrite non-string and non-padding field size
		if fieldType.Name() != "string" && binData.Fields[i].Type != "padding" {
			binData.Fields[i].Size = uint(fieldType.Size())
		}
	}
	return nil
}

func (binData *BinData) getEndiannes() (binary.ByteOrder, error) {
	var endiannes binary.ByteOrder
	cfgEndiannes := strings.ToLower(binData.Endiannes)
	if cfgEndiannes == "" || cfgEndiannes == "be" {
		endiannes = binary.BigEndian
	} else if cfgEndiannes == "le" {
		endiannes = binary.LittleEndian
	} else {
		return nil, fmt.Errorf("invalid bindata_endiannes %s", cfgEndiannes)
	}
	return endiannes, nil
}

func (binData *BinData) getTime(fields map[string]interface{}) (time.Time, error) {
	nilTime := new(time.Time)
	metricTime := time.Now()
	timeValue := fields[timeKey]
	if timeValue != nil {
		var err error
		switch binData.TimeFormat {
		case "unix":
			metricTime, err = internal.ParseTimestamp(binData.TimeFormat, int64(timeValue.(int32)), timezone)
		case "unix_ms", "unix_us", "unix_ns":
			metricTime, err = internal.ParseTimestamp(binData.TimeFormat, int64(timeValue.(int64)), timezone)
		default:
			return *nilTime, fmt.Errorf("invalid time format %s", binData.TimeFormat)
		}
		if err != nil {
			return *nilTime, err
		}
		delete(fields, timeKey)
	}
	return metricTime, nil
}
