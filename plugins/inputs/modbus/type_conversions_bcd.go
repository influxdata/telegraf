package modbus

import (
	"fmt"
)

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

// endiannessConverterBCD returns a function reading the raw register value of a
// BCD type spanning the given number of registers (1, 2 or 4).
func endiannessConverterBCD(registers int, byteOrder string) (func([]byte) uint64, error) {
	switch registers {
	case 1:
		tohost, err := endiannessConverter16(byteOrder)
		if err != nil {
			return nil, err
		}
		return func(b []byte) uint64 { return uint64(tohost(b)) }, nil
	case 2:
		tohost, err := endiannessConverter32(byteOrder)
		if err != nil {
			return nil, err
		}
		return func(b []byte) uint64 { return uint64(tohost(b)) }, nil
	case 4:
		return endiannessConverter64(byteOrder)
	}
	return nil, fmt.Errorf("invalid number of registers %d for BCD", registers)
}

// nativeBCD returns the decoded value as the unsigned integer type of the input width.
func nativeBCD(registers int, value uint64) any {
	switch registers {
	case 1:
		return uint16(value)
	case 2:
		return uint32(value)
	}
	return value
}

// BCD - no scale
// Values that are not valid BCD return nil, so the field is left out of the metric.
func determineConverterBCD(registers int, outType, byteOrder string) (fieldConverterFunc, error) {
	tohost, err := endiannessConverterBCD(registers, byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) any {
			if v, ok := decodeBCD(tohost(b)); ok {
				return nativeBCD(registers, v)
			}
			return nil
		}, nil
	case "INT64":
		return func(b []byte) any {
			if v, ok := decodeBCD(tohost(b)); ok {
				return int64(v)
			}
			return nil
		}, nil
	case "UINT64":
		return func(b []byte) any {
			if v, ok := decodeBCD(tohost(b)); ok {
				return v
			}
			return nil
		}, nil
	case "FLOAT64":
		return func(b []byte) any {
			if v, ok := decodeBCD(tohost(b)); ok {
				return float64(v)
			}
			return nil
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// BCD - scale
// Values that are not valid BCD return nil, so the field is left out of the metric.
func determineConverterBCDScale(registers int, outType, byteOrder string, scale float64) (fieldConverterFunc, error) {
	tohost, err := endiannessConverterBCD(registers, byteOrder)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) any {
			if v, ok := decodeBCD(tohost(b)); ok {
				return nativeBCD(registers, uint64(float64(v)*scale))
			}
			return nil
		}, nil
	case "INT64":
		return func(b []byte) any {
			if v, ok := decodeBCD(tohost(b)); ok {
				return int64(float64(v) * scale)
			}
			return nil
		}, nil
	case "UINT64":
		return func(b []byte) any {
			if v, ok := decodeBCD(tohost(b)); ok {
				return uint64(float64(v) * scale)
			}
			return nil
		}, nil
	case "FLOAT64":
		return func(b []byte) any {
			if v, ok := decodeBCD(tohost(b)); ok {
				return float64(v) * scale
			}
			return nil
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}
