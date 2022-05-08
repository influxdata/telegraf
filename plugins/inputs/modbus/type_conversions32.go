package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
)

type convert32 func([]byte) uint32

func binaryMSWLEU32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(binary.LittleEndian.Uint16(b[0:]))<<16 | uint32(binary.LittleEndian.Uint16(b[2:]))
}

func binaryLSWBEU32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(binary.BigEndian.Uint16(b[2:]))<<16 | uint32(binary.BigEndian.Uint16(b[0:]))
}

func endianessConverter32(byteOrder string) (convert32, error) {
	switch byteOrder {
	case "ABCD": // Big endian (Motorola)
		return binary.BigEndian.Uint32, nil
	case "BADC": // Big endian with bytes swapped
		return binaryMSWLEU32, nil
	case "CDAB": // Little endian with bytes swapped
		return binaryLSWBEU32, nil
	case "DCBA": // Little endian (Intel)
		return binary.LittleEndian.Uint32, nil
	}
	return nil, fmt.Errorf("invalid byte-order: %s", byteOrder)
}

// I32
func determineConverterI32(outType, byteOrder string, scale float64, shift float64) (fieldConverterFunc, error) {
	tohost, err := endianessConverter32(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			in := int32(tohost(b))
			return int32(float64(in)*scale + shift)
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			in := int32(tohost(b))
			return int64(float64(in)*scale + shift)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int32(tohost(b))
			return uint64(math.Max(0, float64(in)*scale+shift))
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int32(tohost(b))
			return float64(in)*scale + shift
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U32
func determineConverterU32(outType, byteOrder string, scale float64, shift float64) (fieldConverterFunc, error) {
	tohost, err := endianessConverter32(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			in := tohost(b)
			return uint32(math.Max(0, float64(in)*scale+shift))
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

// F32
func determineConverterF32(outType, byteOrder string, scale float64, shift float64) (fieldConverterFunc, error) {
	tohost, err := endianessConverter32(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			raw := tohost(b)
			in := math.Float32frombits(raw)
			return float32(float64(in)*scale + shift)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			raw := tohost(b)
			in := math.Float32frombits(raw)
			return float64(in)*scale + shift
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}
