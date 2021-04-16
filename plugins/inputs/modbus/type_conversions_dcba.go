package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
)

// I16 - DCBA
func determineConverterDCBAI16NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int16(binary.LittleEndian.Uint16(b))
			return int64(in)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int16(binary.LittleEndian.Uint16(b))
			return uint64(in)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int16(binary.LittleEndian.Uint16(b))
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterDCBAI16Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int16(binary.LittleEndian.Uint16(b))
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int16(binary.LittleEndian.Uint16(b))
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int16(binary.LittleEndian.Uint16(b))
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U16 - DCBA
func determineConverterDCBAU16NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint16(b)
			return int64(in)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint16(b)
			return uint64(in)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint16(b)
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterDCBAU16Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint16(b)
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint16(b)
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint16(b)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// I32 - DCBA
func determineConverterDCBAI32NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int32(binary.LittleEndian.Uint32(b))
			return int64(in)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int32(binary.LittleEndian.Uint32(b))
			return uint64(in)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int32(binary.LittleEndian.Uint32(b))
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterDCBAI32Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int32(binary.LittleEndian.Uint32(b))
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int32(binary.LittleEndian.Uint32(b))
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int32(binary.LittleEndian.Uint32(b))
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U32 - DCBA
func determineConverterDCBAU32NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint32(b)
			return int64(in)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint32(b)
			return uint64(in)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint32(b)
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterDCBAU32Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint32(b)
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint32(b)
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint32(b)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// I64 - DCBA
func determineConverterDCBAI64NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int64(binary.LittleEndian.Uint64(b))
			return in
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int64(binary.LittleEndian.Uint64(b))
			return uint64(in)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int64(binary.LittleEndian.Uint64(b))
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterDCBAI64Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int64(binary.LittleEndian.Uint64(b))
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int64(binary.LittleEndian.Uint64(b))
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int64(binary.LittleEndian.Uint64(b))
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U64 - DCBA
func determineConverterDCBAU64NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint64(b)
			return int64(in)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint64(b)
			return in
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint64(b)
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterDCBAU64Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint64(b)
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint64(b)
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binary.LittleEndian.Uint64(b)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// F32 - DCBA
func determineConverterDCBAF32NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "FLOAT64":
		return func(b []byte) interface{} {
			raw := binary.LittleEndian.Uint32(b)
			in := math.Float32frombits(raw)
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterDCBAF32Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "FLOAT64":
		return func(b []byte) interface{} {
			raw := binary.LittleEndian.Uint32(b)
			in := math.Float32frombits(raw)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// F64 - DCBA
func determineConverterDCBAF64NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "FLOAT64":
		return func(b []byte) interface{} {
			raw := binary.LittleEndian.Uint64(b)
			in := math.Float64frombits(raw)
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterDCBAF64Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "FLOAT64":
		return func(b []byte) interface{} {
			raw := binary.LittleEndian.Uint64(b)
			in := math.Float64frombits(raw)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}
