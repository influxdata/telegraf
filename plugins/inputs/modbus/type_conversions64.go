package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
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

func rescaleI64AsBigFloat(in int64, value_gain value_gain, value_offset value_offset) *big.Float {
	var t *big.Float = big.NewFloat(0.0)
	t.Mul(big.NewFloat(float64(in)), value_gain.asBigFloat())
	t.Add(t, value_offset.asBigFloat())
	return t
}
func rescaleU64AsBigFloat(in uint64, value_gain value_gain, value_offset value_offset) *big.Float {
	var t *big.Float = big.NewFloat(0.0)
	t.Mul(big.NewFloat(float64(in)), value_gain.asBigFloat())
	t.Add(t, value_offset.asBigFloat())
	return t
}
func rescaleF64AsBigFloat(in float64, value_gain value_gain, value_offset value_offset) *big.Float {
	var t *big.Float = big.NewFloat(0.0)
	t.Mul(big.NewFloat(float64(in)), value_gain.asBigFloat())
	t.Add(t, value_offset.asBigFloat())
	return t
}

// I64
func determineConverterI64(outType, byteOrder string, value_gain value_gain, value_offset value_offset) (fieldConverterFunc, error) {
	tohost, err := endianessConverter64(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native": // I64
		return func(b []byte) interface{} {
			in := int64(tohost(b))
			return int64(forceIntToMinMax(rescaleI64AsBigFloat(in, value_gain, value_offset), "INT64").Int64())
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int64(tohost(b))
			return forceIntToMinMax(rescaleI64AsBigFloat(in, value_gain, value_offset), "UINT64").Uint64()
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int64(tohost(b))
			return forceFloat64ToMinMax(rescaleI64AsBigFloat(in, value_gain, value_offset))
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U64
func determineConverterU64(outType, byteOrder string, value_gain value_gain, value_offset value_offset) (fieldConverterFunc, error) {
	tohost, err := endianessConverter64(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native": // U64
		return func(b []byte) interface{} {
			in := tohost(b)
			return forceIntToMinMax(rescaleU64AsBigFloat(in, value_gain, value_offset), "UINT64").Uint64()
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			in := tohost(b)
			return forceIntToMinMax(rescaleU64AsBigFloat(in, value_gain, value_offset), "INT64").Int64()
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := tohost(b)
			return forceFloat64ToMinMax(rescaleU64AsBigFloat(in, value_gain, value_offset))
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// F64
func determineConverterF64(outType, byteOrder string, value_gain value_gain, value_offset value_offset) (fieldConverterFunc, error) {
	tohost, err := endianessConverter64(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native", "FLOAT64":
		return func(b []byte) interface{} {
			in := math.Float64frombits(tohost(b))
			return forceFloat64ToMinMax(rescaleF64AsBigFloat(in, value_gain, value_offset))
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}
