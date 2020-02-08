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
const defaultStringEncoding = "UTF-8"

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
	byteOrder      binary.ByteOrder
	StringEncoding string
	Fields         []Field
	DefaultTags    map[string]string
}

func NewBinDataParser(
	metricName string,
	timeFormat string,
	endiannes string,
	stringEncoding string,
	fields []Field,
	defaultTags map[string]string,
) (*BinData, error) {

	// Time format
	switch timeFormat {
	case "unix", "unix_ms", "unix_us", "unix_ns":
	default:
		return nil, fmt.Errorf("invalid time format %s", timeFormat)
	}

	// Endiannes
	var byteOrder binary.ByteOrder
	endiannes = strings.ToLower(endiannes)
	if endiannes == "" || endiannes == "be" {
		byteOrder = binary.BigEndian
	} else if endiannes == "le" {
		byteOrder = binary.LittleEndian
	} else {
		return nil, fmt.Errorf("invalid bindata_endiannes %s", endiannes)
	}

	// String encoding
	if stringEncoding == "" {
		stringEncoding = defaultStringEncoding
	}
	stringEncoding = strings.ToUpper(stringEncoding)
	if stringEncoding != defaultStringEncoding {
		return nil, fmt.Errorf(`invalid string encoding %s`, stringEncoding)
	}

	// Fields' sizes
	for i := 0; i < len(fields); i++ {
		fieldType, ok := fieldTypes[strings.ToLower(fields[i].Type)]
		if !ok {
			return nil, fmt.Errorf(`invalid field type %s`, fields[i].Type)
		}
		// Overwrite non-string and non-padding field size
		if fieldType.Name() != "string" && fields[i].Type != "padding" {
			fields[i].Size = uint(fieldType.Size())
		}
	}

	return &BinData{
		MetricName:     metricName,
		TimeFormat:     timeFormat,
		Endiannes:      endiannes,
		byteOrder:      byteOrder,
		StringEncoding: stringEncoding,
		Fields:         fields,
		DefaultTags:    defaultTags,
	}, nil
}

// SetDefaultTags is ...
func (binData *BinData) SetDefaultTags(tags map[string]string) {
	binData.DefaultTags = tags
}

// Parse is ...
func (binData *BinData) Parse(data []byte) ([]telegraf.Metric, error) {

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
				binary.Read(byteReader, binData.byteOrder, fieldValue.Interface())
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
