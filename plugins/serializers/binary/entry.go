package binary

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

type converterFunc func(value interface{}, order binary.ByteOrder) ([]byte, error)

type Entry struct {
	ReadFrom         string `toml:"read_from"`         // field, tag, time, name
	Name             string `toml:"name"`              // name of entry
	DataFormat       string `toml:"data_format"`       // int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64, string
	StringTerminator string `toml:"string_terminator"` // for string metrics: null, 0x00, 00, ....
	StringLength     uint64 `toml:"string_length"`     // for string only, target size
	TimeFormat       string `toml:"time_format"`       // for time metrics: unix, unix_ms, unix_us, unix_ns

	converter   converterFunc
	termination byte
}

func (e *Entry) fillDefaults() error {
	// Normalize
	e.ReadFrom = strings.ToLower(e.ReadFrom)

	// Check input constraints
	switch e.ReadFrom {
	case "":
		e.ReadFrom = "field"
		fallthrough
	case "field", "tag":
		if e.Name == "" {
			return errors.New("missing name")
		}
	case "time":
		switch e.TimeFormat {
		case "":
			e.TimeFormat = "unix"
		case "unix", "unix_ms", "unix_us", "unix_ns":
		default:
			return errors.New("invalid time format")
		}
	case "name":
		if e.DataFormat == "" {
			e.DataFormat = "string"
		} else if e.DataFormat != "string" {
			return errors.New("name data format has to be string")
		}
	default:
		return fmt.Errorf("unknown assignment %q", e.ReadFrom)
	}

	// Check data format
	switch e.DataFormat {
	case "":
		return errors.New("missing data format")
	case "float64":
		e.converter = convertToFloat64
	case "float32":
		e.converter = convertToFloat32
	case "uint64":
		e.converter = convertToUint64
	case "uint32":
		e.converter = convertToUint32
	case "uint16":
		e.converter = convertToUint16
	case "uint8":
		e.converter = convertToUint8
	case "int64":
		e.converter = convertToInt64
	case "int32":
		e.converter = convertToInt32
	case "int16":
		e.converter = convertToInt16
	case "int8":
		e.converter = convertToInt8
	case "string":
		switch e.StringTerminator {
		case "", "null":
			e.termination = 0x00
		default:
			e.StringTerminator = strings.TrimPrefix(e.StringTerminator, "0x")
			termination, err := hex.DecodeString(e.StringTerminator)
			if err != nil {
				return fmt.Errorf("decoding terminator failed for %q: %w", e.Name, err)
			}
			if len(termination) != 1 {
				return fmt.Errorf("terminator must be a single byte, got %q", e.StringTerminator)
			}
			e.termination = termination[0]
		}

		if e.StringLength < 1 {
			return errors.New("string length must be at least 1")
		}
		e.converter = e.convertToString
	default:
		return fmt.Errorf("invalid data format %q for field %q", e.ReadFrom, e.DataFormat)
	}

	return nil
}

func (e *Entry) serializeValue(value interface{}, order binary.ByteOrder) ([]byte, error) {
	// Handle normal fields, tags, etc
	if e.ReadFrom != "time" {
		return e.converter(value, order)
	}

	// We need to serialize the time, make sure we actually do get a time and
	// convert it to the correct timestamp (with scale) first.
	t, ok := value.(time.Time)
	if !ok {
		return nil, fmt.Errorf("time expected but got %T", value)
	}

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
	return e.converter(timestamp, order)
}
