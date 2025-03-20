//go:generate ../../../tools/readme_config_includer/generator
package modbus_server

import (
	"fmt"
	"github.com/x448/float16"
	"math"
)

// Define the size (in addresses) for supported types
var typeSizes = map[string]uint16{
	"BIT":     1, // 1 bit
	"INT8L":   1, // 1 byte
	"INT8H":   1, // 1 byte
	"UINT8L":  1, // 1 byte
	"UINT8H":  1, // 1 byte
	"FLOAT16": 1, // 2 bytes
	"INT16":   1, // 2 bytes
	"UINT16":  1, // 2 bytes
	"FLOAT32": 2, // 4 bytes
	"INT32":   2, // 4 bytes
	"UINT32":  2, // 4 bytes
	"INT64":   4, // 8 bytes
	"UINT64":  4, // 8 bytes
	"FLOAT64": 4, // 8 bytes
}

// MemoryEntry represents a single memory entry with an address and type
type MemoryEntry struct {
	Address          uint16
	CoilInitialValue bool
	Type             string
	Register         string
	Scale            float64
	Bit              uint8
	Length           uint16
	Measurement      string
	HashID           uint64
	Field            string
}

type MemoryLayout []MemoryEntry

// parseEntry calculates the start and end address range for an entry
func (entry MemoryEntry) getBounds() (uint16, uint16, uint8, error) {
	if entry.Register == "coil" {
		return entry.Address, entry.Address + 1, 0, nil
	}

	if entry.Register != "coil" && entry.Length > 0 {
		return entry.Address, entry.Address + entry.Length, 0, nil
	}

	size, ok := typeSizes[entry.Type]
	if !ok {
		return 0, 0, 0, fmt.Errorf("unsupported type: %s", entry.Type)
	}
	start := entry.Address
	end := entry.Address + size
	bit := entry.Bit
	return start, end, bit, nil
}

// HasOverlap checks a list of memory entries for overlaps
func (entries MemoryLayout) HasOverlap() (bool, []string, error) {
	usedAddresses := make(map[int]bool) // Map to track used addresses
	overlaps := []string{}
	usedBits := make(map[int][]bool) // Map to track used bits

	for _, entry := range entries {
		start, end, bit, err := entry.getBounds()
		if err != nil {
			return false, overlaps, err
		}
		//check for BIT overlap
		if entry.Register != "coil" && entry.Type == "BIT" {
			bitIndex := int(entry.Bit)
			if bitIndex > 15 {
				overlaps = append(overlaps, fmt.Sprintf("Entry at address %d overlaps with type %s", entry.Address, entry.Type))
				return true, overlaps, fmt.Errorf("bit index %d out of range", bitIndex)
			}
			if usedBits[int(entry.Address)] == nil {
				usedBits[int(entry.Address)] = make([]bool, 16)
			} else if usedBits[int(entry.Address)][bitIndex] == true {
				overlaps = append(overlaps, fmt.Sprintf("Entry at address %d overlaps with type %s", entry.Address, entry.Type))
			} else {
				usedBits[int(entry.Address)][bitIndex] = true
			}
		}
		// Check for overlaps
		for addr := start; addr < end; addr++ {
			if bit == 0 && usedAddresses[int(addr)] {
				overlaps = append(overlaps, fmt.Sprintf("Entry at address %d overlaps with type %s", entry.Address, entry.Type))
			}
			usedAddresses[int(addr)] = true
		}
	}
	return len(overlaps) > 0, overlaps, nil
}

func (entries MemoryLayout) getMemOffsets() (uint16, uint16, uint16, uint16) {
	var coilOffset, registerOffset uint16 = math.MaxUint16, math.MaxUint16

	var maxAddressCoil, maxAddressRegister uint16 = 0, 0

	for _, entry := range entries {
		if entry.Register == "coil" {
			if coilOffset > entry.Address {
				coilOffset = entry.Address
			}
			if maxAddressCoil < entry.Address {
				maxAddressCoil = entry.Address
			}
		} else {
			if registerOffset > entry.Address {
				registerOffset = entry.Address
			}
			if maxAddressRegister < entry.Address {
				maxAddressRegister = entry.Address + typeSizes[entry.Type] - 1
			}
		}
	}
	return coilOffset, registerOffset, maxAddressCoil, maxAddressRegister
}

