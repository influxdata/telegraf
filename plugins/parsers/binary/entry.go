package binary

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type Entry struct {
	Name       string `toml:"name"`
	Type       string `toml:"type"`
	Bits       uint64 `toml:"bits"`
	Omit       bool   `toml:"omit"`
	Terminator string `toml:"terminator"`
	Timezone   string `toml:"timezone"`
	Assignment string `toml:"assignment"`

	termination []byte
	location    *time.Location
}

func (e *Entry) check() error {
	// Normalize cases
	e.Assignment = strings.ToLower(e.Assignment)
	e.Terminator = strings.ToLower(e.Terminator)
	if e.Assignment != "time" {
		e.Type = strings.ToLower(e.Type)
	}

	// Handle omitted fields
	if e.Omit {
		if e.Bits == 0 && e.Type == "" {
			return errors.New("neither type nor bits given")
		}
		if e.Bits == 0 {
			bits, err := bitsForType(e.Type)
			if err != nil {
				return err
			}
			e.Bits = bits
		}
		return nil
	}

	// Set name for global options
	if e.Assignment == "measurement" || e.Assignment == "time" {
		e.Name = e.Assignment
	}

	// Check the name
	if e.Name == "" {
		return errors.New("missing name")
	}

	// Check the assignment
	var defaultType string
	switch e.Assignment {
	case "measurement":
		defaultType = "string"
		if e.Type != "string" && e.Type != "" {
			return errors.New("'measurement' type has to be 'string'")
		}
	case "time":
		bits := uint64(64)

		switch e.Type {
		// Make 'unix' the default
		case "":
			defaultType = "unix"
		// Special plugin specific names
		case "unix", "unix_ms", "unix_us", "unix_ns":
		// Format-specification string formats
		default:
			bits = uint64(len(e.Type) * 8)
		}
		if e.Bits == 0 {
			e.Bits = bits
		}

		switch e.Timezone {
		case "", "utc":
			// Make UTC the default
			e.location = time.UTC
		case "local":
			e.location = time.Local
		default:
			var err error
			e.location, err = time.LoadLocation(e.Timezone)
			if err != nil {
				return err
			}
		}
	case "tag":
		defaultType = "string"
	case "", "field":
		e.Assignment = "field"
	default:
		return fmt.Errorf("no assignment for %q", e.Name)
	}

	// Check type (special type for "time")
	switch e.Type {
	case "uint8", "int8", "uint16", "int16", "uint32", "int32", "uint64", "int64":
		fallthrough
	case "float32", "float64":
		bits, err := bitsForType(e.Type)
		if err != nil {
			return err
		}
		if e.Bits == 0 {
			e.Bits = bits
		}
		if bits < e.Bits {
			return fmt.Errorf("type overflow for %q", e.Name)
		}
	case "bool":
		if e.Bits == 0 {
			e.Bits = 1
		}
	case "string":
		// Check termination
		switch e.Terminator {
		case "", "fixed":
			e.Terminator = "fixed"
			if e.Bits == 0 {
				return fmt.Errorf("require 'bits' for fixed-length string for %q", e.Name)
			}
		case "null":
			e.termination = []byte{0}
			if e.Bits != 0 {
				return fmt.Errorf("cannot use 'bits' and 'null' terminator together for %q", e.Name)
			}
		default:
			if e.Bits != 0 {
				return fmt.Errorf("cannot use 'bits' and terminator together for %q", e.Name)
			}
			var err error
			e.termination, err = hex.DecodeString(strings.TrimPrefix(e.Terminator, "0x"))
			if err != nil {
				return fmt.Errorf("decoding terminator failed for %q: %w", e.Name, err)
			}
		}

		// We can only handle strings that adhere to byte-bounds
		if e.Bits%8 != 0 {
			return fmt.Errorf("non-byte length for string field %q", e.Name)
		}
	case "":
		if defaultType == "" {
			return fmt.Errorf("no type for %q", e.Name)
		}
		e.Type = defaultType
	default:
		if e.Assignment != "time" {
			return fmt.Errorf("unknown type for %q", e.Name)
		}
	}

	return nil
}

