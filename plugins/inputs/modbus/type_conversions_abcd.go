package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
)

// I16 - ABCD
func determineConverterABCDI16NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int16(binary.BigEndian.Uint16(b))
			return int64(in)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int16(binary.BigEndian.Uint16(b))
			return uint64(in)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int16(binary.BigEndian.Uint16(b))
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterABCDI16Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int16(binary.BigEndian.Uint16(b))
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int16(binary.BigEndian.Uint16(b))
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int16(binary.BigEndian.Uint16(b))
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U16 - ABCD
func determineConverterABCDU16NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint16(b)
			return int64(in)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint16(b)
			return uint64(in)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint16(b)
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterABCDU16Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint16(b)
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint16(b)
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint16(b)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// I32 - ABCD
func determineConverterABCDI32NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int32(binary.BigEndian.Uint32(b))
			return int64(in)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int32(binary.BigEndian.Uint32(b))
			return uint64(in)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int32(binary.BigEndian.Uint32(b))
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterABCDI32Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int32(binary.BigEndian.Uint32(b))
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int32(binary.BigEndian.Uint32(b))
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int32(binary.BigEndian.Uint32(b))
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U32 - ABCD
func determineConverterABCDU32NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint32(b)
			return int64(in)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint32(b)
			return uint64(in)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint32(b)
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterABCDU32Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint32(b)
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint32(b)
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint32(b)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// I64 - ABCD
func determineConverterABCDI64NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int64(binary.BigEndian.Uint64(b))
			return in
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int64(binary.BigEndian.Uint64(b))
			return uint64(in)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int64(binary.BigEndian.Uint64(b))
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterABCDI64Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int64(binary.BigEndian.Uint64(b))
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int64(binary.BigEndian.Uint64(b))
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int64(binary.BigEndian.Uint64(b))
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U64 - ABCD
func determineConverterABCDU64NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint64(b)
			return int64(in)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint64(b)
			return in
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint64(b)
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterABCDU64Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint64(b)
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint64(b)
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binary.BigEndian.Uint64(b)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// F32 - ABCD
func determineConverterABCDF32NoScale(outType string) (fieldConverterFunc, error) {
	if outType != "FLOAT64" {
		return nil, fmt.Errorf("invalid output data-type: %s", outType)
	}

	return func(b []byte) interface{} {
		raw := binary.BigEndian.Uint32(b)
		in := math.Float32frombits(raw)
		return float64(in)
	}, nil
}

func determineConverterABCDF32Scale(outType string, scale float64) (fieldConverterFunc, error) {
	if outType != "FLOAT64" {
		return nil, fmt.Errorf("invalid output data-type: %s", outType)
	}

	return func(b []byte) interface{} {
		raw := binary.BigEndian.Uint32(b)
		in := math.Float32frombits(raw)
		return float64(in) * scale
	}, nil
}

// F64 - ABCD
func determineConverterABCDF64NoScale(outType string) (fieldConverterFunc, error) {
	if outType != "FLOAT64" {
		return nil, fmt.Errorf("invalid output data-type: %s", outType)
	}

	return func(b []byte) interface{} {
		raw := binary.BigEndian.Uint64(b)
		return math.Float64frombits(raw)
	}, nil
}

func determineConverterABCDF64Scale(outType string, scale float64) (fieldConverterFunc, error) {
	if outType != "FLOAT64" {
		return nil, fmt.Errorf("invalid output data-type: %s", outType)
	}

	return func(b []byte) interface{} {
		raw := binary.BigEndian.Uint64(b)
		in := math.Float64frombits(raw)
		return in * scale
	}, nil
}