func ParseMemory(byteOrder string, entry MemoryEntry, coilOffset, registerOffset uint16, coils []bool, registers []uint16) (any, error) {
	if entry.Register == "coil" {
		return coils[entry.Address-coilOffset], nil
	} else {

		startAddr := entry.Address - registerOffset
		endAddr := startAddr + entry.Length - 1

		if entry.Type != "STRING" {
			endAddr = startAddr + typeSizes[entry.Type] - 1
		}

		contents := registers[startAddr : endAddr+1]
		converter, err := determineConverter(entry.Type, byteOrder, "native", entry.Scale, entry.Bit, "")
		if err != nil {
			return nil, err
		}
		converterToBytes, err := endiannessConverterToBytes(byteOrder)
		if err != nil {
			return nil, err
		}
		bytesValue := []byte{}
		for _, content := range contents {
			bytesValue = append(bytesValue, converterToBytes(content)...)
		}
		value := converter(bytesValue)

		if entry.Type == "BIT" {
			value = value != uint8(0)
		}

		return value, nil
	}
}

func (entries MemoryLayout) GetCoilsAndRegisters() ([]bool, []uint16, uint16, uint16) {
	coilOffset, registerOffset, maxCoilAddr, maxRegisterAddr := entries.getMemOffsets()

	coils := make([]bool, maxCoilAddr-coilOffset+1)
	registers := make([]uint16, maxRegisterAddr-registerOffset+1)

	for _, entry := range entries {
		if entry.Register == "coil" {
			coils[entry.Address-coilOffset] = false
		} else {
			for i := uint16(0); i < typeSizes[entry.Type]; i++ {
				registers[entry.Address-registerOffset+i] = 0
			}
		}
	}
	return coils, registers, coilOffset, registerOffset
}

func (entries MemoryLayout) GetMemoryMappedByHashID() (map[uint64]map[string]MemoryEntry, error) {
	memoryMap := make(map[uint64]map[string]MemoryEntry)
	for _, entry := range entries {
		if _, ok := memoryMap[entry.HashID]; ok {
			continue
		}
		memoryMap[entry.HashID] = make(map[string]MemoryEntry)
	}
	for _, entry := range entries {
		memoryMap[entry.HashID][entry.Field] = entry
	}
	return memoryMap, nil
}

// cast a 64-bit number to the specified type
func castToType(value any, valueType string) any {
	type casters map[string]func(any) any

	casterMap := casters{
		"INT8L":   func(v any) any { return int8(v.(int64)) },
		"UINT8L":  func(v any) any { return uint8(v.(uint64)) },
		"INT8H":   func(v any) any { return int8(v.(int64)) },
		"UINT8H":  func(v any) any { return uint8(v.(uint64)) },
		"INT16":   func(v any) any { return int16(v.(int64)) },
		"UINT16":  func(v any) any { return uint16(v.(uint64)) },
		"FLOAT16": func(v any) any { return float16.Fromfloat32(float32(v.(float64))) },
		"FLOAT32": func(v any) any { return float32(v.(float64)) },
		"INT32":   func(v any) any { return int32(v.(int64)) },
		"UINT32":  func(v any) any { return uint32(v.(uint64)) },
		"INT64":   func(v any) any { return v.(int64) },
		"UINT64":  func(v any) any { return v.(uint64) },
		"FLOAT64": func(v any) any { return v.(float64) },
		"STRING":  func(v any) any { return v.(string) },
	}

	if castFunc, exists := casterMap[valueType]; exists {
		return castFunc(value)
	}
	return nil
}

func ParseMetric(byteOrder string, value any, valueType string, scale float64) ([]uint16, error) {
	value = castToType(value, valueType)
	if value == nil {
		return nil, fmt.Errorf("unsupported type: %s", valueType)
	}

	// ignore endianness for strings
	if valueType == "STRING" {
		byteOrder = "ABCD"
	}

	converter, err := determineConverter("UINT16", byteOrder, "native", scale, 0, "")
	if err != nil {
		return nil, err
	}
	converterToBytes, err := endiannessConverterToBytes(byteOrder)
	if err != nil {
		return nil, err
	}
	bytesValue := converterToBytes(value)
	// Add padding for odd-length strings
	if valueType == "STRING" && len(bytesValue)%2 != 0 {
		bytesValue = append(bytesValue, 0)
	}

	registerValues := []uint16{}
	for i := 0; i < len(bytesValue); i++ {
		// Convert the 8-bit values to uint16
		if valueType == "INT8L" || valueType == "UINT8L" {
			registerValues = append(registerValues, uint16(bytesValue[i]))
		} else if valueType == "INT8H" || valueType == "UINT8H" {
			registerValues = append(registerValues, uint16(bytesValue[i])<<8) // Shift the byte to the high position
		} else { // convert >= 16-bit values to uint16
			if i+1 < len(bytesValue) {
				registerValues = append(registerValues, converter(bytesValue[i:i+2]).(uint16))
				i++ // Skip the next byte since we processed two bytes
			} else {
				return nil, fmt.Errorf("unexpected end of bytesValue for %s", valueType)
			}
		}
	}
	return registerValues, nil
}
