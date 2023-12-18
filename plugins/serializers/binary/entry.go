package binary

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf/internal"
	"golang.org/x/exp/constraints"
)

type Entry struct {
	ReadFrom         string `toml:"read_from"`         // field, tag, time, name
	Name             string `toml:"name"`              // name of entry
	DataFormat       string `toml:"data_format"`       // int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64, string
	StringTerminator string `toml:"string_terminator"` // for string metrics: null, 0x00, 00, ....
	StringLength     uint64 `toml:"string_length"`     // for string only, target size
	TimeFormat       string `toml:"time_format"`       // for time metrics: unix, unix_ms, unix_us, unix_ns

	bytes       uint64
	termination byte
}

func (e *Entry) FillDefaults() error {
	// Normalize
	e.ReadFrom = strings.ToLower(e.ReadFrom)

	if e.ReadFrom == "" {
		e.ReadFrom = "field"
	}

	switch e.ReadFrom {
	case "field", "tag":
		if e.Name == "" {
			return errors.New("missing name")
		}
	case "time":
		switch e.TimeFormat {
		// 'unix' the default
		case "":
			e.TimeFormat = "unix"
		// Plugin specific names
		case "unix", "unix_ms", "unix_us", "unix_ns":
		default:
			return errors.New("invalid time format")
		}
	case "name":
	default:
		return fmt.Errorf("unknown assignment %q", e.ReadFrom)
	}

	// Check data format
	switch e.DataFormat {
	case "":
		return errors.New("missing data format")
	case "int64", "uint64", "float64":
		e.bytes = 8
	case "int32", "uint32", "float32":
		e.bytes = 4
	case "int16", "uint16":
		e.bytes = 2
	case "int8", "uint8":
		e.bytes = 1
	case "string":
		if e.StringLength < 1 {
			return errors.New("string length must be at least 1")
		}

		e.bytes = e.StringLength
	}

	// Check string terminator
	switch e.StringTerminator {
	case "", "null":
		e.termination = 0x00
	default:
		terminatorLength := len(e.StringTerminator)

		// support both 0xXX and XX
		termination, err := hex.DecodeString(e.StringTerminator[terminatorLength-2:])

		if err != nil {
			return fmt.Errorf("decoding terminator failed for %q: %w", e.Name, err)
		}

		if len(termination) != 1 {
			return fmt.Errorf("terminator must be a single byte, got %q", e.StringTerminator)
		}

		e.termination = termination[0]
	}

	return nil
}

func (e *Entry) timeToTimestamp(t time.Time) int64 {
	var timestamp int64

	switch e.TimeFormat {
	case "unix":
		timestamp = t.Unix()
	case "unix_ms":
		timestamp = t.UnixMilli()
	case "unix_us":
		timestamp = t.UnixMicro()
	case "unix_ns":
		timestamp = t.UnixNano()
	}

	return timestamp
}

func appendToSlice[T any](arr []T, targetLength uint64, value T) []T {
	for uint64(len(arr)) < targetLength {
		arr = append(arr, value)
	}

	return arr
}

