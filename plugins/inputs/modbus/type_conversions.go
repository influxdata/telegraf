package modbus

import (
	"fmt"
)

func determineUntypedConverter(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "", "UINT16":
		return func(b []byte) any {
			return uint16(b[0])
		}, nil
	case "BOOL":
		return func(b []byte) any {
			return b[0] != 0
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverter(inType, byteOrder, outType string, scale float64, bit uint8, strloc string) (fieldConverterFunc, error) {
	switch inType {
	case "STRING":
		switch strloc {
		case "", "both":
			return determineConverterString(byteOrder)
		case "lower":
			return determineConverterStringLow(byteOrder)
		case "upper":
			return determineConverterStringHigh(byteOrder)
		}
	case "BIT":
		return determineConverterBit(byteOrder, bit)
	}

	if scale != 0.0 {
		return determineConverterScale(inType, byteOrder, outType, scale)
	}
	return determineConverterNoScale(inType, byteOrder, outType)
}

func determineConverterScale(inType, byteOrder, outType string, scale float64) (fieldConverterFunc, error) {
	switch inType {
	case "INT8L":
		return determineConverterI8LScale(outType, byteOrder, scale)
	case "INT8H":
		return determineConverterI8HScale(outType, byteOrder, scale)
	case "UINT8L":
		return determineConverterU8LScale(outType, byteOrder, scale)
	case "UINT8H":
		return determineConverterU8HScale(outType, byteOrder, scale)
	case "INT16":
		return determineConverterI16Scale(outType, byteOrder, scale)
	case "UINT16":
		return determineConverterU16Scale(outType, byteOrder, scale)
	case "INT32":
		return determineConverterI32Scale(outType, byteOrder, scale)
	case "UINT32":
		return determineConverterU32Scale(outType, byteOrder, scale)
	case "INT64":
		return determineConverterI64Scale(outType, byteOrder, scale)
	case "UINT64":
		return determineConverterU64Scale(outType, byteOrder, scale)
	case "FLOAT16":
		return determineConverterF16Scale(outType, byteOrder, scale)
	case "FLOAT32":
		return determineConverterF32Scale(outType, byteOrder, scale)
	case "FLOAT_DEC_8E24M":
		return determineConverterFDec8E24MScale(outType, byteOrder, scale, true)
	case "FLOAT_UDEC_8E24M":
		return determineConverterFDec8E24MScale(outType, byteOrder, scale, false)
	case "BCD16":
		return determineConverterBCD16Scale(outType, byteOrder, scale)
	case "BCD32":
		return determineConverterBCD32Scale(outType, byteOrder, scale)
	case "BCD64":
		return determineConverterBCD64Scale(outType, byteOrder, scale)
	case "FLOAT64":
		return determineConverterF64Scale(outType, byteOrder, scale)
	}
	return nil, fmt.Errorf("invalid input data-type: %s", inType)
}

func determineConverterNoScale(inType, byteOrder, outType string) (fieldConverterFunc, error) {
	switch inType {
	case "INT8L":
		return determineConverterI8L(outType, byteOrder)
	case "INT8H":
		return determineConverterI8H(outType, byteOrder)
	case "UINT8L":
		return determineConverterU8L(outType, byteOrder)
	case "UINT8H":
		return determineConverterU8H(outType, byteOrder)
	case "INT16":
		return determineConverterI16(outType, byteOrder)
	case "UINT16":
		return determineConverterU16(outType, byteOrder)
	case "INT32":
		return determineConverterI32(outType, byteOrder)
	case "UINT32":
		return determineConverterU32(outType, byteOrder)
	case "INT64":
		return determineConverterI64(outType, byteOrder)
	case "UINT64":
		return determineConverterU64(outType, byteOrder)
	case "FLOAT16":
		return determineConverterF16(outType, byteOrder)
	case "FLOAT32":
		return determineConverterF32(outType, byteOrder)
	case "FLOAT_DEC_8E24M":
		return determineConverterFDec8E24M(outType, byteOrder, true)
	case "FLOAT_UDEC_8E24M":
		return determineConverterFDec8E24M(outType, byteOrder, false)
	case "BCD16":
		return determineConverterBCD16(outType, byteOrder)
	case "BCD32":
		return determineConverterBCD32(outType, byteOrder)
	case "BCD64":
		return determineConverterBCD64(outType, byteOrder)
	case "FLOAT64":
		return determineConverterF64(outType, byteOrder)
	}
	return nil, fmt.Errorf("invalid input data-type: %s", inType)
}

// decodeBCD decodes packed binary-coded decimal, where every 4-bit nibble holds
// one decimal digit with the most significant digit first, e.g. 0x1234 is 1234.
// The second return value is false if a nibble is above 9, i.e. not valid BCD.
func decodeBCD(raw uint64) (uint64, bool) {
	var value uint64
	factor := uint64(1)
	for ; raw > 0; raw >>= 4 {
		digit := raw & 0x0f
		if digit > 9 {
			return 0, false
		}
		value += digit * factor
		factor *= 10
	}
	return value, true
}
