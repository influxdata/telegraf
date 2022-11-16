package modbus

import (
	"fmt"
)

func endianessIndex8(byteOrder string, low bool) (int, error) {
	switch byteOrder {
	case "ABCD": // Big endian (Motorola)
		if low {
			return 1, nil
		}
		return 0, nil
	case "DCBA": // Little endian (Intel)
		if low {
			return 0, nil
		}
		return 1, nil
	}
	return -1, fmt.Errorf("invalid byte-order: %s", byteOrder)
}

// I8 lower byte - no scale
func determineConverterI8L(outType, byteOrder string) (fieldConverterFunc, error) {
	idx, err := endianessIndex8(byteOrder, true)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			return int8(b[idx])
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			return int64(int8(b[idx]))
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			return uint64(int8(b[idx]))
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			return float64(int8(b[idx]))
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// I8 higher byte - no scale
func determineConverterI8H(outType, byteOrder string) (fieldConverterFunc, error) {
	idx, err := endianessIndex8(byteOrder, false)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			return int8(b[idx])
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			return int64(int8(b[idx]))
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			return uint64(int8(b[idx]))
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			return float64(int8(b[idx]))
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U8 lower byte - no scale
func determineConverterU8L(outType, byteOrder string) (fieldConverterFunc, error) {
	idx, err := endianessIndex8(byteOrder, true)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			return b[idx]
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			return int64(b[idx])
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			return uint64(b[idx])
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			return float64(b[idx])
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U8 higher byte - no scale
func determineConverterU8H(outType, byteOrder string) (fieldConverterFunc, error) {
	idx, err := endianessIndex8(byteOrder, false)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			return b[idx]
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			return int64(b[idx])
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			return uint64(b[idx])
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			return float64(b[idx])
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// I8 lower byte - scale
func determineConverterI8LScale(outType, byteOrder string, scale float64) (fieldConverterFunc, error) {
	idx, err := endianessIndex8(byteOrder, true)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			in := int8(b[idx])
			return int8(float64(in) * scale)
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			in := int8(b[idx])
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int8(b[idx])
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int8(b[idx])
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// I8 higher byte - scale
func determineConverterI8HScale(outType, byteOrder string, scale float64) (fieldConverterFunc, error) {
	idx, err := endianessIndex8(byteOrder, false)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			in := int8(b[idx])
			return int8(float64(in) * scale)
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			in := int8(b[idx])
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int8(b[idx])
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int8(b[idx])
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U8 lower byte - scale
func determineConverterU8LScale(outType, byteOrder string, scale float64) (fieldConverterFunc, error) {
	idx, err := endianessIndex8(byteOrder, true)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			return uint8(float64(b[idx]) * scale)
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			return int64(float64(b[idx]) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			return uint64(float64(b[idx]) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			return float64(b[idx]) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U8 higher byte - scale
func determineConverterU8HScale(outType, byteOrder string, scale float64) (fieldConverterFunc, error) {
	idx, err := endianessIndex8(byteOrder, false)
	if err != nil {
		return nil, err
	}

	switch outType {
	case "native":
		return func(b []byte) interface{} {
			return uint8(float64(b[idx]) * scale)
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			return int64(float64(b[idx]) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			return uint64(float64(b[idx]) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			return float64(b[idx]) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}
