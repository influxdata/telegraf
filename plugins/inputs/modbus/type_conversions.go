package modbus

import "fmt"

func determineConverter(inType, byteOrder, outType string, scale float64) (fieldConverterFunc, error) {
	switch byteOrder {
	case "ABCD": // Big endian (Motorola)
		return determineConverterABCD(inType, outType, scale)
	case "BADC": // Big endian with bytes swapped
		return determineConverterBADC(inType, outType, scale)
	case "CDAB": // Little endian with bytes swapped
		return determineConverterCDAB(inType, outType, scale)
	case "DCBA": // Little endian (Intel)
		return determineConverterDCBA(inType, outType, scale)
	}
	return nil, fmt.Errorf("invalid byte-order: %s", byteOrder)
}

func determineConverterABCD(inType, outType string, scale float64) (fieldConverterFunc, error) {
	if scale != 0.0 {
		switch inType {
		case "INT16":
			return determineConverterABCDI16Scale(outType, scale)
		case "UINT16":
			return determineConverterABCDU16Scale(outType, scale)
		case "INT32":
			return determineConverterABCDI32Scale(outType, scale)
		case "UINT32":
			return determineConverterABCDU32Scale(outType, scale)
		case "INT64":
			return determineConverterABCDI64Scale(outType, scale)
		case "UINT64":
			return determineConverterABCDU64Scale(outType, scale)
		case "FLOAT32":
			return determineConverterABCDF32Scale(outType, scale)
		case "FLOAT64":
			return determineConverterABCDF64Scale(outType, scale)
		}
		return nil, fmt.Errorf("invalid input data-type: %s", inType)
	}
	switch inType {
	case "INT16":
		return determineConverterABCDI16NoScale(outType)
	case "UINT16":
		return determineConverterABCDU16NoScale(outType)
	case "INT32":
		return determineConverterABCDI32NoScale(outType)
	case "UINT32":
		return determineConverterABCDU32NoScale(outType)
	case "INT64":
		return determineConverterABCDI64NoScale(outType)
	case "UINT64":
		return determineConverterABCDU64NoScale(outType)
	case "FLOAT32":
		return determineConverterABCDF32NoScale(outType)
	case "FLOAT64":
		return determineConverterABCDF64NoScale(outType)
	}
	return nil, fmt.Errorf("invalid input data-type: %s", inType)
}

func determineConverterBADC(inType, outType string, scale float64) (fieldConverterFunc, error) {
	if scale != 0.0 {
		switch inType {
		case "INT32":
			return determineConverterBADCI32Scale(outType, scale)
		case "UINT32":
			return determineConverterBADCU32Scale(outType, scale)
		case "INT64":
			return determineConverterBADCI64Scale(outType, scale)
		case "UINT64":
			return determineConverterBADCU64Scale(outType, scale)
		case "FLOAT32":
			return determineConverterBADCF32Scale(outType, scale)
		case "FLOAT64":
			return determineConverterBADCF64Scale(outType, scale)
		}
		return nil, fmt.Errorf("invalid input data-type: %s", inType)
	}
	switch inType {
	case "INT32":
		return determineConverterBADCI32NoScale(outType)
	case "UINT32":
		return determineConverterBADCU32NoScale(outType)
	case "INT64":
		return determineConverterBADCI64NoScale(outType)
	case "UINT64":
		return determineConverterBADCU64NoScale(outType)
	case "FLOAT32":
		return determineConverterBADCF32NoScale(outType)
	case "FLOAT64":
		return determineConverterBADCF64NoScale(outType)
	}
	return nil, fmt.Errorf("invalid input data-type: %s", inType)
}

func determineConverterCDAB(inType, outType string, scale float64) (fieldConverterFunc, error) {
	if scale != 0.0 {
		switch inType {
		case "INT32":
			return determineConverterCDABI32Scale(outType, scale)
		case "UINT32":
			return determineConverterCDABU32Scale(outType, scale)
		case "INT64":
			return determineConverterCDABI64Scale(outType, scale)
		case "UINT64":
			return determineConverterCDABU64Scale(outType, scale)
		case "FLOAT32":
			return determineConverterCDABF32Scale(outType, scale)
		case "FLOAT64":
			return determineConverterCDABF64Scale(outType, scale)
		}
		return nil, fmt.Errorf("invalid input data-type: %s", inType)
	}
	switch inType {
	case "INT32":
		return determineConverterCDABI32NoScale(outType)
	case "UINT32":
		return determineConverterCDABU32NoScale(outType)
	case "INT64":
		return determineConverterCDABI64NoScale(outType)
	case "UINT64":
		return determineConverterCDABU64NoScale(outType)
	case "FLOAT32":
		return determineConverterCDABF32NoScale(outType)
	case "FLOAT64":
		return determineConverterCDABF64NoScale(outType)
	}
	return nil, fmt.Errorf("invalid input data-type: %s", inType)
}

func determineConverterDCBA(inType, outType string, scale float64) (fieldConverterFunc, error) {
	if scale != 0.0 {
		switch inType {
		case "INT16":
			return determineConverterDCBAI16Scale(outType, scale)
		case "UINT16":
			return determineConverterDCBAU16Scale(outType, scale)
		case "INT32":
			return determineConverterDCBAI32Scale(outType, scale)
		case "UINT32":
			return determineConverterDCBAU32Scale(outType, scale)
		case "INT64":
			return determineConverterDCBAI64Scale(outType, scale)
		case "UINT64":
			return determineConverterDCBAU64Scale(outType, scale)
		case "FLOAT32":
			return determineConverterDCBAF32Scale(outType, scale)
		case "FLOAT64":
			return determineConverterDCBAF64Scale(outType, scale)
		}
		return nil, fmt.Errorf("invalid input data-type: %s", inType)
	}
	switch inType {
	case "INT16":
		return determineConverterDCBAI16NoScale(outType)
	case "UINT16":
		return determineConverterDCBAU16NoScale(outType)
	case "INT32":
		return determineConverterDCBAI32NoScale(outType)
	case "UINT32":
		return determineConverterDCBAU32NoScale(outType)
	case "INT64":
		return determineConverterDCBAI64NoScale(outType)
	case "UINT64":
		return determineConverterDCBAU64NoScale(outType)
	case "FLOAT32":
		return determineConverterDCBAF32NoScale(outType)
	case "FLOAT64":
		return determineConverterDCBAF64NoScale(outType)
	}
	return nil, fmt.Errorf("invalid input data-type: %s", inType)
}
