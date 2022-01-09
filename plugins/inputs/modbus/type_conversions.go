package modbus

import "fmt"

func determineConverter(inType, byteOrder, outType string, scale float64) (fieldConverterFunc, error) {
	if scale != 0.0 {
		return determineConverterScale(inType, byteOrder, outType, scale)
	}
	return determineConverterNoScale(inType, byteOrder, outType)
}

func determineConverterScale(inType, byteOrder, outType string, scale float64) (fieldConverterFunc, error) {
	switch inType {
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
	case "FLOAT32":
		return determineConverterF32Scale(outType, byteOrder, scale)
	case "FLOAT64":
		return determineConverterF64Scale(outType, byteOrder, scale)
	}
	return nil, fmt.Errorf("invalid input data-type: %s", inType)
}

func determineConverterNoScale(inType, byteOrder, outType string) (fieldConverterFunc, error) {
	switch inType {
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
	case "FLOAT32":
		return determineConverterF32(outType, byteOrder)
	case "FLOAT64":
		return determineConverterF64(outType, byteOrder)
	}
	return nil, fmt.Errorf("invalid input data-type: %s", inType)
}
