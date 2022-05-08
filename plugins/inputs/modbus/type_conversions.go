package modbus

import "fmt"

func determineConverter(inType, byteOrder, outType string, scale float64, shift float64) (fieldConverterFunc, error) {
	if scale == 0.0 {
		scale = 1.0
	}
	switch inType {
	case "INT16":
		return determineConverterI16(outType, byteOrder, scale, shift)
	case "UINT16":
		return determineConverterU16(outType, byteOrder, scale, shift)
	case "INT32":
		return determineConverterI32(outType, byteOrder, scale, shift)
	case "UINT32":
		return determineConverterU32(outType, byteOrder, scale, shift)
	case "INT64":
		return determineConverterI64(outType, byteOrder, scale, shift)
	case "UINT64":
		return determineConverterU64(outType, byteOrder, scale, shift)
	case "FLOAT32":
		return determineConverterF32(outType, byteOrder, scale, shift)
	case "FLOAT64":
		return determineConverterF64(outType, byteOrder, scale, shift)
	}
	return nil, fmt.Errorf("invalid input data-type: %s", inType)
}