func (e *Entry) SerializeValue(value interface{}, converter binary.ByteOrder) ([]byte, error) {
	if e.DataFormat != "string" {
		// time to int64
		switch v := value.(type) {
		case time.Time:
			value = e.timeToTimestamp(v)
		case string:
			var err error
			value, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("string to float conversion failed: %w", err)
			}
		}
	}

	// to any of int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint types
	if strings.HasPrefix(e.DataFormat, "int") || strings.HasPrefix(e.DataFormat, "uint") {
		v := reflect.ValueOf(value)
		v = reflect.Indirect(v)

		switch value.(type) {
		case int64, int32, int16, int8, int:
			// check if overflow/data loss is possible
			in := v.Convert(reflect.TypeOf(int64(0))).Int()
			isIntegerConversionPossible(in, e.DataFormat, e.bytes)

			// convert to target type
			uin := v.Convert(getCorrespondingUintType(e.DataFormat)).Uint()
			bytes := make([]byte, 8)
			converter.PutUint64(bytes, uin)

			return trimByteSliceByEndianness(bytes, e.bytes, converter), nil
		case uint64, uint32, uint16, uint8, uint:
			// check if overflow/data loss is possible
			in := v.Convert(reflect.TypeOf(uint64(0))).Uint()
			isUIntegerConversionPossible(in, e.DataFormat, e.bytes)

			// convert to target type
			uin := v.Convert(getCorrespondingUintType(e.DataFormat)).Uint()
			bytes := make([]byte, 8)
			converter.PutUint64(bytes, uin)

			return trimByteSliceByEndianness(bytes, e.bytes, converter), nil
		case float32, float64:
			// check if overflow/data loss is possible
			log.Printf("I! [binary] float to integer conversion detected. Loss of data is very likely")
			uin := v.Convert(getCorrespondingUintType(e.DataFormat)).Uint()
			bytes := make([]byte, 8)
			converter.PutUint64(bytes, uin)

			return trimByteSliceByEndianness(bytes, e.bytes, converter), nil
		}
	}

	// to float32, float64
	if e.DataFormat == "float32" || e.DataFormat == "float64" {
		v := reflect.ValueOf(value)
		v = reflect.Indirect(v)
		f := v.Convert(reflect.TypeOf(float64(0))).Float()

		isFloatConversionPossible(f, e.DataFormat)

		switch e.DataFormat {
		case "float32":
			bytes := make([]byte, 4)
			bits32 := math.Float32bits(float32(f))
			converter.PutUint32(bytes, bits32)

			return bytes, nil
		case "float64":
			bytes := make([]byte, 8)
			bits64 := math.Float64bits(f)
			converter.PutUint64(bytes, bits64)

			return bytes, nil
		}
	}

	// to string
	if e.DataFormat == "string" {
		str, err := internal.ToString(value)
		if err != nil {
			return nil, err
		}

		bytes := []byte(str)

		// If string is longer than target length, truncate it and append terminator.
		// Thus, there is one less place for the data so that the terminator can be placed.
		if len(bytes) >= int(e.bytes) {
			dataLength := int(e.bytes) - 1
			return append(bytes[:dataLength], e.termination), nil
		}

		// If string is shorter than target length, fill the rest with terminator.
		return appendToSlice(bytes, e.bytes, e.termination), nil
	}

	return nil, fmt.Errorf("unknown type %T", value)
}

func trimByteSliceByEndianness(in []byte, targetSize uint64, order binary.ByteOrder) []byte {
	if order == binary.BigEndian {
		return in[8-targetSize:]
	}

	return in[:targetSize]
}

func getCorrespondingUintType(targetType string) reflect.Type {
	switch targetType {
	case "int64", "uint64", "float64":
		return reflect.TypeOf(uint64(0))
	case "int32", "uint32", "float32":
		return reflect.TypeOf(uint32(0))
	case "int16", "uint16":
		return reflect.TypeOf(uint16(0))
	case "int8", "uint8":
		return reflect.TypeOf(uint8(0))
	}

	return nil
}

func isIntegerConversionPossible[T constraints.Signed](value T, targetType string, targetBytes uint64) {
	if strings.HasPrefix(targetType, "int") {
		if int64(value) > (1<<(targetBytes*8)-1)/2 || int64(value) < -1<<(targetBytes*8)/2 {
			log.Printf("I! [binary] overflow detected while converting to %s. Loss of data is very likely", targetType)
		}
	}

	if strings.HasPrefix(targetType, "uint") {
		if int64(value) < 0 {
			log.Printf("I! [binary] negative value to unsigned type %s conversion detected", targetType)
		}

		if int64(value) > 0 && uint64(value) > 1<<(targetBytes*8)-1 {
			log.Printf("I! [binary] overflow detected while converting to %s. Loss of data is very likely", targetType)
		}
	}
}

func isUIntegerConversionPossible[T constraints.Unsigned](value T, targetType string, targetBytes uint64) {
	if strings.HasPrefix(targetType, "int") {
		if uint64(value) > (1<<(targetBytes*8)-1)/2 {
			log.Printf("I! [binary] overflow detected while converting to %s. Loss of data is very likely", targetType)
		}
	}

	if strings.HasPrefix(targetType, "uint") {
		if uint64(value) > 1<<(targetBytes*8)-1 {
			log.Printf("I! [binary] overflow detected while converting to %s. Loss of data is very likely", targetType)
		}
	}
}

func isFloatConversionPossible[T constraints.Float](value T, targetType string) {
	if targetType == "float32" {
		if float64(value) > math.MaxFloat32 || float64(value) < -math.MaxFloat32 {
			log.Printf("I! [binary] overflow detected while converting to %s. Loss of data is very likely", targetType)
		}
	}

	if targetType == "float64" {
		if float64(value) > math.MaxFloat64 || float64(value) < -math.MaxFloat64 {
			log.Printf("I! [binary] overflow detected while converting to %s. Loss of data is very likely", targetType)
		}
	}
}
