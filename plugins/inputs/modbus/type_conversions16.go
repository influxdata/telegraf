package modbus

import (
	"encoding/binary"
	"fmt"

	"github.com/x448/float16"
)

type convert16 func([]byte) uint16

func endiannessConverter16(byteOrder string) (convert16, error) {
	switch byteOrder {
	case "ABCD", "CDAB": // Big endian (Motorola)
		return binary.BigEndian.Uint16, nil
	case "DCBA", "BADC": // Little endian (Intel)
		return binary.LittleEndian.Uint16, nil
	}
	return nil, fmt.Errorf("invalid byte-order: %s", byteOrder)
}

// I16 - no scale
func determineConverterI16(outType, byteOrder string) (fieldConverterFunc, error) {
	tohost, err := endiannessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) any {
			return int16(tohost(b))
		}, nil
	case "INT64":
		return func(b []byte) any {
			return int64(int16(tohost(b)))
		}, nil
	case "UINT64":
		return func(b []byte) any {
			return uint64(int16(tohost(b)))
		}, nil
	case "FLOAT64":
		return func(b []byte) any {
			return float64(int16(tohost(b)))
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U16 - no scale
func determineConverterU16(outType, byteOrder string) (fieldConverterFunc, error) {
	tohost, err := endiannessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) any {
			return tohost(b)
		}, nil
	case "INT64":
		return func(b []byte) any {
			return int64(tohost(b))
		}, nil
	case "UINT64":
		return func(b []byte) any {
			return uint64(tohost(b))
		}, nil
	case "FLOAT64":
		return func(b []byte) any {
			return float64(tohost(b))
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// F16 - no scale
func determineConverterF16(outType, byteOrder string) (fieldConverterFunc, error) {
	tohost, err := endiannessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) any {
			raw := tohost(b)
			return float16.Frombits(raw).Float32()
		}, nil
	case "FLOAT64":
		return func(b []byte) any {
			raw := tohost(b)
			in := float16.Frombits(raw).Float32()
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// I16 - scale
func determineConverterI16Scale(outType, byteOrder string, scale float64) (fieldConverterFunc, error) {
	tohost, err := endiannessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) any {
			in := int16(tohost(b))
			return int16(float64(in) * scale)
		}, nil
	case "INT64":
		return func(b []byte) any {
			in := int16(tohost(b))
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) any {
			in := int16(tohost(b))
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) any {
			in := int16(tohost(b))
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U16 - scale
func determineConverterU16Scale(outType, byteOrder string, scale float64) (fieldConverterFunc, error) {
	tohost, err := endiannessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) any {
			in := tohost(b)
			return uint16(float64(in) * scale)
		}, nil
	case "INT64":
		return func(b []byte) any {
			in := tohost(b)
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) any {
			in := tohost(b)
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) any {
			in := tohost(b)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// F16 - scale
func determineConverterF16Scale(outType, byteOrder string, scale float64) (fieldConverterFunc, error) {
	tohost, err := endiannessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) any {
			raw := tohost(b)
			in := float16.Frombits(raw)
			return in.Float32() * float32(scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) any {
			raw := tohost(b)
			in := float16.Frombits(raw)
			return float64(in.Float32()) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// BCD16 - no scale
// Values that are not valid BCD return nil, so the field is left out of the metric.
func determineConverterBCD16(outType, byteOrder string) (fieldConverterFunc, error) {
	tohost, err := endiannessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) any {
			if v, ok := decodeBCD(uint64(tohost(b))); ok {
				return uint16(v)
			}
			return nil
		}, nil
	case "INT64":
		return func(b []byte) any {
			if v, ok := decodeBCD(uint64(tohost(b))); ok {
				return int64(v)
			}
			return nil
		}, nil
	case "UINT64":
		return func(b []byte) any {
			if v, ok := decodeBCD(uint64(tohost(b))); ok {
				return v
			}
			return nil
		}, nil
	case "FLOAT64":
		return func(b []byte) any {
			if v, ok := decodeBCD(uint64(tohost(b))); ok {
				return float64(v)
			}
			return nil
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// BCD16 - scale
// Values that are not valid BCD return nil, so the field is left out of the metric.
func determineConverterBCD16Scale(outType, byteOrder string, scale float64) (fieldConverterFunc, error) {
	tohost, err := endiannessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) any {
			if v, ok := decodeBCD(uint64(tohost(b))); ok {
				return uint16(float64(v) * scale)
			}
			return nil
		}, nil
	case "INT64":
		return func(b []byte) any {
			if v, ok := decodeBCD(uint64(tohost(b))); ok {
				return int64(float64(v) * scale)
			}
			return nil
		}, nil
	case "UINT64":
		return func(b []byte) any {
			if v, ok := decodeBCD(uint64(tohost(b))); ok {
				return uint64(float64(v) * scale)
			}
			return nil
		}, nil
	case "FLOAT64":
		return func(b []byte) any {
			if v, ok := decodeBCD(uint64(tohost(b))); ok {
				return float64(v) * scale
			}
			return nil
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}