func (e *Entry) extract(in []byte, offset uint64) ([]byte, uint64, error) {
	if e.Bits > 0 {
		data, err := extractPart(in, offset, e.Bits)
		return data, e.Bits, err
	}

	if e.Type != "string" {
		return nil, 0, fmt.Errorf("unexpected entry: %v", e)
	}

	inbits := uint64(len(in)) * 8

	// Read up to the termination
	var found bool
	var data []byte
	var termOffset int
	var n uint64
	for offset+n+8 <= inbits {
		buf, err := extractPart(in, offset+n, 8)
		if err != nil {
			return nil, 0, err
		}
		if len(buf) != 1 {
			return nil, 0, fmt.Errorf("unexpected length %d", len(buf))
		}
		data = append(data, buf[0])
		n += 8

		// Check for terminator
		if buf[0] == e.termination[termOffset] {
			termOffset++
		}
		if termOffset == len(e.termination) {
			found = true
			break
		}
	}
	if !found {
		return nil, n, fmt.Errorf("terminator not found for %q", e.Name)
	}

	// Strip the terminator
	return data[:len(data)-len(e.termination)], n, nil
}

func (e *Entry) convertType(in []byte, order binary.ByteOrder) (interface{}, error) {
	switch e.Type {
	case "uint8", "int8", "uint16", "int16", "uint32", "int32", "float32", "uint64", "int64", "float64":
		return convertNumericType(in, e.Type, order)
	case "bool":
		return convertBoolType(in), nil
	case "string":
		return convertStringType(in), nil
	}

	return nil, fmt.Errorf("cannot handle type %q", e.Type)
}

func (e *Entry) convertTimeType(in []byte, order binary.ByteOrder) (time.Time, error) {
	factor := int64(1)

	switch e.Type {
	case "unix":
		factor *= 1000
		fallthrough
	case "unix_ms":
		factor *= 1000
		fallthrough
	case "unix_us":
		factor *= 1000
		fallthrough
	case "unix_ns":
		raw, err := convertNumericType(in, "int64", order)
		if err != nil {
			return time.Unix(0, 0), err
		}
		v := raw.(int64)
		return time.Unix(0, v*factor).In(e.location), nil
	}
	// We have a format specification (hopefully)
	v := convertStringType(in)
	return time.ParseInLocation(e.Type, v, e.location)
}

func convertStringType(in []byte) string {
	return string(in)
}

func convertNumericType(in []byte, t string, order binary.ByteOrder) (interface{}, error) {
	bits, err := bitsForType(t)
	if err != nil {
		return nil, err
	}

	inlen := uint64(len(in))
	expected := bits / 8
	if inlen > expected {
		// Should never happen
		return 0, fmt.Errorf("too many bytes %d vs %d", len(in), expected)
	}

	// Pad the data if shorter than the datatype length
	buf := make([]byte, expected-inlen, expected)
	buf = append(buf, in...)

	switch t {
	case "uint8":
		return buf[0], nil
	case "int8":
		return int8(buf[0]), nil
	case "uint16":
		return order.Uint16(buf), nil
	case "int16":
		v := order.Uint16(buf)
		return int16(v), nil
	case "uint32":
		return order.Uint32(buf), nil
	case "int32":
		v := order.Uint32(buf)
		return int32(v), nil
	case "uint64":
		return order.Uint64(buf), nil
	case "int64":
		v := order.Uint64(buf)
		return int64(v), nil
	case "float32":
		v := order.Uint32(buf)
		return math.Float32frombits(v), nil
	case "float64":
		v := order.Uint64(buf)
		return math.Float64frombits(v), nil
	}
	return nil, fmt.Errorf("no numeric type %q", t)
}

func convertBoolType(in []byte) bool {
	for _, x := range in {
		if x != 0 {
			return true
		}
	}
	return false
}
