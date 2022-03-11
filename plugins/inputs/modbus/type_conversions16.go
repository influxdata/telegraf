package modbus

import (
	"encoding/binary"
	"fmt"
)

type convert16 func([]byte) uint16

func endianessConverter16(byteOrder string) (convert16, error) {
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
	tohost, err := endianessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			return int16(tohost(b))
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			return int64(int16(tohost(b)))
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			return uint64(int16(tohost(b)))
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			return float64(int16(tohost(b)))
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U16 - no scale
func determineConverterU16(outType, byteOrder string) (fieldConverterFunc, error) {
	tohost, err := endianessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			return tohost(b)
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			return int64(tohost(b))
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			return uint64(tohost(b))
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			return float64(tohost(b))
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// I16 - scale
func determineConverterI16Scale(outType, byteOrder string, scale float64) (fieldConverterFunc, error) {
	tohost, err := endianessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			in := int16(tohost(b))
			return int16(float64(in) * scale)
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			in := int16(tohost(b))
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int16(tohost(b))
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int16(tohost(b))
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U16 - scale
func determineConverterU16Scale(outType, byteOrder string, scale float64) (fieldConverterFunc, error) {
	tohost, err := endianessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			in := tohost(b)
			return uint16(float64(in) * scale)
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			in := tohost(b)
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := tohost(b)
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := tohost(b)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}
