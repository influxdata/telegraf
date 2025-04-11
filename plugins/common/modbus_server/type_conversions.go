package modbus_server

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type fieldConverterFunc func([]byte) interface{}

type convertToBytes func(any) []byte

func endiannessConverterToBytes(byteOrder string) (convertToBytes, error) {
	switch byteOrder {
	case "ABCD", "CDAB": // Big endian (Motorola)
		return func(b any) []byte {
			switch v := b.(type) {
			case string:
				return []byte(v)
			default:
				buf := new(bytes.Buffer)
				err := binary.Write(buf, binary.BigEndian, b)
				if err != nil {
					return nil
				}
				return buf.Bytes()
			}
		}, nil
	case "DCBA", "BADC": // Little endian (Intel)
		return func(b any) []byte {
			switch v := b.(type) {
			case string:
				return []byte(v)
			default:
				buf := new(bytes.Buffer)
				err := binary.Write(buf, binary.LittleEndian, b)
				if err != nil {
					return nil
				}
				return buf.Bytes()
			}
		}, nil
	}
	return nil, fmt.Errorf("invalid byte-order: %s", byteOrder)
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
	case "FLOAT64":
		return determineConverterF64(outType, byteOrder)
	}
	return nil, fmt.Errorf("invalid input data-type: %s", inType)
}
