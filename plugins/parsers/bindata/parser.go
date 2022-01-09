package bindata

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
const defaultTimeFormat = "unix"

// Field is a binary data field descriptor
type Field struct {
	Name string
	Type string
	Size uint
}

// BinData is a binary data parser
type BinData struct {
	metricName     string
	timeFormat     string
	endiannes      string
	byteOrder      binary.ByteOrder
	stringEncoding string
	fields         []Field
	DefaultTags    map[string]string
}

// Supported field types
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

// NewBinDataParser is BinData factory
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
	case "":
		timeFormat = defaultTimeFormat
	case "unix", "unix_ms", "unix_us", "unix_ns":
	default:
		return nil, fmt.Errorf("invalid time format %s", timeFormat)
	}

	// Endiannes
	var byteOrder binary.ByteOrder
	endiannes = strings.ToLower(endiannes)
	switch endiannes {
	case "", "be":
		byteOrder = binary.BigEndian
	case "le":
		byteOrder = binary.LittleEndian
	default:
		return nil, fmt.Errorf("invalid endiannes %q", endiannes)
	}

	// String encoding
	if stringEncoding == "" {
		stringEncoding = defaultStringEncoding
	}
	stringEncoding = strings.ToUpper(stringEncoding)
	if stringEncoding != defaultStringEncoding {
		return nil, fmt.Errorf(`invalid string encoding %s`, stringEncoding)
	}

	// Field types, names and sizes
	knownFields := make(map[string]bool)
	for i, field := range fields {
		fieldType, ok := fieldTypes[strings.ToLower(field.Type)]
		if !ok {
			return nil, fmt.Errorf(`invalid field type %s`, fields[i].Type)
		}

		if field.Type == "padding" {
			// Ignore padding fields
			continue
		}

		// Check for duplicate field names
		fieldName := field.Name
		if _, ok := knownFields[fieldName]; ok {
			return nil, fmt.Errorf(`duplicate field name %s`, fieldName)
		}
		knownFields[fieldName] = true

		// Time field type check
		if fieldName == "time" {
			switch timeFormat {
			case "unix":
				if field.Type != "int32" {
					return nil, fmt.Errorf(`invalid time type, must be int32`)
				}
			case "unix_ms", "unix_us", "unix_ns":
				if field.Type != "int64" {
					return nil, fmt.Errorf(`invalid time type, must be int64`)
				}
			}
		}

		// Overwrite non-string and non-padding field size
		if field.Type != "string" {
			fields[i].Size = uint(fieldType.Size())
		}
	}

	return &BinData{
		metricName:     metricName,
		timeFormat:     timeFormat,
		endiannes:      endiannes,
		byteOrder:      byteOrder,
		stringEncoding: stringEncoding,
		fields:         fields,
		DefaultTags:    defaultTags,
	}, nil
}

// SetDefaultTags implements Parser.SetDefaultTags()
func (binData *BinData) SetDefaultTags(tags map[string]string) {
	binData.DefaultTags = tags
}

// Parse implements Parser.Parse()
func (binData *BinData) Parse(data []byte) ([]telegraf.Metric, error) {

	fields := make(map[string]interface{})
	var offset uint = 0
	for _, field := range binData.fields {
		if offset > uint(len(data)) || offset+field.Size > uint(len(data)) {
			return nil, fmt.Errorf("invalid offset/size in field %s", field.Name)
		}
		if field.Type != "padding" {
			fieldBuffer := data[offset : offset+field.Size]
			switch field.Type {
			case "string":
				fields[field.Name] = string(fieldBuffer)
			default:
				fieldValue := reflect.New(fieldTypes[field.Type])
				byteReader := bytes.NewReader(fieldBuffer)
				binary.Read(byteReader, binData.byteOrder, fieldValue.Interface())
				fields[field.Name] = fieldValue.Elem().Interface()
			}
		}
		offset += field.Size
	}

	metricTime, err := binData.getTime(fields)
	if err != nil {
		return nil, err
	}

	metric, err := metric.New(binData.metricName, binData.DefaultTags, fields, metricTime)
	if err != nil {
		return nil, err
	}

	return []telegraf.Metric{metric}, err
}

// ParseLine implements Parser.ParseLine()
func (binData *BinData) ParseLine(line string) (telegraf.Metric, error) {
	return nil, fmt.Errorf("BinData.ParseLine() not supported")
}

func (binData *BinData) getTime(fields map[string]interface{}) (time.Time, error) {
	nilTime := new(time.Time)
	metricTime := time.Now()
	timeValue := fields[timeKey]
	if timeValue != nil {
		var err error
		switch binData.timeFormat {
		case "unix":
			if _, ok := timeValue.(int32); !ok {
				return *nilTime, fmt.Errorf("invalid time type, must be int32")
			}
			metricTime, err = internal.ParseTimestamp(binData.timeFormat, int64(timeValue.(int32)), timezone)
		case "unix_ms", "unix_us", "unix_ns":
			if _, ok := timeValue.(int64); !ok {
				return *nilTime, fmt.Errorf("invalid time type, must be int64")
			}
			metricTime, err = internal.ParseTimestamp(binData.timeFormat, int64(timeValue.(int64)), timezone)
		default:
			return *nilTime, fmt.Errorf("invalid time format %s", binData.timeFormat)
		}
		if err != nil {
			return *nilTime, err
		}
		delete(fields, timeKey)
	}
	return metricTime, nil
}
