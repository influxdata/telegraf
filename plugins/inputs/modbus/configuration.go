package modbus

import (
	"fmt"
)

const (
	maxQuantityDiscreteInput    = uint16(2000)
	maxQuantityCoils            = uint16(2000)
	maxQuantityInputRegisters   = uint16(125)
	maxQuantityHoldingRegisters = uint16(125)
)

type Configuration interface {
	Check() error
	Process(m *Modbus) error
}

func removeDuplicates(elements []uint16) []uint16 {
	encountered := map[uint16]bool{}
	result := []uint16{}

	for _, addr := range elements {
		if !encountered[addr] {
			encountered[addr] = true
			result = append(result, addr)
		}
	}

	return result
}

func normalizeInputDatatype(dataType string) (string, error) {
	switch dataType {
	case "INT16", "UINT16", "INT32", "UINT32", "INT64", "UINT64", "FLOAT32", "FLOAT64":
		return dataType, nil
	}
	return "unknown", fmt.Errorf("unknown type %q", dataType)
}

func normalizeOutputDatatype(dataType string) (string, error) {
	switch dataType {
	case "INT16", "INT32", "INT64":
		return "INT64", nil
	case "UINT16", "UINT32", "UINT64":
		return "UINT64", nil
	case "FLOAT32", "FLOAT64":
		return "FLOAT64", nil
	}
	return "unknown", fmt.Errorf("unknown type %q", dataType)
}

func normalizeByteOrder(byteOrder string) (string, error) {
	switch byteOrder {
	case "ABCD", "MSW-BE", "MSW": // Big endian (Motorola)
		return "ABCD", nil
	case "BADC", "MSW-LE": // Big endian with bytes swapped
		return "BADC", nil
	case "CDAB", "LSW-BE": // Little endian with bytes swapped
		return "CDAB", nil
	case "DCBA", "LSW-LE", "LSW": // Little endian (Intel)
		return "DCBA", nil
	}
	return "unknown", fmt.Errorf("unknown byte-order %q", byteOrder)
}
