package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
)

type convert64 func([]byte) uint64

func binaryMSWLEU64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(binary.LittleEndian.Uint16(b[0:]))<<48 | uint64(binary.LittleEndian.Uint16(b[2:]))<<32 | uint64(binary.LittleEndian.Uint16(b[4:]))<<16 | uint64(binary.LittleEndian.Uint16(b[6:]))
}

func binaryLSWBEU64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(binary.BigEndian.Uint16(b[6:]))<<48 | uint64(binary.BigEndian.Uint16(b[4:]))<<32 | uint64(binary.BigEndian.Uint16(b[2:]))<<16 | uint64(binary.BigEndian.Uint16(b[0:]))
}

func endianessConverter64(byteOrder string) (convert64, error) {
	switch byteOrder {
	case "ABCD": // Big endian (Motorola)
		return binary.BigEndian.Uint64, nil
	case "BADC": // Big endian with bytes swapped
		return binaryMSWLEU64, nil
	case "CDAB": // Little endian with bytes swapped
		return binaryLSWBEU64, nil
	case "DCBA": // Little endian (Intel)
		return binary.LittleEndian.Uint64, nil
	}
	return nil, fmt.Errorf("invalid byte-order: %s", byteOrder)
}

// I64
func determineConverterI64(outType, byteOrder string, scale float64, shift float64) (fieldConverterFunc, error) {
	tohost, err := endianessConverter64(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			in := int64(tohost(b))
			return int64(float64(in)*scale + shift)
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			in := int64(tohost(b))
			return int64(float64(in)*scale + shift)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int64(tohost(b))
			return uint64(math.Max(0, float64(in)*scale+shift))
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int64(tohost(b))
			return float64(in)*scale + shift
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U64
func determineConverterU64(outType, byteOrder string, scale float64, shift float64) (fieldConverterFunc, error) {
	tohost, err := endianessConverter64(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			in := tohost(b)
			return uint64(math.Max(0, float64(in)*scale+shift))
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

// F64
func determineConverterF64(outType, byteOrder string, scale float64, shift float64) (fieldConverterFunc, error) {
	tohost, err := endianessConverter64(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native", "FLOAT64":
		return func(b []byte) interface{} {
			raw := tohost(b)
			in := math.Float64frombits(raw)
			return in*scale + shift
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}
