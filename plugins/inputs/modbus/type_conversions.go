package modbus

import (
	"fmt"
	"math"
	"math/big"
)

func determineConverter(inType, byteOrder, outType string, gain value_gain, offset value_offset) (fieldConverterFunc, error) {
	switch inType {
	case "INT16":
		return determineConverterI16(outType, byteOrder, gain, offset)
	case "UINT16":
		return determineConverterU16(outType, byteOrder, gain, offset)
	case "INT32":
		return determineConverterI32(outType, byteOrder, gain, offset)
	case "UINT32":
		return determineConverterU32(outType, byteOrder, gain, offset)
	case "INT64":
		return determineConverterI64(outType, byteOrder, gain, offset)
	case "UINT64":
		return determineConverterU64(outType, byteOrder, gain, offset)
	case "FLOAT32":
		return determineConverterF32(outType, byteOrder, gain, offset)
	case "FLOAT64":
		return determineConverterF64(outType, byteOrder, gain, offset)
	}
	return nil, fmt.Errorf("invalid input data-type: %s", inType)
}

func forceIntToMinMax(value *big.Float, outType string) *big.Int {
	var min, max *big.Float = big.NewFloat(0), big.NewFloat(0)
	switch outType {
	case "INT16":
		min = big.NewFloat(-math.MaxInt16 - 1)
		max = big.NewFloat(math.MaxInt16)
	case "UINT16":
		min = big.NewFloat(0)
		max = big.NewFloat(math.MaxUint16)
	case "INT32":
		min = big.NewFloat(-math.MaxInt32 - 1)
		max = big.NewFloat(math.MaxInt32)
	case "UINT32":
		min = big.NewFloat(0)
		max = big.NewFloat(math.MaxUint32)
	case "INT64":
		min = big.NewFloat(-math.MaxInt64 - 1)
		max = big.NewFloat(math.MaxInt64)
	case "UINT64":
		min = big.NewFloat(0)
		max = max.SetUint64(math.MaxUint64)
	}
	var out *big.Int
	if value.Cmp(min) == -1 {
		out, _ = min.Int(out)
	} else if value.Cmp(max) == 1 {
		out, _ = max.Int(out)
		if outType == "UINT64" {
			out = out.SetUint64(math.MaxUint64)
		}
	} else {
		out, _ = value.Int(out)
	}
	return out
}

func forceFloat32ToMinMax(value *big.Float) float32 {
	var min, max *big.Float
	var out float32
	min = big.NewFloat(-math.MaxFloat64)
	max = big.NewFloat(math.MaxFloat64)

	if value.Cmp(min) == -1 {
		out, _ = min.Float32()
	} else if value.Cmp(max) == 1 {
		out, _ = max.Float32()
	} else {
		out, _ = value.Float32()
	}
	return out
}

func forceFloat64ToMinMax(value *big.Float) float64 {
	var min, max *big.Float
	var out float64
	min = big.NewFloat(-math.MaxFloat64)
	max = big.NewFloat(math.MaxFloat64)

	if value.Cmp(min) == -1 {
		out, _ = min.Float64()
	} else if value.Cmp(max) == 1 {
		out, _ = max.Float64()
	} else {
		out, _ = value.Float64()
	}
	return out
}
