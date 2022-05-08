package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
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

// I16
func determineConverterI16(outType, byteOrder string, scale float64, shift float64) (fieldConverterFunc, error) {
	tohost, err := endianessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			in := int16(tohost(b))
			return int16(float64(in)*scale + shift)
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			in := int16(tohost(b))
			return int64(float64(in)*scale + shift)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int16(tohost(b))
			return uint64(math.Max(0, float64(in)*scale+shift))
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int16(tohost(b))
			return float64(in)*scale + shift
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U16
func determineConverterU16(outType, byteOrder string, scale float64, shift float64) (fieldConverterFunc, error) {
	tohost, err := endianessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}
	fmt.Print("OutType is : " + outType + "\n")
	switch outType {
	case "native":
		return func(b []byte) interface{} {
			in := tohost(b)
			return uint16(math.Max(0, float64(in)*scale+shift))
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			in := tohost(b)
			return int64(float64(in)*scale + shift)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := tohost(b)
			return uint64(math.Max(0, float64(in)*scale+shift))
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := tohost(b)
			return float64(in)*scale + shift
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}
