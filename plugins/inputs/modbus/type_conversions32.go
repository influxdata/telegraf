package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
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
func rescaleI32AsBigFloat(in int32, value_gain value_gain, value_offset value_offset) *big.Float {
	var t *big.Float = big.NewFloat(0.0)
	t.Mul(big.NewFloat(float64(in)), value_gain.asBigFloat())
	t.Add(t, value_offset.asBigFloat())
	return t
}
func rescaleU32AsBigFloat(in uint32, value_gain value_gain, value_offset value_offset) *big.Float {
	var t *big.Float = big.NewFloat(0.0)
	t.Mul(big.NewFloat(float64(in)), value_gain.asBigFloat())
	t.Add(t, value_offset.asBigFloat())
	return t
}
func rescaleF32AsBigFloat(in float32, value_gain value_gain, value_offset value_offset) *big.Float {
	var t *big.Float = big.NewFloat(0.0)
	t.Mul(big.NewFloat(float64(in)), value_gain.asBigFloat())
	t.Add(t, value_offset.asBigFloat())
	return t
}

// I32
func determineConverterI32(outType, byteOrder string, value_gain value_gain, value_offset value_offset) (fieldConverterFunc, error) {
	tohost, err := endianessConverter32(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native": // I32
		return func(b []byte) interface{} {
			in := int32(tohost(b))
			return int32(forceIntToMinMax(rescaleI32AsBigFloat(in, value_gain, value_offset), "INT32").Int64())
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			in := int32(tohost(b))
			return forceIntToMinMax(rescaleI32AsBigFloat(in, value_gain, value_offset), "INT64").Int64()
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int32(tohost(b))
			return forceIntToMinMax(rescaleI32AsBigFloat(in, value_gain, value_offset), "UINT64").Uint64()
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int32(tohost(b))
			return forceFloat64ToMinMax(rescaleI32AsBigFloat(in, value_gain, value_offset))
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U32
func determineConverterU32(outType, byteOrder string, value_gain value_gain, value_offset value_offset) (fieldConverterFunc, error) {
	tohost, err := endianessConverter32(byteOrder)
	if err != nil {
		return nil, err
	}
	switch outType {
	case "native": // U32
		return func(b []byte) interface{} {
			in := tohost(b)
			return uint32(forceIntToMinMax(rescaleU32AsBigFloat(in, value_gain, value_offset), "UINT32").Int64())
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			in := tohost(b)
			return forceIntToMinMax(rescaleU32AsBigFloat(in, value_gain, value_offset), "INT64").Int64()
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := tohost(b)
			return forceIntToMinMax(rescaleU32AsBigFloat(in, value_gain, value_offset), "UINT64").Uint64()
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := tohost(b)
			return forceFloat64ToMinMax(rescaleU32AsBigFloat(in, value_gain, value_offset))
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// F32
func determineConverterF32(outType, byteOrder string, value_gain value_gain, value_offset value_offset) (fieldConverterFunc, error) {
	tohost, err := endianessConverter32(byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			in := math.Float32frombits(tohost(b))
			fmt.Printf("%v\t%v\n", in, rescaleF32AsBigFloat(in, value_gain, value_offset))
			return forceFloat32ToMinMax(rescaleF32AsBigFloat(in, value_gain, value_offset))

		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := math.Float32frombits(tohost(b))
			return forceFloat64ToMinMax(rescaleF32AsBigFloat(in, value_gain, value_offset))

		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}
